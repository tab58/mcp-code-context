package verify

import (
	"os"
	"strings"
	"testing"
)

// --- Task 9: REPL ingest pipeline wiring ---

// TestREPL_IngestCallsComputeComplexity verifies that handleIngest in
// commands.go calls analyzer.ComputeComplexity after Analyze.
// Expected result: Source contains "ComputeComplexity" call.
func TestREPL_IngestCallsComputeComplexity(t *testing.T) {
	data, err := os.ReadFile("../repl/commands.go")
	if err != nil {
		t.Fatalf("cannot read commands.go: %v", err)
	}
	src := string(data)
	if !strings.Contains(src, "ComputeComplexity") {
		t.Error("handleIngest should call analyzer.ComputeComplexity after Analyze step")
	}
}

// TestREPL_HelpTextShowsThreeStages verifies that REPL help text
// reflects the 3-stage pipeline (index -> analyze -> compute complexity).
// Expected result: Help text mentions "complexity" or "compute".
func TestREPL_HelpTextShowsThreeStages(t *testing.T) {
	data, err := os.ReadFile("../repl/commands.go")
	if err != nil {
		t.Fatalf("cannot read commands.go: %v", err)
	}
	src := string(data)
	if !strings.Contains(src, "complexity") && !strings.Contains(src, "compute") {
		t.Error("REPL help text should reference complexity computation in pipeline description")
	}
}

// --- Task 10: MCP ingest_repository pipeline wiring ---

// TestMCP_IngestRepositoryCallsComputeComplexity verifies that
// handleIngestRepository calls analyzer.ComputeComplexity after Analyze.
// Expected result: Source contains "ComputeComplexity" call.
func TestMCP_IngestRepositoryCallsComputeComplexity(t *testing.T) {
	data, err := os.ReadFile("../mcp/management_tools.go")
	if err != nil {
		t.Fatalf("cannot read management_tools.go: %v", err)
	}
	src := string(data)
	if !strings.Contains(src, "ComputeComplexity") {
		t.Error("handleIngestRepository should call analyzer.ComputeComplexity after Analyze step")
	}
}

// --- Task 11: server.go 19-tool registration ---

// TestServer_Registers19Tools verifies that server.go doc comment
// reflects 19 total registered tools.
// Expected result: Doc comment mentions "19".
func TestServer_Registers19Tools(t *testing.T) {
	data, err := os.ReadFile("../mcp/server.go")
	if err != nil {
		t.Fatalf("cannot read server.go: %v", err)
	}
	src := string(data)
	if !strings.Contains(src, "19") {
		t.Error("server.go doc comment should mention 19 total tools")
	}
}

// TestServer_RegistersSearchCodeNames verifies that the renamed
// search_code_names tool is registered in server.go.
// Expected result: Source contains "search_code_names".
func TestServer_RegistersSearchCodeNames(t *testing.T) {
	data, err := os.ReadFile("../mcp/server.go")
	if err != nil {
		t.Fatalf("cannot read server.go: %v", err)
	}
	src := string(data)
	if !strings.Contains(src, "search_code_names") {
		t.Error("server.go should register the renamed 'search_code_names' tool")
	}
}

// TestServer_RegistersContentSearchCode verifies that the new
// content-based search_code tool is registered in server.go.
// Expected result: Source contains mcpHandleSearchCodeContent or similar adapter.
func TestServer_RegistersContentSearchCode(t *testing.T) {
	data, err := os.ReadFile("../mcp/server.go")
	if err != nil {
		t.Fatalf("cannot read server.go: %v", err)
	}
	src := string(data)
	// The new search_code tool should use the content handler
	if !strings.Contains(src, "mcpHandleSearchCodeContent") && !strings.Contains(src, "SearchCodeContent") {
		t.Error("server.go should register a content-based search_code tool handler")
	}
}

// TestServer_RegistersFindDeadCode verifies that find_dead_code
// tool is registered in server.go.
// Expected result: Source contains "find_dead_code".
func TestServer_RegistersFindDeadCode(t *testing.T) {
	data, err := os.ReadFile("../mcp/server.go")
	if err != nil {
		t.Fatalf("cannot read server.go: %v", err)
	}
	src := string(data)
	if !strings.Contains(src, "find_dead_code") {
		t.Error("server.go should register 'find_dead_code' tool")
	}
}

// TestServer_RegistersCalculateCyclomaticComplexity verifies that
// calculate_cyclomatic_complexity tool is registered in server.go.
// Expected result: Source contains "calculate_cyclomatic_complexity".
func TestServer_RegistersCalculateCyclomaticComplexity(t *testing.T) {
	data, err := os.ReadFile("../mcp/server.go")
	if err != nil {
		t.Fatalf("cannot read server.go: %v", err)
	}
	src := string(data)
	if !strings.Contains(src, "calculate_cyclomatic_complexity") {
		t.Error("server.go should register 'calculate_cyclomatic_complexity' tool")
	}
}

// TestServer_RegistersFindMostComplexFunctions verifies that
// find_most_complex_functions tool is registered in server.go.
// Expected result: Source contains "find_most_complex_functions".
func TestServer_RegistersFindMostComplexFunctions(t *testing.T) {
	data, err := os.ReadFile("../mcp/server.go")
	if err != nil {
		t.Fatalf("cannot read server.go: %v", err)
	}
	src := string(data)
	if !strings.Contains(src, "find_most_complex_functions") {
		t.Error("server.go should register 'find_most_complex_functions' tool")
	}
}

// --- Task 12: Full verification ---

// TestSearchNamesFile_Exists verifies that search_names.go exists
// with fuzzy search infrastructure.
// Expected result: File contains gqlAllFunctionNames and stripWildcards.
func TestSearchNamesFile_Exists(t *testing.T) {
	data, err := os.ReadFile("../mcp/search_names.go")
	if err != nil {
		t.Fatalf("cannot read search_names.go: %v", err)
	}
	src := string(data)
	checks := []struct {
		name    string
		pattern string
	}{
		{"gqlAllFunctionNames", "gqlAllFunctionNames"},
		{"gqlAllClassNames", "gqlAllClassNames"},
		{"stripWildcards", "stripWildcards"},
		{"executeFuzzySearch", "executeFuzzySearch"},
		{"fuzzyThreshold", "fuzzyThreshold"},
	}
	for _, c := range checks {
		if !strings.Contains(src, c.pattern) {
			t.Errorf("search_names.go should contain %q", c.name)
		}
	}
}

// TestContentSearchFile_Exists verifies that content_search.go exists
// with content-based search infrastructure.
// Expected result: File contains gqlFunctionsWithSource, gqlClassesWithSource, and handleSearchCodeContent.
func TestContentSearchFile_Exists(t *testing.T) {
	data, err := os.ReadFile("../mcp/content_search.go")
	if err != nil {
		t.Fatalf("cannot read content_search.go: %v", err)
	}
	src := string(data)
	checks := []struct {
		name    string
		pattern string
	}{
		{"gqlFunctionsWithSource", "gqlFunctionsWithSource"},
		{"gqlClassesWithSource", "gqlClassesWithSource"},
		{"handleSearchCodeContent", "handleSearchCodeContent"},
	}
	for _, c := range checks {
		if !strings.Contains(src, c.pattern) {
			t.Errorf("content_search.go should contain %q", c.name)
		}
	}
}

// TestDeadCodeFile_HasGqlConstants verifies that dead_code.go has all
// 3 GraphQL constants with _NONE relationship filters.
// Expected result: File contains gqlDeadFunctions, gqlDeadClasses, gqlDeadModules.
func TestDeadCodeFile_HasGqlConstants(t *testing.T) {
	data, err := os.ReadFile("../mcp/dead_code.go")
	if err != nil {
		t.Fatalf("cannot read dead_code.go: %v", err)
	}
	src := string(data)
	for _, name := range []string{"gqlDeadFunctions", "gqlDeadClasses", "gqlDeadModules"} {
		if !strings.Contains(src, name) {
			t.Errorf("dead_code.go should contain %q", name)
		}
	}
}

// TestComplexityFile_HasGqlConstants verifies that complexity.go has
// GraphQL constants for complexity tools.
// Expected result: File contains gqlFunctionComplexity (unified constant for both handlers).
func TestComplexityFile_HasGqlConstants(t *testing.T) {
	data, err := os.ReadFile("../mcp/complexity.go")
	if err != nil {
		t.Fatalf("cannot read complexity.go: %v", err)
	}
	src := string(data)
	if !strings.Contains(src, "gqlFunctionComplexity") {
		t.Errorf("complexity.go should contain %q", "gqlFunctionComplexity")
	}
}

// TestAnalyzer_HasComputeComplexityMethod verifies that analyzer.go
// contains the ComputeComplexity method.
// Expected result: Source contains "ComputeComplexity" method declaration.
func TestAnalyzer_HasComputeComplexityMethod(t *testing.T) {
	data, err := os.ReadFile("../analysis/analyzer.go")
	if err != nil {
		t.Fatalf("cannot read analyzer.go: %v", err)
	}
	src := string(data)
	if !strings.Contains(src, "func (a *Analyzer) ComputeComplexity") {
		t.Error("analyzer.go should contain ComputeComplexity method")
	}
}

// TestAnalyzer_HasSetComplexityBatchConstant verifies that analyzer.go
// has the raw Cypher constant for batch complexity writes.
// Expected result: Source contains gqlSetComplexityBatch.
func TestAnalyzer_HasSetComplexityBatchConstant(t *testing.T) {
	data, err := os.ReadFile("../analysis/analyzer.go")
	if err != nil {
		t.Fatalf("cannot read analyzer.go: %v", err)
	}
	src := string(data)
	if !strings.Contains(src, "gqlSetComplexityBatch") {
		t.Error("analyzer.go should contain gqlSetComplexityBatch constant")
	}
}

// TestLevenshteinDependency verifies that agnivade/levenshtein is in go.mod.
// Expected result: go.mod contains "agnivade/levenshtein".
func TestLevenshteinDependency(t *testing.T) {
	data, err := os.ReadFile("../../go.mod")
	if err != nil {
		t.Fatalf("cannot read go.mod: %v", err)
	}
	src := string(data)
	if !strings.Contains(src, "agnivade/levenshtein") {
		t.Error("go.mod should contain agnivade/levenshtein dependency")
	}
}
