# httputils

The `httputils` package provides utility functions for parsing HTTP headers and status codes from strings. These functions are designed to facilitate working with HTTP-related configurations that are passed as strings, such as environment variables or configuration files.

## Features

- Parse HTTP status codes and ranges from a string.
- Parse HTTP headers into a key-value map.
- Support for handling duplicate headers.

## Installation

To use the `httputils` package, add it to your Go project:

```sh
go get github.com/containerish/portpatrol/pkg/httputils
```

## Usage

### ParseStatusCodes

Parses a comma-separated string of HTTP status codes and ranges into a slice of integers.

__Example:__

```go
package main

import (
    "fmt"
    "log"
    "httputils"
)

func main() {
    statusString := "200,300-302,404"
    statusCodes, err := httputils.ParseStatusCodes(statusString)
    if err != nil {
        log.Fatalf("Error parsing status codes: %v", err)
    }
    fmt.Println("Parsed Status Codes:", statusCodes)
}
```

__Parameters:__

- `statusRanges` (string): Comma-separated string of single status codes (e.g., `200`) and/or ranges (e.g., `200-204`).

__Returns:__

- `[]int`: A slice of status codes.
- `error`: An error if the parsing fails.

__Output:__

```bash
Parsed Status Codes: [200 300 301 302 404]
```

## ParseHeaders

Parses a comma-separated string of HTTP headers into a key-value map.

__Example:__

```go
package main

import (
    "fmt"
    "log"
    "httputils"
)

func main() {
    headerString := "Content-Type=application/json,Authorization=Bearer token,X-Custom-Header="
    headers, err := httputils.ParseHeaders(headerString, false)
    if err != nil {
        log.Fatalf("Error parsing headers: %v", err)
    }
    fmt.Println("Parsed Headers:", headers)
}
```

__Parameters:__

- `headers` (string): Comma-separated string of headers in `Key=Value` format. Keys must not be empty.
- `allowDuplicates` (bool): If `true`, overrides previous values for duplicate keys. If `false`, returns an error on duplicate keys.

__Returns:__

- `map[string]string`: A map of header names to values.
- `error`: An error if the parsing fails.

__Output:__

```bash
Parsed Headers: map[Content-Type:application/json Authorization:Bearer token X-Custom-Header:]

```

## Error Handling

Both `ParseStatusCodes` and `ParseHeaders` return descriptive errors for invalid input, such as:

- Invalid HTTP status codes or ranges.
- Empty or malformed header keys.
- Duplicate header keys (when `allowDuplicates` is `false`).

