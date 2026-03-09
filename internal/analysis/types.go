package analysis

// Symbol represents a code entity extracted from an AST.
type Symbol struct {
	Name       string   // symbol name (e.g., "MyFunc", "UserService")
	Kind       string   // "function", "method", "class", "struct", "interface", "module"
	Path       string   // file path where this symbol is defined
	Language   string   // "go", "typescript", "tsx"
	Signature  string   // function/method signature string
	Visibility string   // "public", "private", "protected", "internal", "package"
	Source     string   // full source text of the symbol
	LineNumber int      // 1-based start line
	LineCount  int      // number of lines
	Decorators []string // annotations/decorators (Go tags, TS decorators)
	ParentName string   // for methods: the class/struct/interface name
	ImportPath string   // for module symbols: fully qualified import path
	ModuleKind string   // for module symbols: "package", "esm", "cjs"
}

// Reference represents a cross-symbol relationship found in source code.
type Reference struct {
	FromSymbol         string // qualified name of source symbol (e.g., "pkg.MyFunc")
	ToName             string // unresolved target name (e.g., "OtherFunc", "fmt.Println")
	Kind               string // "calls", "imports", "inherits", "implements", "overrides"
	FilePath           string // file where this reference was found
	IsExternal         bool   // true = targets an external package (not repo-internal)
	ExternalImportPath string // full import path for external refs (e.g., "fmt", "react")
}

// FileAnalysis holds the results of analyzing a single source file.
type FileAnalysis struct {
	FilePath   string
	Language   string
	Symbols    []Symbol
	References []Reference
}
