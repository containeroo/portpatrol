package parser

import (
	"fmt"
	"strings"
	"time"

	"github.com/containeroo/portpatrol/internal/checker"
)

const (
	ParamPrefix   string = "target"
	ParamType     string = "type"
	ParamName     string = "name"
	ParamAddress  string = "address"
	ParamInterval string = "interval"
)

// TargetChecker represents a checker with its interval.
type TargetChecker struct {
	Interval time.Duration
	Checker  checker.Checker
}

// InitializeTargetCheckers creates a slice of TargetChecker based on the provided target configurations.
func InitializeTargetCheckers(targets map[string]map[string]string, defaultInterval time.Duration) ([]TargetChecker, error) {
	var targetCheckers []TargetChecker

	for targetName, params := range targets {
		address, ok := params[ParamAddress]
		if !ok || address == "" {
			return nil, fmt.Errorf("missing %q for target %q", ParamAddress, targetName)
		}

		// Determine the check type
		checkTypeStr, ok := params[ParamType]
		if !ok || checkTypeStr == "" {
			// Try to infer the type from the address scheme
			address := params[ParamAddress]
			parts := strings.SplitN(address, "://", 2)
			if len(parts) != 2 {
				return nil, fmt.Errorf("missing %q parameter for target %q", ParamType, targetName)
			}
			checkTypeStr = parts[0]
		}

		checkType, err := checker.GetCheckTypeFromString(checkTypeStr)
		if err != nil {
			return nil, fmt.Errorf("unsupported check type %q for target %q", checkTypeStr, targetName)
		}

		// Use identifier as name if name not explicitly set
		name := targetName
		if n, ok := params[ParamName]; ok && n != "" {
			name = n
		}

		// Get interval from parameters or use default
		interval := defaultInterval
		if intervalStr, ok := params[ParamInterval]; ok && intervalStr != "" {
			interval, err = time.ParseDuration(intervalStr)
			if err != nil {
				return nil, fmt.Errorf("invalid %q for target '%s': %w", ParamInterval, targetName, err)
			}
		}

		// Remove common parameters from params map
		delete(params, ParamType)
		delete(params, ParamName)
		delete(params, ParamAddress)
		delete(params, ParamInterval)

		// Collect functional options based on the check type
		var options []checker.Option
		switch checkType {
		case checker.HTTP:
			httpOpts, err := parseHTTPCheckerOptions(params)
			if err != nil {
				return nil, fmt.Errorf("failed to parse HTTP options for target %q: %w", targetName, err)
			}
			options = append(options, httpOpts...)
		case checker.TCP:
			tcpOpts, err := parseTCPCheckerOptions(params)
			if err != nil {
				return nil, fmt.Errorf("failed to parse TCP options for target '%s': %w", targetName, err)
			}
			options = append(options, tcpOpts...)
		case checker.ICMP:
			icmpOpts, err := parseICMPCheckerOptions(params)
			if err != nil {
				return nil, fmt.Errorf("failed to parse ICMP options for target %q: %w", targetName, err)
			}
			options = append(options, icmpOpts...)
		default:
			return nil, fmt.Errorf("unsupported check type %q for target %q", checkTypeStr, targetName)
		}

		// Create the checker using the functional options
		chk, err := checker.NewChecker(checkType, name, address, options...)
		if err != nil {
			return nil, fmt.Errorf("failed to create checker for target %q: %w", targetName, err)
		}

		targetCheckers = append(targetCheckers, TargetChecker{
			Interval: interval,
			Checker:  chk,
		})
	}

	return targetCheckers, nil
}
