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

// Update changes to the dotfiles directory and runs git pull
func Update() error {
	dotfilesDir, err := GetDotfilesDir()
	if err != nil {
		return err
	}

	// Check if the dotfiles directory exists
	if _, err := os.Stat(dotfilesDir); os.IsNotExist(err) {
		return fmt.Errorf("dotfiles directory %s does not exist", dotfilesDir)
	}

	// Execute git pull command in the dotfiles directory
	cmd := exec.Command("git", "pull")
	cmd.Dir = dotfilesDir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to update dotfiles repository: %w", err)
	}

	return nil
}

// Open opens the dotfiles directory in the system file manager
func Open() error {
	dotfilesDir, err := GetDotfilesDir()
	if err != nil {
		return err
	}

	// Check if the dotfiles directory exists
	if _, err := os.Stat(dotfilesDir); os.IsNotExist(err) {
		return fmt.Errorf("dotfiles directory %s does not exist", dotfilesDir)
	}

	// Determine the command based on the operating system
	// Try different commands in order of likelihood
	var cmd *exec.Cmd
	var cmdErr error

	// Try macOS first
	if _, err := exec.LookPath("open"); err == nil {
		cmd = exec.Command("open", dotfilesDir)
		cmdErr = cmd.Run()
		if cmdErr == nil {
			return nil
		}
	}

	// Try Linux/Unix with xdg-open
	if _, err := exec.LookPath("xdg-open"); err == nil {
		cmd = exec.Command("xdg-open", dotfilesDir)
		cmdErr = cmd.Run()
		if cmdErr == nil {
			return nil
		}
	}

	// Try Windows
	if _, err := exec.LookPath("explorer"); err == nil {
		cmd = exec.Command("explorer", dotfilesDir)
		cmdErr = cmd.Run()
		if cmdErr == nil {
			return nil
		}
	}

	if cmdErr != nil {
		return fmt.Errorf("failed to open dotfiles directory: %w", cmdErr)
	}

	return fmt.Errorf("no suitable file manager command found (tried: open, xdg-open, explorer)")
}
