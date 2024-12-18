package resolver

import (
	"fmt"
	"os"
)

// EnvResolver resolves values using environment variables.
// Usage: "env:MY_VAR" -> returns value of MY_VAR
type EnvResolver struct{}

func (r *EnvResolver) Resolve(value string) (string, error) {
	res, found := os.LookupEnv(value)
	if !found {
		return "", fmt.Errorf("environment variable '%s' not found", value)
	}
	return res, nil
}
