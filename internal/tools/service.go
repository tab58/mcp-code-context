package mcp

import (
	"github.com/tab58/code-context/internal/analysis"
	codedb "github.com/tab58/code-context/internal/clients/code_db"
)

// Manager holds the business logic dependencies for code analysis operations.
// It is protocol-agnostic — no MCP types are referenced.
type Manager struct {
	db *codedb.CodeDB
}

// NewService creates a new Service with the given dependencies.
func NewManager(db *codedb.CodeDB, analyzer *analysis.Analyzer) *Manager {
	return &Manager{
		db: db,
	}
}
