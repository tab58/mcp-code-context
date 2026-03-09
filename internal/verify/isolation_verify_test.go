package verify

import (
	"os"
	"strings"
	"testing"
)

// --- Task 1: FalkorDBConfig has GraphName (not Database) ---

// TestFalkorDBConfig_HasGraphName verifies that internal/config/falkordb.go
// contains a "GraphName" field in the FalkorDBConfig struct.
// Expected result: "GraphName" in falkordb.go.
func TestFalkorDBConfig_HasGraphName(t *testing.T) {
	data, err := os.ReadFile("../../internal/config/falkordb.go")
	if err != nil {
		t.Fatalf("failed to read falkordb.go: %v", err)
	}
	source := string(data)
	if !strings.Contains(source, "GraphName") {
		t.Error("FalkorDBConfig should contain a GraphName field")
	}
}

// TestFalkorDBConfig_NoDatabaseField verifies that internal/config/falkordb.go
// does NOT contain a "Database" field in the FalkorDBConfig struct.
// Expected result: No line matching "Database string" in falkordb.go.
func TestFalkorDBConfig_NoDatabaseField(t *testing.T) {
	data, err := os.ReadFile("../../internal/config/falkordb.go")
	if err != nil {
		t.Fatalf("failed to read falkordb.go: %v", err)
	}
	source := string(data)
	if strings.Contains(source, "Database string") {
		t.Error("FalkorDBConfig should NOT contain a Database field — use GraphName instead")
	}
}

// TestAppConfig_HasFalkorDBGraph verifies that cmd/codectx/config/config.go
// contains a "FalkorDBGraph" field.
// Expected result: "FalkorDBGraph" in config.go.
func TestAppConfig_HasFalkorDBGraph(t *testing.T) {
	data, err := os.ReadFile("../../cmd/codectx/config/config.go")
	if err != nil {
		t.Fatalf("failed to read config.go: %v", err)
	}
	source := string(data)
	if !strings.Contains(source, "FalkorDBGraph") {
		t.Error("app Config should contain FalkorDBGraph field")
	}
}

// TestAppConfig_NoFalkorDBDatabase verifies that cmd/codectx/config/config.go
// does NOT contain a "FalkorDBDatabase" field.
// Expected result: No "FalkorDBDatabase" in config.go.
func TestAppConfig_NoFalkorDBDatabase(t *testing.T) {
	data, err := os.ReadFile("../../cmd/codectx/config/config.go")
	if err != nil {
		t.Fatalf("failed to read config.go: %v", err)
	}
	source := string(data)
	if strings.Contains(source, "FalkorDBDatabase") {
		t.Error("app Config should NOT contain FalkorDBDatabase field")
	}
}

// TestAppConfig_NoDefaultFalkorDBDatabase verifies that cmd/codectx/config/config.go
// does NOT contain the DefaultFalkorDBDatabase constant.
// Expected result: No "DefaultFalkorDBDatabase" in config.go.
func TestAppConfig_NoDefaultFalkorDBDatabase(t *testing.T) {
	data, err := os.ReadFile("../../cmd/codectx/config/config.go")
	if err != nil {
		t.Fatalf("failed to read config.go: %v", err)
	}
	source := string(data)
	if strings.Contains(source, "DefaultFalkorDBDatabase") {
		t.Error("app config should NOT contain DefaultFalkorDBDatabase constant")
	}
}

// TestAppConfig_NoFALKORDB_DATABASE_EnvVar verifies that cmd/codectx/config/config.go
// does NOT reference the FALKORDB_DATABASE env var.
// Expected result: No "FALKORDB_DATABASE" in config.go.
func TestAppConfig_NoFALKORDB_DATABASE_EnvVar(t *testing.T) {
	data, err := os.ReadFile("../../cmd/codectx/config/config.go")
	if err != nil {
		t.Fatalf("failed to read config.go: %v", err)
	}
	source := string(data)
	if strings.Contains(source, "FALKORDB_DATABASE") {
		t.Error("app config should NOT reference FALKORDB_DATABASE env var")
	}
}

// TestMainGo_HasGraphNameInFalkorDBConfig verifies that cmd/codectx/main.go
// sets GraphName when constructing falkorDBConfig.
// Expected result: "GraphName:" in main.go.
func TestMainGo_HasGraphNameInFalkorDBConfig(t *testing.T) {
	data, err := os.ReadFile("../../cmd/codectx/main.go")
	if err != nil {
		t.Fatalf("failed to read main.go: %v", err)
	}
	source := string(data)
	if !strings.Contains(source, "GraphName:") {
		t.Error("main.go should set GraphName field on FalkorDBConfig")
	}
}

// --- Task 3: Migrate indexer to ForRepo ---

// TestIndexer_NoClientCall verifies that indexer.go does not call db.Client()
// directly — all access should go through ForRepo.
// Expected result: No "db.Client()" or "idx.db.Client()" in indexer.go.
func TestIndexer_NoClientCall(t *testing.T) {
	data, err := os.ReadFile("../../internal/indexer/indexer.go")
	if err != nil {
		t.Fatalf("failed to read indexer.go: %v", err)
	}
	source := string(data)
	if strings.Contains(source, ".Client()") {
		t.Error("indexer.go should not call .Client() — use ForRepo instead")
	}
}

// TestIndexer_UsesForRepo verifies that indexer.go calls ForRepo to get
// a repo-scoped client.
// Expected result: "ForRepo" appears in indexer.go.
func TestIndexer_UsesForRepo(t *testing.T) {
	data, err := os.ReadFile("../../internal/indexer/indexer.go")
	if err != nil {
		t.Fatalf("failed to read indexer.go: %v", err)
	}
	source := string(data)
	if !strings.Contains(source, "ForRepo") {
		t.Error("indexer.go should use db.ForRepo to get a repo-scoped client")
	}
}

// --- Task 4: Migrate analyzer to ForRepo ---

// TestAnalyzer_NoClientCall verifies that analyzer.go does not call db.Client()
// directly — all access should go through ForRepo.
// Expected result: No ".Client()" in analyzer.go.
func TestAnalyzer_NoClientCall(t *testing.T) {
	data, err := os.ReadFile("../../internal/analysis/analyzer.go")
	if err != nil {
		t.Fatalf("failed to read analyzer.go: %v", err)
	}
	source := string(data)
	if strings.Contains(source, ".Client()") {
		t.Error("analyzer.go should not call .Client() — use ForRepo instead")
	}
}

// TestAnalyzer_UsesForRepo verifies that analyzer.go calls ForRepo.
// Expected result: "ForRepo" appears in analyzer.go.
func TestAnalyzer_UsesForRepo(t *testing.T) {
	data, err := os.ReadFile("../../internal/analysis/analyzer.go")
	if err != nil {
		t.Fatalf("failed to read analyzer.go: %v", err)
	}
	source := string(data)
	if !strings.Contains(source, "ForRepo") {
		t.Error("analyzer.go should use db.ForRepo to get a repo-scoped client")
	}
}

// --- Task 6: Migrate MCP tools to ForRepo ---

// TestMCPTools_NoClientCall verifies that tools.go does not call db.Client()
// directly — all access should go through ForRepo.
// Expected result: No ".Client()" in tools.go.
func TestMCPTools_NoClientCall(t *testing.T) {
	data, err := os.ReadFile("../../internal/mcp/tools.go")
	if err != nil {
		t.Fatalf("failed to read tools.go: %v", err)
	}
	source := string(data)
	if strings.Contains(source, ".Client()") {
		t.Error("tools.go should not call .Client() — use ForRepo instead")
	}
}

// TestMCPTools_UsesForRepo verifies that tools.go calls ForRepo.
// Expected result: "ForRepo" appears in tools.go.
func TestMCPTools_UsesForRepo(t *testing.T) {
	data, err := os.ReadFile("../../internal/mcp/tools.go")
	if err != nil {
		t.Fatalf("failed to read tools.go: %v", err)
	}
	source := string(data)
	if !strings.Contains(source, "ForRepo") {
		t.Error("tools.go should use db.ForRepo to get a repo-scoped client")
	}
}

// --- Shared graph architecture ---

// TestREPL_ListUsesListRepos verifies that commands.go calls
// ListRepos for the list command.
// Expected result: "ListRepos" appears in commands.go.
func TestREPL_ListUsesListRepos(t *testing.T) {
	data, err := os.ReadFile("../../internal/repl/commands.go")
	if err != nil {
		t.Fatalf("failed to read commands.go: %v", err)
	}
	source := string(data)
	if !strings.Contains(source, "ListRepos") {
		t.Error("commands.go should use db.ListRepos for the list command")
	}
}

// TestREPL_NoFalkorDBDatabase verifies that repl.go no longer contains
// FalkorDBDatabase in the StatusInfo struct.
// Expected result: No "FalkorDBDatabase" in repl.go.
func TestREPL_NoFalkorDBDatabase(t *testing.T) {
	data, err := os.ReadFile("../../internal/repl/repl.go")
	if err != nil {
		t.Fatalf("failed to read repl.go: %v", err)
	}
	source := string(data)
	if strings.Contains(source, "FalkorDBDatabase") {
		t.Error("repl.go StatusInfo should not contain FalkorDBDatabase")
	}
}

// TestCodeDB_HasForRepoMethod verifies ForRepo method exists on CodeDB.
// Expected result: "func (db *CodeDB) ForRepo" in codedb.go.
func TestCodeDB_HasForRepoMethod(t *testing.T) {
	data, err := os.ReadFile("../../internal/clients/code_db/codedb.go")
	if err != nil {
		t.Fatalf("failed to read codedb.go: %v", err)
	}
	source := string(data)
	if !strings.Contains(source, "func (db *CodeDB) ForRepo") {
		t.Error("codedb.go should define ForRepo method")
	}
}

// TestCodeDB_HasListReposMethod verifies ListRepos method exists on CodeDB.
// Expected result: "func (db *CodeDB) ListRepos" in codedb.go.
func TestCodeDB_HasListReposMethod(t *testing.T) {
	data, err := os.ReadFile("../../internal/clients/code_db/codedb.go")
	if err != nil {
		t.Fatalf("failed to read codedb.go: %v", err)
	}
	source := string(data)
	if !strings.Contains(source, "func (db *CodeDB) ListRepos") {
		t.Error("codedb.go should define ListRepos method")
	}
}

// TestCodeDB_SharedGraphArchitecture verifies CodeDB uses a shared graph
// (sharedClient) instead of per-repo graphs (repos map).
func TestCodeDB_SharedGraphArchitecture(t *testing.T) {
	data, err := os.ReadFile("../../internal/clients/code_db/codedb.go")
	if err != nil {
		t.Fatalf("failed to read codedb.go: %v", err)
	}
	source := string(data)
	if !strings.Contains(source, "sharedClient") {
		t.Error("codedb.go should define a sharedClient struct")
	}
	if strings.Contains(source, "repos  map[string]") {
		t.Error("codedb.go should not have a per-repo map — all repos share one graph")
	}
	if strings.Contains(source, "GraphLister") {
		t.Error("codedb.go should not have GraphLister — repos are listed via query")
	}
}

// TestNoClientCalls_Anywhere verifies no production source files use
// db.Client() or .Client() to access the go-ormql client directly.
func TestNoClientCalls_Anywhere(t *testing.T) {
	files := []struct {
		path string
		name string
	}{
		{"../../internal/indexer/indexer.go", "indexer.go"},
		{"../../internal/analysis/analyzer.go", "analyzer.go"},
		{"../../internal/mcp/tools.go", "tools.go"},
		{"../../internal/repl/commands.go", "commands.go"},
	}
	for _, f := range files {
		data, err := os.ReadFile(f.path)
		if err != nil {
			t.Errorf("failed to read %s: %v", f.name, err)
			continue
		}
		source := string(data)
		if strings.Contains(source, ".Client()") {
			t.Errorf("%s should not contain .Client() — use ForRepo", f.name)
		}
	}
}
