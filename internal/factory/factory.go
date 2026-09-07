package factory

import (
	"cmp"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/containeroo/never/internal/backoff"
	"github.com/containeroo/never/internal/checker"

	"github.com/containeroo/httputils"
	"github.com/containeroo/resolver"
)

const (
	userAgentHeader = "User-Agent"
	userAgentPrefix = "never/"
)

// TargetConfig contains settings shared by every checker target.
// Exactly one checker-specific config must be set.
type TargetConfig struct {
	ID          string
	Name        string
	Address     string
	Interval    time.Duration
	MaxAttempts int
	Backoff     backoff.Mode
	MaxInterval time.Duration

	HTTP *HTTPConfig
	TCP  *TCPConfig
	ICMP *ICMPConfig
}

// HTTPConfig contains HTTP-specific target settings after CLI parsing.
type HTTPConfig struct {
	Method                string
	Headers               []string
	AllowDuplicateHeaders bool
	ExpectedStatusCodes   []string
	FollowRedirects       bool
	MaxRedirects          int
	SkipTLSVerify         bool
	Timeout               time.Duration
}

// TCPConfig contains TCP-specific target settings after CLI parsing.
type TCPConfig struct {
	Timeout time.Duration
}

// ICMPConfig contains ICMP-specific target settings after CLI parsing.
type ICMPConfig struct {
	Timeout      time.Duration
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
}

// CheckerWithInterval represents a checker with its retry settings.
type CheckerWithInterval struct {
	Interval time.Duration
	Checker  checker.Checker
	// MaxAttempts is fully resolved; -1 means endless retries.
	MaxAttempts int
	Backoff     backoff.Mode
	MaxInterval time.Duration
}

// BuildCheckers creates a list of CheckerWithInterval from typed target configuration.
func BuildCheckers(
	targets []TargetConfig,
	defaultInterval time.Duration,
	maxAttempts int,
	version string,
	showPath bool,
) ([]CheckerWithInterval, error) {
	if err := validateDefaultRetrySettings(defaultInterval, maxAttempts); err != nil {
		return nil, err
	}

	checkers := make([]CheckerWithInterval, 0, len(targets))
	for _, target := range targets {
		checkType, err := target.checkType()
		if err != nil {
			return nil, err
		}
		if err := target.validateRetrySettings(); err != nil {
			return nil, err
		}

		address := strings.TrimSpace(target.Address)
		name := cmp.Or(target.Name, target.ID)
		instance, err := buildChecker(target, checkType, name, address, version, showPath)
		if err != nil {
			return nil, fmt.Errorf("target %q: failed to create %s checker: %w", target.ID, checkType, err)
		}

		checkers = append(checkers, CheckerWithInterval{
			Interval:    cmp.Or(target.Interval, defaultInterval),
			Checker:     instance,
			MaxAttempts: cmp.Or(target.MaxAttempts, maxAttempts),
			Backoff:     cmp.Or(target.Backoff, backoff.ModeLinear),
			MaxInterval: target.MaxInterval,
		})
	}

	return checkers, nil
}

// validateDefaultRetrySettings validates retry settings shared by all targets.
func validateDefaultRetrySettings(defaultInterval time.Duration, maxAttempts int) error {
	if defaultInterval < 0 {
		return fmt.Errorf("default interval must be non-negative")
	}
	if !isValidMaxAttempts(maxAttempts) {
		return fmt.Errorf("max attempts must be -1 or positive")
	}

	return nil
}

// isValidMaxAttempts reports whether maxAttempts is endless or a positive limit.
func isValidMaxAttempts(maxAttempts int) bool {
	return maxAttempts == -1 || maxAttempts > 0
}

// validateRetrySettings validates retry overrides for one target.
func (t TargetConfig) validateRetrySettings() error {
	if t.Interval < 0 {
		return fmt.Errorf("invalid retry settings for %s", t.ID)
	}
	if t.MaxInterval < 0 {
		return fmt.Errorf("invalid retry settings for %s", t.ID)
	}
	if t.MaxAttempts < -1 {
		return fmt.Errorf("invalid retry settings for %s", t.ID)
	}
	if !isValidBackoff(t.Backoff) {
		return fmt.Errorf("invalid backoff for %s", t.ID)
	}

	return nil
}

// isValidBackoff reports whether mode is unset or one of the supported retry modes.
func isValidBackoff(mode backoff.Mode) bool {
	switch mode {
	case "", backoff.ModeLinear, backoff.ModeExponential:
		return true
	default:
		return false
	}
}

// checkType returns the configured checker type and rejects ambiguous target configuration.
func (t TargetConfig) checkType() (checker.CheckType, error) {
	count := 0
	var checkType checker.CheckType

	if t.HTTP != nil {
		count++
		checkType = checker.HTTP
	}
	if t.TCP != nil {
		count++
		checkType = checker.TCP
	}
	if t.ICMP != nil {
		count++
		checkType = checker.ICMP
	}

	if count != 1 {
		return "", fmt.Errorf("target %q must configure exactly one checker", t.ID)
	}

	return checkType, nil
}

// buildChecker converts one target's checker-specific config into a runtime checker.
func buildChecker(
	target TargetConfig,
	checkType checker.CheckType,
	name string,
	address string,
	version string,
	showPath bool,
) (checker.Checker, error) {
	switch checkType {
	case checker.HTTP:
		return buildHTTPChecker(name, address, *target.HTTP, version, showPath)
	case checker.TCP:
		return buildTCPChecker(name, address, *target.TCP)
	case checker.ICMP:
		return buildICMPChecker(name, address, *target.ICMP)
	default:
		panic("unreachable checker type")
	}
}

func buildHTTPChecker(name, address string, target HTTPConfig, version string, showPath bool) (checker.Checker, error) {
	cfg := checker.DefaultHTTPConfig()
	cfg.Method = cmp.Or(target.Method, cfg.Method)

	headers, err := createHTTPHeadersMap(target.Headers, target.AllowDuplicateHeaders)
	if err != nil {
		return nil, fmt.Errorf("invalid HTTP header: %w", err)
	}
	setDefaultUserAgent(headers, version)
	cfg.Headers = headers

	if len(target.ExpectedStatusCodes) > 0 {
		cfg.ExpectedStatusCodes, err = httputils.ParseStatusCodes(strings.Join(target.ExpectedStatusCodes, ","))
		if err != nil {
			return nil, fmt.Errorf("invalid expected status codes: %w", err)
		}
	}

	cfg.FollowRedirects = target.FollowRedirects
	cfg.MaxRedirects = target.MaxRedirects
	cfg.SkipTLSVerify = target.SkipTLSVerify
	cfg.Timeout = cmp.Or(target.Timeout, cfg.Timeout)

	instance, err := checker.NewHTTPChecker(name, address, cfg)
	if err != nil {
		return nil, err
	}

	logAddress, err := httpLogAddress(address, showPath)
	if err != nil {
		return nil, err
	}

	return checker.OverrideAddress(instance, logAddress), nil
}

func buildTCPChecker(name, address string, target TCPConfig) (checker.Checker, error) {
	cfg := checker.DefaultTCPConfig()
	cfg.Timeout = cmp.Or(target.Timeout, cfg.Timeout)
	return checker.NewTCPChecker(name, address, cfg)
}

func buildICMPChecker(name, address string, target ICMPConfig) (checker.Checker, error) {
	cfg := checker.DefaultICMPConfig()
	cfg.ReadTimeout = cmp.Or(target.ReadTimeout, target.Timeout, cfg.ReadTimeout)
	cfg.WriteTimeout = cmp.Or(target.WriteTimeout, target.Timeout, cfg.WriteTimeout)
	return checker.NewICMPChecker(name, address, cfg)
}

// httpLogAddress returns the HTTP address exposed through Checker.Address.
// User credentials are always hidden; paths are opt-in.
func httpLogAddress(address string, showPath bool) (string, error) {
	u, err := url.Parse(address)
	if err != nil {
		return "", fmt.Errorf("invalid HTTP URL: %w", err)
	}

	u.User = nil
	if !showPath {
		u.Path = ""
		u.RawPath = ""
		u.RawQuery = ""
		u.ForceQuery = false
		u.Fragment = ""
		u.RawFragment = ""
	}

	return u.String(), nil
}

// setDefaultUserAgent adds the application user agent unless the target already configured one.
func setDefaultUserAgent(headers http.Header, version string) {
	if headers.Get(userAgentHeader) != "" {
		return
	}

	headers.Set(userAgentHeader, userAgentPrefix+version)
}

// createHTTPHeadersMap creates an http.Header from a slice of key=value strings.
// If allowDuplicateHeaders is true, headers with the same key will be appended.
func createHTTPHeadersMap(headers []string, allowDuplicateHeaders bool) (http.Header, error) {
	headersMap := make(http.Header)

	if headers == nil {
		return headersMap, nil
	}

	for _, header := range headers {
		parts := strings.SplitN(header, "=", 2)

		if len(parts) != 2 || parts[0] == "" {
			return nil, fmt.Errorf("invalid header format: %q", header)
		}

		key := http.CanonicalHeaderKey(strings.TrimSpace(parts[0]))
		value := strings.TrimSpace(parts[1])

		resolved, err := resolver.ResolveVariable(value)
		if err != nil {
			return nil, fmt.Errorf("failed to resolve variable in header: %w", err)
		}

		if _, exists := headersMap[key]; exists && !allowDuplicateHeaders {
			return nil, fmt.Errorf("duplicate header: %q", header)
		}

		if allowDuplicateHeaders {
			headersMap.Add(key, resolved)
			continue
		}

		headersMap.Set(key, resolved)
	}

	return headersMap, nil
}
