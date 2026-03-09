package analysis

import (
	"context"
	"os"
	"path/filepath"
	"sync"
	"testing"
)

// === Task 4 + 10: AnalyzeOption with progress callback ===

// TestAnalyze_NoOption verifies backward compatibility — Analyze works without options.
// Expected result: no crash, returns result.
func TestAnalyze_NoOption(t *testing.T) {
	registry := NewRegistry()
	a := NewAnalyzer(registry, nil)

	result, err := a.Analyze(context.Background(), "test-repo", "", nil)
	if err != nil {
		t.Errorf("Analyze with no options failed: %v", err)
	}
	if result == nil {
		t.Error("Analyze returned nil result")
	}
}

// TestWithAnalyzeProgress_ReceivesCallbacks verifies that the progress
// callback is called per-file during Pass 1 with stage="analyzing".
// Expected result: progress function is called for each analyzed file.
func TestWithAnalyzeProgress_ReceivesCallbacks(t *testing.T) {
	// Create a temp Go file for analysis
	root := t.TempDir()
	goFile := filepath.Join(root, "main.go")
	if err := os.WriteFile(goFile, []byte("package main\n\nfunc main() {}\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	registry := NewRegistry()
	a := NewAnalyzer(registry, nil)

	var mu sync.Mutex
	var calls []struct{ stage, message string }
	progress := func(stage, message string) {
		mu.Lock()
		defer mu.Unlock()
		calls = append(calls, struct{ stage, message string }{stage, message})
	}

	_, err := a.Analyze(context.Background(), "test-repo", "", []string{goFile}, WithAnalyzeProgress(progress))
	if err != nil {
		t.Fatalf("Analyze with progress failed: %v", err)
	}

	mu.Lock()
	defer mu.Unlock()
	if len(calls) == 0 {
		t.Error("progress callback was never called during analysis")
	}

	foundAnalyzing := false
	for _, c := range calls {
		if c.stage == "analyzing" {
			foundAnalyzing = true
			break
		}
	}
	if !foundAnalyzing {
		t.Error("progress callback was never called with stage 'analyzing'")
	}
}

// TestWithAnalyzeProgress_ReportsFilePaths verifies that progress messages
// contain file paths during analysis.
// Expected result: at least one progress message contains the file path.
func TestWithAnalyzeProgress_ReportsFilePaths(t *testing.T) {
	root := t.TempDir()
	goFile := filepath.Join(root, "main.go")
	if err := os.WriteFile(goFile, []byte("package main\n\nfunc main() {}\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	registry := NewRegistry()
	a := NewAnalyzer(registry, nil)

	var mu sync.Mutex
	var messages []string
	progress := func(_, message string) {
		mu.Lock()
		defer mu.Unlock()
		messages = append(messages, message)
	}

	_, err := a.Analyze(context.Background(), "test-repo", "", []string{goFile}, WithAnalyzeProgress(progress))
	if err != nil {
		t.Fatalf("Analyze with progress failed: %v", err)
	}

	mu.Lock()
	defer mu.Unlock()
	foundPath := false
	for _, msg := range messages {
		if msg == goFile || filepath.Ext(msg) != "" {
			foundPath = true
			break
		}
	}
	if !foundPath {
		t.Errorf("progress messages should contain file paths, got: %v", messages)
	}
}

// TestWithAnalyzeProgress_NilProgressSafe verifies that nil progress doesn't panic.
// Expected result: no panic, Analyze works normally.
func TestWithAnalyzeProgress_NilProgressSafe(t *testing.T) {
	registry := NewRegistry()
	a := NewAnalyzer(registry, nil)

	_, err := a.Analyze(context.Background(), "test-repo", "", nil, WithAnalyzeProgress(nil))
	if err != nil {
		t.Errorf("Analyze with nil progress should not fail: %v", err)
	}
}
