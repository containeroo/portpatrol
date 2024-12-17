package httputils

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseHTTPHeaders(t *testing.T) {
	t.Parallel()

	t.Run("Valid headers", func(t *testing.T) {
		t.Parallel()

		headers := "Content-Type=application/json,Auportpatrolization=Bearer token"
		result, err := ParseHeaders(headers, true)

		assert.NoError(t, err)
		assert.ObjectsAreEqual(map[string]string{"Content-Type": "application/json", "Auportpatrolization": "Bearer token"}, result)
	})

	t.Run("Single header", func(t *testing.T) {
		t.Parallel()

		headers := "Content-Type=application/json"
		result, err := ParseHeaders(headers, true)

		assert.NoError(t, err)
		assert.ObjectsAreEqual(map[string]string{"Content-Type": "application/json"}, result)
	})

	t.Run("Empty headers string", func(t *testing.T) {
		t.Parallel()

		headers := ""
		result, err := ParseHeaders(headers, true)

		assert.NoError(t, err)
		assert.ObjectsAreEqual(map[string]string{}, result)
	})

	t.Run("Malformed header (missing =)", func(t *testing.T) {
		t.Parallel()

		headers := "Content-Type=application/json,AuportpatrolizationBearer token"
		_, err := ParseHeaders(headers, true)

		assert.Error(t, err)
		assert.EqualError(t, err, "invalid header format: AuportpatrolizationBearer token")
	})

	t.Run("Header with spaces", func(t *testing.T) {
		t.Parallel()

		headers := "  Content-Type = application/json  , Auportpatrolization = Bearer token  "
		result, err := ParseHeaders(headers, true)

		assert.NoError(t, err)
		assert.ObjectsAreEqual(map[string]string{"Content-Type": "application/json", "Auportpatrolization": "Bearer token"}, result)
	})

	t.Run("Header with empty key", func(t *testing.T) {
		t.Parallel()

		headers := "=value"
		_, err := ParseHeaders(headers, true)

		assert.Error(t, err)
		assert.EqualError(t, err, "header key cannot be empty: =value")
	})

	t.Run("Header with empty value", func(t *testing.T) {
		t.Parallel()

		headers := "key="
		result, err := ParseHeaders(headers, true)

		assert.NoError(t, err)
		assert.ObjectsAreEqual(map[string]string{"key": ""}, result)
	})

	t.Run("Trailing comma", func(t *testing.T) {
		t.Parallel()

		headers := "Content-Type=application/json,"
		result, err := ParseHeaders(headers, true)
		assert.NoError(t, err)

		assert.ObjectsAreEqual(map[string]string{"Content-Type": "application/json"}, result)
	})

	t.Run("Valid header with duplicate headers (allowDuplicates=true)", func(t *testing.T) {
		t.Parallel()

		headers := "Content-Type=application/json,Content-Type=application/json"
		h, err := ParseHeaders(headers, true)
		assert.NoError(t, err)
		assert.ObjectsAreEqual(map[string]string{"Content-Type": "application/json"}, h)
	})

	t.Run("Invalid header with duplicate headers (allowDuplicates=false)", func(t *testing.T) {
		t.Parallel()

		headers := "Content-Type=application/json,Content-Type=application/json"
		_, err := ParseHeaders(headers, false)

		assert.Error(t, err)
		assert.EqualError(t, err, "duplicate header key found: Content-Type")
	})
}

func TestParseHTTPStatusCodes(t *testing.T) {
	t.Parallel()

	t.Run("Valid status code", func(t *testing.T) {
		t.Parallel()

		statuses, err := ParseStatusCodes("200")

		assert.NoError(t, err)
		assert.ObjectsAreEqual([]int{200}, statuses)
	})

	t.Run("Valid multiple status codes", func(t *testing.T) {
		t.Parallel()

		statuses, err := ParseStatusCodes("200,404,500")

		assert.NoError(t, err)
		assert.ObjectsAreEqual([]int{200, 404, 500}, statuses)
	})

	t.Run("Valid status code range", func(t *testing.T) {
		t.Parallel()

		statuses, err := ParseStatusCodes("200-202")

		assert.NoError(t, err)
		assert.ObjectsAreEqual([]int{200, 201, 202}, statuses)
	})

	t.Run("Valid multiple status code ranges", func(t *testing.T) {
		t.Parallel()

		statuses, err := ParseStatusCodes("200-202,300-301,500")

		assert.NoError(t, err)
		assert.ObjectsAreEqual([]int{200, 201, 202, 300, 301, 500}, statuses)
	})

	t.Run("Invalid status code", func(t *testing.T) {
		t.Parallel()

		_, err := ParseStatusCodes("abc")

		assert.Error(t, err)
		assert.EqualError(t, err, "invalid status code: abc")
	})

	t.Run("Invalid status range double dash", func(t *testing.T) {
		t.Parallel()

		_, err := ParseStatusCodes("200--202")

		assert.Error(t, err)
		assert.EqualError(t, err, "invalid status range: 200--202")
	})

	t.Run("Invalid status range (start > end)", func(t *testing.T) {
		t.Parallel()

		_, err := ParseStatusCodes("201-200")

		assert.Error(t, err)
		assert.EqualError(t, err, "invalid status range: 201-200")
	})
}
