package config

import (
	"bytes"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/containeroo/dynflags"

	flag "github.com/spf13/pflag"
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
func ParseFlags(args []string, version string, w io.Writer) (*ParsedFlags, error) {
	// Create global fs and dynamic flags
	fs := setupGlobalFlags()
	df := setupDynamicFlags()

	// Set output for flagSet and dynFlags
	fs.SetOutput(w)
	df.SetOutput(w)

	// Set up custom usage function
	setupUsage(fs, df)

	// Parse unknown arguments with dynamic flags
	if err := df.Parse(args); err != nil {
		return nil, fmt.Errorf("error parsing dynamic flags: %w", err)
	}

	unknownArgs := df.UnknownArgs()

	// Parse known flags
	if err := fs.Parse(unknownArgs); err != nil {
		return nil, fmt.Errorf("Flag parsing error: %s", err.Error())
	}

	// Handle special flags (e.g., --help or --version)
	if err := handleSpecialFlags(fs, version); err != nil {
		return nil, err
	}

	// Retrieve the default interval value
	defaultInterval, err := getDurationFlag(fs, paramDefaultInterval, defaultCheckInterval)
	if err != nil {
		return nil, err
	}

	return &ParsedFlags{
		DefaultCheckInterval: defaultInterval,
		DynFlags:             df,
	}, nil
}

// setupGlobalFlags sets up global application flags.
func setupGlobalFlags() *flag.FlagSet {
	fs := flag.NewFlagSet("PortPatrol", flag.ContinueOnError)
	fs.SortFlags = false

	fs.Duration(paramDefaultInterval, defaultCheckInterval, "Default interval between checks. Can be overridden for each target.")
	fs.Bool("version", false, "Show version and exit.")
	fs.BoolP("help", "h", false, "Show help.")

	return fs
}

// setupDynamicFlags sets up dynamic flags for HTTP, TCP, ICMP.
func setupDynamicFlags() *dynflags.DynFlags {
	df := dynflags.New(dynflags.ContinueOnError)
	df.Epilog("For more information, see https://github.com/containeroo/portpatrol")
	df.SortGroups = true
	df.SortFlags = true

	// HTTP flags
	http := df.Group("http")
	http.String("name", "", "Name of the HTTP checker")
	http.String("method", "GET", "HTTP method to use")
	http.String("address", "", "HTTP target URL")
	http.Duration("interval", 1*time.Second, "Time between HTTP requests. Can be overwritten with --default-interval.")
	http.StringSlices("header", nil, "HTTP headers to send")
	http.Bool("allow-duplicate-headers", defaultHTTPAllowDuplicateHeaders, "Allow duplicate HTTP headers")
	http.String("expected-status-codes", "200", "Expected HTTP status codes")
	http.Bool("skip-tls-verify", defaultHTTPSkipTLSVerify, "Skip TLS verification")
	http.Duration("timeout", 2*time.Second, "Timeout in seconds")

	// ICMP flags
	icmp := df.Group("icmp")
	icmp.String("name", "", "Name of the ICMP checker")
	icmp.String("address", "", "ICMP target address")
	icmp.Duration("interval", 1*time.Second, "Time between ICMP requests. Can be overwritten with --default-interval.")
	icmp.Duration("read-timeout", 2*time.Second, "Timeout for ICMP read")
	icmp.Duration("write-timeout", 2*time.Second, "Timeout for ICMP write")

	// TCP flags
	tcp := df.Group("tcp")
	tcp.String("name", "", "Name of the TCP checker")
	tcp.String("address", "", "TCP target address")
	tcp.Duration("timeout", 2*time.Second, "Timeout for TCP connection")
	tcp.Duration("interval", 1*time.Second, "Time between TCP requests. Can be overwritten with --default-interval.")

	return df
}

// setupUsage sets the custom usage function.
func setupUsage(fs *flag.FlagSet, df *dynflags.DynFlags) {
	fs.Usage = func() {
		fmt.Fprintf(fs.Output(), "Usage: %s [FLAGS] [DYNAMIC FLAGS..]\n", strings.ToLower(fs.Name()))

		fmt.Fprintln(fs.Output(), "\nGlobal Flags:")
		fs.PrintDefaults()

		fmt.Fprintln(fs.Output(), "\nDynamic Flags:")
		df.PrintDefaults()
	}
}

// handleSpecialFlags handles help and version flags.
func handleSpecialFlags(fs *flag.FlagSet, version string) error {
	help := fs.Lookup("help")
	if help != nil && help.Value.String() == "true" {
		// create a buffer to capture the output to pass to the HelpRequested error message
		buffer := &bytes.Buffer{}
		fs.SetOutput(buffer)
		fs.Usage()
		return &HelpRequested{Message: buffer.String()}
	}

	versionFlag := fs.Lookup("version")
	if versionFlag != nil && versionFlag.Value.String() == "true" {
		return &HelpRequested{Message: fmt.Sprintf("%s version %s\n", fs.Name(), version)}
	}

	return nil
}

// Example of getting a flag value as a time.Duration
func getDurationFlag(flagSet *flag.FlagSet, name string, defaultValue time.Duration) (time.Duration, error) {
	flag := flagSet.Lookup(name)
	if flag == nil {
		return defaultValue, nil
	}

	// Parse the flag val from string to time.Duration
	val, err := time.ParseDuration(flag.Value.String())
	if err != nil {
		return defaultValue, fmt.Errorf("invalid duration for flag '%s'", flag.Value.String())
	}

	return val, nil
}
