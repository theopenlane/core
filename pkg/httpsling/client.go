package httpsling

import (
	"crypto/tls"
	"net/http"
	"net/http/cookiejar"
	"sync"
	"time"
)

// Client represents an HTTP client and is the main control mechanism for making HTTP requests
type Client struct {
	// mu is a mutex to protect the client's configuration
	mu sync.RWMutex
	// BaseURL is the base URL for all httpsling made by this client
	BaseURL string
	// Headers are the default headers to be sent with each request
	Headers *http.Header
	// Cookies are the default cookies to be sent with each request
	Cookies []*http.Cookie
	// Middlewares are the request/response manipulation middlewares
	Middlewares []Middleware
	// TLSConfig is the TLS configuration for the client
	TLSConfig *tls.Config
	// MaxRetries is the maximum number of retry attempts
	MaxRetries int
	// RetryStrategy is the backoff strategy for retries
	RetryStrategy BackoffStrategy
	// RetryIf is the custom retry condition function
	RetryIf RetryIfFunc
	// HTTPClient is the underlying HTTP client
	HTTPClient *http.Client
	// JSONEncoder is the JSON encoder for the client
	JSONEncoder Encoder
	// JSONDecoder is the JSON decoder for the client
	JSONDecoder Decoder
	// XMLEncoder is the XML encoder for the client
	XMLEncoder Encoder
	// XMLDecoder is the XML decoder for the client
	XMLDecoder Decoder
	// YAMLEncoder is the YAML encoder for the client
	YAMLEncoder Encoder
	// YAMLDecoder is the YAML decoder for the client
	YAMLDecoder Decoder
	// Logger is the logger instance for the client
	Logger Logger
	// auth is the authentication method for the client
	Auth AuthMethod
}

// Config sets up the initial configuration for the HTTP client - you need to initialize multiple if you want the behaviors to be different
type Config struct {
	// The base URL for all httpsling made by this client
	BaseURL string
	// Default headers to be sent with each request
	Headers *http.Header
	// Default Cookies to be sent with each request
	Cookies map[string]string
	// Timeout for httpsling
	Timeout time.Duration
	// Cookie jar for the client
	CookieJar *cookiejar.Jar
	// Middlewares for request/response manipulation
	Middlewares []Middleware
	// TLS configuration for the client
	TLSConfig *tls.Config
	// Custom transport for the client
	Transport http.RoundTripper
	// Maximum number of retry attempts
	MaxRetries int
	// RetryStrategy defines the backoff strategy for retries
	RetryStrategy BackoffStrategy
	// RetryIf defines the custom retry condition function
	RetryIf RetryIfFunc
	// Logger instance for the client
	Logger Logger
}

// URL creates a new HTTP client with the given base URL
func URL(baseURL string) *Client {
	return Create(&Config{BaseURL: baseURL})
}

// Create initializes a new HTTP client with the given configuration
func Create(config *Config) *Client {
	cfg, httpClient := setInitialClientDetails(config)

	// Return a new Client instance
	client := &Client{
		BaseURL:     cfg.BaseURL,
		HTTPClient:  httpClient,
		JSONEncoder: DefaultJSONEncoder,
		JSONDecoder: DefaultJSONDecoder,
		XMLEncoder:  DefaultXMLEncoder,
		XMLDecoder:  DefaultXMLDecoder,
		YAMLEncoder: DefaultYAMLEncoder,
		YAMLDecoder: DefaultYAMLDecoder,
		TLSConfig:   cfg.TLSConfig,
	}

	if config != nil {
		client.Headers = config.Headers
	}

	return finalizeClientChecks(client, cfg, httpClient)
}

// finalizeClientChecks is a helper function to finalize the client configuration
func finalizeClientChecks(client *Client, config *Config, httpClient *http.Client) *Client {
	if client.TLSConfig != nil && httpClient.Transport != nil {
		httpTransport := httpClient.Transport.(*http.Transport)
		httpTransport.TLSClientConfig = client.TLSConfig
	} else if client.TLSConfig != nil {
		httpClient.Transport = &http.Transport{
			TLSClientConfig: client.TLSConfig,
		}
	}

	if config.Middlewares != nil {
		client.Middlewares = config.Middlewares
	} else {
		client.Middlewares = make([]Middleware, 0)
	}

	if config.Cookies != nil {
		client.SetDefaultCookies(config.Cookies)
	}

	if config.MaxRetries != 0 {
		client.MaxRetries = config.MaxRetries
	}

	if config.RetryStrategy != nil {
		client.RetryStrategy = config.RetryStrategy
	} else {
		client.RetryStrategy = DefaultBackoffStrategy(1 * time.Second)
	}

	if config.RetryIf != nil {
		client.RetryIf = config.RetryIf
	} else {
		client.RetryIf = DefaultRetryIf
	}

	if config.Logger != nil {
		client.Logger = config.Logger
	}

	return client
}

// setInitialClientDetails is a helper function that sets the initial configuration for the client and mostly breaks up how large of a function check the Create function is
func setInitialClientDetails(config *Config) (*Config, *http.Client) {
	if config == nil {
		config = &Config{}
	}

	httpClient := &http.Client{}

	if config.Transport != nil {
		httpClient.Transport = config.Transport
	}

	if config.Timeout != 0 {
		httpClient.Timeout = config.Timeout
	}

	if config.CookieJar != nil {
		httpClient.Jar = config.CookieJar
	}

	return config, httpClient
}

// SetBaseURL sets the base URL for the client
func (c *Client) SetBaseURL(baseURL string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.BaseURL = baseURL
}

// AddMiddleware adds a middleware to the client
func (c *Client) AddMiddleware(middlewares ...Middleware) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.Middlewares = append(c.Middlewares, middlewares...)
}

// SetTLSConfig sets the TLS configuration for the client
func (c *Client) SetTLSConfig(config *tls.Config) *Client {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.TLSConfig = config

	if c.HTTPClient == nil {
		c.HTTPClient = &http.Client{}
	}

	// Apply the TLS configuration to the existing transport if possible
	// If the current transport is not an *http.Transport, replace it
	if transport, ok := c.HTTPClient.Transport.(*http.Transport); ok {
		transport.TLSClientConfig = config
	} else {
		c.HTTPClient.Transport = &http.Transport{
			TLSClientConfig: config,
		}
	}

	return c
}

// InsecureSkipVerify sets the TLS configuration to skip certificate verification
func (c *Client) InsecureSkipVerify() *Client {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.TLSConfig == nil {
		c.TLSConfig = &tls.Config{
			MinVersion: tls.VersionTLS12,
		}
	}

	c.TLSConfig.InsecureSkipVerify = true

	if c.HTTPClient == nil {
		c.HTTPClient = &http.Client{}
	}

	if transport, ok := c.HTTPClient.Transport.(*http.Transport); ok {
		transport.TLSClientConfig = c.TLSConfig
	} else {
		c.HTTPClient.Transport = &http.Transport{
			TLSClientConfig: c.TLSConfig,
		}
	}

	return c
}

// SetHTTPClient sets the HTTP client for the client
func (c *Client) SetHTTPClient(httpClient *http.Client) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.HTTPClient = httpClient
}

// SetDefaultHeaders sets the default headers for the client
func (c *Client) SetDefaultHeaders(headers *http.Header) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.Headers = headers
}

// SetDefaultHeader adds or updates a default header
func (c *Client) SetDefaultHeader(key, value string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.Headers == nil {
		c.Headers = &http.Header{}
	}

	c.Headers.Set(key, value)
}

// AddDefaultHeader adds a default header
func (c *Client) AddDefaultHeader(key, value string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.Headers == nil {
		c.Headers = &http.Header{}
	}

	c.Headers.Add(key, value)
}

// DelDefaultHeader removes a default header
func (c *Client) DelDefaultHeader(key string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.Headers != nil { // only attempt to delete if initialized
		c.Headers.Del(key)
	}
}

// SetDefaultContentType sets the default content type for the client
func (c *Client) SetDefaultContentType(contentType string) {
	c.SetDefaultHeader(HeaderContentType, contentType)
}

// SetDefaultAccept sets the default accept header for the client
func (c *Client) SetDefaultAccept(accept string) {
	c.SetDefaultHeader(HeaderAccept, accept)
}

// SetDefaultUserAgent sets the default user agent for the client
func (c *Client) SetDefaultUserAgent(userAgent string) {
	c.SetDefaultHeader(HeaderUserAgent, userAgent)
}

// SetDefaultReferer sets the default referer for the client
func (c *Client) SetDefaultReferer(referer string) {
	c.SetDefaultHeader(HeaderReferer, referer)
}

// SetDefaultTimeout sets the default timeout for the client
func (c *Client) SetDefaultTimeout(timeout time.Duration) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.HTTPClient.Timeout = timeout
}

// SetDefaultTransport sets the default transport for the client
func (c *Client) SetDefaultTransport(transport http.RoundTripper) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.HTTPClient.Transport = transport
}

// SetDefaultCookieJar sets the default cookie jar for the client
func (c *Client) SetCookieJar(jar *cookiejar.Jar) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.HTTPClient.Jar = jar
}

// SetDefaultCookieJar sets the creates a new cookie jar and sets it for the client
func (c *Client) SetDefaultCookieJar() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Create a new cookie jar
	jar, err := cookiejar.New(nil)
	if err != nil {
		return err
	}

	c.HTTPClient.Jar = jar

	return nil
}

// SetDefaultCookies sets the default cookies for the client
func (c *Client) SetDefaultCookies(cookies map[string]string) {
	for name, value := range cookies {
		c.SetDefaultCookie(name, value)
	}
}

// SetDefaultCookie sets a default cookie for the client
func (c *Client) SetDefaultCookie(name, value string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.Cookies == nil {
		c.Cookies = make([]*http.Cookie, 0)
	}

	c.Cookies = append(c.Cookies, &http.Cookie{Name: name, Value: value})
}

// DelDefaultCookie removes a default cookie from the client
func (c *Client) DelDefaultCookie(name string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.Cookies != nil { // Only attempt to delete if Cookies is initialized
		for i, cookie := range c.Cookies {
			if cookie.Name == name {
				c.Cookies = append(c.Cookies[:i], c.Cookies[i+1:]...)
				break
			}
		}
	}
}

// SetJSONMarshal sets the JSON marshal function for the client's JSONEncoder
func (c *Client) SetJSONMarshal(marshalFunc func(v any) ([]byte, error)) {
	c.JSONEncoder = &JSONEncoder{
		MarshalFunc: marshalFunc,
	}
}

// SetJSONUnmarshal sets the JSON unmarshal function for the client's JSONDecoder
func (c *Client) SetJSONUnmarshal(unmarshalFunc func(data []byte, v any) error) {
	c.JSONDecoder = &JSONDecoder{
		UnmarshalFunc: unmarshalFunc,
	}
}

// SetXMLMarshal sets the XML marshal function for the client's XMLEncoder
func (c *Client) SetXMLMarshal(marshalFunc func(v any) ([]byte, error)) {
	c.XMLEncoder = &XMLEncoder{
		MarshalFunc: marshalFunc,
	}
}

// SetXMLUnmarshal sets the XML unmarshal function for the client's XMLDecoder
func (c *Client) SetXMLUnmarshal(unmarshalFunc func(data []byte, v any) error) {
	c.XMLDecoder = &XMLDecoder{
		UnmarshalFunc: unmarshalFunc,
	}
}

// SetYAMLMarshal sets the YAML marshal function for the client's YAMLEncoder
func (c *Client) SetYAMLMarshal(marshalFunc func(v any) ([]byte, error)) {
	c.YAMLEncoder = &YAMLEncoder{
		MarshalFunc: marshalFunc,
	}
}

// SetYAMLUnmarshal sets the YAML unmarshal function for the client's YAMLDecoder
func (c *Client) SetYAMLUnmarshal(unmarshalFunc func(data []byte, v any) error) {
	c.YAMLDecoder = &YAMLDecoder{
		UnmarshalFunc: unmarshalFunc,
	}
}

// SetMaxRetries sets the maximum number of retry attempts
func (c *Client) SetMaxRetries(maxRetries int) *Client {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.MaxRetries = maxRetries

	return c
}

// SetRetryStrategy sets the backoff strategy for retries
func (c *Client) SetRetryStrategy(strategy BackoffStrategy) *Client {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.RetryStrategy = strategy

	return c
}

// SetRetryIf sets the custom retry condition function
func (c *Client) SetRetryIf(retryIf RetryIfFunc) *Client {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.RetryIf = retryIf

	return c
}

// SetAuth configures an authentication method for the client
func (c *Client) SetAuth(auth AuthMethod) {
	if auth.Valid() {
		c.Auth = auth
	}
}

// SetLogger sets logger instance in client
func (c *Client) SetLogger(logger Logger) *Client {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.Logger = logger

	return c
}

// Get initiates a GET request
func (c *Client) Get(path string) *RequestBuilder {
	return c.NewRequestBuilder(http.MethodGet, path)
}

// Post initiates a POST request
func (c *Client) Post(path string) *RequestBuilder {
	return c.NewRequestBuilder(http.MethodPost, path)
}

// Delete initiates a DELETE request
func (c *Client) Delete(path string) *RequestBuilder {
	return c.NewRequestBuilder(http.MethodDelete, path)
}

// Put initiates a PUT request
func (c *Client) Put(path string) *RequestBuilder {
	return c.NewRequestBuilder(http.MethodPut, path)
}

// Patch initiates a PATCH request
func (c *Client) Patch(path string) *RequestBuilder {
	return c.NewRequestBuilder(http.MethodPatch, path)
}

// Options initiates an OPTIONS request
func (c *Client) Options(path string) *RequestBuilder {
	return c.NewRequestBuilder(http.MethodOptions, path)
}

// Head initiates a HEAD request
func (c *Client) Head(path string) *RequestBuilder {
	return c.NewRequestBuilder(http.MethodHead, path)
}

// Connect initiates a CONNECT request
func (c *Client) Connect(path string) *RequestBuilder {
	return c.NewRequestBuilder(http.MethodConnect, path)
}

// Trace initiates a TRACE request
func (c *Client) Trace(path string) *RequestBuilder {
	return c.NewRequestBuilder(http.MethodTrace, path)
}

// Custom initiates a custom request
func (c *Client) Custom(path, method string) *RequestBuilder {
	return c.NewRequestBuilder(method, path)
}
