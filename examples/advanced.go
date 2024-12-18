package main

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/containeroo/portpatrol/pkg/dynflags"
	"github.com/spf13/pflag"
)

func main() {
	// args := os.Args[1:]
	args := []string{
		"--http.idenitfier1.method", "POST",
		"--http.idenitfier1.address", "https://example.com",
		"--tcp.idenitfier1.address", "127.0.0.1",
		"--tcp.idenitfier1.timeout", "10s",
		"--unknown.identifier2.name", "example 2",
	}

	var output strings.Builder // create a io.Writer to capture output

	// Initialize pflag with ContinueOnError behavior
	flagSet := pflag.NewFlagSet("advanced", pflag.ContinueOnError)
	// Add some flags
	flagSet.Bool("debug", false, "Set debug mode")
	flagSet.SetOutput(&output) // Output to the io.Writer

	// Initialize DynFlags with ContinueOnError behavior
	dynFlags := dynflags.New(dynflags.ContinueOnError)

	// Set the output for the DynFlags instance to the same io.Writer
	dynFlags.SetOutput(&output)

	// Create a custom usage function for the flagSet instance
	flagSet.Usage = func() {
		fmt.Fprintln(&output, "Usage: advanced [FLAGS] [DYNAMIC FLAGS..]")

		fmt.Fprintln(&output, "\nGlobal Flags:")
		flagSet.PrintDefaults()

		fmt.Fprintln(&output, "\nDynamic Flags:")
		dynFlags.PrintDefaults()
	}

	// Add a title and description for the usage output
	dynFlags.Title("DynFlags Example Application")
	dynFlags.Description("This application demonstrates the usage of DynFlags for managing hierarchical flags dynamically.")
	dynFlags.Epilog("For more information, see https://github.com/containerish/portpatrol")

	// Register groups and flags
	httpGroup := dynFlags.Group("http")
	httpGroup.String("method", "GET", "HTTP method to use")
	httpGroup.String("address", "", "HTTP target URL")

	tcpGroup := dynFlags.Group("tcp")
	tcpGroup.String("address", "", "TCP target address")
	tcpGroup.Duration("timeout", 10*time.Second, "TCP timeout")

	// Parse first with dynflags
	if err := dynFlags.Parse(args); err != nil {
		fmt.Fprintf(os.Stderr, "Error parsing flags: %v\n", err)
		os.Exit(1)
	}

	// Print unparsable flags
	fmt.Println("Unparsable flags:")
	for _, flag := range dynFlags.UnknownArgs() {
		fmt.Printf("  %s\n", flag)
	}

	// Retrieve values from parsed group http
	method := dynFlags.Parsed().Lookup("http").Lookup("idenitfier1").Lookup("method")
	httpAddress := dynFlags.Parsed().Lookup("http").Lookup("idenitfier1").Lookup("address")

	fmt.Println("Method:", method)
	fmt.Println("Address:", httpAddress)

	// Retrieve values from parsed group tcp
	tcpAddress := dynFlags.Parsed().Lookup("tcp").Lookup("idenitfier1").Lookup("address")
	tcpTimeout := dynFlags.Parsed().Lookup("tcp").Lookup("idenitfier1").Lookup("timeout")

	fmt.Println("TCP Address:", tcpAddress)
	fmt.Println("TCP Timeout:", tcpTimeout)
}
