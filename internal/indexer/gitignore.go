package indexer

import (
	"os"
	"path/filepath"

	ignore "github.com/sabhiram/go-gitignore"
)

// GitIgnoreMatcher handles .gitignore pattern loading and matching.
// Loads root .gitignore and nested .gitignore files as directories are entered.
type GitIgnoreMatcher struct {
	repoPath string
	matchers []*ignore.GitIgnore
}

// NewGitIgnoreMatcher creates a matcher initialized with the root .gitignore
// and .cctxignore patterns from the given repository path. Always skips .git/ directory.
// .cctxignore uses the same syntax as .gitignore and lets users exclude paths
// from the ingest pipeline without modifying .gitignore.
func NewGitIgnoreMatcher(repoPath string) (*GitIgnoreMatcher, error) {
	m := &GitIgnoreMatcher{repoPath: repoPath}

	for _, name := range []string{".gitignore", ".cctxignore"} {
		path := filepath.Join(repoPath, name)
		if _, err := os.Stat(path); err == nil {
			gi, err := ignore.CompileIgnoreFile(path)
			if err != nil {
				return nil, err
			}
			m.matchers = append(m.matchers, gi)
		}
	}

	return m, nil
}

// ShouldIgnore returns true if the given path (relative to repo root)
// should be ignored based on .gitignore patterns.
func (m *GitIgnoreMatcher) ShouldIgnore(relPath string, isDir bool) bool {
	// Always skip .git directory
	base := filepath.Base(relPath)
	if base == ".git" && isDir {
		return true
	}

	// Skip submodule directories (they contain a .git file, not a .git directory)
	if isDir && m.isSubmodule(relPath) {
		return true
	}

	for _, gi := range m.matchers {
		if gi.MatchesPath(relPath) {
			return true
		}
		// Directory patterns like "build/" need a trailing slash to match
		if isDir && gi.MatchesPath(relPath+"/") {
			return true
		}
	}
	return false
}

// isSubmodule checks if the given directory (relative to repo root) is a git
// submodule by looking for a .git file (not directory) inside it.
func (m *GitIgnoreMatcher) isSubmodule(relPath string) bool {
	gitPath := filepath.Join(m.repoPath, relPath, ".git")
	fi, err := os.Lstat(gitPath)
	if err != nil {
		return false
	}
	return !fi.IsDir()
}

// EnterDirectory loads any .gitignore or .cctxignore files found in the given
// directory, extending the current pattern set. Patterns are cumulative.
func (m *GitIgnoreMatcher) EnterDirectory(dirPath string) {
	for _, name := range []string{".gitignore", ".cctxignore"} {
		path := filepath.Join(dirPath, name)
		if _, err := os.Stat(path); err == nil {
			gi, err := ignore.CompileIgnoreFile(path)
			if err == nil {
				m.matchers = append(m.matchers, gi)
			}
		}
	}
}
