package analysis

import (
	"context"
	"fmt"
	"path/filepath"

	"github.com/tab58/go-ormql/pkg/client"
)

// GraphQL mutations for ExternalReference graph writes via Client().Execute().
const (
	gqlMergeExternalReferences = `mutation($input: [ExternalReferenceMergeInput!]!) {
  mergeExternalReferences(input: $input) { externalReferences { id name } }
}`

	gqlConnectFileExternalImports = `mutation($input: [ConnectFileExternalImportsInput!]!) {
  connectFileExternalImports(input: $input) { relationshipsCreated }
}`

	gqlConnectFunctionExternalCalls = `mutation($input: [ConnectFunctionExternalCallsInput!]!) {
  connectFunctionExternalCalls(input: $input) { relationshipsCreated }
}`

	gqlConnectExternalReferenceRepository = `mutation($input: [ConnectExternalReferenceRepositoryInput!]!) {
  connectExternalReferenceRepository(input: $input) { relationshipsCreated }
}`
)

// writeExternalReferences creates ExternalReference nodes and edges for all
// external references found in the analyses. Deduplicates by name+importPath.
func (a *Analyzer) writeExternalReferences(ctx context.Context, c *client.Client, repoID string, repoPath string, analyses []FileAnalysis) error {
	// Deduplicate external refs by name+importPath
	type extRefKey struct{ name, importPath string }
	seen := make(map[extRefKey]bool)
	var mergeInputs []map[string]any
	var belongsToEdges []map[string]any
	var importEdges []map[string]any
	var callEdges []map[string]any

	for _, fa := range analyses {
		// File nodes use relative paths; compute for IMPORTS edge matching.
		fileRelPath, _ := filepath.Rel(repoPath, fa.FilePath)

		for _, ref := range fa.References {
			if !ref.IsExternal {
				continue
			}

			key := extRefKey{ref.ToName, ref.ExternalImportPath}
			if !seen[key] {
				seen[key] = true
				mergeInputs = append(mergeInputs, map[string]any{
					"match":    map[string]any{"name": ref.ToName, "importPath": ref.ExternalImportPath},
					"onCreate": map[string]any{},
					"onMatch":  map[string]any{},
				})
				belongsToEdges = append(belongsToEdges, map[string]any{
					"from": map[string]any{"name": ref.ToName, "importPath": ref.ExternalImportPath},
					"to":   map[string]any{"name": repoID},
				})
			}

			switch ref.Kind {
			case "imports":
				importEdges = append(importEdges, map[string]any{
					"from": map[string]any{"path": fileRelPath},
					"to":   map[string]any{"name": ref.ToName, "importPath": ref.ExternalImportPath},
				})
			case "calls":
				callEdges = append(callEdges, map[string]any{
					"from": map[string]any{"name": ref.FromSymbol, "path": ref.FilePath},
					"to":   map[string]any{"name": ref.ToName, "importPath": ref.ExternalImportPath},
				})
			}
		}
	}

	if err := batchMutate(ctx, c, mergeInputs, gqlMergeExternalReferences, mergeBatchSize); err != nil {
		return fmt.Errorf("mergeExternalReferences: %w", err)
	}
	if err := batchMutate(ctx, c, belongsToEdges, gqlConnectExternalReferenceRepository, edgeBatchSize); err != nil {
		return fmt.Errorf("connectExternalReferenceRepository: %w", err)
	}
	if err := batchMutate(ctx, c, importEdges, gqlConnectFileExternalImports, edgeBatchSize); err != nil {
		return fmt.Errorf("connectFileExternalImports: %w", err)
	}
	if err := batchMutate(ctx, c, callEdges, gqlConnectFunctionExternalCalls, edgeBatchSize); err != nil {
		return fmt.Errorf("connectFunctionExternalCalls: %w", err)
	}

	return nil
}
