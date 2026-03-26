package mcp

import (
	"context"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/tab58/go-ormql/pkg/driver"
)

// === Task 8: Update MCP SearchResult struct ===
//
// SearchResult should use StartingLine/EndingLine instead of LineNumber/LineCount.
// These tests verify the struct definition via source inspection since the
// fields don't exist yet (compile-time safety).

// TestSearchResult_HasStartingLineField verifies that SearchResult has a
// StartingLine field (not LineNumber).
// Expected result: types.go contains "StartingLine" in the SearchResult struct.
func TestSearchResult_HasStartingLineField(t *testing.T) {
	content := readMCPTypesFile(t)
	if !strings.Contains(content, "StartingLine") {
		t.Error("SearchResult missing 'StartingLine' field — should replace 'LineNumber'")
	}
}

// TestSearchResult_HasEndingLineField verifies that SearchResult has an
// EndingLine field (not LineCount).
// Expected result: types.go contains "EndingLine" in the SearchResult struct.
func TestSearchResult_HasEndingLineField(t *testing.T) {
	content := readMCPTypesFile(t)
	if !strings.Contains(content, "EndingLine") {
		t.Error("SearchResult missing 'EndingLine' field — should replace 'LineCount'")
	}
}

// TestSearchResult_NoLineNumberField verifies that SearchResult no longer
// has a LineNumber field.
// Expected result: types.go does NOT contain "LineNumber" in the struct.
func TestSearchResult_NoLineNumberField(t *testing.T) {
	content := readMCPTypesFile(t)
	if strings.Contains(content, "LineNumber") {
		t.Error("SearchResult still contains 'LineNumber' — should be replaced with 'StartingLine'")
	}
}

// TestSearchResult_NoLineCountField verifies that SearchResult no longer
// has a LineCount field.
// Expected result: types.go does NOT contain "LineCount" in the struct.
func TestSearchResult_NoLineCountField(t *testing.T) {
	content := readMCPTypesFile(t)
	if strings.Contains(content, "LineCount") {
		t.Error("SearchResult still contains 'LineCount' — should be replaced with 'EndingLine'")
	}
}

// TestSearchResult_StartingLineJSONTag verifies that the StartingLine
// field has the correct JSON tag "startingLine".
// Expected result: types.go contains json tag "startingLine".
func TestSearchResult_StartingLineJSONTag(t *testing.T) {
	content := readMCPTypesFile(t)
	if !strings.Contains(content, `"startingLine`) {
		t.Error("SearchResult.StartingLine missing json tag 'startingLine'")
	}
}

// TestSearchResult_EndingLineJSONTag verifies that the EndingLine
// field has the correct JSON tag "endingLine".
// Expected result: types.go contains json tag "endingLine".
func TestSearchResult_EndingLineJSONTag(t *testing.T) {
	content := readMCPTypesFile(t)
	if !strings.Contains(content, `"endingLine`) {
		t.Error("SearchResult.EndingLine missing json tag 'endingLine'")
	}
}

// === Task 9: Update MCP GraphQL queries ===
//
// All function/class GraphQL queries should use startingLine/endingLine
// instead of lineNumber/lineCount.

// TestGqlFindFunctions_UsesStartingLine verifies that gqlFindFunctions
// queries "startingLine" instead of "lineNumber".
// Expected result: gqlFindFunctions query contains "startingLine".
func TestGqlFindFunctions_UsesStartingLine(t *testing.T) {
	content := readMCPToolsFile(t)
	block := extractGqlConst(content, "gqlFindFunctions")
	if !strings.Contains(block, "startingLine") {
		t.Error("gqlFindFunctions query missing 'startingLine' — should replace 'lineNumber'")
	}
}

// TestGqlFindFunctions_UsesEndingLine verifies that gqlFindFunctions
// queries "endingLine" instead of "lineCount".
// Expected result: gqlFindFunctions query contains "endingLine".
func TestGqlFindFunctions_UsesEndingLine(t *testing.T) {
	content := readMCPToolsFile(t)
	block := extractGqlConst(content, "gqlFindFunctions")
	if !strings.Contains(block, "endingLine") {
		t.Error("gqlFindFunctions query missing 'endingLine' — should replace 'lineCount'")
	}
}

// TestGqlFindFunctions_NoLineNumber verifies that gqlFindFunctions
// no longer queries "lineNumber".
// Expected result: gqlFindFunctions query does NOT contain "lineNumber".
func TestGqlFindFunctions_NoLineNumber(t *testing.T) {
	content := readMCPToolsFile(t)
	block := extractGqlConst(content, "gqlFindFunctions")
	if strings.Contains(block, "lineNumber") {
		t.Error("gqlFindFunctions query still contains 'lineNumber' — should be 'startingLine'")
	}
}

// gqlFunctionsSimilar and gqlClassesSimilar were removed with vector search.

// === Task 10: Update MCP result parsing ===
//
// handleFindFunction and parseSimilarResults should read startingLine/endingLine
// from response data and populate the corresponding SearchResult fields.

// TestFindFunction_ParsesStartingLine verifies that handleFindFunction
// reads "startingLine" from query results (not "lineNumber").
// Expected result: tools.go contains intVal(m, "startingLine") call.
func TestFindFunction_ParsesStartingLine(t *testing.T) {
	content := readMCPToolsFile(t)
	if !strings.Contains(content, `intVal(m, "startingLine")`) && !strings.Contains(content, `intVal(node, "startingLine")`) {
		t.Error("handleFindFunction/parseSimilarResults not reading 'startingLine' — should replace 'lineNumber' in intVal calls")
	}
}

// TestFindFunction_ParsesEndingLine verifies that handleFindFunction
// reads "endingLine" from query results (not "lineCount").
// Expected result: tools.go contains intVal(m, "endingLine") call.
func TestFindFunction_ParsesEndingLine(t *testing.T) {
	content := readMCPToolsFile(t)
	if !strings.Contains(content, `intVal(m, "endingLine")`) && !strings.Contains(content, `intVal(node, "endingLine")`) {
		t.Error("handleFindFunction/parseSimilarResults not reading 'endingLine' — should replace 'lineCount' in intVal calls")
	}
}

// TestFindFunction_NoLineNumberParsing verifies that handleFindFunction
// no longer reads "lineNumber" from results.
// Expected result: tools.go does NOT contain intVal(m, "lineNumber") or
// intVal(node, "lineNumber").
func TestFindFunction_NoLineNumberParsing(t *testing.T) {
	content := readMCPToolsFile(t)
	if strings.Contains(content, `intVal(m, "lineNumber")`) || strings.Contains(content, `intVal(node, "lineNumber")`) {
		t.Error("tools.go still reads 'lineNumber' — should be replaced with 'startingLine'")
	}
}

// TestFindFunction_NoLineCountParsing verifies that handleFindFunction
// no longer reads "lineCount" from Function/Class results.
// Note: intVal(m, "lineCount") in handleFindFile for File nodes is OK.
// Expected result: Function/Class parsing does NOT use intVal(m/node, "lineCount").
func TestFindFunction_NoLineCountParsing(t *testing.T) {
	content := readMCPToolsFile(t)
	// Check specifically in the parseSimilarResults function and handleFindFunction
	if strings.Contains(content, `intVal(node, "lineCount")`) {
		t.Error("parseSimilarResults still reads 'lineCount' from nodes — should be 'endingLine'")
	}
}

// TestFindFunction_BehavioralStartingLine verifies end-to-end that
// function search results contain the correct starting line from response data.
// Expected result: search result has StartingLine populated.
func TestFindFunction_BehavioralStartingLine(t *testing.T) {
	svc, _ := newTestServiceWithResponses(t, []driver.Result{
		makeResult(map[string]any{
			"functions": []any{
				map[string]any{
					"name":         "handler",
					"path":         "api/handler.go",
					"source":       "func handler() {}",
					"signature":    "func handler()",
					"language":     "go",
					"visibility":   "public",
					"startingLine": float64(42),
					"endingLine":   float64(60),
				},
			},
		}),
	})

	resp, err := svc.HandleFindFunction(context.Background(), "myrepo", "handler")
	if err != nil {
		t.Fatalf("handleFindFunction returned error: %v", err)
	}
	if len(resp.Results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(resp.Results))
	}

	r := resp.Results[0]
	if r.StartingLine != 42 {
		t.Errorf("StartingLine = %d, want 42", r.StartingLine)
	}
	if r.EndingLine != 60 {
		t.Errorf("EndingLine = %d, want 60", r.EndingLine)
	}
}

// === Helpers ===

// repoRoot returns the absolute path to the repository root.
func repoRoot(t *testing.T) string {
	t.Helper()
	_, filename, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("runtime.Caller failed")
	}
	// internal/mcp/enrichment_test.go -> go up 3 levels
	return filepath.Join(filepath.Dir(filename), "..", "..")
}

// readMCPTypesFile reads the mcp/types.go source file.
func readMCPTypesFile(t *testing.T) string {
	t.Helper()
	path := filepath.Join(repoRoot(t), "internal", "mcp", "types.go")
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("failed to read types.go: %v", err)
	}
	return string(data)
}

// readMCPToolsFile reads the mcp/tools.go source file.
func readMCPToolsFile(t *testing.T) string {
	t.Helper()
	path := filepath.Join(repoRoot(t), "internal", "mcp", "tools.go")
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("failed to read tools.go: %v", err)
	}
	return string(data)
}

// extractGqlConst extracts the value of a GraphQL constant from Go source.
func extractGqlConst(content, constName string) string {
	idx := strings.Index(content, constName)
	if idx == -1 {
		return ""
	}
	rest := content[idx:]
	start := strings.Index(rest, "`")
	if start == -1 {
		return ""
	}
	rest = rest[start+1:]
	end := strings.Index(rest, "`")
	if end == -1 {
		return ""
	}
	return rest[:end]
}
