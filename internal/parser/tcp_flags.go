package parser

import (
	"fmt"
	"time"

	"github.com/containeroo/portpatrol/internal/checker"
	"github.com/containeroo/portpatrol/internal/flags"
)

const ParamTCPTimeout string = "timeout"

var tcpFlagDocs = []flags.FlagDoc{
	{
		Flag:        fmt.Sprintf("--%s.<identifier>.%s=string", ParamPrefix, ParamAddress),
		Description: "The IP address or hostname of the target in the following format: tcp://hostname:port",
	},
	{
		Flag:        fmt.Sprintf("--%s.<identifier>.%s=string", ParamPrefix, ParamName),
		Description: "The name of the target. If not specified, it's derived from the target address.",
	},
	{
		Flag:        fmt.Sprintf("--%s.<identifier>.%s=string", ParamPrefix, ParamType),
		Description: "The type of check to perform. If the scheme (tcp://) is specified in --%s.<identifier>.address, this flag can be omitted as the type will be inferred.",
	},
	{
		Flag:        fmt.Sprintf("--%s.<identifier>.%s=duration", ParamPrefix, ParamInterval),
		Description: "Override the default interval for this target (e.g., 5s).",
	},
	{
		Flag:        fmt.Sprintf("--%s.<identifier>.%s=duration", ParamPrefix, ParamTCPTimeout),
		Description: "The timeout for the TCP request (e.g., 5s).",
	},
}

// parseTCPCheckerOptions parses TCP checker-specific options from parameters.
func parseTCPCheckerOptions(params map[string]string) ([]checker.Option, error) {
	var opts []checker.Option
	unrecognizedParams := trackUnusedParams(params)

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
		return nil, fmt.Errorf("unrecognized parameters for TCP checker: %v", extractMapKeys(unrecognizedParams))
	}

	return opts, nil
}
