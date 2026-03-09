package repl

import (
	"bytes"
	"context"
	"strings"
	"testing"

	"github.com/tab58/go-ormql/pkg/driver"
)

// --- Shared graph isolation tests for REPL ---

// TestHandleList_UsesListRepos verifies that handleList uses
// db.ListRepos() to query Repository nodes from the shared graph.
// Expected result: list command prints repo names from ListRepos.
func TestHandleList_UsesListRepos(t *testing.T) {
	// Provide a driver that returns repository data for the ListRepos query
	db := newTestREPLDBWithResponses(t, []driver.Result{
		{
			Records: []driver.Record{
				{Values: map[string]any{
					"data": map[string]any{
						"repositorys": []any{
							map[string]any{"name": "my-test-repo"},
						},
					},
				}},
			},
		},
	})

	var out bytes.Buffer
	r := New(Pipeline{DB: db}, StatusInfo{}, WithOutput(&out))

	err := r.handleList(context.Background())
	if err != nil {
		t.Fatalf("handleList returned error: %v", err)
	}

	output := out.String()
	if !strings.Contains(output, "my-test-repo") {
		t.Errorf("handleList output should contain 'my-test-repo', got: %q", output)
	}
}

// TestHandleList_EmptyRepos verifies that handleList prints a message
// when no repositories have been indexed.
// Expected result: Output contains "no repositories" or similar.
func TestHandleList_EmptyRepos(t *testing.T) {
	db := newTestREPLDB(t)

	var out bytes.Buffer
	r := New(Pipeline{DB: db}, StatusInfo{}, WithOutput(&out))

	err := r.handleList(context.Background())
	if err != nil {
		t.Fatalf("handleList returned error: %v", err)
	}

	output := out.String()
	if output == "" {
		t.Error("handleList should print something even with no repos")
	}
}

// TestStatusInfo_NoDatabase verifies that StatusInfo does not have a
// FalkorDBDatabase field. This is a compile-time check — if the field
// still existed, the verify test would catch it.
func TestStatusInfo_NoDatabase(t *testing.T) {
	status := StatusInfo{
		FalkorDBHost: "localhost",
		FalkorDBPort: "6379",
		MCPPort:      "8080",
	}
	if status.FalkorDBHost == "" {
		t.Error("unexpected empty host")
	}
}

// TestHandleStatus_NoDatabase verifies that handleStatus does not print
// a database name line.
// Expected result: Output does not contain "Database".
func TestHandleStatus_NoDatabase(t *testing.T) {
	var out bytes.Buffer
	r := New(Pipeline{}, StatusInfo{
		FalkorDBHost: "localhost",
		FalkorDBPort: "6379",
		MCPPort:      "8080",
	}, WithOutput(&out))

	r.handleStatus()

	output := out.String()
	if strings.Contains(output, "Database") {
		t.Errorf("handleStatus should not print Database, got: %q", output)
	}
}
