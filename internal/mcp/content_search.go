package mcp

import (
	"context"
	"errors"
	"fmt"
	"strings"
)

// GraphQL query constants for content-based search.
// Queries function/class nodes WITH source for substring matching,
// then returns matched symbols directly (no separate container resolution needed).
const (
	gqlFunctionsWithSource = `query($where: FunctionWhere) {
  functions(where: $where) { name path signature source startingLine endingLine language }
}`

	gqlClassesWithSource = `query($where: ClassWhere) {
  classs(where: $where) { name path kind source startingLine endingLine language }
}`
)

// handleSearchCodeContent handles the content-based search_code tool.
// Searches within function/class source code for substring matches and
// returns matched symbols directly.
func (s *Server) handleSearchCodeContent(ctx context.Context, repo, query string, limit int) (*SearchResponse, error) {
	if repo == "" {
		return nil, errors.New("repository is required")
	}
	if query == "" {
		return nil, errors.New("query is required")
	}

	c, err := s.requireRepoClient(ctx, repo)
	if err != nil {
		return nil, err
	}

	vars := repoWhere(repo)
	seen := make(map[string]bool)
	var results []SearchResult

	// Query functions with source and search for substring matches
	funcResult, err := c.Execute(ctx, gqlFunctionsWithSource, vars)
	if err != nil {
		return nil, fmt.Errorf("content search functions query failed: %w", err)
	}
	if items, ok := funcResult.Data()["functions"].([]any); ok {
		for _, item := range items {
			m, ok := item.(map[string]any)
			if !ok {
				continue
			}
			source := strVal(m, "source")
			if source == "" || !strings.Contains(source, query) {
				continue
			}
			name := strVal(m, "name")
			path := strVal(m, "path")
			key := "function:" + name + ":" + path
			if !seen[key] {
				seen[key] = true
				results = append(results, SearchResult{
					Type:         "function",
					Name:         name,
					Path:         path,
					Signature:    strVal(m, "signature"),
					StartingLine: intVal(m, "startingLine"),
					EndingLine:   intVal(m, "endingLine"),
					Language:     strVal(m, "language"),
					Score:        0.8,
					Strategy:     "content",
				})
			}
		}
	}

	// Query classes with source and search for substring matches
	classResult, err := c.Execute(ctx, gqlClassesWithSource, vars)
	if err != nil {
		return nil, fmt.Errorf("content search classes query failed: %w", err)
	}
	if items, ok := classResult.Data()["classs"].([]any); ok {
		for _, item := range items {
			m, ok := item.(map[string]any)
			if !ok {
				continue
			}
			source := strVal(m, "source")
			if source == "" || !strings.Contains(source, query) {
				continue
			}
			name := strVal(m, "name")
			path := strVal(m, "path")
			key := "class:" + name + ":" + path
			if !seen[key] {
				seen[key] = true
				results = append(results, SearchResult{
					Type:         "class",
					Name:         name,
					Path:         path,
					Kind:         strVal(m, "kind"),
					StartingLine: intVal(m, "startingLine"),
					EndingLine:   intVal(m, "endingLine"),
					Language:     strVal(m, "language"),
					Score:        0.8,
					Strategy:     "content",
				})
			}
		}
	}

	if limit > 0 && len(results) > limit {
		results = results[:limit]
	}

	return &SearchResponse{
		Results:  results,
		Query:    query,
		Strategy: "content",
		Total:    len(results),
	}, nil
}
