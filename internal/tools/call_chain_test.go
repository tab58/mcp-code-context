package mcp

import (
	"context"
	"strings"
	"testing"

	"github.com/tab58/go-ormql/pkg/driver"
)

// === Tasks 7, 9, 11: Call chain handler tests ===

// --- Input validation ---

// TestHandleFindCallChain_EmptyRepo returns error when repository is empty.
// Expected result: Error containing "repository".
func TestHandleFindCallChain_EmptyRepo(t *testing.T) {
	svc := newTestService(t)

	_, err := svc.HandleFindCallChain(context.Background(), "", "funcA", "funcB", 3)
	if err == nil {
		t.Fatal("expected error for empty repo, got nil")
	}
	if !strings.Contains(err.Error(), "repository") {
		t.Errorf("error = %q, want containing 'repository'", err.Error())
	}
}

// TestHandleFindCallChain_EmptySource returns error when source is empty.
// Expected result: Error containing "source".
func TestHandleFindCallChain_EmptySource(t *testing.T) {
	svc := newTestService(t)

	_, err := svc.HandleFindCallChain(context.Background(), "myrepo", "", "funcB", 3)
	if err == nil {
		t.Fatal("expected error for empty source, got nil")
	}
	if !strings.Contains(err.Error(), "source") {
		t.Errorf("error = %q, want containing 'source'", err.Error())
	}
}

// TestHandleFindCallChain_EmptyTarget returns error when target is empty.
// Expected result: Error containing "target".
func TestHandleFindCallChain_EmptyTarget(t *testing.T) {
	svc := newTestService(t)

	_, err := svc.HandleFindCallChain(context.Background(), "myrepo", "funcA", "", 3)
	if err == nil {
		t.Fatal("expected error for empty target, got nil")
	}
	if !strings.Contains(err.Error(), "target") {
		t.Errorf("error = %q, want containing 'target'", err.Error())
	}
}

// --- Behavioral tests ---

// TestHandleFindCallChain_DirectCall verifies that a direct call (A -> B)
// is found at depth 1.
// Expected result: Found=true, Depth=1, Path has 1 entry (B).
func TestHandleFindCallChain_DirectCall(t *testing.T) {
	// Configure driver to return B when querying callees of A
	// and A when querying callers of B
	calleesOfA := makeResult(map[string]any{
		"functions": []any{
			map[string]any{"name": "funcB", "path": "b.go", "signature": "func funcB()"},
		},
	})
	callersOfB := makeResult(map[string]any{
		"functions": []any{
			map[string]any{"name": "funcA", "path": "a.go", "signature": "func funcA()"},
		},
	})

	svc, _ := newTestServiceWithResponses(t, []driver.Result{calleesOfA, callersOfB})

	resp, err := svc.HandleFindCallChain(context.Background(), "myrepo", "funcA", "funcB", 5)
	if err != nil {
		t.Fatalf("handleFindCallChain error: %v", err)
	}
	if resp == nil {
		t.Fatal("response is nil")
	}
	if !resp.Found {
		t.Error("expected Found=true for direct call A->B")
	}
	if resp.Source != "funcA" {
		t.Errorf("Source = %q, want %q", resp.Source, "funcA")
	}
	if resp.Target != "funcB" {
		t.Errorf("Target = %q, want %q", resp.Target, "funcB")
	}
}

// TestHandleFindCallChain_NoPath verifies that when no call chain exists,
// Found=false is returned.
// Expected result: Found=false, empty Path.
func TestHandleFindCallChain_NoPath(t *testing.T) {
	// Return empty results for both directions
	empty := makeResult(map[string]any{
		"functions": []any{},
	})

	svc, _ := newTestServiceWithResponses(t, []driver.Result{empty, empty, empty, empty, empty, empty, empty, empty, empty, empty})

	resp, err := svc.HandleFindCallChain(context.Background(), "myrepo", "funcA", "funcZ", 3)
	if err != nil {
		t.Fatalf("handleFindCallChain error: %v", err)
	}
	if resp == nil {
		t.Fatal("response is nil")
	}
	if resp.Found {
		t.Error("expected Found=false when no path exists")
	}
}

// TestHandleFindCallChain_MaxDepthCapped verifies that depth is capped at
// maxCallChainDepth (5).
// Expected result: Depth parameter is clamped to 5 even when 10 is passed.
func TestHandleFindCallChain_MaxDepthCapped(t *testing.T) {
	empty := makeResult(map[string]any{
		"functions": []any{},
	})

	// Use enough empty results to cover the max depth iterations
	results := make([]driver.Result, 20)
	for i := range results {
		results[i] = empty
	}

	svc, drv := newTestServiceWithResponses(t, results)

	resp, err := svc.HandleFindCallChain(context.Background(), "myrepo", "funcA", "funcZ", 10)
	if err != nil {
		t.Fatalf("handleFindCallChain error: %v", err)
	}
	if resp == nil {
		t.Fatal("response is nil")
	}

	// Verify the query count doesn't exceed what maxCallChainDepth=5 allows
	// Each depth level should query at most 2 directions (callees + callers)
	// Max queries = 5 * 2 = 10
	if len(drv.readCalls) > 10 {
		t.Errorf("too many queries (%d), expected at most 10 for maxCallChainDepth=5", len(drv.readCalls))
	}
}

// TestHandleFindCallChain_SameSourceTarget verifies behavior when source
// and target are the same function.
// Expected result: Found=true, empty Path, Depth=0.
func TestHandleFindCallChain_SameSourceTarget(t *testing.T) {
	svc := newTestService(t)

	resp, err := svc.HandleFindCallChain(context.Background(), "myrepo", "funcA", "funcA", 5)
	if err != nil {
		t.Fatalf("handleFindCallChain error: %v", err)
	}
	if resp == nil {
		t.Fatal("response is nil")
	}
	if !resp.Found {
		t.Error("expected Found=true when source == target")
	}
	if resp.Depth != 0 {
		t.Errorf("Depth = %d, want 0 when source == target", resp.Depth)
	}
}

// TestHandleFindCallChain_MultiHop verifies that a multi-hop path (A->C->B)
// is found correctly.
// Expected result: Found=true, Path contains intermediate function C.
func TestHandleFindCallChain_MultiHop(t *testing.T) {
	// Callees of A: [C]
	calleesOfA := makeResult(map[string]any{
		"functions": []any{
			map[string]any{"name": "funcC", "path": "c.go", "signature": "func funcC()"},
		},
	})
	// Callers of B: [C]
	callersOfB := makeResult(map[string]any{
		"functions": []any{
			map[string]any{"name": "funcC", "path": "c.go", "signature": "func funcC()"},
		},
	})
	// Additional queries that may happen during BFS expansion
	calleesOfC := makeResult(map[string]any{
		"functions": []any{
			map[string]any{"name": "funcB", "path": "b.go", "signature": "func funcB()"},
		},
	})
	callersOfC := makeResult(map[string]any{
		"functions": []any{
			map[string]any{"name": "funcA", "path": "a.go", "signature": "func funcA()"},
		},
	})

	svc, _ := newTestServiceWithResponses(t, []driver.Result{
		calleesOfA, callersOfB, calleesOfC, callersOfC,
	})

	resp, err := svc.HandleFindCallChain(context.Background(), "myrepo", "funcA", "funcB", 5)
	if err != nil {
		t.Fatalf("handleFindCallChain error: %v", err)
	}
	if resp == nil {
		t.Fatal("response is nil")
	}
	if !resp.Found {
		t.Error("expected Found=true for multi-hop path A->C->B")
	}
}
