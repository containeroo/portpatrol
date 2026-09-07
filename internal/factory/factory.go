package factory

import (
	"cmp"
	"fmt"
	"time"

	"github.com/containeroo/never/internal/backoff"
	"github.com/containeroo/never/internal/checker"
)

// TargetConfig contains settings shared by every checker target.
type TargetConfig struct {
	ID          string
	Name        string
	Address     string
	Interval    time.Duration
	MaxAttempts int
	Backoff     backoff.Mode
	MaxInterval time.Duration
	Config      checker.Config
}

// CheckerWithInterval represents a checker with its retry settings.
type CheckerWithInterval struct {
	Interval    time.Duration
	Checker     checker.Checker
	MaxAttempts int // MaxAttempts is fully resolved; -1 means endless retries.
	Backoff     backoff.Mode
	MaxInterval time.Duration
}

// BuildCheckers creates a list of CheckerWithInterval from typed target configuration.
func BuildCheckers(
	targets []TargetConfig,
	defaultInterval time.Duration,
	maxAttempts int,
) ([]CheckerWithInterval, error) {
	checkers := make([]CheckerWithInterval, 0, len(targets))
	for _, target := range targets {
		if target.Config == nil {
			return nil, fmt.Errorf("target %q has no checker config", target.ID)
		}

		name := cmp.Or(target.Name, target.ID)
		instance, err := target.Config.NewChecker(name, target.Address)
		if err != nil {
			return nil, fmt.Errorf("target %q: failed to create checker: %w", target.ID, err)
		}

		checkers = append(checkers, CheckerWithInterval{
			Checker:     instance,
			Interval:    cmp.Or(target.Interval, defaultInterval),
			MaxAttempts: cmp.Or(target.MaxAttempts, maxAttempts),
			Backoff:     cmp.Or(target.Backoff, backoff.ModeLinear),
			MaxInterval: target.MaxInterval,
		})
	}

	return checkers, nil
}
