package indexer

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"
)

// --- Task 8: Indexer directory walker core tests ---

// createTestRepo creates a temporary directory structure for testing:
//
//	repo/
//	  README.md (text file)
//	  src/
//	    main.go (text file)
//	    util.go (text file)
//	  docs/
//	    guide.md (text file)
func createTestRepo(t *testing.T) string {
	t.Helper()
	root := t.TempDir()

	dirs := []string{
		filepath.Join(root, "src"),
		filepath.Join(root, "docs"),
	}
	for _, d := range dirs {
		if err := os.MkdirAll(d, 0o755); err != nil {
			t.Fatalf("failed to create dir %s: %v", d, err)
		}
	}

	files := map[string]string{
		filepath.Join(root, "README.md"):      "# Test Repo\n",
		filepath.Join(root, "src", "main.go"): "package main\n\nfunc main() {}\n",
		filepath.Join(root, "src", "util.go"): "package main\n\nfunc helper() {}\n",
		filepath.Join(root, "docs", "guide.md"): "# Guide\n",
	}
	for path, content := range files {
		if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
			t.Fatalf("failed to write %s: %v", path, err)
		}
	}

	return root
}

// TestNewIndexer_ReturnsNonNil verifies that NewIndexer returns a non-nil Indexer.
// Expected result: Non-nil *Indexer.
func TestNewIndexer_ReturnsNonNil(t *testing.T) {
	idx := NewIndexer(nil) // nil db is acceptable for construction
	if idx == nil {
		t.Error("NewIndexer returned nil, expected non-nil *Indexer")
	}
}

// TestIndexRepository_ReturnsNonZeroCounts verifies that indexing a directory
// with known files/folders returns correct counts in IndexResult.
// Expected result: FilesIndexed=4, FoldersIndexed=2 for the test repo.
func TestIndexRepository_ReturnsNonZeroCounts(t *testing.T) {
	repoPath := createTestRepo(t)
	idx := NewIndexer(nil)

	result, err := idx.IndexRepository(context.Background(), repoPath)
	if err != nil {
		t.Fatalf("IndexRepository returned error: %v", err)
	}

	if result.FilesIndexed != 4 {
		t.Errorf("FilesIndexed = %d, want 4", result.FilesIndexed)
	}
	if result.FoldersIndexed != 2 {
		t.Errorf("FoldersIndexed = %d, want 2", result.FoldersIndexed)
	}
}

// TestIndexRepository_SetsRepoID verifies that IndexResult.RepoID is set to
// a non-empty value after indexing.
// Expected result: Non-empty RepoID.
func TestIndexRepository_SetsRepoID(t *testing.T) {
	repoPath := createTestRepo(t)
	idx := NewIndexer(nil)

	result, err := idx.IndexRepository(context.Background(), repoPath)
	if err != nil {
		t.Fatalf("IndexRepository returned error: %v", err)
	}

	if result.RepoID == "" {
		t.Error("RepoID is empty, expected non-empty string")
	}
}

// TestIndexRepository_NonExistentPath verifies that IndexRepository returns
// an error when given a path that does not exist.
// Expected result: Non-nil error.
func TestIndexRepository_NonExistentPath(t *testing.T) {
	idx := NewIndexer(nil)
	_, err := idx.IndexRepository(context.Background(), "/nonexistent/path/that/does/not/exist")
	if err == nil {
		t.Error("IndexRepository with non-existent path should return error, got nil")
	}
}

// TestIndexRepository_FileNotDirectory verifies that IndexRepository returns
// an error when given a path to a file (not a directory).
// Expected result: Non-nil error.
func TestIndexRepository_FileNotDirectory(t *testing.T) {
	tmpFile := filepath.Join(t.TempDir(), "file.txt")
	if err := os.WriteFile(tmpFile, []byte("hello"), 0o644); err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	idx := NewIndexer(nil)
	_, err := idx.IndexRepository(context.Background(), tmpFile)
	if err == nil {
		t.Error("IndexRepository with file path (not dir) should return error, got nil")
	}
}

// TestIndexRepository_EmptyDirectory verifies that indexing an empty directory
// returns zero counts but no error.
// Expected result: FilesIndexed=0, FoldersIndexed=0, no error.
func TestIndexRepository_EmptyDirectory(t *testing.T) {
	emptyDir := t.TempDir()
	idx := NewIndexer(nil)

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
}

// TestIndexRepository_CollectsErrors verifies that indexing continues when
// individual files fail and errors are collected in IndexResult.Errors.
// Expected result: Errors slice is populated, indexing completes.
func TestIndexRepository_CollectsErrors(t *testing.T) {
	root := t.TempDir()
	// Create a file, then make it unreadable
	unreadable := filepath.Join(root, "secret.go")
	if err := os.WriteFile(unreadable, []byte("package x"), 0o000); err != nil {
		t.Fatalf("failed to create unreadable file: %v", err)
	}
	t.Cleanup(func() { os.Chmod(unreadable, 0o644) })

	idx := NewIndexer(nil)
	result, err := idx.IndexRepository(context.Background(), root)
	if err != nil {
		t.Fatalf("IndexRepository should not return top-level error for permission issues: %v", err)
	}

	if len(result.Errors) == 0 {
		t.Error("expected Errors to contain permission error, got empty slice")
	}
}

// --- Task 9: Gitignore support tests ---

// TestGitIgnoreMatcher_SkipsGitDir verifies that .git/ directory is always skipped.
// Expected result: ShouldIgnore returns true for ".git" directory.
func TestGitIgnoreMatcher_SkipsGitDir(t *testing.T) {
	repoPath := t.TempDir()
	matcher, err := NewGitIgnoreMatcher(repoPath)
	if err != nil {
		t.Fatalf("NewGitIgnoreMatcher failed: %v", err)
	}

	if !matcher.ShouldIgnore(".git", true) {
		t.Error("ShouldIgnore(.git, true) = false, want true (.git always skipped)")
	}
}

// TestGitIgnoreMatcher_RespectsRootGitignore verifies that patterns from
// the root .gitignore file are applied.
// Expected result: Files matching patterns are ignored.
func TestGitIgnoreMatcher_RespectsRootGitignore(t *testing.T) {
	repoPath := t.TempDir()
	gitignoreContent := "*.log\nbuild/\n"
	if err := os.WriteFile(filepath.Join(repoPath, ".gitignore"), []byte(gitignoreContent), 0o644); err != nil {
		t.Fatalf("failed to write .gitignore: %v", err)
	}

	matcher, err := NewGitIgnoreMatcher(repoPath)
	if err != nil {
		t.Fatalf("NewGitIgnoreMatcher failed: %v", err)
	}

	tests := []struct {
		path    string
		isDir   bool
		ignored bool
	}{
		{"debug.log", false, true},
		{"build", true, true},
		{"main.go", false, false},
		{"src", true, false},
	}

	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			got := matcher.ShouldIgnore(tt.path, tt.isDir)
			if got != tt.ignored {
				t.Errorf("ShouldIgnore(%q, %v) = %v, want %v", tt.path, tt.isDir, got, tt.ignored)
			}
		})
	}
}

// TestGitIgnoreMatcher_NoGitignoreFile verifies that when no .gitignore exists,
// no files are ignored (except .git/).
// Expected result: Only .git is ignored.
func TestGitIgnoreMatcher_NoGitignoreFile(t *testing.T) {
	repoPath := t.TempDir()
	matcher, err := NewGitIgnoreMatcher(repoPath)
	if err != nil {
		t.Fatalf("NewGitIgnoreMatcher failed: %v", err)
	}

	if matcher.ShouldIgnore("main.go", false) {
		t.Error("ShouldIgnore(main.go) = true, want false (no .gitignore)")
	}
	if matcher.ShouldIgnore("build", true) {
		t.Error("ShouldIgnore(build) = true, want false (no .gitignore)")
	}
}

// TestGitIgnoreMatcher_NestedGitignore verifies that nested .gitignore files
// extend the pattern set cumulatively.
// Expected result: Nested patterns are applied in addition to root patterns.
func TestGitIgnoreMatcher_NestedGitignore(t *testing.T) {
	repoPath := t.TempDir()

	// Root .gitignore ignores *.log
	if err := os.WriteFile(filepath.Join(repoPath, ".gitignore"), []byte("*.log\n"), 0o644); err != nil {
		t.Fatalf("failed to write root .gitignore: %v", err)
	}

	// Create src/ with its own .gitignore ignoring *.tmp
	srcDir := filepath.Join(repoPath, "src")
	if err := os.MkdirAll(srcDir, 0o755); err != nil {
		t.Fatalf("failed to create src dir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(srcDir, ".gitignore"), []byte("*.tmp\n"), 0o644); err != nil {
		t.Fatalf("failed to write src/.gitignore: %v", err)
	}

	matcher, err := NewGitIgnoreMatcher(repoPath)
	if err != nil {
		t.Fatalf("NewGitIgnoreMatcher failed: %v", err)
	}

	// Enter src/ to load nested gitignore
	matcher.EnterDirectory(srcDir)

	// After entering src/, both *.log (root) and *.tmp (nested) should be ignored
	if !matcher.ShouldIgnore("src/data.tmp", false) {
		t.Error("ShouldIgnore(src/data.tmp) = false, want true (nested .gitignore)")
	}
	if !matcher.ShouldIgnore("src/debug.log", false) {
		t.Error("ShouldIgnore(src/debug.log) = false, want true (root .gitignore still applies)")
	}
}

// TestIndexRepository_SkipsGitIgnoredFiles verifies that IndexRepository
// respects .gitignore patterns and skips matching files/folders.
// Expected result: FilesSkipped > 0 when .gitignore excludes files.
func TestIndexRepository_SkipsGitIgnoredFiles(t *testing.T) {
	root := t.TempDir()

	// Create files
	if err := os.WriteFile(filepath.Join(root, "main.go"), []byte("package main"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "debug.log"), []byte("log data"), 0o644); err != nil {
		t.Fatal(err)
	}
	// Create .gitignore
	if err := os.WriteFile(filepath.Join(root, ".gitignore"), []byte("*.log\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	idx := NewIndexer(nil)
	result, err := idx.IndexRepository(context.Background(), root)
	if err != nil {
		t.Fatalf("IndexRepository error: %v", err)
	}

	if result.FilesSkipped == 0 {
		t.Error("FilesSkipped = 0, want > 0 (.gitignore should skip *.log files)")
	}
}

// --- Task 10: Binary detection, language detection, incremental re-index tests ---

// TestIsBinary_DetectsNullBytes verifies that files containing null bytes
// in the first 512 bytes are classified as binary.
// Expected result: IsBinary returns true for binary files.
func TestIsBinary_DetectsNullBytes(t *testing.T) {
	binaryFile := filepath.Join(t.TempDir(), "binary.dat")
	content := make([]byte, 100)
	content[50] = 0x00 // null byte
	if err := os.WriteFile(binaryFile, content, 0o644); err != nil {
		t.Fatalf("failed to create binary file: %v", err)
	}

	isBin, err := IsBinary(binaryFile)
	if err != nil {
		t.Fatalf("IsBinary returned error: %v", err)
	}
	if !isBin {
		t.Error("IsBinary = false for file with null byte, want true")
	}
}

// TestIsBinary_TextFileReturnsFalse verifies that text files without null
// bytes are not classified as binary.
// Expected result: IsBinary returns false for text files.
func TestIsBinary_TextFileReturnsFalse(t *testing.T) {
	textFile := filepath.Join(t.TempDir(), "text.go")
	if err := os.WriteFile(textFile, []byte("package main\n\nfunc main() {}\n"), 0o644); err != nil {
		t.Fatalf("failed to create text file: %v", err)
	}

	isBin, err := IsBinary(textFile)
	if err != nil {
		t.Fatalf("IsBinary returned error: %v", err)
	}
	if isBin {
		t.Error("IsBinary = true for text file, want false")
	}
}

// TestDetectLanguage verifies that file extensions are correctly mapped
// to language strings.
// Expected result: Each extension maps to the expected language.
func TestDetectLanguage(t *testing.T) {
	tests := []struct {
		filePath string
		language string
	}{
		{"main.go", "go"},
		{"app.ts", "typescript"},
		{"component.tsx", "tsx"},
		{"script.js", "javascript"},
		{"component.jsx", "jsx"},
		{"app.py", "python"},
		{"server.rb", "ruby"},
		{"Main.java", "java"},
		{"lib.rs", "rust"},
		{"main.c", "c"},
		{"util.h", "c"},
		{"main.cpp", "cpp"},
		{"util.hpp", "cpp"},
		{"main.cc", "cpp"},
		{"unknown.xyz", ""},
		{"noextension", ""},
	}

	for _, tt := range tests {
		t.Run(tt.filePath, func(t *testing.T) {
			got := DetectLanguage(tt.filePath)
			if got != tt.language {
				t.Errorf("DetectLanguage(%q) = %q, want %q", tt.filePath, got, tt.language)
			}
		})
	}
}

// TestCountLines verifies that CountLines correctly counts the number of
// lines in a text file.
// Expected result: Correct line count for known content.
func TestCountLines(t *testing.T) {
	tests := []struct {
		name    string
		content string
		lines   int
	}{
		{"empty", "", 0},
		{"one line", "hello\n", 1},
		{"three lines", "line1\nline2\nline3\n", 3},
		{"no trailing newline", "line1\nline2", 2},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := filepath.Join(t.TempDir(), "test.txt")
			if err := os.WriteFile(f, []byte(tt.content), 0o644); err != nil {
				t.Fatal(err)
			}
			got, err := CountLines(f)
			if err != nil {
				t.Fatalf("CountLines error: %v", err)
			}
			if got != tt.lines {
				t.Errorf("CountLines = %d, want %d", got, tt.lines)
			}
		})
	}
}

// TestIsSymlink verifies that symlinks are correctly detected.
// Expected result: IsSymlink returns true for symlinks, false for regular files.
func TestIsSymlink(t *testing.T) {
	dir := t.TempDir()
	realFile := filepath.Join(dir, "real.txt")
	if err := os.WriteFile(realFile, []byte("hello"), 0o644); err != nil {
		t.Fatal(err)
	}
	symlink := filepath.Join(dir, "link.txt")
	if err := os.Symlink(realFile, symlink); err != nil {
		t.Fatal(err)
	}

	t.Run("regular file", func(t *testing.T) {
		isLink, err := IsSymlink(realFile)
		if err != nil {
			t.Fatalf("IsSymlink error: %v", err)
		}
		if isLink {
			t.Error("IsSymlink = true for regular file, want false")
		}
	})

	t.Run("symlink", func(t *testing.T) {
		isLink, err := IsSymlink(symlink)
		if err != nil {
			t.Fatalf("IsSymlink error: %v", err)
		}
		if !isLink {
			t.Error("IsSymlink = false for symlink, want true")
		}
	})
}

// TestIndexRepository_SkipsBinaryFiles verifies that binary files are not
// indexed and are counted in FilesSkipped.
// Expected result: Binary files are skipped.
func TestIndexRepository_SkipsBinaryFiles(t *testing.T) {
	root := t.TempDir()

	// Create a text file
	if err := os.WriteFile(filepath.Join(root, "main.go"), []byte("package main"), 0o644); err != nil {
		t.Fatal(err)
	}
	// Create a binary file
	binary := make([]byte, 100)
	binary[10] = 0x00
	if err := os.WriteFile(filepath.Join(root, "data.bin"), binary, 0o644); err != nil {
		t.Fatal(err)
	}

	idx := NewIndexer(nil)
	result, err := idx.IndexRepository(context.Background(), root)
	if err != nil {
		t.Fatalf("IndexRepository error: %v", err)
	}

	if result.FilesSkipped == 0 {
		t.Error("FilesSkipped = 0, want > 0 (binary file should be skipped)")
	}
	// main.go should be indexed, data.bin should be skipped
	if result.FilesIndexed != 1 {
		t.Errorf("FilesIndexed = %d, want 1 (only main.go)", result.FilesIndexed)
	}
}

// --- Task 2: pendingFolder, pendingFile types, and batchSize constant ---

// TestPendingFolder_HasExpectedFields verifies that pendingFolder struct has
// Path (string), ParentPath (string), and ModTime (time.Time) fields.
// Expected result: Compiles and fields are assignable.
func TestPendingFolder_HasExpectedFields(t *testing.T) {
	pf := pendingFolder{
		Path:       "src/utils",
		ParentPath: "src",
		ModTime:    time.Now(),
	}
	if pf.Path != "src/utils" {
		t.Errorf("pendingFolder.Path = %q, want %q", pf.Path, "src/utils")
	}
	if pf.ParentPath != "src" {
		t.Errorf("pendingFolder.ParentPath = %q, want %q", pf.ParentPath, "src")
	}
	if pf.ModTime.IsZero() {
		t.Error("pendingFolder.ModTime is zero, expected non-zero")
	}
}

// TestPendingFile_HasExpectedFields verifies that pendingFile struct has
// Path (string), ParentPath (string), Language (string), LineCount (int),
// and ModTime (time.Time) fields.
// Expected result: Compiles and fields are assignable.
func TestPendingFile_HasExpectedFields(t *testing.T) {
	pf := pendingFile{
		Path:       "src/main.go",
		ParentPath: "src",
		Language:   "go",
		LineCount:  42,
		ModTime:    time.Now(),
	}
	if pf.Path != "src/main.go" {
		t.Errorf("pendingFile.Path = %q, want %q", pf.Path, "src/main.go")
	}
	if pf.ParentPath != "src" {
		t.Errorf("pendingFile.ParentPath = %q, want %q", pf.ParentPath, "src")
	}
	if pf.Language != "go" {
		t.Errorf("pendingFile.Language = %q, want %q", pf.Language, "go")
	}
	if pf.LineCount != 42 {
		t.Errorf("pendingFile.LineCount = %d, want 42", pf.LineCount)
	}
	if pf.ModTime.IsZero() {
		t.Error("pendingFile.ModTime is zero, expected non-zero")
	}
}

// TestBatchSize_Is10 verifies that batchSize constant is 10.
// Kept small to avoid OOM-killing FalkorDB in memory-constrained environments.
func TestBatchSize_Is10(t *testing.T) {
	if batchSize != 10 {
		t.Errorf("batchSize = %d, want 10", batchSize)
	}
}

// TestIndexRepository_SkipsSymlinks verifies that symlinks (both files and dirs)
// are skipped during indexing.
// Expected result: Symlinks are not followed or indexed.
func TestIndexRepository_SkipsSymlinks(t *testing.T) {
	root := t.TempDir()

	// Create a real file and directory
	if err := os.WriteFile(filepath.Join(root, "real.go"), []byte("package main"), 0o644); err != nil {
		t.Fatal(err)
	}
	realDir := filepath.Join(root, "realdir")
	if err := os.MkdirAll(realDir, 0o755); err != nil {
		t.Fatal(err)
	}

	// Create symlinks
	if err := os.Symlink(filepath.Join(root, "real.go"), filepath.Join(root, "link.go")); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(realDir, filepath.Join(root, "linkdir")); err != nil {
		t.Fatal(err)
	}

	idx := NewIndexer(nil)
	result, err := idx.IndexRepository(context.Background(), root)
	if err != nil {
		t.Fatalf("IndexRepository error: %v", err)
	}

	// Only real.go should be indexed (1 file), not link.go
	// realdir should be indexed (1 folder), not linkdir
	if result.FilesIndexed != 1 {
		t.Errorf("FilesIndexed = %d, want 1 (only real.go, not symlink)", result.FilesIndexed)
	}
}
