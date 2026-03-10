package typescript

import (
	"testing"

	"github.com/tab58/code-context/internal/analysis"
)

// === Task 6: TypeScript/TSX extractor reference enhancements ===

// TestTSExtractor_NonRelativeImportIsExternal verifies that non-relative
// imports (e.g., "react", "lodash") are marked as external references.
// Expected result: Import reference has IsExternal=true and ExternalImportPath set.
func TestTSExtractor_NonRelativeImportIsExternal(t *testing.T) {
	source := []byte(`import React from 'react';
import { merge } from 'lodash';

function App() {}
`)
	tree := parseTS(t, source)
	ext := NewTypeScriptExtractor()

	refs, err := ext.ExtractReferences(tree, source, "app.ts", "")
	if err != nil {
		t.Fatalf("ExtractReferences error: %v", err)
	}

	var reactImport *analysis.Reference
	for i, ref := range refs {
		if ref.Kind == "imports" && (ref.ToName == "react" || ref.ExternalImportPath == "react") {
			reactImport = &refs[i]
			break
		}
	}

	if reactImport == nil {
		t.Fatal("expected to find import reference for 'react'")
	}
	if !reactImport.IsExternal {
		t.Error("non-relative import 'react' should have IsExternal=true")
	}
	if reactImport.ExternalImportPath != "react" {
		t.Errorf("ExternalImportPath = %q, want %q", reactImport.ExternalImportPath, "react")
	}
}

// TestTSExtractor_RelativeImportIsInternal verifies that relative imports
// (starting with "." or "/") are marked as internal.
// Expected result: Import reference has IsExternal=false.
func TestTSExtractor_RelativeImportIsInternal(t *testing.T) {
	source := []byte(`import { helper } from './utils';

function main() {
  helper();
}
`)
	tree := parseTS(t, source)
	ext := NewTypeScriptExtractor()

	refs, err := ext.ExtractReferences(tree, source, "app.ts", "")
	if err != nil {
		t.Fatalf("ExtractReferences error: %v", err)
	}

	var utilsImport *analysis.Reference
	for i, ref := range refs {
		if ref.Kind == "imports" && ref.ToName == "./utils" {
			utilsImport = &refs[i]
			break
		}
	}

	if utilsImport == nil {
		t.Fatal("expected to find import reference for './utils'")
	}
	if utilsImport.IsExternal {
		t.Error("relative import './utils' should have IsExternal=false")
	}
}

// TestTSExtractor_NormalizesMemberExpression verifies that member expression
// call targets (e.g., console.log) are normalized to property name only.
// Expected result: Reference ToName is "log", not "console.log".
func TestTSExtractor_NormalizesMemberExpression(t *testing.T) {
	source := []byte(`function main(): void {
  console.log("hello");
}
`)
	tree := parseTS(t, source)
	ext := NewTypeScriptExtractor()

	refs, err := ext.ExtractReferences(tree, source, "app.ts", "")
	if err != nil {
		t.Fatalf("ExtractReferences error: %v", err)
	}

	for _, ref := range refs {
		if ref.Kind == "calls" && ref.ToName == "console.log" {
			t.Error("member_expression not normalized — should be filtered or bare property name, not 'console.log'")
		}
	}
	// console.log is normalized to "log" which is in jsBuiltins, so it gets filtered.
}

// TestTSExtractor_SkipsAnonymousFunctionCalls verifies that anonymous
// function calls are skipped.
// Expected result: No call reference for anonymous function expressions.
func TestTSExtractor_SkipsAnonymousFunctionCalls(t *testing.T) {
	source := []byte(`const result = (function() { return 42; })();
`)
	tree := parseTS(t, source)
	ext := NewTypeScriptExtractor()

	refs, err := ext.ExtractReferences(tree, source, "app.ts", "")
	if err != nil {
		t.Fatalf("ExtractReferences error: %v", err)
	}

	for _, ref := range refs {
		if ref.Kind == "calls" {
			t.Errorf("unexpected call reference %q — anonymous function calls should be skipped", ref.ToName)
		}
	}
}

// TestTSExtractor_SkipsArrowFunctionCalls verifies that arrow function
// call targets are skipped.
// Expected result: No call reference for arrow function expressions.
func TestTSExtractor_SkipsArrowFunctionCalls(t *testing.T) {
	source := []byte(`const result = (() => 42)();
`)
	tree := parseTS(t, source)
	ext := NewTypeScriptExtractor()

	refs, err := ext.ExtractReferences(tree, source, "app.ts", "")
	if err != nil {
		t.Fatalf("ExtractReferences error: %v", err)
	}

	for _, ref := range refs {
		if ref.Kind == "calls" {
			t.Errorf("unexpected call reference %q — arrow function calls should be skipped", ref.ToName)
		}
	}
}

// TestTSExtractor_ScopedPackageIsExternal verifies that scoped npm packages
// (e.g., @scope/pkg) are classified as external.
// Expected result: Import for '@angular/core' has IsExternal=true.
func TestTSExtractor_ScopedPackageIsExternal(t *testing.T) {
	source := []byte(`import { Component } from '@angular/core';

function main() {}
`)
	tree := parseTS(t, source)
	ext := NewTypeScriptExtractor()

	refs, err := ext.ExtractReferences(tree, source, "app.ts", "")
	if err != nil {
		t.Fatalf("ExtractReferences error: %v", err)
	}

	var angularImport *analysis.Reference
	for i, ref := range refs {
		if ref.Kind == "imports" && ref.ExternalImportPath == "@angular/core" {
			angularImport = &refs[i]
			break
		}
	}

	if angularImport == nil {
		t.Fatal("expected to find import reference for '@angular/core'")
	}
	if !angularImport.IsExternal {
		t.Error("scoped package '@angular/core' should have IsExternal=true")
	}
}

// TestTSXExtractor_NonRelativeImportIsExternal verifies that TSX extractor
// has the same external reference behavior as TypeScript.
// Expected result: Non-relative import has IsExternal=true.
func TestTSXExtractor_NonRelativeImportIsExternal(t *testing.T) {
	source := []byte(`import React from 'react';

export function App(): JSX.Element {
  return <div>Hello</div>;
}
`)
	tree := parseTSX(t, source)
	ext := NewTSXExtractor()

	refs, err := ext.ExtractReferences(tree, source, "App.tsx", "")
	if err != nil {
		t.Fatalf("ExtractReferences error: %v", err)
	}

	var reactImport *analysis.Reference
	for i, ref := range refs {
		if ref.Kind == "imports" && (ref.ToName == "react" || ref.ExternalImportPath == "react") {
			reactImport = &refs[i]
			break
		}
	}

	if reactImport == nil {
		t.Fatal("expected to find import reference for 'react' in TSX")
	}
	if !reactImport.IsExternal {
		t.Error("non-relative import 'react' in TSX should have IsExternal=true")
	}
}
