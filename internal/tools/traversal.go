package mcp

import (
	"context"
	"errors"
	"fmt"
	"path/filepath"
	"strings"

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

// parseNodeResult creates a TraversalResult from a query response map.
func parseNodeResult(nodeType string) func(map[string]any, int, string, string) TraversalResult {
	return func(m map[string]any, depth int, edgeType, direction string) TraversalResult {
		r := TraversalResult{
			Type:      nodeType,
			Path:      strVal(m, "path"),
			Language:  strVal(m, "language"),
			Depth:     depth,
			EdgeType:  edgeType,
			Direction: direction,
		}
		switch nodeType {
		case "function":
			r.Name = strVal(m, "name")
			r.Signature = strVal(m, "signature")
		case "file":
			r.Name = strVal(m, "path")
		default: // class, module
			r.Name = strVal(m, "name")
			r.Kind = strVal(m, "kind")
		}
		return r
	}
}

// Pre-built parse functions for each node type.
var (
	parseFunctionResult = parseNodeResult("function")
	parseClassResult    = parseNodeResult("class")
	parseModuleResult   = parseNodeResult("module")
	parseFileResult     = parseNodeResult("file")
)

// traverseHops performs iterative multi-hop graph traversal.
func (svc *Manager) traverseHops(
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

// dedupByNamePath removes duplicate TraversalResults by name+path key.
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

// HandleGetCallers handles the get_callers tool call.
func (svc *Manager) HandleGetCallers(ctx context.Context, repo, name string, depth int) (*TraversalResponse, error) {
	if err := validateRepoName(repo, name); err != nil {
		return nil, err
	}
	d := clampDepth(depth)
	c, err := svc.requireRepoClient(ctx, repo)
	if err != nil {
		return nil, err
	}

	results, err := svc.traverseHops(ctx, c, []string{name},
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

// HandleGetCallees handles the get_callees tool call.
func (svc *Manager) HandleGetCallees(ctx context.Context, repo, name string, depth int) (*TraversalResponse, error) {
	if err := validateRepoName(repo, name); err != nil {
		return nil, err
	}
	d := clampDepth(depth)
	c, err := svc.requireRepoClient(ctx, repo)
	if err != nil {
		return nil, err
	}

	results, err := svc.traverseHops(ctx, c, []string{name},
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

// HandleGetClassHierarchy handles the get_class_hierarchy tool call.
func (svc *Manager) HandleGetClassHierarchy(ctx context.Context, repo, name, direction string, depth int) (*TraversalResponse, error) {
	if err := validateRepoName(repo, name); err != nil {
		return nil, err
	}
	d := clampDepth(depth)
	c, err := svc.requireRepoClient(ctx, repo)
	if err != nil {
		return nil, err
	}

	if direction == "" {
		direction = "both"
	}

	var results []TraversalResult

	if direction == "up" || direction == "both" {
		up1, err := svc.traverseHops(ctx, c, []string{name},
			gqlFindParentClasses, "inheritedBy_some", "name", "classs", "inherits", "up", d, parseClassResult)
		if err != nil {
			return nil, err
		}
		results = append(results, up1...)

		up2, err := svc.traverseHops(ctx, c, []string{name},
			gqlFindImplementedInterfaces, "implementedBy_some", "name", "classs", "implements", "up", d, parseClassResult)
		if err != nil {
			return nil, err
		}
		results = append(results, up2...)
	}

	if direction == "down" || direction == "both" {
		down1, err := svc.traverseHops(ctx, c, []string{name},
			gqlFindChildClasses, "inherits_some", "name", "classs", "inherits", "down", d, parseClassResult)
		if err != nil {
			return nil, err
		}
		results = append(results, down1...)

		down2, err := svc.traverseHops(ctx, c, []string{name},
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

// HandleGetDependencies handles the get_dependencies tool call.
func (svc *Manager) HandleGetDependencies(ctx context.Context, repo, name string, depth int) (*TraversalResponse, error) {
	if err := validateRepoName(repo, name); err != nil {
		return nil, err
	}
	d := clampDepth(depth)
	c, err := svc.requireRepoClient(ctx, repo)
	if err != nil {
		return nil, err
	}

	var results []TraversalResult

	if isFilePath(name) {
		r, err := svc.traverseHops(ctx, c, []string{name},
			gqlFindFileImports, "importedBy_some", "path", "modules", "imports", "", 1, parseModuleResult)
		if err != nil {
			return nil, err
		}
		results = r
	} else {
		r, err := svc.traverseHops(ctx, c, []string{name},
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

// HandleGetReferences handles the get_references tool call.
func (svc *Manager) HandleGetReferences(ctx context.Context, repo, name string) (*TraversalResponse, error) {
	if err := validateRepoName(repo, name); err != nil {
		return nil, err
	}
	c, err := svc.requireRepoClient(ctx, repo)
	if err != nil {
		return nil, err
	}

	var results []TraversalResult

	// Detect which types this name matches
	isFunc := svc.detectExists(ctx, c, gqlDetectFunction, "functions", "name", name)
	isClass := svc.detectExists(ctx, c, gqlDetectClass, "classs", "name", name)
	isModule := svc.detectExists(ctx, c, gqlFindModuleDeps, "modules", "name", name)

	// Function references: calledBy + overriddenBy
	if isFunc {
		callers, err := svc.traverseHops(ctx, c, []string{name},
			gqlFindCallers, "calls_some", "name", "functions", "calls", "", 1, parseFunctionResult)
		if err != nil {
			return nil, err
		}
		results = append(results, callers...)

		overriders, err := svc.traverseHops(ctx, c, []string{name},
			gqlFindCallers, "overrides_some", "name", "functions", "overrides", "", 1, parseFunctionResult)
		if err != nil {
			return nil, err
		}
		results = append(results, overriders...)
	}

	// Class references: inheritedBy + implementedBy
	if isClass {
		children, err := svc.traverseHops(ctx, c, []string{name},
			gqlFindChildClasses, "inherits_some", "name", "classs", "inherits", "", 1, parseClassResult)
		if err != nil {
			return nil, err
		}
		results = append(results, children...)

		impls, err := svc.traverseHops(ctx, c, []string{name},
			gqlFindImplementors, "implements_some", "name", "classs", "implements", "", 1, parseClassResult)
		if err != nil {
			return nil, err
		}
		results = append(results, impls...)
	}

	// Module references: dependsOn + imports
	if isModule {
		deps, err := svc.traverseHops(ctx, c, []string{name},
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

// detectExists checks if a node with the given name exists.
func (svc *Manager) detectExists(ctx context.Context, c *client.Client, gqlQuery, resultKey, field, name string) bool {
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
