package golang

import (
	"bufio"
	"os"
	"path/filepath"
	"strings"
	"unicode"
	"unicode/utf8"

	sitter "github.com/smacker/go-tree-sitter"
	"github.com/tab58/code-context/internal/analysis"
)

// Register registers the Go language extractor with the given registry.
func Register(r *analysis.Registry) {
	r.RegisterExtractor("go", NewGoExtractor())
}

// GoExtractor implements the Extractor interface for Go source files.
type GoExtractor struct{}

// NewGoExtractor creates a new Go language extractor.
func NewGoExtractor() *GoExtractor {
	return &GoExtractor{}
}

// ExtractSymbols walks the Go AST and returns all symbols found in the file.
func (e *GoExtractor) ExtractSymbols(tree *sitter.Tree, source []byte, filePath string, repoPath string) ([]analysis.Symbol, error) {
	root := tree.RootNode()
	var symbols []analysis.Symbol

	moduleName := resolveGoModuleNameFromFile(filePath, repoPath)

	for i := 0; i < int(root.ChildCount()); i++ {
		child := root.Child(i)
		switch child.Type() {
		case "package_clause":
			sym := extractPackageClause(child, source, filePath, repoPath, moduleName)
			if sym != nil {
				symbols = append(symbols, *sym)
			}
		case "function_declaration":
			sym := extractFunctionDecl(child, source, filePath)
			if sym != nil {
				symbols = append(symbols, *sym)
			}
		case "method_declaration":
			sym := extractMethodDecl(child, source, filePath)
			if sym != nil {
				symbols = append(symbols, *sym)
			}
		case "type_declaration":
			syms := extractTypeDecl(child, source, filePath)
			symbols = append(symbols, syms...)
		}
	}

	return symbols, nil
}

// goBuiltins is the set of Go built-in identifiers that should be filtered
// from call references (they are not user-defined symbols).
var goBuiltins = map[string]bool{
	"append": true, "cap": true, "clear": true, "close": true, "complex": true,
	"copy": true, "delete": true, "imag": true, "len": true, "make": true,
	"max": true, "min": true, "new": true, "panic": true, "print": true,
	"println": true, "real": true, "recover": true,
	// Built-in types
	"bool": true, "byte": true, "comparable": true, "complex64": true, "complex128": true,
	"error": true, "float32": true, "float64": true, "int": true, "int8": true,
	"int16": true, "int32": true, "int64": true, "rune": true, "string": true,
	"uint": true, "uint8": true, "uint16": true, "uint32": true, "uint64": true,
	"uintptr": true, "any": true,
	// Built-in constants
	"true": true, "false": true, "iota": true, "nil": true,
}

// importInfo holds classification data for a single import.
type importInfo struct {
	importPath string
	shortName  string
	isStdlib   bool
	isExternal bool
}

// ExtractReferences walks the Go AST and returns all cross-symbol references.
func (e *GoExtractor) ExtractReferences(tree *sitter.Tree, source []byte, filePath string, repoPath string) ([]analysis.Reference, error) {
	root := tree.RootNode()
	var refs []analysis.Reference

	moduleName := resolveGoModuleNameFromFile(filePath, repoPath)
	imports := buildImportMap(root, source, moduleName)

	collectRefs(root, source, filePath, imports, &refs)
	return refs, nil
}

func extractPackageClause(node *sitter.Node, source []byte, filePath, repoPath, moduleName string) *analysis.Symbol {
	var nameNode *sitter.Node
	for i := 0; i < int(node.ChildCount()); i++ {
		if node.Child(i).Type() == "package_identifier" {
			nameNode = node.Child(i)
			break
		}
	}
	if nameNode == nil {
		return nil
	}
	name := nodeText(nameNode, source)

	importPath := goImportPath(moduleName, filePath, repoPath)
	visibility := goModuleVisibility(importPath)

	return &analysis.Symbol{
		Name:       name,
		Kind:       "module",
		Path:       filePath,
		Language:   "go",
		Source:     nodeText(node, source),
		LineNumber: int(node.StartPoint().Row) + 1,
		LineCount:  lineCount(node),
		ImportPath: importPath,
		Visibility: visibility,
		ModuleKind: "package",
	}
}

// resolveGoModuleNameFromFile finds the nearest go.mod by searching from the
// file's directory upward to repoPath. This handles monorepos where go.mod
// is in a subdirectory, not the repo root.
func resolveGoModuleNameFromFile(filePath, repoPath string) string {
	dir := filepath.Dir(filePath)
	absRepo, err := filepath.Abs(repoPath)
	if err != nil {
		return findGoModuleName(repoPath)
	}

	for {
		if name := findGoModuleName(dir); name != "" {
			return name
		}
		if dir == absRepo || dir == "/" || dir == "." {
			break
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}
	return ""
}

// findGoModuleName reads the go.mod file in the given directory and returns
// the module name. Returns empty string if go.mod doesn't exist or can't be parsed.
func findGoModuleName(dir string) string {
	goModPath := filepath.Join(dir, "go.mod")
	f, err := os.Open(goModPath)
	if err != nil {
		return ""
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if strings.HasPrefix(line, "module ") {
			return strings.TrimSpace(strings.TrimPrefix(line, "module "))
		}
	}
	return ""
}

// goImportPath computes the fully qualified import path for a Go package.
// Returns moduleName + relative directory path, or empty string if no module name.
func goImportPath(moduleName, filePath, repoPath string) string {
	if moduleName == "" {
		return ""
	}
	dir := filepath.Dir(filePath)
	rel, err := filepath.Rel(repoPath, dir)
	if err != nil || rel == "." {
		return moduleName
	}
	return moduleName + "/" + filepath.ToSlash(rel)
}

// goModuleVisibility returns "internal" if the import path contains "/internal/",
// or "public" otherwise.
func goModuleVisibility(importPath string) string {
	if strings.Contains(importPath, "/internal/") || strings.HasSuffix(importPath, "/internal") {
		return "internal"
	}
	return "public"
}

func extractFunctionDecl(node *sitter.Node, source []byte, filePath string) *analysis.Symbol {
	nameNode := node.ChildByFieldName("name")
	if nameNode == nil {
		return nil
	}
	name := nodeText(nameNode, source)
	return &analysis.Symbol{
		Name:       name,
		Kind:       "function",
		Path:       filePath,
		Language:   "go",
		Signature:  extractSignature(node, source),
		Visibility: goVisibility(name),
		Source:     nodeText(node, source),
		LineNumber: int(node.StartPoint().Row) + 1,
		LineCount:  lineCount(node),
	}
}

func extractMethodDecl(node *sitter.Node, source []byte, filePath string) *analysis.Symbol {
	nameNode := node.ChildByFieldName("name")
	if nameNode == nil {
		return nil
	}
	name := nodeText(nameNode, source)

	parentName := ""
	receiver := node.ChildByFieldName("receiver")
	if receiver != nil {
		parentName = extractReceiverType(receiver, source)
	}

	return &analysis.Symbol{
		Name:       name,
		Kind:       "method",
		Path:       filePath,
		Language:   "go",
		Signature:  extractSignature(node, source),
		Visibility: goVisibility(name),
		Source:     nodeText(node, source),
		LineNumber: int(node.StartPoint().Row) + 1,
		LineCount:  lineCount(node),
		ParentName: parentName,
	}
}

func extractTypeDecl(node *sitter.Node, source []byte, filePath string) []analysis.Symbol {
	var symbols []analysis.Symbol
	for i := 0; i < int(node.ChildCount()); i++ {
		child := node.Child(i)
		if child.Type() == "type_spec" {
			sym := extractTypeSpec(child, source, filePath)
			if sym != nil {
				symbols = append(symbols, *sym)
			}
		}
	}
	return symbols
}

func extractTypeSpec(node *sitter.Node, source []byte, filePath string) *analysis.Symbol {
	nameNode := node.ChildByFieldName("name")
	if nameNode == nil {
		return nil
	}
	name := nodeText(nameNode, source)

	typeNode := node.ChildByFieldName("type")
	if typeNode == nil {
		return nil
	}

	kind := ""
	switch typeNode.Type() {
	case "struct_type":
		kind = "struct"
	case "interface_type":
		kind = "interface"
	default:
		return nil
	}

	return &analysis.Symbol{
		Name:       name,
		Kind:       kind,
		Path:       filePath,
		Language:   "go",
		Visibility: goVisibility(name),
		Source:     nodeText(node, source),
		LineNumber: int(node.StartPoint().Row) + 1,
		LineCount:  lineCount(node),
	}
}

// buildImportMap builds a map from short package name to importInfo by
// scanning import_declaration nodes in the AST root.
func buildImportMap(root *sitter.Node, source []byte, moduleName string) map[string]importInfo {
	imports := make(map[string]importInfo)
	for i := 0; i < int(root.ChildCount()); i++ {
		child := root.Child(i)
		if child.Type() == "import_declaration" {
			for j := 0; j < int(child.ChildCount()); j++ {
				spec := child.Child(j)
				if spec.Type() == "import_spec" || spec.Type() == "import_spec_list" {
					addImportSpecs(spec, source, moduleName, imports)
				}
			}
		}
	}
	return imports
}

func addImportSpecs(node *sitter.Node, source []byte, moduleName string, imports map[string]importInfo) {
	if node.Type() == "import_spec" {
		pathNode := node.ChildByFieldName("path")
		if pathNode == nil {
			return
		}
		raw := nodeText(pathNode, source)
		ip := strings.Trim(raw, "\"")
		parts := strings.Split(ip, "/")
		shortName := parts[len(parts)-1]

		// Check for alias
		nameNode := node.ChildByFieldName("name")
		if nameNode != nil {
			shortName = nodeText(nameNode, source)
		}

		stdlib := isStdlib(ip)
		isExt := stdlib || (moduleName != "" && !strings.HasPrefix(ip, moduleName))
		imports[shortName] = importInfo{
			importPath: ip,
			shortName:  shortName,
			isStdlib:   stdlib,
			isExternal: isExt,
		}
		return
	}
	// Recurse for import_spec_list
	for i := 0; i < int(node.ChildCount()); i++ {
		addImportSpecs(node.Child(i), source, moduleName, imports)
	}
}

// isStdlib returns true if the import path looks like a Go standard library
// package (no dots in the first path segment).
func isStdlib(importPath string) bool {
	parts := strings.SplitN(importPath, "/", 2)
	return !strings.Contains(parts[0], ".")
}

func collectRefs(node *sitter.Node, source []byte, filePath string, imports map[string]importInfo, refs *[]analysis.Reference) {
	switch node.Type() {
	case "import_spec":
		ref := extractImportRef(node, source, filePath, imports)
		if ref != nil {
			*refs = append(*refs, *ref)
		}
	case "call_expression":
		ref := extractCallRef(node, source, filePath, imports)
		if ref != nil {
			*refs = append(*refs, *ref)
		}
	}

	for i := 0; i < int(node.ChildCount()); i++ {
		collectRefs(node.Child(i), source, filePath, imports, refs)
	}
}

func extractImportRef(node *sitter.Node, source []byte, filePath string, imports map[string]importInfo) *analysis.Reference {
	pathNode := node.ChildByFieldName("path")
	if pathNode == nil {
		return nil
	}
	raw := nodeText(pathNode, source)
	importPath := strings.Trim(raw, "\"")
	parts := strings.Split(importPath, "/")
	shortName := parts[len(parts)-1]

	// Check for alias
	nameNode := node.ChildByFieldName("name")
	if nameNode != nil {
		shortName = nodeText(nameNode, source)
	}

	info, found := imports[shortName]
	if found && info.isExternal {
		return &analysis.Reference{
			ToName:             shortName,
			Kind:               "imports",
			FilePath:           filePath,
			IsExternal:         true,
			ExternalImportPath: info.importPath,
		}
	}

	return &analysis.Reference{
		ToName:   shortName,
		Kind:     "imports",
		FilePath: filePath,
	}
}

func extractCallRef(node *sitter.Node, source []byte, filePath string, imports map[string]importInfo) *analysis.Reference {
	fnNode := node.ChildByFieldName("function")
	if fnNode == nil {
		return nil
	}

	// Skip func_literal calls (anonymous functions)
	if fnNode.Type() == "func_literal" {
		return nil
	}

	// Handle selector_expression (e.g., pkg.Func or obj.Method)
	if fnNode.Type() == "selector_expression" {
		objNode := fnNode.ChildByFieldName("operand")
		fieldNode := fnNode.ChildByFieldName("field")
		if objNode == nil || fieldNode == nil {
			return nil
		}
		objName := nodeText(objNode, source)
		methodName := nodeText(fieldNode, source)

		// Check if the object is an imported package
		info, found := imports[objName]
		if found && info.isExternal {
			return &analysis.Reference{
				ToName:             methodName,
				Kind:               "calls",
				FilePath:           filePath,
				IsExternal:         true,
				ExternalImportPath: info.importPath,
			}
		}

		// Internal package call — normalize to bare method name
		if found {
			return &analysis.Reference{
				ToName:   methodName,
				Kind:     "calls",
				FilePath: filePath,
			}
		}

		// Not a package call — it's a method call on a variable (e.g., server.Start()).
		// Skip: these are method calls on types, not standalone function references.
		return nil
	}

	name := nodeText(fnNode, source)

	// Filter built-in identifiers
	if goBuiltins[name] {
		return nil
	}

	return &analysis.Reference{
		ToName:   name,
		Kind:     "calls",
		FilePath: filePath,
	}
}

func extractReceiverType(node *sitter.Node, source []byte) string {
	text := nodeText(node, source)
	text = strings.TrimPrefix(text, "(")
	text = strings.TrimSuffix(text, ")")
	parts := strings.Fields(text)
	if len(parts) == 0 {
		return ""
	}
	typeName := parts[len(parts)-1]
	typeName = strings.TrimPrefix(typeName, "*")
	return typeName
}

func extractSignature(node *sitter.Node, source []byte) string {
	text := nodeText(node, source)
	if idx := strings.Index(text, "{"); idx > 0 {
		return strings.TrimSpace(text[:idx])
	}
	return text
}

func goVisibility(name string) string {
	r, _ := utf8.DecodeRuneInString(name)
	if unicode.IsUpper(r) {
		return "public"
	}
	return "package"
}

func nodeText(node *sitter.Node, source []byte) string {
	return string(source[node.StartByte():node.EndByte()])
}

func lineCount(node *sitter.Node) int {
	return int(node.EndPoint().Row-node.StartPoint().Row) + 1
}
