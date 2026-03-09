package verify

import (
	"os"
	"strings"
	"testing"
)

// === Task 11: Verify tests for structural correctness ===
//
// Verifies that traversal.go exists, 10 gql constants are present,
// 5 tools are registered in server.go (8 total), maxTraversalDepth exists,
// and TraversalResult/TraversalResponse types exist in types.go.

// TestTraversalFile_Exists verifies that internal/mcp/traversal.go exists.
// Expected result: file exists on disk.
func TestTraversalFile_Exists(t *testing.T) {
	_, err := os.Stat("../../internal/mcp/traversal.go")
	if err != nil {
		t.Fatalf("traversal.go should exist: %v", err)
	}
}

// TestTraversalFile_Has10GqlConstants verifies that traversal.go defines
// all 10 GraphQL query constants.
// Expected result: all 10 constants found in source.
func TestTraversalFile_Has10GqlConstants(t *testing.T) {
	data, err := os.ReadFile("../../internal/mcp/traversal.go")
	if err != nil {
		t.Fatalf("failed to read traversal.go: %v", err)
	}
	src := string(data)

	constants := []string{
		"gqlFindCallers",
		"gqlFindCallees",
		"gqlFindParentClasses",
		"gqlFindImplementedInterfaces",
		"gqlFindChildClasses",
		"gqlFindImplementors",
		"gqlFindModuleDeps",
		"gqlFindFileImports",
		"gqlDetectFunction",
		"gqlDetectClass",
	}

	for _, c := range constants {
		if !strings.Contains(src, c) {
			t.Errorf("traversal.go should contain constant %q", c)
		}
	}
}

// TestTraversalFile_HasTraverseHops verifies that traversal.go defines
// the traverseHops method.
// Expected result: "traverseHops" found in source.
func TestTraversalFile_HasTraverseHops(t *testing.T) {
	data, err := os.ReadFile("../../internal/mcp/traversal.go")
	if err != nil {
		t.Fatalf("failed to read traversal.go: %v", err)
	}
	if !strings.Contains(string(data), "traverseHops") {
		t.Error("traversal.go should contain traverseHops method")
	}
}

// TestTraversalFile_HasClampDepth verifies that traversal.go defines
// the clampDepth function.
// Expected result: "clampDepth" found in source.
func TestTraversalFile_HasClampDepth(t *testing.T) {
	data, err := os.ReadFile("../../internal/mcp/traversal.go")
	if err != nil {
		t.Fatalf("failed to read traversal.go: %v", err)
	}
	if !strings.Contains(string(data), "func clampDepth") {
		t.Error("traversal.go should contain clampDepth function")
	}
}

// TestTraversalFile_HasIsFilePath verifies that traversal.go defines
// the isFilePath function.
// Expected result: "isFilePath" found in source.
func TestTraversalFile_HasIsFilePath(t *testing.T) {
	data, err := os.ReadFile("../../internal/mcp/traversal.go")
	if err != nil {
		t.Fatalf("failed to read traversal.go: %v", err)
	}
	if !strings.Contains(string(data), "func isFilePath") {
		t.Error("traversal.go should contain isFilePath function")
	}
}

// TestTraversalFile_Has4ParseFunctions verifies that traversal.go defines
// all 4 parse functions.
// Expected result: all 4 parse functions found in source.
func TestTraversalFile_Has4ParseFunctions(t *testing.T) {
	data, err := os.ReadFile("../../internal/mcp/traversal.go")
	if err != nil {
		t.Fatalf("failed to read traversal.go: %v", err)
	}
	src := string(data)

	funcs := []string{
		"func parseFunctionResult",
		"func parseClassResult",
		"func parseModuleResult",
		"func parseFileResult",
	}

	for _, f := range funcs {
		if !strings.Contains(src, f) {
			t.Errorf("traversal.go should contain %q", f)
		}
	}
}

// TestTraversalFile_HasToTraversalMCPResult verifies that traversal.go
// defines the toTraversalMCPResult converter.
// Expected result: "toTraversalMCPResult" found in source.
func TestTraversalFile_HasToTraversalMCPResult(t *testing.T) {
	data, err := os.ReadFile("../../internal/mcp/traversal.go")
	if err != nil {
		t.Fatalf("failed to read traversal.go: %v", err)
	}
	if !strings.Contains(string(data), "func toTraversalMCPResult") {
		t.Error("traversal.go should contain toTraversalMCPResult function")
	}
}

// TestTypesFile_HasTraversalTypes verifies that types.go contains
// TraversalResult and TraversalResponse type definitions.
// Expected result: both types found in source.
func TestTypesFile_HasTraversalTypes(t *testing.T) {
	data, err := os.ReadFile("../../internal/mcp/types.go")
	if err != nil {
		t.Fatalf("failed to read types.go: %v", err)
	}
	src := string(data)

	if !strings.Contains(src, "type TraversalResult struct") {
		t.Error("types.go should contain TraversalResult struct")
	}
	if !strings.Contains(src, "type TraversalResponse struct") {
		t.Error("types.go should contain TraversalResponse struct")
	}
}

// TestTypesFile_HasMaxTraversalDepth verifies that types.go contains
// the maxTraversalDepth constant.
// Expected result: "maxTraversalDepth" constant found in source.
func TestTypesFile_HasMaxTraversalDepth(t *testing.T) {
	data, err := os.ReadFile("../../internal/mcp/types.go")
	if err != nil {
		t.Fatalf("failed to read types.go: %v", err)
	}
	if !strings.Contains(string(data), "maxTraversalDepth") {
		t.Error("types.go should contain maxTraversalDepth constant")
	}
}

// TestServerFile_Has8Tools verifies that server.go registers 8 tools
// (3 search + 5 traversal) via AddTool calls.
// Expected result: 8 AddTool calls in server.go.
func TestServerFile_Has8Tools(t *testing.T) {
	data, err := os.ReadFile("../../internal/mcp/server.go")
	if err != nil {
		t.Fatalf("failed to read server.go: %v", err)
	}
	src := string(data)

	count := strings.Count(src, "mcpServer.AddTool(")
	if count != 8 {
		t.Errorf("server.go should have 8 AddTool calls, got %d", count)
	}
}

// TestServerFile_HasTraversalToolNames verifies that server.go registers
// all 5 traversal tool names.
// Expected result: all 5 tool names found in AddTool calls.
func TestServerFile_HasTraversalToolNames(t *testing.T) {
	data, err := os.ReadFile("../../internal/mcp/server.go")
	if err != nil {
		t.Fatalf("failed to read server.go: %v", err)
	}
	src := string(data)

	tools := []string{
		`"get_callers"`,
		`"get_callees"`,
		`"get_class_hierarchy"`,
		`"get_dependencies"`,
		`"get_references"`,
	}

	for _, tool := range tools {
		if !strings.Contains(src, tool) {
			t.Errorf("server.go should register tool %s", tool)
		}
	}
}

// TestServerFile_DocComment8Tools verifies that NewServer doc comment
// mentions "8 tool handlers".
// Expected result: "8 tool" found in server.go.
func TestServerFile_DocComment8Tools(t *testing.T) {
	data, err := os.ReadFile("../../internal/mcp/server.go")
	if err != nil {
		t.Fatalf("failed to read server.go: %v", err)
	}
	if !strings.Contains(string(data), "8 tool") {
		t.Error("server.go NewServer doc comment should mention '8 tool handlers'")
	}
}

// TestTraversalFile_Has5Handlers verifies that traversal.go defines
// all 5 handler methods.
// Expected result: all 5 handler function signatures found.
func TestTraversalFile_Has5Handlers(t *testing.T) {
	data, err := os.ReadFile("../../internal/mcp/traversal.go")
	if err != nil {
		t.Fatalf("failed to read traversal.go: %v", err)
	}
	src := string(data)

	handlers := []string{
		"handleGetCallers",
		"handleGetCallees",
		"handleGetClassHierarchy",
		"handleGetDependencies",
		"handleGetReferences",
	}

	for _, h := range handlers {
		if !strings.Contains(src, h) {
			t.Errorf("traversal.go should contain handler %q", h)
		}
	}
}
