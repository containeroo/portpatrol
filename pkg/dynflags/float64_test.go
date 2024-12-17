package dynflags_test

import (
	"testing"

	"github.com/containeroo/portpatrol/pkg/dynflags"
	"github.com/stretchr/testify/assert"
)

func TestFloat64Value_Parse(t *testing.T) {
	t.Parallel()

	t.Run("Valid Float64", func(t *testing.T) {
		t.Parallel()

		bound := new(float64)
		fv := &dynflags.Float64Value{Bound: bound}

		parsedValue, err := fv.Parse("123.456")
		assert.NoError(t, err)
		assert.Equal(t, 123.456, parsedValue)
	})

	t.Run("Invalid Float64", func(t *testing.T) {
		t.Parallel()

		bound := new(float64)
		fv := &dynflags.Float64Value{Bound: bound}

		_, err := fv.Parse("invalid")
		assert.Error(t, err)
	})
}

func TestFloat64Value_Set(t *testing.T) {
	t.Parallel()

	t.Run("Set Valid Float64", func(t *testing.T) {
		t.Parallel()

		bound := new(float64)
		fv := &dynflags.Float64Value{Bound: bound}

		err := fv.Set(123.456)
		assert.NoError(t, err)
		assert.Equal(t, 123.456, *bound)
	})

	t.Run("Set Invalid Float64", func(t *testing.T) {
		t.Parallel()

		bound := new(float64)
		fv := &dynflags.Float64Value{Bound: bound}

		err := fv.Set("invalid")
		assert.Error(t, err)
	})
}

func TestGroupConfig_Float64(t *testing.T) {
	t.Parallel()

	t.Run("Define Float64 Flag", func(t *testing.T) {
		t.Parallel()

		group := &dynflags.ConfigGroup{
			Flags: make(map[string]*dynflags.Flag),
		}
		value := group.Float64("float64-test", 123.456, "test float64 flag")

		assert.NotNil(t, value)
		assert.Equal(t, 123.456, value.Default)
		assert.Contains(t, group.Flags, "float64-test")
	})
}

func TestParsedGroup_GetFloat64(t *testing.T) {
	t.Parallel()

	t.Run("Get Existing Float64 Value", func(t *testing.T) {
		t.Parallel()

		parsedGroup := &dynflags.ParsedGroup{
			Name:   "test-group",
			Values: map[string]interface{}{"float64-test": 123.456},
		}

		value, err := parsedGroup.GetFloat64("float64-test")
		assert.NoError(t, err)
		assert.Equal(t, 123.456, value)
	})

	t.Run("Get Non-Existing Float64 Value", func(t *testing.T) {
		t.Parallel()

		parsedGroup := &dynflags.ParsedGroup{
			Name:   "test-group",
			Values: map[string]interface{}{},
		}

		_, err := parsedGroup.GetFloat64("non-existing")
		assert.Error(t, err)
	})

	t.Run("Get Invalid Float64 Value", func(t *testing.T) {
		t.Parallel()

		parsedGroup := &dynflags.ParsedGroup{
			Name:   "test-group",
			Values: map[string]interface{}{"invalid-test": "not-a-float"},
		}

		_, err := parsedGroup.GetFloat64("invalid-test")
		assert.Error(t, err)
	})
}

func TestFloat64GetBound(t *testing.T) {
	t.Run("Float64Value - GetBound", func(t *testing.T) {
		var f *float64
		val := 3.14
		f = &val

		floatValue := dynflags.Float64Value{Bound: f}
		assert.Equal(t, 3.14, floatValue.GetBound())

		floatValue = dynflags.Float64Value{Bound: nil}
		assert.Nil(t, floatValue.GetBound())
	})
}
