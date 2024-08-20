package config

import (
	"fmt"
	"strconv"
	"time"

	"github.com/containeroo/toast/pkg/checker"
)

const (
	envTargetName          = "TARGET_NAME"
	envTargetAddress       = "TARGET_ADDRESS"
	envInterval            = "INTERVAL"
	envDialTimeout         = "DIAL_TIMEOUT"
	envCheckType           = "CHECK_TYPE"
	envLogAdditionalFields = "LOG_ADDITIONAL_FIELDS"

	defaultInterval    = 2 * time.Second
	defaultDialTimeout = 1 * time.Second
)

// Config holds the required environment variables.
type Config struct {
	TargetName          string        // The name of the target.
	TargetAddress       string        // The address of the target.
	Interval            time.Duration // The interval between connection attempts.
	DialTimeout         time.Duration // The timeout for dialing the target.
	CheckType           string        // Type of check: "tcp" or "http"
	LogAdditionalFields bool          // Whether to log the fields in the log message.
}

// ParseConfig retrieves and parses the required environment variables.
// Provides default values if the environment variables are not set.
func ParseConfig(getenv func(string) string) (Config, error) {
	cfg := Config{
		TargetName:          getenv(envTargetName),
		TargetAddress:       getenv(envTargetAddress),
		Interval:            defaultInterval,
		DialTimeout:         defaultDialTimeout,
		CheckType:           getenv(envCheckType),
		LogAdditionalFields: false,
	}

	if cfg.TargetAddress == "" {
		return Config{}, fmt.Errorf("%s environment variable is required", envTargetAddress)
	}

	if intervalStr := getenv(envInterval); intervalStr != "" {
		interval, err := time.ParseDuration(intervalStr)
		if err != nil || interval <= 0 {
			return Config{}, fmt.Errorf("invalid %s value: %s", envInterval, intervalStr)
		}
		cfg.Interval = interval
	}

	if dialTimeoutStr := getenv(envDialTimeout); dialTimeoutStr != "" {
		dialTimeout, err := time.ParseDuration(dialTimeoutStr)
		if err != nil || dialTimeout <= 0 {
			return Config{}, fmt.Errorf("invalid %s value: %s", envDialTimeout, dialTimeoutStr)
		}
		cfg.DialTimeout = dialTimeout
	}

	if logFieldsStr := getenv(envLogAdditionalFields); logFieldsStr != "" {
		logAdditionalFields, err := strconv.ParseBool(logFieldsStr)
		if err != nil {
			return Config{}, fmt.Errorf("invalid %s value: %s", envLogAdditionalFields, logFieldsStr)
		}
		cfg.LogAdditionalFields = logAdditionalFields
	}

	if cfg.CheckType == "" {
		checkType, err := checker.InferCheckType(cfg.TargetAddress)
		if err != nil {
			return Config{}, fmt.Errorf("could not infer check type for address %s: %w", cfg.TargetAddress, err)
		}
		cfg.CheckType = checkType
	}

	if !isValidCheckType(cfg.CheckType) {
		return Config{}, fmt.Errorf("unsupported check type: %s", cfg.CheckType)
	}

	return cfg, nil
}

// isValidCheckType validates if the check type is supported.
func isValidCheckType(checkType string) bool {
	return checkType == "tcp" || checkType == "http"
}
