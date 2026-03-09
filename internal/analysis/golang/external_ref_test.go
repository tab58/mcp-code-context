package golang

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/tab58/code-context/internal/analysis"
)

// === Task 5: Go extractor reference enhancements ===

// TestGoExtractor_FiltersBuiltins verifies that Go built-in identifiers
// (append, len, make, etc.) are filtered out of call references.
// Expected result: No reference to "make" or "len" in extracted refs.
func TestGoExtractor_FiltersBuiltins(t *testing.T) {
	source := []byte(`package main

func main() {
	s := make([]int, 0)
	n := len(s)
	s = append(s, n)
	println(n)
}
`)
	tree := parseGo(t, source)
	ext := NewGoExtractor()

	refs, err := ext.ExtractReferences(tree, source, "main.go", "")
	if err != nil {
		t.Fatalf("ExtractReferences error: %v", err)
	}

	builtins := map[string]bool{"make": true, "len": true, "append": true, "println": true}
	for _, ref := range refs {
		if ref.Kind == "calls" && builtins[ref.ToName] {
			t.Errorf("built-in %q should be filtered out of call references", ref.ToName)
		}
	}
}

// TestGoExtractor_SkipsFuncLiteralCalls verifies that anonymous function
// literal calls (func() {}()) are skipped.
// Expected result: No call reference emitted for func literal.
func TestGoExtractor_SkipsFuncLiteralCalls(t *testing.T) {
	source := []byte(`package main

func main() {
	func() {
		// anonymous
	}()
}
`)
	tree := parseGo(t, source)
	ext := NewGoExtractor()

	refs, err := ext.ExtractReferences(tree, source, "main.go", "")
	if err != nil {
		t.Fatalf("ExtractReferences error: %v", err)
	}

	for _, ref := range refs {
		if ref.Kind == "calls" {
			// The only call is the func literal — it should be skipped
			t.Errorf("unexpected call reference %q — func_literal calls should be skipped", ref.ToName)
		}
	}
}

// TestGoExtractor_NormalizesSelectorExpressions verifies that method calls
// via selector expressions (e.g., obj.Method()) are normalized to bare method name.
// Expected result: Reference ToName is "Method", not "obj.Method".
func TestGoExtractor_NormalizesSelectorExpressions(t *testing.T) {
	source := []byte(`package main

type Foo struct{}

func (f *Foo) Method() {}

func main() {
	f := &Foo{}
	f.Method()
}
`)
	tree := parseGo(t, source)
	ext := NewGoExtractor()

	refs, err := ext.ExtractReferences(tree, source, "main.go", "")
	if err != nil {
		t.Fatalf("ExtractReferences error: %v", err)
	}

	found := false
	for _, ref := range refs {
		if ref.Kind == "calls" && ref.ToName == "Method" {
			found = true
		}
		if ref.Kind == "calls" && ref.ToName == "f.Method" {
			t.Error("selector expression not normalized — ToName should be 'Method', not 'f.Method'")
		}
	}
	if !found {
		t.Error("expected normalized call reference to 'Method' (bare method name)")
	}
}

// TestGoExtractor_ClassifiesStdlibImportsAsExternal verifies that stdlib
// imports (no dots in first path segment) are marked as external.
// Expected result: Import reference for "fmt" has IsExternal=true.
func TestGoExtractor_ClassifiesStdlibImportsAsExternal(t *testing.T) {
	source := []byte(`package main

import "fmt"

func main() {
	fmt.Println("hello")
}
`)
	tree := parseGo(t, source)
	ext := NewGoExtractor()

	refs, err := ext.ExtractReferences(tree, source, "main.go", "")
	if err != nil {
		t.Fatalf("ExtractReferences error: %v", err)
	}

	var fmtImport *analysis.Reference
	for i, ref := range refs {
		if ref.Kind == "imports" && ref.ToName == "fmt" {
			fmtImport = &refs[i]
			break
		}
	}

	if fmtImport == nil {
		t.Fatal("expected to find import reference for 'fmt'")
	}
	if !fmtImport.IsExternal {
		t.Error("stdlib import 'fmt' should have IsExternal=true")
	}
	if fmtImport.ExternalImportPath != "fmt" {
		t.Errorf("ExternalImportPath = %q, want %q", fmtImport.ExternalImportPath, "fmt")
	}
}

// TestGoExtractor_ClassifiesExternalDepsAsExternal verifies that external
// dependencies are marked as external.
// Expected result: Import reference for external dep has IsExternal=true.
func TestGoExtractor_ClassifiesExternalDepsAsExternal(t *testing.T) {
	dir := t.TempDir()
	gomod := filepath.Join(dir, "go.mod")
	if err := os.WriteFile(gomod, []byte("module github.com/myorg/myrepo\n\ngo 1.25\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	source := []byte(`package main

import "github.com/other/dep"

func main() {
	dep.DoSomething()
}
`)
	tree := parseGo(t, source)
	ext := NewGoExtractor()

	refs, err := ext.ExtractReferences(tree, source, filepath.Join(dir, "main.go"), dir)
	if err != nil {
		t.Fatalf("ExtractReferences error: %v", err)
	}

	var depImport *analysis.Reference
	for i, ref := range refs {
		if ref.Kind == "imports" && ref.ToName == "dep" {
			depImport = &refs[i]
			break
		}
	}

	if depImport == nil {
		t.Fatal("expected to find import reference for 'dep'")
	}
	if !depImport.IsExternal {
		t.Error("external dep import should have IsExternal=true")
	}
	if depImport.ExternalImportPath != "github.com/other/dep" {
		t.Errorf("ExternalImportPath = %q, want %q", depImport.ExternalImportPath, "github.com/other/dep")
	}
}

// TestGoExtractor_ClassifiesInternalImportsAsInternal verifies that imports
// under the module path are marked as internal (not external).
// Expected result: Import reference for internal package has IsExternal=false.
func TestGoExtractor_ClassifiesInternalImportsAsInternal(t *testing.T) {
	dir := t.TempDir()
	gomod := filepath.Join(dir, "go.mod")
	if err := os.WriteFile(gomod, []byte("module github.com/myorg/myrepo\n\ngo 1.25\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	source := []byte(`package main

import "github.com/myorg/myrepo/internal/utils"

func main() {
	utils.Helper()
}
`)
	tree := parseGo(t, source)
	ext := NewGoExtractor()

	refs, err := ext.ExtractReferences(tree, source, filepath.Join(dir, "main.go"), dir)
	if err != nil {
		t.Fatalf("ExtractReferences error: %v", err)
	}

	var utilsImport *analysis.Reference
	for i, ref := range refs {
		if ref.Kind == "imports" && ref.ToName == "utils" {
			utilsImport = &refs[i]
			break
		}
	}

	if utilsImport == nil {
		t.Fatal("expected to find import reference for 'utils'")
	}
	if utilsImport.IsExternal {
		t.Error("internal import should have IsExternal=false")
	}
}

// TestGoExtractor_ExternalCallHasIsExternal verifies that a call to an
// external package function has IsExternal=true and ExternalImportPath set.
// Expected result: Call to fmt.Println has IsExternal=true, ExternalImportPath="fmt".
func TestGoExtractor_ExternalCallHasIsExternal(t *testing.T) {
	source := []byte(`package main

import "fmt"

func main() {
	fmt.Println("hello")
}
`)
	tree := parseGo(t, source)
	ext := NewGoExtractor()

	refs, err := ext.ExtractReferences(tree, source, "main.go", "")
	if err != nil {
		t.Fatalf("ExtractReferences error: %v", err)
	}

	var printlnCall *analysis.Reference
	for i, ref := range refs {
		if ref.Kind == "calls" && ref.ToName == "Println" {
			printlnCall = &refs[i]
			break
		}
	}

	if printlnCall == nil {
		t.Fatal("expected to find call reference to 'Println' (normalized from fmt.Println)")
	}
	if !printlnCall.IsExternal {
		t.Error("call to fmt.Println should have IsExternal=true")
	}
	if printlnCall.ExternalImportPath != "fmt" {
		t.Errorf("ExternalImportPath = %q, want %q", printlnCall.ExternalImportPath, "fmt")
	}
}

// TestGoExtractor_InternalCallHasIsExternalFalse verifies that a call to
// a repo-internal function has IsExternal=false.
// Expected result: Call to 'helper' has IsExternal=false.
func TestGoExtractor_InternalCallHasIsExternalFalse(t *testing.T) {
	source := []byte(`package main

func helper() {}

func main() {
	helper()
}
`)
	tree := parseGo(t, source)
	ext := NewGoExtractor()

	refs, err := ext.ExtractReferences(tree, source, "main.go", "")
	if err != nil {
		t.Fatalf("ExtractReferences error: %v", err)
	}

	var helperCall *analysis.Reference
	for i, ref := range refs {
		if ref.Kind == "calls" && ref.ToName == "helper" {
			helperCall = &refs[i]
			break
		}
	}

	if helperCall == nil {
		t.Fatal("expected to find call reference to 'helper'")
	}
	if helperCall.IsExternal {
		t.Error("call to internal function 'helper' should have IsExternal=false")
	}
}
