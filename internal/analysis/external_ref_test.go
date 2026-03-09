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

// === Task 3: Reference struct has IsExternal and ExternalImportPath fields ===

// TestReference_HasIsExternalField verifies that Reference struct has IsExternal bool.
// Expected result: Reference with IsExternal=true compiles and retains value.
func TestReference_HasIsExternalField_Behavioral(t *testing.T) {
	r := Reference{
		FromSymbol: "main",
		ToName:     "fmt",
		Kind:       "imports",
		FilePath:   "main.go",
		IsExternal: true,
	}
	if !r.IsExternal {
		t.Error("Reference.IsExternal should be true")
	}
}

// TestReference_HasExternalImportPathField verifies that Reference struct
// has ExternalImportPath string.
// Expected result: Reference with ExternalImportPath set retains value.
func TestReference_HasExternalImportPathField_Behavioral(t *testing.T) {
	r := Reference{
		FromSymbol:         "main",
		ToName:             "fmt",
		Kind:               "imports",
		FilePath:           "main.go",
		IsExternal:         true,
		ExternalImportPath: "fmt",
	}
	if r.ExternalImportPath != "fmt" {
		t.Errorf("ExternalImportPath = %q, want %q", r.ExternalImportPath, "fmt")
	}
}

// TestReference_IsExternalDefaultFalse verifies that IsExternal defaults to false.
// Expected result: Zero-value Reference has IsExternal=false.
func TestReference_IsExternalDefaultFalse(t *testing.T) {
	var r Reference
	if r.IsExternal {
		t.Error("IsExternal should default to false")
	}
}

// === Task 7: writeExternalReferences in analyzer ===

// externalRefExtractor is a mock extractor that returns references with
// IsExternal=true to test writeExternalReferences graph writes.
type externalRefExtractor struct{}

func (e *externalRefExtractor) ExtractSymbols(_ *sitter.Tree, _ []byte, filePath string, _ string) ([]Symbol, error) {
	return []Symbol{
		{Name: "main", Kind: "module", Path: filePath, Language: "go"},
		{Name: "myFunc", Kind: "function", Path: filePath, Language: "go", Visibility: "private"},
	}, nil
}

func (e *externalRefExtractor) ExtractReferences(_ *sitter.Tree, _ []byte, filePath string, _ string) ([]Reference, error) {
	return []Reference{
		// External import
		{FromSymbol: "main", ToName: "fmt", Kind: "imports", FilePath: filePath, IsExternal: true, ExternalImportPath: "fmt"},
		// External call
		{FromSymbol: "myFunc", ToName: "Println", Kind: "calls", FilePath: filePath, IsExternal: true, ExternalImportPath: "fmt"},
		// Internal call (should not trigger external ref writes)
		{FromSymbol: "myFunc", ToName: "helper", Kind: "calls", FilePath: filePath, IsExternal: false},
	}, nil
}

// newExternalRefAnalyzerWithRecorder creates an Analyzer with an
// externalRefExtractor and a recording driver.
func newExternalRefAnalyzerWithRecorder(t *testing.T) (*Analyzer, *analysisRecordingDriver) {
	t.Helper()
	rec := &analysisRecordingDriver{}
	ctx := context.Background()
	db, err := codedb.NewCodeDB(ctx, config.FalkorDBConfig{
		Host: "localhost",
		Port: 6379,
	}, codedb.WithDriver(rec))
	if err != nil {
		t.Fatalf("NewCodeDB failed: %v", err)
	}
	t.Cleanup(func() { db.Close(ctx) })

	r := NewRegistry()
	r.RegisterExtractor("go", &externalRefExtractor{})
	return NewAnalyzer(r, db), rec
}

// createExternalRefTestFiles creates Go source files for external ref tests.
func createExternalRefTestFiles(t *testing.T) []string {
	t.Helper()
	dir := t.TempDir()

	path := filepath.Join(dir, "main.go")
	if err := os.WriteFile(path, []byte("package main\n\nimport \"fmt\"\n\nfunc main() {\n\tfmt.Println(\"hello\")\n}\n"), 0o644); err != nil {
		t.Fatalf("failed to write: %v", err)
	}
	return []string{path}
}

// TestAnalyzer_WriteExternalReferencesCreatesNodes verifies that Analyze
// creates ExternalReference nodes via mergeExternalReferences mutation when
// external references are present.
// Expected result: Recording driver receives a call containing "mergeExternalReferences".
func TestAnalyzer_WriteExternalReferencesCreatesNodes(t *testing.T) {
	files := createExternalRefTestFiles(t)
	a, rec := newExternalRefAnalyzerWithRecorder(t)

	_, err := a.Analyze(context.Background(), "test-repo", "", files)
	if err != nil {
		t.Fatalf("Analyze returned error: %v", err)
	}

	found := false
	for _, call := range append(rec.executeCalls, rec.executeWriteCalls...) {
		if strings.Contains(call.Query, "mergeExternalReferences") {
			found = true
			break
		}
	}

	if !found {
		t.Error("Analyze did not call mergeExternalReferences — expected ExternalReference node creation for external imports/calls")
	}
}

// TestAnalyzer_WriteExternalReferencesCreatesImportsEdges verifies that
// IMPORTS edges from File to ExternalReference are created.
// Expected result: Recording driver receives a call containing "connectFileExternalImports"
// or similar File->ExternalReference edge mutation.
func TestAnalyzer_WriteExternalReferencesCreatesImportsEdges(t *testing.T) {
	files := createExternalRefTestFiles(t)
	a, rec := newExternalRefAnalyzerWithRecorder(t)

	_, err := a.Analyze(context.Background(), "test-repo", "", files)
	if err != nil {
		t.Fatalf("Analyze returned error: %v", err)
	}

	found := false
	for _, call := range append(rec.executeCalls, rec.executeWriteCalls...) {
		if strings.Contains(call.Query, "connectFileExternal") {
			found = true
			break
		}
	}

	if !found {
		t.Error("Analyze did not create File->ExternalReference IMPORTS edges")
	}
}

// TestAnalyzer_WriteExternalReferencesCreatesCallsEdges verifies that
// CALLS edges from Function to ExternalReference are created.
// Expected result: Recording driver receives a call containing "connectFunctionExternalCalls".
func TestAnalyzer_WriteExternalReferencesCreatesCallsEdges(t *testing.T) {
	files := createExternalRefTestFiles(t)
	a, rec := newExternalRefAnalyzerWithRecorder(t)

	_, err := a.Analyze(context.Background(), "test-repo", "", files)
	if err != nil {
		t.Fatalf("Analyze returned error: %v", err)
	}

	found := false
	for _, call := range append(rec.executeCalls, rec.executeWriteCalls...) {
		if strings.Contains(call.Query, "connectFunctionExternalCalls") {
			found = true
			break
		}
	}

	if !found {
		t.Error("Analyze did not create Function->ExternalReference CALLS edges")
	}
}

// TestAnalyzer_WriteExternalReferencesCreatesBelongsTo verifies that
// BELONGS_TO edges from ExternalReference to Repository are created.
// Expected result: Recording driver receives a call containing "connectExternalReferenceRepository".
func TestAnalyzer_WriteExternalReferencesCreatesBelongsTo(t *testing.T) {
	files := createExternalRefTestFiles(t)
	a, rec := newExternalRefAnalyzerWithRecorder(t)

	_, err := a.Analyze(context.Background(), "test-repo", "", files)
	if err != nil {
		t.Fatalf("Analyze returned error: %v", err)
	}

	found := false
	for _, call := range append(rec.executeCalls, rec.executeWriteCalls...) {
		if strings.Contains(call.Query, "connectExternalReferenceRepository") {
			found = true
			break
		}
	}

	if !found {
		t.Error("Analyze did not create ExternalReference->Repository BELONGS_TO edges")
	}
}

// === Task 8: resolvePass summary logging and external filtering ===

// TestResolvePass_SkipsExternalReferences verifies that resolvePass does
// NOT count external references as unresolved. External refs (IsExternal=true)
// should be skipped entirely during symbol table lookup.
// Expected result: External references are not in UnresolvedNames.
func TestResolvePass_SkipsExternalReferences(t *testing.T) {
	analyses := []FileAnalysis{
		{
			FilePath: "main.go",
			Language: "go",
			Symbols: []Symbol{
				{Name: "helper", Kind: "function"},
			},
			References: []Reference{
				{FromSymbol: "main", ToName: "helper", Kind: "calls", IsExternal: false},
				{FromSymbol: "main", ToName: "fmt", Kind: "imports", IsExternal: true, ExternalImportPath: "fmt"},
				{FromSymbol: "main", ToName: "Println", Kind: "calls", IsExternal: true, ExternalImportPath: "fmt"},
			},
		},
	}

	result := resolvePass(analyses)

	// "helper" should resolve. "fmt" and "Println" are external and should be skipped.
	if result.ResolvedReferences != 1 {
		t.Errorf("ResolvedReferences = %d, want 1 (only internal 'helper' should be resolved)", result.ResolvedReferences)
	}

	// External refs should NOT appear in unresolved names
	for _, name := range result.UnresolvedNames {
		if name == "fmt" || name == "Println" {
			t.Errorf("external reference %q should NOT appear in UnresolvedNames", name)
		}
	}
}

// TestResolvePass_CountsOnlyInternalReferences verifies that the total
// reference count in resolvePass only counts internal references.
// Expected result: References count excludes external refs.
func TestResolvePass_CountsOnlyInternalReferences(t *testing.T) {
	analyses := []FileAnalysis{
		{
			FilePath: "main.go",
			Language: "go",
			Symbols:  []Symbol{{Name: "helper", Kind: "function"}},
			References: []Reference{
				{FromSymbol: "main", ToName: "helper", Kind: "calls", IsExternal: false},
				{FromSymbol: "main", ToName: "fmt", Kind: "imports", IsExternal: true, ExternalImportPath: "fmt"},
			},
		},
	}

	result := resolvePass(analyses)

	// Only 1 internal reference (helper), external ref (fmt) should be excluded
	if result.References != 1 {
		t.Errorf("References = %d, want 1 (external refs should be excluded from count)", result.References)
	}
}

// === Task 9: Wire writeExternalReferences into Analyze ===

// TestAnalyzer_WriteExternalRefsCalledAfterWritePass2 verifies that
// writeExternalReferences is called during Analyze (after writePass2).
// This is a higher-level test than the individual edge tests above —
// it verifies the wiring in the Analyze method.
// Expected result: Some graph write calls contain ExternalReference-related mutations.
func TestAnalyzer_WriteExternalRefsCalledAfterWritePass2(t *testing.T) {
	files := createExternalRefTestFiles(t)
	a, rec := newExternalRefAnalyzerWithRecorder(t)

	_, err := a.Analyze(context.Background(), "test-repo", "", files)
	if err != nil {
		t.Fatalf("Analyze returned error: %v", err)
	}

	// Check that at least one ExternalReference-related mutation was called
	found := false
	for _, call := range append(rec.executeCalls, rec.executeWriteCalls...) {
		if strings.Contains(call.Query, "ExternalReference") || strings.Contains(call.Query, "externalReference") {
			found = true
			break
		}
	}

	if !found {
		t.Error("Analyze did not invoke any ExternalReference-related mutations — writeExternalReferences not wired into Analyze")
	}
}

// === Task 10: analyzeFile passes repoPath to ExtractReferences ===

// repoPathRecordingExtractor records the repoPath passed to ExtractReferences.
type repoPathRecordingExtractor struct {
	receivedRepoPath string
}

func (e *repoPathRecordingExtractor) ExtractSymbols(_ *sitter.Tree, _ []byte, filePath string, _ string) ([]Symbol, error) {
	return []Symbol{{Name: "main", Kind: "module", Path: filePath, Language: "go"}}, nil
}

func (e *repoPathRecordingExtractor) ExtractReferences(_ *sitter.Tree, _ []byte, _ string, repoPath string) ([]Reference, error) {
	e.receivedRepoPath = repoPath
	return nil, nil
}

// TestAnalyzeFile_PassesRepoPathToExtractReferences verifies that when
// Analyze calls analyzeFile, the repoPath parameter is forwarded to
// ExtractReferences (not "" or some other value).
// Expected result: The extractor receives the exact repoPath passed to Analyze.
func TestAnalyzeFile_PassesRepoPathToExtractReferences_Behavioral(t *testing.T) {
	dir := t.TempDir()
	goFile := filepath.Join(dir, "main.go")
	if err := os.WriteFile(goFile, []byte("package main\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	ext := &repoPathRecordingExtractor{}
	r := NewRegistry()
	r.RegisterExtractor("go", ext)
	a := NewAnalyzer(r, nil)

	_, err := a.Analyze(context.Background(), "test-repo", "/my/special/repo", []string{goFile})
	if err != nil {
		t.Fatalf("Analyze returned error: %v", err)
	}

	if ext.receivedRepoPath != "/my/special/repo" {
		t.Errorf("ExtractReferences received repoPath = %q, want %q", ext.receivedRepoPath, "/my/special/repo")
	}
}
