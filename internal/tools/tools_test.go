package mcp

import (
	"context"
	"strings"
	"testing"

	"github.com/tab58/go-ormql/pkg/driver"
)

// === Task 5: find_function tool handler ===
//
// Exact name match via functions(where: {name, repository}) GraphQL query.
// Returns results with score=1.0, strategy="exact".

// TestHandleFindFunction_FoundResult verifies that find_function returns
// a matching function with score=1.0 and strategy="exact".
// Expected result: 1 result with correct fields.
func TestHandleFindFunction_FoundResult(t *testing.T) {
	// Set up response driver to return a function result
	svc, _ := newTestServiceWithResponses(t, []driver.Result{
		makeResult(map[string]any{
			"functions": []any{
				map[string]any{
					"name":         "getUserByID",
					"path":         "pkg/user.go",
					"source":       "func getUserByID(id string) *User { return nil }",
					"signature":    "func getUserByID(id string) *User",
					"language":     "go",
					"visibility":   "package",
					"startingLine": float64(10),
					"endingLine":   float64(12),
				},
			},
		}),
	})

	resp, err := svc.HandleFindFunction(context.Background(), "myrepo", "getUserByID")
	if err != nil {
		t.Fatalf("handleFindFunction returned error: %v", err)
	}
	if resp == nil {
		t.Fatal("handleFindFunction returned nil response")
	}
	if len(resp.Results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(resp.Results))
	}

	r := resp.Results[0]
	if r.Name != "getUserByID" {
		t.Errorf("result.Name = %q, want %q", r.Name, "getUserByID")
	}
	if r.Score != 1.0 {
		t.Errorf("result.Score = %v, want 1.0", r.Score)
	}
	if r.Strategy != "exact" {
		t.Errorf("result.Strategy = %q, want %q", r.Strategy, "exact")
	}
	if r.Type != "function" {
		t.Errorf("result.Type = %q, want %q", r.Type, "function")
	}
}

// TestHandleFindFunction_NotFound verifies that find_function returns
// an empty result set when no function matches.
// Expected result: 0 results, no error.
func TestHandleFindFunction_NotFound(t *testing.T) {
	// Response driver returns empty functions list
	svc, _ := newTestServiceWithResponses(t, []driver.Result{
		makeResult(map[string]any{
			"functions": []any{},
		}),
	})

	resp, err := svc.HandleFindFunction(context.Background(), "myrepo", "nonexistent")
	if err != nil {
		t.Fatalf("handleFindFunction returned error: %v", err)
	}
	if resp == nil {
		t.Fatal("handleFindFunction returned nil response")
	}
	if len(resp.Results) != 0 {
		t.Errorf("expected 0 results, got %d", len(resp.Results))
	}
}

// === Task 6: find_file tool handler ===
//
// Glob match on File paths within a repository.
// Returns files with score=0.9, strategy="file".
// For <=5 matches, enriches with symbol names.

// TestHandleFindFile_GlobMatch verifies that find_file matches file paths
// against a glob pattern and returns results with score=0.9.
// Expected result: matched files with score=0.9, strategy="file".
func TestHandleFindFile_GlobMatch(t *testing.T) {
	svc, _ := newTestServiceWithResponses(t, []driver.Result{
		makeResult(map[string]any{
			"files": []any{
				map[string]any{"name": "main.go", "path": "cmd/main.go", "language": "go", "lineCount": float64(50)},
				map[string]any{"name": "util.go", "path": "pkg/util.go", "language": "go", "lineCount": float64(30)},
				map[string]any{"name": "README.md", "path": "README.md", "language": "", "lineCount": float64(10)},
			},
		}),
	})

	resp, err := svc.HandleFindFile(context.Background(), "myrepo", "*.go")
	if err != nil {
		t.Fatalf("handleFindFile returned error: %v", err)
	}
	if resp == nil {
		t.Fatal("handleFindFile returned nil response")
	}
	// Should match main.go and util.go, not README.md
	if len(resp.Results) != 2 {
		t.Fatalf("expected 2 results, got %d", len(resp.Results))
	}
	for _, r := range resp.Results {
		if r.Score != 0.9 {
			t.Errorf("result.Score = %v, want 0.9", r.Score)
		}
		if r.Strategy != "file" {
			t.Errorf("result.Strategy = %q, want %q", r.Strategy, "file")
		}
		if r.Type != "file" {
			t.Errorf("result.Type = %q, want %q", r.Type, "file")
		}
	}
}

// TestHandleFindFile_SymbolEnrichment verifies that find_file enriches
// results with function/class names when <=5 files match.
// Expected result: Symbols field populated with defined function names.
func TestHandleFindFile_SymbolEnrichment(t *testing.T) {
	svc, _ := newTestServiceWithResponses(t, []driver.Result{
		// First call: list all files
		makeResult(map[string]any{
			"files": []any{
				map[string]any{"name": "main.go", "path": "cmd/main.go", "language": "go", "lineCount": float64(50)},
			},
		}),
		// Second call: functions for the matched file
		makeResult(map[string]any{
			"functions": []any{
				map[string]any{"name": "main"},
				map[string]any{"name": "init"},
			},
		}),
		// Third call: classes for the matched file
		makeResult(map[string]any{
			"classs": []any{},
		}),
	})

	resp, err := svc.HandleFindFile(context.Background(), "myrepo", "cmd/main.go")
	if err != nil {
		t.Fatalf("handleFindFile returned error: %v", err)
	}
	if resp == nil {
		t.Fatal("handleFindFile returned nil response")
	}
	if len(resp.Results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(resp.Results))
	}
	r := resp.Results[0]
	if len(r.Symbols) == 0 {
		t.Error("expected symbols to be enriched for <=5 matched files, got empty")
	}
}

// TestHandleFindFile_NoMatch verifies that find_file returns empty results
// when no files match the glob pattern.
// Expected result: 0 results, no error.
func TestHandleFindFile_NoMatch(t *testing.T) {
	svc, _ := newTestServiceWithResponses(t, []driver.Result{
		makeResult(map[string]any{
			"files": []any{
				map[string]any{"name": "main.go", "path": "cmd/main.go", "language": "go", "lineCount": float64(50)},
			},
		}),
	})

	resp, err := svc.HandleFindFile(context.Background(), "myrepo", "*.py")
	if err != nil {
		t.Fatalf("handleFindFile returned error: %v", err)
	}
	if resp == nil {
		t.Fatal("handleFindFile returned nil response")
	}
	if len(resp.Results) != 0 {
		t.Errorf("expected 0 results for non-matching glob, got %d", len(resp.Results))
	}
}

// === Task 8: search_code tool handler ===
//
// Classifies query, dispatches to file/exact strategy.
// Exact falls back to exact supplement on no results.

// TestHandleSearchCode_FuzzyStrategy verifies that search_code dispatches
// glob-like queries to the fuzzy strategy (Levenshtein).
// Expected result: strategy="fuzzy" in response.
func TestHandleSearchCode_FileStrategy(t *testing.T) {
	svc, _ := newTestServiceWithResponses(t, []driver.Result{
		// Functions response for fuzzy search
		makeResult(map[string]any{
			"functions": []any{
				map[string]any{"name": "go", "path": "cmd/main.go", "signature": "func go()", "language": "go"},
			},
		}),
		// Classes response for fuzzy search
		makeResult(map[string]any{
			"classs": []any{},
		}),
	})

	resp, err := svc.HandleSearchCode(context.Background(), "myrepo", "*.go", 10)
	if err != nil {
		t.Fatalf("handleSearchCode returned error: %v", err)
	}
	if resp == nil {
		t.Fatal("handleSearchCode returned nil response")
	}
	if resp.Strategy != "fuzzy" {
		t.Errorf("response.Strategy = %q, want %q for glob query", resp.Strategy, "fuzzy")
	}
}

// TestHandleSearchCode_ExactStrategy verifies that search_code dispatches
// identifier-like queries to the exact strategy.
// Expected result: strategy="exact" in response when results are found.
func TestHandleSearchCode_ExactStrategy(t *testing.T) {
	svc, _ := newTestServiceWithResponses(t, []driver.Result{
		makeResult(map[string]any{
			"functions": []any{
				map[string]any{
					"name": "getUserByID", "path": "pkg/user.go",
					"source": "func getUserByID() {}", "language": "go",
				},
			},
		}),
	})

	resp, err := svc.HandleSearchCode(context.Background(), "myrepo", "getUserByID", 10)
	if err != nil {
		t.Fatalf("handleSearchCode returned error: %v", err)
	}
	if resp == nil {
		t.Fatal("handleSearchCode returned nil response")
	}
	if resp.Strategy != "exact" {
		t.Errorf("response.Strategy = %q, want %q for identifier query", resp.Strategy, "exact")
	}
}

// TestHandleSearchCode_ExactFallback verifies that search_code
// falls back from exact to exact supplement when no exact matches are found.
// Expected result: strategy="exact" in response after fallback.
func TestHandleSearchCode_ExactFallback(t *testing.T) {
	svc, _ := newTestServiceWithResponses(t, []driver.Result{
		// First: exact match returns empty
		makeResult(map[string]any{"functions": []any{}}),
		// Then: exact supplement on "getUserByID" token
		makeResult(map[string]any{
			"functions": []any{
				map[string]any{
					"name": "getUser", "path": "pkg/user.go",
					"source": "func getUser() {}", "language": "go",
				},
			},
		}),
	})

	resp, err := svc.HandleSearchCode(context.Background(), "myrepo", "getUserByID", 10)
	if err != nil {
		t.Fatalf("handleSearchCode returned error: %v", err)
	}
	if resp == nil {
		t.Fatal("handleSearchCode returned nil response")
	}
	if resp.Strategy != "exact" {
		t.Errorf("response.Strategy = %q, want %q after exact fallback", resp.Strategy, "exact")
	}
}

// === Input validation and error handling ===
//
// All tools validate required params and return structured errors.
// Error messages must mention the missing field name specifically.

// TestValidation_FindFunction_MissingRepository verifies that find_function
// returns a validation error mentioning "repository" when it is empty.
// Expected result: error message contains "repository".
func TestValidation_FindFunction_MissingRepository(t *testing.T) {
	svc := newTestService(t)
	_, err := svc.HandleFindFunction(context.Background(), "", "myFunc")
	if err == nil {
		t.Fatal("expected error for missing repository, got nil")
	}
	if !strings.Contains(err.Error(), "repository") {
		t.Errorf("error should mention 'repository', got: %v", err)
	}
}

// TestValidation_FindFunction_MissingName verifies that find_function
// returns a validation error mentioning "name" when it is empty.
// Expected result: error message contains "name".
func TestValidation_FindFunction_MissingName(t *testing.T) {
	svc := newTestService(t)
	_, err := svc.HandleFindFunction(context.Background(), "myrepo", "")
	if err == nil {
		t.Fatal("expected error for missing name, got nil")
	}
	if !strings.Contains(err.Error(), "name") {
		t.Errorf("error should mention 'name', got: %v", err)
	}
}

// TestValidation_FindFile_MissingRepository verifies that find_file
// returns a validation error mentioning "repository" when it is empty.
// Expected result: error message contains "repository".
func TestValidation_FindFile_MissingRepository(t *testing.T) {
	svc := newTestService(t)
	_, err := svc.HandleFindFile(context.Background(), "", "*.go")
	if err == nil {
		t.Fatal("expected error for missing repository, got nil")
	}
	if !strings.Contains(err.Error(), "repository") {
		t.Errorf("error should mention 'repository', got: %v", err)
	}
}

// TestValidation_FindFile_MissingPattern verifies that find_file
// returns a validation error mentioning "pattern" when it is empty.
// Expected result: error message contains "pattern".
func TestValidation_FindFile_MissingPattern(t *testing.T) {
	svc := newTestService(t)
	_, err := svc.HandleFindFile(context.Background(), "myrepo", "")
	if err == nil {
		t.Fatal("expected error for missing pattern, got nil")
	}
	if !strings.Contains(err.Error(), "pattern") {
		t.Errorf("error should mention 'pattern', got: %v", err)
	}
}

// TestValidation_SearchCode_MissingRepository verifies that search_code
// returns a validation error mentioning "repository" when it is empty.
// Expected result: error message contains "repository".
func TestValidation_SearchCode_MissingRepository(t *testing.T) {
	svc := newTestService(t)
	_, err := svc.HandleSearchCode(context.Background(), "", "query", 10)
	if err == nil {
		t.Fatal("expected error for missing repository, got nil")
	}
	if !strings.Contains(err.Error(), "repository") {
		t.Errorf("error should mention 'repository', got: %v", err)
	}
}

// TestValidation_SearchCode_MissingQuery verifies that search_code
// returns a validation error mentioning "query" when it is empty.
// Expected result: error message contains "query".
func TestValidation_SearchCode_MissingQuery(t *testing.T) {
	svc := newTestService(t)
	_, err := svc.HandleSearchCode(context.Background(), "myrepo", "", 10)
	if err == nil {
		t.Fatal("expected error for missing query, got nil")
	}
	if !strings.Contains(err.Error(), "query") {
		t.Errorf("error should mention 'query', got: %v", err)
	}
}
