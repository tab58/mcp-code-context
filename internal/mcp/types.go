package mcp

// SearchResult represents a single result item returned by any search tool.
type SearchResult struct {
	Type       string   `json:"type"`
	Name       string   `json:"name"`
	Path       string   `json:"path"`
	Language   string   `json:"language,omitempty"`
	Signature  string   `json:"signature,omitempty"`
	Kind       string   `json:"kind,omitempty"`
	Visibility string   `json:"visibility,omitempty"`
	Source     string   `json:"source,omitempty"`
	StartingLine int      `json:"startingLine,omitempty"`
	EndingLine   int      `json:"endingLine,omitempty"`
	Score      float64  `json:"score"`
	Strategy   string   `json:"strategy"`
	Symbols    []string `json:"symbols,omitempty"`
}

// dedupKey returns a unique key for deduplication by path and name.
func (r SearchResult) dedupKey() string {
	return r.Path + ":" + r.Name
}

// SearchResponse is the structured response envelope for search tools.
type SearchResponse struct {
	Results  []SearchResult `json:"results"`
	Query    string         `json:"query"`
	Strategy string         `json:"strategy"`
	Total    int            `json:"total"`
}

// maxTraversalDepth is the maximum allowed depth for multi-hop traversal.
const maxTraversalDepth = 3

// TraversalResult represents a single node found via graph relationship traversal.
// Lightweight: name+path+signature, no source code.
type TraversalResult struct {
	Type      string `json:"type"`
	Name      string `json:"name"`
	Path      string `json:"path"`
	Signature string `json:"signature,omitempty"`
	Kind      string `json:"kind,omitempty"`
	Language  string `json:"language,omitempty"`
	Depth     int    `json:"depth"`
	EdgeType  string `json:"edgeType,omitempty"`
	Direction string `json:"direction,omitempty"`
}

// TraversalResponse is the response envelope for graph traversal tools.
type TraversalResponse struct {
	Results []TraversalResult `json:"results"`
	Source  string            `json:"source"`
	Total   int               `json:"total"`
	Depth   int               `json:"depth"`
}

// strategy represents a search dispatch strategy.
type strategy int

const (
	strategyFile  strategy = iota // glob/path-based file search
	strategyExact                 // exact function/class name match
)
