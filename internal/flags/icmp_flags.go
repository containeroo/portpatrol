package flags

import (
	"fmt"
	"time"

	"github.com/containeroo/portpatrol/internal/checker"
)

const (
	paramICMPReadTimeout  string = "read-timeout"
	paramICMPWriteTimeout string = "write-timeout"
)

// parseICMPCheckerOptions parses ICMP checker specific options from parameters.
func parseICMPCheckerOptions(params map[string]string) ([]checker.Option, error) {
	var opts []checker.Option

	// Track unrecognized parameters
	unrecognizedParams := make(map[string]struct{})
	for key := range params {
		unrecognizedParams[key] = struct{}{}
	}

	// ICMP Read Timeout
	if readTimeoutStr, ok := params[paramICMPReadTimeout]; ok && readTimeoutStr != "" {
		rt, err := time.ParseDuration(readTimeoutStr)
		if err != nil {
			return nil, fmt.Errorf("invalid %q: %w", paramICMPReadTimeout, err)
		}
		opts = append(opts, checker.WithICMPReadTimeout(rt))
		delete(unrecognizedParams, paramICMPReadTimeout)
	}

	// ICMP Write Timeout
	if writeTimeoutStr, ok := params[paramICMPWriteTimeout]; ok && writeTimeoutStr != "" {
		wt, err := time.ParseDuration(writeTimeoutStr)
		if err != nil {
			return nil, fmt.Errorf("invalid %q: %w", paramICMPWriteTimeout, err)
		}
		opts = append(opts, checker.WithICMPWriteTimeout(wt))
		delete(unrecognizedParams, paramICMPWriteTimeout)
	}

	// Check for unrecognized parameters
	if len(unrecognizedParams) > 0 {
		var unknownKeys []string
		for key := range unrecognizedParams {
			unknownKeys = append(unknownKeys, key)
		}
		return nil, fmt.Errorf("unrecognized parameters for ICMP checker: %v", unknownKeys)
	}

	return opts, nil
}
