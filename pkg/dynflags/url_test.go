package dynflags_test

import (
	"net/url"
	"testing"

	"github.com/containeroo/portpatrol/pkg/dynflags"
	"github.com/stretchr/testify/assert"
)

func TestURLValue(t *testing.T) {
	t.Parallel()

	t.Run("Parse valid URL", func(t *testing.T) {
		t.Parallel()

		urlValue := dynflags.URLValue{}
		parsed, err := urlValue.Parse("https://example.com")
		assert.NoError(t, err)
		assert.NotNil(t, parsed)

		parsedURL, ok := parsed.(*url.URL)
		assert.True(t, ok)
		assert.Equal(t, "https://example.com", parsedURL.String())
	})

	t.Run("Parse invalid URL", func(t *testing.T) {
		t.Parallel()

		urlValue := dynflags.URLValue{}
		parsed, err := urlValue.Parse("https://invalid-url^")
		assert.Error(t, err)
		assert.Nil(t, parsed)
	})

	t.Run("Set valid URL", func(t *testing.T) {
		t.Parallel()

		bound := &url.URL{}
		urlValue := dynflags.URLValue{Bound: bound}

		parsedURL, _ := url.Parse("https://example.com")
		err := urlValue.Set(parsedURL)
		assert.NoError(t, err)
		assert.Equal(t, "https://example.com", bound.String())
	})

	t.Run("Set invalid type", func(t *testing.T) {
		t.Parallel()

		bound := &url.URL{}
		urlValue := dynflags.URLValue{Bound: bound}

		err := urlValue.Set("not-a-url") // Invalid type
		assert.Error(t, err)
		assert.EqualError(t, err, "invalid value type: expected URL")
	})
}

func TestGroupConfigURL(t *testing.T) {
	t.Parallel()

	t.Run("Define URL flag with default value", func(t *testing.T) {
		t.Parallel()

		group := &dynflags.ConfigGroup{Flags: make(map[string]*dynflags.Flag)}
		defaultValue := "https://default.com"
		urlFlag := group.URL("urlFlag", defaultValue, "A URL flag")

		assert.Equal(t, "https://default.com", urlFlag.Default)
		assert.Contains(t, group.Flags, "urlFlag")
		assert.Equal(t, "A URL flag", group.Flags["urlFlag"].Usage)
		assert.Equal(t, defaultValue, group.Flags["urlFlag"].Default)
	})

	t.Run("Define URL flag with invalid default", func(t *testing.T) {
		t.Parallel()

		group := &dynflags.ConfigGroup{Flags: make(map[string]*dynflags.Flag)}

		assert.PanicsWithValue(t,
			"invalid default URL for flag 'urlFlag': parse \"http://i nvalid-url\": invalid character \" \" in host name",
			func() {
				group.URL("urlFlag", "http://i nvalid-url", "Invalid URL flag")
			})
	})
}

func TestParsedGroupGetURL(t *testing.T) {
	t.Parallel()

	t.Run("Get existing URL flag", func(t *testing.T) {
		t.Parallel()

		parsedURL, _ := url.Parse("https://example.com")
		group := &dynflags.ParsedGroup{
			Values: map[string]interface{}{
				"urlFlag": *parsedURL,
			},
		}
		retrievedURL, err := group.GetURL("urlFlag")
		assert.NoError(t, err)
		assert.NotNil(t, retrievedURL)
		assert.Equal(t, "https://example.com", retrievedURL.String())
	})

	t.Run("Get non-existent URL flag", func(t *testing.T) {
		t.Parallel()

		group := &dynflags.ParsedGroup{
			Values: map[string]interface{}{},
		}
		retrievedURL, err := group.GetURL("urlFlag")
		assert.Error(t, err)
		assert.Nil(t, retrievedURL)
		assert.EqualError(t, err, "flag 'urlFlag' not found in group ''")
	})

	t.Run("Get URL flag with invalid type", func(t *testing.T) {
		t.Parallel()

		group := &dynflags.ParsedGroup{
			Values: map[string]interface{}{
				"urlFlag": "not-a-url", // Invalid type
			},
		}
		retrievedURL, err := group.GetURL("urlFlag")
		assert.Error(t, err)
		assert.Nil(t, retrievedURL)
		assert.EqualError(t, err, "flag 'urlFlag' is not a URL")
	})
}

func TestURLGetBound(t *testing.T) {
	t.Run("URLValue - GetBound", func(t *testing.T) {
		var u *url.URL
		val, _ := url.Parse("http://example.com")
		u = val

		urlValue := dynflags.URLValue{Bound: u}
		assert.Equal(t, *val, urlValue.GetBound())

		urlValue = dynflags.URLValue{Bound: nil}
		assert.Nil(t, urlValue.GetBound())
	})
}
