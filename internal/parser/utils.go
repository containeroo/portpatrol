package parser

import "github.com/containeroo/portpatrol/internal/flags"

// trackUnusedParams tracks parameters for validation.
func trackUnusedParams(params map[string]string) map[string]struct{} {
	unrecognized := make(map[string]struct{})
	for key := range params {
		unrecognized[key] = struct{}{}
	}
	return unrecognized
}

// extractMapKeys extracts keys from a map for easier error reporting.
func extractMapKeys(m map[string]struct{}) []string {
	var keys []string
	for key := range m {
		keys = append(keys, key)
	}
	return keys
}

func GenerateDocs() map[string][]flags.FlagDoc {
	return map[string][]flags.FlagDoc{
		"TCP Checker Properties":  tcpFlagDocs,
		"HTTP Checker Properties": httpFlagDocs,
		"ICMP Checker Properties": icmpFlagDocs,
	}
}
