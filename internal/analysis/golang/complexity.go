package golang

import (
	sitter "github.com/smacker/go-tree-sitter"
)

// GoComplexityExtractor computes cyclomatic complexity for Go function AST subtrees.
type GoComplexityExtractor struct{}

// NewGoComplexityExtractor creates a new Go complexity extractor.
func NewGoComplexityExtractor() *GoComplexityExtractor {
	return &GoComplexityExtractor{}
}

// ComputeComplexity walks the function's AST subtree and counts decision points.
// Returns base complexity of 1 + count of decision nodes.
// Rules: if_statement +1, for_statement +1, expression_case +1,
// communication_case +1, "&&" +1, "||" +1.
func (e *GoComplexityExtractor) ComputeComplexity(node *sitter.Node, source []byte) int {
	if node == nil {
		return 1
	}
	complexity := 1
	walkNode(node, source, &complexity)
	return complexity
}

// walkNode recursively visits AST nodes and increments complexity for decision points.
func walkNode(node *sitter.Node, source []byte, complexity *int) {
	nodeType := node.Type()

	switch nodeType {
	case "if_statement", "for_statement", "expression_case", "communication_case":
		*complexity++
	}

	// Check for logical operators in binary expressions
	if nodeType == "binary_expression" {
		opNode := node.ChildByFieldName("operator")
		if opNode != nil {
			op := opNode.Content(source)
			if op == "&&" || op == "||" {
				*complexity++
			}
		}
	}

	for i := 0; i < int(node.ChildCount()); i++ {
		child := node.Child(i)
		if child != nil {
			walkNode(child, source, complexity)
		}
	}
}
