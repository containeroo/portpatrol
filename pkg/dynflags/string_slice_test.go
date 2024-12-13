package dynflags_test

import (
	"testing"

	"github.com/containeroo/portpatrol/pkg/dynflags"
	"github.com/stretchr/testify/assert"
)

func TestStringSlicesValue(t *testing.T) {
	t.Parallel()

	t.Run("Parse valid string slice value", func(t *testing.T) {
		t.Parallel()

		stringSlicesValue := dynflags.StringSlicesValue{Bound: &[]string{}}
		parsed, err := stringSlicesValue.Parse("example")
		assert.NoError(t, err)
		assert.Equal(t, "example", parsed)
	})

	t.Run("Set valid string slice value", func(t *testing.T) {
		t.Parallel()

		bound := []string{"initial"}
		stringSlicesValue := dynflags.StringSlicesValue{Bound: &bound}

		err := stringSlicesValue.Set("updated")
		assert.NoError(t, err)
		assert.Equal(t, []string{"initial", "updated"}, bound)
	})

	t.Run("Set invalid type", func(t *testing.T) {
		t.Parallel()

		bound := []string{"initial"}
		stringSlicesValue := dynflags.StringSlicesValue{Bound: &bound}

		err := stringSlicesValue.Set(123) // Invalid type
		assert.Error(t, err)
		assert.EqualError(t, err, "invalid value type: expected string")
	})
}

func TestGroupConfigStringSlices(t *testing.T) {
	t.Parallel()

	t.Run("Define string slices flag", func(t *testing.T) {
		t.Parallel()

		group := &dynflags.GroupConfig{Flags: make(map[string]*dynflags.Flag)}
		defaultValue := []string{"default1", "default2"}
		strSlice := group.StringSlices("stringSliceFlag", defaultValue, "A string slices flag")

		assert.Equal(t, []string{"default1", "default2"}, *strSlice)
		assert.Contains(t, group.Flags, "stringSliceFlag")
		assert.Equal(t, "A string slices flag", group.Flags["stringSliceFlag"].Usage)
		assert.Equal(t, "default1,default2", group.Flags["stringSliceFlag"].Default)
	})

	t.Run("Define StringSlicesVar and set value", func(t *testing.T) {
		t.Parallel()

		group := &dynflags.GroupConfig{Flags: make(map[string]*dynflags.Flag)}
		var strSlice []string
		group.StringSlicesVar(&strSlice, "stringSliceFlag", []string{"initial1", "initial2"}, "String slices flag variable")
		assert.Equal(t, []string{"initial1", "initial2"}, strSlice)
	})
}

func TestParsedGroupGetStringSlices(t *testing.T) {
	t.Parallel()

	t.Run("Get existing string slices flag", func(t *testing.T) {
		t.Parallel()

		group := &dynflags.ParsedGroup{
			Values: map[string]interface{}{
				"stringSliceFlag": []string{"value1", "value2"},
			},
		}
		slice, err := group.GetStringSlices("stringSliceFlag")
		assert.NoError(t, err)
		assert.Equal(t, []string{"value1", "value2"}, slice)
	})

	t.Run("Get non-existent string slices flag", func(t *testing.T) {
		t.Parallel()

		group := &dynflags.ParsedGroup{
			Values: map[string]interface{}{},
		}
		slice, err := group.GetStringSlices("stringSliceFlag")
		assert.Error(t, err)
		assert.Nil(t, slice)
		assert.EqualError(t, err, "flag 'stringSliceFlag' not found in group ''")
	})

	t.Run("Get string slices flag with invalid type", func(t *testing.T) {
		t.Parallel()

		group := &dynflags.ParsedGroup{
			Values: map[string]interface{}{
				"stringSliceFlag": 123, // Invalid type
			},
		}
		slice, err := group.GetStringSlices("stringSliceFlag")
		assert.Error(t, err)
		assert.Nil(t, slice)
		assert.EqualError(t, err, "flag 'stringSliceFlag' is not a []string")
	})
}
