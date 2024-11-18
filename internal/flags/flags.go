package flags

import (
	"bytes"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/containeroo/portpatrol/internal/parser"

	"github.com/spf13/pflag"
)

// Constants and parameter keys
const (
	defaultDebug         bool          = false
	paramDefaultInterval string        = "default-interval"
	defaultCheckInterval time.Duration = 2 * time.Second
)

type HelpRequested struct {
	Message string
}

func (e *HelpRequested) Error() string {
	return e.Message
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
		if !strings.HasPrefix(arg, fmt.Sprintf("--%s.", parser.ParamPrefix)) {
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
		fmt.Fprintf(&buf, "Usage: %s [OPTIONS] [--%s.<IDENTIFIER>.<PROPERTY> value]\n\nOptions:\n", flagSetName, parser.ParamPrefix)
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
		return nil, &HelpRequested{Message: fmt.Sprintf("%s version %s", flagSetName, version)}
	}

	// Process dynamic target flags
	targets, err := extractDynamicTargetFlags(dynamicArgs, buf)
	if err != nil {
		return nil, err
	}

	return &ParsedFlags{
		DefaultCheckInterval: *checkInterval,
		Targets:              targets,
	}, nil
}

// extractDynamicTargetFlags processes dynamic target flags and returns target configurations.
func extractDynamicTargetFlags(dynamicArgs []string, buf bytes.Buffer) (map[string]map[string]string, error) {
	targets := make(map[string]map[string]string)

	for i := 0; i < len(dynamicArgs); i++ {
		arg := dynamicArgs[i]

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

		// Check for duplicate flags
		if targets[targetName][parameter] != "" {
			return nil, fmt.Errorf("duplicate target flag: %s\n\n%s", arg, buf.String())
		}

		// Assign parameter value
		targets[targetName][parameter] = value
	}

	return targets, nil
}
