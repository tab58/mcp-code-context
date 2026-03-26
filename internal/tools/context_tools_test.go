package mcp

import (
	"context"
	"encoding/json"
	"strings"
	"testing"

	"github.com/tab58/go-ormql/pkg/driver"
)

// ============================================================================
// Task 1: Context response types in types.go
// ============================================================================

// TestRepoMapResponse_TypeStructure verifies that RepoMapResponse has the
// correct fields: Repository, Directories, TotalFiles, TotalSymbols.
// Expected: struct compiles with all required fields and JSON tags.
func TestRepoMapResponse_TypeStructure(t *testing.T) {
	resp := RepoMapResponse{
		Repository: "myrepo",
		Directories: []RepoMapEntry{
			{
				Directory: "cmd/",
				Files: []RepoMapFile{
					{Name: "main.go", Language: "go", SymbolCount: 5},
				},
			},
		},
		TotalFiles:   10,
		TotalSymbols: 25,
	}

	data, err := json.Marshal(resp)
	if err != nil {
		t.Fatalf("failed to marshal RepoMapResponse: %v", err)
	}
	// Verify JSON field names
	s := string(data)
	for _, field := range []string{"repository", "directories", "totalFiles", "totalSymbols"} {
		if !strings.Contains(s, field) {
			t.Errorf("JSON output missing field %q", field)
		}
	}
}

// TestFileOverviewResponse_TypeStructure verifies that FileOverviewResponse has
// Path, Language, Symbols, Total fields.
// Expected: struct compiles with all required fields.
func TestFileOverviewResponse_TypeStructure(t *testing.T) {
	resp := FileOverviewResponse{
		Path:     "pkg/user.go",
		Language: "go",
		Symbols: []OverviewSymbol{
			{Type: "function", Name: "GetUser", Signature: "func GetUser() *User", Visibility: "public", StartingLine: 10, EndingLine: 15},
		},
		Total: 1,
	}

	data, err := json.Marshal(resp)
	if err != nil {
		t.Fatalf("failed to marshal FileOverviewResponse: %v", err)
	}
	s := string(data)
	for _, field := range []string{"path", "language", "symbols", "total"} {
		if !strings.Contains(s, field) {
			t.Errorf("JSON output missing field %q", field)
		}
	}
}

// TestSymbolContextResponse_TypeStructure verifies SymbolContextResponse has
// Symbol (SymbolDetail), Callers, Callees, Siblings fields.
// Expected: struct compiles with all required fields.
func TestSymbolContextResponse_TypeStructure(t *testing.T) {
	resp := SymbolContextResponse{
		Symbol: SymbolDetail{
			Type: "function", Name: "GetUser", Path: "pkg/user.go",
			Language: "go", Signature: "func GetUser() *User",
			Visibility: "public", Source: "func GetUser() *User { return nil }",
			StartingLine: 10, EndingLine: 15,
		},
		Callers: []SymbolSummary{
			{Type: "function", Name: "main", Path: "cmd/main.go", Signature: "func main()"},
		},
		Callees: []SymbolSummary{
			{Type: "function", Name: "findByID", Path: "pkg/db.go", Signature: "func findByID(id string) *User"},
		},
		Siblings: []OverviewSymbol{
			{Type: "function", Name: "DeleteUser", Visibility: "public"},
		},
	}

	data, err := json.Marshal(resp)
	if err != nil {
		t.Fatalf("failed to marshal SymbolContextResponse: %v", err)
	}
	s := string(data)
	for _, field := range []string{"symbol", "callers", "callees", "siblings"} {
		if !strings.Contains(s, field) {
			t.Errorf("JSON output missing field %q", field)
		}
	}
}

// TestReadSourceResponse_TypeStructure verifies ReadSourceResponse has
// Results ([]ReadSourceResult) and Total fields.
// Expected: struct compiles with all required fields.
func TestReadSourceResponse_TypeStructure(t *testing.T) {
	resp := ReadSourceResponse{
		Results: []ReadSourceResult{
			{Type: "function", Name: "GetUser", Path: "pkg/user.go",
				Source: "func GetUser() *User { return nil }", StartingLine: 10, EndingLine: 15},
		},
		Total: 1,
	}

	data, err := json.Marshal(resp)
	if err != nil {
		t.Fatalf("failed to marshal ReadSourceResponse: %v", err)
	}
	s := string(data)
	for _, field := range []string{"results", "total"} {
		if !strings.Contains(s, field) {
			t.Errorf("JSON output missing field %q", field)
		}
	}
}

// TestOverviewSymbol_KindField verifies that OverviewSymbol has a Kind field
// for class-type symbols (struct, interface, enum).
// Expected: Kind field exists in OverviewSymbol.
func TestOverviewSymbol_KindField(t *testing.T) {
	sym := OverviewSymbol{
		Type: "class",
		Name: "User",
		Kind: "struct",
	}
	if sym.Kind != "struct" {
		t.Errorf("OverviewSymbol.Kind = %q, want %q", sym.Kind, "struct")
	}
}

// TestSymbolDetail_SourceField verifies that SymbolDetail has Source field
// (distinguishing it from OverviewSymbol which has no source).
// Expected: Source field exists and is populated.
func TestSymbolDetail_SourceField(t *testing.T) {
	detail := SymbolDetail{
		Name:   "GetUser",
		Source: "func GetUser() *User { return nil }",
	}
	if detail.Source == "" {
		t.Error("SymbolDetail.Source should be populated, got empty")
	}
}

// ============================================================================
// Task 2: Remove source from gqlFindFunctions
// ============================================================================

// TestFindFunction_NoSource verifies that find_function results do NOT
// include source code (Source field should be empty) even when source data
// is available in the graph response.
// Expected: result.Source == "" even when the driver returns source data.
func TestFindFunction_NoSource(t *testing.T) {
	// The driver returns a function WITH source data — the handler should NOT map it
	svc, _ := newTestServiceWithResponses(t, []driver.Result{
		makeResult(map[string]any{
			"functions": []any{
				map[string]any{
					"name": "GetUser", "path": "pkg/user.go",
					"source": "func GetUser() *User { return nil }",
					"signature": "func GetUser() *User", "language": "go",
					"visibility": "public", "startingLine": float64(10), "endingLine": float64(15),
				},
			},
		}),
	})

	resp, err := svc.HandleFindFunction(context.Background(), "myrepo", "GetUser")
	if err != nil {
		t.Fatalf("handleFindFunction error: %v", err)
	}
	if len(resp.Results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(resp.Results))
	}
	// find_function should NOT include source — even though it's in the data
	if resp.Results[0].Source != "" {
		t.Errorf("find_function should not return source, got %q", resp.Results[0].Source)
	}
}

// TestGqlFindFunctions_NoSourceField verifies the gqlFindFunctions constant
// does NOT contain the word "source".
// Expected: gqlFindFunctions query string does not mention "source".
func TestGqlFindFunctions_NoSourceField(t *testing.T) {
	if strings.Contains(gqlFindFunctions, "source") {
		t.Error("gqlFindFunctions should not contain 'source' field")
	}
}

// ============================================================================
// Task 3: context_tools.go infrastructure (GraphQL constants + helpers)
// ============================================================================

// TestGqlConstants_ContextTools verifies all 7 GraphQL constants are defined.
// Expected: all constants are non-empty strings.
func TestGqlConstants_ContextTools(t *testing.T) {
	constants := map[string]string{
		"gqlRepoFiles":             gqlRepoFiles,
		"gqlRepoFunctionPaths":     gqlRepoFunctionPaths,
		"gqlRepoClassPaths":        gqlRepoClassPaths,
		"gqlFileOverviewFunctions": gqlFileOverviewFunctions,
		"gqlFileOverviewClasses":   gqlFileOverviewClasses,
		"gqlSymbolFunction":        gqlSymbolFunction,
		"gqlSymbolClass":           gqlSymbolClass,
	}
	for name, val := range constants {
		if val == "" {
			t.Errorf("%s is empty", name)
		}
	}
}

// TestGqlSymbolFunction_IncludesSource verifies gqlSymbolFunction includes
// the "source" field (unlike gqlFindFunctions which does not).
// Expected: gqlSymbolFunction contains "source".
func TestGqlSymbolFunction_IncludesSource(t *testing.T) {
	if !strings.Contains(gqlSymbolFunction, "source") {
		t.Error("gqlSymbolFunction should include 'source' field")
	}
}

// TestGqlSymbolClass_IncludesSource verifies gqlSymbolClass includes
// the "source" field for read_source and get_symbol_context.
// Expected: gqlSymbolClass contains "source".
func TestGqlSymbolClass_IncludesSource(t *testing.T) {
	if !strings.Contains(gqlSymbolClass, "source") {
		t.Error("gqlSymbolClass should include 'source' field")
	}
}

// TestGqlFileOverviewFunctions_NoSource verifies the overview query
// does NOT include source (lightweight signatures only).
// Expected: gqlFileOverviewFunctions does not contain "source".
func TestGqlFileOverviewFunctions_NoSource(t *testing.T) {
	if strings.Contains(gqlFileOverviewFunctions, "source") {
		t.Error("gqlFileOverviewFunctions should NOT include 'source' field")
	}
}

// TestTraversalToSummary converts a TraversalResult into a lightweight
// SymbolSummary (name+path+signature only).
// Expected: SymbolSummary populated from TraversalResult fields.
func TestTraversalToSummary(t *testing.T) {
	tr := TraversalResult{
		Type:      "function",
		Name:      "GetUser",
		Path:      "pkg/user.go",
		Signature: "func GetUser() *User",
		Kind:      "method",
	}
	summary := traversalToSummary(tr)
	if summary.Name != "GetUser" {
		t.Errorf("SymbolSummary.Name = %q, want %q", summary.Name, "GetUser")
	}
	if summary.Path != "pkg/user.go" {
		t.Errorf("SymbolSummary.Path = %q, want %q", summary.Path, "pkg/user.go")
	}
	if summary.Signature != "func GetUser() *User" {
		t.Errorf("SymbolSummary.Signature = %q, want %q", summary.Signature, "func GetUser() *User")
	}
	if summary.Type != "function" {
		t.Errorf("SymbolSummary.Type = %q, want %q", summary.Type, "function")
	}
}

// ============================================================================
// Task 4: handleGetRepoMap
// ============================================================================

// TestHandleGetRepoMap_GroupsByDirectory verifies repo map groups files by
// directory and counts symbols per file.
// Expected: directories sorted, files within directories sorted, symbol counts correct.
func TestHandleGetRepoMap_GroupsByDirectory(t *testing.T) {
	svc, _ := newTestServiceWithResponses(t, []driver.Result{
		// gqlRepoFiles — all files in the repo
		makeResult(map[string]any{
			"files": []any{
				map[string]any{"path": "cmd/main.go", "filename": "main.go", "language": "go"},
				map[string]any{"path": "pkg/user.go", "filename": "user.go", "language": "go"},
				map[string]any{"path": "pkg/db.go", "filename": "db.go", "language": "go"},
			},
		}),
		// gqlRepoFunctionPaths — function paths for symbol counting
		makeResult(map[string]any{
			"functions": []any{
				map[string]any{"path": "cmd/main.go"},
				map[string]any{"path": "pkg/user.go"},
				map[string]any{"path": "pkg/user.go"},
			},
		}),
		// gqlRepoClassPaths — class paths for symbol counting
		makeResult(map[string]any{
			"classs": []any{
				map[string]any{"path": "pkg/user.go"},
			},
		}),
	})

	resp, err := svc.HandleGetRepoMap(context.Background(), "myrepo")
	if err != nil {
		t.Fatalf("handleGetRepoMap error: %v", err)
	}
	if resp == nil {
		t.Fatal("handleGetRepoMap returned nil")
	}
	if resp.Repository != "myrepo" {
		t.Errorf("Repository = %q, want %q", resp.Repository, "myrepo")
	}
	if resp.TotalFiles != 3 {
		t.Errorf("TotalFiles = %d, want 3", resp.TotalFiles)
	}
	if resp.TotalSymbols != 4 {
		t.Errorf("TotalSymbols = %d, want 4 (3 functions + 1 class)", resp.TotalSymbols)
	}
	// Should have 2 directories: "cmd" and "pkg"
	if len(resp.Directories) != 2 {
		t.Fatalf("expected 2 directories, got %d", len(resp.Directories))
	}
	// Directories should be sorted alphabetically
	if resp.Directories[0].Directory != "cmd" {
		t.Errorf("first directory = %q, want %q", resp.Directories[0].Directory, "cmd")
	}
	if resp.Directories[1].Directory != "pkg" {
		t.Errorf("second directory = %q, want %q", resp.Directories[1].Directory, "pkg")
	}
	// pkg/user.go should have 3 symbols (2 functions + 1 class)
	pkgDir := resp.Directories[1]
	for _, f := range pkgDir.Files {
		if f.Name == "user.go" && f.SymbolCount != 3 {
			t.Errorf("user.go SymbolCount = %d, want 3", f.SymbolCount)
		}
	}
}

// TestHandleGetRepoMap_EmptyRepo verifies repo map returns empty when
// no files exist in the repository.
// Expected: empty directories, TotalFiles=0, TotalSymbols=0.
func TestHandleGetRepoMap_EmptyRepo(t *testing.T) {
	svc, _ := newTestServiceWithResponses(t, []driver.Result{
		makeResult(map[string]any{"files": []any{}}),
		makeResult(map[string]any{"functions": []any{}}),
		makeResult(map[string]any{"classs": []any{}}),
	})

	resp, err := svc.HandleGetRepoMap(context.Background(), "myrepo")
	if err != nil {
		t.Fatalf("handleGetRepoMap error: %v", err)
	}
	if resp.TotalFiles != 0 {
		t.Errorf("TotalFiles = %d, want 0", resp.TotalFiles)
	}
	if resp.TotalSymbols != 0 {
		t.Errorf("TotalSymbols = %d, want 0", resp.TotalSymbols)
	}
	if len(resp.Directories) != 0 {
		t.Errorf("expected 0 directories, got %d", len(resp.Directories))
	}
}

// ============================================================================
// Task 5: handleGetFileOverview
// ============================================================================

// TestHandleGetFileOverview_SortsByLine verifies file overview returns
// symbols sorted by startingLine ascending.
// Expected: symbols in ascending startingLine order.
func TestHandleGetFileOverview_SortsByLine(t *testing.T) {
	svc, _ := newTestServiceWithResponses(t, []driver.Result{
		// gqlFileOverviewFunctions
		makeResult(map[string]any{
			"functions": []any{
				map[string]any{"name": "Delete", "signature": "func Delete()", "visibility": "public", "startingLine": float64(50), "endingLine": float64(60)},
				map[string]any{"name": "Create", "signature": "func Create()", "visibility": "public", "startingLine": float64(10), "endingLine": float64(20)},
			},
		}),
		// gqlFileOverviewClasses
		makeResult(map[string]any{
			"classs": []any{
				map[string]any{"name": "User", "kind": "struct", "visibility": "public", "startingLine": float64(1), "endingLine": float64(8)},
			},
		}),
	})

	resp, err := svc.HandleGetFileOverview(context.Background(), "myrepo", "pkg/user.go")
	if err != nil {
		t.Fatalf("handleGetFileOverview error: %v", err)
	}
	if resp == nil {
		t.Fatal("handleGetFileOverview returned nil")
	}
	if resp.Path != "pkg/user.go" {
		t.Errorf("Path = %q, want %q", resp.Path, "pkg/user.go")
	}
	if len(resp.Symbols) != 3 {
		t.Fatalf("expected 3 symbols, got %d", len(resp.Symbols))
	}
	// Should be sorted: User(1), Create(10), Delete(50)
	if resp.Symbols[0].Name != "User" {
		t.Errorf("first symbol = %q, want %q (lowest startingLine)", resp.Symbols[0].Name, "User")
	}
	if resp.Symbols[1].Name != "Create" {
		t.Errorf("second symbol = %q, want %q", resp.Symbols[1].Name, "Create")
	}
	if resp.Symbols[2].Name != "Delete" {
		t.Errorf("third symbol = %q, want %q", resp.Symbols[2].Name, "Delete")
	}
	if resp.Total != 3 {
		t.Errorf("Total = %d, want 3", resp.Total)
	}
}

// TestHandleGetFileOverview_EmptyFile verifies empty result for a file
// with no symbols.
// Expected: Symbols is empty, Total=0.
func TestHandleGetFileOverview_EmptyFile(t *testing.T) {
	svc, _ := newTestServiceWithResponses(t, []driver.Result{
		makeResult(map[string]any{"functions": []any{}}),
		makeResult(map[string]any{"classs": []any{}}),
	})

	resp, err := svc.HandleGetFileOverview(context.Background(), "myrepo", "README.md")
	if err != nil {
		t.Fatalf("handleGetFileOverview error: %v", err)
	}
	if resp.Total != 0 {
		t.Errorf("Total = %d, want 0", resp.Total)
	}
	if len(resp.Symbols) != 0 {
		t.Errorf("expected 0 symbols, got %d", len(resp.Symbols))
	}
}

// TestHandleGetFileOverview_MixedTypes verifies both functions and classes
// are merged into the same symbol list.
// Expected: both "function" and "class" types present in results.
func TestHandleGetFileOverview_MixedTypes(t *testing.T) {
	svc, _ := newTestServiceWithResponses(t, []driver.Result{
		makeResult(map[string]any{
			"functions": []any{
				map[string]any{"name": "NewUser", "signature": "func NewUser() *User", "visibility": "public", "startingLine": float64(20), "endingLine": float64(25)},
			},
		}),
		makeResult(map[string]any{
			"classs": []any{
				map[string]any{"name": "User", "kind": "struct", "visibility": "public", "startingLine": float64(1), "endingLine": float64(15)},
			},
		}),
	})

	resp, err := svc.HandleGetFileOverview(context.Background(), "myrepo", "pkg/user.go")
	if err != nil {
		t.Fatalf("handleGetFileOverview error: %v", err)
	}
	if len(resp.Symbols) != 2 {
		t.Fatalf("expected 2 symbols, got %d", len(resp.Symbols))
	}
	hasFunc, hasClass := false, false
	for _, sym := range resp.Symbols {
		if sym.Type == "function" {
			hasFunc = true
		}
		if sym.Type == "class" {
			hasClass = true
		}
	}
	if !hasFunc {
		t.Error("expected at least one function type symbol")
	}
	if !hasClass {
		t.Error("expected at least one class type symbol")
	}
}

// ============================================================================
// Task 6: handleGetSymbolContext
// ============================================================================

// TestHandleGetSymbolContext_Function verifies get_symbol_context returns
// a function with source, callers, callees, and siblings.
// Expected: SymbolDetail populated, callers/callees/siblings populated.
func TestHandleGetSymbolContext_Function(t *testing.T) {
	svc, _ := newTestServiceWithResponses(t, []driver.Result{
		// gqlSymbolFunction — found the function
		makeResult(map[string]any{
			"functions": []any{
				map[string]any{
					"name": "GetUser", "path": "pkg/user.go",
					"source": "func GetUser() *User { return nil }",
					"signature": "func GetUser() *User", "language": "go",
					"visibility": "public", "startingLine": float64(10), "endingLine": float64(15),
				},
			},
		}),
		// gqlFindCallers — callers of GetUser
		makeResult(map[string]any{
			"functions": []any{
				map[string]any{"name": "main", "path": "cmd/main.go", "signature": "func main()", "language": "go"},
			},
		}),
		// gqlFindCallees — callees of GetUser
		makeResult(map[string]any{
			"functions": []any{
				map[string]any{"name": "findByID", "path": "pkg/db.go", "signature": "func findByID(id string) *User", "language": "go"},
			},
		}),
		// gqlFileOverviewFunctions — siblings (same file)
		makeResult(map[string]any{
			"functions": []any{
				map[string]any{"name": "GetUser", "signature": "func GetUser() *User", "visibility": "public", "startingLine": float64(10), "endingLine": float64(15)},
				map[string]any{"name": "DeleteUser", "signature": "func DeleteUser()", "visibility": "public", "startingLine": float64(20), "endingLine": float64(25)},
			},
		}),
		// gqlFileOverviewClasses — siblings (same file, classes)
		makeResult(map[string]any{"classs": []any{}}),
	})

	resp, err := svc.HandleGetSymbolContext(context.Background(), "myrepo", "GetUser")
	if err != nil {
		t.Fatalf("handleGetSymbolContext error: %v", err)
	}
	if resp == nil {
		t.Fatal("handleGetSymbolContext returned nil")
	}
	// Symbol detail should have source
	if resp.Symbol.Name != "GetUser" {
		t.Errorf("Symbol.Name = %q, want %q", resp.Symbol.Name, "GetUser")
	}
	if resp.Symbol.Source == "" {
		t.Error("Symbol.Source should be populated")
	}
	if resp.Symbol.Type != "function" {
		t.Errorf("Symbol.Type = %q, want %q", resp.Symbol.Type, "function")
	}
	// Should have callers
	if len(resp.Callers) == 0 {
		t.Error("expected callers to be populated")
	}
	// Should have callees
	if len(resp.Callees) == 0 {
		t.Error("expected callees to be populated")
	}
	// Should have siblings (excluding self)
	if len(resp.Siblings) == 0 {
		t.Error("expected siblings to be populated (excluding self)")
	}
	// Self should be excluded from siblings
	for _, sib := range resp.Siblings {
		if sib.Name == "GetUser" {
			t.Error("siblings should not include the target symbol itself")
		}
	}
}

// TestHandleGetSymbolContext_Class verifies get_symbol_context works for
// class symbols (no callers/callees, has siblings).
// Expected: SymbolDetail type="class", no callers/callees.
func TestHandleGetSymbolContext_Class(t *testing.T) {
	svc, _ := newTestServiceWithResponses(t, []driver.Result{
		// gqlSymbolFunction — not found (empty)
		makeResult(map[string]any{"functions": []any{}}),
		// gqlSymbolClass — found the class
		makeResult(map[string]any{
			"classs": []any{
				map[string]any{
					"name": "User", "path": "pkg/user.go",
					"source": "type User struct { ID string }",
					"kind": "struct", "language": "go",
					"visibility": "public", "startingLine": float64(1), "endingLine": float64(8),
				},
			},
		}),
		// gqlFileOverviewFunctions — siblings
		makeResult(map[string]any{
			"functions": []any{
				map[string]any{"name": "NewUser", "signature": "func NewUser() *User", "visibility": "public", "startingLine": float64(10), "endingLine": float64(15)},
			},
		}),
		// gqlFileOverviewClasses — siblings
		makeResult(map[string]any{
			"classs": []any{
				map[string]any{"name": "User", "kind": "struct", "visibility": "public", "startingLine": float64(1), "endingLine": float64(8)},
			},
		}),
	})

	resp, err := svc.HandleGetSymbolContext(context.Background(), "myrepo", "User")
	if err != nil {
		t.Fatalf("handleGetSymbolContext error: %v", err)
	}
	if resp == nil {
		t.Fatal("handleGetSymbolContext returned nil")
	}
	if resp.Symbol.Type != "class" {
		t.Errorf("Symbol.Type = %q, want %q", resp.Symbol.Type, "class")
	}
	if resp.Symbol.Source == "" {
		t.Error("Symbol.Source should be populated for class")
	}
	// Classes should not have callers/callees
	if len(resp.Callers) > 0 {
		t.Errorf("classes should have no callers, got %d", len(resp.Callers))
	}
	if len(resp.Callees) > 0 {
		t.Errorf("classes should have no callees, got %d", len(resp.Callees))
	}
}

// TestHandleGetSymbolContext_NotFound verifies get_symbol_context returns
// an error when the symbol doesn't exist.
// Expected: error containing "not found".
func TestHandleGetSymbolContext_NotFound(t *testing.T) {
	svc, _ := newTestServiceWithResponses(t, []driver.Result{
		makeResult(map[string]any{"functions": []any{}}),
		makeResult(map[string]any{"classs": []any{}}),
	})

	_, err := svc.HandleGetSymbolContext(context.Background(), "myrepo", "NonExistent")
	if err == nil {
		t.Fatal("expected error for missing symbol, got nil")
	}
	if !strings.Contains(err.Error(), "not found") {
		t.Errorf("error should mention 'not found', got: %v", err)
	}
}

// ============================================================================
// Task 7: handleReadSource
// ============================================================================

// TestHandleReadSource_BatchFetch verifies read_source fetches source for
// multiple symbols in one call.
// Expected: results contain source for each found symbol.
func TestHandleReadSource_BatchFetch(t *testing.T) {
	svc, _ := newTestServiceWithResponses(t, []driver.Result{
		// First name "GetUser": try function — found
		makeResult(map[string]any{
			"functions": []any{
				map[string]any{
					"name": "GetUser", "path": "pkg/user.go",
					"source": "func GetUser() *User { return nil }",
					"startingLine": float64(10), "endingLine": float64(15),
				},
			},
		}),
		// Second name "User": try function — not found
		makeResult(map[string]any{"functions": []any{}}),
		// Second name "User": try class — found
		makeResult(map[string]any{
			"classs": []any{
				map[string]any{
					"name": "User", "path": "pkg/user.go",
					"source": "type User struct { ID string }",
					"startingLine": float64(1), "endingLine": float64(8),
				},
			},
		}),
	})

	resp, err := svc.HandleReadSource(context.Background(), "myrepo", []string{"GetUser", "User"})
	if err != nil {
		t.Fatalf("handleReadSource error: %v", err)
	}
	if resp == nil {
		t.Fatal("handleReadSource returned nil")
	}
	if resp.Total != 2 {
		t.Errorf("Total = %d, want 2", resp.Total)
	}
	if len(resp.Results) != 2 {
		t.Fatalf("expected 2 results, got %d", len(resp.Results))
	}
	// Check types
	if resp.Results[0].Type != "function" {
		t.Errorf("first result type = %q, want %q", resp.Results[0].Type, "function")
	}
	if resp.Results[1].Type != "class" {
		t.Errorf("second result type = %q, want %q", resp.Results[1].Type, "class")
	}
	// All should have source
	for i, r := range resp.Results {
		if r.Source == "" {
			t.Errorf("result[%d].Source should be populated", i)
		}
	}
}

// TestHandleReadSource_MissingSymbolSkipped verifies that read_source
// silently skips symbols that don't exist.
// Expected: no error, only found symbols in results.
func TestHandleReadSource_MissingSymbolSkipped(t *testing.T) {
	svc, _ := newTestServiceWithResponses(t, []driver.Result{
		// "GetUser": function — found
		makeResult(map[string]any{
			"functions": []any{
				map[string]any{
					"name": "GetUser", "path": "pkg/user.go",
					"source": "func GetUser() {}", "startingLine": float64(10), "endingLine": float64(15),
				},
			},
		}),
		// "NonExistent": function — not found
		makeResult(map[string]any{"functions": []any{}}),
		// "NonExistent": class — not found
		makeResult(map[string]any{"classs": []any{}}),
	})

	resp, err := svc.HandleReadSource(context.Background(), "myrepo", []string{"GetUser", "NonExistent"})
	if err != nil {
		t.Fatalf("handleReadSource error: %v", err)
	}
	if resp.Total != 1 {
		t.Errorf("Total = %d, want 1 (one found, one missing)", resp.Total)
	}
	if len(resp.Results) != 1 {
		t.Errorf("expected 1 result, got %d", len(resp.Results))
	}
}

// TestHandleReadSource_EmptyNames verifies read_source returns error for
// empty names list.
// Expected: error mentioning "names".
func TestHandleReadSource_EmptyNames(t *testing.T) {
	svc := newTestService(t)
	_, err := svc.HandleReadSource(context.Background(), "myrepo", []string{})
	if err == nil {
		t.Fatal("expected error for empty names, got nil")
	}
	if !strings.Contains(err.Error(), "names") {
		t.Errorf("error should mention 'names', got: %v", err)
	}
}

// ============================================================================
// Input validation for all 4 context handlers
// ============================================================================

// TestValidation_GetRepoMap_MissingRepository verifies get_repo_map
// returns error when repository is empty.
// Expected: error mentioning "repository".
func TestValidation_GetRepoMap_MissingRepository(t *testing.T) {
	svc := newTestService(t)
	_, err := svc.HandleGetRepoMap(context.Background(), "")
	if err == nil {
		t.Fatal("expected error for missing repository, got nil")
	}
	if !strings.Contains(err.Error(), "repository") {
		t.Errorf("error should mention 'repository', got: %v", err)
	}
}

// TestValidation_GetFileOverview_MissingRepository verifies get_file_overview
// returns error when repository is empty.
// Expected: error mentioning "repository".
func TestValidation_GetFileOverview_MissingRepository(t *testing.T) {
	svc := newTestService(t)
	_, err := svc.HandleGetFileOverview(context.Background(), "", "some/path.go")
	if err == nil {
		t.Fatal("expected error for missing repository, got nil")
	}
	if !strings.Contains(err.Error(), "repository") {
		t.Errorf("error should mention 'repository', got: %v", err)
	}
}

// TestValidation_GetFileOverview_MissingPath verifies get_file_overview
// returns error when path is empty.
// Expected: error mentioning "path".
func TestValidation_GetFileOverview_MissingPath(t *testing.T) {
	svc := newTestService(t)
	_, err := svc.HandleGetFileOverview(context.Background(), "myrepo", "")
	if err == nil {
		t.Fatal("expected error for missing path, got nil")
	}
	if !strings.Contains(err.Error(), "path") {
		t.Errorf("error should mention 'path', got: %v", err)
	}
}

// TestValidation_GetSymbolContext_MissingRepository verifies get_symbol_context
// returns error when repository is empty.
// Expected: error mentioning "repository".
func TestValidation_GetSymbolContext_MissingRepository(t *testing.T) {
	svc := newTestService(t)
	_, err := svc.HandleGetSymbolContext(context.Background(), "", "GetUser")
	if err == nil {
		t.Fatal("expected error for missing repository, got nil")
	}
	if !strings.Contains(err.Error(), "repository") {
		t.Errorf("error should mention 'repository', got: %v", err)
	}
}

// TestValidation_GetSymbolContext_MissingName verifies get_symbol_context
// returns error when name is empty.
// Expected: error mentioning "name".
func TestValidation_GetSymbolContext_MissingName(t *testing.T) {
	svc := newTestService(t)
	_, err := svc.HandleGetSymbolContext(context.Background(), "myrepo", "")
	if err == nil {
		t.Fatal("expected error for missing name, got nil")
	}
	if !strings.Contains(err.Error(), "name") {
		t.Errorf("error should mention 'name', got: %v", err)
	}
}

// TestValidation_ReadSource_MissingRepository verifies read_source
// returns error when repository is empty.
// Expected: error mentioning "repository".
func TestValidation_ReadSource_MissingRepository(t *testing.T) {
	svc := newTestService(t)
	_, err := svc.HandleReadSource(context.Background(), "", []string{"fn"})
	if err == nil {
		t.Fatal("expected error for missing repository, got nil")
	}
	if !strings.Contains(err.Error(), "repository") {
		t.Errorf("error should mention 'repository', got: %v", err)
	}
}

// TestValidation_ReadSource_NilNames verifies read_source
// returns error when names is nil.
// Expected: error mentioning "names".
func TestValidation_ReadSource_NilNames(t *testing.T) {
	svc := newTestService(t)
	_, err := svc.HandleReadSource(context.Background(), "myrepo", nil)
	if err == nil {
		t.Fatal("expected error for nil names, got nil")
	}
	if !strings.Contains(err.Error(), "names") {
		t.Errorf("error should mention 'names', got: %v", err)
	}
}

// ============================================================================
// Task 10: Verify search_code delegation after find_function source removal
// ============================================================================

// TestSearchCode_DelegatesToFindFunction_NoSource verifies that search_code
// (which delegates to find_function) also returns no source even when
// source data is available in the graph response.
// Expected: results from search_code have empty Source field.
func TestSearchCode_DelegatesToFindFunction_NoSource(t *testing.T) {
	svc, _ := newTestServiceWithResponses(t, []driver.Result{
		makeResult(map[string]any{
			"functions": []any{
				map[string]any{
					"name": "GetUser", "path": "pkg/user.go",
					"source": "func GetUser() *User { return nil }",
					"signature": "func GetUser() *User", "language": "go",
					"startingLine": float64(10), "endingLine": float64(15),
				},
			},
		}),
	})

	resp, err := svc.HandleSearchCode(context.Background(), "myrepo", "GetUser", 10)
	if err != nil {
		t.Fatalf("handleSearchCode error: %v", err)
	}
	if len(resp.Results) == 0 {
		t.Fatal("expected results from search_code")
	}
	for i, r := range resp.Results {
		if r.Source != "" {
			t.Errorf("search_code result[%d].Source should be empty, got %q", i, r.Source)
		}
	}
}

// ============================================================================
// Task 11: Unit tests for context tool handlers (covered above, plus extras)
// ============================================================================

// TestHandleGetRepoMap_FilesWithinDirSorted verifies files within each
// directory are sorted alphabetically by name.
// Expected: files within "pkg" are db.go then user.go.
func TestHandleGetRepoMap_FilesWithinDirSorted(t *testing.T) {
	svc, _ := newTestServiceWithResponses(t, []driver.Result{
		makeResult(map[string]any{
			"files": []any{
				map[string]any{"path": "pkg/user.go", "filename": "user.go", "language": "go"},
				map[string]any{"path": "pkg/db.go", "filename": "db.go", "language": "go"},
			},
		}),
		makeResult(map[string]any{"functions": []any{}}),
		makeResult(map[string]any{"classs": []any{}}),
	})

	resp, err := svc.HandleGetRepoMap(context.Background(), "myrepo")
	if err != nil {
		t.Fatalf("handleGetRepoMap error: %v", err)
	}
	if len(resp.Directories) != 1 {
		t.Fatalf("expected 1 directory, got %d", len(resp.Directories))
	}
	files := resp.Directories[0].Files
	if len(files) != 2 {
		t.Fatalf("expected 2 files, got %d", len(files))
	}
	if files[0].Name != "db.go" {
		t.Errorf("first file = %q, want %q (alphabetical)", files[0].Name, "db.go")
	}
	if files[1].Name != "user.go" {
		t.Errorf("second file = %q, want %q (alphabetical)", files[1].Name, "user.go")
	}
}

// TestHandleReadSource_AllMissing verifies read_source returns 0 total
// when none of the requested symbols exist.
// Expected: Total=0, empty results.
func TestHandleReadSource_AllMissing(t *testing.T) {
	svc, _ := newTestServiceWithResponses(t, []driver.Result{
		// "Alpha" function — not found
		makeResult(map[string]any{"functions": []any{}}),
		// "Alpha" class — not found
		makeResult(map[string]any{"classs": []any{}}),
	})

	resp, err := svc.HandleReadSource(context.Background(), "myrepo", []string{"Alpha"})
	if err != nil {
		t.Fatalf("handleReadSource error: %v", err)
	}
	if resp.Total != 0 {
		t.Errorf("Total = %d, want 0", resp.Total)
	}
	if len(resp.Results) != 0 {
		t.Errorf("expected 0 results, got %d", len(resp.Results))
	}
}

// ============================================================================
// Task 12: Unit tests for find_function source removal (covered above)
// ============================================================================

// TestFindFunction_FieldsPresent verifies find_function still returns
// all metadata fields (name, path, signature, etc.) after source removal.
// Expected: all non-source fields populated.
func TestFindFunction_FieldsPresent(t *testing.T) {
	svc, _ := newTestServiceWithResponses(t, []driver.Result{
		makeResult(map[string]any{
			"functions": []any{
				map[string]any{
					"name": "GetUser", "path": "pkg/user.go",
					"signature": "func GetUser() *User", "language": "go",
					"visibility": "public", "startingLine": float64(10), "endingLine": float64(15),
				},
			},
		}),
	})

	resp, err := svc.HandleFindFunction(context.Background(), "myrepo", "GetUser")
	if err != nil {
		t.Fatalf("error: %v", err)
	}
	r := resp.Results[0]
	if r.Name != "GetUser" {
		t.Errorf("Name = %q, want %q", r.Name, "GetUser")
	}
	if r.Signature != "func GetUser() *User" {
		t.Errorf("Signature = %q", r.Signature)
	}
	if r.Language != "go" {
		t.Errorf("Language = %q", r.Language)
	}
	if r.Visibility != "public" {
		t.Errorf("Visibility = %q", r.Visibility)
	}
	if r.StartingLine != 10 {
		t.Errorf("StartingLine = %d", r.StartingLine)
	}
	if r.EndingLine != 15 {
		t.Errorf("EndingLine = %d", r.EndingLine)
	}
}

// ============================================================================
// Task 13: Build/vet/race verification — structural tests
// ============================================================================

// TestContextToolsCompile verifies that context_tools.go compiles and all
// handler methods are callable on *Service.
// Expected: this test compiles and runs.
func TestContextToolsCompile(t *testing.T) {
	svc := newTestService(t)
	// Verify all handler methods exist as compile-time check
	var _ func(context.Context, string) (*RepoMapResponse, error) = svc.HandleGetRepoMap
	var _ func(context.Context, string, string) (*FileOverviewResponse, error) = svc.HandleGetFileOverview
	var _ func(context.Context, string, string) (*SymbolContextResponse, error) = svc.HandleGetSymbolContext
	var _ func(context.Context, string, []string) (*ReadSourceResponse, error) = svc.HandleReadSource
}
