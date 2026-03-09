package analysis

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	sitter "github.com/smacker/go-tree-sitter"
	codedb "github.com/tab58/code-context/internal/clients/code_db"
	"github.com/tab58/go-ormql/pkg/client"
)

// mergeBatchSize is the number of items per merge mutation call.
// Kept small because merge mutations include full source code in parameters,
// which FalkorDB serializes into the CYPHER header. Large batches with source
// code cause FalkorDB to OOM in memory-constrained Docker environments.
const mergeBatchSize = 2

// edgeBatchSize is the number of items per connect mutation call.
// Even with property indexes, UNWIND + double MATCH + MERGE accumulates
// intermediate result sets in FalkorDB. Keep batches moderate.
const edgeBatchSize = 10

// maxSourceLen caps source code stored per symbol. The nomic-embed-text model
// has a 2048-token context window (~8000 chars), so truncating beyond that
// loses no embedding fidelity while drastically reducing query string size.
const maxSourceLen = 6000

// GraphQL mutations for code analysis graph writes via Client().Execute().
const (
	gqlMergeFunctions = `mutation($input: [FunctionMergeInput!]!) {
  mergeFunctions(input: $input) { functions { id name } }
}`

	gqlMergeClasss = `mutation($input: [ClassMergeInput!]!) {
  mergeClasss(input: $input) { classs { id name } }
}`

	gqlConnectFileFunctions = `mutation($input: [ConnectFileFunctionsInput!]!) {
  connectFileFunctions(input: $input) { relationshipsCreated }
}`

	gqlConnectFileClasses = `mutation($input: [ConnectFileClassesInput!]!) {
  connectFileClasses(input: $input) { relationshipsCreated }
}`

	gqlConnectClassMethods = `mutation($input: [ConnectClassMethodsInput!]!) {
  connectClassMethods(input: $input) { relationshipsCreated }
}`

	gqlConnectFunctionRepository = `mutation($input: [ConnectFunctionRepositoryInput!]!) {
  connectFunctionRepository(input: $input) { relationshipsCreated }
}`

	gqlConnectClassRepository = `mutation($input: [ConnectClassRepositoryInput!]!) {
  connectClassRepository(input: $input) { relationshipsCreated }
}`

	gqlMergeModules = `mutation($input: [ModuleMergeInput!]!) {
  mergeModules(input: $input) { modules { id name } }
}`

	gqlConnectModuleRepository = `mutation($input: [ConnectModuleRepositoryInput!]!) {
  connectModuleRepository(input: $input) { relationshipsCreated }
}`

	gqlConnectFunctionCalls = `mutation($input: [ConnectFunctionCallsInput!]!) {
  connectFunctionCalls(input: $input) { relationshipsCreated }
}`

	gqlConnectFileImports = `mutation($input: [ConnectFileImportsInput!]!) {
  connectFileImports(input: $input) { relationshipsCreated }
}`

	gqlConnectClassInherits = `mutation($input: [ConnectClassInheritsInput!]!) {
  connectClassInherits(input: $input) { relationshipsCreated }
}`

	gqlConnectClassImplements = `mutation($input: [ConnectClassImplementsInput!]!) {
  connectClassImplements(input: $input) { relationshipsCreated }
}`

	gqlConnectFunctionOverrides = `mutation($input: [ConnectFunctionOverridesInput!]!) {
  connectFunctionOverrides(input: $input) { relationshipsCreated }
}`

	gqlConnectModuleDependsOn = `mutation($input: [ConnectModuleDependsOnInput!]!) {
  connectModuleDependsOn(input: $input) { relationshipsCreated }
}`

	gqlConnectModuleFunctions = `mutation($input: [ConnectModuleFunctionsInput!]!) {
  connectModuleFunctions(input: $input) { relationshipsCreated }
}`

	gqlConnectModuleClasses = `mutation($input: [ConnectModuleClassesInput!]!) {
  connectModuleClasses(input: $input) { relationshipsCreated }
}`

)

// AnalyzeOption configures an Analyze call.
type AnalyzeOption func(*analyzeOptions)

type analyzeOptions struct {
	progress func(stage, message string)
}

// WithAnalyzeProgress returns an AnalyzeOption that sets the progress callback.
func WithAnalyzeProgress(fn func(stage, message string)) AnalyzeOption {
	return func(o *analyzeOptions) {
		o.progress = fn
	}
}

// AnalyzeResult holds the outcome of a two-pass analysis run.
type AnalyzeResult struct {
	Files              int
	Symbols            int
	References         int
	ResolvedReferences int
	UnresolvedNames    []string
}

// Analyzer coordinates two-pass AST analysis across files in a repository.
type Analyzer struct {
	registry *Registry
	db       *codedb.CodeDB
}

// NewAnalyzer creates an Analyzer with the given registry and database.
func NewAnalyzer(registry *Registry, db *codedb.CodeDB) *Analyzer {
	return &Analyzer{registry: registry, db: db}
}

// Analyze runs the full two-pass analysis on the given files for a repository.
// Pass 1: Parse all files and extract symbols + references.
// Pass 2: Build a symbol table and resolve references against it.
// Returns an AnalyzeResult summarizing what was found.
func (a *Analyzer) Analyze(ctx context.Context, repoID string, repoPath string, files []string, opts ...AnalyzeOption) (*AnalyzeResult, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	var options analyzeOptions
	for _, opt := range opts {
		opt(&options)
	}

	var allAnalyses []FileAnalysis

	// Pass 1: Parse files and extract symbols + references
	// Skip test files — they don't belong in the code knowledge graph
	// and can triple the number of function nodes.
	for _, filePath := range files {
		if err := ctx.Err(); err != nil {
			return nil, err
		}

		if isTestFile(filePath) {
			continue
		}

		if options.progress != nil {
			options.progress("analyzing", filePath)
		}

		fa, ok := a.analyzeFile(ctx, filePath, repoPath)
		if ok {
			allAnalyses = append(allAnalyses, fa)
		}
	}

	// Pass 1 graph writes: create nodes and structural edges
	var c *client.Client
	if a.db != nil {
		var err error
		c, err = a.db.ForRepo(ctx, repoID)
		if err != nil {
			return nil, fmt.Errorf("analyzer ForRepo(%s): %w", repoID, err)
		}

		if err := a.writePass1(ctx, c, repoID, allAnalyses); err != nil {
			return nil, fmt.Errorf("analyzer pass 1 graph writes: %w", err)
		}
		// Write source code separately to avoid bloating MERGE queries.
		// Source is inlined in the Redis command by the FalkorDB Go client,
		// so including it in MERGE parameters causes massive query strings.
		if err := a.writeSourceCode(ctx, c, allAnalyses); err != nil {
			return nil, fmt.Errorf("analyzer source code writes: %w", err)
		}
	}

	// Pass 2: Build symbol table and resolve references
	result := resolvePass(allAnalyses)
	result.Files = len(allAnalyses)

	// Pass 2 graph writes: create relationship edges for resolved references
	if a.db != nil {
		if err := a.writePass2(ctx, c, allAnalyses); err != nil {
			return nil, fmt.Errorf("analyzer pass 2 graph writes: %w", err)
		}
		if err := a.writeExternalReferences(ctx, c, repoID, allAnalyses); err != nil {
			return nil, fmt.Errorf("analyzer external reference writes: %w", err)
		}
	}

	return result, nil
}

// resolvePass builds a symbol table from all extracted symbols and resolves
// internal references against it. External references (IsExternal=true) are
// skipped — they are handled via ExternalReference nodes. Returns counts and
// lists unresolved names with a single summary log line.
func resolvePass(analyses []FileAnalysis) *AnalyzeResult {
	// Build symbol table: name -> list of symbols with that name
	symbolTable := make(map[string][]Symbol)
	// Build module path index: normalized import path -> module symbol
	// This allows relative imports like "../../core/services/auth.service"
	// to resolve against module symbols by their ImportPath.
	modulePathIndex := make(map[string]struct{})
	totalSymbols := 0
	totalRefs := 0

	for _, fa := range analyses {
		for _, sym := range fa.Symbols {
			symbolTable[sym.Name] = append(symbolTable[sym.Name], sym)
			totalSymbols++
			if sym.Kind == "module" {
				if sym.ImportPath != "" {
					// Index by import path (without extension) for matching
					normalized := normalizeModulePath(sym.ImportPath)
					modulePathIndex[normalized] = struct{}{}
				}
				if sym.Path != "" {
					// Also index by absolute path for resolving relative imports
					normalized := normalizeModulePath(sym.Path)
					modulePathIndex[normalized] = struct{}{}
				}
			}
		}
	}

	// Resolve only internal references against the symbol table
	resolved := 0
	unresolvedSet := make(map[string]struct{})

	for _, fa := range analyses {
		for _, ref := range fa.References {
			if ref.IsExternal {
				continue // skip external refs — handled via ExternalReference nodes
			}
			totalRefs++
			if _, found := symbolTable[ref.ToName]; found {
				resolved++
			} else if ref.Kind == "imports" && isRelativePath(ref.ToName) {
				resolvedAbs := resolveRelativeImport(fa.FilePath, ref.ToName)
				_, foundAbs := modulePathIndex[resolvedAbs]
				if !foundAbs {
					// Try index file (barrel export: import from "./dir" → "./dir/index")
					_, foundAbs = modulePathIndex[resolvedAbs+"/index"]
				}
				if foundAbs {
					resolved++
				} else {
					unresolvedSet[ref.ToName] = struct{}{}
				}
			} else {
				unresolvedSet[ref.ToName] = struct{}{}
			}
		}
	}

	var unresolved []string
	for name := range unresolvedSet {
		unresolved = append(unresolved, name)
	}

	if len(unresolved) > 0 {
		log.Printf("analyzer: %d unresolved references out of %d total", len(unresolved), totalRefs)
	}

	return &AnalyzeResult{
		Symbols:            totalSymbols,
		References:         totalRefs,
		ResolvedReferences: resolved,
		UnresolvedNames:    unresolved,
	}
}

// isRelativePath returns true if the path starts with "./" or "../".
func isRelativePath(p string) bool {
	return strings.HasPrefix(p, "./") || strings.HasPrefix(p, "../")
}

// normalizeModulePath strips common extensions from a module import path
// for consistent matching.
func normalizeModulePath(p string) string {
	p = filepath.ToSlash(p)
	for _, ext := range []string{".ts", ".tsx", ".js", ".jsx", ".go"} {
		p = strings.TrimSuffix(p, ext)
	}
	return p
}

// resolveRelativeImport resolves a relative import path against the directory
// of the referring file and returns a normalized path for matching.
func resolveRelativeImport(fromFile, relImport string) string {
	dir := filepath.Dir(fromFile)
	resolved := filepath.Join(dir, relImport)
	resolved = filepath.ToSlash(resolved)
	return normalizeModulePath(resolved)
}

// analyzeFile parses a single file with tree-sitter and extracts symbols and
// references. Returns the analysis and true on success, or zero value and false
// if the file is unsupported or any step fails (errors are logged, not fatal).
func (a *Analyzer) analyzeFile(ctx context.Context, filePath string, repoPath string) (FileAnalysis, bool) {
	lang, ok := a.registry.LanguageForFile(filePath)
	if !ok {
		return FileAnalysis{}, false
	}

	ext, ok := a.registry.ExtractorForLanguage(lang.Name)
	if !ok {
		return FileAnalysis{}, false
	}

	source, err := os.ReadFile(filePath)
	if err != nil {
		log.Printf("analyzer: failed to read %s: %v", filePath, err)
		return FileAnalysis{}, false
	}

	parser := sitter.NewParser()
	parser.SetLanguage(lang.Grammar)
	tree, err := parser.ParseCtx(ctx, nil, source)
	if err != nil {
		log.Printf("analyzer: failed to parse %s: %v", filePath, err)
		return FileAnalysis{}, false
	}

	symbols, err := ext.ExtractSymbols(tree, source, filePath, repoPath)
	if err != nil {
		log.Printf("analyzer: symbol extraction error in %s: %v", filePath, err)
		return FileAnalysis{}, false
	}

	refs, err := ext.ExtractReferences(tree, source, filePath, repoPath)
	if err != nil {
		log.Printf("analyzer: reference extraction error in %s: %v", filePath, err)
		return FileAnalysis{}, false
	}

	return FileAnalysis{
		FilePath:   filePath,
		Language:   lang.Name,
		Symbols:    symbols,
		References: refs,
	}, true
}

// writePass1 creates graph nodes for extracted symbols and structural edges
// (DEFINES, HAS_METHOD, BELONGS_TO) via Client().Execute() merge/connect mutations.
func (a *Analyzer) writePass1(ctx context.Context, c *client.Client, repoID string, analyses []FileAnalysis) error {
	var moduleInputs, funcInputs, classInputs []map[string]any
	var defFuncEdges, defClassEdges []map[string]any
	var methodEdges []map[string]any
	var moduleBelongsTo, funcBelongsTo, classBelongsTo []map[string]any

	for _, fa := range analyses {
		for _, sym := range fa.Symbols {
			switch sym.Kind {
			case "module":
				endLine := computeEndingLine(sym.LineNumber, sym.LineCount)
				moduleFields := map[string]any{
					"language":     sym.Language,
					"startingLine": sym.LineNumber,
					"endingLine":   endLine,
					"importPath":   sym.ImportPath,
					"visibility":   sym.Visibility,
					"kind":         sym.ModuleKind,
				}
				moduleInputs = append(moduleInputs, map[string]any{
					"match":    map[string]any{"name": sym.Name, "path": sym.Path},
					"onCreate": moduleFields,
					"onMatch":  moduleFields,
				})
				moduleBelongsTo = append(moduleBelongsTo, map[string]any{
					"from": map[string]any{"name": sym.Name, "path": sym.Path},
					"to":   map[string]any{"name": repoID},
				})

			case "function", "method":
				fields := buildFuncFields(sym)
				funcInputs = append(funcInputs, map[string]any{
					"match":    map[string]any{"name": sym.Name, "path": sym.Path},
					"onCreate": fields,
					"onMatch":  fields,
				})
				defFuncEdges = append(defFuncEdges, map[string]any{
					"from": map[string]any{"path": sym.Path},
					"to":   map[string]any{"name": sym.Name, "path": sym.Path},
				})
				funcBelongsTo = append(funcBelongsTo, map[string]any{
					"from": map[string]any{"name": sym.Name, "path": sym.Path},
					"to":   map[string]any{"name": repoID},
				})
				if sym.ParentName != "" {
					methodEdges = append(methodEdges, map[string]any{
						"from": map[string]any{"name": sym.ParentName},
						"to":   map[string]any{"name": sym.Name, "path": sym.Path},
					})
				}

			case "class", "struct", "interface", "enum":
				fields := buildFuncFields(sym)
				classInputs = append(classInputs, map[string]any{
					"match":    map[string]any{"name": sym.Name, "path": sym.Path},
					"onCreate": withKind(fields, sym.Kind),
					"onMatch":  withKind(fields, sym.Kind),
				})
				defClassEdges = append(defClassEdges, map[string]any{
					"from": map[string]any{"path": sym.Path},
					"to":   map[string]any{"name": sym.Name, "path": sym.Path},
				})
				classBelongsTo = append(classBelongsTo, map[string]any{
					"from": map[string]any{"name": sym.Name, "path": sym.Path},
					"to":   map[string]any{"name": repoID},
				})
			}
		}
	}

	// Merge nodes (small batches — source code in parameters)
	if err := batchMutate(ctx, c, moduleInputs, gqlMergeModules, mergeBatchSize); err != nil {
		return fmt.Errorf("mergeModules: %w", err)
	}
	if err := batchMutate(ctx, c, funcInputs, gqlMergeFunctions, mergeBatchSize); err != nil {
		return fmt.Errorf("mergeFunctions: %w", err)
	}
	if err := batchMutate(ctx, c, classInputs, gqlMergeClasss, mergeBatchSize); err != nil {
		return fmt.Errorf("mergeClasss: %w", err)
	}

	// DEFINES edges (large batches — small match keys only)
	if err := batchMutate(ctx, c, defFuncEdges, gqlConnectFileFunctions, edgeBatchSize); err != nil {
		return fmt.Errorf("connectFileFunctions: %w", err)
	}
	if err := batchMutate(ctx, c, defClassEdges, gqlConnectFileClasses, edgeBatchSize); err != nil {
		return fmt.Errorf("connectFileClasses: %w", err)
	}

	// HAS_METHOD edges
	if err := batchMutate(ctx, c, methodEdges, gqlConnectClassMethods, edgeBatchSize); err != nil {
		return fmt.Errorf("connectClassMethods: %w", err)
	}

	// BELONGS_TO edges
	if err := batchMutate(ctx, c, moduleBelongsTo, gqlConnectModuleRepository, edgeBatchSize); err != nil {
		return fmt.Errorf("connectModuleRepository: %w", err)
	}
	if err := batchMutate(ctx, c, funcBelongsTo, gqlConnectFunctionRepository, edgeBatchSize); err != nil {
		return fmt.Errorf("connectFunctionRepository: %w", err)
	}
	if err := batchMutate(ctx, c, classBelongsTo, gqlConnectClassRepository, edgeBatchSize); err != nil {
		return fmt.Errorf("connectClassRepository: %w", err)
	}

	return nil
}

// writePass2 creates relationship edges for resolved references
// (CALLS, IMPORTS, INHERITS, IMPLEMENTS, OVERRIDES, DEPENDS_ON, EXPORTS)
// via Client().Execute() connect mutations.
func (a *Analyzer) writePass2(ctx context.Context, c *client.Client, analyses []FileAnalysis) error {
	// Build symbol table for resolution checks
	symbolTable := make(map[string]struct{})
	for _, fa := range analyses {
		for _, sym := range fa.Symbols {
			symbolTable[sym.Name] = struct{}{}
		}
	}

	var callEdges, importEdges []map[string]any
	var inheritsEdges, implementsEdges, overridesEdges []map[string]any
	var dependsOnEdges, exportEdges []map[string]any

	for _, fa := range analyses {
		for _, ref := range fa.References {
			if _, found := symbolTable[ref.ToName]; !found {
				continue // skip unresolved
			}

			switch ref.Kind {
			case "calls":
				callEdges = append(callEdges, map[string]any{
					"from": map[string]any{"name": ref.FromSymbol, "path": ref.FilePath},
					"to":   map[string]any{"name": ref.ToName},
					"edge": map[string]any{"callType": "direct"},
				})
			case "imports":
				importEdges = append(importEdges, map[string]any{
					"from": map[string]any{"path": ref.FilePath},
					"to":   map[string]any{"name": ref.ToName},
				})
			case "inherits":
				inheritsEdges = append(inheritsEdges, map[string]any{
					"from": map[string]any{"name": ref.FromSymbol, "path": ref.FilePath},
					"to":   map[string]any{"name": ref.ToName},
				})
			case "implements":
				implementsEdges = append(implementsEdges, map[string]any{
					"from": map[string]any{"name": ref.FromSymbol, "path": ref.FilePath},
					"to":   map[string]any{"name": ref.ToName},
				})
			case "overrides":
				overridesEdges = append(overridesEdges, map[string]any{
					"from": map[string]any{"name": ref.FromSymbol, "path": ref.FilePath},
					"to":   map[string]any{"name": ref.ToName},
				})
			case "depends_on":
				dependsOnEdges = append(dependsOnEdges, map[string]any{
					"from": map[string]any{"name": ref.FromSymbol, "path": ref.FilePath},
					"to":   map[string]any{"name": ref.ToName},
				})
			case "exports":
				exportEdges = append(exportEdges, map[string]any{
					"from": map[string]any{"name": ref.FromSymbol, "path": ref.FilePath},
					"to":   map[string]any{"name": ref.ToName},
				})
			}
		}
	}

	if err := batchMutate(ctx, c, callEdges, gqlConnectFunctionCalls, edgeBatchSize); err != nil {
		return fmt.Errorf("connectFunctionCalls: %w", err)
	}
	if err := batchMutate(ctx, c, importEdges, gqlConnectFileImports, edgeBatchSize); err != nil {
		return fmt.Errorf("connectFileImports: %w", err)
	}
	if err := batchMutate(ctx, c, inheritsEdges, gqlConnectClassInherits, edgeBatchSize); err != nil {
		return fmt.Errorf("connectClassInherits: %w", err)
	}
	if err := batchMutate(ctx, c, implementsEdges, gqlConnectClassImplements, edgeBatchSize); err != nil {
		return fmt.Errorf("connectClassImplements: %w", err)
	}
	if err := batchMutate(ctx, c, overridesEdges, gqlConnectFunctionOverrides, edgeBatchSize); err != nil {
		return fmt.Errorf("connectFunctionOverrides: %w", err)
	}
	if err := batchMutate(ctx, c, dependsOnEdges, gqlConnectModuleDependsOn, edgeBatchSize); err != nil {
		return fmt.Errorf("connectModuleDependsOn: %w", err)
	}
	if err := batchMutate(ctx, c, exportEdges, gqlConnectModuleFunctions, edgeBatchSize); err != nil {
		return fmt.Errorf("connectModuleExports: %w", err)
	}

	return nil
}

// gqlSetSource is a raw Cypher query that sets source on a node matched by name+path.
// Uses MATCH+SET (no MERGE) so no graph scan for existence checking is needed.
const gqlSetSource = `MATCH (n {name: $name, path: $path}) SET n.source = $source`

// writeSourceCode sets source code on Function and Class nodes one at a time.
// Separated from MERGE to keep merge queries small — the FalkorDB Go client
// inlines all parameters into the Redis command string, so large source code
// values in MERGE batches create massive commands that overwhelm memory.
func (a *Analyzer) writeSourceCode(ctx context.Context, c *client.Client, analyses []FileAnalysis) error {
	for _, fa := range analyses {
		for _, sym := range fa.Symbols {
			if sym.Source == "" {
				continue
			}
			switch sym.Kind {
			case "function", "method", "class", "struct", "interface", "enum":
				if err := ctx.Err(); err != nil {
					return err
				}
				_, err := c.ExecuteRaw(ctx, gqlSetSource, map[string]any{
					"name":   sym.Name,
					"path":   sym.Path,
					"source": truncateSource(sym.Source),
				})
				if err != nil {
					return fmt.Errorf("setSource %s/%s: %w", sym.Path, sym.Name, err)
				}
			}
		}
	}
	return nil
}

// isTestFile returns true if the file path looks like a test file that should
// be excluded from the code knowledge graph. Matches Go test files (*_test.go),
// JavaScript/TypeScript test files (*.test.*, *.spec.*), and test directories.
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
	return false
}

// truncateSource returns source trimmed to maxSourceLen characters.
// Truncation is safe because the embedding model's context window is smaller
// than maxSourceLen, and the stored source is only used for embeddings and display.
func truncateSource(s string) string {
	if len(s) <= maxSourceLen {
		return s
	}
	return s[:maxSourceLen]
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
