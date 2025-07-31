package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/BurntSushi/toml"
)

// Profile represents a mapping of source paths to target paths
type Profile map[string]string

// Config represents the entire .mappings configuration
type Config struct {
	Profiles map[string]Profile
}

// ParseConfig reads and parses the .mappings file from the dotfiles directory
func ParseConfig(dotfilesDir string) (*Config, error) {
	mappingsPath := filepath.Join(dotfilesDir, ".mappings")

	// Check if .mappings file exists
	if _, err := os.Stat(mappingsPath); os.IsNotExist(err) {
		return nil, fmt.Errorf(".mappings file not found at %s", mappingsPath)
	}

	var config Config
	if _, err := toml.DecodeFile(mappingsPath, &config.Profiles); err != nil {
		return nil, fmt.Errorf("failed to parse .mappings file: %w", err)
	}

	// Validate that [general] profile exists
	if config.Profiles == nil {
		config.Profiles = make(map[string]Profile)
	}

	if _, exists := config.Profiles["general"]; !exists {
		return nil, fmt.Errorf("[general] profile is required but not found in .mappings")
	}

	return &config, nil
}

// GetProfiles returns the profiles for the given profile names
// If no profiles are specified, returns [general] profile
// Later profiles override earlier ones when they map to the same target
func (c *Config) GetProfiles(profileNames []string) (Profile, error) {
	if len(profileNames) == 0 {
		profileNames = []string{"general"}
	}

	result := make(Profile)
	targetToSource := make(map[string]string) // track target -> source mapping for precedence

	// Start with [general] as base (lowest precedence)
	if general, exists := c.Profiles["general"]; exists {
		for src, target := range general {
			result[src] = target
			targetToSource[target] = src
		}
	}

	// Apply other profiles in order (last one wins for same target)
	for _, profileName := range profileNames {
		if profileName == "general" {
			continue // Already applied above
		}

		profile, exists := c.Profiles[profileName]
		if !exists {
			return nil, fmt.Errorf("profile [%s] not found in .mappings", profileName)
		}

		for src, target := range profile {
			// If this target already exists from a previous profile, remove the old mapping
			if oldSrc, exists := targetToSource[target]; exists {
				delete(result, oldSrc)
			}

			result[src] = target
			targetToSource[target] = src
		}
	}

	return result, nil
}
