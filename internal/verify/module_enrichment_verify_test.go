package verify

import (
	"os"
	"strings"
	"testing"
)

const schemaPath = "../../internal/clients/code_db/schema.graphql"

// === Task 1: Update schema.graphql ===

// TestSchema_ModuleHasImportPath verifies that the Module type in schema.graphql
// includes an importPath field.
// Expected result: schema contains "importPath: String" inside the Module type block.
func TestSchema_ModuleHasImportPath(t *testing.T) {
	content := readSchema(t)
	moduleBlock := extractTypeBlock(content, "Module")
	if moduleBlock == "" {
		t.Fatal("Module type block not found in schema.graphql")
	}
	if !strings.Contains(moduleBlock, "importPath: String") {
		t.Error("Module type missing 'importPath: String' field in schema.graphql")
	}
}

// TestSchema_ModuleHasVisibility verifies that the Module type in schema.graphql
// includes a visibility field.
// Expected result: schema contains "visibility: String" inside the Module type block.
func TestSchema_ModuleHasVisibility(t *testing.T) {
	content := readSchema(t)
	moduleBlock := extractTypeBlock(content, "Module")
	if moduleBlock == "" {
		t.Fatal("Module type block not found in schema.graphql")
	}
	if !strings.Contains(moduleBlock, "visibility: String") {
		t.Error("Module type missing 'visibility: String' field in schema.graphql")
	}
}

// TestSchema_ModuleHasKind verifies that the Module type in schema.graphql
// includes a kind field.
// Expected result: schema contains "kind: String" inside the Module type block.
func TestSchema_ModuleHasKind(t *testing.T) {
	content := readSchema(t)
	moduleBlock := extractTypeBlock(content, "Module")
	if moduleBlock == "" {
		t.Fatal("Module type block not found in schema.graphql")
	}
	if !strings.Contains(moduleBlock, "kind: String") {
		t.Error("Module type missing 'kind: String' field in schema.graphql")
	}
}

// === Task 2: Regenerate go-ormql code ===

const modelsGenPath = "../../internal/clients/code_db/generated/models_gen.go"
const augmentedSchemaPath = "../../internal/clients/code_db/generated/schema.graphql"

// TestCodegen_ModelsGenHasImportPath verifies that the generated models_gen.go
// contains an ImportPath field for Module-related structs.
// Expected result: models_gen.go contains "ImportPath" field.
func TestCodegen_ModelsGenHasImportPath(t *testing.T) {
	data, err := os.ReadFile(modelsGenPath)
	if err != nil {
		t.Fatalf("failed to read models_gen.go: %v", err)
	}
	content := string(data)
	if !strings.Contains(content, "ImportPath") {
		t.Error("generated models_gen.go missing 'ImportPath' field — run 'task generate' after schema change")
	}
}

// TestCodegen_ModelsGenHasModuleVisibility verifies that the generated models_gen.go
// contains a Visibility field for Module-related structs.
// Expected result: models_gen.go contains "Visibility" in a Module-related struct.
func TestCodegen_ModelsGenHasModuleVisibility(t *testing.T) {
	data, err := os.ReadFile(modelsGenPath)
	if err != nil {
		t.Fatalf("failed to read models_gen.go: %v", err)
	}
	content := string(data)
	// Look for Visibility in a Module merge input struct context
	if !strings.Contains(content, "ModuleMergeInput") {
		t.Fatal("generated models_gen.go missing 'ModuleMergeInput' — schema may not have been regenerated")
	}
	// The Module node already has language/startingLine/endingLine in merge inputs.
	// After adding visibility to schema, the merge input should include it.
	moduleBlock := extractGoStructBlock(content, "ModuleCreateInput")
	if moduleBlock == "" {
		moduleBlock = extractGoStructBlock(content, "ModuleMergeInput")
	}
	if !strings.Contains(moduleBlock, "Visibility") {
		t.Error("generated Module create/merge input missing 'Visibility' field — run 'task generate' after schema change")
	}
}

// TestCodegen_ModelsGenHasModuleKind verifies that the generated models_gen.go
// contains a Kind field for Module-related structs.
// Expected result: models_gen.go contains "Kind" in a Module-related struct.
func TestCodegen_ModelsGenHasModuleKind(t *testing.T) {
	data, err := os.ReadFile(modelsGenPath)
	if err != nil {
		t.Fatalf("failed to read models_gen.go: %v", err)
	}
	content := string(data)
	moduleBlock := extractGoStructBlock(content, "ModuleCreateInput")
	if moduleBlock == "" {
		moduleBlock = extractGoStructBlock(content, "ModuleMergeInput")
	}
	if !strings.Contains(moduleBlock, "Kind") {
		t.Error("generated Module create/merge input missing 'Kind' field — run 'task generate' after schema change")
	}
}

// TestCodegen_AugmentedSchemaHasModuleFields verifies that the augmented
// schema.graphql in generated/ includes the 3 new Module fields.
// Expected result: augmented schema contains importPath, visibility, kind on Module.
func TestCodegen_AugmentedSchemaHasModuleFields(t *testing.T) {
	data, err := os.ReadFile(augmentedSchemaPath)
	if err != nil {
		t.Fatalf("failed to read augmented schema.graphql: %v", err)
	}
	content := string(data)
	moduleBlock := extractTypeBlock(content, "Module")
	if moduleBlock == "" {
		t.Fatal("Module type block not found in augmented schema.graphql")
	}

	fields := []string{"importPath", "visibility", "kind"}
	for _, f := range fields {
		if !strings.Contains(moduleBlock, f) {
			t.Errorf("augmented schema Module type missing '%s' field", f)
		}
	}
}

// extractGoStructBlock extracts a Go struct block by name from source code.
func extractGoStructBlock(content, structName string) string {
	marker := "type " + structName + " struct"
	idx := strings.Index(content, marker)
	if idx < 0 {
		return ""
	}
	rest := content[idx:]
	braceStart := strings.Index(rest, "{")
	if braceStart < 0 {
		return ""
	}
	depth := 0
	for i := braceStart; i < len(rest); i++ {
		switch rest[i] {
		case '{':
			depth++
		case '}':
			depth--
			if depth == 0 {
				return rest[:i+1]
			}
		}
	}
	return ""
}

// === Task 11: Verify build, vet, and race detector ===

const extractorGoPath = "../../internal/analysis/extractor.go"
const analyzerGoPath = "../../internal/analysis/analyzer.go"
const typesGoPath = "../../internal/analysis/types.go"
const goExtractorPath = "../../internal/analysis/golang/extractor.go"
const tsExtractorPath = "../../internal/analysis/typescript/extractor.go"
const tsxExtractorPath = "../../internal/analysis/typescript/tsx_extractor.go"
const replCommandsPath = "../../internal/repl/commands.go"

// TestVerify_SymbolHasImportPathField verifies types.go has ImportPath field.
// Expected result: types.go contains "ImportPath string".
func TestVerify_SymbolHasImportPathField(t *testing.T) {
	data, err := os.ReadFile(typesGoPath)
	if err != nil {
		t.Fatalf("failed to read types.go: %v", err)
	}
	if !strings.Contains(string(data), "ImportPath") {
		t.Error("types.go Symbol struct missing 'ImportPath' field")
	}
}

// TestVerify_SymbolHasModuleKindField verifies types.go has ModuleKind field.
// Expected result: types.go contains "ModuleKind string".
func TestVerify_SymbolHasModuleKindField(t *testing.T) {
	data, err := os.ReadFile(typesGoPath)
	if err != nil {
		t.Fatalf("failed to read types.go: %v", err)
	}
	if !strings.Contains(string(data), "ModuleKind") {
		t.Error("types.go Symbol struct missing 'ModuleKind' field")
	}
}

// TestVerify_ExtractorHasRepoPathParam verifies the Extractor interface
// ExtractSymbols has 4 params including repoPath.
// Expected result: extractor.go contains "repoPath string" in the interface.
func TestVerify_ExtractorHasRepoPathParam(t *testing.T) {
	data, err := os.ReadFile(extractorGoPath)
	if err != nil {
		t.Fatalf("failed to read extractor.go: %v", err)
	}
	if !strings.Contains(string(data), "repoPath string") {
		t.Error("extractor.go missing 'repoPath string' in ExtractSymbols signature")
	}
}

// TestVerify_AnalyzerHasRepoPathParam verifies Analyze method has repoPath param.
// Expected result: analyzer.go contains "repoPath string" in the Analyze method.
func TestVerify_AnalyzerHasRepoPathParam(t *testing.T) {
	data, err := os.ReadFile(analyzerGoPath)
	if err != nil {
		t.Fatalf("failed to read analyzer.go: %v", err)
	}
	if !strings.Contains(string(data), "repoPath string") {
		t.Error("analyzer.go missing 'repoPath string' in Analyze signature")
	}
}

// TestVerify_GoExtractorHasRepoPathParam verifies Go extractor has repoPath.
// Expected result: golang/extractor.go contains "repoPath string".
func TestVerify_GoExtractorHasRepoPathParam(t *testing.T) {
	data, err := os.ReadFile(goExtractorPath)
	if err != nil {
		t.Fatalf("failed to read golang/extractor.go: %v", err)
	}
	if !strings.Contains(string(data), "repoPath string") {
		t.Error("golang/extractor.go missing 'repoPath string' in ExtractSymbols")
	}
}

// TestVerify_TSExtractorHasRepoPathParam verifies TS extractor has repoPath.
// Expected result: typescript/extractor.go contains "repoPath string".
func TestVerify_TSExtractorHasRepoPathParam(t *testing.T) {
	data, err := os.ReadFile(tsExtractorPath)
	if err != nil {
		t.Fatalf("failed to read typescript/extractor.go: %v", err)
	}
	if !strings.Contains(string(data), "repoPath string") {
		t.Error("typescript/extractor.go missing 'repoPath string' in ExtractSymbols")
	}
}

// TestVerify_TSXExtractorHasRepoPathParam verifies TSX extractor has repoPath.
// Expected result: typescript/tsx_extractor.go contains "repoPath string".
func TestVerify_TSXExtractorHasRepoPathParam(t *testing.T) {
	data, err := os.ReadFile(tsxExtractorPath)
	if err != nil {
		t.Fatalf("failed to read typescript/tsx_extractor.go: %v", err)
	}
	if !strings.Contains(string(data), "repoPath string") {
		t.Error("typescript/tsx_extractor.go missing 'repoPath string' in ExtractSymbols")
	}
}

// TestVerify_REPLPassesRepoPath verifies REPL passes path to Analyze.
// Expected result: commands.go contains "Analyze(ctx, result.RepoID, path,".
func TestVerify_REPLPassesRepoPath(t *testing.T) {
	data, err := os.ReadFile(replCommandsPath)
	if err != nil {
		t.Fatalf("failed to read commands.go: %v", err)
	}
	if !strings.Contains(string(data), "Analyze(ctx, result.RepoID, path,") {
		t.Error("commands.go should pass 'path' as repoPath to Analyzer.Analyze()")
	}
}

// TestVerify_GoExtractorHasResolveGoModuleName verifies the Go extractor
// has a resolveGoModuleName helper.
// Expected result: golang/extractor.go contains "resolveGoModuleName".
func TestVerify_GoExtractorHasResolveGoModuleName(t *testing.T) {
	data, err := os.ReadFile(goExtractorPath)
	if err != nil {
		t.Fatalf("failed to read golang/extractor.go: %v", err)
	}
	if !strings.Contains(string(data), "resolveGoModuleName") {
		t.Error("golang/extractor.go missing 'resolveGoModuleName' helper function")
	}
}

// TestVerify_TSExtractorHasDetectModuleKind verifies the TypeScript extractor
// has a detectModuleKind helper.
// Expected result: typescript/extractor.go contains "detectModuleKind".
func TestVerify_TSExtractorHasDetectModuleKind(t *testing.T) {
	data, err := os.ReadFile(tsExtractorPath)
	if err != nil {
		t.Fatalf("failed to read typescript/extractor.go: %v", err)
	}
	if !strings.Contains(string(data), "detectModuleKind") {
		t.Error("typescript/extractor.go missing 'detectModuleKind' helper function")
	}
}

// readSchema reads the schema.graphql source-of-truth file.
func readSchema(t *testing.T) string {
	t.Helper()
	data, err := os.ReadFile(schemaPath)
	if err != nil {
		t.Fatalf("failed to read schema.graphql: %v", err)
	}
	return string(data)
}

// extractTypeBlock extracts the content between "type <name>" and the closing "}".
func extractTypeBlock(content, typeName string) string {
	marker := "type " + typeName
	idx := strings.Index(content, marker)
	if idx < 0 {
		return ""
	}
	rest := content[idx:]
	braceStart := strings.Index(rest, "{")
	if braceStart < 0 {
		return ""
	}
	depth := 0
	for i := braceStart; i < len(rest); i++ {
		switch rest[i] {
		case '{':
			depth++
		case '}':
			depth--
			if depth == 0 {
				return rest[:i+1]
			}
		}
	}
	return ""
}
