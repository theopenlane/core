package httpsling

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"
	"slices"
	"strings"
	"time"

	"github.com/google/go-querystring/query"

	"github.com/theopenlane/utils/rout"
)

// RequestBuilder facilitates building and executing HTTP requests
type RequestBuilder struct {
	// client is the HTTP client instance
	client *Client
	// method is the HTTP method for the request
	method string
	// path is the URL path for the request
	path string
	// headers contains the request headers
	headers *http.Header
	// cookies contains the request cookies
	cookies []*http.Cookie
	// queries contains the request query parameters
	queries url.Values
	// pathParams contains the request path parameters
	pathParams map[string]string
	// formFields contains the request form fields
	formFields url.Values
	// formFiles contains the request form files
	formFiles []*File
	// boundary is the custom boundary for multipart requests
	boundary string
	// bodyData is the request body
	bodyData interface{}
	// timeout is the request timeout
	timeout time.Duration
	// middlewares contains the request middlewares
	middlewares []Middleware
	// maxRetries is the maximum number of retry attempts
	maxRetries int
	// retryStrategy is the backoff strategy for retries
	retryStrategy BackoffStrategy
	// retryIf is the custom retry condition function
	retryIf RetryIfFunc
	// auth is the authentication method for the request
	auth AuthMethod
	// stream is the stream callback for the request
	stream StreamCallback
	// streamErr is the error callback for the request
	streamErr StreamErrCallback
	// streamDone is the done callback for the request
	streamDone StreamDoneCallback
	// BeforeRequest is a hook that can be used to modify the request object
	// before the request has been fired. This is useful for adding authentication
	// and other functionality not provided in this library
	BeforeRequest func(req *http.Request) error
}

// NewRequestBuilder creates a new RequestBuilder with default settings
func (c *Client) NewRequestBuilder(method, path string) *RequestBuilder {
	rb := &RequestBuilder{
		client:  c,
		method:  method,
		path:    path,
		queries: url.Values{},
		headers: &http.Header{},
	}

	if c.Headers != nil {
		rb.headers = c.Headers
	}

	return rb
}

// AddMiddleware adds a middleware to the request
func (b *RequestBuilder) AddMiddleware(middlewares ...Middleware) {
	if b.middlewares == nil {
		b.middlewares = []Middleware{}
	}

	b.middlewares = append(b.middlewares, middlewares...)
}

// Method sets the HTTP method for the request
func (b *RequestBuilder) Method(method string) *RequestBuilder {
	b.method = method

	return b
}

// Path sets the URL path for the request
func (b *RequestBuilder) Path(path string) *RequestBuilder {
	b.path = path

	return b
}

// PathParams sets multiple path params fields and their values at one go in the RequestBuilder instance
func (b *RequestBuilder) PathParams(params map[string]string) *RequestBuilder {
	if b.pathParams == nil {
		b.pathParams = map[string]string{}
	}

	for key, value := range params {
		b.pathParams[key] = value
	}

	return b
}

// PathParam sets a single path param field and its value in the RequestBuilder instance
func (b *RequestBuilder) PathParam(key, value string) *RequestBuilder {
	if b.pathParams == nil {
		b.pathParams = map[string]string{}
	}

	b.pathParams[key] = value

	return b
}

// DelPathParam removes one or more path params fields from the RequestBuilder instance
func (b *RequestBuilder) DelPathParam(key ...string) *RequestBuilder {
	if b.pathParams != nil {
		for _, k := range key {
			delete(b.pathParams, k)
		}
	}

	return b
}

// preparePath replaces path parameters in the URL path
func (b *RequestBuilder) preparePath() string {
	if b.pathParams == nil {
		return b.path
	}

	preparedPath := b.path

	for key, value := range b.pathParams {
		placeholder := "{" + key + "}"
		preparedPath = strings.ReplaceAll(preparedPath, placeholder, url.PathEscape(value))
	}

	return preparedPath
}

// Queries adds query parameters to the request
func (b *RequestBuilder) Queries(params url.Values) *RequestBuilder {
	for key, values := range params {
		for _, value := range values {
			b.queries.Add(key, value)
		}
	}

	return b
}

// Query adds a single query parameter to the request
func (b *RequestBuilder) Query(key, value string) *RequestBuilder {
	b.queries.Add(key, value)

	return b
}

// DelQuery removes one or more query parameters from the request
func (b *RequestBuilder) DelQuery(key ...string) *RequestBuilder {
	for _, k := range key {
		b.queries.Del(k)
	}

	return b
}

// QueriesStruct adds query parameters to the request based on a struct tagged with url tags
func (b *RequestBuilder) QueriesStruct(queryStruct interface{}) *RequestBuilder {
	values, _ := query.Values(queryStruct) // safely ignore error for simplicity

	for key, value := range values {
		for _, v := range value {
			b.queries.Add(key, v)
		}
	}

	return b
}

// Headers set headers to the request
func (b *RequestBuilder) Headers(headers http.Header) *RequestBuilder {
	for key, values := range headers {
		for _, value := range values {
			b.headers.Set(key, value)
		}
	}

	return b
}

// Header sets (or replaces) a header in the request
func (b *RequestBuilder) Header(key, value string) *RequestBuilder {
	b.headers.Set(key, value)

	return b
}

// AddHeader adds a header to the request
func (b *RequestBuilder) AddHeader(key, value string) *RequestBuilder {
	b.headers.Add(key, value)

	return b
}

// DelHeader removes one or more headers from the request
func (b *RequestBuilder) DelHeader(key ...string) *RequestBuilder {
	for _, k := range key {
		b.headers.Del(k)
	}

	return b
}

// Cookies method for map
func (b *RequestBuilder) Cookies(cookies map[string]string) *RequestBuilder {
	for key, value := range cookies {
		b.Cookie(key, value)
	}

	return b
}

// Cookie adds a cookie to the request
func (b *RequestBuilder) Cookie(key, value string) *RequestBuilder {
	if b.cookies == nil {
		b.cookies = []*http.Cookie{}
	}

	b.cookies = append(b.cookies, &http.Cookie{Name: key, Value: value})

	return b
}

// DelCookie removes one or more cookies from the request
func (b *RequestBuilder) DelCookie(key ...string) *RequestBuilder {
	if b.cookies != nil {
		for i, cookie := range b.cookies {
			if slices.Contains(key, cookie.Name) {
				b.cookies = append(b.cookies[:i], b.cookies[i+1:]...)
			}
		}
	}

	return b
}

// ContentType sets the Content-Type header for the request
func (b *RequestBuilder) ContentType(contentType string) *RequestBuilder {
	b.headers.Set(HeaderContentType, contentType)

	return b
}

// Accept sets the Accept header for the request
func (b *RequestBuilder) Accept(accept string) *RequestBuilder {
	b.headers.Set(HeaderAccept, accept)

	return b
}

// UserAgent sets the User-Agent header for the request
func (b *RequestBuilder) UserAgent(userAgent string) *RequestBuilder {
	b.headers.Set(HeaderUserAgent, userAgent)

	return b
}

// Referer sets the Referer header for the request
func (b *RequestBuilder) Referer(referer string) *RequestBuilder {
	b.headers.Set(HeaderReferer, referer)

	return b
}

// Auth applies an authentication method to the request
func (b *RequestBuilder) Auth(auth AuthMethod) *RequestBuilder {
	if auth.Valid() {
		b.auth = auth
	}

	return b
}

// Form sets form fields and files for the request
func (b *RequestBuilder) Form(v any) *RequestBuilder {
	formFields, formFiles, err := parseForm(v)

	if err != nil {
		if b.client.Logger != nil {
			b.client.Logger.Errorf("Error parsing form: %v", err)
		}

		return b
	}

	if formFields != nil {
		b.formFields = formFields
	}

	if formFiles != nil {
		b.formFiles = formFiles
	}

	return b
}

// FormFields sets multiple form fields at once
func (b *RequestBuilder) FormFields(fields any) *RequestBuilder {
	if b.formFields == nil {
		b.formFields = url.Values{}
	}

	values, err := parseFormFields(fields)
	if err != nil {
		if b.client.Logger != nil {
			b.client.Logger.Errorf("Error parsing form fields: %v", err)
		}

		return b
	}

	for key, value := range values {
		for _, v := range value {
			b.formFields.Add(key, v)
		}
	}

	return b
}

// FormField adds or updates a form field
func (b *RequestBuilder) FormField(key, val string) *RequestBuilder {
	if b.formFields == nil {
		b.formFields = url.Values{}
	}

	b.formFields.Add(key, val)

	return b
}

// DelFormField removes one or more form fields
func (b *RequestBuilder) DelFormField(key ...string) *RequestBuilder {
	if b.formFields != nil {
		for _, k := range key {
			b.formFields.Del(k)
		}
	}

	return b
}

// Files sets multiple files at once
func (b *RequestBuilder) Files(files ...*File) *RequestBuilder {
	if b.formFiles == nil {
		b.formFiles = []*File{}
	}

	b.formFiles = append(b.formFiles, files...)

	return b
}

// File adds a file to the request
func (b *RequestBuilder) File(key, filename string, content io.ReadCloser) *RequestBuilder {
	if b.formFiles == nil {
		b.formFiles = []*File{}
	}

	b.formFiles = append(b.formFiles, &File{
		Name:     key,
		FileName: filename,
		Content:  content,
	})

	return b
}

// DelFile removes one or more files from the request
func (b *RequestBuilder) DelFile(key ...string) *RequestBuilder {
	if b.formFiles != nil {
		for i, file := range b.formFiles {
			if slices.Contains(key, file.Name) {
				b.formFiles = append(b.formFiles[:i], b.formFiles[i+1:]...)
			}
		}
	}

	return b
}

// Body sets the request body
func (b *RequestBuilder) Body(body interface{}) *RequestBuilder {
	b.bodyData = body

	return b
}

// JSONBody sets the request body as JSON
func (b *RequestBuilder) JSONBody(v interface{}) *RequestBuilder {
	b.bodyData = v
	b.headers.Set(HeaderContentType, ContentTypeJSON)

	return b
}

// XMLBody sets the request body as XML
func (b *RequestBuilder) XMLBody(v interface{}) *RequestBuilder {
	b.bodyData = v
	b.headers.Set(HeaderContentType, ContentTypeXML)

	return b
}

// YAMLBody sets the request body as YAML
func (b *RequestBuilder) YAMLBody(v interface{}) *RequestBuilder {
	b.bodyData = v
	b.headers.Set(HeaderContentType, ContentTypeYAML)

	return b
}

// TextBody sets the request body as plain text
func (b *RequestBuilder) TextBody(v string) *RequestBuilder {
	b.bodyData = v
	b.headers.Set(HeaderContentType, ContentTypeText)

	return b
}

// RawBody sets the request body as raw bytes
func (b *RequestBuilder) RawBody(v []byte) *RequestBuilder {
	b.bodyData = v

	return b
}

// Timeout sets the request timeout
func (b *RequestBuilder) Timeout(timeout time.Duration) *RequestBuilder {
	b.timeout = timeout

	return b
}

// MaxRetries sets the maximum number of retry attempts
func (b *RequestBuilder) MaxRetries(maxRetries int) *RequestBuilder {
	b.maxRetries = maxRetries

	return b
}

// RetryStrategy sets the backoff strategy for retries
func (b *RequestBuilder) RetryStrategy(strategy BackoffStrategy) *RequestBuilder {
	b.retryStrategy = strategy

	return b
}

// RetryIf sets the custom retry condition function
func (b *RequestBuilder) RetryIf(retryIf RetryIfFunc) *RequestBuilder {
	b.retryIf = retryIf

	return b
}

func (b *RequestBuilder) do(ctx context.Context, req *http.Request) (*http.Response, error) {
	finalHandler := MiddlewareHandlerFunc(func(req *http.Request) (*http.Response, error) {
		var maxRetries = b.client.MaxRetries
		if b.maxRetries > 0 {
			maxRetries = b.maxRetries
		}

		var retryStrategy = b.client.RetryStrategy

		if b.retryStrategy != nil {
			retryStrategy = b.retryStrategy
		}

		var retryIf = b.client.RetryIf

		if b.retryIf != nil {
			retryIf = b.retryIf
		}

		if maxRetries < 1 {
			return b.client.HTTPClient.Do(req)
		}

		var lastErr error

		var resp *http.Response

		for attempt := 0; attempt <= maxRetries; attempt++ {
			resp, lastErr = b.client.HTTPClient.Do(req)

			shouldRetry := lastErr != nil || (resp != nil && retryIf != nil && retryIf(req, resp, lastErr))
			if !shouldRetry || attempt == maxRetries {
				if lastErr != nil {
					if b.client.Logger != nil {
						b.client.Logger.Errorf("Error after %d attempts: %v", attempt+1, lastErr)
					}
				}

				break
			}

			if resp != nil {
				if err := resp.Body.Close(); err != nil {
					if b.client.Logger != nil {
						b.client.Logger.Errorf("Error closing response body: %v", err)
					}
				}
			}

			if b.client.Logger != nil {
				b.client.Logger.Infof("Retrying request (attempt %d) after backoff", attempt+1)
			}

			// Logging context cancellation as an error condition
			select {
			case <-ctx.Done():
				if b.client.Logger != nil {
					b.client.Logger.Errorf("Request canceled or timed out: %v", ctx.Err())
				}

				return nil, ctx.Err()

			case <-time.After(retryStrategy(attempt)):
			}
		}

		return resp, lastErr
	})

	if b.middlewares != nil {
		for i := len(b.middlewares) - 1; i >= 0; i-- {
			finalHandler = b.middlewares[i](finalHandler)
		}
	}

	if b.client.Middlewares != nil {
		for i := len(b.client.Middlewares) - 1; i >= 0; i-- {
			finalHandler = b.client.Middlewares[i](finalHandler)
		}
	}

	return finalHandler(req)
}

// Stream sets the stream callback for the request
func (b *RequestBuilder) Stream(callback StreamCallback) *RequestBuilder {
	b.stream = callback

	return b
}

// StreamErr sets the error callback for the request.
func (b *RequestBuilder) StreamErr(callback StreamErrCallback) *RequestBuilder {
	b.streamErr = callback

	return b
}

// StreamDone sets the done callback for the request.
func (b *RequestBuilder) StreamDone(callback StreamDoneCallback) *RequestBuilder {
	b.streamDone = callback

	return b
}

func (b *RequestBuilder) setContentType() (io.Reader, string, error) {
	var body io.Reader

	var contentType string

	var err error

	switch {
	case len(b.formFiles) > 0:
		// If the request includes files, indicating multipart/form-data encoding is required
		body, contentType, err = b.prepareMultipartBody()

	case len(b.formFields) > 0:
		// For form fields without files, use application/x-www-form-urlencoded encoding
		body, contentType = b.prepareFormFieldsBody()

	case b.bodyData != nil:
		// Fallback to handling as per original logic for JSON, XML, etc
		body, contentType, err = b.prepareBodyBasedOnContentType()
	}

	if err != nil {
		if b.client.Logger != nil {
			// surface to the client logger as well
			b.client.Logger.Errorf("Error preparing request body: %v", err)
		}

		return nil, contentType, err
	}

	if contentType != "" {
		b.headers.Set(HeaderContentType, contentType)
	}

	return body, contentType, nil
}

func (b *RequestBuilder) requestChecks(req *http.Request) *http.Request {
	// apply the authentication method to the request
	if b.auth != nil {
		b.auth.Apply(req)
	} else if b.client.Auth != nil {
		b.client.Auth.Apply(req)
	}

	// set the headers from the client
	if b.client.Headers != nil {
		for key := range *b.client.Headers {
			values := (*b.client.Headers)[key]
			for _, value := range values {
				req.Header.Set(key, value)
			}
		}
	}
	// set the headers from the request builder
	if b.headers != nil {
		for key := range *b.headers {
			values := (*b.headers)[key]
			for _, value := range values {
				req.Header.Set(key, value)
			}
		}
	}

	// merge cookies from the client
	if b.client.Cookies != nil {
		for _, cookie := range b.client.Cookies {
			req.AddCookie(cookie)
		}
	}
	// merge cookies from the request builder
	if b.cookies != nil {
		for _, cookie := range b.cookies {
			req.AddCookie(cookie)
		}
	}

	return req
}

// Send executes the HTTP request
func (b *RequestBuilder) Send(ctx context.Context) (*Response, error) {
	body, _, err := b.setContentType()
	if err != nil {
		return nil, err
	}

	parsedURL, err := url.Parse(b.client.BaseURL + b.preparePath())
	if err != nil {
		if b.client.Logger != nil {
			// surface the error to the client logger as well
			b.client.Logger.Errorf("Error parsing URL: %v", err)
		}

		return nil, err
	}

	query := parsedURL.Query()

	for key, values := range b.queries {
		for _, value := range values {
			query.Set(key, value)
		}
	}

	parsedURL.RawQuery = query.Encode()

	var cancel context.CancelFunc

	if _, ok := ctx.Deadline(); !ok {
		if b.timeout > 0 {
			ctx, cancel = context.WithTimeout(ctx, b.timeout)
			defer cancel()
		}
	}

	req, err := http.NewRequestWithContext(ctx, b.method, parsedURL.String(), body)
	if err != nil {
		if b.client.Logger != nil {
			b.client.Logger.Errorf("Error creating request: %v", err)
		}

		return nil, rout.HTTPErrorResponse(err)
	}

	req = b.requestChecks(req)

	// Execute the HTTP request
	resp, err := b.do(ctx, req)
	if err != nil {
		if b.client.Logger != nil {
			b.client.Logger.Errorf("Error executing request: %v", err)
		}

		if resp != nil {
			_ = resp.Body.Close()
		}

		return nil, err
	}

	if resp == nil {
		if b.client.Logger != nil {
			b.client.Logger.Errorf("Response is nil")
		}

		return nil, fmt.Errorf("%w: %v", ErrResponseNil, err)
	}

	// Wrap and return the response
	return NewResponse(ctx, resp, b.client, b.stream, b.streamErr, b.streamDone)
}

func (b *RequestBuilder) prepareMultipartBody() (io.Reader, string, error) {
	var buf bytes.Buffer
	writer := multipart.NewWriter(&buf)

	// if a custom boundary is set, use it
	if b.boundary != "" {
		if err := writer.SetBoundary(b.boundary); err != nil {
			return nil, "", rout.HTTPErrorResponse(err)
		}
	}

	// add form fields
	for key, vals := range b.formFields {
		for _, val := range vals {
			if err := writer.WriteField(key, val); err != nil {
				return nil, "", rout.HTTPErrorResponse(err)
			}
		}
	}

	// add form files
	for _, file := range b.formFiles {
		// create a new multipart part for the file
		part, err := writer.CreateFormFile(file.Name, file.FileName)

		if err != nil {
			return nil, "", rout.HTTPErrorResponse(err)
		}
		// copy the file content to the part
		if _, err = io.Copy(part, file.Content); err != nil {
			return nil, "", rout.HTTPErrorResponse(err)
		}

		// close the file content if it's a closer
		if closer, ok := file.Content.(io.Closer); ok {
			if err = closer.Close(); err != nil {
				return nil, "", rout.HTTPErrorResponse(err)
			}
		}
	}

	// close the multipart writer
	if err := writer.Close(); err != nil {
		return nil, "", rout.HTTPErrorResponse(err)
	}

	return &buf, writer.FormDataContentType(), nil
}

func (b *RequestBuilder) prepareFormFieldsBody() (io.Reader, string) {
	data := b.formFields.Encode()

	return strings.NewReader(data), ContentTypeForm
}

func (b *RequestBuilder) prepareBodyBasedOnContentType() (io.Reader, string, error) {
	contentType := b.headers.Get(HeaderContentType)

	if contentType == "" && b.bodyData != nil {
		switch b.bodyData.(type) {
		case url.Values, map[string][]string, map[string]string:
			contentType = ContentTypeForm
		case map[string]interface{}, []interface{}, struct{}:
			contentType = ContentTypeJSONUTF8
		case string, []byte:
			contentType = ContentTypeText
		}
		b.headers.Set(HeaderContentType, contentType)
	}

	var body io.Reader

	var err error

	switch contentType {
	case ContentTypeJSON, ContentTypeJSONUTF8:
		body, err = b.client.JSONEncoder.Encode(b.bodyData)
	case ContentTypeXML:
		body, err = b.client.XMLEncoder.Encode(b.bodyData)
	case ContentTypeYAML:
		body, err = b.client.YAMLEncoder.Encode(b.bodyData)
	case ContentTypeForm:
		body, err = DefaultFormEncoder.Encode(b.bodyData)
	case ContentTypeText, ContentTypeApplicationOctetStream:
		switch data := b.bodyData.(type) {
		case string:
			body = strings.NewReader(data)
		case []byte:
			body = bytes.NewReader(data)
		default:
			err = fmt.Errorf("%w: %s", ErrUnsupportedContentType, contentType)
		}
	default:
		err = fmt.Errorf("%w: %s", ErrUnsupportedContentType, contentType)
	}

	return body, contentType, err
}
