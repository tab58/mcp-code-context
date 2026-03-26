package mcp

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/mark3labs/mcp-go/mcp"
)

// defaultLimit is the default number of results for search handlers.
const defaultLimit = 10

// maxCallChainDepth mirrors the business logic constant for parameter defaults.
const maxCallChainDepth = 5

// marshalMCPResult converts any response type to an MCP CallToolResult via JSON.
func marshalMCPResult(resp any) (*mcp.CallToolResult, error) {
	data, err := json.Marshal(resp)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal response: %w", err)
	}
	return mcp.NewToolResultText(string(data)), nil
}

// mcpHandleFindFunction is the MCP tool handler for find_function.
func (s *Server) mcpHandleFindFunction(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	repo := request.GetString("repository", "")
	name := request.GetString("name", "")

	resp, err := s.application.HandleFindFunction(ctx, repo, name)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	return marshalMCPResult(resp)
}

// mcpHandleFindFile is the MCP tool handler for find_file.
func (s *Server) mcpHandleFindFile(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	repo := request.GetString("repository", "")
	pattern := request.GetString("pattern", "")

	resp, err := s.application.HandleFindFile(ctx, repo, pattern)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	return marshalMCPResult(resp)
}

// mcpHandleSearchCode is the MCP tool handler for search_code_names.
func (s *Server) mcpHandleSearchCode(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	repo := request.GetString("repository", "")
	query := request.GetString("query", "")
	limit := int(request.GetFloat("limit", defaultLimit))

	resp, err := s.application.HandleSearchCode(ctx, repo, query, limit)
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

	resp, err := s.application.HandleGetCallers(ctx, repo, name, depth)
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

	resp, err := s.application.HandleGetCallees(ctx, repo, name, depth)
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

	resp, err := s.application.HandleGetClassHierarchy(ctx, repo, name, direction, depth)
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

	resp, err := s.application.HandleGetDependencies(ctx, repo, name, depth)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	return marshalMCPResult(resp)
}

// mcpHandleGetReferences is the MCP tool handler for get_references.
func (s *Server) mcpHandleGetReferences(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	repo := request.GetString("repository", "")
	name := request.GetString("name", "")

	resp, err := s.application.HandleGetReferences(ctx, repo, name)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	return marshalMCPResult(resp)
}

// mcpHandleGetRepoMap is the MCP tool handler for get_repo_map.
func (s *Server) mcpHandleGetRepoMap(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	repo := request.GetString("repository", "")

	resp, err := s.application.HandleGetRepoMap(ctx, repo)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	return marshalMCPResult(resp)
}

// mcpHandleGetFileOverview is the MCP tool handler for get_file_overview.
func (s *Server) mcpHandleGetFileOverview(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	repo := request.GetString("repository", "")
	path := request.GetString("path", "")

	resp, err := s.application.HandleGetFileOverview(ctx, repo, path)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	return marshalMCPResult(resp)
}

// mcpHandleGetSymbolContext is the MCP tool handler for get_symbol_context.
func (s *Server) mcpHandleGetSymbolContext(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	repo := request.GetString("repository", "")
	name := request.GetString("name", "")

	resp, err := s.application.HandleGetSymbolContext(ctx, repo, name)
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

	resp, err := s.application.HandleReadSource(ctx, repo, names)
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

	resp, err := s.application.HandleSearchCodeContent(ctx, repo, query, limit)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	return marshalMCPResult(resp)
}

// mcpHandleDeleteRepository is the MCP tool handler for delete_repository.
func (s *Server) mcpHandleDeleteRepository(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	repo := request.GetString("repository", "")

	resp, err := s.application.HandleDeleteRepository(ctx, repo)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	return marshalMCPResult(resp)
}

// mcpHandleGetRepositoryStats is the MCP tool handler for get_repository_stats.
func (s *Server) mcpHandleGetRepositoryStats(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	repo := request.GetString("repository", "")

	resp, err := s.application.HandleGetRepositoryStats(ctx, repo)
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

	resp, err := s.application.HandleFindDeadCode(ctx, repo, excludeDecorated, excludePatterns, limit)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	return marshalMCPResult(resp)
}

// mcpHandleCalculateCyclomaticComplexity is the MCP tool handler for calculate_cyclomatic_complexity.
func (s *Server) mcpHandleCalculateCyclomaticComplexity(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	repo := request.GetString("repository", "")
	name := request.GetString("name", "")

	resp, err := s.application.HandleCalculateCyclomaticComplexity(ctx, repo, name)
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

	resp, err := s.application.HandleFindMostComplexFunctions(ctx, repo, minComplexity, limit)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	return marshalMCPResult(resp)
}

// mcpHandleFindCallChain is the MCP tool handler for find_call_chain.
func (s *Server) mcpHandleFindCallChain(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	repo := request.GetString("repository", "")
	source := request.GetString("source_function", "")
	target := request.GetString("target_function", "")
	maxDepth := int(request.GetFloat("max_depth", float64(maxCallChainDepth)))

	resp, err := s.application.HandleFindCallChain(ctx, repo, source, target, maxDepth)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	return marshalMCPResult(resp)
}
