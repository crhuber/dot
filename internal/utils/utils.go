package utils

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// ExpandPath expands ~ to the user's home directory
func ExpandPath(path string) string {
	if !strings.HasPrefix(path, "~") {
		return path
	}

	homeDir, err := os.UserHomeDir()
	if err != nil {
		// If we can't get home directory, return path as-is
		return path
	}

	if path == "~" {
		return homeDir
	}

	if strings.HasPrefix(path, "~/") {
		return filepath.Join(homeDir, path[2:])
	}

	return path
}

// BackupFile creates a backup of a file or directory by adding .bak suffix
// Overwrites existing .bak file if present
func BackupFile(path string) error {
	backupPath := path + ".bak"

	// Remove existing backup if it exists
	if _, err := os.Stat(backupPath); err == nil {
		if err := os.RemoveAll(backupPath); err != nil {
			return fmt.Errorf("failed to remove existing backup %s: %w", backupPath, err)
		}
	}

	// Create backup by renaming
	if err := os.Rename(path, backupPath); err != nil {
		return fmt.Errorf("failed to create backup %s: %w", backupPath, err)
	}

	return nil
}

// IsSymlink checks if a path is a symbolic link
func IsSymlink(path string) (bool, error) {
	stat, err := os.Lstat(path)
	if err != nil {
		return false, err
	}

	return stat.Mode()&os.ModeSymlink != 0, nil
}

// ReadSymlink safely reads a symbolic link target
func ReadSymlink(path string) (string, error) {
	isLink, err := IsSymlink(path)
	if err != nil {
		return "", err
	}

	if !isLink {
		return "", fmt.Errorf("%s is not a symbolic link", path)
	}

	return os.Readlink(path)
}

// FileExists checks if a file or directory exists
func FileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

// LogInfo writes an informational message to stdout
func LogInfo(format string, args ...interface{}) {
	fmt.Printf(format+"\n", args...)
}

// LogError writes an error message to stderr
func LogError(format string, args ...interface{}) {
	fmt.Fprintf(os.Stderr, format+"\n", args...)
}

// LogWarning writes a warning message to stderr
func LogWarning(format string, args ...interface{}) {
	fmt.Fprintf(os.Stderr, "Warning: "+format+"\n", args...)
}

// Color constants
const (
	Reset  = "\033[0m"
	Red    = "\033[31m"
	Green  = "\033[32m"
	Yellow = "\033[33m"
	Blue   = "\033[34m"
	Purple = "\033[35m"
	Cyan   = "\033[36m"
	Gray   = "\033[37m"
	White  = "\033[97m"
)

// PrintLn prints text with color
func PrintLn(text string, colorChoice string) {
	switch colorChoice {
	case "red":
		fmt.Println(Red + text + Reset)
	case "green":
		fmt.Println(Green + text + Reset)
	case "yellow":
		fmt.Println(Yellow + text + Reset)
	case "blue":
		fmt.Println(Blue + text + Reset)
	case "gray":
		fmt.Println(Gray + text + Reset)
	default:
		fmt.Println(White + text + Reset)
	}
}

// PrintfColor prints formatted text with color
func PrintfColor(colorChoice string, format string, args ...interface{}) {
	var color string
	switch colorChoice {
	case "red":
		color = Red
	case "green":
		color = Green
	case "yellow":
		color = Yellow
	case "blue":
		color = Blue
	case "gray":
		color = Gray
	default:
		color = White
	}
	fmt.Printf(color+format+Reset, args...)
}

// FprintfColor prints formatted text with color to a specific writer
func FprintfColor(writer *os.File, colorChoice string, format string, args ...interface{}) {
	var color string
	switch colorChoice {
	case "red":
		color = Red
	case "green":
		color = Green
	case "yellow":
		color = Yellow
	case "blue":
		color = Blue
	case "gray":
		color = Gray
	default:
		color = White
	}
	fmt.Fprintf(writer, color+format+Reset, args...)
}
