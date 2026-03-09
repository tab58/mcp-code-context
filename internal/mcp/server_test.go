package mcp

import (
	"context"
	"testing"

	"github.com/tab58/code-context/internal/analysis"
)

// === Task 3: Server skeleton ===
//
// Server struct holds pipeline components and exposes NewServer + Serve.
// NewServer registers 3 tool definitions with mcp-go.
// Serve starts HTTP transport and blocks until ctx is cancelled.

// TestNewServer_ReturnsNonNil verifies that NewServer creates a valid Server.
// Expected result: NewServer returns a non-nil *Server.
func TestNewServer_ReturnsNonNil(t *testing.T) {
	db := newTestDB(t)
	s := NewServer(db, nil, nil)
	if s == nil {
		t.Fatal("NewServer returned nil, expected a non-nil *Server")
	}
}

// TestNewServer_StoresDB verifies that NewServer stores the CodeDB reference.
// Expected result: Server.db is the same as the passed CodeDB.
func TestNewServer_StoresDB(t *testing.T) {
	db := newTestDB(t)
	s := NewServer(db, nil, nil)
	if s == nil {
		t.Fatal("NewServer returned nil")
	}
	if s.db != db {
		t.Error("NewServer did not store the CodeDB reference")
	}
}

// TestNewServer_StoresAllComponents verifies that NewServer stores all
// pipeline components (indexer, analyzer).
// Expected result: All fields are set.
func TestNewServer_StoresAllComponents(t *testing.T) {
	db := newTestDB(t)
	analyzer := analysis.NewAnalyzer(nil, nil)

	s := NewServer(db, nil, analyzer)
	if s == nil {
		t.Fatal("NewServer returned nil")
	}
	if s.analyzer != analyzer {
		t.Error("NewServer did not store the Analyzer reference")
	}
}

// TestServe_ReturnsOnCancel verifies that Serve blocks until ctx is cancelled,
// then returns without error (or returns an error if stdio is not connected).
// Expected result: Serve returns when context is cancelled.
func TestServe_ReturnsOnCancel(t *testing.T) {
	db := newTestDB(t)
	s := NewServer(db, nil, nil)
	if s == nil {
		t.Fatal("NewServer returned nil")
	}

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // cancel immediately

	err := s.Serve(ctx, ":0")
	// Serve should return when ctx is cancelled, possibly with a nil error
	// or a context.Canceled error, but NOT "not implemented".
	if err != nil && err.Error() == "not implemented" {
		t.Error("Serve returned 'not implemented' — should be wired to mcp-go HTTP transport")
	}
}
