package parser

import (
	"fmt"
	"time"

	"github.com/containeroo/portpatrol/internal/checker"
)

const ParamTCPTimeout string = "timeout"

// parseTCPCheckerOptions parses TCP checker specific options from parameters.
func parseTCPCheckerOptions(params map[string]string) ([]checker.Option, error) {
	var opts []checker.Option

	// Track unrecognized parameters
	unrecognizedParams := make(map[string]struct{})
	for key := range params {
		unrecognizedParams[key] = struct{}{}
	}

	// TCP Timeout
	if timeoutStr, ok := params[ParamTCPTimeout]; ok && timeoutStr != "" {
		t, err := time.ParseDuration(timeoutStr)
		if err != nil {
			return nil, fmt.Errorf("invalid %q: %w", ParamTCPTimeout, err)
		}
		opts = append(opts, checker.WithTCPTimeout(t))
		delete(unrecognizedParams, ParamTCPTimeout)
	}

	// Check for unrecognized parameters
	if len(unrecognizedParams) > 0 {
		var unknownKeys []string
		for key := range unrecognizedParams {
			unknownKeys = append(unknownKeys, key)
		}
		return nil, fmt.Errorf("unrecognized parameters for TCP checker: %v", unknownKeys)
	}

	return opts, nil
}
