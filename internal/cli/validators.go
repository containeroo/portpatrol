package cli

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/containeroo/httputils"
)

// validateHTTPStatusCodes validates a single status code value or range expression.
func validateHTTPStatusCodes(codes string) error {
	for code := range strings.SplitSeq(codes, ",") {
		_, err := httputils.ParseStatusCodes(code)
		if err != nil {
			return fmt.Errorf("invalid status code: %w", err)
		}
	}

	return nil
}

// validateMaxAttempts validates the global max-attempts flag.
func validateMaxAttempts(v int) error {
	if v == 0 {
		return errors.New("max-attempts must be -1 or positive")
	}
	if v < -1 {
		return errors.New("max-attempts must be -1 or positive")
	}

	return nil
}

// validateOptionalMaxAttempts validates per-target max-attempts where zero means inherit global.
func validateOptionalMaxAttempts(v int) error {
	if v == 0 {
		return nil
	}

	return validateMaxAttempts(v)
}

// validateTimeoutDuration validates a timeout flag.
func validateTimeoutDuration() func(time.Duration) error {
	return func(d time.Duration) error {
		if d <= 0 {
			return errors.New("timeout must be positive")
		}

		return nil
	}
}

// validateNonNegativeDuration returns a validator that rejects negative durations.
func validateNonNegativeDuration(name string) func(time.Duration) error {
	return func(d time.Duration) error {
		if d < 0 {
			return fmt.Errorf("%s must be non-negative", name)
		}

		return nil
	}
}

// validateNonNegativeInt returns a validator that rejects negative integers.
func validateNonNegativeInt(name string) func(int) error {
	return func(v int) error {
		if v < 0 {
			return fmt.Errorf("%s must be non-negative", name)
		}

		return nil
	}
}
