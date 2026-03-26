package verify

import (
	"os"
	"strings"
	"testing"
)

// === Task 1: Remove embedding field and @vector from Class and Function ===

// TestSchema_NoEmbeddingFieldOnClass verifies that the Class type
// in schema.graphql does not contain an 'embedding' field.
// Expected result: schema.graphql Class type has no 'embedding' line.
func TestSchema_NoEmbeddingFieldOnClass(t *testing.T) {
	data, err := os.ReadFile("../../internal/clients/code_db/schema.graphql")
	if err != nil {
		t.Fatalf("failed to read schema.graphql: %v", err)
	}
	src := string(data)
	if strings.Contains(src, "class_embedding") {
		t.Error("schema.graphql still contains 'class_embedding' — Task 1 requires removal of @vector on Class")
	}
	// Check for embedding field within Class type block
	lines := strings.Split(src, "\n")
	inClass := false
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "type Class") {
			inClass = true
			continue
		}
		if inClass && trimmed == "}" {
			break
		}
		if inClass && strings.Contains(trimmed, "embedding") {
			t.Errorf("Class type still has embedding field: %s", trimmed)
		}
	}
}

// TestSchema_NoEmbeddingFieldOnFunction verifies that the Function type
// in schema.graphql does not contain an 'embedding' field.
// Expected result: schema.graphql Function type has no 'embedding' line.
func TestSchema_NoEmbeddingFieldOnFunction(t *testing.T) {
	data, err := os.ReadFile("../../internal/clients/code_db/schema.graphql")
	if err != nil {
		t.Fatalf("failed to read schema.graphql: %v", err)
	}
	src := string(data)
	if strings.Contains(src, "function_embedding") {
		t.Error("schema.graphql still contains 'function_embedding' — Task 1 requires removal of @vector on Function")
	}
	// Check for embedding field within Function type block
	lines := strings.Split(src, "\n")
	inFunction := false
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "type Function") {
			inFunction = true
			continue
		}
		if inFunction && trimmed == "}" {
			break
		}
		if inFunction && strings.Contains(trimmed, "embedding") {
			t.Errorf("Function type still has embedding field: %s", trimmed)
		}
	}
}

// TestSchema_NoVectorDirective verifies that schema.graphql contains
// no @vector directives at all after removal.
// Expected result: no @vector anywhere in the file.
func TestSchema_NoVectorDirective(t *testing.T) {
	data, err := os.ReadFile("../../internal/clients/code_db/schema.graphql")
	if err != nil {
		t.Fatalf("failed to read schema.graphql: %v", err)
	}
	if strings.Contains(string(data), "@vector") {
		t.Error("schema.graphql still contains @vector directive — Task 1 requires full removal")
	}
}

// === Task 2: Regenerate go-ormql code (CreateIndexes becomes no-op) ===

// TestGenerated_NoEmbeddingFieldInModels verifies that generated models
// do not contain an Embedding field after codegen.
// Expected result: models_gen.go has no 'Embedding' struct field.
func TestGenerated_NoEmbeddingFieldInModels(t *testing.T) {
	data, err := os.ReadFile("../../internal/clients/code_db/generated/models_gen.go")
	if err != nil {
		t.Fatalf("failed to read models_gen.go: %v", err)
	}
	src := string(data)
	if strings.Contains(src, "Embedding") {
		t.Error("models_gen.go still contains 'Embedding' field — Task 2 requires regeneration after schema change")
	}
}

// TestGenerated_NoVectorIndexInIndexes verifies that indexes_gen.go
// contains no vector index creation code after regeneration.
// Expected result: indexes_gen.go has no vector/cosine/embedding references.
func TestGenerated_NoVectorIndexInIndexes(t *testing.T) {
	data, err := os.ReadFile("../../internal/clients/code_db/generated/indexes_gen.go")
	if err != nil {
		t.Fatalf("failed to read indexes_gen.go: %v", err)
	}
	src := string(data)
	for _, keyword := range []string{"vector", "cosine", "embedding", "768"} {
		if strings.Contains(strings.ToLower(src), keyword) {
			t.Errorf("indexes_gen.go still contains %q — Task 2 requires regeneration with no vector indexes", keyword)
		}
	}
}

// === Task 3: Delete internal/search/ package ===

// TestSearchPackageDeleted verifies that internal/search/ directory
// does not exist after cleanup.
// Expected result: directory does not exist.
func TestSearchPackageDeleted(t *testing.T) {
	_, err := os.Stat("../../internal/search")
	if err == nil {
		t.Error("internal/search/ directory still exists — Task 3 requires full deletion")
	}
}

// === Task 4: Delete internal/embedding/ package ===

// TestEmbeddingPackageDeleted verifies that internal/embedding/ directory
// does not exist after cleanup.
// Expected result: directory does not exist.
func TestEmbeddingPackageDeleted(t *testing.T) {
	_, err := os.Stat("../../internal/embedding")
	if err == nil {
		t.Error("internal/embedding/ directory still exists — Task 4 requires full deletion")
	}
}

// === Task 5: Delete cmd/embed/ directory ===

// TestCmdEmbedDeleted verifies that cmd/embed/ directory
// does not exist after cleanup.
// Expected result: directory does not exist.
func TestCmdEmbedDeleted(t *testing.T) {
	_, err := os.Stat("../../cmd/embed")
	if err == nil {
		t.Error("cmd/embed/ directory still exists — Task 5 requires full deletion")
	}
}

// === Task 6: Update MCP server — remove vector/embedding references ===

// TestMCPServer_NoEmbedImport verifies that server.go does not import
// embedding or search packages.
// Expected result: no embedding/search imports.
func TestMCPServer_NoEmbedImport(t *testing.T) {
	data, err := os.ReadFile("../../internal/api/mcp/server.go")
	if err != nil {
		t.Fatalf("failed to read server.go: %v", err)
	}
	src := string(data)
	if strings.Contains(src, "internal/embedding") {
		t.Error("server.go still imports internal/embedding — Task 6 requires removal")
	}
	if strings.Contains(src, "internal/search") {
		t.Error("server.go still imports internal/search — Task 6 requires removal")
	}
}

// TestMCPServer_NoEmbedFunc verifies that server.go does not contain
// embedFunc type or embed field.
// Expected result: no embedFunc type declaration or embed field.
func TestMCPServer_NoEmbedFunc(t *testing.T) {
	data, err := os.ReadFile("../../internal/api/mcp/server.go")
	if err != nil {
		t.Fatalf("failed to read server.go: %v", err)
	}
	src := string(data)
	if strings.Contains(src, "embedFunc") {
		t.Error("server.go still contains embedFunc type — Task 6 requires removal")
	}
	if strings.Contains(src, "embed    embed") || strings.Contains(src, "embed embedFunc") {
		t.Error("server.go still contains embed field — Task 6 requires removal")
	}
}

// TestMCPServer_NoEmbedderField verifies that Server struct does not have
// embedder field.
// Expected result: no embedder field in Server struct.
func TestMCPServer_NoEmbedderField(t *testing.T) {
	data, err := os.ReadFile("../../internal/api/mcp/server.go")
	if err != nil {
		t.Fatalf("failed to read server.go: %v", err)
	}
	if strings.Contains(string(data), "embedder") {
		t.Error("server.go still contains embedder field — Task 6 requires removal")
	}
}

// TestMCPServer_NoVectorSearchTool verifies that server.go does not
// register a vector_search tool.
// Expected result: no vector_search registration.
func TestMCPServer_NoVectorSearchTool(t *testing.T) {
	data, err := os.ReadFile("../../internal/api/mcp/server.go")
	if err != nil {
		t.Fatalf("failed to read server.go: %v", err)
	}
	if strings.Contains(string(data), "vector_search") {
		t.Error("server.go still registers vector_search tool — Task 6 requires removal")
	}
}

// TestMCPServer_NewServerThreeParams verifies that NewServer takes
// 3 parameters (db, idx, analyzer) not 4.
// Expected result: NewServer signature has 3 params.
func TestMCPServer_NewServerThreeParams(t *testing.T) {
	data, err := os.ReadFile("../../internal/api/mcp/server.go")
	if err != nil {
		t.Fatalf("failed to read server.go: %v", err)
	}
	src := string(data)
	// The old signature has 4 params including embedder
	if strings.Contains(src, "embedder *search.Embedder") {
		t.Error("NewServer still takes embedder parameter — Task 6 requires 3-param signature")
	}
}

// TestMCPTools_NoVectorSearchHandler verifies that tools.go does not
// contain handleVectorSearch function.
// Expected result: no handleVectorSearch function.
func TestMCPTools_NoVectorSearchHandler(t *testing.T) {
	data, err := os.ReadFile("../../internal/mcp/tools.go")
	if err != nil {
		t.Fatalf("failed to read tools.go: %v", err)
	}
	if strings.Contains(string(data), "handleVectorSearch") {
		t.Error("tools.go still contains handleVectorSearch — Task 6 requires removal")
	}
}

// TestMCPTools_NoSimilarQueryConstants verifies that tools.go does not
// contain functionsSimilar or classsSimilar GraphQL constants.
// Expected result: no similarity query constants.
func TestMCPTools_NoSimilarQueryConstants(t *testing.T) {
	data, err := os.ReadFile("../../internal/mcp/tools.go")
	if err != nil {
		t.Fatalf("failed to read tools.go: %v", err)
	}
	src := string(data)
	for _, keyword := range []string{"gqlFunctionsSimilar", "gqlClassesSimilar", "functionsSimilar", "classsSimilar"} {
		if strings.Contains(src, keyword) {
			t.Errorf("tools.go still contains %q — Task 6 requires removal of vector search constants", keyword)
		}
	}
}

// TestMCPTools_NoParseSimilarResults verifies that tools.go does not
// contain parseSimilarResults helper function.
// Expected result: no parseSimilarResults function.
func TestMCPTools_NoParseSimilarResults(t *testing.T) {
	data, err := os.ReadFile("../../internal/mcp/tools.go")
	if err != nil {
		t.Fatalf("failed to read tools.go: %v", err)
	}
	if strings.Contains(string(data), "parseSimilarResults") {
		t.Error("tools.go still contains parseSimilarResults — Task 6 requires removal")
	}
}

// TestMCPTools_NoVecToFloat64 verifies that tools.go does not contain
// vecToFloat64 helper function.
// Expected result: no vecToFloat64 function.
func TestMCPTools_NoVecToFloat64(t *testing.T) {
	data, err := os.ReadFile("../../internal/mcp/tools.go")
	if err != nil {
		t.Fatalf("failed to read tools.go: %v", err)
	}
	if strings.Contains(string(data), "vecToFloat64") {
		t.Error("tools.go still contains vecToFloat64 — Task 6 requires removal")
	}
}

// TestMCPSearch_NoHybridStrategy verifies that search.go/types.go do not
// contain strategyHybrid constant.
// Expected result: no strategyHybrid.
func TestMCPSearch_NoHybridStrategy(t *testing.T) {
	data, err := os.ReadFile("../../internal/mcp/types.go")
	if err != nil {
		t.Fatalf("failed to read types.go: %v", err)
	}
	if strings.Contains(string(data), "strategyHybrid") {
		t.Error("types.go still contains strategyHybrid — Task 6 requires removal (no vector fallback)")
	}
}

// TestMCPTools_NoExecuteHybrid verifies that tools.go does not contain
// executeHybrid function.
// Expected result: no executeHybrid function.
func TestMCPTools_NoExecuteHybrid(t *testing.T) {
	data, err := os.ReadFile("../../internal/mcp/tools.go")
	if err != nil {
		t.Fatalf("failed to read tools.go: %v", err)
	}
	if strings.Contains(string(data), "executeHybrid") {
		t.Error("tools.go still contains executeHybrid — Task 6 requires removal")
	}
}

// === Task 7: Update REPL — remove Embedder from Pipeline, EmbeddingModel from StatusInfo ===

// TestREPL_NoEmbedderInPipeline verifies that Pipeline struct does not
// contain Embedder field.
// Expected result: repl.go Pipeline has no Embedder.
func TestREPL_NoEmbedderInPipeline(t *testing.T) {
	data, err := os.ReadFile("../../internal/repl/repl.go")
	if err != nil {
		t.Fatalf("failed to read repl.go: %v", err)
	}
	if strings.Contains(string(data), "Embedder") {
		t.Error("repl.go Pipeline still contains Embedder field — Task 7 requires removal")
	}
}

// TestREPL_NoSearchImport verifies that repl.go does not import
// the search package.
// Expected result: no search import.
func TestREPL_NoSearchImport(t *testing.T) {
	data, err := os.ReadFile("../../internal/repl/repl.go")
	if err != nil {
		t.Fatalf("failed to read repl.go: %v", err)
	}
	if strings.Contains(string(data), "internal/search") {
		t.Error("repl.go still imports internal/search — Task 7 requires removal")
	}
}

// TestREPL_NoEmbeddingModelInStatus verifies that StatusInfo struct does not
// contain EmbeddingModel field.
// Expected result: repl.go StatusInfo has no EmbeddingModel.
func TestREPL_NoEmbeddingModelInStatus(t *testing.T) {
	data, err := os.ReadFile("../../internal/repl/repl.go")
	if err != nil {
		t.Fatalf("failed to read repl.go: %v", err)
	}
	if strings.Contains(string(data), "EmbeddingModel") {
		t.Error("repl.go StatusInfo still contains EmbeddingModel — Task 7 requires removal")
	}
}

// TestREPLCommands_NoEmbedStep verifies that commands.go does not contain
// Step 3 (embedding) in handleIngest.
// Expected result: no Embedder reference in commands.go.
func TestREPLCommands_NoEmbedStep(t *testing.T) {
	data, err := os.ReadFile("../../internal/repl/commands.go")
	if err != nil {
		t.Fatalf("failed to read commands.go: %v", err)
	}
	src := string(data)
	if strings.Contains(src, "Embedder") {
		t.Error("commands.go still references Embedder — Task 7 requires removal of embed step")
	}
	if strings.Contains(src, "EmbedNodes") {
		t.Error("commands.go still calls EmbedNodes — Task 7 requires removal of embed step")
	}
}

// TestREPLCommands_NoSearchImport verifies that commands.go does not import
// the search package.
// Expected result: no search import.
func TestREPLCommands_NoSearchImport(t *testing.T) {
	data, err := os.ReadFile("../../internal/repl/commands.go")
	if err != nil {
		t.Fatalf("failed to read commands.go: %v", err)
	}
	if strings.Contains(string(data), "internal/search") {
		t.Error("commands.go still imports internal/search — Task 7 requires removal")
	}
}

// TestREPLCommands_NoEmbeddingModelInStatus verifies that handleStatus
// does not print Embedding Model line.
// Expected result: no "Embedding Model" in status output.
func TestREPLCommands_NoEmbeddingModelInStatus(t *testing.T) {
	data, err := os.ReadFile("../../internal/repl/commands.go")
	if err != nil {
		t.Fatalf("failed to read commands.go: %v", err)
	}
	if strings.Contains(string(data), "Embedding Model") {
		t.Error("commands.go still prints 'Embedding Model' in status — Task 7 requires removal")
	}
}

// TestREPLHelp_IndexAnalyzeOnly verifies that handleHelp describes
// pipeline as "index -> analyze" not "index -> analyze -> embed".
// Expected result: help text says "index -> analyze", not "embed".
func TestREPLHelp_IndexAnalyzeOnly(t *testing.T) {
	data, err := os.ReadFile("../../internal/repl/commands.go")
	if err != nil {
		t.Fatalf("failed to read commands.go: %v", err)
	}
	src := string(data)
	if strings.Contains(src, "embed)") || strings.Contains(src, "-> embed") {
		t.Error("commands.go help still mentions 'embed' in pipeline description — Task 7 requires 'index -> analyze' only")
	}
}

// === Task 8: Simplify main.go — remove embedding/search refs ===

// TestMainGo_NoEmbeddingImport verifies that main.go does not import
// the embedding package.
// Expected result: no embedding import.
func TestMainGo_NoEmbeddingImport(t *testing.T) {
	data, err := os.ReadFile("../../cmd/codectx/main.go")
	if err != nil {
		t.Fatalf("failed to read main.go: %v", err)
	}
	if strings.Contains(string(data), "internal/embedding") {
		t.Error("main.go still imports internal/embedding — Task 8 requires removal")
	}
}

// TestMainGo_NoSearchImport verifies that main.go does not import
// the search package.
// Expected result: no search import.
func TestMainGo_NoSearchImport(t *testing.T) {
	data, err := os.ReadFile("../../cmd/codectx/main.go")
	if err != nil {
		t.Fatalf("failed to read main.go: %v", err)
	}
	if strings.Contains(string(data), "internal/search") {
		t.Error("main.go still imports internal/search — Task 8 requires removal")
	}
}

// TestMainGo_NoDefaultModelPath verifies that main.go does not contain
// defaultModelPath constant.
// Expected result: no defaultModelPath.
func TestMainGo_NoDefaultModelPath(t *testing.T) {
	data, err := os.ReadFile("../../cmd/codectx/main.go")
	if err != nil {
		t.Fatalf("failed to read main.go: %v", err)
	}
	if strings.Contains(string(data), "defaultModelPath") {
		t.Error("main.go still contains defaultModelPath — Task 8 requires removal")
	}
}

// TestMainGo_NoEmbeddingInit verifies that main.go does not call
// embedding.Init or embedding.Close or embedding.SuppressLog.
// Expected result: no embedding.Init/Close/SuppressLog calls.
func TestMainGo_NoEmbeddingInit(t *testing.T) {
	data, err := os.ReadFile("../../cmd/codectx/main.go")
	if err != nil {
		t.Fatalf("failed to read main.go: %v", err)
	}
	src := string(data)
	for _, fn := range []string{"embedding.Init", "embedding.Close", "embedding.SuppressLog"} {
		if strings.Contains(src, fn) {
			t.Errorf("main.go still calls %s — Task 8 requires removal", fn)
		}
	}
}

// TestMainGo_NoEmbedderCreation verifies that main.go does not create
// an Embedder instance.
// Expected result: no search.NewEmbedder call.
func TestMainGo_NoEmbedderCreation(t *testing.T) {
	data, err := os.ReadFile("../../cmd/codectx/main.go")
	if err != nil {
		t.Fatalf("failed to read main.go: %v", err)
	}
	if strings.Contains(string(data), "NewEmbedder") {
		t.Error("main.go still creates Embedder — Task 8 requires removal")
	}
}

// TestMainGo_NoEmbedderInPipeline verifies that main.go does not pass
// Embedder to Pipeline or NewServer.
// Expected result: no embedder references in pipeline/server construction.
func TestMainGo_NoEmbedderInPipeline(t *testing.T) {
	data, err := os.ReadFile("../../cmd/codectx/main.go")
	if err != nil {
		t.Fatalf("failed to read main.go: %v", err)
	}
	src := string(data)
	if strings.Contains(src, "Embedder:") || strings.Contains(src, "embedder,") || strings.Contains(src, "EmbeddingModel:") {
		t.Error("main.go still passes embedder to Pipeline or StatusInfo — Task 8 requires removal")
	}
}

// === Task 9: Update Taskfile.yml — remove build-llama and embed tasks ===

// TestTaskfile_NoBuildLlamaTask verifies that Taskfile.yml does not contain
// build-llama task.
// Expected result: no build-llama task definition.
func TestTaskfile_NoBuildLlamaTask(t *testing.T) {
	data, err := os.ReadFile("../../Taskfile.yml")
	if err != nil {
		t.Fatalf("failed to read Taskfile.yml: %v", err)
	}
	if strings.Contains(string(data), "build-llama") {
		t.Error("Taskfile.yml still contains build-llama task — Task 9 requires removal")
	}
}

// TestTaskfile_NoEmbedTask verifies that Taskfile.yml does not contain
// embed task.
// Expected result: no embed task definition.
func TestTaskfile_NoEmbedTask(t *testing.T) {
	data, err := os.ReadFile("../../Taskfile.yml")
	if err != nil {
		t.Fatalf("failed to read Taskfile.yml: %v", err)
	}
	if strings.Contains(string(data), "embed:") || strings.Contains(string(data), "cmd/embed") {
		t.Error("Taskfile.yml still contains embed task — Task 9 requires removal")
	}
}

// === Task 12: Archive semantic-search spec ===

// === Task 10: Update all tests referencing removed types ===

// TestMCPTests_NoVectorSearchTests verifies that mcp test files do not
// reference handleVectorSearch or vector_search.
// Expected result: no vector_search test functions.
func TestMCPTests_NoVectorSearchTests(t *testing.T) {
	data, err := os.ReadFile("../../internal/mcp/tools_test.go")
	if err != nil {
		t.Fatalf("failed to read tools_test.go: %v", err)
	}
	src := string(data)
	if strings.Contains(src, "handleVectorSearch") {
		t.Error("tools_test.go still tests handleVectorSearch — Task 10 requires removal")
	}
	if strings.Contains(src, "VectorSearch") {
		t.Error("tools_test.go still references VectorSearch — Task 10 requires removal")
	}
}

// TestMCPTests_NoHybridTests verifies that mcp test files do not
// reference executeHybrid or hybrid strategy.
// Expected result: no hybrid test functions.
func TestMCPTests_NoHybridTests(t *testing.T) {
	data, err := os.ReadFile("../../internal/mcp/tools_test.go")
	if err != nil {
		t.Fatalf("failed to read tools_test.go: %v", err)
	}
	src := string(data)
	if strings.Contains(src, "executeHybrid") {
		t.Error("tools_test.go still tests executeHybrid — Task 10 requires removal")
	}
	if strings.Contains(src, "HybridDedup") {
		t.Error("tools_test.go still has HybridDedup test — Task 10 requires removal")
	}
}

// TestMCPTests_NoMockEmbed verifies that mcp test helpers do not contain
// mockEmbed function (no longer needed without vector search).
// Expected result: no mockEmbed function.
func TestMCPTests_NoMockEmbed(t *testing.T) {
	data, err := os.ReadFile("../../internal/mcp/test_helpers_test.go")
	if err != nil {
		t.Fatalf("failed to read test_helpers_test.go: %v", err)
	}
	if strings.Contains(string(data), "mockEmbed") {
		t.Error("test_helpers_test.go still contains mockEmbed — Task 10 requires removal")
	}
}

// TestMCPTestHelpers_NoEmbedField verifies that test server constructors
// do not set embed field.
// Expected result: no embed field in test server construction.
func TestMCPTestHelpers_NoEmbedField(t *testing.T) {
	data, err := os.ReadFile("../../internal/mcp/test_helpers_test.go")
	if err != nil {
		t.Fatalf("failed to read test_helpers_test.go: %v", err)
	}
	if strings.Contains(string(data), "embed:") {
		t.Error("test_helpers_test.go still sets embed field — Task 10 requires removal")
	}
}

// TestREPLTests_NoEmbedderRef verifies that repl test files do not
// reference Embedder or EmbeddingModel.
// Expected result: no Embedder or EmbeddingModel in test files.
func TestREPLTests_NoEmbedderRef(t *testing.T) {
	for _, file := range []string{
		"../../internal/repl/repl_test.go",
		"../../internal/repl/commands_test.go",
		"../../internal/repl/test_helpers_test.go",
	} {
		data, err := os.ReadFile(file)
		if err != nil {
			continue // file may be deleted, which is fine
		}
		src := string(data)
		if strings.Contains(src, "EmbeddingModel") {
			t.Errorf("%s still references EmbeddingModel — Task 10 requires removal", file)
		}
	}
}

// === Task 11: Full build/vet/race verification ===

// TestBuildCompiles verifies that the project compiles without errors.
// This test itself compiling and running means the verify package is ok,
// but we need source inspection to verify no broken imports remain.
// Expected result: main.go, server.go, etc. have no embedding/search imports.
func TestBuildVerify_NoEmbeddingImportsAnywhere(t *testing.T) {
	files := []string{
		"../../cmd/codectx/main.go",
		"../../internal/api/mcp/server.go",
		"../../internal/mcp/tools.go",
		"../../internal/repl/repl.go",
		"../../internal/repl/commands.go",
	}
	for _, file := range files {
		data, err := os.ReadFile(file)
		if err != nil {
			t.Fatalf("failed to read %s: %v", file, err)
		}
		src := string(data)
		if strings.Contains(src, "internal/embedding") {
			t.Errorf("%s still imports internal/embedding — Task 11 build will fail", file)
		}
		if strings.Contains(src, "internal/search") {
			t.Errorf("%s still imports internal/search — Task 11 build will fail", file)
		}
	}
}

// === Task 12: Archive semantic-search spec ===

// TestSemanticSearchSpec_Archived verifies that the semantic-search spec
// is marked as Archived in INDEX.md (if the spec index file exists).
// Expected result: INDEX.md shows Archived for semantic-search, or file doesn't exist.
func TestSemanticSearchSpec_Archived(t *testing.T) {
	data, err := os.ReadFile("../../.claudemod/spec/INDEX.md")
	if err != nil {
		// File doesn't exist — spec directory removed, which is acceptable
		t.Skip("INDEX.md not found — spec directory may have been removed")
		return
	}
	src := string(data)
	if !strings.Contains(src, "Archived") {
		t.Error("INDEX.md does not show semantic-search as Archived — Task 12 requires archival")
	}
}
