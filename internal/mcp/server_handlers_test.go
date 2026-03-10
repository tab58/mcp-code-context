package mcp

import (
	"context"
	"encoding/json"
	"strings"
	"testing"

	mcplib "github.com/mark3labs/mcp-go/mcp"
	"github.com/tab58/go-ormql/pkg/driver"
)

// makeMCPRequest builds a CallToolRequest with the given tool name and arguments map.
func makeMCPRequest(name string, args map[string]any) mcplib.CallToolRequest {
	return mcplib.CallToolRequest{
		Params: mcplib.CallToolParams{
			Name:      name,
			Arguments: args,
		},
	}
}

// --- mcpHandleFindFunction ---

func TestMCPHandleFindFunction_Success(t *testing.T) {
	s, _ := newTestServerWithResponses(t, []driver.Result{
		makeResult(map[string]any{
			"functions": []any{
				map[string]any{
					"name": "myFunc", "path": "pkg/a.go",
					"source": "func myFunc() {}", "signature": "func myFunc()",
					"language": "go", "visibility": "public",
					"startingLine": float64(1), "endingLine": float64(3),
				},
			},
		}),
	})

	req := makeMCPRequest("find_function", map[string]any{
		"repository": "myrepo",
		"name":       "myFunc",
	})
	result, err := s.mcpHandleFindFunction(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result == nil {
		t.Fatal("expected non-nil result")
	}
	if result.IsError {
		t.Error("expected success, got error result")
	}
}

func TestMCPHandleFindFunction_ValidationError(t *testing.T) {
	s := newTestServer(t)
	req := makeMCPRequest("find_function", map[string]any{
		"repository": "",
		"name":       "myFunc",
	})
	result, err := s.mcpHandleFindFunction(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected Go error: %v", err)
	}
	if result == nil {
		t.Fatal("expected non-nil result")
	}
	if !result.IsError {
		t.Error("expected error result for missing repository")
	}
}

// --- mcpHandleFindFile ---

func TestMCPHandleFindFile_Success(t *testing.T) {
	s, _ := newTestServerWithResponses(t, []driver.Result{
		makeResult(map[string]any{
			"files": []any{
				map[string]any{"name": "main.go", "path": "cmd/main.go", "language": "go", "lineCount": float64(50)},
			},
		}),
	})

	req := makeMCPRequest("find_file", map[string]any{
		"repository": "myrepo",
		"pattern":    "*.go",
	})
	result, err := s.mcpHandleFindFile(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result == nil || result.IsError {
		t.Error("expected successful result")
	}
}

func TestMCPHandleFindFile_ValidationError(t *testing.T) {
	s := newTestServer(t)
	req := makeMCPRequest("find_file", map[string]any{
		"repository": "myrepo",
		"pattern":    "",
	})
	result, err := s.mcpHandleFindFile(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected Go error: %v", err)
	}
	if !result.IsError {
		t.Error("expected error result for missing pattern")
	}
}

// --- mcpHandleSearchCode ---

func TestMCPHandleSearchCode_Success(t *testing.T) {
	s, _ := newTestServerWithResponses(t, []driver.Result{
		makeResult(map[string]any{
			"functions": []any{
				map[string]any{
					"name": "getUserByID", "path": "pkg/user.go",
					"source": "func getUserByID() {}", "language": "go",
				},
			},
		}),
	})

	req := makeMCPRequest("search_code_names", map[string]any{
		"repository": "myrepo",
		"query":      "getUserByID",
		"limit":      float64(5),
	})
	result, err := s.mcpHandleSearchCode(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result == nil || result.IsError {
		t.Error("expected successful result")
	}
}

func TestMCPHandleSearchCode_DefaultLimit(t *testing.T) {
	s, _ := newTestServerWithResponses(t, []driver.Result{
		makeResult(map[string]any{
			"functions": []any{
				map[string]any{
					"name": "getUserByID", "path": "pkg/user.go",
					"source": "func getUserByID() {}", "language": "go",
				},
			},
		}),
	})

	// No limit param — should use defaultLimit
	req := makeMCPRequest("search_code_names", map[string]any{
		"repository": "myrepo",
		"query":      "getUserByID",
	})
	result, err := s.mcpHandleSearchCode(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result == nil || result.IsError {
		t.Error("expected successful result")
	}
}

func TestMCPHandleSearchCode_ValidationError(t *testing.T) {
	s := newTestServer(t)
	req := makeMCPRequest("search_code_names", map[string]any{
		"repository": "myrepo",
		"query":      "",
	})
	result, err := s.mcpHandleSearchCode(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected Go error: %v", err)
	}
	if !result.IsError {
		t.Error("expected error result for missing query")
	}
}

// --- mcpHandleSearchCodeContent ---

func TestMCPHandleSearchCodeContent_Success(t *testing.T) {
	s, _ := newTestServerWithResponses(t, []driver.Result{
		makeResult(map[string]any{
			"functions": []any{
				map[string]any{
					"name": "doStuff", "path": "pkg/stuff.go",
					"source": "func doStuff() { fmt.Println(\"hello\") }", "language": "go",
					"signature": "func doStuff()", "startingLine": float64(1), "endingLine": float64(3),
				},
			},
		}),
	})

	req := makeMCPRequest("search_code", map[string]any{
		"repository": "myrepo",
		"query":      "hello",
		"limit":      float64(5),
	})
	result, err := s.mcpHandleSearchCodeContent(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result == nil || result.IsError {
		t.Error("expected successful result")
	}
}

func TestMCPHandleSearchCodeContent_ValidationError(t *testing.T) {
	s := newTestServer(t)
	req := makeMCPRequest("search_code", map[string]any{
		"repository": "",
		"query":      "hello",
	})
	result, err := s.mcpHandleSearchCodeContent(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected Go error: %v", err)
	}
	if !result.IsError {
		t.Error("expected error result for missing repository")
	}
}

// --- mcpHandleGetCallers ---

func TestMCPHandleGetCallers_Success(t *testing.T) {
	s, _ := newTestServerWithResponses(t, []driver.Result{
		makeResult(map[string]any{
			"functions": []any{
				map[string]any{
					"name": "callerFunc", "path": "pkg/a.go",
					"signature": "func callerFunc()", "language": "go",
				},
			},
		}),
	})

	req := makeMCPRequest("get_callers", map[string]any{
		"repository": "myrepo",
		"name":       "myFunc",
		"depth":      float64(1),
	})
	result, err := s.mcpHandleGetCallers(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result == nil || result.IsError {
		t.Error("expected successful result")
	}
}

func TestMCPHandleGetCallers_DefaultDepth(t *testing.T) {
	s, _ := newTestServerWithResponses(t, []driver.Result{
		makeResult(map[string]any{
			"functions": []any{},
		}),
	})

	// No depth param — should default to 1
	req := makeMCPRequest("get_callers", map[string]any{
		"repository": "myrepo",
		"name":       "myFunc",
	})
	result, err := s.mcpHandleGetCallers(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result == nil || result.IsError {
		t.Error("expected successful result")
	}
}

func TestMCPHandleGetCallers_ValidationError(t *testing.T) {
	s := newTestServer(t)
	req := makeMCPRequest("get_callers", map[string]any{
		"repository": "myrepo",
		"name":       "",
	})
	result, err := s.mcpHandleGetCallers(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected Go error: %v", err)
	}
	if !result.IsError {
		t.Error("expected error result for missing name")
	}
}

// --- mcpHandleGetCallees ---

func TestMCPHandleGetCallees_Success(t *testing.T) {
	s, _ := newTestServerWithResponses(t, []driver.Result{
		makeResult(map[string]any{
			"functions": []any{
				map[string]any{
					"name": "calleeFunc", "path": "pkg/b.go",
					"signature": "func calleeFunc()", "language": "go",
				},
			},
		}),
	})

	req := makeMCPRequest("get_callees", map[string]any{
		"repository": "myrepo",
		"name":       "myFunc",
		"depth":      float64(2),
	})
	result, err := s.mcpHandleGetCallees(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result == nil || result.IsError {
		t.Error("expected successful result")
	}
}

func TestMCPHandleGetCallees_ValidationError(t *testing.T) {
	s := newTestServer(t)
	req := makeMCPRequest("get_callees", map[string]any{
		"repository": "",
		"name":       "myFunc",
	})
	result, err := s.mcpHandleGetCallees(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected Go error: %v", err)
	}
	if !result.IsError {
		t.Error("expected error result for missing repository")
	}
}

// --- mcpHandleGetClassHierarchy ---

func TestMCPHandleGetClassHierarchy_Success(t *testing.T) {
	s, _ := newTestServerWithResponses(t, []driver.Result{
		makeResult(map[string]any{
			"classs": []any{
				map[string]any{
					"name": "BaseClass", "path": "pkg/base.go",
					"signature": "type BaseClass struct{}", "language": "go",
				},
			},
		}),
	})

	req := makeMCPRequest("get_class_hierarchy", map[string]any{
		"repository": "myrepo",
		"name":       "MyClass",
		"direction":  "up",
		"depth":      float64(2),
	})
	result, err := s.mcpHandleGetClassHierarchy(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result == nil || result.IsError {
		t.Error("expected successful result")
	}
}

func TestMCPHandleGetClassHierarchy_DefaultDirection(t *testing.T) {
	s, _ := newTestServerWithResponses(t, []driver.Result{
		makeResult(map[string]any{
			"classs": []any{},
		}),
	})

	// No direction or depth — should default to "both" and 1
	req := makeMCPRequest("get_class_hierarchy", map[string]any{
		"repository": "myrepo",
		"name":       "MyClass",
	})
	result, err := s.mcpHandleGetClassHierarchy(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result == nil || result.IsError {
		t.Error("expected successful result")
	}
}

func TestMCPHandleGetClassHierarchy_ValidationError(t *testing.T) {
	s := newTestServer(t)
	req := makeMCPRequest("get_class_hierarchy", map[string]any{
		"repository": "myrepo",
		"name":       "",
	})
	result, err := s.mcpHandleGetClassHierarchy(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected Go error: %v", err)
	}
	if !result.IsError {
		t.Error("expected error result for missing name")
	}
}

// --- mcpHandleGetDependencies ---

func TestMCPHandleGetDependencies_Success(t *testing.T) {
	s, _ := newTestServerWithResponses(t, []driver.Result{
		makeResult(map[string]any{
			"modules": []any{
				map[string]any{
					"name": "dep-module", "path": "pkg/dep",
					"language": "go",
				},
			},
		}),
	})

	req := makeMCPRequest("get_dependencies", map[string]any{
		"repository": "myrepo",
		"name":       "myModule",
		"depth":      float64(1),
	})
	result, err := s.mcpHandleGetDependencies(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result == nil || result.IsError {
		t.Error("expected successful result")
	}
}

func TestMCPHandleGetDependencies_ValidationError(t *testing.T) {
	s := newTestServer(t)
	req := makeMCPRequest("get_dependencies", map[string]any{
		"repository": "myrepo",
		"name":       "",
	})
	result, err := s.mcpHandleGetDependencies(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected Go error: %v", err)
	}
	if !result.IsError {
		t.Error("expected error result for missing name")
	}
}

// --- mcpHandleGetReferences ---

func TestMCPHandleGetReferences_Success(t *testing.T) {
	s, _ := newTestServerWithResponses(t, []driver.Result{
		// First query: find the symbol (function match)
		makeResult(map[string]any{
			"functions": []any{
				map[string]any{
					"name": "myFunc", "path": "pkg/a.go",
					"signature": "func myFunc()", "language": "go",
				},
			},
		}),
		// Second query: callers
		makeResult(map[string]any{
			"functions": []any{},
		}),
	})

	req := makeMCPRequest("get_references", map[string]any{
		"repository": "myrepo",
		"name":       "myFunc",
	})
	result, err := s.mcpHandleGetReferences(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result == nil || result.IsError {
		t.Error("expected successful result")
	}
}

func TestMCPHandleGetReferences_ValidationError(t *testing.T) {
	s := newTestServer(t)
	req := makeMCPRequest("get_references", map[string]any{
		"repository": "",
		"name":       "myFunc",
	})
	result, err := s.mcpHandleGetReferences(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected Go error: %v", err)
	}
	if !result.IsError {
		t.Error("expected error result for missing repository")
	}
}

// --- mcpHandleGetRepoMap ---

func TestMCPHandleGetRepoMap_Success(t *testing.T) {
	s, _ := newTestServerWithResponses(t, []driver.Result{
		makeResult(map[string]any{
			"files": []any{
				map[string]any{"name": "main.go", "path": "cmd/main.go", "language": "go", "lineCount": float64(50)},
			},
		}),
		// functions for symbol count
		makeResult(map[string]any{
			"functions": []any{
				map[string]any{"name": "main"},
			},
		}),
		// classes for symbol count
		makeResult(map[string]any{
			"classs": []any{},
		}),
	})

	req := makeMCPRequest("get_repo_map", map[string]any{
		"repository": "myrepo",
	})
	result, err := s.mcpHandleGetRepoMap(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result == nil || result.IsError {
		t.Error("expected successful result")
	}
}

func TestMCPHandleGetRepoMap_ValidationError(t *testing.T) {
	s := newTestServer(t)
	req := makeMCPRequest("get_repo_map", map[string]any{
		"repository": "",
	})
	result, err := s.mcpHandleGetRepoMap(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected Go error: %v", err)
	}
	if !result.IsError {
		t.Error("expected error result for missing repository")
	}
}

// --- mcpHandleGetFileOverview ---

func TestMCPHandleGetFileOverview_Success(t *testing.T) {
	s, _ := newTestServerWithResponses(t, []driver.Result{
		makeResult(map[string]any{
			"functions": []any{
				map[string]any{
					"name": "main", "path": "cmd/main.go",
					"signature": "func main()", "language": "go",
					"startingLine": float64(1), "endingLine": float64(5),
				},
			},
		}),
		makeResult(map[string]any{
			"classs": []any{},
		}),
	})

	req := makeMCPRequest("get_file_overview", map[string]any{
		"repository": "myrepo",
		"path":       "cmd/main.go",
	})
	result, err := s.mcpHandleGetFileOverview(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result == nil || result.IsError {
		t.Error("expected successful result")
	}
}

func TestMCPHandleGetFileOverview_ValidationError(t *testing.T) {
	s := newTestServer(t)
	req := makeMCPRequest("get_file_overview", map[string]any{
		"repository": "myrepo",
		"path":       "",
	})
	result, err := s.mcpHandleGetFileOverview(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected Go error: %v", err)
	}
	if !result.IsError {
		t.Error("expected error result for missing path")
	}
}

// --- mcpHandleGetSymbolContext ---

func TestMCPHandleGetSymbolContext_Success(t *testing.T) {
	s, _ := newTestServerWithResponses(t, []driver.Result{
		// Find the symbol
		makeResult(map[string]any{
			"functions": []any{
				map[string]any{
					"name": "myFunc", "path": "pkg/a.go",
					"source": "func myFunc() {}", "signature": "func myFunc()",
					"language": "go", "visibility": "public",
					"startingLine": float64(1), "endingLine": float64(3),
				},
			},
		}),
		// Callers
		makeResult(map[string]any{
			"functions": []any{},
		}),
		// Callees
		makeResult(map[string]any{
			"functions": []any{},
		}),
		// Siblings (functions in same file)
		makeResult(map[string]any{
			"functions": []any{},
		}),
		// Siblings (classes in same file)
		makeResult(map[string]any{
			"classs": []any{},
		}),
	})

	req := makeMCPRequest("get_symbol_context", map[string]any{
		"repository": "myrepo",
		"name":       "myFunc",
	})
	result, err := s.mcpHandleGetSymbolContext(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result == nil || result.IsError {
		t.Error("expected successful result")
	}
}

func TestMCPHandleGetSymbolContext_ValidationError(t *testing.T) {
	s := newTestServer(t)
	req := makeMCPRequest("get_symbol_context", map[string]any{
		"repository": "",
		"name":       "myFunc",
	})
	result, err := s.mcpHandleGetSymbolContext(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected Go error: %v", err)
	}
	if !result.IsError {
		t.Error("expected error result for missing repository")
	}
}

// --- mcpHandleReadSource ---

func TestMCPHandleReadSource_Success(t *testing.T) {
	s, _ := newTestServerWithResponses(t, []driver.Result{
		makeResult(map[string]any{
			"functions": []any{
				map[string]any{
					"name": "funcA", "path": "pkg/a.go",
					"source": "func funcA() {}", "language": "go",
					"startingLine": float64(1), "endingLine": float64(3),
				},
			},
		}),
		makeResult(map[string]any{
			"functions": []any{
				map[string]any{
					"name": "funcB", "path": "pkg/b.go",
					"source": "func funcB() {}", "language": "go",
					"startingLine": float64(10), "endingLine": float64(12),
				},
			},
		}),
	})

	req := makeMCPRequest("read_source", map[string]any{
		"repository": "myrepo",
		"names":      "funcA, funcB",
	})
	result, err := s.mcpHandleReadSource(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result == nil || result.IsError {
		t.Error("expected successful result")
	}
	// Verify the JSON contains both results
	if len(result.Content) > 0 {
		text := result.Content[0].(mcplib.TextContent).Text
		var resp ReadSourceResponse
		if err := json.Unmarshal([]byte(text), &resp); err == nil {
			if resp.Total != 2 {
				t.Errorf("expected 2 results, got %d", resp.Total)
			}
		}
	}
}

func TestMCPHandleReadSource_CommaParsing(t *testing.T) {
	s, _ := newTestServerWithResponses(t, []driver.Result{
		makeResult(map[string]any{
			"functions": []any{
				map[string]any{
					"name": "funcA", "path": "pkg/a.go",
					"source": "func funcA() {}", "language": "go",
					"startingLine": float64(1), "endingLine": float64(3),
				},
			},
		}),
	})

	// names with extra spaces and trailing comma
	req := makeMCPRequest("read_source", map[string]any{
		"repository": "myrepo",
		"names":      " funcA , , ",
	})
	result, err := s.mcpHandleReadSource(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result == nil || result.IsError {
		t.Error("expected successful result; empty names after trim should be filtered")
	}
}

func TestMCPHandleReadSource_ValidationError(t *testing.T) {
	s := newTestServer(t)
	req := makeMCPRequest("read_source", map[string]any{
		"repository": "",
		"names":      "funcA",
	})
	result, err := s.mcpHandleReadSource(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected Go error: %v", err)
	}
	if !result.IsError {
		t.Error("expected error result for missing repository")
	}
}

// --- mcpHandleIngestRepository ---

func TestMCPHandleIngestRepository_ValidationError(t *testing.T) {
	s := newTestServer(t)
	req := makeMCPRequest("ingest_repository", map[string]any{
		"repository_path": "",
	})
	result, err := s.mcpHandleIngestRepository(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected Go error: %v", err)
	}
	if !result.IsError {
		t.Error("expected error result for missing repository_path")
	}
}

// --- mcpHandleDeleteRepository ---

func TestMCPHandleDeleteRepository_Success(t *testing.T) {
	s, _ := newTestServerWithResponses(t, []driver.Result{})

	req := makeMCPRequest("delete_repository", map[string]any{
		"repository": "myrepo",
	})
	result, err := s.mcpHandleDeleteRepository(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result == nil || result.IsError {
		t.Error("expected successful result")
	}
}

func TestMCPHandleDeleteRepository_ValidationError(t *testing.T) {
	s := newTestServer(t)
	req := makeMCPRequest("delete_repository", map[string]any{
		"repository": "",
	})
	result, err := s.mcpHandleDeleteRepository(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected Go error: %v", err)
	}
	if !result.IsError {
		t.Error("expected error result for missing repository")
	}
}

// --- mcpHandleGetRepositoryStats ---

func TestMCPHandleGetRepositoryStats_Success(t *testing.T) {
	s, _ := newTestServerWithResponses(t, []driver.Result{
		makeResult(map[string]any{
			"files":              []any{},
			"functions":          []any{},
			"classs":             []any{},
			"modules":            []any{},
			"externalReferences": []any{},
		}),
		makeResult(map[string]any{"functions": []any{}}),
		makeResult(map[string]any{"classs": []any{}}),
		makeResult(map[string]any{"modules": []any{}}),
		makeResult(map[string]any{"externalReferences": []any{}}),
	})

	req := makeMCPRequest("get_repository_stats", map[string]any{
		"repository": "myrepo",
	})
	result, err := s.mcpHandleGetRepositoryStats(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result == nil || result.IsError {
		t.Error("expected successful result")
	}
}

func TestMCPHandleGetRepositoryStats_ValidationError(t *testing.T) {
	s := newTestServer(t)
	req := makeMCPRequest("get_repository_stats", map[string]any{
		"repository": "",
	})
	result, err := s.mcpHandleGetRepositoryStats(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected Go error: %v", err)
	}
	if !result.IsError {
		t.Error("expected error result for missing repository")
	}
}

// --- mcpHandleFindDeadCode ---

func TestMCPHandleFindDeadCode_Success(t *testing.T) {
	s, _ := newTestServerWithResponses(t, []driver.Result{
		// Dead functions query
		makeResult(map[string]any{
			"functions": []any{
				map[string]any{
					"name": "unusedFunc", "path": "pkg/unused.go",
					"signature": "func unusedFunc()", "startingLine": float64(1), "endingLine": float64(3),
				},
			},
		}),
		// Dead classes query
		makeResult(map[string]any{
			"classs": []any{},
		}),
		// Dead modules query
		makeResult(map[string]any{
			"modules": []any{},
		}),
	})

	req := makeMCPRequest("find_dead_code", map[string]any{
		"repository":       "myrepo",
		"exclude_decorated": true,
		"exclude_patterns": "test*,mock*",
		"limit":            float64(20),
	})
	result, err := s.mcpHandleFindDeadCode(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result == nil || result.IsError {
		t.Error("expected successful result")
	}
}

func TestMCPHandleFindDeadCode_DefaultParams(t *testing.T) {
	s, _ := newTestServerWithResponses(t, []driver.Result{
		makeResult(map[string]any{"functions": []any{}}),
		makeResult(map[string]any{"classs": []any{}}),
		makeResult(map[string]any{"modules": []any{}}),
	})

	// No optional params — should use defaults (exclude_decorated=false, limit=50)
	req := makeMCPRequest("find_dead_code", map[string]any{
		"repository": "myrepo",
	})
	result, err := s.mcpHandleFindDeadCode(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result == nil || result.IsError {
		t.Error("expected successful result")
	}
}

func TestMCPHandleFindDeadCode_ValidationError(t *testing.T) {
	s := newTestServer(t)
	req := makeMCPRequest("find_dead_code", map[string]any{
		"repository": "",
	})
	result, err := s.mcpHandleFindDeadCode(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected Go error: %v", err)
	}
	if !result.IsError {
		t.Error("expected error result for missing repository")
	}
}

// --- mcpHandleCalculateCyclomaticComplexity ---

func TestMCPHandleCalculateCyclomaticComplexity_Success(t *testing.T) {
	s, _ := newTestServerWithResponses(t, []driver.Result{
		makeResult(map[string]any{
			"functions": []any{
				map[string]any{
					"name": "complexFunc", "path": "pkg/complex.go",
					"signature": "func complexFunc()", "cyclomaticComplexity": float64(8),
					"startingLine": float64(1), "endingLine": float64(30),
				},
			},
		}),
	})

	req := makeMCPRequest("calculate_cyclomatic_complexity", map[string]any{
		"repository": "myrepo",
		"name":       "complexFunc",
	})
	result, err := s.mcpHandleCalculateCyclomaticComplexity(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result == nil || result.IsError {
		t.Error("expected successful result")
	}
}

func TestMCPHandleCalculateCyclomaticComplexity_ValidationError(t *testing.T) {
	s := newTestServer(t)
	req := makeMCPRequest("calculate_cyclomatic_complexity", map[string]any{
		"repository": "myrepo",
		"name":       "",
	})
	result, err := s.mcpHandleCalculateCyclomaticComplexity(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected Go error: %v", err)
	}
	if !result.IsError {
		t.Error("expected error result for missing name")
	}
}

// --- mcpHandleFindMostComplexFunctions ---

func TestMCPHandleFindMostComplexFunctions_Success(t *testing.T) {
	s, _ := newTestServerWithResponses(t, []driver.Result{
		makeResult(map[string]any{
			"functions": []any{
				map[string]any{
					"name": "bigFunc", "path": "pkg/big.go",
					"signature": "func bigFunc()", "cyclomaticComplexity": float64(15),
					"startingLine": float64(1), "endingLine": float64(100),
				},
			},
		}),
	})

	req := makeMCPRequest("find_most_complex_functions", map[string]any{
		"repository":     "myrepo",
		"min_complexity": float64(10),
		"limit":          float64(5),
	})
	result, err := s.mcpHandleFindMostComplexFunctions(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result == nil || result.IsError {
		t.Error("expected successful result")
	}
}

func TestMCPHandleFindMostComplexFunctions_DefaultParams(t *testing.T) {
	s, _ := newTestServerWithResponses(t, []driver.Result{
		makeResult(map[string]any{
			"functions": []any{},
		}),
	})

	// No optional params — should default to min_complexity=5, limit=10
	req := makeMCPRequest("find_most_complex_functions", map[string]any{
		"repository": "myrepo",
	})
	result, err := s.mcpHandleFindMostComplexFunctions(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result == nil || result.IsError {
		t.Error("expected successful result")
	}
}

func TestMCPHandleFindMostComplexFunctions_ValidationError(t *testing.T) {
	s := newTestServer(t)
	req := makeMCPRequest("find_most_complex_functions", map[string]any{
		"repository": "",
	})
	result, err := s.mcpHandleFindMostComplexFunctions(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected Go error: %v", err)
	}
	if !result.IsError {
		t.Error("expected error result for missing repository")
	}
}

// --- Verify JSON output structure ---

func TestMCPHandleFindFunction_JSONOutput(t *testing.T) {
	s, _ := newTestServerWithResponses(t, []driver.Result{
		makeResult(map[string]any{
			"functions": []any{
				map[string]any{
					"name": "myFunc", "path": "pkg/a.go",
					"source": "func myFunc() {}", "signature": "func myFunc()",
					"language": "go", "visibility": "public",
					"startingLine": float64(1), "endingLine": float64(3),
				},
			},
		}),
	})

	req := makeMCPRequest("find_function", map[string]any{
		"repository": "myrepo",
		"name":       "myFunc",
	})
	result, err := s.mcpHandleFindFunction(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result.Content) == 0 {
		t.Fatal("expected content in result")
	}
	text := result.Content[0].(mcplib.TextContent).Text
	if !strings.Contains(text, "myFunc") {
		t.Error("JSON output should contain function name")
	}
	// Verify it's valid JSON
	var raw map[string]any
	if err := json.Unmarshal([]byte(text), &raw); err != nil {
		t.Errorf("result is not valid JSON: %v", err)
	}
}
