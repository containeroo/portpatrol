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

// IP defines an net.IP flag with specified name, default value, and usage string.
// The return value is the address of an net.IP variable that stores the value of the flag.
func (g *ConfigGroup) IP(name, value, usage string) *Flag {
	bound := new(*net.IP)
	if value != "" {
		parsed := net.ParseIP(value)
		if parsed == nil {
			panic(fmt.Sprintf("%s has a invalid default IP flag '%s'", name, value))
		}
		*bound = &parsed // Copy the parsed URL into bound
	}
	flag := &Flag{
		Type:    FlagTypeIP,
		Default: value,
		Usage:   usage,
		Value:   &IPValue{Bound: *bound},
	}
	g.Flags[name] = flag
	g.flagOrder = append(g.flagOrder, name)
	return flag
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
