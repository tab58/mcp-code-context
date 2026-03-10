package ruby

import (
	"context"
	"testing"

	sitter "github.com/smacker/go-tree-sitter"
	"github.com/smacker/go-tree-sitter/ruby"
	"github.com/tab58/code-context/internal/analysis"
)

// === Task 4: Ruby extractor tests ===

// parseRb is a test helper that parses Ruby source code into a tree-sitter tree.
func parseRb(t *testing.T, source []byte) *sitter.Tree {
	t.Helper()
	parser := sitter.NewParser()
	parser.SetLanguage(ruby.GetLanguage())
	tree, err := parser.ParseCtx(context.Background(), nil, source)
	if err != nil {
		t.Fatalf("failed to parse Ruby source: %v", err)
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

// findMethodNode recursively searches for a method node.
func findMethodNode(node *sitter.Node) *sitter.Node {
	if node.Type() == "method" {
		return node
	}
	for i := 0; i < int(node.NamedChildCount()); i++ {
		if found := findMethodNode(node.NamedChild(i)); found != nil {
			return found
		}
	}
	return nil
}

// parseRbMethod parses Ruby source and returns the first method node.
func parseRbMethod(t *testing.T, source string) *sitter.Node {
	t.Helper()
	parser := sitter.NewParser()
	parser.SetLanguage(ruby.GetLanguage())
	tree, err := parser.ParseCtx(context.Background(), nil, []byte(source))
	if err != nil {
		t.Fatalf("parse failed: %v", err)
	}
	return findMethodNode(tree.RootNode())
}

// --- Interface compliance ---

// TestRubyExtractor_ImplementsExtractor verifies that RubyExtractor satisfies
// the analysis.Extractor interface.
// Expected result: Compiles without errors.
func TestRubyExtractor_ImplementsExtractor(t *testing.T) {
	var _ analysis.Extractor = &RubyExtractor{}
}

// TestRubyComplexityExtractor_ImplementsComplexityExtractor verifies that
// RubyComplexityExtractor satisfies the analysis.ComplexityExtractor interface.
// Expected result: Compiles without errors.
func TestRubyComplexityExtractor_ImplementsComplexityExtractor(t *testing.T) {
	var _ analysis.ComplexityExtractor = &RubyComplexityExtractor{}
}

// --- Register pattern ---

// TestRubyRegister_RegistersExtractorAndComplexity verifies that Register
// wires both Extractor and ComplexityExtractor into the registry.
// Expected result: Both extractor types are retrievable for "ruby".
func TestRubyRegister_RegistersExtractorAndComplexity(t *testing.T) {
	r := analysis.NewRegistry()
	Register(r)

	ext, ok := r.ExtractorForLanguage("ruby")
	if !ok {
		t.Fatal("ExtractorForLanguage(ruby) returned false after Register")
	}
	if ext == nil {
		t.Fatal("ExtractorForLanguage(ruby) returned nil after Register")
	}

	ce, ok := r.ComplexityExtractorForLanguage("ruby")
	if !ok {
		t.Fatal("ComplexityExtractorForLanguage(ruby) returned false after Register")
	}
	if ce == nil {
		t.Fatal("ComplexityExtractorForLanguage(ruby) returned nil after Register")
	}
}

// --- Symbol extraction ---

// TestRubyExtractor_ExtractsMethodDefinition verifies extraction of top-level
// method definitions in Ruby.
// Expected result: Symbol with Kind="function", correct name, language="ruby".
func TestRubyExtractor_ExtractsMethodDefinition(t *testing.T) {
	source := []byte(`def greet(name)
  "Hello, #{name}"
end
`)
	tree := parseRb(t, source)
	ext := NewRubyExtractor()

	symbols, err := ext.ExtractSymbols(tree, source, "app.rb", "/repo")
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
	if sym.Language != "ruby" {
		t.Errorf("Language = %q, want %q", sym.Language, "ruby")
	}
}

// TestRubyExtractor_ExtractsClassDefinition verifies extraction of class definitions.
// Expected result: Symbol with Kind="class", correct name.
func TestRubyExtractor_ExtractsClassDefinition(t *testing.T) {
	source := []byte(`class UserService
  def initialize(db)
    @db = db
  end

  def get_user(id)
    @db.find(id)
  end
end
`)
	tree := parseRb(t, source)
	ext := NewRubyExtractor()

	symbols, err := ext.ExtractSymbols(tree, source, "service.rb", "/repo")
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

// TestRubyExtractor_ExtractsClassMethods verifies extraction of methods inside classes.
// Expected result: Method symbols with Kind="method" and correct ParentName.
func TestRubyExtractor_ExtractsClassMethods(t *testing.T) {
	source := []byte(`class Calc
  def add(a, b)
    a + b
  end

  def sub(a, b)
    a - b
  end
end
`)
	tree := parseRb(t, source)
	ext := NewRubyExtractor()

	symbols, err := ext.ExtractSymbols(tree, source, "calc.rb", "/repo")
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

// TestRubyExtractor_ExtractsModuleDefinition verifies extraction of Ruby modules.
// Expected result: Symbol with Kind="module", correct name.
func TestRubyExtractor_ExtractsModuleDefinition(t *testing.T) {
	source := []byte(`module Helpers
  def self.format(s)
    s.strip
  end
end
`)
	tree := parseRb(t, source)
	ext := NewRubyExtractor()

	symbols, err := ext.ExtractSymbols(tree, source, "helpers.rb", "/repo")
	if err != nil {
		t.Fatalf("ExtractSymbols error: %v", err)
	}

	sym := findSymbol(symbols, "Helpers")
	if sym == nil {
		t.Fatal("expected to find symbol 'Helpers', got nil")
	}
	if sym.Kind != "module" {
		t.Errorf("Kind = %q, want %q", sym.Kind, "module")
	}
}

// TestRubyExtractor_VisibilityModifiers verifies that Ruby visibility modifiers
// (public/private/protected) are correctly applied.
// Expected result: Methods after "private" keyword have Visibility="private".
func TestRubyExtractor_VisibilityModifiers(t *testing.T) {
	source := []byte(`class MyClass
  def public_method
    42
  end

  private

  def private_method
    "secret"
  end
end
`)
	tree := parseRb(t, source)
	ext := NewRubyExtractor()

	symbols, err := ext.ExtractSymbols(tree, source, "my_class.rb", "/repo")
	if err != nil {
		t.Fatalf("ExtractSymbols error: %v", err)
	}

	pubSym := findSymbol(symbols, "public_method")
	if pubSym == nil {
		t.Fatal("expected to find 'public_method'")
	}
	if pubSym.Visibility != "public" {
		t.Errorf("public_method Visibility = %q, want %q", pubSym.Visibility, "public")
	}

	privSym := findSymbol(symbols, "private_method")
	if privSym == nil {
		t.Fatal("expected to find 'private_method'")
	}
	if privSym.Visibility != "private" {
		t.Errorf("private_method Visibility = %q, want %q", privSym.Visibility, "private")
	}
}

// --- Reference extraction ---

// TestRubyExtractor_ExtractsRequireReferences verifies extraction of require references.
// Expected result: Reference with Kind="imports" for require statements.
func TestRubyExtractor_ExtractsRequireReferences(t *testing.T) {
	source := []byte(`require 'json'
require_relative 'helper'

def main
  JSON.parse("{}")
end
`)
	tree := parseRb(t, source)
	ext := NewRubyExtractor()

	refs, err := ext.ExtractReferences(tree, source, "app.rb", "/repo")
	if err != nil {
		t.Fatalf("ExtractReferences error: %v", err)
	}

	jsonRef := findRef(refs, "json")
	if jsonRef == nil {
		t.Fatal("expected to find import reference for 'json'")
	}
	if jsonRef.Kind != "imports" {
		t.Errorf("Kind = %q, want %q", jsonRef.Kind, "imports")
	}
}

// TestRubyExtractor_RequireIsExternal verifies that require (non-relative) imports
// are marked as external.
// Expected result: Import reference for 'json' has IsExternal=true.
func TestRubyExtractor_RequireIsExternal(t *testing.T) {
	source := []byte(`require 'json'

def main
  JSON.parse("{}")
end
`)
	tree := parseRb(t, source)
	ext := NewRubyExtractor()

	refs, err := ext.ExtractReferences(tree, source, "app.rb", "/repo")
	if err != nil {
		t.Fatalf("ExtractReferences error: %v", err)
	}

	jsonRef := findRef(refs, "json")
	if jsonRef == nil {
		t.Fatal("expected to find import reference for 'json'")
	}
	if !jsonRef.IsExternal {
		t.Error("require 'json' should have IsExternal=true")
	}
}

// TestRubyExtractor_ExtractsCallReferences verifies extraction of method calls.
// Expected result: Reference with Kind="calls" for method calls.
func TestRubyExtractor_ExtractsCallReferences(t *testing.T) {
	source := []byte(`def caller
  do_work
end
`)
	tree := parseRb(t, source)
	ext := NewRubyExtractor()

	refs, err := ext.ExtractReferences(tree, source, "app.rb", "/repo")
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

// --- Complexity ---

// TestRubyComplexityExtractor_EmptyMethod verifies base complexity of 1.
// Expected result: Complexity = 1.
func TestRubyComplexityExtractor_EmptyMethod(t *testing.T) {
	source := `def empty
end
`
	node := parseRbMethod(t, source)
	if node == nil {
		t.Fatal("no method node found")
	}
	ext := NewRubyComplexityExtractor()
	result := ext.ComputeComplexity(node, []byte(source))
	if result != 1 {
		t.Errorf("empty method complexity = %d, want 1", result)
	}
}

// TestRubyComplexityExtractor_IfStatement verifies +1 for if.
// Expected result: Complexity = 2 (base 1 + 1 if).
func TestRubyComplexityExtractor_IfStatement(t *testing.T) {
	source := `def with_if(x)
  if x > 0
    true
  end
end
`
	node := parseRbMethod(t, source)
	if node == nil {
		t.Fatal("no method node found")
	}
	ext := NewRubyComplexityExtractor()
	result := ext.ComputeComplexity(node, []byte(source))
	if result != 2 {
		t.Errorf("method with if complexity = %d, want 2", result)
	}
}

// TestRubyComplexityExtractor_UnlessStatement verifies +1 for unless.
// Expected result: Complexity = 2 (base 1 + 1 unless).
func TestRubyComplexityExtractor_UnlessStatement(t *testing.T) {
	source := `def with_unless(x)
  unless x.nil?
    x.to_s
  end
end
`
	node := parseRbMethod(t, source)
	if node == nil {
		t.Fatal("no method node found")
	}
	ext := NewRubyComplexityExtractor()
	result := ext.ComputeComplexity(node, []byte(source))
	if result != 2 {
		t.Errorf("method with unless complexity = %d, want 2", result)
	}
}

// TestRubyComplexityExtractor_WhileStatement verifies +1 for while.
// Expected result: Complexity = 2 (base 1 + 1 while).
func TestRubyComplexityExtractor_WhileStatement(t *testing.T) {
	source := `def with_while
  while true
    break
  end
end
`
	node := parseRbMethod(t, source)
	if node == nil {
		t.Fatal("no method node found")
	}
	ext := NewRubyComplexityExtractor()
	result := ext.ComputeComplexity(node, []byte(source))
	if result != 2 {
		t.Errorf("method with while complexity = %d, want 2", result)
	}
}

// TestRubyComplexityExtractor_UntilStatement verifies +1 for until.
// Expected result: Complexity = 2 (base 1 + 1 until).
func TestRubyComplexityExtractor_UntilStatement(t *testing.T) {
	source := `def with_until
  until done?
    process
  end
end
`
	node := parseRbMethod(t, source)
	if node == nil {
		t.Fatal("no method node found")
	}
	ext := NewRubyComplexityExtractor()
	result := ext.ComputeComplexity(node, []byte(source))
	if result != 2 {
		t.Errorf("method with until complexity = %d, want 2", result)
	}
}

// TestRubyComplexityExtractor_WhenClause verifies +1 for when (case clause).
// Expected result: Complexity = 4 (base 1 + 3 when).
func TestRubyComplexityExtractor_WhenClause(t *testing.T) {
	source := `def with_case(x)
  case x
  when 1
    "one"
  when 2
    "two"
  when 3
    "three"
  end
end
`
	node := parseRbMethod(t, source)
	if node == nil {
		t.Fatal("no method node found")
	}
	ext := NewRubyComplexityExtractor()
	result := ext.ComputeComplexity(node, []byte(source))
	if result != 4 {
		t.Errorf("method with 3 when clauses complexity = %d, want 4", result)
	}
}

// TestRubyComplexityExtractor_RescueClause verifies +1 for rescue.
// Expected result: Complexity = 2 (base 1 + 1 rescue).
func TestRubyComplexityExtractor_RescueClause(t *testing.T) {
	source := `def with_rescue
  begin
    do_work
  rescue StandardError
    handle_error
  end
end
`
	node := parseRbMethod(t, source)
	if node == nil {
		t.Fatal("no method node found")
	}
	ext := NewRubyComplexityExtractor()
	result := ext.ComputeComplexity(node, []byte(source))
	if result != 2 {
		t.Errorf("method with rescue complexity = %d, want 2", result)
	}
}

// TestRubyComplexityExtractor_LogicalOperators verifies +1 for && and ||.
// Expected result: Complexity = 4 (base 1 + 1 if + 1 && + 1 ||).
func TestRubyComplexityExtractor_LogicalOperators(t *testing.T) {
	source := `def with_logic(a, b, c)
  if a && b || c
    true
  end
end
`
	node := parseRbMethod(t, source)
	if node == nil {
		t.Fatal("no method node found")
	}
	ext := NewRubyComplexityExtractor()
	result := ext.ComputeComplexity(node, []byte(source))
	if result != 4 {
		t.Errorf("method with if && || complexity = %d, want 4", result)
	}
}
