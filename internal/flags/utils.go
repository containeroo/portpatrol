package flags

import (
	"bytes"
	"fmt"
	"text/tabwriter"
)

type FlagDoc struct {
	Flag        string
	Description string
}

func displayCheckerProperties(buf *bytes.Buffer, docs map[string][]FlagDoc) {
	w := tabwriter.NewWriter(buf, 0, 0, 2, ' ', 0)

	appendFlagDocs := func(title string, docs []FlagDoc) {
		fmt.Fprintf(w, "\n%s:\n", title)
		fmt.Fprintln(w, "  Flag\tDescription")
		for _, doc := range docs {
			fmt.Fprintf(w, "  %s\t%s\n", doc.Flag, doc.Description)
		}
	}

	for title, doc := range docs {
		appendFlagDocs(title, doc)
	}

	w.Flush()
}
