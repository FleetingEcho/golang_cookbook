package utils

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/google/uuid"
)

func SafeFileName(name string) string {
	return strings.Map(func(r rune) rune {
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') ||
			(r >= '0' && r <= '9') || r == '.' || r == '-' || r == '_' {
			return r
		}
		return '_'
	}, name)
}

func StoredFileName(original string) string {
	return fmt.Sprintf("%s-%s", uuid.NewString(), SafeFileName(original))
}

func EnsureInsideUploadDir(uploadDir, storedName string) (string, error) {
	if strings.Contains(storedName, "/") || strings.Contains(storedName, "\\") || strings.Contains(storedName, "..") {
		return "", fmt.Errorf("invalid file name")
	}
	full := filepath.Join(uploadDir, storedName)
	absDir, _ := filepath.Abs(uploadDir)
	absFull, _ := filepath.Abs(full)
	if !strings.HasPrefix(absFull, absDir) {
		return "", fmt.Errorf("path traversal detected")
	}
	return full, nil
}
