package urlx

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/theopenlane/httpsling"
	"github.com/theopenlane/httpsling/httpclient"
	"gotest.tools/v3/assert"
	is "gotest.tools/v3/assert/cmp"
)

func TestNewHTTPClient(t *testing.T) {
	t.Parallel()

	client, err := NewHTTPClient()
	assert.NilError(t, err)
	assert.Check(t, is.Equal(defaultHTTPTimeout, client.Timeout))

	client, err = NewHTTPClient(httpclient.Timeout(defaultHTTPTimeout * 2))
	assert.NilError(t, err)
	assert.Check(t, is.Equal(defaultHTTPTimeout*2, client.Timeout))
}

func TestNewRequester(t *testing.T) {
	t.Parallel()

	requester, err := NewRequester()
	assert.NilError(t, err)
	assert.Assert(t, requester.HTTPClient() != nil)
	assert.Check(t, is.Equal(defaultHTTPTimeout, requester.HTTPClient().Timeout))
}

func TestMaxSizeValidator(t *testing.T) {
	t.Parallel()

	vf := MaxSizeValidator(10)

	assert.Check(t, vf(httpsling.File{Size: 10}))
	assert.Check(t, is.ErrorIs(vf(httpsling.File{Size: 11}), ErrSizeLimitExceeded))
}

func TestReadBody(t *testing.T) {
	t.Parallel()

	t.Run("nil validator reads full body", func(t *testing.T) {
		t.Parallel()

		resp := &http.Response{Body: http.NoBody}
		body, err := ReadBody(resp, nil)
		assert.NilError(t, err)
		assert.Check(t, is.Len(body, 0))
	})

	t.Run("body within limit", func(t *testing.T) {
		t.Parallel()

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			fmt.Fprint(w, "hello")
		}))
		defer server.Close()

		resp, err := http.Get(server.URL)
		assert.NilError(t, err)

		body, err := ReadBody(resp, MaxSizeValidator(5))
		assert.NilError(t, err)
		assert.Check(t, is.Equal("hello", string(body)))
	})

	t.Run("advertised content length rejected before read", func(t *testing.T) {
		t.Parallel()

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.Header().Set(httpsling.HeaderContentLength, "100")
			fmt.Fprint(w, strings.Repeat("a", 100))
		}))
		defer server.Close()

		resp, err := http.Get(server.URL)
		assert.NilError(t, err)

		_, err = ReadBody(resp, MaxSizeValidator(10))
		assert.Check(t, is.ErrorIs(err, ErrSizeLimitExceeded))
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
		assert.NilError(t, err)

		_, err = ReadBody(resp, MaxSizeValidator(readChunkSize*2))
		assert.Check(t, is.ErrorIs(err, ErrSizeLimitExceeded))
	})

	t.Run("content type passed to validator", func(t *testing.T) {
		t.Parallel()

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.Header().Set(httpsling.HeaderContentType, httpsling.ContentTypeJSON)
			fmt.Fprint(w, `{}`)
		}))
		defer server.Close()

		resp, err := http.Get(server.URL)
		assert.NilError(t, err)

		var seen string

		_, err = ReadBody(resp, func(f httpsling.File) error {
			seen = f.MimeType
			return nil
		})
		assert.NilError(t, err)
		assert.Check(t, is.Equal(httpsling.ContentTypeJSON, seen))
	})
}
