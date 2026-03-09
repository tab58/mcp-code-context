package codedb

import (
	"context"
	"os"
	"strings"
	"testing"

	"github.com/tab58/go-ormql/pkg/client"
)

// --- Shared graph isolation tests ---

// TestForRepo_ReturnsClient verifies that ForRepo returns a non-nil
// *client.Client for a valid repo name.
// Expected result: Non-nil *client.Client.
func TestForRepo_ReturnsClient(t *testing.T) {
	db := newTestCodeDB(t)
	ctx := context.Background()

	c, err := db.ForRepo(ctx, "my-repo")
	if err != nil {
		t.Fatalf("ForRepo returned error: %v", err)
	}
	var typed *client.Client = c
	if typed == nil {
		t.Error("ForRepo returned nil *client.Client, expected non-nil")
	}
}

// TestForRepo_CachesSameGraph verifies that ForRepo returns the same
// *client.Client for repeated calls with the same repo name.
// Expected result: Same pointer for both calls.
func TestForRepo_CachesSameGraph(t *testing.T) {
	db := newTestCodeDB(t)
	ctx := context.Background()

	c1, err := db.ForRepo(ctx, "my-repo")
	if err != nil {
		t.Fatalf("first ForRepo returned error: %v", err)
	}
	c2, err := db.ForRepo(ctx, "my-repo")
	if err != nil {
		t.Fatalf("second ForRepo returned error: %v", err)
	}
	if c1 != c2 {
		t.Error("ForRepo should return cached client for the same repo name")
	}
}

// TestForRepo_SharedGraph verifies that ForRepo returns the same
// *client.Client for different repo names (single shared graph).
// Expected result: Same pointer for different names.
func TestForRepo_SharedGraph(t *testing.T) {
	db := newTestCodeDB(t)
	ctx := context.Background()

	c1, err := db.ForRepo(ctx, "repo-a")
	if err != nil {
		t.Fatalf("ForRepo repo-a returned error: %v", err)
	}
	c2, err := db.ForRepo(ctx, "repo-b")
	if err != nil {
		t.Fatalf("ForRepo repo-b returned error: %v", err)
	}
	if c1 != c2 {
		t.Error("ForRepo should return the same shared client for all repo names")
	}
}

// TestForRepo_ErrorAfterClose verifies that ForRepo returns an error
// when called after Close.
// Expected result: Non-nil error.
func TestForRepo_ErrorAfterClose(t *testing.T) {
	db := newTestCodeDB(t)
	ctx := context.Background()

	_ = db.Close(ctx)

	_, err := db.ForRepo(ctx, "my-repo")
	if err == nil {
		t.Error("ForRepo after Close should return error, got nil")
	}
}

// TestListRepos_Empty verifies that ListRepos returns empty when
// no repositories have been indexed.
// Expected result: Empty slice.
func TestListRepos_Empty(t *testing.T) {
	db := newTestCodeDB(t)

	repos, err := db.ListRepos(context.Background())
	if err != nil {
		t.Fatalf("ListRepos returned error: %v", err)
	}
	if len(repos) != 0 {
		t.Errorf("ListRepos on fresh CodeDB should return empty, got %v", repos)
	}
}

// TestClose_ClosesSharedDriver verifies that Close shuts down the shared
// driver. After close, ForRepo should return error.
func TestClose_ClosesSharedDriver(t *testing.T) {
	db := newTestCodeDB(t)
	ctx := context.Background()

	_, _ = db.ForRepo(ctx, "repo-1")

	err := db.Close(ctx)
	if err != nil {
		t.Errorf("Close returned error: %v", err)
	}

	// After close, ForRepo should fail
	_, err = db.ForRepo(ctx, "repo-1")
	if err == nil {
		t.Error("ForRepo after Close should return error")
	}
}

// TestNewCodeDB_NoConnectionAtBoot verifies that NewCodeDB does NOT
// create a driver or call CreateIndexes at construction time.
// Expected result: No error.
func TestNewCodeDB_NoConnectionAtBoot(t *testing.T) {
	ctx := context.Background()
	// Use a failWriteDriver — if NewCodeDB tried to create indexes at boot,
	// it would fail. With lazy init, it should succeed.
	db, err := NewCodeDB(ctx, validConfig(),
		WithDriver(failWriteDriverInstance()),
	)
	if err != nil {
		t.Fatalf("NewCodeDB should not connect at boot, got error: %v", err)
	}
	if db == nil {
		t.Fatal("NewCodeDB returned nil")
	}
}

// TestWithDriver_WorksWithForRepo verifies that the injected driver via
// WithDriver is used by ForRepo.
// Expected result: ForRepo succeeds with injected driver.
func TestWithDriver_WorksWithForRepo(t *testing.T) {
	ctx := context.Background()
	db, err := NewCodeDB(ctx, validConfig(), WithDriver(noopDriverInstance()))
	if err != nil {
		t.Fatalf("NewCodeDB returned error: %v", err)
	}

	c, err := db.ForRepo(ctx, "test-graph")
	if err != nil {
		t.Fatalf("ForRepo with injected driver returned error: %v", err)
	}
	if c == nil {
		t.Error("ForRepo returned nil client with injected driver")
	}
}

// TestClientMethod_Removed verifies that the Client() method no longer
// exists on CodeDB. This is a source inspection test.
// Expected result: No "func (db *CodeDB) Client()" in codedb.go.
func TestClientMethod_Removed(t *testing.T) {
	data, err := readSourceFile("codedb.go")
	if err != nil {
		t.Fatalf("failed to read codedb.go: %v", err)
	}
	source := string(data)
	if strings.Contains(source, "func (db *CodeDB) Client()") {
		t.Error("Client() method should be removed from CodeDB — use ForRepo instead")
	}
}

// TestCodeDB_HasSharedClient verifies that CodeDB has a shared field
// for the single graph client.
// Expected result: "shared" field exists in codedb.go.
func TestCodeDB_HasSharedClient(t *testing.T) {
	data, err := readSourceFile("codedb.go")
	if err != nil {
		t.Fatalf("failed to read codedb.go: %v", err)
	}
	source := string(data)
	if !strings.Contains(source, "shared") {
		t.Error("CodeDB should have a 'shared' field for the single graph client")
	}
}

// TestCodeDB_HasSharedClientType verifies that codedb.go defines a sharedClient
// struct for the driver+client pair.
// Expected result: "sharedClient" type exists in codedb.go.
func TestCodeDB_HasSharedClientType(t *testing.T) {
	data, err := readSourceFile("codedb.go")
	if err != nil {
		t.Fatalf("failed to read codedb.go: %v", err)
	}
	source := string(data)
	if !strings.Contains(source, "sharedClient") {
		t.Error("codedb.go should define a sharedClient struct")
	}
}

// TestCodeDB_NoPerRepoMap verifies that CodeDB no longer has a per-repo
// map field (repos map).
// Expected result: No "repos" map field in codedb.go.
func TestCodeDB_NoPerRepoMap(t *testing.T) {
	data, err := readSourceFile("codedb.go")
	if err != nil {
		t.Fatalf("failed to read codedb.go: %v", err)
	}
	source := string(data)
	if strings.Contains(source, "repos  map[string]") {
		t.Error("CodeDB should not have a 'repos' map — all repos share one graph")
	}
}

// readSourceFile reads a source file from the current package directory.
func readSourceFile(name string) ([]byte, error) {
	return os.ReadFile(name)
}
