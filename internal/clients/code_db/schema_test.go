package codedb_test

import (
	"os"
	"strings"
	"testing"
)

// readSchema is a test helper that reads the schema.graphql file from the package directory.
func readSchema(t *testing.T) string {
	t.Helper()
	data, err := os.ReadFile("schema.graphql")
	if err != nil {
		t.Fatalf("failed to read schema.graphql: %v", err)
	}
	return string(data)
}

// schemaContains is a test helper that asserts the schema contains a given pattern.
func schemaContains(t *testing.T, schema, pattern, description string) {
	t.Helper()
	if !strings.Contains(schema, pattern) {
		t.Errorf("schema.graphql missing %s (expected pattern: %q)", description, pattern)
	}
}

// schemaNotContains is a test helper that asserts the schema does NOT contain a given pattern.
func schemaNotContains(t *testing.T, schema, pattern, description string) {
	t.Helper()
	if strings.Contains(schema, pattern) {
		t.Errorf("schema.graphql should not contain %s (found pattern: %q)", description, pattern)
	}
}

// TestSchemaNodeTypes verifies that schema.graphql defines all 6 required node types
// from the knowledge graph spec: Repository, Folder, File, Module, Class, Function.
// Expected result: All 6 types are present with @node directive.
func TestSchemaNodeTypes(t *testing.T) {
	schema := readSchema(t)

	nodeTypes := []struct {
		name string
	}{
		{"Repository"},
		{"Folder"},
		{"File"},
		{"Module"},
		{"Class"},
		{"Function"},
	}

	for _, tt := range nodeTypes {
		t.Run(tt.name, func(t *testing.T) {
			pattern := "type " + tt.name + " @node"
			schemaContains(t, schema, pattern, "node type "+tt.name)
		})
	}
}

// TestSchemaCallProperties verifies that CallProperties relationship type exists
// with the callType field for typed cross-repo CALLS edges.
// Expected result: CallProperties @relationshipProperties with callType: String field.
func TestSchemaCallProperties(t *testing.T) {
	schema := readSchema(t)

	schemaContains(t, schema, "type CallProperties @relationshipProperties", "CallProperties relationship type")
	schemaContains(t, schema, "callType: String", "callType field on CallProperties")
}

// TestSchemaNoVectorIndexes verifies that @vector annotations have been removed
// from the schema. Embedding fields are no longer present.
// Expected result: No @vector directives, no embedding fields.
func TestSchemaNoVectorIndexes(t *testing.T) {
	schema := readSchema(t)

	schemaNotContains(t, schema, "@vector", "@vector directive (embedding removed)")
	schemaNotContains(t, schema, "embedding", "embedding field (embedding removed)")
}

// TestSchemaModuleNoVector verifies that the Module node type does NOT have
// an embedding field or @vector directive. Module nodes are structural only.
// Expected result: No module_embedding index, no embedding field on Module.
func TestSchemaModuleNoVector(t *testing.T) {
	schema := readSchema(t)

	// Module should NOT have a vector index
	schemaNotContains(t, schema, "module_embedding", "module_embedding vector index (Module is structural only)")

	// Find the Module type block and verify no embedding field
	typeStart := strings.Index(schema, "type Module @node")
	if typeStart == -1 {
		t.Fatal("type Module not found in schema")
	}
	typeBlock := schema[typeStart:]
	closingBrace := strings.Index(typeBlock, "}")
	if closingBrace == -1 {
		t.Fatal("could not find closing brace for type Module")
	}
	typeBlock = typeBlock[:closingBrace+1]

	if strings.Contains(typeBlock, "embedding") {
		t.Error("Module type should NOT have an embedding field (Module is structural only)")
	}
	if strings.Contains(typeBlock, "@vector") {
		t.Error("Module type should NOT have @vector directive (Module is structural only)")
	}
}

// TestSchemaBelongsToRelationships verifies that every non-Repository node type
// has a BELONGS_TO → Repository relationship for cross-repo query denormalization.
// Expected result: Folder, File, Module, Class, Function all have BELONGS_TO.
func TestSchemaBelongsToRelationships(t *testing.T) {
	schema := readSchema(t)

	// Each of these types must have a repository field with BELONGS_TO
	types := []string{"Folder", "File", "Module", "Class", "Function"}

	for _, typeName := range types {
		t.Run(typeName, func(t *testing.T) {
			// Find the type block and check it contains BELONGS_TO
			typeStart := strings.Index(schema, "type "+typeName+" @node")
			if typeStart == -1 {
				t.Fatalf("type %s not found in schema", typeName)
			}
			// Find the closing brace of this type block
			typeBlock := schema[typeStart:]
			closingBrace := strings.Index(typeBlock, "}")
			if closingBrace == -1 {
				t.Fatalf("could not find closing brace for type %s", typeName)
			}
			typeBlock = typeBlock[:closingBrace+1]

			if !strings.Contains(typeBlock, "BELONGS_TO") {
				t.Errorf("type %s missing BELONGS_TO → Repository relationship", typeName)
			}
		})
	}
}

// TestSchemaKeyRelationships verifies that all key relationship types exist in the schema.
// Expected result: All relationship types from the spec are present.
func TestSchemaKeyRelationships(t *testing.T) {
	schema := readSchema(t)

	relationships := []struct {
		name    string
		pattern string
	}{
		{"CONTAINS", `type: "CONTAINS"`},
		{"BELONGS_TO", `type: "BELONGS_TO"`},
		{"DEFINES", `type: "DEFINES"`},
		{"IMPORTS", `type: "IMPORTS"`},
		{"EXPORTS", `type: "EXPORTS"`},
		{"DEPENDS_ON", `type: "DEPENDS_ON"`},
		{"HAS_METHOD", `type: "HAS_METHOD"`},
		{"INHERITS", `type: "INHERITS"`},
		{"IMPLEMENTS", `type: "IMPLEMENTS"`},
		{"CALLS", `type: "CALLS"`},
		{"OVERRIDES", `type: "OVERRIDES"`},
		{"HAS_MODULE", `type: "HAS_MODULE"`},
	}

	for _, tt := range relationships {
		t.Run(tt.name, func(t *testing.T) {
			schemaContains(t, schema, tt.pattern, "relationship "+tt.name)
		})
	}
}

// TestSchemaNoLegacyTypes verifies that the old placeholder types (RepositoryFolder,
// RepositoryFile) have been removed and renamed to Folder and File.
// Expected result: No RepositoryFolder or RepositoryFile types exist.
func TestSchemaNoLegacyTypes(t *testing.T) {
	schema := readSchema(t)

	schemaNotContains(t, schema, "RepositoryFolder", "legacy RepositoryFolder type (should be Folder)")
	schemaNotContains(t, schema, "RepositoryFile", "legacy RepositoryFile type (should be File)")
}

// TestSchemaKeyFields verifies that key fields exist on each node type.
// Expected result: Each type has its required fields from the spec.
func TestSchemaKeyFields(t *testing.T) {
	schema := readSchema(t)

	tests := []struct {
		typeName string
		fields   []string
	}{
		{"Repository", []string{"id: ID!", "name: String!", "lastIndexed: DateTime!"}},
		{"Folder", []string{"id: ID!", "path: String!", "lastUpdated: DateTime!"}},
		{"File", []string{"id: ID!", "path: String!", "language: String", "lineCount: Int", "lastUpdated: DateTime!"}},
		{"Module", []string{"id: ID!", "name: String!", "path: String!"}},
		{"Class", []string{"id: ID!", "name: String!", "kind: String!", "source: String"}},
		{"Function", []string{"id: ID!", "name: String!", "signature: String", "cyclomaticComplexity: Int"}},
	}

	for _, tt := range tests {
		t.Run(tt.typeName, func(t *testing.T) {
			typeStart := strings.Index(schema, "type "+tt.typeName+" @node")
			if typeStart == -1 {
				t.Fatalf("type %s not found in schema", tt.typeName)
			}
			typeBlock := schema[typeStart:]
			closingBrace := strings.Index(typeBlock, "}")
			if closingBrace == -1 {
				t.Fatalf("could not find closing brace for type %s", tt.typeName)
			}
			typeBlock = typeBlock[:closingBrace+1]

			for _, field := range tt.fields {
				if !strings.Contains(typeBlock, field) {
					t.Errorf("type %s missing field %q", tt.typeName, field)
				}
			}
		})
	}
}
