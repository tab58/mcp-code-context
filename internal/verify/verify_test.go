package verify_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// findProjectRoot walks up from the current working directory to find the
// project root (the directory containing go.mod).
func findProjectRoot(t *testing.T) string {
	t.Helper()
	dir, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to get working directory: %v", err)
	}
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			t.Fatal("could not find project root (no go.mod found)")
		}
		dir = parent
	}
}

// readProjectFile reads a file relative to the project root.
func readProjectFile(t *testing.T, relPath string) string {
	t.Helper()
	root := findProjectRoot(t)
	data, err := os.ReadFile(filepath.Join(root, relPath))
	if err != nil {
		t.Fatalf("failed to read %s: %v", relPath, err)
	}
	return string(data)
}

// projectFileExists checks if a file exists relative to the project root.
func projectFileExists(t *testing.T, relPath string) bool {
	t.Helper()
	root := findProjectRoot(t)
	_, err := os.Stat(filepath.Join(root, relPath))
	return !os.IsNotExist(err)
}

// --- Existing: go-ormql dependency ---

// TestGoModContainsOrmql verifies that go.mod includes the go-ormql dependency.
// Expected result: go.mod contains a require line for github.com/tab58/go-ormql.
func TestGoModContainsOrmql(t *testing.T) {
	gomod := readProjectFile(t, "go.mod")

	if !strings.Contains(gomod, "github.com/tab58/go-ormql") {
		t.Error("go.mod missing go-ormql dependency (expected: github.com/tab58/go-ormql)")
	}
}

// --- Task 3: FalkorDB and tree-sitter dependencies in go.mod ---

// TestGoModContainsTreeSitter verifies that go.mod includes the tree-sitter dependency.
// Expected result: go.mod contains smacker/go-tree-sitter.
func TestGoModContainsTreeSitter(t *testing.T) {
	gomod := readProjectFile(t, "go.mod")

	if !strings.Contains(gomod, "github.com/smacker/go-tree-sitter") {
		t.Error("go.mod missing tree-sitter dependency (expected: github.com/smacker/go-tree-sitter)")
	}
}

// TestGoModContainsGitignoreLib verifies that go.mod includes a gitignore library.
// Expected result: go.mod contains sabhiram/go-gitignore.
func TestGoModContainsGitignoreLib(t *testing.T) {
	gomod := readProjectFile(t, "go.mod")

	if !strings.Contains(gomod, "github.com/sabhiram/go-gitignore") {
		t.Error("go.mod missing gitignore library (expected: github.com/sabhiram/go-gitignore)")
	}
}

// TestGoModContainsFalkorDBDriver verifies that go.mod includes a FalkorDB driver
// dependency (either directly or transitively via go-ormql with --target falkordb).
// Expected result: go.mod contains falkordb reference.
func TestGoModContainsFalkorDBDriver(t *testing.T) {
	gomod := readProjectFile(t, "go.mod")

	if !strings.Contains(gomod, "falkordb") {
		t.Error("go.mod missing FalkorDB driver dependency (expected: falkordb reference in go.mod)")
	}
}

// --- Existing + Task 4: Taskfile generate and build tasks ---

// TestTaskfileHasGenerateTask verifies that Taskfile.yml defines a 'generate' task
// that runs gormql generate with the correct schema path and output directory.
// Expected result: Taskfile.yml contains a generate task with the correct command.
func TestTaskfileHasGenerateTask(t *testing.T) {
	taskfile := readProjectFile(t, "Taskfile.yml")

	if !strings.Contains(taskfile, "generate:") {
		t.Error("Taskfile.yml missing 'generate' task definition")
	}
	if !strings.Contains(taskfile, "gormql generate") {
		t.Error("Taskfile.yml generate task missing 'gormql generate' command")
	}
	if !strings.Contains(taskfile, "internal/clients/code_db/schema.graphql") {
		t.Error("Taskfile.yml generate task missing schema path")
	}
	if !strings.Contains(taskfile, "internal/clients/code_db/generated") {
		t.Error("Taskfile.yml generate task missing output directory")
	}
}

// TestTaskfileHasBuildTask verifies that Taskfile.yml defines a 'build' task.
// Expected result: Taskfile.yml contains a build task.
func TestTaskfileHasBuildTask(t *testing.T) {
	taskfile := readProjectFile(t, "Taskfile.yml")

	if !strings.Contains(taskfile, "build:") {
		t.Error("Taskfile.yml missing 'build' task definition")
	}
	if !strings.Contains(taskfile, "go build") {
		t.Error("Taskfile.yml build task missing 'go build' command")
	}
}

// TestTaskfileGenerateHasFalkorDBTarget verifies that the generate task includes
// the --target falkordb flag for FalkorDB-compatible code generation.
// Expected result: Taskfile.yml generate command includes --target falkordb.
func TestTaskfileGenerateHasFalkorDBTarget(t *testing.T) {
	taskfile := readProjectFile(t, "Taskfile.yml")

	if !strings.Contains(taskfile, "--target falkordb") {
		t.Error("Taskfile.yml generate task missing '--target falkordb' flag")
	}
}

// TestTaskfileBuildHasCGo verifies that the build task sets CGO_ENABLED=1
// for tree-sitter and llama.cpp dependencies.
// Expected result: Taskfile.yml build task includes CGO_ENABLED=1.
func TestTaskfileBuildHasCGo(t *testing.T) {
	taskfile := readProjectFile(t, "Taskfile.yml")

	if !strings.Contains(taskfile, "CGO_ENABLED") {
		t.Error("Taskfile.yml build task missing CGO_ENABLED setting")
	}
}

// --- Task 5: Codegen with FalkorDB ---

// TestGeneratedCodeNoModuleEmbedding verifies that the generated indexes_gen.go
// does NOT contain a module_embedding vector index after regeneration with FalkorDB.
// Expected result: indexes_gen.go has only 2 vector indexes (function, class).
func TestGeneratedCodeNoModuleEmbedding(t *testing.T) {
	indexesContent := readProjectFile(t, "internal/clients/code_db/generated/indexes_gen.go")

	if strings.Contains(indexesContent, "module_embedding") {
		t.Error("indexes_gen.go should NOT contain module_embedding after schema update")
	}
}

// TestGeneratedCodeNoVectorIndexes verifies embedding vector indexes are removed.
// Expected result: indexes_gen.go does NOT contain function_embedding or class_embedding.
func TestGeneratedCodeNoVectorIndexes(t *testing.T) {
	indexesContent := readProjectFile(t, "internal/clients/code_db/generated/indexes_gen.go")

	if strings.Contains(indexesContent, "function_embedding") {
		t.Error("indexes_gen.go still contains function_embedding — embeddings removed")
	}
	if strings.Contains(indexesContent, "class_embedding") {
		t.Error("indexes_gen.go still contains class_embedding — embeddings removed")
	}
}

// --- Task 7: cmd/main.go exists ---

// TestCmdMainGoExists verifies that cmd/main.go exists as the application entry point.
// Expected result: cmd/main.go file exists.
func TestCmdMainGoExists(t *testing.T) {
	if !projectFileExists(t, "cmd/codectx/main.go") {
		t.Error("cmd/main.go does not exist (needed for startup sequence)")
	}
}

// TestCmdMainGoImportsConfig verifies that cmd/main.go imports the config package
// for LoadFalkorDBConfig().
// Expected result: cmd/main.go contains config import.
func TestCmdMainGoImportsConfig(t *testing.T) {
	if !projectFileExists(t, "cmd/codectx/main.go") {
		t.Fatal("cmd/main.go does not exist (required for startup sequence)")
	}
	content := readProjectFile(t, "cmd/codectx/main.go")

	if !strings.Contains(content, "internal/config") {
		t.Error("cmd/main.go should import internal/config for LoadFalkorDBConfig()")
	}
}

// TestCmdMainGoImportsCodeDB verifies that cmd/main.go imports the codedb package.
// Expected result: cmd/main.go contains codedb import.
func TestCmdMainGoImportsCodeDB(t *testing.T) {
	if !projectFileExists(t, "cmd/codectx/main.go") {
		t.Fatal("cmd/main.go does not exist (required for startup sequence)")
	}
	content := readProjectFile(t, "cmd/codectx/main.go")

	if !strings.Contains(content, "internal/clients/code_db") {
		t.Error("cmd/main.go should import internal/clients/code_db for NewCodeDB()")
	}
}

// TestCmdMainGoHasSignalHandling verifies that cmd/main.go includes signal handling
// for graceful shutdown (SIGINT/SIGTERM).
// Expected result: cmd/main.go contains signal handling code.
func TestCmdMainGoHasSignalHandling(t *testing.T) {
	if !projectFileExists(t, "cmd/codectx/main.go") {
		t.Fatal("cmd/main.go does not exist (required for startup sequence)")
	}
	content := readProjectFile(t, "cmd/codectx/main.go")

	if !strings.Contains(content, "os/signal") && !strings.Contains(content, "signal.Notify") {
		t.Error("cmd/main.go should include signal handling for graceful shutdown")
	}
}

// --- Task 17: Pipeline wiring ---

// TestCmdMainGoImportsApp verifies that cmd/main.go imports the internal/app package.
// After refactoring, pipeline wiring (indexer, analysis, extractors) moved to internal/app.
// Expected result: cmd/main.go contains internal/app import.
func TestCmdMainGoImportsIndexer(t *testing.T) {
	if !projectFileExists(t, "cmd/codectx/main.go") {
		t.Fatal("cmd/main.go does not exist (required for pipeline wiring)")
	}
	content := readProjectFile(t, "cmd/codectx/main.go")

	if !strings.Contains(content, "internal/app") {
		t.Error("cmd/main.go should import internal/app for pipeline wiring (composition root)")
	}
}

// TestCmdMainGoImportsAppPackage verifies that cmd/main.go imports internal/app
// which wires analysis and indexer (composition root pattern).
// Expected result: cmd/main.go contains internal/app import.
func TestCmdMainGoImportsAnalysis(t *testing.T) {
	if !projectFileExists(t, "cmd/codectx/main.go") {
		t.Fatal("cmd/main.go does not exist (required for pipeline wiring)")
	}
	content := readProjectFile(t, "cmd/codectx/main.go")

	if !strings.Contains(content, "internal/app") {
		t.Error("cmd/main.go should import internal/app — analysis wiring is done in the app package")
	}
}

// TestCmdMainGoNoSearchImport verifies that cmd/main.go no longer imports
// the search package (embeddings removed).
// Expected result: cmd/main.go does NOT contain internal/search import.
func TestCmdMainGoNoSearchImport(t *testing.T) {
	if !projectFileExists(t, "cmd/codectx/main.go") {
		t.Fatal("cmd/main.go does not exist")
	}
	content := readProjectFile(t, "cmd/codectx/main.go")

	if strings.Contains(content, "internal/search") {
		t.Error("cmd/main.go should not import internal/search — embeddings removed")
	}
}

// --- Task 18: Build pipeline verification ---

// TestBuildPipelineComponents verifies that Taskfile.yml contains
// all required pipeline components.
// Expected result: All pipeline components are present.
func TestBuildPipelineComponents(t *testing.T) {
	taskfile := readProjectFile(t, "Taskfile.yml")

	components := []struct {
		name    string
		pattern string
	}{
		{"generate task", "generate:"},
		{"gormql command", "gormql generate"},
		{"build task", "build:"},
		{"go build command", "go build"},
	}

	for _, c := range components {
		t.Run(c.name, func(t *testing.T) {
			if !strings.Contains(taskfile, c.pattern) {
				t.Errorf("Taskfile.yml missing pipeline component %q (pattern: %q)", c.name, c.pattern)
			}
		})
	}
}

// TestAllNewPackagesExist verifies that all new packages from the feature
// have been created with at least one .go file.
// Expected result: All package directories and Go files exist.
func TestAllNewPackagesExist(t *testing.T) {
	packages := []struct {
		name string
		path string
	}{
		{"config/falkordb", "internal/config/falkordb.go"},
		{"indexer", "internal/indexer/indexer.go"},
		{"indexer/gitignore", "internal/indexer/gitignore.go"},
		{"indexer/detect", "internal/indexer/detect.go"},
		{"analysis/types", "internal/analysis/types.go"},
		{"analysis/extractor", "internal/analysis/extractor.go"},
		{"analysis/registry", "internal/analysis/registry.go"},
		{"analysis/analyzer", "internal/analysis/analyzer.go"},
		{"analysis/golang", "internal/analysis/golang/extractor.go"},
		{"analysis/typescript", "internal/analysis/typescript/extractor.go"},
		{"analysis/tsx", "internal/analysis/typescript/tsx_extractor.go"},
		// search/embedder removed — embeddings deleted
	}

	for _, pkg := range packages {
		t.Run(pkg.name, func(t *testing.T) {
			if !projectFileExists(t, pkg.path) {
				t.Errorf("package file missing: %s", pkg.path)
			}
		})
	}
}
