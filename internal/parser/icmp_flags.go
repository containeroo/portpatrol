package parser

import (
	"fmt"
	"time"

	"github.com/containeroo/portpatrol/internal/checker"
)

const (
	ParamICMPReadTimeout  string = "read-timeout"
	ParamICMPWriteTimeout string = "write-timeout"
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
	if readTimeoutStr, ok := params[ParamICMPReadTimeout]; ok && readTimeoutStr != "" {
		rt, err := time.ParseDuration(readTimeoutStr)
		if err != nil {
			return nil, fmt.Errorf("invalid %q: %w", ParamICMPReadTimeout, err)
		}
		opts = append(opts, checker.WithICMPReadTimeout(rt))
		delete(unrecognizedParams, ParamICMPReadTimeout)
	}

	// ICMP Write Timeout
	if writeTimeoutStr, ok := params[ParamICMPWriteTimeout]; ok && writeTimeoutStr != "" {
		wt, err := time.ParseDuration(writeTimeoutStr)
		if err != nil {
			return nil, fmt.Errorf("invalid %q: %w", ParamICMPWriteTimeout, err)
		}
		opts = append(opts, checker.WithICMPWriteTimeout(wt))
		delete(unrecognizedParams, ParamICMPWriteTimeout)
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
