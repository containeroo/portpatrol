package checker

import (
	"fmt"
	"strings"
)

// extractScheme extracts the scheme from the address if it exists.
// If the address does not have a scheme, it returns an empty string.
func extractScheme(address string) (string, error) {
	parts := strings.SplitN(address, "://", 2)
	if len(parts) != 2 {
		return "", fmt.Errorf("no scheme found in address: %s", address)
	}

	return parts[0], nil
}
