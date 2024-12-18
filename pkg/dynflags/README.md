# DynFlags

**DynFlags** is a Go package designed for dynamically managing hierarchical command-line flags. It supports parsing flags with a structure like `--group.identifier.flag=value` while allowing dynamic group and flag registration at runtime. For "POSIX/GNU-style --flags" use the library [pflag](https://github.com/spf13/pflag).

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

## Usage

`dynflags` is a Go package that provides a simple way to manage hierarchical command-line flags.
It supports parsing flags with a structure like `--group.identifier.flag=value` while allowing dynamic group and flag registration at runtime.
For POSIX/GNU-style `--flags` use the library [pflag](https://github.com/spf13/pflag). `dynflags` can be used together with `pflag`.

```go
import "github.com/containeroo/portpatrol/pkg/dynflags"
```

Create a new `DynFlags` instance:

```go
dynFlags := dynflags.New(dynflags.ContinueOnError)
```

Add groups to the `DynFlags` instance:

```go
httpGroup := dynFlags.Group("http")
```

Add flags to the `DynFlags` instance:

```go
httpGroup.String("method", "GET", "HTTP method for requests")
httpGroup.Int("timeout", 5, "Timeout duration for HTTP requests in seconds")
// httpGroup.Bool, httpGroup.Float64, httpGroup.Duration, etc.
```

After all flags are defined, call

```go
args := os.Args[1:] // Skip the first argument (the executable name)
dynflags.Parse(args)
```

to parse the command line into the defined flags. `args` are the command-line arguments to parse.
When integrating with `pflag`, set the `ParseBehavior` to `dynflags.ContinueOnError`. Parse `dynflags` arguments first, followed by `pflag`. Refer to `./examples/advanced/main.go` for a detailed example.
Unparsed arguments are stored in `dynflags.UnparsedArgs()`.

```go
args := os.Args[1:] // Skip the first argument (the executable name)

// Separate known and unknown flags
if err := dynFlags.Parse(args); err != nil {
    return err
}

unknownArgs := dynFlags.UnparsedArgs()

// Parse known flags
if err := flagSet.Parse(unknownArgs); err != nil {
    return err
}
```

`dynflags` provides 2 Groups:

- `dynflags.Config()` returns a `ConfigGroups` instance that provides direct access to the static configuration of the `DynFlags` instance.
- `dynflags.Parsed()` returns a `ParsedGroups` instance that provides direct access to the parsed configuration of the `DynFlags` instance.

Each of these Groups provides a `Lookup("SEARCH")` method that can be used to retrieve a specific group or flag.

```go
// Retrieve the "http" group
httpGroups := dynFlags.Parsed().Lookup("http")
// Retrieve "identifier1" object within "http"
httpIdentifier1 := httpGroups.Lookup("identifier1")
// Retrieve "method" object within "identifier1"
method := httpIdentifier1.Lookup("method")
// Show value of "method" within "identifier1"
value := method.Value()
fmt.Printf("Method: %s\n", value)
```

and each of these Groups provides a `Groups()` method that can be used to iterate over all groups.

```go
for groupName, groups := range dynFlags.Parsed().Groups() {
    fmt.Printf("Group: %s\n", groupName)
    for _, group := range groups {
        fmt.Printf("  Identifier: %s\n", group.Name)
        for flagName, value := range group.Values {
            fmt.Printf("    Flag: %s, Value: %v\n", flagName, value)
        }
    }
}
```

Unrecognized or unparsed arguments can be retrieved via `dynflags.UnknownArgs()`.

## Title, Description, and Epilog

`dynflags` allows you to set a title, description, and epilog for the help message.
You can also change the default usage output by setting the `Usage` field of a group. If not set, it uses the Group name in uppercase.

**Example:**

```go
dynFlags := dynflags.New(dynflags.ContinueOnError)
dynFlags.Title("DynFlags Example Application")
dynFlags.Description("This application demonstrates the usage of DynFlags for managing hierarchical flags dynamically.")
dynFlags.Epilog("For more information, see https://github.com/containerish/portpatrol")

tcpGroup := dynFlags.Group("tcp")
tcpGroup.Usage("TCP flags")
tcpGroup.String("Timeout", "10s", "TCP timeout")
tcpGroup.String("address", "127.0.0.1:8080", "TCP target address")

httpGroup := dynFlags.Group("http")
httpGroup.Usage("HTTP flags")
httpGroup.String("method", "GET", "HTTP method to use")
httpGroup.String("address", "https://example.com", "HTTP target URL")

dynFlags.PrintDefaults()
```

**Output:**

```text
DynFlags Example Application

This application demonstrates the usage of DynFlags for managing hierarchical flags dynamically.

TCP flags
  Flag                               Usage
  --tcp.<IDENTIFIER>.Timeout STRING  TCP timeout (default: 10s)
  --tcp.<IDENTIFIER>.address STRING  TCP target address (default: 127.0.0.1:8080)

HTTP flags
  Flag                                Usage
  --http.<IDENTIFIER>.method STRING   HTTP method to use (default: GET)
  --http.<IDENTIFIER>.address STRING  HTTP target URL (default: https://example.com)


For more information, see https://github.com/containerish/portpatrol
```

## Disable sorting of flags

`dynflags` allows you to disable sorting of groups and flags for help and usage message. Sort is disabled by default.

**Example:**

```go
dynFlags := dynflags.New(dynflags.ContinueOnError)
tcpGroup := dynFlags.Group("tcp")
tcpGroup.String("Timeout", "10s", "TCP timeout")
tcpGroup.String("address", "127.0.0.1:8080", "TCP target address")

httpGroup := dynFlags.Group("http")
httpGroup.String("method", "GET", "HTTP method to use")
httpGroup.String("address", "https://example.com", "HTTP target URL")

dynFlags.SortGroups = true
dynFlags.SortFlags = true
dynFlags.PrintDefaults()
```

**Output:**

```text
HTTP
  Flag                                Usage
  --http.<IDENTIFIER>.address STRING  HTTP target URL (default: https://example.com)
  --http.<IDENTIFIER>.method STRING   HTTP method to use (default: GET)

TCP
  Flag                               Usage
  --tcp.<IDENTIFIER>.Timeout STRING  TCP timeout (default: 10s)
  --tcp.<IDENTIFIER>.address STRING  TCP target address (default: 127.0.0.1:8080)
```

## MetaVar

`MetaVar` is a string that is used to represent the flag in the usage message. It defaults to the flag type in uppercase.

- String flags: `--<IDENTIFIER>.<FLAG>.<FLAG> STRING`
- Boolean flags: `--<IDENTIFIER>.<FLAG>.<FLAG> BOOL`

Slices will have a `MetaVar` with the base type in uppercase, followed by a lowercase `s`.

- String Slices: `--<IDENTIFIER>.<FLAG>.<FLAG> ..STRINGs`
- Boolean Slices: `--<IDENTIFIER>.<FLAG>.<FLAG> ..BOOLs`

To change the `MetaVar` for a flag, set the `MetaVar` field on the flag.

**Example:**

```go
dynFlags := dynflags.New(dynflags.ContinueOnError)
tcpGroup := dynFlags.Group("tcp")
timeout := tcpGroup.String("Timeout", "10s", "TCP timeout")
timeout.MetaVar("TIMEOUT")
dynFlags.PrintDefaults()
```

**Output:**

```text
TCP
  Flag                                   Usage
  --tcp.<IDENTIFIER>.Timeout TIMEOUT     TCP timeout (default: 10s)
```

## Examples

The `examples` directory contains a simple example that demonstrates the usage of `dynflags`, as well as an advanced example that shows how to use `dynflags` with `pflag`.

