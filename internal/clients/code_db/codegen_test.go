package codedb_test

import (
	"os"
	"testing"
)

// TestGeneratedFilesExist verifies that go-ormql code generation has produced
// the expected files in the generated/ directory. These files are build artifacts
// created by `gormql generate`.
// Expected result: All 5 expected generated files exist.
func TestGeneratedFilesExist(t *testing.T) {
	expectedFiles := []struct {
		name string
		path string
	}{
		{"models_gen.go", "generated/models_gen.go"},
		{"graphmodel_gen.go", "generated/graphmodel_gen.go"},
		{"client_gen.go", "generated/client_gen.go"},
		{"indexes_gen.go", "generated/indexes_gen.go"},
		{"augmented schema", "generated/schema.graphql"},
	}

	for _, f := range expectedFiles {
		t.Run(f.name, func(t *testing.T) {
			if _, err := os.Stat(f.path); os.IsNotExist(err) {
				t.Errorf("generated file missing: %s (run 'task generate' to create)", f.path)
			}
		})
	}
}

// TestGeneratedDirectoryExists verifies the generated/ output directory exists.
// Expected result: The generated/ directory exists.
func TestGeneratedDirectoryExists(t *testing.T) {
	info, err := os.Stat("generated")
	if os.IsNotExist(err) {
		t.Fatal("generated/ directory does not exist (run 'task generate' to create)")
	}
	if err != nil {
		t.Fatalf("error checking generated/ directory: %v", err)
	}
	if !info.IsDir() {
		t.Fatal("generated is not a directory")
	}
}
