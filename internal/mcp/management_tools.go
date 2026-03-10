package mcp

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/tab58/code-context/internal/analysis"
	"github.com/tab58/code-context/internal/indexer"
	"github.com/tab58/go-ormql/pkg/client"
)

// countNodeType executes a count query and returns the number of items
// in the result list identified by resultKey. Returns 0 on error.
func countNodeType(ctx context.Context, c *client.Client, query string, vars map[string]any, resultKey string) int {
	result, err := c.Execute(ctx, query, vars)
	if err != nil {
		return 0
	}
	if items, ok := result.Data()[resultKey].([]any); ok {
		return len(items)
	}
	return 0
}

// GraphQL query constants for management and stats tool handlers.
const (
	gqlCountFiles = `query($where: FileWhere) {
  files(where: $where) { path }
}`

	gqlCountFunctions = `query($where: FunctionWhere) {
  functions(where: $where) { name }
}`

	gqlCountClasses = `query($where: ClassWhere) {
  classs(where: $where) { name }
}`

	gqlCountModules = `query($where: ModuleWhere) {
  modules(where: $where) { name }
}`

	gqlCountExternalRefs = `query($where: ExternalReferenceWhere) {
  externalReferences(where: $where) { name }
}`
)

// IngestResponse is the response for ingest_repository.
type IngestResponse struct {
	Repository     string `json:"repository"`
	FilesIndexed   int    `json:"filesIndexed"`
	FoldersIndexed int    `json:"foldersIndexed"`
	FilesSkipped   int    `json:"filesSkipped"`
	SymbolsFound   int    `json:"symbolsFound"`
}

// DeleteResponse is the response for delete_repository.
type DeleteResponse struct {
	Repository string `json:"repository"`
	Deleted    bool   `json:"deleted"`
}

// RepoStatsResponse is the response for get_repository_stats.
type RepoStatsResponse struct {
	Repository        string `json:"repository"`
	Files             int    `json:"files"`
	Functions         int    `json:"functions"`
	Classes           int    `json:"classes"`
	Modules           int    `json:"modules"`
	ExternalReferences int   `json:"externalReferences"`
}

// handleIngestRepository handles the ingest_repository MCP tool call.
// Runs the full pipeline (index -> analyze) on a local directory path.
func (s *Server) handleIngestRepository(ctx context.Context, repoPath string) (*IngestResponse, error) {
	if repoPath == "" {
		return nil, errors.New("repository_path is required")
	}

	fi, err := os.Stat(repoPath)
	if err != nil {
		return nil, fmt.Errorf("path does not exist: %w", err)
	}
	if !fi.IsDir() {
		return nil, fmt.Errorf("path is not a directory: %s", repoPath)
	}

	if s.idx == nil {
		return nil, errors.New("indexer not configured")
	}

	repoName := filepath.Base(repoPath)

	// Step 1: Index
	result, err := s.idx.IndexRepository(ctx, repoPath, indexer.WithProgress(func(_, _ string) {}))
	if err != nil {
		return nil, fmt.Errorf("indexing failed: %w", err)
	}

	resp := IngestResponse{
		Repository:     repoName,
		FilesIndexed:   result.FilesIndexed,
		FoldersIndexed: result.FoldersIndexed,
		FilesSkipped:   result.FilesSkipped,
	}

	// Step 2: Analyze (if analyzer is configured and files were indexed)
	if s.analyzer != nil && len(result.FilePaths) > 0 {
		analyzeResult, analyzeErr := s.analyzer.Analyze(ctx, result.RepoID, repoPath, result.FilePaths, analysis.WithAnalyzeProgress(func(_, _ string) {}))
		if analyzeErr != nil {
			return nil, fmt.Errorf("analysis failed: %w", analyzeErr)
		}
		if analyzeResult != nil {
			resp.SymbolsFound = analyzeResult.Symbols
		}
	}

	// Step 3: Compute complexity
	if s.analyzer != nil && len(result.FilePaths) > 0 {
		if err := s.analyzer.ComputeComplexity(ctx, result.RepoID, repoPath, result.FilePaths); err != nil {
			return nil, fmt.Errorf("complexity computation failed: %w", err)
		}
	}

	return &resp, nil
}

// handleDeleteRepository handles the delete_repository MCP tool call.
// Removes all graph data for a repository.
func (s *Server) handleDeleteRepository(ctx context.Context, repo string) (*DeleteResponse, error) {
	if repo == "" {
		return nil, errors.New("repository is required")
	}

	if err := s.db.DeleteRepo(ctx, repo); err != nil {
		return nil, fmt.Errorf("delete_repository failed: %w", err)
	}

	return &DeleteResponse{
		Repository: repo,
		Deleted:    true,
	}, nil
}

// handleGetRepositoryStats handles the get_repository_stats MCP tool call.
// Returns node counts per type for a repository.
func (s *Server) handleGetRepositoryStats(ctx context.Context, repo string) (*RepoStatsResponse, error) {
	if repo == "" {
		return nil, errors.New("repository is required")
	}

	c, err := s.requireRepoClient(ctx, repo)
	if err != nil {
		return nil, err
	}

	vars := repoWhere(repo)

	stats := RepoStatsResponse{
		Repository:         repo,
		Files:              countNodeType(ctx, c, gqlCountFiles, vars, "files"),
		Functions:          countNodeType(ctx, c, gqlCountFunctions, vars, "functions"),
		Classes:            countNodeType(ctx, c, gqlCountClasses, vars, "classs"),
		Modules:            countNodeType(ctx, c, gqlCountModules, vars, "modules"),
		ExternalReferences: countNodeType(ctx, c, gqlCountExternalRefs, vars, "externalReferences"),
	}

	return &stats, nil
}
