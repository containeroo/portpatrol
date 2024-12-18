package main

import (
	"fmt"
	"time"

	"github.com/containeroo/portpatrol/pkg/dynflags"
)

func main() {
	// Initialize DynFlags
	dynFlags := dynflags.New(dynflags.ContinueOnError)

	// Define configuration groups and flags
	httpGroup := dynFlags.Group("http")
	httpGroup.String("method", "GET", "HTTP method to use")
	httpGroup.String("address", "", "HTTP target URL")
	httpGroup.Bool("secure", true, "Use secure connection (HTTPS)")
	httpGroup.Duration("timeout", 5*time.Second, "Request timeout")

	tcpGroup := dynFlags.Group("tcp")
	tcpGroup.String("address", "", "TCP target address")
	tcpGroup.Duration("timeout", 10*time.Second, "TCP timeout")

	// Simulate CLI arguments
	args := []string{
		"--http.identifier1.method", "POST",
		"--http.identifier1.address", "https://example.com",
		"--tcp.identifier2.address", "127.0.0.1",
		"--tcp.identifier2.timeout", "15s",
		"--unknown.identifier3.flag", "unknownValue",
	}

	// Parse arguments
	if err := dynFlags.Parse(args); err != nil {
		fmt.Printf("Error parsing flags: %v\n", err)
		return
	}

	// Print unparsable flags
	fmt.Println("Unparsable flags:")
	for _, flag := range dynFlags.UnknownArgs() {
		fmt.Printf("  %s\n", flag)
	}

	// ITERATION: Iterate over all config groups
	fmt.Println("=== Iterating over Config Groups ===")
	for groupName, group := range dynFlags.Config().Groups() {
		fmt.Printf("Group: %s\n", groupName)
		for flagName, flag := range group.Flags {
			fmt.Printf("  Flag: %s, Default: %v, Usage: %s\n", flagName, flag.Default, flag.Usage)
		}
	}

	// ITERATION: Iterate over all parsed groups
	fmt.Println("\n=== Iterating over Parsed Groups ===")
	for groupName, groups := range dynFlags.Parsed().Groups() {
		fmt.Printf("Group: %s\n", groupName)
		for _, group := range groups {
			fmt.Printf("  Identifier: %s\n", group.Name)
			for flagName, value := range group.Values {
				fmt.Printf("    Flag: %s, Value: %v\n", flagName, value)
			}
		}
	}

	// LOOKUP: Direct access using Lookup methods
	fmt.Println("\n=== Lookup Example ===")

	// Lookup a config group
	httpConfig := dynFlags.Config().Lookup("http")
	if httpConfig != nil {
		fmt.Printf("Config Group 'http' exists, Flags: %v\n", httpConfig.Flags)
	}

	// Lookup the "http" group
	httpGroups := dynFlags.Parsed().Lookup("http")
	if httpGroups != nil {
		// Lookup "identifier1" within the "http" group
		httpIdentifier1 := httpGroups.Lookup("identifier1")
		if httpIdentifier1 != nil {
			method := httpIdentifier1.Lookup("method")
			fmt.Printf("HTTP Method (Lookup): %s\n", method)
		}
	}

	// LOOKUP: Direct flag retrieval from a config group
	fmt.Println("\n=== Direct Flag Lookup ===")
	if httpConfig != nil {
		methodFlag := httpConfig.Lookup("method")
		if methodFlag != nil {
			fmt.Printf("HTTP Method Flag: Default = %v, Usage = %s\n", methodFlag.Default, methodFlag.Usage)
		}
	}
}
