package typescript

import (
	"context"
	"testing"

	sitter "github.com/smacker/go-tree-sitter"
	"github.com/smacker/go-tree-sitter/typescript/tsx"
	"github.com/smacker/go-tree-sitter/typescript/typescript"
	"github.com/tab58/code-context/internal/analysis"
)

// --- Task 14: TypeScript + TSX extractor tests ---

// parseTS is a test helper that parses TypeScript source into a tree-sitter tree.
func parseTS(t *testing.T, source []byte) *sitter.Tree {
	t.Helper()
	parser := sitter.NewParser()
	parser.SetLanguage(typescript.GetLanguage())
	tree, err := parser.ParseCtx(context.Background(), nil, source)
	if err != nil {
		t.Fatalf("failed to parse TypeScript source: %v", err)
	}
	return tree
}

// parseTSX is a test helper that parses TSX source into a tree-sitter tree.
func parseTSX(t *testing.T, source []byte) *sitter.Tree {
	t.Helper()
	parser := sitter.NewParser()
	parser.SetLanguage(tsx.GetLanguage())
	tree, err := parser.ParseCtx(context.Background(), nil, source)
	if err != nil {
		t.Fatalf("failed to parse TSX source: %v", err)
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

// TestTypeScriptExtractor_ImplementsExtractor verifies that TypeScriptExtractor
// satisfies the analysis.Extractor interface (compile-time check).
// Expected result: Compiles without errors.
func TestTypeScriptExtractor_ImplementsExtractor(t *testing.T) {
	var _ analysis.Extractor = &TypeScriptExtractor{}
}

// TestTSXExtractor_ImplementsExtractor verifies that TSXExtractor
// satisfies the analysis.Extractor interface (compile-time check).
// Expected result: Compiles without errors.
func TestTSXExtractor_ImplementsExtractor(t *testing.T) {
	var _ analysis.Extractor = &TSXExtractor{}
}

// TestTSExtractor_ExtractsFunctionDeclaration verifies that the TypeScript
// extractor extracts named function declarations.
// Expected result: Symbol with Kind="function".
func TestTSExtractor_ExtractsFunctionDeclaration(t *testing.T) {
	source := []byte(`function greet(name: string): string {
  return "Hello, " + name;
}
`)
	tree := parseTS(t, source)
	ext := NewTypeScriptExtractor()

	symbols, err := ext.ExtractSymbols(tree, source, "app.ts", "")
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
	if sym.Language != "typescript" {
		t.Errorf("Language = %q, want %q", sym.Language, "typescript")
	}
	if sym.Source == "" {
		t.Error("Source should not be empty")
	}
}

// TestTSExtractor_ExtractsClassDeclaration verifies that the TypeScript
// extractor extracts class declarations.
// Expected result: Symbol with Kind="class".
func TestTSExtractor_ExtractsClassDeclaration(t *testing.T) {
	source := []byte(`export class UserService {
  private name: string;

  constructor(name: string) {
    this.name = name;
  }

  getName(): string {
    return this.name;
  }
}
`)
	tree := parseTS(t, source)
	ext := NewTypeScriptExtractor()

	symbols, err := ext.ExtractSymbols(tree, source, "service.ts", "")
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
	if sym.Visibility != "public" {
		t.Errorf("Visibility = %q, want %q (export keyword)", sym.Visibility, "public")
	}
}

// TestTSExtractor_ExtractsInterfaceDeclaration verifies that the TypeScript
// extractor extracts interface declarations.
// Expected result: Symbol with Kind="interface".
func TestTSExtractor_ExtractsInterfaceDeclaration(t *testing.T) {
	source := []byte(`interface User {
  name: string;
  age: number;
}
`)
	tree := parseTS(t, source)
	ext := NewTypeScriptExtractor()

	symbols, err := ext.ExtractSymbols(tree, source, "types.ts", "")
	if err != nil {
		t.Fatalf("ExtractSymbols error: %v", err)
	}

	sym := findSymbol(symbols, "User")
	if sym == nil {
		t.Fatal("expected to find symbol 'User', got nil")
	}
	if sym.Kind != "interface" {
		t.Errorf("Kind = %q, want %q", sym.Kind, "interface")
	}
}

// TestTSExtractor_ExtractsEnumDeclaration verifies that the TypeScript
// extractor extracts enum declarations as class kind="enum".
// Expected result: Symbol with Kind="class", and some indication of enum.
func TestTSExtractor_ExtractsEnumDeclaration(t *testing.T) {
	source := []byte(`enum Direction {
  Up,
  Down,
  Left,
  Right
}
`)
	tree := parseTS(t, source)
	ext := NewTypeScriptExtractor()

	symbols, err := ext.ExtractSymbols(tree, source, "enums.ts", "")
	if err != nil {
		t.Fatalf("ExtractSymbols error: %v", err)
	}

	sym := findSymbol(symbols, "Direction")
	if sym == nil {
		t.Fatal("expected to find symbol 'Direction', got nil")
	}
	// Spec says enums map to Kind="class" with kind="enum" in the Class node
	// At the Symbol level, we just need to capture it
	if sym.Kind == "" {
		t.Error("Kind should not be empty for enum")
	}
}

// TestTSExtractor_ExportVisibility verifies that exported declarations get
// visibility="public" and non-exported get visibility="private".
// Expected result: Correct visibility based on export keyword.
func TestTSExtractor_ExportVisibility(t *testing.T) {
	source := []byte(`export function publicFunc(): void {}
function privateFunc(): void {}
`)
	tree := parseTS(t, source)
	ext := NewTypeScriptExtractor()

	symbols, err := ext.ExtractSymbols(tree, source, "mod.ts", "")
	if err != nil {
		t.Fatalf("ExtractSymbols error: %v", err)
	}

	pub := findSymbol(symbols, "publicFunc")
	if pub == nil {
		t.Fatal("expected to find symbol 'publicFunc'")
	}
	if pub.Visibility != "public" {
		t.Errorf("publicFunc.Visibility = %q, want %q", pub.Visibility, "public")
	}

	priv := findSymbol(symbols, "privateFunc")
	if priv == nil {
		t.Fatal("expected to find symbol 'privateFunc'")
	}
	if priv.Visibility != "private" {
		t.Errorf("privateFunc.Visibility = %q, want %q", priv.Visibility, "private")
	}
}

// TestTSExtractor_ExtractsImportReferences verifies that import statements
// are extracted as references with Kind="imports".
// Expected result: Reference with Kind="imports" for each import.
func TestTSExtractor_ExtractsImportReferences(t *testing.T) {
	source := []byte(`import { readFile } from 'fs';
import express from 'express';

function main() {}
`)
	tree := parseTS(t, source)
	ext := NewTypeScriptExtractor()

	refs, err := ext.ExtractReferences(tree, source, "app.ts", "")
	if err != nil {
		t.Fatalf("ExtractReferences error: %v", err)
	}

	fsRef := findRef(refs, "fs")
	if fsRef == nil {
		t.Error("expected to find reference to 'fs', got nil")
	}
	if fsRef != nil && fsRef.Kind != "imports" {
		t.Errorf("fs reference Kind = %q, want %q", fsRef.Kind, "imports")
	}

	expressRef := findRef(refs, "express")
	if expressRef == nil {
		t.Error("expected to find reference to 'express', got nil")
	}
}

// TestTSExtractor_ExtractsCallReferences verifies that function calls are
// extracted as references with Kind="calls".
// Expected result: Reference with Kind="calls".
func TestTSExtractor_ExtractsCallReferences(t *testing.T) {
	source := []byte(`function helper(): void {}

function main(): void {
  console.log("hello");
  helper();
}
`)
	tree := parseTS(t, source)
	ext := NewTypeScriptExtractor()

	refs, err := ext.ExtractReferences(tree, source, "app.ts", "")
	if err != nil {
		t.Fatalf("ExtractReferences error: %v", err)
	}

	var foundHelper bool
	for _, ref := range refs {
		if ref.Kind == "calls" && ref.ToName == "helper" {
			foundHelper = true
		}
	}
	if !foundHelper {
		t.Error("expected to find call reference to 'helper'")
	}
}

// TestTSExtractor_ExtractsMethodDefinition verifies that class methods are
// extracted with Kind="method" and ParentName set to the class name.
// Expected result: Symbol with Kind="method", ParentName="Greeter".
func TestTSExtractor_ExtractsMethodDefinition(t *testing.T) {
	source := []byte(`class Greeter {
  greet(name: string): string {
    return "Hello, " + name;
  }
}
`)
	tree := parseTS(t, source)
	ext := NewTypeScriptExtractor()

	symbols, err := ext.ExtractSymbols(tree, source, "greeter.ts", "")
	if err != nil {
		t.Fatalf("ExtractSymbols error: %v", err)
	}

	sym := findSymbol(symbols, "greet")
	if sym == nil {
		t.Fatal("expected to find symbol 'greet'")
	}
	if sym.Kind != "method" {
		t.Errorf("Kind = %q, want %q", sym.Kind, "method")
	}
	if sym.ParentName != "Greeter" {
		t.Errorf("ParentName = %q, want %q", sym.ParentName, "Greeter")
	}
}

// TestTSXExtractor_HandlesJSX verifies that the TSX extractor can parse
// source code containing JSX syntax without errors.
// Expected result: Non-nil symbols, no error.
func TestTSXExtractor_HandlesJSX(t *testing.T) {
	source := []byte(`export function App(): JSX.Element {
  return <div><h1>Hello</h1></div>;
}
`)
	tree := parseTSX(t, source)
	ext := NewTSXExtractor()

	symbols, err := ext.ExtractSymbols(tree, source, "App.tsx", "")
	if err != nil {
		t.Fatalf("ExtractSymbols error: %v", err)
	}

	sym := findSymbol(symbols, "App")
	if sym == nil {
		t.Fatal("expected to find symbol 'App' in TSX file")
	}
	if sym.Kind != "function" {
		t.Errorf("Kind = %q, want %q", sym.Kind, "function")
	}
	if sym.Language != "tsx" {
		t.Errorf("Language = %q, want %q", sym.Language, "tsx")
	}
}

// TestTSExtractor_ExtractsDecorators verifies that decorators are captured
// on class and method declarations.
// Expected result: Decorators slice is populated.
func TestTSExtractor_ExtractsDecorators(t *testing.T) {
	source := []byte(`function Injectable() { return function(target: any) {}; }

@Injectable()
class UserService {
  doWork(): void {}
}
`)
	tree := parseTS(t, source)
	ext := NewTypeScriptExtractor()

	symbols, err := ext.ExtractSymbols(tree, source, "service.ts", "")
	if err != nil {
		t.Fatalf("ExtractSymbols error: %v", err)
	}

	sym := findSymbol(symbols, "UserService")
	if sym == nil {
		t.Fatal("expected to find symbol 'UserService'")
	}
	if len(sym.Decorators) == 0 {
		t.Error("Decorators should not be empty for @Injectable class")
	}
}
