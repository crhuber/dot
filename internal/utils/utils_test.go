package utils

import (
	"bytes"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestExpandPath(t *testing.T) {
	// Save original HOME
	originalHome := os.Getenv("HOME")
	defer os.Setenv("HOME", originalHome)

	// Set a test HOME directory
	testHome := "/test/home"
	os.Setenv("HOME", testHome)

	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "Expand tilde only",
			input:    "~",
			expected: testHome,
		},
		{
			name:     "Expand tilde with path",
			input:    "~/.vimrc",
			expected: filepath.Join(testHome, ".vimrc"),
		},
		{
			name:     "Expand tilde with nested path",
			input:    "~/.config/nvim/init.vim",
			expected: filepath.Join(testHome, ".config/nvim/init.vim"),
		},
		{
			name:     "No expansion for absolute path",
			input:    "/absolute/path",
			expected: "/absolute/path",
		},
		{
			name:     "No expansion for relative path",
			input:    "relative/path",
			expected: "relative/path",
		},
		{
			name:     "No expansion for tilde in middle",
			input:    "/path/~/file",
			expected: "/path/~/file",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ExpandPath(tt.input)
			if result != tt.expected {
				t.Errorf("ExpandPath(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestExpandPathWithoutHome(t *testing.T) {
	// Temporarily unset HOME to test error handling
	originalHome := os.Getenv("HOME")
	os.Unsetenv("HOME")
	defer os.Setenv("HOME", originalHome)

	result := ExpandPath("~/test")
	expected := "~/test" // Should return unchanged when HOME is not available
	if result != expected {
		t.Errorf("ExpandPath with no HOME = %q, want %q", result, expected)
	}
}

func TestBackupFile(t *testing.T) {
	t.Run("Backup regular file", func(t *testing.T) {
		tempDir := t.TempDir()
		testFile := filepath.Join(tempDir, "test.txt")
		backupFile := testFile + ".bak"

		// Create test file
		content := "test content"
		if err := os.WriteFile(testFile, []byte(content), 0644); err != nil {
			t.Fatalf("Failed to create test file: %v", err)
		}

		// Backup the file
		err := BackupFile(testFile)
		if err != nil {
			t.Errorf("BackupFile failed: %v", err)
		}

		// Verify original file is gone
		if FileExists(testFile) {
			t.Error("Original file should not exist after backup")
		}

		// Verify backup file exists with correct content
		if !FileExists(backupFile) {
			t.Error("Backup file should exist")
		}

		backupContent, err := os.ReadFile(backupFile)
		if err != nil {
			t.Errorf("Failed to read backup file: %v", err)
		}

		if string(backupContent) != content {
			t.Errorf("Backup content = %q, want %q", string(backupContent), content)
		}
	})

	t.Run("Backup directory", func(t *testing.T) {
		tempDir := t.TempDir()
		testDir := filepath.Join(tempDir, "testdir")
		backupDir := testDir + ".bak"

		// Create test directory with a file
		os.MkdirAll(testDir, 0755)
		testFile := filepath.Join(testDir, "file.txt")
		content := "test content"
		if err := os.WriteFile(testFile, []byte(content), 0644); err != nil {
			t.Fatalf("Failed to create test file: %v", err)
		}

		// Backup the directory
		err := BackupFile(testDir)
		if err != nil {
			t.Errorf("BackupFile failed: %v", err)
		}

		// Verify original directory is gone
		if FileExists(testDir) {
			t.Error("Original directory should not exist after backup")
		}

		// Verify backup directory exists with correct content
		if !FileExists(backupDir) {
			t.Error("Backup directory should exist")
		}

		backupFile := filepath.Join(backupDir, "file.txt")
		if !FileExists(backupFile) {
			t.Error("File in backup directory should exist")
		}

		backupContent, err := os.ReadFile(backupFile)
		if err != nil {
			t.Errorf("Failed to read backup file: %v", err)
		}

		if string(backupContent) != content {
			t.Errorf("Backup content = %q, want %q", string(backupContent), content)
		}
	})

	t.Run("Overwrite existing backup", func(t *testing.T) {
		tempDir := t.TempDir()
		testFile := filepath.Join(tempDir, "test.txt")
		backupFile := testFile + ".bak"

		// Create test file
		newContent := "new content"
		if err := os.WriteFile(testFile, []byte(newContent), 0644); err != nil {
			t.Fatalf("Failed to create test file: %v", err)
		}

		// Create existing backup file
		oldContent := "old backup content"
		if err := os.WriteFile(backupFile, []byte(oldContent), 0644); err != nil {
			t.Fatalf("Failed to create existing backup: %v", err)
		}

		// Backup the file (should overwrite existing backup)
		err := BackupFile(testFile)
		if err != nil {
			t.Errorf("BackupFile failed: %v", err)
		}

		// Verify backup was overwritten with new content
		backupContent, err := os.ReadFile(backupFile)
		if err != nil {
			t.Errorf("Failed to read backup file: %v", err)
		}

		if string(backupContent) != newContent {
			t.Errorf("Backup content = %q, want %q", string(backupContent), newContent)
		}
	})

	t.Run("Fail to backup non-existent file", func(t *testing.T) {
		tempDir := t.TempDir()
		nonExistentFile := filepath.Join(tempDir, "nonexistent.txt")

		err := BackupFile(nonExistentFile)
		if err == nil {
			t.Error("Expected error when backing up non-existent file")
		}
	})
}

func TestIsSymlink(t *testing.T) {
	tempDir := t.TempDir()

	t.Run("Regular file is not symlink", func(t *testing.T) {
		testFile := filepath.Join(tempDir, "regular.txt")
		if err := os.WriteFile(testFile, []byte("content"), 0644); err != nil {
			t.Fatalf("Failed to create test file: %v", err)
		}

		isLink, err := IsSymlink(testFile)
		if err != nil {
			t.Errorf("IsSymlink failed: %v", err)
		}
		if isLink {
			t.Error("Regular file should not be identified as symlink")
		}
	})

	t.Run("Directory is not symlink", func(t *testing.T) {
		testDir := filepath.Join(tempDir, "testdir")
		if err := os.MkdirAll(testDir, 0755); err != nil {
			t.Fatalf("Failed to create test directory: %v", err)
		}

		isLink, err := IsSymlink(testDir)
		if err != nil {
			t.Errorf("IsSymlink failed: %v", err)
		}
		if isLink {
			t.Error("Directory should not be identified as symlink")
		}
	})

	t.Run("Symlink is identified correctly", func(t *testing.T) {
		testFile := filepath.Join(tempDir, "target.txt")
		testLink := filepath.Join(tempDir, "link.txt")

		// Create target file and symlink
		if err := os.WriteFile(testFile, []byte("content"), 0644); err != nil {
			t.Fatalf("Failed to create target file: %v", err)
		}
		if err := os.Symlink(testFile, testLink); err != nil {
			t.Fatalf("Failed to create symlink: %v", err)
		}

		isLink, err := IsSymlink(testLink)
		if err != nil {
			t.Errorf("IsSymlink failed: %v", err)
		}
		if !isLink {
			t.Error("Symlink should be identified as symlink")
		}
	})

	t.Run("Non-existent path returns error", func(t *testing.T) {
		nonExistentPath := filepath.Join(tempDir, "nonexistent")

		_, err := IsSymlink(nonExistentPath)
		if err == nil {
			t.Error("Expected error for non-existent path")
		}
	})
}

func TestReadSymlink(t *testing.T) {
	tempDir := t.TempDir()

	t.Run("Read valid symlink", func(t *testing.T) {
		targetFile := filepath.Join(tempDir, "target.txt")
		linkFile := filepath.Join(tempDir, "link.txt")

		// Create target file and symlink
		if err := os.WriteFile(targetFile, []byte("content"), 0644); err != nil {
			t.Fatalf("Failed to create target file: %v", err)
		}
		if err := os.Symlink(targetFile, linkFile); err != nil {
			t.Fatalf("Failed to create symlink: %v", err)
		}

		target, err := ReadSymlink(linkFile)
		if err != nil {
			t.Errorf("ReadSymlink failed: %v", err)
		}
		if target != targetFile {
			t.Errorf("ReadSymlink = %q, want %q", target, targetFile)
		}
	})

	t.Run("Read regular file returns error", func(t *testing.T) {
		regularFile := filepath.Join(tempDir, "regular.txt")
		if err := os.WriteFile(regularFile, []byte("content"), 0644); err != nil {
			t.Fatalf("Failed to create regular file: %v", err)
		}

		_, err := ReadSymlink(regularFile)
		if err == nil {
			t.Error("Expected error when reading regular file as symlink")
		}
		if !strings.Contains(err.Error(), "is not a symbolic link") {
			t.Errorf("Expected 'is not a symbolic link' error, got: %v", err)
		}
	})

	t.Run("Read non-existent path returns error", func(t *testing.T) {
		nonExistentPath := filepath.Join(tempDir, "nonexistent")

		_, err := ReadSymlink(nonExistentPath)
		if err == nil {
			t.Error("Expected error for non-existent path")
		}
	})
}

func TestFileExists(t *testing.T) {
	tempDir := t.TempDir()

	t.Run("Regular file exists", func(t *testing.T) {
		testFile := filepath.Join(tempDir, "exists.txt")
		if err := os.WriteFile(testFile, []byte("content"), 0644); err != nil {
			t.Fatalf("Failed to create test file: %v", err)
		}

		if !FileExists(testFile) {
			t.Error("FileExists should return true for existing file")
		}
	})

	t.Run("Directory exists", func(t *testing.T) {
		testDir := filepath.Join(tempDir, "testdir")
		if err := os.MkdirAll(testDir, 0755); err != nil {
			t.Fatalf("Failed to create test directory: %v", err)
		}

		if !FileExists(testDir) {
			t.Error("FileExists should return true for existing directory")
		}
	})

	t.Run("Symlink exists", func(t *testing.T) {
		targetFile := filepath.Join(tempDir, "target.txt")
		linkFile := filepath.Join(tempDir, "link.txt")

		// Create target file and symlink
		if err := os.WriteFile(targetFile, []byte("content"), 0644); err != nil {
			t.Fatalf("Failed to create target file: %v", err)
		}
		if err := os.Symlink(targetFile, linkFile); err != nil {
			t.Fatalf("Failed to create symlink: %v", err)
		}

		if !FileExists(linkFile) {
			t.Error("FileExists should return true for existing symlink")
		}
	})

	t.Run("Non-existent file", func(t *testing.T) {
		nonExistentFile := filepath.Join(tempDir, "nonexistent.txt")

		if FileExists(nonExistentFile) {
			t.Error("FileExists should return false for non-existent file")
		}
	})
}

func TestLogFunctions(t *testing.T) {
	t.Run("LogInfo outputs to stdout", func(t *testing.T) {
		// Capture stdout
		oldStdout := os.Stdout
		r, w, _ := os.Pipe()
		os.Stdout = w

		LogInfo("Test info message: %s", "value")

		w.Close()
		os.Stdout = oldStdout

		var buf bytes.Buffer
		io.Copy(&buf, r)
		output := buf.String()

		expected := "Test info message: value\n"
		if output != expected {
			t.Errorf("LogInfo output = %q, want %q", output, expected)
		}
	})

	t.Run("LogError outputs to stderr", func(t *testing.T) {
		// Capture stderr
		oldStderr := os.Stderr
		r, w, _ := os.Pipe()
		os.Stderr = w

		LogError("Test error message: %s", "value")

		w.Close()
		os.Stderr = oldStderr

		var buf bytes.Buffer
		io.Copy(&buf, r)
		output := buf.String()

		expected := "Test error message: value\n"
		if output != expected {
			t.Errorf("LogError output = %q, want %q", output, expected)
		}
	})

	t.Run("LogWarning outputs to stderr with prefix", func(t *testing.T) {
		// Capture stderr
		oldStderr := os.Stderr
		r, w, _ := os.Pipe()
		os.Stderr = w

		LogWarning("Test warning message: %s", "value")

		w.Close()
		os.Stderr = oldStderr

		var buf bytes.Buffer
		io.Copy(&buf, r)
		output := buf.String()

		expected := "Warning: Test warning message: value\n"
		if output != expected {
			t.Errorf("LogWarning output = %q, want %q", output, expected)
		}
	})
}
