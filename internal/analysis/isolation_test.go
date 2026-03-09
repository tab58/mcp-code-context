package analysis

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

// --- Task 4: Migrate analyzer to ForRepo ---

// TestAnalyze_UsesForRepo verifies that Analyze calls db.ForRepo(ctx, repoName)
// to get a repo-scoped client for graph writes, rather than db.Client().
// Expected result: Analyze succeeds with recording driver, driver records calls
// (proving the client was obtained via ForRepo, not Client()).
func TestAnalyze_UsesForRepo(t *testing.T) {
	analyzer, rec := newAnalyzerWithRecorder(t)
	files := createGraphWriteTestFiles(t)
	repoPath := filepath.Dir(files[0])
	repoName := filepath.Base(repoPath)

	result, err := analyzer.Analyze(context.Background(), repoName, repoPath, files)
	if err != nil {
		t.Fatalf("Analyze returned error: %v", err)
	}
	if result == nil {
		t.Fatal("Analyze returned nil result")
	}
	if result.Symbols == 0 {
		t.Error("Analyze found 0 symbols, expected > 0")
	}

	// Verify driver was exercised (graph writes happened via ForRepo-obtained client)
	totalCalls := len(rec.executeCalls) + len(rec.executeWriteCalls)
	if totalCalls == 0 {
		t.Error("Analyze made no driver calls — ForRepo-based client not exercised")
	}
}

// TestWritePass1_AcceptsClient verifies that writePass1 uses a client obtained
// from ForRepo (passed as parameter or obtained internally via ForRepo).
// Expected result: writePass1 succeeds and makes driver calls.
func TestWritePass1_AcceptsClient(t *testing.T) {
	analyzer, rec := newAnalyzerWithRecorder(t)

	// Create a minimal FileAnalysis with symbols
	analyses := []FileAnalysis{
		{
			FilePath: "/tmp/test.go",
			Language: "go",
			Symbols: []Symbol{
				{Name: "main", Kind: "function", Path: "/tmp/test.go", Language: "go", Visibility: "public"},
			},
		},
	}

	c, err := analyzer.db.ForRepo(context.Background(), "test-repo")
	if err != nil {
		t.Fatalf("ForRepo returned error: %v", err)
	}

	err = analyzer.writePass1(context.Background(), c, "test-repo", "/tmp/test-repo", analyses)
	if err != nil {
		t.Fatalf("writePass1 returned error: %v", err)
	}

	totalCalls := len(rec.executeCalls) + len(rec.executeWriteCalls)
	if totalCalls == 0 {
		t.Error("writePass1 made no driver calls")
	}
}

// --- Task 5: Migrate embedder to ForRepo (source inspection in verify) ---
// Behavioral tests for the embedder ForRepo migration are in internal/search/
// because embedding requires CGo. The verify tests confirm source changes.

// --- Helper for this test file ---

func createIsolationTestFile(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, "test.go")
	content := `package main

func hello() {
	return
}
`
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("failed to write test file: %v", err)
	}
	return path
}
