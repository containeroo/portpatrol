package parser

// trackUnrecognizedParams tracks parameters for validation.
func trackUnrecognizedParams(params map[string]string) map[string]struct{} {
	unrecognized := make(map[string]struct{})
	for key := range params {
		unrecognized[key] = struct{}{}
	}
	return unrecognized
}

// mapKeys extracts keys from a map for easier error reporting.
func mapKeys(m map[string]struct{}) []string {
	var keys []string
	for key := range m {
		keys = append(keys, key)
	}
	return keys
}
