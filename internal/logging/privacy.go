package logging

import (
	"net/url"
	"regexp"
)

var httpURL = regexp.MustCompile(`(?i)https?://[^\s"'<>]+`)

// RedactURLs hides URL credentials, and hides paths, queries and fragments unless opted in.
// Apply before log encoding so both JSON and text errors receive identical treatment.
func RedactURLs(text string, showPath bool) string {
	return httpURL.ReplaceAllStringFunc(text, func(raw string) string {
		u, err := url.Parse(raw)
		if err != nil || u.Host == "" {
			return "[redacted URL]"
		}
		u.User = nil
		if !showPath {
			u.Path = ""
			u.RawPath = ""
			u.RawQuery = ""
			u.ForceQuery = false
			u.Fragment = ""
			u.RawFragment = ""
		}
		return u.String()
	})
}
