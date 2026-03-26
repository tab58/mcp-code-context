package app

import (
	"context"

	tools "github.com/tab58/code-context/internal/tools"
)

func (a *application) HandleFindCallChain(ctx context.Context, repo, source, target string, maxDepth int) (*tools.CallChainResponse, error) {
	return a.mcpTools.HandleFindCallChain(ctx, repo, source, target, maxDepth)
}

func (a *application) HandleGetRepoMap(ctx context.Context, repo string) (*tools.RepoMapResponse, error) {
	return a.mcpTools.HandleGetRepoMap(ctx, repo)
}

func (a *application) HandleGetFileOverview(ctx context.Context, repo, path string) (*tools.FileOverviewResponse, error) {
	return a.mcpTools.HandleGetFileOverview(ctx, repo, path)
}

func (a *application) HandleGetSymbolContext(ctx context.Context, repo, name string) (*tools.SymbolContextResponse, error) {
	return a.mcpTools.HandleGetSymbolContext(ctx, repo, name)
}

func (a *application) HandleReadSource(ctx context.Context, repo string, names []string) (*tools.ReadSourceResponse, error) {
	return a.mcpTools.HandleReadSource(ctx, repo, names)
}

func (a *application) HandleFindDeadCode(ctx context.Context, repo string, excludeDecorated bool, excludePatterns string, limit int) (*tools.DeadCodeResponse, error) {
	return a.mcpTools.HandleFindDeadCode(ctx, repo, excludeDecorated, excludePatterns, limit)
}

func (a *application) HandleSearchCodeContent(ctx context.Context, repo, query string, limit int) (*tools.SearchResponse, error) {
	return a.mcpTools.HandleSearchCodeContent(ctx, repo, query, limit)
}

func (a *application) HandleDeleteRepository(ctx context.Context, repo string) (*tools.DeleteResponse, error) {
	return a.mcpTools.HandleDeleteRepository(ctx, repo)
}

func (a *application) HandleGetRepositoryStats(ctx context.Context, repo string) (*tools.RepoStatsResponse, error) {
	return a.mcpTools.HandleGetRepositoryStats(ctx, repo)
}

func (a *application) HandleCalculateCyclomaticComplexity(ctx context.Context, repo, name string) (*tools.ComplexityResponse, error) {
	return a.mcpTools.HandleCalculateCyclomaticComplexity(ctx, repo, name)
}

func (a *application) HandleFindMostComplexFunctions(ctx context.Context, repo string, minComplexity, limit int) (*tools.ComplexityResponse, error) {
	return a.mcpTools.HandleFindMostComplexFunctions(ctx, repo, minComplexity, limit)
}

func (a *application) HandleFindFunction(ctx context.Context, repo, name string) (*tools.SearchResponse, error) {
	return a.mcpTools.HandleFindFunction(ctx, repo, name)
}

func (a *application) HandleFindFile(ctx context.Context, repo, pattern string) (*tools.SearchResponse, error) {
	return a.mcpTools.HandleFindFile(ctx, repo, pattern)
}

func (a *application) HandleSearchCode(ctx context.Context, repo, query string, limit int) (*tools.SearchResponse, error) {
	return a.mcpTools.HandleSearchCode(ctx, repo, query, limit)
}

func (a *application) HandleGetCallers(ctx context.Context, repo, name string, depth int) (*tools.TraversalResponse, error) {
	return a.mcpTools.HandleGetCallers(ctx, repo, name, depth)
}

func (a *application) HandleGetCallees(ctx context.Context, repo, name string, depth int) (*tools.TraversalResponse, error) {
	return a.mcpTools.HandleGetCallees(ctx, repo, name, depth)
}

func (a *application) HandleGetClassHierarchy(ctx context.Context, repo, name, direction string, depth int) (*tools.TraversalResponse, error) {
	return a.mcpTools.HandleGetClassHierarchy(ctx, repo, name, direction, depth)
}

func (a *application) HandleGetDependencies(ctx context.Context, repo, name string, depth int) (*tools.TraversalResponse, error) {
	return a.mcpTools.HandleGetDependencies(ctx, repo, name, depth)
}

func (a *application) HandleGetReferences(ctx context.Context, repo, name string) (*tools.TraversalResponse, error) {
	return a.mcpTools.HandleGetReferences(ctx, repo, name)
}
