package analysis

import (
	sitter "github.com/smacker/go-tree-sitter"
)

// Extractor extracts symbols and references from a parsed tree-sitter AST.
// Each supported language implements this interface.
type Extractor interface {
	// ExtractSymbols walks the AST and returns all symbols (functions, classes, modules)
	// found in the file. repoPath is the absolute filesystem path to the repository root.
	ExtractSymbols(tree *sitter.Tree, source []byte, filePath string, repoPath string) ([]Symbol, error)

	// ExtractReferences walks the AST and returns all cross-symbol references
	// (calls, imports, inheritance, etc.) found in the file. repoPath is needed
	// to classify imports as internal vs external (e.g., reading go.mod module name).
	ExtractReferences(tree *sitter.Tree, source []byte, filePath string, repoPath string) ([]Reference, error)
}

// ComplexityExtractor computes cyclomatic complexity for function AST subtrees.
// Each supported language implements this interface.
type ComplexityExtractor interface {
	// ComputeComplexity walks the function's AST subtree and counts decision points.
	// Returns base complexity of 1 + count of decision nodes.
	ComputeComplexity(node *sitter.Node, source []byte) int
}
