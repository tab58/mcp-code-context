package indexer

import (
	"context"
	"os"
	"path/filepath"
	"sync"
	"testing"
)

// === Task 3 + 10: IndexOption with progress callback ===

// TestWithProgress_NoOption verifies that IndexRepository works without any options.
// Expected result: backward-compatible, no crash.
func TestWithProgress_NoOption(t *testing.T) {
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, "main.go"), []byte("package main\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	idx := NewIndexer(nil) // no DB
	_, err := idx.IndexRepository(context.Background(), root)
	if err != nil {
		t.Errorf("IndexRepository with no options failed: %v", err)
	}
}

// TestWithProgress_ReceivesCallbacks verifies that the progress callback
// is called during indexing with stage="indexing" and file/folder paths.
// Expected result: progress function is called at least once with stage "indexing".
func TestWithProgress_ReceivesCallbacks(t *testing.T) {
	root := t.TempDir()
	srcDir := filepath.Join(root, "src")
	if err := os.MkdirAll(srcDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "main.go"), []byte("package main\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(srcDir, "util.go"), []byte("package src\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	var mu sync.Mutex
	var calls []struct{ stage, message string }
	progress := func(stage, message string) {
		mu.Lock()
		defer mu.Unlock()
		calls = append(calls, struct{ stage, message string }{stage, message})
	}

	idx := NewIndexer(nil)
	_, err := idx.IndexRepository(context.Background(), root, WithProgress(progress))
	if err != nil {
		t.Fatalf("IndexRepository with progress failed: %v", err)
	}

	mu.Lock()
	defer mu.Unlock()
	if len(calls) == 0 {
		t.Error("progress callback was never called during indexing")
	}

	foundIndexing := false
	for _, c := range calls {
		if c.stage == "indexing" {
			foundIndexing = true
			break
		}
	}
	if !foundIndexing {
		t.Error("progress callback was never called with stage 'indexing'")
	}
}

// TestWithProgress_ReportsFilePaths verifies that progress messages contain
// relative file paths during indexing.
// Expected result: at least one progress message contains a file path.
func TestWithProgress_ReportsFilePaths(t *testing.T) {
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, "main.go"), []byte("package main\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	var mu sync.Mutex
	var messages []string
	progress := func(stage, message string) {
		mu.Lock()
		defer mu.Unlock()
		messages = append(messages, message)
	}

	idx := NewIndexer(nil)
	_, err := idx.IndexRepository(context.Background(), root, WithProgress(progress))
	if err != nil {
		t.Fatalf("IndexRepository with progress failed: %v", err)
	}

	mu.Lock()
	defer mu.Unlock()
	foundPath := false
	for _, msg := range messages {
		if msg == "main.go" || filepath.Ext(msg) != "" {
			foundPath = true
			break
		}
	}
	if !foundPath {
		t.Errorf("progress messages should contain file paths, got: %v", messages)
	}
}

// TestIndexOption_NilProgress verifies that passing nil progress function
// does not cause a panic.
// Expected result: no panic, IndexRepository works normally.
func TestIndexOption_NilProgress(t *testing.T) {
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, "main.go"), []byte("package main\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	idx := NewIndexer(nil)
	// WithProgress(nil) should be safe
	_, err := idx.IndexRepository(context.Background(), root, WithProgress(nil))
	if err != nil {
		t.Errorf("IndexRepository with nil progress should not fail: %v", err)
	}
}
