// flags/flags.go

package flags

import (
	"bytes"
	"errors"
	"fmt"
	"strings"
	"text/tabwriter"
	"time"

	"github.com/spf13/pflag"
)

// Constants and parameter keys
const (
	defaultDebug         bool          = false
	paramDefaultInterval string        = "default-interval"
	defaultCheckInterval time.Duration = 2 * time.Second

	paramPrefix   string = "target"
	paramType     string = "type"
	paramName     string = "name"
	paramAddress  string = "address"
	paramInterval string = "interval"
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

// ParsedFlags holds the parsed command-line flags.
type ParsedFlags struct {
	ShowHelp             bool
	ShowVersion          bool
	Version              string
	DefaultCheckInterval time.Duration
	Targets              map[string]map[string]string
}

// ParseCommandLineFlags parses command line arguments and returns the parsed flags.
func ParseCommandLineFlags(args []string, version string) (*ParsedFlags, error) {
	var knownArgs []string
	var dynamicArgs []string

	// Separate known flags and dynamic target flags
	for i := 0; i < len(args); i++ {
		arg := args[i]
		if !strings.HasPrefix(arg, fmt.Sprintf("--%s.", paramPrefix)) {
			knownArgs = append(knownArgs, arg)
			continue
		}

		dynamicArgs = append(dynamicArgs, arg)
		// Check if the next argument is not a flag (happens when "--target.identifier.param value" has no = between param and value)
		if !strings.Contains(arg, "=") && i+1 < len(args) && !strings.HasPrefix(args[i+1], "-") {
			i++ // Use next argument as value
			dynamicArgs = append(dynamicArgs, args[i])
		}
	}

	flagSetName := "portpatrol"
	flagSet := pflag.NewFlagSet(flagSetName, pflag.ContinueOnError)
	flagSet.SortFlags = false

	// Buffer to capture help and error messages
	var buf bytes.Buffer
	flagSet.SetOutput(&buf)

	// Custom usage function
	flagSet.Usage = func() {
		fmt.Fprintf(&buf, "Usage: %s [OPTIONS] [--%s.<IDENTIFIER>.<PROPERTY> value]\n\nOptions:\n", flagSetName, paramPrefix)
		flagSet.PrintDefaults()
		displayCheckerProperties(&buf)
	}

	// Define known flags
	checkInterval := flagSet.Duration(paramDefaultInterval, defaultCheckInterval, "Default interval between checks. Can be overwritten for each target.")
	showVersion := flagSet.Bool("version", false, "Show version and exit.")
	showHelp := flagSet.BoolP("help", "h", false, "Show help.")

	// Parse known flags
	if err := flagSet.Parse(knownArgs); err != nil {
		buf.WriteString(err.Error())
		buf.WriteString("\n\n")
		flagSet.Usage()
		return nil, errors.New(buf.String())
	}

	// Handle help
	if *showHelp {
		flagSet.Usage()
		return nil, &HelpRequested{Message: buf.String()}
	}

	// Handle version
	if *showVersion {
		return nil, &VersionRequested{Version: fmt.Sprintf("%s version %s", flagSetName, version)}
	}

	// Process dynamic target flags
	targets, err := extractDynamicTargetFlags(dynamicArgs, buf)
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
		Targets:              targets,
	}, nil
}

// displayCheckerProperties appends checker properties documentation to the buffer, including the expected type for each flag.
func displayCheckerProperties(buf *bytes.Buffer) {
	// Initialize a new tab writer with 2-space indentation
	w := tabwriter.NewWriter(buf, 0, 0, 2, ' ', 0)

	// TCP Checker properties
	fmt.Fprintln(w, "\nTCP Checker Properties:")
	fmt.Fprintln(w, "  Flag\tDescription")
	fmt.Fprintf(w, "  --%s.<IDENTIFIER>.%s string\tThe IP address or hostname of the target in the following format: tcp://hostname:port\n\tIf the (tcp://) is specified, the check type is automatically inferred,\n\tmaking the --%s.<IDENTIFIER>%s flag optional.\n", paramPrefix, paramAddress, paramPrefix, paramType)
	fmt.Fprintf(w, "  --%s.<IDENTIFIER>.%s string\tThe name of the target. If not specified, it's derived from the target address.\n", paramPrefix, paramName)
	fmt.Fprintf(w, "  --%s.<IDENTIFIER>.%s string\tThe type of check to perform. If the scheme (tcp://) is specified in --%s.<IDENTIFIER>.%s,\n\tthis flag can be omitted as the type will be inferred.\n", paramPrefix, paramType, paramPrefix, paramAddress)
	fmt.Fprintf(w, "  --%s.<IDENTIFIER>.%s duration\tThe timeout for the TCP connection (e.g., 2s, 500ms).\n", paramPrefix, paramTCPTimeout)
	fmt.Fprintf(w, "  --%s.<IDENTIFIER>.%s duration\toverride the default interval for this target (e.g., 5s).\n", paramPrefix, paramInterval)
	fmt.Fprintln(w, "")

	// HTTP Checker properties
	fmt.Fprintln(w, "HTTP Checker Properties:")
	fmt.Fprintln(w, "  Flag\tDescription")
	fmt.Fprintf(w, "  --%s.<IDENTIFIER>.%s string\tThe IP address or hostname of the target in the following format: scheme://hostname:[port]\n\tIf a scheme (e.g. http://) is specified, the check type is automatically inferred,\n\tmaking the --%s.<IDENTIFIER>%s flag optional.\n", paramPrefix, paramAddress, paramPrefix, paramType)
	fmt.Fprintf(w, "  --%s.<IDENTIFIER>.%s string\tThe name of the target. If not specified, it's derived from the target address.\n", paramPrefix, paramName)
	fmt.Fprintf(w, "  --%s.<IDENTIFIER>.%s string\tThe type of check to perform. If a scheme (e.g. http://) is specified in --%s.<IDENTIFIER>.%s,\n\tthis flag can be omitted as the type will be inferred.\n", paramPrefix, paramType, paramPrefix, paramAddress)
	fmt.Fprintf(w, "  --%s.<IDENTIFIER>.%s string\tThe HTTP method to use (e.g., GET, POST). Defaults to \"GET\".\n", paramPrefix, paramHTTPMethod)
	fmt.Fprintf(w, "  --%s.<IDENTIFIER>.%s string\tA comma-separated list of HTTP headers to include in the request in \"key=value\" format.\n\tExample: Authorization=Bearer token,Content-Type=application/json\n", paramPrefix, paramHTTPHeaders)
	fmt.Fprintf(w, "  --%s.<IDENTIFIER>.%s string\tA comma-separated list of expected HTTP status codes or ranges. Defaults to 200.\n\tExample: \"200,301,404\" or \"200,300-302\" or \"200,301-302,404,500-502\"\n", paramPrefix, paramHTTPExpectedStatusCodes)
	fmt.Fprintf(w, "  --%s.<IDENTIFIER>.%s bool\tWhether to skip TLS verification. Defaults to false.\n", paramPrefix, paramHTTPSkipTLSVerify)
	fmt.Fprintf(w, "  --%s.<IDENTIFIER>.%s duration\tThe timeout for the HTTP request (e.g., 5s).\n", paramPrefix, paramHTTPTimeout)
	fmt.Fprintf(w, "  --%s.<IDENTIFIER>.%s duration\t(Optional) Override the default interval for this target (e.g., 10s).\n", paramPrefix, paramInterval)
	fmt.Fprintln(w, "")

	// ICMP Checker properties
	fmt.Fprintln(w, "ICMP Checker Properties:")
	fmt.Fprintln(w, "  Flag\tDescription")
	fmt.Fprintf(w, "  --%s.<IDENTIFIER>.%s string\tThe IP address or hostname of the target in the following format: icmp://hostname (no port allowed).\n\tIf the scheme (icmp://) is specified, the check type is automatically inferred,\n\tmaking the --%s.<IDENTIFIER>%s flag optional.\n", paramPrefix, paramAddress, paramPrefix, paramType)
	fmt.Fprintf(w, "  --%s.<IDENTIFIER>.%s string\tThe name of the target. If not specified, it's derived from the target address.\n", paramPrefix, paramName)
	fmt.Fprintf(w, "  --%s.<IDENTIFIER>.%s string\tThe type of check to perform. If the scheme (icmp://) is specified in --%s.<IDENTIFIER>.%s,\n\tthis flag can be omitted as the type will be inferred.\n", paramPrefix, paramType, paramPrefix, paramAddress)

	fmt.Fprintf(w, "  --%s.<IDENTIFIER>.%s duration\tThe read timeout for the ICMP connection (e.g., 1s).\n", paramPrefix, paramICMPReadTimeout)
	fmt.Fprintf(w, "  --%s.<IDENTIFIER>.%s duration\tThe write timeout for the ICMP connection (e.g., 1s).\n", paramPrefix, paramICMPWriteTimeout)
	fmt.Fprintf(w, "  --%s.<IDENTIFIER>.%s duration\t(Optional) Override the default interval for this target (e.g., 5s).\n", paramPrefix, paramInterval)

	// Flush the writer to ensure all data is written to the buffer
	w.Flush()
}

// extractDynamicTargetFlags processes dynamic target flags and returns target configurations.
func extractDynamicTargetFlags(dynamicArgs []string, buf bytes.Buffer) (map[string]map[string]string, error) {
	targets := make(map[string]map[string]string)

	for i := 0; i < len(dynamicArgs); i++ {
		arg := dynamicArgs[i]

		// Remove the "--" prefix
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

		// Split the flag name into parts
		nameParts := strings.Split(flagName, ".")
		if len(nameParts) < 3 {
			return nil, fmt.Errorf("invalid target flag format: %s\n\n%s", flagName, buf.String())
		}

		targetName := nameParts[1]                    // e.g., "web"
		parameter := strings.Join(nameParts[2:], ".") // e.g., "address"

		// Initialize target map if necessary
		if _, exists := targets[targetName]; !exists {
			targets[targetName] = make(map[string]string)
		}

		// Assign parameter value
		targets[targetName][parameter] = value
	}

	return targets, nil
}
