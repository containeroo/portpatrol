package wait

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/containeroo/portpatrol/internal/checks"
)

// WaitUntilReady continuously attempts to connect to the specified target until it becomes available or the context is canceled.
func WaitUntilReady(ctx context.Context, interval time.Duration, checker checks.Checker, logger *slog.Logger) error {
	logger = logger.With(
		slog.String("target", checker.GetName()),
		slog.String("type", checker.GetType()),
		slog.String("address", checker.GetAddress()),
		slog.Duration("interval", interval),
	)

	logger.Info(fmt.Sprintf("Waiting for %s to become ready...", checker.GetName()))

	for {
		err := checker.Check(ctx)
		if err == nil {
			logger.Info(fmt.Sprintf("%s is ready ✓", checker.GetName()))
			return nil // Successfully connected to the target
		}

		logger.Warn(fmt.Sprintf("%s is not ready ✗", checker.GetName()), slog.String("error", err.Error()))

		select {
		case <-time.After(interval):
			// Continue to the next connection attempt after the interval
		case <-ctx.Done():
			if ctx.Err() == context.Canceled {
				return nil // Treat context cancellation as expected behavior
			}
			return ctx.Err()
		}
	}
}