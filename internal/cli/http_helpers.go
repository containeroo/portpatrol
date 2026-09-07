package cli

import (
	"fmt"
	"net/http"

	"github.com/containeroo/resolver"
)

const httpUserAgentPrefix string = "never"

// resolveHTTPHeaderValues resolves configured variables in parsed HTTP header values.
func resolveHTTPHeaderValues(headers http.Header) error {
	for _, values := range headers {
		for i, value := range values {
			resolved, err := resolver.ResolveVariable(value)
			if err != nil {
				return fmt.Errorf("failed to resolve variable in header: %w", err)
			}
			values[i] = resolved
		}
	}

	return nil
}
