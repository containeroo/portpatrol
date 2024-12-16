package factory_test

import (
	"testing"
	"time"

	"github.com/containeroo/portpatrol/internal/factory"
	"github.com/containeroo/portpatrol/pkg/dynflags"
	"github.com/stretchr/testify/assert"
)

func TestBuildCheckers(t *testing.T) {
	t.Parallel()

	t.Run("Valid HTTP Checker", func(t *testing.T) {
		t.Parallel()

		df := dynflags.New(dynflags.ContinueOnError)
		httpGroup := df.Group("http")
		httpGroup.String("address", "http://example.com", "HTTP target address")
		httpGroup.String("method", "GET", "HTTP method")
		httpGroup.Duration("interval", 5*time.Second, "Request interval")
		httpGroup.String("headers", "Content-Type: application/json", "HTTP headers")
		httpGroup.Bool("skip-tls-verify", false, "Skip TLS verification")
		httpGroup.Duration("timeout", 2*time.Second, "Timeout")

		err := df.Parse([]string{
			"--http.mygroup.address=http://example.com",
			"--http.mygroup.method=GET",
			"--http.mygroup.interval=5s",
			"--http.mygroup.headers=\"Content-Type=application/json\"",
			"--http.mygroup.skip-tls-verify=true",
			"--http.mygroup.timeout=2s",
		})

		assert.NoError(t, err)

		checkers, err := factory.BuildCheckers(df, 2*time.Second)
		assert.NoError(t, err)
		assert.Len(t, checkers, 1)
		assert.Equal(t, "http://example.com", checkers[0].Checker.GetAddress())
		assert.Equal(t, 5*time.Second, checkers[0].Interval)
	})

	t.Run("Missing Address", func(t *testing.T) {
		t.Parallel()

		df := dynflags.New(dynflags.ContinueOnError)
		httpGroup := df.Group("http")
		httpGroup.String("method", "GET", "HTTP method")

		df.Parse([]string{"--http.mygroup.method=GET"})

		checkers, err := factory.BuildCheckers(df, 2*time.Second)
		assert.Nil(t, checkers)
		assert.ErrorContains(t, err, "missing address for http checker")
	})

	t.Run("Invalid Check Type", func(t *testing.T) {
		t.Parallel()

		df := dynflags.New(dynflags.ContinueOnError)
		invalidGroup := df.Group("invalid")
		invalidGroup.String("address", "invalid-address", "Invalid target address")

		df.Parse([]string{"--invalid.mygroup.address=invalid-address"})

		checkers, err := factory.BuildCheckers(df, 2*time.Second)
		assert.Nil(t, checkers)
		assert.ErrorContains(t, err, "invalid check type 'invalid'")
	})

	t.Run("Invalid Header Parsing", func(t *testing.T) {
		t.Parallel()

		df := dynflags.New(dynflags.ContinueOnError)
		httpGroup := df.Group("http")
		httpGroup.String("address", "http://example.com", "HTTP target address")
		httpGroup.String("headers", "InvalidHeaderFormat", "HTTP headers")

		df.Parse([]string{"--http.mygroup.address=http://example.com", "--http.mygroup.headers=InvalidHeaderFormat"})

		checkers, err := factory.BuildCheckers(df, 2*time.Second)
		assert.Nil(t, checkers)
		assert.ErrorContains(t, err, "invalid \"--http.mygroup.headers\"")
	})

	t.Run("Valid TCP Checker", func(t *testing.T) {
		t.Parallel()

		df := dynflags.New(dynflags.ContinueOnError)
		tcpGroup := df.Group("tcp")
		tcpGroup.String("address", "127.0.0.1:8080", "TCP target address")
		tcpGroup.Duration("timeout", 3*time.Second, "Timeout")

		df.Parse([]string{"--tcp.mygroup.address=127.0.0.1:8080", "--tcp.mygroup.timeout=3s"})

		checkers, err := factory.BuildCheckers(df, 2*time.Second)
		assert.NoError(t, err)
		assert.Len(t, checkers, 1)
		assert.Equal(t, "127.0.0.1:8080", checkers[0].Checker.GetAddress())
	})

	t.Run("Valid ICMP Checker", func(t *testing.T) {
		t.Parallel()

		df := dynflags.New(dynflags.ContinueOnError)
		icmpGroup := df.Group("icmp")
		icmpGroup.String("address", "8.8.8.8", "ICMP target address")
		icmpGroup.Duration("read-timeout", 2*time.Second, "Read timeout")
		icmpGroup.Duration("write-timeout", 2*time.Second, "Write timeout")

		args := []string{
			"--icmp.mygroup.address=8.8.8.8",
			"--icmp.mygroup.read-timeout=2s",
			"--icmp.mygroup.write-timeout=2s",
		}

		err := df.Parse(args)
		assert.NoError(t, err)

		checkers, err := factory.BuildCheckers(df, 2*time.Second)
		assert.NoError(t, err)
		assert.Len(t, checkers, 1)
		assert.Equal(t, "8.8.8.8", checkers[0].Checker.GetAddress())
	})

	t.Run("Invalid ICMP Checker", func(t *testing.T) {
		t.Parallel()

		df := dynflags.New(dynflags.ContinueOnError)
		icmpGroup := df.Group("icmp")
		icmpGroup.String("address", "8.8.8.8", "ICMP target address")

		args := []string{
			"--icmp.mygroup.address=://invalid-url",
		}

		err := df.Parse(args)
		assert.NoError(t, err)

		checker, err := factory.BuildCheckers(df, 2*time.Second)
		assert.Nil(t, checker)
		assert.Error(t, err)
	})

	t.Run("Checker Creation Failure", func(t *testing.T) {
		t.Parallel()

		df := dynflags.New(dynflags.ContinueOnError)
		httpGroup := df.Group("http")
		httpGroup.String("address", "", "HTTP target address")

		df.Parse([]string{"--http.mygroup.address="})

		checkers, err := factory.BuildCheckers(df, 2*time.Second)
		assert.NotNil(t, checkers)
		assert.NoError(t, err)
	})
}
