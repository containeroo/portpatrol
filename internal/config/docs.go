package config

import (
	"bytes"
	"fmt"
	"text/tabwriter"
)

type FlagDoc struct {
	Flag        string
	Description string
}

var tcpFlagDocs = []FlagDoc{
	{
		Flag:        fmt.Sprintf("--%s.<identifier>.%s=string", ParamPrefix, ParamAddress),
		Description: "The IP address or hostname of the target in the following format: tcp://hostname:port",
	},
	{
		Flag:        fmt.Sprintf("--%s.<identifier>.%s=string", ParamPrefix, ParamName),
		Description: "The name of the target. If not specified, it's derived from the target address.",
	},
	{
		Flag:        fmt.Sprintf("--%s.<identifier>.%s=string", ParamPrefix, ParamType),
		Description: "The type of check to perform. If the scheme (tcp://) is specified in --%s.<identifier>.address, this flag can be omitted as the type will be inferred.",
	},
	{
		Flag:        fmt.Sprintf("--%s.<identifier>.%s=duration", ParamPrefix, ParamInterval),
		Description: "Override the default interval for this target (e.g., 5s).",
	},
	{
		Flag:        fmt.Sprintf("--%s.<identifier>.%s=duration", ParamPrefix, ParamTCPTimeout),
		Description: "The timeout for the TCP request (e.g., 5s).",
	},
}

var httpFlagDocs = []FlagDoc{
	{
		Flag:        fmt.Sprintf("--%s.<identifier>.%s=string", ParamPrefix, ParamAddress),
		Description: "The IP address or hostname of the target in the following format: scheme://hostname[:port]",
	},
	{
		Flag:        fmt.Sprintf("--%s.<identifier>.%s=string", ParamPrefix, ParamName),
		Description: "The name of the target. If not specified, it's derived from the target address.",
	},
	{
		Flag:        fmt.Sprintf("--%s.<identifier>.%s=string", ParamPrefix, ParamType),
		Description: "The type of check to perform. If a scheme (e.g. http://) is specified in --%s.<identifier>.address, this flag can be omitted as the type will be inferred.",
	},
	{
		Flag:        fmt.Sprintf("--%s.<identifier>.%s=string", ParamPrefix, ParamHTTPMethod),
		Description: "The HTTP method to use (e.g., GET, POST). Defaults to \"GET\".",
	},
	{
		Flag:        fmt.Sprintf("--%s.<identifier>.%s=string", ParamPrefix, ParamHTTPHeaders),
		Description: "A comma-separated list of HTTP headers to include in the request in \"key=value\" format.\n\tExample: Authorization=Bearer token,Content-Type=application/json",
	},
	{
		Flag:        fmt.Sprintf("--%s.<identifier>%s=string", ParamPrefix, ParamHTTPExpectedStatusCodes),
		Description: "A comma-separated list of expected HTTP status codes or ranges. Defaults to 200.\n\tExample: \"200,301,404\" or \"200,300-302\" or \"200,301-302,404,500-502\"",
	},
	{
		Flag:        fmt.Sprintf("--%s.<identifier>.%s=bool", ParamPrefix, ParamHTTPSkipTLSVerify),
		Description: "Whether to skip TLS verification. Defaults to false.",
	},
	{
		Flag:        fmt.Sprintf("--%s.<identifier>.%s=duration", ParamPrefix, ParamHTTPTimeout),
		Description: "The timeout for the HTTP request (e.g., 5s).",
	},
	{
		Flag:        fmt.Sprintf("--%s.<identifier>.%s=duration", ParamPrefix, ParamInterval),
		Description: "Override the default interval for this target (e.g., 10s).",
	},
}

var icmpFlagDocs = []FlagDoc{
	{
		Flag:        fmt.Sprintf("--%s.<identifier>.%s=string", ParamPrefix, ParamAddress),
		Description: "The IP address or hostname of the target in the following format: icmp://hostname (no port allowed).",
	},
	{
		Flag:        fmt.Sprintf("--%s.<identifier>.%s=string", ParamPrefix, ParamName),
		Description: "The name of the target. If not specified, it's derived from the target address.",
	},
	{
		Flag:        fmt.Sprintf("--%s.<identifier>.%s=string", ParamPrefix, ParamType),
		Description: "The type of check to perform. If the scheme (icmp://) is specified in --%s.<identifier>.address, this flag can be omitted as the type will be inferred.",
	},
	{
		Flag:        fmt.Sprintf("--%s.<identifier>.%s=duration", ParamPrefix, ParamICMPReadTimeout),
		Description: "The read timeout for the ICMP connection (e.g., 1s).",
	},
	{
		Flag:        fmt.Sprintf("--%s.<identifier>.%s=duration", ParamPrefix, ParamICMPWriteTimeout),
		Description: "The write timeout for the ICMP connection (e.g., 1s).",
	},
	{
		Flag:        fmt.Sprintf("--%s.<identifier>.%s=duration", ParamPrefix, ParamInterval),
		Description: "Override the default interval for this target (e.g., 5s).",
	},
}

func displayCheckerProperties(buf *bytes.Buffer) {
	w := tabwriter.NewWriter(buf, 0, 0, 2, ' ', 0)

	appendFlagDocs := func(title string, docs []FlagDoc) {
		fmt.Fprintf(w, "\n%s:\n", title)
		fmt.Fprintln(w, "  Flag\tDescription")
		for _, doc := range docs {
			fmt.Fprintf(w, "  %s\t%s\n", doc.Flag, doc.Description)
		}
	}

	appendFlagDocs("TCP Checker Properties", tcpFlagDocs)
	appendFlagDocs("HTTP Checker Properties", httpFlagDocs)
	appendFlagDocs("ICMP Checker Properties", icmpFlagDocs)

	w.Flush()
}
