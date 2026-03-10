package mcp

import (
	"context"
	"strings"
	"time"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"github.com/tab58/code-context/internal/analysis"
	codedb "github.com/tab58/code-context/internal/clients/code_db"
	"github.com/tab58/code-context/internal/indexer"
)

// Server wraps mark3labs/mcp-go and registers tool handlers backed by
// the code-context pipeline (indexer, analyzer, CodeDB).
type Server struct {
	db       *codedb.CodeDB
	idx      *indexer.Indexer
	analyzer *analysis.Analyzer
	mcp      *server.MCPServer
	http     *server.StreamableHTTPServer
}

// NewServer creates an MCP server with all 19 tool handlers (4 search + 5 traversal + 4 context + 3 management + 3 analysis) registered.
func NewServer(db *codedb.CodeDB, idx *indexer.Indexer, analyzer *analysis.Analyzer) *Server {
	s := &Server{
		db:       db,
		idx:      idx,
		analyzer: analyzer,
	}

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

	// Register ingest_repository tool
	ingestRepoTool := mcp.NewTool("ingest_repository",
		mcp.WithDescription("Index a local directory into the code graph (index -> analyze pipeline)"),
		mcp.WithString("repository_path", mcp.Required(), mcp.Description("Absolute path to the local directory to index")),
	)
	mcpServer.AddTool(ingestRepoTool, s.mcpHandleIngestRepository)

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

// mcpHandleFindFunction is the MCP tool handler for find_function.
func (s *Server) mcpHandleFindFunction(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	repo := request.GetString("repository", "")
	name := request.GetString("name", "")

	resp, err := s.handleFindFunction(ctx, repo, name)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	return marshalMCPResult(resp)
}

// mcpHandleFindFile is the MCP tool handler for find_file.
func (s *Server) mcpHandleFindFile(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	repo := request.GetString("repository", "")
	pattern := request.GetString("pattern", "")

	resp, err := s.handleFindFile(ctx, repo, pattern)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	return marshalMCPResult(resp)
}

// mcpHandleSearchCode is the MCP tool handler for search_code.
func (s *Server) mcpHandleSearchCode(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	repo := request.GetString("repository", "")
	query := request.GetString("query", "")
	limit := int(request.GetFloat("limit", defaultLimit))

	resp, err := s.handleSearchCode(ctx, repo, query, limit)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	return marshalMCPResult(resp)
}

// mcpHandleGetCallers is the MCP tool handler for get_callers.
func (s *Server) mcpHandleGetCallers(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	repo := request.GetString("repository", "")
	name := request.GetString("name", "")
	depth := int(request.GetFloat("depth", 1))

	resp, err := s.handleGetCallers(ctx, repo, name, depth)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	return marshalMCPResult(resp)
}

// mcpHandleGetCallees is the MCP tool handler for get_callees.
func (s *Server) mcpHandleGetCallees(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	repo := request.GetString("repository", "")
	name := request.GetString("name", "")
	depth := int(request.GetFloat("depth", 1))

	resp, err := s.handleGetCallees(ctx, repo, name, depth)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	return marshalMCPResult(resp)
}

// mcpHandleGetClassHierarchy is the MCP tool handler for get_class_hierarchy.
func (s *Server) mcpHandleGetClassHierarchy(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	repo := request.GetString("repository", "")
	name := request.GetString("name", "")
	direction := request.GetString("direction", "both")
	depth := int(request.GetFloat("depth", 1))

	resp, err := s.handleGetClassHierarchy(ctx, repo, name, direction, depth)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	return marshalMCPResult(resp)
}

// mcpHandleGetDependencies is the MCP tool handler for get_dependencies.
func (s *Server) mcpHandleGetDependencies(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	repo := request.GetString("repository", "")
	name := request.GetString("name", "")
	depth := int(request.GetFloat("depth", 1))

	resp, err := s.handleGetDependencies(ctx, repo, name, depth)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	return marshalMCPResult(resp)
}

// mcpHandleGetReferences is the MCP tool handler for get_references.
func (s *Server) mcpHandleGetReferences(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	repo := request.GetString("repository", "")
	name := request.GetString("name", "")

	resp, err := s.handleGetReferences(ctx, repo, name)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	return marshalMCPResult(resp)
}

// mcpHandleGetRepoMap is the MCP tool handler for get_repo_map.
func (s *Server) mcpHandleGetRepoMap(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	repo := request.GetString("repository", "")

	resp, err := s.handleGetRepoMap(ctx, repo)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	return marshalMCPResult(resp)
}

// mcpHandleGetFileOverview is the MCP tool handler for get_file_overview.
func (s *Server) mcpHandleGetFileOverview(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	repo := request.GetString("repository", "")
	path := request.GetString("path", "")

	resp, err := s.handleGetFileOverview(ctx, repo, path)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	return marshalMCPResult(resp)
}

// mcpHandleGetSymbolContext is the MCP tool handler for get_symbol_context.
func (s *Server) mcpHandleGetSymbolContext(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	repo := request.GetString("repository", "")
	name := request.GetString("name", "")

	resp, err := s.handleGetSymbolContext(ctx, repo, name)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	return marshalMCPResult(resp)
}

// mcpHandleIngestRepository is the MCP tool handler for ingest_repository.
func (s *Server) mcpHandleIngestRepository(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	repoPath := request.GetString("repository_path", "")

	resp, err := s.handleIngestRepository(ctx, repoPath)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	return marshalMCPResult(resp)
}

// mcpHandleDeleteRepository is the MCP tool handler for delete_repository.
func (s *Server) mcpHandleDeleteRepository(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	repo := request.GetString("repository", "")

	resp, err := s.handleDeleteRepository(ctx, repo)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	return marshalMCPResult(resp)
}

// mcpHandleGetRepositoryStats is the MCP tool handler for get_repository_stats.
func (s *Server) mcpHandleGetRepositoryStats(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	repo := request.GetString("repository", "")

	resp, err := s.handleGetRepositoryStats(ctx, repo)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	return marshalMCPResult(resp)
}

// mcpHandleReadSource is the MCP tool handler for read_source.
func (s *Server) mcpHandleReadSource(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	repo := request.GetString("repository", "")
	namesStr := request.GetString("names", "")

	var names []string
	for _, n := range strings.Split(namesStr, ",") {
		trimmed := strings.TrimSpace(n)
		if trimmed != "" {
			names = append(names, trimmed)
		}
	}

	resp, err := s.handleReadSource(ctx, repo, names)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	return marshalMCPResult(resp)
}

// mcpHandleSearchCodeContent is the MCP tool handler for content-based search_code.
func (s *Server) mcpHandleSearchCodeContent(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	repo := request.GetString("repository", "")
	query := request.GetString("query", "")
	limit := int(request.GetFloat("limit", float64(defaultLimit)))

	resp, err := s.handleSearchCodeContent(ctx, repo, query, limit)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	return marshalMCPResult(resp)
}

// mcpHandleFindDeadCode is the MCP tool handler for find_dead_code.
func (s *Server) mcpHandleFindDeadCode(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	repo := request.GetString("repository", "")
	excludeDecorated := request.GetBool("exclude_decorated", false)
	excludePatterns := request.GetString("exclude_patterns", "")
	limit := int(request.GetFloat("limit", 50))

	resp, err := s.handleFindDeadCode(ctx, repo, excludeDecorated, excludePatterns, limit)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	return marshalMCPResult(resp)
}

// mcpHandleCalculateCyclomaticComplexity is the MCP tool handler for calculate_cyclomatic_complexity.
func (s *Server) mcpHandleCalculateCyclomaticComplexity(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	repo := request.GetString("repository", "")
	name := request.GetString("name", "")

	resp, err := s.handleCalculateCyclomaticComplexity(ctx, repo, name)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	return marshalMCPResult(resp)
}

// mcpHandleFindMostComplexFunctions is the MCP tool handler for find_most_complex_functions.
func (s *Server) mcpHandleFindMostComplexFunctions(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	repo := request.GetString("repository", "")
	minComplexity := int(request.GetFloat("min_complexity", 5))
	limit := int(request.GetFloat("limit", float64(defaultLimit)))

	resp, err := s.handleFindMostComplexFunctions(ctx, repo, minComplexity, limit)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	return marshalMCPResult(resp)
}

