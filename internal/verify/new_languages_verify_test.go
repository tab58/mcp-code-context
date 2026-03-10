package verify

import (
	"os"
	"strings"
	"testing"
)

// === Tasks 1, 5, 6, 8, 9, 12: Structural verification tests ===

// --- Task 1: Registry grammars ---

// TestVerify_RegistryImportsJavaScriptGrammar verifies that registry.go imports
// the JavaScript tree-sitter grammar package.
// Expected result: registry.go contains "go-tree-sitter/javascript" import.
func TestVerify_RegistryImportsJavaScriptGrammar(t *testing.T) {
	content, err := os.ReadFile("../../internal/analysis/registry.go")
	if err != nil {
		t.Fatalf("failed to read registry.go: %v", err)
	}
	s := string(content)
	if !strings.Contains(s, `go-tree-sitter/javascript`) {
		t.Error("registry.go does not import go-tree-sitter/javascript")
	}
}

// TestVerify_RegistryImportsPythonGrammar verifies that registry.go imports
// the Python tree-sitter grammar package.
// Expected result: registry.go contains "go-tree-sitter/python" import.
func TestVerify_RegistryImportsPythonGrammar(t *testing.T) {
	content, err := os.ReadFile("../../internal/analysis/registry.go")
	if err != nil {
		t.Fatalf("failed to read registry.go: %v", err)
	}
	s := string(content)
	if !strings.Contains(s, `go-tree-sitter/python`) {
		t.Error("registry.go does not import go-tree-sitter/python")
	}
}

// TestVerify_RegistryImportsRubyGrammar verifies that registry.go imports
// the Ruby tree-sitter grammar package.
// Expected result: registry.go contains "go-tree-sitter/ruby" import.
func TestVerify_RegistryImportsRubyGrammar(t *testing.T) {
	content, err := os.ReadFile("../../internal/analysis/registry.go")
	if err != nil {
		t.Fatalf("failed to read registry.go: %v", err)
	}
	s := string(content)
	if !strings.Contains(s, `go-tree-sitter/ruby`) {
		t.Error("registry.go does not import go-tree-sitter/ruby")
	}
}

// TestVerify_RegistryRegistersJavaScriptInNewRegistry verifies that NewRegistry
// calls registerLanguage for JavaScript.
// Expected result: registry.go contains 'Name: "javascript"'.
func TestVerify_RegistryRegistersJavaScriptInNewRegistry(t *testing.T) {
	content, err := os.ReadFile("../../internal/analysis/registry.go")
	if err != nil {
		t.Fatalf("failed to read registry.go: %v", err)
	}
	s := string(content)
	if !strings.Contains(s, `"javascript"`) || !strings.Contains(s, `".js"`) {
		t.Error("registry.go does not register JavaScript language with .js extension")
	}
}

// TestVerify_RegistryRegistersPythonInNewRegistry verifies that NewRegistry
// calls registerLanguage for Python.
// Expected result: registry.go contains 'Name: "python"'.
func TestVerify_RegistryRegistersPythonInNewRegistry(t *testing.T) {
	content, err := os.ReadFile("../../internal/analysis/registry.go")
	if err != nil {
		t.Fatalf("failed to read registry.go: %v", err)
	}
	s := string(content)
	if !strings.Contains(s, `"python"`) || !strings.Contains(s, `".py"`) {
		t.Error("registry.go does not register Python language with .py extension")
	}
}

// TestVerify_RegistryRegistersRubyInNewRegistry verifies that NewRegistry
// calls registerLanguage for Ruby.
// Expected result: registry.go contains 'Name: "ruby"'.
func TestVerify_RegistryRegistersRubyInNewRegistry(t *testing.T) {
	content, err := os.ReadFile("../../internal/analysis/registry.go")
	if err != nil {
		t.Fatalf("failed to read registry.go: %v", err)
	}
	s := string(content)
	if !strings.Contains(s, `"ruby"`) || !strings.Contains(s, `".rb"`) {
		t.Error("registry.go does not register Ruby language with .rb extension")
	}
}

// --- Task 5: isTestFile Python/Ruby patterns ---

// TestVerify_IsTestFile_PythonTestPrefix verifies that isTestFile recognizes
// Python test files with test_ prefix.
// Expected result: graph_helpers.go contains 'test_' pattern.
func TestVerify_IsTestFile_PythonTestPrefix(t *testing.T) {
	content, err := os.ReadFile("../../internal/analysis/graph_helpers.go")
	if err != nil {
		t.Fatalf("failed to read graph_helpers.go: %v", err)
	}
	s := string(content)
	if !strings.Contains(s, `test_`) {
		t.Error("graph_helpers.go isTestFile does not contain Python test_ prefix pattern")
	}
}

// TestVerify_IsTestFile_PythonTestSuffix verifies that isTestFile recognizes
// Python test files with _test.py suffix.
// Expected result: graph_helpers.go contains '_test.py' pattern.
func TestVerify_IsTestFile_PythonTestSuffix(t *testing.T) {
	content, err := os.ReadFile("../../internal/analysis/graph_helpers.go")
	if err != nil {
		t.Fatalf("failed to read graph_helpers.go: %v", err)
	}
	s := string(content)
	if !strings.Contains(s, `_test.py`) {
		t.Error("graph_helpers.go isTestFile does not contain Python _test.py suffix pattern")
	}
}

// TestVerify_IsTestFile_RubySpecSuffix verifies that isTestFile recognizes
// Ruby spec files with _spec.rb suffix.
// Expected result: graph_helpers.go contains '_spec.rb' pattern.
func TestVerify_IsTestFile_RubySpecSuffix(t *testing.T) {
	content, err := os.ReadFile("../../internal/analysis/graph_helpers.go")
	if err != nil {
		t.Fatalf("failed to read graph_helpers.go: %v", err)
	}
	s := string(content)
	if !strings.Contains(s, `_spec.rb`) {
		t.Error("graph_helpers.go isTestFile does not contain Ruby _spec.rb suffix pattern")
	}
}

// TestVerify_IsTestFile_RubyTestSuffix verifies that isTestFile recognizes
// Ruby test files with _test.rb suffix.
// Expected result: graph_helpers.go contains '_test.rb' pattern.
func TestVerify_IsTestFile_RubyTestSuffix(t *testing.T) {
	content, err := os.ReadFile("../../internal/analysis/graph_helpers.go")
	if err != nil {
		t.Fatalf("failed to read graph_helpers.go: %v", err)
	}
	s := string(content)
	if !strings.Contains(s, `_test.rb`) {
		t.Error("graph_helpers.go isTestFile does not contain Ruby _test.rb suffix pattern")
	}
}

// --- Task 6: main.go registration ---

// TestVerify_MainImportsJSExtractor verifies that main.go imports the
// JavaScript extractor sub-package.
// Expected result: main.go contains "analysis/javascript" import.
func TestVerify_MainImportsJSExtractor(t *testing.T) {
	content, err := os.ReadFile("../../cmd/codectx/main.go")
	if err != nil {
		t.Fatalf("failed to read main.go: %v", err)
	}
	s := string(content)
	if !strings.Contains(s, `analysis/javascript`) {
		t.Error("main.go does not import analysis/javascript sub-package")
	}
}

// TestVerify_MainImportsPyExtractor verifies that main.go imports the
// Python extractor sub-package.
// Expected result: main.go contains "analysis/python" import.
func TestVerify_MainImportsPyExtractor(t *testing.T) {
	content, err := os.ReadFile("../../cmd/codectx/main.go")
	if err != nil {
		t.Fatalf("failed to read main.go: %v", err)
	}
	s := string(content)
	if !strings.Contains(s, `analysis/python`) {
		t.Error("main.go does not import analysis/python sub-package")
	}
}

// TestVerify_MainImportsRbExtractor verifies that main.go imports the
// Ruby extractor sub-package.
// Expected result: main.go contains "analysis/ruby" import.
func TestVerify_MainImportsRbExtractor(t *testing.T) {
	content, err := os.ReadFile("../../cmd/codectx/main.go")
	if err != nil {
		t.Fatalf("failed to read main.go: %v", err)
	}
	s := string(content)
	if !strings.Contains(s, `analysis/ruby`) {
		t.Error("main.go does not import analysis/ruby sub-package")
	}
}

// TestVerify_MainRegistersJSExtractor verifies that main.go calls
// jsextractor.Register(registry) or equivalent.
// Expected result: main.go contains JavaScript Register call.
func TestVerify_MainRegistersJSExtractor(t *testing.T) {
	content, err := os.ReadFile("../../cmd/codectx/main.go")
	if err != nil {
		t.Fatalf("failed to read main.go: %v", err)
	}
	s := string(content)
	// Look for any Register call for JS extractor (alias may vary)
	if !strings.Contains(s, "Register(registry)") || !strings.Contains(s, "javascript") {
		t.Error("main.go does not register JavaScript extractor with registry")
	}
}

// --- Task 7: call_chain.go existence ---

// TestVerify_CallChainFileExists verifies that call_chain.go exists in internal/mcp/.
// Expected result: File exists and is non-empty.
func TestVerify_CallChainFileExists(t *testing.T) {
	info, err := os.Stat("../../internal/mcp/call_chain.go")
	if err != nil {
		t.Fatalf("call_chain.go does not exist: %v", err)
	}
	if info.Size() == 0 {
		t.Error("call_chain.go is empty")
	}
}

// TestVerify_CallChainHasHandleFunction verifies that call_chain.go contains
// the handleFindCallChain method.
// Expected result: call_chain.go contains "handleFindCallChain".
func TestVerify_CallChainHasHandleFunction(t *testing.T) {
	content, err := os.ReadFile("../../internal/mcp/call_chain.go")
	if err != nil {
		t.Fatalf("failed to read call_chain.go: %v", err)
	}
	s := string(content)
	if !strings.Contains(s, "handleFindCallChain") {
		t.Error("call_chain.go does not contain handleFindCallChain method")
	}
}

// TestVerify_CallChainHasMaxDepthConstant verifies that call_chain.go defines
// maxCallChainDepth constant.
// Expected result: call_chain.go contains "maxCallChainDepth".
func TestVerify_CallChainHasMaxDepthConstant(t *testing.T) {
	content, err := os.ReadFile("../../internal/mcp/call_chain.go")
	if err != nil {
		t.Fatalf("failed to read call_chain.go: %v", err)
	}
	s := string(content)
	if !strings.Contains(s, "maxCallChainDepth") {
		t.Error("call_chain.go does not define maxCallChainDepth constant")
	}
}

// --- Task 8: server.go 20 tools ---

// TestVerify_ServerRegisters20Tools verifies that server.go registers 20 tools.
// Expected result: server.go contains "20 tool handlers" in doc comment.
func TestVerify_ServerRegisters20Tools(t *testing.T) {
	content, err := os.ReadFile("../../internal/mcp/server.go")
	if err != nil {
		t.Fatalf("failed to read server.go: %v", err)
	}
	s := string(content)
	if !strings.Contains(s, "20 tool") {
		t.Error("server.go doc comment does not reference 20 tools")
	}
}

// TestVerify_ServerHasFindCallChainTool verifies that server.go registers
// the find_call_chain tool.
// Expected result: server.go contains 'find_call_chain' tool registration.
func TestVerify_ServerHasFindCallChainTool(t *testing.T) {
	content, err := os.ReadFile("../../internal/mcp/server.go")
	if err != nil {
		t.Fatalf("failed to read server.go: %v", err)
	}
	s := string(content)
	if !strings.Contains(s, `"find_call_chain"`) {
		t.Error("server.go does not register find_call_chain tool")
	}
}

// TestVerify_ServerHasMCPHandleFindCallChain verifies that server.go has
// the mcpHandleFindCallChain adapter method.
// Expected result: server.go contains "mcpHandleFindCallChain".
func TestVerify_ServerHasMCPHandleFindCallChain(t *testing.T) {
	content, err := os.ReadFile("../../internal/mcp/server.go")
	if err != nil {
		t.Fatalf("failed to read server.go: %v", err)
	}
	s := string(content)
	if !strings.Contains(s, "mcpHandleFindCallChain") {
		t.Error("server.go does not contain mcpHandleFindCallChain adapter")
	}
}

// --- Task 9: GraphQL constants for call chain ---

// TestVerify_CallChainHasCalleesConstant verifies that call_chain.go contains
// a GraphQL constant for finding callees.
// Expected result: call_chain.go contains "gqlCallees" or reuses gqlFindCallees.
func TestVerify_CallChainHasCalleesConstant(t *testing.T) {
	content, err := os.ReadFile("../../internal/mcp/call_chain.go")
	if err != nil {
		t.Fatalf("failed to read call_chain.go: %v", err)
	}
	s := string(content)
	// Either defines own constant or uses existing traversal constant
	if !strings.Contains(s, "gqlCallChainCallees") && !strings.Contains(s, "gqlFindCallees") {
		t.Error("call_chain.go does not contain callees GraphQL constant or reference")
	}
}

// TestVerify_CallChainHasCallersConstant verifies that call_chain.go contains
// a GraphQL constant for finding callers.
// Expected result: call_chain.go contains "gqlCallChainCallers" or reuses gqlFindCallers.
func TestVerify_CallChainHasCallersConstant(t *testing.T) {
	content, err := os.ReadFile("../../internal/mcp/call_chain.go")
	if err != nil {
		t.Fatalf("failed to read call_chain.go: %v", err)
	}
	s := string(content)
	if !strings.Contains(s, "gqlCallChainCallers") && !strings.Contains(s, "gqlFindCallers") {
		t.Error("call_chain.go does not contain callers GraphQL constant or reference")
	}
}

// --- Task 12: Build verification ---

// TestVerify_JSExtractorFileExists verifies that the JavaScript extractor file exists.
// Expected result: internal/analysis/javascript/extractor.go exists.
func TestVerify_JSExtractorFileExists(t *testing.T) {
	_, err := os.Stat("../../internal/analysis/javascript/extractor.go")
	if err != nil {
		t.Fatalf("javascript/extractor.go does not exist: %v", err)
	}
}

// TestVerify_PythonExtractorFileExists verifies that the Python extractor file exists.
// Expected result: internal/analysis/python/extractor.go exists.
func TestVerify_PythonExtractorFileExists(t *testing.T) {
	_, err := os.Stat("../../internal/analysis/python/extractor.go")
	if err != nil {
		t.Fatalf("python/extractor.go does not exist: %v", err)
	}
}

// TestVerify_RubyExtractorFileExists verifies that the Ruby extractor file exists.
// Expected result: internal/analysis/ruby/extractor.go exists.
func TestVerify_RubyExtractorFileExists(t *testing.T) {
	_, err := os.Stat("../../internal/analysis/ruby/extractor.go")
	if err != nil {
		t.Fatalf("ruby/extractor.go does not exist: %v", err)
	}
}

// TestVerify_CallChainResponseTypeExists verifies that types.go defines
// CallChainResponse.
// Expected result: types.go contains "CallChainResponse".
func TestVerify_CallChainResponseTypeExists(t *testing.T) {
	content, err := os.ReadFile("../../internal/mcp/types.go")
	if err != nil {
		t.Fatalf("failed to read types.go: %v", err)
	}
	s := string(content)
	if !strings.Contains(s, "CallChainResponse") {
		t.Error("types.go does not define CallChainResponse type")
	}
}
