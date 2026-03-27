package rlm

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"
)

// MemoryGraph is an in-memory KnowledgeGraph backed by Go maps.
// Suitable for testing and small knowledge graphs loaded from JSON files.
type MemoryGraph struct {
	nodes map[string]Node
	edges map[string]Edge
}

// NewMemoryGraph creates an empty in-memory graph.
func NewMemoryGraph() *MemoryGraph {
	return &MemoryGraph{
		nodes: make(map[string]Node),
		edges: make(map[string]Edge),
	}
}

// graphJSON is the on-disk format for loading graphs from JSON files.
type graphJSON struct {
	Nodes []Node `json:"nodes"`
	Edges []Edge `json:"edges"`
}

// NewMemoryGraphFromFile loads an in-memory graph from a JSON file.
func NewMemoryGraphFromFile(path string) (*MemoryGraph, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read graph file: %w", err)
	}

	var gj graphJSON
	if err := json.Unmarshal(data, &gj); err != nil {
		return nil, fmt.Errorf("parse graph file: %w", err)
	}

	g := NewMemoryGraph()
	for _, n := range gj.Nodes {
		g.nodes[n.ID] = n
	}
	for _, e := range gj.Edges {
		g.edges[e.ID] = e
	}
	return g, nil
}

func (g *MemoryGraph) Metadata(_ context.Context) (Metadata, error) {
	labelCounts := make(map[string]int)
	for _, n := range g.nodes {
		for _, l := range n.Labels {
			labelCounts[l]++
		}
	}

	edgeTypeCounts := make(map[string]int)
	for _, e := range g.edges {
		edgeTypeCounts[e.Type]++
	}

	sampleNodes := make([]Node, 0, 5)
	for _, n := range g.nodes {
		if len(sampleNodes) >= 5 {
			break
		}
		sampleNodes = append(sampleNodes, n)
	}

	sampleEdges := make([]Edge, 0, 5)
	for _, e := range g.edges {
		if len(sampleEdges) >= 5 {
			break
		}
		sampleEdges = append(sampleEdges, e)
	}

	return Metadata{
		TotalNodes:  len(g.nodes),
		TotalEdges:  len(g.edges),
		NodeLabels:  labelCounts,
		EdgeTypes:   edgeTypeCounts,
		SampleNodes: sampleNodes,
		SampleEdges: sampleEdges,
	}, nil
}

func (g *MemoryGraph) Search(_ context.Context, query string, labels []string, limit int) ([]Node, error) {
	if limit <= 0 {
		limit = 20
	}
	q := strings.ToLower(query)
	labelSet := make(map[string]bool, len(labels))
	for _, l := range labels {
		labelSet[l] = true
	}

	var results []Node
	for _, n := range g.nodes {
		if len(results) >= limit {
			break
		}
		if len(labelSet) > 0 && !hasMatchingLabel(n.Labels, labelSet) {
			continue
		}
		if q == "" || matchesProperties(n.Properties, q) {
			results = append(results, n)
		}
	}
	return results, nil
}

func hasMatchingLabel(nodeLabels []string, labelSet map[string]bool) bool {
	for _, l := range nodeLabels {
		if labelSet[l] {
			return true
		}
	}
	return false
}

func matchesProperties(props map[string]any, query string) bool {
	for _, v := range props {
		s, ok := v.(string)
		if ok && strings.Contains(strings.ToLower(s), query) {
			return true
		}
	}
	return false
}

func (g *MemoryGraph) Neighbors(_ context.Context, nodeID string, edgeType string, limit int) (SubGraph, error) {
	if limit <= 0 {
		limit = 50
	}
	if _, ok := g.nodes[nodeID]; !ok {
		return SubGraph{}, fmt.Errorf("node not found: %s", nodeID)
	}

	var matchedEdges []Edge
	neighborIDs := make(map[string]bool)

	for _, e := range g.edges {
		if len(matchedEdges) >= limit {
			break
		}
		src, tgt := e.SourceID, e.TargetID
		if src != nodeID && tgt != nodeID {
			continue
		}
		if edgeType != "" && e.Type != edgeType {
			continue
		}
		matchedEdges = append(matchedEdges, e)
		if src == nodeID {
			neighborIDs[tgt] = true
		} else {
			neighborIDs[src] = true
		}
	}

	nodes := []Node{g.nodes[nodeID]}
	for id := range neighborIDs {
		if n, ok := g.nodes[id]; ok {
			nodes = append(nodes, n)
		}
	}

	return SubGraph{Nodes: nodes, Edges: matchedEdges}, nil
}

func (g *MemoryGraph) GetNode(_ context.Context, id string) (Node, error) {
	n, ok := g.nodes[id]
	if !ok {
		return Node{}, fmt.Errorf("node not found: %s", id)
	}
	return n, nil
}

func (g *MemoryGraph) ShortestPath(_ context.Context, fromID, toID string, maxHops int) (SubGraph, error) {
	if maxHops <= 0 {
		maxHops = 6
	}
	if maxHops > 20 {
		maxHops = 20
	}
	if _, ok := g.nodes[fromID]; !ok {
		return SubGraph{}, fmt.Errorf("node not found: %s", fromID)
	}
	if _, ok := g.nodes[toID]; !ok {
		return SubGraph{}, fmt.Errorf("node not found: %s", toID)
	}

	// Build adjacency: nodeID -> []Edge
	adj := make(map[string][]Edge)
	for _, e := range g.edges {
		adj[e.SourceID] = append(adj[e.SourceID], e)
		adj[e.TargetID] = append(adj[e.TargetID], e)
	}

	// BFS
	type bfsState struct {
		nodeID string
		path   []string // alternating nodeID, edgeID, nodeID, edgeID, ...
	}
	visited := map[string]bool{fromID: true}
	queue := []bfsState{{nodeID: fromID, path: []string{fromID}}}

	for len(queue) > 0 {
		cur := queue[0]
		queue = queue[1:]

		hops := len(cur.path) / 2
		if hops >= maxHops {
			continue
		}

		for _, e := range adj[cur.nodeID] {
			neighbor := e.TargetID
			if neighbor == cur.nodeID {
				neighbor = e.SourceID
			}
			if visited[neighbor] {
				continue
			}
			visited[neighbor] = true

			newPath := make([]string, len(cur.path), len(cur.path)+2)
			copy(newPath, cur.path)
			newPath = append(newPath, e.ID, neighbor)

			if neighbor == toID {
				return g.buildPathSubGraph(newPath), nil
			}
			queue = append(queue, bfsState{nodeID: neighbor, path: newPath})
		}
	}

	return SubGraph{}, fmt.Errorf("no path found between %s and %s within %d hops", fromID, toID, maxHops)
}

func (g *MemoryGraph) buildPathSubGraph(path []string) SubGraph {
	var nodes []Node
	var edges []Edge
	for i, id := range path {
		if i%2 == 0 {
			if n, ok := g.nodes[id]; ok {
				nodes = append(nodes, n)
			}
		} else {
			if e, ok := g.edges[id]; ok {
				edges = append(edges, e)
			}
		}
	}
	return SubGraph{Nodes: nodes, Edges: edges}
}

func (g *MemoryGraph) RunCypher(_ context.Context, _ string, _ map[string]any) (SubGraph, error) {
	return SubGraph{}, fmt.Errorf("cypher queries are not supported by the in-memory graph backend; use the typed methods instead")
}

func (g *MemoryGraph) Aggregate(_ context.Context, _ string, _ map[string]any) ([]map[string]any, error) {
	return nil, fmt.Errorf("aggregate queries are not supported by the in-memory graph backend; use the typed methods instead")
}

func (g *MemoryGraph) Close(_ context.Context) error {
	return nil
}
