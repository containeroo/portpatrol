package config

import (
	"fmt"
	"time"

	"github.com/containeroo/portpatrol/internal/checks"
)

const (
	ParamICMPReadTimeout  string = "read-timeout"
	ParamICMPWriteTimeout string = "write-timeout"
)

// parseICMPCheckerOptions parses ICMP checker-specific options from parameters.
func parseICMPCheckerOptions(params map[string]string) ([]checks.Option, error) {
	var opts []checks.Option
	unrecognizedParams := trackUnusedParams(params)

	// ICMP Read Timeout
	if readTimeoutStr, ok := params[ParamICMPReadTimeout]; ok && readTimeoutStr != "" {
		readTimeout, err := time.ParseDuration(readTimeoutStr)
		if err != nil {
			return nil, fmt.Errorf("invalid %q: %w", ParamICMPReadTimeout, err)
		}
		opts = append(opts, checks.WithICMPReadTimeout(readTimeout))
		delete(unrecognizedParams, ParamICMPReadTimeout)
	}

	// ICMP Write Timeout
	if writeTimeoutStr, ok := params[ParamICMPWriteTimeout]; ok && writeTimeoutStr != "" {
		writeTimeout, err := time.ParseDuration(writeTimeoutStr)
		if err != nil {
			return nil, fmt.Errorf("invalid %q: %w", ParamICMPWriteTimeout, err)
		}
		opts = append(opts, checks.WithICMPWriteTimeout(writeTimeout))
		delete(unrecognizedParams, ParamICMPWriteTimeout)
	}

	// Check for unrecognized parameters
	if len(unrecognizedParams) > 0 {
		return nil, fmt.Errorf("unrecognized parameters for ICMP checker: %v", extractMapKeys(unrecognizedParams))
	}

	return opts, nil
}
