package resolver

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func createIniTestFile(t *testing.T) string {
	tempDir := t.TempDir()

	testFilePath := filepath.Join(tempDir, "config.ini")
	fileContent := `
[DEFAULT]
Key1=DefaultVal1

[SectionA]
Key2=SectionAVal2
Key3=SectionAVal3

[SectionB]
Key4=SectionBVal4
`
	err := os.WriteFile(testFilePath, []byte(fileContent), 0666)
	assert.NoError(t, err, "failed to create test INI file")

	return testFilePath
}

func TestINIResolver_Resolve(t *testing.T) {
	t.Parallel()
	resolver := &INIResolver{}

	t.Run("Resolve entire file", func(t *testing.T) {
		t.Parallel()

		testFilePath := createIniTestFile(t)
		val, err := resolver.Resolve(testFilePath)
		assert.NoError(t, err, "unexpected error resolving entire INI file")

		expected := `[DEFAULT]
Key1=DefaultVal1

[SectionA]
Key2=SectionAVal2
Key3=SectionAVal3

[SectionB]
Key4=SectionBVal4`
		assert.Equal(t, expected, val)
	})

	t.Run("Resolve key from default section", func(t *testing.T) {
		t.Parallel()

		testFilePath := createIniTestFile(t)
		val, err := resolver.Resolve(testFilePath + "//Key1")
		assert.NoError(t, err, "unexpected error resolving default section key")
		assert.Equal(t, "DefaultVal1", val)
	})

	t.Run("Resolve key from named section", func(t *testing.T) {
		t.Parallel()

		testFilePath := createIniTestFile(t)
		val, err := resolver.Resolve(testFilePath + "//SectionA.Key3")
		assert.NoError(t, err, "unexpected error resolving section key")
		assert.Equal(t, "SectionAVal3", val)
	})

	t.Run("Resolve missing key", func(t *testing.T) {
		t.Parallel()

		testFilePath := createIniTestFile(t)
		_, err := resolver.Resolve(testFilePath + "//NonExistentKey")
		assert.Error(t, err, "expected an error resolving a missing key, but got none")
	})

	t.Run("Resolve missing section", func(t *testing.T) {
		t.Parallel()

		testFilePath := createIniTestFile(t)
		_, err := resolver.Resolve(testFilePath + "//NonExistentSection.Key")
		assert.Error(t, err, "expected an error resolving a missing section, but got none")
	})

	t.Run("Resolve non-existing file", func(t *testing.T) {
		t.Parallel()

		tempDir := t.TempDir()
		nonExistentFile := filepath.Join(tempDir, "nonexistent.ini")

		_, err := resolver.Resolve(nonExistentFile)
		assert.Error(t, err, "expected an error resolving a non-existing file, but got none")
	})
}
