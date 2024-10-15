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
	defaultCheckInterval  time.Duration = 2 * time.Second
	defaultDialTimeout    time.Duration = 1 * time.Second
	defaultLogExtraFields bool          = false

	defaultHTTPAllowDuplicateHeaders bool = false
	defaultHTTPSkipTLSVerify         bool = false
)

type HelpRequested struct {
	Message string
}

func (e *HelpRequested) Error() string {
	return e.Message
}

type VersionRequested struct {
	Version string
}

func (e *VersionRequested) Error() string {
	return e.Version
}

type TargetChecker struct {
	Interval time.Duration
	Checker  checker.Checker
}

type ParsedFlags struct {
	ShowHelp             bool
	ShowVersion          bool
	Version              string
	DefaultDialTimeout   time.Duration // The timeout for dialing the target.
	DefaultCheckInterval time.Duration // The interval between connection attempts.
	LogExtraFields       bool          // Whether to log the fields in the log message.

	Targets map[string]map[string]string
}

func ParseFlags(args []string, version string) (*ParsedFlags, error) {
	// Preprocess arguments to extract dynamic target flags
	var newArgs []string
	var dynamicArgs []string

	for i := 0; i < len(args); i++ {
		arg := args[i]
		if strings.HasPrefix(arg, "--target.") {
			dynamicArgs = append(dynamicArgs, arg)
			if !strings.Contains(arg, "=") && i+1 < len(args) && !strings.HasPrefix(args[i+1], "-") {
				i++
				dynamicArgs = append(dynamicArgs, args[i])
			}
		} else {
			newArgs = append(newArgs, arg)
		}
	}

	flagSetName := "portpatrol"
	flagSet := pflag.NewFlagSet(flagSetName, pflag.ContinueOnError)
	flagSet.SortFlags = false

	// Set a buffer to capture help messages or error output
	var buf bytes.Buffer
	flagSet.SetOutput(&buf)

	// Set custom usage function
	flagSet.Usage = func() {
		fmt.Fprintf(&buf, "Usage: %s [OPTIONS] [--target.identifier.property=value]\n\nOptions:\n", flagSetName)
		flagSet.PrintDefaults()
		// Show properties for TCP
		buf.WriteString("\nTCP properties:\n")
		buf.WriteString("  name: The name of the Target. If not specified, it's derived from the target address.\n")
		buf.WriteString("  address: The IP address or hostname of the target in the following formats: scheme://hostname:port")
		buf.WriteString("  timeout: The timeout for the TCP connection.\n")
		buf.WriteString("  type: The type of check to perform. Must be \"tcp\".\n")
		// Show properties for HTTP
		buf.WriteString("\nHTTP properties:\n")
		buf.WriteString("  name: The name of the Target. If not specified, it's derived from the target address.\n")
		buf.WriteString("  address: The IP address or hostname of the target in the following formats: [tcp://]hostname[:port]\n")
		buf.WriteString("  timeout: The timeout for the HTTP request.\n")
		buf.WriteString("  type: The type of check to perform. Must be \"http\" or \"https\".\n")
		buf.WriteString("  method: The HTTP method to use. Defaults to \"GET\".\n")
		buf.WriteString("  headers: A comma-separated list of HTTP headers to include in the request.\n")
		buf.WriteString("  allow_duplicate_headers: Whether to allow duplicate headers in the request. Defaults to false.\n")
		buf.WriteString("  expected_status_codes: A comma-separated list of expected status codes. Defaults to 200.\n")
		buf.WriteString("  skip_tls_verify: Whether to skip TLS verification. Defaults to false.\n")
		// Show properties for ICMP
		buf.WriteString("\nICMP properties:\n")
		buf.WriteString("  name: The name of the Target. If not specified, it's derived from the target address.\n")
		buf.WriteString("  address: The IP address or hostname of the target in the following formats: [icmp://]hostname\n")
	}

	// Define known flags
	showVersion := flagSet.Bool("version", false, "Show version")
	showHelp := flagSet.BoolP("help", "h", false, "Show help")
	dialTimeout := flagSet.Duration("dial-timeout", defaultDialTimeout, "Timeout for dialing the target")
	checkInterval := flagSet.Duration("check-interval", defaultCheckInterval, "Interval between checks")
	logExtraFields := flagSet.Bool("log-extra-fields", defaultLogExtraFields, "Log extra fields")

	// Parse known flags with the newArgs slice
	if err := flagSet.Parse(newArgs); err != nil {
		buf.WriteString(err.Error())
		buf.WriteString("\n\n")
		flagSet.Usage()
		return nil, errors.New(buf.String())
	}

	// Handle help request
	if *showHelp {
		flagSet.Usage()
		return nil, &HelpRequested{Message: buf.String()}
	}

	// Handle version request
	if *showVersion {
		return nil, &VersionRequested{Version: fmt.Sprintf("%s version %s", flagSetName, version)}
	}

	// Process the dynamic target flags
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

		// Initialize the target map if necessary and store the parameter value
		if _, exists := targets[targetName]; !exists {
			targets[targetName] = make(map[string]string)
			targets[targetName]["name"] = value
		}

		// Store the parameter value
		targets[targetName][parameter] = value
	}

	// Check for unknown arguments
	for _, arg := range flagSet.Args() {
		if strings.HasPrefix(arg, "-") && !strings.HasPrefix(arg, "--target.") {
			return nil, fmt.Errorf("Warning: Unknown flag ignored: %s\n", arg)
		}
	}

	return &ParsedFlags{
		DefaultCheckInterval: *checkInterval,
		DefaultDialTimeout:   *dialTimeout,
		LogExtraFields:       *logExtraFields,
		Targets:              targets,
	}, nil
}

func ParseChecker(targets map[string]map[string]string, defaultInterval time.Duration) ([]TargetChecker, error) {
	var targetCheckers []TargetChecker

	for targetName, params := range targets {
		// Determine the check type
		checkTypeStr, ok := params["type"]
		if !ok || checkTypeStr == "" {
			// Try to infer the type from the address scheme
			address := params["address"]
			parts := strings.SplitN(address, "://", 2)
			if len(parts) == 2 {
				checkTypeStr = parts[0]
				// params["address"] = parts[1]
			} else {
				return nil, fmt.Errorf("missing 'type' parameter for target '%s'", targetName)
			}
		}

		checkType, err := checker.GetCheckTypeFromString(checkTypeStr)
		if err != nil {
			return nil, fmt.Errorf("unsupported check type '%s' for target '%s'", checkTypeStr, targetName)
		}

		name := targetName
		if n, ok := params["name"]; ok && n != "" {
			name = n
		}

		address, ok := params["address"]
		if !ok || address == "" {
			return nil, fmt.Errorf("missing 'address' for target '%s'", targetName)
		}

		// Create the specific config based on the check type
		var config checker.CheckerConfig

		switch checkType {
		case checker.HTTP:
			cfg, err := parseHTTPCheckerConfig(params)
			if err != nil {
				return nil, fmt.Errorf("failed to parse HTTP config for target '%s': %w", targetName, err)
			}
			config = cfg
		case checker.TCP:
			cfg, err := parseTCPCheckerConfig(params)
			if err != nil {
				return nil, fmt.Errorf("failed to parse TCP config for target '%s': %w", targetName, err)
			}
			config = cfg
		case checker.ICMP:
			cfg, err := parseICMPCheckerConfig(params)
			if err != nil {
				return nil, fmt.Errorf("failed to parse ICMP config for target '%s': %w", targetName, err)
			}
			config = cfg
		default:
			return nil, fmt.Errorf("unsupported check type '%s' for target '%s'", checkTypeStr, targetName)
		}

		// Create the checker
		chk, err := checker.NewChecker(checkType, name, address, config)
		if err != nil {
			return nil, fmt.Errorf("failed to create checker for target '%s': %w", targetName, err)
		}

		// Get interval from parameters or use default
		interval := defaultInterval
		if intervalStr, ok := params["interval"]; ok && intervalStr != "" {
			interval, err = time.ParseDuration(intervalStr)
			if err != nil {
				return nil, fmt.Errorf("invalid 'interval' for target '%s': %w", targetName, err)
			}
		}

		targetCheckers = append(targetCheckers, TargetChecker{
			Interval: interval,
			Checker:  chk,
		})
	}

	return targetCheckers, nil
}

func parseHTTPCheckerConfig(params map[string]string) (checker.HTTPCheckerConfig, error) {
	cfg := checker.HTTPCheckerConfig{}

	if method, ok := params["method"]; ok && method != "" {
		cfg.Method = method
	}

	if headersStr, ok := params["headers"]; ok && headersStr != "" {
		headers, err := httputils.ParseHeaders(headersStr, defaultHTTPAllowDuplicateHeaders)
		if err != nil {
			return cfg, fmt.Errorf("invalid headers: %w", err)
		}
		cfg.Headers = headers
	}

	if codesStr, ok := params["expected_status_codes"]; ok && codesStr != "" {
		codes, err := httputils.ParseStatusCodes(codesStr)
		if err != nil {
			return cfg, fmt.Errorf("invalid expected_status_codes: %w", err)
		}
		cfg.ExpectedStatusCodes = codes
	}

	if skipStr, ok := params["skip_tls_verify"]; ok && skipStr != "" {
		skip, err := strconv.ParseBool(skipStr)
		if err != nil {
			return cfg, fmt.Errorf("invalid skip_tls_verify: %w", err)
		}
		cfg.SkipTLSVerify = skip
	}

	if timeoutStr, ok := params["timeout"]; ok && timeoutStr != "" {
		t, err := time.ParseDuration(timeoutStr)
		if err != nil {
			return cfg, fmt.Errorf("invalid timeout: %w", err)
		}
		cfg.Timeout = t
	}

	if intervalStr, ok := params["interval"]; ok && intervalStr != "" {
		interval, err := time.ParseDuration(intervalStr)
		if err != nil {
			return cfg, fmt.Errorf("invalid interval: %w", err)
		}
		cfg.Interval = interval
	}

	return cfg, nil
}

func parseTCPCheckerConfig(params map[string]string) (checker.TCPCheckerConfig, error) {
	cfg := checker.TCPCheckerConfig{}

	if timeoutStr, ok := params["timeout"]; ok && timeoutStr != "" {
		t, err := time.ParseDuration(timeoutStr)
		if err != nil {
			return cfg, fmt.Errorf("invalid timeout: %w", err)
		}
		cfg.Timeout = t
	}

	if intervalStr, ok := params["interval"]; ok && intervalStr != "" {
		interval, err := time.ParseDuration(intervalStr)
		if err != nil {
			return cfg, fmt.Errorf("invalid interval: %w", err)
		}
		cfg.Interval = interval
	}

	return cfg, nil
}

func parseICMPCheckerConfig(params map[string]string) (checker.ICMPCheckerConfig, error) {
	cfg := checker.ICMPCheckerConfig{}

	if readTimeoutStr, ok := params["read_timeout"]; ok && readTimeoutStr != "" {
		rt, err := time.ParseDuration(readTimeoutStr)
		if err != nil {
			return cfg, fmt.Errorf("invalid read_timeout: %w", err)
		}
		cfg.ReadTimeout = rt
	}

	if writeTimeoutStr, ok := params["write_timeout"]; ok && writeTimeoutStr != "" {
		wt, err := time.ParseDuration(writeTimeoutStr)
		if err != nil {
			return cfg, fmt.Errorf("invalid write_timeout: %w", err)
		}
		cfg.WriteTimeout = wt
	}

	if intervalStr, ok := params["interval"]; ok && intervalStr != "" {
		interval, err := time.ParseDuration(intervalStr)
		if err != nil {
			return cfg, fmt.Errorf("invalid interval: %w", err)
		}
		cfg.Interval = interval
	}

	return cfg, nil
}
