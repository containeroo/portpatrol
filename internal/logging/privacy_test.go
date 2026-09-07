package logging

import "testing"

func TestURLCredentialsAlwaysHidden(t *testing.T) {
	for _, show := range []bool{false, true} {
		got := RedactURLs("Get \"https://user:password@example.com/private?q=secret#fragment\": failed", show)
		want := "Get \"https://example.com\": failed"
		if show {
			want = "Get \"https://example.com/private?q=secret#fragment\": failed"
		}
		if got != want {
			t.Fatalf("got %q want %q", got, want)
		}
	}
}
