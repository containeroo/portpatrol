package dynflags_test

import (
	"testing"

	"github.com/containeroo/portpatrol/pkg/dynflags"
	"github.com/stretchr/testify/assert"
)

func TestBoolValue_Parse(t *testing.T) {
	t.Parallel()

	t.Run("ValidTrueValue", func(t *testing.T) {
		t.Parallel()

		b := &dynflags.BoolValue{Bound: new(bool)}
		value, err := b.Parse("true")
		assert.NoError(t, err)
		assert.Equal(t, true, value)
	})

	t.Run("ValidFalseValue", func(t *testing.T) {
		t.Parallel()

		b := &dynflags.BoolValue{Bound: new(bool)}
		value, err := b.Parse("false")
		assert.NoError(t, err)
		assert.Equal(t, false, value)
	})

	t.Run("InvalidValue", func(t *testing.T) {
		t.Parallel()

		b := &dynflags.BoolValue{Bound: new(bool)}
		value, err := b.Parse("invalid")
		assert.Error(t, err)
		assert.Equal(t, value, false)
	})
}

func TestBoolValue_Set(t *testing.T) {
	t.Parallel()

	t.Run("SetValidTrue", func(t *testing.T) {
		t.Parallel()

		bound := new(bool)
		b := &dynflags.BoolValue{Bound: bound}
		err := b.Set(true)
		assert.NoError(t, err)
		assert.Equal(t, true, *bound)
	})

	t.Run("SetValidFalse", func(t *testing.T) {
		t.Parallel()

		bound := new(bool)
		b := &dynflags.BoolValue{Bound: bound}
		err := b.Set(false)
		assert.NoError(t, err)
		assert.Equal(t, false, *bound)
	})

	t.Run("SetInvalidValue", func(t *testing.T) {
		t.Parallel()

		bound := new(bool)
		b := &dynflags.BoolValue{Bound: bound}
		err := b.Set(123) // Invalid type
		assert.Error(t, err)
		assert.EqualError(t, err, "invalid value type: expected bool")
	})
}

func TestGroupConfig_Bool(t *testing.T) {
	t.Parallel()

	t.Run("DefaultBool", func(t *testing.T) {
		t.Parallel()

		group := &dynflags.ConfigGroup{
			Flags: make(map[string]*dynflags.Flag),
		}
		boolVar := group.Bool("testBool", true, "Test boolean flag")
		assert.Equal(t, true, *boolVar)
		flag := group.Flags["testBool"]
		assert.NotNil(t, flag)
		assert.Equal(t, dynflags.FlagTypeBool, flag.Type)
		assert.Equal(t, true, flag.Default)
	})

	t.Run("BindBoolVar", func(t *testing.T) {
		t.Parallel()

		group := &dynflags.ConfigGroup{
			Flags: make(map[string]*dynflags.Flag),
		}
		var boundBool bool
		group.BoolVar(&boundBool, "testBoolVar", false, "Test bound boolean flag")
		assert.Equal(t, false, boundBool)
		flag := group.Flags["testBoolVar"]
		assert.NotNil(t, flag)
		assert.Equal(t, dynflags.FlagTypeBool, flag.Type)
		assert.Equal(t, false, flag.Default)
	})
}

func TestParsedGroup_GetBool(t *testing.T) {
	t.Parallel()

	t.Run("GetExistingBool", func(t *testing.T) {
		t.Parallel()

		parsedGroup := &dynflags.ParsedGroup{
			Name:   "testGroup",
			Values: map[string]interface{}{"testBool": true},
		}
		value, err := parsedGroup.GetBool("testBool")
		assert.NoError(t, err)
		assert.Equal(t, true, value)
	})

	t.Run("GetNonExistentBool", func(t *testing.T) {
		t.Parallel()

		parsedGroup := &dynflags.ParsedGroup{
			Name:   "testGroup",
			Values: map[string]interface{}{},
		}
		value, err := parsedGroup.GetBool("nonExistent")
		assert.Error(t, err)
		assert.Equal(t, false, value)
		assert.EqualError(t, err, "flag 'nonExistent' not found in group 'testGroup'")
	})

	t.Run("GetInvalidBoolType", func(t *testing.T) {
		t.Parallel()

		parsedGroup := &dynflags.ParsedGroup{
			Name:   "testGroup",
			Values: map[string]interface{}{"invalidBool": "notABool"},
		}
		value, err := parsedGroup.GetBool("invalidBool")
		assert.Error(t, err)
		assert.Equal(t, false, value)
		assert.EqualError(t, err, "flag 'invalidBool' is not a bool")
	})
}
