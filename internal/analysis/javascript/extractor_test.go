package javascript

import (
	"context"
	"testing"

	sitter "github.com/smacker/go-tree-sitter"
	"github.com/smacker/go-tree-sitter/javascript"
	"github.com/tab58/code-context/internal/analysis"
)

// === Task 2: JavaScript extractor tests ===

// parseJS is a test helper that parses JavaScript source code into a tree-sitter tree.
func parseJS(t *testing.T, source []byte) *sitter.Tree {
	t.Helper()
	parser := sitter.NewParser()
	parser.SetLanguage(javascript.GetLanguage())
	tree, err := parser.ParseCtx(context.Background(), nil, source)
	if err != nil {
		t.Fatalf("failed to parse JavaScript source: %v", err)
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

// findFuncNode recursively searches for a function_declaration or method_definition.
func findFuncNode(node *sitter.Node) *sitter.Node {
	if node.Type() == "function_declaration" || node.Type() == "method_definition" {
		return node
	}
	for i := 0; i < int(node.NamedChildCount()); i++ {
		if found := findFuncNode(node.NamedChild(i)); found != nil {
			return found
		}
	}
	return nil
}

// parseJSFunc parses JavaScript source and returns the first function node.
func parseJSFunc(t *testing.T, source string) *sitter.Node {
	t.Helper()
	parser := sitter.NewParser()
	parser.SetLanguage(javascript.GetLanguage())
	tree, err := parser.ParseCtx(context.Background(), nil, []byte(source))
	if err != nil {
		t.Fatalf("parse failed: %v", err)
	}
	return findFuncNode(tree.RootNode())
}

// --- Interface compliance ---

// TestJSExtractor_ImplementsExtractor verifies that JSExtractor satisfies
// the analysis.Extractor interface.
// Expected result: Compiles without errors.
func TestJSExtractor_ImplementsExtractor(t *testing.T) {
	var _ analysis.Extractor = &JSExtractor{}
}

// TestJSComplexityExtractor_ImplementsComplexityExtractor verifies that
// JSComplexityExtractor satisfies the analysis.ComplexityExtractor interface.
// Expected result: Compiles without errors.
func TestJSComplexityExtractor_ImplementsComplexityExtractor(t *testing.T) {
	var _ analysis.ComplexityExtractor = &JSComplexityExtractor{}
}

// --- Register pattern ---

// TestJSRegister_RegistersExtractorAndComplexity verifies that Register
// wires both Extractor and ComplexityExtractor into the registry.
// Expected result: Both extractor types are retrievable for "javascript".
func TestJSRegister_RegistersExtractorAndComplexity(t *testing.T) {
	r := analysis.NewRegistry()
	Register(r)

	ext, ok := r.ExtractorForLanguage("javascript")
	if !ok {
		t.Fatal("ExtractorForLanguage(javascript) returned false after Register")
	}
	if ext == nil {
		t.Fatal("ExtractorForLanguage(javascript) returned nil after Register")
	}

	ce, ok := r.ComplexityExtractorForLanguage("javascript")
	if !ok {
		t.Fatal("ComplexityExtractorForLanguage(javascript) returned false after Register")
	}
	if ce == nil {
		t.Fatal("ComplexityExtractorForLanguage(javascript) returned nil after Register")
	}
}

// --- Symbol extraction ---

// TestJSExtractor_ExtractsFunctionDeclaration verifies extraction of top-level
// function declarations in JavaScript.
// Expected result: Symbol with Kind="function", correct name, language="javascript".
func TestJSExtractor_ExtractsFunctionDeclaration(t *testing.T) {
	source := []byte(`function greet(name) {
  return "Hello, " + name;
}
`)
	tree := parseJS(t, source)
	ext := NewJSExtractor()

	symbols, err := ext.ExtractSymbols(tree, source, "app.js", "/repo")
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
	if sym.Language != "javascript" {
		t.Errorf("Language = %q, want %q", sym.Language, "javascript")
	}
}

// TestJSExtractor_ExtractsClassDeclaration verifies extraction of class declarations.
// Expected result: Symbol with Kind="class", correct name.
func TestJSExtractor_ExtractsClassDeclaration(t *testing.T) {
	source := []byte(`class UserService {
  constructor(db) {
    this.db = db;
  }
  getUser(id) {
    return this.db.find(id);
  }
}
`)
	tree := parseJS(t, source)
	ext := NewJSExtractor()

	symbols, err := ext.ExtractSymbols(tree, source, "service.js", "/repo")
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

// TestJSExtractor_ExtractsClassMethods verifies extraction of methods inside classes.
// Expected result: Method symbols with Kind="method" and correct ParentName.
func TestJSExtractor_ExtractsClassMethods(t *testing.T) {
	source := []byte(`class Calc {
  add(a, b) { return a + b; }
  sub(a, b) { return a - b; }
}
`)
	tree := parseJS(t, source)
	ext := NewJSExtractor()

	symbols, err := ext.ExtractSymbols(tree, source, "calc.js", "/repo")
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

// TestJSExtractor_EmitsModulePerFile verifies that one Module symbol is emitted
// per file with ESM/CJS kind detection.
// Expected result: A module symbol with Kind="module" and correct ModuleKind.
func TestJSExtractor_EmitsModulePerFile(t *testing.T) {
	source := []byte(`import { foo } from './utils';
export function bar() {}
`)
	tree := parseJS(t, source)
	ext := NewJSExtractor()

	symbols, err := ext.ExtractSymbols(tree, source, "/repo/src/app.js", "/repo")
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
	if moduleSym.ModuleKind != "esm" {
		t.Errorf("ModuleKind = %q, want %q", moduleSym.ModuleKind, "esm")
	}
}

// TestJSExtractor_CJSModule verifies CJS module detection for CommonJS files.
// Expected result: Module symbol with ModuleKind="cjs".
func TestJSExtractor_CJSModule(t *testing.T) {
	source := []byte(`const express = require('express');
function handler(req, res) { res.send('ok'); }
module.exports = handler;
`)
	tree := parseJS(t, source)
	ext := NewJSExtractor()

	symbols, err := ext.ExtractSymbols(tree, source, "/repo/server.js", "/repo")
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
		t.Fatal("expected to find module symbol")
	}
	if moduleSym.ModuleKind != "cjs" {
		t.Errorf("ModuleKind = %q, want %q", moduleSym.ModuleKind, "cjs")
	}
}

// TestJSExtractor_ExportedVisibility verifies that exported symbols have public visibility.
// Expected result: Exported function has Visibility="public", non-exported has "private".
func TestJSExtractor_ExportedVisibility(t *testing.T) {
	source := []byte(`export function publicFunc() {}
function privateFunc() {}
`)
	tree := parseJS(t, source)
	ext := NewJSExtractor()

	symbols, err := ext.ExtractSymbols(tree, source, "mod.js", "/repo")
	if err != nil {
		t.Fatalf("ExtractSymbols error: %v", err)
	}

	pubSym := findSymbol(symbols, "publicFunc")
	if pubSym == nil {
		t.Fatal("expected to find 'publicFunc'")
	}
	if pubSym.Visibility != "public" {
		t.Errorf("publicFunc Visibility = %q, want %q", pubSym.Visibility, "public")
	}

	privSym := findSymbol(symbols, "privateFunc")
	if privSym == nil {
		t.Fatal("expected to find 'privateFunc'")
	}
	if privSym.Visibility != "private" {
		t.Errorf("privateFunc Visibility = %q, want %q", privSym.Visibility, "private")
	}
}

// --- Reference extraction ---

// TestJSExtractor_ExtractsImportReferences verifies extraction of import references.
// Expected result: Reference with Kind="imports" for relative imports.
func TestJSExtractor_ExtractsImportReferences(t *testing.T) {
	source := []byte(`import { helper } from './utils';

function main() {
  helper();
}
`)
	tree := parseJS(t, source)
	ext := NewJSExtractor()

	refs, err := ext.ExtractReferences(tree, source, "app.js", "/repo")
	if err != nil {
		t.Fatalf("ExtractReferences error: %v", err)
	}

	importRef := findRef(refs, "helper")
	if importRef == nil {
		t.Fatal("expected to find import reference for 'helper'")
	}
	if importRef.Kind != "imports" {
		t.Errorf("Kind = %q, want %q", importRef.Kind, "imports")
	}
}

// TestJSExtractor_ExtractsCallReferences verifies extraction of function call references.
// Expected result: Reference with Kind="calls" for function calls.
func TestJSExtractor_ExtractsCallReferences(t *testing.T) {
	source := []byte(`function caller() {
  doWork();
}
`)
	tree := parseJS(t, source)
	ext := NewJSExtractor()

	refs, err := ext.ExtractReferences(tree, source, "app.js", "/repo")
	if err != nil {
		t.Fatalf("ExtractReferences error: %v", err)
	}

	callRef := findRef(refs, "doWork")
	if callRef == nil {
		t.Fatal("expected to find call reference for 'doWork'")
	}
	if callRef.Kind != "calls" {
		t.Errorf("Kind = %q, want %q", callRef.Kind, "calls")
	}
}

// TestJSExtractor_NonRelativeImportIsExternal verifies that non-relative
// imports are marked as external references.
// Expected result: Import reference has IsExternal=true.
func TestJSExtractor_NonRelativeImportIsExternal(t *testing.T) {
	source := []byte(`import express from 'express';
function app() {}
`)
	tree := parseJS(t, source)
	ext := NewJSExtractor()

	refs, err := ext.ExtractReferences(tree, source, "app.js", "/repo")
	if err != nil {
		t.Fatalf("ExtractReferences error: %v", err)
	}

	var expressRef *analysis.Reference
	for i, ref := range refs {
		if ref.ExternalImportPath == "express" || ref.ToName == "express" {
			expressRef = &refs[i]
			break
		}
	}
	if expressRef == nil {
		t.Fatal("expected to find import reference for 'express'")
	}
	if !expressRef.IsExternal {
		t.Error("non-relative import 'express' should have IsExternal=true")
	}
}

// --- Complexity ---

// TestJSComplexityExtractor_EmptyFunction verifies base complexity of 1
// for a function with no decision points.
// Expected result: Complexity = 1.
func TestJSComplexityExtractor_EmptyFunction(t *testing.T) {
	source := `function empty() {}`
	node := parseJSFunc(t, source)
	if node == nil {
		t.Fatal("no function node found")
	}
	ext := NewJSComplexityExtractor()
	result := ext.ComputeComplexity(node, []byte(source))
	if result != 1 {
		t.Errorf("empty function complexity = %d, want 1", result)
	}
}

// TestJSComplexityExtractor_IfStatement verifies +1 for if_statement.
// Expected result: Complexity = 2 (base 1 + 1 if).
func TestJSComplexityExtractor_IfStatement(t *testing.T) {
	source := `function withIf(x) {
  if (x > 0) { return true; }
}`
	node := parseJSFunc(t, source)
	if node == nil {
		t.Fatal("no function node found")
	}
	ext := NewJSComplexityExtractor()
	result := ext.ComputeComplexity(node, []byte(source))
	if result != 2 {
		t.Errorf("function with if complexity = %d, want 2", result)
	}
}

// TestJSComplexityExtractor_ForStatement verifies +1 for for_statement.
// Expected result: Complexity = 2 (base 1 + 1 for).
func TestJSComplexityExtractor_ForStatement(t *testing.T) {
	source := `function withFor() {
  for (let i = 0; i < 10; i++) {}
}`
	node := parseJSFunc(t, source)
	if node == nil {
		t.Fatal("no function node found")
	}
	ext := NewJSComplexityExtractor()
	result := ext.ComputeComplexity(node, []byte(source))
	if result != 2 {
		t.Errorf("function with for complexity = %d, want 2", result)
	}
}

// TestJSComplexityExtractor_WhileStatement verifies +1 for while_statement.
// Expected result: Complexity = 2 (base 1 + 1 while).
func TestJSComplexityExtractor_WhileStatement(t *testing.T) {
	source := `function withWhile() {
  while (true) { break; }
}`
	node := parseJSFunc(t, source)
	if node == nil {
		t.Fatal("no function node found")
	}
	ext := NewJSComplexityExtractor()
	result := ext.ComputeComplexity(node, []byte(source))
	if result != 2 {
		t.Errorf("function with while complexity = %d, want 2", result)
	}
}

// TestJSComplexityExtractor_LogicalAndOr verifies +1 for "&&" and "||" operators.
// Expected result: Complexity = 4 (base 1 + 1 if + 1 && + 1 ||).
func TestJSComplexityExtractor_LogicalAndOr(t *testing.T) {
	source := `function withLogical(a, b, c) {
  if (a && b || c) { return true; }
}`
	node := parseJSFunc(t, source)
	if node == nil {
		t.Fatal("no function node found")
	}
	ext := NewJSComplexityExtractor()
	result := ext.ComputeComplexity(node, []byte(source))
	if result != 4 {
		t.Errorf("function with if && || complexity = %d, want 4", result)
	}
}

// TestJSComplexityExtractor_TernaryExpression verifies +1 for ternary_expression.
// Expected result: Complexity = 2 (base 1 + 1 ternary).
func TestJSComplexityExtractor_TernaryExpression(t *testing.T) {
	source := `function withTernary(x) {
  return x > 0 ? "positive" : "negative";
}`
	node := parseJSFunc(t, source)
	if node == nil {
		t.Fatal("no function node found")
	}
	ext := NewJSComplexityExtractor()
	result := ext.ComputeComplexity(node, []byte(source))
	if result != 2 {
		t.Errorf("function with ternary complexity = %d, want 2", result)
	}
}

// TestJSComplexityExtractor_CatchClause verifies +1 for catch_clause.
// Expected result: Complexity = 2 (base 1 + 1 catch).
func TestJSComplexityExtractor_CatchClause(t *testing.T) {
	source := `function withTryCatch() {
  try { doWork(); } catch(e) { handleError(e); }
}`
	node := parseJSFunc(t, source)
	if node == nil {
		t.Fatal("no function node found")
	}
	ext := NewJSComplexityExtractor()
	result := ext.ComputeComplexity(node, []byte(source))
	if result != 2 {
		t.Errorf("function with try/catch complexity = %d, want 2", result)
	}
}
