package mcp

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"strings"

	"github.com/agnivade/levenshtein"
)

// GraphQL query constants for fuzzy name search.
const (
	gqlAllFunctionNames = `query($where: FunctionWhere) {
  functions(where: $where) { name path signature language startingLine endingLine }
}`

	gqlAllClassNames = `query($where: ClassWhere) {
  classs(where: $where) { name path kind language startingLine endingLine }
}`
)

// fuzzyThreshold is the minimum similarity score (1.0 - distance/maxLen)
// for a Levenshtein match to be included in results.
const fuzzyThreshold = 0.6

// stripWildcards removes glob characters (*, ?) from a query string
// so it can be used as a Levenshtein comparison target.
func stripWildcards(query string) string {
	var sb strings.Builder
	for _, r := range query {
		if r != '*' && r != '?' {
			sb.WriteRune(r)
		}
	}
	return sb.String()
}

// executeFuzzySearch queries all function/class names in a repo, computes
// Levenshtein distance against the stripped query, scores as
// 1.0 - (distance / maxLen), and returns results above fuzzyThreshold.
func (svc *Manager) executeFuzzySearch(ctx context.Context, repo, query string, limit int) (*SearchResponse, error) {
	if repo == "" {
		return nil, errors.New("repository is required")
	}

	c, err := svc.requireRepoClient(ctx, repo)
	if err != nil {
		return nil, err
	}

	stripped := stripWildcards(query)
	if stripped == "" {
		return &SearchResponse{
			Results:  nil,
			Query:    query,
			Strategy: "fuzzy",
		}, nil
	}

	vars := repoWhere(repo)

	var results []SearchResult

	// Query all functions
	funcResult, err := c.Execute(ctx, gqlAllFunctionNames, vars)
	if err != nil {
		return nil, fmt.Errorf("fuzzy search functions query failed: %w", err)
	}
	if items, ok := funcResult.Data()["functions"].([]any); ok {
		for _, item := range items {
			m, ok := item.(map[string]any)
			if !ok {
				continue
			}
			name := strVal(m, "name")
			score := fuzzyScore(stripped, name)
			if score >= fuzzyThreshold {
				results = append(results, SearchResult{
					Type:         "function",
					Name:         name,
					Path:         strVal(m, "path"),
					Signature:    strVal(m, "signature"),
					Language:     strVal(m, "language"),
					StartingLine: intVal(m, "startingLine"),
					EndingLine:   intVal(m, "endingLine"),
					Score:        score,
					Strategy:     "fuzzy",
				})
			}
		}
	}

	// Query all classes
	classResult, err := c.Execute(ctx, gqlAllClassNames, vars)
	if err != nil {
		return nil, fmt.Errorf("fuzzy search classes query failed: %w", err)
	}
	if items, ok := classResult.Data()["classs"].([]any); ok {
		for _, item := range items {
			m, ok := item.(map[string]any)
			if !ok {
				continue
			}
			name := strVal(m, "name")
			score := fuzzyScore(stripped, name)
			if score >= fuzzyThreshold {
				results = append(results, SearchResult{
					Type:     "class",
					Name:     name,
					Path:     strVal(m, "path"),
					Kind:     strVal(m, "kind"),
					Language: strVal(m, "language"),
					Score:    score,
					Strategy: "fuzzy",
				})
			}
		}
	}

	// Sort by score descending
	sort.Slice(results, func(i, j int) bool {
		return results[i].Score > results[j].Score
	})

	if limit > 0 && len(results) > limit {
		results = results[:limit]
	}

	return &SearchResponse{
		Results:  results,
		Query:    query,
		Strategy: "fuzzy",
		Total:    len(results),
	}, nil
}

// fuzzyScore computes similarity score between query and target using Levenshtein distance.
// Returns 1.0 - (distance / maxLen), where maxLen is the longer of the two strings.
func fuzzyScore(query, target string) float64 {
	dist := levenshtein.ComputeDistance(query, target)
	maxLen := len(query)
	if len(target) > maxLen {
		maxLen = len(target)
	}
	if maxLen == 0 {
		return 1.0
	}
	return 1.0 - float64(dist)/float64(maxLen)
}
