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

		group := &dynflags.GroupConfig{Flags: make(map[string]*dynflags.Flag)}
		defaultValue := []float64{1.23, 4.56}
		floatSlice := group.Float64Slices("float64SliceFlag", defaultValue, "A float64 slices flag")

		assert.Equal(t, []float64{1.23, 4.56}, *floatSlice)
		assert.Contains(t, group.Flags, "float64SliceFlag")
		assert.Equal(t, "A float64 slices flag", group.Flags["float64SliceFlag"].Usage)
		assert.Equal(t, "1.23,4.56", group.Flags["float64SliceFlag"].Default)
	})

	t.Run("Define Float64SlicesVar and set value", func(t *testing.T) {
		t.Parallel()

		group := &dynflags.GroupConfig{Flags: make(map[string]*dynflags.Flag)}
		var floatSlice []float64
		group.Float64SlicesVar(&floatSlice, "float64SliceFlag", []float64{1.23, 4.56}, "Float64 slices flag variable")
		assert.Equal(t, []float64{1.23, 4.56}, floatSlice)
	})
}

func TestParsedGroupGetFloat64Slices(t *testing.T) {
	t.Parallel()

	t.Run("Get existing float64 slices flag", func(t *testing.T) {
		t.Parallel()

		group := &dynflags.ParsedGroup{
			Values: map[string]interface{}{
				"float64SliceFlag": []float64{1.23, 4.56},
			},
		}
		slice, err := group.GetFloat64Slices("float64SliceFlag")
		assert.NoError(t, err)
		assert.Equal(t, []float64{1.23, 4.56}, slice)
	})

	t.Run("Get non-existent float64 slices flag", func(t *testing.T) {
		t.Parallel()

		group := &dynflags.ParsedGroup{
			Values: map[string]interface{}{},
		}
		slice, err := group.GetFloat64Slices("float64SliceFlag")
		assert.Error(t, err)
		assert.Nil(t, slice)
		assert.EqualError(t, err, "flag 'float64SliceFlag' not found in group ''")
	})

	t.Run("Get float64 slices flag with invalid type", func(t *testing.T) {
		t.Parallel()

		group := &dynflags.ParsedGroup{
			Values: map[string]interface{}{
				"float64SliceFlag": "invalid-type", // Invalid type
			},
		}
		slice, err := group.GetFloat64Slices("float64SliceFlag")
		assert.Error(t, err)
		assert.Nil(t, slice)
		assert.EqualError(t, err, "flag 'float64SliceFlag' is not a []float64")
	})
}
