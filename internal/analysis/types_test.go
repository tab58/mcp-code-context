package analysis

import (
	"testing"

	sitter "github.com/smacker/go-tree-sitter"
)

// --- Task 11: Types and Extractor interface tests ---

// TestSymbol_HasAllFields verifies that Symbol struct has all required fields
// from the spec: Name, Kind, Path, Language, Signature, Visibility, Source,
// LineNumber, LineCount, Decorators, ParentName.
// Expected result: All fields are addressable (compile-time check).
func TestSymbol_HasAllFields(t *testing.T) {
	s := Symbol{
		Name:       "MyFunc",
		Kind:       "function",
		Path:       "/path/to/file.go",
		Language:   "go",
		Signature:  "func MyFunc(x int) error",
		Visibility: "public",
		Source:     "func MyFunc(x int) error {\n\treturn nil\n}",
		LineNumber: 10,
		LineCount:  3,
		Decorators: []string{"//go:generate"},
		ParentName: "MyStruct",
	}

	if s.Name == "" {
		t.Error("Symbol.Name should not be empty")
	}
	if s.Kind == "" {
		t.Error("Symbol.Kind should not be empty")
	}
	if s.Path == "" {
		t.Error("Symbol.Path should not be empty")
	}
	if s.Language == "" {
		t.Error("Symbol.Language should not be empty")
	}
	if s.Signature == "" {
		t.Error("Symbol.Signature should not be empty")
	}
	if s.Visibility == "" {
		t.Error("Symbol.Visibility should not be empty")
	}
	if s.Source == "" {
		t.Error("Symbol.Source should not be empty")
	}
	if s.LineNumber == 0 {
		t.Error("Symbol.LineNumber should not be zero")
	}
	if s.LineCount == 0 {
		t.Error("Symbol.LineCount should not be zero")
	}
	if len(s.Decorators) == 0 {
		t.Error("Symbol.Decorators should not be empty")
	}
	if s.ParentName == "" {
		t.Error("Symbol.ParentName should not be empty")
	}
}

// TestReference_HasAllFields verifies that Reference struct has all required fields.
// Expected result: All fields are addressable (compile-time check).
func TestReference_HasAllFields(t *testing.T) {
	r := Reference{
		FromSymbol: "pkg.MyFunc",
		ToName:     "fmt.Println",
		Kind:       "calls",
		FilePath:   "/path/to/file.go",
	}

	if r.FromSymbol == "" {
		t.Error("Reference.FromSymbol should not be empty")
	}
	if r.ToName == "" {
		t.Error("Reference.ToName should not be empty")
	}
	if r.Kind == "" {
		t.Error("Reference.Kind should not be empty")
	}
	if r.FilePath == "" {
		t.Error("Reference.FilePath should not be empty")
	}
}

// TestFileAnalysis_HasAllFields verifies that FileAnalysis struct has all required fields.
// Expected result: All fields are addressable (compile-time check).
func TestFileAnalysis_HasAllFields(t *testing.T) {
	fa := FileAnalysis{
		FilePath: "/path/to/file.go",
		Language: "go",
		Symbols: []Symbol{
			{Name: "main", Kind: "function"},
		},
		References: []Reference{
			{FromSymbol: "main", ToName: "fmt.Println", Kind: "calls"},
		},
	}

	if fa.FilePath == "" {
		t.Error("FileAnalysis.FilePath should not be empty")
	}
	if fa.Language == "" {
		t.Error("FileAnalysis.Language should not be empty")
	}
	if len(fa.Symbols) == 0 {
		t.Error("FileAnalysis.Symbols should not be empty")
	}
	if len(fa.References) == 0 {
		t.Error("FileAnalysis.References should not be empty")
	}
}

// mockExtractor is a test double that satisfies the Extractor interface.
type mockExtractor struct{}

func (m *mockExtractor) ExtractSymbols(_ *sitter.Tree, _ []byte, _ string, _ string) ([]Symbol, error) {
	return nil, nil
}

func (m *mockExtractor) ExtractReferences(_ *sitter.Tree, _ []byte, _ string, _ string) ([]Reference, error) {
	return nil, nil
}

// TestExtractorInterface_Satisfiable verifies that the Extractor interface
// can be satisfied by a concrete type (compile-time check).
// Expected result: Compiles without errors.
func TestExtractorInterface_Satisfiable(t *testing.T) {
	var e Extractor = &mockExtractor{}
	if e == nil {
		t.Error("Extractor interface should be satisfiable by mockExtractor")
	}
}
