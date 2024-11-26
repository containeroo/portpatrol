package parser

import (
	"fmt"
	"time"

	"github.com/containeroo/portpatrol/internal/checker"
	"github.com/containeroo/portpatrol/internal/flags"
)

const (
	ParamICMPReadTimeout  string = "read-timeout"
	ParamICMPWriteTimeout string = "write-timeout"
)

var icmpFlagDocs = []flags.FlagDoc{
	{
		Flag:        fmt.Sprintf("--%s.<identifier>.%s=string", ParamPrefix, ParamAddress),
		Description: "The IP address or hostname of the target in the following format: icmp://hostname (no port allowed).",
	},
	{
		Flag:        fmt.Sprintf("--%s.<identifier>.%s=string", ParamPrefix, ParamName),
		Description: "The name of the target. If not specified, it's derived from the target address.",
	},
	{
		Flag:        fmt.Sprintf("--%s.<identifier>.%s=string", ParamPrefix, ParamType),
		Description: "The type of check to perform. If the scheme (icmp://) is specified in --%s.<identifier>.address, this flag can be omitted as the type will be inferred.",
	},
	{
		Flag:        fmt.Sprintf("--%s.<identifier>.%s=duration", ParamPrefix, ParamICMPReadTimeout),
		Description: "The read timeout for the ICMP connection (e.g., 1s).",
	},
	{
		Flag:        fmt.Sprintf("--%s.<identifier>.%s=duration", ParamPrefix, ParamICMPWriteTimeout),
		Description: "The write timeout for the ICMP connection (e.g., 1s).",
	},
	{
		Flag:        fmt.Sprintf("--%s.<identifier>.%s=duration", ParamPrefix, ParamInterval),
		Description: "Override the default interval for this target (e.g., 5s).",
	},
}

// parseICMPCheckerOptions parses ICMP checker-specific options from parameters.
func parseICMPCheckerOptions(params map[string]string) ([]checker.Option, error) {
	var opts []checker.Option
	unrecognizedParams := trackUnusedParams(params)

	// ICMP Read Timeout
	if readTimeoutStr, ok := params[ParamICMPReadTimeout]; ok && readTimeoutStr != "" {
		readTimeout, err := time.ParseDuration(readTimeoutStr)
		if err != nil {
			return nil, fmt.Errorf("invalid %q: %w", ParamICMPReadTimeout, err)
		}
		opts = append(opts, checker.WithICMPReadTimeout(readTimeout))
		delete(unrecognizedParams, ParamICMPReadTimeout)
	}

	// ICMP Write Timeout
	if writeTimeoutStr, ok := params[ParamICMPWriteTimeout]; ok && writeTimeoutStr != "" {
		writeTimeout, err := time.ParseDuration(writeTimeoutStr)
		if err != nil {
			return nil, fmt.Errorf("invalid %q: %w", ParamICMPWriteTimeout, err)
		}
		opts = append(opts, checker.WithICMPWriteTimeout(writeTimeout))
		delete(unrecognizedParams, ParamICMPWriteTimeout)
	}

	// Check for unrecognized parameters
	if len(unrecognizedParams) > 0 {
		return nil, fmt.Errorf("unrecognized parameters for ICMP checker: %v", extractMapKeys(unrecognizedParams))
	}

	return opts, nil
}
