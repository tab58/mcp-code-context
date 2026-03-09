package repl

import (
	"context"
	"errors"
	"testing"

	codedb "github.com/tab58/code-context/internal/clients/code_db"
	"github.com/tab58/code-context/internal/config"
	"github.com/tab58/go-ormql/pkg/cypher"
	"github.com/tab58/go-ormql/pkg/driver"
)

// noopREPLDriver is a minimal driver.Driver that returns empty results.
type noopREPLDriver struct{}

func (d *noopREPLDriver) Execute(_ context.Context, _ cypher.Statement) (driver.Result, error) {
	return driver.Result{}, nil
}

func (d *noopREPLDriver) ExecuteWrite(_ context.Context, _ cypher.Statement) (driver.Result, error) {
	return driver.Result{}, nil
}

func (d *noopREPLDriver) BeginTx(_ context.Context) (driver.Transaction, error) {
	return nil, errors.New("noopREPLDriver: transactions not supported")
}

func (d *noopREPLDriver) Close(_ context.Context) error {
	return nil
}

// responseREPLDriver returns pre-configured responses for Execute calls.
type responseREPLDriver struct {
	noopREPLDriver
	readResponses []driver.Result
	readIdx       int
}

func (d *responseREPLDriver) Execute(_ context.Context, _ cypher.Statement) (driver.Result, error) {
	if d.readIdx < len(d.readResponses) {
		r := d.readResponses[d.readIdx]
		d.readIdx++
		return r, nil
	}
	return driver.Result{}, nil
}

// makeREPLResult builds a driver.Result wrapping data for Client.Execute().
func makeREPLResult(data map[string]any) driver.Result {
	return driver.Result{
		Records: []driver.Record{
			{Values: map[string]any{"data": data}},
		},
	}
}

// newTestREPLDB creates a CodeDB with a noop driver for unit tests.
func newTestREPLDB(t *testing.T) *codedb.CodeDB {
	t.Helper()
	ctx := context.Background()
	db, err := codedb.NewCodeDB(ctx, config.FalkorDBConfig{
		Host: "localhost",
		Port: 6379,
	},
		codedb.WithDriver(&noopREPLDriver{}),
	)
	if err != nil {
		t.Fatalf("NewCodeDB with noop driver failed: %v", err)
	}
	t.Cleanup(func() { db.Close(ctx) })
	return db
}

// newTestREPLDBWithResponses creates a CodeDB with a response driver.
func newTestREPLDBWithResponses(t *testing.T, readResponses []driver.Result) *codedb.CodeDB {
	t.Helper()
	drv := &responseREPLDriver{readResponses: readResponses}
	ctx := context.Background()
	db, err := codedb.NewCodeDB(ctx, config.FalkorDBConfig{
		Host: "localhost",
		Port: 6379,
	},
		codedb.WithDriver(drv),
	)
	if err != nil {
		t.Fatalf("NewCodeDB with response driver failed: %v", err)
	}
	t.Cleanup(func() { db.Close(ctx) })
	return db
}
