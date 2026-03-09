package indexer

import (
	"context"
	"fmt"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"time"

	codedb "github.com/tab58/code-context/internal/clients/code_db"
	"github.com/tab58/go-ormql/pkg/client"
)

// batchSize is the number of nodes/edges per FalkorDB mutation call.
// Kept small to avoid OOM-killing FalkorDB in memory-constrained environments.
// UNWIND + MATCH creates cartesian products; smaller batches limit intermediate
// result set size.
const batchSize = 10

// GraphQL queries and mutations for structural graph persistence via Client().Execute().
const (
	gqlMergeRepositorys = `mutation($input: [RepositoryMergeInput!]!) {
  mergeRepositorys(input: $input) { repositorys { id name } }
}`

	gqlQueryFolders = `query($where: FolderWhere) {
  folders(where: $where) { path lastUpdated }
}`

	gqlQueryFiles = `query($where: FileWhere) {
  files(where: $where) { path lastUpdated }
}`

	gqlMergeFolders = `mutation($input: [FolderMergeInput!]!) {
  mergeFolders(input: $input) { folders { id path } }
}`

	gqlMergeFiles = `mutation($input: [FileMergeInput!]!) {
  mergeFiles(input: $input) { files { id path } }
}`

	gqlConnectRepositoryFolders = `mutation($input: [ConnectRepositoryFoldersInput!]!) {
  connectRepositoryFolders(input: $input) { relationshipsCreated }
}`

	gqlConnectRepositoryFiles = `mutation($input: [ConnectRepositoryFilesInput!]!) {
  connectRepositoryFiles(input: $input) { relationshipsCreated }
}`

	gqlConnectFolderSubfolders = `mutation($input: [ConnectFolderSubfoldersInput!]!) {
  connectFolderSubfolders(input: $input) { relationshipsCreated }
}`

	gqlConnectFolderFiles = `mutation($input: [ConnectFolderFilesInput!]!) {
  connectFolderFiles(input: $input) { relationshipsCreated }
}`

	gqlConnectFolderRepository = `mutation($input: [ConnectFolderRepositoryInput!]!) {
  connectFolderRepository(input: $input) { relationshipsCreated }
}`

	gqlConnectFileRepository = `mutation($input: [ConnectFileRepositoryInput!]!) {
  connectFileRepository(input: $input) { relationshipsCreated }
}`
)

// pendingFolder is a folder discovered during the walk, pending creation in FalkorDB.
type pendingFolder struct {
	Path       string
	ParentPath string
	ModTime    time.Time
}

// pendingFile is a file discovered during the walk, pending creation in FalkorDB.
type pendingFile struct {
	Path       string
	ParentPath string
	Language   string
	LineCount  int
	ModTime    time.Time
}

// ProgressFunc is called by pipeline stages to report progress.
type ProgressFunc func(stage, message string)

// IndexOption configures an IndexRepository call.
type IndexOption func(*indexOptions)

type indexOptions struct {
	progress ProgressFunc
}

// WithProgress returns an IndexOption that sets the progress callback.
func WithProgress(fn ProgressFunc) IndexOption {
	return func(o *indexOptions) {
		o.progress = fn
	}
}

// IndexResult reports what was indexed during a repository scan.
type IndexResult struct {
	RepoID         string
	FilesIndexed   int
	FoldersIndexed int
	FilesSkipped   int      // skipped due to .gitignore, binary, or unchanged
	FilePaths      []string // absolute paths of all indexed files
	Errors         []error
}

// Indexer walks directories and creates Repository/Folder/File nodes in FalkorDB.
type Indexer struct {
	db *codedb.CodeDB
	// indexed tracks path -> modTime for incremental re-indexing, scoped per
	// repository name. Supplements the database query for newly created nodes
	// that may not yet be returned by queryExistingNodes.
	indexed map[string]map[string]time.Time // repoName -> relPath -> modTime
}

// isUnchanged returns true if the given path exists in existingNodes with a
// lastUpdated time at or after modTime (compared at second granularity).
func isUnchanged(existingNodes map[string]time.Time, relPath string, modTime time.Time) bool {
	if existing, ok := existingNodes[relPath]; ok {
		return !modTime.After(existing.Truncate(time.Second))
	}
	return false
}

// parentRelPath returns the parent directory of relPath, normalizing "." to ""
// so root-level items get an empty parent (indicating they belong to the repo).
func parentRelPath(relPath string) string {
	p := filepath.Dir(relPath)
	if p == "." {
		return ""
	}
	return p
}

// walkContext holds mutable state accumulated during the directory walk.
type walkContext struct {
	repoPath       string
	result         *IndexResult
	matcher        *GitIgnoreMatcher
	existingNodes  map[string]time.Time
	hasPersistence bool
	pendingFolders []pendingFolder
	pendingFiles   []pendingFile
	progress       ProgressFunc
}

// handleDir processes a directory entry during the walk. Returns fs.SkipDir
// if the directory should be skipped (gitignore, symlink).
func (wc *walkContext) handleDir(path, relPath string, d fs.DirEntry) error {
	if wc.matcher.ShouldIgnore(relPath, true) {
		log.Printf("[DEBUG] SKIP DIR (gitignore): %s", relPath)
		return fs.SkipDir
	}
	wc.matcher.EnterDirectory(path)

	if wc.hasPersistence {
		info, infoErr := d.Info()
		if infoErr != nil {
			wc.result.Errors = append(wc.result.Errors, infoErr)
			return nil
		}
		modTime := info.ModTime().Truncate(time.Second)

		if isUnchanged(wc.existingNodes, relPath, modTime) {
			return nil
		}

		wc.pendingFolders = append(wc.pendingFolders, pendingFolder{
			Path:       relPath,
			ParentPath: parentRelPath(relPath),
			ModTime:    modTime,
		})
	}

	wc.result.FoldersIndexed++
	if wc.progress != nil {
		wc.progress("indexing", relPath)
	}
	return nil
}

// handleFile processes a file entry during the walk. Skips binary files and
// files matching gitignore patterns.
func (wc *walkContext) handleFile(path, relPath string, d fs.DirEntry) error {
	if wc.matcher.ShouldIgnore(relPath, false) {
		log.Printf("[DEBUG] SKIP FILE (gitignore): %s", relPath)
		wc.result.FilesSkipped++
		return nil
	}

	isBin, binErr := IsBinary(path)
	if binErr != nil {
		wc.result.Errors = append(wc.result.Errors, binErr)
		return nil
	}
	if isBin {
		log.Printf("[DEBUG] SKIP FILE (binary): %s", relPath)
		wc.result.FilesSkipped++
		return nil
	}

	if wc.hasPersistence {
		info, infoErr := d.Info()
		if infoErr != nil {
			wc.result.Errors = append(wc.result.Errors, infoErr)
			return nil
		}
		modTime := info.ModTime().Truncate(time.Second)

		if isUnchanged(wc.existingNodes, relPath, modTime) {
			wc.result.FilesSkipped++
			return nil
		}

		lang := DetectLanguage(path)
		lines, _ := CountLines(path)

		wc.pendingFiles = append(wc.pendingFiles, pendingFile{
			Path:       relPath,
			ParentPath: parentRelPath(relPath),
			Language:   lang,
			LineCount:  lines,
			ModTime:    modTime,
		})
	}

	wc.result.FilesIndexed++
	wc.result.FilePaths = append(wc.result.FilePaths, path)
	if wc.progress != nil {
		wc.progress("indexing", relPath)
	}
	return nil
}

// NewIndexer creates an Indexer backed by the given CodeDB.
func NewIndexer(db *codedb.CodeDB) *Indexer {
	return &Indexer{db: db, indexed: make(map[string]map[string]time.Time)}
}

// IndexRepository scans a local directory and creates/updates structural graph
// nodes. If a Repository node already exists for this path, it is updated.
// Returns an IndexResult summarizing what was indexed.
func (idx *Indexer) IndexRepository(ctx context.Context, repoPath string, opts ...IndexOption) (IndexResult, error) {
	var options indexOptions
	for _, opt := range opts {
		opt(&options)
	}

	fi, err := os.Stat(repoPath)
	if err != nil {
		return IndexResult{}, fmt.Errorf("indexer: path does not exist: %w", err)
	}
	if !fi.IsDir() {
		return IndexResult{}, fmt.Errorf("indexer: path is not a directory: %s", repoPath)
	}

	repoName := filepath.Base(repoPath)
	result := IndexResult{
		RepoID: repoName,
	}

	// Persistence: upsert the Repository node (if db is available)
	hasPersistence := idx.db != nil
	existingNodes := make(map[string]time.Time)
	var c *client.Client

	if hasPersistence {
		var forRepoErr error
		c, forRepoErr = idx.db.ForRepo(ctx, repoName)
		if forRepoErr != nil {
			return result, fmt.Errorf("indexer: %w", forRepoErr)
		}

		if _, err := idx.upsertRepository(ctx, c, repoName, repoPath); err != nil {
			return result, fmt.Errorf("indexer: %w", err)
		}

		existingNodes, err = idx.queryExistingNodes(ctx, c, repoName)
		if err != nil {
			return result, fmt.Errorf("indexer: %w", err)
		}
	}

	// Merge in-memory index cache (supplements database query)
	if repoCache, ok := idx.indexed[repoName]; ok {
		for k, v := range repoCache {
			if _, ok := existingNodes[k]; !ok {
				existingNodes[k] = v
			}
		}
	}

	matcher, err := NewGitIgnoreMatcher(repoPath)
	if err != nil {
		return IndexResult{}, fmt.Errorf("indexer: failed to load .gitignore: %w", err)
	}

	wc := &walkContext{
		repoPath:       repoPath,
		result:         &result,
		matcher:        matcher,
		existingNodes:  existingNodes,
		hasPersistence: hasPersistence,
		progress:       options.progress,
	}

	err = filepath.WalkDir(repoPath, func(path string, d fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			wc.result.Errors = append(wc.result.Errors, walkErr)
			return nil
		}

		relPath, _ := filepath.Rel(repoPath, path)
		if relPath == "." {
			return nil
		}

		// Check for symlinks
		isLink, linkErr := IsSymlink(path)
		if linkErr != nil {
			wc.result.Errors = append(wc.result.Errors, linkErr)
			return nil
		}
		if isLink {
			wc.result.FilesSkipped++
			if d.IsDir() {
				return fs.SkipDir
			}
			return nil
		}

		if d.IsDir() {
			return wc.handleDir(path, relPath, d)
		}
		return wc.handleFile(path, relPath, d)
	})

	if err != nil {
		return result, fmt.Errorf("indexer: walk error: %w", err)
	}

	// Pass 1: Create nodes
	if hasPersistence {
		nodeStart := time.Now()
		if err := idx.createNodes(ctx, c, wc.pendingFolders, wc.pendingFiles); err != nil {
			return result, fmt.Errorf("indexer: %w", err)
		}
		log.Printf("[DEBUG] indexer createNodes: %d folders, %d files (%s)", len(wc.pendingFolders), len(wc.pendingFiles), time.Since(nodeStart))

		// Pass 2: Create edges
		edgeStart := time.Now()
		if err := idx.createEdges(ctx, c, repoName, wc.pendingFolders, wc.pendingFiles); err != nil {
			return result, fmt.Errorf("indexer: %w", err)
		}
		log.Printf("[DEBUG] indexer createEdges: (%s)", time.Since(edgeStart))
	}

	// Update in-memory index cache for incremental re-indexing
	if _, ok := idx.indexed[repoName]; !ok {
		idx.indexed[repoName] = make(map[string]time.Time)
	}
	repoCache := idx.indexed[repoName]
	for _, f := range wc.pendingFolders {
		repoCache[f.Path] = f.ModTime
	}
	for _, f := range wc.pendingFiles {
		repoCache[f.Path] = f.ModTime
	}

	return result, nil
}

// upsertRepository creates or updates a Repository node via mergeRepositorys
// GraphQL mutation through the repo-scoped client from ForRepo.
func (idx *Indexer) upsertRepository(ctx context.Context, c *client.Client, repoName string, repoPath string) (string, error) {
	if err := ctx.Err(); err != nil {
		return "", fmt.Errorf("indexer: context error: %w", err)
	}

	now := time.Now().UTC().Format(time.RFC3339)

	vars := map[string]any{
		"input": []any{map[string]any{
			"match":    map[string]any{"name": repoName},
			"onCreate": map[string]any{"name": repoName, "lastIndexed": now, "path": repoPath},
			"onMatch":  map[string]any{"lastIndexed": now, "path": repoPath},
		}},
	}

	_, err := c.Execute(ctx, gqlMergeRepositorys, vars)
	if err != nil {
		return "", fmt.Errorf("indexer: upsertRepository: %w", err)
	}

	return repoName, nil
}

// queryExistingNodes queries all existing Folder/File nodes for a repository
// using relationship WHERE filters via the repo-scoped client and returns a
// map of path -> lastUpdated for incremental re-indexing.
func (idx *Indexer) queryExistingNodes(ctx context.Context, c *client.Client, repoName string) (map[string]time.Time, error) {

	existing := make(map[string]time.Time)
	repoWhere := map[string]any{"repository": map[string]any{"name": repoName}}

	// Query folders
	folderResult, err := c.Execute(ctx, gqlQueryFolders, map[string]any{"where": repoWhere})
	if err != nil {
		return nil, fmt.Errorf("indexer: queryExistingNodes folders: %w", err)
	}
	decodeExistingNodes(existing, folderResult, "folders")

	// Query files
	fileResult, err := c.Execute(ctx, gqlQueryFiles, map[string]any{"where": repoWhere})
	if err != nil {
		return nil, fmt.Errorf("indexer: queryExistingNodes files: %w", err)
	}
	decodeExistingNodes(existing, fileResult, "files")

	return existing, nil
}

// decodeExistingNodes extracts path -> lastUpdated entries from a Client().Execute()
// query result into the provided map.
func decodeExistingNodes(dst map[string]time.Time, result *client.Result, key string) {
	if result == nil {
		return
	}
	data := result.Data()
	items, ok := data[key]
	if !ok {
		return
	}
	slice, ok := items.([]any)
	if !ok {
		return
	}
	for _, item := range slice {
		m, ok := item.(map[string]any)
		if !ok {
			continue
		}
		path, _ := m["path"].(string)
		if path == "" {
			continue
		}
		if ts, ok := m["lastUpdated"].(string); ok {
			if t, err := time.Parse(time.RFC3339, ts); err == nil {
				dst[path] = t
			}
		}
	}
}

// createNodes batch-creates Folder and File nodes in FalkorDB via mergeFolders
// and mergeFiles GraphQL mutations through the repo-scoped client, batched at batchSize.
func (idx *Indexer) createNodes(ctx context.Context, c *client.Client, folders []pendingFolder, files []pendingFile) error {

	// Create folders in batches
	for i := 0; i < len(folders); i += batchSize {
		end := i + batchSize
		if end > len(folders) {
			end = len(folders)
		}
		batch := folders[i:end]

		input := make([]any, len(batch))
		for j, f := range batch {
			input[j] = map[string]any{
				"match":    map[string]any{"path": f.Path},
				"onCreate": map[string]any{"path": f.Path, "lastUpdated": f.ModTime.UTC().Format(time.RFC3339)},
				"onMatch":  map[string]any{"path": f.Path, "lastUpdated": f.ModTime.UTC().Format(time.RFC3339)},
			}
		}

		if _, err := c.Execute(ctx, gqlMergeFolders, map[string]any{"input": input}); err != nil {
			return fmt.Errorf("indexer: createNodes folders: %w", err)
		}
	}

	// Create files in batches
	for i := 0; i < len(files); i += batchSize {
		end := i + batchSize
		if end > len(files) {
			end = len(files)
		}
		batch := files[i:end]

		input := make([]any, len(batch))
		for j, f := range batch {
			fields := map[string]any{
				"path":        f.Path,
				"filename":    filepath.Base(f.Path),
				"language":    f.Language,
				"lineCount":   f.LineCount,
				"lastUpdated": f.ModTime.UTC().Format(time.RFC3339),
			}
			input[j] = map[string]any{
				"match":    map[string]any{"path": f.Path},
				"onCreate": fields,
				"onMatch":  fields,
			}
		}

		if _, err := c.Execute(ctx, gqlMergeFiles, map[string]any{"input": input}); err != nil {
			return fmt.Errorf("indexer: createNodes files: %w", err)
		}
	}

	return nil
}

// indexerEdgeBatchSize is the number of edges per UNWIND batch for structural edges.
const indexerEdgeBatchSize = 50

// createEdges creates CONTAINS and BELONGS_TO edges via batched UNWIND Cypher.
// Uses ExecuteRawBatch to reduce FalkorDB round-trips from N individual calls
// to ceil(N/batchSize) batched UNWIND calls.
func (idx *Indexer) createEdges(ctx context.Context, c *client.Client, repoName string, folders []pendingFolder, files []pendingFile) error {
	if len(folders) == 0 && len(files) == 0 {
		return nil
	}

	// UNWIND-compatible Cypher queries (reference item.field instead of $field)
	const repoContainsFolder = "MATCH (a:Repository {name: item.from_name}) MATCH (b:Folder {path: item.to_path}) MERGE (a)-[:CONTAINS]->(b)"
	const repoContainsFile = "MATCH (a:Repository {name: item.from_name}) MATCH (b:File {path: item.to_path}) MERGE (a)-[:CONTAINS]->(b)"
	const folderContainsFolder = "MATCH (a:Folder {path: item.from_path}) MATCH (b:Folder {path: item.to_path}) MERGE (a)-[:CONTAINS]->(b)"
	const folderContainsFile = "MATCH (a:Folder {path: item.from_path}) MATCH (b:File {path: item.to_path}) MERGE (a)-[:CONTAINS]->(b)"
	const folderBelongsToRepo = "MATCH (a:Folder {path: item.from_path}) MATCH (b:Repository {name: item.to_name}) MERGE (a)-[:BELONGS_TO]->(b)"
	const fileBelongsToRepo = "MATCH (a:File {path: item.from_path}) MATCH (b:Repository {name: item.to_name}) MERGE (a)-[:BELONGS_TO]->(b)"

	// Accumulate edge items by query type
	var repoFolderItems, repoFileItems []map[string]any
	var folderFolderItems, folderFileItems []map[string]any
	var folderBelongsItems, fileBelongsItems []map[string]any

	for _, f := range folders {
		if f.ParentPath == "" {
			repoFolderItems = append(repoFolderItems, map[string]any{"from_name": repoName, "to_path": f.Path})
		} else {
			folderFolderItems = append(folderFolderItems, map[string]any{"from_path": f.ParentPath, "to_path": f.Path})
		}
		folderBelongsItems = append(folderBelongsItems, map[string]any{"from_path": f.Path, "to_name": repoName})
	}

	for _, f := range files {
		if f.ParentPath == "" {
			repoFileItems = append(repoFileItems, map[string]any{"from_name": repoName, "to_path": f.Path})
		} else {
			folderFileItems = append(folderFileItems, map[string]any{"from_path": f.ParentPath, "to_path": f.Path})
		}
		fileBelongsItems = append(fileBelongsItems, map[string]any{"from_path": f.Path, "to_name": repoName})
	}

	log.Printf("[DEBUG] createEdges: repoFolders=%d folderFolders=%d repoFiles=%d folderFiles=%d folderBelongs=%d fileBelongs=%d",
		len(repoFolderItems), len(folderFolderItems), len(repoFileItems), len(folderFileItems), len(folderBelongsItems), len(fileBelongsItems))

	type edgeBatch struct {
		query string
		items []map[string]any
		name  string
	}
	batches := []edgeBatch{
		{repoContainsFolder, repoFolderItems, "repo->folder"},
		{folderContainsFolder, folderFolderItems, "folder->folder"},
		{repoContainsFile, repoFileItems, "repo->file"},
		{folderContainsFile, folderFileItems, "folder->file"},
		{folderBelongsToRepo, folderBelongsItems, "folder->repo"},
		{fileBelongsToRepo, fileBelongsItems, "file->repo"},
	}

	for _, b := range batches {
		if len(b.items) == 0 {
			continue
		}
		if err := c.ExecuteRawBatch(ctx, b.query, b.items, indexerEdgeBatchSize); err != nil {
			return fmt.Errorf("indexer: createEdges %s: %w", b.name, err)
		}
	}

	return nil
}
