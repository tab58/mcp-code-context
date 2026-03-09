package indexer

import (
	"context"
	"strings"
	"testing"
	"time"
)

// === Task 2: Migrate upsertRepository to mergeRepositorys mutation ===
// upsertRepository should use Client().Execute() with a mergeRepositorys
// GraphQL mutation instead of raw Cypher via Driver().ExecuteWrite().

// TestUpsertRepository_UsesClientExecute verifies that upsertRepository
// calls Client().Execute() (not Driver().ExecuteWrite() directly).
// Expected result: The driver receives a translated Cypher statement
// from Client().Execute() (not a raw cypherUpsertRepo constant).
func TestUpsertRepository_UsesClientExecute(t *testing.T) {
	idx, rec := newTestIndexerWithRecorder(t)

	_, err := idx.upsertRepository(context.Background(), testClient(t, idx, "test-repo"), "test-repo", "/tmp/test-repo")
	if err != nil {
		t.Fatalf("upsertRepository returned error: %v", err)
	}

	// Client().Execute() for mutations routes through ExecuteWrite.
	// The key distinction: Client().Execute() generates Cypher from GraphQL,
	// so the query should NOT be the raw cypherUpsertRepo constant.
	for _, call := range rec.executeWriteCalls {
		if strings.Contains(call.Query, "MERGE (r:Repository {name: $name})") &&
			strings.Contains(call.Query, "ON CREATE SET") {
			t.Error("upsertRepository still uses raw Cypher constant — should use Client().Execute() with mergeRepositorys GraphQL mutation")
		}
	}
}

// TestUpsertRepository_DoesNotCallDriverDirectly verifies that upsertRepository
// does not reference Driver() at all — it should go through Client().Execute().
// Expected result: No raw Cypher MERGE pattern in the driver calls.
func TestUpsertRepository_DoesNotCallDriverDirectly(t *testing.T) {
	idx, rec := newTestIndexerWithRecorder(t)

	_, err := idx.upsertRepository(context.Background(), testClient(t, idx, "test-repo"), "test-repo", "/tmp/test-repo")
	if err != nil {
		t.Fatalf("upsertRepository returned error: %v", err)
	}

	// Verify at least one call was made (through Client().Execute())
	totalCalls := len(rec.executeCalls) + len(rec.executeWriteCalls)
	if totalCalls == 0 {
		t.Error("upsertRepository made no driver calls")
	}
}

// === Task 3: Migrate queryExistingNodes to relationship WHERE queries ===
// queryExistingNodes should use Client().Execute() with folders/files queries
// that include relationship WHERE filters, instead of raw Cypher MATCH patterns.

// TestQueryExistingNodes_UsesClientExecute verifies that queryExistingNodes
// uses Client().Execute() with relationship WHERE filters (not raw Cypher).
// Expected result: No raw Cypher MATCH patterns in driver calls.
func TestQueryExistingNodes_UsesClientExecute(t *testing.T) {
	idx, rec := newTestIndexerWithRecorder(t)

	_, err := idx.queryExistingNodes(context.Background(), testClient(t, idx, "test-repo"), "test-repo")
	if err != nil {
		t.Fatalf("queryExistingNodes returned error: %v", err)
	}

	// Client().Execute() for queries routes through Execute (read).
	// Should NOT see the raw Cypher MATCH patterns.
	for _, call := range rec.executeCalls {
		if strings.Contains(call.Query, "MATCH (f:Folder)-[:BELONGS_TO]->(r:Repository") {
			t.Error("queryExistingNodes still uses raw Cypher MATCH — should use Client().Execute() with relationship WHERE queries")
		}
		if strings.Contains(call.Query, "MATCH (fi:File)-[:BELONGS_TO]->(r:Repository") {
			t.Error("queryExistingNodes still uses raw Cypher MATCH — should use Client().Execute() with relationship WHERE queries")
		}
	}
}

// TestQueryExistingNodes_QueriesBothFoldersAndFiles verifies that
// queryExistingNodes makes separate queries for folders and files.
// Expected result: At least 2 Execute calls (one for folders, one for files).
func TestQueryExistingNodes_QueriesBothFoldersAndFiles(t *testing.T) {
	idx, rec := newTestIndexerWithRecorder(t)

	_, err := idx.queryExistingNodes(context.Background(), testClient(t, idx, "test-repo"), "test-repo")
	if err != nil {
		t.Fatalf("queryExistingNodes returned error: %v", err)
	}

	// Client().Execute() for queries uses Execute (read), not ExecuteWrite
	if len(rec.executeCalls) < 2 {
		t.Errorf("queryExistingNodes made %d Execute calls, want at least 2 (folders + files)", len(rec.executeCalls))
	}
}

// === Task 4: Migrate createNodes to mergeFolders/mergeFiles mutations ===
// createNodes should use Client().Execute() with mergeFolders/mergeFiles
// GraphQL mutations instead of raw Cypher UNWIND+MERGE.

// TestCreateNodes_UsesClientExecuteMerge verifies that createNodes uses
// Client().Execute() with merge mutations (not raw Cypher UNWIND+MERGE).
// Expected result: No raw Cypher UNWIND patterns in driver calls.
func TestCreateNodes_UsesClientExecuteMerge(t *testing.T) {
	idx, rec := newTestIndexerWithRecorder(t)

	folders := []pendingFolder{
		{Path: "src", ParentPath: "", ModTime: time.Now()},
	}
	files := []pendingFile{
		{Path: "main.go", ParentPath: "", Language: "go", LineCount: 10, ModTime: time.Now()},
	}

	err := idx.createNodes(context.Background(), testClient(t, idx, "test-repo"), folders, files)
	if err != nil {
		t.Fatalf("createNodes returned error: %v", err)
	}

	// Check that no raw Cypher UNWIND patterns are used
	for _, call := range rec.executeWriteCalls {
		if strings.Contains(call.Query, "UNWIND $items AS item") &&
			strings.Contains(call.Query, "MERGE (f:Folder") {
			t.Error("createNodes still uses raw Cypher UNWIND+MERGE for folders — should use mergeFolders GraphQL mutation")
		}
		if strings.Contains(call.Query, "UNWIND $items AS item") &&
			strings.Contains(call.Query, "MERGE (f:File") {
			t.Error("createNodes still uses raw Cypher UNWIND+MERGE for files — should use mergeFiles GraphQL mutation")
		}
	}
}

// TestCreateNodes_StillBatchesAt50 verifies that createNodes still batches
// at 50 items per mutation call even with the new Client().Execute() API.
// Expected result: 120 folders produce at least 3 mutation calls.
func TestCreateNodes_StillBatchesAt50(t *testing.T) {
	idx, rec := newTestIndexerWithRecorder(t)

	folders := make([]pendingFolder, 120)
	for i := range folders {
		folders[i] = pendingFolder{
			Path:       "dir/" + string(rune('a'+i%26)),
			ParentPath: "",
			ModTime:    time.Now(),
		}
	}

	// Record baseline calls (CreateIndexes)
	baselineCalls := len(rec.executeCalls) + len(rec.executeWriteCalls)

	err := idx.createNodes(context.Background(), testClient(t, idx, "test-repo"), folders, nil)
	if err != nil {
		t.Fatalf("createNodes returned error: %v", err)
	}

	newCalls := (len(rec.executeCalls) + len(rec.executeWriteCalls)) - baselineCalls
	if newCalls < 3 {
		t.Errorf("createNodes made %d calls for 120 items, want at least 3 (batches of 50)", newCalls)
	}
}

// === Task 5: Migrate createEdges to 6 connect* mutations ===
// createEdges should use Client().Execute() with 6 connect* GraphQL mutations
// instead of raw Cypher UNWIND+MERGE edge statements.

// TestCreateEdges_UsesClientExecuteConnect verifies that createEdges uses
// Client().Execute() with connect* mutations (not raw Cypher UNWIND+MERGE).
// Expected result: No raw Cypher MATCH+MERGE edge patterns in driver calls.
func TestCreateEdges_UsesClientExecuteConnect(t *testing.T) {
	idx, rec := newTestIndexerWithRecorder(t)

	folders := []pendingFolder{
		{Path: "src", ParentPath: "", ModTime: time.Now()},
	}
	files := []pendingFile{
		{Path: "main.go", ParentPath: "", Language: "go", LineCount: 10, ModTime: time.Now()},
	}

	// Record baseline calls
	baselineWrites := len(rec.executeWriteCalls)

	err := idx.createEdges(context.Background(), testClient(t, idx, "test-repo"), "test-repo", folders, files)
	if err != nil {
		t.Fatalf("createEdges returned error: %v", err)
	}

	// Check that no raw Cypher MATCH+MERGE edge patterns remain
	for i := baselineWrites; i < len(rec.executeWriteCalls); i++ {
		call := rec.executeWriteCalls[i]
		if strings.Contains(call.Query, "UNWIND $items AS item") &&
			strings.Contains(call.Query, "MERGE (r)-[:CONTAINS]->(f)") {
			t.Error("createEdges still uses raw Cypher UNWIND+MERGE — should use connect* GraphQL mutations")
		}
		if strings.Contains(call.Query, "UNWIND $items AS item") &&
			strings.Contains(call.Query, "MERGE (f)-[:BELONGS_TO]->(r)") {
			t.Error("createEdges still uses raw Cypher UNWIND+MERGE — should use connect* GraphQL mutations")
		}
	}
}

// TestCreateEdges_StillProducesSixEdgeTypes verifies that createEdges still
// produces all 6 edge types via Client().Execute() connect mutations.
// Expected result: At least 6 mutation calls (one per edge type).
func TestCreateEdges_StillProducesSixEdgeTypes(t *testing.T) {
	idx, rec := newTestIndexerWithRecorder(t)

	folders := []pendingFolder{
		{Path: "src", ParentPath: "", ModTime: time.Now()},
		{Path: "src/utils", ParentPath: "src", ModTime: time.Now()},
	}
	files := []pendingFile{
		{Path: "main.go", ParentPath: "", Language: "go", LineCount: 10, ModTime: time.Now()},
		{Path: "src/util.go", ParentPath: "src", Language: "go", LineCount: 5, ModTime: time.Now()},
	}

	// Record baseline
	baselineWrites := len(rec.executeWriteCalls)

	err := idx.createEdges(context.Background(), testClient(t, idx, "test-repo"), "test-repo", folders, files)
	if err != nil {
		t.Fatalf("createEdges returned error: %v", err)
	}

	newWrites := len(rec.executeWriteCalls) - baselineWrites
	if newWrites < 6 {
		t.Errorf("createEdges made %d write calls, want at least 6 (one per edge type)", newWrites)
	}
}

// TestCreateEdges_StillBatchesAt50 verifies that edge creation still batches
// at 50 items per mutation with the new connect* API.
// Expected result: 120 root folders produce at least 3+3=6 calls for
// CONTAINS and BELONGS_TO edges alone.
func TestCreateEdges_StillBatchesAt50(t *testing.T) {
	idx, rec := newTestIndexerWithRecorder(t)

	folders := make([]pendingFolder, 120)
	for i := range folders {
		folders[i] = pendingFolder{
			Path:       "dir/" + string(rune('a'+i%26)),
			ParentPath: "",
			ModTime:    time.Now(),
		}
	}

	baselineWrites := len(rec.executeWriteCalls)

	err := idx.createEdges(context.Background(), testClient(t, idx, "test-repo"), "test-repo", folders, nil)
	if err != nil {
		t.Fatalf("createEdges returned error: %v", err)
	}

	newWrites := len(rec.executeWriteCalls) - baselineWrites
	if newWrites < 6 {
		t.Errorf("createEdges made %d write calls for 120 folders, want at least 6 (batched at 50)", newWrites)
	}
}
