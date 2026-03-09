package golang

import (
	"context"
	"testing"

	sitter "github.com/smacker/go-tree-sitter"
	"github.com/smacker/go-tree-sitter/golang"
	"github.com/tab58/code-context/internal/analysis"
)

// --- Task 13: Go extractor tests ---

// parseGo is a test helper that parses Go source code into a tree-sitter tree.
func parseGo(t *testing.T, source []byte) *sitter.Tree {
	t.Helper()
	parser := sitter.NewParser()
	parser.SetLanguage(golang.GetLanguage())
	tree, err := parser.ParseCtx(context.Background(), nil, source)
	if err != nil {
		t.Fatalf("failed to parse Go source: %v", err)
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

// TestGoExtractor_ImplementsExtractor verifies that GoExtractor satisfies
// the analysis.Extractor interface (compile-time check).
// Expected result: Compiles without errors.
func TestGoExtractor_ImplementsExtractor(t *testing.T) {
	var _ analysis.Extractor = &GoExtractor{}
}

// TestGoExtractor_ExtractsFunctionDeclaration verifies that the Go extractor
// extracts top-level function declarations.
// Expected result: Symbol with Kind="function", correct name, signature, source.
func TestGoExtractor_ExtractsFunctionDeclaration(t *testing.T) {
	source := []byte(`package main

func Hello(name string) string {
	return "Hello, " + name
}
`)
	tree := parseGo(t, source)
	ext := NewGoExtractor()

	symbols, err := ext.ExtractSymbols(tree, source, "main.go", "")
	if err != nil {
		t.Fatalf("ExtractSymbols error: %v", err)
	}

	sym := findSymbol(symbols, "Hello")
	if sym == nil {
		t.Fatal("expected to find symbol 'Hello', got nil")
	}
	if sym.Kind != "function" {
		t.Errorf("Kind = %q, want %q", sym.Kind, "function")
	}
	if sym.Language != "go" {
		t.Errorf("Language = %q, want %q", sym.Language, "go")
	}
	if sym.Source == "" {
		t.Error("Source should not be empty")
	}
	if sym.LineNumber <= 0 {
		t.Errorf("LineNumber = %d, want > 0", sym.LineNumber)
	}
}

// TestGoExtractor_ExtractsMethodDeclaration verifies that the Go extractor
// extracts method declarations with the correct ParentName (receiver type).
// Expected result: Symbol with Kind="method", ParentName="MyStruct".
func TestGoExtractor_ExtractsMethodDeclaration(t *testing.T) {
	source := []byte(`package main

type MyStruct struct{}

func (s *MyStruct) DoWork() error {
	return nil
}
`)
	tree := parseGo(t, source)
	ext := NewGoExtractor()

	symbols, err := ext.ExtractSymbols(tree, source, "main.go", "")
	if err != nil {
		t.Fatalf("ExtractSymbols error: %v", err)
	}

	sym := findSymbol(symbols, "DoWork")
	if sym == nil {
		t.Fatal("expected to find symbol 'DoWork', got nil")
	}
	if sym.Kind != "method" {
		t.Errorf("Kind = %q, want %q", sym.Kind, "method")
	}
	if sym.ParentName != "MyStruct" {
		t.Errorf("ParentName = %q, want %q", sym.ParentName, "MyStruct")
	}
}

// TestGoExtractor_ExtractsStructType verifies that the Go extractor extracts
// struct type declarations as class-kind symbols.
// Expected result: Symbol with Kind="struct".
func TestGoExtractor_ExtractsStructType(t *testing.T) {
	source := []byte(`package main

type User struct {
	Name string
	Age  int
}
`)
	tree := parseGo(t, source)
	ext := NewGoExtractor()

	symbols, err := ext.ExtractSymbols(tree, source, "main.go", "")
	if err != nil {
		t.Fatalf("ExtractSymbols error: %v", err)
	}

	sym := findSymbol(symbols, "User")
	if sym == nil {
		t.Fatal("expected to find symbol 'User', got nil")
	}
	if sym.Kind != "struct" {
		t.Errorf("Kind = %q, want %q", sym.Kind, "struct")
	}
}

// TestGoExtractor_ExtractsInterfaceType verifies that the Go extractor extracts
// interface type declarations.
// Expected result: Symbol with Kind="interface".
func TestGoExtractor_ExtractsInterfaceType(t *testing.T) {
	source := []byte(`package main

type Reader interface {
	Read(p []byte) (n int, err error)
}
`)
	tree := parseGo(t, source)
	ext := NewGoExtractor()

	symbols, err := ext.ExtractSymbols(tree, source, "main.go", "")
	if err != nil {
		t.Fatalf("ExtractSymbols error: %v", err)
	}

	sym := findSymbol(symbols, "Reader")
	if sym == nil {
		t.Fatal("expected to find symbol 'Reader', got nil")
	}
	if sym.Kind != "interface" {
		t.Errorf("Kind = %q, want %q", sym.Kind, "interface")
	}
}

// TestGoExtractor_ExtractsPackageClause verifies that the Go extractor extracts
// the package declaration as a module symbol.
// Expected result: Symbol with Kind="module", Name="main".
func TestGoExtractor_ExtractsPackageClause(t *testing.T) {
	source := []byte(`package mypackage

func init() {}
`)
	tree := parseGo(t, source)
	ext := NewGoExtractor()

	symbols, err := ext.ExtractSymbols(tree, source, "mypackage/file.go", "")
	if err != nil {
		t.Fatalf("ExtractSymbols error: %v", err)
	}

	sym := findSymbol(symbols, "mypackage")
	if sym == nil {
		t.Fatal("expected to find symbol 'mypackage', got nil")
	}
	if sym.Kind != "module" {
		t.Errorf("Kind = %q, want %q", sym.Kind, "module")
	}
}

// TestGoExtractor_GoVisibility verifies that exported names (uppercase) get
// visibility="public" and unexported names get visibility="package".
// Expected result: Correct visibility for each symbol.
func TestGoExtractor_GoVisibility(t *testing.T) {
	source := []byte(`package main

func ExportedFunc() {}
func unexportedFunc() {}
`)
	tree := parseGo(t, source)
	ext := NewGoExtractor()

	symbols, err := ext.ExtractSymbols(tree, source, "main.go", "")
	if err != nil {
		t.Fatalf("ExtractSymbols error: %v", err)
	}

	exported := findSymbol(symbols, "ExportedFunc")
	if exported == nil {
		t.Fatal("expected to find symbol 'ExportedFunc'")
	}
	if exported.Visibility != "public" {
		t.Errorf("ExportedFunc.Visibility = %q, want %q", exported.Visibility, "public")
	}

	unexported := findSymbol(symbols, "unexportedFunc")
	if unexported == nil {
		t.Fatal("expected to find symbol 'unexportedFunc'")
	}
	if unexported.Visibility != "package" {
		t.Errorf("unexportedFunc.Visibility = %q, want %q", unexported.Visibility, "package")
	}
}

// TestGoExtractor_ExtractsImportReferences verifies that import statements
// are extracted as references with Kind="imports".
// Expected result: Reference with Kind="imports" for each import.
func TestGoExtractor_ExtractsImportReferences(t *testing.T) {
	source := []byte(`package main

import (
	"fmt"
	"os"
)

func main() {}
`)
	tree := parseGo(t, source)
	ext := NewGoExtractor()

	refs, err := ext.ExtractReferences(tree, source, "main.go", "")
	if err != nil {
		t.Fatalf("ExtractReferences error: %v", err)
	}

	fmtRef := findRef(refs, "fmt")
	if fmtRef == nil {
		t.Error("expected to find reference to 'fmt', got nil")
	}
	if fmtRef != nil && fmtRef.Kind != "imports" {
		t.Errorf("fmt reference Kind = %q, want %q", fmtRef.Kind, "imports")
	}

	osRef := findRef(refs, "os")
	if osRef == nil {
		t.Error("expected to find reference to 'os', got nil")
	}
}

// TestGoExtractor_ExtractsCallReferences verifies that function call expressions
// are extracted as references with Kind="calls".
// Expected result: Reference with Kind="calls" for each call.
func TestGoExtractor_ExtractsCallReferences(t *testing.T) {
	source := []byte(`package main

import "fmt"

func main() {
	fmt.Println("hello")
	helper()
}

func helper() {}
`)
	tree := parseGo(t, source)
	ext := NewGoExtractor()

	refs, err := ext.ExtractReferences(tree, source, "main.go", "")
	if err != nil {
		t.Fatalf("ExtractReferences error: %v", err)
	}

	// Should find calls to fmt.Println and helper
	var foundPrintln, foundHelper bool
	for _, ref := range refs {
		if ref.Kind == "calls" {
			if ref.ToName == "fmt.Println" || ref.ToName == "Println" {
				foundPrintln = true
			}
			if ref.ToName == "helper" {
				foundHelper = true
			}
		}
	}

	if !foundPrintln {
		t.Error("expected to find call reference to fmt.Println")
	}
	if !foundHelper {
		t.Error("expected to find call reference to helper")
	}
}
