package mcp

import (
	"context"
	"encoding/json"
	"fmt"
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

// NewServer creates an MCP server with all 8 tool handlers (3 search + 5 traversal) registered.
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

	// Register search_code tool
	searchCodeTool := mcp.NewTool("search_code",
		mcp.WithDescription("Unified agentic search with automatic strategy selection"),
		mcp.WithString("repository", mcp.Required(), mcp.Description("Repository name to search in")),
		mcp.WithString("query", mcp.Required(), mcp.Description("Search query (function name, file glob, or natural language)")),
		mcp.WithNumber("limit", mcp.Description("Maximum number of results to return (default 10)")),
	)
	mcpServer.AddTool(searchCodeTool, s.mcpHandleSearchCode)

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
	return toMCPResult(resp)
}

// mcpHandleFindFile is the MCP tool handler for find_file.
func (s *Server) mcpHandleFindFile(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	repo := request.GetString("repository", "")
	pattern := request.GetString("pattern", "")

	resp, err := s.handleFindFile(ctx, repo, pattern)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	return toMCPResult(resp)
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
	return toMCPResult(resp)
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
	return toTraversalMCPResult(resp)
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
	return toTraversalMCPResult(resp)
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
	return toTraversalMCPResult(resp)
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
	return toTraversalMCPResult(resp)
}

// mcpHandleGetReferences is the MCP tool handler for get_references.
func (s *Server) mcpHandleGetReferences(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	repo := request.GetString("repository", "")
	name := request.GetString("name", "")

	resp, err := s.handleGetReferences(ctx, repo, name)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	return toTraversalMCPResult(resp)
}

// toMCPResult converts a SearchResponse to an MCP CallToolResult.
func toMCPResult(resp *SearchResponse) (*mcp.CallToolResult, error) {
	data, err := json.Marshal(resp)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal response: %w", err)
	}
	return mcp.NewToolResultText(string(data)), nil
}
