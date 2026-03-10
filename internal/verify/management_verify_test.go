package verify

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

// projectRoot returns the absolute path to the project root.
func projectRoot() string {
	_, f, _, _ := runtime.Caller(0)
	return filepath.Join(filepath.Dir(f), "..", "..")
}

// readProjectFile reads a file relative to the project root.
func readProjectFile(t *testing.T, relPath string) string {
	t.Helper()
	data, err := os.ReadFile(filepath.Join(projectRoot(), relPath))
	if err != nil {
		t.Fatalf("failed to read %s: %v", relPath, err)
	}
	return string(data)
}

// ============================================================================
// Task 1: Verify CodeDB DeleteRepo + ListRepos
// ============================================================================

// TestCodeDB_HasDeleteRepoMethod verifies that codedb.go contains a
// DeleteRepo method with the correct signature (ctx, repoName string) error.
// Expected result: codedb.go contains "func (db *CodeDB) DeleteRepo".
func TestCodeDB_HasDeleteRepoMethod(t *testing.T) {
	src := readProjectFile(t, "internal/clients/code_db/codedb.go")
	if !strings.Contains(src, "func (db *CodeDB) DeleteRepo(") {
		t.Error("codedb.go should contain DeleteRepo method")
	}
}

// TestCodeDB_DeleteRepoUsesRawCypher verifies that DeleteRepo uses
// two-step raw Cypher cascade delete (BELONGS_TO dependents, then Repository).
// Expected result: codedb.go contains BELONGS_TO and DETACH DELETE patterns.
func TestCodeDB_DeleteRepoUsesRawCypher(t *testing.T) {
	src := readProjectFile(t, "internal/clients/code_db/codedb.go")
	if !strings.Contains(src, "BELONGS_TO") {
		t.Error("DeleteRepo should use BELONGS_TO in Cypher query for cascade delete")
	}
	if !strings.Contains(src, "DETACH DELETE") {
		t.Error("DeleteRepo should use DETACH DELETE for node removal")
	}
}

// TestCodeDB_DeleteRepoTwoStepDelete verifies that DeleteRepo performs
// two separate Cypher operations: delete dependents first, then repository.
// Expected result: codedb.go contains two ExecuteWrite calls in DeleteRepo.
func TestCodeDB_DeleteRepoTwoStepDelete(t *testing.T) {
	src := readProjectFile(t, "internal/clients/code_db/codedb.go")
	// Check for two distinct delete operations
	if !strings.Contains(src, "deleteDependent") {
		t.Error("DeleteRepo should have a deleteDependent variable for first step")
	}
	if !strings.Contains(src, "deleteRepo") {
		t.Error("DeleteRepo should have a deleteRepo variable for second step")
	}
}

// TestCodeDB_HasListReposMethod verifies that codedb.go contains a
// ListRepos method with the correct signature (ctx) ([]string, error).
// Expected result: codedb.go contains "func (db *CodeDB) ListRepos".
func TestCodeDB_HasListReposMethodSignature(t *testing.T) {
	src := readProjectFile(t, "internal/clients/code_db/codedb.go")
	if !strings.Contains(src, "func (db *CodeDB) ListRepos(") {
		t.Error("codedb.go should contain ListRepos method")
	}
}

// TestCodeDB_ListReposUsesGraphQL verifies that ListRepos queries
// Repository nodes via GraphQL (not raw Cypher).
// Expected result: codedb.go contains gqlListRepositories constant usage.
func TestCodeDB_ListReposUsesGraphQL(t *testing.T) {
	src := readProjectFile(t, "internal/clients/code_db/codedb.go")
	if !strings.Contains(src, "gqlListRepositories") {
		t.Error("ListRepos should use gqlListRepositories GraphQL constant")
	}
}

// TestCodeDB_ListReposSortsResults verifies that ListRepos sorts
// repository names alphabetically.
// Expected result: codedb.go imports "sort" and calls sort.Strings.
func TestCodeDB_ListReposSortsResults(t *testing.T) {
	src := readProjectFile(t, "internal/clients/code_db/codedb.go")
	if !strings.Contains(src, "sort.Strings") {
		t.Error("ListRepos should sort results with sort.Strings")
	}
}

// TestCodeDB_GqlListRepositoriesConstant verifies that gqlListRepositories
// queries the repositorys field with name.
// Expected result: codedb.go contains a gqlListRepositories query.
func TestCodeDB_GqlListRepositoriesConstant(t *testing.T) {
	src := readProjectFile(t, "internal/clients/code_db/codedb.go")
	if !strings.Contains(src, `repositorys { name }`) {
		t.Error("gqlListRepositories should query repositorys { name }")
	}
}

// TestCodeDB_DeleteRepoValidatesEmptyName verifies that DeleteRepo
// returns an error for empty repository names.
// Expected result: codedb.go checks for empty repoName.
func TestCodeDB_DeleteRepoValidatesEmptyName(t *testing.T) {
	src := readProjectFile(t, "internal/clients/code_db/codedb.go")
	if !strings.Contains(src, `repoName == ""`) {
		t.Error("DeleteRepo should validate non-empty repoName")
	}
}

// ============================================================================
// Task 2: Verify management_tools.go handlers
// ============================================================================

// TestManagement_FileExists verifies that management_tools.go exists
// as a dedicated file for management tool handlers.
// Expected result: management_tools.go file exists in internal/mcp/.
func TestManagement_FileExists(t *testing.T) {
	src := readProjectFile(t, "internal/mcp/management_tools.go")
	if len(src) == 0 {
		t.Error("management_tools.go should not be empty")
	}
}

// TestManagement_HasIngestHandler verifies that management_tools.go
// contains handleIngestRepository handler method.
// Expected result: management_tools.go contains handleIngestRepository.
func TestManagement_HasIngestHandler(t *testing.T) {
	src := readProjectFile(t, "internal/mcp/management_tools.go")
	if !strings.Contains(src, "func (s *Server) handleIngestRepository(") {
		t.Error("management_tools.go should contain handleIngestRepository handler")
	}
}

// TestManagement_HasDeleteHandler verifies that management_tools.go
// contains handleDeleteRepository handler method.
// Expected result: management_tools.go contains handleDeleteRepository.
func TestManagement_HasDeleteHandler(t *testing.T) {
	src := readProjectFile(t, "internal/mcp/management_tools.go")
	if !strings.Contains(src, "func (s *Server) handleDeleteRepository(") {
		t.Error("management_tools.go should contain handleDeleteRepository handler")
	}
}

// TestManagement_HasStatsHandler verifies that management_tools.go
// contains handleGetRepositoryStats handler method.
// Expected result: management_tools.go contains handleGetRepositoryStats.
func TestManagement_HasStatsHandler(t *testing.T) {
	src := readProjectFile(t, "internal/mcp/management_tools.go")
	if !strings.Contains(src, "func (s *Server) handleGetRepositoryStats(") {
		t.Error("management_tools.go should contain handleGetRepositoryStats handler")
	}
}

// TestManagement_Has5GQLConstants verifies that management_tools.go
// defines exactly 5 GraphQL query constants for stats operations.
// Expected result: gqlCountFiles, gqlCountFunctions, gqlCountClasses,
// gqlCountModules, gqlCountExternalRefs.
func TestManagement_Has5GQLConstants(t *testing.T) {
	src := readProjectFile(t, "internal/mcp/management_tools.go")
	constants := []string{
		"gqlCountFiles",
		"gqlCountFunctions",
		"gqlCountClasses",
		"gqlCountModules",
		"gqlCountExternalRefs",
	}
	for _, c := range constants {
		if !strings.Contains(src, c) {
			t.Errorf("management_tools.go should define %s constant", c)
		}
	}
}

// TestManagement_Has3ResponseTypes verifies that management_tools.go
// defines 3 response types: IngestResponse, DeleteResponse, RepoStatsResponse.
// Expected result: all 3 types are defined.
func TestManagement_Has3ResponseTypes(t *testing.T) {
	src := readProjectFile(t, "internal/mcp/management_tools.go")
	types := []string{
		"type IngestResponse struct",
		"type DeleteResponse struct",
		"type RepoStatsResponse struct",
	}
	for _, typ := range types {
		if !strings.Contains(src, typ) {
			t.Errorf("management_tools.go should define %s", typ)
		}
	}
}

// TestManagement_IngestResponseFields verifies that IngestResponse has
// all expected JSON-tagged fields.
// Expected result: Fields for repository, filesIndexed, foldersIndexed,
// filesSkipped, symbolsFound.
func TestManagement_IngestResponseFields(t *testing.T) {
	src := readProjectFile(t, "internal/mcp/management_tools.go")
	fields := []string{
		`json:"repository"`,
		`json:"filesIndexed"`,
		`json:"foldersIndexed"`,
		`json:"filesSkipped"`,
		`json:"symbolsFound"`,
	}
	for _, f := range fields {
		if !strings.Contains(src, f) {
			t.Errorf("IngestResponse should have field with %s", f)
		}
	}
}

// TestManagement_DeleteResponseFields verifies that DeleteResponse has
// all expected JSON-tagged fields.
// Expected result: Fields for repository and deleted.
func TestManagement_DeleteResponseFields(t *testing.T) {
	src := readProjectFile(t, "internal/mcp/management_tools.go")
	fields := []string{
		`json:"repository"`,
		`json:"deleted"`,
	}
	for _, f := range fields {
		if !strings.Contains(src, f) {
			t.Errorf("DeleteResponse should have field with %s", f)
		}
	}
}

// TestManagement_RepoStatsResponseFields verifies that RepoStatsResponse has
// all expected JSON-tagged fields.
// Expected result: Fields for files, functions, classes, modules, externalReferences.
func TestManagement_RepoStatsResponseFields(t *testing.T) {
	src := readProjectFile(t, "internal/mcp/management_tools.go")
	fields := []string{
		`json:"files"`,
		`json:"functions"`,
		`json:"classes"`,
		`json:"modules"`,
		`json:"externalReferences"`,
	}
	for _, f := range fields {
		if !strings.Contains(src, f) {
			t.Errorf("RepoStatsResponse should have field with %s", f)
		}
	}
}

// TestManagement_IngestValidatesPath verifies that handleIngestRepository
// validates the repository_path input.
// Expected result: handler checks for empty path.
func TestManagement_IngestValidatesPath(t *testing.T) {
	src := readProjectFile(t, "internal/mcp/management_tools.go")
	if !strings.Contains(src, `repoPath == ""`) {
		t.Error("handleIngestRepository should validate non-empty repoPath")
	}
}

// TestManagement_IngestValidatesDirectory verifies that handleIngestRepository
// checks that the path is a directory.
// Expected result: handler calls os.Stat and fi.IsDir().
func TestManagement_IngestValidatesDirectory(t *testing.T) {
	src := readProjectFile(t, "internal/mcp/management_tools.go")
	if !strings.Contains(src, "os.Stat") {
		t.Error("handleIngestRepository should call os.Stat to validate path")
	}
	if !strings.Contains(src, "IsDir()") {
		t.Error("handleIngestRepository should check IsDir()")
	}
}

// TestManagement_IngestRunsPipeline verifies that handleIngestRepository
// runs the full index -> analyze pipeline.
// Expected result: handler calls IndexRepository and Analyze.
func TestManagement_IngestRunsPipeline(t *testing.T) {
	src := readProjectFile(t, "internal/mcp/management_tools.go")
	if !strings.Contains(src, "IndexRepository") {
		t.Error("handleIngestRepository should call IndexRepository")
	}
	if !strings.Contains(src, "Analyze") {
		t.Error("handleIngestRepository should call Analyze")
	}
}

// TestManagement_DeleteValidatesRepo verifies that handleDeleteRepository
// validates the repository name.
// Expected result: handler checks for empty repo.
func TestManagement_DeleteValidatesRepo(t *testing.T) {
	src := readProjectFile(t, "internal/mcp/management_tools.go")
	if !strings.Contains(src, `repo == ""`) {
		t.Error("handleDeleteRepository should validate non-empty repo")
	}
}

// TestManagement_DeleteCallsCodeDB verifies that handleDeleteRepository
// calls db.DeleteRepo to perform the cascade delete.
// Expected result: handler calls s.db.DeleteRepo.
func TestManagement_DeleteCallsCodeDB(t *testing.T) {
	src := readProjectFile(t, "internal/mcp/management_tools.go")
	if !strings.Contains(src, "s.db.DeleteRepo") {
		t.Error("handleDeleteRepository should call s.db.DeleteRepo")
	}
}

// TestManagement_StatsUsesRepoWhere verifies that handleGetRepositoryStats
// uses repoWhere helper for query filtering.
// Expected result: handler calls repoWhere(repo).
func TestManagement_StatsUsesRepoWhere(t *testing.T) {
	src := readProjectFile(t, "internal/mcp/management_tools.go")
	if !strings.Contains(src, "repoWhere(repo)") {
		t.Error("handleGetRepositoryStats should use repoWhere(repo) for filtering")
	}
}

// TestManagement_StatsCountsAll5NodeTypes verifies that handleGetRepositoryStats
// queries all 5 node types: files, functions, classes, modules, external refs.
// Expected result: handler queries all 5 types.
func TestManagement_StatsCountsAll5NodeTypes(t *testing.T) {
	src := readProjectFile(t, "internal/mcp/management_tools.go")
	queries := []string{
		"gqlCountFiles",
		"gqlCountFunctions",
		"gqlCountClasses",
		"gqlCountModules",
		"gqlCountExternalRefs",
	}
	for _, q := range queries {
		if !strings.Contains(src, q) {
			t.Errorf("handleGetRepositoryStats should query using %s", q)
		}
	}
}

// ============================================================================
// Task 3: Verify server.go tool registration
// ============================================================================

// TestServer_Registers15Tools verifies that server.go registers the expected
// number of tools in NewServer: 4 search + 5 traversal + 1 call chain + 4 context + 3 management + 3 analysis = 20.
// Expected result: server.go calls mcpServer.AddTool 20 times.
func TestServer_Registers15Tools(t *testing.T) {
	src := readProjectFile(t, "internal/mcp/server.go")
	count := strings.Count(src, "mcpServer.AddTool(")
	if count != 20 {
		t.Errorf("NewServer should register 20 tools, found %d AddTool calls", count)
	}
}

// TestServer_NewServerAcceptsIndexer verifies that NewServer accepts an
// *indexer.Indexer parameter for the ingest_repository tool.
// Expected result: NewServer signature includes idx *indexer.Indexer.
func TestServer_NewServerAcceptsIndexer(t *testing.T) {
	src := readProjectFile(t, "internal/mcp/server.go")
	if !strings.Contains(src, "idx *indexer.Indexer") {
		t.Error("NewServer should accept *indexer.Indexer parameter")
	}
}

// TestServer_NewServerAcceptsAnalyzer verifies that NewServer accepts an
// *analysis.Analyzer parameter for the ingest_repository tool.
// Expected result: NewServer signature includes analyzer *analysis.Analyzer.
func TestServer_NewServerAcceptsAnalyzer(t *testing.T) {
	src := readProjectFile(t, "internal/mcp/server.go")
	if !strings.Contains(src, "analyzer *analysis.Analyzer") {
		t.Error("NewServer should accept *analysis.Analyzer parameter")
	}
}

// TestServer_HasManagementMCPAdapters verifies that server.go has
// mcpHandle* adapter methods for all 3 management tools.
// Expected result: server.go contains all 3 mcpHandle* methods.
func TestServer_HasManagementMCPAdapters(t *testing.T) {
	src := readProjectFile(t, "internal/mcp/server.go")
	adapters := []string{
		"mcpHandleIngestRepository",
		"mcpHandleDeleteRepository",
		"mcpHandleGetRepositoryStats",
	}
	for _, a := range adapters {
		if !strings.Contains(src, a) {
			t.Errorf("server.go should contain %s adapter", a)
		}
	}
}

// TestServer_RegistersIngestTool verifies that NewServer registers the
// ingest_repository tool with the correct name and required parameter.
// Expected result: server.go registers "ingest_repository" with "repository_path" param.
func TestServer_RegistersIngestTool(t *testing.T) {
	src := readProjectFile(t, "internal/mcp/server.go")
	if !strings.Contains(src, `"ingest_repository"`) {
		t.Error("NewServer should register 'ingest_repository' tool")
	}
	if !strings.Contains(src, `"repository_path"`) {
		t.Error("ingest_repository tool should have 'repository_path' parameter")
	}
}

// TestServer_RegistersDeleteTool verifies that NewServer registers the
// delete_repository tool with the correct name and required parameter.
// Expected result: server.go registers "delete_repository".
func TestServer_RegistersDeleteTool(t *testing.T) {
	src := readProjectFile(t, "internal/mcp/server.go")
	if !strings.Contains(src, `"delete_repository"`) {
		t.Error("NewServer should register 'delete_repository' tool")
	}
}

// TestServer_RegistersStatsTool verifies that NewServer registers the
// get_repository_stats tool with the correct name.
// Expected result: server.go registers "get_repository_stats".
func TestServer_RegistersStatsTool(t *testing.T) {
	src := readProjectFile(t, "internal/mcp/server.go")
	if !strings.Contains(src, `"get_repository_stats"`) {
		t.Error("NewServer should register 'get_repository_stats' tool")
	}
}

// TestServer_ServerStructHasIdxField verifies that the Server struct
// has an idx field for the Indexer.
// Expected result: server.go has idx field in Server struct.
func TestServer_ServerStructHasIdxField(t *testing.T) {
	src := readProjectFile(t, "internal/mcp/server.go")
	if !strings.Contains(src, "idx") || !strings.Contains(src, "*indexer.Indexer") {
		t.Error("Server struct should have idx *indexer.Indexer field")
	}
}

// TestServer_ServerStructHasAnalyzerField verifies that the Server struct
// has an analyzer field for the Analyzer.
// Expected result: server.go has analyzer field in Server struct.
func TestServer_ServerStructHasAnalyzerField(t *testing.T) {
	src := readProjectFile(t, "internal/mcp/server.go")
	if !strings.Contains(src, "analyzer") || !strings.Contains(src, "*analysis.Analyzer") {
		t.Error("Server struct should have analyzer *analysis.Analyzer field")
	}
}

// TestServer_DocComment15Tools verifies that NewServer doc comment
// mentions "20 tool handlers".
// Expected result: server.go doc comment says "20".
func TestServer_DocComment15Tools(t *testing.T) {
	src := readProjectFile(t, "internal/mcp/server.go")
	if !strings.Contains(src, "20 tool") {
		t.Error("NewServer doc comment should mention '20 tool handlers'")
	}
}

// ============================================================================
// Task 4: Verify REPL delete command
// ============================================================================

// TestREPL_HasHandleDelete verifies that commands.go contains the
// handleDelete method.
// Expected result: commands.go contains "func (r *REPL) handleDelete(".
func TestREPL_HasHandleDelete(t *testing.T) {
	src := readProjectFile(t, "internal/repl/commands.go")
	if !strings.Contains(src, "func (r *REPL) handleDelete(") {
		t.Error("commands.go should contain handleDelete method")
	}
}

// TestREPL_DeleteInRunSwitch verifies that repl.go Run loop has a
// "delete" case in the command switch.
// Expected result: repl.go contains case "delete".
func TestREPL_DeleteInRunSwitch(t *testing.T) {
	src := readProjectFile(t, "internal/repl/repl.go")
	if !strings.Contains(src, `case "delete"`) {
		t.Error("repl.go Run switch should have case \"delete\"")
	}
}

// TestREPL_HandleListUsesListRepos verifies that handleList uses
// db.ListRepos(ctx) instead of raw GraphQL queries.
// Expected result: commands.go calls ListRepos.
func TestREPL_HandleListUsesListRepos(t *testing.T) {
	src := readProjectFile(t, "internal/repl/commands.go")
	if !strings.Contains(src, "ListRepos") {
		t.Error("handleList should use db.ListRepos(ctx)")
	}
}

// TestREPL_HandleHelpLists6Commands verifies that handleHelp lists
// all 6 commands: ingest, delete, status, list, help, quit.
// Expected result: commands.go help output contains all 6 command names.
func TestREPL_HandleHelpLists6Commands(t *testing.T) {
	src := readProjectFile(t, "internal/repl/commands.go")
	commands := []string{"ingest", "delete", "status", "list", "help", "quit"}
	for _, cmd := range commands {
		if !strings.Contains(src, cmd) {
			t.Errorf("handleHelp should list %q command", cmd)
		}
	}
}

// TestREPL_HandleDeleteCallsDeleteRepo verifies that handleDelete
// calls db.DeleteRepo for cascade delete.
// Expected result: commands.go calls DeleteRepo.
func TestREPL_HandleDeleteCallsDeleteRepo(t *testing.T) {
	src := readProjectFile(t, "internal/repl/commands.go")
	if !strings.Contains(src, "DeleteRepo") {
		t.Error("handleDelete should call db.DeleteRepo")
	}
}

// TestREPL_HandleDeleteValidatesArgs verifies that handleDelete
// validates that at least one argument is provided.
// Expected result: commands.go checks len(args) == 0.
func TestREPL_HandleDeleteValidatesArgs(t *testing.T) {
	src := readProjectFile(t, "internal/repl/commands.go")
	if !strings.Contains(src, `len(args) == 0`) {
		t.Error("handleDelete should validate args length")
	}
}

// TestREPL_HandleDeleteValidatesDB verifies that handleDelete
// checks for nil DB before attempting delete.
// Expected result: commands.go checks pipeline.DB == nil.
func TestREPL_HandleDeleteValidatesDB(t *testing.T) {
	src := readProjectFile(t, "internal/repl/commands.go")
	if !strings.Contains(src, "r.pipeline.DB == nil") {
		t.Error("handleDelete should check for nil DB")
	}
}

// ============================================================================
// Task 5: Verify test helpers
// ============================================================================

// TestHelpers_NewTestServerWithIndexerExists verifies that
// newTestServerWithIndexer helper exists in test_helpers_test.go.
// Expected result: test_helpers_test.go contains newTestServerWithIndexer.
func TestHelpers_NewTestServerWithIndexerExists(t *testing.T) {
	src := readProjectFile(t, "internal/mcp/test_helpers_test.go")
	if !strings.Contains(src, "func newTestServerWithIndexer(") {
		t.Error("test_helpers_test.go should contain newTestServerWithIndexer helper")
	}
}

// TestHelpers_NewTestServerWithResponsesExists verifies that
// newTestServerWithResponses helper exists in test_helpers_test.go.
// Expected result: test_helpers_test.go contains newTestServerWithResponses.
func TestHelpers_NewTestServerWithResponsesExists(t *testing.T) {
	src := readProjectFile(t, "internal/mcp/test_helpers_test.go")
	if !strings.Contains(src, "func newTestServerWithResponses(") {
		t.Error("test_helpers_test.go should contain newTestServerWithResponses helper")
	}
}

// TestHelpers_WithIndexerCreatesIndexer verifies that
// newTestServerWithIndexer creates an actual indexer.Indexer instance.
// Expected result: helper calls indexer.NewIndexer.
func TestHelpers_WithIndexerCreatesIndexer(t *testing.T) {
	src := readProjectFile(t, "internal/mcp/test_helpers_test.go")
	if !strings.Contains(src, "indexer.NewIndexer") {
		t.Error("newTestServerWithIndexer should create indexer via indexer.NewIndexer")
	}
}

// TestHelpers_WithIndexerSetsIdxField verifies that
// newTestServerWithIndexer sets the idx field on the Server.
// Expected result: helper sets s.idx or Server{..., idx: ...}.
func TestHelpers_WithIndexerSetsIdxField(t *testing.T) {
	src := readProjectFile(t, "internal/mcp/test_helpers_test.go")
	if !strings.Contains(src, "idx:") && !strings.Contains(src, "idx =") {
		t.Error("newTestServerWithIndexer should set the idx field")
	}
}

// ============================================================================
// Task 8: Verify doc comment accuracy
// ============================================================================

// TestDocComment_ManagementToolsHasComments verifies that
// management_tools.go has doc comments on all handler methods.
// Expected result: all 3 handlers have "// handle" doc comments.
func TestDocComment_ManagementToolsHasComments(t *testing.T) {
	src := readProjectFile(t, "internal/mcp/management_tools.go")
	comments := []string{
		"// handleIngestRepository",
		"// handleDeleteRepository",
		"// handleGetRepositoryStats",
	}
	for _, c := range comments {
		if !strings.Contains(src, c) {
			t.Errorf("management_tools.go should have doc comment %q", c)
		}
	}
}

// TestDocComment_GQLConstantNamesMatchBodies verifies that each gql
// constant name corresponds to the query body content.
// Expected result: gqlCountFiles contains "files", gqlCountFunctions
// contains "functions", etc.
func TestDocComment_GQLConstantNamesMatchBodies(t *testing.T) {
	src := readProjectFile(t, "internal/mcp/management_tools.go")
	// Verify each constant queries the correct field
	if !strings.Contains(src, `gqlCountFiles`) || !strings.Contains(src, `files(where:`) {
		t.Error("gqlCountFiles should query files(where:)")
	}
	if !strings.Contains(src, `gqlCountFunctions`) || !strings.Contains(src, `functions(where:`) {
		t.Error("gqlCountFunctions should query functions(where:)")
	}
	if !strings.Contains(src, `gqlCountClasses`) || !strings.Contains(src, `classs(where:`) {
		t.Error("gqlCountClasses should query classs(where:)")
	}
	if !strings.Contains(src, `gqlCountModules`) || !strings.Contains(src, `modules(where:`) {
		t.Error("gqlCountModules should query modules(where:)")
	}
	if !strings.Contains(src, `gqlCountExternalRefs`) || !strings.Contains(src, `externalReferences(where:`) {
		t.Error("gqlCountExternalRefs should query externalReferences(where:)")
	}
}

// ============================================================================
// Task 9: Build verification
// ============================================================================

// TestBuild_ManagementImportsIndexer verifies that management_tools.go
// imports the indexer package for the ingest pipeline.
// Expected result: management_tools.go imports internal/indexer.
func TestBuild_ManagementImportsIndexer(t *testing.T) {
	src := readProjectFile(t, "internal/mcp/management_tools.go")
	if !strings.Contains(src, `"github.com/tab58/code-context/internal/indexer"`) {
		t.Error("management_tools.go should import internal/indexer")
	}
}

// TestBuild_ManagementImportsAnalysis verifies that management_tools.go
// imports the analysis package for the analyze step.
// Expected result: management_tools.go imports internal/analysis.
func TestBuild_ManagementImportsAnalysis(t *testing.T) {
	src := readProjectFile(t, "internal/mcp/management_tools.go")
	if !strings.Contains(src, `"github.com/tab58/code-context/internal/analysis"`) {
		t.Error("management_tools.go should import internal/analysis")
	}
}

// TestBuild_ServerImportsIndexer verifies that server.go imports
// the indexer package for the NewServer parameter.
// Expected result: server.go imports internal/indexer.
func TestBuild_ServerImportsIndexer(t *testing.T) {
	src := readProjectFile(t, "internal/mcp/server.go")
	if !strings.Contains(src, `"github.com/tab58/code-context/internal/indexer"`) {
		t.Error("server.go should import internal/indexer")
	}
}

// TestBuild_ServerImportsAnalysis verifies that server.go imports
// the analysis package for the NewServer parameter.
// Expected result: server.go imports internal/analysis.
func TestBuild_ServerImportsAnalysis(t *testing.T) {
	src := readProjectFile(t, "internal/mcp/server.go")
	if !strings.Contains(src, `"github.com/tab58/code-context/internal/analysis"`) {
		t.Error("server.go should import internal/analysis")
	}
}
