package main

import (
	"fmt"
	"log"
	"time"

	"github.com/containeroo/portpatrol/pkg/dynflags"
)

func main() {
	// Initialize DynFlags with ExitOnError behavior
	dynFlags := dynflags.New(dynflags.ExitOnError)

	// Define the HTTP parent group with static configuration
	httpGroup, err := dynFlags.Group("http")
	if err != nil {
		log.Fatalf("Error creating group: %v", err)
	}
	httpMethod := httpGroup.String("method", "GET", "HTTP method to use")
	httpGroup.URL("address", "", "HTTP target URL")
	httpGroup.Bool("secure", true, "Use secure connection (HTTPS)")
	httpGroup.Int("retries", 3, "Number of retries")
	httpGroup.Float64("timeout", 1.5, "Timeout in seconds")

	// Define the TCP parent group with static configuration
	tcpGroup, err := dynFlags.Group("tcp")
	if err != nil {
		log.Fatalf("Error creating group: %v", err)
	}
	tcpGroup.String("address", "", "The TCP target address")
	tcpGroup.Duration("timeout", 5*time.Second, "Timeout for TCP connections")

	// Parse command-line arguments
	args := []string{
		"--http.IDENTIFIER1.method=POST",
		"--http.IDENTIFIER1.address=http://google.com",
		"--http.IDENTIFIER1.secure=false",
		"--tcp.IDENTIFIER1.address=tcp://service.com",
		"--tcp.IDENTIFIER1.timeout=10s",
	}

	if err := dynFlags.Parse(args); err != nil {
		log.Fatalf("Error parsing flags: %v", err)
	}

	fmt.Println("blabla", *httpMethod)
	fmt.Println(dynFlags)

	dynFlags.Usage()

	// Print parsed groups and their flag values
	fmt.Println("\nParsed Groups:")
	for parentName, childGroups := range dynFlags.GetAllParsedGroups() {
		fmt.Printf("Parent Group: %s\n", parentName)
		for _, group := range childGroups {
			fmt.Printf("  Child Group: %s\n", group.Name)
			for flagName, value := range group.Values {
				fmt.Printf("    %s: %v\n", flagName, value)
			}
		}
	}
}
