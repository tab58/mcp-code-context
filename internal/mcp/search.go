package mcp

import (
	"path/filepath"
	"strings"
	"unicode"
)

// classifyQuery inspects the query string and returns the dispatch strategy.
// Rules applied in priority order (first match wins):
//  1. Contains glob chars (*, ?, **) -> file
//  2. Contains path separator (/) and file extension (.go, .ts, etc.) -> file
//  3. Single token, camelCase or PascalCase -> exact
//  4. Single token, snake_case -> exact
//  5. Everything else (multi-word, natural language) -> exact
func classifyQuery(query string) strategy {
	q := strings.TrimSpace(query)
	if q == "" {
		return strategyExact
	}

	// Rule 1: glob characters
	if strings.ContainsAny(q, "*?") {
		return strategyFile
	}

	// Rule 2: path with file extension
	if strings.Contains(q, "/") && filepath.Ext(q) != "" {
		return strategyFile
	}

	// Rule 5 check: multi-word queries use exact token-by-token search
	if strings.Contains(q, " ") {
		return strategyExact
	}

	// Single token — check for identifier patterns (Rules 3 & 4)

	// Rule 4: snake_case (contains underscore)
	if strings.Contains(q, "_") {
		return strategyExact
	}

	// Rule 3: camelCase or PascalCase (has mixed case transitions)
	if isCamelOrPascal(q) {
		return strategyExact
	}

	// Single word without case transitions — treat as exact (function/class name)
	return strategyExact
}

// isCamelOrPascal returns true if the string has at least one uppercase-to-lowercase
// or lowercase-to-uppercase transition, indicating camelCase or PascalCase.
func isCamelOrPascal(s string) bool {
	runes := []rune(s)
	for i := 1; i < len(runes); i++ {
		prev := runes[i-1]
		cur := runes[i]
		if (unicode.IsLower(prev) && unicode.IsUpper(cur)) ||
			(unicode.IsUpper(prev) && unicode.IsLower(cur)) {
			return true
		}
	}
	return false
}
