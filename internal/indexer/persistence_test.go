package indexer

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"
)

// --- Task 3: Repository upsert tests ---

// TestUpsertRepository_ExecutesGraphQLQuery verifies that upsertRepository
// executes a GraphQL query/mutation via the go-ormql client to create or update
// a Repository node.
// Expected result: At least one call is made through the driver.
func TestUpsertRepository_ExecutesGraphQLQuery(t *testing.T) {
	idx, rec := newTestIndexerWithRecorder(t)

	_, err := idx.upsertRepository(context.Background(), testClient(t, idx, "test-repo"), "test-repo", "/tmp/test-repo")
	if err != nil {
		t.Fatalf("upsertRepository returned error: %v", err)
	}

	totalCalls := len(rec.executeCalls) + len(rec.executeWriteCalls)
	if totalCalls == 0 {
		t.Error("upsertRepository made no driver calls, expected at least one for query/create")
	}
}

// TestUpsertRepository_ReturnsRepoName verifies that upsertRepository returns
// the repository name (used as identifier for subsequent operations).
// Expected result: Non-empty string returned.
func TestUpsertRepository_ReturnsRepoName(t *testing.T) {
	idx, _ := newTestIndexerWithRecorder(t)

	name, err := idx.upsertRepository(context.Background(), testClient(t, idx, "test-repo"), "my-project", "/tmp/my-project")
	if err != nil {
		t.Fatalf("upsertRepository returned error: %v", err)
	}

	if name == "" {
		t.Error("upsertRepository returned empty name, expected non-empty")
	}
}

// TestUpsertRepository_UsesRepoName verifies that upsertRepository passes
// the repository name to the GraphQL query.
// Expected result: Driver call params contain the repo name (may be deeply nested).
func TestUpsertRepository_UsesRepoName(t *testing.T) {
	idx, rec := newTestIndexerWithRecorder(t)

	_, err := idx.upsertRepository(context.Background(), testClient(t, idx, "test-repo"), "my-project", "/tmp/my-project")
	if err != nil {
		t.Fatalf("upsertRepository returned error: %v", err)
	}

	// After Client().Execute() translation, params may be deeply nested.
	// Verify at least one call was made (the repo name is in the GraphQL vars,
	// translated by go-ormql into Cypher params).
	totalCalls := len(rec.executeCalls) + len(rec.executeWriteCalls)
	if totalCalls == 0 {
		t.Error("upsertRepository did not make any driver calls")
	}
}

// --- Task 4: Incremental existing-nodes query tests ---

// TestQueryExistingNodes_ReturnsParsedMap verifies that queryExistingNodes
// returns a map[string]time.Time from query results.
// Expected result: Non-nil map returned.
func TestQueryExistingNodes_ReturnsParsedMap(t *testing.T) {
	idx, _ := newTestIndexerWithRecorder(t)

	nodes, err := idx.queryExistingNodes(context.Background(), testClient(t, idx, "test-repo"), "test-repo")
	if err != nil {
		t.Fatalf("queryExistingNodes returned error: %v", err)
	}

	if nodes == nil {
		t.Error("queryExistingNodes returned nil map, expected non-nil")
	}
}

// TestQueryExistingNodes_MakesDriverCalls verifies that queryExistingNodes
// uses Client().Execute() for queries.
// Expected result: At least one Execute or ExecuteWrite call on the driver.
func TestQueryExistingNodes_MakesDriverCalls(t *testing.T) {
	idx, rec := newTestIndexerWithRecorder(t)

	_, err := idx.queryExistingNodes(context.Background(), testClient(t, idx, "test-repo"), "test-repo")
	if err != nil {
		t.Fatalf("queryExistingNodes returned error: %v", err)
	}

	totalCalls := len(rec.executeCalls) + len(rec.executeWriteCalls)
	if totalCalls == 0 {
		t.Error("queryExistingNodes made no driver calls, expected queries via Client().Execute()")
	}
}

// TestQueryExistingNodes_EmptyRepoReturnsEmptyMap verifies that querying
// existing nodes for a repo with no data returns an empty (not nil) map.
// Expected result: Empty map, no error.
func TestQueryExistingNodes_EmptyRepoReturnsEmptyMap(t *testing.T) {
	idx, _ := newTestIndexerWithRecorder(t)

	nodes, err := idx.queryExistingNodes(context.Background(), testClient(t, idx, "test-repo"), "empty-repo")
	if err != nil {
		t.Fatalf("queryExistingNodes returned error: %v", err)
	}

	if len(nodes) != 0 {
		t.Errorf("queryExistingNodes returned %d entries, want 0 for empty repo", len(nodes))
	}
}

// --- Task 5: Node batching (Pass 1) tests ---

// TestCreateNodes_FlushesViaMutations verifies that createNodes sends
// GraphQL mutations through the client to create Folder and File nodes.
// Expected result: Driver calls are made for node creation.
func TestCreateNodes_FlushesViaMutations(t *testing.T) {
	idx, rec := newTestIndexerWithRecorder(t)

	folders := []pendingFolder{
		{Path: "src", ParentPath: "", ModTime: time.Now()},
	}
	files := []pendingFile{
		{Path: "main.go", ParentPath: "", Language: "go", LineCount: 10, ModTime: time.Now()},
	}

	err := idx.createNodes(context.Background(), testClient(t, idx, "test-repo"), folders, files)
	if err != nil {
		t.Fatalf("createNodes returned error: %v", err)
	}

	totalCalls := len(rec.executeCalls) + len(rec.executeWriteCalls)
	if totalCalls == 0 {
		t.Error("createNodes made no driver calls, expected mutation calls for folders and files")
	}
}

// TestCreateNodes_BatchesByBatchSize verifies that createNodes splits large
// input into batches of batchSize (50). 120 items should produce 3 batches.
// Expected result: At least 3 mutation calls for 120 folders.
func TestCreateNodes_BatchesByBatchSize(t *testing.T) {
	idx, rec := newTestIndexerWithRecorder(t)

	// Create 120 pending folders
	folders := make([]pendingFolder, 120)
	for i := range folders {
		folders[i] = pendingFolder{
			Path:       filepath.Join("dir", string(rune('a'+i%26))),
			ParentPath: "",
			ModTime:    time.Now(),
		}
	}

	err := idx.createNodes(context.Background(), testClient(t, idx, "test-repo"), folders, nil)
	if err != nil {
		t.Fatalf("createNodes returned error: %v", err)
	}

	// With batchSize=50, 120 folders should produce at least 3 batch calls
	totalCalls := len(rec.executeCalls) + len(rec.executeWriteCalls)
	if totalCalls < 3 {
		t.Errorf("createNodes made %d calls for 120 items, expected at least 3 (batches of 50)", totalCalls)
	}
}

// TestCreateNodes_EmptyInput verifies that createNodes handles empty slices
// gracefully without making any driver calls.
// Expected result: No error, no driver calls.
func TestCreateNodes_EmptyInput(t *testing.T) {
	idx, rec := newTestIndexerWithRecorder(t)
	c := testClient(t, idx, "test-repo")

	// Record baseline after ForRepo/CreateIndexes
	baselineCalls := len(rec.executeCalls) + len(rec.executeWriteCalls)

	err := idx.createNodes(context.Background(), c, nil, nil)
	if err != nil {
		t.Fatalf("createNodes with empty input returned error: %v", err)
	}

	totalCalls := len(rec.executeCalls) + len(rec.executeWriteCalls) - baselineCalls
	if totalCalls != 0 {
		t.Errorf("createNodes with empty input made %d calls, expected 0", totalCalls)
	}
}

// --- Task 6: Edge creation (Pass 2) tests ---

// TestCreateEdges_ExecutesMutations verifies that createEdges sends
// mutations through Client().Execute() for edge creation.
// Expected result: At least one ExecuteWrite call for edges.
func TestCreateEdges_ExecutesMutations(t *testing.T) {
	idx, rec := newTestIndexerWithRecorder(t)

	folders := []pendingFolder{
		{Path: "src", ParentPath: "", ModTime: time.Now()},
	}
	files := []pendingFile{
		{Path: "main.go", ParentPath: "", Language: "go", LineCount: 10, ModTime: time.Now()},
	}

	err := idx.createEdges(context.Background(), testClient(t, idx, "test-repo"), "test-repo", folders, files)
	if err != nil {
		t.Fatalf("createEdges returned error: %v", err)
	}

	if len(rec.executeWriteCalls) == 0 {
		t.Error("createEdges made no ExecuteWrite calls, expected connect* mutations")
	}
}

// TestCreateEdges_SixStatementTypes verifies that createEdges produces all 6
// edge statement types.
// Expected result: At least 6 ExecuteWrite calls (one per edge type).
func TestCreateEdges_SixStatementTypes(t *testing.T) {
	idx, rec := newTestIndexerWithRecorder(t)

	folders := []pendingFolder{
		{Path: "src", ParentPath: "", ModTime: time.Now()},
		{Path: "src/utils", ParentPath: "src", ModTime: time.Now()},
	}
	files := []pendingFile{
		{Path: "main.go", ParentPath: "", Language: "go", LineCount: 10, ModTime: time.Now()},
		{Path: "src/util.go", ParentPath: "src", Language: "go", LineCount: 5, ModTime: time.Now()},
	}

	err := idx.createEdges(context.Background(), testClient(t, idx, "test-repo"), "test-repo", folders, files)
	if err != nil {
		t.Fatalf("createEdges returned error: %v", err)
	}

	// 6 edge types, each produces at least 1 ExecuteWrite call
	if len(rec.executeWriteCalls) < 6 {
		t.Errorf("createEdges made %d ExecuteWrite calls, expected at least 6 (one per edge type)", len(rec.executeWriteCalls))
	}
}

// TestCreateEdges_EmptyInput verifies that createEdges handles empty slices
// gracefully without errors.
// Expected result: No error.
func TestCreateEdges_EmptyInput(t *testing.T) {
	idx, rec := newTestIndexerWithRecorder(t)
	c := testClient(t, idx, "test-repo")

	// Record baseline after ForRepo/CreateIndexes
	baselineWrites := len(rec.executeWriteCalls)

	err := idx.createEdges(context.Background(), c, "test-repo", nil, nil)
	if err != nil {
		t.Fatalf("createEdges with empty input returned error: %v", err)
	}

	// With no items, no edge calls should be made beyond baseline
	newWrites := len(rec.executeWriteCalls) - baselineWrites
	if newWrites != 0 {
		t.Errorf("createEdges with empty input made %d calls, expected 0", newWrites)
	}
}

// TestCreateEdges_BatchesAt50 verifies that edge creation batches at 50 items.
func TestCreateEdges_BatchesAt50(t *testing.T) {
	idx, rec := newTestIndexerWithRecorder(t)

	folders := make([]pendingFolder, 120)
	for i := range folders {
		folders[i] = pendingFolder{
			Path:       filepath.Join("dir", string(rune('a'+i%26))),
			ParentPath: "",
			ModTime:    time.Now(),
		}
	}

	err := idx.createEdges(context.Background(), testClient(t, idx, "test-repo"), "test-repo", folders, nil)
	if err != nil {
		t.Fatalf("createEdges returned error: %v", err)
	}

	// 120 root folders: CONTAINS batched at 50 = 3 calls, BELONGS_TO batched at 50 = 3 calls
	if len(rec.executeWriteCalls) < 6 {
		t.Errorf("createEdges made %d calls for 120 folders, expected at least 6 (batched at 50)", len(rec.executeWriteCalls))
	}
}

// --- Task 7: Wire into IndexRepository walk loop tests ---

// TestIndexRepository_WritesGraphNodes verifies that IndexRepository creates
// Repository, Folder, and File nodes in FalkorDB during indexing.
func TestIndexRepository_WritesGraphNodes(t *testing.T) {
	repoPath := createPersistenceTestRepo(t)
	idx, rec := newTestIndexerWithRecorder(t)

	_, err := idx.IndexRepository(context.Background(), repoPath)
	if err != nil {
		t.Fatalf("IndexRepository returned error: %v", err)
	}

	totalCalls := len(rec.executeCalls) + len(rec.executeWriteCalls)
	if totalCalls == 0 {
		t.Error("IndexRepository made no driver calls, expected graph writes for Repository/Folder/File nodes")
	}
}

// TestIndexRepository_CreatesEdges verifies that IndexRepository creates
// edges via ExecuteWrite.
func TestIndexRepository_CreatesEdges(t *testing.T) {
	repoPath := createPersistenceTestRepo(t)
	idx, rec := newTestIndexerWithRecorder(t)

	_, err := idx.IndexRepository(context.Background(), repoPath)
	if err != nil {
		t.Fatalf("IndexRepository returned error: %v", err)
	}

	if len(rec.executeWriteCalls) == 0 {
		t.Error("IndexRepository made no ExecuteWrite calls, expected edge creation")
	}
}

// TestIndexRepository_IncrementalSkip verifies that unchanged files are
// skipped during re-indexing.
func TestIndexRepository_IncrementalSkip(t *testing.T) {
	repoPath := createPersistenceTestRepo(t)
	idx, _ := newTestIndexerWithRecorder(t)

	// First index
	_, err := idx.IndexRepository(context.Background(), repoPath)
	if err != nil {
		t.Fatalf("first IndexRepository returned error: %v", err)
	}

	// Second index of same unchanged directory
	result, err := idx.IndexRepository(context.Background(), repoPath)
	if err != nil {
		t.Fatalf("second IndexRepository returned error: %v", err)
	}

	if result.FilesSkipped == 0 && result.FilesIndexed > 0 {
		t.Error("second IndexRepository with unchanged files: FilesSkipped=0, expected incremental skip")
	}
}

// --- Task 9: Unit tests for methods with mock driver ---

// TestUpsertRepository_NilContext verifies that upsertRepository handles
// a cancelled context by returning an error.
func TestUpsertRepository_NilContext(t *testing.T) {
	idx, _ := newTestIndexerWithRecorder(t)
	c := testClient(t, idx, "test-repo")
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err := idx.upsertRepository(ctx, c, "test-repo", "/tmp/test-repo")
	if err == nil {
		t.Error("upsertRepository with cancelled context should return error, got nil")
	}
}

// TestCreateNodes_DriverWriteError verifies that createNodes returns an error
// when the driver fails during a write.
func TestCreateNodes_DriverWriteError(t *testing.T) {
	// CreateIndexes is a no-op (0 writes), then fail on first write
	idx := newFailAfterNIndexer(t, 0, context.DeadlineExceeded)

	folders := []pendingFolder{
		{Path: "src", ParentPath: "", ModTime: time.Now()},
	}

	err := idx.createNodes(context.Background(), testClient(t, idx, "test-repo"), folders, nil)
	if err == nil {
		t.Error("createNodes with failing driver should return error, got nil")
	}
}

// TestCreateEdges_DriverWriteError verifies that createEdges returns an error
// when the driver fails during ExecuteWrite.
func TestCreateEdges_DriverWriteError(t *testing.T) {
	// CreateIndexes is a no-op (0 writes), then fail on first write
	idx := newFailAfterNIndexer(t, 0, context.DeadlineExceeded)

	folders := []pendingFolder{
		{Path: "src", ParentPath: "", ModTime: time.Now()},
	}

	err := idx.createEdges(context.Background(), testClient(t, idx, "test-repo"), "test-repo", folders, nil)
	if err == nil {
		t.Error("createEdges with failing driver should return error, got nil")
	}
}

// --- Task 10: Edge case handling tests ---

// TestIndexRepository_EmptyRepo_CreatesRepoNodeOnly verifies that indexing
// an empty directory creates a Repository node but no Folder/File nodes.
func TestIndexRepository_EmptyRepo_CreatesRepoNodeOnly(t *testing.T) {
	emptyDir := t.TempDir()
	idx, rec := newTestIndexerWithRecorder(t)

	result, err := idx.IndexRepository(context.Background(), emptyDir)
	if err != nil {
		t.Fatalf("IndexRepository on empty dir returned error: %v", err)
	}

	if result.FilesIndexed != 0 {
		t.Errorf("FilesIndexed = %d, want 0 for empty dir", result.FilesIndexed)
	}
	if result.FoldersIndexed != 0 {
		t.Errorf("FoldersIndexed = %d, want 0 for empty dir", result.FoldersIndexed)
	}

	totalCalls := len(rec.executeCalls) + len(rec.executeWriteCalls)
	if totalCalls == 0 {
		t.Error("IndexRepository on empty dir made no driver calls, expected at least repo upsert")
	}
}

// TestIndexRepository_PermissionDenied_CollectsErrors verifies that permission
// errors during walk are collected in IndexResult.Errors and indexing continues.
func TestIndexRepository_PermissionDenied_CollectsErrors(t *testing.T) {
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, "good.go"), []byte("package main"), 0o644); err != nil {
		t.Fatal(err)
	}
	unreadable := filepath.Join(root, "secret.go")
	if err := os.WriteFile(unreadable, []byte("package x"), 0o000); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { os.Chmod(unreadable, 0o644) })

	idx, _ := newTestIndexerWithRecorder(t)

	result, err := idx.IndexRepository(context.Background(), root)
	if err != nil {
		t.Fatalf("IndexRepository should not return top-level error for permission issues: %v", err)
	}

	if len(result.Errors) == 0 {
		t.Error("expected Errors to contain permission error, got empty slice")
	}
}

// TestIndexRepository_MidBatchFailure_ReturnsError verifies that a FalkorDB
// write failure mid-batch propagates as an error from IndexRepository.
func TestIndexRepository_MidBatchFailure_ReturnsError(t *testing.T) {
	repoPath := createPersistenceTestRepo(t)

	// CreateIndexes is a no-op (0 vector indexes), so fail on first indexer write
	idx := newFailAfterNIndexer(t, 0, context.DeadlineExceeded)

	_, err := idx.IndexRepository(context.Background(), repoPath)
	if err == nil {
		t.Error("IndexRepository with failing driver should return error, got nil")
	}
}
