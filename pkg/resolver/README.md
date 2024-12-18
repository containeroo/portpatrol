# Resolver Package

The `resolver` package provides a flexible and extensible way to resolve configuration values from various sources including environment variables, files, JSON, YAML, INI, and key-value files. It uses a prefix-based system to identify which resolver to use and returns the resolved value or an error if something goes wrong.

## Installation

```bash
go get github.com/yourusername/yourrepo/resolver
```

## Usage

The primary entry point is the `ResolveVariable` function. It takes a single string and attempts to resolve it based on its prefix:

- `env`: – Resolves environment variables.
  Example: `env:PATH` returns the value of the `PATH` environment variable.
- `file`: – Resolves values from a simple key-value file.
  Example: `file:/config/app.txt//KeyName` returns the value associated with `KeyName` in `app.txt`.
- `json`: – Resolves values from a JSON file. Supports dot notation for nested keys.
  Example: `json:/config/app.json//database.host` returns `host` field under `database` in `app.json`. It is also possible to indexing into arrays (e.g., `json:/config/app.json//servers.0.host`).
- `yaml`: – Resolves values from a YAML file. Supports dot notation for nested keys.
  Example: `yaml:/config/app.yaml//server.port` returns `port` under `server` in `app.yaml`.It is also possible to indexing into arrays (e.g., `yaml:/config/app.yaml//servers.0.host`).
- `ini`: – Resolves values from an INI file. Can specify a section and key, or just a key in the default section.
  Example: `ini:/config/app.ini//Section.Key` returns the value of `Key` under `Section`.
- No prefix – Returns the value as-is, unchanged.

## Example

```go
package main

import (
    "fmt"
    "log"
    "os"

    "github.com/containeroo/portpatrol/resolver"
)

func main() {
    os.Setenv("MY_VAR", "HelloWorld")

    val, err := resolver.ResolveVariable("env:MY_VAR")
    if err != nil {
        log.Fatal(err)
    }
    fmt.Println(val) // Output: HelloWorld

    // Resolve a JSON key:
    // Given a JSON file: /config/app.json
    // {
    //   "server": {
    //     "host": "localhost",
    //     "port": 8080
    //   }
    // }
    jsonVal, err := resolver.ResolveVariable("json:/config/app.json//server.host")
    if err != nil {
        log.Fatal(err)
    }
    fmt.Println(jsonVal) // Output: localhost
}
```
