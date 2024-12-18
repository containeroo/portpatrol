package resolver

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func createJsonTestFile(t *testing.T) string {
	tempDir := t.TempDir()

	// Create a temporary JSON file with nested objects and arrays
	testFilePath := filepath.Join(tempDir, "config.json")
	fileContent := `{
  "server": {
    "host": "localhost",
    "port": 8080,
    "nested": {
      "key": "value"
    }
  },
  "servers": [
    { "host": "example.com", "port": 80 },
    { "host": "example.org", "port": 443 }
  ],
  "emptyString": "",
  "nonString": { "inner": true }
}`
	err := os.WriteFile(testFilePath, []byte(fileContent), 0666)
	assert.NoError(t, err, "failed to create test JSON file")

	return testFilePath
}

func TestJSONResolver_Resolve(t *testing.T) {
	t.Parallel()

	t.Run("Resolve entire file", func(t *testing.T) {
		t.Parallel()

		resolver := &JSONResolver{}
		testFilePath := createJsonTestFile(t)

		// Call the resolver without the 'json:' prefix.
		val, err := resolver.Resolve(testFilePath)
		assert.NoError(t, err, "unexpected error resolving entire JSON file")

		expected := `{
  "server": {
    "host": "localhost",
    "port": 8080,
    "nested": {
      "key": "value"
    }
  },
  "servers": [
    { "host": "example.com", "port": 80 },
    { "host": "example.org", "port": 443 }
  ],
  "emptyString": "",
  "nonString": { "inner": true }
}`
		assert.Equal(t, expected, val)
	})

	t.Run("Resolve top-level key", func(t *testing.T) {
		t.Parallel()

		resolver := &JSONResolver{}
		testFilePath := createJsonTestFile(t)

		val, err := resolver.Resolve(testFilePath + "//server.host")
		assert.NoError(t, err, "unexpected error resolving top-level key")
		assert.Equal(t, "localhost", val)
	})

	t.Run("Resolve nested key", func(t *testing.T) {
		t.Parallel()

		resolver := &JSONResolver{}
		testFilePath := createJsonTestFile(t)

		val, err := resolver.Resolve(testFilePath + "//server.nested.key")
		assert.NoError(t, err, "unexpected error resolving nested key")
		assert.Equal(t, "value", val)
	})

	t.Run("Resolve array element", func(t *testing.T) {
		t.Parallel()

		resolver := &JSONResolver{}
		testFilePath := createJsonTestFile(t)

		val, err := resolver.Resolve(testFilePath + "//servers.1.host")
		assert.NoError(t, err, "unexpected error resolving array element")
		assert.Equal(t, "example.org", val)
	})

	t.Run("Resolve empty string key", func(t *testing.T) {
		t.Parallel()

		resolver := &JSONResolver{}
		testFilePath := createJsonTestFile(t)

		val, err := resolver.Resolve(testFilePath + "//emptyString")
		assert.NoError(t, err, "unexpected error resolving empty string key")
		assert.Equal(t, "", val)
	})

	t.Run("Resolve non-string value", func(t *testing.T) {
		t.Parallel()

		resolver := &JSONResolver{}
		testFilePath := createJsonTestFile(t)

		val, err := resolver.Resolve(testFilePath + "//nonString")
		assert.NoError(t, err, "unexpected error resolving non-string value")
		expected := `{"inner":true}`
		assert.Equal(t, expected, val)
	})

	t.Run("Resolve missing key", func(t *testing.T) {
		t.Parallel()

		resolver := &JSONResolver{}
		testFilePath := createJsonTestFile(t)

		_, err := resolver.Resolve(testFilePath + "//server.nested.missingKey")
		assert.Error(t, err, "expected an error resolving a missing key, but got none")
	})

	t.Run("Resolve non-existing file", func(t *testing.T) {
		t.Parallel()

		resolver := &JSONResolver{}

		tempDir := t.TempDir()
		nonExistentFile := filepath.Join(tempDir, "nonexistent.json")

		_, err := resolver.Resolve(nonExistentFile)
		assert.Error(t, err, "expected an error resolving a non-existing file, but got none")
	})
}
