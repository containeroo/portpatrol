package dynflags

import (
	"fmt"
	"sort"
	"strings"
	"text/tabwriter"
)

// PrintDefaults prints all registered flags
func (df *DynFlags) PrintDefaults() {
	w := tabwriter.NewWriter(df.output, 0, 8, 2, ' ', 0)
	defer w.Flush()

	// Print title if present
	if df.title != "" {
		fmt.Fprintln(df.output, df.title)
		fmt.Fprintln(df.output)
	}

	// Print description if present
	if df.description != "" {
		fmt.Fprintln(df.output, df.description)
		fmt.Fprintln(df.output)
	}

	// Sort group names
	if df.SortGroups {
		sort.Strings(df.groupOrder)
	}

	// Iterate over groups in the order they were added
	for _, groupName := range df.groupOrder {
		group := df.configGroups[groupName]

		// Print group usage or fallback to uppercase group name
		if group.usage != "" {
			fmt.Fprintln(w, group.usage)
		} else {
			fmt.Fprintln(w, strings.ToUpper(groupName))
		}

		// Sort flag names
		if df.SortFlags {
			sort.Strings(group.flagOrder)
		}

		// Print flags for the group
		fmt.Fprintln(w, "  Flag\tUsage")
		for _, flagName := range group.flagOrder {
			flag := group.Flags[flagName]
			usage := flag.Usage
			if flag.Default != nil && flag.Default != "" {
				usage = fmt.Sprintf("%s (default: %v)", flag.Usage, flag.Default)
			}
			fmt.Fprintf(w, "  --%s.<IDENTIFIER>.%s %s\t%s\n", groupName, flagName, flag.Type, usage)
		}
		fmt.Fprintln(w, "")
	}

	// Print epilog if present
	if df.epilog != "" {
		fmt.Fprintln(df.output)
		fmt.Fprintln(df.output, df.epilog)
	}
}
