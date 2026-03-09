package analysis

import (
	"testing"
)

// --- Task 12: Language registry tests ---

// TestNewRegistry_ReturnsNonNil verifies that NewRegistry returns a non-nil registry.
// Expected result: Non-nil *Registry.
func TestNewRegistry_ReturnsNonNil(t *testing.T) {
	r := NewRegistry()
	if r == nil {
		t.Error("NewRegistry returned nil, expected non-nil *Registry")
	}
}

// TestNewRegistry_PreregistersGo verifies that NewRegistry pre-registers the Go
// language so that .go files are recognized without explicit Register calls.
// Expected result: LanguageForFile("main.go") returns Go language.
func TestNewRegistry_PreregistersGo(t *testing.T) {
	r := NewRegistry()

	lang, ok := r.LanguageForFile("main.go")
	if !ok {
		t.Fatal("LanguageForFile(main.go) returned false, expected Go to be pre-registered")
	}
	if lang == nil {
		t.Fatal("LanguageForFile(main.go) returned nil Language")
	}
	if lang.Name != "go" {
		t.Errorf("Language.Name = %q, want %q", lang.Name, "go")
	}
	if lang.Grammar == nil {
		t.Error("Language.Grammar is nil, expected compiled-in Go grammar")
	}
}

// TestNewRegistry_PreregistersTypeScript verifies that NewRegistry pre-registers
// TypeScript so that .ts files are recognized.
// Expected result: LanguageForFile("app.ts") returns TypeScript language.
func TestNewRegistry_PreregistersTypeScript(t *testing.T) {
	r := NewRegistry()

	lang, ok := r.LanguageForFile("app.ts")
	if !ok {
		t.Fatal("LanguageForFile(app.ts) returned false, expected TypeScript to be pre-registered")
	}
	if lang == nil {
		t.Fatal("LanguageForFile(app.ts) returned nil Language")
	}
	if lang.Name != "typescript" {
		t.Errorf("Language.Name = %q, want %q", lang.Name, "typescript")
	}
	if lang.Grammar == nil {
		t.Error("Language.Grammar is nil, expected compiled-in TypeScript grammar")
	}
}

// TestNewRegistry_PreregistersTSX verifies that NewRegistry pre-registers
// TSX so that .tsx files are recognized.
// Expected result: LanguageForFile("component.tsx") returns TSX language.
func TestNewRegistry_PreregistersTSX(t *testing.T) {
	r := NewRegistry()

	lang, ok := r.LanguageForFile("component.tsx")
	if !ok {
		t.Fatal("LanguageForFile(component.tsx) returned false, expected TSX to be pre-registered")
	}
	if lang == nil {
		t.Fatal("LanguageForFile(component.tsx) returned nil Language")
	}
	if lang.Name != "tsx" {
		t.Errorf("Language.Name = %q, want %q", lang.Name, "tsx")
	}
	if lang.Grammar == nil {
		t.Error("Language.Grammar is nil, expected compiled-in TSX grammar")
	}
}

// TestRegistry_UnknownExtension verifies that LanguageForFile returns false
// for file extensions that are not registered.
// Expected result: LanguageForFile returns nil, false.
func TestRegistry_UnknownExtension(t *testing.T) {
	r := NewRegistry()

	lang, ok := r.LanguageForFile("data.xyz")
	if ok {
		t.Error("LanguageForFile(data.xyz) returned true, want false for unknown extension")
	}
	if lang != nil {
		t.Error("LanguageForFile(data.xyz) returned non-nil Language, want nil")
	}
}

// TestRegistry_ExtractorForLanguage verifies that extractors are retrievable
// by language name after registration via RegisterExtractor.
// Expected result: ExtractorForLanguage returns non-nil Extractor for registered languages.
func TestRegistry_ExtractorForLanguage(t *testing.T) {
	r := NewRegistry()

	// Extractors are not pre-registered; register them manually
	languages := []string{"go", "typescript", "tsx"}
	for _, lang := range languages {
		r.RegisterExtractor(lang, &mockExtractor{})
	}

	for _, lang := range languages {
		t.Run(lang, func(t *testing.T) {
			ext, ok := r.ExtractorForLanguage(lang)
			if !ok {
				t.Errorf("ExtractorForLanguage(%q) returned false, expected registered extractor", lang)
			}
			if ext == nil {
				t.Errorf("ExtractorForLanguage(%q) returned nil Extractor", lang)
			}
		})
	}
}

// TestRegistry_NoExtractorsPreregistered verifies that NewRegistry does NOT
// pre-register extractors (only grammars). Extractors are wired by the caller.
// Expected result: ExtractorForLanguage returns false for all pre-registered grammars.
func TestRegistry_NoExtractorsPreregistered(t *testing.T) {
	r := NewRegistry()

	for _, lang := range []string{"go", "typescript", "tsx"} {
		t.Run(lang, func(t *testing.T) {
			ext, ok := r.ExtractorForLanguage(lang)
			if ok {
				t.Errorf("ExtractorForLanguage(%q) returned true, want false (extractors not pre-registered)", lang)
			}
			if ext != nil {
				t.Errorf("ExtractorForLanguage(%q) returned non-nil, want nil", lang)
			}
		})
	}
}

// TestRegistry_ExtractorForUnknownLanguage verifies that ExtractorForLanguage
// returns false for languages that are not registered.
// Expected result: ExtractorForLanguage returns nil, false.
func TestRegistry_ExtractorForUnknownLanguage(t *testing.T) {
	r := NewRegistry()

	ext, ok := r.ExtractorForLanguage("ruby")
	if ok {
		t.Error("ExtractorForLanguage(ruby) returned true, want false for unknown language")
	}
	if ext != nil {
		t.Error("ExtractorForLanguage(ruby) returned non-nil Extractor, want nil")
	}
}

// TestRegistry_Register verifies that Register correctly adds a language
// and its extractor to the registry.
// Expected result: After Register, LanguageForFile and ExtractorForLanguage work.
func TestRegistry_Register(t *testing.T) {
	r := NewRegistry()

	lang := Language{
		Name:       "python",
		Extensions: []string{".py"},
		Grammar:    nil, // no grammar needed for this test
	}
	r.Register(lang, &mockExtractor{})

	// Should now find .py files
	found, ok := r.LanguageForFile("script.py")
	if !ok {
		t.Error("LanguageForFile(script.py) returned false after Register")
	}
	if found == nil {
		t.Error("LanguageForFile(script.py) returned nil after Register")
	}

	// Should now find python extractor
	ext, ok := r.ExtractorForLanguage("python")
	if !ok {
		t.Error("ExtractorForLanguage(python) returned false after Register")
	}
	if ext == nil {
		t.Error("ExtractorForLanguage(python) returned nil after Register")
	}
}

// TestRegistry_RegisterExtractor verifies that RegisterExtractor correctly
// adds an extractor for a language that was already registered via NewRegistry.
// Expected result: ExtractorForLanguage returns the registered extractor.
func TestRegistry_RegisterExtractor(t *testing.T) {
	r := NewRegistry()

	// "go" language is pre-registered but has no extractor
	_, ok := r.ExtractorForLanguage("go")
	if ok {
		t.Fatal("expected no extractor before RegisterExtractor")
	}

	r.RegisterExtractor("go", &mockExtractor{})

	ext, ok := r.ExtractorForLanguage("go")
	if !ok {
		t.Error("ExtractorForLanguage(go) returned false after RegisterExtractor")
	}
	if ext == nil {
		t.Error("ExtractorForLanguage(go) returned nil after RegisterExtractor")
	}
}

// TestRegistry_RegisterWithNilExtractor verifies that Register with a nil
// extractor only registers the language, not the extractor.
func TestRegistry_RegisterWithNilExtractor(t *testing.T) {
	r := &Registry{
		languages:  make(map[string]*Language),
		extractors: make(map[string]Extractor),
	}

	r.Register(Language{Name: "rust", Extensions: []string{".rs"}}, nil)

	_, ok := r.LanguageForFile("lib.rs")
	if !ok {
		t.Error("LanguageForFile(lib.rs) returned false after Register with nil extractor")
	}

	_, ok = r.ExtractorForLanguage("rust")
	if ok {
		t.Error("ExtractorForLanguage(rust) returned true, want false (nil extractor)")
	}
}
