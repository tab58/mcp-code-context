package rlm

import (
	"fmt"
	"io"
	"os"
	"sync"
	"time"
)

// TraceLogger receives structured events from the RLM engine loop.
type TraceLogger interface {
	LogLLMResponse(turn int, response string)
	LogREPLInput(turn int, blockIndex int, code string)
	LogREPLOutput(turn int, blockIndex int, output string)
	LogFinal(turn int, answer string)
	Close() error
}

// FileTraceLogger writes RLM trace events to an io.Writer (typically a file).
type FileTraceLogger struct {
	mu sync.Mutex
	w  io.WriteCloser
}

// NewFileTraceLogger opens the given path for append-only writing and returns a logger.
func NewFileTraceLogger(path string) (*FileTraceLogger, error) {
	f, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0600)
	if err != nil {
		return nil, fmt.Errorf("open trace log: %w", err)
	}
	return &FileTraceLogger{w: f}, nil
}

func (l *FileTraceLogger) write(section, body string) {
	l.mu.Lock()
	defer l.mu.Unlock()
	ts := time.Now().UTC().Format(time.RFC3339Nano)
	fmt.Fprintf(l.w, "=== %s [%s] ===\n%s\n=== END %s ===\n\n", section, ts, body, section)
}

func (l *FileTraceLogger) LogLLMResponse(turn int, response string) {
	l.write(fmt.Sprintf("LLM_RESPONSE turn=%d", turn), response)
}

func (l *FileTraceLogger) LogREPLInput(turn int, blockIndex int, code string) {
	l.write(fmt.Sprintf("REPL_INPUT turn=%d block=%d", turn, blockIndex), code)
}

func (l *FileTraceLogger) LogREPLOutput(turn int, blockIndex int, output string) {
	l.write(fmt.Sprintf("REPL_OUTPUT turn=%d block=%d", turn, blockIndex), output)
}

func (l *FileTraceLogger) LogFinal(turn int, answer string) {
	l.write(fmt.Sprintf("FINAL turn=%d", turn), answer)
}

func (l *FileTraceLogger) Close() error {
	l.mu.Lock()
	defer l.mu.Unlock()
	return l.w.Close()
}
