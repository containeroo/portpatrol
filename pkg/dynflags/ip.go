package dynflags

import (
	"fmt"
	"net"
)

// IPValue implementation for URL flags
type IPValue struct {
	Bound *net.IP
}

func (u *IPValue) Parse(value string) (interface{}, error) {
	result := net.ParseIP(value)
	if result == nil {
		return nil, fmt.Errorf("invalid IP address: %s", value)
	}
	return &result, nil
}

func (u *IPValue) Set(value interface{}) error {
	if parsedIP, ok := value.(*net.IP); ok {
		*u.Bound = *parsedIP
		return nil
	}
	return fmt.Errorf("invalid value type: expected IP")
}

// IPVar defines an net.IP flag with specified name, default value, and usage string.
// The argument p points to an net.IP variable in which to store the value of the flag.
func (g *GroupConfig) IPVar(p *net.IP, name, value, usage string) {
	*p = *g.IP(name, value, usage)
}

// IP defines an net.IP flag with specified name, default value, and usage string.
// The return value is the address of an net.IP variable that stores the value of the flag.
func (g *GroupConfig) IP(name, value, usage string) *net.IP {
	bound := new(*net.IP)
	if value != "" {
		parsed := net.ParseIP(value)
		if parsed == nil {
			panic(fmt.Sprintf("%s has a invalid default IP flag '%s'", name, value))
		}
		*bound = &parsed // Copy the parsed URL into bound
	}
	g.Flags[name] = &Flag{
		Type:    FlagTypeURL,
		Default: value,
		Usage:   usage,
		Value:   &IPValue{Bound: *bound},
	}
	return *bound
}

// GetIP returns the net.IP value of a flag with the given name
func (pg *ParsedGroup) GetIP(flagName string) (net.IP, error) {
	value, exists := pg.Values[flagName]
	if !exists {
		return nil, fmt.Errorf("flag '%s' not found in group '%s'", flagName, pg.Name)
	}
	if ip, ok := value.(net.IP); ok {
		return ip, nil
	}

	return nil, fmt.Errorf("flag '%s' is not a IP", flagName)
}
