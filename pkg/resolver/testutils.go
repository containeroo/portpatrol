package resolver

import "regexp"

func sanitizeEnvName(name string) string {
	re := regexp.MustCompile(`[^a-zA-Z0-9_]`)
	return re.ReplaceAllString(name, "_")
}
