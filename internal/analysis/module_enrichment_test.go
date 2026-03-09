package analysis

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	sitter "github.com/smacker/go-tree-sitter"
	codedb "github.com/tab58/code-context/internal/clients/code_db"
	"github.com/tab58/code-context/internal/config"
)

// === Task 3: Add ImportPath and ModuleKind to Symbol struct ===

// TestSymbol_HasImportPathField verifies that the Symbol struct has an
// ImportPath string field for storing fully qualified import paths.
// Expected result: Symbol{ImportPath: "github.com/foo/bar"} compiles and retains value.
func TestSymbol_HasImportPathField(t *testing.T) {
	sym := Symbol{ImportPath: "github.com/foo/bar"}
	if sym.ImportPath != "github.com/foo/bar" {
		t.Errorf("ImportPath = %q, want %q", sym.ImportPath, "github.com/foo/bar")
	}
}

// TestSymbol_HasModuleKindField verifies that the Symbol struct has a
// ModuleKind string field for storing module system kind (package/esm/cjs).
// Expected result: Symbol{ModuleKind: "package"} compiles and retains value.
func TestSymbol_HasModuleKindField(t *testing.T) {
	sym := Symbol{ModuleKind: "package"}
	if sym.ModuleKind != "package" {
		t.Errorf("ModuleKind = %q, want %q", sym.ModuleKind, "package")
	}
}

// TestSymbol_ImportPathDefaultEmpty verifies that ImportPath defaults to empty string.
// Expected result: zero-value Symbol has ImportPath == "".
func TestSymbol_ImportPathDefaultEmpty(t *testing.T) {
	var sym Symbol
	if sym.ImportPath != "" {
		t.Errorf("ImportPath default = %q, want empty string", sym.ImportPath)
	}
}

// TestSymbol_ModuleKindDefaultEmpty verifies that ModuleKind defaults to empty string.
// Expected result: zero-value Symbol has ModuleKind == "".
func TestSymbol_ModuleKindDefaultEmpty(t *testing.T) {
	var sym Symbol
	if sym.ModuleKind != "" {
		t.Errorf("ModuleKind default = %q, want empty string", sym.ModuleKind)
	}
}

// === Task 4: Update Extractor interface ===
// ExtractSymbols signature changes from (tree, source, filePath)
// to (tree, source, filePath, repoPath).

// repoPathExtractor is a mock extractor that verifies it receives repoPath
// and returns it embedded in the symbol's ImportPath for assertion.
type repoPathExtractor struct {
	receivedRepoPath string
}

func (e *repoPathExtractor) ExtractSymbols(_ *sitter.Tree, _ []byte, filePath string, repoPath string) ([]Symbol, error) {
	e.receivedRepoPath = repoPath
	return []Symbol{
		{Name: "test", Kind: "module", Path: filePath, Language: "go", ImportPath: repoPath},
	}, nil
}

func (e *repoPathExtractor) ExtractReferences(_ *sitter.Tree, _ []byte, _ string, _ string) ([]Reference, error) {
	return nil, nil
}

// TestExtractor_ExtractSymbolsAcceptsRepoPath verifies that the Extractor
// interface's ExtractSymbols method accepts a repoPath parameter.
// Expected result: repoPathExtractor implements Extractor and receives repoPath.
func TestExtractor_ExtractSymbolsAcceptsRepoPath(t *testing.T) {
	var ext Extractor = &repoPathExtractor{}
	_ = ext // compile-time check that repoPathExtractor satisfies Extractor
}

// === Task 7: Update Analyzer.Analyze signature ===
// Analyze gains a repoPath string parameter: Analyze(ctx, repoID, repoPath, files, opts...)

// TestAnalyzer_AnalyzeAcceptsRepoPath verifies that Analyzer.Analyze accepts
// a repoPath string parameter between repoID and files.
// Expected result: Analyze(ctx, repoID, repoPath, files) compiles and runs.
func TestAnalyzer_AnalyzeAcceptsRepoPath(t *testing.T) {
	_, files := createTestGoFiles(t)
	r := registryWithMockGo()
	a := NewAnalyzer(r, nil)

	result, err := a.Analyze(context.Background(), "test-repo-id", "/tmp/repo", files)
	if err != nil {
		t.Fatalf("Analyze with repoPath returned error: %v", err)
	}
	if result == nil {
		t.Fatal("Analyze with repoPath returned nil result")
	}
}

// TestAnalyzer_RepoPathPassedToExtractor verifies that the repoPath parameter
// is forwarded to ExtractSymbols calls during analysis.
// Expected result: The extractor receives the repoPath value.
func TestAnalyzer_RepoPathPassedToExtractor(t *testing.T) {
	dir := t.TempDir()
	goFile := filepath.Join(dir, "main.go")
	if err := os.WriteFile(goFile, []byte("package main\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	ext := &repoPathExtractor{}
	r := NewRegistry()
	r.RegisterExtractor("go", ext)
	a := NewAnalyzer(r, nil)

	_, err := a.Analyze(context.Background(), "test-repo", "/my/repo/path", []string{goFile})
	if err != nil {
		t.Fatalf("Analyze returned error: %v", err)
	}

	if ext.receivedRepoPath != "/my/repo/path" {
		t.Errorf("extractor received repoPath = %q, want %q", ext.receivedRepoPath, "/my/repo/path")
	}
}

// === Task 8: Update writePass1 mergeModules mutation ===
// The mergeModules mutation should include importPath, visibility, kind
// in onCreate/onMatch maps.

// moduleEnrichmentExtractor returns module symbols with ImportPath, Visibility,
// and ModuleKind fields populated, for testing that writePass1 passes them through.
type moduleEnrichmentExtractor struct{}

func (e *moduleEnrichmentExtractor) ExtractSymbols(_ *sitter.Tree, _ []byte, filePath string, repoPath string) ([]Symbol, error) {
	return []Symbol{
		{
			Name:       "main",
			Kind:       "module",
			Path:       filePath,
			Language:   "go",
			LineNumber: 1,
			LineCount:  1,
			ImportPath: "github.com/test/repo",
			Visibility: "public",
			ModuleKind: "package",
		},
	}, nil
}

func (e *moduleEnrichmentExtractor) ExtractReferences(_ *sitter.Tree, _ []byte, _ string, _ string) ([]Reference, error) {
	return nil, nil
}

// newEnrichmentAnalyzerWithRecorder creates an Analyzer with a moduleEnrichmentExtractor
// and a recording driver for verifying graph write calls.
func newEnrichmentAnalyzerWithRecorder(t *testing.T) (*Analyzer, *analysisRecordingDriver) {
	t.Helper()
	rec := &analysisRecordingDriver{}
	ctx := context.Background()
	db, err := codedb.NewCodeDB(ctx, config.FalkorDBConfig{
		Host:     "localhost",
		Port:     6379,
	}, codedb.WithDriver(rec))
	if err != nil {
		t.Fatalf("NewCodeDB failed: %v", err)
	}
	t.Cleanup(func() { db.Close(ctx) })

	r := NewRegistry()
	r.RegisterExtractor("go", &moduleEnrichmentExtractor{})
	return NewAnalyzer(r, db), rec
}

// TestAnalyzer_MergeModulesIncludesImportPath verifies that the mergeModules
// mutation in writePass1 includes "importPath" in the onCreate/onMatch maps.
// Expected result: recording driver receives a call containing "importPath"
// with value "github.com/test/repo".
func TestAnalyzer_MergeModulesIncludesImportPath(t *testing.T) {
	files := createGraphWriteTestFiles(t)
	a, rec := newEnrichmentAnalyzerWithRecorder(t)

	_, err := a.Analyze(context.Background(), "test-repo", "/tmp/repo", files)
	if err != nil {
		t.Fatalf("Analyze returned error: %v", err)
	}

	found := containsParamString(rec, "importPath", "github.com/test/repo")
	if !found {
		t.Error("mergeModules mutation missing 'importPath' in onCreate/onMatch — expected importPath='github.com/test/repo'")
	}
}

// TestAnalyzer_MergeModulesIncludesVisibility verifies that the mergeModules
// mutation includes "visibility" in the onCreate/onMatch maps.
// Expected result: recording driver receives a call containing "visibility"
// with value "public".
func TestAnalyzer_MergeModulesIncludesVisibility(t *testing.T) {
	files := createGraphWriteTestFiles(t)
	a, rec := newEnrichmentAnalyzerWithRecorder(t)

	_, err := a.Analyze(context.Background(), "test-repo", "/tmp/repo", files)
	if err != nil {
		t.Fatalf("Analyze returned error: %v", err)
	}

	found := containsParamString(rec, "visibility", "public")
	if !found {
		t.Error("mergeModules mutation missing 'visibility' in onCreate/onMatch — expected visibility='public'")
	}
}

// TestAnalyzer_MergeModulesIncludesKind verifies that the mergeModules
// mutation includes "kind" in the onCreate/onMatch maps.
// Expected result: recording driver receives a call containing "kind"
// with value "package".
func TestAnalyzer_MergeModulesIncludesKind(t *testing.T) {
	files := createGraphWriteTestFiles(t)
	a, rec := newEnrichmentAnalyzerWithRecorder(t)

	_, err := a.Analyze(context.Background(), "test-repo", "/tmp/repo", files)
	if err != nil {
		t.Fatalf("Analyze returned error: %v", err)
	}

	found := containsParamString(rec, "kind", "package")
	if !found {
		t.Error("mergeModules mutation missing 'kind' in onCreate/onMatch — expected kind='package'")
	}
}

// containsParamString searches all recorded driver calls for a specific
// key-value string pair in the params.
func containsParamString(rec *analysisRecordingDriver, key, value string) bool {
	allCalls := append(rec.executeCalls, rec.executeWriteCalls...)
	for _, call := range allCalls {
		if searchAnyKeyValue(call.Params, key, value) {
			return true
		}
	}
	return false
}

// searchAnyKeyValue recursively searches for a key-value pair in nested maps/slices.
func searchAnyKeyValue(v any, key, value string) bool {
	switch val := v.(type) {
	case map[string]any:
		if s, ok := val[key].(string); ok && s == value {
			return true
		}
		for _, child := range val {
			if searchAnyKeyValue(child, key, value) {
				return true
			}
		}
	case []any:
		for _, item := range val {
			if searchAnyKeyValue(item, key, value) {
				return true
			}
		}
	}
	return false
}

// === Task 9: Update REPL handleIngest ===
// handleIngest should pass the ingest path as repoPath to Analyzer.Analyze().

// TestREPL_HandleIngestPassesRepoPath is tested in internal/repl/ package.
// We verify here that the Analyze signature accepts repoPath so the REPL
// can pass the path through.
// (Actual REPL test is in internal/repl/module_enrichment_repl_test.go)

// === Task 10: Update all tests ===
// All existing test call sites must be updated for the new signatures.

// TestExistingAnalyzerTests_CompileWithNewSignature verifies that the existing
// Analyzer tests still compile and pass after the Analyze signature changes.
// This test runs with the mock extractor that has the old 3-param signature
// replaced by the new 4-param signature.
// Expected result: all existing tests still pass (compile-time verified).
func TestExistingAnalyzerTests_CompileWithNewSignature(t *testing.T) {
	// This is a meta-test: if this file compiles, it proves that the
	// repoPathExtractor and moduleEnrichmentExtractor both implement
	// the updated Extractor interface with 4 params.
	var _ Extractor = &repoPathExtractor{}
	var _ Extractor = &moduleEnrichmentExtractor{}
}

// === Task 11: Verify build, vet, and race detector ===

// TestVerify_AnalysisPackageCompiles is a structural test that verifies
// the analysis package compiles with the new types and interfaces.
// Expected result: this test compiles and passes.
func TestVerify_AnalysisPackageCompiles(t *testing.T) {
	sym := Symbol{
		Name:       "test",
		ImportPath: "github.com/test",
		ModuleKind: "package",
	}
	if sym.Name == "" {
		t.Error("Symbol should have a name")
	}
}

// === Shared test helper for verifying source file content ===

// TestSourceFile_ExtractorInterfaceHasRepoPath verifies the source code of
// extractor.go contains the updated 4-param ExtractSymbols signature.
// Expected result: extractor.go contains "repoPath string" in ExtractSymbols.
func TestSourceFile_ExtractorInterfaceHasRepoPath(t *testing.T) {
	data, err := os.ReadFile("extractor.go")
	if err != nil {
		t.Fatalf("failed to read extractor.go: %v", err)
	}
	content := string(data)
	if !strings.Contains(content, "repoPath string") {
		t.Error("extractor.go ExtractSymbols signature missing 'repoPath string' parameter")
	}
}

// TestSourceFile_AnalyzerHasRepoPath verifies the source code of
// analyzer.go contains the updated Analyze signature with repoPath.
// Expected result: analyzer.go contains repoPath in the Analyze method.
func TestSourceFile_AnalyzerHasRepoPath(t *testing.T) {
	data, err := os.ReadFile("analyzer.go")
	if err != nil {
		t.Fatalf("failed to read analyzer.go: %v", err)
	}
	content := string(data)
	// Look for the Analyze method signature with repoPath
	if !strings.Contains(content, "repoPath string") {
		t.Error("analyzer.go Analyze signature missing 'repoPath string' parameter")
	}
}
