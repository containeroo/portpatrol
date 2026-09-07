package checker

import "strings"

// normalizeAddress trims an already-resolved and prevalidated target address.
func normalizeAddress(address string) string {
	return strings.TrimSpace(address)
}
