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

	// ITERATION: Iterate over all config groups
  fmt.Println("=== Iterating over Config Groups ===")
  for groupName, groups := range dynFlags.ConfigGroups().Groups() {

	// ITERATION: Iterate over all parsed groups
	fmt.Println("=== Iterating over Parsed Groups ===")
	for groupName, groups := range dynFlags.Parsed().Groups() {
		fmt.Printf("Group: %s\n", groupName)
		for _, group := range groups {
			fmt.Printf("  Identifier: %s\n", group.Name)
			for flagName, value := range group.Values {
				fmt.Printf("    Flag: %s, Value: %v\n", flagName, value)
			}
		}
	}

	fmt.Println("\n=== Iterating over Unknown Groups ===")
	for groupName, groups := range dynFlags.Unknown().Groups() {
		fmt.Printf("Unknown Group: %s\n", groupName)
		for _, group := range groups {
			fmt.Printf("  Identifier: %s\n", group.Name)
			for flagName, value := range group.Values {
				fmt.Printf("    Flag: %s, Value: %v\n", flagName, value)
			}
		}
	}

	// LOOKUP: Direct access using Lookup methods
	fmt.Println("\n=== Lookup Example ===")

	// Lookup the "http" group
	httpGroups, err := dynFlags.Parsed().Lookup("http")
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	// Lookup "identifier1" within the "http" group
	httpIdentifier1 := httpGroups[0]
	method, err := httpIdentifier1.Lookup("method")
	if err != nil {
		fmt.Printf("Error: %v\n", err)
	} else {
		fmt.Printf("HTTP Method (Lookup): %s\n", method)
	}

	// Lookup the "unknown" group
	unknownGroups, err := dynFlags.Unknown().Lookup("unknown")
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	// Lookup "identifier3.flag" within the "unknown" group
	unknownIdentifier3 := unknownGroups[0]
	unknownValue, err := unknownIdentifier3.Lookup("flag")
	if err != nil {
		fmt.Printf("Error: %v\n", err)
	} else {
		fmt.Printf("Unknown Value (Lookup): %s\n", unknownValue)
	}
}
