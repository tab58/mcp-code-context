package typescript

import (
	"path/filepath"
	"strings"

	sitter "github.com/smacker/go-tree-sitter"
	"github.com/tab58/code-context/internal/analysis"
)

// Register registers the TypeScript and TSX language extractors
// with the given registry.
func Register(r *analysis.Registry) {
	r.RegisterExtractor("typescript", NewTypeScriptExtractor())
	r.RegisterExtractor("tsx", NewTSXExtractor())
}

// TypeScriptExtractor implements the Extractor interface for TypeScript source files.
type TypeScriptExtractor struct{}

// NewTypeScriptExtractor creates a new TypeScript language extractor.
func NewTypeScriptExtractor() *TypeScriptExtractor {
	return &TypeScriptExtractor{}
}

// ExtractSymbols walks the TypeScript AST and returns all symbols found in the file.
func (e *TypeScriptExtractor) ExtractSymbols(tree *sitter.Tree, source []byte, filePath string, repoPath string) ([]analysis.Symbol, error) {
	return extractTSSymbols(tree, source, filePath, repoPath, "typescript")
}

// ExtractReferences walks the TypeScript AST and returns all cross-symbol references.
func (e *TypeScriptExtractor) ExtractReferences(tree *sitter.Tree, source []byte, filePath string, repoPath string) ([]analysis.Reference, error) {
	return extractTSReferences(tree, source, filePath)
}

// extractTSSymbols is the shared implementation for TypeScript and TSX.
func extractTSSymbols(tree *sitter.Tree, source []byte, filePath, repoPath, lang string) ([]analysis.Symbol, error) {
	root := tree.RootNode()
	var symbols []analysis.Symbol
	collectSymbols(root, source, filePath, lang, false, "", &symbols)

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
		Language:   lang,
		Visibility: "public",
		ImportPath: importPath,
		ModuleKind: detectModuleKind(source),
		LineNumber: 1,
		LineCount:  int(root.EndPoint().Row) + 1,
	})

	return symbols, nil
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

// extractTSReferences is the shared implementation for TypeScript and TSX.
func extractTSReferences(tree *sitter.Tree, source []byte, filePath string) ([]analysis.Reference, error) {
	root := tree.RootNode()
	var refs []analysis.Reference
	// Build import alias map: alias -> original name
	aliases := buildImportAliasMap(root, source)
	collectRefs(root, source, filePath, aliases, &refs)
	return refs, nil
}

// buildImportAliasMap collects "import { X as Y }" mappings, returning alias → original.
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

// collectImportAliases recursively finds import_specifier nodes with aliases.
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

func collectSymbols(node *sitter.Node, source []byte, filePath, lang string, exported bool, className string, symbols *[]analysis.Symbol) {
	switch node.Type() {
	case "export_statement":
		for i := 0; i < int(node.ChildCount()); i++ {
			child := node.Child(i)
			collectSymbols(child, source, filePath, lang, true, className, symbols)
		}
		return

	case "function_declaration":
		sym := extractFunction(node, source, filePath, lang, exported)
		if sym != nil {
			*symbols = append(*symbols, *sym)
		}

	case "class_declaration":
		sym, name := extractClass(node, source, filePath, lang, exported)
		if sym != nil {
			*symbols = append(*symbols, *sym)
		}
		// Extract methods inside class body
		body := findChildByType(node, "class_body")
		if body != nil {
			for i := 0; i < int(body.ChildCount()); i++ {
				collectSymbols(body.Child(i), source, filePath, lang, false, name, symbols)
			}
		}
		return

	case "interface_declaration":
		sym := extractInterface(node, source, filePath, lang, exported)
		if sym != nil {
			*symbols = append(*symbols, *sym)
		}

	case "enum_declaration":
		sym := extractEnum(node, source, filePath, lang, exported)
		if sym != nil {
			*symbols = append(*symbols, *sym)
		}

	case "method_definition":
		sym := extractMethod(node, source, filePath, lang, className)
		if sym != nil {
			*symbols = append(*symbols, *sym)
		}
		return
	}

	// Recurse into children (except cases that handle their own children)
	for i := 0; i < int(node.ChildCount()); i++ {
		collectSymbols(node.Child(i), source, filePath, lang, false, className, symbols)
	}
}

// tsVisibility returns "public" if exported, "private" otherwise.
func tsVisibility(exported bool) string {
	if exported {
		return "public"
	}
	return "private"
}

func extractFunction(node *sitter.Node, source []byte, filePath, lang string, exported bool) *analysis.Symbol {
	nameNode := node.ChildByFieldName("name")
	if nameNode == nil {
		return nil
	}
	name := nodeText(nameNode, source)
	return &analysis.Symbol{
		Name:       name,
		Kind:       "function",
		Path:       filePath,
		Language:   lang,
		Visibility: tsVisibility(exported),
		Source:     nodeText(node, source),
		LineNumber: int(node.StartPoint().Row) + 1,
		LineCount:  lineCount(node),
	}
}

func extractClass(node *sitter.Node, source []byte, filePath, lang string, exported bool) (*analysis.Symbol, string) {
	nameNode := findChildByType(node, "type_identifier")
	if nameNode == nil {
		return nil, ""
	}
	name := nodeText(nameNode, source)

	var decorators []string
	for i := 0; i < int(node.ChildCount()); i++ {
		child := node.Child(i)
		if child.Type() == "decorator" {
			decorators = append(decorators, nodeText(child, source))
		}
	}

	sym := &analysis.Symbol{
		Name:       name,
		Kind:       "class",
		Path:       filePath,
		Language:   lang,
		Visibility: tsVisibility(exported),
		Source:     nodeText(node, source),
		LineNumber: int(node.StartPoint().Row) + 1,
		LineCount:  lineCount(node),
		Decorators: decorators,
	}
	return sym, name
}

func extractInterface(node *sitter.Node, source []byte, filePath, lang string, exported bool) *analysis.Symbol {
	nameNode := findChildByType(node, "type_identifier")
	if nameNode == nil {
		return nil
	}
	name := nodeText(nameNode, source)
	return &analysis.Symbol{
		Name:       name,
		Kind:       "interface",
		Path:       filePath,
		Language:   lang,
		Visibility: tsVisibility(exported),
		Source:     nodeText(node, source),
		LineNumber: int(node.StartPoint().Row) + 1,
		LineCount:  lineCount(node),
	}
}

func extractEnum(node *sitter.Node, source []byte, filePath, lang string, exported bool) *analysis.Symbol {
	nameNode := findChildByType(node, "identifier")
	if nameNode == nil {
		return nil
	}
	name := nodeText(nameNode, source)
	return &analysis.Symbol{
		Name:       name,
		Kind:       "class",
		Path:       filePath,
		Language:   lang,
		Visibility: tsVisibility(exported),
		Source:     nodeText(node, source),
		LineNumber: int(node.StartPoint().Row) + 1,
		LineCount:  lineCount(node),
	}
}

func extractMethod(node *sitter.Node, source []byte, filePath, lang, className string) *analysis.Symbol {
	nameNode := node.ChildByFieldName("name")
	if nameNode == nil {
		return nil
	}
	name := nodeText(nameNode, source)
	return &analysis.Symbol{
		Name:       name,
		Kind:       "method",
		Path:       filePath,
		Language:   lang,
		Source:     nodeText(node, source),
		LineNumber: int(node.StartPoint().Row) + 1,
		LineCount:  lineCount(node),
		ParentName: className,
	}
}

func collectRefs(node *sitter.Node, source []byte, filePath string, aliases map[string]string, refs *[]analysis.Reference) {
	switch node.Type() {
	case "import_statement":
		ref := extractImportRef(node, source, filePath)
		if ref != nil {
			*refs = append(*refs, *ref)
		}
	case "call_expression":
		ref := extractCallRef(node, source, filePath)
		if ref != nil {
			// Resolve import aliases: if a call uses an aliased name,
			// replace it with the original name for symbol table lookup.
			if original, ok := aliases[ref.ToName]; ok {
				ref.ToName = original
			}
			*refs = append(*refs, *ref)
		}
	}

	for i := 0; i < int(node.ChildCount()); i++ {
		collectRefs(node.Child(i), source, filePath, aliases, refs)
	}
}

// isRelativeImport returns true if the module path starts with "." or "/",
// indicating a project-internal import.
func isRelativeImport(modulePath string) bool {
	return strings.HasPrefix(modulePath, ".") || strings.HasPrefix(modulePath, "/")
}

func extractImportRef(node *sitter.Node, source []byte, filePath string) *analysis.Reference {
	// Find the string node (module source)
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

	if !isRelativeImport(moduleSource) {
		return &analysis.Reference{
			ToName:             moduleSource,
			Kind:               "imports",
			FilePath:           filePath,
			IsExternal:         true,
			ExternalImportPath: moduleSource,
		}
	}

	return &analysis.Reference{
		ToName:   moduleSource,
		Kind:     "imports",
		FilePath: filePath,
	}
}

// jsBuiltins is the set of JavaScript/TypeScript built-in methods, properties,
// and global functions that should be filtered from call references.
var jsBuiltins = map[string]bool{
	// Array methods
	"push": true, "pop": true, "shift": true, "unshift": true, "splice": true,
	"slice": true, "concat": true, "join": true, "reverse": true, "sort": true,
	"filter": true, "map": true, "reduce": true, "forEach": true, "find": true,
	"findIndex": true, "some": true, "every": true, "includes": true, "indexOf": true,
	"flat": true, "flatMap": true, "fill": true, "entries": true, "keys": true,
	"values": true, "at": true, "from": true, "of": true,
	// String methods
	"charAt": true, "charCodeAt": true, "split": true, "substring": true,
	"toLowerCase": true, "toUpperCase": true, "trim": true, "trimStart": true,
	"trimEnd": true, "replace": true, "replaceAll": true, "match": true,
	"search": true, "startsWith": true, "endsWith": true, "padStart": true,
	"padEnd": true, "repeat": true,
	// Object/Math/JSON/Number/Date methods
	"stringify": true, "parse": true, "assign": true, "freeze": true,
	"floor": true, "ceil": true, "round": true, "random": true, "pow": true,
	"abs": true, "sqrt": true, "max": true, "min": true,
	"parseInt": true, "parseFloat": true, "isNaN": true, "isFinite": true,
	"toFixed": true, "toLocaleDateString": true, "toLocaleTimeString": true,
	"toLocaleString": true, "toString": true, "valueOf": true,
	"now": true, "getFullYear": true, "getMonth": true, "getDate": true,
	"getTime": true, "getHours": true, "getMinutes": true, "getSeconds": true,
	// DOM methods
	"querySelector": true, "querySelectorAll": true, "getElementById": true,
	"getElementsByClassName": true, "getElementsByTagName": true,
	"createElement": true, "appendChild": true, "removeChild": true,
	"insertBefore": true, "addEventListener": true, "removeEventListener": true,
	"getAttribute": true, "setAttribute": true, "removeAttribute": true,
	"getBoundingClientRect": true, "scrollTo": true, "scrollIntoView": true,
	"stopPropagation": true, "preventDefault": true,
	// Promise methods
	"then": true, "catch": true, "finally": true, "resolve": true, "reject": true,
	"all": true, "allSettled": true, "race": true, "any": true,
	// Global functions
	"setTimeout": true, "setInterval": true, "clearTimeout": true, "clearInterval": true,
	"encodeURIComponent": true, "decodeURIComponent": true, "encodeURI": true,
	"decodeURI": true, "btoa": true, "atob": true, "fetch": true,
	"confirm": true, "alert": true, "prompt": true,
	// Console
	"log": true, "warn": true, "error": true, "info": true, "debug": true,
	// Misc built-ins
	"Number": true, "String": true, "Boolean": true, "BigInt": true, "Symbol": true,
	"Array": true, "Object": true, "Map": true, "Set": true, "WeakMap": true,
	"require": true, "getRandomValues": true, "setUint32": true,
	// RxJS operators (commonly used with Angular)
	"pipe": true, "subscribe": true, "unsubscribe": true, "next": true,
	"complete": true, "emit": true, "asObservable": true, "asReadonly": true,
	"catchError": true, "switchMap": true, "mergeMap": true, "concatMap": true,
	"exhaustMap": true, "debounceTime": true, "throttleTime": true,
	"distinctUntilChanged": true, "takeUntil": true, "take": true, "skip": true,
	"delay": true, "finalize": true, "tap": true, "retry": true, "share": true,
	"shareReplay": true, "toPromise": true, "firstValueFrom": true,
	"lastValueFrom": true, "forkJoin": true, "combineLatest": true,
	"merge": true, "fromEvent": true, "throwError": true,
	// Angular form methods
	"patchValue": true, "setValue": true, "getValue": true, "markAsTouched": true,
	"markAsDirty": true, "reset": true, "get": true, "set": true,
	// Navigation
	"navigate": true, "navigateByUrl": true, "createUrlTree": true,
	// Test methods
	"detectChanges": true, "test": true, "describe": true, "it": true, "expect": true,
	"beforeEach": true, "afterEach": true,
	// Generic method-like names
	"apply": true, "call": true, "bind": true, "append": true,
	"delete": true, "update": true, "put": true, "post": true,
	"run": true, "start": true, "stop": true, "close": true,
	// Angular decorators and functions
	"Pipe": true,
	"Component": true, "Injectable": true, "NgModule": true, "Directive": true,
	"Input": true, "Output": true, "ViewChild": true, "ViewChildren": true,
	"ContentChild": true, "ContentChildren": true, "HostListener": true,
	"HostBinding": true, "Optional": true, "SkipSelf": true, "Self": true,
	"Inject": true, "inject": true, "computed": true, "signal": true,
	"output": true, "input": true, "effect": true, "model": true,
	"bootstrapApplication": true, "provideRouter": true, "provideHttpClient": true,
	"provideAnimations": true, "provideAuth": true, "provideFirebaseApp": true,
	"provideZonelessChangeDetection": true, "provideBrowserGlobalErrorListeners": true,
	"withFetch": true, "withInMemoryScrolling": true, "initializeApp": true,
	"getAuth": true, "signInWithEmailAndPassword": true, "signInWithPopup": true,
	"getIdToken": true, "minLength": true, "control": true, "validation": true,
	"group": true, "format": true,
}

func extractCallRef(node *sitter.Node, source []byte, filePath string) *analysis.Reference {
	fnNode := node.ChildByFieldName("function")
	if fnNode == nil {
		return nil
	}

	// Skip anonymous function calls
	if fnNode.Type() == "function" || fnNode.Type() == "arrow_function" {
		return nil
	}

	// Skip parenthesized anonymous functions: (function(){})() or (()=>x)()
	if fnNode.Type() == "parenthesized_expression" {
		return nil
	}

	// Skip await expressions used as function targets
	if fnNode.Type() == "await_expression" {
		return nil
	}

	// Normalize member_expression to property name
	if fnNode.Type() == "member_expression" {
		objNode := fnNode.ChildByFieldName("object")
		propNode := fnNode.ChildByFieldName("property")
		if propNode == nil {
			return nil
		}
		name := nodeText(propNode, source)
		if jsBuiltins[name] {
			return nil
		}
		// Skip this.method() and this.prop.method() calls — these are
		// internal class method calls or calls on class properties.
		if objNode != nil && hasThisRoot(objNode, source) {
			return nil
		}
		return &analysis.Reference{
			ToName:   name,
			Kind:     "calls",
			FilePath: filePath,
		}
	}

	name := nodeText(fnNode, source)
	// Skip multi-word names (tree-sitter artifacts like "await this.http")
	if strings.Contains(name, " ") {
		return nil
	}
	if jsBuiltins[name] {
		return nil
	}
	return &analysis.Reference{
		ToName:   name,
		Kind:     "calls",
		FilePath: filePath,
	}
}

// hasThisRoot returns true if the expression tree is rooted at "this",
// handling chained member expressions like this.foo.bar.
func hasThisRoot(node *sitter.Node, source []byte) bool {
	if node.Type() == "this" {
		return true
	}
	if node.Type() == "member_expression" {
		obj := node.ChildByFieldName("object")
		if obj != nil {
			return hasThisRoot(obj, source)
		}
	}
	return false
}

func findChildByType(node *sitter.Node, typeName string) *sitter.Node {
	for i := 0; i < int(node.ChildCount()); i++ {
		if node.Child(i).Type() == typeName {
			return node.Child(i)
		}
	}
	return nil
}

func nodeText(node *sitter.Node, source []byte) string {
	return string(source[node.StartByte():node.EndByte()])
}

func lineCount(node *sitter.Node) int {
	return int(node.EndPoint().Row-node.StartPoint().Row) + 1
}
