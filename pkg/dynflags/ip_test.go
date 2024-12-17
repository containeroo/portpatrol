package dynflags_test

import (
	"net"
	"testing"

	"github.com/containeroo/portpatrol/pkg/dynflags"
	"github.com/stretchr/testify/assert"
)

func TestIPValue(t *testing.T) {
	t.Parallel()

	t.Run("Parse valid IP address", func(t *testing.T) {
		t.Parallel()

		ipValue := dynflags.IPValue{}
		parsed, err := ipValue.Parse("192.168.1.1")
		assert.NoError(t, err)
		assert.NotNil(t, parsed)
		assert.Equal(t, "192.168.1.1", parsed.(*net.IP).String())
	})

	t.Run("Parse invalid IP address", func(t *testing.T) {
		t.Parallel()

		ipValue := dynflags.IPValue{}
		parsed, err := ipValue.Parse("invalid-ip")
		assert.Error(t, err)
		assert.Nil(t, parsed)
	})

	t.Run("Set valid IP value", func(t *testing.T) {
		t.Parallel()

		bound := net.ParseIP("0.0.0.0")
		ipValue := dynflags.IPValue{Bound: &bound}

		parsed := net.ParseIP("192.168.1.1")
		err := ipValue.Set(&parsed)
		assert.NoError(t, err)
		assert.Equal(t, "192.168.1.1", ipValue.Bound.String())
	})

	t.Run("Set invalid value type", func(t *testing.T) {
		t.Parallel()

		bound := net.ParseIP("0.0.0.0")
		ipValue := dynflags.IPValue{Bound: &bound}

		err := ipValue.Set("invalid-type")
		assert.Error(t, err)
		assert.EqualError(t, err, "invalid value type: expected IP")
	})
}

func TestGroupConfigIP(t *testing.T) {
	t.Parallel()

	t.Run("Define IP flag with valid default", func(t *testing.T) {
		t.Parallel()

		group := &dynflags.ConfigGroup{Flags: make(map[string]*dynflags.Flag)}
		defaultIP := "192.168.1.1"
		group.IP("ipFlag", defaultIP, "An example IP flag")

		assert.Contains(t, group.Flags, "ipFlag")
		assert.Equal(t, "An example IP flag", group.Flags["ipFlag"].Usage)
		assert.Equal(t, defaultIP, group.Flags["ipFlag"].Default)
	})

	t.Run("Define IP flag with invalid default", func(t *testing.T) {
		t.Parallel()

		group := &dynflags.ConfigGroup{Flags: make(map[string]*dynflags.Flag)}

		assert.PanicsWithValue(t,
			"ipFlag has a invalid default IP flag 'invalid-ip'",
			func() {
				group.IP("ipFlag", "invalid-ip", "Invalid IP flag")
			})
	})
}

func TestParsedGroupGetIP(t *testing.T) {
	t.Parallel()

	t.Run("Get existing IP flag", func(t *testing.T) {
		t.Parallel()

		group := &dynflags.ParsedGroup{
			Values: map[string]interface{}{
				"ipFlag": net.ParseIP("192.168.1.1"),
			},
		}
		ip, err := group.GetIP("ipFlag")
		assert.NoError(t, err)
		assert.Equal(t, "192.168.1.1", ip.String())
	})

	t.Run("Get non-existent IP flag", func(t *testing.T) {
		t.Parallel()

		group := &dynflags.ParsedGroup{
			Values: map[string]interface{}{},
		}
		ip, err := group.GetIP("ipFlag")
		assert.Error(t, err)
		assert.Nil(t, ip)
		assert.EqualError(t, err, "flag 'ipFlag' not found in group ''")
	})

	t.Run("Get IP flag with invalid type", func(t *testing.T) {
		t.Parallel()

		group := &dynflags.ParsedGroup{
			Values: map[string]interface{}{
				"ipFlag": "not-an-ip",
			},
		}
		ip, err := group.GetIP("ipFlag")
		assert.Error(t, err)
		assert.Nil(t, ip)
		assert.EqualError(t, err, "flag 'ipFlag' is not a IP")
	})
}

func TestIPGetBound(t *testing.T) {
	t.Run("IPValue - GetBound", func(t *testing.T) {
		var ip *net.IP
		val := net.ParseIP("127.0.0.1")
		ip = &val

		ipValue := dynflags.IPValue{Bound: ip}
		assert.Equal(t, val, ipValue.GetBound())

		ipValue = dynflags.IPValue{Bound: nil}
		assert.Nil(t, ipValue.GetBound())
	})
}
