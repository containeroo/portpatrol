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
	envTargetName      = "TARGET_NAME"
	envTargetAddress   = "TARGET_ADDRESS"
	envTargetCheckType = "TARGET_CHECK_TYPE"
	envCheckInterval   = "CHECK_INTERVAL"
	envDialTimeout     = "DIAL_TIMEOUT"
	envLogExtraFields  = "LOG_EXTRA_FIELDS"

	defaultTargetCheckType = "tcp"
	defaultCheckInterval   = 2 * time.Second
	defaultDialTimeout     = 1 * time.Second
	defaultLogExtraFields  = false
)

// Config holds the required environment variables.
type Config struct {
	Version         string        // The version of the application.
	TargetName      string        // The name of the target.
	TargetAddress   string        // The address of the target.
	TargetCheckType string        // Type of check: "tcp" or "http"
	CheckInterval   time.Duration // The interval between connection attempts.
	DialTimeout     time.Duration // The timeout for dialing the target.
	LogExtraFields  bool          // Whether to log the fields in the log message.
}

// ParseConfig retrieves and parses the required environment variables.
// Provides default values if the environment variables are not set.
func ParseConfig(getEnv func(string) string) (Config, error) {
	cfg := Config{
		TargetName:      getEnv(envTargetName),
		TargetAddress:   getEnv(envTargetAddress),
		TargetCheckType: getEnv(envTargetCheckType),
		CheckInterval:   defaultCheckInterval,
		DialTimeout:     defaultDialTimeout,
		LogExtraFields:  defaultLogExtraFields,
	}

	if cfg.TargetAddress == "" {
		return Config{}, fmt.Errorf("%s environment variable is required", envTargetAddress)
	}

	if cfg.TargetName == "" {
		// Prepend scheme if missing to avoid url.Parse error
		address := cfg.TargetAddress
		if !strings.Contains(address, "://") {
			address = fmt.Sprintf("http://%s", address)
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
		logAdditionalFields, err := strconv.ParseBool(logFieldsStr)
		if err != nil {
			return Config{}, fmt.Errorf("invalid %s value: %s", envLogExtraFields, logFieldsStr)
		}
		cfg.LogExtraFields = logAdditionalFields
	}

	// Infer TargetCheckType if not provided
	if cfg.TargetCheckType == "" {
		checkType, err := checker.InferCheckType(cfg.TargetAddress)
		if err != nil {
			return Config{}, fmt.Errorf("could not infer check type for address %s: %w", cfg.TargetAddress, err)
		}
		if checkType == "" {
			checkType = defaultTargetCheckType
		}
		cfg.TargetCheckType = checkType
	}

	// Validate the check type
	if !checker.IsValidCheckType(cfg.TargetCheckType) {
		return Config{}, fmt.Errorf("unsupported check type: %s", cfg.TargetCheckType)
	}

	return cfg, nil
}
