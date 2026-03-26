package mcp

import (
	"context"
	"testing"

	"github.com/tab58/go-ormql/pkg/driver"
)

// --- Task 6: Migrate MCP tools to ForRepo ---

// TestHandleFindFunction_UsesForRepo verifies that handleFindFunction
// obtains a client via ForRepo (the repo param) rather than requireClient/db.Client().
// Expected result: Function query succeeds with a recording driver
// (proving ForRepo was used to get the repo-scoped client).
func TestHandleFindFunction_UsesForRepo(t *testing.T) {
	svc, _ := newTestServiceWithResponses(t, []driver.Result{
		makeResult(map[string]any{
			"functions": []any{
				map[string]any{
					"name":         "hello",
					"path":         "main.go",
					"source":       "func hello() {}",
					"language":     "go",
					"visibility":   "public",
					"startingLine": float64(1),
					"endingLine":   float64(1),
				},
			},
		}),
	})

	resp, err := svc.HandleFindFunction(context.Background(), "my-repo", "hello")
	if err != nil {
		t.Fatalf("handleFindFunction returned error: %v", err)
	}
	if resp == nil {
		t.Fatal("handleFindFunction returned nil response")
	}
	if len(resp.Results) == 0 {
		t.Error("handleFindFunction returned 0 results, expected at least 1")
	}
}

// TestHandleFindFile_UsesForRepo verifies that handleFindFile obtains
// a client via ForRepo using the repository parameter.
// Expected result: Find file succeeds with a response driver.
func TestHandleFindFile_UsesForRepo(t *testing.T) {
	svc, _ := newTestServiceWithResponses(t, []driver.Result{
		makeResult(map[string]any{
			"files": []any{
				map[string]any{
					"path":     "src/main.go",
					"language": "go",
				},
			},
		}),
	})

	resp, err := svc.HandleFindFile(context.Background(), "my-repo", "*.go")
	if err != nil {
		t.Fatalf("handleFindFile returned error: %v", err)
	}
	if resp == nil {
		t.Fatal("nil response")
	}
}

// TestRequireClient_Removed verifies that requireClient() is replaced
// with a ForRepo-based approach. The requireRepoClient or equivalent
// should accept a repo name parameter.
// This is validated by the source inspection test in isolation_verify_test.go
// which checks that tools.go does not contain ".Client()".
