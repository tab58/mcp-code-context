package repl

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/tab58/code-context/internal/indexer"
	"github.com/tab58/go-ormql/pkg/driver"
)

// === Task 7: Command handlers ===

// TestHandleHelp_PrintsAllCommands verifies that handleHelp prints all 5 commands.
// Expected result: output contains ingest, status, list, help, quit.
func TestHandleHelp_PrintsAllCommands(t *testing.T) {
	var out bytes.Buffer
	r := New(Pipeline{}, StatusInfo{}, WithOutput(&out))
	if r == nil {
		t.Fatal("New returned nil")
	}

	r.handleHelp()

	output := out.String()
	commands := []string{"ingest", "status", "list", "help", "quit"}
	for _, cmd := range commands {
		if !strings.Contains(output, cmd) {
			t.Errorf("handleHelp output missing %q command", cmd)
		}
	}
}

// TestHandleStatus_PrintsConfig verifies that handleStatus prints config fields.
// Expected result: output contains config values.
func TestHandleStatus_PrintsConfig(t *testing.T) {
	var out bytes.Buffer
	status := StatusInfo{
		FalkorDBHost: "db.example.com",
		FalkorDBPort: "6379",
		MCPPort:      "9090",
	}
	r := New(Pipeline{}, status, WithOutput(&out))
	if r == nil {
		t.Fatal("New returned nil")
	}

	r.handleStatus()

	output := out.String()
	expected := []string{"db.example.com", "6379", "9090"}
	for _, val := range expected {
		if !strings.Contains(output, val) {
			t.Errorf("handleStatus output missing config value %q", val)
		}
	}
}

// TestHandleIngest_NoPath verifies that handleIngest with no args prints usage error.
// Expected result: error returned or output contains "usage".
func TestHandleIngest_NoPath(t *testing.T) {
	var out bytes.Buffer
	r := New(Pipeline{}, StatusInfo{}, WithOutput(&out))
	if r == nil {
		t.Fatal("New returned nil")
	}

	err := r.handleIngest(context.Background(), nil)
	// Should return an error or print usage
	if err == nil && !strings.Contains(out.String(), "usage") {
		t.Error("handleIngest with no path should return error or print usage")
	}
}

// TestHandleIngest_NonExistentPath verifies that handleIngest with invalid path fails.
// Expected result: error returned containing path info.
func TestHandleIngest_NonExistentPath(t *testing.T) {
	var out bytes.Buffer
	r := New(Pipeline{}, StatusInfo{}, WithOutput(&out))
	if r == nil {
		t.Fatal("New returned nil")
	}

	err := r.handleIngest(context.Background(), []string{"/nonexistent/path/xyz"})
	if err == nil {
		t.Error("handleIngest with non-existent path should return error")
	}
}

// TestHandleList_NilDB verifies that handleList handles nil DB gracefully.
// Expected result: returns error or prints "not connected" message.
func TestHandleList_NilDB(t *testing.T) {
	var out bytes.Buffer
	r := New(Pipeline{}, StatusInfo{}, WithOutput(&out))
	if r == nil {
		t.Fatal("New returned nil")
	}

	err := r.handleList(context.Background())
	// Should handle nil DB gracefully
	if err == nil && out.Len() == 0 {
		t.Error("handleList with nil DB should return error or print message")
	}
}

// TestRun_HelpCommand verifies that typing "help" in the REPL prints help text.
// Expected result: output contains command listings.
func TestRun_HelpCommand(t *testing.T) {
	in := strings.NewReader("help\nquit\n")
	var out bytes.Buffer
	r := New(Pipeline{}, StatusInfo{}, WithInput(in), WithOutput(&out))
	if r == nil {
		t.Fatal("New returned nil")
	}

	err := r.Run(context.Background())
	if err != nil {
		t.Errorf("Run returned error: %v", err)
	}
	output := out.String()
	if !strings.Contains(output, "ingest") {
		t.Error("help command should list 'ingest' command")
	}
}

// TestRun_StatusCommand verifies that typing "status" in the REPL prints config.
// Expected result: output contains status info.
func TestRun_StatusCommand(t *testing.T) {
	in := strings.NewReader("status\nquit\n")
	var out bytes.Buffer
	status := StatusInfo{
		FalkorDBHost: "testhost",
		MCPPort:      "8080",
	}
	r := New(Pipeline{}, status, WithInput(in), WithOutput(&out))
	if r == nil {
		t.Fatal("New returned nil")
	}

	err := r.Run(context.Background())
	if err != nil {
		t.Errorf("Run returned error: %v", err)
	}
	output := out.String()
	if !strings.Contains(output, "testhost") {
		t.Error("status command should print FalkorDB host")
	}
}

// TestRun_IngestNoArgs verifies that "ingest" with no path shows usage.
// Expected result: output contains usage information.
func TestRun_IngestNoArgs(t *testing.T) {
	in := strings.NewReader("ingest\nquit\n")
	var out bytes.Buffer
	r := New(Pipeline{}, StatusInfo{}, WithInput(in), WithOutput(&out))
	if r == nil {
		t.Fatal("New returned nil")
	}

	err := r.Run(context.Background())
	if err != nil {
		t.Errorf("Run returned error: %v", err)
	}
	output := out.String()
	if !strings.Contains(strings.ToLower(output), "usage") && !strings.Contains(strings.ToLower(output), "ingest <path>") {
		t.Error("ingest with no path should show usage or error")
	}
}

// TestHandleIngest_PathIsFile verifies that ingest with a file path returns error.
func TestHandleIngest_PathIsFile(t *testing.T) {
	tmp := t.TempDir()
	f := filepath.Join(tmp, "file.txt")
	if err := os.WriteFile(f, []byte("hello"), 0o644); err != nil {
		t.Fatal(err)
	}

	var out bytes.Buffer
	r := New(Pipeline{}, StatusInfo{}, WithOutput(&out))
	err := r.handleIngest(context.Background(), []string{f})
	if err == nil {
		t.Error("handleIngest with file path should return error")
	}
	if err != nil && !strings.Contains(err.Error(), "not a directory") {
		t.Errorf("expected 'not a directory' error, got: %v", err)
	}
}

// TestHandleIngest_SuccessfulPipeline verifies the full ingest pipeline runs
// with a real Indexer (nil DB) and prints the done summary.
func TestHandleIngest_SuccessfulPipeline(t *testing.T) {
	tmp := t.TempDir()
	if err := os.WriteFile(filepath.Join(tmp, "main.go"), []byte("package main\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	pipeline := Pipeline{
		Indexer: indexer.NewIndexer(nil),
	}

	var out bytes.Buffer
	r := New(pipeline, StatusInfo{}, WithOutput(&out))
	err := r.handleIngest(context.Background(), []string{tmp})
	if err != nil {
		t.Fatalf("handleIngest failed: %v", err)
	}
	output := out.String()
	if !strings.Contains(output, "done:") {
		t.Errorf("expected done summary, got: %s", output)
	}
}

// TestHandleIngest_ProgressOutput verifies that ingest prints progress lines.
func TestHandleIngest_ProgressOutput(t *testing.T) {
	tmp := t.TempDir()
	if err := os.MkdirAll(filepath.Join(tmp, "src"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(tmp, "src", "util.go"), []byte("package src\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	pipeline := Pipeline{
		Indexer: indexer.NewIndexer(nil),
	}

	var out bytes.Buffer
	r := New(pipeline, StatusInfo{}, WithOutput(&out))
	err := r.handleIngest(context.Background(), []string{tmp})
	if err != nil {
		t.Fatalf("handleIngest failed: %v", err)
	}
	output := out.String()
	if !strings.Contains(output, "[indexing]") {
		t.Errorf("expected progress output with [indexing] stage, got: %s", output)
	}
}

// TestHandleList_NilClient verifies that handleList handles nil client gracefully.
func TestHandleList_NilClient(t *testing.T) {
	var out bytes.Buffer
	r := New(Pipeline{}, StatusInfo{}, WithOutput(&out))

	err := r.handleList(context.Background())
	if err == nil {
		t.Error("handleList with nil DB should return error")
	}
	if err != nil && !strings.Contains(err.Error(), "not connected") {
		t.Errorf("expected 'not connected' error, got: %v", err)
	}
}

// TestHandleList_EmptyResult verifies handleList with no repositories.
func TestHandleList_EmptyResult(t *testing.T) {
	db := newTestREPLDB(t)
	pipeline := Pipeline{DB: db}

	var out bytes.Buffer
	r := New(pipeline, StatusInfo{}, WithOutput(&out))
	err := r.handleList(context.Background())
	if err != nil {
		t.Fatalf("handleList with empty result failed: %v", err)
	}
	if !strings.Contains(out.String(), "no repositories found") {
		t.Errorf("expected 'no repositories found', got: %s", out.String())
	}
}

// TestHandleList_WithRepositories verifies handleList prints repository names.
func TestHandleList_WithRepositories(t *testing.T) {
	// Provide a response driver that returns repository data for ListRepos
	db := newTestREPLDBWithResponses(t, []driver.Result{
		{
			Records: []driver.Record{
				{Values: map[string]any{
					"data": map[string]any{
						"repositorys": []any{
							map[string]any{"name": "my-repo"},
							map[string]any{"name": "other-repo"},
						},
					},
				}},
			},
		},
	})

	pipeline := Pipeline{DB: db}

	var out bytes.Buffer
	r := New(pipeline, StatusInfo{}, WithOutput(&out))
	err := r.handleList(context.Background())
	if err != nil {
		t.Fatalf("handleList failed: %v", err)
	}
	output := out.String()
	if !strings.Contains(output, "my-repo") {
		t.Errorf("expected 'my-repo' in output, got: %s", output)
	}
	if !strings.Contains(output, "other-repo") {
		t.Errorf("expected 'other-repo' in output, got: %s", output)
	}
}


// === Delete command tests ===

// TestHandleDelete_NoArgs verifies that handleDelete with no args returns usage error.
func TestHandleDelete_NoArgs(t *testing.T) {
	var out bytes.Buffer
	r := New(Pipeline{}, StatusInfo{}, WithOutput(&out))
	err := r.handleDelete(context.Background(), nil)
	if err == nil {
		t.Error("handleDelete with no args should return error")
	}
	if err != nil && !strings.Contains(err.Error(), "usage") {
		t.Errorf("expected usage error, got: %v", err)
	}
}

// TestHandleDelete_NilDB verifies that handleDelete with nil DB returns error.
func TestHandleDelete_NilDB(t *testing.T) {
	var out bytes.Buffer
	r := New(Pipeline{}, StatusInfo{}, WithOutput(&out))
	err := r.handleDelete(context.Background(), []string{"myrepo"})
	if err == nil {
		t.Error("handleDelete with nil DB should return error")
	}
	if err != nil && !strings.Contains(err.Error(), "not connected") {
		t.Errorf("expected 'not connected' error, got: %v", err)
	}
}

// TestHandleDelete_Success verifies that handleDelete succeeds and prints confirmation.
func TestHandleDelete_Success(t *testing.T) {
	db := newTestREPLDB(t)
	pipeline := Pipeline{DB: db}

	var out bytes.Buffer
	r := New(pipeline, StatusInfo{}, WithOutput(&out))
	err := r.handleDelete(context.Background(), []string{"myrepo"})
	if err != nil {
		t.Fatalf("handleDelete failed: %v", err)
	}
	if !strings.Contains(out.String(), "deleted") {
		t.Errorf("expected 'deleted' in output, got: %s", out.String())
	}
	if !strings.Contains(out.String(), "myrepo") {
		t.Errorf("expected repo name in output, got: %s", out.String())
	}
}

// TestRun_DeleteCommand verifies the delete command through the Run loop.
func TestRun_DeleteCommand(t *testing.T) {
	db := newTestREPLDB(t)
	pipeline := Pipeline{DB: db}

	in := strings.NewReader("delete myrepo\nquit\n")
	var out bytes.Buffer
	r := New(pipeline, StatusInfo{}, WithInput(in), WithOutput(&out))

	if err := r.Run(context.Background()); err != nil {
		t.Fatalf("Run returned error: %v", err)
	}
	if !strings.Contains(out.String(), "deleted") {
		t.Errorf("expected 'deleted' in output, got: %s", out.String())
	}
}

// TestRun_DeleteNoArgs verifies delete with no args shows error through Run.
func TestRun_DeleteNoArgs(t *testing.T) {
	in := strings.NewReader("delete\nquit\n")
	var out bytes.Buffer
	r := New(Pipeline{}, StatusInfo{}, WithInput(in), WithOutput(&out))

	if err := r.Run(context.Background()); err != nil {
		t.Fatalf("Run returned error: %v", err)
	}
	if !strings.Contains(out.String(), "error:") {
		t.Errorf("expected error output for delete with no args, got: %s", out.String())
	}
}

// TestHandleHelp_IncludesDelete verifies help output includes the delete command.
func TestHandleHelp_IncludesDelete(t *testing.T) {
	var out bytes.Buffer
	r := New(Pipeline{}, StatusInfo{}, WithOutput(&out))
	r.handleHelp()
	if !strings.Contains(out.String(), "delete") {
		t.Error("handleHelp output should include 'delete' command")
	}
}

// TestRun_ListWithNilDB verifies list command through Run with nil DB.
func TestRun_ListWithNilDB(t *testing.T) {
	in := strings.NewReader("list\nquit\n")
	var out bytes.Buffer
	r := New(Pipeline{}, StatusInfo{}, WithInput(in), WithOutput(&out))

	if err := r.Run(context.Background()); err != nil {
		t.Fatalf("Run returned error: %v", err)
	}
	output := out.String()
	if !strings.Contains(output, "error:") {
		t.Errorf("expected error output for list with nil DB, got: %s", output)
	}
}

// TestRun_IngestNonExistentPath verifies ingest with bad path through Run.
func TestRun_IngestNonExistentPath(t *testing.T) {
	in := strings.NewReader("ingest /no/such/path\nquit\n")
	var out bytes.Buffer
	r := New(Pipeline{}, StatusInfo{}, WithInput(in), WithOutput(&out))

	if err := r.Run(context.Background()); err != nil {
		t.Fatalf("Run returned error: %v", err)
	}
	output := out.String()
	if !strings.Contains(output, "error:") {
		t.Errorf("expected error output for bad path, got: %s", output)
	}
}
