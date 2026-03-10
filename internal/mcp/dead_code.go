package mcp

import (
	"context"
	"errors"
	"fmt"
	"path/filepath"
	"strings"
)

// GraphQL query constants for dead code detection using _NONE filters.
const (
	gqlDeadFunctions = `query($where: FunctionWhere) {
  functions(where: $where) { name path signature startingLine endingLine language }
}`

	gqlDeadClasses = `query($where: ClassWhere) {
  classs(where: $where) { name path kind startingLine endingLine language }
}`

	gqlDeadModules = `query($where: ModuleWhere) {
  modules(where: $where) { name path kind startingLine endingLine language }
}`
)

// matchesExcludePattern returns true if name matches any of the comma-separated
// glob patterns in excludePatterns.
func matchesExcludePattern(name, excludePatterns string) bool {
	if excludePatterns == "" {
		return false
	}
	for _, p := range strings.Split(excludePatterns, ",") {
		p = strings.TrimSpace(p)
		if matched, _ := filepath.Match(p, name); matched {
			return true
		}
	}
	return false
}

// parseDeadCodeItems extracts DeadCodeResult items from a GraphQL result list,
// filtering by excludePatterns. nodeType is "function", "class", or "module".
func parseDeadCodeItems(items []any, nodeType, excludePatterns string) []DeadCodeResult {
	var results []DeadCodeResult
	for _, item := range items {
		m, ok := item.(map[string]any)
		if !ok {
			continue
		}
		name := strVal(m, "name")
		if matchesExcludePattern(name, excludePatterns) {
			continue
		}
		results = append(results, DeadCodeResult{
			Type:         nodeType,
			Name:         name,
			Path:         strVal(m, "path"),
			Signature:    strVal(m, "signature"),
			StartingLine: intVal(m, "startingLine"),
			EndingLine:   intVal(m, "endingLine"),
		})
	}
	return results
}

// handleFindDeadCode handles the find_dead_code MCP tool call.
// Queries functions/classes/modules with no inbound relationship edges,
// auto-excludes Go main()/init(), and supports exclude_decorated and exclude_patterns.
func (s *Server) handleFindDeadCode(ctx context.Context, repo string, excludeDecorated bool, excludePatterns string, limit int) (*DeadCodeResponse, error) {
	if repo == "" {
		return nil, errors.New("repository is required")
	}

	c, err := s.requireRepoClient(ctx, repo)
	if err != nil {
		return nil, err
	}

	var results []DeadCodeResult

	// Query dead functions (no callers, not a class method)
	funcVars := map[string]any{
		"where": map[string]any{
			"repository":   map[string]any{"name": repo},
			"calledBy_NONE": map[string]any{},
			"class_NONE":    map[string]any{},
		},
	}
	funcResult, err := c.Execute(ctx, gqlDeadFunctions, funcVars)
	if err != nil {
		return nil, fmt.Errorf("dead code functions query failed: %w", err)
	}
	if items, ok := funcResult.Data()["functions"].([]any); ok {
		for _, item := range items {
			m, ok := item.(map[string]any)
			if !ok {
				continue
			}
			name := strVal(m, "name")
			lang := strVal(m, "language")

			// Auto-exclude Go main() and init()
			if lang == "go" && (name == "main" || name == "init") {
				continue
			}

			// Exclude decorated symbols if requested
			if excludeDecorated {
				if decs, ok := m["decorators"].([]any); ok && len(decs) > 0 {
					continue
				}
			}

			if matchesExcludePattern(name, excludePatterns) {
				continue
			}

			results = append(results, DeadCodeResult{
				Type:         "function",
				Name:         name,
				Path:         strVal(m, "path"),
				Signature:    strVal(m, "signature"),
				StartingLine: intVal(m, "startingLine"),
				EndingLine:   intVal(m, "endingLine"),
			})
		}
	}

	// Query dead classes (no inheritors, no implementors)
	classVars := map[string]any{
		"where": map[string]any{
			"repository":       map[string]any{"name": repo},
			"inheritedBy_NONE": map[string]any{},
			"implementedBy_NONE": map[string]any{},
		},
	}
	classResult, err := c.Execute(ctx, gqlDeadClasses, classVars)
	if err != nil {
		return nil, fmt.Errorf("dead code classes query failed: %w", err)
	}
	if items, ok := classResult.Data()["classs"].([]any); ok {
		results = append(results, parseDeadCodeItems(items, "class", excludePatterns)...)
	}

	// Query dead modules (no importers)
	modVars := map[string]any{
		"where": map[string]any{
			"repository":    map[string]any{"name": repo},
			"importedBy_NONE": map[string]any{},
		},
	}
	modResult, err := c.Execute(ctx, gqlDeadModules, modVars)
	if err != nil {
		return nil, fmt.Errorf("dead code modules query failed: %w", err)
	}
	if items, ok := modResult.Data()["modules"].([]any); ok {
		results = append(results, parseDeadCodeItems(items, "module", excludePatterns)...)
	}

	if limit > 0 && len(results) > limit {
		results = results[:limit]
	}

	return &DeadCodeResponse{
		Repository: repo,
		Results:    results,
		Total:      len(results),
	}, nil
}
