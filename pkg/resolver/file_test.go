package resolver

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func createKeyValueTestFile(t *testing.T) string {
	tempDir := t.TempDir()

	testFilePath := filepath.Join(tempDir, "app.txt")
	fileContent := `Key1 = Value1
Key2=Value2
Key with spaces =  TrimmedValue
`
	err := os.WriteFile(testFilePath, []byte(fileContent), 0666)
	assert.NoError(t, err, "failed to create test file")

	return testFilePath
}

func TestKeyValueFileResolver_Resolve(t *testing.T) {
	t.Parallel()

	resolver := &KeyValueFileResolver{}

	t.Run("Resolve entire file", func(t *testing.T) {
		t.Parallel()
		testFilePath := createKeyValueTestFile(t)

		val, err := resolver.Resolve(testFilePath)
		assert.NoError(t, err, "unexpected error resolving entire file")

		expected := "Key1 = Value1\nKey2=Value2\nKey with spaces =  TrimmedValue"
		assert.Equal(t, expected, val)
	})

	t.Run("Resolve specific key", func(t *testing.T) {
		t.Parallel()
		testFilePath := createKeyValueTestFile(t)

		val, err := resolver.Resolve(testFilePath + "//Key2")
		assert.NoError(t, err, "unexpected error resolving specific key")
		assert.Equal(t, "Value2", val)
	})

	t.Run("Resolve key with spaces", func(t *testing.T) {
		t.Parallel()
		testFilePath := createKeyValueTestFile(t)

		val, err := resolver.Resolve(testFilePath + "//Key with spaces")
		assert.NoError(t, err, "unexpected error resolving key with spaces")
		assert.Equal(t, "TrimmedValue", val)
	})

	t.Run("Resolve missing key", func(t *testing.T) {
		t.Parallel()
		testFilePath := createKeyValueTestFile(t)

		_, err := resolver.Resolve(testFilePath + "//NonExistentKey")
		assert.Error(t, err, "expected an error resolving a missing key, but got none")
	})

	t.Run("Resolve non-existing file", func(t *testing.T) {
		t.Parallel()
		tempDir := t.TempDir()
		nonExistentFile := filepath.Join(tempDir, "nonexistent.txt")

		_, err := resolver.Resolve(nonExistentFile)
		assert.Error(t, err, "expected an error resolving a non-existing file, but got none")
	})
}
