package mcp

import (
	"context"
	"testing"

	"github.com/tab58/go-ormql/pkg/driver"
)

// --- Task 6: Content-based search_code tool ---

// TestHandleSearchCodeContent_RequiresRepository verifies that content search
// returns an error when repository is empty.
func TestHandleSearchCodeContent_RequiresRepository(t *testing.T) {
	srv := newTestService(t)
	_, err := srv.HandleSearchCodeContent(context.Background(), "", "fmt.Println", 10)
	if err == nil {
		t.Fatal("handleSearchCodeContent with empty repo returned nil error")
	}
}

// TestHandleSearchCodeContent_RequiresQuery verifies that content search
// returns an error when query is empty.
func TestHandleSearchCodeContent_RequiresQuery(t *testing.T) {
	srv := newTestService(t)
	_, err := srv.HandleSearchCodeContent(context.Background(), "myrepo", "", 10)
	if err == nil {
		t.Fatal("handleSearchCodeContent with empty query returned nil error")
	}
}

// TestHandleSearchCodeContent_FindsSubstringMatch verifies that content search
// finds a substring match in function source and returns it.
func TestHandleSearchCodeContent_FindsSubstringMatch(t *testing.T) {
	responses := []driver.Result{
		// Functions with source
		makeResult(map[string]any{
			"functions": []any{
				map[string]any{
					"name":         "DoWork",
					"path":         "internal/foo.go",
					"signature":    "func DoWork()",
					"source":       "func DoWork() {\n\tfmt.Println(\"hello\")\n}",
					"startingLine": float64(3),
					"endingLine":   float64(5),
					"language":     "go",
				},
				map[string]any{
					"name":         "Other",
					"path":         "internal/bar.go",
					"signature":    "func Other()",
					"source":       "func Other() {}",
					"startingLine": float64(3),
					"endingLine":   float64(3),
					"language":     "go",
				},
			},
		}),
		// Classes with source
		makeResult(map[string]any{
			"classs": []any{},
		}),
	}

	srv, _ := newTestServiceWithResponses(t, responses)
	resp, err := srv.HandleSearchCodeContent(context.Background(), "myrepo", "fmt.Println", 10)
	if err != nil {
		t.Fatalf("handleSearchCodeContent returned error: %v", err)
	}
	if resp == nil {
		t.Fatal("nil response")
	}
	if len(resp.Results) == 0 {
		t.Error("expected at least 1 content match result")
	}
	if resp.Strategy != "content" {
		t.Errorf("strategy = %q, want \"content\"", resp.Strategy)
	}
}

// TestHandleSearchCodeContent_ResolvesToFunction verifies that content search
// returns function-level results when query matches function source.
func TestHandleSearchCodeContent_ResolvesToFunction(t *testing.T) {
	responses := []driver.Result{
		// Functions with source containing match
		makeResult(map[string]any{
			"functions": []any{
				map[string]any{
					"name":         "DoWork",
					"path":         "a.go",
					"signature":    "func DoWork()",
					"source":       "func DoWork() {\n\tfmt.Println(\"hello\")\n}",
					"startingLine": float64(3),
					"endingLine":   float64(5),
				},
			},
		}),
		// Classes
		makeResult(map[string]any{"classs": []any{}}),
	}

	srv, _ := newTestServiceWithResponses(t, responses)
	resp, err := srv.HandleSearchCodeContent(context.Background(), "myrepo", "fmt.Println", 10)
	if err != nil {
		t.Fatalf("handleSearchCodeContent returned error: %v", err)
	}
	if resp == nil || len(resp.Results) == 0 {
		t.Fatal("expected at least 1 result")
	}

	found := false
	for _, r := range resp.Results {
		if r.Name == "DoWork" && r.Type == "function" {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected result resolved to function 'DoWork'")
	}
}

// TestHandleSearchCodeContent_NoResults verifies that content search
// returns empty results when no source matches the query.
func TestHandleSearchCodeContent_NoResults(t *testing.T) {
	responses := []driver.Result{
		makeResult(map[string]any{
			"functions": []any{
				map[string]any{
					"name":   "DoWork",
					"path":   "a.go",
					"source": "func DoWork() {}",
				},
			},
		}),
		makeResult(map[string]any{"classs": []any{}}),
	}

	srv, _ := newTestServiceWithResponses(t, responses)
	resp, err := srv.HandleSearchCodeContent(context.Background(), "myrepo", "nonexistent_pattern", 10)
	if err != nil {
		t.Fatalf("handleSearchCodeContent returned error: %v", err)
	}
	if resp == nil {
		t.Fatal("nil response")
	}
	if len(resp.Results) != 0 {
		t.Errorf("expected 0 results, got %d", len(resp.Results))
	}
}

// TestHandleSearchCodeContent_RespectsLimit verifies that content search
// truncates results to the limit parameter.
func TestHandleSearchCodeContent_RespectsLimit(t *testing.T) {
	funcs := make([]any, 10)
	for i := range funcs {
		funcs[i] = map[string]any{
			"name":   "Func" + string(rune('A'+i)),
			"path":   "file" + string(rune('A'+i)) + ".go",
			"source": "func X() { fmt.Println(\"match\") }",
		}
	}
	responses := []driver.Result{
		makeResult(map[string]any{"functions": funcs}),
		makeResult(map[string]any{"classs": []any{}}),
	}

	srv, _ := newTestServiceWithResponses(t, responses)
	resp, err := srv.HandleSearchCodeContent(context.Background(), "myrepo", "fmt.Println", 3)
	if err != nil {
		t.Fatalf("handleSearchCodeContent returned error: %v", err)
	}
	if resp == nil {
		t.Fatal("nil response")
	}
	if len(resp.Results) > 3 {
		t.Errorf("results = %d, want <= 3 (limit)", len(resp.Results))
	}
}
