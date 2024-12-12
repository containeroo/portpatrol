package config

import (
	"bytes"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/containeroo/portpatrol/pkg/dynflags"
	"github.com/spf13/pflag"
)

const (
	paramDefaultInterval             string        = "default-interval"
	defaultCheckInterval             time.Duration = 2 * time.Second
	defaultHTTPAllowDuplicateHeaders bool          = false
	defaultHTTPSkipTLSVerify         bool          = false
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
	DynFlags             *dynflags.DynFlags
}

// ParseFlags parses command-line arguments and returns the parsed flags.
func ParseFlags(args []string, version string) (*ParsedFlags, error) {
	flagSet := setupGlobalFlags()
	dynFlags, _ := setupDynamicFlags()

	// Buffer for capturing help and error messages
	var buf bytes.Buffer
	flagSet.SetOutput(&buf)
	dynFlags.SetOutput(&buf)

	// Set up custom usage
	setupUsage(flagSet, dynFlags)

	// Separate known and unknown flags
	knownArgs, unknownArgs := separateKnownAndUnknownArgs(args, flagSet)

	// Parse known flags
	if err := flagSet.Parse(knownArgs); err != nil {
		return parseAndHandleErrors(err, &buf, flagSet, dynFlags)
	}

	// Handle special flags
	if err := handleSpecialFlags(flagSet, &buf, version); err != nil {
		return nil, err
	}

	// Parse unknown arguments with dynamic flags
	if err := dynFlags.Parse(unknownArgs); err != nil {
		return nil, fmt.Errorf("error parsing dynamic flags: %w", err)
	}

	// Get the default interval
	defaultInterval, err := getDurationFlag(flagSet, paramDefaultInterval, defaultCheckInterval)
	if err != nil {
		return nil, err
	}

	return &ParsedFlags{
		DefaultCheckInterval: defaultInterval,
		DynFlags:             dynFlags,
	}, nil
}

// setupGlobalFlags sets up global application flags.
func setupGlobalFlags() *pflag.FlagSet {
	flagSet := pflag.NewFlagSet("portpatrol", pflag.ContinueOnError)
	flagSet.SortFlags = false

	flagSet.ParseErrorsWhitelist.UnknownFlags = true
	flagSet.Bool("version", false, "Show version and exit.")
	flagSet.BoolP("help", "h", false, "Show help.")
	flagSet.Duration(paramDefaultInterval, defaultCheckInterval, "Default interval between checks. Can be overridden for each target.")

	return flagSet
}

// setupDynamicFlags sets up dynamic flags for HTTP, TCP, ICMP.
func setupDynamicFlags() (*dynflags.DynFlags, error) {
	dynFlags := dynflags.New(dynflags.ContinueOnError)

	// HTTP flags
	httpFlags, _ := dynFlags.Group("http")
	httpFlags.String("name", "", "Name of the HTTP checker")
	httpFlags.String("method", "GET", "HTTP method to use")
	httpFlags.String("address", "", "HTTP target URL")
	httpFlags.Bool("secure", true, "Use secure connection (HTTPS)")
	httpFlags.String("headers", "", "HTTP headers to send")
	httpFlags.Bool("allow-duplicate-headers", defaultHTTPAllowDuplicateHeaders, "Allow duplicate HTTP headers")
	httpFlags.String("expected-status-codes", "", "Expected HTTP status codes")
	httpFlags.Bool("skip-tls-verify", defaultHTTPSkipTLSVerify, "Skip TLS verification")
	httpFlags.Duration("timeout", 2*time.Second, "Timeout in seconds")

	// ICMP flags
	icmpFlags, _ := dynFlags.Group("icmp")
	icmpFlags.String("name", "", "Name of the ICMP checker")
	icmpFlags.String("address", "", "ICMP target address")
	icmpFlags.Duration("read-timeout", 2*time.Second, "Timeout for ICMP read")
	icmpFlags.Duration("write-timeout", 2*time.Second, "Timeout for ICMP write")

	// TCP flags
	tcpFlags, _ := dynFlags.Group("tcp")
	tcpFlags.String("name", "", "Name of the TCP checker")
	tcpFlags.String("address", "", "TCP target address")
	tcpFlags.Duration("timeout", 2*time.Second, "Timeout for TCP connection")

	return dynFlags, nil
}

// setupUsage sets the custom usage function.
func setupUsage(flagSet *pflag.FlagSet, dynFlags *dynflags.DynFlags) {
	flagSet.Usage = func() {
		fmt.Println("Usage: portpatrol [OPTIONS]")
		fmt.Println("\nOptions:")
		flagSet.PrintDefaults()
		fmt.Println("\nDynamic Flags:")
		dynFlags.PrintDefaults()
	}
}

// parseAndHandleErrors processes errors during flag parsing.
func parseAndHandleErrors(err error, buf *bytes.Buffer, flagSet *pflag.FlagSet, dynFlags *dynflags.DynFlags) (*ParsedFlags, error) {
	buf.WriteString(err.Error())
	buf.WriteString("\n\n")
	flagSet.Usage()
	return nil, errors.New(buf.String())
}

// handleSpecialFlags handles help and version flags.
func handleSpecialFlags(flagSet *pflag.FlagSet, buf *bytes.Buffer, version string) error {
	if flagSet.Lookup("help").Value.String() == "true" {
		flagSet.Usage()
		buf.WriteString("\n")
		return &HelpRequested{Message: buf.String()}
	}

	if flagSet.Lookup("version").Value.String() == "true" {
		return &HelpRequested{Message: fmt.Sprintf("PortPatrol version %s\n", version)}
	}

	return nil
}

// Example of getting a flag value as a time.Duration
func getDurationFlag(flagSet *pflag.FlagSet, name string, defaultValue time.Duration) (time.Duration, error) {
	flag := flagSet.Lookup(name)
	if flag == nil {
		return defaultValue, nil
	}

	// Parse the flag value from string to time.Duration
	value, err := time.ParseDuration(flag.Value.String())
	if err != nil {
		return defaultValue, fmt.Errorf("invalid duration for flag '%s'", flag.Value.String())
	}

	return value, nil
}

// separateKnownAndUnknownArgs separates known and unknown flags from the command-line arguments.
func separateKnownAndUnknownArgs(args []string, flagSet *pflag.FlagSet) (known []string, unknown []string) {
	for _, arg := range args {
		if !strings.HasPrefix(arg, "--") {
			known = append(known, arg) // Positional arguments are considered known
			continue
		}
		// Extract the flag name
		parts := strings.SplitN(arg[2:], "=", 2)
		flagName := parts[0]

		// Check if the flag is known
		if flagSet.Lookup(flagName) != nil {
			known = append(known, arg)
		} else {
			unknown = append(unknown, arg)
		}
	}
	return known, unknown
}
