package rlm

import "context"

// KnowledgeGraph defines the read-only interface for graph backends.
// The REPL binds each method to a corresponding JavaScript function.
type KnowledgeGraph interface {
	// Metadata returns a constant-size summary of the graph. Called once per query.
	Metadata(ctx context.Context) (Metadata, error)

	// Search performs case-insensitive text search across node string properties.
	// If labels is non-empty, only nodes with at least one matching label are returned.
	// Limit defaults to 20 if <= 0.
	Search(ctx context.Context, query string, labels []string, limit int) ([]Node, error)

	// Neighbors returns the 1-hop neighborhood of a node.
	// If edgeType is non-empty, only edges of that type are followed.
	// Limit defaults to 50 if <= 0.
	Neighbors(ctx context.Context, nodeID string, edgeType string, limit int) (SubGraph, error)

	// GetNode retrieves a single node by ID.
	GetNode(ctx context.Context, id string) (Node, error)

	// ShortestPath finds the shortest path between two nodes up to maxHops.
	// MaxHops defaults to 6 if <= 0.
	ShortestPath(ctx context.Context, fromID, toID string, maxHops int) (SubGraph, error)

	// RunCypher executes a read-only Cypher query and returns all nodes and edges
	// from the result set. Backends that do not support Cypher (e.g. in-memory)
	// should return an error directing the caller to use typed methods.
	RunCypher(ctx context.Context, cypher string, params map[string]any) (SubGraph, error)

	// Aggregate executes a Cypher query that returns tabular (non-graph) results —
	// counts, groupings, sums. Returns rows as generic maps. Backends that do not
	// support Cypher should return an error.
	Aggregate(ctx context.Context, cypher string, params map[string]any) ([]map[string]any, error)

	// Close releases any resources held by the graph backend.
	Close(ctx context.Context) error
}
