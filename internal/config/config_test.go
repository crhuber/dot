package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestParseConfig(t *testing.T) {
	t.Run("Valid TOML with general and other profiles", func(t *testing.T) {
		content := `[general]
"vim/.vimrc" = "~/.vimrc"
"git/.gitconfig" = "~/.gitconfig"

[work]
"vim/.vimrc" = "~/.vimrc"
"git/.gitconfig-work" = "~/.gitconfig"

[minimal]
"vim/.vimrc" = "~/.vimrc"`

		tempDir := createTempMappings(t, content)
		config, err := ParseConfig(tempDir)

		if err != nil {
			t.Fatalf("Expected no error, got: %v", err)
		}

		// Check that all profiles exist
		if len(config.Profiles) != 3 {
			t.Errorf("Expected 3 profiles, got %d", len(config.Profiles))
		}

		// Check general profile
		general, exists := config.Profiles["general"]
		if !exists {
			t.Error("Expected [general] profile to exist")
		}
		if general["vim/.vimrc"] != "~/.vimrc" {
			t.Errorf("Expected vim/.vimrc -> ~/.vimrc, got %s", general["vim/.vimrc"])
		}
		if general["git/.gitconfig"] != "~/.gitconfig" {
			t.Errorf("Expected git/.gitconfig -> ~/.gitconfig, got %s", general["git/.gitconfig"])
		}

		// Check work profile
		work, exists := config.Profiles["work"]
		if !exists {
			t.Error("Expected [work] profile to exist")
		}
		if work["git/.gitconfig-work"] != "~/.gitconfig" {
			t.Errorf("Expected git/.gitconfig-work -> ~/.gitconfig, got %s", work["git/.gitconfig-work"])
		}

		// Check minimal profile
		minimal, exists := config.Profiles["minimal"]
		if !exists {
			t.Error("Expected [minimal] profile to exist")
		}
		if len(minimal) != 1 {
			t.Errorf("Expected minimal profile to have 1 entry, got %d", len(minimal))
		}
	})

	t.Run("Missing general profile should error", func(t *testing.T) {
		content := `[work]
"vim/.vimrc" = "~/.vimrc"
"git/.gitconfig-work" = "~/.gitconfig"`

		tempDir := createTempMappings(t, content)
		_, err := ParseConfig(tempDir)

		if err == nil {
			t.Error("Expected error for missing [general] profile")
		}
		if !strings.Contains(err.Error(), "[general] profile is required") {
			t.Errorf("Expected error about missing [general] profile, got: %v", err)
		}
	})

	t.Run("Invalid TOML syntax should error", func(t *testing.T) {
		content := `[general]
"vim/.vimrc" = "~/.vimrc"
[work
"invalid toml"`

		tempDir := createTempMappings(t, content)
		_, err := ParseConfig(tempDir)

		if err == nil {
			t.Error("Expected error for invalid TOML syntax")
		}
		if !strings.Contains(err.Error(), "failed to parse .mappings file") {
			t.Errorf("Expected parse error, got: %v", err)
		}
	})

	t.Run("Non-existent .mappings file should error", func(t *testing.T) {
		tempDir := t.TempDir()
		_, err := ParseConfig(tempDir)

		if err == nil {
			t.Error("Expected error for non-existent .mappings file")
		}
		if !strings.Contains(err.Error(), ".mappings file not found") {
			t.Errorf("Expected file not found error, got: %v", err)
		}
	})

	t.Run("Empty .mappings file should error", func(t *testing.T) {
		content := ""
		tempDir := createTempMappings(t, content)
		_, err := ParseConfig(tempDir)

		if err == nil {
			t.Error("Expected error for empty .mappings file")
		}
		if !strings.Contains(err.Error(), "[general] profile is required") {
			t.Errorf("Expected error about missing [general] profile, got: %v", err)
		}
	})

	t.Run("Only general profile", func(t *testing.T) {
		content := `[general]
"vim/.vimrc" = "~/.vimrc"
"git/.gitconfig" = "~/.gitconfig"`

		tempDir := createTempMappings(t, content)
		config, err := ParseConfig(tempDir)

		if err != nil {
			t.Fatalf("Expected no error, got: %v", err)
		}

		if len(config.Profiles) != 1 {
			t.Errorf("Expected 1 profile, got %d", len(config.Profiles))
		}

		general, exists := config.Profiles["general"]
		if !exists {
			t.Error("Expected [general] profile to exist")
		}
		if len(general) != 2 {
			t.Errorf("Expected 2 entries in general profile, got %d", len(general))
		}
	})
}

func TestGetProfiles(t *testing.T) {
	// Setup test configuration
	content := `[general]
"vim/.vimrc" = "~/.vimrc"
"git/.gitconfig" = "~/.gitconfig"
"zsh/.zshrc" = "~/.zshrc"

[work]
"git/.gitconfig-work" = "~/.gitconfig"
"ssh/work_config" = "~/.ssh/config"

[minimal]
"vim/.vimrc" = "~/.vimrc"`

	tempDir := createTempMappings(t, content)
	config, err := ParseConfig(tempDir)
	if err != nil {
		t.Fatalf("Failed to parse config: %v", err)
	}

	t.Run("Default to general when no profiles specified", func(t *testing.T) {
		result, err := config.GetProfiles([]string{})
		if err != nil {
			t.Fatalf("Expected no error, got: %v", err)
		}

		if len(result) != 3 {
			t.Errorf("Expected 3 entries from general profile, got %d", len(result))
		}
		if result["vim/.vimrc"] != "~/.vimrc" {
			t.Errorf("Expected vim/.vimrc -> ~/.vimrc, got %s", result["vim/.vimrc"])
		}
		if result["git/.gitconfig"] != "~/.gitconfig" {
			t.Errorf("Expected git/.gitconfig -> ~/.gitconfig, got %s", result["git/.gitconfig"])
		}
	})

	t.Run("Default to general when nil profiles specified", func(t *testing.T) {
		result, err := config.GetProfiles(nil)
		if err != nil {
			t.Fatalf("Expected no error, got: %v", err)
		}

		if len(result) != 3 {
			t.Errorf("Expected 3 entries from general profile, got %d", len(result))
		}
	})

	t.Run("Single profile", func(t *testing.T) {
		result, err := config.GetProfiles([]string{"minimal"})
		if err != nil {
			t.Fatalf("Expected no error, got: %v", err)
		}

		// Should have general + minimal, but minimal overrides general's vim/.vimrc
		expectedEntries := 3 // vim/.vimrc (from minimal), git/.gitconfig, zsh/.zshrc (from general)
		if len(result) != expectedEntries {
			t.Errorf("Expected %d entries, got %d", expectedEntries, len(result))
		}
		if result["vim/.vimrc"] != "~/.vimrc" {
			t.Errorf("Expected vim/.vimrc -> ~/.vimrc, got %s", result["vim/.vimrc"])
		}
	})

	t.Run("Last profile overrides earlier ones", func(t *testing.T) {
		result, err := config.GetProfiles([]string{"general", "work"})
		if err != nil {
			t.Fatalf("Expected no error, got: %v", err)
		}

		// work profile should override git/.gitconfig
		if result["git/.gitconfig-work"] != "~/.gitconfig" {
			t.Errorf("Expected work profile to set git/.gitconfig-work, got %s", result["git/.gitconfig-work"])
		}
		// But general entries should still be there
		if result["vim/.vimrc"] != "~/.vimrc" {
			t.Errorf("Expected vim/.vimrc from general, got %s", result["vim/.vimrc"])
		}
		if result["zsh/.zshrc"] != "~/.zshrc" {
			t.Errorf("Expected zsh/.zshrc from general, got %s", result["zsh/.zshrc"])
		}
	})

	t.Run("General has lowest precedence", func(t *testing.T) {
		result, err := config.GetProfiles([]string{"work", "general"})
		if err != nil {
			t.Fatalf("Expected no error, got: %v", err)
		}

		// Since general comes last, it should NOT override work's git config
		// But this tests our logic - general is always applied first regardless of order
		if result["git/.gitconfig-work"] != "~/.gitconfig" {
			t.Errorf("Expected work profile git/.gitconfig-work to remain, got %s", result["git/.gitconfig-work"])
		}
	})

	t.Run("Multiple profiles with precedence", func(t *testing.T) {
		result, err := config.GetProfiles([]string{"minimal", "work"})
		if err != nil {
			t.Fatalf("Expected no error, got: %v", err)
		}

		// Should have all entries from general as base, then work overrides
		if result["vim/.vimrc"] != "~/.vimrc" {
			t.Errorf("Expected vim/.vimrc from general/minimal, got %s", result["vim/.vimrc"])
		}
		if result["git/.gitconfig-work"] != "~/.gitconfig" {
			t.Errorf("Expected git/.gitconfig-work from work, got %s", result["git/.gitconfig-work"])
		}
		if result["ssh/work_config"] != "~/.ssh/config" {
			t.Errorf("Expected ssh/work_config from work, got %s", result["ssh/work_config"])
		}
	})

	t.Run("Error when requesting non-existent profile", func(t *testing.T) {
		_, err := config.GetProfiles([]string{"nonexistent"})
		if err == nil {
			t.Error("Expected error for non-existent profile")
		}
		if !strings.Contains(err.Error(), "profile [nonexistent] not found") {
			t.Errorf("Expected error about nonexistent profile, got: %v", err)
		}
	})

	t.Run("Mix of valid and invalid profiles", func(t *testing.T) {
		_, err := config.GetProfiles([]string{"general", "nonexistent"})
		if err == nil {
			t.Error("Expected error for mix with non-existent profile")
		}
		if !strings.Contains(err.Error(), "profile [nonexistent] not found") {
			t.Errorf("Expected error about nonexistent profile, got: %v", err)
		}
	})

	t.Run("Explicit general profile", func(t *testing.T) {
		result, err := config.GetProfiles([]string{"general"})
		if err != nil {
			t.Fatalf("Expected no error, got: %v", err)
		}

		if len(result) != 3 {
			t.Errorf("Expected 3 entries from general profile, got %d", len(result))
		}
		if result["vim/.vimrc"] != "~/.vimrc" {
			t.Errorf("Expected vim/.vimrc -> ~/.vimrc, got %s", result["vim/.vimrc"])
		}
	})

	t.Run("Profile precedence with duplicate entries", func(t *testing.T) {
		// Test that later profiles completely override earlier ones for same keys
		result, err := config.GetProfiles([]string{"general", "work", "minimal"})
		if err != nil {
			t.Fatalf("Expected no error, got: %v", err)
		}

		// vim/.vimrc should come from minimal (last profile with this key)
		if result["vim/.vimrc"] != "~/.vimrc" {
			t.Errorf("Expected vim/.vimrc from minimal profile, got %s", result["vim/.vimrc"])
		}
		// work profile entries should still be there
		if result["git/.gitconfig-work"] != "~/.gitconfig" {
			t.Errorf("Expected git/.gitconfig-work from work profile, got %s", result["git/.gitconfig-work"])
		}
		// general profile entries that aren't overridden should be there
		if result["zsh/.zshrc"] != "~/.zshrc" {
			t.Errorf("Expected zsh/.zshrc from general profile, got %s", result["zsh/.zshrc"])
		}
	})

	t.Run("General profile applied even when explicitly specified later", func(t *testing.T) {
		// Test that general is always applied first, regardless of position in list
		result, err := config.GetProfiles([]string{"work", "general"})
		if err != nil {
			t.Fatalf("Expected no error, got: %v", err)
		}

		// work profile should still override general where they conflict
		if result["git/.gitconfig-work"] != "~/.gitconfig" {
			t.Errorf("Expected git/.gitconfig-work from work to remain, got %s", result["git/.gitconfig-work"])
		}
	})
}

// Helper function to create temporary .mappings file for testing
func createTempMappings(t *testing.T, content string) string {
	tempDir := t.TempDir()
	mappingsPath := filepath.Join(tempDir, ".mappings")

	if err := os.WriteFile(mappingsPath, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to create temp .mappings file: %v", err)
	}

	return tempDir
}

// Benchmark tests for performance
func BenchmarkParseConfig(b *testing.B) {
	content := `[general]
"vim/.vimrc" = "~/.vimrc"
"git/.gitconfig" = "~/.gitconfig"
"zsh/.zshrc" = "~/.zshrc"
"tmux/.tmux.conf" = "~/.tmux.conf"

[work]
"git/.gitconfig-work" = "~/.gitconfig"
"ssh/work_config" = "~/.ssh/config"
"vim/.vimrc-work" = "~/.vimrc"

[minimal]
"vim/.vimrc" = "~/.vimrc"
"git/.gitconfig" = "~/.gitconfig"`

	tempDir, err := os.MkdirTemp("", "benchmark-config-*")
	if err != nil {
		b.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	mappingsPath := filepath.Join(tempDir, ".mappings")
	if err := os.WriteFile(mappingsPath, []byte(content), 0644); err != nil {
		b.Fatalf("Failed to create .mappings file: %v", err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := ParseConfig(tempDir)
		if err != nil {
			b.Fatalf("ParseConfig failed: %v", err)
		}
	}
}

func BenchmarkGetProfiles(b *testing.B) {
	content := `[general]
"vim/.vimrc" = "~/.vimrc"
"git/.gitconfig" = "~/.gitconfig"
"zsh/.zshrc" = "~/.zshrc"
"tmux/.tmux.conf" = "~/.tmux.conf"

[work]
"git/.gitconfig-work" = "~/.gitconfig"
"ssh/work_config" = "~/.ssh/config"
"vim/.vimrc-work" = "~/.vimrc"

[minimal]
"vim/.vimrc" = "~/.vimrc"
"git/.gitconfig" = "~/.gitconfig"`

	tempDir, err := os.MkdirTemp("", "benchmark-config-*")
	if err != nil {
		b.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	mappingsPath := filepath.Join(tempDir, ".mappings")
	if err := os.WriteFile(mappingsPath, []byte(content), 0644); err != nil {
		b.Fatalf("Failed to create .mappings file: %v", err)
	}

	config, err := ParseConfig(tempDir)
	if err != nil {
		b.Fatalf("ParseConfig failed: %v", err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := config.GetProfiles([]string{"general", "work", "minimal"})
		if err != nil {
			b.Fatalf("GetProfiles failed: %v", err)
		}
	}
}
