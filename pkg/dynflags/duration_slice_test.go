package dynflags_test

import (
	"testing"
	"time"

	"github.com/containeroo/portpatrol/pkg/dynflags"
	"github.com/stretchr/testify/assert"
)

func TestDurationSlicesValue(t *testing.T) {
	t.Parallel()

	t.Run("Parse valid duration slice value", func(t *testing.T) {
		t.Parallel()

		durationSlicesValue := dynflags.DurationSlicesValue{Bound: &[]time.Duration{}}
		parsed, err := durationSlicesValue.Parse("5s")
		assert.NoError(t, err)
		assert.Equal(t, 5*time.Second, parsed)
	})

	t.Run("Parse invalid duration value", func(t *testing.T) {
		t.Parallel()

		durationSlicesValue := dynflags.DurationSlicesValue{Bound: &[]time.Duration{}}
		parsed, err := durationSlicesValue.Parse("invalid")
		assert.Error(t, err)
		assert.Nil(t, parsed)
	})

	t.Run("Set valid duration value", func(t *testing.T) {
		t.Parallel()

		bound := []time.Duration{1 * time.Second}
		durationSlicesValue := dynflags.DurationSlicesValue{Bound: &bound}

		err := durationSlicesValue.Set(2 * time.Second)
		assert.NoError(t, err)
		assert.Equal(t, []time.Duration{1 * time.Second, 2 * time.Second}, bound)
	})

	t.Run("Set invalid type", func(t *testing.T) {
		t.Parallel()

		bound := []time.Duration{}
		durationSlicesValue := dynflags.DurationSlicesValue{Bound: &bound}

		err := durationSlicesValue.Set("invalid")
		assert.Error(t, err)
		assert.EqualError(t, err, "invalid value type: expected time.Duration")
	})
}

func TestGroupConfigDurationSlices(t *testing.T) {
	t.Parallel()

	t.Run("Define duration slices flag", func(t *testing.T) {
		t.Parallel()

		group := &dynflags.GroupConfig{Flags: make(map[string]*dynflags.Flag)}
		defaultValue := []time.Duration{1 * time.Second, 2 * time.Second}
		durationSlice := group.DurationSlices("durationSliceFlag", defaultValue, "A duration slices flag")

		assert.Equal(t, []time.Duration{1 * time.Second, 2 * time.Second}, *durationSlice)
		assert.Contains(t, group.Flags, "durationSliceFlag")
		assert.Equal(t, "A duration slices flag", group.Flags["durationSliceFlag"].Usage)
		assert.Equal(t, "1s,2s", group.Flags["durationSliceFlag"].Default)
	})

	t.Run("Define DurationSlicesVar and set value", func(t *testing.T) {
		t.Parallel()

		group := &dynflags.GroupConfig{Flags: make(map[string]*dynflags.Flag)}
		var durationSlice []time.Duration
		group.DurationSlicesVar(&durationSlice, "durationSliceFlag", []time.Duration{3 * time.Second}, "Duration slices flag variable")
		assert.Equal(t, []time.Duration{3 * time.Second}, durationSlice)
	})
}

func TestParsedGroupGetDurationSlices(t *testing.T) {
	t.Parallel()

	t.Run("Get existing duration slices flag", func(t *testing.T) {
		t.Parallel()

		group := &dynflags.ParsedGroup{
			Values: map[string]interface{}{
				"durationSliceFlag": []time.Duration{1 * time.Second, 2 * time.Second},
			},
		}
		slice, err := group.GetDurationSlices("durationSliceFlag")
		assert.NoError(t, err)
		assert.Equal(t, []time.Duration{1 * time.Second, 2 * time.Second}, slice)
	})

	t.Run("Get non-existent duration slices flag", func(t *testing.T) {
		t.Parallel()

		group := &dynflags.ParsedGroup{
			Values: map[string]interface{}{},
		}
		slice, err := group.GetDurationSlices("durationSliceFlag")
		assert.Error(t, err)
		assert.Nil(t, slice)
		assert.EqualError(t, err, "flag 'durationSliceFlag' not found in group ''")
	})

	t.Run("Get duration slices flag with invalid type", func(t *testing.T) {
		t.Parallel()

		group := &dynflags.ParsedGroup{
			Values: map[string]interface{}{
				"durationSliceFlag": "invalid-type", // Invalid type
			},
		}
		slice, err := group.GetDurationSlices("durationSliceFlag")
		assert.Error(t, err)
		assert.Nil(t, slice)
		assert.EqualError(t, err, "flag 'durationSliceFlag' is not a []time.Duration")
	})
}
