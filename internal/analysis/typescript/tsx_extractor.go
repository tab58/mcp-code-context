package typescript

import (
	sitter "github.com/smacker/go-tree-sitter"
	"github.com/tab58/code-context/internal/analysis"
)

// TSXExtractor implements the Extractor interface for TSX source files.
// Uses the same extraction logic as TypeScript but with the TSX grammar.
type TSXExtractor struct{}

// NewTSXExtractor creates a new TSX language extractor.
func NewTSXExtractor() *TSXExtractor {
	return &TSXExtractor{}
}

// ExtractSymbols walks the TSX AST and returns all symbols found in the file.
func (e *TSXExtractor) ExtractSymbols(tree *sitter.Tree, source []byte, filePath string, repoPath string) ([]analysis.Symbol, error) {
	return extractTSSymbols(tree, source, filePath, repoPath, "tsx")
}

// ExtractReferences walks the TSX AST and returns all cross-symbol references.
func (e *TSXExtractor) ExtractReferences(tree *sitter.Tree, source []byte, filePath string, repoPath string) ([]analysis.Reference, error) {
	return extractTSReferences(tree, source, filePath)
}
