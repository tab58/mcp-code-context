package typescript

import (
	"context"
	"testing"

	sitter "github.com/smacker/go-tree-sitter"
	"github.com/smacker/go-tree-sitter/typescript/typescript"
	"github.com/tab58/code-context/internal/analysis"
)

// --- Task 3: TSComplexityExtractor ---

// parseTSFunc parses TypeScript source and returns the first function_declaration
// or method_definition node.
func parseTSFunc(t *testing.T, source string) *sitter.Node {
	t.Helper()
	parser := sitter.NewParser()
	parser.SetLanguage(typescript.GetLanguage())
	tree, err := parser.ParseCtx(context.Background(), nil, []byte(source))
	if err != nil {
		t.Fatalf("parse failed: %v", err)
	}
	root := tree.RootNode()
	return findFuncNode(root)
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

// TestTSComplexityExtractor_SatisfiesInterface verifies TSComplexityExtractor
// implements the ComplexityExtractor interface.
// Expected result: Compiles without errors.
func TestTSComplexityExtractor_SatisfiesInterface(t *testing.T) {
	var ce analysis.ComplexityExtractor = NewTSComplexityExtractor()
	if ce == nil {
		t.Fatal("NewTSComplexityExtractor() returned nil")
	}
}

// TestTSComplexityExtractor_EmptyFunction verifies base complexity of 1
// for a function with no decision points.
// Expected result: Complexity = 1.
func TestTSComplexityExtractor_EmptyFunction(t *testing.T) {
	source := `function empty() {}`
	node := parseTSFunc(t, source)
	ext := NewTSComplexityExtractor()
	result := ext.ComputeComplexity(node, []byte(source))
	if result != 1 {
		t.Errorf("empty function complexity = %d, want 1", result)
	}
}

// TestTSComplexityExtractor_IfStatement verifies +1 for if_statement.
// Expected result: Complexity = 2.
func TestTSComplexityExtractor_IfStatement(t *testing.T) {
	source := `function withIf(x: number) {
  if (x > 0) { return; }
}`
	node := parseTSFunc(t, source)
	ext := NewTSComplexityExtractor()
	result := ext.ComputeComplexity(node, []byte(source))
	if result != 2 {
		t.Errorf("function with if complexity = %d, want 2", result)
	}
}

// TestTSComplexityExtractor_ForStatement verifies +1 for for_statement.
// Expected result: Complexity = 2.
func TestTSComplexityExtractor_ForStatement(t *testing.T) {
	source := `function withFor() {
  for (let i = 0; i < 10; i++) {}
}`
	node := parseTSFunc(t, source)
	ext := NewTSComplexityExtractor()
	result := ext.ComputeComplexity(node, []byte(source))
	if result != 2 {
		t.Errorf("function with for complexity = %d, want 2", result)
	}
}

// TestTSComplexityExtractor_WhileStatement verifies +1 for while_statement.
// Expected result: Complexity = 2.
func TestTSComplexityExtractor_WhileStatement(t *testing.T) {
	source := `function withWhile() {
  while (true) { break; }
}`
	node := parseTSFunc(t, source)
	ext := NewTSComplexityExtractor()
	result := ext.ComputeComplexity(node, []byte(source))
	if result != 2 {
		t.Errorf("function with while complexity = %d, want 2", result)
	}
}

// TestTSComplexityExtractor_DoStatement verifies +1 for do_statement.
// Expected result: Complexity = 2.
func TestTSComplexityExtractor_DoStatement(t *testing.T) {
	source := `function withDo() {
  do {} while (false);
}`
	node := parseTSFunc(t, source)
	ext := NewTSComplexityExtractor()
	result := ext.ComputeComplexity(node, []byte(source))
	if result != 2 {
		t.Errorf("function with do complexity = %d, want 2", result)
	}
}

// TestTSComplexityExtractor_SwitchCase verifies +1 for each switch_case.
// Expected result: Complexity = 3 (base 1 + 2 cases).
func TestTSComplexityExtractor_SwitchCase(t *testing.T) {
	source := `function withSwitch(x: number) {
  switch (x) {
    case 1: return "one";
    case 2: return "two";
  }
}`
	node := parseTSFunc(t, source)
	ext := NewTSComplexityExtractor()
	result := ext.ComputeComplexity(node, []byte(source))
	if result != 3 {
		t.Errorf("function with 2 switch cases complexity = %d, want 3", result)
	}
}

// TestTSComplexityExtractor_LogicalAnd verifies +1 for "&&".
// Expected result: Complexity = 3 (base 1 + 1 if + 1 &&).
func TestTSComplexityExtractor_LogicalAnd(t *testing.T) {
	source := `function withAnd(a: boolean, b: boolean) {
  if (a && b) { return; }
}`
	node := parseTSFunc(t, source)
	ext := NewTSComplexityExtractor()
	result := ext.ComputeComplexity(node, []byte(source))
	if result != 3 {
		t.Errorf("function with if && complexity = %d, want 3", result)
	}
}

// TestTSComplexityExtractor_LogicalOr verifies +1 for "||".
// Expected result: Complexity = 3.
func TestTSComplexityExtractor_LogicalOr(t *testing.T) {
	source := `function withOr(a: boolean, b: boolean) {
  if (a || b) { return; }
}`
	node := parseTSFunc(t, source)
	ext := NewTSComplexityExtractor()
	result := ext.ComputeComplexity(node, []byte(source))
	if result != 3 {
		t.Errorf("function with if || complexity = %d, want 3", result)
	}
}

// TestTSComplexityExtractor_NullishCoalescing verifies +1 for "??".
// Expected result: Complexity = 2 (base 1 + 1 ??).
func TestTSComplexityExtractor_NullishCoalescing(t *testing.T) {
	source := `function withNullish(a: string | null) {
  const b = a ?? "default";
}`
	node := parseTSFunc(t, source)
	ext := NewTSComplexityExtractor()
	result := ext.ComputeComplexity(node, []byte(source))
	if result != 2 {
		t.Errorf("function with ?? complexity = %d, want 2", result)
	}
}

// TestTSComplexityExtractor_OptionalChain verifies +1 for optional_chain_expression.
// Expected result: Complexity = 2 (base 1 + 1 ?.).
func TestTSComplexityExtractor_OptionalChain(t *testing.T) {
	source := `function withOptional(obj: any) {
  return obj?.foo;
}`
	node := parseTSFunc(t, source)
	ext := NewTSComplexityExtractor()
	result := ext.ComputeComplexity(node, []byte(source))
	if result != 2 {
		t.Errorf("function with ?. complexity = %d, want 2", result)
	}
}

// TestTSComplexityExtractor_CatchClause verifies +1 for catch_clause.
// Expected result: Complexity = 2 (base 1 + 1 catch).
func TestTSComplexityExtractor_CatchClause(t *testing.T) {
	source := `function withCatch() {
  try { throw new Error(); } catch (e) { }
}`
	node := parseTSFunc(t, source)
	ext := NewTSComplexityExtractor()
	result := ext.ComputeComplexity(node, []byte(source))
	if result != 2 {
		t.Errorf("function with catch complexity = %d, want 2", result)
	}
}

// TestTSComplexityExtractor_TernaryExpression verifies +1 for ternary_expression.
// Expected result: Complexity = 2 (base 1 + 1 ternary).
func TestTSComplexityExtractor_TernaryExpression(t *testing.T) {
	source := `function withTernary(x: number) {
  return x > 0 ? "pos" : "neg";
}`
	node := parseTSFunc(t, source)
	ext := NewTSComplexityExtractor()
	result := ext.ComputeComplexity(node, []byte(source))
	if result != 2 {
		t.Errorf("function with ternary complexity = %d, want 2", result)
	}
}

// TestTSComplexityExtractor_Complex verifies multiple decision points combine.
// Expected result: Complexity = 5 (base 1 + 1 for + 1 if + 1 && + 1 ternary).
func TestTSComplexityExtractor_Complex(t *testing.T) {
	source := `function complex(items: number[]) {
  for (const item of items) {
    if (item > 0 && item < 100) {
      const label = item > 50 ? "big" : "small";
    }
  }
}`
	node := parseTSFunc(t, source)
	ext := NewTSComplexityExtractor()
	result := ext.ComputeComplexity(node, []byte(source))
	if result != 5 {
		t.Errorf("complex function complexity = %d, want 5", result)
	}
}

// TestTSComplexityExtractor_RegisterViaRegister verifies that typescript.Register
// also registers the ComplexityExtractor for both TS and TSX.
// Expected result: ComplexityExtractorForLanguage returns non-nil for typescript and tsx.
func TestTSComplexityExtractor_RegisterViaRegister(t *testing.T) {
	r := analysis.NewRegistry()
	Register(r)

	for _, lang := range []string{"typescript", "tsx"} {
		t.Run(lang, func(t *testing.T) {
			ext, ok := r.ComplexityExtractorForLanguage(lang)
			if !ok {
				t.Fatalf("ComplexityExtractorForLanguage(%q) returned false after Register", lang)
			}
			if ext == nil {
				t.Fatalf("ComplexityExtractorForLanguage(%q) returned nil after Register", lang)
			}
		})
	}
}
