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

		group := &dynflags.ConfigGroup{Flags: make(map[string]*dynflags.Flag)}
		defaultValue := []time.Duration{1 * time.Second, 2 * time.Second}
		group.DurationSlices("durationSliceFlag", defaultValue, "A duration slices flag")

		assert.Contains(t, group.Flags, "durationSliceFlag")
		assert.Equal(t, "A duration slices flag", group.Flags["durationSliceFlag"].Usage)
		assert.Equal(t, "1s,2s", group.Flags["durationSliceFlag"].Default)
	})
}

func TestGetDurationSlices(t *testing.T) {
	t.Parallel()

	t.Run("Retrieve []time.Duration value", func(t *testing.T) {
		t.Parallel()

		parsedGroup := &dynflags.ParsedGroup{
			Name: "testGroup",
			Values: map[string]interface{}{
				"flag1": []time.Duration{1 * time.Second, 2 * time.Second, 3 * time.Second},
			},
		}

		result, err := parsedGroup.GetDurationSlices("flag1")
		assert.NoError(t, err)
		assert.Equal(t, []time.Duration{1 * time.Second, 2 * time.Second, 3 * time.Second}, result)
	})

	t.Run("Retrieve single time.Duration value as []time.Duration", func(t *testing.T) {
		t.Parallel()

		parsedGroup := &dynflags.ParsedGroup{
			Name: "testGroup",
			Values: map[string]interface{}{
				"flag1": 5 * time.Second,
			},
		}

		result, err := parsedGroup.GetDurationSlices("flag1")
		assert.NoError(t, err)
		assert.Equal(t, []time.Duration{5 * time.Second}, result)
	})

	t.Run("Flag not found", func(t *testing.T) {
		t.Parallel()

		parsedGroup := &dynflags.ParsedGroup{
			Name:   "testGroup",
			Values: map[string]interface{}{},
		}

		result, err := parsedGroup.GetDurationSlices("nonExistentFlag")
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

		result, err := parsedGroup.GetDurationSlices("flag1")
		assert.Error(t, err)
		assert.Nil(t, result)
		assert.EqualError(t, err, "flag 'flag1' is not a []time.Duration")
	})
}
