package analysis

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/tab58/go-ormql/pkg/client"
)

// mergeBatchSize is the number of items per merge mutation call.
// Kept small because merge mutations include full source code in parameters,
// which FalkorDB serializes into the CYPHER header. Large batches with source
// code cause FalkorDB to OOM in memory-constrained Docker environments.
const mergeBatchSize = 2

// edgeBatchSize is the number of items per connect mutation call.
// Even with property indexes, UNWIND + double MATCH + MERGE accumulates
// intermediate result sets in FalkorDB. Keep batches small to avoid OOM
// in memory-constrained Docker environments with large repositories.
const edgeBatchSize = 5

// maxSourceLen caps source code stored per symbol (6000 runes).
const maxSourceLen = 6000

// sourceCodeBatchSize limits source code items per UNWIND batch.
// Source code strings can be large, so batches are kept smaller than edge batches
// to avoid oversized Redis commands.
const sourceCodeBatchSize = 5

// gqlSetSourceBatch is a Cypher query for batched source code writes via UNWIND.
// Uses MATCH+SET (no MERGE) so no graph scan for existence checking is needed.
const gqlSetSourceBatch = `MATCH (n {name: item.name, path: item.path}) SET n.source = item.source`

// gqlSetComplexityBatch is a Cypher query for batched complexity writes via UNWIND.
const gqlSetComplexityBatch = `MATCH (f:Function {name: item.name, path: item.path}) SET f.cyclomaticComplexity = item.complexity`

// edgeSpec describes how to create edges between two node types via raw Cypher.
// Uses individual MATCH+CREATE statements instead of UNWIND+MERGE to avoid
// FalkorDB memory spikes from intermediate result accumulation.
type edgeSpec struct {
	FromLabel string            // e.g. "File"
	FromWhere map[string]string // field name → param key, e.g. {"path": "from_path"}
	ToLabel   string            // e.g. "Function"
	ToWhere   map[string]string // field name → param key
	RelType   string            // e.g. "DEFINES"
	EdgeProps map[string]string // optional edge properties, field → param key
}

// createEdgesRaw creates edges via batched UNWIND raw Cypher for efficiency.
// Each item is a map with "from" and "to" sub-maps containing match fields,
// and optionally an "edge" sub-map for relationship properties.
// Uses Client.ExecuteRawBatch to reduce FalkorDB round-trips.
func createEdgesRaw(ctx context.Context, c *client.Client, items []map[string]any, spec edgeSpec) error {
	if len(items) == 0 {
		return nil
	}
	query := buildEdgeCypherBatch(spec)
	if err := c.ExecuteRawBatch(ctx, query, items, edgeBatchSize); err != nil {
		return fmt.Errorf("create edge %s: %w", spec.RelType, err)
	}
	return nil
}

// buildEdgeCypherBatch builds a Cypher query for use with ExecuteRawBatch (UNWIND).
// Uses item.from.field and item.to.field references instead of flat $param names.
func buildEdgeCypherBatch(spec edgeSpec) string {
	var sb strings.Builder
	sb.WriteString("MATCH (a:")
	sb.WriteString(spec.FromLabel)
	sb.WriteString(" {")
	writeBatchWhereProps(&sb, "from", spec.FromWhere)
	sb.WriteString("}) MATCH (b:")
	sb.WriteString(spec.ToLabel)
	sb.WriteString(" {")
	writeBatchWhereProps(&sb, "to", spec.ToWhere)
	sb.WriteString("}) MERGE (a)-[r:")
	sb.WriteString(spec.RelType)
	sb.WriteString("]->(b)")
	if len(spec.EdgeProps) > 0 {
		sb.WriteString(" SET ")
		first := true
		for field := range spec.EdgeProps {
			if !first {
				sb.WriteString(", ")
			}
			sb.WriteString("r.")
			sb.WriteString(field)
			sb.WriteString(" = item.edge.")
			sb.WriteString(field)
			first = false
		}
	}
	return sb.String()
}

// writeBatchWhereProps writes Cypher property match expressions for UNWIND batch queries.
// Uses item.prefix.field syntax (e.g., "name: item.from.name, path: item.from.path").
func writeBatchWhereProps(sb *strings.Builder, prefix string, where map[string]string) {
	first := true
	for field := range where {
		if !first {
			sb.WriteString(", ")
		}
		sb.WriteString(field)
		sb.WriteString(": item.")
		sb.WriteString(prefix)
		sb.WriteString(".")
		sb.WriteString(field)
		first = false
	}
}

// batchMutate executes a GraphQL mutation in batches of the given size
// via Client().Execute(). Converts []map[string]any to []any for
// FalkorDB driver compatibility.
func batchMutate(ctx context.Context, c *client.Client, items []map[string]any, query string, size int) error {
	// Convert to []any so FalkorDB's ToString can handle the slice type.
	anyItems := make([]any, len(items))
	for i, item := range items {
		anyItems[i] = item
	}

	for i := 0; i < len(anyItems); i += size {
		end := i + size
		if end > len(anyItems) {
			end = len(anyItems)
		}
		if _, err := c.Execute(ctx, query, map[string]any{"input": anyItems[i:end]}); err != nil {
			return err
		}
	}
	return nil
}

// computeEndingLine returns the ending line number given a starting line and
// line count. When lineCount is 0 or 1, endingLine equals startingLine.
func computeEndingLine(startingLine, lineCount int) int {
	if lineCount > 1 {
		return startingLine + lineCount - 1
	}
	return startingLine
}

// buildFuncFields creates a fresh map of metadata fields for a symbol.
// Excludes source code to keep merge queries small — source is written
// in a separate pass via writeSourceCode after all structural merges complete.
func buildFuncFields(sym Symbol) map[string]any {
	fields := map[string]any{
		"language":     sym.Language,
		"visibility":   sym.Visibility,
		"startingLine": sym.LineNumber,
		"endingLine":   computeEndingLine(sym.LineNumber, sym.LineCount),
	}
	if sym.Signature != "" {
		fields["signature"] = sym.Signature
	}
	return fields
}

// withKind returns a copy of fields with "kind" added.
func withKind(fields map[string]any, kind string) map[string]any {
	out := make(map[string]any, len(fields)+1)
	for k, v := range fields {
		out[k] = v
	}
	out["kind"] = kind
	return out
}

// isTestFile returns true if the file path looks like a test file that should
// be excluded from the code knowledge graph. Matches Go test files (*_test.go),
// JavaScript/TypeScript test files (*.test.*, *.spec.*), Python test files
// (test_*.py, *_test.py), Ruby test/spec files (*_test.rb, *_spec.rb),
// and files in test/tests/spec directories.
func isTestFile(path string) bool {
	base := filepath.Base(path)
	if strings.HasSuffix(base, "_test.go") {
		return true
	}
	// JS/TS test patterns: foo.test.ts, foo.spec.tsx
	ext := filepath.Ext(base)
	nameWithoutExt := strings.TrimSuffix(base, ext)
	if strings.HasSuffix(nameWithoutExt, ".test") || strings.HasSuffix(nameWithoutExt, ".spec") {
		return true
	}
	// Python test patterns: test_models.py, models_test.py
	if ext == ".py" && (strings.HasPrefix(base, "test_") || strings.HasSuffix(nameWithoutExt, "_test")) {
		return true
	}
	// Ruby test/spec patterns: user_test.rb, user_spec.rb
	if ext == ".rb" && (strings.HasSuffix(nameWithoutExt, "_test") || strings.HasSuffix(nameWithoutExt, "_spec")) {
		return true
	}
	// Test directory patterns: test/, tests/, spec/
	normalized := filepath.ToSlash(path)
	for _, seg := range strings.Split(normalized, "/") {
		if seg == "test" || seg == "tests" || seg == "spec" {
			return true
		}
	}
	return false
}

// isGeneratedFile returns true if the file path looks like auto-generated code
// that should be excluded from the code knowledge graph. Generated code inflates
// the graph with thousands of boilerplate symbols (e.g., ORM predicates, CRUD methods)
// that don't help with code understanding.
func isGeneratedFile(path string) bool {
	// Check for common generated directory patterns
	normalized := filepath.ToSlash(path)
	for _, seg := range strings.Split(normalized, "/") {
		if seg == "generated" || seg == "gen" {
			return true
		}
	}
	return false
}

// truncateSource returns source trimmed to maxSourceLen runes.
// Uses rune-safe truncation to avoid splitting multi-byte UTF-8 characters.
func truncateSource(s string) string {
	runes := []rune(s)
	if len(runes) <= maxSourceLen {
		return s
	}
	return string(runes[:maxSourceLen])
}
