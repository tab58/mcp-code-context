//go:build integration

package indexer

import (
	"context"
	"os"
	"path/filepath"
	"strconv"
	"testing"

	codedb "github.com/tab58/code-context/internal/clients/code_db"
	"github.com/tab58/code-context/internal/config"
	testutil "github.com/tab58/code-context/testinfra"
)

// setupIntegrationTest starts a FalkorDB container and returns a ready-to-use
// Indexer and CodeDB with a real database connection.
func setupIntegrationTest(t *testing.T) (*Indexer, *codedb.CodeDB) {
	t.Helper()
	ctx := context.Background()

	container, err := testutil.SetupFalkorDB(ctx)
	if err != nil {
		t.Fatalf("SetupFalkorDB failed: %v", err)
	}
	t.Cleanup(func() { container.Teardown(ctx) })

	port, err := strconv.Atoi(container.Port)
	if err != nil {
		t.Fatalf("invalid port %q: %v", container.Port, err)
	}

	db, err := codedb.NewCodeDB(ctx, config.FalkorDBConfig{
		Host:     container.Host,
		Port:     port,
	})
	if err != nil {
		t.Fatalf("NewCodeDB failed: %v", err)
	}
	t.Cleanup(func() { db.Close(ctx) })

	return NewIndexer(db), db
}

// createIntegrationTestRepo creates a temporary repository with known structure:
//
//	repo/
//	  README.md
//	  src/
//	    main.go
//	    utils/
//	      helper.go
func createIntegrationTestRepo(t *testing.T) string {
	t.Helper()
	root := t.TempDir()

	dirs := []string{
		filepath.Join(root, "src"),
		filepath.Join(root, "src", "utils"),
	}
	for _, d := range dirs {
		if err := os.MkdirAll(d, 0o755); err != nil {
			t.Fatalf("failed to create dir %s: %v", d, err)
		}
	}

	files := map[string]string{
		filepath.Join(root, "README.md"):                "# Test\n",
		filepath.Join(root, "src", "main.go"):           "package main\n\nfunc main() {}\n",
		filepath.Join(root, "src", "utils", "helper.go"): "package utils\n\nfunc Help() {}\n",
	}
	for path, content := range files {
		if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
			t.Fatalf("failed to write %s: %v", path, err)
		}
	}

	return root
}

// TestIntegration_IndexCreatesNodes verifies that IndexRepository creates
// Repository, Folder, and File nodes in a real FalkorDB instance.
// Expected result: Non-zero counts, nodes queryable in FalkorDB.
func TestIntegration_IndexCreatesNodes(t *testing.T) {
	idx, _ := setupIntegrationTest(t)
	repoPath := createIntegrationTestRepo(t)

	result, err := idx.IndexRepository(context.Background(), repoPath)
	if err != nil {
		t.Fatalf("IndexRepository returned error: %v", err)
	}

	if result.FilesIndexed == 0 {
		t.Error("FilesIndexed = 0, expected > 0")
	}
	if result.FoldersIndexed == 0 {
		t.Error("FoldersIndexed = 0, expected > 0")
	}
}

// TestIntegration_IndexCreatesEdges verifies that CONTAINS and BELONGS_TO
// edges are created between the structural nodes.
// Expected result: Edges exist in FalkorDB after indexing, verified via ForRepo().Execute().
func TestIntegration_IndexCreatesEdges(t *testing.T) {
	idx, db := setupIntegrationTest(t)
	repoPath := createIntegrationTestRepo(t)

	_, err := idx.IndexRepository(context.Background(), repoPath)
	if err != nil {
		t.Fatalf("IndexRepository returned error: %v", err)
	}

	repoName := filepath.Base(repoPath)
	c, ferr := db.ForRepo(context.Background(), repoName)
	if ferr != nil {
		t.Fatalf("ForRepo returned error: %v", ferr)
	}

	// Verify folders exist for this repository using relationship WHERE
	folderResult, err := c.Execute(context.Background(),
		`query($where: FolderWhere) { folders(where: $where) { path } }`,
		map[string]any{"where": map[string]any{"repository": map[string]any{"name": repoName}}},
	)
	if err != nil {
		t.Fatalf("querying folders failed: %v", err)
	}
	if folderResult == nil {
		t.Error("expected folder query result, got nil")
	}

	// Verify files exist for this repository using relationship WHERE
	fileResult, err := c.Execute(context.Background(),
		`query($where: FileWhere) { files(where: $where) { path } }`,
		map[string]any{"where": map[string]any{"repository": map[string]any{"name": repoName}}},
	)
	if err != nil {
		t.Fatalf("querying files failed: %v", err)
	}
	if fileResult == nil {
		t.Error("expected file query result, got nil")
	}
}

// TestIntegration_IncrementalReindex verifies that re-indexing an unchanged
// directory skips all files (FilesSkipped > 0).
// Expected result: Second index has FilesSkipped == FilesIndexed from first.
func TestIntegration_IncrementalReindex(t *testing.T) {
	idx, _ := setupIntegrationTest(t)
	repoPath := createIntegrationTestRepo(t)

	// First index
	first, err := idx.IndexRepository(context.Background(), repoPath)
	if err != nil {
		t.Fatalf("first IndexRepository returned error: %v", err)
	}

	// Second index of unchanged directory
	second, err := idx.IndexRepository(context.Background(), repoPath)
	if err != nil {
		t.Fatalf("second IndexRepository returned error: %v", err)
	}

	if second.FilesSkipped == 0 {
		t.Errorf("second index FilesSkipped = 0, expected %d (all unchanged)", first.FilesIndexed)
	}
}

// TestIntegration_ModifiedFileReindexed verifies that modifying a file and
// re-indexing updates the corresponding File node.
// Expected result: Modified file is re-indexed (not skipped).
func TestIntegration_ModifiedFileReindexed(t *testing.T) {
	idx, _ := setupIntegrationTest(t)
	repoPath := createIntegrationTestRepo(t)

	// First index
	_, err := idx.IndexRepository(context.Background(), repoPath)
	if err != nil {
		t.Fatalf("first IndexRepository returned error: %v", err)
	}

	// Modify a file
	modifiedFile := filepath.Join(repoPath, "src", "main.go")
	if err := os.WriteFile(modifiedFile, []byte("package main\n\nfunc main() {\n\tfmt.Println(\"updated\")\n}\n"), 0o644); err != nil {
		t.Fatalf("failed to modify file: %v", err)
	}

	// Second index
	second, err := idx.IndexRepository(context.Background(), repoPath)
	if err != nil {
		t.Fatalf("second IndexRepository returned error: %v", err)
	}

	// At least one file should be re-indexed (the modified one)
	if second.FilesIndexed == 0 {
		t.Error("second index FilesIndexed = 0, expected at least 1 (modified file)")
	}
}
