package dotfiles

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

// GetDotfilesDir returns the dotfiles directory path
// Uses $DOT_DIR environment variable if set, otherwise defaults to ~/.dotfiles
func GetDotfilesDir() (string, error) {
	if dotDir := os.Getenv("DOT_DIR"); dotDir != "" {
		return dotDir, nil
	}

	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to get user home directory: %w", err)
	}

	return filepath.Join(homeDir, ".dotfiles"), nil
}

// Clone clones a repository to the dotfiles directory
func Clone(repoURL string) error {
	dotfilesDir, err := GetDotfilesDir()
	if err != nil {
		return err
	}

	// Check if destination exists and is non-empty
	if stat, err := os.Stat(dotfilesDir); err == nil {
		if stat.IsDir() {
			entries, err := os.ReadDir(dotfilesDir)
			if err != nil {
				return fmt.Errorf("failed to read dotfiles directory: %w", err)
			}
			if len(entries) > 0 {
				return fmt.Errorf("dotfiles directory %s already exists and is non-empty", dotfilesDir)
			}
		} else {
			return fmt.Errorf("dotfiles path %s exists but is not a directory", dotfilesDir)
		}
	}

	// Execute git clone command
	cmd := exec.Command("git", "clone", repoURL, dotfilesDir)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to clone repository: %w", err)
	}

	// Validate that .mappings file exists
	mappingsPath := filepath.Join(dotfilesDir, ".mappings")
	if _, err := os.Stat(mappingsPath); os.IsNotExist(err) {
		return fmt.Errorf("cloned repository does not contain a .mappings file")
	}

	return nil
}

// PrintRoot prints the dotfiles directory path
func PrintRoot() error {
	dotfilesDir, err := GetDotfilesDir()
	if err != nil {
		return err
	}

	fmt.Println(dotfilesDir)
	return nil
}
