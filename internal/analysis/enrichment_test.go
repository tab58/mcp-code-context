package analysis

import (
	"context"
	"testing"

	sitter "github.com/smacker/go-tree-sitter"
)

// === Task 6: Update analyzer buildFuncFields ===
//
// buildFuncFields should use "startingLine"/"endingLine" instead of
// "lineNumber"/"lineCount". endingLine = LineNumber + LineCount - 1.
// Guard: when LineCount is 0, endingLine should equal startingLine.

// TestBuildFuncFields_UsesStartingLine verifies that buildFuncFields
// produces a map with "startingLine" key instead of "lineNumber".
// Expected result: map contains "startingLine" and NOT "lineNumber".
func TestBuildFuncFields_UsesStartingLine(t *testing.T) {
	sym := Symbol{
		Name:       "myFunc",
		Kind:       "function",
		Language:   "go",
		Visibility: "public",
		LineNumber: 10,
		LineCount:  5,
	}

	fields := buildFuncFields(sym)

	if _, ok := fields["startingLine"]; !ok {
		t.Error("buildFuncFields missing 'startingLine' key — should replace 'lineNumber'")
	}
	if _, ok := fields["lineNumber"]; ok {
		t.Error("buildFuncFields still contains 'lineNumber' key — should be replaced with 'startingLine'")
	}
}

// TestBuildFuncFields_UsesEndingLine verifies that buildFuncFields
// produces a map with "endingLine" key instead of "lineCount".
// Expected result: map contains "endingLine" and NOT "lineCount".
func TestBuildFuncFields_UsesEndingLine(t *testing.T) {
	sym := Symbol{
		Name:       "myFunc",
		Kind:       "function",
		Language:   "go",
		Visibility: "public",
		LineNumber: 10,
		LineCount:  5,
	}

	fields := buildFuncFields(sym)

	if _, ok := fields["endingLine"]; !ok {
		t.Error("buildFuncFields missing 'endingLine' key — should replace 'lineCount'")
	}
	if _, ok := fields["lineCount"]; ok {
		t.Error("buildFuncFields still contains 'lineCount' key — should be replaced with 'endingLine'")
	}
}

// TestBuildFuncFields_StartingLineValue verifies that startingLine
// is set to sym.LineNumber (1-based line number).
// Expected result: startingLine == 10 for LineNumber=10.
func TestBuildFuncFields_StartingLineValue(t *testing.T) {
	sym := Symbol{
		Name:       "myFunc",
		Kind:       "function",
		Language:   "go",
		LineNumber: 10,
		LineCount:  5,
	}

	fields := buildFuncFields(sym)

	val, ok := fields["startingLine"]
	if !ok {
		t.Fatal("buildFuncFields missing 'startingLine' key")
	}
	if val != 10 {
		t.Errorf("startingLine = %v, want 10", val)
	}
}

// TestBuildFuncFields_EndingLineValue verifies that endingLine
// is computed as LineNumber + LineCount - 1.
// Expected result: endingLine == 14 for LineNumber=10, LineCount=5.
func TestBuildFuncFields_EndingLineValue(t *testing.T) {
	sym := Symbol{
		Name:       "myFunc",
		Kind:       "function",
		Language:   "go",
		LineNumber: 10,
		LineCount:  5,
	}

	fields := buildFuncFields(sym)

	val, ok := fields["endingLine"]
	if !ok {
		t.Fatal("buildFuncFields missing 'endingLine' key")
	}
	if val != 14 {
		t.Errorf("endingLine = %v, want 14 (LineNumber=10 + LineCount=5 - 1)", val)
	}
}

// TestBuildFuncFields_EndingLineGuard_ZeroLineCount verifies that when
// LineCount is 0, endingLine equals startingLine (not startingLine - 1).
// Expected result: endingLine == 10 for LineNumber=10, LineCount=0.
func TestBuildFuncFields_EndingLineGuard_ZeroLineCount(t *testing.T) {
	sym := Symbol{
		Name:       "myFunc",
		Kind:       "function",
		Language:   "go",
		LineNumber: 10,
		LineCount:  0,
	}

	fields := buildFuncFields(sym)

	startVal, ok := fields["startingLine"]
	if !ok {
		t.Fatal("buildFuncFields missing 'startingLine' key")
	}
	endVal, ok := fields["endingLine"]
	if !ok {
		t.Fatal("buildFuncFields missing 'endingLine' key")
	}

	if endVal != startVal {
		t.Errorf("when LineCount=0, endingLine should equal startingLine (%v), got %v", startVal, endVal)
	}
}

// TestBuildFuncFields_SingleLineFunction verifies endingLine for a single-line
// function (LineCount=1).
// Expected result: endingLine == startingLine for LineCount=1.
func TestBuildFuncFields_SingleLineFunction(t *testing.T) {
	sym := Symbol{
		Name:       "oneLineFn",
		Kind:       "function",
		Language:   "go",
		LineNumber: 5,
		LineCount:  1,
	}

	fields := buildFuncFields(sym)

	startVal := fields["startingLine"]
	endVal := fields["endingLine"]

	if startVal != 5 {
		t.Errorf("startingLine = %v, want 5", startVal)
	}
	if endVal != 5 {
		t.Errorf("endingLine = %v, want 5 (single line: 5 + 1 - 1 = 5)", endVal)
	}
}

// === Task 7: Update analyzer module merge — add startingLine/endingLine ===
//
// Module merge mutations in writePass1 should include "startingLine" and
// "endingLine" in the onCreate and onMatch maps.

// TestAnalyzer_ModuleMergeIncludesStartingLine verifies that when the
// Analyzer creates Module nodes in Pass 1, the merge mutation includes
// "startingLine" in the node data.
// Expected result: Module merge params contain a startingLine-like value.
func TestAnalyzer_ModuleMergeIncludesStartingLine(t *testing.T) {
	files := createGraphWriteTestFiles(t)
	a, rec := newAnalyzerWithRecorder(t)

	// Override the extractor to return a module symbol with line info
	a.registry.RegisterExtractor("go", &moduleLineExtractor{})

	_, err := a.Analyze(context.Background(), "test-repo", "", files)
	if err != nil {
		t.Fatalf("Analyze returned error: %v", err)
	}

	// Look through all driver calls for evidence that module merge included
	// a startingLine value. The module symbol has LineNumber=1, so we check
	// for that value in the params.
	found := false
	for _, call := range append(rec.executeCalls, rec.executeWriteCalls...) {
		if containsParamValue(call.Params, 1) {
			found = true
			break
		}
	}

	if !found {
		t.Error("Module merge mutation did not include startingLine — expected 'startingLine' in onCreate/onMatch for module nodes")
	}
}

// TestAnalyzer_ModuleMergeIncludesEndingLine verifies that Module merge
// mutations include "endingLine" computed from LineNumber + LineCount - 1.
// Expected result: Module merge params contain the computed endingLine value.
func TestAnalyzer_ModuleMergeIncludesEndingLine(t *testing.T) {
	files := createGraphWriteTestFiles(t)
	a, rec := newAnalyzerWithRecorder(t)

	// Override the extractor to return a module symbol with line info
	a.registry.RegisterExtractor("go", &moduleLineExtractor{})

	_, err := a.Analyze(context.Background(), "test-repo", "", files)
	if err != nil {
		t.Fatalf("Analyze returned error: %v", err)
	}

	// Module symbol: LineNumber=1, LineCount=3 -> endingLine=3
	found := false
	for _, call := range append(rec.executeCalls, rec.executeWriteCalls...) {
		if containsParamValue(call.Params, 3) {
			found = true
			break
		}
	}

	if !found {
		t.Error("Module merge mutation did not include endingLine — expected 'endingLine' in onCreate/onMatch for module nodes (LineNumber=1 + LineCount=3 - 1 = 3)")
	}
}

// === Test tooling ===

// moduleLineExtractor is a mock extractor that returns a module symbol with
// line number information to verify startingLine/endingLine on module merges.
type moduleLineExtractor struct{}

func (e *moduleLineExtractor) ExtractSymbols(_ *sitter.Tree, _ []byte, filePath string, _ string) ([]Symbol, error) {
	return []Symbol{
		{Name: "main", Kind: "module", Path: filePath, Language: "go", LineNumber: 1, LineCount: 3},
		{Name: "myFunc", Kind: "function", Path: filePath, Language: "go", Visibility: "public", Source: "func myFunc() {}", LineNumber: 5, LineCount: 2},
	}, nil
}

func (e *moduleLineExtractor) ExtractReferences(_ *sitter.Tree, _ []byte, _ string, _ string) ([]Reference, error) {
	return nil, nil
}

// containsParamValue recursively searches params for an int value.
func containsParamValue(params map[string]any, target int) bool {
	for _, v := range params {
		if searchAnyInt(v, target) {
			return true
		}
	}
	return false
}

func searchAnyInt(v any, target int) bool {
	switch val := v.(type) {
	case int:
		return val == target
	case int64:
		return int(val) == target
	case float64:
		return int(val) == target
	case map[string]any:
		return containsParamValue(val, target)
	case []any:
		for _, item := range val {
			if searchAnyInt(item, target) {
				return true
			}
		}
	}
	return false
}
