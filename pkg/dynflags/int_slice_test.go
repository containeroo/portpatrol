package dynflags_test

import (
	"testing"

	"github.com/containeroo/portpatrol/pkg/dynflags"
	"github.com/stretchr/testify/assert"
)

func TestIntSlicesValue(t *testing.T) {
	t.Parallel()

	t.Run("Parse valid int slice value", func(t *testing.T) {
		t.Parallel()

		intSlicesValue := dynflags.IntSlicesValue{Bound: &[]int{}}
		parsed, err := intSlicesValue.Parse("123")
		assert.NoError(t, err)
		assert.Equal(t, 123, parsed)
	})

	t.Run("Parse invalid int slice value", func(t *testing.T) {
		t.Parallel()

		intSlicesValue := dynflags.IntSlicesValue{Bound: &[]int{}}
		parsed, err := intSlicesValue.Parse("invalid")
		assert.Error(t, err)
		assert.Nil(t, parsed)
	})

	t.Run("Set valid int slice value", func(t *testing.T) {
		t.Parallel()

		bound := []int{1}
		intSlicesValue := dynflags.IntSlicesValue{Bound: &bound}

		err := intSlicesValue.Set(2)
		assert.NoError(t, err)
		assert.Equal(t, []int{1, 2}, bound)
	})

	t.Run("Set invalid type", func(t *testing.T) {
		t.Parallel()

		bound := []int{1}
		intSlicesValue := dynflags.IntSlicesValue{Bound: &bound}

		err := intSlicesValue.Set("invalid") // Invalid type
		assert.Error(t, err)
		assert.EqualError(t, err, "invalid value type: expected int")
	})
}

func TestGroupConfigIntSlices(t *testing.T) {
	t.Parallel()

	t.Run("Define int slices flag", func(t *testing.T) {
		t.Parallel()

		group := &dynflags.ConfigGroup{Flags: make(map[string]*dynflags.Flag)}
		defaultValue := []int{1, 2}
		intSlice := group.IntSlices("intSliceFlag", defaultValue, "An int slices flag")

		assert.Equal(t, []int{1, 2}, *intSlice)
		assert.Contains(t, group.Flags, "intSliceFlag")
		assert.Equal(t, "An int slices flag", group.Flags["intSliceFlag"].Usage)
		assert.Equal(t, "1,2", group.Flags["intSliceFlag"].Default)
	})

	t.Run("Define IntSlicesVar and set value", func(t *testing.T) {
		t.Parallel()

		group := &dynflags.ConfigGroup{Flags: make(map[string]*dynflags.Flag)}
		var intSlice []int
		group.IntSlicesVar(&intSlice, "intSliceFlag", []int{1, 2}, "Int slices flag variable")
		assert.Equal(t, []int{1, 2}, intSlice)
	})
}

func TestParsedGroupGetIntSlices(t *testing.T) {
	t.Parallel()

	t.Run("Get existing int slices flag", func(t *testing.T) {
		t.Parallel()

		group := &dynflags.ParsedGroup{
			Values: map[string]interface{}{
				"intSliceFlag": []int{1, 2},
			},
		}
		slice, err := group.GetIntSlices("intSliceFlag")
		assert.NoError(t, err)
		assert.Equal(t, []int{1, 2}, slice)
	})

	t.Run("Get non-existent int slices flag", func(t *testing.T) {
		t.Parallel()

		group := &dynflags.ParsedGroup{
			Values: map[string]interface{}{},
		}
		slice, err := group.GetIntSlices("intSliceFlag")
		assert.Error(t, err)
		assert.Nil(t, slice)
		assert.EqualError(t, err, "flag 'intSliceFlag' not found in group ''")
	})

	t.Run("Get int slices flag with invalid type", func(t *testing.T) {
		t.Parallel()

		group := &dynflags.ParsedGroup{
			Values: map[string]interface{}{
				"intSliceFlag": "invalid", // Invalid type
			},
		}
		slice, err := group.GetIntSlices("intSliceFlag")
		assert.Error(t, err)
		assert.Nil(t, slice)
		assert.EqualError(t, err, "flag 'intSliceFlag' is not a []int")
	})
}
