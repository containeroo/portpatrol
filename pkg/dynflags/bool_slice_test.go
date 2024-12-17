package dynflags_test

import (
	"testing"

	"github.com/containeroo/portpatrol/pkg/dynflags"
	"github.com/stretchr/testify/assert"
)

func TestBoolSlicesValue(t *testing.T) {
	t.Parallel()

	t.Run("Parse valid bool value", func(t *testing.T) {
		t.Parallel()

		boolSlicesValue := dynflags.BoolSlicesValue{Bound: &[]bool{}}
		parsed, err := boolSlicesValue.Parse("true")
		assert.NoError(t, err)
		assert.Equal(t, true, parsed)
	})

	t.Run("Parse invalid bool value", func(t *testing.T) {
		t.Parallel()

		boolSlicesValue := dynflags.BoolSlicesValue{Bound: &[]bool{}}
		parsed, err := boolSlicesValue.Parse("invalid")
		assert.Error(t, err)
		assert.Nil(t, parsed)
	})

	t.Run("Set valid bool value", func(t *testing.T) {
		t.Parallel()

		bound := []bool{true}
		boolSlicesValue := dynflags.BoolSlicesValue{Bound: &bound}

		err := boolSlicesValue.Set(false)
		assert.NoError(t, err)
		assert.Equal(t, []bool{true, false}, bound)
	})

	t.Run("Set invalid type", func(t *testing.T) {
		t.Parallel()

		bound := []bool{}
		boolSlicesValue := dynflags.BoolSlicesValue{Bound: &bound}

		err := boolSlicesValue.Set("invalid")
		assert.Error(t, err)
		assert.EqualError(t, err, "invalid value type: expected bool")
	})
}

func TestGroupConfigBoolSlices(t *testing.T) {
	t.Parallel()

	t.Run("Define bool slices flag", func(t *testing.T) {
		t.Parallel()

		group := &dynflags.ConfigGroup{Flags: make(map[string]*dynflags.Flag)}
		defaultValue := []bool{true, false}
		boolSlice := group.BoolSlices("boolSliceFlag", defaultValue, "A bool slices flag")

		assert.Equal(t, []bool{true, false}, *boolSlice)
		assert.Contains(t, group.Flags, "boolSliceFlag")
		assert.Equal(t, "A bool slices flag", group.Flags["boolSliceFlag"].Usage)
		assert.Equal(t, "true,false", group.Flags["boolSliceFlag"].Default)
	})

	t.Run("Define BoolSlicesVar and set value", func(t *testing.T) {
		t.Parallel()

		group := &dynflags.ConfigGroup{Flags: make(map[string]*dynflags.Flag)}
		var boolSlice []bool
		group.BoolSlicesVar(&boolSlice, "boolSliceFlag", []bool{true}, "Bool slices flag variable")
		assert.Equal(t, []bool{true}, boolSlice)
	})
}

func TestParsedGroupGetBoolSlices(t *testing.T) {
	t.Parallel()

	t.Run("Get existing bool slices flag", func(t *testing.T) {
		t.Parallel()

		group := &dynflags.ParsedGroup{
			Values: map[string]interface{}{
				"boolSliceFlag": []bool{true, false},
			},
		}
		slice, err := group.GetBoolSlices("boolSliceFlag")
		assert.NoError(t, err)
		assert.Equal(t, []bool{true, false}, slice)
	})

	t.Run("Get non-existent bool slices flag", func(t *testing.T) {
		t.Parallel()

		group := &dynflags.ParsedGroup{
			Values: map[string]interface{}{},
		}
		slice, err := group.GetBoolSlices("boolSliceFlag")
		assert.Error(t, err)
		assert.Nil(t, slice)
		assert.EqualError(t, err, "flag 'boolSliceFlag' not found in group ''")
	})

	t.Run("Get bool slices flag with invalid type", func(t *testing.T) {
		t.Parallel()

		group := &dynflags.ParsedGroup{
			Values: map[string]interface{}{
				"boolSliceFlag": "invalid-type", // Invalid type
			},
		}
		slice, err := group.GetBoolSlices("boolSliceFlag")
		assert.Error(t, err)
		assert.Nil(t, slice)
		assert.EqualError(t, err, "flag 'boolSliceFlag' is not a []bool")
	})
}
