package verify_test

import (
	"strings"
	"testing"
)

// === Task 1: Regenerated go-ormql files — range index removal ===
//
// The upstream go-ormql OOM fix removed range property indexes from
// CreateIndexes because FalkorDB doesn't need them (auto-indexes).
// Vector indexes have also been removed.

// TestIndexesGenHasNoRangeIndexes verifies that indexes_gen.go no longer
// contains a rangeIndexes variable. Range indexes were removed upstream
// because FalkorDB auto-indexes properties.
// Expected result: indexes_gen.go does NOT contain "rangeIndexes".
func TestIndexesGenHasNoRangeIndexes(t *testing.T) {
	content := readProjectFile(t, "internal/clients/code_db/generated/indexes_gen.go")

	if strings.Contains(content, "rangeIndexes") {
		t.Error("indexes_gen.go still contains rangeIndexes variable — should be removed in upstream OOM fix")
	}
}

// TestIndexesGenHasNoRangeIndexStatements verifies that indexes_gen.go
// does not contain any CREATE INDEX FOR range statements.
// Expected result: indexes_gen.go does NOT contain "CREATE INDEX FOR" (without VECTOR).
func TestIndexesGenHasNoRangeIndexStatements(t *testing.T) {
	content := readProjectFile(t, "internal/clients/code_db/generated/indexes_gen.go")

	// Range indexes use "CREATE INDEX FOR" without VECTOR keyword
	lines := strings.Split(content, "\n")
	for _, line := range lines {
		if strings.Contains(line, "CREATE INDEX FOR") && !strings.Contains(line, "VECTOR") {
			t.Errorf("indexes_gen.go contains range index statement: %s", strings.TrimSpace(line))
		}
	}
}

// TestIndexesGenNoVectorIndexes verifies that indexes_gen.go has no
// vector index creation code (embeddings removed from schema).
// Expected result: CreateIndexes is a no-op with no vector index statements.
func TestIndexesGenNoVectorIndexes(t *testing.T) {
	content := readProjectFile(t, "internal/clients/code_db/generated/indexes_gen.go")

	if strings.Contains(content, "CREATE VECTOR INDEX") {
		t.Error("indexes_gen.go still contains vector index statements — embedding fields removed from schema")
	}
}

// === Task 2: REPL refactor — goroutine+channel+select pattern ===
//
// The REPL was refactored from a plain bufio.Scanner loop to a
// goroutine+channel+select pattern for proper context cancellation.

// TestReplHasScanResultType verifies that repl.go defines the scanResult
// struct used for channel-based stdin communication.
// Expected result: repl.go contains "type scanResult struct".
func TestReplHasScanResultType(t *testing.T) {
	content := readProjectFile(t, "internal/repl/repl.go")

	if !strings.Contains(content, "type scanResult struct") {
		t.Error("repl.go missing scanResult struct — needed for goroutine+channel REPL pattern")
	}
}

// TestReplUsesChannelForStdin verifies that Run() creates a channel for
// stdin reads instead of blocking directly on bufio.Scanner.
// Expected result: repl.go contains "make(chan scanResult)".
func TestReplUsesChannelForStdin(t *testing.T) {
	content := readProjectFile(t, "internal/repl/repl.go")

	if !strings.Contains(content, "make(chan scanResult)") {
		t.Error("repl.go does not use channel for stdin reads — goroutine+channel pattern required")
	}
}

// TestReplUsesSelectForCancellation verifies that Run() uses a select
// statement to multiplex between stdin reads and context cancellation.
// Expected result: repl.go contains "select {" and "case <-ctx.Done()".
func TestReplUsesSelectForCancellation(t *testing.T) {
	content := readProjectFile(t, "internal/repl/repl.go")

	if !strings.Contains(content, "select {") {
		t.Error("repl.go does not use select statement — required for context cancellation multiplexing")
	}
	if !strings.Contains(content, "case <-ctx.Done()") {
		t.Error("repl.go does not check ctx.Done() in select — context cancellation will not work")
	}
}

// TestReplUsesGoroutineForScanner verifies that Run() launches a goroutine
// for the bufio.Scanner loop (non-blocking stdin reads).
// Expected result: repl.go contains "go func()" for scanner goroutine.
func TestReplUsesGoroutineForScanner(t *testing.T) {
	content := readProjectFile(t, "internal/repl/repl.go")

	if !strings.Contains(content, "go func()") {
		t.Error("repl.go does not launch goroutine for scanner — blocking stdin reads will prevent context cancellation")
	}
}

// TestReplNoDirectCtxErrCheck verifies that Run() does not use the old
// pattern of checking ctx.Err() at the top of the loop (replaced by select).
// Expected result: repl.go does NOT contain "ctx.Err()" direct check.
func TestReplNoDirectCtxErrCheck(t *testing.T) {
	content := readProjectFile(t, "internal/repl/repl.go")

	if strings.Contains(content, "ctx.Err()") {
		t.Error("repl.go still uses ctx.Err() direct check — should use select+ctx.Done() pattern instead")
	}
}

// === Task 4: Verify go vet clean ===
//
// Structural verification that regenerated files have valid Go patterns.

// TestGeneratedModelsGenHasPackageDecl verifies that models_gen.go
// has a valid package declaration (basic structural check).
// Expected result: First non-comment line declares package generated.
func TestGeneratedModelsGenHasPackageDecl(t *testing.T) {
	content := readProjectFile(t, "internal/clients/code_db/generated/models_gen.go")

	if !strings.Contains(content, "package generated") {
		t.Error("models_gen.go missing 'package generated' declaration")
	}
}

// TestGeneratedGraphmodelGenHasPackageDecl verifies that graphmodel_gen.go
// has a valid package declaration.
// Expected result: File declares package generated.
func TestGeneratedGraphmodelGenHasPackageDecl(t *testing.T) {
	content := readProjectFile(t, "internal/clients/code_db/generated/graphmodel_gen.go")

	if !strings.Contains(content, "package generated") {
		t.Error("graphmodel_gen.go missing 'package generated' declaration")
	}
}

// === Task 5: Verify generated code compiles with go build ===
//
// This task verifies that the regenerated type ordering doesn't break
// compilation. The type ordering change (Module before Class) should
// be cosmetic only.

// TestGeneratedModelsHaveAllNodeTypes verifies that all 7 node types
// still exist in models_gen.go after regeneration.
// Expected result: All 7 types present (Repository, Folder, File, Module, Class, Function, ExternalReference).
func TestGeneratedModelsHaveAllNodeTypes(t *testing.T) {
	content := readProjectFile(t, "internal/clients/code_db/generated/models_gen.go")

	types := []string{
		"Repository",
		"Folder",
		"File",
		"Module",
		"Class",
		"Function",
		"ExternalReference",
	}

	for _, typeName := range types {
		t.Run(typeName, func(t *testing.T) {
			if !strings.Contains(content, "type "+typeName+" struct") {
				t.Errorf("models_gen.go missing %s struct after regeneration", typeName)
			}
		})
	}
}

// TestGeneratedSchemaHasAllNodeTypes verifies that the augmented
// schema.graphql still contains all 7 node types after regeneration.
// Expected result: All 7 types present.
func TestGeneratedSchemaHasAllNodeTypes(t *testing.T) {
	content := readProjectFile(t, "internal/clients/code_db/generated/schema.graphql")

	types := []string{"Repository", "Folder", "File", "Module", "Class", "Function", "ExternalReference"}

	for _, typeName := range types {
		t.Run(typeName, func(t *testing.T) {
			if !strings.Contains(content, typeName) {
				t.Errorf("generated schema.graphql missing %s type after regeneration", typeName)
			}
		})
	}
}
