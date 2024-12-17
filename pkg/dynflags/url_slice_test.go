package dynflags_test

import (
	"net/url"
	"testing"

	"github.com/containeroo/portpatrol/pkg/dynflags"
	"github.com/stretchr/testify/assert"
)

func TestURLSlicesValue(t *testing.T) {
	t.Parallel()

	t.Run("Parse valid URL", func(t *testing.T) {
		t.Parallel()

		urlSlicesValue := dynflags.URLSlicesValue{Bound: &[]*url.URL{}}
		parsed, err := urlSlicesValue.Parse("https://example.com")
		assert.NoError(t, err)
		assert.Equal(t, "https://example.com", parsed.(*url.URL).String())
	})

	t.Run("Parse invalid URL", func(t *testing.T) {
		t.Parallel()

		urlSlicesValue := dynflags.URLSlicesValue{Bound: &[]*url.URL{}}
		parsed, err := urlSlicesValue.Parse("://invalid-url")
		assert.Error(t, err)
		assert.Nil(t, parsed)
	})

	t.Run("Set valid URL", func(t *testing.T) {
		t.Parallel()

		bound := []*url.URL{{Scheme: "https", Host: "example.com"}}
		urlSlicesValue := dynflags.URLSlicesValue{Bound: &bound}

		err := urlSlicesValue.Set(&url.URL{Scheme: "http", Host: "localhost"})
		assert.NoError(t, err)
		assert.Equal(t, []*url.URL{
			{Scheme: "https", Host: "example.com"},
			{Scheme: "http", Host: "localhost"},
		}, bound)
	})

	t.Run("Set invalid type", func(t *testing.T) {
		t.Parallel()

		bound := []*url.URL{}
		urlSlicesValue := dynflags.URLSlicesValue{Bound: &bound}

		err := urlSlicesValue.Set("invalid-type")
		assert.Error(t, err)
		assert.EqualError(t, err, "invalid value type: expected *url.URL")
	})
}

func TestGroupConfigURLSlices(t *testing.T) {
	t.Parallel()

	t.Run("Define URL slices flag", func(t *testing.T) {
		t.Parallel()

		group := &dynflags.ConfigGroup{Flags: make(map[string]*dynflags.Flag)}
		defaultValue := []*url.URL{
			{Scheme: "https", Host: "example.com"},
			{Scheme: "http", Host: "localhost"},
		}
		urlSlice := group.URLSlices("urlSliceFlag", defaultValue, "A URL slices flag")

		assert.Equal(t, []*url.URL{
			{Scheme: "https", Host: "example.com"},
			{Scheme: "http", Host: "localhost"},
		}, *urlSlice)
		assert.Contains(t, group.Flags, "urlSliceFlag")
		assert.Equal(t, "A URL slices flag", group.Flags["urlSliceFlag"].Usage)
		assert.Equal(t, "https://example.com,http://localhost", group.Flags["urlSliceFlag"].Default)
	})

	t.Run("Define URLSlicesVar and set value", func(t *testing.T) {
		t.Parallel()

		group := &dynflags.ConfigGroup{Flags: make(map[string]*dynflags.Flag)}
		var urlSlice []*url.URL
		group.URLSlicesVar(&urlSlice, "urlSliceFlag", []*url.URL{{Scheme: "https", Host: "google.com"}}, "URL slices flag variable")
		assert.Equal(t, []*url.URL{{Scheme: "https", Host: "google.com"}}, urlSlice)
	})
}

func TestGetURLSlices(t *testing.T) {
	t.Parallel()

	t.Run("Retrieve []*url.URL value", func(t *testing.T) {
		t.Parallel()

		parsedURL1, _ := url.Parse("https://example.com")
		parsedURL2, _ := url.Parse("https://example.org")

		parsedGroup := &dynflags.ParsedGroup{
			Name:   "testGroup",
			Values: map[string]interface{}{"flag1": []*url.URL{parsedURL1, parsedURL2}},
		}

		result, err := parsedGroup.GetURLSlices("flag1")
		assert.NoError(t, err)
		assert.Equal(t, []*url.URL{parsedURL1, parsedURL2}, result)
	})

	t.Run("Retrieve single *url.URL value as []*url.URL", func(t *testing.T) {
		t.Parallel()

		parsedURL, _ := url.Parse("https://example.com")

		parsedGroup := &dynflags.ParsedGroup{
			Name:   "testGroup",
			Values: map[string]interface{}{"flag1": parsedURL},
		}

		result, err := parsedGroup.GetURLSlices("flag1")
		assert.NoError(t, err)
		assert.Equal(t, []*url.URL{parsedURL}, result)
	})

	t.Run("Flag not found", func(t *testing.T) {
		t.Parallel()

		parsedGroup := &dynflags.ParsedGroup{
			Name:   "testGroup",
			Values: map[string]interface{}{},
		}

		result, err := parsedGroup.GetURLSlices("nonExistentFlag")
		assert.Error(t, err)
		assert.Nil(t, result)
		assert.EqualError(t, err, "flag 'nonExistentFlag' not found in group 'testGroup'")
	})

	t.Run("Flag value is invalid type", func(t *testing.T) {
		t.Parallel()

		parsedGroup := &dynflags.ParsedGroup{
			Name:   "testGroup",
			Values: map[string]interface{}{"flag1": 123}, // Invalid type (int)
		}

		result, err := parsedGroup.GetURLSlices("flag1")
		assert.Error(t, err)
		assert.Nil(t, result)
		assert.EqualError(t, err, "flag 'flag1' is not a []*url.URL")
	})
}
