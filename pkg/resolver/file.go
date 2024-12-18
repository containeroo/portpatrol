package resolver

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strings"
)

// Resolves values from key = values files. The format can be:
// "file:/config/app.txt//Key"
// If no key is provided, returns the whole file.
type KeyValueFileResolver struct{}

func (f *KeyValueFileResolver) Resolve(value string) (string, error) {
	filePath, keyPath := splitFileAndKey(value)
	filePath = os.ExpandEnv(filePath)

	file, err := os.Open(filePath)
	if err != nil {
		return "", fmt.Errorf("Failed to open file '%s'. %v", filePath, err)
	}
	defer file.Close()

	if keyPath != "" {
		return searchKeyInFile(file, keyPath)
	}

	// No key specified, read the whole file
	data, err := io.ReadAll(file)
	if err != nil {
		return "", fmt.Errorf("Failed to read file '%s'. %v", filePath, err)
	}
	return strings.TrimSpace(string(data)), nil
}

// searchKeyInFile searches for a specified key in a file and returns its associated value.
func searchKeyInFile(file *os.File, key string) (string, error) {
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		pair := strings.SplitN(line, "=", 2)
		if len(pair) == 2 && strings.TrimSpace(pair[0]) == key {
			return strings.TrimSpace(pair[1]), nil
		}
	}
	return "", fmt.Errorf("Key '%s' not found in file '%s'.", key, file.Name())
}
