package repl

import (
	"os"
	"strings"
	"testing"
)

// === Task 9: Update REPL handleIngest ===
// handleIngest should pass the ingest path as repoPath to Analyzer.Analyze().

// TestREPL_HandleIngestPassesRepoPath verifies that the REPL commands.go
// source code passes path to Analyzer.Analyze as the repoPath parameter.
// Expected result: commands.go contains a call to Analyze with path as repoPath.
func TestREPL_HandleIngestPassesRepoPath(t *testing.T) {
	data, err := os.ReadFile("commands.go")
	if err != nil {
		t.Fatalf("failed to read commands.go: %v", err)
	}
	content := string(data)

	// The Analyze call should include the path variable between repoID and filePaths.
	// Old: Analyze(ctx, result.RepoID, result.FilePaths, ...)
	// New: Analyze(ctx, result.RepoID, path, result.FilePaths, ...)
	if !strings.Contains(content, "Analyze(ctx, result.RepoID, path,") {
		t.Error("commands.go should pass 'path' as repoPath to Analyzer.Analyze() — expected 'Analyze(ctx, result.RepoID, path, result.FilePaths, ...)'")
	}
}
