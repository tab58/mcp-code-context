package python

import (
	"path/filepath"
	"strings"

	sitter "github.com/smacker/go-tree-sitter"
	"github.com/tab58/code-context/internal/analysis"
)

// Register registers the Python language extractor and complexity extractor
// with the given registry.
func Register(r *analysis.Registry) {
	r.RegisterExtractor("python", NewPythonExtractor())
	r.RegisterComplexityExtractor("python", NewPythonComplexityExtractor())
}

// PythonExtractor implements the Extractor interface for Python source files.
type PythonExtractor struct{}

// NewPythonExtractor creates a new Python language extractor.
func NewPythonExtractor() *PythonExtractor {
	return &PythonExtractor{}
}

// ExtractSymbols walks the Python AST and returns all symbols found in the file.
func (e *PythonExtractor) ExtractSymbols(tree *sitter.Tree, source []byte, filePath string, repoPath string) ([]analysis.Symbol, error) {
	root := tree.RootNode()
	var symbols []analysis.Symbol
	collectPySymbols(root, source, filePath, "", &symbols)

	// Emit one Module symbol per file with dotted import path
	base := filepath.Base(filePath)
	ext := filepath.Ext(base)
	name := strings.TrimSuffix(base, ext)

	importPath := name
	if rel, err := filepath.Rel(repoPath, filePath); err == nil {
		// Convert to dotted module path: src/app.py -> src.app
		rel = filepath.ToSlash(rel)
		rel = strings.TrimSuffix(rel, ext)
		importPath = strings.ReplaceAll(rel, "/", ".")
	}

	symbols = append(symbols, analysis.Symbol{
		Name:       name,
		Kind:       "module",
		Path:       filePath,
		Language:   "python",
		Visibility: "public",
		ImportPath: importPath,
		LineNumber: 1,
		LineCount:  int(root.EndPoint().Row) + 1,
	})

	return symbols, nil
}

// ExtractReferences walks the Python AST and returns all cross-symbol references.
func (e *PythonExtractor) ExtractReferences(tree *sitter.Tree, source []byte, filePath string, repoPath string) ([]analysis.Reference, error) {
	root := tree.RootNode()
	var refs []analysis.Reference
	collectPyRefs(root, source, filePath, &refs)
	return refs, nil
}

func collectPySymbols(node *sitter.Node, source []byte, filePath, className string, symbols *[]analysis.Symbol) {
	switch node.Type() {
	case "function_definition":
		sym := extractPyFunction(node, source, filePath, className)
		if sym != nil {
			*symbols = append(*symbols, *sym)
		}
		return // don't recurse into function body for nested defs

	case "class_definition":
		sym, name := extractPyClass(node, source, filePath)
		if sym != nil {
			*symbols = append(*symbols, *sym)
		}
		// Extract methods inside class body
		body := node.ChildByFieldName("body")
		if body != nil {
			for i := 0; i < int(body.NamedChildCount()); i++ {
				collectPySymbols(body.NamedChild(i), source, filePath, name, symbols)
			}
		}
		return

	case "decorated_definition":
		// Extract decorators and recurse into the decorated node
		var decorators []string
		for i := 0; i < int(node.NamedChildCount()); i++ {
			child := node.NamedChild(i)
			if child.Type() == "decorator" {
				decorators = append(decorators, nodeText(child, source))
			}
		}
		// Find the actual definition inside
		for i := 0; i < int(node.NamedChildCount()); i++ {
			child := node.NamedChild(i)
			if child.Type() == "function_definition" {
				sym := extractPyFunction(child, source, filePath, className)
				if sym != nil {
					sym.Decorators = decorators
					*symbols = append(*symbols, *sym)
				}
				return
			}
			if child.Type() == "class_definition" {
				sym, name := extractPyClass(child, source, filePath)
				if sym != nil {
					sym.Decorators = decorators
					*symbols = append(*symbols, *sym)
				}
				body := child.ChildByFieldName("body")
				if body != nil {
					for j := 0; j < int(body.NamedChildCount()); j++ {
						collectPySymbols(body.NamedChild(j), source, filePath, name, symbols)
					}
				}
				return
			}
		}
		return
	}

	for i := 0; i < int(node.NamedChildCount()); i++ {
		collectPySymbols(node.NamedChild(i), source, filePath, className, symbols)
	}
}

func extractPyFunction(node *sitter.Node, source []byte, filePath, className string) *analysis.Symbol {
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
		Language:   "python",
		Visibility: pyVisibility(name),
		Source:     nodeText(node, source),
		LineNumber: int(node.StartPoint().Row) + 1,
		LineCount:  lineCount(node),
		ParentName: className,
	}
}

func extractPyClass(node *sitter.Node, source []byte, filePath string) (*analysis.Symbol, string) {
	nameNode := node.ChildByFieldName("name")
	if nameNode == nil {
		return nil, ""
	}
	name := nodeText(nameNode, source)
	sym := &analysis.Symbol{
		Name:       name,
		Kind:       "class",
		Path:       filePath,
		Language:   "python",
		Visibility: pyVisibility(name),
		Source:     nodeText(node, source),
		LineNumber: int(node.StartPoint().Row) + 1,
		LineCount:  lineCount(node),
	}
	return sym, name
}

// pyVisibility returns "private" if name starts with underscore, "public" otherwise.
func pyVisibility(name string) string {
	if strings.HasPrefix(name, "_") {
		return "private"
	}
	return "public"
}

func collectPyRefs(node *sitter.Node, source []byte, filePath string, refs *[]analysis.Reference) {
	switch node.Type() {
	case "import_statement":
		// import os, import json
		for i := 0; i < int(node.NamedChildCount()); i++ {
			child := node.NamedChild(i)
			if child.Type() == "dotted_name" {
				name := nodeText(child, source)
				*refs = append(*refs, analysis.Reference{
					ToName:             name,
					Kind:               "imports",
					FilePath:           filePath,
					IsExternal:         true,
					ExternalImportPath: name,
				})
			}
		}

	case "import_from_statement":
		// from pathlib import Path
		var moduleName string
		for i := 0; i < int(node.NamedChildCount()); i++ {
			child := node.NamedChild(i)
			if child.Type() == "dotted_name" || child.Type() == "relative_import" {
				moduleName = nodeText(child, source)
				break
			}
		}
		isExternal := moduleName != "" && !strings.HasPrefix(moduleName, ".")
		for i := 0; i < int(node.NamedChildCount()); i++ {
			child := node.NamedChild(i)
			if child.Type() == "dotted_name" && nodeText(child, source) != moduleName {
				name := nodeText(child, source)
				ref := analysis.Reference{
					ToName:     name,
					Kind:       "imports",
					FilePath:   filePath,
					IsExternal: isExternal,
				}
				if isExternal {
					ref.ExternalImportPath = moduleName
				}
				*refs = append(*refs, ref)
			}
		}

	case "call":
		ref := extractPyCallRef(node, source, filePath)
		if ref != nil {
			*refs = append(*refs, *ref)
		}
	}

	for i := 0; i < int(node.NamedChildCount()); i++ {
		collectPyRefs(node.NamedChild(i), source, filePath, refs)
	}
}

func extractPyCallRef(node *sitter.Node, source []byte, filePath string) *analysis.Reference {
	fnNode := node.ChildByFieldName("function")
	if fnNode == nil {
		return nil
	}

	// Normalize attribute access to just the method name
	if fnNode.Type() == "attribute" {
		attrNode := fnNode.ChildByFieldName("attribute")
		if attrNode == nil {
			return nil
		}
		name := nodeText(attrNode, source)
		if pyBuiltins[name] {
			return nil
		}
		return &analysis.Reference{
			ToName:   name,
			Kind:     "calls",
			FilePath: filePath,
		}
	}

	name := nodeText(fnNode, source)
	if pyBuiltins[name] {
		return nil
	}
	return &analysis.Reference{
		ToName:   name,
		Kind:     "calls",
		FilePath: filePath,
	}
}

var pyBuiltins = map[string]bool{
	"print": true, "len": true, "range": true, "type": true, "isinstance": true,
	"str": true, "int": true, "float": true, "bool": true, "list": true,
	"dict": true, "set": true, "tuple": true, "super": true, "property": true,
	"staticmethod": true, "classmethod": true, "enumerate": true, "zip": true,
	"map": true, "filter": true, "sorted": true, "reversed": true,
	"open": true, "input": true, "abs": true, "min": true, "max": true,
	"sum": true, "round": true, "hasattr": true, "getattr": true, "setattr": true,
	"append": true, "extend": true, "insert": true, "remove": true, "pop": true,
	"keys": true, "values": true, "items": true, "get": true, "update": true,
	"format": true, "join": true, "split": true, "strip": true,
}

func nodeText(node *sitter.Node, source []byte) string {
	return analysis.NodeText(node, source)
}

func lineCount(node *sitter.Node) int {
	return analysis.LineCount(node)
}

// PythonComplexityExtractor computes cyclomatic complexity for Python
// function AST subtrees.
type PythonComplexityExtractor struct{}

// NewPythonComplexityExtractor creates a new Python complexity extractor.
func NewPythonComplexityExtractor() *PythonComplexityExtractor {
	return &PythonComplexityExtractor{}
}

// ComputeComplexity walks the function's AST subtree and counts decision points.
// Returns base complexity of 1 + count of decision nodes.
// Rules: if, for, while, except, and, or, elif, comprehension if, assert.
func (e *PythonComplexityExtractor) ComputeComplexity(node *sitter.Node, source []byte) int {
	if node == nil {
		return 1
	}
	complexity := 1
	walkPyNode(node, source, &complexity)
	return complexity
}

func walkPyNode(node *sitter.Node, source []byte, complexity *int) {
	switch node.Type() {
	case "if_statement", "for_statement", "while_statement",
		"except_clause", "elif_clause", "assert_statement":
		*complexity++
	case "boolean_operator":
		opNode := node.ChildByFieldName("operator")
		if opNode != nil {
			op := nodeText(opNode, source)
			if op == "and" || op == "or" {
				*complexity++
			}
		}
	}

	for i := 0; i < int(node.ChildCount()); i++ {
		child := node.Child(i)
		if child != nil {
			walkPyNode(child, source, complexity)
		}
	}
}
