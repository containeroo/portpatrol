package dynflags

import (
	"fmt"
	"net"
	"strings"
)

// IPSlicesValue implementation for IP slice flags
type IPSlicesValue struct {
	Bound *[]net.IP
}

func (s *IPSlicesValue) Parse(value string) (interface{}, error) {
	ip := net.ParseIP(value)
	if ip == nil {
		return nil, fmt.Errorf("invalid IP address: %s", value)
	}
	return ip, nil
}

func (s *IPSlicesValue) Set(value interface{}) error {
	if ip, ok := value.(net.IP); ok {
		*s.Bound = append(*s.Bound, ip)
		return nil
	}
	return fmt.Errorf("invalid value type: expected net.IP")
}

// IPSlicesVar defines an IP slice flag with specified name, default value, and usage string.
// The argument p points to a slice of IPs in which to store the value of the flag.
func (g *ConfigGroup) IPSlicesVar(p *[]net.IP, name string, value []net.IP, usage string) {
	*p = *g.IPSlices(name, value, usage)
}

// IPSlices defines an IP slice flag with specified name, default value, and usage string.
// The return value is the address of a slice of IPs that stores the value of the flag.
func (g *ConfigGroup) IPSlices(name string, value []net.IP, usage string) *[]net.IP {
	bound := &value
	defaultValue := make([]string, len(value))
	for i, ip := range value {
		defaultValue[i] = ip.String()
	}

	g.Flags[name] = &Flag{
		Type:    FlagTypeIPSlice,
		Default: strings.Join(defaultValue, ","),
		Usage:   usage,
		Value:   &IPSlicesValue{Bound: bound},
	}
	g.flagOrder = append(g.flagOrder, name)
	return bound
}

// GetIPSlices returns the []net.IP value of a flag with the given name
func (pg *ParsedGroup) GetIPSlices(flagName string) ([]net.IP, error) {
	value, exists := pg.Values[flagName]
	if !exists {
		return nil, fmt.Errorf("flag '%s' not found in group '%s'", flagName, pg.Name)
	}
	if slice, ok := value.([]net.IP); ok {
		return slice, nil
	}

	return nil, fmt.Errorf("flag '%s' is not a []net.IP", flagName)
}
