package cli

import (
	"time"

	"github.com/containeroo/never/internal/checker"
	"github.com/containeroo/tinyflags"
)

// registerTCPFlags registers TCP-related flags and binds them to cfg.
func registerTCPFlags(tf *tinyflags.FlagSet) {
	tcp := tf.DynamicGroup("tcp").Title("TCP")
	tcp.String("name", "", "Name of the TCP checker. Defaults to <ID>.")
	tcp.String("address", "", "TCP target address").
		Validate(validateTCPAddress).
		Required()
	tcp.Duration("timeout", checker.DefaultTCPConfig().Timeout, "Timeout for TCP connection").
		Validate(validateTimeoutDuration()).
		Placeholder("DURATION")
	tcp.Duration("interval", 0*time.Second, "Time between TCP requests. Defaults to --default-interval when unset or 0.").
		Validate(validateNonNegativeDuration("interval")).
		Placeholder("DURATION")
	tcp.Int("max-attempts", 0, "Maximum attempts before giving up. Inherits --max-attempts when unset; 0 means endless retries.").
		Validate(validateNonNegativeInt("max-attempts")).
		Placeholder("N")
	registerRetryFlags(tcp)
}
