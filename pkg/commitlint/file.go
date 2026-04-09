package commitlint

import (
	"fmt"
	"os"
	"path/filepath"
)

// readFile reads the content of a file
func readFile(filePath string) (string, error) {
	content, err := os.ReadFile(filepath.Clean(filePath))
	if err != nil {
		return "", fmt.Errorf("reading %s: %w", filePath, err)
	}
	return string(content), nil
}
