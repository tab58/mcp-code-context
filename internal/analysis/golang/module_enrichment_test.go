package golang

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	sitter "github.com/smacker/go-tree-sitter"
	"github.com/smacker/go-tree-sitter/golang"
)

// === Task 5: Update Go extractor ===
// Go extractor should populate Module symbols with ImportPath, Visibility,
// ModuleKind when repoPath is provided.

// parseGoSource is a test helper that parses Go source code into a tree-sitter tree.
func parseGoSource(t *testing.T, source string) *sitter.Tree {
	t.Helper()
	parser := sitter.NewParser()
	parser.SetLanguage(golang.GetLanguage())
	tree, err := parser.ParseCtx(context.Background(), nil, []byte(source))
	if err != nil {
		t.Fatalf("failed to parse Go source: %v", err)
	}
	return tree
}

// createGoModRepo creates a temp dir with a go.mod file and returns the path.
func createGoModRepo(t *testing.T, moduleName string) string {
	t.Helper()
	dir := t.TempDir()
	goMod := filepath.Join(dir, "go.mod")
	content := "module " + moduleName + "\n\ngo 1.25\n"
	if err := os.WriteFile(goMod, []byte(content), 0o644); err != nil {
		t.Fatalf("failed to write go.mod: %v", err)
	}
	return dir
}

// TestGoExtractor_ModuleHasImportPath verifies that the Go extractor populates
// ImportPath on module symbols using go.mod module name + relative directory.
// Expected result: Module symbol has ImportPath = "github.com/test/repo".
func TestGoExtractor_ModuleHasImportPath(t *testing.T) {
	repoPath := createGoModRepo(t, "github.com/test/repo")
	filePath := filepath.Join(repoPath, "main.go")
	source := "package main\n"

	tree := parseGoSource(t, source)
	ext := NewGoExtractor()

	symbols, err := ext.ExtractSymbols(tree, []byte(source), filePath, repoPath)
	if err != nil {
		t.Fatalf("ExtractSymbols error: %v", err)
	}

	var moduleSymbol *struct{ importPath string }
	for _, sym := range symbols {
		if sym.Kind == "module" {
			moduleSymbol = &struct{ importPath string }{sym.ImportPath}
			break
		}
	}

	if moduleSymbol == nil {
		t.Fatal("no module symbol found")
	}
	if moduleSymbol.importPath != "github.com/test/repo" {
		t.Errorf("Module.ImportPath = %q, want %q", moduleSymbol.importPath, "github.com/test/repo")
	}
}

// TestGoExtractor_ModuleImportPathSubpackage verifies that the Go extractor
// computes ImportPath as moduleName + "/" + relativeDir for files in subdirectories.
// Expected result: Module symbol has ImportPath = "github.com/test/repo/internal/pkg".
func TestGoExtractor_ModuleImportPathSubpackage(t *testing.T) {
	repoPath := createGoModRepo(t, "github.com/test/repo")
	subDir := filepath.Join(repoPath, "internal", "pkg")
	if err := os.MkdirAll(subDir, 0o755); err != nil {
		t.Fatal(err)
	}
	filePath := filepath.Join(subDir, "util.go")
	source := "package pkg\n"

	tree := parseGoSource(t, source)
	ext := NewGoExtractor()

	symbols, err := ext.ExtractSymbols(tree, []byte(source), filePath, repoPath)
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

	if importPath != "github.com/test/repo/internal/pkg" {
		t.Errorf("Module.ImportPath = %q, want %q", importPath, "github.com/test/repo/internal/pkg")
	}
}

// TestGoExtractor_ModuleVisibilityInternal verifies that the Go extractor sets
// Visibility to "internal" when the import path contains "/internal/".
// Expected result: Module.Visibility = "internal".
func TestGoExtractor_ModuleVisibilityInternal(t *testing.T) {
	repoPath := createGoModRepo(t, "github.com/test/repo")
	subDir := filepath.Join(repoPath, "internal", "pkg")
	if err := os.MkdirAll(subDir, 0o755); err != nil {
		t.Fatal(err)
	}
	filePath := filepath.Join(subDir, "util.go")
	source := "package pkg\n"

	tree := parseGoSource(t, source)
	ext := NewGoExtractor()

	symbols, err := ext.ExtractSymbols(tree, []byte(source), filePath, repoPath)
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

	if visibility != "internal" {
		t.Errorf("Module.Visibility = %q, want %q for path containing /internal/", visibility, "internal")
	}
}

// TestGoExtractor_ModuleVisibilityPublic verifies that the Go extractor sets
// Visibility to "public" when the import path does NOT contain "/internal/".
// Expected result: Module.Visibility = "public".
func TestGoExtractor_ModuleVisibilityPublic(t *testing.T) {
	repoPath := createGoModRepo(t, "github.com/test/repo")
	filePath := filepath.Join(repoPath, "main.go")
	source := "package main\n"

	tree := parseGoSource(t, source)
	ext := NewGoExtractor()

	symbols, err := ext.ExtractSymbols(tree, []byte(source), filePath, repoPath)
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
		t.Errorf("Module.Visibility = %q, want %q for root package", visibility, "public")
	}
}

// TestGoExtractor_ModuleKindIsPackage verifies that the Go extractor always
// sets ModuleKind to "package" for Go modules.
// Expected result: Module.ModuleKind = "package".
func TestGoExtractor_ModuleKindIsPackage(t *testing.T) {
	repoPath := createGoModRepo(t, "github.com/test/repo")
	filePath := filepath.Join(repoPath, "main.go")
	source := "package main\n"

	tree := parseGoSource(t, source)
	ext := NewGoExtractor()

	symbols, err := ext.ExtractSymbols(tree, []byte(source), filePath, repoPath)
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

	if moduleKind != "package" {
		t.Errorf("Module.ModuleKind = %q, want %q", moduleKind, "package")
	}
}

// TestGoExtractor_NoGoMod verifies that when there is no go.mod file,
// the extractor still returns a module symbol (graceful degradation).
// Expected result: Module symbol exists with empty ImportPath.
func TestGoExtractor_NoGoMod(t *testing.T) {
	repoPath := t.TempDir() // no go.mod
	filePath := filepath.Join(repoPath, "main.go")
	source := "package main\n"

	tree := parseGoSource(t, source)
	ext := NewGoExtractor()

	symbols, err := ext.ExtractSymbols(tree, []byte(source), filePath, repoPath)
	if err != nil {
		t.Fatalf("ExtractSymbols error: %v", err)
	}

	found := false
	for _, sym := range symbols {
		if sym.Kind == "module" {
			found = true
			// ImportPath should be empty when no go.mod exists
			if sym.ImportPath != "" {
				t.Errorf("Module.ImportPath = %q without go.mod, want empty string", sym.ImportPath)
			}
			break
		}
	}
	if !found {
		t.Error("no module symbol found — should still extract package declaration")
	}
}

// TestGoExtractor_ResolveGoModuleName verifies that the Go source file
// contains a resolveGoModuleName function.
// Expected result: extractor.go contains "resolveGoModuleName".
func TestGoExtractor_ResolveGoModuleName(t *testing.T) {
	data, err := os.ReadFile("extractor.go")
	if err != nil {
		t.Fatalf("failed to read extractor.go: %v", err)
	}
	if !strings.Contains(string(data), "resolveGoModuleName") {
		t.Error("extractor.go missing 'resolveGoModuleName' helper function")
	}
}
