package urlx

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/theopenlane/httpsling"
	"github.com/theopenlane/httpsling/httpclient"
)

func TestNewHTTPClient(t *testing.T) {
	t.Parallel()

	client, err := NewHTTPClient()
	require.NoError(t, err)
	assert.Equal(t, defaultHTTPTimeout, client.Timeout)

	client, err = NewHTTPClient(httpclient.Timeout(defaultHTTPTimeout * 2))
	require.NoError(t, err)
	assert.Equal(t, defaultHTTPTimeout*2, client.Timeout)
}

func TestNewRequester(t *testing.T) {
	t.Parallel()

	requester, err := NewRequester()
	require.NoError(t, err)
	require.NotNil(t, requester.HTTPClient())
	assert.Equal(t, defaultHTTPTimeout, requester.HTTPClient().Timeout)
}

func TestMaxSizeValidator(t *testing.T) {
	t.Parallel()

	vf := MaxSizeValidator(10)

	assert.NoError(t, vf(httpsling.File{Size: 10}))
	assert.ErrorIs(t, vf(httpsling.File{Size: 11}), ErrSizeLimitExceeded)
}

func TestReadBody(t *testing.T) {
	t.Parallel()

	t.Run("nil validator reads full body", func(t *testing.T) {
		t.Parallel()

		resp := &http.Response{Body: http.NoBody}
		body, err := ReadBody(resp, nil)
		require.NoError(t, err)
		assert.Empty(t, body)
	})

	t.Run("body within limit", func(t *testing.T) {
		t.Parallel()

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			fmt.Fprint(w, "hello")
		}))
		defer server.Close()

		resp, err := http.Get(server.URL)
		require.NoError(t, err)

		body, err := ReadBody(resp, MaxSizeValidator(5))
		require.NoError(t, err)
		assert.Equal(t, "hello", string(body))
	})

	t.Run("advertised content length rejected before read", func(t *testing.T) {
		t.Parallel()

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.Header().Set(httpsling.HeaderContentLength, "100")
			fmt.Fprint(w, strings.Repeat("a", 100))
		}))
		defer server.Close()

		resp, err := http.Get(server.URL)
		require.NoError(t, err)

		_, err = ReadBody(resp, MaxSizeValidator(10))
		assert.ErrorIs(t, err, ErrSizeLimitExceeded)
	})

	t.Run("chunked body aborted mid stream", func(t *testing.T) {
		t.Parallel()

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			flusher := w.(http.Flusher)
			for range 8 {
				fmt.Fprint(w, strings.Repeat("a", readChunkSize))
				flusher.Flush()
			}
		}))
		defer server.Close()

		resp, err := http.Get(server.URL)
		require.NoError(t, err)

		_, err = ReadBody(resp, MaxSizeValidator(readChunkSize*2))
		assert.ErrorIs(t, err, ErrSizeLimitExceeded)
	})

	t.Run("content type passed to validator", func(t *testing.T) {
		t.Parallel()

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.Header().Set(httpsling.HeaderContentType, httpsling.ContentTypeJSON)
			fmt.Fprint(w, `{}`)
		}))
		defer server.Close()

		resp, err := http.Get(server.URL)
		require.NoError(t, err)

		var seen string

		_, err = ReadBody(resp, func(f httpsling.File) error {
			seen = f.MimeType
			return nil
		})
		require.NoError(t, err)
		assert.Equal(t, httpsling.ContentTypeJSON, seen)
	})
}
