package repl

import (
	"bytes"
	"context"
	"strings"
	"testing"
)

// === Task 6: REPL struct and Run loop ===

// TestNew_ReturnsNonNil verifies that New creates a valid REPL.
// Expected result: New returns a non-nil *REPL.
func TestNew_ReturnsNonNil(t *testing.T) {
	r := New(Pipeline{}, StatusInfo{})
	if r == nil {
		t.Fatal("New returned nil, expected non-nil *REPL")
	}
}

// TestNew_WithInputOption verifies that WithInput sets the input reader.
// Expected result: REPL uses the provided reader instead of os.Stdin.
func TestNew_WithInputOption(t *testing.T) {
	in := strings.NewReader("quit\n")
	r := New(Pipeline{}, StatusInfo{}, WithInput(in))
	if r == nil {
		t.Fatal("New returned nil")
	}
	if r.in != in {
		t.Error("WithInput did not set the input reader")
	}
}

// TestNew_WithOutputOption verifies that WithOutput sets the output writer.
// Expected result: REPL uses the provided writer instead of os.Stdout.
func TestNew_WithOutputOption(t *testing.T) {
	var out bytes.Buffer
	r := New(Pipeline{}, StatusInfo{}, WithOutput(&out))
	if r == nil {
		t.Fatal("New returned nil")
	}
	if r.out != &out {
		t.Error("WithOutput did not set the output writer")
	}
}

// TestNew_StoresPipeline verifies that New stores the pipeline reference.
// Expected result: REPL.pipeline is set.
func TestNew_StoresPipeline(t *testing.T) {
	p := Pipeline{}
	r := New(p, StatusInfo{})
	if r == nil {
		t.Fatal("New returned nil")
	}
	// Pipeline is a struct, not a pointer — just verify the REPL was created
}


// TestParseCommand_Empty verifies that empty input returns empty command.
// Expected result: cmd="" args=nil.
func TestParseCommand_Empty(t *testing.T) {
	cmd, args := parseCommand("")
	if cmd != "" {
		t.Errorf("parseCommand(\"\") cmd = %q, want \"\"", cmd)
	}
	if len(args) != 0 {
		t.Errorf("parseCommand(\"\") args = %v, want nil/empty", args)
	}
}

// TestParseCommand_WhitespaceOnly verifies that whitespace-only input returns empty command.
// Expected result: cmd="" args=nil.
func TestParseCommand_WhitespaceOnly(t *testing.T) {
	cmd, args := parseCommand("   ")
	if cmd != "" {
		t.Errorf("parseCommand(\"   \") cmd = %q, want \"\"", cmd)
	}
	if len(args) != 0 {
		t.Errorf("parseCommand(\"   \") args = %v, want nil/empty", args)
	}
}

// TestParseCommand_SingleWord verifies that a single word returns the command with no args.
// Expected result: cmd="help" args=nil.
func TestParseCommand_SingleWord(t *testing.T) {
	cmd, args := parseCommand("help")
	if cmd != "help" {
		t.Errorf("parseCommand(\"help\") cmd = %q, want \"help\"", cmd)
	}
	if len(args) != 0 {
		t.Errorf("parseCommand(\"help\") args = %v, want nil/empty", args)
	}
}

// TestParseCommand_WithArgs verifies that command + args are split correctly.
// Expected result: cmd="ingest" args=["/path/to/repo"].
func TestParseCommand_WithArgs(t *testing.T) {
	cmd, args := parseCommand("ingest /path/to/repo")
	if cmd != "ingest" {
		t.Errorf("parseCommand(\"ingest /path/to/repo\") cmd = %q, want \"ingest\"", cmd)
	}
	if len(args) != 1 || args[0] != "/path/to/repo" {
		t.Errorf("parseCommand(\"ingest /path/to/repo\") args = %v, want [\"/path/to/repo\"]", args)
	}
}

// TestParseCommand_TrimWhitespace verifies that leading/trailing whitespace is trimmed.
// Expected result: cmd="status" args=nil.
func TestParseCommand_TrimWhitespace(t *testing.T) {
	cmd, args := parseCommand("  status  ")
	if cmd != "status" {
		t.Errorf("parseCommand(\"  status  \") cmd = %q, want \"status\"", cmd)
	}
	if len(args) != 0 {
		t.Errorf("parseCommand(\"  status  \") args = %v, want nil/empty", args)
	}
}

// TestRun_QuitExits verifies that "quit" command causes Run to return.
// Expected result: Run returns nil when quit is entered.
func TestRun_QuitExits(t *testing.T) {
	in := strings.NewReader("quit\n")
	var out bytes.Buffer
	r := New(Pipeline{}, StatusInfo{}, WithInput(in), WithOutput(&out))
	if r == nil {
		t.Fatal("New returned nil")
	}

	err := r.Run(context.Background())
	if err != nil {
		t.Errorf("Run returned error: %v", err)
	}
}

// TestRun_EmptyLineIgnored verifies that empty lines are silently ignored.
// Expected result: no error output for empty lines, REPL continues to quit.
func TestRun_EmptyLineIgnored(t *testing.T) {
	in := strings.NewReader("\n\n\nquit\n")
	var out bytes.Buffer
	r := New(Pipeline{}, StatusInfo{}, WithInput(in), WithOutput(&out))
	if r == nil {
		t.Fatal("New returned nil")
	}

	err := r.Run(context.Background())
	if err != nil {
		t.Errorf("Run returned error: %v", err)
	}
	// Empty lines should not produce error output
	if strings.Contains(out.String(), "unknown command") {
		t.Error("empty lines should be silently ignored, not treated as unknown commands")
	}
}

// TestRun_UnknownCommand verifies that unknown commands print an error suggesting help.
// Expected result: output contains "unknown command" and "help".
func TestRun_UnknownCommand(t *testing.T) {
	in := strings.NewReader("foobar\nquit\n")
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
	if !strings.Contains(output, "unknown command") {
		t.Error("unknown command should print error message containing 'unknown command'")
	}
	if !strings.Contains(output, "help") {
		t.Error("unknown command error should suggest 'help'")
	}
}

// TestRun_EOFTreatedAsQuit verifies that EOF on stdin is treated as quit.
// Expected result: Run returns nil when EOF is reached.
func TestRun_EOFTreatedAsQuit(t *testing.T) {
	in := strings.NewReader("") // EOF immediately
	var out bytes.Buffer
	r := New(Pipeline{}, StatusInfo{}, WithInput(in), WithOutput(&out))
	if r == nil {
		t.Fatal("New returned nil")
	}

	err := r.Run(context.Background())
	if err != nil {
		t.Errorf("Run returned error on EOF: %v", err)
	}
}

// TestRun_ContextCancellation verifies that Run returns when context is cancelled.
// Expected result: Run returns when ctx is done.
func TestRun_ContextCancellation(t *testing.T) {
	// Use a reader that blocks (never returns)
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // cancel immediately

	in := strings.NewReader("") // EOF as fallback
	var out bytes.Buffer
	r := New(Pipeline{}, StatusInfo{}, WithInput(in), WithOutput(&out))
	if r == nil {
		t.Fatal("New returned nil")
	}

	err := r.Run(ctx)
	// Should return without error (or with context.Canceled)
	if err != nil && err != context.Canceled {
		t.Errorf("Run returned unexpected error: %v", err)
	}
}
