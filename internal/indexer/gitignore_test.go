package indexer

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

// TestCctxignoreRootPattern verifies that patterns in .cctxignore are loaded
// and used to exclude paths from indexing.
func TestCctxignoreRootPattern(t *testing.T) {
	root := t.TempDir()

	// Create directory structure
	os.MkdirAll(filepath.Join(root, "src"), 0o755)
	os.MkdirAll(filepath.Join(root, "vendor", "lib"), 0o755)

	// .cctxignore excludes vendor/
	os.WriteFile(filepath.Join(root, ".cctxignore"), []byte("vendor\n"), 0o644)

	m, err := NewGitIgnoreMatcher(root)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	tests := []struct {
		name   string
		path   string
		isDir  bool
		expect bool
	}{
		{"src dir not ignored", "src", true, false},
		{"vendor dir ignored by cctxignore", "vendor", true, true},
		{"vendor subdir ignored by cctxignore", "vendor/lib", true, true},
		{"file in src not ignored", "src/main.go", false, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := m.ShouldIgnore(tt.path, tt.isDir)
			if got != tt.expect {
				t.Errorf("ShouldIgnore(%q, %v) = %v, want %v", tt.path, tt.isDir, got, tt.expect)
			}
		})
	}
}

// TestCctxignoreAndGitignoreCombine verifies that both .gitignore and .cctxignore
// patterns are applied together.
func TestCctxignoreAndGitignoreCombine(t *testing.T) {
	root := t.TempDir()

	os.MkdirAll(filepath.Join(root, "build"), 0o755)
	os.MkdirAll(filepath.Join(root, "vendor"), 0o755)
	os.MkdirAll(filepath.Join(root, "src"), 0o755)

	// .gitignore excludes build/
	os.WriteFile(filepath.Join(root, ".gitignore"), []byte("build\n"), 0o644)
	// .cctxignore excludes vendor/
	os.WriteFile(filepath.Join(root, ".cctxignore"), []byte("vendor\n"), 0o644)

	m, err := NewGitIgnoreMatcher(root)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	tests := []struct {
		name   string
		path   string
		isDir  bool
		expect bool
	}{
		{"build ignored by gitignore", "build", true, true},
		{"vendor ignored by cctxignore", "vendor", true, true},
		{"src not ignored by either", "src", true, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := m.ShouldIgnore(tt.path, tt.isDir)
			if got != tt.expect {
				t.Errorf("ShouldIgnore(%q, %v) = %v, want %v", tt.path, tt.isDir, got, tt.expect)
			}
		})
	}
}

// TestCctxignoreNested verifies that .cctxignore in subdirectories is loaded
// when EnterDirectory is called.
func TestCctxignoreNested(t *testing.T) {
	root := t.TempDir()

	subdir := filepath.Join(root, "internal")
	os.MkdirAll(filepath.Join(subdir, "generated"), 0o755)
	os.MkdirAll(filepath.Join(subdir, "core"), 0o755)

	// Nested .cctxignore in internal/
	os.WriteFile(filepath.Join(subdir, ".cctxignore"), []byte("generated\n"), 0o644)

	m, err := NewGitIgnoreMatcher(root)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Before entering internal/, "generated" is not ignored
	if m.ShouldIgnore("generated", true) {
		t.Error("generated should not be ignored before EnterDirectory")
	}

	// Enter the internal/ directory
	m.EnterDirectory(subdir)

	// After entering internal/, "generated" should be ignored
	if !m.ShouldIgnore("generated", true) {
		t.Error("generated should be ignored after EnterDirectory loads nested .cctxignore")
	}
	if m.ShouldIgnore("core", true) {
		t.Error("core should not be ignored")
	}
}

// TestCctxignoreMissing verifies no error when .cctxignore doesn't exist.
func TestCctxignoreMissing(t *testing.T) {
	root := t.TempDir()

	m, err := NewGitIgnoreMatcher(root)
	if err != nil {
		t.Fatalf("unexpected error when no ignore files: %v", err)
	}

	if m.ShouldIgnore("anything", false) {
		t.Error("nothing should be ignored with no ignore files")
	}
}

// TestCctxignoreInIndexResult verifies that files excluded by .cctxignore
// are not counted as indexed by IndexRepository.
func TestCctxignoreInIndexResult(t *testing.T) {
	root := t.TempDir()

	// Create dirs + files
	os.MkdirAll(filepath.Join(root, "src"), 0o755)
	os.MkdirAll(filepath.Join(root, "vendor", "dep"), 0o755)
	os.WriteFile(filepath.Join(root, "src", "main.go"), []byte("package main\n"), 0o644)
	os.WriteFile(filepath.Join(root, "vendor", "dep", "lib.go"), []byte("package dep\n"), 0o644)

	// Exclude vendor via .cctxignore
	os.WriteFile(filepath.Join(root, ".cctxignore"), []byte("vendor\n"), 0o644)

	idx := NewIndexer(nil) // no DB, just walker
	result, err := idx.IndexRepository(context.Background(), root)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Should index src/main.go + .cctxignore but not vendor/dep/lib.go
	if result.FilesIndexed != 2 {
		t.Errorf("expected 2 files indexed (main.go + .cctxignore), got %d", result.FilesIndexed)
	}
	// Should index src/ but not vendor/ or vendor/dep/
	if result.FoldersIndexed != 1 {
		t.Errorf("expected 1 folder indexed, got %d", result.FoldersIndexed)
	}
}
