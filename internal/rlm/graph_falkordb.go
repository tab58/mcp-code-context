package rlm

import (
	"context"
	"fmt"
	"regexp"
	"strconv"
	"strings"

	fdb "github.com/FalkorDB/falkordb-go/v2"
	"github.com/tab58/go-ormql/pkg/cypher"
	"github.com/tab58/go-ormql/pkg/driver"
)

var validCypherIdentRe = regexp.MustCompile(`^[a-zA-Z_][a-zA-Z0-9_]*$`)

// cypherWriteRe matches Cypher keywords that indicate a write operation,
// using word boundaries to avoid false positives on identifiers like "SETTINGS".
var cypherWriteRe = regexp.MustCompile(`(?i)\b(CREATE|MERGE|DELETE|SET|REMOVE|DETACH|CALL)\b`)

// FalkorGraph implements KnowledgeGraph against a FalkorDB database
// using the go-ormql driver for query execution.
type FalkorGraph struct {
	drv driver.Driver
}

// NewFalkorGraph creates a KnowledgeGraph backed by a FalkorDB driver.
func NewFalkorGraph(drv driver.Driver) *FalkorGraph {
	return &FalkorGraph{drv: drv}
}

func (g *FalkorGraph) execute(ctx context.Context, query string, params map[string]any) (driver.Result, error) {
	return g.drv.Execute(ctx, cypher.Statement{Query: query, Params: params})
}

func (g *FalkorGraph) Metadata(ctx context.Context) (Metadata, error) {
	var meta Metadata

	// Total node count
	res, err := g.execute(ctx, "MATCH (n) RETURN count(n) AS c", nil)
	if err != nil {
		return Metadata{}, fmt.Errorf("metadata node count: %w", err)
	}
	if len(res.Records) > 0 {
		meta.TotalNodes = toInt(res.Records[0].Values["c"])
	}

	// Total edge count
	res, err = g.execute(ctx, "MATCH ()-[r]->() RETURN count(r) AS c", nil)
	if err != nil {
		return Metadata{}, fmt.Errorf("metadata edge count: %w", err)
	}
	if len(res.Records) > 0 {
		meta.TotalEdges = toInt(res.Records[0].Values["c"])
	}

	// Label distribution
	meta.NodeLabels = make(map[string]int)
	res, err = g.execute(ctx, "MATCH (n) UNWIND labels(n) AS lbl RETURN lbl, count(*) AS cnt", nil)
	if err != nil {
		return Metadata{}, fmt.Errorf("metadata labels: %w", err)
	}
	for _, rec := range res.Records {
		meta.NodeLabels[fmt.Sprint(rec.Values["lbl"])] = toInt(rec.Values["cnt"])
	}

	// Edge type distribution
	meta.EdgeTypes = make(map[string]int)
	res, err = g.execute(ctx, "MATCH ()-[r]->() RETURN type(r) AS t, count(*) AS cnt", nil)
	if err != nil {
		return Metadata{}, fmt.Errorf("metadata edge types: %w", err)
	}
	for _, rec := range res.Records {
		meta.EdgeTypes[fmt.Sprint(rec.Values["t"])] = toInt(rec.Values["cnt"])
	}

	// Sample nodes (up to 5)
	res, err = g.execute(ctx, "MATCH (n) RETURN n LIMIT 5", nil)
	if err != nil {
		return Metadata{}, fmt.Errorf("metadata sample nodes: %w", err)
	}
	for _, rec := range res.Records {
		if n, ok := toRLMNode(rec.Values["n"]); ok {
			meta.SampleNodes = append(meta.SampleNodes, n)
		}
	}

	// Sample edges (up to 5)
	res, err = g.execute(ctx, "MATCH (a)-[r]->(b) RETURN r LIMIT 5", nil)
	if err != nil {
		return Metadata{}, fmt.Errorf("metadata sample edges: %w", err)
	}
	for _, rec := range res.Records {
		if e, ok := toRLMEdge(rec.Values["r"]); ok {
			meta.SampleEdges = append(meta.SampleEdges, e)
		}
	}

	return meta, nil
}

func (g *FalkorGraph) Search(ctx context.Context, query string, labels []string, limit int) ([]Node, error) {
	if limit <= 0 {
		limit = 20
	}

	q := "MATCH (n)"
	var conditions []string

	if len(labels) > 0 {
		conditions = append(conditions, "any(lbl IN labels(n) WHERE lbl IN $labels)")
	}
	if query != "" {
		conditions = append(conditions, "any(key IN keys(n) WHERE tostring(n[key]) CONTAINS $q)")
	}
	if len(conditions) > 0 {
		q += " WHERE " + strings.Join(conditions, " AND ")
	}
	q += " RETURN n LIMIT $limit"

	params := map[string]any{"q": query, "labels": labels, "limit": limit}
	res, err := g.execute(ctx, q, params)
	if err != nil {
		return nil, fmt.Errorf("search: %w", err)
	}

	var nodes []Node
	for _, rec := range res.Records {
		if n, ok := toRLMNode(rec.Values["n"]); ok {
			nodes = append(nodes, n)
		}
	}
	return nodes, nil
}

func (g *FalkorGraph) Neighbors(ctx context.Context, nodeID string, edgeType string, limit int) (SubGraph, error) {
	if limit <= 0 {
		limit = 50
	}

	id, err := strconv.ParseUint(nodeID, 10, 64)
	if err != nil {
		return SubGraph{}, fmt.Errorf("invalid node ID %q: %w", nodeID, err)
	}

	q := "MATCH (n)-[r]-(m) WHERE id(n) = $id"
	if edgeType != "" {
		if !validCypherIdentRe.MatchString(edgeType) {
			return SubGraph{}, fmt.Errorf("invalid edge type: %q", edgeType)
		}
		q = fmt.Sprintf("MATCH (n)-[r:%s]-(m) WHERE id(n) = $id", edgeType)
	}
	q += " RETURN n, r, m LIMIT $limit"

	params := map[string]any{"id": int64(id), "limit": limit}
	res, err := g.execute(ctx, q, params)
	if err != nil {
		return SubGraph{}, fmt.Errorf("neighbors: %w", err)
	}

	return collectSubGraph(res), nil
}

func (g *FalkorGraph) GetNode(ctx context.Context, nodeID string) (Node, error) {
	id, err := strconv.ParseUint(nodeID, 10, 64)
	if err != nil {
		return Node{}, fmt.Errorf("invalid node ID %q: %w", nodeID, err)
	}

	res, err := g.execute(ctx, "MATCH (n) WHERE id(n) = $id RETURN n", map[string]any{"id": int64(id)})
	if err != nil {
		return Node{}, fmt.Errorf("get node: %w", err)
	}
	if len(res.Records) == 0 {
		return Node{}, fmt.Errorf("node not found: %s", nodeID)
	}
	if n, ok := toRLMNode(res.Records[0].Values["n"]); ok {
		return n, nil
	}
	return Node{}, fmt.Errorf("node not found: %s", nodeID)
}

func (g *FalkorGraph) ShortestPath(ctx context.Context, fromID, toID string, maxHops int) (SubGraph, error) {
	if maxHops <= 0 {
		maxHops = 6
	}
	if maxHops > 20 {
		maxHops = 20
	}

	from, err := strconv.ParseUint(fromID, 10, 64)
	if err != nil {
		return SubGraph{}, fmt.Errorf("invalid from ID %q: %w", fromID, err)
	}
	to, err := strconv.ParseUint(toID, 10, 64)
	if err != nil {
		return SubGraph{}, fmt.Errorf("invalid to ID %q: %w", toID, err)
	}

	q := fmt.Sprintf(
		"MATCH p = shortestPath((a)-[*..%d]-(b)) WHERE id(a) = $from AND id(b) = $to "+
			"RETURN nodes(p) AS ns, relationships(p) AS rs", maxHops)
	params := map[string]any{"from": int64(from), "to": int64(to)}

	res, err := g.execute(ctx, q, params)
	if err != nil {
		return SubGraph{}, fmt.Errorf("shortest path: %w", err)
	}
	if len(res.Records) == 0 {
		return SubGraph{}, fmt.Errorf("no path found between %s and %s within %d hops", fromID, toID, maxHops)
	}

	var sg SubGraph
	rec := res.Records[0]
	if ns, ok := rec.Values["ns"].([]any); ok {
		for _, raw := range ns {
			if n, ok := toRLMNode(raw); ok {
				sg.Nodes = append(sg.Nodes, n)
			}
		}
	}
	if rs, ok := rec.Values["rs"].([]any); ok {
		for _, raw := range rs {
			if e, ok := toRLMEdge(raw); ok {
				sg.Edges = append(sg.Edges, e)
			}
		}
	}
	return sg, nil
}

func (g *FalkorGraph) RunCypher(ctx context.Context, query string, params map[string]any) (SubGraph, error) {
	if containsWriteOp(query) {
		return SubGraph{}, fmt.Errorf("write operations are not allowed: query contains a prohibited keyword")
	}
	res, err := g.execute(ctx, query, params)
	if err != nil {
		return SubGraph{}, fmt.Errorf("run cypher: %w", err)
	}
	return collectSubGraph(res), nil
}

func (g *FalkorGraph) Aggregate(ctx context.Context, query string, params map[string]any) ([]map[string]any, error) {
	if containsWriteOp(query) {
		return nil, fmt.Errorf("write operations are not allowed: query contains a prohibited keyword")
	}
	res, err := g.execute(ctx, query, params)
	if err != nil {
		return nil, fmt.Errorf("aggregate: %w", err)
	}
	rows := make([]map[string]any, 0, len(res.Records))
	for _, rec := range res.Records {
		rows = append(rows, rec.Values)
	}
	return rows, nil
}

func (g *FalkorGraph) Close(ctx context.Context) error {
	return g.drv.Close(ctx)
}

// --- type conversion helpers ---

func toRLMNode(val any) (Node, bool) {
	n, ok := val.(*fdb.Node)
	if !ok {
		return Node{}, false
	}
	props := make(map[string]any, len(n.Properties))
	for k, v := range n.Properties {
		// Rename "id" property to "ext_id" to avoid confusion with the
		// graph-internal node ID used by graph_neighbors, graph_get_node, etc.
		if k == "id" {
			props["ext_id"] = v
		} else {
			props[k] = v
		}
	}
	return Node{
		ID:         strconv.FormatUint(n.ID, 10),
		Labels:     n.Labels,
		Properties: props,
	}, true
}

func toRLMEdge(val any) (Edge, bool) {
	e, ok := val.(*fdb.Edge)
	if !ok {
		return Edge{}, false
	}
	props := make(map[string]any, len(e.Properties))
	for k, v := range e.Properties {
		props[k] = v
	}
	srcID := strconv.FormatUint(e.SourceNodeID(), 10)
	dstID := strconv.FormatUint(e.DestNodeID(), 10)
	return Edge{
		ID:         strconv.FormatUint(e.ID, 10),
		Type:       e.Relation,
		SourceID:   srcID,
		TargetID:   dstID,
		Properties: props,
	}, true
}

func toInt(v any) int {
	switch n := v.(type) {
	case int64:
		return int(n)
	case int:
		return n
	case float64:
		return int(n)
	case uint64:
		return int(n)
	default:
		return 0
	}
}

func collectSubGraph(res driver.Result) SubGraph {
	nodeMap := make(map[string]Node)
	edgeMap := make(map[string]Edge)

	for _, rec := range res.Records {
		for _, val := range rec.Values {
			if n, ok := toRLMNode(val); ok {
				nodeMap[n.ID] = n
			} else if e, ok := toRLMEdge(val); ok {
				edgeMap[e.ID] = e
			}
		}
	}

	sg := SubGraph{
		Nodes: make([]Node, 0, len(nodeMap)),
		Edges: make([]Edge, 0, len(edgeMap)),
	}
	for _, n := range nodeMap {
		sg.Nodes = append(sg.Nodes, n)
	}
	for _, e := range edgeMap {
		sg.Edges = append(sg.Edges, e)
	}
	return sg
}

func containsWriteOp(query string) bool {
	return cypherWriteRe.MatchString(query)
}
