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
When using `pflag`, use `pflag.Parse()`

