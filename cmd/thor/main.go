package main

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/signal"
	"syscall"

	"github.com/containeroo/thor/pkg/checker"
	"github.com/containeroo/thor/pkg/config"
	"github.com/containeroo/thor/pkg/logger"
	"github.com/containeroo/thor/pkg/runner"
)

const version = "0.0.8"

// run is the main function of the application
func run(ctx context.Context, getenv func(string) string, output io.Writer) error {
	ctx, cancel := signal.NotifyContext(ctx, os.Interrupt, syscall.SIGTERM)
	defer cancel()

	cfg, err := config.ParseConfig(getenv)
	if err != nil {
		return fmt.Errorf("configuration error: %w", err)
	}
	cfg.Version = version

	logger := logger.SetupLogger(cfg, output)

	checkerInstance, err := checker.NewChecker(cfg.CheckType, cfg.TargetName, cfg.TargetAddress, cfg.DialTimeout, getenv)
	if err != nil {
		return fmt.Errorf("failed to initialize checker: %w", err)
	}

	return runner.RunLoop(ctx, cfg, checkerInstance, logger)
}

func main() {
	ctx := context.Background()

	if err := run(ctx, os.Getenv, os.Stdout); err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err)
		os.Exit(1)
	}
}
