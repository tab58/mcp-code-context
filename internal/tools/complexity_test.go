package mcp

import (
	"context"
	"testing"

	"github.com/tab58/go-ormql/pkg/driver"
)

// --- Task 8: calculate_cyclomatic_complexity + find_most_complex_functions tools ---

// TestHandleCalculateCyclomaticComplexity_RequiresRepository verifies that
// calculate_cyclomatic_complexity returns an error when repository is empty.
// Expected result: Error with "repository is required".
func TestHandleCalculateCyclomaticComplexity_RequiresRepository(t *testing.T) {
	srv := newTestService(t)
	_, err := srv.HandleCalculateCyclomaticComplexity(context.Background(), "", "doWork")
	if err == nil {
		t.Fatal("handleCalculateCyclomaticComplexity with empty repo returned nil error")
	}
}

// TestHandleCalculateCyclomaticComplexity_RequiresName verifies that
// calculate_cyclomatic_complexity returns an error when name is empty.
// Expected result: Error with "name is required".
func TestHandleCalculateCyclomaticComplexity_RequiresName(t *testing.T) {
	srv := newTestService(t)
	_, err := srv.HandleCalculateCyclomaticComplexity(context.Background(), "myrepo", "")
	if err == nil {
		t.Fatal("handleCalculateCyclomaticComplexity with empty name returned nil error")
	}
}

// TestHandleCalculateCyclomaticComplexity_ReturnsComplexity verifies that
// calculate_cyclomatic_complexity queries the pre-computed complexity value.
// Expected result: ComplexityResponse with function's cyclomaticComplexity.
func TestHandleCalculateCyclomaticComplexity_ReturnsComplexity(t *testing.T) {
	responses := []driver.Result{
		makeResult(map[string]any{
			"functions": []any{
				map[string]any{
					"name":                 "doWork",
					"path":                 "internal/foo.go",
					"signature":            "func doWork()",
					"cyclomaticComplexity": float64(5),
					"startingLine":         float64(10),
					"endingLine":           float64(25),
				},
			},
		}),
	}

	srv, _ := newTestServiceWithResponses(t, responses)
	resp, err := srv.HandleCalculateCyclomaticComplexity(context.Background(), "myrepo", "doWork")
	if err != nil {
		t.Fatalf("handleCalculateCyclomaticComplexity returned error: %v", err)
	}
	if resp == nil {
		t.Fatal("nil response")
	}
	if len(resp.Results) == 0 {
		t.Fatal("expected at least 1 result")
	}
	if resp.Results[0].CyclomaticComplexity != 5 {
		t.Errorf("cyclomaticComplexity = %d, want 5", resp.Results[0].CyclomaticComplexity)
	}
}

// TestHandleFindMostComplexFunctions_RequiresRepository verifies that
// find_most_complex_functions returns an error when repository is empty.
// Expected result: Error with "repository is required".
func TestHandleFindMostComplexFunctions_RequiresRepository(t *testing.T) {
	srv := newTestService(t)
	_, err := srv.HandleFindMostComplexFunctions(context.Background(), "", 5, 10)
	if err == nil {
		t.Fatal("handleFindMostComplexFunctions with empty repo returned nil error")
	}
}

// TestHandleFindMostComplexFunctions_ReturnsResults verifies that
// find_most_complex_functions queries functions above min_complexity threshold
// and returns them sorted descending by complexity.
// Expected result: ComplexityResponse with filtered + sorted results.
func TestHandleFindMostComplexFunctions_ReturnsResults(t *testing.T) {
	responses := []driver.Result{
		makeResult(map[string]any{
			"functions": []any{
				map[string]any{
					"name":                 "complexFunc",
					"path":                 "a.go",
					"signature":            "func complexFunc()",
					"cyclomaticComplexity": float64(12),
					"startingLine":         float64(1),
					"endingLine":           float64(50),
				},
				map[string]any{
					"name":                 "simpleFunc",
					"path":                 "b.go",
					"signature":            "func simpleFunc()",
					"cyclomaticComplexity": float64(1),
					"startingLine":         float64(1),
					"endingLine":           float64(3),
				},
				map[string]any{
					"name":                 "mediumFunc",
					"path":                 "c.go",
					"signature":            "func mediumFunc()",
					"cyclomaticComplexity": float64(7),
					"startingLine":         float64(1),
					"endingLine":           float64(20),
				},
			},
		}),
	}

	srv, _ := newTestServiceWithResponses(t, responses)
	resp, err := srv.HandleFindMostComplexFunctions(context.Background(), "myrepo", 5, 10)
	if err != nil {
		t.Fatalf("handleFindMostComplexFunctions returned error: %v", err)
	}
	if resp == nil {
		t.Fatal("nil response")
	}
	// With min_complexity=5, should exclude simpleFunc (complexity=1)
	for _, r := range resp.Results {
		if r.CyclomaticComplexity < 5 {
			t.Errorf("result %q has complexity %d, want >= 5 (min_complexity)", r.Name, r.CyclomaticComplexity)
		}
	}
	if len(resp.Results) != 2 {
		t.Errorf("results = %d, want 2 (complexFunc + mediumFunc)", len(resp.Results))
	}
	// Results should be sorted descending by complexity
	if len(resp.Results) >= 2 && resp.Results[0].CyclomaticComplexity < resp.Results[1].CyclomaticComplexity {
		t.Error("results should be sorted descending by cyclomaticComplexity")
	}
}

// TestHandleFindMostComplexFunctions_RespectsLimit verifies that
// find_most_complex_functions truncates results to the limit parameter.
// Expected result: At most limit results returned.
func TestHandleFindMostComplexFunctions_RespectsLimit(t *testing.T) {
	funcs := make([]any, 10)
	for i := range funcs {
		funcs[i] = map[string]any{
			"name":                 "func" + string(rune('A'+i)),
			"path":                 "a.go",
			"cyclomaticComplexity": float64(10 + i),
		}
	}
	responses := []driver.Result{
		makeResult(map[string]any{"functions": funcs}),
	}

	srv, _ := newTestServiceWithResponses(t, responses)
	resp, err := srv.HandleFindMostComplexFunctions(context.Background(), "myrepo", 1, 3)
	if err != nil {
		t.Fatalf("handleFindMostComplexFunctions returned error: %v", err)
	}
	if resp == nil {
		t.Fatal("nil response")
	}
	if len(resp.Results) > 3 {
		t.Errorf("results = %d, want <= 3 (limit)", len(resp.Results))
	}
}
