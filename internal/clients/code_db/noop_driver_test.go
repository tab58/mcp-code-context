package codedb

import (
	"context"
	"errors"

	"github.com/tab58/go-ormql/pkg/cypher"
	"github.com/tab58/go-ormql/pkg/driver"
)

// noopDriver is a minimal driver.Driver implementation for testing.
type noopDriver struct{}

func (d *noopDriver) Execute(_ context.Context, _ cypher.Statement) (driver.Result, error) {
	return driver.Result{}, nil
}

func (d *noopDriver) ExecuteWrite(_ context.Context, _ cypher.Statement) (driver.Result, error) {
	return driver.Result{}, nil
}

func (d *noopDriver) BeginTx(_ context.Context) (driver.Transaction, error) {
	return nil, errors.New("noopDriver: transactions not supported")
}

func (d *noopDriver) Close(_ context.Context) error {
	return nil
}

// noopDriverInstance returns a no-op driver for unit testing.
func noopDriverInstance() driver.Driver {
	return &noopDriver{}
}

// failWriteDriver is a driver that fails on ExecuteWrite (for testing CreateIndexes error).
type failWriteDriver struct {
	noopDriver
}

func (d *failWriteDriver) ExecuteWrite(_ context.Context, _ cypher.Statement) (driver.Result, error) {
	return driver.Result{}, errors.New("failWriteDriver: write failed")
}

func failWriteDriverInstance() driver.Driver {
	return &failWriteDriver{}
}
