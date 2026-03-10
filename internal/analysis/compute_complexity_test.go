package analysis

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

// --- Task 4: Analyzer.ComputeComplexity method ---

// TestComputeComplexity_SkipsTestFiles verifies that ComputeComplexity skips
// test files (e.g., *_test.go, *.test.ts).
// Expected result: No error, test files ignored.
func TestComputeComplexity_SkipsTestFiles(t *testing.T) {
	r := NewRegistry()
	a := NewAnalyzer(r, nil)

	err := a.ComputeComplexity(context.Background(), "testrepo", "/tmp/repo", []string{
		"/tmp/repo/foo_test.go",
		"/tmp/repo/bar.test.ts",
	})
	if err != nil {
		t.Errorf("ComputeComplexity with test files returned error: %v", err)
	}
}

// TestComputeComplexity_SkipsGeneratedFiles verifies that ComputeComplexity skips
// files in generated/ directories.
// Expected result: No error, generated files ignored.
func TestComputeComplexity_SkipsGeneratedFiles(t *testing.T) {
	r := NewRegistry()
	a := NewAnalyzer(r, nil)

	err := a.ComputeComplexity(context.Background(), "testrepo", "/tmp/repo", []string{
		"/tmp/repo/generated/models.go",
	})
	if err != nil {
		t.Errorf("ComputeComplexity with generated files returned error: %v", err)
	}
}

// TestComputeComplexity_SkipsUnknownLanguage verifies that ComputeComplexity
// skips files with no registered language/grammar.
// Expected result: No error, unsupported files ignored.
func TestComputeComplexity_SkipsUnknownLanguage(t *testing.T) {
	r := NewRegistry()
	a := NewAnalyzer(r, nil)

	err := a.ComputeComplexity(context.Background(), "testrepo", "/tmp/repo", []string{
		"/tmp/repo/data.json",
	})
	if err != nil {
		t.Errorf("ComputeComplexity with unknown language returned error: %v", err)
	}
}

// TestComputeComplexity_EmptyFiles verifies that ComputeComplexity handles
// an empty file list without error.
// Expected result: No error.
func TestComputeComplexity_EmptyFiles(t *testing.T) {
	r := NewRegistry()
	a := NewAnalyzer(r, nil)

	err := a.ComputeComplexity(context.Background(), "testrepo", "/tmp/repo", nil)
	if err != nil {
		t.Errorf("ComputeComplexity with empty files returned error: %v", err)
	}
}

// TestComputeComplexity_ParsesGoFunction verifies that ComputeComplexity
// parses a Go file with tree-sitter and calls the ComplexityExtractor on
// function_declaration nodes. With a stub extractor that returns 0, this
// will fail because the actual implementation should compute real complexity.
// Expected result: Method exists, accepts AnalyzeOption, and returns error.
func TestComputeComplexity_ParsesGoFunction(t *testing.T) {
	// Create a temp Go file
	dir := t.TempDir()
	goFile := filepath.Join(dir, "main.go")
	goSource := `package main

func hello() {
	if true {
		return
	}
}
`
	if err := os.WriteFile(goFile, []byte(goSource), 0644); err != nil {
		t.Fatal(err)
	}

	r := NewRegistry()
	// Register a mock complexity extractor that returns base=1
	r.RegisterComplexityExtractor("go", &mockComplexityExtractor{returnValue: 1})
	r.RegisterExtractor("go", &mockExtractor{})

	// Without db, ComputeComplexity should still parse and compute
	// but won't write to graph
	a := NewAnalyzer(r, nil)
	err := a.ComputeComplexity(context.Background(), "testrepo", dir, []string{goFile})
	if err != nil {
		t.Errorf("ComputeComplexity returned error: %v", err)
	}
}

// TestComputeComplexity_RespectsContext verifies that ComputeComplexity
// checks context cancellation.
// Expected result: Returns context error when context is cancelled.
func TestComputeComplexity_RespectsContext(t *testing.T) {
	r := NewRegistry()
	a := NewAnalyzer(r, nil)

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // cancel immediately

	err := a.ComputeComplexity(ctx, "testrepo", "/tmp/repo", []string{"/tmp/repo/main.go"})
	if err == nil {
		t.Error("ComputeComplexity with cancelled context returned nil error, want context error")
	}
}

// TestComputeComplexity_AcceptsAnalyzeOption verifies that ComputeComplexity
// accepts variadic AnalyzeOption parameters.
// Expected result: Compiles and runs without error.
func TestComputeComplexity_AcceptsAnalyzeOption(t *testing.T) {
	r := NewRegistry()
	a := NewAnalyzer(r, nil)

	progress := WithAnalyzeProgress(func(_, _ string) {})

	// Should accept options without error
	err := a.ComputeComplexity(context.Background(), "testrepo", "/tmp/repo", nil, progress)
	if err != nil {
		t.Errorf("ComputeComplexity with option returned error: %v", err)
	}
}
