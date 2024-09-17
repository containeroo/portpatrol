package config

import (
	"fmt"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/containeroo/portpatrol/internal/checker"
)

const (
	envTargetName      string = "TARGET_NAME"
	envTargetAddress   string = "TARGET_ADDRESS"
	envTargetCheckType string = "TARGET_CHECK_TYPE"
	envCheckInterval   string = "CHECK_INTERVAL"
	envDialTimeout     string = "DIAL_TIMEOUT"
	envLogExtraFields  string = "LOG_EXTRA_FIELDS"

	defaultTargetCheckType checker.CheckType = checker.TCP
	defaultCheckInterval   time.Duration     = 2 * time.Second
	defaultDialTimeout     time.Duration     = 1 * time.Second
	defaultLogExtraFields  bool              = false
)

// Config holds the required environment variables.
type Config struct {
	Version         string            // The version of the application.
	TargetName      string            // The name of the target.
	TargetAddress   string            // The address of the target.
	TargetCheckType checker.CheckType // Type of check: "tcp", "http" or "icmp".
	CheckInterval   time.Duration     // The interval between connection attempts.
	DialTimeout     time.Duration     // The timeout for dialing the target.
	LogExtraFields  bool              // Whether to log the fields in the log message.
}

// ParseConfig retrieves and parses the required environment variables.
// Provides default values if the environment variables are not set.
func ParseConfig(getEnv func(string) string) (Config, error) {
	cfg := Config{
		TargetName:      getEnv(envTargetName),
		TargetAddress:   getEnv(envTargetAddress),
		TargetCheckType: defaultTargetCheckType,
		CheckInterval:   defaultCheckInterval,
		DialTimeout:     defaultDialTimeout,
		LogExtraFields:  defaultLogExtraFields,
	}

	if cfg.TargetAddress == "" {
		return Config{}, fmt.Errorf("%s environment variable is required", envTargetAddress)
	}

	if cfg.TargetName == "" {
		address := cfg.TargetAddress
		if !strings.Contains(address, "://") {
			address = fmt.Sprintf("http://%s", address) // Prepend scheme if missing to avoid url.Parse error
		}

		// Use url.Parse to handle both cases: with and without a port
		parsedURL, err := url.Parse(address)
		if err != nil {
			return Config{}, fmt.Errorf("could not parse target address: %w", err)
		}

		hostname := parsedURL.Hostname() // Extract the hostname
		if hostname == "" {
			return Config{}, fmt.Errorf("could not extract hostname from target address: %s", cfg.TargetAddress)
		}

		cfg.TargetName = hostname
	}

	// Parse the interval
	if intervalStr := getEnv(envCheckInterval); intervalStr != "" {
		interval, err := time.ParseDuration(intervalStr)
		if err != nil || interval <= 0 {
			return Config{}, fmt.Errorf("invalid %s value: %s", envCheckInterval, intervalStr)
		}
		cfg.CheckInterval = interval
	}

	// Parse the dial timeout
	if dialTimeoutStr := getEnv(envDialTimeout); dialTimeoutStr != "" {
		dialTimeout, err := time.ParseDuration(dialTimeoutStr)
		if err != nil || dialTimeout <= 0 {
			return Config{}, fmt.Errorf("invalid %s value: %s", envDialTimeout, dialTimeoutStr)
		}
		cfg.DialTimeout = dialTimeout
	}

	// Parse the log additional fields
	if logFieldsStr := getEnv(envLogExtraFields); logFieldsStr != "" {
		logExtraFields, err := strconv.ParseBool(logFieldsStr)
		if err != nil {
			return Config{}, fmt.Errorf("invalid %s value: %s", envLogExtraFields, logFieldsStr)
		}
		cfg.LogExtraFields = logExtraFields
	}

	// Resolve TargetCheckType
	if err := resolveTargetCheckType(&cfg, getEnv); err != nil {
		return Config{}, err
	}

	return cfg, nil
}

// resolveTargetCheckType handles the logic for determining the check type
func resolveTargetCheckType(cfg *Config, getEnv func(string) string) error {
	// First, check if envTargetCheckType is explicitly set
	if checkTypeStr := getEnv(envTargetCheckType); checkTypeStr != "" {
		checkType, err := checker.GetCheckTypeFromString(checkTypeStr)
		if err != nil {
			return fmt.Errorf("invalid check type from environment: %w", err)
		}
		cfg.TargetCheckType = checkType
		return nil
	}

	// If not set, try to infer from the TargetAddress scheme
	parts := strings.SplitN(cfg.TargetAddress, "://", 2) // parts[0] is the scheme, parts[1] is the address
	if len(parts) == 2 {
		checkType, err := checker.GetCheckTypeFromString(parts[0])
		if err != nil {
			return fmt.Errorf("could not infer check type from address %s: %w", cfg.TargetAddress, err)
		}
		cfg.TargetCheckType = checkType
		return nil
	}

	// Fallback to default check type if neither is set or inferred
	cfg.TargetCheckType = defaultTargetCheckType
	return nil
}
