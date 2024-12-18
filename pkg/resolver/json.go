package resolver

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
)

// Resolves a value by loading a JSON file and extracting a nested key.
// The value after the prefix should be in the format "path/to/file.json//key1.key2.keyN"
// If no key is provided, returns the entire JSON file as a string.
// Example:
// "json:/config/app.json//server.host"
// would load app.json, parse it as JSON, and then return the value at server.host.
//
// Keys are navigated via dot notation.
// If no key is provided (no "//" present), returns the entire JSON file as string.
type JSONResolver struct{}

func (r *JSONResolver) Resolve(value string) (string, error) {
	filePath, keyPath := splitFileAndKey(value)
	filePath = os.ExpandEnv(filePath)

	data, err := os.ReadFile(filePath)
	if err != nil {
		return "", fmt.Errorf("failed to read JSON file '%s': %w", filePath, err)
	}

	if keyPath == "" {
		// Return whole file
		return strings.TrimSpace(string(data)), nil
	}

	var content map[string]interface{}
	if err := json.Unmarshal(data, &content); err != nil {
		return "", fmt.Errorf("failed to parse JSON in '%s': %w", filePath, err)
	}

	val, err := navigateData(content, strings.Split(keyPath, "."))
	if err != nil {
		return "", fmt.Errorf("key path '%s' not found in JSON '%s': %w", keyPath, filePath, err)
	}

	strVal, ok := val.(string)
	if !ok {
		// If the value isn't a string, return its JSON representation
		jData, _ := json.Marshal(val)
		return string(jData), nil
	}
	return strVal, nil
}
