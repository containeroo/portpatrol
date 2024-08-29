package runner

import (
	"context"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/containeroo/portpatrol/internal/checker"
	"github.com/containeroo/portpatrol/internal/config"
	"github.com/containeroo/portpatrol/internal/logger"
)

func TestLoopUntilReadyHTTP(t *testing.T) {
	t.Parallel()

	t.Run("HTTP target is ready", func(t *testing.T) {
		t.Parallel()

		server := &http.Server{Addr: ":9082"}
		http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		})
		go func() {
			// Run the server in a goroutine so that it does not block the test
			_ = server.ListenAndServe()
		}()

		defer server.Close()

		cfg := config.Config{
			TargetName:    "HTTPServer",
			TargetAddress: "http://localhost:9082/",
			CheckInterval: 50 * time.Millisecond,
			DialTimeout:   50 * time.Millisecond,
		}

		mockEnv := func(key string) string {
			env := map[string]string{
				"METHOD":            "GET",
				"EXPECTED_STATUSES": "200",
			}
			return env[key]
		}

		checker, err := checker.NewHTTPChecker(cfg.TargetName, cfg.TargetAddress, cfg.DialTimeout, mockEnv)
		if err != nil {
			t.Fatalf("Failed to create HTTPChecker: %q", err)
		}

		var stdOut strings.Builder
		logger := slog.New(slog.NewTextHandler(&stdOut, nil))

		ctx, cancel := context.WithTimeout(context.Background(), cfg.CheckInterval*4)
		defer cancel()

		err = LoopUntilReady(ctx, cfg.CheckInterval, checker, logger)
		if err != nil {
			t.Errorf("Unexpected error: %q", err)
		}

		expected := "HTTPServer is ready ✓"
		if !strings.Contains(stdOut.String(), expected) {
			t.Errorf("Expected output to contain %q but got %q", expected, stdOut.String())
		}
	})

	t.Run("HTTP Target with path is ready", func(t *testing.T) {
		t.Parallel()

		server := &http.Server{Addr: ":9081"}
		http.HandleFunc("/ping", func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		})
		go func() {
			// Run the server in a goroutine so that it does not block the test
			_ = server.ListenAndServe()
		}()
		defer server.Close()

		cfg := config.Config{
			TargetName:    "HTTPServer",
			TargetAddress: "http://localhost:9081/ping",
			CheckInterval: 50 * time.Millisecond,
			DialTimeout:   50 * time.Millisecond,
		}

		mockEnv := func(key string) string {
			env := map[string]string{
				"METHOD":            "GET",
				"EXPECTED_STATUSES": "200",
			}
			return env[key]
		}

		checker, err := checker.NewHTTPChecker(cfg.TargetName, cfg.TargetAddress, cfg.DialTimeout, mockEnv)
		if err != nil {
			t.Fatalf("Failed to create HTTPChecker: %q", err)
		}

		var stdOut strings.Builder
		logger := slog.New(slog.NewTextHandler(&stdOut, nil))

		ctx, cancel := context.WithTimeout(context.Background(), cfg.CheckInterval*4)
		defer cancel()

		err = LoopUntilReady(ctx, cfg.CheckInterval, checker, logger)
		if err != nil {
			t.Errorf("Unexpected error: %q", err)
		}

		expected := "HTTPServer is ready ✓"
		if !strings.Contains(stdOut.String(), expected) {
			t.Errorf("Expected output to contain %q but got %q", expected, stdOut.String())
		}
	})

	t.Run("Successful HTTP target run after 3 attempts", func(t *testing.T) {
		t.Parallel()

		cfg := config.Config{
			TargetName:      "HTTPServer",
			TargetAddress:   "http://localhost:6081/success",
			CheckInterval:   500 * time.Millisecond,
			DialTimeout:     500 * time.Millisecond,
			TargetCheckType: "http",
			LogExtraFields:  true,
			Version:         "1.0.0",
		}

		parsedURL, err := url.Parse(cfg.TargetAddress)
		if err != nil {
			t.Fatalf("Failed to parse URL: %q", err)
		}

		host := parsedURL.Host

		_, addressPort, err := net.SplitHostPort(host)
		if err != nil {
			t.Fatalf("Failed to split host and port: %q", err)
		}

		var wg sync.WaitGroup
		wg.Add(1)

		server := &http.Server{Addr: fmt.Sprintf(":%s", addressPort)}
		http.HandleFunc("/success", func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		})

		go func() {
			// Run the server in a goroutine so that it does not block the test
			// Wait 3 times the interval before starting the server
			defer wg.Done() // Mark the WaitGroup as done when the goroutine completes
			time.Sleep(cfg.CheckInterval * 3)
			err := server.ListenAndServe()

			if err != nil && err != http.ErrServerClosed { // After Server.Shutdown the returned error is ErrServerClosed.
				panic("failed to listen: " + err.Error())
			}
			time.Sleep(200 * time.Millisecond) // Ensure runloop get a successful attempt
		}()

		ctx, cancel := context.WithTimeout(context.Background(), cfg.CheckInterval*4)
		defer cancel()

		go func() {
			// Wait for the context to be canceled
			<-ctx.Done()
			_ = server.Shutdown(context.Background()) // Gracefully shutdown the server
		}()

		mockEnv := func(key string) string {
			env := map[string]string{
				"METHOD":            "GET",
				"EXPECTED_STATUSES": "200",
			}
			return env[key]
		}

		checker, err := checker.NewHTTPChecker(cfg.TargetName, cfg.TargetAddress, cfg.DialTimeout, mockEnv)
		if err != nil {
			t.Fatalf("Failed to create HTTPChecker: %q", err)
		}

		var stdOut strings.Builder
		logger := logger.SetupLogger(cfg, &stdOut)

		err = LoopUntilReady(ctx, cfg.CheckInterval, checker, logger)
		if err != nil {
			t.Errorf("Unexpected error: %q", err)
		}

		wg.Wait() // Ensure server is closed after the test

		stdOutEntries := strings.Split(strings.TrimSpace(stdOut.String()), "\n")
		// output must be:
		// 0: Waiting for HTTPServer to become ready...
		// 1: HTTPServer is not ready ✗
		// 2: HTTPServer is not ready ✗
		// 3: HTTPServer is not ready ✗
		// 4: HTTPServer is ready ✓
		lenExpectedOuts := 5

		if len(stdOutEntries) != lenExpectedOuts {
			t.Errorf("Expected output to contain '%d' lines but got '%d'.", lenExpectedOuts, len(stdOutEntries))
		}

		// First log entry: "Waiting for HTTPServer to become ready..."
		expected := fmt.Sprintf("Waiting for %s to become ready...", cfg.TargetName)
		if !strings.Contains(stdOutEntries[0], expected) {
			t.Errorf("Expected output to contain %q but got %q", expected, stdOutEntries[0])
		}

		from := 1
		to := 3
		for i := from; i < to; i++ {
			expected := fmt.Sprintf("%s is not ready ✗", cfg.TargetName)
			if !strings.Contains(stdOutEntries[i], expected) {
				t.Errorf("Expected output to contain %q but got %q", expected, stdOutEntries[i])
			}

			expected = fmt.Sprintf("error=\"Get \\\"%s\\\": dial tcp [::1]:%s: connect: connection refused\"", cfg.TargetAddress, addressPort)
			if !strings.Contains(stdOutEntries[i], expected) {
				t.Errorf("Expected output to contain %q but got %q", expected, stdOutEntries[i])
			}
		}

		// Last log entry: "HTTPServer is ready ✓"
		expected = fmt.Sprintf("%s is ready ✓", cfg.TargetName)
		if !strings.Contains(stdOutEntries[lenExpectedOuts-1], expected) { // lenExpectedOuts -1 = last element
			t.Errorf("Expected output to contain %q but got %q", expected, stdOutEntries[1])
		}

		// Check version in the last entry
		expected = fmt.Sprintf("version=%s", cfg.Version)
		if !strings.Contains(stdOutEntries[lenExpectedOuts-1], expected) { // lenExpectedOuts -1 = last element
			t.Errorf("Expected output to contain %q but got %q", expected, stdOutEntries[1])
		}
	})

	t.Run("Successful HTTP target run after 3 wrong responses", func(t *testing.T) {
		t.Parallel()

		cfg := config.Config{
			TargetName:      "HTTPServer",
			TargetAddress:   "http://localhost:2081/wrong",
			CheckInterval:   500 * time.Millisecond,
			DialTimeout:     500 * time.Millisecond,
			TargetCheckType: "http",
			LogExtraFields:  true,
			Version:         "1.0.0",
		}

		parsedURL, err := url.Parse(cfg.TargetAddress)
		if err != nil {
			t.Fatalf("Failed to parse URL: %q", err)
		}

		host := parsedURL.Host

		_, addressPort, err := net.SplitHostPort(host)
		if err != nil {
			t.Fatalf("Failed to split host and port: %q", err)
		}

		counter := 0

		server := &http.Server{Addr: fmt.Sprintf(":%s", addressPort)}
		http.HandleFunc("/wrong", func(w http.ResponseWriter, r *http.Request) {
			if counter < 3 {
				counter++
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			w.WriteHeader(http.StatusOK)
		})

		go func() {
			// Run the server in a goroutine so that it does not block the test
			_ = server.ListenAndServe()
		}()

		ctx, cancel := context.WithTimeout(context.Background(), cfg.CheckInterval*4)
		defer cancel()

		mockEnv := func(key string) string {
			env := map[string]string{
				"METHOD":            "GET",
				"EXPECTED_STATUSES": "200",
			}
			return env[key]
		}

		checker, err := checker.NewHTTPChecker(cfg.TargetName, cfg.TargetAddress, cfg.DialTimeout, mockEnv)
		if err != nil {
			t.Fatalf("Failed to create HTTPChecker: %q", err)
		}

		var stdOut strings.Builder
		logger := logger.SetupLogger(cfg, &stdOut)

		err = LoopUntilReady(ctx, cfg.CheckInterval, checker, logger)
		if err != nil {
			t.Errorf("Unexpected error: %q", err)
		}

		stdOutEntries := strings.Split(strings.TrimSpace(stdOut.String()), "\n")
		// output must be:
		// 0: Waiting for HTTPServer to become ready...
		// 1: HTTPServer is not ready ✗
		// 2: HTTPServer is not ready ✗
		// 3: HTTPServer is not ready ✗
		// 4: HTTPServer is ready ✓
		lenExpectedOuts := 5

		if len(stdOutEntries) != lenExpectedOuts {
			t.Errorf("Expected output to contain '%d' lines but got '%d'.", lenExpectedOuts, len(stdOutEntries))
		}

		// First log entry: "Waiting for HTTPServer to become ready..."
		expected := fmt.Sprintf("Waiting for %s to become ready...", cfg.TargetName)
		if !strings.Contains(stdOutEntries[0], expected) {
			t.Errorf("Expected output to contain %q but got %q", expected, stdOutEntries[0])
		}

		from := 1
		to := 3
		for i := from; i < to; i++ {
			expected := fmt.Sprintf("%s is not ready ✗", cfg.TargetName)
			if !strings.Contains(stdOutEntries[i], expected) {
				t.Errorf("Expected output to contain %q but got %q", expected, stdOutEntries[i])
			}

			expected = "error=\"unexpected status code: got 500, expected one of [200]\""
			if !strings.Contains(stdOutEntries[i], expected) {
				t.Errorf("Expected output to contain %q but got %q", expected, stdOutEntries[i])
			}
		}

		// Last log entry: "HTTPServer is ready ✓"
		expected = fmt.Sprintf("%s is ready ✓", cfg.TargetName)
		if !strings.Contains(stdOutEntries[lenExpectedOuts-1], expected) { // lenExpectedOuts -1 = last element
			t.Errorf("Expected output to contain %q but got %q", expected, stdOutEntries[1])
		}

		// Check version in the last entry
		expected = fmt.Sprintf("version=%s", cfg.Version)
		if !strings.Contains(stdOutEntries[lenExpectedOuts-1], expected) { // lenExpectedOuts -1 = last element
			t.Errorf("Expected output to contain %q but got %q", expected, stdOutEntries[1])
		}
	})

	t.Run("HTTP target context cancled", func(t *testing.T) {
		t.Parallel()

		cfg := config.Config{
			TargetName:      "HTTPServer",
			TargetAddress:   "http://localhost:7083/fail",
			CheckInterval:   50 * time.Millisecond,
			DialTimeout:     50 * time.Millisecond,
			TargetCheckType: "http",
		}

		mockEnv := func(key string) string {
			env := map[string]string{
				"METHOD":            "GET",
				"EXPECTED_STATUSES": "200",
			}
			return env[key]
		}

		checker, err := checker.NewHTTPChecker(cfg.TargetName, cfg.TargetAddress, cfg.DialTimeout, mockEnv)
		if err != nil {
			t.Fatalf("Failed to create HTTPChecker: %q", err)
		}

		var stdOut strings.Builder
		logger := slog.New(slog.NewTextHandler(&stdOut, nil))

		ctx, cancel := context.WithTimeout(context.Background(), cfg.CheckInterval*4)

		go func() {
			// Wait for the context to be canceled
			time.Sleep(100 * time.Millisecond)
			cancel()
		}()

		err = LoopUntilReady(ctx, cfg.CheckInterval, checker, logger)
		if err != nil && err != context.Canceled {
			t.Errorf("Expected context canceled error, got %q", err)
		}

		expected := fmt.Sprintf("Waiting for %s to become ready...", cfg.TargetName)
		if !strings.Contains(stdOut.String(), expected) {
			t.Errorf("Expected output to contain %q but got %q", expected, stdOut.String())
		}

		expected = fmt.Sprintf("%s is not ready ✗", cfg.TargetName)
		if !strings.Contains(stdOut.String(), expected) {
			t.Errorf("Expected output to contain %q but got %q", expected, stdOut.String())
		}
	})
}

func TestLoopUntilReadyTCP(t *testing.T) {
	t.Parallel()

	t.Run("TCP Target is ready", func(t *testing.T) {
		t.Parallel()

		listener, err := net.Listen("tcp", "localhost:5082")
		if err != nil {
			t.Fatalf("Failed to start TCP server: %q", err)
		}
		defer listener.Close()

		cfg := config.Config{
			TargetName:      "TCPServer",
			TargetAddress:   listener.Addr().String(),
			CheckInterval:   50 * time.Millisecond,
			DialTimeout:     50 * time.Millisecond,
			TargetCheckType: "tcp",
		}

		mockEnv := func(key string) string {
			return ""
		}

		checker, err := checker.NewTCPChecker(cfg.TargetName, cfg.TargetAddress, cfg.DialTimeout, mockEnv)
		if err != nil {
			t.Fatalf("Failed to create TCPChecker: %q", err)
		}

		var stdOut strings.Builder
		logger := slog.New(slog.NewTextHandler(&stdOut, nil))

		ctx, cancel := context.WithTimeout(context.Background(), cfg.CheckInterval*4)
		defer cancel()

		err = LoopUntilReady(ctx, cfg.CheckInterval, checker, logger)
		if err != nil {
			t.Errorf("Unexpected error: %q", err)
		}

		expected := "TCPServer is ready ✓"
		if !strings.Contains(stdOut.String(), expected) {
			t.Errorf("Expected output to contain %q but got %q", expected, stdOut.String())
		}
	})

	t.Run("TCP Target is ready without type", func(t *testing.T) {
		t.Parallel()

		listener, err := net.Listen("tcp", "localhost:7082")
		if err != nil {
			t.Fatalf("Failed to start TCP server: %q", err)
		}
		defer listener.Close()

		cfg := config.Config{
			TargetName:    "TCPServer",
			TargetAddress: fmt.Sprintf("tcp://%s", listener.Addr().String()),
			CheckInterval: 50 * time.Millisecond,
			DialTimeout:   50 * time.Millisecond,
		}

		mockEnv := func(key string) string {
			return ""
		}

		checker, err := checker.NewTCPChecker(cfg.TargetName, cfg.TargetAddress, cfg.DialTimeout, mockEnv)
		if err != nil {
			t.Fatalf("Failed to create TCPChecker: %q", err)
		}

		var stdOut strings.Builder
		logger := slog.New(slog.NewTextHandler(&stdOut, nil))

		ctx, cancel := context.WithTimeout(context.Background(), cfg.CheckInterval*4)
		defer cancel()

		err = LoopUntilReady(ctx, cfg.CheckInterval, checker, logger)
		if err != nil {
			t.Errorf("Unexpected error: %q", err)
		}

		expected := "TCPServer is ready ✓"
		if !strings.Contains(stdOut.String(), expected) {
			t.Errorf("Expected output to contain %q but got %q", expected, stdOut.String())
		}
	})

	t.Run("Successful TCP target run after 3 attempts", func(t *testing.T) {
		t.Parallel()

		cfg := config.Config{
			TargetName:      "TCPServer",
			TargetAddress:   "localhost:5081",
			CheckInterval:   500 * time.Millisecond,
			DialTimeout:     500 * time.Millisecond,
			TargetCheckType: "tcp",
			LogExtraFields:  true,
			Version:         "1.0.0",
		}

		addressPort := strings.Split(cfg.TargetAddress, ":")[1]

		var wg sync.WaitGroup
		wg.Add(1)

		var lis net.Listener

		go func() {
			// Run the server in a goroutine so that it does not block the test
			// Wait 3 times the interval before starting the server
			defer wg.Done() // Mark the WaitGroup as done when the goroutine completes
			time.Sleep(cfg.CheckInterval * 3)
			var err error
			lis, err = net.Listen("tcp", cfg.TargetAddress)
			if err != nil {
				panic("failed to listen: " + err.Error())
			}
			time.Sleep(200 * time.Millisecond) // Ensure runloop get a successful attempt
		}()

		ctx, cancel := context.WithTimeout(context.Background(), cfg.CheckInterval*4)
		defer cancel()

		mockEnv := func(key string) string {
			return ""
		}

		checker, err := checker.NewTCPChecker(cfg.TargetName, cfg.TargetAddress, cfg.DialTimeout, mockEnv)
		if err != nil {
			t.Fatalf("Failed to create HTTPChecker: %q", err)
		}

		var stdOut strings.Builder
		logger := logger.SetupLogger(cfg, &stdOut)

		err = LoopUntilReady(ctx, cfg.CheckInterval, checker, logger)
		if err != nil {
			t.Errorf("Unexpected error: %q", err)
		}

		wg.Wait()
		defer lis.Close() // listener must be closed after waiting group is done

		stdOutEntries := strings.Split(strings.TrimSpace(stdOut.String()), "\n")
		// output must be:
		// 0: Waiting for TCPServer to become ready...
		// 1: TCPServer is not ready ✗
		// 2: TCPServer is not ready ✗
		// 3: TCPServer is not ready ✗
		// 4: TCPServer is ready ✓
		lenExpectedOuts := 5

		if len(stdOutEntries) != lenExpectedOuts {
			t.Errorf("Expected output to contain '%d' lines but got '%d'.", lenExpectedOuts, len(stdOutEntries))
		}

		// First log entry: "Waiting for HTTPServer to become ready..."
		expected := fmt.Sprintf("Waiting for %s to become ready...", cfg.TargetName)
		if !strings.Contains(stdOutEntries[0], expected) {
			t.Errorf("Expected output to contain %q but got %q", expected, stdOutEntries[0])
		}

		from := 1
		to := 3
		for i := from; i < to; i++ {
			expected := fmt.Sprintf("%s is not ready ✗", cfg.TargetName)
			if !strings.Contains(stdOutEntries[i], expected) {
				t.Errorf("Expected output to contain %q but got %q", expected, stdOutEntries[i])
			}

			expected = fmt.Sprintf("error=\"dial tcp [::1]:%s: connect: connection refused\"", addressPort)
			if !strings.Contains(stdOutEntries[i], expected) {
				t.Errorf("Expected output to contain %q but got %q", expected, stdOutEntries[i])
			}
		}

		// Last log entry: "HTTPServer is ready ✓"
		expected = fmt.Sprintf("%s is ready ✓", cfg.TargetName)
		if !strings.Contains(stdOutEntries[lenExpectedOuts-1], expected) { // lenExpectedOuts -1 = last element
			t.Errorf("Expected output to contain %q but got %q", expected, stdOutEntries[1])
		}

		// Check version in the last entry
		expected = fmt.Sprintf("version=%s", cfg.Version)
		if !strings.Contains(stdOutEntries[lenExpectedOuts-1], expected) { // lenExpectedOuts -1 = last element
			t.Errorf("Expected output to contain %q but got %q", expected, stdOutEntries[1])
		}
	})

	t.Run("TCP target context cancled", func(t *testing.T) {
		t.Parallel()

		cfg := config.Config{
			TargetName:      "TCPServer",
			TargetAddress:   "localhost:7084",
			CheckInterval:   50 * time.Millisecond,
			DialTimeout:     50 * time.Millisecond,
			TargetCheckType: "tcp",
		}

		mockEnv := func(key string) string {
			return ""
		}

		checker, err := checker.NewTCPChecker(cfg.TargetName, cfg.TargetAddress, cfg.DialTimeout, mockEnv)
		if err != nil {
			t.Fatalf("Failed to create TCPChecker: %q", err)
		}

		var stdOut strings.Builder
		logger := slog.New(slog.NewTextHandler(&stdOut, nil))

		ctx, cancel := context.WithTimeout(context.Background(), cfg.CheckInterval*4)

		go func() {
			time.Sleep(100 * time.Millisecond)
			cancel()
		}()

		err = LoopUntilReady(ctx, cfg.CheckInterval, checker, logger)
		if err != nil && err != context.Canceled {
			t.Errorf("Expected context canceled error, got %q", err)
		}

		expected := fmt.Sprintf("Waiting for %s to become ready...", cfg.TargetName)
		if !strings.Contains(stdOut.String(), expected) {
			t.Errorf("Expected output to contain %q but got %q", expected, stdOut.String())
		}

		expected = fmt.Sprintf("%s is not ready ✗", cfg.TargetName)
		if !strings.Contains(stdOut.String(), expected) {
			t.Errorf("Expected output to contain %q but got %q", expected, stdOut.String())
		}
	})

	t.Run("TCP target context deadline exceeded", func(t *testing.T) {
		t.Parallel()

		cfg := config.Config{
			TargetName:      "TCPServer",
			TargetAddress:   "localhost:7084",
			CheckInterval:   50 * time.Millisecond,
			DialTimeout:     50 * time.Millisecond,
			TargetCheckType: "tcp",
		}

		mockEnv := func(key string) string {
			return ""
		}

		checker, err := checker.NewTCPChecker(cfg.TargetName, cfg.TargetAddress, cfg.DialTimeout, mockEnv)
		if err != nil {
			t.Fatalf("Failed to create TCPChecker: %q", err)
		}

		var stdOut strings.Builder
		logger := slog.New(slog.NewTextHandler(&stdOut, nil))

		ctx, cancel := context.WithDeadline(context.Background(), time.Now().Add(50*time.Millisecond))
		defer cancel() // Ensure cancel is called to free resources

		err = LoopUntilReady(ctx, cfg.CheckInterval, checker, logger)
		if err != context.DeadlineExceeded {
			t.Errorf("Expected context canceled error, got %q", err)
		}
	})
}

func TestLoopUntilReadyICMP(t *testing.T) {
	t.Parallel()

	t.Run("ICMP Target is ready", func(t *testing.T) {
		t.Parallel()

		// Configuration for the ICMP target
		cfg := config.Config{
			TargetName:    "ICMPServer",
			TargetAddress: "icmp://8.8.8.8", // Google's public DNS server
			CheckInterval: 50 * time.Millisecond,
			DialTimeout:   50 * time.Millisecond,
		}

		// Set up the environment function to return a read timeout
		mockEnv := func(key string) string {
			env := map[string]string{
				"ICMP_READ_TIMEOUT": "1s",
			}
			return env[key]
		}

		// Create a new ICMPChecker
		checker, err := checker.NewICMPChecker(cfg.TargetName, cfg.TargetAddress, cfg.DialTimeout, mockEnv)
		if err != nil {
			t.Fatalf("Failed to create ICMPChecker: %q", err)
		}

		// Set up a logger to capture output
		var stdOut strings.Builder
		logger := slog.New(slog.NewTextHandler(&stdOut, nil))

		// Create a context with a timeout for the test run
		ctx, cancel := context.WithTimeout(context.Background(), cfg.CheckInterval*4)
		defer cancel()

		// Run the check
		err = LoopUntilReady(ctx, cfg.CheckInterval, checker, logger)
		if err != nil {
			t.Errorf("Unexpected error: %q", err)
		}

		// Check the expected output
		expected := "ICMPServer is ready ✓"
		if !strings.Contains(stdOut.String(), expected) {
			t.Errorf("Expected output to contain %q but got %q", expected, stdOut.String())
		}
	})

	t.Run("ICMP Target Unreachable", func(t *testing.T) {
		t.Parallel()

		// Configuration for an unreachable ICMP target
		cfg := config.Config{
			TargetName:    "ICMPServerUnreachable",
			TargetAddress: "icmp://192.0.2.1", // This is a TEST-NET IP address that should not be reachable
			CheckInterval: 50 * time.Millisecond,
			DialTimeout:   50 * time.Millisecond,
		}

		// Set up the environment function to return a read timeout
		mockEnv := func(key string) string {
			env := map[string]string{
				"ICMP_READ_TIMEOUT": "1s",
			}
			return env[key]
		}

		// Create a new ICMPChecker
		checker, err := checker.NewICMPChecker(cfg.TargetName, cfg.TargetAddress, cfg.DialTimeout, mockEnv)
		if err != nil {
			t.Fatalf("Failed to create ICMPChecker: %q", err)
		}

		// Set up a logger to capture output
		var stdOut strings.Builder
		logger := slog.New(slog.NewTextHandler(&stdOut, nil))

		// Create a context with a timeout for the test run
		ctx, cancel := context.WithTimeout(context.Background(), cfg.CheckInterval*4)
		defer cancel()

		// Run the check
		err = LoopUntilReady(ctx, cfg.CheckInterval, checker, logger)
		if err == nil {
			t.Error("Expected an error due to unreachable target, but got none")
		}

		// Check the expected output
		expected := "ICMPServerUnreachable is not ready ✗"
		if !strings.Contains(stdOut.String(), expected) {
			t.Errorf("Expected output to contain %q but got %q", expected, stdOut.String())
		}
	})

	t.Run("ICMP Target Timeout", func(t *testing.T) {
		t.Parallel()

		// Configuration for a target that will time out
		cfg := config.Config{
			TargetName:    "ICMPServerTimeout",
			TargetAddress: "icmp://192.0.2.1", // Using the same TEST-NET IP
			CheckInterval: 50 * time.Millisecond,
			DialTimeout:   50 * time.Millisecond,
		}

		// Set up the environment function to return a very short read timeout
		mockEnv := func(key string) string {
			env := map[string]string{
				"ICMP_READ_TIMEOUT": "10ms", // Very short timeout to force a timeout error
			}
			return env[key]
		}

		// Create a new ICMPChecker
		checker, err := checker.NewICMPChecker(cfg.TargetName, cfg.TargetAddress, cfg.DialTimeout, mockEnv)
		if err != nil {
			t.Fatalf("Failed to create ICMPChecker: %q", err)
		}

		// Set up a logger to capture output
		var stdOut strings.Builder
		logger := slog.New(slog.NewTextHandler(&stdOut, nil))

		// Create a context with a timeout for the test run
		ctx, cancel := context.WithTimeout(context.Background(), cfg.CheckInterval*4)
		defer cancel()

		// Run the check
		err = LoopUntilReady(ctx, cfg.CheckInterval, checker, logger)
		if err == nil {
			t.Error("Expected a timeout error, but got none")
		}

		// Check the expected output
		expected := "ICMPServerTimeout is not ready ✗"
		if !strings.Contains(stdOut.String(), expected) {
			t.Errorf("Expected output to contain %q but got %q", expected, stdOut.String())
		}
	})

	t.Run("ICMP Target Context Canceled", func(t *testing.T) {
		t.Parallel()

		// Configuration for an ICMP target
		cfg := config.Config{
			TargetName:    "ICMPServerCancel",
			TargetAddress: "icmp://8.8.8.8", // Google's public DNS server
			CheckInterval: 50 * time.Millisecond,
			DialTimeout:   50 * time.Millisecond,
		}

		// Set up the environment function to return a read timeout
		mockEnv := func(key string) string {
			env := map[string]string{
				"ICMP_READ_TIMEOUT": "1s",
			}
			return env[key]
		}

		// Create a new ICMPChecker
		checker, err := checker.NewICMPChecker(cfg.TargetName, cfg.TargetAddress, cfg.DialTimeout, mockEnv)
		if err != nil {
			t.Fatalf("Failed to create ICMPChecker: %q", err)
		}

		// Set up a logger to capture output
		var stdOut strings.Builder
		logger := slog.New(slog.NewTextHandler(&stdOut, nil))

		// Create a context with a timeout for the test run
		ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond) // Cancel after 100ms
		defer cancel()

		// Simulate a delay to ensure the context cancels before the check can complete
		time.Sleep(200 * time.Millisecond)

		// Run the check
		err = LoopUntilReady(ctx, cfg.CheckInterval, checker, logger)
		if err == nil || !strings.Contains(err.Error(), "context deadline exceeded") {
			t.Errorf("Expected context deadline exceeded error, but got: %v", err)
		}

		// Check the expected output
		expected := "ICMPServerCancel is not ready ✗"
		if !strings.Contains(stdOut.String(), expected) {
			t.Errorf("Expected output to contain %q but got %q", expected, stdOut.String())
		}
	})
}
