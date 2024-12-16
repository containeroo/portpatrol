package factory

import (
	"fmt"
	"time"

	"github.com/containeroo/portpatrol/internal/checker"
	"github.com/containeroo/portpatrol/pkg/dynflags"
	"github.com/containeroo/portpatrol/pkg/httputils"
)

// CheckerWithInterval represents a checker with its interval.
type CheckerWithInterval struct {
	Interval time.Duration
	Checker  checker.Checker
}

// BuildCheckers creates a list of CheckerWithInterval from the parsed dynflags configuration.
func BuildCheckers(dynFlags *dynflags.DynFlags, defaultInterval time.Duration) ([]CheckerWithInterval, error) {
	var checkers []CheckerWithInterval

	// Iterate over all parsed groups
	for parentName, childGroups := range dynFlags.Parsed().Groups() {
		checkType, err := checker.GetCheckTypeFromString(parentName)
		if err != nil {
			return nil, fmt.Errorf("invalid check type '%s': %w", parentName, err)
		}

		// Process each parsed group (child) under the parent group
		for _, group := range childGroups {
			address, err := group.GetString("address")
			if err != nil {
				return nil, fmt.Errorf("missing address for %s checker: %w", parentName, err)
			}

			// Default interval for the checker
			interval := defaultInterval
			if customInterval, err := group.GetDuration("interval"); err == nil {
				interval = customInterval
			}

			// Prepare options based on the checker type
			var opts []checker.Option

			switch checkType {
			case checker.HTTP:
				if method, err := group.GetString("method"); err == nil {
					opts = append(opts, checker.WithHTTPMethod(method))
				}
				if headers, err := group.GetString("headers"); err == nil && headers != "" {
					headersMap, err := httputils.ParseHeaders(headers, true)
					if err != nil {
						return nil, fmt.Errorf("invalid \"--%s.%s.headers\": %w", parentName, group.Name, err)
					}
					opts = append(opts, checker.WithHTTPHeaders(headersMap))
				}
				if skipTLS, err := group.GetBool("skip-tls-verify"); err == nil {
					opts = append(opts, checker.WithHTTPSkipTLSVerify(skipTLS))
				}
				if timeout, err := group.GetDuration("timeout"); err == nil {
					opts = append(opts, checker.WithHTTPTimeout(timeout))
				}

			case checker.TCP:
				if timeout, err := group.GetDuration("timeout"); err == nil {
					opts = append(opts, checker.WithHTTPTimeout(timeout)) // Could have a TCP-specific timeout option
				}

			case checker.ICMP:
				if readTimeout, err := group.GetDuration("read-timeout"); err == nil {
					opts = append(opts, checker.WithICMPReadTimeout(readTimeout))
				}
				if writeTimeout, err := group.GetDuration("write-timeout"); err == nil {
					opts = append(opts, checker.WithICMPWriteTimeout(writeTimeout))
				}
			}

			name, _ := group.GetString("name")
			if name == "" {
				name = group.Name
			}

			instance, err := checker.NewChecker(checkType, name, address, opts...)
			if err != nil {
				return nil, fmt.Errorf("failed to create %s checker: %w", parentName, err)
			}

			// Wrap the checker with its interval and add to the list
			checkers = append(checkers, CheckerWithInterval{
				Interval: interval,
				Checker:  instance,
			})
		}
	}

	return checkers, nil
}
