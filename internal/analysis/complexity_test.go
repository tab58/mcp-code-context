package analysis

import (
	"testing"

	sitter "github.com/smacker/go-tree-sitter"
)

// --- Task 1: ComplexityExtractor interface + Registry support ---

// mockComplexityExtractor is a test double that satisfies the ComplexityExtractor interface.
type mockComplexityExtractor struct {
	returnValue int
}

func (m *mockComplexityExtractor) ComputeComplexity(_ *sitter.Node, _ []byte) int {
	return m.returnValue
}

// TestComplexityExtractorInterface_Satisfiable verifies that the ComplexityExtractor
// interface can be satisfied by a concrete type (compile-time check).
// Expected result: Compiles without errors.
func TestComplexityExtractorInterface_Satisfiable(t *testing.T) {
	var ce ComplexityExtractor = &mockComplexityExtractor{returnValue: 1}
	if ce == nil {
		t.Error("ComplexityExtractor interface should be satisfiable by mockComplexityExtractor")
	}
}

// TestComplexityExtractor_BaseComplexity verifies that a minimal ComplexityExtractor
// returns base complexity of 1 for a no-op function.
// Expected result: ComputeComplexity returns 1.
func TestComplexityExtractor_BaseComplexity(t *testing.T) {
	ce := &mockComplexityExtractor{returnValue: 1}
	result := ce.ComputeComplexity(nil, nil)
	if result != 1 {
		t.Errorf("ComputeComplexity() = %d, want 1 (base complexity)", result)
	}
}

// TestRegistry_RegisterComplexityExtractor verifies that RegisterComplexityExtractor
// stores a ComplexityExtractor retrievable by language name.
// Expected result: ComplexityExtractorForLanguage returns the registered extractor.
func TestRegistry_RegisterComplexityExtractor(t *testing.T) {
	r := NewRegistry()
	mock := &mockComplexityExtractor{returnValue: 5}

	r.RegisterComplexityExtractor("go", mock)

	ext, ok := r.ComplexityExtractorForLanguage("go")
	if !ok {
		t.Fatal("ComplexityExtractorForLanguage(go) returned false after registration")
	}
	if ext == nil {
		t.Fatal("ComplexityExtractorForLanguage(go) returned nil after registration")
	}
	if ext.ComputeComplexity(nil, nil) != 5 {
		t.Errorf("registered extractor returned %d, want 5", ext.ComputeComplexity(nil, nil))
	}
}

// TestRegistry_ComplexityExtractorForUnknownLanguage verifies that
// ComplexityExtractorForLanguage returns false for unregistered languages.
// Expected result: Returns nil, false.
func TestRegistry_ComplexityExtractorForUnknownLanguage(t *testing.T) {
	r := NewRegistry()

	ext, ok := r.ComplexityExtractorForLanguage("ruby")
	if ok {
		t.Error("ComplexityExtractorForLanguage(ruby) returned true, want false")
	}
	if ext != nil {
		t.Error("ComplexityExtractorForLanguage(ruby) returned non-nil, want nil")
	}
}

// TestRegistry_ComplexityExtractorNotPreregistered verifies that NewRegistry
// does not pre-register any ComplexityExtractors.
// Expected result: ComplexityExtractorForLanguage returns false for all languages.
func TestRegistry_ComplexityExtractorNotPreregistered(t *testing.T) {
	r := NewRegistry()

	for _, lang := range []string{"go", "typescript", "tsx"} {
		t.Run(lang, func(t *testing.T) {
			_, ok := r.ComplexityExtractorForLanguage(lang)
			if ok {
				t.Errorf("ComplexityExtractorForLanguage(%q) returned true, want false (not pre-registered)", lang)
			}
		})
	}
}

// TestRegistry_RegisterMultipleComplexityExtractors verifies that different
// languages can have different ComplexityExtractors registered.
// Expected result: Each language returns its own extractor.
func TestRegistry_RegisterMultipleComplexityExtractors(t *testing.T) {
	r := NewRegistry()
	goExt := &mockComplexityExtractor{returnValue: 3}
	tsExt := &mockComplexityExtractor{returnValue: 7}

	r.RegisterComplexityExtractor("go", goExt)
	r.RegisterComplexityExtractor("typescript", tsExt)

	goResult, ok := r.ComplexityExtractorForLanguage("go")
	if !ok || goResult.ComputeComplexity(nil, nil) != 3 {
		t.Errorf("Go extractor returned %v, want 3", goResult)
	}

	tsResult, ok := r.ComplexityExtractorForLanguage("typescript")
	if !ok || tsResult.ComputeComplexity(nil, nil) != 7 {
		t.Errorf("TS extractor returned %v, want 7", tsResult)
	}
}
