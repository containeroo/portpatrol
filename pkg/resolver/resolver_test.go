package resolver

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestResolveVariable(t *testing.T) {
	t.Parallel()

	t.Run("NoPrefixReturnsValueAsIs", func(t *testing.T) {
		t.Parallel()
		val, err := ResolveVariable("somevalue")
		assert.NoError(t, err, "unexpected error")
		assert.Equal(t, "somevalue", val)
	})

	t.Run("EnvPrefix", func(t *testing.T) {
		t.Parallel()

		envName := "RESOLVER_TEST_ENV_" + sanitizeEnvName(t.Name())
		os.Setenv(envName, "envValue")

		val, err := ResolveVariable("env:" + envName)
		assert.NoError(t, err, "unexpected error resolving env variable")
		assert.Equal(t, "envValue", val)
	})

	t.Run("JsonPrefix", func(t *testing.T) {
		t.Parallel()

		tempDir := t.TempDir()
		jsonPath := filepath.Join(tempDir, "config.json")
		jsonContent := `{"key":"jsonValue"}`
		err := os.WriteFile(jsonPath, []byte(jsonContent), 0666)
		assert.NoError(t, err, "failed to write test JSON file")

		envName := "RESOLVER_TEST_JSON_" + sanitizeEnvName(t.Name())
		os.Setenv(envName, jsonPath)

		val, err := ResolveVariable("json:$" + envName + "//key")
		assert.NoError(t, err, "unexpected error resolving json key")
		assert.Equal(t, "jsonValue", val)
	})

	t.Run("YamlPrefix", func(t *testing.T) {
		t.Parallel()

		tempDir := t.TempDir()
		yamlPath := filepath.Join(tempDir, "config.yaml")
		yamlContent := "key: yamlValue"
		err := os.WriteFile(yamlPath, []byte(yamlContent), 0666)
		assert.NoError(t, err, "failed to write test YAML file")

		envName := "RESOLVER_TEST_YAML_" + sanitizeEnvName(t.Name())
		os.Setenv(envName, yamlPath)

		val, err := ResolveVariable("yaml:$" + envName + "//key")
		assert.NoError(t, err, "unexpected error resolving yaml key")
		assert.Equal(t, "yamlValue", val)
	})

	t.Run("IniPrefix", func(t *testing.T) {
		t.Parallel()

		tempDir := t.TempDir()
		iniPath := filepath.Join(tempDir, "config.ini")
		iniContent := `[DEFAULT]
Key=iniValue`
		err := os.WriteFile(iniPath, []byte(iniContent), 0666)
		assert.NoError(t, err, "failed to write test INI file")

		envName := "RESOLVER_TEST_INI_" + sanitizeEnvName(t.Name())
		os.Setenv(envName, iniPath)

		val, err := ResolveVariable("ini:$" + envName + "//Key")
		assert.NoError(t, err, "unexpected error resolving ini key")
		assert.Equal(t, "iniValue", val)
	})

	t.Run("FilePrefix", func(t *testing.T) {
		t.Parallel()

		// According to the code, filePrefix maps to INIResolver. We'll test similarly to INI.
		tempDir := t.TempDir()
		filePath := filepath.Join(tempDir, "config.txt")
		fileContent := `[DEFAULT]
FileKey=fileValue`
		err := os.WriteFile(filePath, []byte(fileContent), 0666)
		assert.NoError(t, err, "failed to write test file")

		envName := "RESOLVER_TEST_FILE_" + sanitizeEnvName(t.Name())
		os.Setenv(envName, filePath)

		val, err := ResolveVariable("file:$" + envName + "//FileKey")
		assert.NoError(t, err, "unexpected error resolving file key")
		assert.Equal(t, "fileValue", val)
	})

	t.Run("UnknownPrefix", func(t *testing.T) {
		t.Parallel()

		val, err := ResolveVariable("unknown:somevalue")
		assert.NoError(t, err, "unexpected error for unknown prefix")
		assert.Equal(t, "unknown:somevalue", val, "For unknown prefix, returns the value as-is")
	})
}
