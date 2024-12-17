package dynflags_test

import (
	"net"
	"testing"

	"github.com/containeroo/portpatrol/pkg/dynflags"
	"github.com/stretchr/testify/assert"
)

func TestIPSlicesValue(t *testing.T) {
	t.Parallel()

	t.Run("Parse valid IP", func(t *testing.T) {
		t.Parallel()

		ipSlicesValue := dynflags.IPSlicesValue{Bound: &[]net.IP{}}
		parsed, err := ipSlicesValue.Parse("192.168.0.1")
		assert.NoError(t, err)
		assert.Equal(t, net.ParseIP("192.168.0.1"), parsed)
	})

	t.Run("Parse invalid IP", func(t *testing.T) {
		t.Parallel()

		ipSlicesValue := dynflags.IPSlicesValue{Bound: &[]net.IP{}}
		parsed, err := ipSlicesValue.Parse("invalid-ip")
		assert.Error(t, err)
		assert.Nil(t, parsed)
	})

	t.Run("Set valid IP", func(t *testing.T) {
		t.Parallel()

		bound := []net.IP{net.ParseIP("192.168.0.1")}
		ipSlicesValue := dynflags.IPSlicesValue{Bound: &bound}

		err := ipSlicesValue.Set(net.ParseIP("10.0.0.1"))
		assert.NoError(t, err)
		assert.Equal(t, []net.IP{net.ParseIP("192.168.0.1"), net.ParseIP("10.0.0.1")}, bound)
	})

	t.Run("Set invalid type", func(t *testing.T) {
		t.Parallel()

		bound := []net.IP{}
		ipSlicesValue := dynflags.IPSlicesValue{Bound: &bound}

		err := ipSlicesValue.Set("invalid-ip-type")
		assert.Error(t, err)
		assert.EqualError(t, err, "invalid value type: expected net.IP")
	})
}

func TestGroupConfigIPSlices(t *testing.T) {
	t.Parallel()

	t.Run("Define IP slices flag", func(t *testing.T) {
		t.Parallel()

		group := &dynflags.ConfigGroup{Flags: make(map[string]*dynflags.Flag)}
		defaultValue := []net.IP{net.ParseIP("192.168.0.1"), net.ParseIP("10.0.0.1")}
		ipSlice := group.IPSlices("ipSliceFlag", defaultValue, "An IP slices flag")

		assert.Equal(t, []net.IP{net.ParseIP("192.168.0.1"), net.ParseIP("10.0.0.1")}, *ipSlice)
		assert.Contains(t, group.Flags, "ipSliceFlag")
		assert.Equal(t, "An IP slices flag", group.Flags["ipSliceFlag"].Usage)
		assert.Equal(t, "192.168.0.1,10.0.0.1", group.Flags["ipSliceFlag"].Default)
	})

	t.Run("Define IPSlicesVar and set value", func(t *testing.T) {
		t.Parallel()

		group := &dynflags.ConfigGroup{Flags: make(map[string]*dynflags.Flag)}
		var ipSlice []net.IP
		group.IPSlicesVar(&ipSlice, "ipSliceFlag", []net.IP{net.ParseIP("8.8.8.8")}, "IP slices flag variable")
		assert.Equal(t, []net.IP{net.ParseIP("8.8.8.8")}, ipSlice)
	})
}

func TestGetIPSlices(t *testing.T) {
	t.Parallel()

	t.Run("Retrieve []net.IP value", func(t *testing.T) {
		t.Parallel()

		ip1 := net.ParseIP("192.168.1.1")
		ip2 := net.ParseIP("10.0.0.1")

		parsedGroup := &dynflags.ParsedGroup{
			Name:   "testGroup",
			Values: map[string]interface{}{"flag1": []net.IP{ip1, ip2}},
		}

		result, err := parsedGroup.GetIPSlices("flag1")
		assert.NoError(t, err)
		assert.Equal(t, []net.IP{ip1, ip2}, result)
	})

	t.Run("Retrieve single net.IP value as []net.IP", func(t *testing.T) {
		t.Parallel()

		ip := net.ParseIP("127.0.0.1")

		parsedGroup := &dynflags.ParsedGroup{
			Name:   "testGroup",
			Values: map[string]interface{}{"flag1": ip},
		}

		result, err := parsedGroup.GetIPSlices("flag1")
		assert.NoError(t, err)
		assert.Equal(t, []net.IP{ip}, result)
	})

	t.Run("Flag not found", func(t *testing.T) {
		t.Parallel()

		parsedGroup := &dynflags.ParsedGroup{
			Name:   "testGroup",
			Values: map[string]interface{}{},
		}

		result, err := parsedGroup.GetIPSlices("nonExistentFlag")
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

		result, err := parsedGroup.GetIPSlices("flag1")
		assert.Error(t, err)
		assert.Nil(t, result)
		assert.EqualError(t, err, "flag 'flag1' is not a []net.IP")
	})
}
