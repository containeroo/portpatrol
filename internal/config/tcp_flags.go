package config

import (
	"fmt"
	"time"

	"github.com/containeroo/portpatrol/internal/checks"
)

const ParamTCPTimeout string = "timeout"

// parseTCPCheckerOptions parses TCP checker-specific options from parameters.
func parseTCPCheckerOptions(params map[string]string) ([]checks.Option, error) {
	var opts []checks.Option
	unrecognizedParams := trackUnusedParams(params)

	// TCP Timeout
	if timeoutStr, ok := params[ParamTCPTimeout]; ok && timeoutStr != "" {
		timeout, err := time.ParseDuration(timeoutStr)
		if err != nil {
			return nil, fmt.Errorf("invalid %q: %w", ParamTCPTimeout, err)
		}
		opts = append(opts, checks.WithTCPTimeout(timeout))
		delete(unrecognizedParams, ParamTCPTimeout)
	}

	// Check for unrecognized parameters
	if len(unrecognizedParams) > 0 {
		return nil, fmt.Errorf("unrecognized parameters for TCP checker: %v", extractMapKeys(unrecognizedParams))
	}

	return opts, nil
}