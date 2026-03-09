package mcp

import (
	"testing"
)

// === Task 2 + 11: MCP transport switch tests ===
//
// These tests verify the transport switch from stdio to streamable HTTP.
// They will compile against the current API but assert the new behavior.

// TestNewServer_CreatesStreamableHTTP verifies that NewServer creates a Server
// with the StreamableHTTPServer field populated (http field).
// Expected result: Server has an http field that is non-nil.
func TestNewServer_CreatesStreamableHTTP(t *testing.T) {
	db := newTestDB(t)
	s := NewServer(db, nil, nil)
	if s == nil {
		t.Fatal("NewServer returned nil")
	}
	// The http field should exist and be non-nil after transport switch
	if s.http == nil {
		t.Error("NewServer did not create StreamableHTTPServer — transport not switched to HTTP")
	}
}
