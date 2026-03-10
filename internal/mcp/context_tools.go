package mcp

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"path/filepath"
	"sort"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/tab58/go-ormql/pkg/client"
)

// GraphQL query constants for context tool handlers.
const (
	gqlRepoFiles = `query($where: FileWhere) {
  files(where: $where) { path filename language }
}`

	gqlRepoFunctionPaths = `query($where: FunctionWhere) {
  functions(where: $where) { path }
}`

	gqlRepoClassPaths = `query($where: ClassWhere) {
  classs(where: $where) { path }
}`

	gqlFileOverviewFunctions = `query($where: FunctionWhere) {
  functions(where: $where) { name signature visibility startingLine endingLine }
}`

	gqlFileOverviewClasses = `query($where: ClassWhere) {
  classs(where: $where) { name kind visibility startingLine endingLine }
}`

	gqlSymbolFunction = `query($where: FunctionWhere) {
  functions(where: $where) { name path source signature language visibility startingLine endingLine }
}`

	gqlSymbolClass = `query($where: ClassWhere) {
  classs(where: $where) { name path source kind language visibility startingLine endingLine }
}`
)

// handleGetRepoMap handles the get_repo_map MCP tool call.
func (s *Server) handleGetRepoMap(ctx context.Context, repo string) (*RepoMapResponse, error) {
	if repo == "" {
		return nil, errors.New("repository is required")
	}

	c, err := s.requireRepoClient(ctx, repo)
	if err != nil {
		return nil, err
	}

	vars := repoWhere(repo)

	// Query all files
	fileResult, err := c.Execute(ctx, gqlRepoFiles, vars)
	if err != nil {
		return nil, fmt.Errorf("repo map files query failed: %w", err)
	}

	// Query function paths for symbol counting
	funcResult, err := c.Execute(ctx, gqlRepoFunctionPaths, vars)
	if err != nil {
		return nil, fmt.Errorf("repo map function paths query failed: %w", err)
	}

	// Query class paths for symbol counting
	classResult, err := c.Execute(ctx, gqlRepoClassPaths, vars)
	if err != nil {
		return nil, fmt.Errorf("repo map class paths query failed: %w", err)
	}

	// Build symbol count map (path -> count)
	symbolCounts := make(map[string]int)
	totalSymbols := 0
	totalSymbols += countSymbolPaths(funcResult.Data(), "functions", symbolCounts)
	totalSymbols += countSymbolPaths(classResult.Data(), "classs", symbolCounts)

	// Group files by directory
	dirMap := make(map[string][]RepoMapFile)
	totalFiles := 0
	if items, ok := fileResult.Data()["files"].([]any); ok {
		for _, item := range items {
			m, ok := item.(map[string]any)
			if !ok {
				continue
			}
			path := strVal(m, "path")
			filename := strVal(m, "filename")
			language := strVal(m, "language")
			dir := filepath.Dir(path)

			dirMap[dir] = append(dirMap[dir], RepoMapFile{
				Name:        filename,
				Language:    language,
				SymbolCount: symbolCounts[path],
			})
			totalFiles++
		}
	}

	// Sort directories alphabetically
	dirNames := make([]string, 0, len(dirMap))
	for d := range dirMap {
		dirNames = append(dirNames, d)
	}
	sort.Strings(dirNames)

	directories := make([]RepoMapEntry, 0, len(dirNames))
	for _, d := range dirNames {
		files := dirMap[d]
		sort.Slice(files, func(i, j int) bool {
			return files[i].Name < files[j].Name
		})
		directories = append(directories, RepoMapEntry{
			Directory: d,
			Files:     files,
		})
	}

	return &RepoMapResponse{
		Repository:   repo,
		Directories:  directories,
		TotalFiles:   totalFiles,
		TotalSymbols: totalSymbols,
	}, nil
}

// handleGetFileOverview handles the get_file_overview MCP tool call.
func (s *Server) handleGetFileOverview(ctx context.Context, repo, path string) (*FileOverviewResponse, error) {
	if repo == "" {
		return nil, errors.New("repository is required")
	}
	if path == "" {
		return nil, errors.New("path is required")
	}

	c, err := s.requireRepoClient(ctx, repo)
	if err != nil {
		return nil, err
	}

	where := map[string]any{
		"where": map[string]any{
			"path":       path,
			"repository": map[string]any{"name": repo},
		},
	}

	// Query functions at this path
	funcResult, err := c.Execute(ctx, gqlFileOverviewFunctions, where)
	if err != nil {
		return nil, fmt.Errorf("file overview functions query failed: %w", err)
	}

	// Query classes at this path
	classResult, err := c.Execute(ctx, gqlFileOverviewClasses, where)
	if err != nil {
		return nil, fmt.Errorf("file overview classes query failed: %w", err)
	}

	var symbols []OverviewSymbol
	symbols = append(symbols, parseOverviewItems(funcResult.Data(), "functions", parseFuncOverview)...)
	symbols = append(symbols, parseOverviewItems(classResult.Data(), "classs", parseClassOverview)...)

	// Sort by startingLine ascending
	sort.Slice(symbols, func(i, j int) bool {
		return symbols[i].StartingLine < symbols[j].StartingLine
	})

	return &FileOverviewResponse{
		Path:    path,
		Symbols: symbols,
		Total:   len(symbols),
	}, nil
}

// handleGetSymbolContext handles the get_symbol_context MCP tool call.
func (s *Server) handleGetSymbolContext(ctx context.Context, repo, name string) (*SymbolContextResponse, error) {
	if repo == "" {
		return nil, errors.New("repository is required")
	}
	if name == "" {
		return nil, errors.New("name is required")
	}

	c, err := s.requireRepoClient(ctx, repo)
	if err != nil {
		return nil, err
	}

	where := map[string]any{
		"where": map[string]any{
			"name":       name,
			"repository": map[string]any{"name": repo},
		},
	}

	symbolType, m, err := findSymbol(ctx, c, where)
	if err != nil {
		return nil, err
	}
	if m == nil {
		return nil, fmt.Errorf("symbol %q not found", name)
	}

	symbol := buildSymbolDetail(symbolType, m)
	isFunction := symbolType == "function"

	resp := SymbolContextResponse{Symbol: symbol}

	// Callers/callees only for functions
	if isFunction {
		callers, err := s.traverseHops(ctx, c, []string{name},
			gqlFindCallers, "calls_some", "name", "functions", "calls", "", 1, parseFunctionResult)
		if err == nil {
			for _, r := range callers {
				resp.Callers = append(resp.Callers, traversalToSummary(r))
			}
		}

		callees, err := s.traverseHops(ctx, c, []string{name},
			gqlFindCallees, "calledBy_some", "name", "functions", "calls", "", 1, parseFunctionResult)
		if err == nil {
			for _, r := range callees {
				resp.Callees = append(resp.Callees, traversalToSummary(r))
			}
		}
	}

	// Siblings: symbols in the same file, excluding self
	overview, err := s.handleGetFileOverview(ctx, repo, symbol.Path)
	if err == nil {
		for _, sym := range overview.Symbols {
			if sym.Name != name {
				resp.Siblings = append(resp.Siblings, sym)
			}
		}
	}

	return &resp, nil
}

// handleReadSource handles the read_source MCP tool call.
func (s *Server) handleReadSource(ctx context.Context, repo string, names []string) (*ReadSourceResponse, error) {
	if repo == "" {
		return nil, errors.New("repository is required")
	}
	if len(names) == 0 {
		return nil, errors.New("names is required")
	}

	c, err := s.requireRepoClient(ctx, repo)
	if err != nil {
		return nil, err
	}

	var results []ReadSourceResult

	for _, name := range names {
		where := map[string]any{
			"where": map[string]any{
				"name":       name,
				"repository": map[string]any{"name": repo},
			},
		}

		symbolType, m, err := findSymbol(ctx, c, where)
		if err != nil || m == nil {
			continue
		}
		results = append(results, ReadSourceResult{
			Type:         symbolType,
			Name:         strVal(m, "name"),
			Path:         strVal(m, "path"),
			Source:       strVal(m, "source"),
			StartingLine: intVal(m, "startingLine"),
			EndingLine:   intVal(m, "endingLine"),
		})
	}

	return &ReadSourceResponse{
		Results: results,
		Total:   len(results),
	}, nil
}

// traversalToSummary converts a TraversalResult to a SymbolSummary.
func traversalToSummary(r TraversalResult) SymbolSummary {
	return SymbolSummary{
		Type:      r.Type,
		Name:      r.Name,
		Path:      r.Path,
		Signature: r.Signature,
		Kind:      r.Kind,
	}
}

// findSymbol queries for a symbol by trying function first, then class.
// Returns the type ("function" or "class"), the raw map, and any error.
// If not found, returns ("", nil, nil).
func findSymbol(ctx context.Context, c *client.Client, where map[string]any) (string, map[string]any, error) {
	funcResult, err := c.Execute(ctx, gqlSymbolFunction, where)
	if err != nil {
		return "", nil, fmt.Errorf("symbol function query failed: %w", err)
	}
	if items, ok := funcResult.Data()["functions"].([]any); ok && len(items) > 0 {
		if m, ok := items[0].(map[string]any); ok {
			return "function", m, nil
		}
	}

	classResult, err := c.Execute(ctx, gqlSymbolClass, where)
	if err != nil {
		return "", nil, fmt.Errorf("symbol class query failed: %w", err)
	}
	if items, ok := classResult.Data()["classs"].([]any); ok && len(items) > 0 {
		if m, ok := items[0].(map[string]any); ok {
			return "class", m, nil
		}
	}

	return "", nil, nil
}

// buildSymbolDetail constructs a SymbolDetail from a raw GraphQL result map.
func buildSymbolDetail(symbolType string, m map[string]any) SymbolDetail {
	detail := SymbolDetail{
		Type:         symbolType,
		Name:         strVal(m, "name"),
		Path:         strVal(m, "path"),
		Source:       strVal(m, "source"),
		Language:     strVal(m, "language"),
		Visibility:   strVal(m, "visibility"),
		StartingLine: intVal(m, "startingLine"),
		EndingLine:   intVal(m, "endingLine"),
	}
	if symbolType == "function" {
		detail.Signature = strVal(m, "signature")
	} else {
		detail.Kind = strVal(m, "kind")
	}
	return detail
}

// repoWhere builds a GraphQL variables map scoped to a repository.
func repoWhere(repo string) map[string]any {
	return map[string]any{
		"where": map[string]any{
			"repository": map[string]any{"name": repo},
		},
	}
}

// countSymbolPaths counts symbol occurrences per file path from a GraphQL result.
// Increments the counts map and returns the total number of symbols counted.
func countSymbolPaths(data map[string]any, key string, counts map[string]int) int {
	total := 0
	if items, ok := data[key].([]any); ok {
		for _, item := range items {
			if m, ok := item.(map[string]any); ok {
				p := strVal(m, "path")
				counts[p]++
				total++
			}
		}
	}
	return total
}

// parseFuncOverview creates an OverviewSymbol from a function query result map.
func parseFuncOverview(m map[string]any) OverviewSymbol {
	return OverviewSymbol{
		Type:         "function",
		Name:         strVal(m, "name"),
		Signature:    strVal(m, "signature"),
		Visibility:   strVal(m, "visibility"),
		StartingLine: intVal(m, "startingLine"),
		EndingLine:   intVal(m, "endingLine"),
	}
}

// parseClassOverview creates an OverviewSymbol from a class query result map.
func parseClassOverview(m map[string]any) OverviewSymbol {
	return OverviewSymbol{
		Type:         "class",
		Name:         strVal(m, "name"),
		Kind:         strVal(m, "kind"),
		Visibility:   strVal(m, "visibility"),
		StartingLine: intVal(m, "startingLine"),
		EndingLine:   intVal(m, "endingLine"),
	}
}

// parseOverviewItems extracts items from a GraphQL result data map and
// converts each to an OverviewSymbol using the provided parse function.
func parseOverviewItems(data map[string]any, key string, parse func(map[string]any) OverviewSymbol) []OverviewSymbol {
	items, ok := data[key].([]any)
	if !ok {
		return nil
	}
	var out []OverviewSymbol
	for _, item := range items {
		if m, ok := item.(map[string]any); ok {
			out = append(out, parse(m))
		}
	}
	return out
}

// marshalMCPResult converts any response type to an MCP CallToolResult via JSON.
func marshalMCPResult(resp any) (*mcp.CallToolResult, error) {
	data, err := json.Marshal(resp)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal response: %w", err)
	}
	return mcp.NewToolResultText(string(data)), nil
}
