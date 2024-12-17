package dynflags_test

import (
	"testing"

	"github.com/containeroo/portpatrol/pkg/dynflags"
	"github.com/stretchr/testify/assert"
)

func TestFloat64SlicesValue(t *testing.T) {
	t.Parallel()

	t.Run("Parse valid float64 value", func(t *testing.T) {
		t.Parallel()

		float64SlicesValue := dynflags.Float64SlicesValue{Bound: &[]float64{}}
		parsed, err := float64SlicesValue.Parse("3.14159")
		assert.NoError(t, err)
		assert.Equal(t, 3.14159, parsed)
	})

	t.Run("Parse invalid float64 value", func(t *testing.T) {
		t.Parallel()

		float64SlicesValue := dynflags.Float64SlicesValue{Bound: &[]float64{}}
		parsed, err := float64SlicesValue.Parse("invalid")
		assert.Error(t, err)
		assert.Nil(t, parsed)
	})

	t.Run("Set valid float64 value", func(t *testing.T) {
		t.Parallel()

		bound := []float64{1.23}
		float64SlicesValue := dynflags.Float64SlicesValue{Bound: &bound}

		err := float64SlicesValue.Set(4.56)
		assert.NoError(t, err)
		assert.Equal(t, []float64{1.23, 4.56}, bound)
	})

	t.Run("Set invalid type", func(t *testing.T) {
		t.Parallel()

		bound := []float64{}
		float64SlicesValue := dynflags.Float64SlicesValue{Bound: &bound}

		err := float64SlicesValue.Set("invalid")
		assert.Error(t, err)
		assert.EqualError(t, err, "invalid value type: expected float64")
	})
}

func TestGroupConfigFloat64Slices(t *testing.T) {
	t.Parallel()

	t.Run("Define float64 slices flag", func(t *testing.T) {
		t.Parallel()

		group := &dynflags.ConfigGroup{Flags: make(map[string]*dynflags.Flag)}
		defaultValue := []float64{1.23, 4.56}
		group.Float64Slices("float64SliceFlag", defaultValue, "A float64 slices flag")

		assert.Contains(t, group.Flags, "float64SliceFlag")
		assert.Equal(t, "A float64 slices flag", group.Flags["float64SliceFlag"].Usage)
		assert.Equal(t, "1.23,4.56", group.Flags["float64SliceFlag"].Default)
	})
}

func TestGetFloat64Slices(t *testing.T) {
	t.Parallel()

	t.Run("Retrieve []float64 value", func(t *testing.T) {
		t.Parallel()

		parsedGroup := &dynflags.ParsedGroup{
			Name:   "testGroup",
			Values: map[string]interface{}{"flag1": []float64{1.1, 2.2, 3.3}},
		}

		result, err := parsedGroup.GetFloat64Slices("flag1")
		assert.NoError(t, err)
		assert.Equal(t, []float64{1.1, 2.2, 3.3}, result)
	})

	t.Run("Retrieve single float64 value as []float64", func(t *testing.T) {
		t.Parallel()

		parsedGroup := &dynflags.ParsedGroup{
			Name:   "testGroup",
			Values: map[string]interface{}{"flag1": 42.42},
		}

		result, err := parsedGroup.GetFloat64Slices("flag1")
		assert.NoError(t, err)
		assert.Equal(t, []float64{42.42}, result)
	})

	t.Run("Flag not found", func(t *testing.T) {
		t.Parallel()

		parsedGroup := &dynflags.ParsedGroup{
			Name:   "testGroup",
			Values: map[string]interface{}{},
		}

		result, err := parsedGroup.GetFloat64Slices("nonExistentFlag")
		assert.Error(t, err)
		assert.Nil(t, result)
		assert.EqualError(t, err, "flag 'nonExistentFlag' not found in group 'testGroup'")
	})

	t.Run("Flag value is invalid type", func(t *testing.T) {
		t.Parallel()

		parsedGroup := &dynflags.ParsedGroup{
			Name:   "testGroup",
			Values: map[string]interface{}{"flag1": "invalid"},
		}

		result, err := parsedGroup.GetFloat64Slices("flag1")
		assert.Error(t, err)
		assert.Nil(t, result)
		assert.EqualError(t, err, "flag 'flag1' is not a []float64")
	})
}

func TestFloat64SlicesGetBound(t *testing.T) {
	t.Run("Float64SlicesValue - GetBound", func(t *testing.T) {
		var slices *[]float64
		val := []float64{1.1, 2.2, 3.3}
		slices = &val

		floatSlicesValue := dynflags.Float64SlicesValue{Bound: slices}
		assert.Equal(t, val, floatSlicesValue.GetBound())

		floatSlicesValue = dynflags.Float64SlicesValue{Bound: nil}
		assert.Nil(t, floatSlicesValue.GetBound())
	})
}
