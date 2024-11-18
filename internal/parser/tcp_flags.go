package parser

import (
	"fmt"
	"time"

	"github.com/containeroo/portpatrol/internal/checker"
)

const ParamTCPTimeout string = "timeout"

// parseTCPCheckerOptions parses TCP checker-specific options from parameters.
func parseTCPCheckerOptions(params map[string]string) ([]checker.Option, error) {
	var opts []checker.Option
	unrecognizedParams := trackUnrecognizedParams(params)

	// TCP Timeout
	if timeoutStr, ok := params[ParamTCPTimeout]; ok && timeoutStr != "" {
		timeout, err := time.ParseDuration(timeoutStr)
		if err != nil {
			return nil, fmt.Errorf("invalid %q: %w", ParamTCPTimeout, err)
		}
		opts = append(opts, checker.WithTCPTimeout(timeout))
		delete(unrecognizedParams, ParamTCPTimeout)
	}

	// Check for unrecognized parameters
	if len(unrecognizedParams) > 0 {
		return nil, fmt.Errorf("unrecognized parameters for TCP checker: %v", mapKeys(unrecognizedParams))
	}

	return opts, nil
}
