package indexer

import (
	"os"
	"path/filepath"
	"strings"

	ignore "github.com/sabhiram/go-gitignore"
)

// scopedMatcher pairs a compiled gitignore with the repo-root-relative directory
// it was loaded from, so patterns are checked against paths relative to that directory.
type scopedMatcher struct {
	gi     *ignore.GitIgnore
	relDir string // "" for root-level matchers
}

// GitIgnoreMatcher handles .gitignore pattern loading and matching.
// Loads root .gitignore and nested .gitignore files as directories are entered.
// Each matcher is scoped to the directory it was loaded from, so patterns
// like "megatron-api" in services/megatron-api/.gitignore only match paths
// relative to that directory (not ancestor path components).
type GitIgnoreMatcher struct {
	repoPath string
	matchers []scopedMatcher
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
			m.matchers = append(m.matchers, scopedMatcher{gi: gi, relDir: ""})
		}
	}

	return m, nil
}

// ShouldIgnore returns true if the given path (relative to repo root)
// should be ignored based on .gitignore patterns. Each nested .gitignore
// matcher only checks paths relative to its own directory.
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

	for _, sm := range m.matchers {
		checkPath := relPathForMatcher(relPath, sm.relDir)
		if checkPath == "" {
			// Path is not under this matcher's directory scope
			continue
		}
		if sm.gi.MatchesPath(checkPath) {
			return true
		}
		// Directory patterns like "build/" need a trailing slash to match
		if isDir && sm.gi.MatchesPath(checkPath+"/") {
			return true
		}
	}
	return false
}

// relPathForMatcher returns the path to check against a matcher scoped to relDir.
// For root matchers (relDir == ""), returns relPath unchanged.
// For nested matchers, strips the relDir prefix so patterns are checked relative
// to the directory containing the .gitignore.
// Returns "" if relPath is not under relDir (meaning this matcher doesn't apply).
func relPathForMatcher(relPath, relDir string) string {
	if relDir == "" {
		return relPath
	}
	prefix := relDir + string(filepath.Separator)
	if !strings.HasPrefix(relPath, prefix) {
		return ""
	}
	return relPath[len(prefix):]
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
// directory, extending the current pattern set. Each loaded matcher is scoped
// to dirPath so its patterns only apply to paths within that directory.
func (m *GitIgnoreMatcher) EnterDirectory(dirPath string) {
	relDir, err := filepath.Rel(m.repoPath, dirPath)
	if err != nil {
		return
	}
	if relDir == "." {
		relDir = ""
	}

	for _, name := range []string{".gitignore", ".cctxignore"} {
		path := filepath.Join(dirPath, name)
		if _, err := os.Stat(path); err == nil {
			gi, err := ignore.CompileIgnoreFile(path)
			if err == nil {
				m.matchers = append(m.matchers, scopedMatcher{gi: gi, relDir: relDir})
			}
		}
	}
}
