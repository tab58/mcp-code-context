package repl

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/tab58/code-context/internal/analysis"
	codedb "github.com/tab58/code-context/internal/clients/code_db"
	"github.com/tab58/code-context/internal/indexer"
)

// ProgressFunc is called by pipeline stages to report progress.
// stage is "indexing" or "analyzing".
// message is a human-readable status line (e.g., file path or node name).
type ProgressFunc func(stage, message string)

// Pipeline holds the shared pipeline components used by REPL commands.
type Pipeline struct {
	DB       *codedb.CodeDB
	Indexer  *indexer.Indexer
	Analyzer *analysis.Analyzer
}

// StatusInfo holds displayable configuration for the status command.
type StatusInfo struct {
	FalkorDBHost string
	FalkorDBPort string
	MCPPort      string
}

// REPL provides an interactive command-line interface on stdin/stdout.
type REPL struct {
	pipeline Pipeline
	status   StatusInfo
	in       io.Reader
	out      io.Writer
}

// Option configures REPL construction.
type Option func(*REPL)

// WithInput sets the input reader (default: os.Stdin). Used for testing.
func WithInput(r io.Reader) Option {
	return func(repl *REPL) {
		repl.in = r
	}
}

// WithOutput sets the output writer (default: os.Stdout). Used for testing.
func WithOutput(w io.Writer) Option {
	return func(repl *REPL) {
		repl.out = w
	}
}

// New creates a REPL wired to the given pipeline.
func New(pipeline Pipeline, status StatusInfo, opts ...Option) *REPL {
	r := &REPL{
		pipeline: pipeline,
		status:   status,
		in:       os.Stdin,
		out:      os.Stdout,
	}
	for _, opt := range opts {
		opt(r)
	}
	return r
}

// scanResult holds a line read from stdin or signals EOF.
type scanResult struct {
	line string
	eof  bool
}

// Run starts the REPL loop. Reads lines from stdin, parses commands, dispatches
// to handlers. Blocks until "quit" command or ctx cancellation.
func (r *REPL) Run(ctx context.Context) error {
	lines := make(chan scanResult)
	go func() {
		defer close(lines)
		scanner := bufio.NewScanner(r.in)
		for scanner.Scan() {
			lines <- scanResult{line: scanner.Text()}
		}
		lines <- scanResult{eof: true}
	}()

	for {
		fmt.Fprint(r.out, "> ")

		select {
		case <-ctx.Done():
			return nil
		case result, ok := <-lines:
			if !ok || result.eof {
				return nil
			}

			cmd, args := parseCommand(result.line)
			if cmd == "" {
				continue
			}

			switch cmd {
			case "quit":
				return nil
			case "help":
				r.handleHelp()
			case "status":
				r.handleStatus()
			case "list":
				if err := r.handleList(ctx); err != nil {
					fmt.Fprintf(r.out, "error: %v\n", err)
				}
			case "ingest":
				if err := r.handleIngest(ctx, args); err != nil {
					fmt.Fprintf(r.out, "error: %v\n", err)
				}
			case "delete":
				if err := r.handleDelete(ctx, args); err != nil {
					fmt.Fprintf(r.out, "error: %v\n", err)
				}
			default:
				fmt.Fprintf(r.out, "unknown command %q. Type \"help\" for available commands.\n", cmd)
			}
		}
	}
}

// parseCommand splits an input line into command name and arguments.
// Returns empty command for blank lines.
func parseCommand(line string) (cmd string, args []string) {
	fields := strings.Fields(line)
	if len(fields) == 0 {
		return "", nil
	}
	return fields[0], fields[1:]
}
