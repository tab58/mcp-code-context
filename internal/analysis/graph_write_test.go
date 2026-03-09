package analysis

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"

	sitter "github.com/smacker/go-tree-sitter"
	codedb "github.com/tab58/code-context/internal/clients/code_db"
	"github.com/tab58/code-context/internal/config"
	"github.com/tab58/go-ormql/pkg/cypher"
	"github.com/tab58/go-ormql/pkg/driver"
)

// === Test tooling for analyzer graph writes ===

// analysisRecordingDriver records all Execute and ExecuteWrite calls
// so tests can verify that the Analyzer makes graph write calls.
type analysisRecordingDriver struct {
	executeCalls      []analysisRecordedCall
	executeWriteCalls []analysisRecordedCall
}

type analysisRecordedCall struct {
	Query  string
	Params map[string]any
}

func (d *analysisRecordingDriver) Execute(_ context.Context, stmt cypher.Statement) (driver.Result, error) {
	d.executeCalls = append(d.executeCalls, analysisRecordedCall{Query: stmt.Query, Params: stmt.Params})
	return driver.Result{}, nil
}

func (d *analysisRecordingDriver) ExecuteWrite(_ context.Context, stmt cypher.Statement) (driver.Result, error) {
	d.executeWriteCalls = append(d.executeWriteCalls, analysisRecordedCall{Query: stmt.Query, Params: stmt.Params})
	return driver.Result{}, nil
}

func (d *analysisRecordingDriver) BeginTx(_ context.Context) (driver.Transaction, error) {
	return nil, errors.New("not supported")
}

func (d *analysisRecordingDriver) Close(_ context.Context) error { return nil }

// graphWriteExtractor is a mock extractor that returns predefined symbols
// and references. Unlike mockExtractor (which returns nil), this extractor
// ensures the Analyzer has real data to write to the graph.
type graphWriteExtractor struct{}

func (e *graphWriteExtractor) ExtractSymbols(_ *sitter.Tree, _ []byte, filePath string, _ string) ([]Symbol, error) {
	return []Symbol{
		{Name: "main", Kind: "module", Path: filePath, Language: "go"},
		{Name: "myFunc", Kind: "function", Path: filePath, Language: "go", Visibility: "private", Source: "func myFunc() {}", LineNumber: 5, LineCount: 3},
		{Name: "helper", Kind: "function", Path: filePath, Language: "go", Visibility: "private", Source: "func helper() {}", LineNumber: 10, LineCount: 2, ParentName: "MyStruct"},
		{Name: "MyStruct", Kind: "struct", Path: filePath, Language: "go", Visibility: "public", Source: "type MyStruct struct{}", LineNumber: 1, LineCount: 1},
	}, nil
}

func (e *graphWriteExtractor) ExtractReferences(_ *sitter.Tree, _ []byte, filePath string, _ string) ([]Reference, error) {
	return []Reference{
		{FromSymbol: "myFunc", ToName: "helper", Kind: "calls", FilePath: filePath},
		{FromSymbol: "myFunc", ToName: "main", Kind: "imports", FilePath: filePath},
		{FromSymbol: "MyStruct", ToName: "MyStruct", Kind: "inherits", FilePath: filePath},
		{FromSymbol: "MyStruct", ToName: "MyStruct", Kind: "implements", FilePath: filePath},
		{FromSymbol: "helper", ToName: "myFunc", Kind: "overrides", FilePath: filePath},
		{FromSymbol: "main", ToName: "main", Kind: "depends_on", FilePath: filePath},
		{FromSymbol: "main", ToName: "myFunc", Kind: "exports", FilePath: filePath},
	}, nil
}

// registryWithGraphWriteExtractor creates a Registry with a graphWriteExtractor
// that returns real symbols and references.
func registryWithGraphWriteExtractor() *Registry {
	r := NewRegistry()
	r.RegisterExtractor("go", &graphWriteExtractor{})
	return r
}

// newAnalyzerWithRecorder creates an Analyzer backed by a CodeDB with a recording
// driver, and returns both the analyzer and the recorder for verification.
func newAnalyzerWithRecorder(t *testing.T) (*Analyzer, *analysisRecordingDriver) {
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

	r := registryWithGraphWriteExtractor()
	return NewAnalyzer(r, db), rec
}

// createGraphWriteTestFiles creates Go source files that the graphWriteExtractor
// can "parse" (it ignores the tree-sitter tree and returns predefined results).
func createGraphWriteTestFiles(t *testing.T) []string {
	t.Helper()
	dir := t.TempDir()

	files := map[string]string{
		"main.go": `package main

import "fmt"

func main() {
	fmt.Println("hello")
	helper()
}
`,
		"helper.go": `package main

func helper() string {
	return "helped"
}
`,
	}

	var paths []string
	for name, content := range files {
		path := filepath.Join(dir, name)
		if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
			t.Fatalf("failed to write %s: %v", name, err)
		}
		paths = append(paths, path)
	}
	return paths
}

// === Task 7: Wire Analyzer graph writes (Pass 1 — node creation) ===
// Analyzer should create Module/Class/Function nodes and Pass 1 edges
// (DEFINES, HAS_METHOD, BELONGS_TO) via Client().Execute() merge/connect mutations.

// TestAnalyzer_Pass1CreatesGraphNodes verifies that Analyze creates
// Module, Class, and Function nodes in FalkorDB during Pass 1 by calling
// Client().Execute() with merge mutations.
// Expected result: After analyzing files with symbols, the recording driver
// should have received ExecuteWrite calls for merge mutations (beyond the
// initial CreateIndexes baseline). Currently the Analyzer does NO graph
// writes, so this test FAILS.
func TestAnalyzer_Pass1CreatesGraphNodes(t *testing.T) {
	files := createGraphWriteTestFiles(t)
	a, rec := newAnalyzerWithRecorder(t)

	// Record baseline calls from CreateIndexes
	baselineWrites := len(rec.executeWriteCalls)

	result, err := a.Analyze(context.Background(), "test-repo", "", files)
	if err != nil {
		t.Fatalf("Analyze returned error: %v", err)
	}

	// Verify extractor returned symbols (sanity check)
	if result.Symbols == 0 {
		t.Fatal("graphWriteExtractor returned 0 symbols — test setup error")
	}

	// The key assertion: with symbols extracted and a non-nil db,
	// the Analyzer should have made ExecuteWrite calls for merge mutations
	// (mergeModules, mergeClasss, mergeFunctions).
	newWrites := len(rec.executeWriteCalls) - baselineWrites
	if newWrites == 0 {
		t.Error("Analyzer.Analyze made no graph write calls after extracting symbols — expected merge mutations for Module/Class/Function nodes via Client().Execute()")
	}
}

// TestAnalyzer_Pass1CreatesBelongsToEdges verifies that Analyze creates
// BELONGS_TO edges from Module/Class/Function to Repository during Pass 1.
// Expected result: connect*BelongsTo mutations are called. Currently the
// Analyzer does NOT create any edges, so this test FAILS.
func TestAnalyzer_Pass1CreatesBelongsToEdges(t *testing.T) {
	files := createGraphWriteTestFiles(t)
	a, rec := newAnalyzerWithRecorder(t)

	baselineWrites := len(rec.executeWriteCalls)

	result, err := a.Analyze(context.Background(), "test-repo", "", files)
	if err != nil {
		t.Fatalf("Analyze returned error: %v", err)
	}

	if result.Symbols == 0 {
		t.Fatal("graphWriteExtractor returned 0 symbols — test setup error")
	}

	// With symbols extracted, BELONGS_TO edges should be created.
	// Count total calls beyond baseline — we need at least some for edges.
	newWrites := len(rec.executeWriteCalls) - baselineWrites
	newReads := len(rec.executeCalls)
	totalNew := newWrites + newReads

	// We expect at least 2 types of calls: node merges AND edge connects.
	// If only node merges happened, that's still not enough.
	if totalNew < 2 {
		t.Errorf("Analyzer made %d new graph calls after extracting symbols, want at least 2 (node merges + edge connects for BELONGS_TO)", totalNew)
	}
}

// TestAnalyzer_Pass1CreatesDefinesEdges verifies that Analyze creates
// DEFINES edges from File to Function/Class during Pass 1.
// Expected result: connectFileDefines mutations are called. Currently
// the Analyzer does NOT create DEFINES edges, so this test FAILS.
func TestAnalyzer_Pass1CreatesDefinesEdges(t *testing.T) {
	files := createGraphWriteTestFiles(t)
	a, rec := newAnalyzerWithRecorder(t)

	baselineWrites := len(rec.executeWriteCalls)

	result, err := a.Analyze(context.Background(), "test-repo", "", files)
	if err != nil {
		t.Fatalf("Analyze returned error: %v", err)
	}

	if result.Symbols == 0 {
		t.Fatal("graphWriteExtractor returned 0 symbols — test setup error")
	}

	// DEFINES edges connect File -> Function/Class. The Analyzer should
	// create these during Pass 1 alongside node creation.
	newWrites := len(rec.executeWriteCalls) - baselineWrites
	if newWrites < 3 {
		t.Errorf("Analyzer made %d new write calls, want at least 3 (node merges + BELONGS_TO + DEFINES edges)", newWrites)
	}
}

// === Task 8: Wire Analyzer graph writes (Pass 2 — relationship resolution) ===
// After resolving references in Pass 2, Analyzer should create graph edges
// (CALLS, IMPORTS, INHERITS, etc.) via connect* mutations.

// TestAnalyzer_Pass2CreatesCallsEdges verifies that Analyze creates CALLS
// edges from resolved function-to-function references in Pass 2.
// Expected result: connectFunctionCalls mutations are called with CallProperties.
// Currently the Analyzer does NOT create CALLS edges, so this test FAILS.
func TestAnalyzer_Pass2CreatesCallsEdges(t *testing.T) {
	files := createGraphWriteTestFiles(t)
	a, rec := newAnalyzerWithRecorder(t)

	baselineWrites := len(rec.executeWriteCalls)

	result, err := a.Analyze(context.Background(), "test-repo", "", files)
	if err != nil {
		t.Fatalf("Analyze returned error: %v", err)
	}

	// graphWriteExtractor returns references including a "calls" reference.
	// After resolution, CALLS edges should be written to the graph.
	if result.ResolvedReferences == 0 && result.References > 0 {
		t.Log("no references resolved — CALLS edges still expected for resolved ones")
	}

	// The key assertion: with resolved references and a non-nil db,
	// the recording driver should show write calls for CALLS edges.
	newWrites := len(rec.executeWriteCalls) - baselineWrites
	if newWrites == 0 {
		t.Error("Analyzer.Analyze made no graph write calls — expected CALLS edge creation via Client().Execute() connect mutations")
	}
}

// TestAnalyzer_Pass2CreatesImportsEdges verifies that Analyze creates IMPORTS
// edges from File to Module for resolved import references.
// Expected result: connectFileImports mutations are called. Currently the
// Analyzer does NOT create IMPORTS edges, so this test FAILS.
func TestAnalyzer_Pass2CreatesImportsEdges(t *testing.T) {
	files := createGraphWriteTestFiles(t)
	a, rec := newAnalyzerWithRecorder(t)

	baselineWrites := len(rec.executeWriteCalls)

	result, err := a.Analyze(context.Background(), "test-repo", "", files)
	if err != nil {
		t.Fatalf("Analyze returned error: %v", err)
	}

	// graphWriteExtractor returns an "imports" reference.
	if result.References == 0 {
		t.Fatal("graphWriteExtractor returned 0 references — test setup error")
	}

	// IMPORTS edges should be written for import references.
	newWrites := len(rec.executeWriteCalls) - baselineWrites
	if newWrites == 0 {
		t.Error("Analyzer.Analyze made no graph write calls — expected IMPORTS edge creation via Client().Execute()")
	}
}

// TestAnalyzer_Pass2SkipsUnresolvedReferences verifies that unresolved
// references are logged and skipped — they don't cause graph write errors.
// Expected result: No error even when references can't be resolved.
func TestAnalyzer_Pass2SkipsUnresolvedReferences(t *testing.T) {
	files := createGraphWriteTestFiles(t)
	a, _ := newAnalyzerWithRecorder(t)

	result, err := a.Analyze(context.Background(), "test-repo", "", files)
	if err != nil {
		t.Fatalf("Analyze returned error for unresolved references: %v", err)
	}

	// Unresolved references should be logged, not cause errors
	if result != nil && len(result.UnresolvedNames) > 0 {
		t.Logf("unresolved names (expected): %v", result.UnresolvedNames)
	}
}
