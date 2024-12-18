package dynflags

import (
	"fmt"
	"net"
	"strings"
)

type IPSlicesValue struct {
	Bound *[]net.IP
}

func (i *IPSlicesValue) GetBound() interface{} {
	if i.Bound == nil {
		return nil
	}
	return *i.Bound
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

// IPSlices defines an IP slice flag with the specified name, default value, and usage description.
// The flag is added to the group's flag list and returned as a *Flag instance.
func (g *ConfigGroup) IPSlices(name string, value []net.IP, usage string) *Flag {
	bound := &value
	defaultValue := make([]string, len(value))
	for i, ip := range value {
		defaultValue[i] = ip.String()
	}

	flag := &Flag{
		Type:    FlagTypeIPSlice,
		Default: strings.Join(defaultValue, ","),
		Usage:   usage,
		value:   &IPSlicesValue{Bound: bound},
	}
	g.Flags[name] = flag
	g.flagOrder = append(g.flagOrder, name)
	return flag
}

// GetIPSlices returns the []net.IP value of a flag with the given name
func (pg *ParsedGroup) GetIPSlices(flagName string) ([]net.IP, error) {
	value, exists := pg.Values[flagName]
	if !exists {
		return nil, fmt.Errorf("flag '%s' not found in group '%s'", flagName, pg.Name)
	}

	if ipSlice, ok := value.([]net.IP); ok {
		return ipSlice, nil
	}

	if i, ok := value.(net.IP); ok {
		return []net.IP{i}, nil
	}

	return nil, fmt.Errorf("flag '%s' is not a []net.IP", flagName)
}
