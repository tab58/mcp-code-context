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

// --- Context response types ---

// RepoMapFile represents a file in a repo map directory entry.
type RepoMapFile struct {
	Name        string `json:"name"`
	Language    string `json:"language,omitempty"`
	SymbolCount int    `json:"symbolCount"`
}

// RepoMapEntry represents a directory in the repo map.
type RepoMapEntry struct {
	Directory string        `json:"directory"`
	Files     []RepoMapFile `json:"files"`
}

// RepoMapResponse is the response for get_repo_map.
type RepoMapResponse struct {
	Repository   string         `json:"repository"`
	Directories  []RepoMapEntry `json:"directories"`
	TotalFiles   int            `json:"totalFiles"`
	TotalSymbols int            `json:"totalSymbols"`
}

// OverviewSymbol represents a symbol in a file overview (no source).
type OverviewSymbol struct {
	Type         string `json:"type"`
	Name         string `json:"name"`
	Signature    string `json:"signature,omitempty"`
	Kind         string `json:"kind,omitempty"`
	Visibility   string `json:"visibility,omitempty"`
	StartingLine int    `json:"startingLine,omitempty"`
	EndingLine   int    `json:"endingLine,omitempty"`
}

// FileOverviewResponse is the response for get_file_overview.
type FileOverviewResponse struct {
	Path     string           `json:"path"`
	Language string           `json:"language,omitempty"`
	Symbols  []OverviewSymbol `json:"symbols"`
	Total    int              `json:"total"`
}

// SymbolDetail holds a symbol's full details including source.
type SymbolDetail struct {
	Type         string `json:"type"`
	Name         string `json:"name"`
	Path         string `json:"path"`
	Language     string `json:"language,omitempty"`
	Signature    string `json:"signature,omitempty"`
	Kind         string `json:"kind,omitempty"`
	Visibility   string `json:"visibility,omitempty"`
	Source       string `json:"source"`
	StartingLine int    `json:"startingLine,omitempty"`
	EndingLine   int    `json:"endingLine,omitempty"`
}

// SymbolSummary holds a lightweight symbol reference (no source).
type SymbolSummary struct {
	Type      string `json:"type"`
	Name      string `json:"name"`
	Path      string `json:"path"`
	Signature string `json:"signature,omitempty"`
	Kind      string `json:"kind,omitempty"`
}

// SymbolContextResponse is the response for get_symbol_context.
type SymbolContextResponse struct {
	Symbol   SymbolDetail     `json:"symbol"`
	Callers  []SymbolSummary  `json:"callers,omitempty"`
	Callees  []SymbolSummary  `json:"callees,omitempty"`
	Siblings []OverviewSymbol `json:"siblings,omitempty"`
}

// ReadSourceResult represents a single symbol's source code.
type ReadSourceResult struct {
	Type         string `json:"type"`
	Name         string `json:"name"`
	Path         string `json:"path"`
	Source       string `json:"source"`
	StartingLine int    `json:"startingLine,omitempty"`
	EndingLine   int    `json:"endingLine,omitempty"`
}

// ReadSourceResponse is the response for read_source.
type ReadSourceResponse struct {
	Results []ReadSourceResult `json:"results"`
	Total   int                `json:"total"`
}

// --- Dead Code + Complexity response types ---

// DeadCodeResult represents a single potentially dead symbol.
type DeadCodeResult struct {
	Type         string `json:"type"`
	Name         string `json:"name"`
	Path         string `json:"path"`
	Signature    string `json:"signature,omitempty"`
	StartingLine int    `json:"startingLine,omitempty"`
	EndingLine   int    `json:"endingLine,omitempty"`
}

// DeadCodeResponse is the response envelope for find_dead_code.
type DeadCodeResponse struct {
	Repository string           `json:"repository"`
	Results    []DeadCodeResult `json:"results"`
	Total      int              `json:"total"`
}

// ComplexityResult represents a function with its complexity score.
type ComplexityResult struct {
	Name                 string `json:"name"`
	Path                 string `json:"path"`
	Signature            string `json:"signature,omitempty"`
	CyclomaticComplexity int    `json:"cyclomaticComplexity"`
	StartingLine         int    `json:"startingLine,omitempty"`
	EndingLine           int    `json:"endingLine,omitempty"`
}

// ComplexityResponse is the response envelope for complexity tools.
type ComplexityResponse struct {
	Repository string             `json:"repository"`
	Results    []ComplexityResult  `json:"results"`
	Total      int                `json:"total"`
}

// CallChainResponse is the response envelope for find_call_chain.
type CallChainResponse struct {
	Source string            `json:"source"`
	Target string            `json:"target"`
	Path   []TraversalResult `json:"path"`
	Depth  int               `json:"depth"`
	Found  bool              `json:"found"`
}

// strategy represents a search dispatch strategy.
type strategy int

const (
	strategyFile    strategy = iota // glob/path-based file search
	strategyExact                   // exact function/class name match
	strategyFuzzy                   // Levenshtein fuzzy match (wildcards)
	strategyContent                 // content-based search within source
)
