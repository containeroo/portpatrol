package checker

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewChecker(t *testing.T) {
	t.Parallel()

	t.Run("Valid HTTP checker", func(t *testing.T) {
		t.Parallel()

		check, err := NewChecker(HTTP, "example", "http://example.com")

		assert.NoError(t, err)
		assert.Equal(t, check.Name(), "example")
		assert.Equal(t, check.Type(), "HTTP")
	})

	t.Run("Valid TCP checker", func(t *testing.T) {
		t.Parallel()

		check, err := NewChecker(TCP, "example", "example.com:80")

		assert.NoError(t, err)
		assert.Equal(t, check.Name(), "example")
		assert.Equal(t, check.Type(), "TCP")
	})

	t.Run("Valid ICMP checker", func(t *testing.T) {
		t.Parallel()

		check, err := NewChecker(ICMP, "example", "example.com")

		assert.NoError(t, err)
		assert.Equal(t, check.Name(), "example")
		assert.Equal(t, check.Type(), "ICMP")
	})

	t.Run("Invalid checker type", func(t *testing.T) {
		t.Parallel()

		_, err := NewChecker("", "example", "example.com")

		assert.Error(t, err)
		assert.EqualError(t, err, "unsupported check type: ")
	})
}

func TestParseCheckType(t *testing.T) {
	t.Parallel()

	t.Run("Check type HTTP", func(t *testing.T) {
		t.Parallel()

		result, err := ParseCheckType("HTTP")

		assert.NoError(t, err)
		assert.Equal(t, result, HTTP)
	})

	t.Run("Check type http", func(t *testing.T) {
		t.Parallel()

		result, err := ParseCheckType("http")

		assert.NoError(t, err)
		assert.Equal(t, result, HTTP)
	})

	t.Run("Check type TCP", func(t *testing.T) {
		t.Parallel()

		result, err := ParseCheckType("tcp")

		assert.NoError(t, err)
		assert.Equal(t, result, TCP)
	})

	t.Run("Check type tcp", func(t *testing.T) {
		t.Parallel()

		result, err := ParseCheckType("tcp")

		assert.NoError(t, err)
		assert.Equal(t, result, TCP)
	})

	t.Run("Check type ICMP", func(t *testing.T) {
		t.Parallel()

		result, err := ParseCheckType("ICMP")

		assert.NoError(t, err)
		assert.Equal(t, result, ICMP)
	})

	t.Run("Check type icmp", func(t *testing.T) {
		t.Parallel()

		result, err := ParseCheckType("icmp")

		assert.NoError(t, err)
		assert.Equal(t, result, ICMP)
	})

	t.Run("Invalid check type", func(t *testing.T) {
		t.Parallel()

		_, err := ParseCheckType("invalid")

		assert.Error(t, err)
		assert.EqualError(t, err, "unsupported check type: invalid")
	})
}
