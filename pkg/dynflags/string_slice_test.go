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

	t.Run("Multiple Occurrences Append Correctly", func(t *testing.T) {
		t.Parallel()

		var bound []string
		value := &dynflags.StringSlicesValue{Bound: &bound}

		assert.NoError(t, value.Set("Content-Type=application/json"))
		assert.NoError(t, value.Set("MyHeader=header1"))
		assert.NoError(t, value.Set("Header1=value1,Header2=value2"))

		assert.Equal(t, []string{
			"Content-Type=application/json",
			"MyHeader=header1",
			"Header1=value1,Header2=value2",
		}, bound)
	})

	t.Run("Single Value Append", func(t *testing.T) {
		t.Parallel()

		var bound []string
		value := &dynflags.StringSlicesValue{Bound: &bound}

		assert.NoError(t, value.Set("Content-Type=application/json"))
		assert.Equal(t, []string{"Content-Type=application/json"}, bound)
	})

	t.Run("Invalid Value Type", func(t *testing.T) {
		t.Parallel()

		var bound []string
		value := &dynflags.StringSlicesValue{Bound: &bound}

		err := value.Set(123) // Invalid type
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid value type")
	})
}

func TestGroupConfigStringSlices(t *testing.T) {
	t.Parallel()

	t.Run("Define string slices flag", func(t *testing.T) {
		t.Parallel()

		group := &dynflags.ConfigGroup{Flags: make(map[string]*dynflags.Flag)}
		defaultValue := []string{"default1", "default2"}
		group.StringSlices("stringSliceFlag", defaultValue, "A string slices flag")

		assert.Contains(t, group.Flags, "stringSliceFlag")
		assert.Equal(t, "A string slices flag", group.Flags["stringSliceFlag"].Usage)
		assert.Equal(t, "default1,default2", group.Flags["stringSliceFlag"].Default)
	})
}

func TestGetStringSlices(t *testing.T) {
	t.Parallel()

	t.Run("Retrieve []string value", func(t *testing.T) {
		t.Parallel()

		parsedGroup := &dynflags.ParsedGroup{
			Name:   "testGroup",
			Values: map[string]interface{}{"flag1": []string{"value1", "value2"}},
		}

		result, err := parsedGroup.GetStringSlices("flag1")
		assert.NoError(t, err)
		assert.Equal(t, []string{"value1", "value2"}, result)
	})

	t.Run("Retrieve single string value as []string", func(t *testing.T) {
		t.Parallel()

		parsedGroup := &dynflags.ParsedGroup{
			Name:   "testGroup",
			Values: map[string]interface{}{"flag1": "singleValue"},
		}

		result, err := parsedGroup.GetStringSlices("flag1")
		assert.NoError(t, err)
		assert.Equal(t, []string{"singleValue"}, result)
	})

	t.Run("Flag not found", func(t *testing.T) {
		t.Parallel()

		parsedGroup := &dynflags.ParsedGroup{
			Name:   "testGroup",
			Values: map[string]interface{}{},
		}

		result, err := parsedGroup.GetStringSlices("nonExistentFlag")
		assert.Error(t, err)
		assert.Nil(t, result)
		assert.EqualError(t, err, "flag 'nonExistentFlag' not found in group 'testGroup'")
	})

	t.Run("Flag value is invalid type", func(t *testing.T) {
		t.Parallel()

		parsedGroup := &dynflags.ParsedGroup{
			Name:   "testGroup",
			Values: map[string]interface{}{"flag1": 123}, // Invalid type (int)
		}

		result, err := parsedGroup.GetStringSlices("flag1")
		assert.Error(t, err)
		assert.Nil(t, result)
		assert.EqualError(t, err, "flag 'flag1' is not a []string")
	})
}

func TestGetStringSlicesGetBound(t *testing.T) {
	t.Run("StringSlicesValue - GetBound", func(t *testing.T) {
		var slices *[]string
		val := []string{"a", "b", "c"}
		slices = &val

		stringSlicesValue := dynflags.StringSlicesValue{Bound: slices}
		assert.Equal(t, val, stringSlicesValue.GetBound())

		stringSlicesValue = dynflags.StringSlicesValue{Bound: nil}
		assert.Nil(t, stringSlicesValue.GetBound())
	})
}
