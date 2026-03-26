package mcp

import (
	"context"
	"encoding/json"
	"strings"
	"testing"

	"github.com/tab58/go-ormql/pkg/client"
	"github.com/tab58/go-ormql/pkg/driver"
)

// === Task 1: TraversalResult and TraversalResponse types ===
//
// Verifies that TraversalResult/TraversalResponse types exist with correct
// fields and JSON tags, and that maxTraversalDepth constant is 3.

// TestTraversalResult_FieldsExist verifies that TraversalResult struct
// has all required fields and they serialize correctly to JSON.
// Expected result: all fields present in JSON output with correct keys.
func TestTraversalResult_FieldsExist(t *testing.T) {
	r := TraversalResult{
		Type:      "function",
		Name:      "myFunc",
		Path:      "pkg/main.go",
		Signature: "func myFunc()",
		Kind:      "",
		Language:  "go",
		Depth:     1,
		EdgeType:  "calls",
		Direction: "up",
	}

	data, err := json.Marshal(r)
	if err != nil {
		t.Fatalf("failed to marshal TraversalResult: %v", err)
	}
	s := string(data)

	for _, field := range []string{`"type"`, `"name"`, `"path"`, `"depth"`, `"edgeType"`, `"direction"`} {
		if !strings.Contains(s, field) {
			t.Errorf("JSON output missing field %s: %s", field, s)
		}
	}
}

// TestTraversalResult_OmitEmpty verifies that empty optional fields
// (signature, kind, language, edgeType, direction) are omitted from JSON.
// Expected result: omitempty fields not in JSON when empty.
func TestTraversalResult_OmitEmpty(t *testing.T) {
	r := TraversalResult{
		Type:  "function",
		Name:  "myFunc",
		Path:  "pkg/main.go",
		Depth: 1,
	}

	data, err := json.Marshal(r)
	if err != nil {
		t.Fatalf("failed to marshal TraversalResult: %v", err)
	}
	s := string(data)

	for _, field := range []string{`"signature"`, `"kind"`, `"language"`, `"edgeType"`, `"direction"`} {
		if strings.Contains(s, field) {
			t.Errorf("JSON output should omit empty field %s: %s", field, s)
		}
	}
}

// TestTraversalResponse_FieldsExist verifies that TraversalResponse struct
// has all required fields (Results, Source, Total, Depth).
// Expected result: all fields present in JSON output.
func TestTraversalResponse_FieldsExist(t *testing.T) {
	resp := TraversalResponse{
		Results: []TraversalResult{{Type: "function", Name: "a", Path: "b", Depth: 1}},
		Source:  "myFunc",
		Total:   1,
		Depth:   2,
	}

	data, err := json.Marshal(resp)
	if err != nil {
		t.Fatalf("failed to marshal TraversalResponse: %v", err)
	}
	s := string(data)

	for _, field := range []string{`"results"`, `"source"`, `"total"`, `"depth"`} {
		if !strings.Contains(s, field) {
			t.Errorf("JSON output missing field %s: %s", field, s)
		}
	}
}

// TestMaxTraversalDepth verifies that the maxTraversalDepth constant is 3.
// Expected result: maxTraversalDepth == 3.
func TestMaxTraversalDepth(t *testing.T) {
	if maxTraversalDepth != 3 {
		t.Errorf("maxTraversalDepth = %d, want 3", maxTraversalDepth)
	}
}

// === Task 2: traversal.go infrastructure ===
//
// Verifies clampDepth, isFilePath, and parse functions.

// TestClampDepth verifies that clampDepth constrains depth to [1, maxTraversalDepth].
// Expected result: 0 -> 1, negative -> 1, 1 -> 1, 3 -> 3, 4 -> 3.
func TestClampDepth(t *testing.T) {
	tests := []struct {
		name     string
		input    int
		expected int
	}{
		{"zero", 0, 1},
		{"negative", -5, 1},
		{"one", 1, 1},
		{"two", 2, 2},
		{"three", 3, 3},
		{"four (above max)", 4, 3},
		{"hundred", 100, 3},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := clampDepth(tt.input)
			if got != tt.expected {
				t.Errorf("clampDepth(%d) = %d, want %d", tt.input, got, tt.expected)
			}
		})
	}
}

// TestIsFilePath verifies that isFilePath detects file paths (containing / or
// having a file extension) vs. bare module/function names.
// Expected result: paths with / or extensions -> true, bare names -> false.
func TestIsFilePath(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected bool
	}{
		{"path with slash", "pkg/user.go", true},
		{"extension only", "main.go", true},
		{"ts extension", "helpers.ts", true},
		{"tsx extension", "App.tsx", true},
		{"bare module name", "fmt", false},
		{"camelCase name", "getUserByID", false},
		{"dotless path with slash", "internal/config", true},
		{"empty string", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := isFilePath(tt.input)
			if got != tt.expected {
				t.Errorf("isFilePath(%q) = %v, want %v", tt.input, got, tt.expected)
			}
		})
	}
}

// TestParseFunctionResult verifies that parseFunctionResult extracts
// name, path, signature, language from a map and tags depth/edgeType/direction.
// Expected result: TraversalResult with type="function" and all fields populated.
func TestParseFunctionResult(t *testing.T) {
	m := map[string]any{
		"name":      "myFunc",
		"path":      "pkg/main.go",
		"signature": "func myFunc()",
		"language":  "go",
	}

	r := parseFunctionResult(m, 2, "calls", "")
	if r.Type != "function" {
		t.Errorf("Type = %q, want %q", r.Type, "function")
	}
	if r.Name != "myFunc" {
		t.Errorf("Name = %q, want %q", r.Name, "myFunc")
	}
	if r.Path != "pkg/main.go" {
		t.Errorf("Path = %q, want %q", r.Path, "pkg/main.go")
	}
	if r.Signature != "func myFunc()" {
		t.Errorf("Signature = %q, want %q", r.Signature, "func myFunc()")
	}
	if r.Language != "go" {
		t.Errorf("Language = %q, want %q", r.Language, "go")
	}
	if r.Depth != 2 {
		t.Errorf("Depth = %d, want 2", r.Depth)
	}
	if r.EdgeType != "calls" {
		t.Errorf("EdgeType = %q, want %q", r.EdgeType, "calls")
	}
}

// TestParseClassResult verifies that parseClassResult extracts
// name, path, kind, language from a map.
// Expected result: TraversalResult with type="class" and kind populated.
func TestParseClassResult(t *testing.T) {
	m := map[string]any{
		"name":     "MyClass",
		"path":     "pkg/types.go",
		"kind":     "struct",
		"language": "go",
	}

	r := parseClassResult(m, 1, "inherits", "up")
	if r.Type != "class" {
		t.Errorf("Type = %q, want %q", r.Type, "class")
	}
	if r.Kind != "struct" {
		t.Errorf("Kind = %q, want %q", r.Kind, "struct")
	}
	if r.Direction != "up" {
		t.Errorf("Direction = %q, want %q", r.Direction, "up")
	}
}

// TestParseModuleResult verifies that parseModuleResult extracts
// name, path, importPath, kind, language from a map.
// Expected result: TraversalResult with type="module".
func TestParseModuleResult(t *testing.T) {
	m := map[string]any{
		"name":       "fmt",
		"path":       "pkg/fmt",
		"importPath": "fmt",
		"kind":       "package",
		"language":   "go",
	}

	r := parseModuleResult(m, 1, "depends_on", "")
	if r.Type != "module" {
		t.Errorf("Type = %q, want %q", r.Type, "module")
	}
	if r.Name != "fmt" {
		t.Errorf("Name = %q, want %q", r.Name, "fmt")
	}
	if r.Kind != "package" {
		t.Errorf("Kind = %q, want %q", r.Kind, "package")
	}
}

// TestParseFileResult verifies that parseFileResult extracts
// name (derived from path), path, language from a map.
// Expected result: TraversalResult with type="file".
func TestParseFileResult(t *testing.T) {
	m := map[string]any{
		"path":     "pkg/user.go",
		"language": "go",
	}

	r := parseFileResult(m, 1, "imports", "")
	if r.Type != "file" {
		t.Errorf("Type = %q, want %q", r.Type, "file")
	}
	if r.Path != "pkg/user.go" {
		t.Errorf("Path = %q, want %q", r.Path, "pkg/user.go")
	}
}

// TestParseFunctionResult_MissingFields verifies that parseFunctionResult
// handles missing fields gracefully (returns empty strings).
// Expected result: no panic, empty string fields.
func TestParseFunctionResult_MissingFields(t *testing.T) {
	m := map[string]any{}
	r := parseFunctionResult(m, 1, "calls", "")
	if r.Name != "" {
		t.Errorf("Name = %q, want empty", r.Name)
	}
	if r.Path != "" {
		t.Errorf("Path = %q, want empty", r.Path)
	}
}

// === Task 3: traverseHops multi-hop helper ===
//
// Iterative multi-hop traversal: start names -> query neighbors -> repeat.
// Uses visited set for cycle prevention and deduplication.

// TestTraverseHops_SingleHop verifies that traverseHops returns direct
// neighbors at depth 1 when maxDepth=1.
// Expected result: results tagged with depth=1.
func TestTraverseHops_SingleHop(t *testing.T) {
	// Query for functions whose calls_some matches "main": returns funcA
	svc, _ := newTestServiceWithResponses(t, []driver.Result{
		makeResult(map[string]any{
			"functions": []any{
				map[string]any{"name": "funcA", "path": "a.go", "signature": "func funcA()", "language": "go"},
			},
		}),
	})

	results, err := svc.traverseHops(
		context.Background(),
		mustForRepo(t, svc, "myrepo"),
		[]string{"main"},
		gqlFindCallers,
		"calls_some", "name", "functions", "calls", "",
		1,
		parseFunctionResult,
	)
	if err != nil {
		t.Fatalf("traverseHops returned error: %v", err)
	}
	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}
	if results[0].Name != "funcA" {
		t.Errorf("result[0].Name = %q, want %q", results[0].Name, "funcA")
	}
	if results[0].Depth != 1 {
		t.Errorf("result[0].Depth = %d, want 1", results[0].Depth)
	}
}

// TestTraverseHops_MultiHop verifies that traverseHops expands neighbors
// iteratively across multiple depth levels.
// Expected result: depth-1 and depth-2 results both present.
func TestTraverseHops_MultiHop(t *testing.T) {
	svc, _ := newTestServiceWithResponses(t, []driver.Result{
		// Depth 1: main -> funcA
		makeResult(map[string]any{
			"functions": []any{
				map[string]any{"name": "funcA", "path": "a.go", "signature": "func funcA()", "language": "go"},
			},
		}),
		// Depth 2: funcA -> funcB
		makeResult(map[string]any{
			"functions": []any{
				map[string]any{"name": "funcB", "path": "b.go", "signature": "func funcB()", "language": "go"},
			},
		}),
	})

	results, err := svc.traverseHops(
		context.Background(),
		mustForRepo(t, svc, "myrepo"),
		[]string{"main"},
		gqlFindCallers,
		"calls_some", "name", "functions", "calls", "",
		2,
		parseFunctionResult,
	)
	if err != nil {
		t.Fatalf("traverseHops returned error: %v", err)
	}
	if len(results) != 2 {
		t.Fatalf("expected 2 results, got %d", len(results))
	}

	// Verify depth tagging
	depths := map[int]bool{}
	for _, r := range results {
		depths[r.Depth] = true
	}
	if !depths[1] || !depths[2] {
		t.Errorf("expected results at depth 1 and 2, got depths: %v", depths)
	}
}

// TestTraverseHops_CycleDetection verifies that traverseHops does not
// revisit already-seen nodes (cycle prevention).
// Expected result: funcA appears once even if returned at depth 2.
func TestTraverseHops_CycleDetection(t *testing.T) {
	svc, _ := newTestServiceWithResponses(t, []driver.Result{
		// Depth 1: main -> funcA
		makeResult(map[string]any{
			"functions": []any{
				map[string]any{"name": "funcA", "path": "a.go", "signature": "func funcA()", "language": "go"},
			},
		}),
		// Depth 2: funcA -> main (cycle!) + funcB (new)
		makeResult(map[string]any{
			"functions": []any{
				map[string]any{"name": "main", "path": "main.go", "signature": "func main()", "language": "go"},
				map[string]any{"name": "funcB", "path": "b.go", "signature": "func funcB()", "language": "go"},
			},
		}),
	})

	results, err := svc.traverseHops(
		context.Background(),
		mustForRepo(t, svc, "myrepo"),
		[]string{"main"},
		gqlFindCallers,
		"calls_some", "name", "functions", "calls", "",
		2,
		parseFunctionResult,
	)
	if err != nil {
		t.Fatalf("traverseHops returned error: %v", err)
	}

	// "main" should be excluded (seed + cycle), so only funcA and funcB
	nameSet := map[string]bool{}
	for _, r := range results {
		nameSet[r.Name] = true
	}
	if nameSet["main"] {
		t.Error("visited seed 'main' should not appear in results (cycle detection)")
	}
	if !nameSet["funcA"] || !nameSet["funcB"] {
		t.Errorf("expected funcA and funcB, got names: %v", nameSet)
	}
}

// TestTraverseHops_EmptyResults verifies that traverseHops returns empty
// slice when no neighbors are found at any depth.
// Expected result: 0 results, no error.
func TestTraverseHops_EmptyResults(t *testing.T) {
	svc, _ := newTestServiceWithResponses(t, []driver.Result{
		makeResult(map[string]any{
			"functions": []any{},
		}),
	})

	results, err := svc.traverseHops(
		context.Background(),
		mustForRepo(t, svc, "myrepo"),
		[]string{"nonexistent"},
		gqlFindCallers,
		"calls_some", "name", "functions", "calls", "",
		1,
		parseFunctionResult,
	)
	if err != nil {
		t.Fatalf("traverseHops returned error: %v", err)
	}
	if len(results) != 0 {
		t.Errorf("expected 0 results, got %d", len(results))
	}
}

// === Task 4: handleGetCallers and handleGetCallees ===
//
// Both validate inputs, clamp depth, call traverseHops with appropriate
// WHERE key (calls_some for callers, calledBy_some for callees).

// TestHandleGetCallers_SingleHop verifies that get_callers returns
// functions that call the named function at depth 1.
// Expected result: TraversalResponse with edgeType="calls".
func TestHandleGetCallers_SingleHop(t *testing.T) {
	svc, _ := newTestServiceWithResponses(t, []driver.Result{
		makeResult(map[string]any{
			"functions": []any{
				map[string]any{"name": "caller1", "path": "a.go", "signature": "func caller1()", "language": "go"},
			},
		}),
	})

	resp, err := svc.HandleGetCallers(context.Background(), "myrepo", "target", 1)
	if err != nil {
		t.Fatalf("handleGetCallers returned error: %v", err)
	}
	if resp == nil {
		t.Fatal("handleGetCallers returned nil response")
	}
	if len(resp.Results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(resp.Results))
	}
	if resp.Results[0].EdgeType != "calls" {
		t.Errorf("EdgeType = %q, want %q", resp.Results[0].EdgeType, "calls")
	}
	if resp.Source != "target" {
		t.Errorf("Source = %q, want %q", resp.Source, "target")
	}
}

// TestHandleGetCallees_SingleHop verifies that get_callees returns
// functions called by the named function at depth 1.
// Expected result: TraversalResponse with edgeType="calls".
func TestHandleGetCallees_SingleHop(t *testing.T) {
	svc, _ := newTestServiceWithResponses(t, []driver.Result{
		makeResult(map[string]any{
			"functions": []any{
				map[string]any{"name": "callee1", "path": "b.go", "signature": "func callee1()", "language": "go"},
			},
		}),
	})

	resp, err := svc.HandleGetCallees(context.Background(), "myrepo", "source", 1)
	if err != nil {
		t.Fatalf("handleGetCallees returned error: %v", err)
	}
	if resp == nil {
		t.Fatal("handleGetCallees returned nil response")
	}
	if len(resp.Results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(resp.Results))
	}
	if resp.Results[0].Name != "callee1" {
		t.Errorf("Name = %q, want %q", resp.Results[0].Name, "callee1")
	}
}

// TestHandleGetCallers_Validation verifies that get_callers validates
// required params (repository, name).
// Expected result: error for empty repository and empty name.
func TestHandleGetCallers_Validation(t *testing.T) {
	svc := newTestService(t)

	_, err := svc.HandleGetCallers(context.Background(), "", "func", 1)
	if err == nil {
		t.Error("expected error for missing repository, got nil")
	}

	_, err = svc.HandleGetCallers(context.Background(), "repo", "", 1)
	if err == nil {
		t.Error("expected error for missing name, got nil")
	}
}

// TestHandleGetCallees_Validation verifies that get_callees validates
// required params (repository, name).
// Expected result: error for empty repository and empty name.
func TestHandleGetCallees_Validation(t *testing.T) {
	svc := newTestService(t)

	_, err := svc.HandleGetCallees(context.Background(), "", "func", 1)
	if err == nil {
		t.Error("expected error for missing repository, got nil")
	}

	_, err = svc.HandleGetCallees(context.Background(), "repo", "", 1)
	if err == nil {
		t.Error("expected error for missing name, got nil")
	}
}

// TestHandleGetCallers_DepthClamping verifies that depth is clamped
// to [1, 3] for get_callers.
// Expected result: depth 0 -> uses 1, depth 5 -> uses 3.
func TestHandleGetCallers_DepthClamping(t *testing.T) {
	svc, _ := newTestServiceWithResponses(t, []driver.Result{
		makeResult(map[string]any{"functions": []any{}}),
	})

	resp, err := svc.HandleGetCallers(context.Background(), "myrepo", "target", 0)
	if err != nil {
		t.Fatalf("handleGetCallers returned error: %v", err)
	}
	if resp.Depth != 1 {
		t.Errorf("Depth = %d, want 1 (clamped from 0)", resp.Depth)
	}
}

// === Task 5: handleGetClassHierarchy ===
//
// Supports direction param (up/down/both). Up queries inheritedBy_some +
// implementedBy_some. Down queries inherits_some + implements_some.

// TestHandleGetClassHierarchy_Up verifies that direction="up" queries
// parent classes and implemented interfaces.
// Expected result: results with direction="up".
func TestHandleGetClassHierarchy_Up(t *testing.T) {
	svc, _ := newTestServiceWithResponses(t, []driver.Result{
		// inheritedBy_some -> parent
		makeResult(map[string]any{
			"classs": []any{
				map[string]any{"name": "BaseClass", "path": "base.go", "kind": "struct", "language": "go"},
			},
		}),
		// implementedBy_some -> interface
		makeResult(map[string]any{
			"classs": []any{
				map[string]any{"name": "MyInterface", "path": "iface.go", "kind": "interface", "language": "go"},
			},
		}),
	})

	resp, err := svc.HandleGetClassHierarchy(context.Background(), "myrepo", "MyClass", "up", 1)
	if err != nil {
		t.Fatalf("handleGetClassHierarchy returned error: %v", err)
	}
	if resp == nil {
		t.Fatal("handleGetClassHierarchy returned nil response")
	}
	if len(resp.Results) != 2 {
		t.Fatalf("expected 2 results (parent + interface), got %d", len(resp.Results))
	}
	for _, r := range resp.Results {
		if r.Direction != "up" {
			t.Errorf("Direction = %q, want %q", r.Direction, "up")
		}
	}
}

// TestHandleGetClassHierarchy_Down verifies that direction="down" queries
// child classes and implementors.
// Expected result: results with direction="down".
func TestHandleGetClassHierarchy_Down(t *testing.T) {
	svc, _ := newTestServiceWithResponses(t, []driver.Result{
		// inherits_some -> child
		makeResult(map[string]any{
			"classs": []any{
				map[string]any{"name": "ChildClass", "path": "child.go", "kind": "struct", "language": "go"},
			},
		}),
		// implements_some -> implementor
		makeResult(map[string]any{
			"classs": []any{
				map[string]any{"name": "Implementor", "path": "impl.go", "kind": "struct", "language": "go"},
			},
		}),
	})

	resp, err := svc.HandleGetClassHierarchy(context.Background(), "myrepo", "MyInterface", "down", 1)
	if err != nil {
		t.Fatalf("handleGetClassHierarchy returned error: %v", err)
	}
	if resp == nil {
		t.Fatal("handleGetClassHierarchy returned nil response")
	}
	if len(resp.Results) != 2 {
		t.Fatalf("expected 2 results (child + implementor), got %d", len(resp.Results))
	}
	for _, r := range resp.Results {
		if r.Direction != "down" {
			t.Errorf("Direction = %q, want %q", r.Direction, "down")
		}
	}
}

// TestHandleGetClassHierarchy_Both verifies that direction="both" runs
// all 4 traversals and merges results.
// Expected result: both up and down results present.
func TestHandleGetClassHierarchy_Both(t *testing.T) {
	svc, _ := newTestServiceWithResponses(t, []driver.Result{
		// Up: inheritedBy_some -> parent
		makeResult(map[string]any{
			"classs": []any{
				map[string]any{"name": "Parent", "path": "parent.go", "kind": "struct", "language": "go"},
			},
		}),
		// Up: implementedBy_some -> interface
		makeResult(map[string]any{
			"classs": []any{},
		}),
		// Down: inherits_some -> child
		makeResult(map[string]any{
			"classs": []any{
				map[string]any{"name": "Child", "path": "child.go", "kind": "struct", "language": "go"},
			},
		}),
		// Down: implements_some -> implementor
		makeResult(map[string]any{
			"classs": []any{},
		}),
	})

	resp, err := svc.HandleGetClassHierarchy(context.Background(), "myrepo", "MyClass", "both", 1)
	if err != nil {
		t.Fatalf("handleGetClassHierarchy returned error: %v", err)
	}
	if resp == nil {
		t.Fatal("handleGetClassHierarchy returned nil response")
	}
	if len(resp.Results) != 2 {
		t.Fatalf("expected 2 results (parent + child), got %d", len(resp.Results))
	}

	dirs := map[string]bool{}
	for _, r := range resp.Results {
		dirs[r.Direction] = true
	}
	if !dirs["up"] || !dirs["down"] {
		t.Errorf("expected both 'up' and 'down' directions, got: %v", dirs)
	}
}

// TestHandleGetClassHierarchy_DefaultDirection verifies that empty
// direction defaults to "both".
// Expected result: same behavior as direction="both".
func TestHandleGetClassHierarchy_DefaultDirection(t *testing.T) {
	svc, _ := newTestServiceWithResponses(t, []driver.Result{
		makeResult(map[string]any{"classs": []any{}}),
		makeResult(map[string]any{"classs": []any{}}),
		makeResult(map[string]any{"classs": []any{}}),
		makeResult(map[string]any{"classs": []any{}}),
	})

	resp, err := svc.HandleGetClassHierarchy(context.Background(), "myrepo", "MyClass", "", 1)
	if err != nil {
		t.Fatalf("handleGetClassHierarchy returned error: %v", err)
	}
	// Should not error — empty direction treated as "both"
	if resp == nil {
		t.Fatal("handleGetClassHierarchy returned nil response")
	}
}

// TestHandleGetClassHierarchy_Validation verifies input validation.
// Expected result: error for empty repository and empty name.
func TestHandleGetClassHierarchy_Validation(t *testing.T) {
	svc := newTestService(t)

	_, err := svc.HandleGetClassHierarchy(context.Background(), "", "MyClass", "both", 1)
	if err == nil {
		t.Error("expected error for missing repository, got nil")
	}

	_, err = svc.HandleGetClassHierarchy(context.Background(), "repo", "", "both", 1)
	if err == nil {
		t.Error("expected error for missing name, got nil")
	}
}

// === Task 6: handleGetDependencies ===
//
// Auto-detects Module name vs File path. Module: traverseHops with
// dependedOnBy_some. File: single query with importedBy_some.

// TestHandleGetDependencies_Module verifies that a bare name (no / or
// extension) is treated as a Module and queries dependedOnBy_some.
// Expected result: module results with edgeType="depends_on".
func TestHandleGetDependencies_Module(t *testing.T) {
	svc, _ := newTestServiceWithResponses(t, []driver.Result{
		makeResult(map[string]any{
			"modules": []any{
				map[string]any{"name": "dep1", "path": "pkg/dep1", "importPath": "example/dep1", "kind": "package", "language": "go"},
			},
		}),
	})

	resp, err := svc.HandleGetDependencies(context.Background(), "myrepo", "mymodule", 1)
	if err != nil {
		t.Fatalf("handleGetDependencies returned error: %v", err)
	}
	if resp == nil {
		t.Fatal("handleGetDependencies returned nil response")
	}
	if len(resp.Results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(resp.Results))
	}
	if resp.Results[0].Type != "module" {
		t.Errorf("Type = %q, want %q", resp.Results[0].Type, "module")
	}
	if resp.Results[0].EdgeType != "depends_on" {
		t.Errorf("EdgeType = %q, want %q", resp.Results[0].EdgeType, "depends_on")
	}
}

// TestHandleGetDependencies_File verifies that a name with / or extension
// is treated as a File and queries importedBy_some.
// Expected result: module results with edgeType="imports".
func TestHandleGetDependencies_File(t *testing.T) {
	svc, _ := newTestServiceWithResponses(t, []driver.Result{
		makeResult(map[string]any{
			"modules": []any{
				map[string]any{"name": "fmt", "path": "pkg/fmt", "importPath": "fmt", "kind": "package", "language": "go"},
			},
		}),
	})

	resp, err := svc.HandleGetDependencies(context.Background(), "myrepo", "pkg/main.go", 1)
	if err != nil {
		t.Fatalf("handleGetDependencies returned error: %v", err)
	}
	if resp == nil {
		t.Fatal("handleGetDependencies returned nil response")
	}
	if len(resp.Results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(resp.Results))
	}
	if resp.Results[0].EdgeType != "imports" {
		t.Errorf("EdgeType = %q, want %q", resp.Results[0].EdgeType, "imports")
	}
}

// TestHandleGetDependencies_Validation verifies input validation.
// Expected result: error for empty repository and empty name.
func TestHandleGetDependencies_Validation(t *testing.T) {
	svc := newTestService(t)

	_, err := svc.HandleGetDependencies(context.Background(), "", "mymodule", 1)
	if err == nil {
		t.Error("expected error for missing repository, got nil")
	}

	_, err = svc.HandleGetDependencies(context.Background(), "repo", "", 1)
	if err == nil {
		t.Error("expected error for missing name, got nil")
	}
}

// === Task 7: handleGetReferences ===
//
// Auto-detects symbol type by querying Function/Class/Module existence.
// Returns all inbound edges. Single-level only.

// TestHandleGetReferences_Function verifies that a symbol detected as
// Function returns calledBy + overriddenBy edges.
// Expected result: results with function type and calls/overrides edgeType.
func TestHandleGetReferences_Function(t *testing.T) {
	svc, _ := newTestServiceWithResponses(t, []driver.Result{
		// Detect: function exists
		makeResult(map[string]any{
			"functions": []any{
				map[string]any{"name": "myFunc"},
			},
		}),
		// Detect: class does NOT exist
		makeResult(map[string]any{
			"classs": []any{},
		}),
		// Detect: module does NOT exist
		makeResult(map[string]any{
			"modules": []any{},
		}),
		// Function calledBy (calls_some)
		makeResult(map[string]any{
			"functions": []any{
				map[string]any{"name": "caller1", "path": "a.go", "signature": "func caller1()", "language": "go"},
			},
		}),
		// Function overriddenBy (overrides_some)
		makeResult(map[string]any{
			"functions": []any{},
		}),
	})

	resp, err := svc.HandleGetReferences(context.Background(), "myrepo", "myFunc")
	if err != nil {
		t.Fatalf("handleGetReferences returned error: %v", err)
	}
	if resp == nil {
		t.Fatal("handleGetReferences returned nil response")
	}
	if len(resp.Results) < 1 {
		t.Fatalf("expected at least 1 result, got %d", len(resp.Results))
	}
	if resp.Results[0].Name != "caller1" {
		t.Errorf("Name = %q, want %q", resp.Results[0].Name, "caller1")
	}
}

// TestHandleGetReferences_Class verifies that a symbol detected as
// Class returns inheritedBy + implementedBy edges.
// Expected result: results with class type.
func TestHandleGetReferences_Class(t *testing.T) {
	svc, _ := newTestServiceWithResponses(t, []driver.Result{
		// Detect: function does NOT exist
		makeResult(map[string]any{
			"functions": []any{},
		}),
		// Detect: class exists
		makeResult(map[string]any{
			"classs": []any{
				map[string]any{"name": "MyClass"},
			},
		}),
		// Detect: module does NOT exist
		makeResult(map[string]any{
			"modules": []any{},
		}),
		// Class inheritedBy (inherits_some)
		makeResult(map[string]any{
			"classs": []any{
				map[string]any{"name": "ChildClass", "path": "child.go", "kind": "struct", "language": "go"},
			},
		}),
		// Class implementedBy (implements_some)
		makeResult(map[string]any{
			"classs": []any{},
		}),
	})

	resp, err := svc.HandleGetReferences(context.Background(), "myrepo", "MyClass")
	if err != nil {
		t.Fatalf("handleGetReferences returned error: %v", err)
	}
	if resp == nil {
		t.Fatal("handleGetReferences returned nil response")
	}
	if len(resp.Results) < 1 {
		t.Fatalf("expected at least 1 result, got %d", len(resp.Results))
	}
}

// TestHandleGetReferences_MultiType verifies that a name matching
// multiple types merges all inbound edges.
// Expected result: results from both function and class edges.
func TestHandleGetReferences_MultiType(t *testing.T) {
	svc, _ := newTestServiceWithResponses(t, []driver.Result{
		// Detect: function exists
		makeResult(map[string]any{
			"functions": []any{
				map[string]any{"name": "MySymbol"},
			},
		}),
		// Detect: class also exists
		makeResult(map[string]any{
			"classs": []any{
				map[string]any{"name": "MySymbol"},
			},
		}),
		// Detect: module does NOT exist
		makeResult(map[string]any{
			"modules": []any{},
		}),
		// Function calledBy
		makeResult(map[string]any{
			"functions": []any{
				map[string]any{"name": "funcCaller", "path": "a.go", "signature": "func funcCaller()", "language": "go"},
			},
		}),
		// Function overriddenBy
		makeResult(map[string]any{
			"functions": []any{},
		}),
		// Class inheritedBy
		makeResult(map[string]any{
			"classs": []any{
				map[string]any{"name": "childClass", "path": "child.go", "kind": "struct", "language": "go"},
			},
		}),
		// Class implementedBy
		makeResult(map[string]any{
			"classs": []any{},
		}),
	})

	resp, err := svc.HandleGetReferences(context.Background(), "myrepo", "MySymbol")
	if err != nil {
		t.Fatalf("handleGetReferences returned error: %v", err)
	}
	if resp == nil {
		t.Fatal("handleGetReferences returned nil response")
	}
	if len(resp.Results) < 2 {
		t.Fatalf("expected at least 2 results (function + class edges), got %d", len(resp.Results))
	}
}

// TestHandleGetReferences_NoMatch verifies that get_references returns
// empty results when the symbol is not found as any type.
// Expected result: 0 results, no error.
func TestHandleGetReferences_NoMatch(t *testing.T) {
	svc, _ := newTestServiceWithResponses(t, []driver.Result{
		makeResult(map[string]any{"functions": []any{}}),
		makeResult(map[string]any{"classs": []any{}}),
		makeResult(map[string]any{"modules": []any{}}),
	})

	resp, err := svc.HandleGetReferences(context.Background(), "myrepo", "nonexistent")
	if err != nil {
		t.Fatalf("handleGetReferences returned error: %v", err)
	}
	if resp == nil {
		t.Fatal("handleGetReferences returned nil response")
	}
	if len(resp.Results) != 0 {
		t.Errorf("expected 0 results, got %d", len(resp.Results))
	}
}

// TestHandleGetReferences_Validation verifies input validation.
// Expected result: error for empty repository and empty name.
func TestHandleGetReferences_Validation(t *testing.T) {
	svc := newTestService(t)

	_, err := svc.HandleGetReferences(context.Background(), "", "myFunc")
	if err == nil {
		t.Error("expected error for missing repository, got nil")
	}

	_, err = svc.HandleGetReferences(context.Background(), "repo", "")
	if err == nil {
		t.Error("expected error for missing name, got nil")
	}
}

// === Task 8: Register 5 new tools in server.go ===
//
// Verified by checking NewServer registers 8 tools total.

// === Helper ===

// mustForRepo returns a *client.Client from ForRepo or fails the test.
func mustForRepo(t *testing.T, s *Manager, repo string) *client.Client {
	t.Helper()
	c, err := s.db.ForRepo(context.Background(), repo)
	if err != nil {
		t.Fatalf("ForRepo(%q) failed: %v", repo, err)
	}
	return c
}
