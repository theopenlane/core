package httpsling

import (
	"bufio"
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"github.com/theopenlane/utils/rout"
)

// Response represents an HTTP response
type Response struct {
	// stream is the callback function for streaming responses
	stream StreamCallback
	// streamErr is the callback function for streaming errors
	streamErr StreamErrCallback
	// streamDone is the callback function for when the stream is done
	streamDone StreamDoneCallback
	// RawResponse is the original HTTP response
	RawResponse *http.Response
	// BodyBytes is the response body as a juicy byte slice
	BodyBytes []byte
	// Context is the request context
	Context context.Context
	// Client is the HTTP client
	Client *Client
}

// NewResponse creates a new wrapped response object leveraging the buffer pool
func NewResponse(ctx context.Context, resp *http.Response, client *Client, stream StreamCallback, streamErr StreamErrCallback, streamDone StreamDoneCallback) (*Response, error) {
	response := &Response{
		RawResponse: resp,
		Context:     ctx,
		BodyBytes:   nil,
		stream:      stream,
		streamErr:   streamErr,
		streamDone:  streamDone,
		Client:      client,
	}

	if response.stream != nil {
		go response.handleStream()
	} else if err := response.handleNonStream(); err != nil {
		return nil, err
	}

	return response, nil
}

var maxStreamBufferSize = 512 * 1024

// handleStream processes the HTTP response as a stream
func (r *Response) handleStream() {
	defer func() {
		if err := r.RawResponse.Body.Close(); err != nil {
			r.Client.Logger.Errorf("failed to close response body: %v", err)
		}
	}()

	scanner := bufio.NewScanner(r.RawResponse.Body)

	scanBuf := make([]byte, 0, maxStreamBufferSize)

	scanner.Buffer(scanBuf, maxStreamBufferSize)

	for scanner.Scan() {
		if err := r.stream(scanner.Bytes()); err != nil {
			break
		}
	}

	if err := scanner.Err(); err != nil && r.streamErr != nil {
		r.streamErr(err)
	}

	if r.streamDone != nil {
		r.streamDone()
	}
}

// handleNonStream reads the HTTP response body into a buffer for non-streaming responses
func (r *Response) handleNonStream() error {
	buf := GetBuffer()
	defer PutBuffer(buf)

	_, err := buf.ReadFrom(r.RawResponse.Body)
	if err != nil {
		return fmt.Errorf("%w: %v", ErrResponseReadFailed, err)
	}

	_ = r.RawResponse.Body.Close()

	r.RawResponse.Body = io.NopCloser(bytes.NewReader(buf.B))
	r.BodyBytes = buf.B

	return nil
}

// StatusCode returns the HTTP status code of the response
func (r *Response) StatusCode() int {
	return r.RawResponse.StatusCode
}

// Status returns the status string of the response
func (r *Response) Status() string {
	return r.RawResponse.Status
}

// Header returns the response headers
func (r *Response) Header() http.Header {
	return r.RawResponse.Header
}

// Cookies parses and returns the cookies set in the response
func (r *Response) Cookies() []*http.Cookie {
	return r.RawResponse.Cookies()
}

// Location returns the URL redirected address
func (r *Response) Location() (*url.URL, error) {
	return r.RawResponse.Location()
}

// URL returns the request URL that elicited the response
func (r *Response) URL() *url.URL {
	return r.RawResponse.Request.URL
}

// ContentType returns the value of the HeaderContentType header
func (r *Response) ContentType() string {
	return r.Header().Get(HeaderContentType)
}

// IsContentType Checks if the response Content-Type header matches a given content type
func (r *Response) IsContentType(contentType string) bool {
	return strings.Contains(r.ContentType(), contentType)
}

// IsJSON checks if the response Content-Type indicates JSON
func (r *Response) IsJSON() bool {
	return r.IsContentType(ContentTypeJSON)
}

// IsXML checks if the response Content-Type indicates XML
func (r *Response) IsXML() bool {
	return r.IsContentType(ContentTypeXML)
}

// IsYAML checks if the response Content-Type indicates YAML
func (r *Response) IsYAML() bool {
	return r.IsContentType(ContentTypeYAML)
}

// ContentLength returns the length of the response body
func (r *Response) ContentLength() int {
	if r.BodyBytes == nil {
		return 0
	}

	return len(r.BodyBytes)
}

// IsEmpty checks if the response body is empty
func (r *Response) IsEmpty() bool {
	return r.ContentLength() == 0
}

// IsSuccess checks if the response status code indicates success
func (r *Response) IsSuccess() bool {
	code := r.StatusCode()

	return code >= http.StatusOK && code <= http.StatusIMUsed
}

// Body returns the response body as a juicy byte slice
func (r *Response) Body() []byte {
	return r.BodyBytes
}

// String returns the response body as a string
func (r *Response) String() string {
	return string(r.BodyBytes)
}

// Scan attempts to unmarshal the response body based on its content type
func (r *Response) Scan(v interface{}) error {
	switch {
	case r.IsJSON():
		return r.ScanJSON(v)
	case r.IsXML():
		return r.ScanXML(v)
	case r.IsYAML():
		return r.ScanYAML(v)
	}

	return fmt.Errorf("%w: %s", ErrUnsupportedContentType, r.ContentType())
}

// ScanJSON unmarshals the response body into a struct via JSON decoding
func (r *Response) ScanJSON(v interface{}) error {
	if r.BodyBytes == nil {
		return nil
	}

	return r.Client.JSONDecoder.Decode(bytes.NewReader(r.BodyBytes), v)
}

// ScanXML unmarshals the response body into a struct via XML decoding
func (r *Response) ScanXML(v interface{}) error {
	if r.BodyBytes == nil {
		return nil
	}

	return r.Client.XMLDecoder.Decode(bytes.NewReader(r.BodyBytes), v)
}

// ScanYAML unmarshals the response body into a struct via YAML decoding
func (r *Response) ScanYAML(v interface{}) error {
	if r.BodyBytes == nil {
		return nil
	}

	return r.Client.YAMLDecoder.Decode(bytes.NewReader(r.BodyBytes), v)
}

const dirPermissions = 0755

// Save saves the response body to a file or io.Writer
func (r *Response) Save(v any) error {
	switch p := v.(type) {
	case string:
		file := filepath.Clean(p)
		dir := filepath.Dir(file)

		// Create the directory if it doesn't exist
		if _, err := os.Stat(dir); err != nil {
			if !errors.Is(err, os.ErrNotExist) {
				return rout.HTTPErrorResponse(err)
			}

			if err = os.MkdirAll(dir, dirPermissions); err != nil {
				return rout.HTTPErrorResponse(err)
			}
		}

		// Create and open the file for writing
		outFile, err := os.Create(file)
		if err != nil {
			return rout.HTTPErrorResponse(err)
		}

		defer func() {
			if err := outFile.Close(); err != nil {
				r.Client.Logger.Errorf("failed to close file: %v", err)
			}
		}()

		// Write the response body to the file
		_, err = io.Copy(outFile, bytes.NewReader(r.Body()))
		if err != nil {
			return rout.HTTPErrorResponse(err)
		}

		return nil
	case io.Writer:
		// Write the response body directly to the provided io.Writer
		_, err := io.Copy(p, bytes.NewReader(r.Body()))
		if err != nil {
			return rout.HTTPErrorResponse(err)
		}

		if pc, ok := p.(io.WriteCloser); ok {
			if err := pc.Close(); err != nil {
				r.Client.Logger.Errorf("failed to close io.Writer: %v", err)
			}
		}

		return nil
	default:
		return ErrNotSupportSaveMethod
	}
}

// Close closes the response body
func (r *Response) Close() error {
	return r.RawResponse.Body.Close()
}
