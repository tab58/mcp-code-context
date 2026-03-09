package indexer

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	codedb "github.com/tab58/code-context/internal/clients/code_db"
	"github.com/tab58/code-context/internal/config"
	"github.com/tab58/go-ormql/pkg/client"
	"github.com/tab58/go-ormql/pkg/cypher"
	"github.com/tab58/go-ormql/pkg/driver"
)

// --- Test tooling: recording driver ---

// recordingDriver is a mock driver.Driver that records all Execute and
// ExecuteWrite calls for verification in tests.
type recordingDriver struct {
	executeCalls      []recordedCall
	executeWriteCalls []recordedCall
	// executeResult is returned by Execute calls.
	executeResult driver.Result
	// executeWriteResult is returned by ExecuteWrite calls.
	executeWriteResult driver.Result
	// executeErr is returned by Execute calls when non-nil.
	executeErr error
	// executeWriteErr is returned by ExecuteWrite calls when non-nil.
	executeWriteErr error
}

type recordedCall struct {
	Query  string
	Params map[string]any
}

func (d *recordingDriver) Execute(_ context.Context, stmt cypher.Statement) (driver.Result, error) {
	d.executeCalls = append(d.executeCalls, recordedCall{Query: stmt.Query, Params: stmt.Params})
	return d.executeResult, d.executeErr
}

func (d *recordingDriver) ExecuteWrite(_ context.Context, stmt cypher.Statement) (driver.Result, error) {
	d.executeWriteCalls = append(d.executeWriteCalls, recordedCall{Query: stmt.Query, Params: stmt.Params})
	return d.executeWriteResult, d.executeWriteErr
}

func (d *recordingDriver) BeginTx(_ context.Context) (driver.Transaction, error) {
	return nil, nil
}

func (d *recordingDriver) Close(_ context.Context) error {
	return nil
}

// failAfterNDriver wraps recordingDriver but fails ExecuteWrite calls after
// the first N succeed. This allows NewCodeDB's CreateIndexes calls to pass
// while subsequent writes (from indexer methods) fail.
type failAfterNDriver struct {
	recordingDriver
	allowedWrites int
	writeCount    int
	failErr       error
}

func (d *failAfterNDriver) ExecuteWrite(ctx context.Context, stmt cypher.Statement) (driver.Result, error) {
	d.writeCount++
	if d.writeCount > d.allowedWrites {
		return driver.Result{}, d.failErr
	}
	return d.recordingDriver.ExecuteWrite(ctx, stmt)
}

// newFailAfterNIndexer creates an Indexer with a driver that allows the first N
// ExecuteWrite calls (for CreateIndexes) then fails all subsequent writes.
func newFailAfterNIndexer(t *testing.T, allowedWrites int, failErr error) *Indexer {
	t.Helper()
	drv := &failAfterNDriver{
		allowedWrites: allowedWrites,
		failErr:       failErr,
	}
	ctx := context.Background()
	db, err := codedb.NewCodeDB(ctx, config.FalkorDBConfig{
		Host:     "localhost",
		Port:     6379,
	}, codedb.WithDriver(drv))
	if err != nil {
		t.Fatalf("NewCodeDB with failAfterN driver failed: %v", err)
	}
	t.Cleanup(func() { db.Close(ctx) })
	return NewIndexer(db)
}

// newTestIndexerWithRecorder creates an Indexer backed by a recording driver
// and returns both the indexer and the recorder for verification.
func newTestIndexerWithRecorder(t *testing.T) (*Indexer, *recordingDriver) {
	t.Helper()
	rec := &recordingDriver{}
	ctx := context.Background()
	db, err := codedb.NewCodeDB(ctx, config.FalkorDBConfig{
		Host:     "localhost",
		Port:     6379,
	}, codedb.WithDriver(rec))
	if err != nil {
		t.Fatalf("NewCodeDB with recording driver failed: %v", err)
	}
	t.Cleanup(func() { db.Close(ctx) })
	return NewIndexer(db), rec
}

// testClient returns a *client.Client for direct method testing via ForRepo.
func testClient(t *testing.T, idx *Indexer, graphName string) *client.Client {
	t.Helper()
	c, err := idx.db.ForRepo(context.Background(), graphName)
	if err != nil {
		t.Fatalf("ForRepo returned error: %v", err)
	}
	return c
}

// createPersistenceTestRepo creates a temp directory with known structure:
//
//	repo/
//	  main.go (3 lines)
//	  src/
//	    util.go (2 lines)
func createPersistenceTestRepo(t *testing.T) string {
	t.Helper()
	root := t.TempDir()

	srcDir := filepath.Join(root, "src")
	if err := os.MkdirAll(srcDir, 0o755); err != nil {
		t.Fatalf("failed to create src dir: %v", err)
	}

	files := map[string]string{
		filepath.Join(root, "main.go"):       "package main\n\nfunc main() {}\n",
		filepath.Join(root, "src", "util.go"): "package main\n\nfunc helper() {}\n",
	}
	for path, content := range files {
		if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
			t.Fatalf("failed to write %s: %v", path, err)
		}
	}

	return root
}
