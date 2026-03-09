package verify_test

import (
	"strings"
	"testing"
)

// === Task 1: Schema changes — Graph Node Attribute Enrichment ===
//
// Verifies that schema.graphql has the new/changed fields:
// - Repository: path: String
// - File: filename: String
// - Module: startingLine: Int, endingLine: Int
// - Class: startingLine/endingLine replace lineNumber/lineCount
// - Function: startingLine/endingLine replace lineNumber/lineCount

// TestSchema_RepositoryHasPath verifies that the Repository node type
// in schema.graphql includes a "path" field for the absolute filesystem path.
// Expected result: schema.graphql contains "path: String" within the Repository type.
func TestSchema_RepositoryHasPath(t *testing.T) {
	content := readProjectFile(t, "internal/clients/code_db/schema.graphql")

	// Extract the Repository type block
	repoBlock := extractTypeBlock(t, content, "Repository")
	if !strings.Contains(repoBlock, "path:") {
		t.Error("Repository type in schema.graphql is missing 'path' field — expected 'path: String' for absolute filesystem path")
	}
}

// TestSchema_FileHasFilename verifies that the File node type
// in schema.graphql includes a "filename" field.
// Expected result: schema.graphql contains "filename: String" within the File type.
func TestSchema_FileHasFilename(t *testing.T) {
	content := readProjectFile(t, "internal/clients/code_db/schema.graphql")

	fileBlock := extractTypeBlock(t, content, "File")
	if !strings.Contains(fileBlock, "filename:") {
		t.Error("File type in schema.graphql is missing 'filename' field — expected 'filename: String' for base file name")
	}
}

// TestSchema_ModuleHasStartingLine verifies that the Module node type
// in schema.graphql includes a "startingLine" field.
// Expected result: schema.graphql contains "startingLine: Int" within the Module type.
func TestSchema_ModuleHasStartingLine(t *testing.T) {
	content := readProjectFile(t, "internal/clients/code_db/schema.graphql")

	moduleBlock := extractTypeBlock(t, content, "Module")
	if !strings.Contains(moduleBlock, "startingLine:") {
		t.Error("Module type in schema.graphql is missing 'startingLine' field")
	}
}

// TestSchema_ModuleHasEndingLine verifies that the Module node type
// in schema.graphql includes an "endingLine" field.
// Expected result: schema.graphql contains "endingLine: Int" within the Module type.
func TestSchema_ModuleHasEndingLine(t *testing.T) {
	content := readProjectFile(t, "internal/clients/code_db/schema.graphql")

	moduleBlock := extractTypeBlock(t, content, "Module")
	if !strings.Contains(moduleBlock, "endingLine:") {
		t.Error("Module type in schema.graphql is missing 'endingLine' field")
	}
}

// TestSchema_ClassHasStartingLine verifies that the Class node type
// in schema.graphql uses "startingLine" instead of "lineNumber".
// Expected result: Class type contains "startingLine: Int", not "lineNumber".
func TestSchema_ClassHasStartingLine(t *testing.T) {
	content := readProjectFile(t, "internal/clients/code_db/schema.graphql")

	classBlock := extractTypeBlock(t, content, "Class")
	if !strings.Contains(classBlock, "startingLine:") {
		t.Error("Class type in schema.graphql is missing 'startingLine' field — should replace 'lineNumber'")
	}
}

// TestSchema_ClassHasEndingLine verifies that the Class node type
// in schema.graphql uses "endingLine" instead of "lineCount".
// Expected result: Class type contains "endingLine: Int", not "lineCount".
func TestSchema_ClassHasEndingLine(t *testing.T) {
	content := readProjectFile(t, "internal/clients/code_db/schema.graphql")

	classBlock := extractTypeBlock(t, content, "Class")
	if !strings.Contains(classBlock, "endingLine:") {
		t.Error("Class type in schema.graphql is missing 'endingLine' field — should replace 'lineCount'")
	}
}

// TestSchema_ClassNoLineNumber verifies that the Class node type
// no longer contains the old "lineNumber" field.
// Expected result: Class type does NOT contain "lineNumber".
func TestSchema_ClassNoLineNumber(t *testing.T) {
	content := readProjectFile(t, "internal/clients/code_db/schema.graphql")

	classBlock := extractTypeBlock(t, content, "Class")
	if strings.Contains(classBlock, "lineNumber:") {
		t.Error("Class type in schema.graphql still contains 'lineNumber' — should be replaced with 'startingLine'")
	}
}

// TestSchema_ClassNoLineCount verifies that the Class node type
// no longer contains the old "lineCount" field.
// Expected result: Class type does NOT contain "lineCount".
func TestSchema_ClassNoLineCount(t *testing.T) {
	content := readProjectFile(t, "internal/clients/code_db/schema.graphql")

	classBlock := extractTypeBlock(t, content, "Class")
	if strings.Contains(classBlock, "lineCount:") {
		t.Error("Class type in schema.graphql still contains 'lineCount' — should be replaced with 'endingLine'")
	}
}

// TestSchema_FunctionHasStartingLine verifies that the Function node type
// in schema.graphql uses "startingLine" instead of "lineNumber".
// Expected result: Function type contains "startingLine: Int".
func TestSchema_FunctionHasStartingLine(t *testing.T) {
	content := readProjectFile(t, "internal/clients/code_db/schema.graphql")

	funcBlock := extractTypeBlock(t, content, "Function")
	if !strings.Contains(funcBlock, "startingLine:") {
		t.Error("Function type in schema.graphql is missing 'startingLine' field — should replace 'lineNumber'")
	}
}

// TestSchema_FunctionHasEndingLine verifies that the Function node type
// in schema.graphql uses "endingLine" instead of "lineCount".
// Expected result: Function type contains "endingLine: Int".
func TestSchema_FunctionHasEndingLine(t *testing.T) {
	content := readProjectFile(t, "internal/clients/code_db/schema.graphql")

	funcBlock := extractTypeBlock(t, content, "Function")
	if !strings.Contains(funcBlock, "endingLine:") {
		t.Error("Function type in schema.graphql is missing 'endingLine' field — should replace 'lineCount'")
	}
}

// TestSchema_FunctionNoLineNumber verifies that the Function node type
// no longer contains the old "lineNumber" field.
// Expected result: Function type does NOT contain "lineNumber".
func TestSchema_FunctionNoLineNumber(t *testing.T) {
	content := readProjectFile(t, "internal/clients/code_db/schema.graphql")

	funcBlock := extractTypeBlock(t, content, "Function")
	if strings.Contains(funcBlock, "lineNumber:") {
		t.Error("Function type in schema.graphql still contains 'lineNumber' — should be replaced with 'startingLine'")
	}
}

// TestSchema_FunctionNoLineCount verifies that the Function node type
// no longer contains the old "lineCount" field (not to be confused with File.lineCount).
// Expected result: Function type does NOT contain "lineCount".
func TestSchema_FunctionNoLineCount(t *testing.T) {
	content := readProjectFile(t, "internal/clients/code_db/schema.graphql")

	funcBlock := extractTypeBlock(t, content, "Function")
	if strings.Contains(funcBlock, "lineCount:") {
		t.Error("Function type in schema.graphql still contains 'lineCount' — should be replaced with 'endingLine'")
	}
}

// === Task 2: Regenerate go-ormql code ===
// After schema changes, generated code should reflect new fields.

// TestGenerated_ModelsHasStartingLine verifies that the regenerated models
// contain a StartingLine field (from the schema change).
// Expected result: models_gen.go contains "StartingLine".
func TestGenerated_ModelsHasStartingLine(t *testing.T) {
	content := readProjectFile(t, "internal/clients/code_db/generated/models_gen.go")

	if !strings.Contains(content, "StartingLine") {
		t.Error("models_gen.go is missing 'StartingLine' field — regeneration needed after schema change")
	}
}

// TestGenerated_ModelsHasEndingLine verifies that the regenerated models
// contain an EndingLine field.
// Expected result: models_gen.go contains "EndingLine".
func TestGenerated_ModelsHasEndingLine(t *testing.T) {
	content := readProjectFile(t, "internal/clients/code_db/generated/models_gen.go")

	if !strings.Contains(content, "EndingLine") {
		t.Error("models_gen.go is missing 'EndingLine' field — regeneration needed after schema change")
	}
}

// TestGenerated_ModelsNoLineNumber verifies that the regenerated models
// no longer contain a LineNumber field on Function/Class.
// Expected result: models_gen.go does NOT contain "LineNumber".
func TestGenerated_ModelsNoLineNumber(t *testing.T) {
	content := readProjectFile(t, "internal/clients/code_db/generated/models_gen.go")

	if strings.Contains(content, "LineNumber") {
		t.Error("models_gen.go still contains 'LineNumber' — regeneration needed after schema change (should be StartingLine)")
	}
}

// TestGenerated_ModelsHasFilename verifies that the regenerated File model
// contains a Filename field.
// Expected result: models_gen.go contains "Filename".
func TestGenerated_ModelsHasFilename(t *testing.T) {
	content := readProjectFile(t, "internal/clients/code_db/generated/models_gen.go")

	if !strings.Contains(content, "Filename") {
		t.Error("models_gen.go is missing 'Filename' field on File model — regeneration needed after schema change")
	}
}

// TestGenerated_RepositoryHasPath verifies that the regenerated Repository model
// contains a Path field.
// Expected result: models_gen.go contains Path in the Repository struct.
func TestGenerated_RepositoryHasPath(t *testing.T) {
	content := readProjectFile(t, "internal/clients/code_db/generated/models_gen.go")

	// The Repository struct should have a Path field
	if !strings.Contains(content, "Path") {
		t.Error("models_gen.go is missing 'Path' field — regeneration needed after schema change (Repository.path)")
	}
}

// === Task 11: Verify no old field references in implementation ===

// TestAnalyzerNoLineNumberReference verifies that analyzer.go no longer
// references "lineNumber" in field maps (should use "startingLine").
// Expected result: analyzer.go does NOT contain the string "lineNumber" as a map key.
func TestAnalyzerNoLineNumberReference(t *testing.T) {
	content := readProjectFile(t, "internal/analysis/analyzer.go")

	if strings.Contains(content, `"lineNumber"`) {
		t.Error("analyzer.go still contains '\"lineNumber\"' — should be replaced with '\"startingLine\"'")
	}
}

// TestAnalyzerNoLineCountReference verifies that analyzer.go no longer
// references "lineCount" in field maps (should use "endingLine").
// Expected result: analyzer.go does NOT contain the string "lineCount" as a map key.
func TestAnalyzerNoLineCountReference(t *testing.T) {
	content := readProjectFile(t, "internal/analysis/analyzer.go")

	if strings.Contains(content, `"lineCount"`) {
		t.Error("analyzer.go still contains '\"lineCount\"' — should be replaced with '\"endingLine\"'")
	}
}

// TestMCPTypesNoLineNumber verifies that mcp/types.go no longer has
// a LineNumber field (should use StartingLine).
// Expected result: types.go does NOT contain "LineNumber".
func TestMCPTypesNoLineNumber(t *testing.T) {
	content := readProjectFile(t, "internal/mcp/types.go")

	if strings.Contains(content, "LineNumber") {
		t.Error("mcp/types.go still contains 'LineNumber' field — should be replaced with 'StartingLine'")
	}
}

// TestMCPTypesNoLineCount verifies that mcp/types.go no longer has
// a LineCount field on SearchResult (should use EndingLine).
// Note: File results may still reference lineCount from the File node,
// but SearchResult should use EndingLine.
// Expected result: types.go does NOT contain "LineCount" as a struct field.
func TestMCPTypesNoLineCount(t *testing.T) {
	content := readProjectFile(t, "internal/mcp/types.go")

	if strings.Contains(content, "LineCount") {
		t.Error("mcp/types.go still contains 'LineCount' field — should be replaced with 'EndingLine'")
	}
}

// TestMCPTypesHasStartingLine verifies that mcp/types.go has a
// StartingLine field on SearchResult.
// Expected result: types.go contains "StartingLine".
func TestMCPTypesHasStartingLine(t *testing.T) {
	content := readProjectFile(t, "internal/mcp/types.go")

	if !strings.Contains(content, "StartingLine") {
		t.Error("mcp/types.go is missing 'StartingLine' field on SearchResult")
	}
}

// TestMCPTypesHasEndingLine verifies that mcp/types.go has an
// EndingLine field on SearchResult.
// Expected result: types.go contains "EndingLine".
func TestMCPTypesHasEndingLine(t *testing.T) {
	content := readProjectFile(t, "internal/mcp/types.go")

	if !strings.Contains(content, "EndingLine") {
		t.Error("mcp/types.go is missing 'EndingLine' field on SearchResult")
	}
}

// TestMCPToolsNoLineNumberQuery verifies that mcp/tools.go no longer
// contains "lineNumber" in GraphQL query strings.
// Expected result: tools.go does NOT contain "lineNumber".
func TestMCPToolsNoLineNumberQuery(t *testing.T) {
	content := readProjectFile(t, "internal/mcp/tools.go")

	if strings.Contains(content, "lineNumber") {
		t.Error("mcp/tools.go still contains 'lineNumber' in queries — should be replaced with 'startingLine'")
	}
}

// TestMCPToolsNoLineCountQuery verifies that mcp/tools.go no longer
// contains "lineCount" in GraphQL query strings for Function/Class queries.
// Note: gqlListFiles may still query lineCount from the File node (File.lineCount is unchanged).
// Expected result: tools.go Function/Class queries do NOT contain "lineCount".
func TestMCPToolsNoLineCountQuery(t *testing.T) {
	content := readProjectFile(t, "internal/mcp/tools.go")

	// Check Function query
	if strings.Contains(content, "gqlFindFunctions") {
		funcQuery := extractConstBlock(content, "gqlFindFunctions")
		if strings.Contains(funcQuery, "lineCount") {
			t.Error("gqlFindFunctions query in tools.go still contains 'lineCount' — should use 'endingLine'")
		}
	}
	// Check similarity queries
	if strings.Contains(content, "gqlFunctionsSimilar") {
		funcSimilarQuery := extractConstBlock(content, "gqlFunctionsSimilar")
		if strings.Contains(funcSimilarQuery, "lineCount") {
			t.Error("gqlFunctionsSimilar query in tools.go still contains 'lineCount' — should use 'endingLine'")
		}
	}
	if strings.Contains(content, "gqlClassesSimilar") {
		classSimilarQuery := extractConstBlock(content, "gqlClassesSimilar")
		if strings.Contains(classSimilarQuery, "lineCount") {
			t.Error("gqlClassesSimilar query in tools.go still contains 'lineCount' — should use 'endingLine'")
		}
	}
}

// TestMCPToolsHasStartingLineQuery verifies that mcp/tools.go queries
// contain "startingLine" instead of "lineNumber".
// Expected result: tools.go contains "startingLine" in function/class queries.
func TestMCPToolsHasStartingLineQuery(t *testing.T) {
	content := readProjectFile(t, "internal/mcp/tools.go")

	if !strings.Contains(content, "startingLine") {
		t.Error("mcp/tools.go is missing 'startingLine' in queries — should replace 'lineNumber'")
	}
}

// TestMCPToolsHasEndingLineQuery verifies that mcp/tools.go queries
// contain "endingLine" instead of "lineCount" for Function/Class queries.
// Expected result: tools.go contains "endingLine" in function/class queries.
func TestMCPToolsHasEndingLineQuery(t *testing.T) {
	content := readProjectFile(t, "internal/mcp/tools.go")

	if !strings.Contains(content, "endingLine") {
		t.Error("mcp/tools.go is missing 'endingLine' in queries — should replace 'lineCount' in Function/Class queries")
	}
}

// === Helpers ===

// extractTypeBlock extracts the content between "type <name> @node {" and the
// closing "}" for a GraphQL type definition. Returns the block content.
func extractTypeBlock(t *testing.T, schema, typeName string) string {
	t.Helper()
	marker := "type " + typeName + " @node"
	idx := strings.Index(schema, marker)
	if idx == -1 {
		t.Fatalf("type %s not found in schema", typeName)
	}
	rest := schema[idx:]
	// Find the closing brace (simple: count braces)
	depth := 0
	for i, ch := range rest {
		if ch == '{' {
			depth++
		}
		if ch == '}' {
			depth--
			if depth == 0 {
				return rest[:i+1]
			}
		}
	}
	t.Fatalf("unclosed type block for %s", typeName)
	return ""
}

// extractConstBlock extracts the value of a const declaration from Go source.
// Returns the string between the const name's backtick delimiters.
func extractConstBlock(content, constName string) string {
	idx := strings.Index(content, constName)
	if idx == -1 {
		return ""
	}
	rest := content[idx:]
	// Find first backtick
	start := strings.Index(rest, "`")
	if start == -1 {
		return ""
	}
	rest = rest[start+1:]
	end := strings.Index(rest, "`")
	if end == -1 {
		return ""
	}
	return rest[:end]
}
