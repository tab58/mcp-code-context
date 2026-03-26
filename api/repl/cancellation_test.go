package repl

import (
	"bytes"
	"context"
	"io"
	"testing"
	"time"
)

// === Task 2: REPL goroutine+channel+select context cancellation ===
//
// The REPL was refactored to use a goroutine reading stdin into a channel,
// with a select statement that checks both the channel and ctx.Done().
// This ensures Run() exits promptly when the context is cancelled,
// even if stdin is blocked (no input available).

// TestRun_ContextCancellationWithBlockingPipe verifies that Run() exits
// promptly when context is cancelled even if stdin is a blocking pipe
// with no data and no close. With the old plain-Scanner pattern, Run()
// would hang forever because scanner.Scan() blocks in the pipe Read.
// The goroutine+select pattern lets the select case <-ctx.Done() fire
// while the scanner goroutine remains blocked on the pipe.
// Expected result: Run returns within 500ms of context cancellation.
func TestRun_ContextCancellationWithBlockingPipe(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())

	// Use a pipe where we never write and never close the write end.
	// scanner.Scan() will block in the pipe Read call indefinitely.
	pr, pw := io.Pipe()
	defer pw.Close() // cleanup after test

	var out bytes.Buffer
	r := New(Pipeline{}, StatusInfo{}, WithInput(pr), WithOutput(&out))

	done := make(chan error, 1)
	go func() {
		done <- r.Run(ctx)
	}()

	// Give Run a moment to start and block on the pipe
	time.Sleep(50 * time.Millisecond)

	// Cancel context — Run should exit via select case <-ctx.Done()
	// even though the scanner goroutine is still blocked reading the pipe
	cancel()

	select {
	case err := <-done:
		if err != nil {
			t.Errorf("Run returned error: %v, expected nil", err)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("Run did not exit within 2s after context cancellation — " +
			"select+ctx.Done() pattern not working (scanner blocking on pipe)")
	}
}

// TestRun_ConcurrentContextCancelAndInput verifies that Run() handles
// the race between context cancellation and stdin input correctly.
// Expected result: Run returns without panic or hang.
func TestRun_ConcurrentContextCancelAndInput(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())

	// Pipe: write commands while cancelling concurrently
	pr, pw := io.Pipe()
	var out bytes.Buffer
	r := New(Pipeline{}, StatusInfo{}, WithInput(pr), WithOutput(&out))

	done := make(chan error, 1)
	go func() {
		done <- r.Run(ctx)
	}()

	// Write a command and cancel concurrently
	go func() {
		time.Sleep(10 * time.Millisecond)
		_, _ = pw.Write([]byte("help\n"))
		time.Sleep(10 * time.Millisecond)
		cancel()
		pw.Close()
	}()

	select {
	case err := <-done:
		if err != nil && err != context.Canceled {
			t.Errorf("Run returned unexpected error: %v", err)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("Run did not exit within 2s — possible deadlock in channel+select pattern")
	}
}
