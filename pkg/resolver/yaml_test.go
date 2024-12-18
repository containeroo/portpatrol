package resolver

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func createYamlTestFile(t *testing.T) string {
	tempDir := t.TempDir()

	testFilePath := filepath.Join(tempDir, "config.yaml")
	fileContent := `server:
  host: localhost
  port: 8080
  nested:
    key: value
servers:
  - host: example.com
    port: 80
  - host: example.org
    port: 443
emptyString: ""
nonString:
  inner: true
`
	err := os.WriteFile(testFilePath, []byte(fileContent), 0666)
	assert.NoError(t, err, "failed to create test YAML file")

	return testFilePath
}

func TestYAMLResolver_Resolve(t *testing.T) {
	t.Parallel()
	resolver := &YAMLResolver{}

	t.Run("Resolve entire file", func(t *testing.T) {
		t.Parallel()

		testFilePath := createYamlTestFile(t)
		val, err := resolver.Resolve(testFilePath)
		assert.NoError(t, err, "unexpected error resolving entire YAML file")

		expected := `server:
  host: localhost
  port: 8080
  nested:
    key: value
servers:
  - host: example.com
    port: 80
  - host: example.org
    port: 443
emptyString: ""
nonString:
  inner: true`
		assert.Equal(t, expected, val)
	})

	t.Run("Resolve top-level key", func(t *testing.T) {
		t.Parallel()

		testFilePath := createYamlTestFile(t)
		val, err := resolver.Resolve(testFilePath + "//server.host")
		assert.NoError(t, err, "unexpected error resolving top-level key")
		assert.Equal(t, "localhost", val)
	})

	t.Run("Resolve nested key", func(t *testing.T) {
		t.Parallel()

		testFilePath := createYamlTestFile(t)
		val, err := resolver.Resolve(testFilePath + "//server.nested.key")
		assert.NoError(t, err, "unexpected error resolving nested key")
		assert.Equal(t, "value", val)
	})

	t.Run("Resolve array element", func(t *testing.T) {
		t.Parallel()

		testFilePath := createYamlTestFile(t)
		val, err := resolver.Resolve(testFilePath + "//servers.1.host")
		assert.NoError(t, err, "unexpected error resolving array element")
		assert.Equal(t, "example.org", val)
	})

	t.Run("Resolve empty string key", func(t *testing.T) {
		t.Parallel()

		testFilePath := createYamlTestFile(t)
		val, err := resolver.Resolve(testFilePath + "//emptyString")
		assert.NoError(t, err, "unexpected error resolving empty string key")
		assert.Equal(t, "", val)
	})

	t.Run("Resolve non-string value", func(t *testing.T) {
		t.Parallel()

		testFilePath := createYamlTestFile(t)
		val, err := resolver.Resolve(testFilePath + "//nonString")
		assert.NoError(t, err, "unexpected error resolving non-string value")

		expected := `inner: true`
		assert.Equal(t, expected, val)
	})

	t.Run("Resolve missing key", func(t *testing.T) {
		t.Parallel()

		testFilePath := createYamlTestFile(t)
		_, err := resolver.Resolve(testFilePath + "//server.nested.missingKey")
		assert.Error(t, err, "expected an error resolving a missing key, but got none")
	})

	t.Run("Resolve non-existing file", func(t *testing.T) {
		t.Parallel()

		tempDir := t.TempDir()
		nonExistentFile := filepath.Join(tempDir, "nonexistent.yaml")

		_, err := resolver.Resolve(nonExistentFile)
		assert.Error(t, err, "expected an error resolving a non-existing file, but got none")
	})

	t.Run("Invalid YAML", func(t *testing.T) {
		t.Parallel()

		tempDir := t.TempDir()
		testFilePath := filepath.Join(tempDir, "config.yaml")
		fileContent := "key: \"unclosed string"
		err := os.WriteFile(testFilePath, []byte(fileContent), 0666)
		assert.NoError(t, err)

		_, err = resolver.Resolve(testFilePath)
		assert.Error(t, err)
		expected := fmt.Sprintf("failed to parse YAML in '%s': yaml: found unexpected end of stream", testFilePath)
		assert.EqualError(t, err, expected)
	})
}
