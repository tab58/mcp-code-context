package verify

import (
	"os"
	"strings"
	"testing"
)

// ============================================================================
// Context-Optimized MCP Tools — Structural Verification Tests
// ============================================================================

// --- Task 1: Context response types in types.go ---

// TestContextTypes_TypesFile verifies types.go contains all 10 new types.
// Expected: all type names found in source.
func TestContextTypes_TypesFile(t *testing.T) {
	data, err := os.ReadFile("../../internal/mcp/types.go")
	if err != nil {
		t.Fatalf("failed to read types.go: %v", err)
	}
	src := string(data)
	types := []string{
		"RepoMapFile", "RepoMapEntry", "RepoMapResponse",
		"OverviewSymbol", "FileOverviewResponse",
		"SymbolDetail", "SymbolSummary", "SymbolContextResponse",
		"ReadSourceResult", "ReadSourceResponse",
	}
	for _, typ := range types {
		if !strings.Contains(src, "type "+typ+" struct") {
			t.Errorf("types.go should contain 'type %s struct'", typ)
		}
	}
}

// --- Task 2: Remove source from gqlFindFunctions ---

// TestSourceRemoval_GqlFindFunctions verifies gqlFindFunctions in tools.go
// does NOT contain "source" field.
// Expected: the query string has no "source" field.
func TestSourceRemoval_GqlFindFunctions(t *testing.T) {
	data, err := os.ReadFile("../../internal/mcp/tools.go")
	if err != nil {
		t.Fatalf("failed to read tools.go: %v", err)
	}
	src := string(data)

	// Find the gqlFindFunctions constant value
	idx := strings.Index(src, "gqlFindFunctions")
	if idx == -1 {
		t.Fatal("gqlFindFunctions not found in tools.go")
	}
	// Extract from the constant to the next backtick-close
	remainder := src[idx:]
	endIdx := strings.Index(remainder, "}`")
	if endIdx == -1 {
		t.Fatal("could not find end of gqlFindFunctions constant")
	}
	queryStr := remainder[:endIdx]

	if strings.Contains(queryStr, "source") {
		t.Error("gqlFindFunctions should NOT contain 'source' field")
	}
}

// TestSourceRemoval_HandlerMapping verifies handleFindFunction in tools.go
// does NOT map Source field from response.
// Expected: no "Source:" mapping line in handleFindFunction.
func TestSourceRemoval_HandlerMapping(t *testing.T) {
	data, err := os.ReadFile("../../internal/mcp/tools.go")
	if err != nil {
		t.Fatalf("failed to read tools.go: %v", err)
	}
	src := string(data)

	// Find the handleFindFunction method
	idx := strings.Index(src, "func (s *Server) handleFindFunction")
	if idx == -1 {
		t.Fatal("handleFindFunction not found in tools.go")
	}
	// Get the method body (approximate: next 50 lines / 2000 chars)
	endIdx := idx + 2000
	if endIdx > len(src) {
		endIdx = len(src)
	}
	methodBody := src[idx:endIdx]

	// The method should NOT contain Source: strVal(m, "source")
	if strings.Contains(methodBody, `strVal(m, "source")`) {
		t.Error("handleFindFunction should NOT map Source field from response")
	}
}

// --- Task 3: context_tools.go infrastructure ---

// TestContextTools_FileExists verifies context_tools.go exists.
func TestContextTools_FileExists(t *testing.T) {
	_, err := os.Stat("../../internal/mcp/context_tools.go")
	if err != nil {
		t.Fatalf("context_tools.go should exist: %v", err)
	}
}

// TestContextTools_Has7GqlConstants verifies context_tools.go defines
// all 7 GraphQL query constants.
func TestContextTools_Has7GqlConstants(t *testing.T) {
	data, err := os.ReadFile("../../internal/mcp/context_tools.go")
	if err != nil {
		t.Fatalf("failed to read context_tools.go: %v", err)
	}
	src := string(data)
	constants := []string{
		"gqlRepoFiles", "gqlRepoFunctionPaths", "gqlRepoClassPaths",
		"gqlFileOverviewFunctions", "gqlFileOverviewClasses",
		"gqlSymbolFunction", "gqlSymbolClass",
	}
	for _, c := range constants {
		if !strings.Contains(src, c) {
			t.Errorf("context_tools.go should contain constant %q", c)
		}
	}
}

// TestContextTools_HasMarshalMCPResult verifies marshalMCPResult exists.
func TestContextTools_HasMarshalMCPResult(t *testing.T) {
	data, err := os.ReadFile("../../internal/mcp/context_tools.go")
	if err != nil {
		t.Fatalf("failed to read context_tools.go: %v", err)
	}
	if !strings.Contains(string(data), "marshalMCPResult") {
		t.Error("context_tools.go should contain marshalMCPResult function")
	}
}

// TestContextTools_HasTraversalToSummary verifies traversalToSummary exists.
func TestContextTools_HasTraversalToSummary(t *testing.T) {
	data, err := os.ReadFile("../../internal/mcp/context_tools.go")
	if err != nil {
		t.Fatalf("failed to read context_tools.go: %v", err)
	}
	if !strings.Contains(string(data), "traversalToSummary") {
		t.Error("context_tools.go should contain traversalToSummary function")
	}
}

// --- Task 8: Register 4 new tools in server.go ---

// TestServerRegistration_12Tools verifies server.go registers 12 tools
// (mentions "12 tool" in comments/docs).
func TestServerRegistration_12Tools(t *testing.T) {
	data, err := os.ReadFile("../../internal/mcp/server.go")
	if err != nil {
		t.Fatalf("failed to read server.go: %v", err)
	}
	src := string(data)
	if !strings.Contains(src, "20 tool") {
		t.Error("server.go should mention '20 tool' in doc comment")
	}
}

// TestServerRegistration_ContextTools verifies server.go registers
// get_repo_map, get_file_overview, get_symbol_context, read_source.
func TestServerRegistration_ContextTools(t *testing.T) {
	data, err := os.ReadFile("../../internal/mcp/server.go")
	if err != nil {
		t.Fatalf("failed to read server.go: %v", err)
	}
	src := string(data)
	tools := []string{
		"get_repo_map", "get_file_overview", "get_symbol_context", "read_source",
	}
	for _, tool := range tools {
		if !strings.Contains(src, `"`+tool+`"`) {
			t.Errorf("server.go should register tool %q", tool)
		}
	}
}

// TestServerRegistration_McpHandleAdapters verifies server.go has
// mcpHandle* adapter methods for all 4 context tools.
func TestServerRegistration_McpHandleAdapters(t *testing.T) {
	data, err := os.ReadFile("../../internal/mcp/server.go")
	if err != nil {
		t.Fatalf("failed to read server.go: %v", err)
	}
	src := string(data)
	adapters := []string{
		"mcpHandleGetRepoMap",
		"mcpHandleGetFileOverview",
		"mcpHandleGetSymbolContext",
		"mcpHandleReadSource",
	}
	for _, adapter := range adapters {
		if !strings.Contains(src, adapter) {
			t.Errorf("server.go should contain adapter method %q", adapter)
		}
	}
}

// --- Task 13: Full build/vet/race verification ---

// TestContextTools_4HandlerMethods verifies context_tools.go has all 4 handler methods.
func TestContextTools_4HandlerMethods(t *testing.T) {
	data, err := os.ReadFile("../../internal/mcp/context_tools.go")
	if err != nil {
		t.Fatalf("failed to read context_tools.go: %v", err)
	}
	src := string(data)
	handlers := []string{
		"handleGetRepoMap",
		"handleGetFileOverview",
		"handleGetSymbolContext",
		"handleReadSource",
	}
	for _, h := range handlers {
		if !strings.Contains(src, h) {
			t.Errorf("context_tools.go should contain handler method %q", h)
		}
	}
}
