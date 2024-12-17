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
		group.BoolSlices("boolSliceFlag", defaultValue, "A bool slices flag")

		assert.Contains(t, group.Flags, "boolSliceFlag")
		assert.Equal(t, "A bool slices flag", group.Flags["boolSliceFlag"].Usage)
		assert.Equal(t, "true,false", group.Flags["boolSliceFlag"].Default)
	})
}

func TestGetBoolSlices(t *testing.T) {
	t.Parallel()

	t.Run("Retrieve []bool value", func(t *testing.T) {
		t.Parallel()

		parsedGroup := &dynflags.ParsedGroup{
			Name: "testGroup",
			Values: map[string]interface{}{
				"flag1": []bool{true, false, true},
			},
		}

		result, err := parsedGroup.GetBoolSlices("flag1")
		assert.NoError(t, err)
		assert.Equal(t, []bool{true, false, true}, result)
	})

	t.Run("Retrieve single bool value as []bool", func(t *testing.T) {
		t.Parallel()

		parsedGroup := &dynflags.ParsedGroup{
			Name: "testGroup",
			Values: map[string]interface{}{
				"flag1": true,
			},
		}

		result, err := parsedGroup.GetBoolSlices("flag1")
		assert.NoError(t, err)
		assert.Equal(t, []bool{true}, result)
	})

	t.Run("Flag not found", func(t *testing.T) {
		t.Parallel()

		parsedGroup := &dynflags.ParsedGroup{
			Name:   "testGroup",
			Values: map[string]interface{}{},
		}

		result, err := parsedGroup.GetBoolSlices("nonExistentFlag")
		assert.Error(t, err)
		assert.Nil(t, result)
		assert.EqualError(t, err, "flag 'nonExistentFlag' not found in group 'testGroup'")
	})

	t.Run("Flag value is invalid type", func(t *testing.T) {
		t.Parallel()

		parsedGroup := &dynflags.ParsedGroup{
			Name: "testGroup",
			Values: map[string]interface{}{
				"flag1": "invalid",
			},
		}

		result, err := parsedGroup.GetBoolSlices("flag1")
		assert.Error(t, err)
		assert.Nil(t, result)
		assert.EqualError(t, err, "flag 'flag1' is not a []bool")
	})
}

func TestBoolSlicesGetBound(t *testing.T) {
	t.Run("BoolSlicesValue - GetBound", func(t *testing.T) {
		var slices *[]bool
		val := []bool{true, false, true}
		slices = &val

		boolSlicesValue := dynflags.BoolSlicesValue{Bound: slices}
		assert.Equal(t, val, boolSlicesValue.GetBound())

		boolSlicesValue = dynflags.BoolSlicesValue{Bound: nil}
		assert.Nil(t, boolSlicesValue.GetBound())
	})
}
