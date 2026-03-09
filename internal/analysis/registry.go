package analysis

import (
	"path/filepath"
	"strings"

	sitter "github.com/smacker/go-tree-sitter"
	"github.com/smacker/go-tree-sitter/golang"
	"github.com/smacker/go-tree-sitter/typescript/tsx"
	"github.com/smacker/go-tree-sitter/typescript/typescript"
)

// Language represents a supported programming language with its tree-sitter grammar.
type Language struct {
	Name       string           // "go", "typescript", "tsx"
	Extensions []string         // [".go"], [".ts"], [".tsx"]
	Grammar    *sitter.Language // compiled-in tree-sitter grammar
}

// Registry maps file extensions to language grammars and extractors.
type Registry struct {
	languages  map[string]*Language // ".go" -> Language
	extractors map[string]Extractor // "go" -> GoExtractor
}

// NewRegistry creates a Registry with Go, TypeScript, and TSX grammars
// pre-registered. Extractors must be registered separately via
// RegisterExtractor or by calling the sub-package Register functions
// (e.g., golang.Register, typescript.Register).
func NewRegistry() *Registry {
	r := &Registry{
		languages:  make(map[string]*Language),
		extractors: make(map[string]Extractor),
	}

	r.registerLanguage(Language{Name: "go", Extensions: []string{".go"}, Grammar: golang.GetLanguage()})
	r.registerLanguage(Language{Name: "typescript", Extensions: []string{".ts"}, Grammar: typescript.GetLanguage()})
	r.registerLanguage(Language{Name: "tsx", Extensions: []string{".tsx"}, Grammar: tsx.GetLanguage()})

	return r
}

// Register adds a language and its extractor to the registry.
func (r *Registry) Register(lang Language, ext Extractor) {
	r.registerLanguage(lang)
	if ext != nil {
		r.extractors[lang.Name] = ext
	}
}

// RegisterExtractor registers an extractor for a language by name.
// The language must already be registered via NewRegistry or Register.
func (r *Registry) RegisterExtractor(langName string, ext Extractor) {
	r.extractors[langName] = ext
}

// registerLanguage adds a language (grammar + extensions) without an extractor.
func (r *Registry) registerLanguage(lang Language) {
	l := &lang
	for _, e := range lang.Extensions {
		r.languages[strings.ToLower(e)] = l
	}
}

// LanguageForFile returns the Language for a given file path based on its extension.
// Returns nil, false if the file extension is not registered.
func (r *Registry) LanguageForFile(path string) (*Language, bool) {
	ext := strings.ToLower(filepath.Ext(path))
	lang, ok := r.languages[ext]
	return lang, ok
}

// ExtractorForLanguage returns the Extractor for a given language name.
func (r *Registry) ExtractorForLanguage(name string) (Extractor, bool) {
	ext, ok := r.extractors[name]
	return ext, ok
}
