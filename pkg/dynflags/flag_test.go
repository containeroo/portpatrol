package dynflags_test

import (
	"testing"

	"github.com/containeroo/portpatrol/pkg/dynflags"
	"github.com/stretchr/testify/assert"
)

func TestFlagGetValue(t *testing.T) {
	t.Parallel()

	t.Run("Nil Flag - GetValue", func(t *testing.T) {
		t.Parallel()
		var flag *dynflags.Flag
		assert.Nil(t, flag.GetValue(), "Expected nil when flag is nil")
	})

	t.Run("String Flag - GetValue", func(t *testing.T) {
		t.Parallel()

		group := &dynflags.ConfigGroup{
			Name:  "testGroup",
			Flags: make(map[string]*dynflags.Flag),
		}

		flag := group.String("example", "default-value", "An example string flag")
		assert.Equal(t, "default-value", flag.GetValue(), "Expected GetValue() to return the default value")
	})
}
