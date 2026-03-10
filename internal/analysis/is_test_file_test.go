package analysis

import (
	"testing"
)

// === Task 5: isTestFile Python/Ruby conventions ===

// TestIsTestFile_PythonTestPrefix verifies that Python test files with test_
// prefix are recognized.
// Expected result: isTestFile("test_models.py") returns true.
func TestIsTestFile_PythonTestPrefix(t *testing.T) {
	if !isTestFile("test_models.py") {
		t.Error("isTestFile(test_models.py) = false, want true")
	}
}

// TestIsTestFile_PythonTestSuffix verifies that Python test files with _test.py
// suffix are recognized.
// Expected result: isTestFile("models_test.py") returns true.
func TestIsTestFile_PythonTestSuffix(t *testing.T) {
	if !isTestFile("models_test.py") {
		t.Error("isTestFile(models_test.py) = false, want true")
	}
}

// TestIsTestFile_PythonTestDir verifies that files in test/ or tests/ dirs
// are recognized as tests.
// Expected result: isTestFile("tests/test_app.py") returns true.
func TestIsTestFile_PythonTestDir(t *testing.T) {
	tests := []struct {
		path string
		want bool
	}{
		{"tests/test_app.py", true},
		{"test/test_app.py", true},
		{"tests/conftest.py", true},
	}
	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			got := isTestFile(tt.path)
			if got != tt.want {
				t.Errorf("isTestFile(%q) = %v, want %v", tt.path, got, tt.want)
			}
		})
	}
}

// TestIsTestFile_RubySpecSuffix verifies that Ruby spec files with _spec.rb
// suffix are recognized.
// Expected result: isTestFile("user_spec.rb") returns true.
func TestIsTestFile_RubySpecSuffix(t *testing.T) {
	if !isTestFile("user_spec.rb") {
		t.Error("isTestFile(user_spec.rb) = false, want true")
	}
}

// TestIsTestFile_RubyTestSuffix verifies that Ruby test files with _test.rb
// suffix are recognized.
// Expected result: isTestFile("user_test.rb") returns true.
func TestIsTestFile_RubyTestSuffix(t *testing.T) {
	if !isTestFile("user_test.rb") {
		t.Error("isTestFile(user_test.rb) = false, want true")
	}
}

// TestIsTestFile_RubySpecDir verifies that files in spec/ dir are recognized.
// Expected result: isTestFile("spec/user_spec.rb") returns true.
func TestIsTestFile_RubySpecDir(t *testing.T) {
	tests := []struct {
		path string
		want bool
	}{
		{"spec/user_spec.rb", true},
		{"spec/models/user_spec.rb", true},
		{"test/user_test.rb", true},
	}
	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			got := isTestFile(tt.path)
			if got != tt.want {
				t.Errorf("isTestFile(%q) = %v, want %v", tt.path, got, tt.want)
			}
		})
	}
}

// TestIsTestFile_PythonNonTestFile verifies that regular Python files are
// not identified as test files.
// Expected result: isTestFile("models.py") returns false.
func TestIsTestFile_PythonNonTestFile(t *testing.T) {
	if isTestFile("models.py") {
		t.Error("isTestFile(models.py) = true, want false")
	}
}

// TestIsTestFile_RubyNonTestFile verifies that regular Ruby files are
// not identified as test files.
// Expected result: isTestFile("user.rb") returns false.
func TestIsTestFile_RubyNonTestFile(t *testing.T) {
	if isTestFile("user.rb") {
		t.Error("isTestFile(user.rb) = true, want false")
	}
}
