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

var tcpFlagDocs = []FlagDoc{
	{fmt.Sprintf("--%s.<identifier>.%s=string", paramPrefix, paramAddress), "The IP address or hostname of the target in the following format: tcp://hostname:port"},
	{fmt.Sprintf("--%s.<identifier>.%s=string", paramPrefix, paramName), "The name of the target. If not specified, it's derived from the target address."},
	{fmt.Sprintf("--%s.<identifier>.%s=string", paramPrefix, paramType), "The type of check to perform. If the scheme (tcp://) is specified in --%s.<identifier>.address, this flag can be omitted as the type will be inferred."},
	{fmt.Sprintf("--%s.<identifier>.%s=duration", paramPrefix, paramInterval), "Override the default interval for this target (e.g., 5s)."},
}

var httpFlagDocs = []FlagDoc{
	{fmt.Sprintf("--%s.<identifier>.%s=string", paramPrefix, paramAddress), "The IP address or hostname of the target in the following format: scheme://hostname[:port]"},
	{fmt.Sprintf("--%s.<identifier>.%s=string", paramPrefix, paramName), "The name of the target. If not specified, it's derived from the target address."},
	{fmt.Sprintf("--%s.<identifier>.%s=string", paramPrefix, paramType), "The type of check to perform. If a scheme (e.g. http://) is specified in --%s.<identifier>.address, this flag can be omitted as the type will be inferred."},
	{fmt.Sprintf("--%s.<identifier>.%s=string", paramPrefix, paramHTTPMethod), "The HTTP method to use (e.g., GET, POST). Defaults to \"GET\"."},
	{fmt.Sprintf("--%s.<identifier>.%s=string", paramPrefix, paramHTTPHeaders), "A comma-separated list of HTTP headers to include in the request in \"key=value\" format.\n\tExample: Authorization=Bearer token,Content-Type=application/json"},
	{fmt.Sprintf("--%s.<identifier>%s=string", paramPrefix, paramHTTPExpectedStatusCodes), "A comma-separated list of expected HTTP status codes or ranges. Defaults to 200.\n\tExample: \"200,301,404\" or \"200,300-302\" or \"200,301-302,404,500-502\""},
	{fmt.Sprintf("--%s.<identifier>.%s=bool", paramPrefix, paramHTTPSkipTLSVerify), "Whether to skip TLS verification. Defaults to false."},
	{fmt.Sprintf("--%s.<identifier>.%s=duration", paramPrefix, paramHTTPTimeout), "The timeout for the HTTP request (e.g., 5s)."},
	{fmt.Sprintf("--%s.<identifier>.%s=duration", paramPrefix, paramInterval), "Override the default interval for this target (e.g., 10s)."},
}

var icmpFlagDocs = []FlagDoc{
	{fmt.Sprintf("--%s.<identifier>.%s=string", paramPrefix, paramAddress), "The IP address or hostname of the target in the following format: icmp://hostname (no port allowed)."},
	{fmt.Sprintf("--%s.<identifier>.%s=string", paramPrefix, paramName), "The name of the target. If not specified, it's derived from the target address."},
	{fmt.Sprintf("--%s.<identifier>.%s=string", paramPrefix, paramType), "The type of check to perform. If the scheme (icmp://) is specified in --%s.<identifier>.address, this flag can be omitted as the type will be inferred."},
	{fmt.Sprintf("--%s.<identifier>.%s=duration", paramPrefix, paramICMPReadTimeout), "The read timeout for the ICMP connection (e.g., 1s)."},
	{fmt.Sprintf("--%s.<identifier>.%s=duration", paramPrefix, paramICMPWriteTimeout), "The write timeout for the ICMP connection (e.g., 1s)."},
	{fmt.Sprintf("--%s.<identifier>.%s=duration", paramPrefix, paramInterval), "Override the default interval for this target (e.g., 5s)."},
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
