package mcp

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/tab58/go-ormql/pkg/client"
)

// GraphQL query constants for traversal tool handlers.
const (
	gqlFindCallers = `query($where: FunctionWhere) {
  functions(where: $where) { name path signature language }
}`

	gqlFindCallees = `query($where: FunctionWhere) {
  functions(where: $where) { name path signature language }
}`

	gqlFindParentClasses = `query($where: ClassWhere) {
  classs(where: $where) { name path kind language }
}`

	gqlFindImplementedInterfaces = `query($where: ClassWhere) {
  classs(where: $where) { name path kind language }
}`

	gqlFindChildClasses = `query($where: ClassWhere) {
  classs(where: $where) { name path kind language }
}`

	gqlFindImplementors = `query($where: ClassWhere) {
  classs(where: $where) { name path kind language }
}`

	gqlFindModuleDeps = `query($where: ModuleWhere) {
  modules(where: $where) { name path importPath language kind }
}`

	gqlFindFileImports = `query($where: ModuleWhere) {
  modules(where: $where) { name path importPath language kind }
}`

	gqlDetectFunction = `query($where: FunctionWhere) {
  functions(where: $where) { name }
}`

	gqlDetectClass = `query($where: ClassWhere) {
  classs(where: $where) { name }
}`
)

// clampDepth ensures depth is within [1, maxTraversalDepth].
// Returns 1 if depth <= 0, maxTraversalDepth if depth > max.
func clampDepth(depth int) int {
	if depth < 1 {
		return 1
	}
	if depth > maxTraversalDepth {
		return maxTraversalDepth
	}
	return depth
}

// isFilePath returns true if the name looks like a file path
// (contains "/" or has a recognized file extension).
func isFilePath(name string) bool {
	if name == "" {
		return false
	}
	if strings.Contains(name, "/") {
		return true
	}
	ext := filepath.Ext(name)
	return ext != ""
}

// parseFunctionResult creates a TraversalResult from a function query response map.
func parseFunctionResult(m map[string]any, depth int, edgeType, direction string) TraversalResult {
	return TraversalResult{
		Type:      "function",
		Name:      strVal(m, "name"),
		Path:      strVal(m, "path"),
		Signature: strVal(m, "signature"),
		Language:  strVal(m, "language"),
		Depth:     depth,
		EdgeType:  edgeType,
		Direction: direction,
	}
}

// parseClassResult creates a TraversalResult from a class query response map.
func parseClassResult(m map[string]any, depth int, edgeType, direction string) TraversalResult {
	return TraversalResult{
		Type:      "class",
		Name:      strVal(m, "name"),
		Path:      strVal(m, "path"),
		Kind:      strVal(m, "kind"),
		Language:  strVal(m, "language"),
		Depth:     depth,
		EdgeType:  edgeType,
		Direction: direction,
	}
}

// parseModuleResult creates a TraversalResult from a module query response map.
func parseModuleResult(m map[string]any, depth int, edgeType, direction string) TraversalResult {
	return TraversalResult{
		Type:      "module",
		Name:      strVal(m, "name"),
		Path:      strVal(m, "path"),
		Kind:      strVal(m, "kind"),
		Language:  strVal(m, "language"),
		Depth:     depth,
		EdgeType:  edgeType,
		Direction: direction,
	}
}

// parseFileResult creates a TraversalResult from a file query response map.
func parseFileResult(m map[string]any, depth int, edgeType, direction string) TraversalResult {
	return TraversalResult{
		Type:      "file",
		Name:      strVal(m, "path"),
		Path:      strVal(m, "path"),
		Language:  strVal(m, "language"),
		Depth:     depth,
		EdgeType:  edgeType,
		Direction: direction,
	}
}

// toTraversalMCPResult converts a TraversalResponse to an MCP CallToolResult.
func toTraversalMCPResult(resp *TraversalResponse) (*mcp.CallToolResult, error) {
	data, err := json.Marshal(resp)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal traversal response: %w", err)
	}
	return mcp.NewToolResultText(string(data)), nil
}

// traverseHops performs iterative multi-hop graph traversal.
// At each depth level it queries for neighbors matching the WHERE key,
// collects new (unvisited) names, and repeats up to maxDepth.
func (s *Server) traverseHops(
	ctx context.Context,
	c *client.Client,
	startNames []string,
	gqlQuery string,
	whereKey string,
	whereField string,
	resultKey string,
	edgeType string,
	direction string,
	maxDepth int,
	parseFunc func(map[string]any, int, string, string) TraversalResult,
) ([]TraversalResult, error) {
	visited := make(map[string]bool)
	for _, name := range startNames {
		visited[name] = true
	}

	var allResults []TraversalResult
	frontier := startNames

	for depth := 1; depth <= maxDepth; depth++ {
		if len(frontier) == 0 {
			break
		}

		var nextFrontier []string
		for _, name := range frontier {
			vars := map[string]any{
				"where": map[string]any{
					whereKey: map[string]any{whereField: name},
				},
			}

			result, err := c.Execute(ctx, gqlQuery, vars)
			if err != nil {
				return nil, fmt.Errorf("traversal query at depth %d failed: %w", depth, err)
			}

			data := result.Data()
			items, ok := data[resultKey].([]any)
			if !ok {
				continue
			}

			for _, item := range items {
				m, ok := item.(map[string]any)
				if !ok {
					continue
				}

				r := parseFunc(m, depth, edgeType, direction)
				key := r.Name + ":" + r.Path
				if visited[key] || visited[r.Name] {
					continue
				}
				visited[key] = true
				visited[r.Name] = true
				allResults = append(allResults, r)
				nextFrontier = append(nextFrontier, r.Name)
			}
		}
		frontier = nextFrontier
	}

	return allResults, nil
}

// validateRepoName returns an error if repo or name is empty.
func validateRepoName(repo, name string) error {
	if repo == "" {
		return errors.New("repository is required")
	}
	if name == "" {
		return errors.New("name is required")
	}
	return nil
}

// dedupByNamePath removes duplicate TraversalResults by name+path key,
// preserving the order of first occurrence.
func dedupByNamePath(results []TraversalResult) []TraversalResult {
	seen := make(map[string]bool)
	var out []TraversalResult
	for _, r := range results {
		key := r.Name + ":" + r.Path
		if seen[key] {
			continue
		}
		seen[key] = true
		out = append(out, r)
	}
	return out
}

// handleGetCallers handles the get_callers MCP tool call.
func (s *Server) handleGetCallers(ctx context.Context, repo, name string, depth int) (*TraversalResponse, error) {
	if err := validateRepoName(repo, name); err != nil {
		return nil, err
	}
	d := clampDepth(depth)
	c, err := s.requireRepoClient(ctx, repo)
	if err != nil {
		return nil, err
	}

	results, err := s.traverseHops(ctx, c, []string{name},
		gqlFindCallers, "calls_some", "name", "functions", "calls", "", d, parseFunctionResult)
	if err != nil {
		return nil, err
	}

	return &TraversalResponse{
		Results: results,
		Source:  name,
		Total:   len(results),
		Depth:   d,
	}, nil
}

// handleGetCallees handles the get_callees MCP tool call.
func (s *Server) handleGetCallees(ctx context.Context, repo, name string, depth int) (*TraversalResponse, error) {
	if err := validateRepoName(repo, name); err != nil {
		return nil, err
	}
	d := clampDepth(depth)
	c, err := s.requireRepoClient(ctx, repo)
	if err != nil {
		return nil, err
	}

	results, err := s.traverseHops(ctx, c, []string{name},
		gqlFindCallees, "calledBy_some", "name", "functions", "calls", "", d, parseFunctionResult)
	if err != nil {
		return nil, err
	}

	return &TraversalResponse{
		Results: results,
		Source:  name,
		Total:   len(results),
		Depth:   d,
	}, nil
}

// handleGetClassHierarchy handles the get_class_hierarchy MCP tool call.
func (s *Server) handleGetClassHierarchy(ctx context.Context, repo, name, direction string, depth int) (*TraversalResponse, error) {
	if err := validateRepoName(repo, name); err != nil {
		return nil, err
	}
	d := clampDepth(depth)
	c, err := s.requireRepoClient(ctx, repo)
	if err != nil {
		return nil, err
	}

	if direction == "" {
		direction = "both"
	}

	var results []TraversalResult

	if direction == "up" || direction == "both" {
		up1, err := s.traverseHops(ctx, c, []string{name},
			gqlFindParentClasses, "inheritedBy_some", "name", "classs", "inherits", "up", d, parseClassResult)
		if err != nil {
			return nil, err
		}
		results = append(results, up1...)

		up2, err := s.traverseHops(ctx, c, []string{name},
			gqlFindImplementedInterfaces, "implementedBy_some", "name", "classs", "implements", "up", d, parseClassResult)
		if err != nil {
			return nil, err
		}
		results = append(results, up2...)
	}

	if direction == "down" || direction == "both" {
		down1, err := s.traverseHops(ctx, c, []string{name},
			gqlFindChildClasses, "inherits_some", "name", "classs", "inherits", "down", d, parseClassResult)
		if err != nil {
			return nil, err
		}
		results = append(results, down1...)

		down2, err := s.traverseHops(ctx, c, []string{name},
			gqlFindImplementors, "implements_some", "name", "classs", "implements", "down", d, parseClassResult)
		if err != nil {
			return nil, err
		}
		results = append(results, down2...)
	}

	deduped := dedupByNamePath(results)

	return &TraversalResponse{
		Results: deduped,
		Source:  name,
		Total:   len(deduped),
		Depth:   d,
	}, nil
}

// handleGetDependencies handles the get_dependencies MCP tool call.
func (s *Server) handleGetDependencies(ctx context.Context, repo, name string, depth int) (*TraversalResponse, error) {
	if err := validateRepoName(repo, name); err != nil {
		return nil, err
	}
	d := clampDepth(depth)
	c, err := s.requireRepoClient(ctx, repo)
	if err != nil {
		return nil, err
	}

	var results []TraversalResult

	if isFilePath(name) {
		// File: query modules imported by this file (single-level)
		r, err := s.traverseHops(ctx, c, []string{name},
			gqlFindFileImports, "importedBy_some", "path", "modules", "imports", "", 1, parseModuleResult)
		if err != nil {
			return nil, err
		}
		results = r
	} else {
		// Module: query modules that depend on this module
		r, err := s.traverseHops(ctx, c, []string{name},
			gqlFindModuleDeps, "dependedOnBy_some", "name", "modules", "depends_on", "", d, parseModuleResult)
		if err != nil {
			return nil, err
		}
		results = r
	}

	return &TraversalResponse{
		Results: results,
		Source:  name,
		Total:   len(results),
		Depth:   d,
	}, nil
}

// handleGetReferences handles the get_references MCP tool call.
func (s *Server) handleGetReferences(ctx context.Context, repo, name string) (*TraversalResponse, error) {
	if err := validateRepoName(repo, name); err != nil {
		return nil, err
	}
	c, err := s.requireRepoClient(ctx, repo)
	if err != nil {
		return nil, err
	}

	var results []TraversalResult

	// Detect which types this name matches
	isFunc := s.detectExists(ctx, c, gqlDetectFunction, "functions", "name", name)
	isClass := s.detectExists(ctx, c, gqlDetectClass, "classs", "name", name)
	isModule := s.detectExists(ctx, c, gqlFindModuleDeps, "modules", "name", name)

	// Function references: calledBy + overriddenBy
	if isFunc {
		callers, err := s.traverseHops(ctx, c, []string{name},
			gqlFindCallers, "calls_some", "name", "functions", "calls", "", 1, parseFunctionResult)
		if err != nil {
			return nil, err
		}
		results = append(results, callers...)

		overriders, err := s.traverseHops(ctx, c, []string{name},
			gqlFindCallers, "overrides_some", "name", "functions", "overrides", "", 1, parseFunctionResult)
		if err != nil {
			return nil, err
		}
		results = append(results, overriders...)
	}

	// Class references: inheritedBy + implementedBy
	if isClass {
		children, err := s.traverseHops(ctx, c, []string{name},
			gqlFindChildClasses, "inherits_some", "name", "classs", "inherits", "", 1, parseClassResult)
		if err != nil {
			return nil, err
		}
		results = append(results, children...)

		impls, err := s.traverseHops(ctx, c, []string{name},
			gqlFindImplementors, "implements_some", "name", "classs", "implements", "", 1, parseClassResult)
		if err != nil {
			return nil, err
		}
		results = append(results, impls...)
	}

	// Module references: dependsOn + imports
	if isModule {
		deps, err := s.traverseHops(ctx, c, []string{name},
			gqlFindModuleDeps, "dependedOnBy_some", "name", "modules", "depends_on", "", 1, parseModuleResult)
		if err != nil {
			return nil, err
		}
		results = append(results, deps...)
	}

	return &TraversalResponse{
		Results: results,
		Source:  name,
		Total:   len(results),
		Depth:   1,
	}, nil
}

// detectExists checks if a node with the given name exists by querying
// and checking if the result list is non-empty.
func (s *Server) detectExists(ctx context.Context, c *client.Client, gqlQuery, resultKey, field, name string) bool {
	vars := map[string]any{
		"where": map[string]any{field: name},
	}
	result, err := c.Execute(ctx, gqlQuery, vars)
	if err != nil {
		return false
	}
	data := result.Data()
	items, ok := data[resultKey].([]any)
	return ok && len(items) > 0
}
