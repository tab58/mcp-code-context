package ruby

import (
	"strings"

	sitter "github.com/smacker/go-tree-sitter"
	"github.com/tab58/code-context/internal/analysis"
)

// Register registers the Ruby language extractor and complexity extractor
// with the given registry.
func Register(r *analysis.Registry) {
	r.RegisterExtractor("ruby", NewRubyExtractor())
	r.RegisterComplexityExtractor("ruby", NewRubyComplexityExtractor())
}

// RubyExtractor implements the Extractor interface for Ruby source files.
type RubyExtractor struct{}

// NewRubyExtractor creates a new Ruby language extractor.
func NewRubyExtractor() *RubyExtractor {
	return &RubyExtractor{}
}

// ExtractSymbols walks the Ruby AST and returns all symbols found in the file.
func (e *RubyExtractor) ExtractSymbols(tree *sitter.Tree, source []byte, filePath string, repoPath string) ([]analysis.Symbol, error) {
	root := tree.RootNode()
	var symbols []analysis.Symbol
	collectRbSymbols(root, source, filePath, "", "public", &symbols)
	return symbols, nil
}

// ExtractReferences walks the Ruby AST and returns all cross-symbol references.
func (e *RubyExtractor) ExtractReferences(tree *sitter.Tree, source []byte, filePath string, repoPath string) ([]analysis.Reference, error) {
	root := tree.RootNode()
	var refs []analysis.Reference
	collectRbRefs(root, source, filePath, &refs)
	return refs, nil
}

func collectRbSymbols(node *sitter.Node, source []byte, filePath, className, currentVisibility string, symbols *[]analysis.Symbol) {
	switch node.Type() {
	case "method":
		sym := extractRbMethod(node, source, filePath, className, currentVisibility)
		if sym != nil {
			*symbols = append(*symbols, *sym)
		}
		return

	case "class":
		sym, name := extractRbClass(node, source, filePath)
		if sym != nil {
			*symbols = append(*symbols, *sym)
		}
		// Recurse into class body with visibility tracking
		body := node.ChildByFieldName("body")
		if body != nil {
			vis := "public"
			for i := 0; i < int(body.NamedChildCount()); i++ {
				child := body.NamedChild(i)
				if newVis := checkVisibilityModifier(child, source); newVis != "" {
					vis = newVis
					continue
				}
				collectRbSymbols(child, source, filePath, name, vis, symbols)
			}
		}
		return

	case "module":
		sym := extractRbModule(node, source, filePath)
		if sym != nil {
			*symbols = append(*symbols, *sym)
		}
		// Recurse into module body
		body := node.ChildByFieldName("body")
		if body != nil {
			for i := 0; i < int(body.NamedChildCount()); i++ {
				collectRbSymbols(body.NamedChild(i), source, filePath, "", currentVisibility, symbols)
			}
		}
		return

	case "singleton_method":
		sym := extractRbSingletonMethod(node, source, filePath, className)
		if sym != nil {
			*symbols = append(*symbols, *sym)
		}
		return
	}

	for i := 0; i < int(node.NamedChildCount()); i++ {
		collectRbSymbols(node.NamedChild(i), source, filePath, className, currentVisibility, symbols)
	}
}

// checkVisibilityModifier returns the new visibility if the node is a bare
// visibility modifier call (private/public/protected without arguments).
func checkVisibilityModifier(node *sitter.Node, source []byte) string {
	if node.Type() == "identifier" {
		text := nodeText(node, source)
		if text == "private" || text == "public" || text == "protected" {
			return text
		}
	}
	return ""
}

func extractRbMethod(node *sitter.Node, source []byte, filePath, className, visibility string) *analysis.Symbol {
	nameNode := node.ChildByFieldName("name")
	if nameNode == nil {
		return nil
	}
	name := nodeText(nameNode, source)
	kind := "function"
	if className != "" {
		kind = "method"
	}
	return &analysis.Symbol{
		Name:       name,
		Kind:       kind,
		Path:       filePath,
		Language:   "ruby",
		Visibility: visibility,
		Source:     nodeText(node, source),
		LineNumber: int(node.StartPoint().Row) + 1,
		LineCount:  lineCount(node),
		ParentName: className,
	}
}

func extractRbSingletonMethod(node *sitter.Node, source []byte, filePath, className string) *analysis.Symbol {
	nameNode := node.ChildByFieldName("name")
	if nameNode == nil {
		return nil
	}
	name := nodeText(nameNode, source)
	return &analysis.Symbol{
		Name:       name,
		Kind:       "method",
		Path:       filePath,
		Language:   "ruby",
		Visibility: "public",
		Source:     nodeText(node, source),
		LineNumber: int(node.StartPoint().Row) + 1,
		LineCount:  lineCount(node),
		ParentName: className,
	}
}

func extractRbClass(node *sitter.Node, source []byte, filePath string) (*analysis.Symbol, string) {
	nameNode := node.ChildByFieldName("name")
	if nameNode == nil {
		return nil, ""
	}
	name := nodeText(nameNode, source)
	sym := &analysis.Symbol{
		Name:       name,
		Kind:       "class",
		Path:       filePath,
		Language:   "ruby",
		Visibility: "public",
		Source:     nodeText(node, source),
		LineNumber: int(node.StartPoint().Row) + 1,
		LineCount:  lineCount(node),
	}
	return sym, name
}

func extractRbModule(node *sitter.Node, source []byte, filePath string) *analysis.Symbol {
	nameNode := node.ChildByFieldName("name")
	if nameNode == nil {
		return nil
	}
	name := nodeText(nameNode, source)
	return &analysis.Symbol{
		Name:       name,
		Kind:       "module",
		Path:       filePath,
		Language:   "ruby",
		Visibility: "public",
		Source:     nodeText(node, source),
		LineNumber: int(node.StartPoint().Row) + 1,
		LineCount:  lineCount(node),
	}
}

func collectRbRefs(node *sitter.Node, source []byte, filePath string, refs *[]analysis.Reference) {
	if node.Type() == "call" {
		extractRbCallRefs(node, source, filePath, refs)
		return
	}

	// Ruby bare method calls (no parens) appear as identifiers in body_statement
	if node.Type() == "identifier" && node.Parent() != nil && node.Parent().Type() == "body_statement" {
		name := nodeText(node, source)
		if !rbBuiltins[name] && !rbKeywords[name] {
			*refs = append(*refs, analysis.Reference{
				ToName:   name,
				Kind:     "calls",
				FilePath: filePath,
			})
		}
	}

	for i := 0; i < int(node.NamedChildCount()); i++ {
		collectRbRefs(node.NamedChild(i), source, filePath, refs)
	}
}

func extractRbCallRefs(node *sitter.Node, source []byte, filePath string, refs *[]analysis.Reference) {
	// Ruby call nodes: first named child is typically the method identifier or receiver
	// For `require 'json'`: identifier("require") + argument_list("'json'")
	// For `obj.method(args)`: receiver + "." + identifier + argument_list
	var methodName string
	var args *sitter.Node

	for i := 0; i < int(node.NamedChildCount()); i++ {
		child := node.NamedChild(i)
		switch child.Type() {
		case "identifier":
			methodName = nodeText(child, source)
		case "argument_list":
			args = child
		}
	}

	if methodName == "" {
		// Recurse into children
		for i := 0; i < int(node.NamedChildCount()); i++ {
			collectRbRefs(node.NamedChild(i), source, filePath, refs)
		}
		return
	}

	// Handle require/require_relative
	if methodName == "require" || methodName == "require_relative" {
		if args != nil {
			for i := 0; i < int(args.NamedChildCount()); i++ {
				arg := args.NamedChild(i)
				if arg.Type() == "string" {
					moduleName := extractStringContent(arg, source)
					isExternal := methodName == "require"
					ref := analysis.Reference{
						ToName:     moduleName,
						Kind:       "imports",
						FilePath:   filePath,
						IsExternal: isExternal,
					}
					if isExternal {
						ref.ExternalImportPath = moduleName
					}
					*refs = append(*refs, ref)
					return
				}
			}
		}
		return
	}

	if !rbBuiltins[methodName] {
		*refs = append(*refs, analysis.Reference{
			ToName:   methodName,
			Kind:     "calls",
			FilePath: filePath,
		})
	}

	// Recurse into argument list
	if args != nil {
		for i := 0; i < int(args.NamedChildCount()); i++ {
			collectRbRefs(args.NamedChild(i), source, filePath, refs)
		}
	}
}

// extractStringContent gets the text content from a Ruby string node.
func extractStringContent(node *sitter.Node, source []byte) string {
	for i := 0; i < int(node.NamedChildCount()); i++ {
		child := node.NamedChild(i)
		if child.Type() == "string_content" {
			return nodeText(child, source)
		}
	}
	// Fallback: strip delimiters
	raw := nodeText(node, source)
	return strings.Trim(raw, "'\"")
}

// rbKeywords are Ruby keywords that appear as identifiers but aren't method calls.
var rbKeywords = map[string]bool{
	"private": true, "public": true, "protected": true,
	"true": true, "false": true, "nil": true,
	"self": true, "end": true, "begin": true,
}

var rbBuiltins = map[string]bool{
	"puts": true, "print": true, "p": true, "pp": true,
	"raise": true, "attr_reader": true, "attr_writer": true, "attr_accessor": true,
	"include": true, "extend": true, "prepend": true,
	"new": true, "initialize": true, "super": true,
	"to_s": true, "to_i": true, "to_f": true, "to_a": true, "to_h": true,
	"each": true, "map": true, "select": true, "reject": true, "reduce": true,
	"find": true, "any?": true, "all?": true, "none?": true,
	"push": true, "pop": true, "shift": true, "unshift": true,
	"freeze": true, "dup": true, "clone": true,
}

func nodeText(node *sitter.Node, source []byte) string {
	return analysis.NodeText(node, source)
}

func lineCount(node *sitter.Node) int {
	return analysis.LineCount(node)
}

// RubyComplexityExtractor computes cyclomatic complexity for Ruby
// function AST subtrees.
type RubyComplexityExtractor struct{}

// NewRubyComplexityExtractor creates a new Ruby complexity extractor.
func NewRubyComplexityExtractor() *RubyComplexityExtractor {
	return &RubyComplexityExtractor{}
}

// ComputeComplexity walks the function's AST subtree and counts decision points.
// Returns base complexity of 1 + count of decision nodes.
// Rules: if, unless, while, until, for, when, rescue, &&, ||.
func (e *RubyComplexityExtractor) ComputeComplexity(node *sitter.Node, source []byte) int {
	if node == nil {
		return 1
	}
	complexity := 1
	walkRbNode(node, source, &complexity)
	return complexity
}

func walkRbNode(node *sitter.Node, source []byte, complexity *int) {
	// Only count named nodes to avoid counting unnamed keyword tokens
	// (e.g., the `if` keyword child inside an `if` statement node).
	if node.IsNamed() {
		switch node.Type() {
		case "if", "unless", "while", "until", "for", "when", "rescue":
			*complexity++
		case "binary":
			opNode := node.ChildByFieldName("operator")
			if opNode != nil {
				op := nodeText(opNode, source)
				if op == "&&" || op == "||" {
					*complexity++
				}
			}
		}
	}

	for i := 0; i < int(node.ChildCount()); i++ {
		child := node.Child(i)
		if child != nil {
			walkRbNode(child, source, complexity)
		}
	}
}
