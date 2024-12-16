package config

import (
	"fmt"
	"io"
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

// Is returns true if the error is a HelpRequested error.
func (e *HelpRequested) Is(target error) bool {
	_, ok := target.(*HelpRequested)
	return ok
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
func ParseFlags(args []string, version string, output io.Writer) (*ParsedFlags, error) {
	// Create global flagSet and dynamic flags
	flagSet := setupGlobalFlags()
	dynFlags := setupDynamicFlags()

	// Set output for flagSet and dynFlags
	flagSet.SetOutput(output)
	dynFlags.SetOutput(output)

	// Set up custom usage function
	setupUsage(output, flagSet, dynFlags)

	// Separate known and unknown flags
	knownArgs, unknownArgs := dynFlags.SeparateKnownAndUnknownArgs(args)

	// Parse known flags
	if err := flagSet.Parse(unknownArgs); err != nil {
		return parseAndHandleErrors(err, output, flagSet)
	}

	// TODO: is this necessary?

	// Handle special flags (e.g., --help or --version)
	if err := handleSpecialFlags(flagSet, output, version); err != nil {
		return nil, err
	}

	// Parse unknown arguments with dynamic flags
	if err := dynFlags.Parse(knownArgs); err != nil {
		return nil, fmt.Errorf("error parsing dynamic flags: %w", err)
	}

	// Retrieve the default interval value
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

	flagSet.Duration(paramDefaultInterval, defaultCheckInterval, "Default interval between checks. Can be overridden for each target.")
	flagSet.Bool("version", false, "Show version and exit.")
	flagSet.BoolP("help", "h", false, "Show help.")

	return flagSet
}

// setupDynamicFlags sets up dynamic flags for HTTP, TCP, ICMP.
func setupDynamicFlags() *dynflags.DynFlags {
	dynFlags := dynflags.New(dynflags.ExitOnError)
	dynFlags.Epilog("For more information, see https://github.com/containeroo/portpatrol")
	dynFlags.SortGroups = true
	dynFlags.SortFlags = true

	// HTTP flags
	httpFlags := dynFlags.Group("http")
	httpFlags.String("name", "", "Name of the HTTP checker")
	httpFlags.String("method", "GET", "HTTP method to use")
	httpFlags.String("address", "", "HTTP target URL")
	httpFlags.Duration("interval", 1*time.Second, "Time between HTTP requests. Can be overwritten with --default-interval.")
	httpFlags.String("headers", "", "HTTP headers to send")
	httpFlags.Bool("allow-duplicate-headers", defaultHTTPAllowDuplicateHeaders, "Allow duplicate HTTP headers")
	httpFlags.String("expected-status-codes", "", "Expected HTTP status codes")
	httpFlags.Bool("skip-tls-verify", defaultHTTPSkipTLSVerify, "Skip TLS verification")
	httpFlags.Duration("timeout", 2*time.Second, "Timeout in seconds")

	// ICMP flags
	icmpFlags := dynFlags.Group("icmp")
	icmpFlags.String("name", "", "Name of the ICMP checker")
	icmpFlags.String("address", "", "ICMP target address")
	icmpFlags.Duration("interval", 1*time.Second, "Time between ICMP requests. Can be overwritten with --default-interval.")
	icmpFlags.Duration("read-timeout", 2*time.Second, "Timeout for ICMP read")
	icmpFlags.Duration("write-timeout", 2*time.Second, "Timeout for ICMP write")

	// TCP flags
	tcpFlags := dynFlags.Group("tcp")
	tcpFlags.String("name", "", "Name of the TCP checker")
	tcpFlags.String("address", "", "TCP target address")
	tcpFlags.Duration("timeout", 2*time.Second, "Timeout for TCP connection")
	tcpFlags.Duration("interval", 1*time.Second, "Time between TCP requests. Can be overwritten with --default-interval.")

	return dynFlags
}

// setupUsage sets the custom usage function.
func setupUsage(output io.Writer, flagSet *pflag.FlagSet, dynFlags *dynflags.DynFlags) {
	flagSet.Usage = func() {
		fmt.Fprintln(output, "Usage: portpatrol [FLAGS] [DYNAMIC FLAGS..]")

		fmt.Fprintln(output, "\nGlobal Flags:")
		flagSet.PrintDefaults()

		fmt.Fprintln(output, "\nDynamic Flags:")
		dynFlags.PrintDefaults()
	}
}

// parseAndHandleErrors processes errors during flag parsing.
func parseAndHandleErrors(err error, output io.Writer, flagSet *pflag.FlagSet) (*ParsedFlags, error) {
	fmt.Fprintf(output, "%s\n\n", err.Error())
	flagSet.Usage()
	return nil, fmt.Errorf("Flag parsing error: %s", err.Error())
}

// handleSpecialFlags handles help and version flags.
func handleSpecialFlags(flagSet *pflag.FlagSet, output io.Writer, version string) error {
	if flagSet.Lookup("help").Value.String() == "true" {
		flagSet.Usage()
		return &HelpRequested{Message: ""}
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
