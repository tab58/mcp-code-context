package indexer

import (
	"bytes"
	"io"
	"os"
	"path/filepath"
	"strings"
)

// extensionToLanguage maps file extensions to language strings.
var extensionToLanguage = map[string]string{
	".go":  "go",
	".ts":  "typescript",
	".tsx": "tsx",
	".js":  "javascript",
	".jsx": "jsx",
	".py":  "python",
	".rb":  "ruby",
	".java": "java",
	".rs":  "rust",
	".c":   "c",
	".h":   "c",
	".cpp": "cpp",
	".hpp": "cpp",
	".cc":  "cpp",
}

// binaryCheckBufSize is the number of bytes read from the start of a file
// to detect binary content. Matches the heuristic used by Git.
const binaryCheckBufSize = 512

// IsBinary checks if a file is binary by reading the first binaryCheckBufSize
// bytes and looking for null bytes (0x00).
func IsBinary(filePath string) (bool, error) {
	f, err := os.Open(filePath)
	if err != nil {
		return false, err
	}
	defer f.Close()

	buf := make([]byte, binaryCheckBufSize)
	n, err := f.Read(buf)
	if err != nil && err != io.EOF {
		return false, err
	}

	return bytes.Contains(buf[:n], []byte{0x00}), nil
}

// DetectLanguage maps a file extension to a language string.
// Returns empty string for unknown extensions.
func DetectLanguage(filePath string) string {
	ext := strings.ToLower(filepath.Ext(filePath))
	return extensionToLanguage[ext]
}

// CountLines returns the number of lines in a text file.
func CountLines(filePath string) (int, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return 0, err
	}
	if len(data) == 0 {
		return 0, nil
	}
	count := bytes.Count(data, []byte{'\n'})
	// If file doesn't end with newline, the last line still counts
	if len(data) > 0 && data[len(data)-1] != '\n' {
		count++
	}
	return count, nil
}

// IsSymlink checks if the given path is a symbolic link.
func IsSymlink(filePath string) (bool, error) {
	fi, err := os.Lstat(filePath)
	if err != nil {
		return false, err
	}
	return fi.Mode()&os.ModeSymlink != 0, nil
}
