package mcp

import (
	"context"
	"errors"
	"fmt"

	"github.com/tab58/go-ormql/pkg/client"
)

// maxCallChainDepth is the maximum allowed depth for call chain analysis.
const maxCallChainDepth = 5

// HandleFindCallChain finds the call path between two functions using
// bidirectional BFS. Expands callees from source and callers from target
// simultaneously, checking frontier intersection at each depth level.
func (svc *Manager) HandleFindCallChain(ctx context.Context, repo, source, target string, maxDepth int) (*CallChainResponse, error) {
	if repo == "" {
		return nil, errors.New("repository is required")
	}
	if source == "" {
		return nil, errors.New("source function is required")
	}
	if target == "" {
		return nil, errors.New("target function is required")
	}

	// Same function: trivial path
	if source == target {
		return &CallChainResponse{
			Source: source,
			Target: target,
			Found:  true,
			Depth:  0,
		}, nil
	}

	// Clamp depth
	if maxDepth <= 0 || maxDepth > maxCallChainDepth {
		maxDepth = maxCallChainDepth
	}

	c, err := svc.requireRepoClient(ctx, repo)
	if err != nil {
		return nil, err
	}

	// Forward frontier: names reachable from source via callees
	forwardVisited := map[string]bool{source: true}
	forwardFrontier := []string{source}

	// Backward frontier: names reaching target via callers
	backwardVisited := map[string]bool{target: true}
	backwardFrontier := []string{target}

	resp := CallChainResponse{Source: source, Target: target}

	for depth := 1; depth <= maxDepth; depth++ {
		// Expand forward
		next, meetPoint, err := expandFrontier(ctx, c, forwardFrontier, forwardVisited, backwardVisited, gqlFindCallees, "forward", depth)
		if err != nil {
			return nil, err
		}
		forwardFrontier = next
		if meetPoint != "" {
			resp.Found, resp.Depth, resp.Path = true, depth, buildChainPath(meetPoint)
			return &resp, nil
		}
		if forwardVisited[target] {
			resp.Found, resp.Depth = true, depth
			return &resp, nil
		}

		// Expand backward
		next, meetPoint, err = expandFrontier(ctx, c, backwardFrontier, backwardVisited, forwardVisited, gqlFindCallers, "backward", depth)
		if err != nil {
			return nil, err
		}
		backwardFrontier = next
		if meetPoint != "" {
			resp.Found, resp.Depth, resp.Path = true, depth, buildChainPath(meetPoint)
			return &resp, nil
		}
		if backwardVisited[source] {
			resp.Found, resp.Depth = true, depth
			return &resp, nil
		}
	}

	return &resp, nil
}

// expandFrontier expands one BFS frontier by querying connected functions.
func expandFrontier(ctx context.Context, c *client.Client, frontier []string, visited, opposite map[string]bool, query, direction string, depth int) ([]string, string, error) {
	if len(frontier) == 0 {
		return nil, "", nil
	}
	var next []string
	for _, name := range frontier {
		vars := map[string]any{
			"where": map[string]any{
				"calls_some": map[string]any{"name": name},
			},
		}
		result, err := c.Execute(ctx, query, vars)
		if err != nil {
			return nil, "", fmt.Errorf("call chain %s query at depth %d: %w", direction, depth, err)
		}
		items, ok := result.Data()["functions"].([]any)
		if !ok {
			continue
		}
		for _, item := range items {
			m, ok := item.(map[string]any)
			if !ok {
				continue
			}
			n, _ := m["name"].(string)
			if n == "" || visited[n] {
				continue
			}
			visited[n] = true
			next = append(next, n)
			if opposite[n] {
				return next, n, nil
			}
		}
	}
	return next, "", nil
}

// buildChainPath creates a minimal path with an intermediate node.
func buildChainPath(meetPoint string) []TraversalResult {
	return []TraversalResult{
		{Name: meetPoint},
	}
}
