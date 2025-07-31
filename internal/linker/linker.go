package linker

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/yourusername/dot/internal/config"
	"github.com/yourusername/dot/internal/dotfiles"
	"github.com/yourusername/dot/internal/utils"
)

// Check verifies that symbolic links exist and point to correct source files
func Check(profiles []string) error {
	dotfilesDir, err := dotfiles.GetDotfilesDir()
	if err != nil {
		return err
	}

	cfg, err := config.ParseConfig(dotfilesDir)
	if err != nil {
		return err
	}

	profileMap, err := cfg.GetProfiles(profiles)
	if err != nil {
		return err
	}

	var issues []string

	for source, target := range profileMap {
		targetPath := utils.ExpandPath(target)
		sourcePath := filepath.Join(dotfilesDir, source)

		// Check if target exists
		stat, err := os.Lstat(targetPath)
		if os.IsNotExist(err) {
			issues = append(issues, fmt.Sprintf("Missing link: %s", targetPath))
			continue
		}
		if err != nil {
			issues = append(issues, fmt.Sprintf("Error checking %s: %v", targetPath, err))
			continue
		}

		// Check if target is a symbolic link
		if stat.Mode()&os.ModeSymlink == 0 {
			issues = append(issues, fmt.Sprintf("Not a symlink: %s", targetPath))
			continue
		}

		// Check if link points to correct source
		linkTarget, err := os.Readlink(targetPath)
		if err != nil {
			issues = append(issues, fmt.Sprintf("Error reading link %s: %v", targetPath, err))
			continue
		}

		if linkTarget != sourcePath {
			issues = append(issues, fmt.Sprintf("Incorrect link: %s -> %s (expected: %s)", targetPath, linkTarget, sourcePath))
		}
	}

	if len(issues) == 0 {
		fmt.Println("All links are correct")
	} else {
		for _, issue := range issues {
			fmt.Fprintf(os.Stderr, "%s\n", issue)
		}
		return fmt.Errorf("found %d issue(s)", len(issues))
	}

	return nil
}

// Clean removes all registered symbolic links
func Clean(profiles []string) error {
	dotfilesDir, err := dotfiles.GetDotfilesDir()
	if err != nil {
		return err
	}

	cfg, err := config.ParseConfig(dotfilesDir)
	if err != nil {
		return err
	}

	profileMap, err := cfg.GetProfiles(profiles)
	if err != nil {
		return err
	}

	for _, target := range profileMap {
		targetPath := utils.ExpandPath(target)

		// Check if target exists and is a symlink
		stat, err := os.Lstat(targetPath)
		if os.IsNotExist(err) {
			fmt.Printf("Skipped (not found): %s\n", targetPath)
			continue
		}
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error checking %s: %v\n", targetPath, err)
			continue
		}

		if stat.Mode()&os.ModeSymlink == 0 {
			fmt.Printf("Skipped (not a symlink): %s\n", targetPath)
			continue
		}

		// Remove the symlink
		if err := os.Remove(targetPath); err != nil {
			fmt.Fprintf(os.Stderr, "Error removing %s: %v\n", targetPath, err)
		} else {
			fmt.Printf("Removed: %s\n", targetPath)
		}
	}

	return nil
}

// Link creates symbolic links based on the .mappings file
func Link(profiles []string, dryRun bool) error {
	dotfilesDir, err := dotfiles.GetDotfilesDir()
	if err != nil {
		return err
	}

	cfg, err := config.ParseConfig(dotfilesDir)
	if err != nil {
		return err
	}

	profileMap, err := cfg.GetProfiles(profiles)
	if err != nil {
		return err
	}

	for source, target := range profileMap {
		targetPath := utils.ExpandPath(target)
		sourcePath := filepath.Join(dotfilesDir, source)

		// Check if source file exists
		if _, err := os.Stat(sourcePath); os.IsNotExist(err) {
			fmt.Fprintf(os.Stderr, "Warning: Source file does not exist: %s\n", sourcePath)
			continue
		}

		// Handle existing target
		if stat, err := os.Lstat(targetPath); err == nil {
			if stat.Mode()&os.ModeSymlink != 0 {
				// Target is a symlink
				linkTarget, err := os.Readlink(targetPath)
				if err != nil {
					fmt.Fprintf(os.Stderr, "Error reading existing link %s: %v\n", targetPath, err)
					continue
				}

				if linkTarget == sourcePath {
					fmt.Printf("Skipped (already correct): %s -> %s\n", targetPath, sourcePath)
					continue
				} else {
					// Remove existing symlink to override it
					if !dryRun {
						if err := os.Remove(targetPath); err != nil {
							fmt.Fprintf(os.Stderr, "Error removing existing link %s: %v\n", targetPath, err)
							continue
						}
					}
					fmt.Printf("Overriding: %s (was pointing to %s)\n", targetPath, linkTarget)
				}
			} else {
				// Target is a file or directory, back it up
				if !dryRun {
					if err := utils.BackupFile(targetPath); err != nil {
						fmt.Fprintf(os.Stderr, "Error backing up %s: %v\n", targetPath, err)
						continue
					}
				}
				fmt.Printf("Backed up: %s -> %s.bak\n", targetPath, targetPath)
			}
		}

		// Create the symlink
		if dryRun {
			fmt.Printf("Would create: %s -> %s\n", targetPath, sourcePath)
		} else {
			// Ensure target directory exists
			if err := os.MkdirAll(filepath.Dir(targetPath), 0755); err != nil {
				fmt.Fprintf(os.Stderr, "Error creating directory for %s: %v\n", targetPath, err)
				continue
			}

			if err := os.Symlink(sourcePath, targetPath); err != nil {
				fmt.Fprintf(os.Stderr, "Error creating link %s -> %s: %v\n", targetPath, sourcePath, err)
			} else {
				fmt.Printf("Created: %s -> %s\n", targetPath, sourcePath)
			}
		}
	}

	return nil
}

// ParseProfiles parses a comma-separated list of profile names
func ParseProfiles(profileStr string) []string {
	if profileStr == "" {
		return []string{"general"}
	}

	profiles := strings.Split(profileStr, ",")
	for i, profile := range profiles {
		profiles[i] = strings.TrimSpace(profile)
	}

	return profiles
}

// List shows all symbolic links that are currently set based on the profiles
func List(profiles []string) error {
	dotfilesDir, err := dotfiles.GetDotfilesDir()
	if err != nil {
		return err
	}

	cfg, err := config.ParseConfig(dotfilesDir)
	if err != nil {
		return err
	}

	profileMap, err := cfg.GetProfiles(profiles)
	if err != nil {
		return err
	}

	fmt.Printf("Dotfiles links for profile(s): %s\n", strings.Join(profiles, ", "))
	fmt.Println()

	linksFound := false

	for source, target := range profileMap {
		targetPath := utils.ExpandPath(target)
		sourcePath := filepath.Join(dotfilesDir, source)

		// Check if target exists and what type it is
		if stat, err := os.Lstat(targetPath); err == nil {
			if stat.Mode()&os.ModeSymlink != 0 {
				// Target is a symlink
				linkTarget, err := os.Readlink(targetPath)
				if err != nil { //nolint:gocritic
					fmt.Printf("❌ %s -> ??? (error reading link: %v)\n", targetPath, err)
				} else if linkTarget == sourcePath {
					// Check if source actually exists
					if utils.FileExists(sourcePath) {
						fmt.Printf("✅ %s -> %s\n", targetPath, sourcePath)
					} else {
						fmt.Printf("⚠️  %s -> %s (source missing)\n", targetPath, sourcePath)
					}
				} else {
					fmt.Printf("❌ %s -> %s (expected: %s)\n", targetPath, linkTarget, sourcePath)
				}
				linksFound = true
			} else {
				fmt.Printf("❌ %s (exists but not a symlink)\n", targetPath)
				linksFound = true
			}
		} else {
			fmt.Printf("❌ %s (not linked)\n", targetPath)
			linksFound = true
		}
	}

	if !linksFound {
		fmt.Println("No dotfile mappings found in the specified profile(s).")
	}

	return nil
}
