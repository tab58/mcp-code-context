package mcp

import (
	"context"
	"errors"
	"fmt"
	"path/filepath"
	"sort"
	"strings"

	"github.com/tab58/go-ormql/pkg/client"
)

// defaultLimit is the default number of results returned by search tools.
const defaultLimit = 10

// GraphQL query constants for tool handlers.
const (
	gqlFindFunctions = `query($where: FunctionWhere) {
  functions(where: $where) { name path source signature language visibility startingLine endingLine }
}`

	gqlListFiles = `query($where: FileWhere) {
  files(where: $where) { path language lineCount }
}`

	gqlFindFunctionsForFile = `query($where: FunctionWhere) {
  functions(where: $where) { name }
}`

	gqlFindClassesForFile = `query($where: ClassWhere) {
  classs(where: $where) { name }
}`
)

// handleFindFunction handles the find_function MCP tool call.
// Queries functions by exact name within a repository.
func (s *Server) handleFindFunction(ctx context.Context, repo, name string) (*SearchResponse, error) {
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

	vars := map[string]any{
		"where": map[string]any{
			"name":       name,
			"repository": map[string]any{"name": repo},
		},
	}

	result, err := c.Execute(ctx, gqlFindFunctions, vars)
	if err != nil {
		return nil, fmt.Errorf("find_function query failed: %w", err)
	}

	data := result.Data()
	var results []SearchResult
	if items, ok := data["functions"].([]any); ok {
		for _, item := range items {
			m, ok := item.(map[string]any)
			if !ok {
				continue
			}
			results = append(results, SearchResult{
				Type:         "function",
				Name:         strVal(m, "name"),
				Path:         strVal(m, "path"),
				Source:       strVal(m, "source"),
				Signature:    strVal(m, "signature"),
				Language:     strVal(m, "language"),
				Visibility:   strVal(m, "visibility"),
				StartingLine: intVal(m, "startingLine"),
				EndingLine:   intVal(m, "endingLine"),
				Score:        1.0,
				Strategy:     "exact",
			})
		}
	}

	return &SearchResponse{
		Results:  results,
		Query:    name,
		Strategy: "exact",
		Total:    len(results),
	}, nil
}

// handleFindFile handles the find_file MCP tool call.
// Glob matches file paths within a repository.
func (s *Server) handleFindFile(ctx context.Context, repo, pattern string) (*SearchResponse, error) {
	if repo == "" {
		return nil, errors.New("repository is required")
	}
	if pattern == "" {
		return nil, errors.New("pattern is required")
	}

	c, err := s.requireRepoClient(ctx, repo)
	if err != nil {
		return nil, err
	}

	vars := map[string]any{
		"where": map[string]any{
			"repository": map[string]any{"name": repo},
		},
	}

	result, err := c.Execute(ctx, gqlListFiles, vars)
	if err != nil {
		return nil, fmt.Errorf("find_file query failed: %w", err)
	}

	data := result.Data()
	var matched []SearchResult
	if items, ok := data["files"].([]any); ok {
		for _, item := range items {
			m, ok := item.(map[string]any)
			if !ok {
				continue
			}
			path := strVal(m, "path")

			ok, _ = filepath.Match(pattern, path)
			if !ok {
				// Also try matching just the filename
				ok, _ = filepath.Match(pattern, filepath.Base(path))
			}
			if !ok {
				continue
			}

			matched = append(matched, SearchResult{
				Type:     "file",
				Name:     filepath.Base(path),
				Path:     path,
				Language: strVal(m, "language"),
				Score:    0.9,
				Strategy: "file",
			})
		}
	}

	// Symbol enrichment for <=5 matched files
	if len(matched) > 0 && len(matched) <= 5 {
		for i := range matched {
			symbols := s.fetchSymbolsForFile(ctx, repo, matched[i].Path)
			if len(symbols) > 0 {
				matched[i].Symbols = symbols
			}
		}
	}

	return &SearchResponse{
		Results:  matched,
		Query:    pattern,
		Strategy: "file",
		Total:    len(matched),
	}, nil
}

// fetchSymbolsForFile queries function and class names defined in a file.
func (s *Server) fetchSymbolsForFile(ctx context.Context, repo, path string) []string {
	c, err := s.db.ForRepo(ctx, repo)
	if err != nil {
		return nil
	}

	var symbols []string
	where := map[string]any{
		"where": map[string]any{
			"path":       path,
			"repository": map[string]any{"name": repo},
		},
	}

	// Fetch functions
	if result, err := c.Execute(ctx, gqlFindFunctionsForFile, where); err == nil {
		data := result.Data()
		if items, ok := data["functions"].([]any); ok {
			for _, item := range items {
				if m, ok := item.(map[string]any); ok {
					symbols = append(symbols, strVal(m, "name"))
				}
			}
		}
	}

	// Fetch classes
	if result, err := c.Execute(ctx, gqlFindClassesForFile, where); err == nil {
		data := result.Data()
		if items, ok := data["classs"].([]any); ok {
			for _, item := range items {
				if m, ok := item.(map[string]any); ok {
					symbols = append(symbols, strVal(m, "name"))
				}
			}
		}
	}

	return symbols
}

// handleSearchCode handles the search_code MCP tool call.
// Classifies query, dispatches to appropriate strategy.
func (s *Server) handleSearchCode(ctx context.Context, repo, query string, limit int) (*SearchResponse, error) {
	if repo == "" {
		return nil, errors.New("repository is required")
	}
	if query == "" {
		return nil, errors.New("query is required")
	}
	if limit <= 0 {
		limit = defaultLimit
	}

	strat := classifyQuery(query)

	switch strat {
	case strategyFile:
		return s.handleFindFile(ctx, repo, query)

	case strategyExact:
		resp, err := s.handleFindFunction(ctx, repo, query)
		if err != nil {
			return nil, err
		}
		if len(resp.Results) > 0 {
			return resp, nil
		}
		// Fall back to exact supplement on each token
		return s.executeExactSupplement(ctx, repo, query, limit)

	default:
		return s.executeExactSupplement(ctx, repo, query, limit)
	}
}

// executeExactSupplement runs exact match on each single-word token, deduplicates, and ranks.
func (s *Server) executeExactSupplement(ctx context.Context, repo, query string, limit int) (*SearchResponse, error) {
	var results []SearchResult
	tokens := strings.Fields(query)
	for _, token := range tokens {
		resp, err := s.handleFindFunction(ctx, repo, token)
		if err != nil {
			continue
		}
		results = append(results, resp.Results...)
	}

	// Deduplicate by path+name (keep higher score)
	seen := make(map[string]int)
	var merged []SearchResult
	for _, r := range results {
		key := r.dedupKey()
		if idx, exists := seen[key]; exists {
			if r.Score > merged[idx].Score {
				merged[idx].Score = r.Score
			}
		} else {
			seen[key] = len(merged)
			merged = append(merged, r)
		}
	}

	merged = rankAndTruncate(merged, limit)

	return &SearchResponse{
		Results:  merged,
		Query:    query,
		Strategy: "exact",
		Total:    len(merged),
	}, nil
}

// rankAndTruncate sorts results by score descending and truncates to limit.
func rankAndTruncate(results []SearchResult, limit int) []SearchResult {
	sort.Slice(results, func(i, j int) bool {
		return results[i].Score > results[j].Score
	})
	if len(results) > limit {
		results = results[:limit]
	}
	return results
}

// requireRepoClient returns a repo-scoped client via ForRepo, or an error.
func (s *Server) requireRepoClient(ctx context.Context, repo string) (*client.Client, error) {
	c, err := s.db.ForRepo(ctx, repo)
	if err != nil {
		return nil, fmt.Errorf("database connection for %s: %w", repo, err)
	}
	return c, nil
}

// strVal safely extracts a string value from a map.
func strVal(m map[string]any, key string) string {
	if v, ok := m[key].(string); ok {
		return v
	}
	return ""
}

// intVal safely extracts an int value from a map (handles float64 from JSON).
func intVal(m map[string]any, key string) int {
	if v, ok := m[key].(float64); ok {
		return int(v)
	}
	if v, ok := m[key].(int); ok {
		return v
	}
	return 0
}

