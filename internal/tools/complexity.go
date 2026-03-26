package mcp

import (
	"context"
	"errors"
	"fmt"
	"sort"
)

// gqlFunctionComplexity queries functions with their pre-computed complexity scores.
const gqlFunctionComplexity = `query($where: FunctionWhere) {
  functions(where: $where) { name path signature cyclomaticComplexity startingLine endingLine }
}`

// parseComplexityResult extracts a ComplexityResult from a GraphQL result map.
func parseComplexityResult(m map[string]any) ComplexityResult {
	return ComplexityResult{
		Name:                 strVal(m, "name"),
		Path:                 strVal(m, "path"),
		Signature:            strVal(m, "signature"),
		CyclomaticComplexity: intVal(m, "cyclomaticComplexity"),
		StartingLine:         intVal(m, "startingLine"),
		EndingLine:           intVal(m, "endingLine"),
	}
}

// HandleCalculateCyclomaticComplexity handles the calculate_cyclomatic_complexity tool.
func (svc *Manager) HandleCalculateCyclomaticComplexity(ctx context.Context, repo, name string) (*ComplexityResponse, error) {
	if repo == "" {
		return nil, errors.New("repository is required")
	}
	if name == "" {
		return nil, errors.New("name is required")
	}

	c, err := svc.requireRepoClient(ctx, repo)
	if err != nil {
		return nil, err
	}

	vars := map[string]any{
		"where": map[string]any{
			"name":       name,
			"repository": map[string]any{"name": repo},
		},
	}

	result, err := c.Execute(ctx, gqlFunctionComplexity, vars)
	if err != nil {
		return nil, fmt.Errorf("complexity query failed: %w", err)
	}

	var results []ComplexityResult
	if items, ok := result.Data()["functions"].([]any); ok {
		for _, item := range items {
			m, ok := item.(map[string]any)
			if !ok {
				continue
			}
			results = append(results, parseComplexityResult(m))
		}
	}

	return &ComplexityResponse{
		Repository: repo,
		Results:    results,
		Total:      len(results),
	}, nil
}

// HandleFindMostComplexFunctions handles the find_most_complex_functions tool.
func (svc *Manager) HandleFindMostComplexFunctions(ctx context.Context, repo string, minComplexity, limit int) (*ComplexityResponse, error) {
	if repo == "" {
		return nil, errors.New("repository is required")
	}

	c, err := svc.requireRepoClient(ctx, repo)
	if err != nil {
		return nil, err
	}

	vars := map[string]any{
		"where": map[string]any{
			"repository": map[string]any{"name": repo},
		},
	}

	result, err := c.Execute(ctx, gqlFunctionComplexity, vars)
	if err != nil {
		return nil, fmt.Errorf("complexity query failed: %w", err)
	}

	var results []ComplexityResult
	if items, ok := result.Data()["functions"].([]any); ok {
		for _, item := range items {
			m, ok := item.(map[string]any)
			if !ok {
				continue
			}
			cr := parseComplexityResult(m)
			if cr.CyclomaticComplexity < minComplexity {
				continue
			}
			results = append(results, cr)
		}
	}

	// Sort descending by complexity
	sort.Slice(results, func(i, j int) bool {
		return results[i].CyclomaticComplexity > results[j].CyclomaticComplexity
	})

	if limit > 0 && len(results) > limit {
		results = results[:limit]
	}

	return &ComplexityResponse{
		Repository: repo,
		Results:    results,
		Total:      len(results),
	}, nil
}
