package linker

import (
	"bytes"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestParseProfiles(t *testing.T) {
	t.Run("Default to general when empty", func(t *testing.T) {
		result := ParseProfiles("")
		expected := []string{"general"}

		if len(result) != len(expected) {
			t.Errorf("Expected %v, got %v", expected, result)
		}
		if result[0] != expected[0] {
			t.Errorf("Expected %s, got %s", expected[0], result[0])
		}
	})

	t.Run("Parse single profile", func(t *testing.T) {
		result := ParseProfiles("work")
		expected := []string{"work"}

		if len(result) != len(expected) {
			t.Errorf("Expected %v, got %v", expected, result)
		}
		if result[0] != expected[0] {
			t.Errorf("Expected %s, got %s", expected[0], result[0])
		}
	})

	t.Run("Parse comma-separated profiles", func(t *testing.T) {
		result := ParseProfiles("general,work,minimal")
		expected := []string{"general", "work", "minimal"}

		if len(result) != len(expected) {
			t.Errorf("Expected %v, got %v", expected, result)
		}
		for i, exp := range expected {
			if result[i] != exp {
				t.Errorf("Expected %s at index %d, got %s", exp, i, result[i])
			}
		}
	})

	t.Run("Trim whitespace from profiles", func(t *testing.T) {
		result := ParseProfiles("  general  ,  work  ,  minimal  ")
		expected := []string{"general", "work", "minimal"}

		if len(result) != len(expected) {
			t.Errorf("Expected %v, got %v", expected, result)
		}
		for i, exp := range expected {
			if result[i] != exp {
				t.Errorf("Expected %s at index %d, got %s", exp, i, result[i])
			}
		}
	})
}

func TestCheck(t *testing.T) {
	// Save original DOT_DIR
	originalDotDir := os.Getenv("DOT_DIR")
	defer func() {
		if originalDotDir != "" {
			os.Setenv("DOT_DIR", originalDotDir)
		} else {
			os.Unsetenv("DOT_DIR")
		}
	}()

	t.Run("All links correct", func(t *testing.T) {
		tempDir := t.TempDir()
		dotfilesDir := filepath.Join(tempDir, "dotfiles")
		homeDir := filepath.Join(tempDir, "home")
		os.Setenv("DOT_DIR", dotfilesDir)

		// Setup test environment
		setupTestEnvironment(t, dotfilesDir, homeDir)

		// Create correct symlinks
		sourcePath := filepath.Join(dotfilesDir, "vim/.vimrc")
		targetPath := filepath.Join(homeDir, ".vimrc")
		if err := os.Symlink(sourcePath, targetPath); err != nil {
			t.Fatalf("Failed to create test symlink: %v", err)
		}

		// Capture output
		oldStdout := os.Stdout
		oldStderr := os.Stderr
		r, w, _ := os.Pipe()
		os.Stdout = w
		os.Stderr = w

		err := Check([]string{"general"})

		w.Close()
		os.Stdout = oldStdout
		os.Stderr = oldStderr

		var buf bytes.Buffer
		io.Copy(&buf, r)
		output := buf.String()

		if err != nil {
			t.Errorf("Expected no error, got: %v", err)
		}
		if !strings.Contains(output, "All links are correct") {
			t.Errorf("Expected success message, got: %s", output)
		}
	})

	t.Run("Missing symlinks", func(t *testing.T) {
		tempDir := t.TempDir()
		dotfilesDir := filepath.Join(tempDir, "dotfiles")
		homeDir := filepath.Join(tempDir, "home")
		os.Setenv("DOT_DIR", dotfilesDir)

		// Setup test environment but don't create symlinks
		setupTestEnvironment(t, dotfilesDir, homeDir)

		// Capture stderr
		oldStderr := os.Stderr
		r, w, _ := os.Pipe()
		os.Stderr = w

		err := Check([]string{"general"})

		w.Close()
		os.Stderr = oldStderr

		var buf bytes.Buffer
		io.Copy(&buf, r)
		output := buf.String()

		if err == nil {
			t.Error("Expected error for missing links")
		}
		if !strings.Contains(output, "Missing link:") {
			t.Errorf("Expected missing link message, got: %s", output)
		}
	})

	t.Run("Incorrect symlinks", func(t *testing.T) {
		tempDir := t.TempDir()
		dotfilesDir := filepath.Join(tempDir, "dotfiles")
		homeDir := filepath.Join(tempDir, "home")
		os.Setenv("DOT_DIR", dotfilesDir)

		// Setup test environment
		setupTestEnvironment(t, dotfilesDir, homeDir)

		// Create incorrect symlink
		wrongSource := filepath.Join(tempDir, "wrong-target")
		targetPath := filepath.Join(homeDir, ".vimrc")
		if err := os.WriteFile(wrongSource, []byte("wrong"), 0644); err != nil {
			t.Fatalf("Failed to create wrong source: %v", err)
		}
		if err := os.Symlink(wrongSource, targetPath); err != nil {
			t.Fatalf("Failed to create incorrect symlink: %v", err)
		}

		// Capture stderr
		oldStderr := os.Stderr
		r, w, _ := os.Pipe()
		os.Stderr = w

		err := Check([]string{"general"})

		w.Close()
		os.Stderr = oldStderr

		var buf bytes.Buffer
		io.Copy(&buf, r)
		output := buf.String()

		if err == nil {
			t.Error("Expected error for incorrect links")
		}
		if !strings.Contains(output, "Incorrect link:") {
			t.Errorf("Expected incorrect link message, got: %s", output)
		}
	})

	t.Run("Non-symlink files at target paths", func(t *testing.T) {
		tempDir := t.TempDir()
		dotfilesDir := filepath.Join(tempDir, "dotfiles")
		homeDir := filepath.Join(tempDir, "home")
		os.Setenv("DOT_DIR", dotfilesDir)

		// Setup test environment
		setupTestEnvironment(t, dotfilesDir, homeDir)

		// Create regular file at target path
		targetPath := filepath.Join(homeDir, ".vimrc")
		if err := os.WriteFile(targetPath, []byte("content"), 0644); err != nil {
			t.Fatalf("Failed to create regular file: %v", err)
		}

		// Capture stderr
		oldStderr := os.Stderr
		r, w, _ := os.Pipe()
		os.Stderr = w

		err := Check([]string{"general"})

		w.Close()
		os.Stderr = oldStderr

		var buf bytes.Buffer
		io.Copy(&buf, r)
		output := buf.String()

		if err == nil {
			t.Error("Expected error for non-symlink files")
		}
		if !strings.Contains(output, "Not a symlink:") {
			t.Errorf("Expected not a symlink message, got: %s", output)
		}
	})
}

func TestClean(t *testing.T) {
	// Save original DOT_DIR
	originalDotDir := os.Getenv("DOT_DIR")
	defer func() {
		if originalDotDir != "" {
			os.Setenv("DOT_DIR", originalDotDir)
		} else {
			os.Unsetenv("DOT_DIR")
		}
	}()

	t.Run("Remove valid symlinks", func(t *testing.T) {
		tempDir := t.TempDir()
		dotfilesDir := filepath.Join(tempDir, "dotfiles")
		homeDir := filepath.Join(tempDir, "home")
		os.Setenv("DOT_DIR", dotfilesDir)

		// Setup test environment
		setupTestEnvironment(t, dotfilesDir, homeDir)

		// Create symlink to remove
		sourcePath := filepath.Join(dotfilesDir, "vim/.vimrc")
		targetPath := filepath.Join(homeDir, ".vimrc")
		if err := os.Symlink(sourcePath, targetPath); err != nil {
			t.Fatalf("Failed to create test symlink: %v", err)
		}

		// Capture output
		oldStdout := os.Stdout
		r, w, _ := os.Pipe()
		os.Stdout = w

		err := Clean([]string{"general"})

		w.Close()
		os.Stdout = oldStdout

		var buf bytes.Buffer
		io.Copy(&buf, r)
		output := buf.String()

		if err != nil {
			t.Errorf("Expected no error, got: %v", err)
		}
		if !strings.Contains(output, "Removed:") {
			t.Errorf("Expected removed message, got: %s", output)
		}

		// Verify symlink was removed
		if _, err := os.Lstat(targetPath); !os.IsNotExist(err) {
			t.Error("Expected symlink to be removed")
		}
	})

	t.Run("Skip non-existent targets", func(t *testing.T) {
		tempDir := t.TempDir()
		dotfilesDir := filepath.Join(tempDir, "dotfiles")
		homeDir := filepath.Join(tempDir, "home")
		os.Setenv("DOT_DIR", dotfilesDir)

		// Setup test environment but don't create symlinks
		setupTestEnvironment(t, dotfilesDir, homeDir)

		// Capture output
		oldStdout := os.Stdout
		r, w, _ := os.Pipe()
		os.Stdout = w

		err := Clean([]string{"general"})

		w.Close()
		os.Stdout = oldStdout

		var buf bytes.Buffer
		io.Copy(&buf, r)
		output := buf.String()

		if err != nil {
			t.Errorf("Expected no error, got: %v", err)
		}
		if !strings.Contains(output, "Skipped (not found):") {
			t.Errorf("Expected skipped message, got: %s", output)
		}
	})

	t.Run("Skip non-symlink files", func(t *testing.T) {
		tempDir := t.TempDir()
		dotfilesDir := filepath.Join(tempDir, "dotfiles")
		homeDir := filepath.Join(tempDir, "home")
		os.Setenv("DOT_DIR", dotfilesDir)

		// Setup test environment
		setupTestEnvironment(t, dotfilesDir, homeDir)

		// Create regular file at target path
		targetPath := filepath.Join(homeDir, ".vimrc")
		if err := os.WriteFile(targetPath, []byte("content"), 0644); err != nil {
			t.Fatalf("Failed to create regular file: %v", err)
		}

		// Capture output
		oldStdout := os.Stdout
		r, w, _ := os.Pipe()
		os.Stdout = w

		err := Clean([]string{"general"})

		w.Close()
		os.Stdout = oldStdout

		var buf bytes.Buffer
		io.Copy(&buf, r)
		output := buf.String()

		if err != nil {
			t.Errorf("Expected no error, got: %v", err)
		}
		if !strings.Contains(output, "Skipped (not a symlink):") {
			t.Errorf("Expected skipped message, got: %s", output)
		}

		// Verify file was not removed
		if _, err := os.Stat(targetPath); os.IsNotExist(err) {
			t.Error("Expected regular file to remain")
		}
	})
}

func TestLink(t *testing.T) {
	// Save original DOT_DIR
	originalDotDir := os.Getenv("DOT_DIR")
	defer func() {
		if originalDotDir != "" {
			os.Setenv("DOT_DIR", originalDotDir)
		} else {
			os.Unsetenv("DOT_DIR")
		}
	}()

	t.Run("Create new symlinks", func(t *testing.T) {
		tempDir := t.TempDir()
		dotfilesDir := filepath.Join(tempDir, "dotfiles")
		homeDir := filepath.Join(tempDir, "home")
		os.Setenv("DOT_DIR", dotfilesDir)

		// Setup test environment
		setupTestEnvironment(t, dotfilesDir, homeDir)

		// Capture output
		oldStdout := os.Stdout
		r, w, _ := os.Pipe()
		os.Stdout = w

		err := Link([]string{"general"}, false)

		w.Close()
		os.Stdout = oldStdout

		var buf bytes.Buffer
		io.Copy(&buf, r)
		output := buf.String()

		if err != nil {
			t.Errorf("Expected no error, got: %v", err)
		}
		if !strings.Contains(output, "Created:") {
			t.Errorf("Expected created message, got: %s", output)
		}

		// Verify symlink was created
		targetPath := filepath.Join(homeDir, ".vimrc")
		if _, err := os.Lstat(targetPath); os.IsNotExist(err) {
			t.Error("Expected symlink to be created")
		}
	})

	t.Run("Skip existing correct symlinks", func(t *testing.T) {
		tempDir := t.TempDir()
		dotfilesDir := filepath.Join(tempDir, "dotfiles")
		homeDir := filepath.Join(tempDir, "home")
		os.Setenv("DOT_DIR", dotfilesDir)

		// Setup test environment
		setupTestEnvironment(t, dotfilesDir, homeDir)

		// Create correct symlink first
		sourcePath := filepath.Join(dotfilesDir, "vim/.vimrc")
		targetPath := filepath.Join(homeDir, ".vimrc")
		if err := os.Symlink(sourcePath, targetPath); err != nil {
			t.Fatalf("Failed to create test symlink: %v", err)
		}

		// Capture output
		oldStdout := os.Stdout
		r, w, _ := os.Pipe()
		os.Stdout = w

		err := Link([]string{"general"}, false)

		w.Close()
		os.Stdout = oldStdout

		var buf bytes.Buffer
		io.Copy(&buf, r)
		output := buf.String()

		if err != nil {
			t.Errorf("Expected no error, got: %v", err)
		}
		if !strings.Contains(output, "Skipped (already correct):") {
			t.Errorf("Expected skipped message, got: %s", output)
		}
	})

	t.Run("Override existing incorrect symlinks", func(t *testing.T) {
		tempDir := t.TempDir()
		dotfilesDir := filepath.Join(tempDir, "dotfiles")
		homeDir := filepath.Join(tempDir, "home")
		os.Setenv("DOT_DIR", dotfilesDir)

		// Setup test environment
		setupTestEnvironment(t, dotfilesDir, homeDir)

		// Create incorrect symlink
		wrongSource := filepath.Join(tempDir, "wrong-target")
		targetPath := filepath.Join(homeDir, ".vimrc")
		if err := os.WriteFile(wrongSource, []byte("wrong"), 0644); err != nil {
			t.Fatalf("Failed to create wrong source: %v", err)
		}
		if err := os.Symlink(wrongSource, targetPath); err != nil {
			t.Fatalf("Failed to create incorrect symlink: %v", err)
		}

		err := Link([]string{"general"}, false)

		if err != nil {
			t.Errorf("Expected no error, got: %v", err)
		}

		// Verify the symlink was overridden correctly
		target, err := os.Readlink(targetPath)
		if err != nil {
			t.Errorf("Expected symlink to exist, got error: %v", err)
		}
		expectedTarget := filepath.Join(dotfilesDir, "vim", ".vimrc")
		if target != expectedTarget {
			t.Errorf("Expected symlink to point to %s, got %s", expectedTarget, target)
		}
	})

	t.Run("Backup existing files", func(t *testing.T) {
		tempDir := t.TempDir()
		dotfilesDir := filepath.Join(tempDir, "dotfiles")
		homeDir := filepath.Join(tempDir, "home")
		os.Setenv("DOT_DIR", dotfilesDir)

		// Setup test environment
		setupTestEnvironment(t, dotfilesDir, homeDir)

		// Create existing file
		targetPath := filepath.Join(homeDir, ".vimrc")
		if err := os.WriteFile(targetPath, []byte("existing content"), 0644); err != nil {
			t.Fatalf("Failed to create existing file: %v", err)
		}

		// Capture output
		oldStdout := os.Stdout
		r, w, _ := os.Pipe()
		os.Stdout = w

		err := Link([]string{"general"}, false)

		w.Close()
		os.Stdout = oldStdout

		var buf bytes.Buffer
		io.Copy(&buf, r)
		output := buf.String()

		if err != nil {
			t.Errorf("Expected no error, got: %v", err)
		}
		if !strings.Contains(output, "Backed up:") {
			t.Errorf("Expected backup message, got: %s", output)
		}

		// Verify backup was created
		backupPath := targetPath + ".bak"
		if _, err := os.Stat(backupPath); os.IsNotExist(err) {
			t.Error("Expected backup file to be created")
		}
	})

	t.Run("Dry-run behavior", func(t *testing.T) {
		tempDir := t.TempDir()
		dotfilesDir := filepath.Join(tempDir, "dotfiles")
		homeDir := filepath.Join(tempDir, "home")
		os.Setenv("DOT_DIR", dotfilesDir)

		// Setup test environment
		setupTestEnvironment(t, dotfilesDir, homeDir)

		// Capture output
		oldStdout := os.Stdout
		r, w, _ := os.Pipe()
		os.Stdout = w

		err := Link([]string{"general"}, true)

		w.Close()
		os.Stdout = oldStdout

		var buf bytes.Buffer
		io.Copy(&buf, r)
		output := buf.String()

		if err != nil {
			t.Errorf("Expected no error, got: %v", err)
		}
		if !strings.Contains(output, "Would create:") {
			t.Errorf("Expected dry-run message, got: %s", output)
		}

		// Verify no symlink was actually created
		targetPath := filepath.Join(homeDir, ".vimrc")
		if _, err := os.Lstat(targetPath); !os.IsNotExist(err) {
			t.Error("Expected no symlink to be created in dry-run mode")
		}
	})
}

// Test error handling scenarios
func TestLinkErrorHandling(t *testing.T) {
	// Save original DOT_DIR
	originalDotDir := os.Getenv("DOT_DIR")
	defer func() {
		if originalDotDir != "" {
			os.Setenv("DOT_DIR", originalDotDir)
		} else {
			os.Unsetenv("DOT_DIR")
		}
	}()

	t.Run("Warning for missing source files", func(t *testing.T) {
		tempDir := t.TempDir()
		dotfilesDir := filepath.Join(tempDir, "dotfiles")
		homeDir := filepath.Join(tempDir, "home")
		os.Setenv("DOT_DIR", dotfilesDir)

		// Setup environment but don't create source files
		if err := os.MkdirAll(dotfilesDir, 0755); err != nil {
			t.Fatalf("Failed to create dotfiles directory: %v", err)
		}
		if err := os.MkdirAll(homeDir, 0755); err != nil {
			t.Fatalf("Failed to create home directory: %v", err)
		}

		// Create .mappings without creating source files
		mappingsContent := `[general]
"vim/.vimrc" = "` + filepath.Join(homeDir, ".vimrc") + `"`

		mappingsPath := filepath.Join(dotfilesDir, ".mappings")
		if err := os.WriteFile(mappingsPath, []byte(mappingsContent), 0644); err != nil {
			t.Fatalf("Failed to create .mappings: %v", err)
		}

		// Capture stderr
		oldStderr := os.Stderr
		r, w, _ := os.Pipe()
		os.Stderr = w

		err := Link([]string{"general"}, false)

		w.Close()
		os.Stderr = oldStderr

		var buf bytes.Buffer
		io.Copy(&buf, r)
		output := buf.String()

		if err != nil {
			t.Errorf("Expected no error, got: %v", err)
		}
		if !strings.Contains(output, "Warning: Source file does not exist:") {
			t.Errorf("Expected missing source warning, got: %s", output)
		}
	})

	t.Run("Handle invalid .mappings file", func(t *testing.T) {
		tempDir := t.TempDir()
		dotfilesDir := filepath.Join(tempDir, "dotfiles")
		os.Setenv("DOT_DIR", dotfilesDir)

		// Create dotfiles directory
		if err := os.MkdirAll(dotfilesDir, 0755); err != nil {
			t.Fatalf("Failed to create dotfiles directory: %v", err)
		}

		// Create invalid .mappings file
		mappingsPath := filepath.Join(dotfilesDir, ".mappings")
		if err := os.WriteFile(mappingsPath, []byte("invalid toml ["), 0644); err != nil {
			t.Fatalf("Failed to create invalid .mappings: %v", err)
		}

		err := Link([]string{"general"}, false)
		if err == nil {
			t.Error("Expected error for invalid .mappings file")
		}
		if !strings.Contains(err.Error(), "failed to parse .mappings file") {
			t.Errorf("Expected parse error, got: %v", err)
		}
	})

	t.Run("Handle non-existent profile", func(t *testing.T) {
		tempDir := t.TempDir()
		dotfilesDir := filepath.Join(tempDir, "dotfiles")
		homeDir := filepath.Join(tempDir, "home")
		os.Setenv("DOT_DIR", dotfilesDir)

		// Setup basic environment
		setupTestEnvironment(t, dotfilesDir, homeDir)

		err := Link([]string{"nonexistent"}, false)
		if err == nil {
			t.Error("Expected error for non-existent profile")
		}
		if !strings.Contains(err.Error(), "profile [nonexistent] not found") {
			t.Errorf("Expected profile not found error, got: %v", err)
		}
	})
}

// Test profile precedence
func TestProfilePrecedence(t *testing.T) {
	// Save original DOT_DIR
	originalDotDir := os.Getenv("DOT_DIR")
	defer func() {
		if originalDotDir != "" {
			os.Setenv("DOT_DIR", originalDotDir)
		} else {
			os.Unsetenv("DOT_DIR")
		}
	}()

	t.Run("Profile precedence in link command", func(t *testing.T) {
		tempDir := t.TempDir()
		dotfilesDir := filepath.Join(tempDir, "dotfiles")
		homeDir := filepath.Join(tempDir, "home")
		os.Setenv("DOT_DIR", dotfilesDir)

		// Create dotfiles directory structure
		vimDir := filepath.Join(dotfilesDir, "vim")
		if err := os.MkdirAll(vimDir, 0755); err != nil {
			t.Fatalf("Failed to create vim directory: %v", err)
		}
		if err := os.MkdirAll(homeDir, 0755); err != nil {
			t.Fatalf("Failed to create home directory: %v", err)
		}

		// Create source files
		generalVimrc := filepath.Join(vimDir, ".vimrc")
		workVimrc := filepath.Join(vimDir, ".vimrc-work")
		if err := os.WriteFile(generalVimrc, []byte("general vim config"), 0644); err != nil {
			t.Fatalf("Failed to create general .vimrc: %v", err)
		}
		if err := os.WriteFile(workVimrc, []byte("work vim config"), 0644); err != nil {
			t.Fatalf("Failed to create work .vimrc: %v", err)
		}

		// Create .mappings with profile precedence
		mappingsContent := `[general]
"vim/.vimrc" = "` + filepath.Join(homeDir, ".vimrc") + `"

[work]
"vim/.vimrc-work" = "` + filepath.Join(homeDir, ".vimrc") + `"`

		mappingsPath := filepath.Join(dotfilesDir, ".mappings")
		if err := os.WriteFile(mappingsPath, []byte(mappingsContent), 0644); err != nil {
			t.Fatalf("Failed to create .mappings: %v", err)
		}

		// Test that work profile overrides general
		err := Link([]string{"general", "work"}, false)
		if err != nil {
			t.Errorf("Expected no error, got: %v", err)
		}

		// Verify the correct symlink was created (work should override general)
		targetPath := filepath.Join(homeDir, ".vimrc")
		linkTarget, err := os.Readlink(targetPath)
		if err != nil {
			t.Fatalf("Failed to read symlink: %v", err)
		}

		expectedTarget := workVimrc
		if linkTarget != expectedTarget {
			t.Errorf("Expected link to point to %s, got %s", expectedTarget, linkTarget)
		}
	})
}

// Helper function to setup test environment with dotfiles and .mappings
func setupTestEnvironment(t *testing.T, dotfilesDir, homeDir string) {
	// Create dotfiles directory structure
	vimDir := filepath.Join(dotfilesDir, "vim")
	if err := os.MkdirAll(vimDir, 0755); err != nil {
		t.Fatalf("Failed to create vim directory: %v", err)
	}

	// Create home directory
	if err := os.MkdirAll(homeDir, 0755); err != nil {
		t.Fatalf("Failed to create home directory: %v", err)
	}

	// Create source files
	vimrcPath := filepath.Join(vimDir, ".vimrc")
	if err := os.WriteFile(vimrcPath, []byte("\" vim config"), 0644); err != nil {
		t.Fatalf("Failed to create .vimrc: %v", err)
	}

	// Create .mappings file with home directory references
	mappingsContent := `[general]
"vim/.vimrc" = "` + filepath.Join(homeDir, ".vimrc") + `"

[work]
"vim/.vimrc" = "` + filepath.Join(homeDir, ".vimrc") + `"`

	mappingsPath := filepath.Join(dotfilesDir, ".mappings")
	if err := os.WriteFile(mappingsPath, []byte(mappingsContent), 0644); err != nil {
		t.Fatalf("Failed to create .mappings: %v", err)
	}
}

func TestList(t *testing.T) {
	// Save original DOT_DIR
	originalDotDir := os.Getenv("DOT_DIR")
	defer func() {
		if originalDotDir != "" {
			os.Setenv("DOT_DIR", originalDotDir)
		} else {
			os.Unsetenv("DOT_DIR")
		}
	}()

	t.Run("List with correct symlinks", func(t *testing.T) {
		tempDir := t.TempDir()
		dotfilesDir := filepath.Join(tempDir, "dotfiles")
		homeDir := filepath.Join(tempDir, "home")
		os.Setenv("DOT_DIR", dotfilesDir)

		// Setup test environment
		setupTestEnvironment(t, dotfilesDir, homeDir)

		// Create correct symlinks
		sourcePath := filepath.Join(dotfilesDir, "vim", ".vimrc")
		targetPath := filepath.Join(homeDir, ".vimrc")
		if err := os.Symlink(sourcePath, targetPath); err != nil {
			t.Fatalf("Failed to create symlink: %v", err)
		}

		// Capture stdout
		oldStdout := os.Stdout
		r, w, _ := os.Pipe()
		os.Stdout = w

		err := List([]string{"general"})

		w.Close()
		os.Stdout = oldStdout

		var buf bytes.Buffer
		io.Copy(&buf, r)
		output := buf.String()

		if err != nil {
			t.Errorf("Expected no error, got: %v", err)
		}
		if !strings.Contains(output, "✅") {
			t.Errorf("Expected success indicator, got: %s", output)
		}
		if !strings.Contains(output, ".vimrc") {
			t.Errorf("Expected .vimrc in output, got: %s", output)
		}
	})

	t.Run("List with missing symlinks", func(t *testing.T) {
		tempDir := t.TempDir()
		dotfilesDir := filepath.Join(tempDir, "dotfiles")
		homeDir := filepath.Join(tempDir, "home")
		os.Setenv("DOT_DIR", dotfilesDir)

		// Setup test environment
		setupTestEnvironment(t, dotfilesDir, homeDir)

		// Don't create any symlinks

		// Capture stdout
		oldStdout := os.Stdout
		r, w, _ := os.Pipe()
		os.Stdout = w

		err := List([]string{"general"})

		w.Close()
		os.Stdout = oldStdout

		var buf bytes.Buffer
		io.Copy(&buf, r)
		output := buf.String()

		if err != nil {
			t.Errorf("Expected no error, got: %v", err)
		}
		if !strings.Contains(output, "❌") {
			t.Errorf("Expected error indicator, got: %s", output)
		}
		if !strings.Contains(output, "(not linked)") {
			t.Errorf("Expected 'not linked' message, got: %s", output)
		}
	})

	t.Run("List with incorrect symlinks", func(t *testing.T) {
		tempDir := t.TempDir()
		dotfilesDir := filepath.Join(tempDir, "dotfiles")
		homeDir := filepath.Join(tempDir, "home")
		os.Setenv("DOT_DIR", dotfilesDir)

		// Setup test environment
		setupTestEnvironment(t, dotfilesDir, homeDir)

		// Create incorrect symlink
		wrongTarget := filepath.Join(tempDir, "wrong.txt")
		targetPath := filepath.Join(homeDir, ".vimrc")
		if err := os.WriteFile(wrongTarget, []byte("wrong"), 0644); err != nil {
			t.Fatalf("Failed to create wrong target: %v", err)
		}
		if err := os.Symlink(wrongTarget, targetPath); err != nil {
			t.Fatalf("Failed to create incorrect symlink: %v", err)
		}

		// Capture stdout
		oldStdout := os.Stdout
		r, w, _ := os.Pipe()
		os.Stdout = w

		err := List([]string{"general"})

		w.Close()
		os.Stdout = oldStdout

		var buf bytes.Buffer
		io.Copy(&buf, r)
		output := buf.String()

		if err != nil {
			t.Errorf("Expected no error, got: %v", err)
		}
		if !strings.Contains(output, "❌") {
			t.Errorf("Expected error indicator, got: %s", output)
		}
		if !strings.Contains(output, "(expected:") {
			t.Errorf("Expected 'expected:' message, got: %s", output)
		}
	})

	t.Run("List with missing source files", func(t *testing.T) {
		tempDir := t.TempDir()
		dotfilesDir := filepath.Join(tempDir, "dotfiles")
		homeDir := filepath.Join(tempDir, "home")

		// Setup test environment without creating source files
		os.MkdirAll(dotfilesDir, 0755)
		os.MkdirAll(homeDir, 0755)
		os.Setenv("DOT_DIR", dotfilesDir)

		// Create .mappings file
		mappingsContent := `[general]
"vim/.vimrc" = "~/.vimrc"`
		mappingsPath := filepath.Join(dotfilesDir, ".mappings")
		if err := os.WriteFile(mappingsPath, []byte(mappingsContent), 0644); err != nil {
			t.Fatalf("Failed to create .mappings: %v", err)
		}

		// Override HOME for this test
		oldHome := os.Getenv("HOME")
		os.Setenv("HOME", homeDir)
		defer os.Setenv("HOME", oldHome)

		// Create correct symlink but with missing source
		sourcePath := filepath.Join(dotfilesDir, "vim", ".vimrc")
		targetPath := filepath.Join(homeDir, ".vimrc")
		os.MkdirAll(filepath.Dir(targetPath), 0755)
		if err := os.Symlink(sourcePath, targetPath); err != nil {
			t.Fatalf("Failed to create symlink: %v", err)
		}

		// Capture stdout
		oldStdout := os.Stdout
		r, w, _ := os.Pipe()
		os.Stdout = w

		err := List([]string{"general"})

		w.Close()
		os.Stdout = oldStdout

		var buf bytes.Buffer
		io.Copy(&buf, r)
		output := buf.String()

		if err != nil {
			t.Errorf("Expected no error, got: %v", err)
		}
		if !strings.Contains(output, "⚠️") {
			t.Errorf("Expected warning indicator, got: %s", output)
		}
		if !strings.Contains(output, "(source missing)") {
			t.Errorf("Expected 'source missing' message, got: %s", output)
		}
	})

	t.Run("List with regular file at target path", func(t *testing.T) {
		tempDir := t.TempDir()
		dotfilesDir := filepath.Join(tempDir, "dotfiles")
		homeDir := filepath.Join(tempDir, "home")
		os.Setenv("DOT_DIR", dotfilesDir)

		// Setup test environment
		setupTestEnvironment(t, dotfilesDir, homeDir)

		// Create regular file at target path
		targetPath := filepath.Join(homeDir, ".vimrc")
		if err := os.WriteFile(targetPath, []byte("regular file"), 0644); err != nil {
			t.Fatalf("Failed to create regular file: %v", err)
		}

		// Capture stdout
		oldStdout := os.Stdout
		r, w, _ := os.Pipe()
		os.Stdout = w

		err := List([]string{"general"})

		w.Close()
		os.Stdout = oldStdout

		var buf bytes.Buffer
		io.Copy(&buf, r)
		output := buf.String()

		if err != nil {
			t.Errorf("Expected no error, got: %v", err)
		}
		if !strings.Contains(output, "❌") {
			t.Errorf("Expected error indicator, got: %s", output)
		}
		if !strings.Contains(output, "(exists but not a symlink)") {
			t.Errorf("Expected 'exists but not a symlink' message, got: %s", output)
		}
	})

	t.Run("List with multiple profiles", func(t *testing.T) {
		tempDir := t.TempDir()
		dotfilesDir := filepath.Join(tempDir, "dotfiles")
		homeDir := filepath.Join(tempDir, "home")

		// Create mappings with multiple profiles
		os.MkdirAll(dotfilesDir, 0755)
		os.MkdirAll(homeDir, 0755)
		os.Setenv("DOT_DIR", dotfilesDir)

		// Create .mappings file
		mappingsContent := `[general]
"vim/.vimrc" = "~/.vimrc"

[work]
"work/.workrc" = "~/.workrc"`
		mappingsPath := filepath.Join(dotfilesDir, ".mappings")
		if err := os.WriteFile(mappingsPath, []byte(mappingsContent), 0644); err != nil {
			t.Fatalf("Failed to create .mappings: %v", err)
		}

		// Create source files
		os.MkdirAll(filepath.Join(dotfilesDir, "vim"), 0755)
		os.MkdirAll(filepath.Join(dotfilesDir, "work"), 0755)
		os.WriteFile(filepath.Join(dotfilesDir, "vim", ".vimrc"), []byte("vim config"), 0644)
		os.WriteFile(filepath.Join(dotfilesDir, "work", ".workrc"), []byte("work config"), 0644)

		// Override HOME for this test
		oldHome := os.Getenv("HOME")
		os.Setenv("HOME", homeDir)
		defer os.Setenv("HOME", oldHome)

		// Capture stdout
		oldStdout := os.Stdout
		r, w, _ := os.Pipe()
		os.Stdout = w

		err := List([]string{"general", "work"})

		w.Close()
		os.Stdout = oldStdout

		var buf bytes.Buffer
		io.Copy(&buf, r)
		output := buf.String()

		if err != nil {
			t.Errorf("Expected no error, got: %v", err)
		}
		if !strings.Contains(output, "general, work") {
			t.Errorf("Expected profile names in output, got: %s", output)
		}
		if !strings.Contains(output, ".vimrc") {
			t.Errorf("Expected .vimrc in output, got: %s", output)
		}
		if !strings.Contains(output, ".workrc") {
			t.Errorf("Expected .workrc in output, got: %s", output)
		}
	})
}
