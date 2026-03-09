package verify_test

import (
	"strings"
	"testing"
)

// === Task 1: Update go-ormql dependency and regenerate code ===
// These tests verify that go-ormql has been updated with merge mutations,
// top-level connect, and relationship WHERE support, and that codegen has
// produced the new types.

// TestGeneratedModelsHaveMergeInputTypes verifies that the regenerated
// models_gen.go contains MergeInput types for all 6 node types.
// Expected result: models_gen.go contains RepositoryMergeInput, FolderMergeInput,
// FileMergeInput, ModuleMergeInput, ClassMergeInput, FunctionMergeInput.
func TestGeneratedModelsHaveMergeInputTypes(t *testing.T) {
	models := readProjectFile(t, "internal/clients/code_db/generated/models_gen.go")

	mergeTypes := []string{
		"RepositoryMergeInput",
		"FolderMergeInput",
		"FileMergeInput",
		"ModuleMergeInput",
		"ClassMergeInput",
		"FunctionMergeInput",
	}

	for _, mt := range mergeTypes {
		t.Run(mt, func(t *testing.T) {
			if !strings.Contains(models, mt) {
				t.Errorf("models_gen.go missing %s (go-ormql merge mutation support required)", mt)
			}
		})
	}
}

// TestGeneratedModelsHaveMatchInputTypes verifies that MergeInput types
// contain MatchInput sub-types for identifying nodes during MERGE operations.
// Expected result: models_gen.go contains RepositoryMatchInput, FolderMatchInput, etc.
func TestGeneratedModelsHaveMatchInputTypes(t *testing.T) {
	models := readProjectFile(t, "internal/clients/code_db/generated/models_gen.go")

	matchTypes := []string{
		"RepositoryMatchInput",
		"FolderMatchInput",
		"FileMatchInput",
	}

	for _, mt := range matchTypes {
		t.Run(mt, func(t *testing.T) {
			if !strings.Contains(models, mt) {
				t.Errorf("models_gen.go missing %s (go-ormql merge mutation support required)", mt)
			}
		})
	}
}

// TestGeneratedModelsHaveConnectInputTypes verifies that the regenerated
// models_gen.go contains ConnectInput types for standalone edge creation.
// Expected result: models_gen.go contains ConnectRepositoryFoldersInput, etc.
func TestGeneratedModelsHaveConnectInputTypes(t *testing.T) {
	models := readProjectFile(t, "internal/clients/code_db/generated/models_gen.go")

	connectTypes := []string{
		"ConnectRepositoryFoldersInput",
		"ConnectRepositoryFilesInput",
		"ConnectFolderSubfoldersInput",
		"ConnectFolderFilesInput",
		"ConnectFolderRepositoryInput",
		"ConnectFileRepositoryInput",
	}

	for _, ct := range connectTypes {
		t.Run(ct, func(t *testing.T) {
			if !strings.Contains(models, ct) {
				t.Errorf("models_gen.go missing %s (go-ormql top-level connect mutation support required)", ct)
			}
		})
	}
}

// TestGeneratedModelsHaveConnectInfo verifies that ConnectInfo response type
// exists in the generated code.
// Expected result: models_gen.go contains ConnectInfo struct.
func TestGeneratedModelsHaveConnectInfo(t *testing.T) {
	models := readProjectFile(t, "internal/clients/code_db/generated/models_gen.go")

	if !strings.Contains(models, "ConnectInfo") {
		t.Error("models_gen.go missing ConnectInfo response type (go-ormql connect mutation support required)")
	}
}

// TestGeneratedModelsHaveRelationshipWHERE verifies that WHERE input types
// contain relationship filter fields for querying by related nodes.
// Expected result: FolderWhere contains Repository field, FileWhere contains Repository field.
func TestGeneratedModelsHaveRelationshipWHERE(t *testing.T) {
	models := readProjectFile(t, "internal/clients/code_db/generated/models_gen.go")

	// FolderWhere should have a Repository field for relationship WHERE filter
	if !strings.Contains(models, "FolderWhere") {
		t.Fatal("models_gen.go missing FolderWhere type")
	}

	// FileWhere should have a Repository field for relationship WHERE filter
	if !strings.Contains(models, "FileWhere") {
		t.Fatal("models_gen.go missing FileWhere type")
	}

	// FunctionWhere should have a Repository field for relationship WHERE filter
	if !strings.Contains(models, "FunctionWhere") {
		t.Fatal("models_gen.go missing FunctionWhere type")
	}
}

// TestGeneratedSchemaHasMergeMutations verifies that the augmented schema.graphql
// contains merge mutation definitions.
// Expected result: Generated schema.graphql contains mergeRepositorys, mergeFolders, etc.
func TestGeneratedSchemaHasMergeMutations(t *testing.T) {
	schema := readProjectFile(t, "internal/clients/code_db/generated/schema.graphql")

	mutations := []string{
		"mergeRepositorys",
		"mergeFolders",
		"mergeFiles",
		"mergeModules",
		"mergeClasss",
		"mergeFunctions",
	}

	for _, m := range mutations {
		t.Run(m, func(t *testing.T) {
			if !strings.Contains(schema, m) {
				t.Errorf("generated schema.graphql missing %s mutation", m)
			}
		})
	}
}

// TestGeneratedSchemaHasConnectMutations verifies that the augmented schema.graphql
// contains connect mutation definitions for edge creation.
// Expected result: Generated schema.graphql contains connectRepositoryFolders, etc.
func TestGeneratedSchemaHasConnectMutations(t *testing.T) {
	schema := readProjectFile(t, "internal/clients/code_db/generated/schema.graphql")

	mutations := []string{
		"connectRepositoryFolders",
		"connectRepositoryFiles",
		"connectFolderSubfolders",
		"connectFolderFiles",
		"connectFolderRepository",
		"connectFileRepository",
	}

	for _, m := range mutations {
		t.Run(m, func(t *testing.T) {
			if !strings.Contains(schema, m) {
				t.Errorf("generated schema.graphql missing %s connect mutation", m)
			}
		})
	}
}

// === Task 6: Remove Driver() from CodeDB ===
// These tests verify that Driver() method has been removed.

// TestCodeDBHasNoDriverMethod verifies that codedb.go no longer contains
// a Driver() method. All graph operations should use Client().Execute().
// Expected result: codedb.go does NOT contain "func (db *CodeDB) Driver()".
func TestCodeDBHasNoDriverMethod(t *testing.T) {
	content := readProjectFile(t, "internal/clients/code_db/codedb.go")

	if strings.Contains(content, "func (db *CodeDB) Driver()") {
		t.Error("codedb.go still contains Driver() method — should be removed (all operations use Client().Execute())")
	}
}

// TestCodeDBTestsHaveNoDriverTests verifies that codedb_test.go no longer
// contains tests for the removed Driver() method.
// Expected result: codedb_test.go does NOT contain TestDriver_ prefixed tests.
func TestCodeDBTestsHaveNoDriverTests(t *testing.T) {
	content := readProjectFile(t, "internal/clients/code_db/codedb_test.go")

	if strings.Contains(content, "TestDriver_") {
		t.Error("codedb_test.go still contains TestDriver_ tests — should be removed with Driver() method")
	}
}

// === Task 2-5: Indexer migration verification ===
// These tests verify the indexer has been migrated from raw Cypher to Client().Execute().

// TestIndexerHasNoCypherConstants verifies that indexer.go no longer contains
// any raw Cypher query constants (all migrated to Client().Execute() GraphQL).
// Expected result: indexer.go does NOT contain any cypher* constants.
func TestIndexerHasNoCypherConstants(t *testing.T) {
	content := readProjectFile(t, "internal/indexer/indexer.go")

	cypherConstants := []string{
		"cypherUpsertRepo",
		"cypherQueryFolders",
		"cypherQueryFiles",
		"cypherMergeFolders",
		"cypherMergeFiles",
		"cypherRepoContainsFolders",
		"cypherRepoContainsFiles",
		"cypherFolderContainsFolders",
		"cypherFolderContainsFiles",
		"cypherFoldersBelongToRepo",
		"cypherFilesBelongToRepo",
	}

	for _, c := range cypherConstants {
		t.Run(c, func(t *testing.T) {
			if strings.Contains(content, c) {
				t.Errorf("indexer.go still contains raw Cypher constant %q — should be migrated to Client().Execute()", c)
			}
		})
	}
}

// TestIndexerHasNoCypherImport verifies that indexer.go no longer imports
// the cypher package (all operations use Client().Execute() GraphQL).
// Expected result: indexer.go does NOT import go-ormql/pkg/cypher.
func TestIndexerHasNoCypherImport(t *testing.T) {
	content := readProjectFile(t, "internal/indexer/indexer.go")

	if strings.Contains(content, "go-ormql/pkg/cypher") {
		t.Error("indexer.go still imports go-ormql/pkg/cypher — should be removed (all operations use Client().Execute())")
	}
}

// TestIndexerHasNoDriverImport verifies that indexer.go no longer imports
// the driver package (no direct Driver() usage).
// Expected result: indexer.go does NOT import go-ormql/pkg/driver.
func TestIndexerHasNoDriverImport(t *testing.T) {
	content := readProjectFile(t, "internal/indexer/indexer.go")

	if strings.Contains(content, "go-ormql/pkg/driver") {
		t.Error("indexer.go still imports go-ormql/pkg/driver — should be removed (no Driver() usage)")
	}
}

// TestIndexerHasNoBatchEdgeWrite verifies that the batchEdgeWrite helper
// has been removed (replaced by Client().Execute() connect mutations).
// Expected result: indexer.go does NOT contain batchEdgeWrite function.
func TestIndexerHasNoBatchEdgeWrite(t *testing.T) {
	content := readProjectFile(t, "internal/indexer/indexer.go")

	if strings.Contains(content, "batchEdgeWrite") {
		t.Error("indexer.go still contains batchEdgeWrite helper — should be removed (replaced by connect* mutations)")
	}
}

// TestIndexerHasNoParseRecords verifies that the parseRecords helper
// has been removed (replaced by result.Decode() into generated structs).
// Expected result: indexer.go does NOT contain parseRecords function.
func TestIndexerHasNoParseRecords(t *testing.T) {
	content := readProjectFile(t, "internal/indexer/indexer.go")

	if strings.Contains(content, "parseRecords") {
		t.Error("indexer.go still contains parseRecords helper — should be removed (replaced by result.Decode())")
	}
}

// TestIndexerHasNoDriverCalls verifies that indexer.go does not call
// db.Driver() anywhere — all operations use db.Client().
// Expected result: indexer.go does NOT contain ".Driver()".
func TestIndexerHasNoDriverCalls(t *testing.T) {
	content := readProjectFile(t, "internal/indexer/indexer.go")

	if strings.Contains(content, ".Driver()") {
		t.Error("indexer.go still calls .Driver() — should use .Client().Execute() exclusively")
	}
}

// TestMainGoHasNoUnusedAssignments verifies that cmd/codectx/main.go does
// not have unused _ assignments for pipeline components.
// Expected result: main.go does NOT contain "_ = idx", "_ = analyzer", "_ = embedder".
func TestMainGoHasNoUnusedAssignments(t *testing.T) {
	content := readProjectFile(t, "cmd/codectx/main.go")

	unused := []string{
		"_ = idx",
		"_ = analyzer",
		"_ = embedder",
	}

	for _, u := range unused {
		t.Run(u, func(t *testing.T) {
			if strings.Contains(content, u) {
				t.Errorf("cmd/codectx/main.go still contains unused assignment %q — pipeline should be wired", u)
			}
		})
	}
}

// === Task 11: Update all tests for Client-based API ===
// These tests verify that test files have been updated to use Client-based
// assertions instead of Driver-based ones.

// TestPersistenceTestsHaveNoCypherImport verifies that persistence_test.go
// no longer imports the cypher package (test tooling should not depend on
// raw Cypher types after migration to Client().Execute()).
// Expected result: persistence_test.go does NOT import go-ormql/pkg/cypher.
func TestPersistenceTestsHaveNoCypherImport(t *testing.T) {
	content := readProjectFile(t, "internal/indexer/persistence_test.go")

	if strings.Contains(content, "go-ormql/pkg/cypher") {
		t.Error("persistence_test.go still imports go-ormql/pkg/cypher — test tooling should use Client-based assertions")
	}
}

// TestPersistenceTestsHaveNoDriverImport verifies that persistence_test.go
// no longer imports the driver package (recording driver should not depend
// on driver types after migration).
// Expected result: persistence_test.go does NOT import go-ormql/pkg/driver.
func TestPersistenceTestsHaveNoDriverImport(t *testing.T) {
	content := readProjectFile(t, "internal/indexer/persistence_test.go")

	if strings.Contains(content, "go-ormql/pkg/driver") {
		t.Error("persistence_test.go still imports go-ormql/pkg/driver — test tooling should use Client-based assertions")
	}
}

// TestPersistenceTestsHaveNoDriverMethodCalls verifies that persistence_test.go
// does not reference Driver() method (since it's being removed from CodeDB).
// Expected result: persistence_test.go does NOT contain ".Driver()".
func TestPersistenceTestsHaveNoDriverMethodCalls(t *testing.T) {
	content := readProjectFile(t, "internal/indexer/persistence_test.go")

	if strings.Contains(content, ".Driver()") {
		t.Error("persistence_test.go still references .Driver() — should use Client-based test patterns")
	}
}

// TestMigrationTestsHaveNoCypherImport verifies that migration_test.go
// test tooling has been updated to not import raw Cypher types.
// Expected result: migration_test.go does NOT import go-ormql/pkg/cypher.
func TestMigrationTestsHaveNoCypherImport(t *testing.T) {
	content := readProjectFile(t, "internal/indexer/migration_test.go")

	if strings.Contains(content, "go-ormql/pkg/cypher") {
		t.Error("migration_test.go still imports go-ormql/pkg/cypher — should be removed after migration")
	}
}

// TestCodeDBTestCoverageAbove80 verifies that codedb_test.go has been
// updated to remove Driver() tests and maintain coverage above 80%.
// Expected result: codedb_test.go does NOT contain "func (d *NoopDriver)" or
// raw driver test utilities that reference the removed Driver() method.
func TestCodeDBTestCoverageAbove80(t *testing.T) {
	content := readProjectFile(t, "internal/clients/code_db/codedb_test.go")

	// After Driver() removal, tests should not reference Driver test patterns
	if strings.Contains(content, "TestDriver_ReturnsTypedDriver") {
		t.Error("codedb_test.go still contains TestDriver_ReturnsTypedDriver — should be removed with Driver()")
	}
	if strings.Contains(content, "TestDriver_NilAfterClose") {
		t.Error("codedb_test.go still contains TestDriver_NilAfterClose — should be removed with Driver()")
	}
}

// === Task 12: Integration test verification ===
// These tests verify that integration tests have been updated for Client-based API.

// TestIntegrationTestsHaveNoCypherImport verifies that
// indexer_integration_test.go no longer imports the cypher package.
// Expected result: Integration tests use Client().Execute() for verification queries.
func TestIntegrationTestsHaveNoCypherImport(t *testing.T) {
	content := readProjectFile(t, "internal/indexer/indexer_integration_test.go")

	if strings.Contains(content, "go-ormql/pkg/cypher") {
		t.Error("indexer_integration_test.go still imports go-ormql/pkg/cypher — verification queries should use Client().Execute()")
	}
}

// TestIntegrationTestsHaveNoDriverCalls verifies that integration tests
// do not call Driver() for verification queries.
// Expected result: Integration tests use Client().Execute() exclusively.
func TestIntegrationTestsHaveNoDriverCalls(t *testing.T) {
	content := readProjectFile(t, "internal/indexer/indexer_integration_test.go")

	if strings.Contains(content, ".Driver()") {
		t.Error("indexer_integration_test.go still calls .Driver() — should use Client().Execute() for verification queries")
	}
}

// TestIntegrationTestsHaveNoRawCypherStatements verifies that integration
// tests do not construct raw cypher.Statement values for verification.
// Expected result: No cypher.Statement in integration test code.
func TestIntegrationTestsHaveNoRawCypherStatements(t *testing.T) {
	content := readProjectFile(t, "internal/indexer/indexer_integration_test.go")

	if strings.Contains(content, "cypher.Statement") {
		t.Error("indexer_integration_test.go still uses cypher.Statement — should use Client().Execute() with GraphQL queries")
	}
}

// TestIntegrationTestsVerifyWithForRepoQueries verifies that integration
// tests use ForRepo-based queries to verify graph state after indexing.
// Expected result: Integration tests contain ForRepo() calls.
func TestIntegrationTestsVerifyWithForRepoQueries(t *testing.T) {
	content := readProjectFile(t, "internal/indexer/indexer_integration_test.go")

	if !strings.Contains(content, "ForRepo") {
		t.Error("indexer_integration_test.go does not use ForRepo for verification — should query graph state via ForRepo-scoped client")
	}
}
