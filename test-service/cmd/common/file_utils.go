package common

import (
	"fmt"
	"os"
	"strings"
)

// FileExists checks if a file exists
func FileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

// ReadFileContent reads and returns the content of a file
func ReadFileContent(path string) (string, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return "", fmt.Errorf("failed to read file: %w", err)
	}
	return string(content), nil
}

// EnsureDir creates a directory if it doesn't exist
func EnsureDir(path string) error {
	if err := os.MkdirAll(path, 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}
	return nil
}

// CheckFileHasContent checks if a file contains specific content
func CheckFileHasContent(filePath, content string) (bool, error) {
	fileContent, err := ReadFileContent(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}
		return false, err
	}
	return strings.Contains(fileContent, content), nil
}

// CheckFileHasAllStrings checks if a file contains all specified strings
func CheckFileHasAllStrings(filePath string, searchStrings []string) (bool, error) {
	if !FileExists(filePath) {
		return false, nil
	}

	fileContent, err := ReadFileContent(filePath)
	if err != nil {
		return false, err
	}

	for _, searchStr := range searchStrings {
		if !strings.Contains(fileContent, searchStr) {
			return false, nil
		}
	}

	return true, nil
}

// FindMissingStrings returns which strings are missing from the file
func FindMissingStrings(filePath string, searchStrings []string) ([]string, error) {
	if !FileExists(filePath) {
		return searchStrings, nil
	}

	fileContent, err := ReadFileContent(filePath)
	if err != nil {
		return nil, err
	}

	var missing []string
	for _, searchStr := range searchStrings {
		if !strings.Contains(fileContent, searchStr) {
			missing = append(missing, searchStr)
		}
	}

	return missing, nil
}

// DeleteFileIfExists deletes a file if it exists, returns nil if file doesn't exist
func DeleteFileIfExists(path string) error {
	if !FileExists(path) {
		return nil
	}

	if err := os.Remove(path); err != nil {
		return fmt.Errorf("failed to delete file: %w", err)
	}

	return nil
}

// CopyFile copies a file from src to dst
func CopyFile(src, dst string) error {
	content, err := os.ReadFile(src)
	if err != nil {
		return fmt.Errorf("failed to read source file: %w", err)
	}

	if err := os.WriteFile(dst, content, 0644); err != nil {
		return fmt.Errorf("failed to write destination file: %w", err)
	}

	return nil
}
