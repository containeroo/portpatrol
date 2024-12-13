# DynFlags

**DynFlags** is a Go package designed for dynamically managing hierarchical command-line flags. It supports parsing flags with a structure like `--group.identifier.flag=value` while allowing dynamic group and flag registration at runtime.

## Features

- Dynamically register groups and flags at runtime.
- Hierarchical structure for flags (`group.identifier.flag`).
- Supports multiple data types: `string`, `int`, `bool`, `float64`, `time.Duration`, etc.
- Handles unknown groups and flags with configurable behavior.
- Provides a customizable usage output.
- Designed with testability in mind by accepting `io.Writer` for output.

## Installation

Install the package using:

```bash
go get github.com/containeroo/portpatrol/pkg/dynflags
```

## Example Usage

Here’s an example of how to use DynFlags in your application:

```go
package main

import (
  "fmt"
  "os"
  "time"

  "github.com/containeroo/portpatrol/pkg/dynflags"
)

func main() {
  // Initialize DynFlags with ContinueOnError behavior
  dynFlags := dynflags.New(dynflags.ContinueOnError)

  // Add a title and description for the usage output
  dynFlags.Title("DynFlags Example Application")
  dynFlags.Description("This application demonstrates the usage of DynFlags for managing hierarchical flags dynamically.")
  dynFlags.Epilogue("For more information, see https://github.com/containerish/portpatrol")

  // Register groups and flags
  httpGroup := dynFlags.Group("http")
  httpGroup.String("method", "GET", "HTTP method to use")
  httpGroup.String("address", "", "HTTP target URL")
  httpGroup.Bool("secure", true, "Use secure connection (HTTPS)")
  httpGroup.Duration("timeout", 5*time.Second, "Request timeout")

  tcpGroup := dynFlags.Group("tcp")
  tcpGroup.String("address", "", "TCP target address")
  tcpGroup.Duration("timeout", 10*time.Second, "TCP timeout")

  // Parse command-line arguments
  args := os.Args[1:]
  if err := dynFlags.Parse(args); err != nil {
    fmt.Fprintf(os.Stderr, "Error parsing flags: %v\n", err)
    os.Exit(1)
  }

  // Access parsed values
  for groupName, groups := range dynFlags.GetAllParsedGroups() {
    for _, group := range groups {
      fmt.Printf("Group: %s, Identifier: %s\n", groupName, group.Name)
      if method, err := group.GetString("method"); err == nil {
        fmt.Printf("  Method: %s\n", method)
      }
      if address, err := group.GetString("address"); err == nil {
        fmt.Printf("  Address: %s\n", address)
      }
      if timeout, err := group.GetDuration("timeout"); err == nil {
        fmt.Printf("  Timeout: %s\n", timeout)
      }
    }
  }

  // Handle unknown groups or flags
  unknownGroups := dynFlags.GetUnknownGroups()
  for groupName, groups := range unknownGroups {
    fmt.Printf("Unknown Group: %s\n", groupName)
    for _, group := range groups {
      fmt.Printf("  Identifier: %s\n", group.Name)
      for key, value := range group.GetUnknownValues() {
        fmt.Printf("    Unknown Flag: %s, Value: %v\n", key, value)
      }
    }
  }
}
```

## Output Example

Running the application with:

```bash
go run main.go --http.api.address=http://example.com --http.api.timeout=10s --tcp.server.address=tcp://127.0.0.1 --unknown.group.flag=value
```

Produces the following output:

```bash
Group: http, Identifier: api
  Address: http://example.com
  Timeout: 10s

Group: tcp, Identifier: server
  Address: tcp://127.0.0.1
  Timeout: 10s

Unknown Group: unknown
  Identifier: group
    Unknown Flag: flag, Value: value
```

## Advanced Usage

### Handling Unknown Groups and Flags

DynFlags supports three behaviors for handling unknown flags:

- `ExitOnError`: Stops execution with an error.
- `ContinueOnError`: Skips unknown flags and continues parsing.
- `IgnoreUnknown`: Collects unknown flags and groups them under `unknownGroups`.

### Customizing Usage Output

You can customize the usage output by setting the title, description, and epilog:

```go
dynFlags.AddTitle("Custom Application")
dynFlags.AddDescription("Manage dynamic flags with ease!")
dynFlags.AddEpilog("For more information, visit: https://github.com/yourusername/dynflags")
```

Testing with Custom Output
To test or capture usage output, set a custom io.Writer:

```go
var buf bytes.Buffer
dynFlags.SetOutput(&buf)
dynFlags.Usage()
fmt.Println(buf.String())
```
