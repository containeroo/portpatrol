package wait

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
	"github.com/containeroo/portpatrol/internal/testutils"

	"golang.org/x/net/icmp"
	"golang.org/x/net/ipv4"
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

		err = WaitUntilReady(ctx, cfg.CheckInterval, checker, logger)
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

		err = WaitUntilReady(ctx, cfg.CheckInterval, checker, logger)
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
			TargetCheckType: checker.HTTP,
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

		err = WaitUntilReady(ctx, cfg.CheckInterval, checker, logger)
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
			TargetCheckType: checker.HTTP,
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

		err = WaitUntilReady(ctx, cfg.CheckInterval, checker, logger)
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
			TargetCheckType: checker.HTTP,
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

		err = WaitUntilReady(ctx, cfg.CheckInterval, checker, logger)
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
			TargetCheckType: checker.TCP,
		}

		checker, err := checker.NewTCPChecker(cfg.TargetName, cfg.TargetAddress, cfg.DialTimeout)
		if err != nil {
			t.Fatalf("Failed to create TCPChecker: %q", err)
		}

		var stdOut strings.Builder
		logger := slog.New(slog.NewTextHandler(&stdOut, nil))

		ctx, cancel := context.WithTimeout(context.Background(), cfg.CheckInterval*4)
		defer cancel()

		err = WaitUntilReady(ctx, cfg.CheckInterval, checker, logger)
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

		checker, err := checker.NewTCPChecker(cfg.TargetName, cfg.TargetAddress, cfg.DialTimeout)
		if err != nil {
			t.Fatalf("Failed to create TCPChecker: %q", err)
		}

		var stdOut strings.Builder
		logger := slog.New(slog.NewTextHandler(&stdOut, nil))

		ctx, cancel := context.WithTimeout(context.Background(), cfg.CheckInterval*4)
		defer cancel()

		err = WaitUntilReady(ctx, cfg.CheckInterval, checker, logger)
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
			TargetCheckType: checker.TCP,
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

		checker, err := checker.NewTCPChecker(cfg.TargetName, cfg.TargetAddress, cfg.DialTimeout)
		if err != nil {
			t.Fatalf("Failed to create HTTPChecker: %q", err)
		}

		var stdOut strings.Builder
		logger := logger.SetupLogger(cfg, &stdOut)

		err = WaitUntilReady(ctx, cfg.CheckInterval, checker, logger)
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
			TargetName:    "TCPServer",
			TargetAddress: "localhost:7084",
			CheckInterval: 50 * time.Millisecond,
			DialTimeout:   50 * time.Millisecond,
		}

		checker, err := checker.NewTCPChecker(cfg.TargetName, cfg.TargetAddress, cfg.DialTimeout)
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

		err = WaitUntilReady(ctx, cfg.CheckInterval, checker, logger)
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
			TargetCheckType: checker.TCP,
		}

		checker, err := checker.NewTCPChecker(cfg.TargetName, cfg.TargetAddress, cfg.DialTimeout)
		if err != nil {
			t.Fatalf("Failed to create TCPChecker: %q", err)
		}

		var stdOut strings.Builder
		logger := slog.New(slog.NewTextHandler(&stdOut, nil))

		ctx, cancel := context.WithDeadline(context.Background(), time.Now().Add(50*time.Millisecond))
		defer cancel() // Ensure cancel is called to free resources

		err = WaitUntilReady(ctx, cfg.CheckInterval, checker, logger)
		if err != context.DeadlineExceeded {
			t.Errorf("Expected context canceled error, got %q", err)
		}
	})
}

func TestICMPChecker_Check_SuccessfulICMPCheck(t *testing.T) {
	t.Run("Successful ICMP Check", func(t *testing.T) {
		var generatedIdentifier uint16
		var generatedSequence uint16

		mockPacketConn := &testutils.MockPacketConn{
			WriteToFunc: func(b []byte, addr net.Addr) (int, error) {
				// Capture the identifier and sequence number generated by the ICMPChecker
				msg, err := icmp.ParseMessage(1, b)
				if err != nil {
					return 0, err
				}
				echo, ok := msg.Body.(*icmp.Echo)
				if !ok {
					return 0, fmt.Errorf("invalid ICMP message body")
				}
				generatedIdentifier = uint16(echo.ID)
				generatedSequence = uint16(echo.Seq)
				return len(b), nil
			},
			ReadFromFunc: func(b []byte) (int, net.Addr, error) {
				// Create a response with the captured identifier and sequence number
				msg := icmp.Message{
					Type: ipv4.ICMPTypeEchoReply,
					Code: 0,
					Body: &icmp.Echo{
						ID:   int(generatedIdentifier),
						Seq:  int(generatedSequence),
						Data: []byte("HELLO-R-U-THERE"),
					},
				}
				msgBytes, _ := msg.Marshal(nil)
				copy(b, msgBytes)
				return len(msgBytes), &net.IPAddr{IP: net.IPv4(127, 0, 0, 1)}, nil
			},
			SetReadDeadlineFunc: func(t time.Time) error {
				return nil
			},
			CloseFunc: func() error {
				return nil
			},
		}

		mockProtocol := &testutils.MockProtocol{
			MakeRequestFunc: func(identifier, sequence uint16) ([]byte, error) {
				body := &icmp.Echo{
					ID:   int(identifier),
					Seq:  int(sequence),
					Data: []byte("HELLO-R-U-THERE"),
				}
				msg := icmp.Message{
					Type: ipv4.ICMPTypeEcho,
					Code: 0,
					Body: body,
				}
				return msg.Marshal(nil)
			},
			ValidateReplyFunc: func(reply []byte, identifier, sequence uint16) error {
				parsedMsg, err := icmp.ParseMessage(1, reply)
				if err != nil {
					return err
				}
				body, ok := parsedMsg.Body.(*icmp.Echo)
				if !ok || body.ID != int(identifier) || body.Seq != int(sequence) {
					return fmt.Errorf("identifier or sequence mismatch")
				}
				return nil
			},
			NetworkFunc: func() string {
				return "ip4:icmp"
			},
			ListenPacketFunc: func(ctx context.Context, network, address string) (net.PacketConn, error) {
				return mockPacketConn, nil
			},
		}

		checker := &checker.ICMPChecker{
			Name:        "TestChecker",
			Address:     "127.0.0.1",
			Protocol:    mockProtocol,
			ReadTimeout: 2 * time.Second,
		}

		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()

		err := checker.Check(ctx)
		if err != nil {
			t.Errorf("expected no error, got %v", err)
		}
	})
}
