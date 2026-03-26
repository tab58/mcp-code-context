package app

import (
	"context"
	"fmt"

	"github.com/tab58/code-context/internal/analysis"
	codedb "github.com/tab58/code-context/internal/clients/code_db"
	"github.com/tab58/code-context/internal/indexer"
	tools "github.com/tab58/code-context/internal/tools"
)

type Application interface {
	GetAppVersion() string
	Ingest(ctx context.Context, path string) (*IngestResult, error)
	ListRepositories(ctx context.Context) ([]string, error)

	HandleFindCallChain(ctx context.Context, repo, source, target string, maxDepth int) (*tools.CallChainResponse, error)
	HandleGetRepoMap(ctx context.Context, repo string) (*tools.RepoMapResponse, error)
	HandleGetFileOverview(ctx context.Context, repo, path string) (*tools.FileOverviewResponse, error)
	HandleGetSymbolContext(ctx context.Context, repo, name string) (*tools.SymbolContextResponse, error)
	HandleReadSource(ctx context.Context, repo string, names []string) (*tools.ReadSourceResponse, error)
	HandleFindDeadCode(ctx context.Context, repo string, excludeDecorated bool, excludePatterns string, limit int) (*tools.DeadCodeResponse, error)
	HandleSearchCodeContent(ctx context.Context, repo, query string, limit int) (*tools.SearchResponse, error)
	HandleDeleteRepository(ctx context.Context, repo string) (*tools.DeleteResponse, error)
	HandleGetRepositoryStats(ctx context.Context, repo string) (*tools.RepoStatsResponse, error)
	HandleCalculateCyclomaticComplexity(ctx context.Context, repo, name string) (*tools.ComplexityResponse, error)
	HandleFindMostComplexFunctions(ctx context.Context, repo string, minComplexity, limit int) (*tools.ComplexityResponse, error)
	HandleFindFunction(ctx context.Context, repo, name string) (*tools.SearchResponse, error)
	HandleFindFile(ctx context.Context, repo, pattern string) (*tools.SearchResponse, error)
	HandleSearchCode(ctx context.Context, repo, query string, limit int) (*tools.SearchResponse, error)
	HandleGetCallers(ctx context.Context, repo, name string, depth int) (*tools.TraversalResponse, error)
	HandleGetCallees(ctx context.Context, repo, name string, depth int) (*tools.TraversalResponse, error)
	HandleGetClassHierarchy(ctx context.Context, repo, name, direction string, depth int) (*tools.TraversalResponse, error)
	HandleGetDependencies(ctx context.Context, repo, name string, depth int) (*tools.TraversalResponse, error)
	HandleGetReferences(ctx context.Context, repo, name string) (*tools.TraversalResponse, error)
}

type application struct {
	appVersion string
	db         *codedb.CodeDB
	indexer    *indexer.Indexer
	analyzer   *analysis.Analyzer
	mcpTools   *tools.Manager
}

type ApplicationConfig struct {
	AppVersion string
	DB         *codedb.CodeDB
	Indexer    *indexer.Indexer
	Analyzer   *analysis.Analyzer
}

func NewApplication(config *ApplicationConfig) Application {
	toolManager := tools.NewManager(config.DB, config.Analyzer)
	return &application{
		appVersion: config.AppVersion,
		db:         config.DB,
		indexer:    config.Indexer,
		analyzer:   config.Analyzer,
		mcpTools:   toolManager,
	}
}

// IngestResult holds the outcome of a full ingest pipeline run.
type IngestResult struct {
	Repository     string `json:"repository"`
	FilesIndexed   int    `json:"filesIndexed"`
	FoldersIndexed int    `json:"foldersIndexed"`
	FilesSkipped   int    `json:"filesSkipped"`
	SymbolsFound   int    `json:"symbolsFound"`
}

func (a *application) GetAppVersion() string {
	return a.appVersion
}

// Ingest runs the full pipeline (index -> analyze -> complexity) on a local directory.
func (a *application) Ingest(ctx context.Context, path string) (*IngestResult, error) {
	result, err := a.indexer.IndexRepository(ctx, path)
	if err != nil {
		return nil, fmt.Errorf("indexing failed: %w", err)
	}

	res := &IngestResult{
		Repository:     result.RepoID,
		FilesIndexed:   result.FilesIndexed,
		FoldersIndexed: result.FoldersIndexed,
		FilesSkipped:   result.FilesSkipped,
	}

	if len(result.FilePaths) == 0 {
		return res, nil
	}

	if a.analyzer != nil {
		analyzeResult, err := a.analyzer.Analyze(ctx, result.RepoID, path, result.FilePaths)
		if err != nil {
			return nil, fmt.Errorf("analysis failed: %w", err)
		}
		res.SymbolsFound = analyzeResult.Symbols

		if err := a.analyzer.ComputeComplexity(ctx, result.RepoID, path, result.FilePaths); err != nil {
			return nil, fmt.Errorf("complexity computation failed: %w", err)
		}
	}

	return res, nil
}

// ListRepositories returns all indexed repository names.
func (a *application) ListRepositories(ctx context.Context) ([]string, error) {
	return a.db.ListRepos(ctx)
}
