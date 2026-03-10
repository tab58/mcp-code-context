package repl

import (
	"context"
	"errors"
	"fmt"
	"os"

	"github.com/tab58/code-context/internal/analysis"
	"github.com/tab58/code-context/internal/indexer"
)

// handleIngest runs the full pipeline on a local directory path.
func (r *REPL) handleIngest(ctx context.Context, args []string) error {
	if len(args) == 0 {
		return errors.New("usage: ingest <path>")
	}

	path := args[0]
	fi, err := os.Stat(path)
	if err != nil {
		return fmt.Errorf("path does not exist: %w", err)
	}
	if !fi.IsDir() {
		return fmt.Errorf("path is not a directory: %s", path)
	}

	progress := func(stage, message string) {
		fmt.Fprintf(r.out, "[%s]\t%s\n", stage, message)
	}

	// Step 1: Index
	result, err := r.pipeline.Indexer.IndexRepository(ctx, path, indexer.WithProgress(progress))
	if err != nil {
		return fmt.Errorf("indexing failed: %w", err)
	}

	// Skip analysis if no files changed
	if len(result.FilePaths) == 0 {
		fmt.Fprintf(r.out, "done: no changes detected")
		if result.FilesSkipped > 0 {
			fmt.Fprintf(r.out, " (%d unchanged)", result.FilesSkipped)
		}
		fmt.Fprintln(r.out)
		return nil
	}

	// Step 2: Analyze
	var analyzeResult *analysis.AnalyzeResult
	if r.pipeline.Analyzer != nil {
		analyzeResult, err = r.pipeline.Analyzer.Analyze(ctx, result.RepoID, path, result.FilePaths, analysis.WithAnalyzeProgress(progress))
		if err != nil {
			return fmt.Errorf("analysis failed: %w", err)
		}
	}

	// Step 3: Compute complexity
	if r.pipeline.Analyzer != nil {
		if err := r.pipeline.Analyzer.ComputeComplexity(ctx, result.RepoID, path, result.FilePaths); err != nil {
			return fmt.Errorf("complexity computation failed: %w", err)
		}
	}

	fmt.Fprintf(r.out, "done: indexed %d files, %d folders", result.FilesIndexed, result.FoldersIndexed)
	if analyzeResult != nil {
		fmt.Fprintf(r.out, ", analyzed %d symbols", analyzeResult.Symbols)
	}
	fmt.Fprintln(r.out)
	return nil
}

// handleStatus prints the current configuration to stdout.
func (r *REPL) handleStatus() {
	fmt.Fprintf(r.out, "FalkorDB Host:     %s\n", r.status.FalkorDBHost)
	fmt.Fprintf(r.out, "FalkorDB Port:     %s\n", r.status.FalkorDBPort)
	fmt.Fprintf(r.out, "MCP Port:          %s\n", r.status.MCPPort)
}

// handleList prints all graph names from the CodeDB registry.
func (r *REPL) handleList(ctx context.Context) error {
	if r.pipeline.DB == nil {
		return errors.New("not connected to database")
	}

	graphs, err := r.pipeline.DB.ListRepos(ctx)
	if err != nil {
		return fmt.Errorf("failed to list repositories: %w", err)
	}
	if len(graphs) == 0 {
		fmt.Fprintln(r.out, "no repositories found")
		return nil
	}

	for _, name := range graphs {
		fmt.Fprintf(r.out, "  %s\n", name)
	}

	return nil
}

// handleDelete removes all graph data for a named repository.
func (r *REPL) handleDelete(ctx context.Context, args []string) error {
	if len(args) == 0 {
		return errors.New("usage: delete <repository>")
	}

	name := args[0]
	if r.pipeline.DB == nil {
		return errors.New("not connected to database")
	}

	if err := r.pipeline.DB.DeleteRepo(ctx, name); err != nil {
		return fmt.Errorf("failed to delete repository: %w", err)
	}

	fmt.Fprintf(r.out, "deleted repository %q\n", name)
	return nil
}

// handleHelp prints the command listing to stdout.
func (r *REPL) handleHelp() {
	fmt.Fprintln(r.out, "Available commands:")
	fmt.Fprintln(r.out, "  ingest <path>  Run full pipeline (index -> analyze -> compute complexity) on a local directory")
	fmt.Fprintln(r.out, "  delete <name>  Delete all graph data for a repository")
	fmt.Fprintln(r.out, "  status         Display current configuration")
	fmt.Fprintln(r.out, "  list           List indexed repositories")
	fmt.Fprintln(r.out, "  help           Show this help message")
	fmt.Fprintln(r.out, "  quit           Shut down the application")
}
