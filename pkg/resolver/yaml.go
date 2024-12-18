package resolver

import (
	"fmt"
	"os"
	"strings"

	"gopkg.in/yaml.v3"
)

// Resolves a value by loading a YAML file and extracting a nested key using dot notation.
// Similar usage as JSONResolver:
// "yaml:/config/app.yaml//server.host"
type YAMLResolver struct{}

func (r *YAMLResolver) Resolve(value string) (string, error) {
	filePath, keyPath := splitFileAndKey(value)
	filePath = os.ExpandEnv(filePath)

	data, err := os.ReadFile(filePath)
	if err != nil {
		return "", fmt.Errorf("failed to read YAML file '%s': %w", filePath, err)
	}

	var content interface{}
	if err := yaml.Unmarshal(data, &content); err != nil {
		return "", fmt.Errorf("failed to parse YAML in '%s': %w", filePath, err)
	}

	// Convert YAML to map[string]interface{} if needed
	contentMap, err := convertToMapStringInterface(content)
	if err != nil {
		return "", fmt.Errorf("failed to process YAML '%s': %w", filePath, err)
	}

	if keyPath == "" {
		// Return whole file as YAML string
		return strings.TrimSpace(string(data)), nil
	}

	val, err := navigateData(contentMap, strings.Split(keyPath, "."))
	if err != nil {
		return "", fmt.Errorf("key path '%s' not found in YAML '%s': %w", keyPath, filePath, err)
	}

	// If the value isn't a string, return its YAML representation
	switch typedVal := val.(type) {
	case string:
		return typedVal, nil
	default:
		yData, _ := yaml.Marshal(typedVal)
		return strings.TrimSpace(string(yData)), nil
	}
}
