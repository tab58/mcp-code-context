package indexer

import (
	"context"
	"testing"
)

// --- Task 3: Migrate indexer to ForRepo ---

// TestIndexRepository_UsesForRepo verifies that IndexRepository calls
// db.ForRepo(ctx, repoName) to get a repo-scoped client, rather than
// db.Client(). The recording driver should see calls made after ForRepo
// lazy-initializes the graph.
// Expected result: IndexRepository succeeds with a recording driver,
// and the driver records calls (proving the client was obtained via ForRepo).
func TestIndexRepository_UsesForRepo(t *testing.T) {
	idx, rec := newTestIndexerWithRecorder(t)
	repo := createPersistenceTestRepo(t)

	result, err := idx.IndexRepository(context.Background(), repo)
	if err != nil {
		t.Fatalf("IndexRepository returned error: %v", err)
	}

	// Should have indexed files
	if result.FilesIndexed == 0 {
		t.Error("IndexRepository indexed 0 files, expected > 0")
	}

	// Should have made driver calls (upsert + query + create + edges)
	totalCalls := len(rec.executeCalls) + len(rec.executeWriteCalls)
	if totalCalls == 0 {
		t.Error("IndexRepository made no driver calls — ForRepo-based client not exercised")
	}
}

// TestUpsertRepository_AcceptsClient verifies that upsertRepository
// accepts a *client.Client parameter (from ForRepo) instead of
// accessing idx.db.Client() internally.
// Expected result: Method signature accepts *client.Client.
// This is validated by the source inspection test in isolation_verify_test.go.
// Here we test behavior: upsert should work when the client is obtained
// through ForRepo (which is what IndexRepository does).
func TestUpsertRepository_AcceptsClient(t *testing.T) {
	idx, _ := newTestIndexerWithRecorder(t)

	_, err := idx.upsertRepository(context.Background(), testClient(t, idx, "test-repo"), "test-repo", "/tmp/test-repo")
	if err != nil {
		t.Fatalf("upsertRepository returned error: %v", err)
	}
}
