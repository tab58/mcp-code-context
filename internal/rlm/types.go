package rlm

// Node represents a node in the knowledge graph.
type Node struct {
	ID         string         `json:"id"`
	Labels     []string       `json:"labels"`
	Properties map[string]any `json:"properties"`
}

// Edge represents a directed edge in the knowledge graph.
type Edge struct {
	ID         string         `json:"id"`
	Type       string         `json:"type"`
	SourceID   string         `json:"source_id"`
	TargetID   string         `json:"target_id"`
	Properties map[string]any `json:"properties"`
}

// SubGraph is a collection of nodes and edges returned by graph queries.
type SubGraph struct {
	Nodes []Node `json:"nodes"`
	Edges []Edge `json:"edges"`
}

// Metadata is a constant-size summary of the graph, safe to embed in a prompt.
type Metadata struct {
	TotalNodes  int            `json:"total_nodes"`
	TotalEdges  int            `json:"total_edges"`
	NodeLabels  map[string]int `json:"node_labels"`
	EdgeTypes   map[string]int `json:"edge_types"`
	SampleNodes []Node         `json:"sample_nodes"`
	SampleEdges []Edge         `json:"sample_edges"`
}

// Message represents a single message in an LLM conversation.
type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}
