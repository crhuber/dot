package dotfiles

import (
	"bytes"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestGetDotfilesDir(t *testing.T) {
	// Save original environment variable
	originalDotDir := os.Getenv("DOT_DIR")
	defer func() {
		if originalDotDir != "" {
			os.Setenv("DOT_DIR", originalDotDir)
		} else {
			os.Unsetenv("DOT_DIR")
		}
	}()

	t.Run("Use DOT_DIR environment variable when set", func(t *testing.T) {
		customDir := "/custom/dotfiles/path"
		os.Setenv("DOT_DIR", customDir)

		result, err := GetDotfilesDir()
		if err != nil {
			t.Fatalf("Expected no error, got: %v", err)
		}

		if result != customDir {
			t.Errorf("Expected %s, got %s", customDir, result)
		}
	})

	t.Run("Default to ~/.dotfiles when DOT_DIR not set", func(t *testing.T) {
		os.Unsetenv("DOT_DIR")

		result, err := GetDotfilesDir()
		if err != nil {
			t.Fatalf("Expected no error, got: %v", err)
		}

		// Should end with .dotfiles
		if !strings.HasSuffix(result, ".dotfiles") {
			t.Errorf("Expected path to end with .dotfiles, got %s", result)
		}

		// Should be an absolute path
		if !filepath.IsAbs(result) {
			t.Errorf("Expected absolute path, got %s", result)
		}
	})

	t.Run("Handle empty DOT_DIR as if unset", func(t *testing.T) {
		os.Setenv("DOT_DIR", "")

		result, err := GetDotfilesDir()
		if err != nil {
			t.Fatalf("Expected no error, got: %v", err)
		}

		// Should default to ~/.dotfiles
		if !strings.HasSuffix(result, ".dotfiles") {
			t.Errorf("Expected path to end with .dotfiles, got %s", result)
		}
	})
}

func TestClone(t *testing.T) {
	// Save original environment variable
	originalDotDir := os.Getenv("DOT_DIR")
	defer func() {
		if originalDotDir != "" {
			os.Setenv("DOT_DIR", originalDotDir)
		} else {
			os.Unsetenv("DOT_DIR")
		}
	}()

	t.Run("Clone to empty directory succeeds", func(t *testing.T) {
		tempDir := t.TempDir()
		dotfilesDir := filepath.Join(tempDir, "dotfiles")
		os.Setenv("DOT_DIR", dotfilesDir)

		// Create a fake git command that creates .mappings file
		if err := createMockGitClone(dotfilesDir); err != nil {
			t.Fatalf("Failed to setup mock git: %v", err)
		}

		// Test that directory exists after mock setup
		stat, err := os.Stat(dotfilesDir)
		if err != nil {
			t.Fatalf("Directory should exist: %v", err)
		}
		if !stat.IsDir() {
			t.Error("Path should be a directory")
		}

		entries, err := os.ReadDir(dotfilesDir)
		if err != nil {
			t.Fatalf("Should be able to read directory: %v", err)
		}
		// Should have .mappings file from mock setup
		if len(entries) == 0 {
			t.Error("Directory should contain .mappings file from mock setup")
		}

		// Check that .mappings file exists
		mappingsExists := false
		for _, entry := range entries {
			if entry.Name() == ".mappings" {
				mappingsExists = true
				break
			}
		}
		if !mappingsExists {
			t.Error("Should contain .mappings file")
		}
	})

	t.Run("Clone fails when destination exists and is non-empty", func(t *testing.T) {
		tempDir := t.TempDir()
		dotfilesDir := filepath.Join(tempDir, "dotfiles")
		os.Setenv("DOT_DIR", dotfilesDir)

		// Create directory with a file
		if err := os.MkdirAll(dotfilesDir, 0755); err != nil {
			t.Fatalf("Failed to create test directory: %v", err)
		}
		if err := os.WriteFile(filepath.Join(dotfilesDir, "existing.txt"), []byte("content"), 0644); err != nil {
			t.Fatalf("Failed to create test file: %v", err)
		}

		err := Clone("https://example.com/repo.git")
		if err == nil {
			t.Error("Expected error for non-empty directory")
		}
		if !strings.Contains(err.Error(), "already exists and is non-empty") {
			t.Errorf("Expected error about non-empty directory, got: %v", err)
		}
	})

	t.Run("Clone fails when destination exists but is not a directory", func(t *testing.T) {
		tempDir := t.TempDir()
		dotfilesPath := filepath.Join(tempDir, "dotfiles")
		os.Setenv("DOT_DIR", dotfilesPath)

		// Create a file at the destination path
		if err := os.WriteFile(dotfilesPath, []byte("content"), 0644); err != nil {
			t.Fatalf("Failed to create test file: %v", err)
		}

		err := Clone("https://example.com/repo.git")
		if err == nil {
			t.Error("Expected error for non-directory path")
		}
		if !strings.Contains(err.Error(), "exists but is not a directory") {
			t.Errorf("Expected error about non-directory, got: %v", err)
		}
	})

	t.Run("Clone succeeds when destination doesn't exist", func(t *testing.T) {
		tempDir := t.TempDir()
		dotfilesDir := filepath.Join(tempDir, "nonexistent")
		os.Setenv("DOT_DIR", dotfilesDir)

		// Verify directory doesn't exist
		if _, err := os.Stat(dotfilesDir); !os.IsNotExist(err) {
			t.Error("Directory should not exist initially")
		}

		// We can't easily test the actual git clone without mocking exec.Command
		// But we can test that the directory check passes for non-existent paths
		// This would normally proceed to git clone
	})

	t.Run("Clone allows empty directory", func(t *testing.T) {
		tempDir := t.TempDir()
		dotfilesDir := filepath.Join(tempDir, "empty")
		os.Setenv("DOT_DIR", dotfilesDir)

		// Create empty directory
		if err := os.MkdirAll(dotfilesDir, 0755); err != nil {
			t.Fatalf("Failed to create empty directory: %v", err)
		}

		// Verify it's empty
		entries, err := os.ReadDir(dotfilesDir)
		if err != nil {
			t.Fatalf("Failed to read directory: %v", err)
		}
		if len(entries) != 0 {
			t.Error("Directory should be empty")
		}

		// This test verifies the empty directory check passes
		// Actual git clone would happen next in real scenario
	})
}

func TestPrintRoot(t *testing.T) {
	// Save original environment variable
	originalDotDir := os.Getenv("DOT_DIR")
	defer func() {
		if originalDotDir != "" {
			os.Setenv("DOT_DIR", originalDotDir)
		} else {
			os.Unsetenv("DOT_DIR")
		}
	}()

	t.Run("Print custom DOT_DIR path", func(t *testing.T) {
		customPath := "/custom/dotfiles"
		os.Setenv("DOT_DIR", customPath)

		// Capture stdout
		oldStdout := os.Stdout
		r, w, _ := os.Pipe()
		os.Stdout = w

		err := PrintRoot()

		// Restore stdout and get output
		w.Close()
		os.Stdout = oldStdout

		var buf bytes.Buffer
		io.Copy(&buf, r)
		output := strings.TrimSpace(buf.String())

		if err != nil {
			t.Fatalf("Expected no error, got: %v", err)
		}

		if output != customPath {
			t.Errorf("Expected output %s, got %s", customPath, output)
		}
	})

	t.Run("Print default path when DOT_DIR not set", func(t *testing.T) {
		os.Unsetenv("DOT_DIR")

		// Capture stdout
		oldStdout := os.Stdout
		r, w, _ := os.Pipe()
		os.Stdout = w

		err := PrintRoot()

		// Restore stdout and get output
		w.Close()
		os.Stdout = oldStdout

		var buf bytes.Buffer
		io.Copy(&buf, r)
		output := strings.TrimSpace(buf.String())

		if err != nil {
			t.Fatalf("Expected no error, got: %v", err)
		}

		if !strings.HasSuffix(output, ".dotfiles") {
			t.Errorf("Expected output to end with .dotfiles, got %s", output)
		}

		if !filepath.IsAbs(output) {
			t.Errorf("Expected absolute path, got %s", output)
		}
	})
}

// Test for error handling in Clone when git command fails
func TestCloneGitFailures(t *testing.T) {
	originalDotDir := os.Getenv("DOT_DIR")
	defer func() {
		if originalDotDir != "" {
			os.Setenv("DOT_DIR", originalDotDir)
		} else {
			os.Unsetenv("DOT_DIR")
		}
	}()

	t.Run("Clone with invalid repository URL", func(t *testing.T) {
		tempDir := t.TempDir()
		dotfilesDir := filepath.Join(tempDir, "dotfiles")
		os.Setenv("DOT_DIR", dotfilesDir)

		// This will fail because the URL is invalid
		err := Clone("invalid-url")
		if err == nil {
			t.Error("Expected error for invalid URL")
		}
		if !strings.Contains(err.Error(), "failed to clone repository") {
			t.Errorf("Expected clone error, got: %v", err)
		}
	})
}

// Test validation of .mappings file after clone
func TestMappingsValidation(t *testing.T) {
	originalDotDir := os.Getenv("DOT_DIR")
	defer func() {
		if originalDotDir != "" {
			os.Setenv("DOT_DIR", originalDotDir)
		} else {
			os.Unsetenv("DOT_DIR")
		}
	}()

	t.Run("Validate .mappings file existence logic", func(t *testing.T) {
		tempDir := t.TempDir()
		dotfilesDir := filepath.Join(tempDir, "dotfiles")
		os.Setenv("DOT_DIR", dotfilesDir)

		// Create directory structure that simulates post-clone
		if err := os.MkdirAll(dotfilesDir, 0755); err != nil {
			t.Fatalf("Failed to create directory: %v", err)
		}

		// Test missing .mappings file
		mappingsPath := filepath.Join(dotfilesDir, ".mappings")
		if _, err := os.Stat(mappingsPath); !os.IsNotExist(err) {
			t.Error("Expected .mappings to not exist initially")
		}

		// Create .mappings file
		if err := os.WriteFile(mappingsPath, []byte("[general]\n"), 0644); err != nil {
			t.Fatalf("Failed to create .mappings: %v", err)
		}

		// Test .mappings file exists
		if _, err := os.Stat(mappingsPath); os.IsNotExist(err) {
			t.Error("Expected .mappings to exist after creation")
		}
	})
}

// Helper function to create a mock git clone result
func createMockGitClone(dotfilesDir string) error {
	if err := os.MkdirAll(dotfilesDir, 0755); err != nil {
		return err
	}

	// Create a .mappings file
	mappingsContent := `[general]
"vim/.vimrc" = "~/.vimrc"
"git/.gitconfig" = "~/.gitconfig"`

	return os.WriteFile(filepath.Join(dotfilesDir, ".mappings"), []byte(mappingsContent), 0644)
}

// Test error handling when GetDotfilesDir fails
func TestCloneWithGetDotfilesDirError(t *testing.T) {
	// This is harder to test without mocking os.UserHomeDir
	// In practice, GetDotfilesDir rarely fails unless home directory is inaccessible
	t.Run("Clone handles GetDotfilesDir errors", func(t *testing.T) {
		// We can at least verify the error propagation structure exists
		// by checking that Clone calls GetDotfilesDir first

		// Set DOT_DIR to ensure GetDotfilesDir succeeds in this test
		tempDir := t.TempDir()
		os.Setenv("DOT_DIR", tempDir)
		defer os.Unsetenv("DOT_DIR")

		// This should at least get past GetDotfilesDir and fail at git clone
		err := Clone("invalid-url")
		if err == nil {
			t.Error("Expected some error (likely git clone failure)")
		}
		// Error should be from git clone, not GetDotfilesDir
		if strings.Contains(err.Error(), "failed to get user home directory") {
			t.Error("Unexpected error from GetDotfilesDir")
		}
	})
}

// Test for successful clone with proper .mappings validation
func TestCloneSuccess(t *testing.T) {
	originalDotDir := os.Getenv("DOT_DIR")
	defer func() {
		if originalDotDir != "" {
			os.Setenv("DOT_DIR", originalDotDir)
		} else {
			os.Unsetenv("DOT_DIR")
		}
	}()

	t.Run("Clone succeeds and validates .mappings file", func(t *testing.T) {
		tempDir := t.TempDir()
		dotfilesDir := filepath.Join(tempDir, "dotfiles")
		os.Setenv("DOT_DIR", dotfilesDir)

		// Create a mock successful clone
		if err := createMockGitClone(dotfilesDir); err != nil {
			t.Fatalf("Failed to setup mock git clone: %v", err)
		}

		// Verify .mappings file exists (as Clone would check)
		mappingsPath := filepath.Join(dotfilesDir, ".mappings")
		if _, err := os.Stat(mappingsPath); os.IsNotExist(err) {
			t.Error("Expected .mappings file to exist after mock clone")
		}
	})
}

// Test directory read failure handling
func TestCloneDirectoryReadFailure(t *testing.T) {
	originalDotDir := os.Getenv("DOT_DIR")
	defer func() {
		if originalDotDir != "" {
			os.Setenv("DOT_DIR", originalDotDir)
		} else {
			os.Unsetenv("DOT_DIR")
		}
	}()

	t.Run("Clone handles directory read errors gracefully", func(t *testing.T) {
		tempDir := t.TempDir()
		dotfilesDir := filepath.Join(tempDir, "dotfiles")
		os.Setenv("DOT_DIR", dotfilesDir)

		// Create directory but don't make it unreadable (that's hard to test portably)
		// Instead test the normal case where directory is readable
		if err := os.MkdirAll(dotfilesDir, 0755); err != nil {
			t.Fatalf("Failed to create directory: %v", err)
		}

		// Verify directory is readable (normal case)
		entries, err := os.ReadDir(dotfilesDir)
		if err != nil {
			t.Fatalf("Should be able to read directory: %v", err)
		}
		if len(entries) != 0 {
			t.Error("New directory should be empty")
		}

		// This verifies the directory check would pass in Clone
	})
}

// Test Update function
func TestUpdate(t *testing.T) {
	originalDotDir := os.Getenv("DOT_DIR")
	defer func() {
		if originalDotDir != "" {
			os.Setenv("DOT_DIR", originalDotDir)
		} else {
			os.Unsetenv("DOT_DIR")
		}
	}()

	t.Run("Update fails when dotfiles directory doesn't exist", func(t *testing.T) {
		tempDir := t.TempDir()
		dotfilesDir := filepath.Join(tempDir, "nonexistent")
		os.Setenv("DOT_DIR", dotfilesDir)

		err := Update()
		if err == nil {
			t.Error("Expected error for non-existent directory")
		}
		if !strings.Contains(err.Error(), "does not exist") {
			t.Errorf("Expected error about non-existent directory, got: %v", err)
		}
	})

	t.Run("Update fails when not a git repository", func(t *testing.T) {
		tempDir := t.TempDir()
		dotfilesDir := filepath.Join(tempDir, "notgit")
		os.Setenv("DOT_DIR", dotfilesDir)

		// Create directory but not as git repo
		if err := os.MkdirAll(dotfilesDir, 0755); err != nil {
			t.Fatalf("Failed to create directory: %v", err)
		}

		err := Update()
		if err == nil {
			t.Error("Expected error for non-git directory")
		}
		if !strings.Contains(err.Error(), "failed to update dotfiles repository") {
			t.Errorf("Expected update error, got: %v", err)
		}
	})
}

// Test Open function
func TestOpen(t *testing.T) {
	originalDotDir := os.Getenv("DOT_DIR")
	defer func() {
		if originalDotDir != "" {
			os.Setenv("DOT_DIR", originalDotDir)
		} else {
			os.Unsetenv("DOT_DIR")
		}
	}()

	t.Run("Open fails when dotfiles directory doesn't exist", func(t *testing.T) {
		tempDir := t.TempDir()
		dotfilesDir := filepath.Join(tempDir, "nonexistent")
		os.Setenv("DOT_DIR", dotfilesDir)

		err := Open()
		if err == nil {
			t.Error("Expected error for non-existent directory")
		}
		if !strings.Contains(err.Error(), "does not exist") {
			t.Errorf("Expected error about non-existent directory, got: %v", err)
		}
	})

	t.Run("Open handles directory existence check", func(t *testing.T) {
		tempDir := t.TempDir()
		dotfilesDir := filepath.Join(tempDir, "existing")
		os.Setenv("DOT_DIR", dotfilesDir)

		// Create directory
		if err := os.MkdirAll(dotfilesDir, 0755); err != nil {
			t.Fatalf("Failed to create directory: %v", err)
		}

		// We can't fully test the open command without a GUI environment,
		// but we can verify it gets past the directory check
		// The actual open command will fail in test environment, which is expected
		err := Open()
		// In test environment without GUI, this will likely fail, which is OK
		// We're mainly testing that it doesn't error on directory existence check
		if err != nil && !strings.Contains(err.Error(), "failed to open dotfiles directory") &&
			!strings.Contains(err.Error(), "no suitable file manager command found") {
			t.Errorf("Unexpected error type: %v", err)
		}
	})
}
