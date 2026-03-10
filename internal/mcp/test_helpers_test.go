package mcp

import (
	"context"
	"errors"
	"testing"

	codedb "github.com/tab58/code-context/internal/clients/code_db"
	"github.com/tab58/code-context/internal/config"
	"github.com/tab58/code-context/internal/indexer"
	"github.com/tab58/go-ormql/pkg/cypher"
	"github.com/tab58/go-ormql/pkg/driver"
)

// --- Test tooling: noop driver ---

// noopMCPDriver is a minimal driver.Driver that returns empty results.
type noopMCPDriver struct{}

func (d *noopMCPDriver) Execute(_ context.Context, _ cypher.Statement) (driver.Result, error) {
	return driver.Result{}, nil
}

func (d *noopMCPDriver) ExecuteWrite(_ context.Context, _ cypher.Statement) (driver.Result, error) {
	return driver.Result{}, nil
}

func (d *noopMCPDriver) BeginTx(_ context.Context) (driver.Transaction, error) {
	return nil, errors.New("noopMCPDriver: transactions not supported")
}

func (d *noopMCPDriver) Close(_ context.Context) error {
	return nil
}

// --- Test tooling: response driver ---

// responseDriver returns pre-configured responses for Execute calls (reads)
// and ExecuteWrite calls (writes). Responses are returned in order.
type responseDriver struct {
	noopMCPDriver
	readResponses  []driver.Result
	writeResponses []driver.Result
	readIdx        int
	writeIdx       int
	readCalls      []recordedMCPCall
	writeCalls     []recordedMCPCall
}

type recordedMCPCall struct {
	Query  string
	Params map[string]any
}

func (d *responseDriver) Execute(_ context.Context, stmt cypher.Statement) (driver.Result, error) {
	d.readCalls = append(d.readCalls, recordedMCPCall{Query: stmt.Query, Params: stmt.Params})
	if d.readIdx < len(d.readResponses) {
		r := d.readResponses[d.readIdx]
		d.readIdx++
		return r, nil
	}
	return driver.Result{}, nil
}

func (d *responseDriver) ExecuteWrite(_ context.Context, stmt cypher.Statement) (driver.Result, error) {
	d.writeCalls = append(d.writeCalls, recordedMCPCall{Query: stmt.Query, Params: stmt.Params})
	if d.writeIdx < len(d.writeResponses) {
		r := d.writeResponses[d.writeIdx]
		d.writeIdx++
		return r, nil
	}
	return driver.Result{}, nil
}

// makeResult builds a driver.Result that wraps data in records[0].Values["data"].
// This mirrors how go-ormql Client.Execute() extracts response data.
func makeResult(data map[string]any) driver.Result {
	return driver.Result{
		Records: []driver.Record{
			{Values: map[string]any{"data": data}},
		},
	}
}

// --- Test helpers ---

// newTestDB creates a CodeDB with a noop driver for unit tests.
func newTestDB(t *testing.T) *codedb.CodeDB {
	t.Helper()
	ctx := context.Background()
	db, err := codedb.NewCodeDB(ctx, config.FalkorDBConfig{
		Host: "localhost",
		Port: 6379,
	}, codedb.WithDriver(&noopMCPDriver{}))
	if err != nil {
		t.Fatalf("NewCodeDB with noop driver failed: %v", err)
	}
	t.Cleanup(func() { db.Close(ctx) })
	return db
}

// newTestDBWithResponses creates a CodeDB with a response driver that returns
// canned responses for Execute (read) calls.
func newTestDBWithResponses(t *testing.T, readResponses []driver.Result) (*codedb.CodeDB, *responseDriver) {
	t.Helper()
	drv := &responseDriver{readResponses: readResponses}
	ctx := context.Background()
	db, err := codedb.NewCodeDB(ctx, config.FalkorDBConfig{
		Host: "localhost",
		Port: 6379,
	}, codedb.WithDriver(drv))
	if err != nil {
		t.Fatalf("NewCodeDB with response driver failed: %v", err)
	}
	t.Cleanup(func() { db.Close(ctx) })
	return db, drv
}

// newTestServer creates a Server backed by a noop driver.
func newTestServer(t *testing.T) *Server {
	t.Helper()
	db := newTestDB(t)
	return &Server{db: db}
}

// newTestServerWithIndexer creates a Server backed by a noop driver with an Indexer wired up.
func newTestServerWithIndexer(t *testing.T) *Server {
	t.Helper()
	db := newTestDB(t)
	idx := indexer.NewIndexer(db)
	return &Server{db: db, idx: idx}
}

// newTestServerWithResponses creates a Server backed by a response driver.
func newTestServerWithResponses(t *testing.T, readResponses []driver.Result) (*Server, *responseDriver) {
	t.Helper()
	db, drv := newTestDBWithResponses(t, readResponses)
	return &Server{db: db}, drv
}
