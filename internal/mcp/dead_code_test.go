package mcp

import (
	"context"
	"testing"

	"github.com/tab58/go-ormql/pkg/driver"
)

// --- Task 7: find_dead_code tool ---

// TestHandleFindDeadCode_RequiresRepository verifies that find_dead_code
// returns an error when repository is empty.
// Expected result: Error with "repository is required".
func TestHandleFindDeadCode_RequiresRepository(t *testing.T) {
	srv := newTestServer(t)
	_, err := srv.handleFindDeadCode(context.Background(), "", false, "", 50)
	if err == nil {
		t.Fatal("handleFindDeadCode with empty repo returned nil error")
	}
}

// TestHandleFindDeadCode_ReturnsDeadFunctions verifies that find_dead_code
// queries functions with _NONE relationship filters and returns results.
// Expected result: DeadCodeResponse with results from canned data.
func TestHandleFindDeadCode_ReturnsDeadFunctions(t *testing.T) {
	responses := []driver.Result{
		// Dead functions query response
		makeResult(map[string]any{
			"functions": []any{
				map[string]any{
					"name":         "unusedFunc",
					"path":         "internal/foo.go",
					"signature":    "func unusedFunc()",
					"startingLine": float64(10),
					"endingLine":   float64(15),
					"language":     "go",
				},
			},
		}),
		// Dead classes query response
		makeResult(map[string]any{
			"classs": []any{},
		}),
		// Dead modules query response
		makeResult(map[string]any{
			"modules": []any{},
		}),
	}

	srv, _ := newTestServerWithResponses(t, responses)
	resp, err := srv.handleFindDeadCode(context.Background(), "myrepo", false, "", 50)
	if err != nil {
		t.Fatalf("handleFindDeadCode returned error: %v", err)
	}
	if resp == nil {
		t.Fatal("handleFindDeadCode returned nil response")
	}
	if resp.Total == 0 {
		t.Error("handleFindDeadCode returned 0 results, expected at least 1 dead function")
	}
}

// TestHandleFindDeadCode_ExcludesGoMainInit verifies that Go main() and init()
// functions are auto-excluded from dead code results.
// Expected result: main and init are filtered out.
func TestHandleFindDeadCode_ExcludesGoMainInit(t *testing.T) {
	responses := []driver.Result{
		// Dead functions including main() and init()
		makeResult(map[string]any{
			"functions": []any{
				map[string]any{"name": "main", "path": "cmd/main.go", "language": "go"},
				map[string]any{"name": "init", "path": "pkg/init.go", "language": "go"},
				map[string]any{"name": "unused", "path": "pkg/foo.go", "language": "go"},
			},
		}),
		makeResult(map[string]any{"classs": []any{}}),
		makeResult(map[string]any{"modules": []any{}}),
	}

	srv, _ := newTestServerWithResponses(t, responses)
	resp, err := srv.handleFindDeadCode(context.Background(), "myrepo", false, "", 50)
	if err != nil {
		t.Fatalf("handleFindDeadCode returned error: %v", err)
	}
	if resp == nil {
		t.Fatal("nil response")
	}

	for _, r := range resp.Results {
		if r.Name == "main" || r.Name == "init" {
			t.Errorf("Go %s() should be auto-excluded from dead code results", r.Name)
		}
	}
	if resp.Total != 1 {
		t.Errorf("total = %d, want 1 (only 'unused')", resp.Total)
	}
}

// TestHandleFindDeadCode_ExcludeDecorated verifies that exclude_decorated=true
// filters out symbols with non-empty decorators.
// Expected result: Decorated symbols are excluded.
func TestHandleFindDeadCode_ExcludeDecorated(t *testing.T) {
	responses := []driver.Result{
		makeResult(map[string]any{
			"functions": []any{
				map[string]any{"name": "decorated", "path": "a.ts", "language": "typescript",
					"decorators": []any{"@Injectable"}},
				map[string]any{"name": "plain", "path": "b.ts", "language": "typescript"},
			},
		}),
		makeResult(map[string]any{"classs": []any{}}),
		makeResult(map[string]any{"modules": []any{}}),
	}

	srv, _ := newTestServerWithResponses(t, responses)
	resp, err := srv.handleFindDeadCode(context.Background(), "myrepo", true, "", 50)
	if err != nil {
		t.Fatalf("handleFindDeadCode returned error: %v", err)
	}
	if resp == nil {
		t.Fatal("nil response")
	}

	for _, r := range resp.Results {
		if r.Name == "decorated" {
			t.Error("decorated function should be excluded when exclude_decorated=true")
		}
	}
}

// TestHandleFindDeadCode_ExcludePatterns verifies that exclude_patterns filters
// out symbols matching the glob pattern on name.
// Expected result: Symbols matching pattern are excluded.
func TestHandleFindDeadCode_ExcludePatterns(t *testing.T) {
	responses := []driver.Result{
		makeResult(map[string]any{
			"functions": []any{
				map[string]any{"name": "TestSomething", "path": "a.go", "language": "go"},
				map[string]any{"name": "doWork", "path": "b.go", "language": "go"},
			},
		}),
		makeResult(map[string]any{"classs": []any{}}),
		makeResult(map[string]any{"modules": []any{}}),
	}

	srv, _ := newTestServerWithResponses(t, responses)
	resp, err := srv.handleFindDeadCode(context.Background(), "myrepo", false, "Test*", 50)
	if err != nil {
		t.Fatalf("handleFindDeadCode returned error: %v", err)
	}
	if resp == nil {
		t.Fatal("nil response")
	}

	for _, r := range resp.Results {
		if r.Name == "TestSomething" {
			t.Error("TestSomething should be excluded by pattern 'Test*'")
		}
	}
}

// TestHandleFindDeadCode_RespectsLimit verifies that find_dead_code respects
// the limit parameter.
// Expected result: Results truncated to limit.
func TestHandleFindDeadCode_RespectsLimit(t *testing.T) {
	funcs := make([]any, 10)
	for i := range funcs {
		funcs[i] = map[string]any{"name": "func" + string(rune('A'+i)), "path": "a.go", "language": "go"}
	}
	responses := []driver.Result{
		makeResult(map[string]any{"functions": funcs}),
		makeResult(map[string]any{"classs": []any{}}),
		makeResult(map[string]any{"modules": []any{}}),
	}

	srv, _ := newTestServerWithResponses(t, responses)
	resp, err := srv.handleFindDeadCode(context.Background(), "myrepo", false, "", 3)
	if err != nil {
		t.Fatalf("handleFindDeadCode returned error: %v", err)
	}
	if resp == nil {
		t.Fatal("nil response")
	}
	if len(resp.Results) > 3 {
		t.Errorf("results = %d, want <= 3 (limit)", len(resp.Results))
	}
}
