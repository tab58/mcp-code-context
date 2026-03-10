package analysis

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	sitter "github.com/smacker/go-tree-sitter"
	codedb "github.com/tab58/code-context/internal/clients/code_db"
	"github.com/tab58/go-ormql/pkg/client"
)

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

		if isTestFile(filePath) || isGeneratedFile(filePath) {
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

		pass1Start := time.Now()
		if err := a.writePass1(ctx, c, repoID, repoPath, allAnalyses); err != nil {
			return nil, fmt.Errorf("analyzer pass 1 graph writes: %w", err)
		}
		log.Printf("[DEBUG] pass1 total: %s", time.Since(pass1Start))

		// Write source code separately to avoid bloating MERGE queries.
		// Source is inlined in the Redis command by the FalkorDB Go client,
		// so including it in MERGE parameters causes massive query strings.
		sourceStart := time.Now()
		if err := a.writeSourceCode(ctx, c, allAnalyses); err != nil {
			return nil, fmt.Errorf("analyzer source code writes: %w", err)
		}
		log.Printf("[DEBUG] writeSourceCode total: %s", time.Since(sourceStart))
	}

	// Pass 2: Build symbol table and resolve references
	result := resolvePass(allAnalyses)
	result.Files = len(allAnalyses)

	// Pass 2 graph writes: create relationship edges for resolved references
	if a.db != nil {
		pass2Start := time.Now()
		if err := a.writePass2(ctx, c, repoPath, allAnalyses); err != nil {
			return nil, fmt.Errorf("analyzer pass 2 graph writes: %w", err)
		}
		log.Printf("[DEBUG] pass2 edges total: %s", time.Since(pass2Start))
		if err := a.writeExternalReferences(ctx, c, repoID, repoPath, allAnalyses); err != nil {
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
func (a *Analyzer) writePass1(ctx context.Context, c *client.Client, repoID string, repoPath string, analyses []FileAnalysis) error {
	var moduleInputs, funcInputs, classInputs []map[string]any
	var defFuncEdges, defClassEdges []map[string]any
	var methodEdges []map[string]any
	var moduleBelongsTo, funcBelongsTo, classBelongsTo []map[string]any

	for _, fa := range analyses {
		// File nodes use relative paths; compute relative path for File matching.
		fileRelPath, _ := filepath.Rel(repoPath, fa.FilePath)

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
					"from": map[string]any{"path": fileRelPath},
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
					"from": map[string]any{"path": fileRelPath},
					"to":   map[string]any{"name": sym.Name, "path": sym.Path},
				})
				classBelongsTo = append(classBelongsTo, map[string]any{
					"from": map[string]any{"name": sym.Name, "path": sym.Path},
					"to":   map[string]any{"name": repoID},
				})
			}
		}
	}

	log.Printf("[DEBUG] Symbol counts: modules=%d functions=%d classes=%d defFuncEdges=%d defClassEdges=%d methodEdges=%d",
		len(moduleInputs), len(funcInputs), len(classInputs), len(defFuncEdges), len(defClassEdges), len(methodEdges))

	// Merge nodes (small batches — source code in parameters)
	mergeStart := time.Now()
	if err := batchMutate(ctx, c, moduleInputs, gqlMergeModules, mergeBatchSize); err != nil {
		return fmt.Errorf("mergeModules: %w", err)
	}
	log.Printf("[DEBUG] mergeModules complete (%d items, %s)", len(moduleInputs), time.Since(mergeStart))
	funcStart := time.Now()
	if err := batchMutate(ctx, c, funcInputs, gqlMergeFunctions, mergeBatchSize); err != nil {
		return fmt.Errorf("mergeFunctions: %w", err)
	}
	log.Printf("[DEBUG] mergeFunctions complete (%d items, %s)", len(funcInputs), time.Since(funcStart))
	classStart := time.Now()
	if err := batchMutate(ctx, c, classInputs, gqlMergeClasss, mergeBatchSize); err != nil {
		return fmt.Errorf("mergeClasss: %w", err)
	}
	log.Printf("[DEBUG] mergeClasss complete (%d items, %s)", len(classInputs), time.Since(classStart))

	// DEFINES edges — batched UNWIND MATCH+CREATE
	edgeStart := time.Now()
	defFuncSpec := edgeSpec{
		FromLabel: "File", FromWhere: map[string]string{"path": "from_path"},
		ToLabel: "Function", ToWhere: map[string]string{"name": "to_name", "path": "to_path"},
		RelType: "DEFINES",
	}
	if err := createEdgesRaw(ctx, c, defFuncEdges, defFuncSpec); err != nil {
		return fmt.Errorf("connectFileFunctions: %w", err)
	}
	defClassSpec := edgeSpec{
		FromLabel: "File", FromWhere: map[string]string{"path": "from_path"},
		ToLabel: "Class", ToWhere: map[string]string{"name": "to_name", "path": "to_path"},
		RelType: "DEFINES",
	}
	if err := createEdgesRaw(ctx, c, defClassEdges, defClassSpec); err != nil {
		return fmt.Errorf("connectFileClasses: %w", err)
	}

	// HAS_METHOD edges
	methodSpec := edgeSpec{
		FromLabel: "Class", FromWhere: map[string]string{"name": "from_name"},
		ToLabel: "Function", ToWhere: map[string]string{"name": "to_name", "path": "to_path"},
		RelType: "HAS_METHOD",
	}
	if err := createEdgesRaw(ctx, c, methodEdges, methodSpec); err != nil {
		return fmt.Errorf("connectClassMethods: %w", err)
	}

	// BELONGS_TO edges
	moduleBelongsToSpec := edgeSpec{
		FromLabel: "Module", FromWhere: map[string]string{"name": "from_name", "path": "from_path"},
		ToLabel: "Repository", ToWhere: map[string]string{"name": "to_name"},
		RelType: "BELONGS_TO",
	}
	if err := createEdgesRaw(ctx, c, moduleBelongsTo, moduleBelongsToSpec); err != nil {
		return fmt.Errorf("connectModuleRepository: %w", err)
	}
	funcBelongsToSpec := edgeSpec{
		FromLabel: "Function", FromWhere: map[string]string{"name": "from_name", "path": "from_path"},
		ToLabel: "Repository", ToWhere: map[string]string{"name": "to_name"},
		RelType: "BELONGS_TO",
	}
	if err := createEdgesRaw(ctx, c, funcBelongsTo, funcBelongsToSpec); err != nil {
		return fmt.Errorf("connectFunctionRepository: %w", err)
	}
	classBelongsToSpec := edgeSpec{
		FromLabel: "Class", FromWhere: map[string]string{"name": "from_name", "path": "from_path"},
		ToLabel: "Repository", ToWhere: map[string]string{"name": "to_name"},
		RelType: "BELONGS_TO",
	}
	if err := createEdgesRaw(ctx, c, classBelongsTo, classBelongsToSpec); err != nil {
		return fmt.Errorf("connectClassRepository: %w", err)
	}
	log.Printf("[DEBUG] pass1 edges complete (%s)", time.Since(edgeStart))

	return nil
}

// writePass2 creates relationship edges for resolved references
// (CALLS, IMPORTS, INHERITS, IMPLEMENTS, OVERRIDES, DEPENDS_ON, EXPORTS)
// via Client().Execute() connect mutations.
func (a *Analyzer) writePass2(ctx context.Context, c *client.Client, repoPath string, analyses []FileAnalysis) error {
	// Build symbol count table: name -> count of symbols with that name.
	// Names that appear more than once are ambiguous — creating edges to all
	// matching targets produces a cartesian product (e.g., 32 "constructor"
	// functions × 100 call sites = 3,200 spurious edges).
	symbolCount := make(map[string]int)
	for _, fa := range analyses {
		for _, sym := range fa.Symbols {
			symbolCount[sym.Name]++
		}
	}

	var callEdges, importEdges []map[string]any
	var inheritsEdges, implementsEdges, overridesEdges []map[string]any
	var dependsOnEdges, exportEdges []map[string]any
	skippedAmbiguous := 0

	for _, fa := range analyses {
		// File nodes use relative paths; compute for IMPORTS edge matching.
		fileRelPath, _ := filepath.Rel(repoPath, fa.FilePath)

		for _, ref := range fa.References {
			cnt := symbolCount[ref.ToName]
			if cnt == 0 {
				continue // skip unresolved
			}
			if cnt > 1 {
				skippedAmbiguous++
				continue // skip ambiguous — would create cartesian product
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
					"from": map[string]any{"path": fileRelPath},
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

	if skippedAmbiguous > 0 {
		log.Printf("[DEBUG] pass2: skipped %d ambiguous references (target name matches multiple symbols)", skippedAmbiguous)
	}

	// Raw Cypher MATCH+CREATE for all pass 2 edges (avoids UNWIND+MERGE memory spikes)
	callSpec := edgeSpec{
		FromLabel: "Function", FromWhere: map[string]string{"name": "from_name", "path": "from_path"},
		ToLabel: "Function", ToWhere: map[string]string{"name": "to_name"},
		RelType: "CALLS",
		EdgeProps: map[string]string{"callType": "edge_callType"},
	}
	if err := createEdgesRaw(ctx, c, callEdges, callSpec); err != nil {
		return fmt.Errorf("connectFunctionCalls: %w", err)
	}
	importSpec := edgeSpec{
		FromLabel: "File", FromWhere: map[string]string{"path": "from_path"},
		ToLabel: "Module", ToWhere: map[string]string{"name": "to_name"},
		RelType: "IMPORTS",
	}
	if err := createEdgesRaw(ctx, c, importEdges, importSpec); err != nil {
		return fmt.Errorf("connectFileImports: %w", err)
	}
	inheritsSpec := edgeSpec{
		FromLabel: "Class", FromWhere: map[string]string{"name": "from_name", "path": "from_path"},
		ToLabel: "Class", ToWhere: map[string]string{"name": "to_name"},
		RelType: "INHERITS",
	}
	if err := createEdgesRaw(ctx, c, inheritsEdges, inheritsSpec); err != nil {
		return fmt.Errorf("connectClassInherits: %w", err)
	}
	implementsSpec := edgeSpec{
		FromLabel: "Class", FromWhere: map[string]string{"name": "from_name", "path": "from_path"},
		ToLabel: "Class", ToWhere: map[string]string{"name": "to_name"},
		RelType: "IMPLEMENTS",
	}
	if err := createEdgesRaw(ctx, c, implementsEdges, implementsSpec); err != nil {
		return fmt.Errorf("connectClassImplements: %w", err)
	}
	overridesSpec := edgeSpec{
		FromLabel: "Function", FromWhere: map[string]string{"name": "from_name", "path": "from_path"},
		ToLabel: "Function", ToWhere: map[string]string{"name": "to_name"},
		RelType: "OVERRIDES",
	}
	if err := createEdgesRaw(ctx, c, overridesEdges, overridesSpec); err != nil {
		return fmt.Errorf("connectFunctionOverrides: %w", err)
	}
	dependsOnSpec := edgeSpec{
		FromLabel: "Module", FromWhere: map[string]string{"name": "from_name", "path": "from_path"},
		ToLabel: "Module", ToWhere: map[string]string{"name": "to_name"},
		RelType: "DEPENDS_ON",
	}
	if err := createEdgesRaw(ctx, c, dependsOnEdges, dependsOnSpec); err != nil {
		return fmt.Errorf("connectModuleDependsOn: %w", err)
	}
	exportSpec := edgeSpec{
		FromLabel: "Module", FromWhere: map[string]string{"name": "from_name", "path": "from_path"},
		ToLabel: "Function", ToWhere: map[string]string{"name": "to_name"},
		RelType: "EXPORTS",
	}
	if err := createEdgesRaw(ctx, c, exportEdges, exportSpec); err != nil {
		return fmt.Errorf("connectModuleExports: %w", err)
	}

	return nil
}

// ComputeComplexity computes cyclomatic complexity for all Function symbols
// in the given files and writes values to the cyclomaticComplexity field in FalkorDB.
// This is a post-processing step called after Analyze.
func (a *Analyzer) ComputeComplexity(ctx context.Context, repoID string, repoPath string, files []string, opts ...AnalyzeOption) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	var options analyzeOptions
	for _, opt := range opts {
		opt(&options)
	}

	var items []map[string]any

	for _, filePath := range files {
		if err := ctx.Err(); err != nil {
			return err
		}

		if isTestFile(filePath) || isGeneratedFile(filePath) {
			continue
		}

		lang, ok := a.registry.LanguageForFile(filePath)
		if !ok {
			continue
		}

		ce, ok := a.registry.ComplexityExtractorForLanguage(lang.Name)
		if !ok {
			continue
		}

		source, err := os.ReadFile(filePath)
		if err != nil {
			log.Printf("complexity: failed to read %s: %v", filePath, err)
			continue
		}

		parser := sitter.NewParser()
		parser.SetLanguage(lang.Grammar)
		tree, err := parser.ParseCtx(ctx, nil, source)
		if err != nil {
			log.Printf("complexity: failed to parse %s: %v", filePath, err)
			continue
		}

		if options.progress != nil {
			options.progress("complexity", filePath)
		}

		// Walk the AST to find function/method nodes and compute complexity
		root := tree.RootNode()
		walkForFunctions(root, source, filePath, lang.Name, ce, &items)
	}

	// Write complexity values to graph via batched Cypher
	if a.db != nil && len(items) > 0 {
		c, err := a.db.ForRepo(ctx, repoID)
		if err != nil {
			return fmt.Errorf("complexity ForRepo(%s): %w", repoID, err)
		}
		if err := c.ExecuteRawBatch(ctx, gqlSetComplexityBatch, items, edgeBatchSize); err != nil {
			return fmt.Errorf("complexity batch write: %w", err)
		}
	}

	return nil
}

// walkForFunctions recursively finds function/method AST nodes and computes
// their cyclomatic complexity.
func walkForFunctions(node *sitter.Node, source []byte, filePath, langName string, ce ComplexityExtractor, items *[]map[string]any) {
	nodeType := node.Type()

	isFuncNode := false
	switch langName {
	case "go":
		isFuncNode = nodeType == "function_declaration" || nodeType == "method_declaration"
	case "typescript", "tsx":
		isFuncNode = nodeType == "function_declaration" || nodeType == "method_definition"
	}

	if isFuncNode {
		complexity := ce.ComputeComplexity(node, source)
		nameNode := node.ChildByFieldName("name")
		if nameNode != nil {
			*items = append(*items, map[string]any{
				"name":       nameNode.Content(source),
				"path":       filePath,
				"complexity": complexity,
			})
		}
		return // don't recurse into function bodies (nested functions are separate)
	}

	for i := 0; i < int(node.NamedChildCount()); i++ {
		child := node.NamedChild(i)
		if child != nil {
			walkForFunctions(child, source, filePath, langName, ce, items)
		}
	}
}

// writeSourceCode sets source code on Function and Class nodes via batched UNWIND.
// Separated from MERGE to keep merge queries small — the FalkorDB Go client
// inlines all parameters into the Redis command string, so large source code
// values in MERGE batches create massive commands that overwhelm memory.
func (a *Analyzer) writeSourceCode(ctx context.Context, c *client.Client, analyses []FileAnalysis) error {
	var items []map[string]any
	for _, fa := range analyses {
		for _, sym := range fa.Symbols {
			if sym.Source == "" {
				continue
			}
			switch sym.Kind {
			case "function", "method", "class", "struct", "interface", "enum":
				items = append(items, map[string]any{
					"name":   sym.Name,
					"path":   sym.Path,
					"source": truncateSource(sym.Source),
				})
			}
		}
	}
	if len(items) == 0 {
		return nil
	}
	log.Printf("[DEBUG] writeSourceCode: %d items in batches of %d", len(items), sourceCodeBatchSize)
	return c.ExecuteRawBatch(ctx, gqlSetSourceBatch, items, sourceCodeBatchSize)
}


