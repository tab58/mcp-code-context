package mcp

import (
	"context"
	"encoding/json"
	"strings"
	"testing"

	"github.com/tab58/go-ormql/pkg/driver"
)

// ============================================================================
// Feature 1: ingest_repository MCP Tool
// ============================================================================

// func TestHandleIngestRepository_MissingPath(t *testing.T) {
// 	svc := newTestService(t)
// 	_, err := svc.HandleIngestRepository(context.Background(), "")
// 	if err == nil {
// 		t.Fatal("expected error for empty path")
// 	}
// 	if !strings.Contains(err.Error(), "repository_path is required") {
// 		t.Errorf("unexpected error: %v", err)
// 	}
// }

// func TestHandleIngestRepository_PathDoesNotExist(t *testing.T) {
// 	svc := newTestService(t)
// 	_, err := svc.HandleIngestRepository(context.Background(), "/nonexistent/path/xyz")
// 	if err == nil {
// 		t.Fatal("expected error for nonexistent path")
// 	}
// 	if !strings.Contains(err.Error(), "does not exist") {
// 		t.Errorf("unexpected error: %v", err)
// 	}
// }

// func TestHandleIngestRepository_PathIsFile(t *testing.T) {
// 	tmpFile := filepath.Join(t.TempDir(), "file.txt")
// 	if err := os.WriteFile(tmpFile, []byte("hello"), 0644); err != nil {
// 		t.Fatal(err)
// 	}

// 	svc := newTestService(t)
// 	_, err := svc.HandleIngestRepository(context.Background(), tmpFile)
// 	if err == nil {
// 		t.Fatal("expected error for file path")
// 	}
// 	if !strings.Contains(err.Error(), "not a directory") {
// 		t.Errorf("unexpected error: %v", err)
// 	}
// }

// func TestHandleIngestRepository_NoIndexer(t *testing.T) {
// 	svc := newTestService(t)
// 	svc.idx = nil
// 	_, err := svc.HandleIngestRepository(context.Background(), t.TempDir())
// 	if err == nil {
// 		t.Fatal("expected error when indexer is nil")
// 	}
// 	if !strings.Contains(err.Error(), "indexer not configured") {
// 		t.Errorf("unexpected error: %v", err)
// 	}
// }

// func TestHandleIngestRepository_Success(t *testing.T) {
// 	// Create a temp dir with a file
// 	dir := t.TempDir()
// 	if err := os.WriteFile(filepath.Join(dir, "main.go"), []byte("package main\n"), 0644); err != nil {
// 		t.Fatal(err)
// 	}

// 	svc := newTestServiceWithIndexer(t)
// 	resp, err := svc.HandleIngestRepository(context.Background(), dir)
// 	if err != nil {
// 		t.Fatalf("unexpected error: %v", err)
// 	}
// 	if resp.Repository != filepath.Base(dir) {
// 		t.Errorf("expected repository=%q, got %q", filepath.Base(dir), resp.Repository)
// 	}
// 	if resp.FilesIndexed == 0 {
// 		t.Error("expected at least 1 file indexed")
// 	}
// }

func TestIngestResponse_JSON(t *testing.T) {
	resp := IngestResponse{
		Repository:     "myrepo",
		FilesIndexed:   10,
		FoldersIndexed: 3,
		FilesSkipped:   2,
		SymbolsFound:   15,
	}
	data, err := json.Marshal(resp)
	if err != nil {
		t.Fatalf("failed to marshal: %v", err)
	}
	s := string(data)
	for _, field := range []string{"repository", "filesIndexed", "foldersIndexed", "filesSkipped", "symbolsFound"} {
		if !strings.Contains(s, field) {
			t.Errorf("JSON missing field %q", field)
		}
	}
}

// ============================================================================
// Feature 2: delete_repository MCP Tool
// ============================================================================

func TestHandleDeleteRepository_MissingRepo(t *testing.T) {
	svc := newTestService(t)
	_, err := svc.HandleDeleteRepository(context.Background(), "")
	if err == nil {
		t.Fatal("expected error for empty repository")
	}
	if !strings.Contains(err.Error(), "repository is required") {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestHandleDeleteRepository_Success(t *testing.T) {
	svc, drv := newTestServiceWithResponses(t, nil)
	_ = drv
	resp, err := svc.HandleDeleteRepository(context.Background(), "myrepo")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Repository != "myrepo" {
		t.Errorf("expected repository=myrepo, got %q", resp.Repository)
	}
	if !resp.Deleted {
		t.Error("expected deleted=true")
	}
}

func TestHandleDeleteRepository_WriteCalls(t *testing.T) {
	svc, drv := newTestServiceWithResponses(t, nil)
	_, err := svc.HandleDeleteRepository(context.Background(), "myrepo")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// Write calls include createIndexes (11 range indexes) + 2 delete calls.
	// Verify the last 2 calls are the delete operations.
	n := len(drv.writeCalls)
	if n < 2 {
		t.Fatalf("expected at least 2 write calls, got %d", n)
	}
	// Second-to-last call should delete dependent nodes via BELONGS_TO
	if !strings.Contains(drv.writeCalls[n-2].Query, "BELONGS_TO") {
		t.Errorf("second-to-last write should delete BELONGS_TO dependent nodes, got: %s", drv.writeCalls[n-2].Query)
	}
	// Last call should delete the Repository node
	if !strings.Contains(drv.writeCalls[n-1].Query, "Repository") {
		t.Errorf("last write should delete Repository node, got: %s", drv.writeCalls[n-1].Query)
	}
}

func TestDeleteResponse_JSON(t *testing.T) {
	resp := DeleteResponse{Repository: "test", Deleted: true}
	data, err := json.Marshal(resp)
	if err != nil {
		t.Fatalf("failed to marshal: %v", err)
	}
	s := string(data)
	for _, field := range []string{"repository", "deleted"} {
		if !strings.Contains(s, field) {
			t.Errorf("JSON missing field %q", field)
		}
	}
}

// ============================================================================
// Feature 3: get_repository_stats MCP Tool
// ============================================================================

func TestHandleGetRepositoryStats_MissingRepo(t *testing.T) {
	svc := newTestService(t)
	_, err := svc.HandleGetRepositoryStats(context.Background(), "")
	if err == nil {
		t.Fatal("expected error for empty repository")
	}
}

func TestHandleGetRepositoryStats_EmptyRepo(t *testing.T) {
	svc := newTestService(t)
	resp, err := svc.HandleGetRepositoryStats(context.Background(), "empty-repo")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Repository != "empty-repo" {
		t.Errorf("expected repository=empty-repo, got %q", resp.Repository)
	}
	if resp.Files != 0 || resp.Functions != 0 || resp.Classes != 0 || resp.Modules != 0 || resp.ExternalReferences != 0 {
		t.Error("expected all counts to be 0 for empty repo")
	}
}

func TestHandleGetRepositoryStats_WithData(t *testing.T) {
	responses := []driver.Result{
		// files query
		makeResult(map[string]any{
			"files": []any{
				map[string]any{"path": "a.go"},
				map[string]any{"path": "b.go"},
				map[string]any{"path": "c.go"},
			},
		}),
		// functions query
		makeResult(map[string]any{
			"functions": []any{
				map[string]any{"name": "foo"},
				map[string]any{"name": "bar"},
			},
		}),
		// classes query
		makeResult(map[string]any{
			"classs": []any{
				map[string]any{"name": "MyClass"},
			},
		}),
		// modules query
		makeResult(map[string]any{
			"modules": []any{
				map[string]any{"name": "main"},
				map[string]any{"name": "utils"},
			},
		}),
		// external references query
		makeResult(map[string]any{
			"externalReferences": []any{
				map[string]any{"name": "fmt.Println"},
			},
		}),
	}

	svc, _ := newTestServiceWithResponses(t, responses)
	resp, err := svc.HandleGetRepositoryStats(context.Background(), "myrepo")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if resp.Files != 3 {
		t.Errorf("expected files=3, got %d", resp.Files)
	}
	if resp.Functions != 2 {
		t.Errorf("expected functions=2, got %d", resp.Functions)
	}
	if resp.Classes != 1 {
		t.Errorf("expected classes=1, got %d", resp.Classes)
	}
	if resp.Modules != 2 {
		t.Errorf("expected modules=2, got %d", resp.Modules)
	}
	if resp.ExternalReferences != 1 {
		t.Errorf("expected externalReferences=1, got %d", resp.ExternalReferences)
	}
}

func TestRepoStatsResponse_JSON(t *testing.T) {
	resp := RepoStatsResponse{
		Repository:         "test",
		Files:              10,
		Functions:          20,
		Classes:            5,
		Modules:            3,
		ExternalReferences: 8,
	}
	data, err := json.Marshal(resp)
	if err != nil {
		t.Fatalf("failed to marshal: %v", err)
	}
	s := string(data)
	for _, field := range []string{"repository", "files", "functions", "classes", "modules", "externalReferences"} {
		if !strings.Contains(s, field) {
			t.Errorf("JSON missing field %q", field)
		}
	}
}

// ============================================================================
// GraphQL constant checks
// ============================================================================

func TestManagementGQLConstants(t *testing.T) {
	tests := []struct {
		name  string
		query string
		want  string
	}{
		{"gqlCountFiles", gqlCountFiles, "files"},
		{"gqlCountFunctions", gqlCountFunctions, "functions"},
		{"gqlCountClasses", gqlCountClasses, "classs"},
		{"gqlCountModules", gqlCountModules, "modules"},
		{"gqlCountExternalRefs", gqlCountExternalRefs, "externalReferences"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if !strings.Contains(tt.query, tt.want) {
				t.Errorf("%s does not contain %q", tt.name, tt.want)
			}
		})
	}
}
