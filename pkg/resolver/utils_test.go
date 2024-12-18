package resolver

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestUtils(t *testing.T) {
	t.Parallel()

	t.Run("splitFileAndKey", func(t *testing.T) {
		t.Parallel()

		t.Run("NoDelimiter", func(t *testing.T) {
			t.Parallel()
			file, key := splitFileAndKey("path/to/file")
			assert.Equal(t, "path/to/file", file)
			assert.Equal(t, "", key)
		})

		t.Run("DelimiterStart", func(t *testing.T) {
			t.Parallel()
			file, key := splitFileAndKey("//key")
			assert.Equal(t, "", file)
			assert.Equal(t, "key", key)
		})

		t.Run("DelimiterEnd", func(t *testing.T) {
			t.Parallel()
			file, key := splitFileAndKey("path/to/file//")
			assert.Equal(t, "path/to/file", file)
			assert.Equal(t, "", key)
		})

		t.Run("DelimiterMiddle", func(t *testing.T) {
			t.Parallel()
			file, key := splitFileAndKey("path/to/file//somekey")
			assert.Equal(t, "path/to/file", file)
			assert.Equal(t, "somekey", key)
		})

		t.Run("MultipleDelimiters", func(t *testing.T) {
			t.Parallel()
			file, key := splitFileAndKey("path//to//file//key")
			assert.Equal(t, "path//to//file", file)
			assert.Equal(t, "key", key)
		})
	})

	t.Run("navigateData", func(t *testing.T) {
		t.Parallel()

		data := map[string]interface{}{
			"server": map[string]interface{}{
				"host":  "localhost",
				"ports": []interface{}{80, 443},
				"nested": map[string]interface{}{
					"key": "value",
				},
			},
			"array": []interface{}{
				"zero", "one", map[string]interface{}{"nested": "val"},
			},
		}

		t.Run("SimpleMapKey", func(t *testing.T) {
			t.Parallel()
			val, err := navigateData(data, []string{"server"})
			assert.NoError(t, err)
			assert.Equal(t, data["server"], val)
		})

		t.Run("NestedMapKey", func(t *testing.T) {
			t.Parallel()
			val, err := navigateData(data, []string{"server", "nested", "key"})
			assert.NoError(t, err)
			assert.Equal(t, "value", val)
		})

		t.Run("ArrayIndex", func(t *testing.T) {
			t.Parallel()
			val, err := navigateData(data, []string{"server", "ports", "1"})
			assert.NoError(t, err)
			assert.Equal(t, 443, val)
		})

		t.Run("ArrayOutOfBounds", func(t *testing.T) {
			t.Parallel()
			_, err := navigateData(data, []string{"server", "ports", "10"})
			assert.Error(t, err)
		})

		t.Run("InvalidArrayIndex", func(t *testing.T) {
			t.Parallel()
			_, err := navigateData(data, []string{"server", "ports", "notAnInt"})
			assert.Error(t, err)
		})

		t.Run("MissingKey", func(t *testing.T) {
			t.Parallel()
			_, err := navigateData(data, []string{"server", "missing"})
			assert.Error(t, err)
		})

		t.Run("MixedArrayAndMap", func(t *testing.T) {
			t.Parallel()
			val, err := navigateData(data, []string{"array", "2", "nested"})
			assert.NoError(t, err)
			assert.Equal(t, "val", val)
		})

		t.Run("PathSegmentOnNonMapNonArray", func(t *testing.T) {
			t.Parallel()
			_, err := navigateData(data, []string{"server", "host", "sub"})
			assert.Error(t, err)
		})
	})

	t.Run("convertToMapStringInterface_and_convertValue", func(t *testing.T) {
		t.Parallel()

		t.Run("TopLevelMap", func(t *testing.T) {
			t.Parallel()
			input := map[string]interface{}{
				"key": "value",
				"nested": map[string]interface{}{
					"subKey": []interface{}{"val1", "val2"},
				},
			}
			want := map[string]interface{}{
				"key": "value",
				"nested": map[string]interface{}{
					"subKey": []interface{}{"val1", "val2"},
				},
			}
			got, err := convertToMapStringInterface(input)
			assert.NoError(t, err)
			assert.Equal(t, want, got)
		})

		t.Run("NonMapTopLevel", func(t *testing.T) {
			t.Parallel()
			input := []interface{}{"val1", "val2"}
			want := map[string]interface{}{}
			got, err := convertToMapStringInterface(input)
			assert.NoError(t, err)
			assert.Equal(t, want, got)
		})

		t.Run("AlreadyClean", func(t *testing.T) {
			t.Parallel()
			input := map[string]interface{}{"simple": "val"}
			want := map[string]interface{}{"simple": "val"}
			got, err := convertToMapStringInterface(input)
			assert.NoError(t, err)
			assert.Equal(t, want, got)
		})

		t.Run("ComplexNestedStructures", func(t *testing.T) {
			t.Parallel()
			input := map[string]interface{}{
				"level1": map[string]interface{}{
					"level2": []interface{}{
						map[string]interface{}{"key": "val"},
						"stringVal",
					},
				},
			}
			want := map[string]interface{}{
				"level1": map[string]interface{}{
					"level2": []interface{}{
						map[string]interface{}{"key": "val"},
						"stringVal",
					},
				},
			}
			got, err := convertToMapStringInterface(input)
			assert.NoError(t, err)
			assert.Equal(t, want, got)
		})
	})
}
