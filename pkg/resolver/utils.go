package resolver

import (
	"fmt"
	"strconv"
	"strings"
)

// splitFileAndKey splits a value by "//" to separate file path and key path.
func splitFileAndKey(value string) (string, string) {
	const keyDelim = "//"
	idx := strings.LastIndex(value, keyDelim)
	if idx == -1 {
		return value, ""
	}
	return value[:idx], value[idx+len(keyDelim):]
}

// navigateData walks through a nested structure (maps and arrays) using a slice of keys.
// Keys can be:
// - Map keys: e.g. "server", "host"
// - Array indices: e.g. "0", "1"
//
// Example:
// For JSON or YAML structures, "servers.0.host" means:
//   - Look up the "servers" field in a map
//   - Expect that value to be a slice (array)
//   - Take the 0th element of that slice
//   - Expect that element to be a map with a "host" key
//   - Return the value at "host"
func navigateData(data interface{}, keys []string) (interface{}, error) {
	current := data
	for _, k := range keys {
		switch curr := current.(type) {
		case map[string]interface{}:
			// Current data is a map, so k is a field name
			val, ok := curr[k]
			if !ok {
				return nil, fmt.Errorf("key '%s' not found", k)
			}
			current = val

		case []interface{}:
			// Current data is a slice, so k should be a numeric index
			idx, err := strconv.Atoi(k)
			if err != nil {
				return nil, fmt.Errorf("'%s' is not a valid array index", k)
			}
			if idx < 0 || idx >= len(curr) {
				return nil, fmt.Errorf("array index '%d' out of bounds", idx)
			}
			current = curr[idx]

		default:
			// We've reached a non-map, non-array type but still have keys
			return nil, fmt.Errorf("path segment '%s' not found", k)
		}
	}
	return current, nil
}

// convertToMapStringInterface attempts to convert arbitrary YAML-parsed data into a map[string]interface{} for uniform handling.
// It recursively ensures arrays and maps are properly converted.
func convertToMapStringInterface(val interface{}) (map[string]interface{}, error) {
	switch v := val.(type) {
	case map[string]interface{}:
		// Recursively convert values
		for key, val2 := range v {
			converted, err := convertValue(val2)
			if err != nil {
				return nil, err
			}
			v[key] = converted
		}
		return v, nil
	default:
		// If it's not a map at the root level, return an empty map
		return map[string]interface{}{}, nil
	}
}

func convertValue(val interface{}) (interface{}, error) {
	// Recursively convert slices and maps
	switch vv := val.(type) {
	case map[string]interface{}:
		for k, v := range vv {
			converted, err := convertValue(v)
			if err != nil {
				return nil, err
			}
			vv[k] = converted
		}
		return vv, nil
	case []interface{}:
		for i, elem := range vv {
			converted, err := convertValue(elem)
			if err != nil {
				return nil, err
			}
			vv[i] = converted
		}
		return vv, nil
	default:
		return vv, nil
	}
}
