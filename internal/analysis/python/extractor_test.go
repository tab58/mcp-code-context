package python

import (
	"context"
	"testing"

	sitter "github.com/smacker/go-tree-sitter"
	"github.com/smacker/go-tree-sitter/python"
	"github.com/tab58/code-context/internal/analysis"
)

// === Task 3: Python extractor tests ===

// parsePy is a test helper that parses Python source code into a tree-sitter tree.
func parsePy(t *testing.T, source []byte) *sitter.Tree {
	t.Helper()
	parser := sitter.NewParser()
	parser.SetLanguage(python.GetLanguage())
	tree, err := parser.ParseCtx(context.Background(), nil, source)
	if err != nil {
		t.Fatalf("failed to parse Python source: %v", err)
	}
	return tree
}

// findSymbol is a test helper that finds a symbol by name in a slice.
func findSymbol(symbols []analysis.Symbol, name string) *analysis.Symbol {
	for i := range symbols {
		if symbols[i].Name == name {
			return &symbols[i]
		}
	}
	return nil
}

// findRef is a test helper that finds a reference by ToName in a slice.
func findRef(refs []analysis.Reference, toName string) *analysis.Reference {
	for i := range refs {
		if refs[i].ToName == toName {
			return &refs[i]
		}
	}
	return nil
}

// findFuncNode recursively searches for a function_definition node.
func findFuncNode(node *sitter.Node) *sitter.Node {
	if node.Type() == "function_definition" {
		return node
	}
	for i := 0; i < int(node.NamedChildCount()); i++ {
		if found := findFuncNode(node.NamedChild(i)); found != nil {
			return found
		}
	}
	return nil
}

// parsePyFunc parses Python source and returns the first function_definition node.
func parsePyFunc(t *testing.T, source string) *sitter.Node {
	t.Helper()
	parser := sitter.NewParser()
	parser.SetLanguage(python.GetLanguage())
	tree, err := parser.ParseCtx(context.Background(), nil, []byte(source))
	if err != nil {
		t.Fatalf("parse failed: %v", err)
	}
	return findFuncNode(tree.RootNode())
}

// --- Interface compliance ---

// TestPythonExtractor_ImplementsExtractor verifies that PythonExtractor
// satisfies the analysis.Extractor interface.
// Expected result: Compiles without errors.
func TestPythonExtractor_ImplementsExtractor(t *testing.T) {
	var _ analysis.Extractor = &PythonExtractor{}
}

// TestPythonComplexityExtractor_ImplementsComplexityExtractor verifies that
// PythonComplexityExtractor satisfies the analysis.ComplexityExtractor interface.
// Expected result: Compiles without errors.
func TestPythonComplexityExtractor_ImplementsComplexityExtractor(t *testing.T) {
	var _ analysis.ComplexityExtractor = &PythonComplexityExtractor{}
}

// --- Register pattern ---

// TestPythonRegister_RegistersExtractorAndComplexity verifies that Register
// wires both Extractor and ComplexityExtractor into the registry.
// Expected result: Both extractor types are retrievable for "python".
func TestPythonRegister_RegistersExtractorAndComplexity(t *testing.T) {
	r := analysis.NewRegistry()
	Register(r)

	ext, ok := r.ExtractorForLanguage("python")
	if !ok {
		t.Fatal("ExtractorForLanguage(python) returned false after Register")
	}
	if ext == nil {
		t.Fatal("ExtractorForLanguage(python) returned nil after Register")
	}

	ce, ok := r.ComplexityExtractorForLanguage("python")
	if !ok {
		t.Fatal("ComplexityExtractorForLanguage(python) returned false after Register")
	}
	if ce == nil {
		t.Fatal("ComplexityExtractorForLanguage(python) returned nil after Register")
	}
}

// --- Symbol extraction ---

// TestPythonExtractor_ExtractsFunctionDefinition verifies extraction of
// top-level function definitions in Python.
// Expected result: Symbol with Kind="function", correct name, language="python".
func TestPythonExtractor_ExtractsFunctionDefinition(t *testing.T) {
	source := []byte(`def greet(name):
    return f"Hello, {name}"
`)
	tree := parsePy(t, source)
	ext := NewPythonExtractor()

	symbols, err := ext.ExtractSymbols(tree, source, "app.py", "/repo")
	if err != nil {
		t.Fatalf("ExtractSymbols error: %v", err)
	}

	sym := findSymbol(symbols, "greet")
	if sym == nil {
		t.Fatal("expected to find symbol 'greet', got nil")
	}
	if sym.Kind != "function" {
		t.Errorf("Kind = %q, want %q", sym.Kind, "function")
	}
	if sym.Language != "python" {
		t.Errorf("Language = %q, want %q", sym.Language, "python")
	}
}

// TestPythonExtractor_ExtractsClassDefinition verifies extraction of class definitions.
// Expected result: Symbol with Kind="class", correct name.
func TestPythonExtractor_ExtractsClassDefinition(t *testing.T) {
	source := []byte(`class UserService:
    def __init__(self, db):
        self.db = db

    def get_user(self, user_id):
        return self.db.find(user_id)
`)
	tree := parsePy(t, source)
	ext := NewPythonExtractor()

	symbols, err := ext.ExtractSymbols(tree, source, "service.py", "/repo")
	if err != nil {
		t.Fatalf("ExtractSymbols error: %v", err)
	}

	sym := findSymbol(symbols, "UserService")
	if sym == nil {
		t.Fatal("expected to find symbol 'UserService', got nil")
	}
	if sym.Kind != "class" {
		t.Errorf("Kind = %q, want %q", sym.Kind, "class")
	}
}

// TestPythonExtractor_ExtractsClassMethods verifies extraction of methods inside classes.
// Expected result: Method symbols with Kind="method" and correct ParentName.
func TestPythonExtractor_ExtractsClassMethods(t *testing.T) {
	source := []byte(`class Calc:
    def add(self, a, b):
        return a + b

    def sub(self, a, b):
        return a - b
`)
	tree := parsePy(t, source)
	ext := NewPythonExtractor()

	symbols, err := ext.ExtractSymbols(tree, source, "calc.py", "/repo")
	if err != nil {
		t.Fatalf("ExtractSymbols error: %v", err)
	}

	addSym := findSymbol(symbols, "add")
	if addSym == nil {
		t.Fatal("expected to find method 'add', got nil")
	}
	if addSym.Kind != "method" {
		t.Errorf("Kind = %q, want %q", addSym.Kind, "method")
	}
	if addSym.ParentName != "Calc" {
		t.Errorf("ParentName = %q, want %q", addSym.ParentName, "Calc")
	}
}

// TestPythonExtractor_UnderscoreVisibility verifies that leading underscore
// makes symbols private per Python convention.
// Expected result: _private_func has Visibility="private", public_func has "public".
func TestPythonExtractor_UnderscoreVisibility(t *testing.T) {
	source := []byte(`def public_func():
    pass

def _private_func():
    pass
`)
	tree := parsePy(t, source)
	ext := NewPythonExtractor()

	symbols, err := ext.ExtractSymbols(tree, source, "mod.py", "/repo")
	if err != nil {
		t.Fatalf("ExtractSymbols error: %v", err)
	}

	pubSym := findSymbol(symbols, "public_func")
	if pubSym == nil {
		t.Fatal("expected to find 'public_func'")
	}
	if pubSym.Visibility != "public" {
		t.Errorf("public_func Visibility = %q, want %q", pubSym.Visibility, "public")
	}

	privSym := findSymbol(symbols, "_private_func")
	if privSym == nil {
		t.Fatal("expected to find '_private_func'")
	}
	if privSym.Visibility != "private" {
		t.Errorf("_private_func Visibility = %q, want %q", privSym.Visibility, "private")
	}
}

// TestPythonExtractor_EmitsModulePerFile verifies one Module symbol per file.
// Expected result: Module with dotted import path relative to repo.
func TestPythonExtractor_EmitsModulePerFile(t *testing.T) {
	source := []byte(`def main():
    pass
`)
	tree := parsePy(t, source)
	ext := NewPythonExtractor()

	symbols, err := ext.ExtractSymbols(tree, source, "/repo/src/app.py", "/repo")
	if err != nil {
		t.Fatalf("ExtractSymbols error: %v", err)
	}

	var moduleSym *analysis.Symbol
	for i := range symbols {
		if symbols[i].Kind == "module" {
			moduleSym = &symbols[i]
			break
		}
	}
	if moduleSym == nil {
		t.Fatal("expected to find module symbol, got nil")
	}
	// Python uses dotted module path: src.app
	if moduleSym.ImportPath != "src.app" {
		t.Errorf("ImportPath = %q, want %q", moduleSym.ImportPath, "src.app")
	}
}

// TestPythonExtractor_ExtractsDecorators verifies that decorators are captured.
// Expected result: Function with decorator has non-empty Decorators slice.
func TestPythonExtractor_ExtractsDecorators(t *testing.T) {
	source := []byte(`import functools

@functools.cache
def expensive(n):
    return n * 2
`)
	tree := parsePy(t, source)
	ext := NewPythonExtractor()

	symbols, err := ext.ExtractSymbols(tree, source, "mod.py", "/repo")
	if err != nil {
		t.Fatalf("ExtractSymbols error: %v", err)
	}

	sym := findSymbol(symbols, "expensive")
	if sym == nil {
		t.Fatal("expected to find 'expensive'")
	}
	if len(sym.Decorators) == 0 {
		t.Error("expected non-empty Decorators for decorated function")
	}
}

// --- Reference extraction ---

// TestPythonExtractor_ExtractsImportStatement verifies extraction of import references.
// Expected result: Reference with Kind="imports" for import statements.
func TestPythonExtractor_ExtractsImportStatement(t *testing.T) {
	source := []byte(`import os
from pathlib import Path

def main():
    p = Path(".")
`)
	tree := parsePy(t, source)
	ext := NewPythonExtractor()

	refs, err := ext.ExtractReferences(tree, source, "app.py", "/repo")
	if err != nil {
		t.Fatalf("ExtractReferences error: %v", err)
	}

	osRef := findRef(refs, "os")
	if osRef == nil {
		t.Fatal("expected to find import reference for 'os'")
	}
	if osRef.Kind != "imports" {
		t.Errorf("Kind = %q, want %q", osRef.Kind, "imports")
	}
}

// TestPythonExtractor_ExtractsCallReferences verifies extraction of function calls.
// Expected result: Reference with Kind="calls" for function calls.
func TestPythonExtractor_ExtractsCallReferences(t *testing.T) {
	source := []byte(`def caller():
    do_work()
`)
	tree := parsePy(t, source)
	ext := NewPythonExtractor()

	refs, err := ext.ExtractReferences(tree, source, "app.py", "/repo")
	if err != nil {
		t.Fatalf("ExtractReferences error: %v", err)
	}

	callRef := findRef(refs, "do_work")
	if callRef == nil {
		t.Fatal("expected to find call reference for 'do_work'")
	}
	if callRef.Kind != "calls" {
		t.Errorf("Kind = %q, want %q", callRef.Kind, "calls")
	}
}

// TestPythonExtractor_StdlibImportIsExternal verifies that stdlib imports
// are marked as external.
// Expected result: Import reference for 'os' has IsExternal=true.
func TestPythonExtractor_StdlibImportIsExternal(t *testing.T) {
	source := []byte(`import os

def main():
    os.listdir(".")
`)
	tree := parsePy(t, source)
	ext := NewPythonExtractor()

	refs, err := ext.ExtractReferences(tree, source, "app.py", "/repo")
	if err != nil {
		t.Fatalf("ExtractReferences error: %v", err)
	}

	osRef := findRef(refs, "os")
	if osRef == nil {
		t.Fatal("expected to find import reference for 'os'")
	}
	if !osRef.IsExternal {
		t.Error("stdlib import 'os' should have IsExternal=true")
	}
}

// --- Complexity ---

// TestPythonComplexityExtractor_EmptyFunction verifies base complexity of 1.
// Expected result: Complexity = 1.
func TestPythonComplexityExtractor_EmptyFunction(t *testing.T) {
	source := `def empty():
    pass
`
	node := parsePyFunc(t, source)
	if node == nil {
		t.Fatal("no function node found")
	}
	ext := NewPythonComplexityExtractor()
	result := ext.ComputeComplexity(node, []byte(source))
	if result != 1 {
		t.Errorf("empty function complexity = %d, want 1", result)
	}
}

// TestPythonComplexityExtractor_IfStatement verifies +1 for if_statement.
// Expected result: Complexity = 2 (base 1 + 1 if).
func TestPythonComplexityExtractor_IfStatement(t *testing.T) {
	source := `def with_if(x):
    if x > 0:
        return True
`
	node := parsePyFunc(t, source)
	if node == nil {
		t.Fatal("no function node found")
	}
	ext := NewPythonComplexityExtractor()
	result := ext.ComputeComplexity(node, []byte(source))
	if result != 2 {
		t.Errorf("function with if complexity = %d, want 2", result)
	}
}

// TestPythonComplexityExtractor_ForStatement verifies +1 for for_statement.
// Expected result: Complexity = 2 (base 1 + 1 for).
func TestPythonComplexityExtractor_ForStatement(t *testing.T) {
	source := `def with_for(items):
    for item in items:
        print(item)
`
	node := parsePyFunc(t, source)
	if node == nil {
		t.Fatal("no function node found")
	}
	ext := NewPythonComplexityExtractor()
	result := ext.ComputeComplexity(node, []byte(source))
	if result != 2 {
		t.Errorf("function with for complexity = %d, want 2", result)
	}
}

// TestPythonComplexityExtractor_WhileStatement verifies +1 for while_statement.
// Expected result: Complexity = 2 (base 1 + 1 while).
func TestPythonComplexityExtractor_WhileStatement(t *testing.T) {
	source := `def with_while():
    while True:
        break
`
	node := parsePyFunc(t, source)
	if node == nil {
		t.Fatal("no function node found")
	}
	ext := NewPythonComplexityExtractor()
	result := ext.ComputeComplexity(node, []byte(source))
	if result != 2 {
		t.Errorf("function with while complexity = %d, want 2", result)
	}
}

// TestPythonComplexityExtractor_ExceptClause verifies +1 for except_clause.
// Expected result: Complexity = 2 (base 1 + 1 except).
func TestPythonComplexityExtractor_ExceptClause(t *testing.T) {
	source := `def with_try():
    try:
        do_work()
    except ValueError:
        handle_error()
`
	node := parsePyFunc(t, source)
	if node == nil {
		t.Fatal("no function node found")
	}
	ext := NewPythonComplexityExtractor()
	result := ext.ComputeComplexity(node, []byte(source))
	if result != 2 {
		t.Errorf("function with try/except complexity = %d, want 2", result)
	}
}

// TestPythonComplexityExtractor_BooleanOperators verifies +1 for "and" and "or".
// Expected result: Complexity = 4 (base 1 + 1 if + 1 and + 1 or).
func TestPythonComplexityExtractor_BooleanOperators(t *testing.T) {
	source := `def with_logic(a, b, c):
    if a and b or c:
        return True
`
	node := parsePyFunc(t, source)
	if node == nil {
		t.Fatal("no function node found")
	}
	ext := NewPythonComplexityExtractor()
	result := ext.ComputeComplexity(node, []byte(source))
	if result != 4 {
		t.Errorf("function with if and or complexity = %d, want 4", result)
	}
}

// TestPythonComplexityExtractor_ElifClause verifies +1 for elif_clause.
// Expected result: Complexity = 3 (base 1 + 1 if + 1 elif).
func TestPythonComplexityExtractor_ElifClause(t *testing.T) {
	source := `def with_elif(x):
    if x > 0:
        return "positive"
    elif x < 0:
        return "negative"
    else:
        return "zero"
`
	node := parsePyFunc(t, source)
	if node == nil {
		t.Fatal("no function node found")
	}
	ext := NewPythonComplexityExtractor()
	result := ext.ComputeComplexity(node, []byte(source))
	if result != 3 {
		t.Errorf("function with if/elif complexity = %d, want 3", result)
	}
}

// TestPythonComplexityExtractor_Assert verifies +1 for assert.
// Expected result: Complexity = 2 (base 1 + 1 assert).
func TestPythonComplexityExtractor_Assert(t *testing.T) {
	source := `def with_assert(x):
    assert x > 0
`
	node := parsePyFunc(t, source)
	if node == nil {
		t.Fatal("no function node found")
	}
	ext := NewPythonComplexityExtractor()
	result := ext.ComputeComplexity(node, []byte(source))
	if result != 2 {
		t.Errorf("function with assert complexity = %d, want 2", result)
	}
}
