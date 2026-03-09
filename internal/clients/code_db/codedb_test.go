package codedb

import (
	"context"
	"testing"

	"github.com/tab58/code-context/internal/config"
	"github.com/tab58/go-ormql/pkg/client"
)

// validConfig returns a FalkorDBConfig with all fields populated for testing.
func validConfig() config.FalkorDBConfig {
	return config.FalkorDBConfig{
		Host:      "localhost",
		Port:      6379,
		Password:  "testpassword",
		GraphName: "test-graph",
	}
}

// newTestCodeDB creates a CodeDB with a no-op driver for testing.
func newTestCodeDB(t *testing.T) *CodeDB {
	t.Helper()
	ctx := context.Background()
	db, err := NewCodeDB(ctx, validConfig(),
		WithDriver(noopDriverInstance()),
	)
	if err != nil {
		t.Fatalf("NewCodeDB returned unexpected error: %v", err)
	}
	return db
}

// --- CodeDB with FalkorDB interface tests ---

// TestNewCodeDB_AcceptsFalkorDBConfig verifies that NewCodeDB accepts a
// config.FalkorDBConfig struct.
// Expected result: Non-nil *CodeDB when given valid FalkorDBConfig.
func TestNewCodeDB_AcceptsFalkorDBConfig(t *testing.T) {
	db := newTestCodeDB(t)
	if db == nil {
		t.Error("NewCodeDB returned nil *CodeDB, expected non-nil")
	}
}

// TestNewCodeDB_NilContext verifies that NewCodeDB returns an error when
// given a nil context (invalid input).
// Expected result: Non-nil error.
func TestNewCodeDB_NilContext(t *testing.T) {
	//nolint:staticcheck // intentionally passing nil context to test error handling
	_, err := NewCodeDB(nil, validConfig())
	if err == nil {
		t.Error("NewCodeDB with nil context should return error, got nil")
	}
}

// TestNewCodeDB_EmptyHost verifies that NewCodeDB returns an error when
// the Host field is empty.
// Expected result: Non-nil error.
func TestNewCodeDB_EmptyHost(t *testing.T) {
	ctx := context.Background()
	cfg := validConfig()
	cfg.Host = ""
	_, err := NewCodeDB(ctx, cfg)
	if err == nil {
		t.Error("NewCodeDB with empty Host should return error, got nil")
	}
}

// TestNewCodeDB_InvalidPort verifies that NewCodeDB returns an error when
// the Port is zero or negative.
// Expected result: Non-nil error.
func TestNewCodeDB_InvalidPort(t *testing.T) {
	tests := []struct {
		name string
		port int
	}{
		{"zero", 0},
		{"negative", -1},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			cfg := validConfig()
			cfg.Port = tt.port
			_, err := NewCodeDB(ctx, cfg)
			if err == nil {
				t.Errorf("NewCodeDB with port=%d should return error, got nil", tt.port)
			}
		})
	}
}

// TestNewCodeDB_CancelledContext verifies that NewCodeDB returns an error when
// the context is already cancelled before connection is attempted.
// Expected result: Non-nil error (context.Canceled).
func TestNewCodeDB_CancelledContext(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err := NewCodeDB(ctx, validConfig())
	if err == nil {
		t.Error("NewCodeDB with cancelled context should return error, got nil")
	}
}

// TestNewCodeDB_PasswordOptional verifies that NewCodeDB succeeds when
// Password is empty (FalkorDB supports no-auth for local dev).
// Expected result: No error, valid *CodeDB returned.
func TestNewCodeDB_PasswordOptional(t *testing.T) {
	ctx := context.Background()
	cfg := validConfig()
	cfg.Password = ""
	db, err := NewCodeDB(ctx, cfg, WithDriver(noopDriverInstance()))
	if err != nil {
		t.Fatalf("NewCodeDB should not require Password, got error: %v", err)
	}
	if db == nil {
		t.Error("NewCodeDB returned nil *CodeDB, expected non-nil")
	}
}

// TestForRepo_ReturnsTypedClient verifies that ForRepo() returns a *client.Client
// (not any or interface{}). This is a type-system test.
// Expected result: Non-nil *client.Client after ForRepo call.
func TestForRepo_ReturnsTypedClient(t *testing.T) {
	db := newTestCodeDB(t)

	// This assignment proves the return type is *client.Client at compile time
	var c *client.Client
	var err error
	c, err = db.ForRepo(context.Background(), "test-repo")
	if err != nil {
		t.Fatalf("ForRepo returned error: %v", err)
	}
	if c == nil {
		t.Error("ForRepo returned nil, expected non-nil *client.Client")
	}
}

// TestClose_ReturnsNilOnSuccess verifies that Close() returns nil error
// on a properly constructed CodeDB.
// Expected result: nil error.
func TestClose_ReturnsNilOnSuccess(t *testing.T) {
	db := newTestCodeDB(t)

	err := db.Close(context.Background())
	if err != nil {
		t.Errorf("Close() returned error: %v, expected nil", err)
	}
}

// TestClose_Idempotent verifies that calling Close() multiple times does not
// error or panic.
// Expected result: Both Close() calls return nil error.
func TestClose_Idempotent(t *testing.T) {
	db := newTestCodeDB(t)

	if err := db.Close(context.Background()); err != nil {
		t.Errorf("first Close() returned error: %v", err)
	}
	if err := db.Close(context.Background()); err != nil {
		t.Errorf("second Close() returned error: %v, expected idempotent nil", err)
	}
}

// TestForRepo_ErrorAfterClose_Existing verifies that ForRepo returns error
// after Close, preventing use-after-close.
func TestForRepo_ErrorAfterClose_Existing(t *testing.T) {
	db := newTestCodeDB(t)

	_ = db.Close(context.Background())

	_, err := db.ForRepo(context.Background(), "test")
	if err == nil {
		t.Error("ForRepo after Close() should return error")
	}
}

// TestNewCodeDB_ErrorCases is a table-driven test covering all error paths
// for NewCodeDB with config.FalkorDBConfig.
// Expected result: All cases return non-nil error.
func TestNewCodeDB_ErrorCases(t *testing.T) {
	tests := []struct {
		name string
		cfg  config.FalkorDBConfig
	}{
		{
			name: "empty config",
			cfg:  config.FalkorDBConfig{},
		},
		{
			name: "missing host",
			cfg: config.FalkorDBConfig{
				Port:     6379,
				Password: "pass",
			},
		},
		{
			name: "zero port",
			cfg: config.FalkorDBConfig{
				Host:     "localhost",
				Port:     0,
				Password: "pass",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			_, err := NewCodeDB(ctx, tt.cfg)
			if err == nil {
				t.Errorf("NewCodeDB(%+v) should return error, got nil", tt.cfg)
			}
		})
	}
}

// TestWithDriver_InjectedDriverUsed verifies that WithDriver option injects
// a custom driver and bypasses the real FalkorDB connection path.
func TestWithDriver_InjectedDriverUsed(t *testing.T) {
	ctx := context.Background()
	drv := noopDriverInstance()
	db, err := NewCodeDB(ctx, validConfig(), WithDriver(drv))
	if err != nil {
		t.Fatalf("NewCodeDB with injected driver returned error: %v", err)
	}
	// Verify ForRepo is usable with injected driver
	c, ferr := db.ForRepo(ctx, "test-graph")
	if ferr != nil {
		t.Fatalf("ForRepo returned error: %v", ferr)
	}
	if c == nil {
		t.Error("ForRepo returned nil with injected driver")
	}
	// Verify close works
	if err := db.Close(ctx); err != nil {
		t.Errorf("Close() returned error: %v", err)
	}
}

// TestClose_NoSharedClient verifies that Close handles the case where
// no ForRepo call was ever made (shared client is nil).
func TestClose_NoSharedClient(t *testing.T) {
	ctx := context.Background()
	db, err := NewCodeDB(ctx, validConfig(), WithDriver(noopDriverInstance()))
	if err != nil {
		t.Fatalf("NewCodeDB returned error: %v", err)
	}
	err = db.Close(ctx)
	if err != nil {
		t.Errorf("Close() with no shared client returned error: %v", err)
	}
}

// TestForRepo_CreateIndexesNoOp verifies that ForRepo succeeds even with a
// fail-write driver because CreateIndexes is a no-op (no vector indexes).
func TestForRepo_CreateIndexesNoOp(t *testing.T) {
	ctx := context.Background()
	db, err := NewCodeDB(ctx, validConfig(), WithDriver(failWriteDriverInstance()))
	if err != nil {
		t.Fatalf("NewCodeDB should succeed (no boot connection): %v", err)
	}
	_, err = db.ForRepo(ctx, "test-graph")
	if err != nil {
		t.Errorf("ForRepo should succeed with no-op CreateIndexes, got: %v", err)
	}
}
