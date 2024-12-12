package dynflags

import (
	"fmt"
	"strings"
	"text/tabwriter"
)

// PrintDefaults prints all registered flags
func (df *DynFlags) PrintDefaults() {
	w := tabwriter.NewWriter(df.output, 0, 8, 2, ' ', 0)
	defer w.Flush()

	if df.title != "" {
		fmt.Fprintln(df.output, df.title)
	}

	if df.description != "" {
		fmt.Fprintln(df.output, df.description)
	}

	for groupName, group := range df.configGroups {

		if group.usage != "" {
			fmt.Fprintln(w, group.usage)
		} else {
			fmt.Fprintln(w, strings.ToUpper(groupName))
		}

		fmt.Fprintln(w, "  Flag\tUsage")
		for flagName, flag := range group.Flags {
			usage := flag.Usage
			if flag.Default != nil && flag.Default != "" {
				usage = fmt.Sprintf("%s (default: %v)", flag.Usage, flag.Default)
			}
			fmt.Fprintf(w, "  --%s.<IDENTIFIER>.%s %s\t%s\n", groupName, flagName, flag.Type, usage)
		}
		fmt.Fprintln(w, "")
	}

	if df.epilog != "" {
		fmt.Fprintln(df.output, df.epilog)
	}
}
