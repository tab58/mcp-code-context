package mcp

import (
	"strings"
	"testing"
)

// === Task 4 + Task 11: classifyQuery heuristic classifier ===
//
// classifyQuery inspects a query string and classifies it into one of 2
// search strategies: file or exact. Rules are applied in priority order
// (first match wins). This is a pure function with no DB dependency.

// TestClassifyQuery_GlobChars verifies Rule 1: queries containing glob
// characters (*, ?, **) are classified as file strategy.
// Expected result: strategyFile for all glob patterns.
func TestClassifyQuery_GlobChars(t *testing.T) {
	tests := []struct {
		name  string
		query string
	}{
		{"star extension", "*.go"},
		{"double star", "src/**/*.ts"},
		{"question mark", "main.?o"},
		{"star only", "*"},
		{"double star only", "**"},
		{"star in middle", "internal/*/config.go"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := classifyQuery(tt.query)
			if got != strategyFile {
				t.Errorf("classifyQuery(%q) = %v, want strategyFile", tt.query, got)
			}
		})
	}
}

// TestClassifyQuery_PathWithExtension verifies Rule 2: queries containing
// path separators and file extensions are classified as file strategy.
// Expected result: strategyFile for paths like "cmd/main.go".
func TestClassifyQuery_PathWithExtension(t *testing.T) {
	tests := []struct {
		name  string
		query string
	}{
		{"go file path", "cmd/main.go"},
		{"ts file path", "src/utils/helpers.ts"},
		{"tsx file path", "components/App.tsx"},
		{"nested path", "internal/clients/code_db/codedb.go"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := classifyQuery(tt.query)
			if got != strategyFile {
				t.Errorf("classifyQuery(%q) = %v, want strategyFile", tt.query, got)
			}
		})
	}
}

// TestClassifyQuery_CamelCasePascalCase verifies Rule 3: single tokens in
// camelCase or PascalCase are classified as exact strategy.
// Expected result: strategyExact for identifier-like tokens.
func TestClassifyQuery_CamelCasePascalCase(t *testing.T) {
	tests := []struct {
		name  string
		query string
	}{
		{"camelCase", "getUserByID"},
		{"PascalCase", "HTTPHandler"},
		{"PascalCase impl", "SerializerImpl"},
		{"camelCase short", "myFunc"},
		{"PascalCase short", "MyClass"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := classifyQuery(tt.query)
			if got != strategyExact {
				t.Errorf("classifyQuery(%q) = %v, want strategyExact", tt.query, got)
			}
		})
	}
}

// TestClassifyQuery_SnakeCase verifies Rule 4: single tokens in snake_case
// are classified as exact strategy.
// Expected result: strategyExact for snake_case identifiers.
func TestClassifyQuery_SnakeCase(t *testing.T) {
	tests := []struct {
		name  string
		query string
	}{
		{"snake_case", "get_user_by_id"},
		{"snake_case short", "http_handler"},
		{"snake_case single underscore", "my_func"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := classifyQuery(tt.query)
			if got != strategyExact {
				t.Errorf("classifyQuery(%q) = %v, want strategyExact", tt.query, got)
			}
		})
	}
}

// TestClassifyQuery_NaturalLanguage verifies Rule 5: multi-word queries
// and natural language are classified as exact strategy (token-by-token search).
// Expected result: strategyExact for multi-word queries.
func TestClassifyQuery_NaturalLanguage(t *testing.T) {
	tests := []struct {
		name  string
		query string
	}{
		{"multi-word", "HTTP handler for authentication"},
		{"function description", "function that parses JSON"},
		{"class description", "class implementing the observer pattern"},
		{"two words", "login handler"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := classifyQuery(tt.query)
			if got != strategyExact {
				t.Errorf("classifyQuery(%q) = %v, want strategyExact", tt.query, got)
			}
		})
	}
}

// TestClassifyQuery_EdgeCases verifies edge case handling.
// Expected result: empty string -> exact, very long input -> exact.
func TestClassifyQuery_EdgeCases(t *testing.T) {
	tests := []struct {
		name     string
		query    string
		expected strategy
	}{
		{"empty string", "", strategyExact},
		{"whitespace only", "   ", strategyExact},
		{"very long input", strings.Repeat("search query ", 100), strategyExact},
		// Single lowercase word without underscores: not camelCase/PascalCase/snake_case
		{"single lowercase word", "handler", strategyExact},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := classifyQuery(tt.query)
			if got != tt.expected {
				t.Errorf("classifyQuery(%q) = %v, want %v", tt.query, got, tt.expected)
			}
		})
	}
}

// TestClassifyQuery_GlobWinsOverExact verifies that glob characters take
// priority over identifier patterns (Rule 1 > Rule 3/4).
// Expected result: "*" is classified as file, not exact.
func TestClassifyQuery_GlobWinsOverExact(t *testing.T) {
	got := classifyQuery("*")
	if got != strategyFile {
		t.Errorf("classifyQuery(\"*\") = %v, want strategyFile (glob should win over exact)", got)
	}
}
