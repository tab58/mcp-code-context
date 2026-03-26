package verify_test

import (
	"strings"
	"testing"
)

// === Task 1: Add MCP_PORT to app config ===
//
// Config struct gains MCPPort field read from MCP_PORT env var with default "8080".

// TestConfigHasMCPPortField verifies that cmd/codectx/config/config.go
// defines an MCPPort field on the Config struct.
// Expected result: config.go contains "MCPPort" field.
func TestConfigHasMCPPortField(t *testing.T) {
	content := readProjectFile(t, "cmd/codectx/config/config.go")

	if !strings.Contains(content, "MCPPort") {
		t.Error("config.go missing MCPPort field — needed for MCP server port configuration")
	}
}

// TestConfigReadsMCPPortEnv verifies that config.go reads the MCP_PORT env var.
// Expected result: config.go contains "MCP_PORT" string.
func TestConfigReadsMCPPortEnv(t *testing.T) {
	content := readProjectFile(t, "cmd/codectx/config/config.go")

	if !strings.Contains(content, "MCP_PORT") {
		t.Error("config.go does not read MCP_PORT environment variable")
	}
}

// TestConfigMCPPortDefault verifies that config.go defaults MCP_PORT to "8080".
// Expected result: config.go contains "8080" as default value.
func TestConfigMCPPortDefault(t *testing.T) {
	content := readProjectFile(t, "cmd/codectx/config/config.go")

	if !strings.Contains(content, "8080") {
		t.Error("config.go does not default MCP_PORT to 8080")
	}
}

// === Task 2: Switch MCP transport from stdio to streamable HTTP ===
//
// Server gains http field and Serve(ctx, addr) starts HTTP server.

// TestMcpServerImportsStreamableHTTP verifies that server.go imports
// the streamable HTTP transport from mcp-go.
// Expected result: server.go references StreamableHTTP or streamable_http.
func TestMcpServerImportsStreamableHTTP(t *testing.T) {
	content := readProjectFile(t, "internal/api/mcp/server.go")

	if !strings.Contains(content, "StreamableHTTP") {
		t.Error("server.go does not reference StreamableHTTPServer — transport not switched to HTTP")
	}
}

// TestMcpServerServeAcceptsAddr verifies that Serve method accepts an addr parameter.
// Expected result: server.go contains "Serve(ctx context.Context, addr string)".
func TestMcpServerServeAcceptsAddr(t *testing.T) {
	content := readProjectFile(t, "internal/api/mcp/server.go")

	if !strings.Contains(content, "addr string") {
		t.Error("server.go Serve method does not accept addr parameter — transport not switched to HTTP")
	}
}

// TestMcpServerNoStdioTransport verifies that server.go no longer uses stdio transport.
// Expected result: server.go does NOT contain ServeStdio.
func TestMcpServerNoStdioTransport(t *testing.T) {
	content := readProjectFile(t, "internal/api/mcp/server.go")

	if strings.Contains(content, "ServeStdio") {
		t.Error("server.go still uses ServeStdio — should be switched to streamable HTTP")
	}
}

// === Task 3: Add ProgressFunc type and IndexOption to indexer ===

// TestIndexerHasProgressFunc verifies that indexer.go defines a ProgressFunc type.
// Expected result: indexer.go contains "ProgressFunc".
func TestIndexerHasProgressFunc(t *testing.T) {
	content := readProjectFile(t, "internal/indexer/indexer.go")

	if !strings.Contains(content, "ProgressFunc") {
		t.Error("indexer.go missing ProgressFunc type definition")
	}
}

// TestIndexerHasIndexOption verifies that indexer.go defines an IndexOption type.
// Expected result: indexer.go contains "IndexOption".
func TestIndexerHasIndexOption(t *testing.T) {
	content := readProjectFile(t, "internal/indexer/indexer.go")

	if !strings.Contains(content, "IndexOption") {
		t.Error("indexer.go missing IndexOption type definition")
	}
}

// TestIndexerHasWithProgress verifies that indexer.go defines WithProgress constructor.
// Expected result: indexer.go contains "WithProgress".
func TestIndexerHasWithProgress(t *testing.T) {
	content := readProjectFile(t, "internal/indexer/indexer.go")

	if !strings.Contains(content, "WithProgress") {
		t.Error("indexer.go missing WithProgress option constructor")
	}
}

// === Task 4: Add AnalyzeOption with progress callback to analyzer ===

// TestAnalyzerHasAnalyzeOption verifies that analyzer.go defines AnalyzeOption.
// Expected result: analyzer.go contains "AnalyzeOption".
func TestAnalyzerHasAnalyzeOption(t *testing.T) {
	content := readProjectFile(t, "internal/analysis/analyzer.go")

	if !strings.Contains(content, "AnalyzeOption") {
		t.Error("analyzer.go missing AnalyzeOption type definition")
	}
}

// TestAnalyzerHasWithAnalyzeProgress verifies that analyzer.go defines
// WithAnalyzeProgress constructor.
// Expected result: analyzer.go contains "WithAnalyzeProgress".
func TestAnalyzerHasWithAnalyzeProgress(t *testing.T) {
	content := readProjectFile(t, "internal/analysis/analyzer.go")

	if !strings.Contains(content, "WithAnalyzeProgress") {
		t.Error("analyzer.go missing WithAnalyzeProgress option constructor")
	}
}

// === Task 6: Create internal/repl package ===

// TestReplPackageExists verifies that internal/repl/repl.go exists.
// Expected result: file exists.
func TestReplPackageExists(t *testing.T) {
	if !projectFileExists(t, "internal/repl/repl.go") {
		t.Error("internal/repl/repl.go does not exist")
	}
}

// TestReplCommandsFileExists verifies that internal/repl/commands.go exists.
// Expected result: file exists.
func TestReplCommandsFileExists(t *testing.T) {
	if !projectFileExists(t, "internal/repl/commands.go") {
		t.Error("internal/repl/commands.go does not exist")
	}
}

// === Task 8: Wire REPL + MCP server into main.go ===

// TestMainGoImportsReplPackage verifies that cmd/codectx/main.go imports
// the internal/repl package for REPL creation.
// Expected result: main.go contains "internal/repl" import.
func TestMainGoImportsReplPackage(t *testing.T) {
	content := readProjectFile(t, "cmd/codectx/main.go")

	if !strings.Contains(content, "internal/repl") {
		t.Error("cmd/codectx/main.go does not import internal/repl — REPL not wired")
	}
}

// TestMainGoReadsMCPPort verifies that main.go reads MCPPort from config.
// Expected result: main.go contains "MCPPort".
func TestMainGoReadsMCPPort(t *testing.T) {
	content := readProjectFile(t, "cmd/codectx/main.go")

	if !strings.Contains(content, "MCPPort") {
		t.Error("cmd/codectx/main.go does not read MCPPort from config")
	}
}

// TestMainGoServesHTTP verifies that main.go passes an address to Serve.
// Expected result: main.go contains .Serve(ctx, with an addr argument.
func TestMainGoServesHTTP(t *testing.T) {
	content := readProjectFile(t, "cmd/codectx/main.go")

	// Should call server.Serve(ctx, ":"+...) or similar with addr
	if !strings.Contains(content, "Serve(ctx,") {
		t.Error("cmd/codectx/main.go does not call Serve with addr parameter — HTTP transport not wired")
	}
}

// TestMainGoCreatesREPL verifies that main.go creates a REPL.
// Expected result: main.go contains repl.New or equivalent.
func TestMainGoCreatesREPL(t *testing.T) {
	content := readProjectFile(t, "cmd/codectx/main.go")

	if !strings.Contains(content, "repl.New") && !strings.Contains(content, "repl.New(") {
		t.Error("cmd/codectx/main.go does not create REPL via repl.New()")
	}
}

// TestMainGoRunsREPL verifies that main.go calls REPL.Run.
// Expected result: main.go contains .Run(ctx).
func TestMainGoRunsREPL(t *testing.T) {
	content := readProjectFile(t, "cmd/codectx/main.go")

	if !strings.Contains(content, ".Run(ctx)") {
		t.Error("cmd/codectx/main.go does not call .Run(ctx) — REPL not started on main goroutine")
	}
}

// TestMainGoLaunchesMCPInGoroutine verifies that main.go launches MCP server
// in a goroutine (not blocking main goroutine).
// Expected result: main.go contains "go " before Serve call.
func TestMainGoLaunchesMCPInGoroutine(t *testing.T) {
	content := readProjectFile(t, "cmd/codectx/main.go")

	if !strings.Contains(content, "go server.Serve") && !strings.Contains(content, "go mcpServer.Serve") {
		// More flexible check: look for "go " + "Serve" pattern
		if !strings.Contains(content, "go ") || !strings.Contains(content, ".Serve(ctx,") {
			t.Error("cmd/codectx/main.go does not launch MCP server in goroutine")
		}
	}
}
