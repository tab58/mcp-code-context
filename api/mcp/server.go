package mcp

import (
	"context"
	"time"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"github.com/tab58/code-context/internal/app"
)

// Server wraps mark3labs/mcp-go and registers tool handlers backed by
// the code-context Service (business logic layer).
type Server struct {
	application app.Application
	mcp         *server.MCPServer
	http        *server.StreamableHTTPServer
}

// NewServer creates an MCP server with all 20 tool handlers registered.
func NewServer(application app.Application) *Server {
	s := &Server{application: application}

	mcpServer := server.NewMCPServer(
		"code-context",
		"1.0.0",
	)

	// Register find_function tool
	findFunctionTool := mcp.NewTool("find_function",
		mcp.WithDescription("Exact function name match within a repository"),
		mcp.WithString("repository", mcp.Required(), mcp.Description("Repository name to search in")),
		mcp.WithString("name", mcp.Required(), mcp.Description("Exact function name to find")),
	)
	mcpServer.AddTool(findFunctionTool, s.mcpHandleFindFunction)

	// Register find_file tool
	findFileTool := mcp.NewTool("find_file",
		mcp.WithDescription("Glob pattern match on file paths within a repository"),
		mcp.WithString("repository", mcp.Required(), mcp.Description("Repository name to search in")),
		mcp.WithString("pattern", mcp.Required(), mcp.Description("Glob pattern to match file paths")),
	)
	mcpServer.AddTool(findFileTool, s.mcpHandleFindFile)

	// Register search_code_names tool (renamed from search_code — name-based search)
	searchCodeNamesTool := mcp.NewTool("search_code_names",
		mcp.WithDescription("Name-based search with automatic strategy selection (exact, fuzzy, or file glob)"),
		mcp.WithString("repository", mcp.Required(), mcp.Description("Repository name to search in")),
		mcp.WithString("query", mcp.Required(), mcp.Description("Search query (function name, file glob, or natural language)")),
		mcp.WithNumber("limit", mcp.Description("Maximum number of results to return (default 10)")),
	)
	mcpServer.AddTool(searchCodeNamesTool, s.mcpHandleSearchCode)

	// Register search_code tool (content-based source code search)
	searchCodeContentTool := mcp.NewTool("search_code",
		mcp.WithDescription("Content-based search within source code with container resolution"),
		mcp.WithString("repository", mcp.Required(), mcp.Description("Repository name to search in")),
		mcp.WithString("query", mcp.Required(), mcp.Description("Substring to search for in source code")),
		mcp.WithNumber("limit", mcp.Description("Maximum number of results to return (default 10)")),
	)
	mcpServer.AddTool(searchCodeContentTool, s.mcpHandleSearchCodeContent)

	// Register get_callers tool
	getCallersTool := mcp.NewTool("get_callers",
		mcp.WithDescription("Find functions that call a given function"),
		mcp.WithString("repository", mcp.Required(), mcp.Description("Repository name")),
		mcp.WithString("name", mcp.Required(), mcp.Description("Function name to find callers of")),
		mcp.WithNumber("depth", mcp.Description("Traversal depth 1-3 (default 1)")),
	)
	mcpServer.AddTool(getCallersTool, s.mcpHandleGetCallers)

	// Register get_callees tool
	getCalleesTool := mcp.NewTool("get_callees",
		mcp.WithDescription("Find functions called by a given function"),
		mcp.WithString("repository", mcp.Required(), mcp.Description("Repository name")),
		mcp.WithString("name", mcp.Required(), mcp.Description("Function name to find callees of")),
		mcp.WithNumber("depth", mcp.Description("Traversal depth 1-3 (default 1)")),
	)
	mcpServer.AddTool(getCalleesTool, s.mcpHandleGetCallees)

	// Register get_class_hierarchy tool
	getClassHierarchyTool := mcp.NewTool("get_class_hierarchy",
		mcp.WithDescription("Find parent/child classes and interface implementations"),
		mcp.WithString("repository", mcp.Required(), mcp.Description("Repository name")),
		mcp.WithString("name", mcp.Required(), mcp.Description("Class or interface name")),
		mcp.WithString("direction", mcp.Description("Direction: up, down, or both (default both)")),
		mcp.WithNumber("depth", mcp.Description("Traversal depth 1-3 (default 1)")),
	)
	mcpServer.AddTool(getClassHierarchyTool, s.mcpHandleGetClassHierarchy)

	// Register get_dependencies tool
	getDependenciesTool := mcp.NewTool("get_dependencies",
		mcp.WithDescription("Find module dependencies or file imports"),
		mcp.WithString("repository", mcp.Required(), mcp.Description("Repository name")),
		mcp.WithString("name", mcp.Required(), mcp.Description("Module name or file path")),
		mcp.WithNumber("depth", mcp.Description("Traversal depth 1-3 (default 1)")),
	)
	mcpServer.AddTool(getDependenciesTool, s.mcpHandleGetDependencies)

	// Register get_references tool
	getReferencesTool := mcp.NewTool("get_references",
		mcp.WithDescription("Find all references to a symbol (auto-detects type)"),
		mcp.WithString("repository", mcp.Required(), mcp.Description("Repository name")),
		mcp.WithString("name", mcp.Required(), mcp.Description("Symbol name to find references to")),
	)
	mcpServer.AddTool(getReferencesTool, s.mcpHandleGetReferences)

	// Register get_repo_map tool
	getRepoMapTool := mcp.NewTool("get_repo_map",
		mcp.WithDescription("Get repository directory tree with per-file symbol counts"),
		mcp.WithString("repository", mcp.Required(), mcp.Description("Repository name")),
	)
	mcpServer.AddTool(getRepoMapTool, s.mcpHandleGetRepoMap)

	// Register get_file_overview tool
	getFileOverviewTool := mcp.NewTool("get_file_overview",
		mcp.WithDescription("Get all symbols in a file (signatures only, no source)"),
		mcp.WithString("repository", mcp.Required(), mcp.Description("Repository name")),
		mcp.WithString("path", mcp.Required(), mcp.Description("File path within the repository")),
	)
	mcpServer.AddTool(getFileOverviewTool, s.mcpHandleGetFileOverview)

	// Register get_symbol_context tool
	getSymbolContextTool := mcp.NewTool("get_symbol_context",
		mcp.WithDescription("Get symbol source + callers/callees/siblings signatures"),
		mcp.WithString("repository", mcp.Required(), mcp.Description("Repository name")),
		mcp.WithString("name", mcp.Required(), mcp.Description("Symbol name to get context for")),
	)
	mcpServer.AddTool(getSymbolContextTool, s.mcpHandleGetSymbolContext)

	// Register read_source tool
	readSourceTool := mcp.NewTool("read_source",
		mcp.WithDescription("Batch-fetch source code for multiple named symbols"),
		mcp.WithString("repository", mcp.Required(), mcp.Description("Repository name")),
		mcp.WithString("names", mcp.Required(), mcp.Description("Comma-separated symbol names to fetch source for")),
	)
	mcpServer.AddTool(readSourceTool, s.mcpHandleReadSource)

	// Register delete_repository tool
	deleteRepoTool := mcp.NewTool("delete_repository",
		mcp.WithDescription("Remove all graph data for a repository"),
		mcp.WithString("repository", mcp.Required(), mcp.Description("Repository name to delete")),
	)
	mcpServer.AddTool(deleteRepoTool, s.mcpHandleDeleteRepository)

	// Register get_repository_stats tool
	getRepoStatsTool := mcp.NewTool("get_repository_stats",
		mcp.WithDescription("Get node counts (files, functions, classes, modules, external references) for a repository"),
		mcp.WithString("repository", mcp.Required(), mcp.Description("Repository name")),
	)
	mcpServer.AddTool(getRepoStatsTool, s.mcpHandleGetRepositoryStats)

	// Register find_dead_code tool
	findDeadCodeTool := mcp.NewTool("find_dead_code",
		mcp.WithDescription("Find potentially dead code (functions/classes/modules with no inbound references)"),
		mcp.WithString("repository", mcp.Required(), mcp.Description("Repository name")),
		mcp.WithBoolean("exclude_decorated", mcp.Description("Exclude decorated symbols (TS-only, default false)")),
		mcp.WithString("exclude_patterns", mcp.Description("Comma-separated glob patterns to exclude by name")),
		mcp.WithNumber("limit", mcp.Description("Maximum number of results (default 50)")),
	)
	mcpServer.AddTool(findDeadCodeTool, s.mcpHandleFindDeadCode)

	// Register calculate_cyclomatic_complexity tool
	calcComplexityTool := mcp.NewTool("calculate_cyclomatic_complexity",
		mcp.WithDescription("Look up pre-computed cyclomatic complexity for a function"),
		mcp.WithString("repository", mcp.Required(), mcp.Description("Repository name")),
		mcp.WithString("name", mcp.Required(), mcp.Description("Function name")),
	)
	mcpServer.AddTool(calcComplexityTool, s.mcpHandleCalculateCyclomaticComplexity)

	// Register find_most_complex_functions tool
	findMostComplexTool := mcp.NewTool("find_most_complex_functions",
		mcp.WithDescription("Find functions with highest cyclomatic complexity"),
		mcp.WithString("repository", mcp.Required(), mcp.Description("Repository name")),
		mcp.WithNumber("min_complexity", mcp.Description("Minimum complexity threshold (default 5)")),
		mcp.WithNumber("limit", mcp.Description("Maximum number of results (default 10)")),
	)
	mcpServer.AddTool(findMostComplexTool, s.mcpHandleFindMostComplexFunctions)

	// Register find_call_chain tool
	findCallChainTool := mcp.NewTool("find_call_chain",
		mcp.WithDescription("Find the call path between two functions using bidirectional BFS"),
		mcp.WithString("repository", mcp.Required(), mcp.Description("Repository name")),
		mcp.WithString("source_function", mcp.Required(), mcp.Description("Source function name")),
		mcp.WithString("target_function", mcp.Required(), mcp.Description("Target function name")),
		mcp.WithNumber("max_depth", mcp.Description("Maximum traversal depth 1-5 (default 5)")),
	)
	mcpServer.AddTool(findCallChainTool, s.mcpHandleFindCallChain)

	s.mcp = mcpServer
	s.http = server.NewStreamableHTTPServer(mcpServer)
	return s
}

// Serve starts the MCP server on streamable HTTP transport at the given address
// (e.g., ":8080"). Blocks until ctx is cancelled, then calls Shutdown() for
// graceful drain.
func (s *Server) Serve(ctx context.Context, addr string) error {
	errCh := make(chan error, 1)
	go func() {
		errCh <- s.http.Start(addr)
	}()

	select {
	case <-ctx.Done():
		shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer shutdownCancel()
		return s.http.Shutdown(shutdownCtx)
	case err := <-errCh:
		return err
	}
}
