package main

import (
	"fmt"
	"os"
	"time"

	"github.com/containeroo/portpatrol/pkg/dynflags"
)

func main() {
	// Initialize DynFlags
	dynFlags := dynflags.New(dynflags.ContinueOnError)

	// Define HTTP group
	httpGroup := dynFlags.Group("http")
	httpGroup.String("method", "GET", "HTTP method to use")
	httpGroup.URL("address", "", "HTTP target URL")
	httpGroup.Bool("secure", true, "Use secure connection (HTTPS)")
	httpGroup.Int("retries", 3, "Number of retries")
	httpGroup.Float("timeout", 1.5, "Timeout in seconds")

	// Define TCP group
	tcpGroup := dynFlags.Group("tcp")
	tcpGroup.String("address", "", "The TCP target address")
	tcpGroup.Duration("timeout", 5*time.Second, "Timeout for TCP connections")

	// Parse CLI arguments
	err := dynFlags.Parse(os.Args[1:])
	if err != nil {
		fmt.Printf("Error parsing flags: %v\n", err)
		os.Exit(1)
	}

	dynFlags.Usage()
}
