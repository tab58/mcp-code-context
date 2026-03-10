package golang

import (
	"context"
	"testing"

	sitter "github.com/smacker/go-tree-sitter"
	"github.com/smacker/go-tree-sitter/golang"
	"github.com/tab58/code-context/internal/analysis"
)

// --- Task 2: GoComplexityExtractor ---

// parseGoFunc parses Go source and returns the function_declaration node.
func parseGoFunc(t *testing.T, source string) *sitter.Node {
	t.Helper()
	parser := sitter.NewParser()
	parser.SetLanguage(golang.GetLanguage())
	tree, err := parser.ParseCtx(context.Background(), nil, []byte(source))
	if err != nil {
		t.Fatalf("parse failed: %v", err)
	}
	root := tree.RootNode()
	// Find first function_declaration or method_declaration
	for i := 0; i < int(root.NamedChildCount()); i++ {
		child := root.NamedChild(i)
		if child.Type() == "function_declaration" || child.Type() == "method_declaration" {
			return child
		}
	}
	t.Fatal("no function_declaration or method_declaration found")
	return nil
}

// TestGoComplexityExtractor_SatisfiesInterface verifies GoComplexityExtractor
// implements the ComplexityExtractor interface.
// Expected result: Compiles without errors.
func TestGoComplexityExtractor_SatisfiesInterface(t *testing.T) {
	var ce analysis.ComplexityExtractor = NewGoComplexityExtractor()
	if ce == nil {
		t.Fatal("NewGoComplexityExtractor() returned nil")
	}
}

// TestGoComplexityExtractor_EmptyFunction verifies base complexity of 1
// for a function with no decision points.
// Expected result: Complexity = 1.
func TestGoComplexityExtractor_EmptyFunction(t *testing.T) {
	source := `package main
func empty() {
}
`
	node := parseGoFunc(t, source)
	ext := NewGoComplexityExtractor()
	result := ext.ComputeComplexity(node, []byte(source))
	if result != 1 {
		t.Errorf("empty function complexity = %d, want 1", result)
	}
}

// TestGoComplexityExtractor_IfStatement verifies +1 for each if_statement.
// Expected result: Complexity = 2 (base 1 + 1 if).
func TestGoComplexityExtractor_IfStatement(t *testing.T) {
	source := `package main
func withIf(x int) {
	if x > 0 {
		return
	}
}
`
	node := parseGoFunc(t, source)
	ext := NewGoComplexityExtractor()
	result := ext.ComputeComplexity(node, []byte(source))
	if result != 2 {
		t.Errorf("function with if complexity = %d, want 2", result)
	}
}

// TestGoComplexityExtractor_ForStatement verifies +1 for for_statement (includes range).
// Expected result: Complexity = 2 (base 1 + 1 for).
func TestGoComplexityExtractor_ForStatement(t *testing.T) {
	source := `package main
func withFor() {
	for i := 0; i < 10; i++ {
	}
}
`
	node := parseGoFunc(t, source)
	ext := NewGoComplexityExtractor()
	result := ext.ComputeComplexity(node, []byte(source))
	if result != 2 {
		t.Errorf("function with for complexity = %d, want 2", result)
	}
}

// TestGoComplexityExtractor_SwitchCase verifies +1 for each expression_case
// (switch case clause).
// Expected result: Complexity = 4 (base 1 + 3 cases).
func TestGoComplexityExtractor_SwitchCase(t *testing.T) {
	source := `package main
func withSwitch(x int) string {
	switch x {
	case 1:
		return "one"
	case 2:
		return "two"
	case 3:
		return "three"
	}
	return ""
}
`
	node := parseGoFunc(t, source)
	ext := NewGoComplexityExtractor()
	result := ext.ComputeComplexity(node, []byte(source))
	if result != 4 {
		t.Errorf("function with 3 switch cases complexity = %d, want 4", result)
	}
}

// TestGoComplexityExtractor_SelectCase verifies +1 for each communication_case
// (select case clause).
// Expected result: Complexity = 3 (base 1 + 2 select cases).
func TestGoComplexityExtractor_SelectCase(t *testing.T) {
	source := `package main
func withSelect(ch1 chan int, ch2 chan string) {
	select {
	case v := <-ch1:
		_ = v
	case s := <-ch2:
		_ = s
	}
}
`
	node := parseGoFunc(t, source)
	ext := NewGoComplexityExtractor()
	result := ext.ComputeComplexity(node, []byte(source))
	if result != 3 {
		t.Errorf("function with 2 select cases complexity = %d, want 3", result)
	}
}

// TestGoComplexityExtractor_LogicalAnd verifies +1 for "&&" operator.
// Expected result: Complexity = 3 (base 1 + 1 if + 1 &&).
func TestGoComplexityExtractor_LogicalAnd(t *testing.T) {
	source := `package main
func withAnd(a, b bool) {
	if a && b {
		return
	}
}
`
	node := parseGoFunc(t, source)
	ext := NewGoComplexityExtractor()
	result := ext.ComputeComplexity(node, []byte(source))
	if result != 3 {
		t.Errorf("function with if && complexity = %d, want 3", result)
	}
}

// TestGoComplexityExtractor_LogicalOr verifies +1 for "||" operator.
// Expected result: Complexity = 3 (base 1 + 1 if + 1 ||).
func TestGoComplexityExtractor_LogicalOr(t *testing.T) {
	source := `package main
func withOr(a, b bool) {
	if a || b {
		return
	}
}
`
	node := parseGoFunc(t, source)
	ext := NewGoComplexityExtractor()
	result := ext.ComputeComplexity(node, []byte(source))
	if result != 3 {
		t.Errorf("function with if || complexity = %d, want 3", result)
	}
}

// TestGoComplexityExtractor_Complex verifies multiple decision points combine.
// Expected result: Complexity = 6 (base 1 + 1 for + 1 if + 1 if + 1 && + 1 ||).
func TestGoComplexityExtractor_Complex(t *testing.T) {
	source := `package main
func complex(items []int) int {
	count := 0
	for _, item := range items {
		if item > 0 {
			count++
		}
		if item > 10 && item < 100 || item == 0 {
			count += 2
		}
	}
	return count
}
`
	node := parseGoFunc(t, source)
	ext := NewGoComplexityExtractor()
	result := ext.ComputeComplexity(node, []byte(source))
	if result != 6 {
		t.Errorf("complex function complexity = %d, want 6", result)
	}
}

// TestGoComplexityExtractor_RegisterViaRegister verifies that golang.Register
// also registers the ComplexityExtractor with the registry.
// Expected result: ComplexityExtractorForLanguage("go") returns non-nil.
func TestGoComplexityExtractor_RegisterViaRegister(t *testing.T) {
	r := analysis.NewRegistry()
	Register(r)

	ext, ok := r.ComplexityExtractorForLanguage("go")
	if !ok {
		t.Fatal("ComplexityExtractorForLanguage(go) returned false after Register")
	}
	if ext == nil {
		t.Fatal("ComplexityExtractorForLanguage(go) returned nil after Register")
	}
}
