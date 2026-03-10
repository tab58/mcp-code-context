package javascript

import (
	"path/filepath"
	"strings"

	sitter "github.com/smacker/go-tree-sitter"
	"github.com/tab58/code-context/internal/analysis"
)

// Register registers the JavaScript language extractor and complexity extractor
// with the given registry.
func Register(r *analysis.Registry) {
	r.RegisterExtractor("javascript", NewJSExtractor())
	r.RegisterComplexityExtractor("javascript", NewJSComplexityExtractor())
}

// JSExtractor implements the Extractor interface for JavaScript source files.
type JSExtractor struct{}

// NewJSExtractor creates a new JavaScript language extractor.
func NewJSExtractor() *JSExtractor {
	return &JSExtractor{}
}

// ExtractSymbols walks the JavaScript AST and returns all symbols found in the file.
func (e *JSExtractor) ExtractSymbols(tree *sitter.Tree, source []byte, filePath string, repoPath string) ([]analysis.Symbol, error) {
	root := tree.RootNode()
	var symbols []analysis.Symbol
	collectJSSymbols(root, source, filePath, false, "", &symbols)

	// Emit one Module symbol per file
	base := filepath.Base(filePath)
	ext := filepath.Ext(base)
	name := strings.TrimSuffix(base, ext)

	importPath := filePath
	if rel, err := filepath.Rel(repoPath, filePath); err == nil {
		importPath = filepath.ToSlash(rel)
	}

	symbols = append(symbols, analysis.Symbol{
		Name:       name,
		Kind:       "module",
		Path:       filePath,
		Language:   "javascript",
		Visibility: "public",
		ImportPath: importPath,
		ModuleKind: detectModuleKind(source),
		LineNumber: 1,
		LineCount:  int(root.EndPoint().Row) + 1,
	})

	return symbols, nil
}

// ExtractReferences walks the JavaScript AST and returns all cross-symbol references.
func (e *JSExtractor) ExtractReferences(tree *sitter.Tree, source []byte, filePath string, repoPath string) ([]analysis.Reference, error) {
	root := tree.RootNode()
	var refs []analysis.Reference
	aliases := buildImportAliasMap(root, source)
	collectJSRefs(root, source, filePath, aliases, &refs)
	return refs, nil
}

// detectModuleKind returns "esm" if the source contains import or export
// keywords, "cjs" otherwise.
func detectModuleKind(source []byte) string {
	s := string(source)
	if strings.Contains(s, "import ") || strings.Contains(s, "export ") {
		return "esm"
	}
	return "cjs"
}

// buildImportAliasMap collects "import { X as Y }" mappings.
func buildImportAliasMap(root *sitter.Node, source []byte) map[string]string {
	aliases := make(map[string]string)
	for i := 0; i < int(root.ChildCount()); i++ {
		child := root.Child(i)
		if child.Type() == "import_statement" {
			collectImportAliases(child, source, aliases)
		}
	}
	return aliases
}

func collectImportAliases(node *sitter.Node, source []byte, aliases map[string]string) {
	if node.Type() == "import_specifier" {
		nameNode := node.ChildByFieldName("name")
		aliasNode := node.ChildByFieldName("alias")
		if nameNode != nil && aliasNode != nil {
			original := nodeText(nameNode, source)
			alias := nodeText(aliasNode, source)
			if original != alias {
				aliases[alias] = original
			}
		}
		return
	}
	for i := 0; i < int(node.ChildCount()); i++ {
		collectImportAliases(node.Child(i), source, aliases)
	}
}

func collectJSSymbols(node *sitter.Node, source []byte, filePath string, exported bool, className string, symbols *[]analysis.Symbol) {
	switch node.Type() {
	case "export_statement":
		for i := 0; i < int(node.ChildCount()); i++ {
			collectJSSymbols(node.Child(i), source, filePath, true, className, symbols)
		}
		return

	case "function_declaration":
		sym := extractJSFunction(node, source, filePath, exported)
		if sym != nil {
			*symbols = append(*symbols, *sym)
		}

	case "class_declaration":
		sym, name := extractJSClass(node, source, filePath, exported)
		if sym != nil {
			*symbols = append(*symbols, *sym)
		}
		body := findChildByType(node, "class_body")
		if body != nil {
			for i := 0; i < int(body.ChildCount()); i++ {
				collectJSSymbols(body.Child(i), source, filePath, false, name, symbols)
			}
		}
		return

	case "method_definition":
		sym := extractJSMethod(node, source, filePath, className)
		if sym != nil {
			*symbols = append(*symbols, *sym)
		}
		return
	}

	for i := 0; i < int(node.ChildCount()); i++ {
		collectJSSymbols(node.Child(i), source, filePath, false, className, symbols)
	}
}

func extractJSFunction(node *sitter.Node, source []byte, filePath string, exported bool) *analysis.Symbol {
	nameNode := node.ChildByFieldName("name")
	if nameNode == nil {
		return nil
	}
	return &analysis.Symbol{
		Name:       nodeText(nameNode, source),
		Kind:       "function",
		Path:       filePath,
		Language:   "javascript",
		Visibility: jsVisibility(exported),
		Source:     nodeText(node, source),
		LineNumber: int(node.StartPoint().Row) + 1,
		LineCount:  lineCount(node),
	}
}

func extractJSClass(node *sitter.Node, source []byte, filePath string, exported bool) (*analysis.Symbol, string) {
	nameNode := findChildByType(node, "identifier")
	if nameNode == nil {
		return nil, ""
	}
	name := nodeText(nameNode, source)
	sym := &analysis.Symbol{
		Name:       name,
		Kind:       "class",
		Path:       filePath,
		Language:   "javascript",
		Visibility: jsVisibility(exported),
		Source:     nodeText(node, source),
		LineNumber: int(node.StartPoint().Row) + 1,
		LineCount:  lineCount(node),
	}
	return sym, name
}

func extractJSMethod(node *sitter.Node, source []byte, filePath, className string) *analysis.Symbol {
	nameNode := node.ChildByFieldName("name")
	if nameNode == nil {
		return nil
	}
	return &analysis.Symbol{
		Name:       nodeText(nameNode, source),
		Kind:       "method",
		Path:       filePath,
		Language:   "javascript",
		Source:     nodeText(node, source),
		LineNumber: int(node.StartPoint().Row) + 1,
		LineCount:  lineCount(node),
		ParentName: className,
	}
}

func collectJSRefs(node *sitter.Node, source []byte, filePath string, aliases map[string]string, refs *[]analysis.Reference) {
	switch node.Type() {
	case "import_statement":
		ref := extractJSImportRef(node, source, filePath)
		if ref != nil {
			*refs = append(*refs, *ref)
		}
	case "call_expression":
		ref := extractJSCallRef(node, source, filePath)
		if ref != nil {
			if original, ok := aliases[ref.ToName]; ok {
				ref.ToName = original
			}
			*refs = append(*refs, *ref)
		}
	}

	for i := 0; i < int(node.ChildCount()); i++ {
		collectJSRefs(node.Child(i), source, filePath, aliases, refs)
	}
}

func extractJSImportRef(node *sitter.Node, source []byte, filePath string) *analysis.Reference {
	// Find import specifiers for named imports
	var importedName string
	for i := 0; i < int(node.ChildCount()); i++ {
		child := node.Child(i)
		if child.Type() == "import_clause" {
			// Check for default import (identifier child) or named imports
			for j := 0; j < int(child.ChildCount()); j++ {
				sub := child.Child(j)
				if sub.Type() == "identifier" {
					importedName = nodeText(sub, source)
				} else if sub.Type() == "named_imports" {
					for k := 0; k < int(sub.ChildCount()); k++ {
						spec := sub.Child(k)
						if spec.Type() == "import_specifier" {
							nameNode := spec.ChildByFieldName("name")
							if nameNode != nil {
								importedName = nodeText(nameNode, source)
							}
						}
					}
				}
			}
		}
	}

	// Find the module source string
	var moduleSource string
	for i := 0; i < int(node.ChildCount()); i++ {
		child := node.Child(i)
		if child.Type() == "string" {
			raw := nodeText(child, source)
			moduleSource = strings.Trim(raw, "'\"")
			break
		}
	}
	if moduleSource == "" {
		return nil
	}

	toName := importedName
	if toName == "" {
		toName = moduleSource
	}

	if !isRelativeImport(moduleSource) {
		return &analysis.Reference{
			ToName:             toName,
			Kind:               "imports",
			FilePath:           filePath,
			IsExternal:         true,
			ExternalImportPath: moduleSource,
		}
	}

	return &analysis.Reference{
		ToName:   toName,
		Kind:     "imports",
		FilePath: filePath,
	}
}

func extractJSCallRef(node *sitter.Node, source []byte, filePath string) *analysis.Reference {
	fnNode := node.ChildByFieldName("function")
	if fnNode == nil {
		return nil
	}

	// Skip anonymous function calls
	if fnNode.Type() == "function" || fnNode.Type() == "arrow_function" || fnNode.Type() == "parenthesized_expression" {
		return nil
	}

	// Normalize member_expression to property name
	if fnNode.Type() == "member_expression" {
		propNode := fnNode.ChildByFieldName("property")
		if propNode == nil {
			return nil
		}
		name := nodeText(propNode, source)
		if jsBuiltins[name] {
			return nil
		}
		return &analysis.Reference{
			ToName:   name,
			Kind:     "calls",
			FilePath: filePath,
		}
	}

	name := nodeText(fnNode, source)
	if strings.Contains(name, " ") || jsBuiltins[name] {
		return nil
	}
	return &analysis.Reference{
		ToName:   name,
		Kind:     "calls",
		FilePath: filePath,
	}
}

func isRelativeImport(modulePath string) bool {
	return strings.HasPrefix(modulePath, ".") || strings.HasPrefix(modulePath, "/")
}

func jsVisibility(exported bool) string {
	if exported {
		return "public"
	}
	return "private"
}

func findChildByType(node *sitter.Node, typeName string) *sitter.Node {
	return analysis.FindChildByType(node, typeName)
}

func nodeText(node *sitter.Node, source []byte) string {
	return analysis.NodeText(node, source)
}

func lineCount(node *sitter.Node) int {
	return analysis.LineCount(node)
}

// jsBuiltins is the set of JavaScript built-in functions to filter from calls.
var jsBuiltins = map[string]bool{
	"push": true, "pop": true, "shift": true, "unshift": true,
	"slice": true, "concat": true, "join": true, "reverse": true, "sort": true,
	"filter": true, "map": true, "reduce": true, "forEach": true, "find": true,
	"log": true, "warn": true, "error": true, "info": true,
	"setTimeout": true, "setInterval": true, "clearTimeout": true, "clearInterval": true,
	"parseInt": true, "parseFloat": true, "isNaN": true,
	"then": true, "catch": true, "finally": true, "resolve": true, "reject": true,
	"require": true, "console": true,
	"stringify": true, "parse": true,
	"Number": true, "String": true, "Boolean": true, "Array": true, "Object": true,
}

// JSComplexityExtractor computes cyclomatic complexity for JavaScript
// function AST subtrees.
type JSComplexityExtractor struct{}

// NewJSComplexityExtractor creates a new JavaScript complexity extractor.
func NewJSComplexityExtractor() *JSComplexityExtractor {
	return &JSComplexityExtractor{}
}

// ComputeComplexity walks the function's AST subtree and counts decision points.
// Returns base complexity of 1 + count of decision nodes.
func (e *JSComplexityExtractor) ComputeComplexity(node *sitter.Node, source []byte) int {
	if node == nil {
		return 1
	}
	complexity := 1
	walkJSNode(node, source, &complexity)
	return complexity
}

func walkJSNode(node *sitter.Node, source []byte, complexity *int) {
	switch node.Type() {
	case "if_statement", "for_statement", "for_in_statement", "while_statement", "do_statement",
		"switch_case", "catch_clause", "ternary_expression":
		*complexity++
	case "binary_expression":
		opNode := node.ChildByFieldName("operator")
		if opNode != nil {
			op := nodeText(opNode, source)
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
			walkJSNode(child, source, complexity)
		}
	}
}
