// flags/flags.go

package flags

import (
	"bytes"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/containeroo/portpatrol/internal/checker"
	"github.com/containeroo/portpatrol/pkg/httputils"
	"github.com/spf13/pflag"
)

const (
	// Options
	defaultDebug         bool          = false
	paramDefaultInterval string        = "default-interval"
	defaultCheckInterval time.Duration = 2 * time.Second

	paramPrefix   string = "target"
	paramType     string = "type"
	paramName     string = "name"
	paramAddress  string = "address"
	paramInterval string = "interval"

	// HTTPChecker parameter keys
	paramHTTPMethod                  string = "method"
	paramHTTPHeaders                 string = "headers"
	paramHTTPAllowDuplicateHeaders   string = "allow-duplicate-headers"
	paramHTTPExpectedStatusCodes     string = "expected-status-codes"
	paramHTTPSkipTLSVerify           string = "skip-tls-verify"
	paramHTTPTimeout                 string = "timeout"
	defaultHTTPAllowDuplicateHeaders bool   = false
	defaultHTTPSkipTLSVerify         bool   = false

	// TCPChecker parameter keys
	paramTCPTimeout string = "timeout"

	// ICMPChecker parameter keys
	paramICMPReadTimeout  string = "read-timeout"
	paramICMPWriteTimeout string = "write-timeout"
)

type TargetChecker struct {
	Interval time.Duration
	Checker  checker.Checker
}

type ParsedFlags struct {
	ShowHelp             bool
	ShowVersion          bool
	Version              string
	DefaultCheckInterval time.Duration
	Debug                bool
	Targets              map[string]map[string]string
}

// ParseFlags parses command line arguments and returns the parsed flags.
func ParseFlags(args []string, version string) (*ParsedFlags, error) {
	var knownArgs []string
	var dynamicArgs []string

	// Preprocess arguments to extract dynamic target flags
	for i := 0; i < len(args); i++ {
		arg := args[i]
		if !strings.HasPrefix(arg, fmt.Sprintf("--%s.", paramPrefix)) {
			knownArgs = append(knownArgs, arg)
			continue
		}

		dynamicArgs = append(dynamicArgs, arg)
		if !strings.Contains(arg, "=") && i+1 < len(args) && !strings.HasPrefix(args[i+1], "-") {
			i++ // Use next argument as value
			dynamicArgs = append(dynamicArgs, args[i])
		}
	}

	flagSetName := "portpatrol"
	flagSet := pflag.NewFlagSet(flagSetName, pflag.ContinueOnError)
	flagSet.SortFlags = false

	// Set a buffer to capture help messages or error output
	var buf bytes.Buffer
	flagSet.SetOutput(&buf)

	// Custom usage function to display help information
	flagSet.Usage = func() {
		fmt.Fprintf(&buf, "Usage: %s [OPTIONS] [--%s.<IDENTIFIER>.<PROPERTY>=value]\n\nOptions:\n", flagSetName, paramPrefix)
		flagSet.PrintDefaults()
		displayCheckerProperties(&buf)
	}

	// Define known flags
	checkInterval := flagSet.Duration(paramDefaultInterval, defaultCheckInterval, "Default interval between checks. Can be overwritten for each target.")
	logExtraFields := flagSet.Bool("debug", defaultDebug, "Log extra fields.")
	showVersion := flagSet.Bool("version", false, "Show version and exit.")
	showHelp := flagSet.BoolP("help", "h", false, "Show help.")

	// Parse known flags
	if err := flagSet.Parse(knownArgs); err != nil {
		buf.WriteString(err.Error())
		buf.WriteString("\n\n")
		flagSet.Usage()
		return nil, errors.New(buf.String())
	}

	// Handle help request
	if *showHelp {
		flagSet.Usage()
		return nil, errors.New(buf.String())
	}

	// Handle version request
	if *showVersion {
		return nil, fmt.Errorf("%s version %s", flagSetName, version)
	}

	// Process the dynamic target flags
	targets, err := processDynamicArgs(dynamicArgs, buf)
	if err != nil {
		return nil, err
	}

	// Check for unknown arguments
	for _, arg := range flagSet.Args() {
		if strings.HasPrefix(arg, "-") && !strings.HasPrefix(arg, fmt.Sprintf("--%s.", paramPrefix)) {
			return nil, fmt.Errorf("Warning: Unknown flag ignored: %s\n", arg)
		}
	}

	return &ParsedFlags{
		DefaultCheckInterval: *checkInterval,
		Debug:                *logExtraFields,
		Targets:              targets,
	}, nil
}

// displayCheckerProperties appends checker properties documentation to the buffer.
func displayCheckerProperties(buf *bytes.Buffer) {
	// Display TCP Checker properties
	buf.WriteString(fmt.Sprintf("  --%s.<IDENTIFIER>.%s: The type of check to perform. Must be \"tcp\".\n", paramPrefix, paramType))
	buf.WriteString(fmt.Sprintf("  --%s.<IDENTIFIER>.%s: The name of the target. If not specified, it's derived from the target address.\n", paramPrefix, paramName))
	buf.WriteString(fmt.Sprintf("  --%s.<IDENTIFIER>.%s: The IP address or hostname of the target in the following format: tcp://hostname:port\n", paramPrefix, paramAddress))
	buf.WriteString(fmt.Sprintf("  --%s.<IDENTIFIER>.%s: The timeout for the TCP connection.\n", paramPrefix, paramTCPTimeout))
	buf.WriteString(fmt.Sprintf("  --%s.<IDENTIFIER>.%s: (Optional) Override the default interval for this target.\n", paramPrefix, paramInterval))

	// Display HTTP Checker properties
	buf.WriteString("\nHTTP Checker Properties:\n")
	buf.WriteString(fmt.Sprintf("  --%s.<IDENTIFIER>.%s: The type of check to perform. Must be \"http\" or \"https\".\n", paramPrefix, paramType))
	buf.WriteString(fmt.Sprintf("  --%s.<IDENTIFIER>.%s: The name of the target. If not specified, it's derived from the target address.\n", paramPrefix, paramName))
	buf.WriteString(fmt.Sprintf("  --%s.<IDENTIFIER>.%s: The IP address or hostname of the target in the following format: http://hostname[:port]\n", paramPrefix, paramAddress))
	buf.WriteString(fmt.Sprintf("  --%s.<IDENTIFIER>.%s: The HTTP method to use. Defaults to \"GET\".\n", paramPrefix, paramHTTPMethod))
	buf.WriteString(fmt.Sprintf("  --%s.<IDENTIFIER>.%s: A comma-separated list of HTTP headers to include in the request.\n", paramPrefix, paramHTTPHeaders))
	buf.WriteString(fmt.Sprintf("  --%s.<IDENTIFIER>.%s: A comma-separated list of expected status codes. Defaults to 200.\n", paramPrefix, paramHTTPExpectedStatusCodes))
	buf.WriteString(fmt.Sprintf("  --%s.<IDENTIFIER>.%s: Whether to skip TLS verification. Defaults to false.\n", paramPrefix, paramHTTPSkipTLSVerify))
	buf.WriteString(fmt.Sprintf("  --%s.<IDENTIFIER>.%s: The timeout for the HTTP request.\n", paramPrefix, paramHTTPTimeout))
	buf.WriteString(fmt.Sprintf("  --%s.<IDENTIFIER>.%s: (Optional) Override the default interval for this target.\n", paramPrefix, paramInterval))

	// Display ICMP Checker properties
	buf.WriteString("\nICMP Checker Properties:\n")
	buf.WriteString(fmt.Sprintf("  --%s.<IDENTIFIER>.%s: The type of check to perform. Must be \"icmp\".\n", paramPrefix, paramType))
	buf.WriteString(fmt.Sprintf("  --%s.<IDENTIFIER>.%s: The name of the target. If not specified, it's derived from the target address.\n", paramPrefix, paramName))
	buf.WriteString(fmt.Sprintf("  --%s.<IDENTIFIER>.%s: The IP address or hostname of the target in the following format: icmp://hostname\n", paramPrefix, paramAddress))
	buf.WriteString(fmt.Sprintf("  --%s.<IDENTIFIER>.%s: The read timeout for the ICMP connection.\n", paramPrefix, paramICMPReadTimeout))
	buf.WriteString(fmt.Sprintf("  --%s.<IDENTIFIER>.%s: The write timeout for the ICMP connection.\n", paramPrefix, paramICMPWriteTimeout))
	buf.WriteString(fmt.Sprintf("  --%s.<IDENTIFIER>.%s: (Optional) Override the default interval for this target.\n", paramPrefix, paramInterval))
}

// processDynamicArgs processes dynamic target flags and returns target configurations.
func processDynamicArgs(dynamicArgs []string, buf bytes.Buffer) (map[string]map[string]string, error) {
	// Map to store target configurations
	targets := make(map[string]map[string]string)

	// Process the dynamic target flags
	for i := 0; i < len(dynamicArgs); i++ {
		arg := dynamicArgs[i]

		// Extract the flag name by removing the "--" prefix
		flagName := strings.TrimPrefix(arg, "--")

		var value string
		if strings.Contains(flagName, "=") {
			// Handle "--target.name.param=value" format
			parts := strings.SplitN(flagName, "=", 2)
			flagName = parts[0]
			value = parts[1]
		} else if i+1 < len(dynamicArgs) && !strings.HasPrefix(dynamicArgs[i+1], "--") {
			// Handle "--target.name.param value" format
			value = dynamicArgs[i+1]
			i++ // Skip the value in the next iteration
		} else {
			return nil, fmt.Errorf("missing value for flag: %s\n\n%s", arg, buf.String())
		}

		// Split the flag name to extract target name and parameter
		nameParts := strings.Split(flagName, ".")
		if len(nameParts) < 3 {
			return nil, fmt.Errorf("invalid target flag format: %s\n\n%s", flagName, buf.String())
		}

		// Extract the target name and parameter
		targetName := nameParts[1]                    // e.g., "postgres"
		parameter := strings.Join(nameParts[2:], ".") // e.g., "address"

		// Initialize the target map if necessary
		if _, exists := targets[targetName]; !exists {
			targets[targetName] = make(map[string]string)
		}

		// Store the parameter value
		targets[targetName][parameter] = value
	}

	return targets, nil
}

// ParseTargets creates a slice of TargetChecker based on the provided target configurations.
func ParseTargets(targets map[string]map[string]string, defaultInterval time.Duration) ([]TargetChecker, error) {
	var targetCheckers []TargetChecker

	for targetName, params := range targets {
		address, ok := params[paramAddress]
		if !ok || address == "" {
			return nil, fmt.Errorf("missing %q for target %q", paramAddress, targetName)
		}

		// Determine the check type
		checkTypeStr, ok := params[paramType]
		if !ok || checkTypeStr == "" {
			// Try to infer the type from the address scheme
			address := params[paramAddress]
			parts := strings.SplitN(address, "://", 2)
			if len(parts) != 2 {
				return nil, fmt.Errorf("missing %q parameter for target %q", paramType, targetName)
			}
			checkTypeStr = parts[0]
		}

		checkType, err := checker.GetCheckTypeFromString(checkTypeStr)
		if err != nil {
			return nil, fmt.Errorf("unsupported check type %q for target %q", checkTypeStr, targetName)
		}

		name := targetName
		if n, ok := params[paramName]; ok && n != "" {
			name = n
		}

		// Get interval from parameters or use default
		interval := defaultInterval
		if intervalStr, ok := params[paramInterval]; ok && intervalStr != "" {
			interval, err = time.ParseDuration(intervalStr)
			if err != nil {
				return nil, fmt.Errorf("invalid %q for target '%s': %w", paramInterval, targetName, err)
			}
		}

		// Remove common parameters from params map
		delete(params, paramType)
		delete(params, paramName)
		delete(params, paramAddress)
		delete(params, paramInterval)

		// Collect functional options based on the check type
		var options []checker.Option
		switch checkType {
		case checker.HTTP:
			httpOpts, err := parseHTTPCheckerOptions(params)
			if err != nil {
				return nil, fmt.Errorf("failed to parse HTTP options for target %q: %w", targetName, err)
			}
			options = append(options, httpOpts...)
		case checker.TCP:
			tcpOpts, err := parseTCPCheckerOptions(params)
			if err != nil {
				return nil, fmt.Errorf("failed to parse TCP options for target '%s': %w", targetName, err)
			}
			options = append(options, tcpOpts...)
		case checker.ICMP:
			icmpOpts, err := parseICMPCheckerOptions(params)
			if err != nil {
				return nil, fmt.Errorf("failed to parse ICMP options for target %q: %w", targetName, err)
			}
			options = append(options, icmpOpts...)
		default:
			return nil, fmt.Errorf("unsupported check type %q for target %q", checkTypeStr, targetName)
		}

		// Create the checker using the functional options
		chk, err := checker.NewChecker(checkType, name, address, options...)
		if err != nil {
			return nil, fmt.Errorf("failed to create checker for target %q: %w", targetName, err)
		}

		targetCheckers = append(targetCheckers, TargetChecker{
			Interval: interval,
			Checker:  chk,
		})
	}

	return targetCheckers, nil
}

// parseHTTPCheckerOptions parses HTTP checker specific options from parameters.
func parseHTTPCheckerOptions(params map[string]string) ([]checker.Option, error) {
	var opts []checker.Option

	// Create a copy of params to track unrecognized parameters
	unrecognizedParams := make(map[string]struct{})
	for key := range params {
		unrecognizedParams[key] = struct{}{}
	}

	if method, ok := params[paramHTTPMethod]; ok && method != "" {
		opts = append(opts, checker.WithHTTPMethod(method))
		delete(unrecognizedParams, paramHTTPMethod)
	}

	allowDupHeaders := defaultHTTPAllowDuplicateHeaders
	if allowDupHeadersStr, ok := params[paramHTTPAllowDuplicateHeaders]; ok && allowDupHeadersStr != "" {
		var err error
		allowDupHeaders, err = strconv.ParseBool(allowDupHeadersStr)
		if err != nil {
			return nil, fmt.Errorf("invalid %q: %w", paramHTTPAllowDuplicateHeaders, err)
		}
		delete(unrecognizedParams, paramHTTPAllowDuplicateHeaders)
	}

	if headersStr, ok := params[paramHTTPHeaders]; ok && headersStr != "" {
		headers, err := httputils.ParseHeaders(headersStr, allowDupHeaders)
		if err != nil {
			return nil, fmt.Errorf("invalid %q: %w", paramHTTPHeaders, err)
		}
		opts = append(opts, checker.WithHTTPHeaders(headers))
		delete(unrecognizedParams, paramHTTPHeaders)
	}

	if codesStr, ok := params[paramHTTPExpectedStatusCodes]; ok && codesStr != "" {
		codes, err := httputils.ParseStatusCodes(codesStr)
		if err != nil {
			return nil, fmt.Errorf("invalid %q: %w", paramHTTPExpectedStatusCodes, err)
		}
		opts = append(opts, checker.WithExpectedStatusCodes(codes))
		delete(unrecognizedParams, paramHTTPExpectedStatusCodes)
	}

	if skipStr, ok := params[paramHTTPSkipTLSVerify]; ok && skipStr != "" {
		skip, err := strconv.ParseBool(skipStr)
		if err != nil {
			return nil, fmt.Errorf("invalid %s: %w", paramHTTPSkipTLSVerify, err)
		}
		opts = append(opts, checker.WithHTTPSkipTLSVerify(skip))
		delete(unrecognizedParams, paramHTTPSkipTLSVerify)
	}

	if timeoutStr, ok := params[paramHTTPTimeout]; ok && timeoutStr != "" {
		t, err := time.ParseDuration(timeoutStr)
		if err != nil || t <= 0 {
			return nil, fmt.Errorf("invalid %q: %w", paramHTTPTimeout, err)
		}
		opts = append(opts, checker.WithHTTPTimeout(t))
		delete(unrecognizedParams, paramHTTPTimeout)
	}

	// After processing known parameters, check for unrecognized ones
	if len(unrecognizedParams) > 0 {
		var unknownKeys []string
		for key := range unrecognizedParams {
			unknownKeys = append(unknownKeys, key)
		}
		return nil, fmt.Errorf("unrecognized parameters for HTTP checker: %v", unknownKeys)
	}

	return opts, nil
}

// parseTCPCheckerOptions parses TCP checker specific options from parameters.
func parseTCPCheckerOptions(params map[string]string) ([]checker.Option, error) {
	var opts []checker.Option

	// Create a copy of params to track unrecognized parameters
	unrecognizedParams := make(map[string]struct{})
	for key := range params {
		unrecognizedParams[key] = struct{}{}
	}

	if timeoutStr, ok := params[paramTCPTimeout]; ok && timeoutStr != "" {
		t, err := time.ParseDuration(timeoutStr)
		if err != nil {
			return nil, fmt.Errorf("invalid %q: %w", paramTCPTimeout, err)
		}
		opts = append(opts, checker.WithTCPTimeout(t))
		delete(unrecognizedParams, paramTCPTimeout)
	}

	// After processing known parameters, check for unrecognized ones
	if len(unrecognizedParams) > 0 {
		var unknownKeys []string
		for key := range unrecognizedParams {
			unknownKeys = append(unknownKeys, key)
		}
		return nil, fmt.Errorf("unrecognized parameters for TCP checker: %v", unknownKeys)
	}

	return opts, nil
}

// parseICMPCheckerOptions parses ICMP checker specific options from parameters.
func parseICMPCheckerOptions(params map[string]string) ([]checker.Option, error) {
	var opts []checker.Option

	// Create a copy of params to track unrecognized parameters
	unrecognizedParams := make(map[string]struct{})
	for key := range params {
		unrecognizedParams[key] = struct{}{}
	}

	if readTimeoutStr, ok := params[paramICMPReadTimeout]; ok && readTimeoutStr != "" {
		rt, err := time.ParseDuration(readTimeoutStr)
		if err != nil {
			return nil, fmt.Errorf("invalid %q: %w", paramICMPReadTimeout, err)
		}
		opts = append(opts, checker.WithICMPReadTimeout(rt))
		delete(unrecognizedParams, paramICMPReadTimeout)
	}

	if writeTimeoutStr, ok := params[paramICMPWriteTimeout]; ok && writeTimeoutStr != "" {
		wt, err := time.ParseDuration(writeTimeoutStr)
		if err != nil {
			return nil, fmt.Errorf("invalid %q: %w", paramICMPWriteTimeout, err)
		}
		opts = append(opts, checker.WithICMPWriteTimeout(wt))
		delete(unrecognizedParams, paramICMPWriteTimeout)
	}

	// After processing known parameters, check for unrecognized ones
	if len(unrecognizedParams) > 0 {
		var unknownKeys []string
		for key := range unrecognizedParams {
			unknownKeys = append(unknownKeys, key)
		}
		return nil, fmt.Errorf("unrecognized parameters for ICMP checker: %v", unknownKeys)
	}

	return opts, nil
}
