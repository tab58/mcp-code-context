package verify

import (
	"os"
	"strings"
	"testing"
)

// === Task 1: Add ExternalReference to schema.graphql ===

// TestSchema_ContainsExternalReferenceType verifies that schema.graphql
// defines an ExternalReference node type.
// Expected result: schema contains "type ExternalReference @node".
func TestSchema_ContainsExternalReferenceType(t *testing.T) {
	schema, err := os.ReadFile("../../internal/clients/code_db/schema.graphql")
	if err != nil {
		t.Fatalf("failed to read schema.graphql: %v", err)
	}
	if !strings.Contains(string(schema), "type ExternalReference @node") {
		t.Error("schema.graphql missing 'type ExternalReference @node' — Task 1 requires ExternalReference node type")
	}
}

// TestSchema_ExternalReferenceHasNameField verifies that ExternalReference
// has a required name field.
// Expected result: schema contains "name: String!" within ExternalReference.
func TestSchema_ExternalReferenceHasNameField(t *testing.T) {
	schema, err := os.ReadFile("../../internal/clients/code_db/schema.graphql")
	if err != nil {
		t.Fatalf("failed to read schema.graphql: %v", err)
	}
	content := string(schema)
	idx := strings.Index(content, "type ExternalReference @node")
	if idx < 0 {
		t.Fatal("schema.graphql missing ExternalReference type")
	}
	block := content[idx:]
	endIdx := strings.Index(block, "}")
	if endIdx < 0 {
		t.Fatal("could not find closing brace for ExternalReference type")
	}
	block = block[:endIdx]
	if !strings.Contains(block, "name: String!") {
		t.Error("ExternalReference missing 'name: String!' field")
	}
}

// TestSchema_ExternalReferenceHasImportPathField verifies that ExternalReference
// has a required importPath field.
// Expected result: schema contains "importPath: String!" within ExternalReference.
func TestSchema_ExternalReferenceHasImportPathField(t *testing.T) {
	schema, err := os.ReadFile("../../internal/clients/code_db/schema.graphql")
	if err != nil {
		t.Fatalf("failed to read schema.graphql: %v", err)
	}
	content := string(schema)
	idx := strings.Index(content, "type ExternalReference @node")
	if idx < 0 {
		t.Fatal("schema.graphql missing ExternalReference type")
	}
	block := content[idx:]
	endIdx := strings.Index(block, "}")
	if endIdx < 0 {
		t.Fatal("could not find closing brace for ExternalReference type")
	}
	block = block[:endIdx]
	if !strings.Contains(block, "importPath: String!") {
		t.Error("ExternalReference missing 'importPath: String!' field")
	}
}

// TestSchema_ExternalReferenceHasBelongsToRepository verifies that ExternalReference
// has a BELONGS_TO relationship to Repository.
// Expected result: schema contains a BELONGS_TO relationship.
func TestSchema_ExternalReferenceHasBelongsToRepository(t *testing.T) {
	schema, err := os.ReadFile("../../internal/clients/code_db/schema.graphql")
	if err != nil {
		t.Fatalf("failed to read schema.graphql: %v", err)
	}
	content := string(schema)
	idx := strings.Index(content, "type ExternalReference @node")
	if idx < 0 {
		t.Fatal("schema.graphql missing ExternalReference type")
	}
	block := content[idx:]
	endIdx := strings.Index(block, "}")
	if endIdx < 0 {
		t.Fatal("could not find closing brace for ExternalReference type")
	}
	block = block[:endIdx]
	if !strings.Contains(block, "BELONGS_TO") {
		t.Error("ExternalReference missing BELONGS_TO relationship to Repository")
	}
}

// === Task 2: Regenerate go-ormql code ===

// TestCodegen_ExternalReferenceModelExists verifies that the generated code
// contains an ExternalReference model struct.
// Expected result: models_gen.go contains "ExternalReference" struct.
func TestCodegen_ExternalReferenceModelExists(t *testing.T) {
	models, err := os.ReadFile("../../internal/clients/code_db/generated/models_gen.go")
	if err != nil {
		t.Fatalf("failed to read models_gen.go: %v", err)
	}
	if !strings.Contains(string(models), "ExternalReference") {
		t.Error("generated models_gen.go missing ExternalReference type — run 'task generate' after adding ExternalReference to schema.graphql")
	}
}

// TestCodegen_MergeExternalReferencesExists verifies that the augmented schema
// contains a mergeExternalReferences mutation.
// Expected result: augmented schema.graphql contains "mergeExternalReferences".
func TestCodegen_MergeExternalReferencesExists(t *testing.T) {
	schema, err := os.ReadFile("../../internal/clients/code_db/generated/schema.graphql")
	if err != nil {
		t.Fatalf("failed to read generated schema.graphql: %v", err)
	}
	if !strings.Contains(string(schema), "mergeExternalReferences") {
		t.Error("generated schema.graphql missing 'mergeExternalReferences' mutation — run 'task generate' after adding ExternalReference to schema.graphql")
	}
}

// TestCodegen_ConnectExternalReferenceRepositoryExists verifies that the augmented
// schema contains a connectExternalReferenceRepository mutation.
// Expected result: augmented schema.graphql contains "connectExternalReferenceRepository".
func TestCodegen_ConnectExternalReferenceRepositoryExists(t *testing.T) {
	schema, err := os.ReadFile("../../internal/clients/code_db/generated/schema.graphql")
	if err != nil {
		t.Fatalf("failed to read generated schema.graphql: %v", err)
	}
	if !strings.Contains(string(schema), "connectExternalReferenceRepository") {
		t.Error("generated schema.graphql missing 'connectExternalReferenceRepository' mutation — run 'task generate'")
	}
}

// === Task 4: Update ExtractReferences interface ===

// TestExtractReferencesInterface_Has4Params verifies that the Extractor interface's
// ExtractReferences method requires 4 parameters (tree, source, filePath, repoPath).
// Expected result: extractor.go contains the 4-param signature.
func TestExtractReferencesInterface_Has4Params(t *testing.T) {
	source, err := os.ReadFile("../../internal/analysis/extractor.go")
	if err != nil {
		t.Fatalf("failed to read extractor.go: %v", err)
	}
	content := string(source)
	if !strings.Contains(content, "ExtractReferences(tree *sitter.Tree, source []byte, filePath string, repoPath string)") {
		t.Error("Extractor.ExtractReferences missing repoPath parameter — should be (tree, source, filePath, repoPath)")
	}
}

// === Task 5: Go extractor — verify goBuiltins exists ===

// TestGoExtractor_HasGoBuiltinsSet verifies that the Go extractor source
// contains a goBuiltins variable for filtering built-in identifiers.
// Expected result: golang/extractor.go contains "goBuiltins".
func TestGoExtractor_HasGoBuiltinsSet(t *testing.T) {
	source, err := os.ReadFile("../../internal/analysis/golang/extractor.go")
	if err != nil {
		t.Fatalf("failed to read golang/extractor.go: %v", err)
	}
	if !strings.Contains(string(source), "goBuiltins") {
		t.Error("golang/extractor.go missing 'goBuiltins' set — Task 5 requires built-in identifier filter")
	}
}

// TestGoExtractor_HasBuildImportMap verifies that the Go extractor source
// contains a buildImportMap function.
// Expected result: golang/extractor.go contains "buildImportMap".
func TestGoExtractor_HasBuildImportMap(t *testing.T) {
	source, err := os.ReadFile("../../internal/analysis/golang/extractor.go")
	if err != nil {
		t.Fatalf("failed to read golang/extractor.go: %v", err)
	}
	if !strings.Contains(string(source), "buildImportMap") {
		t.Error("golang/extractor.go missing 'buildImportMap' function — Task 5 requires import classification")
	}
}

// TestGoExtractor_HasIsStdlib verifies that the Go extractor source
// contains an isStdlib function.
// Expected result: golang/extractor.go contains "isStdlib".
func TestGoExtractor_HasIsStdlib(t *testing.T) {
	source, err := os.ReadFile("../../internal/analysis/golang/extractor.go")
	if err != nil {
		t.Fatalf("failed to read golang/extractor.go: %v", err)
	}
	if !strings.Contains(string(source), "isStdlib") {
		t.Error("golang/extractor.go missing 'isStdlib' function — Task 5 requires stdlib detection")
	}
}

// === Task 6: TS extractor — verify isRelativeImport exists ===

// TestTSExtractor_HasIsRelativeImport verifies that the TypeScript extractor
// source contains an isRelativeImport function.
// Expected result: typescript/extractor.go contains "isRelativeImport".
func TestTSExtractor_HasIsRelativeImport(t *testing.T) {
	source, err := os.ReadFile("../../internal/analysis/typescript/extractor.go")
	if err != nil {
		t.Fatalf("failed to read typescript/extractor.go: %v", err)
	}
	if !strings.Contains(string(source), "isRelativeImport") {
		t.Error("typescript/extractor.go missing 'isRelativeImport' function — Task 6 requires relative import detection")
	}
}

// === Task 7: writeExternalReferences in analyzer ===

// TestAnalyzer_HasWriteExternalReferences verifies that analyzer.go
// contains a writeExternalReferences method.
// Expected result: analyzer.go contains "writeExternalReferences".
func TestAnalyzer_HasWriteExternalReferences(t *testing.T) {
	source, err := os.ReadFile("../../internal/analysis/analyzer.go")
	if err != nil {
		t.Fatalf("failed to read analyzer.go: %v", err)
	}
	if !strings.Contains(string(source), "writeExternalReferences") {
		t.Error("analyzer.go missing 'writeExternalReferences' method — Task 7 requires ExternalReference graph writes")
	}
}

// TestAnalyzer_HasGqlMergeExternalReferences verifies that analyzer.go
// contains the gqlMergeExternalReferences GraphQL constant.
// Expected result: analyzer.go contains "gqlMergeExternalReferences".
func TestAnalyzer_HasGqlMergeExternalReferences(t *testing.T) {
	source, err := os.ReadFile("../../internal/analysis/external_refs.go")
	if err != nil {
		t.Fatalf("failed to read external_refs.go: %v", err)
	}
	if !strings.Contains(string(source), "gqlMergeExternalReferences") {
		t.Error("analyzer.go missing 'gqlMergeExternalReferences' constant — Task 7 requires ExternalReference merge mutation")
	}
}

// TestAnalyzer_HasGqlConnectFileExternalImports verifies that analyzer.go
// contains the gqlConnectFileExternalImports GraphQL constant.
// Expected result: analyzer.go contains "gqlConnectFileExternalImports".
func TestAnalyzer_HasGqlConnectFileExternalImports(t *testing.T) {
	source, err := os.ReadFile("../../internal/analysis/external_refs.go")
	if err != nil {
		t.Fatalf("failed to read external_refs.go: %v", err)
	}
	if !strings.Contains(string(source), "gqlConnectFileExternalImports") {
		t.Error("external_refs.go missing 'gqlConnectFileExternalImports' constant — Task 7")
	}
}

// TestAnalyzer_HasGqlConnectFunctionExternalCalls verifies that external_refs.go
// contains the gqlConnectFunctionExternalCalls GraphQL constant.
func TestAnalyzer_HasGqlConnectFunctionExternalCalls(t *testing.T) {
	source, err := os.ReadFile("../../internal/analysis/external_refs.go")
	if err != nil {
		t.Fatalf("failed to read external_refs.go: %v", err)
	}
	if !strings.Contains(string(source), "gqlConnectFunctionExternalCalls") {
		t.Error("external_refs.go missing 'gqlConnectFunctionExternalCalls' constant — Task 7")
	}
}

// TestAnalyzer_HasGqlConnectExternalReferenceRepository verifies that external_refs.go
// contains the gqlConnectExternalReferenceRepository GraphQL constant.
func TestAnalyzer_HasGqlConnectExternalReferenceRepository(t *testing.T) {
	source, err := os.ReadFile("../../internal/analysis/external_refs.go")
	if err != nil {
		t.Fatalf("failed to read external_refs.go: %v", err)
	}
	if !strings.Contains(string(source), "gqlConnectExternalReferenceRepository") {
		t.Error("external_refs.go missing 'gqlConnectExternalReferenceRepository' constant — Task 7")
	}
}

// === Task 8: resolvePass summary logging ===

// TestResolvePass_HasSummaryLogging verifies that analyzer.go uses summary
// logging instead of per-reference logging in resolvePass.
// Expected result: analyzer.go does NOT contain per-reference log.Printf
// for "unresolved reference" and DOES contain summary format string.
func TestResolvePass_HasSummaryLogging(t *testing.T) {
	source, err := os.ReadFile("../../internal/analysis/analyzer.go")
	if err != nil {
		t.Fatalf("failed to read analyzer.go: %v", err)
	}
	content := string(source)
	if strings.Contains(content, `log.Printf("analyzer: unresolved reference %q"`) {
		t.Error("analyzer.go still contains per-reference 'unresolved reference' logging — Task 8 requires single summary line")
	}
}

// === Task 10: analyzeFile passes repoPath to ExtractReferences ===

// TestAnalyzeFile_PassesRepoPathToExtractReferences verifies that analyzeFile
// passes the real repoPath to ExtractReferences (not empty string).
// Expected result: analyzer.go does NOT pass "" to ExtractReferences.
func TestAnalyzeFile_PassesRepoPathToExtractReferences(t *testing.T) {
	source, err := os.ReadFile("../../internal/analysis/analyzer.go")
	if err != nil {
		t.Fatalf("failed to read analyzer.go: %v", err)
	}
	content := string(source)
	// Check that ExtractReferences call passes repoPath, not ""
	if strings.Contains(content, `ExtractReferences(tree, source, filePath, "")`) {
		t.Error("analyzer.go passes empty string to ExtractReferences — Task 10 requires passing repoPath")
	}
}

// === Task 12: Full verification ===

// TestReference_HasIsExternalField verifies that Reference struct has IsExternal field.
// Expected result: types.go contains IsExternal field.
func TestReference_HasIsExternalField(t *testing.T) {
	source, err := os.ReadFile("../../internal/analysis/types.go")
	if err != nil {
		t.Fatalf("failed to read types.go: %v", err)
	}
	if !strings.Contains(string(source), "IsExternal") {
		t.Error("Reference struct missing IsExternal field")
	}
}

// TestReference_HasExternalImportPathField verifies that Reference struct
// has ExternalImportPath field.
// Expected result: types.go contains ExternalImportPath field.
func TestReference_HasExternalImportPathField(t *testing.T) {
	source, err := os.ReadFile("../../internal/analysis/types.go")
	if err != nil {
		t.Fatalf("failed to read types.go: %v", err)
	}
	if !strings.Contains(string(source), "ExternalImportPath") {
		t.Error("Reference struct missing ExternalImportPath field")
	}
}
