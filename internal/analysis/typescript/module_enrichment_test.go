package typescript

import (
	"context"
	"os"
	"strings"
	"testing"

	sitter "github.com/smacker/go-tree-sitter"
	"github.com/smacker/go-tree-sitter/typescript/typescript"
	"github.com/smacker/go-tree-sitter/typescript/tsx"
)

// === Task 6: Update TypeScript/TSX extractors ===
// TypeScript and TSX extractors should emit one Module symbol per file.
// Name = filename without extension, ImportPath = relative path from repo root,
// Visibility = "public", ModuleKind = "esm" or "cjs" based on import/export keywords.

// parseTSSource is a test helper that parses TypeScript source code.
func parseTSSource(t *testing.T, source string) *sitter.Tree {
	t.Helper()
	parser := sitter.NewParser()
	parser.SetLanguage(typescript.GetLanguage())
	tree, err := parser.ParseCtx(context.Background(), nil, []byte(source))
	if err != nil {
		t.Fatalf("failed to parse TypeScript source: %v", err)
	}
	return tree
}

// parseTSXSource is a test helper that parses TSX source code.
func parseTSXSource(t *testing.T, source string) *sitter.Tree {
	t.Helper()
	parser := sitter.NewParser()
	parser.SetLanguage(tsx.GetLanguage())
	tree, err := parser.ParseCtx(context.Background(), nil, []byte(source))
	if err != nil {
		t.Fatalf("failed to parse TSX source: %v", err)
	}
	return tree
}

// TestTSExtractor_EmitsModuleSymbol verifies that the TypeScript extractor
// emits exactly one Module symbol per file.
// Expected result: symbols contain one symbol with Kind="module".
func TestTSExtractor_EmitsModuleSymbol(t *testing.T) {
	source := `export function greet(name: string): string {
  return "Hello, " + name;
}
`
	tree := parseTSSource(t, source)
	ext := NewTypeScriptExtractor()

	symbols, err := ext.ExtractSymbols(tree, []byte(source), "/repo/src/greet.ts", "/repo")
	if err != nil {
		t.Fatalf("ExtractSymbols error: %v", err)
	}

	moduleCount := 0
	for _, sym := range symbols {
		if sym.Kind == "module" {
			moduleCount++
		}
	}

	if moduleCount != 1 {
		t.Errorf("TypeScript extractor emitted %d module symbols, want 1", moduleCount)
	}
}

// TestTSExtractor_ModuleName verifies that the Module symbol's Name
// is the filename without extension.
// Expected result: Module.Name = "greet" for file "greet.ts".
func TestTSExtractor_ModuleName(t *testing.T) {
	source := `export function greet() {}`
	tree := parseTSSource(t, source)
	ext := NewTypeScriptExtractor()

	symbols, err := ext.ExtractSymbols(tree, []byte(source), "/repo/src/greet.ts", "/repo")
	if err != nil {
		t.Fatalf("ExtractSymbols error: %v", err)
	}

	var moduleName string
	for _, sym := range symbols {
		if sym.Kind == "module" {
			moduleName = sym.Name
			break
		}
	}

	if moduleName != "greet" {
		t.Errorf("Module.Name = %q, want %q", moduleName, "greet")
	}
}

// TestTSExtractor_ModuleImportPath verifies that the Module symbol's ImportPath
// is the relative file path from the repo root.
// Expected result: Module.ImportPath = "src/greet.ts" for file "/repo/src/greet.ts".
func TestTSExtractor_ModuleImportPath(t *testing.T) {
	source := `export function greet() {}`
	tree := parseTSSource(t, source)
	ext := NewTypeScriptExtractor()

	symbols, err := ext.ExtractSymbols(tree, []byte(source), "/repo/src/greet.ts", "/repo")
	if err != nil {
		t.Fatalf("ExtractSymbols error: %v", err)
	}

	var importPath string
	for _, sym := range symbols {
		if sym.Kind == "module" {
			importPath = sym.ImportPath
			break
		}
	}

	if importPath != "src/greet.ts" {
		t.Errorf("Module.ImportPath = %q, want %q", importPath, "src/greet.ts")
	}
}

// TestTSExtractor_ModuleVisibilityAlwaysPublic verifies that the Module symbol's
// Visibility is always "public" for TypeScript.
// Expected result: Module.Visibility = "public".
func TestTSExtractor_ModuleVisibilityAlwaysPublic(t *testing.T) {
	source := `function privateFunc() {}`
	tree := parseTSSource(t, source)
	ext := NewTypeScriptExtractor()

	symbols, err := ext.ExtractSymbols(tree, []byte(source), "/repo/src/helper.ts", "/repo")
	if err != nil {
		t.Fatalf("ExtractSymbols error: %v", err)
	}

	var visibility string
	for _, sym := range symbols {
		if sym.Kind == "module" {
			visibility = sym.Visibility
			break
		}
	}

	if visibility != "public" {
		t.Errorf("Module.Visibility = %q, want %q", visibility, "public")
	}
}

// TestTSExtractor_ModuleKindESM verifies that the Module symbol's ModuleKind
// is "esm" when the source contains import or export keywords.
// Expected result: Module.ModuleKind = "esm".
func TestTSExtractor_ModuleKindESM(t *testing.T) {
	source := `import { foo } from 'bar';
export function greet() {}
`
	tree := parseTSSource(t, source)
	ext := NewTypeScriptExtractor()

	symbols, err := ext.ExtractSymbols(tree, []byte(source), "/repo/src/greet.ts", "/repo")
	if err != nil {
		t.Fatalf("ExtractSymbols error: %v", err)
	}

	var moduleKind string
	for _, sym := range symbols {
		if sym.Kind == "module" {
			moduleKind = sym.ModuleKind
			break
		}
	}

	if moduleKind != "esm" {
		t.Errorf("Module.ModuleKind = %q, want %q for source with import/export", moduleKind, "esm")
	}
}

// TestTSExtractor_ModuleKindCJS verifies that the Module symbol's ModuleKind
// is "cjs" when the source has no import or export keywords.
// Expected result: Module.ModuleKind = "cjs".
func TestTSExtractor_ModuleKindCJS(t *testing.T) {
	source := `function helper() { return 42; }
const x = helper();
`
	tree := parseTSSource(t, source)
	ext := NewTypeScriptExtractor()

	symbols, err := ext.ExtractSymbols(tree, []byte(source), "/repo/src/helper.ts", "/repo")
	if err != nil {
		t.Fatalf("ExtractSymbols error: %v", err)
	}

	var moduleKind string
	for _, sym := range symbols {
		if sym.Kind == "module" {
			moduleKind = sym.ModuleKind
			break
		}
	}

	if moduleKind != "cjs" {
		t.Errorf("Module.ModuleKind = %q, want %q for source without import/export", moduleKind, "cjs")
	}
}

// TestTSXExtractor_EmitsModuleSymbol verifies that the TSX extractor also
// emits one Module symbol per file (same rules as TypeScript).
// Expected result: symbols contain one symbol with Kind="module".
func TestTSXExtractor_EmitsModuleSymbol(t *testing.T) {
	source := `export function App() { return <div>Hello</div>; }
`
	tree := parseTSXSource(t, source)
	ext := NewTSXExtractor()

	symbols, err := ext.ExtractSymbols(tree, []byte(source), "/repo/src/App.tsx", "/repo")
	if err != nil {
		t.Fatalf("ExtractSymbols error: %v", err)
	}

	moduleCount := 0
	for _, sym := range symbols {
		if sym.Kind == "module" {
			moduleCount++
		}
	}

	if moduleCount != 1 {
		t.Errorf("TSX extractor emitted %d module symbols, want 1", moduleCount)
	}
}

// TestTSXExtractor_ModuleKindESM verifies that the TSX extractor detects ESM
// module kind for JSX source files that use export.
// Expected result: Module.ModuleKind = "esm".
func TestTSXExtractor_ModuleKindESM(t *testing.T) {
	source := `export function App() { return <div>Hello</div>; }
`
	tree := parseTSXSource(t, source)
	ext := NewTSXExtractor()

	symbols, err := ext.ExtractSymbols(tree, []byte(source), "/repo/src/App.tsx", "/repo")
	if err != nil {
		t.Fatalf("ExtractSymbols error: %v", err)
	}

	var moduleKind string
	for _, sym := range symbols {
		if sym.Kind == "module" {
			moduleKind = sym.ModuleKind
			break
		}
	}

	if moduleKind != "esm" {
		t.Errorf("TSX Module.ModuleKind = %q, want %q", moduleKind, "esm")
	}
}

// TestTSExtractor_DetectModuleKindExists verifies that the TypeScript package
// contains a detectModuleKind helper function.
// Expected result: extractor.go contains "detectModuleKind".
func TestTSExtractor_DetectModuleKindExists(t *testing.T) {
	data, err := os.ReadFile("extractor.go")
	if err != nil {
		t.Fatalf("failed to read extractor.go: %v", err)
	}
	if !strings.Contains(string(data), "detectModuleKind") {
		t.Error("extractor.go missing 'detectModuleKind' helper function")
	}
}
