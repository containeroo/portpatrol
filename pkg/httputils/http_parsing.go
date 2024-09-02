// Package httputils provides utility functions for parsing HTTP headers and status codes
// from strings. These functions are designed to help in scenarios where HTTP-related configurations
// are passed as strings, such as in environment variables or configuration files.
package httputils

import (
	"fmt"
	"strconv"
	"strings"
)

// ParseStatusCodes parses a comma-separated string of HTTP status codes and ranges into a slice of integers.
//
// Parameters:
//   - statusRanges: Comma-separated string of status codes and/or ranges. It supports combinations of single codes
//     (e.g., "200") and ranges (e.g., "200-204"), including mixed combinations like "200,300-301,404".
//
// Returns:
//   - A slice of status codes, or an error if parsing fails.
func ParseStatusCodes(statusRanges string) ([]int, error) {
	var statusCodes []int

	ranges := strings.Split(statusRanges, ",")
	for _, r := range ranges {
		trimmed := strings.TrimSpace(r)

		if !strings.Contains(trimmed, "-") {
			// Handle individual status codes like "200"
			code, err := strconv.Atoi(trimmed)
			if err != nil {
				return nil, fmt.Errorf("invalid status code: %s", trimmed)
			}
			statusCodes = append(statusCodes, code)
			continue
		}

		// Handle ranges like "200-204"
		parts := strings.Split(trimmed, "-") // Split the range into start and end
		if len(parts) != 2 {
			return nil, fmt.Errorf("invalid status range: %s", trimmed)
		}

		// Parse the start and end status codes
		start, err1 := strconv.Atoi(strings.TrimSpace(parts[0]))
		end, err2 := strconv.Atoi(strings.TrimSpace(parts[1]))

		// Check if parsing failed or if the start is greater than the end
		if err1 != nil || err2 != nil || start > end {
			return nil, fmt.Errorf("invalid status range: %s", trimmed)
		}

		// Generate a slice of status codes in the range
		for i := start; i <= end; i++ {
			statusCodes = append(statusCodes, i)
		}
	}

	return statusCodes, nil
}

// ParseHeaders parses a comma-separated string of HTTP headers into a map.
//
// Parameters:
//   - headers: Comma-separated string of headers in "Key=Value" format.
//     The value can be empty (e.g., "X-Empty-Header="), but the key must not be empty.
//   - allowDuplicates: If true, allows duplicate keys with the last occurrence taking precedence.
//
// Returns:
//   - A map of header names to values, or an error if parsing fails.
func ParseHeaders(headers string, allowDuplicates bool) (map[string]string, error) {
	headerMap := make(map[string]string)
	if headers == "" {
		return headerMap, nil
	}

	// Split the headers into key=value pairs
	pairs := strings.Split(headers, ",")
	for _, pair := range pairs {
		trimmedPair := strings.TrimSpace(pair)
		if trimmedPair == "" {
			continue // Skip any empty parts resulting from trailing commas
		}

		// Split the pair into key and value
		parts := strings.SplitN(trimmedPair, "=", 2)
		if len(parts) != 2 {
			return nil, fmt.Errorf("invalid header format: %s", pair)
		}

		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])

		if key == "" {
			return nil, fmt.Errorf("header key cannot be empty: %s", pair)
		}

		if _, exists := headerMap[key]; exists && !allowDuplicates {
			return nil, fmt.Errorf("duplicate header key found: %s", key)
		}

		headerMap[key] = value
	}

	return headerMap, nil
}
