package verify_test

import (
	"strings"
	"testing"
)

// === Task 1: Add mark3labs/mcp-go dependency ===
//
// mark3labs/mcp-go provides the MCP protocol implementation: stdio transport,
// tool registration, JSON schema validation, and JSON-RPC handling.

// TestGoModContainsMcpGo verifies that go.mod includes the mark3labs/mcp-go
// dependency for MCP protocol handling.
// Expected result: go.mod contains mark3labs/mcp-go.
func TestGoModContainsMcpGo(t *testing.T) {
	gomod := readProjectFile(t, "go.mod")

	if !strings.Contains(gomod, "github.com/mark3labs/mcp-go") {
		t.Error("go.mod missing mark3labs/mcp-go dependency (required for MCP server)")
	}
}

// === Task 2: internal/mcp/types.go exists ===

// TestMcpTypesFileExists verifies that internal/mcp/types.go has been created.
// Expected result: file exists.
func TestMcpTypesFileExists(t *testing.T) {
	if !projectFileExists(t, "internal/mcp/types.go") {
		t.Error("internal/mcp/types.go does not exist")
	}
}

// === Task 3: internal/mcp/server.go exists ===

// TestMcpServerFileExists verifies that internal/mcp/server.go has been created.
// Expected result: file exists.
func TestMcpServerFileExists(t *testing.T) {
	if !projectFileExists(t, "internal/mcp/server.go") {
		t.Error("internal/mcp/server.go does not exist")
	}
}

// TestMcpSearchFileExists verifies that internal/mcp/search.go has been created.
// Expected result: file exists.
func TestMcpSearchFileExists(t *testing.T) {
	if !projectFileExists(t, "internal/mcp/search.go") {
		t.Error("internal/mcp/search.go does not exist")
	}
}

// TestMcpToolsFileExists verifies that internal/mcp/tools.go has been created.
// Expected result: file exists.
func TestMcpToolsFileExists(t *testing.T) {
	if !projectFileExists(t, "internal/mcp/tools.go") {
		t.Error("internal/mcp/tools.go does not exist")
	}
}

// === Task 9: Wire MCP server into cmd/codectx/main.go ===
//
// main.go should import internal/mcp, create a Server with pipeline
// components, and call Serve() instead of the placeholder println.

// TestMainGoImportsMcpPackage verifies that cmd/codectx/main.go imports
// the internal/mcp package for MCP server creation.
// Expected result: main.go contains "internal/mcp" import.
func TestMainGoImportsMcpPackage(t *testing.T) {
	content := readProjectFile(t, "cmd/codectx/main.go")

	if !strings.Contains(content, "internal/mcp") {
		t.Error("cmd/codectx/main.go does not import internal/mcp — MCP server not wired")
	}
}

// TestMainGoCreatesMcpServer verifies that main.go calls mcp.NewServer()
// with the pipeline components.
// Expected result: main.go contains mcp.NewServer or equivalent.
func TestMainGoCreatesMcpServer(t *testing.T) {
	content := readProjectFile(t, "cmd/codectx/main.go")

	// Check for mcp.NewServer or mcpserver.NewServer (either import alias)
	if !strings.Contains(content, "NewServer") || !strings.Contains(content, "internal/mcp") {
		t.Error("cmd/codectx/main.go does not create MCP server via NewServer()")
	}
}

// TestMainGoCallsServe verifies that main.go calls server.Serve(ctx)
// to start the MCP stdio transport.
// Expected result: main.go contains .Serve( call.
func TestMainGoCallsServe(t *testing.T) {
	content := readProjectFile(t, "cmd/codectx/main.go")

	if !strings.Contains(content, ".Serve(") {
		t.Error("cmd/codectx/main.go does not call .Serve() — MCP server not started")
	}
}

// TestMainGoNoPlaceholderPrintln verifies that the placeholder
// "code-context server ready" println has been replaced with MCP server.
// Expected result: main.go does NOT contain the placeholder.
func TestMainGoNoPlaceholderPrintln(t *testing.T) {
	content := readProjectFile(t, "cmd/codectx/main.go")

	if strings.Contains(content, `"code-context server ready"`) {
		t.Error("cmd/codectx/main.go still has placeholder println — should be replaced with MCP server")
	}
}

// TestMainGoNoWaitForShutdown verifies that main.go no longer blocks on
// ctx.Done() waiting for shutdown — Serve() handles that.
// Expected result: main.go does NOT have "<-ctx.Done()" at the end.
func TestMainGoNoWaitForShutdown(t *testing.T) {
	content := readProjectFile(t, "cmd/codectx/main.go")

	// The old pattern was: <-ctx.Done() followed by "shutting down"
	// With Serve(), the server blocks until shutdown, no separate wait needed
	if strings.Contains(content, `<-ctx.Done()`) && strings.Contains(content, `"code-context server ready"`) {
		t.Error("cmd/codectx/main.go still has placeholder shutdown pattern — Serve() should handle blocking")
	}
}

// === Task 13: MCP server end-to-end verification ===
//
// Verifies that all 4 MCP tool definitions are registered.

// TestMcpPackageHasAllHandlers verifies that the internal/mcp package
// contains handler implementations for all 3 tools.
// Expected result: tools.go contains handleFindFunction, handleFindFile,
// handleSearchCode.
func TestMcpPackageHasAllHandlers(t *testing.T) {
	content := readProjectFile(t, "internal/mcp/tools.go")

	handlers := []string{
		"handleFindFunction",
		"handleFindFile",
		"handleSearchCode",
	}

	for _, h := range handlers {
		t.Run(h, func(t *testing.T) {
			if !strings.Contains(content, h) {
				t.Errorf("internal/mcp/tools.go missing %s handler", h)
			}
		})
	}
}

// TestMcpServerRegistersTools verifies that server.go registers all 3 tools
// with mcp-go tool definitions.
// Expected result: server.go contains tool name strings for all 3 tools.
func TestMcpServerRegistersTools(t *testing.T) {
	content := readProjectFile(t, "internal/mcp/server.go")

	tools := []string{
		"find_function",
		"find_file",
		"search_code",
	}

	for _, tool := range tools {
		t.Run(tool, func(t *testing.T) {
			if !strings.Contains(content, tool) {
				t.Errorf("internal/mcp/server.go does not register %q tool", tool)
			}
		})
	}
}

// TestMcpServerUsesStdioTransport verifies that server.go imports the mcp-go
// library (not just mentions it in comments) and uses stdio transport.
// Expected result: server.go has a Go import line for mark3labs/mcp-go.
func TestMcpServerUsesStdioTransport(t *testing.T) {
	content := readProjectFile(t, "internal/mcp/server.go")

	// Check for actual Go import of mcp-go (inside import block, with quotes)
	if !strings.Contains(content, `"github.com/mark3labs/mcp-go`) {
		t.Error("internal/mcp/server.go does not import mcp-go — stdio transport requires mcp-go library import")
	}
}
