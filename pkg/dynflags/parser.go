package dynflags

import (
	"fmt"
	"strings"
)

func (df *DynFlags) Parse(args []string) error {
	for i := 0; i < len(args); i++ {
		arg := args[i]
		if !strings.HasPrefix(arg, "--") {
			return fmt.Errorf("invalid flag format: %s", arg)
		}

		var fullKey, value string
		if strings.Contains(arg, "=") {
			parts := strings.SplitN(arg[2:], "=", 2)
			fullKey, value = parts[0], parts[1]
		} else {
			fullKey = arg[2:]
			if i+1 < len(args) && !strings.HasPrefix(args[i+1], "--") {
				value = args[i+1]
				i++
			} else {
				return fmt.Errorf("missing value for flag: %s", fullKey)
			}
		}

		keyParts := strings.Split(fullKey, ".")
		if len(keyParts) < 3 {
			return fmt.Errorf("flag must follow the pattern: --group.identifier.key=value")
		}
		groupName, identifier, flagName := keyParts[0], keyParts[1], keyParts[2]

		if _, exists := df.Groups[groupName]; !exists {
			df.Groups[groupName] = make(map[string]*Group)
		}
		if _, exists := df.Groups[groupName][identifier]; !exists {
			df.Groups[groupName][identifier] = &Group{
				Name:  identifier,
				Flags: make(map[string]*Flag),
			}
		}

		group := df.Groups[groupName][identifier]
		flag, exists := group.Flags[flagName]
		if !exists {
			switch df.ParseBehavior {
			case ExitOnError:
				return fmt.Errorf("unknown flag '%s' in group '%s.%s'", flagName, groupName, identifier)
			case ContinueOnError:
				continue
			case IgnoreUnknown:
				// Do nothing and continue parsing
				continue
			}
		}

		parsedValue, err := flag.Parser.Parse(value)
		if err != nil {
			return fmt.Errorf("failed to parse flag '%s': %v", fullKey, err)
		}
		flag.Value = parsedValue
	}
	return nil
}
