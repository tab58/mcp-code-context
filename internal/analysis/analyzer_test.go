package analysis

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

// --- Task 15: Analyzer orchestrator (two-pass) tests ---

// registryWithMockGo creates a Registry with a mock Go extractor registered.
func registryWithMockGo() *Registry {
	r := NewRegistry()
	r.RegisterExtractor("go", &mockExtractor{})
	return r
}

// TestNewAnalyzer_ReturnsNonNil verifies that NewAnalyzer returns a non-nil Analyzer.
// Expected result: Non-nil *Analyzer.
func TestNewAnalyzer_ReturnsNonNil(t *testing.T) {
	r := NewRegistry()
	a := NewAnalyzer(r, nil)
	if a == nil {
		t.Error("NewAnalyzer returned nil, expected non-nil *Analyzer")
	}
}

// createTestGoFiles is a test helper that creates a temp directory with Go source files.
func createTestGoFiles(t *testing.T) (string, []string) {
	t.Helper()
	dir := t.TempDir()

	files := map[string]string{
		"main.go": `package main

import "fmt"

func main() {
	fmt.Println("hello")
	helper()
}
`,
		"helper.go": `package main

func helper() string {
	return "helped"
}
`,
	}

	var paths []string
	for name, content := range files {
		path := filepath.Join(dir, name)
		if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
			t.Fatalf("failed to write %s: %v", name, err)
		}
		paths = append(paths, path)
	}

	return dir, paths
}

// TestAnalyzer_AnalyzeReturnsNoError verifies that Analyze completes without
// error for valid Go source files.
// Expected result: nil error, non-nil result.
func TestAnalyzer_AnalyzeReturnsNoError(t *testing.T) {
	_, files := createTestGoFiles(t)
	r := registryWithMockGo()
	a := NewAnalyzer(r, nil)

	result, err := a.Analyze(context.Background(), "test-repo-id", "", files)
	if err != nil {
		t.Fatalf("Analyze returned error: %v", err)
	}
	if result == nil {
		t.Fatal("Analyze returned nil result")
	}
}

// TestAnalyzer_SkipsUnknownLanguages verifies that files with unregistered
// extensions are skipped without error.
// Expected result: nil error (unknown files silently skipped).
func TestAnalyzer_SkipsUnknownLanguages(t *testing.T) {
	dir := t.TempDir()
	unknownFile := filepath.Join(dir, "data.xyz")
	if err := os.WriteFile(unknownFile, []byte("unknown content"), 0o644); err != nil {
		t.Fatal(err)
	}

	r := NewRegistry()
	a := NewAnalyzer(r, nil)

	_, err := a.Analyze(context.Background(), "test-repo-id", "", []string{unknownFile})
	if err != nil {
		t.Errorf("Analyze returned error for unknown language: %v (should skip silently)", err)
	}
}

// TestAnalyzer_HandlesEmptyFileList verifies that Analyze handles an empty
// file list gracefully.
// Expected result: nil error.
func TestAnalyzer_HandlesEmptyFileList(t *testing.T) {
	r := NewRegistry()
	a := NewAnalyzer(r, nil)

	result, err := a.Analyze(context.Background(), "test-repo-id", "", []string{})
	if err != nil {
		t.Errorf("Analyze returned error for empty file list: %v", err)
	}
	if result == nil {
		t.Fatal("Analyze returned nil result for empty file list")
	}
	if result.Files != 0 {
		t.Errorf("Files = %d, want 0", result.Files)
	}
}

// TestAnalyzer_HandlesSyntaxErrors verifies that files with syntax errors
// are skipped (logged) and do not cause Analyze to return an error.
// Expected result: nil error (syntax errors logged, not fatal).
func TestAnalyzer_HandlesSyntaxErrors(t *testing.T) {
	dir := t.TempDir()
	badFile := filepath.Join(dir, "bad.go")
	if err := os.WriteFile(badFile, []byte("package main\n\nfunc {\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	r := registryWithMockGo()
	a := NewAnalyzer(r, nil)

	_, err := a.Analyze(context.Background(), "test-repo-id", "", []string{badFile})
	if err != nil {
		t.Errorf("Analyze returned error for syntax errors: %v (should log and continue)", err)
	}
}

// TestAnalyzer_TwoPassArchitecture verifies that the Analyzer runs two passes
// and returns an AnalyzeResult.
// Expected result: Analyze processes multiple files without error.
func TestAnalyzer_TwoPassArchitecture(t *testing.T) {
	_, files := createTestGoFiles(t)

	r := registryWithMockGo()
	a := NewAnalyzer(r, nil)

	result, err := a.Analyze(context.Background(), "test-repo-id", "", files)
	if err != nil {
		t.Fatalf("Analyze returned error: %v", err)
	}
	if result == nil {
		t.Fatal("Analyze returned nil result")
	}
	// Mock extractor returns nil symbols/refs, so counts are 0
	// but Files should reflect how many were analyzed
	if result.Files != len(files) {
		t.Errorf("Files = %d, want %d", result.Files, len(files))
	}
}

// TestAnalyzer_CancelledContext verifies that Analyze respects context cancellation.
// Expected result: Non-nil error when context is cancelled.
func TestAnalyzer_CancelledContext(t *testing.T) {
	_, files := createTestGoFiles(t)

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // cancel immediately

	r := registryWithMockGo()
	a := NewAnalyzer(r, nil)

	_, err := a.Analyze(ctx, "test-repo-id", "", files)
	if err == nil {
		t.Error("Analyze with cancelled context should return error, got nil")
	}
}

// TestAnalyzer_SkipsFilesWithNoExtractor verifies that files with a registered
// grammar but no extractor are skipped silently.
func TestAnalyzer_SkipsFilesWithNoExtractor(t *testing.T) {
	_, files := createTestGoFiles(t)

	// Registry with grammars but NO extractors
	r := NewRegistry()
	a := NewAnalyzer(r, nil)

	result, err := a.Analyze(context.Background(), "test-repo-id", "", files)
	if err != nil {
		t.Fatalf("Analyze should skip files with no extractor, got error: %v", err)
	}
	if result.Files != 0 {
		t.Errorf("Files = %d, want 0 (no extractor registered)", result.Files)
	}
}

// TestResolvePass_ResolvesKnownSymbols verifies that resolvePass correctly
// matches references to symbols in the symbol table.
func TestResolvePass_ResolvesKnownSymbols(t *testing.T) {
	analyses := []FileAnalysis{
		{
			FilePath: "main.go",
			Language: "go",
			Symbols: []Symbol{
				{Name: "main", Kind: "function"},
				{Name: "helper", Kind: "function"},
			},
			References: []Reference{
				{FromSymbol: "main", ToName: "helper", Kind: "calls"},
				{FromSymbol: "main", ToName: "fmt.Println", Kind: "calls"},
			},
		},
		{
			FilePath: "helper.go",
			Language: "go",
			Symbols: []Symbol{
				{Name: "helper", Kind: "function"},
			},
		},
	}

	result := resolvePass(analyses)

	if result.Symbols != 3 {
		t.Errorf("Symbols = %d, want 3", result.Symbols)
	}
	if result.References != 2 {
		t.Errorf("References = %d, want 2", result.References)
	}
	if result.ResolvedReferences != 1 {
		t.Errorf("ResolvedReferences = %d, want 1 (helper is in symbol table)", result.ResolvedReferences)
	}
	if len(result.UnresolvedNames) != 1 {
		t.Errorf("UnresolvedNames = %d, want 1", len(result.UnresolvedNames))
	}
}

// TestResolvePass_EmptyAnalyses verifies that resolvePass handles empty input.
func TestResolvePass_EmptyAnalyses(t *testing.T) {
	result := resolvePass(nil)

	if result.Symbols != 0 {
		t.Errorf("Symbols = %d, want 0", result.Symbols)
	}
	if result.References != 0 {
		t.Errorf("References = %d, want 0", result.References)
	}
}
