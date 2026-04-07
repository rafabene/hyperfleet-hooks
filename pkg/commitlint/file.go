package commitlint

import (
	"os"
)

// readFile reads the content of a file
func readFile(filePath string) (string, error) {
	content, err := os.ReadFile(filePath)
	if err != nil {
		return "", err
	}
	return string(content), nil
}
