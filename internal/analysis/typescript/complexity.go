package typescript

import (
	sitter "github.com/smacker/go-tree-sitter"
)

// TSComplexityExtractor computes cyclomatic complexity for TypeScript/TSX
// function AST subtrees.
type TSComplexityExtractor struct{}

// NewTSComplexityExtractor creates a new TypeScript complexity extractor.
func NewTSComplexityExtractor() *TSComplexityExtractor {
	return &TSComplexityExtractor{}
}

// ComputeComplexity walks the function's AST subtree and counts decision points.
// Returns base complexity of 1 + count of decision nodes.
// Rules: if_statement, for_statement, while_statement, do_statement,
// switch_case, "&&", "||", "??", optional_chain_expression, catch_clause,
// ternary_expression.
func (e *TSComplexityExtractor) ComputeComplexity(node *sitter.Node, source []byte) int {
	if node == nil {
		return 1
	}
	complexity := 1
	walkTSNode(node, source, &complexity)
	return complexity
}

// walkTSNode recursively visits AST nodes and increments complexity for decision points.
func walkTSNode(node *sitter.Node, source []byte, complexity *int) {
	nodeType := node.Type()

	switch nodeType {
	case "if_statement", "for_statement", "for_in_statement", "while_statement", "do_statement",
		"switch_case", "catch_clause", "ternary_expression":
		*complexity++
	case "binary_expression":
		opNode := node.ChildByFieldName("operator")
		if opNode != nil {
			op := opNode.Content(source)
			if op == "&&" || op == "||" || op == "??" {
				*complexity++
			}
		}
	case "optional_chain":
		*complexity++
	}

	for i := 0; i < int(node.ChildCount()); i++ {
		child := node.Child(i)
		if child != nil {
			walkTSNode(child, source, complexity)
		}
	}
}
