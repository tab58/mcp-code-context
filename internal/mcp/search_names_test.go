package mcp

import (
	"context"
	"testing"

	"github.com/tab58/go-ormql/pkg/driver"
)

// --- Task 5: Rename search_code to search_code_names + add Levenshtein fuzzy ---

// TestClassifyQuery_GlobReturnsFuzzy verifies that after the search refactoring,
// glob characters (*, ?) classify as strategyFuzzy instead of strategyFile.
// Expected result: strategyFuzzy for all glob patterns.
func TestClassifyQuery_GlobReturnsFuzzy(t *testing.T) {
	tests := []struct {
		name  string
		query string
	}{
		{"star extension", "*.go"},
		{"question mark", "main.?o"},
		{"star only", "*"},
		{"star in middle", "get*Handler"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := classifyQuery(tt.query)
			if got != strategyFuzzy {
				t.Errorf("classifyQuery(%q) = %v, want strategyFuzzy", tt.query, got)
			}
		})
	}
}

// TestStripWildcards verifies that glob characters are removed from the query.
// Expected result: "get*Handler" -> "getHandler", "*.go" -> ".go", "*" -> "".
func TestStripWildcards(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"star in middle", "get*Handler", "getHandler"},
		{"star extension", "*.go", ".go"},
		{"star only", "*", ""},
		{"question mark", "main.?o", "main.o"},
		{"no wildcards", "handler", "handler"},
		{"multiple wildcards", "get*User*By?ID", "getUserByID"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := stripWildcards(tt.input)
			if got != tt.expected {
				t.Errorf("stripWildcards(%q) = %q, want %q", tt.input, got, tt.expected)
			}
		})
	}
}

// TestExecuteFuzzySearch_RequiresRepository verifies that fuzzy search
// returns an error when repository is empty.
// Expected result: Error with "repository is required".
func TestExecuteFuzzySearch_RequiresRepository(t *testing.T) {
	srv := newTestServer(t)
	_, err := srv.executeFuzzySearch(context.Background(), "", "getUser", 10)
	if err == nil {
		t.Fatal("executeFuzzySearch with empty repo returned nil error")
	}
}

// TestExecuteFuzzySearch_ReturnsFuzzyMatches verifies that fuzzy search
// queries all functions/classes and returns Levenshtein-scored results.
// Expected result: Results with score > fuzzyThreshold, strategy "fuzzy".
func TestExecuteFuzzySearch_ReturnsFuzzyMatches(t *testing.T) {
	responses := []driver.Result{
		// All functions in repo
		makeResult(map[string]any{
			"functions": []any{
				map[string]any{"name": "getUser", "path": "a.go", "signature": "func getUser()"},
				map[string]any{"name": "getUserByID", "path": "b.go", "signature": "func getUserByID()"},
				map[string]any{"name": "deleteItem", "path": "c.go", "signature": "func deleteItem()"},
			},
		}),
		// All classes in repo
		makeResult(map[string]any{
			"classs": []any{},
		}),
	}

	srv, _ := newTestServerWithResponses(t, responses)
	resp, err := srv.executeFuzzySearch(context.Background(), "myrepo", "getUser", 10)
	if err != nil {
		t.Fatalf("executeFuzzySearch returned error: %v", err)
	}
	if resp == nil {
		t.Fatal("executeFuzzySearch returned nil response")
	}
	if resp.Strategy != "fuzzy" {
		t.Errorf("strategy = %q, want \"fuzzy\"", resp.Strategy)
	}
	if len(resp.Results) == 0 {
		t.Error("expected at least 1 fuzzy match result")
	}
	// "getUser" should match "getUser" exactly (score ~1.0)
	// "getUserByID" should match with lower score
	// "deleteItem" should NOT match (distance too high)
	for _, r := range resp.Results {
		if r.Score < fuzzyThreshold {
			t.Errorf("result %q has score %f below threshold %f", r.Name, r.Score, fuzzyThreshold)
		}
	}
}

// TestExecuteFuzzySearch_RespectsLimit verifies that fuzzy search
// truncates results to the limit parameter.
// Expected result: At most limit results returned.
func TestExecuteFuzzySearch_RespectsLimit(t *testing.T) {
	funcs := make([]any, 20)
	for i := range funcs {
		funcs[i] = map[string]any{"name": "handler", "path": "a.go"}
	}
	responses := []driver.Result{
		makeResult(map[string]any{"functions": funcs}),
		makeResult(map[string]any{"classs": []any{}}),
	}

	srv, _ := newTestServerWithResponses(t, responses)
	resp, err := srv.executeFuzzySearch(context.Background(), "myrepo", "handler", 3)
	if err != nil {
		t.Fatalf("executeFuzzySearch returned error: %v", err)
	}
	if resp == nil {
		t.Fatal("nil response")
	}
	if len(resp.Results) > 3 {
		t.Errorf("results = %d, want <= 3 (limit)", len(resp.Results))
	}
}

// TestSearchCodeNames_DispatchesFuzzyForGlob verifies that after renaming
// search_code to search_code_names, glob queries dispatch to fuzzy strategy.
// Expected result: handleSearchCode dispatches to executeFuzzySearch for glob queries.
func TestSearchCodeNames_DispatchesFuzzyForGlob(t *testing.T) {
	responses := []driver.Result{
		makeResult(map[string]any{
			"functions": []any{
				map[string]any{"name": "getHandler", "path": "a.go", "signature": "func getHandler()"},
			},
		}),
		makeResult(map[string]any{"classs": []any{}}),
	}

	srv, _ := newTestServerWithResponses(t, responses)
	resp, err := srv.handleSearchCode(context.Background(), "myrepo", "get*Handler", 10)
	if err != nil {
		t.Fatalf("handleSearchCode returned error: %v", err)
	}
	if resp == nil {
		t.Fatal("nil response")
	}
	// After refactoring, glob queries should use fuzzy strategy, not file strategy
	if resp.Strategy != "fuzzy" {
		t.Errorf("strategy = %q, want \"fuzzy\" for glob query", resp.Strategy)
	}
}
