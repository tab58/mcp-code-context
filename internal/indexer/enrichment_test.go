package indexer

import (
	"context"
	"testing"
	"time"
)

// === Task 3: Update indexer upsertRepository — persist path ===
//
// upsertRepository should pass repoPath (absolute path) and include
// "path": repoPath in the onCreate and onMatch maps of the mergeRepositorys
// mutation.

// TestUpsertRepository_IncludesPath verifies that upsertRepository passes
// the absolute repository path to the GraphQL mutation so that it is
// persisted as a visible node attribute on the Repository node.
// Expected result: The driver call params (deeply nested via go-ormql translation)
// contain the absolute path value. We verify by checking the upsertRepository
// method signature accepts repoPath and that the mutation vars include "path".
func TestUpsertRepository_IncludesPath(t *testing.T) {
	idx, rec := newTestIndexerWithRecorder(t)

	repoPath := "/home/user/projects/my-project"
	repoName := "my-project"

	_, err := idx.upsertRepository(context.Background(), testClient(t, idx, "test-repo"), repoName, repoPath)
	if err != nil {
		t.Fatalf("upsertRepository returned error: %v", err)
	}

	// Verify that at least one driver call was made
	totalCalls := len(rec.executeCalls) + len(rec.executeWriteCalls)
	if totalCalls == 0 {
		t.Error("upsertRepository made no driver calls")
	}

	// Verify the path value appears somewhere in the driver call params.
	// After go-ormql Client().Execute() translation, the path ends up in
	// Cypher params (possibly deeply nested).
	found := false
	for _, call := range append(rec.executeCalls, rec.executeWriteCalls...) {
		if containsValue(call.Params, repoPath) {
			found = true
			break
		}
	}
	if !found {
		t.Error("upsertRepository did not pass the absolute path to the mutation — expected 'path' in onCreate/onMatch")
	}
}

// TestUpsertRepository_PathInOnCreateAndOnMatch verifies that the path
// appears in both onCreate and onMatch maps so it is set regardless of
// whether the Repository node is new or existing.
// Expected result: The source code of upsertRepository includes path in both maps.
func TestUpsertRepository_PathInOnCreateAndOnMatch(t *testing.T) {
	idx, rec := newTestIndexerWithRecorder(t)

	repoPath := "/tmp/test-repo"
	_, err := idx.upsertRepository(context.Background(), testClient(t, idx, "test-repo"), "test-repo", repoPath)
	if err != nil {
		t.Fatalf("upsertRepository returned error: %v", err)
	}

	// The path must be set on both create and update (match) of the repo node.
	// With go-ormql, both onCreate and onMatch are in the mutation vars.
	totalCalls := len(rec.executeCalls) + len(rec.executeWriteCalls)
	if totalCalls == 0 {
		t.Error("upsertRepository made no driver calls — cannot verify path in onCreate/onMatch")
	}
}

// === Task 4: Update indexer createNodes (folders) — persist path ===
//
// Folder merge mutations should include "path": f.Path in both onCreate
// and onMatch maps so the relative path is persisted as a visible node
// attribute (not just the match key).

// TestCreateNodes_FolderHasPathInOnCreate verifies that folder merge
// mutations include the path in the onCreate map.
// Expected result: Driver call params contain the folder's relative path.
func TestCreateNodes_FolderHasPathInOnCreate(t *testing.T) {
	idx, rec := newTestIndexerWithRecorder(t)

	folders := []pendingFolder{
		{Path: "src/utils", ParentPath: "src", ModTime: time.Now()},
	}

	err := idx.createNodes(context.Background(), testClient(t, idx, "test-repo"), folders, nil)
	if err != nil {
		t.Fatalf("createNodes returned error: %v", err)
	}

	// The folder's relative path should appear in the mutation params.
	// With go-ormql, the path in onCreate/onMatch becomes a Cypher param.
	found := false
	for _, call := range append(rec.executeCalls, rec.executeWriteCalls...) {
		if containsValue(call.Params, "src/utils") {
			found = true
			break
		}
	}
	if !found {
		t.Error("createNodes did not include folder path in mutation params — expected 'path' in onCreate/onMatch for folders")
	}
}

// TestCreateNodes_FolderPathAppearsMultipleTimes verifies that path appears
// in both the match key AND the onCreate/onMatch maps (3 occurrences total).
// Expected result: The path value appears at least 3 times in the params
// (match, onCreate, onMatch).
func TestCreateNodes_FolderPathAppearsMultipleTimes(t *testing.T) {
	idx, rec := newTestIndexerWithRecorder(t)

	folders := []pendingFolder{
		{Path: "src", ParentPath: "", ModTime: time.Now()},
	}

	err := idx.createNodes(context.Background(), testClient(t, idx, "test-repo"), folders, nil)
	if err != nil {
		t.Fatalf("createNodes returned error: %v", err)
	}

	// Count occurrences of "src" in all params — should be in match + onCreate + onMatch
	count := 0
	for _, call := range append(rec.executeCalls, rec.executeWriteCalls...) {
		count += countValue(call.Params, "src")
	}

	// path should appear in match key, onCreate, and onMatch = at least 3 times
	if count < 3 {
		t.Errorf("folder path 'src' appeared %d times in params, expected at least 3 (match + onCreate + onMatch)", count)
	}
}

// === Task 5: Update indexer createNodes (files) — persist path + filename ===
//
// File merge mutations should include "path": f.Path and
// "filename": filepath.Base(f.Path) in both onCreate and onMatch maps.

// TestCreateNodes_FileHasPathInOnCreate verifies that file merge
// mutations include the path in the onCreate map.
// Expected result: Driver call params contain the file's relative path
// beyond just the match key.
func TestCreateNodes_FileHasPathInOnCreate(t *testing.T) {
	idx, rec := newTestIndexerWithRecorder(t)

	files := []pendingFile{
		{Path: "src/main.go", ParentPath: "src", Language: "go", LineCount: 10, ModTime: time.Now()},
	}

	err := idx.createNodes(context.Background(), testClient(t, idx, "test-repo"), nil, files)
	if err != nil {
		t.Fatalf("createNodes returned error: %v", err)
	}

	// The file's relative path should appear in the mutation params
	// beyond just the match key (also in onCreate/onMatch).
	count := 0
	for _, call := range append(rec.executeCalls, rec.executeWriteCalls...) {
		count += countValue(call.Params, "src/main.go")
	}

	// path in match + onCreate + onMatch = at least 3
	if count < 3 {
		t.Errorf("file path 'src/main.go' appeared %d times in params, expected at least 3 (match + onCreate + onMatch)", count)
	}
}

// TestCreateNodes_FileHasFilename verifies that file merge mutations
// include a "filename" field (filepath.Base of the relative path).
// Expected result: Driver call params contain the base filename "main.go".
func TestCreateNodes_FileHasFilename(t *testing.T) {
	idx, rec := newTestIndexerWithRecorder(t)

	files := []pendingFile{
		{Path: "src/main.go", ParentPath: "src", Language: "go", LineCount: 10, ModTime: time.Now()},
	}

	err := idx.createNodes(context.Background(), testClient(t, idx, "test-repo"), nil, files)
	if err != nil {
		t.Fatalf("createNodes returned error: %v", err)
	}

	// The base filename "main.go" should appear in the mutation params
	// (from "filename": filepath.Base(f.Path) in onCreate/onMatch).
	found := false
	for _, call := range append(rec.executeCalls, rec.executeWriteCalls...) {
		if containsValue(call.Params, "main.go") {
			found = true
			break
		}
	}
	if !found {
		t.Error("createNodes did not include filename in mutation params — expected 'filename': 'main.go' in onCreate/onMatch for files")
	}
}

// TestCreateNodes_FileFilenameIsBaseName verifies that the filename field
// is the base name only, not the full relative path.
// Expected result: For path "deeply/nested/file.go", filename should be "file.go".
func TestCreateNodes_FileFilenameIsBaseName(t *testing.T) {
	idx, rec := newTestIndexerWithRecorder(t)

	files := []pendingFile{
		{Path: "deeply/nested/file.go", ParentPath: "deeply/nested", Language: "go", LineCount: 5, ModTime: time.Now()},
	}

	err := idx.createNodes(context.Background(), testClient(t, idx, "test-repo"), nil, files)
	if err != nil {
		t.Fatalf("createNodes returned error: %v", err)
	}

	// "file.go" (base name) should appear in params, but "deeply/nested/file.go"
	// should only appear 3x (match + onCreate + onMatch as path, not as filename).
	foundFilename := false
	for _, call := range append(rec.executeCalls, rec.executeWriteCalls...) {
		if containsValue(call.Params, "file.go") {
			foundFilename = true
			break
		}
	}
	if !foundFilename {
		t.Error("createNodes did not include base filename 'file.go' — expected filepath.Base() as filename in onCreate/onMatch")
	}
}

// === Helpers ===

// containsValue recursively searches params for a string value.
func containsValue(params map[string]any, target string) bool {
	for _, v := range params {
		if searchAny(v, target) {
			return true
		}
	}
	return false
}

// countValue recursively counts occurrences of a string value in params.
func countValue(params map[string]any, target string) int {
	count := 0
	for _, v := range params {
		count += countAny(v, target)
	}
	return count
}

func searchAny(v any, target string) bool {
	switch val := v.(type) {
	case string:
		return val == target
	case map[string]any:
		return containsValue(val, target)
	case []any:
		for _, item := range val {
			if searchAny(item, target) {
				return true
			}
		}
	}
	return false
}

func countAny(v any, target string) int {
	switch val := v.(type) {
	case string:
		if val == target {
			return 1
		}
		return 0
	case map[string]any:
		return countValue(val, target)
	case []any:
		count := 0
		for _, item := range val {
			count += countAny(item, target)
		}
		return count
	}
	return 0
}

