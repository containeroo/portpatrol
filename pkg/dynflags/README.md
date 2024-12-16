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
httpGroup.String("method", "GET", "HTTP method to use")
httpGroup.Int("timeout", 5, "Timeout for HTTP requests")
// httpGroup.Bool, httpGroup.Float64, httpGroup.Duration, etc.
```

After all flags are defined, call

```go
args := os.Args[1:] // Skip the first argument (the executable name)
dynflags.Parse(args)
```

to parse the command line into the defined flags. `args` are the command-line arguments to parse.
When using `pflag`, the the ParseBehavior is set to `dynflags.ContinueOnError` and parse first `dynflags` and then `pflag`.
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

`dynflags` provides 3 Groups:

- `dynflags.Config()` returns a `ConfigGroups` instance that provides direct access to the static configuration of the `DynFlags` instance.
- `dynflags.Parsed()` returns a `ParsedGroups` instance that provides direct access to the parsed configuration of the `DynFlags` instance.
- `dynflags.Unknown()` returns a `UnknownGroups` instance that provides direct access to the unknown configuration of the `DynFlags` instance.

Each of these Groups provides a `Lookup("SEARCH")` method that can be used to retrieve a specific group or flag.

```go
// Retrieve the "http" group
httpGroups := dynFlags.Parsed().Lookup("http")
// Retrieve "identifier1" within
httpIdentifier1 := httpGroups.Lookup("identifier1")
// Retrieve "method" within "identifier1"
httpIdentifier1.Lookup("method")
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

## Examples

The `examples` directory contains a simple example that demonstrates the usage of `dynflags`, as well as an advanced example that shows how to use `dynflags` with `pflag`.

