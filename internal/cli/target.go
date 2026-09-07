package cli

import (
	"fmt"
	"strings"
	"time"

	"github.com/containeroo/never/internal/backoff"
	"github.com/containeroo/never/internal/checker"
	"github.com/containeroo/never/internal/factory"
	"github.com/containeroo/resolver"
	"github.com/containeroo/tinyflags"
)

// parseTargetConfigs converts parsed dynamic flag groups into typed target config.
func parseTargetConfigs(dynamicGroups []*tinyflags.DynamicGroup) ([]factory.TargetConfig, error) {
	var targets []factory.TargetConfig

	for _, group := range dynamicGroups {
		checkType, err := checker.ParseCheckType(group.Name())
		if err != nil {
			return nil, err
		}

		for _, id := range group.Instances() {
			address, err := resolveTargetAddress(
				tinyflags.GetOrDefaultDynamic[string](group, id, "address"),
				checkType,
			)
			if err != nil {
				return nil, fmt.Errorf("%s target %q: %w", checkType, id, err)
			}

			target := factory.TargetConfig{
				ID:          id,
				Name:        tinyflags.GetOrDefaultDynamic[string](group, id, "name"),
				Address:     address,
				Interval:    getDynamicDuration(group, id, "interval"),
				MaxAttempts: getDynamicInt(group, id, "max-attempts"),
				Backoff:     getDynamicBackoffMode(group, id, "backoff"),
				MaxInterval: getDynamicDuration(group, id, "max-interval"),
			}

			applyCheckerConfig(&target, group, id, checkType)
			targets = append(targets, target)
		}
	}

	return targets, nil
}

// resolveTargetAddress resolves a target address and validates the concrete value.
func resolveTargetAddress(value string, checkType checker.CheckType) (string, error) {
	address, err := resolver.ResolveVariable(strings.TrimSpace(value))
	if err != nil {
		return "", fmt.Errorf("failed to resolve address: %w", err)
	}

	address = strings.TrimSpace(address)
	if err := validateResolvedTargetAddress(address, checkType); err != nil {
		return "", err
	}

	return address, nil
}

// validateResolvedTargetAddress dispatches concrete address validation by checker type.
func validateResolvedTargetAddress(address string, checkType checker.CheckType) error {
	switch checkType {
	case checker.HTTP:
		return validateResolvedHTTPAddress(address)
	case checker.TCP:
		return validateResolvedTCPAddress(address)
	case checker.ICMP:
		return validateResolvedICMPAddress(address)
	default:
		panic("unreachable checker type")
	}
}

// applyCheckerConfig attaches the checker-specific settings for one target.
func applyCheckerConfig(target *factory.TargetConfig, group *tinyflags.DynamicGroup, id string, checkType checker.CheckType) {
	switch checkType {
	case checker.HTTP:
		target.HTTP = &factory.HTTPConfig{
			Method:                tinyflags.GetOrDefaultDynamic[string](group, id, "method"),
			Headers:               tinyflags.GetOrDefaultDynamic[[]string](group, id, "header"),
			AllowDuplicateHeaders: tinyflags.GetOrDefaultDynamic[bool](group, id, "allow-duplicate-headers"),
			ExpectedStatusCodes:   tinyflags.GetOrDefaultDynamic[[]string](group, id, "expected-status-codes"),
			FollowRedirects:       tinyflags.GetOrDefaultDynamic[bool](group, id, "follow-redirects"),
			MaxRedirects:          tinyflags.GetOrDefaultDynamic[int](group, id, "max-redirects"),
			SkipTLSVerify:         tinyflags.GetOrDefaultDynamic[bool](group, id, "skip-tls-verify"),
			Timeout:               tinyflags.GetOrDefaultDynamic[time.Duration](group, id, "timeout"),
		}
	case checker.TCP:
		target.TCP = &factory.TCPConfig{
			Timeout: tinyflags.GetOrDefaultDynamic[time.Duration](group, id, "timeout"),
		}
	case checker.ICMP:
		target.ICMP = &factory.ICMPConfig{
			Timeout:      tinyflags.GetOrDefaultDynamic[time.Duration](group, id, "timeout"),
			ReadTimeout:  getDynamicDuration(group, id, "read-timeout"),
			WriteTimeout: getDynamicDuration(group, id, "write-timeout"),
		}
	}
}

// getDynamicDuration returns a dynamic duration flag value or zero when unset.
func getDynamicDuration(group *tinyflags.DynamicGroup, id, name string) time.Duration {
	v, _ := tinyflags.GetDynamic[time.Duration](group, id, name)
	return v
}

// getDynamicInt returns a dynamic int flag value or zero when unset.
func getDynamicInt(group *tinyflags.DynamicGroup, id, name string) int {
	v, _ := tinyflags.GetDynamic[int](group, id, name)
	return v
}

// getDynamicBackoffMode returns the configured backoff mode or the zero value when unset.
func getDynamicBackoffMode(group *tinyflags.DynamicGroup, id, name string) backoff.Mode {
	return tinyflags.GetOrDefaultDynamic[backoff.Mode](group, id, name)
}
