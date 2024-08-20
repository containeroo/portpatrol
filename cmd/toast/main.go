package main

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/containeroo/toast/pkg/checker"
	"github.com/containeroo/toast/pkg/config"
	"github.com/containeroo/toast/pkg/runner"
)

const version = "0.0.1"

// setupLogger configures the logger based on the configuration
func setupLogger(cfg config.Config, output io.Writer) *slog.Logger {
	handlerOpts := &slog.HandlerOptions{}

	if cfg.LogAdditionalFields {
		// Add additional fields to the logger
		return slog.New(slog.NewTextHandler(output, handlerOpts)).With(
			slog.String("target_address", cfg.TargetAddress),
			slog.String("interval", cfg.Interval.String()),
			slog.String("dial_timeout", cfg.DialTimeout.String()),
			slog.String("checker_type", cfg.CheckType),
			slog.String("version", version),
		)
	}

	handlerOpts.ReplaceAttr = func(groups []string, a slog.Attr) slog.Attr {
		// Remove the "error" key from the attributes
		if a.Key == "error" {
			return slog.Attr{}
		}
		return a
	}
	return slog.New(slog.NewTextHandler(output, handlerOpts))
}

// run runs the main loop
func run(ctx context.Context, getenv func(string) string, output io.Writer) error {
	cfg, err := config.ParseConfig(getenv)
	if err != nil {
		return fmt.Errorf("configuration error: %w", err)
	}

	logger := setupLogger(cfg, output)

	checkerInstance, err := checker.NewChecker(ctx, cfg.CheckType, cfg.TargetName, cfg.TargetAddress, cfg.DialTimeout, getenv)
	if err != nil {
		return fmt.Errorf("failed to initialize checker: %w", err)
	}

	return runner.RunLoop(ctx, cfg, checkerInstance, logger)
}

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	if err := run(ctx, os.Getenv, os.Stdout); err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err)
		os.Exit(1)
	}
}
