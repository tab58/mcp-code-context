package analysis

import (
	"testing"
)

// === Task 1: Register 3 new tree-sitter grammars in registry.go ===

// TestNewRegistry_PreregistersJavaScript verifies that NewRegistry pre-registers
// JavaScript so that .js and .jsx files are recognized.
// Expected result: LanguageForFile("app.js") returns JavaScript language with grammar.
func TestNewRegistry_PreregistersJavaScript(t *testing.T) {
	r := NewRegistry()

	tests := []struct {
		file string
		ext  string
	}{
		{"app.js", ".js"},
		{"component.jsx", ".jsx"},
	}

	for _, tt := range tests {
		t.Run(tt.file, func(t *testing.T) {
			lang, ok := r.LanguageForFile(tt.file)
			if !ok {
				t.Fatalf("LanguageForFile(%s) returned false, expected JavaScript to be pre-registered", tt.file)
			}
			if lang == nil {
				t.Fatalf("LanguageForFile(%s) returned nil Language", tt.file)
			}
			if lang.Name != "javascript" {
				t.Errorf("Language.Name = %q, want %q", lang.Name, "javascript")
			}
			if lang.Grammar == nil {
				t.Error("Language.Grammar is nil, expected compiled-in JavaScript grammar")
			}
		})
	}
}

// TestNewRegistry_PreregistersPython verifies that NewRegistry pre-registers
// Python so that .py files are recognized.
// Expected result: LanguageForFile("script.py") returns Python language with grammar.
func TestNewRegistry_PreregistersPython(t *testing.T) {
	r := NewRegistry()

	lang, ok := r.LanguageForFile("script.py")
	if !ok {
		t.Fatal("LanguageForFile(script.py) returned false, expected Python to be pre-registered")
	}
	if lang == nil {
		t.Fatal("LanguageForFile(script.py) returned nil Language")
	}
	if lang.Name != "python" {
		t.Errorf("Language.Name = %q, want %q", lang.Name, "python")
	}
	if lang.Grammar == nil {
		t.Error("Language.Grammar is nil, expected compiled-in Python grammar")
	}
}

// TestNewRegistry_PreregistersRuby verifies that NewRegistry pre-registers
// Ruby so that .rb files are recognized.
// Expected result: LanguageForFile("app.rb") returns Ruby language with grammar.
func TestNewRegistry_PreregistersRuby(t *testing.T) {
	r := NewRegistry()

	lang, ok := r.LanguageForFile("app.rb")
	if !ok {
		t.Fatal("LanguageForFile(app.rb) returned false, expected Ruby to be pre-registered")
	}
	if lang == nil {
		t.Fatal("LanguageForFile(app.rb) returned nil Language")
	}
	if lang.Name != "ruby" {
		t.Errorf("Language.Name = %q, want %q", lang.Name, "ruby")
	}
	if lang.Grammar == nil {
		t.Error("Language.Grammar is nil, expected compiled-in Ruby grammar")
	}
}

// TestNewRegistry_AllSixLanguages verifies that after pre-registration,
// the registry knows about all 6 languages (Go, TS, TSX, JS, Python, Ruby).
// Expected result: All 6 extensions resolve to the correct language name.
func TestNewRegistry_AllSixLanguages(t *testing.T) {
	r := NewRegistry()

	tests := []struct {
		file     string
		wantLang string
	}{
		{"main.go", "go"},
		{"app.ts", "typescript"},
		{"component.tsx", "tsx"},
		{"script.js", "javascript"},
		{"util.jsx", "javascript"},
		{"model.py", "python"},
		{"worker.rb", "ruby"},
	}

	for _, tt := range tests {
		t.Run(tt.file, func(t *testing.T) {
			lang, ok := r.LanguageForFile(tt.file)
			if !ok {
				t.Fatalf("LanguageForFile(%s) returned false", tt.file)
			}
			if lang.Name != tt.wantLang {
				t.Errorf("Language.Name = %q, want %q", lang.Name, tt.wantLang)
			}
		})
	}
}
