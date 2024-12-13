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

		group := &dynflags.GroupConfig{Flags: make(map[string]*dynflags.Flag)}
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

		group := &dynflags.GroupConfig{Flags: make(map[string]*dynflags.Flag)}
		var urlSlice []*url.URL
		group.URLSlicesVar(&urlSlice, "urlSliceFlag", []*url.URL{{Scheme: "https", Host: "google.com"}}, "URL slices flag variable")
		assert.Equal(t, []*url.URL{{Scheme: "https", Host: "google.com"}}, urlSlice)
	})
}

func TestParsedGroupGetURLSlices(t *testing.T) {
	t.Parallel()

	t.Run("Get existing URL slices flag", func(t *testing.T) {
		t.Parallel()

		group := &dynflags.ParsedGroup{
			Values: map[string]interface{}{
				"urlSliceFlag": []*url.URL{
					{Scheme: "https", Host: "example.com"},
					{Scheme: "http", Host: "localhost"},
				},
			},
		}
		slice, err := group.GetURLSlices("urlSliceFlag")
		assert.NoError(t, err)
		assert.Equal(t, []*url.URL{
			{Scheme: "https", Host: "example.com"},
			{Scheme: "http", Host: "localhost"},
		}, slice)
	})

	t.Run("Get non-existent URL slices flag", func(t *testing.T) {
		t.Parallel()

		group := &dynflags.ParsedGroup{
			Values: map[string]interface{}{},
		}
		slice, err := group.GetURLSlices("urlSliceFlag")
		assert.Error(t, err)
		assert.Nil(t, slice)
		assert.EqualError(t, err, "flag 'urlSliceFlag' not found in group ''")
	})

	t.Run("Get URL slices flag with invalid type", func(t *testing.T) {
		t.Parallel()

		group := &dynflags.ParsedGroup{
			Values: map[string]interface{}{
				"urlSliceFlag": "invalid-type", // Invalid type
			},
		}
		slice, err := group.GetURLSlices("urlSliceFlag")
		assert.Error(t, err)
		assert.Nil(t, slice)
		assert.EqualError(t, err, "flag 'urlSliceFlag' is not a []*url.URL")
	})
}
