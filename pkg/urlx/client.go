package urlx

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"time"

	"github.com/theopenlane/httpsling"
	"github.com/theopenlane/httpsling/httpclient"
)

const (
	// defaultHTTPTimeout is the org default timeout applied to outbound HTTP clients
	defaultHTTPTimeout = 30 * time.Second
	// readChunkSize is the read granularity used when validating a response body as it arrives
	readChunkSize = 32 * 1024
)

// NewHTTPClient returns an *http.Client built on the httpclient default transport
// with the org default timeout, overridable via opts
func NewHTTPClient(opts ...httpclient.Option) (*http.Client, error) {
	return httpclient.New(append([]httpclient.Option{httpclient.Timeout(defaultHTTPTimeout)}, opts...)...)
}

// NewRequester returns an httpsling.Requester backed by NewHTTPClient so callers
// never fall back to http.DefaultClient, with opts applied after the defaults
func NewRequester(opts ...httpsling.Option) (*httpsling.Requester, error) {
	client, err := NewHTTPClient()
	if err != nil {
		return nil, err
	}

	return httpsling.New(append([]httpsling.Option{httpsling.WithDoer(client)}, opts...)...)
}

// MaxSizeValidator returns a validator that rejects payloads whose size exceeds maxBytes
func MaxSizeValidator(maxBytes int64) httpsling.ValidationFunc {
	return func(f httpsling.File) error {
		if f.Size > maxBytes {
			return fmt.Errorf("%w: %d bytes exceeds limit of %d", ErrSizeLimitExceeded, f.Size, maxBytes)
		}

		return nil
	}
}

// ReadBody reads and closes resp.Body, invoking vf with the response content type
// and the accumulated size as data arrives so a failing validator aborts the read
// mid-stream; the advertised Content-Length is validated before anything is read
func ReadBody(resp *http.Response, vf httpsling.ValidationFunc) ([]byte, error) {
	if resp == nil || resp.Body == nil || resp.Body == http.NoBody {
		return nil, nil
	}

	defer resp.Body.Close()

	contentType := resp.Header.Get(httpsling.HeaderContentType)

	var advertised int64
	if header := resp.Header.Get(httpsling.HeaderContentLength); header != "" {
		advertised, _ = strconv.ParseInt(header, 10, 64)
	}

	if vf != nil {
		if err := vf(httpsling.File{MimeType: contentType, Size: advertised}); err != nil {
			return nil, err
		}
	}

	var buf bytes.Buffer
	if advertised > 0 {
		buf.Grow(int(advertised))
	}

	if vf == nil {
		if _, err := buf.ReadFrom(resp.Body); err != nil {
			return nil, err
		}

		return buf.Bytes(), nil
	}

	chunk := make([]byte, readChunkSize)

	for {
		n, err := resp.Body.Read(chunk)
		if n > 0 {
			buf.Write(chunk[:n])

			if verr := vf(httpsling.File{MimeType: contentType, Size: int64(buf.Len())}); verr != nil {
				return nil, verr
			}
		}

		switch {
		case errors.Is(err, io.EOF):
			return buf.Bytes(), nil
		case err != nil:
			return nil, err
		}
	}
}
