package httpsling

import (
	"net/http"
	"net/url"

	"github.com/theopenlane/utils/rout"
)

// verifyProxy validates the given proxy URL, supporting http, https, and socks5 schemes
func verifyProxy(proxyURL string) (*url.URL, error) {
	parsedURL, err := url.Parse(proxyURL)
	if err != nil {
		return nil, err
	}

	// Check if the scheme is supported
	switch parsedURL.Scheme {
	case "http", "https", "socks5":
		return parsedURL, nil
	default:
		return nil, ErrUnsupportedScheme
	}
}

// SetProxy configures the client to use a proxy. Supports http, https, and socks5 proxies
func (c *Client) SetProxy(proxyURL string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Validate and parse the proxy URL
	validatedProxyURL, err := verifyProxy(proxyURL)
	if err != nil {
		return err
	}

	// Ensure the HTTPClient's Transport is properly initialized
	if c.HTTPClient.Transport == nil {
		c.HTTPClient.Transport = &http.Transport{}
	}

	// Assert the Transport to *http.Transport to access the Proxy field
	transport, ok := c.HTTPClient.Transport.(*http.Transport)
	if !ok {
		return rout.HTTPErrorResponse(err)
	}

	transport.Proxy = http.ProxyURL(validatedProxyURL)

	return nil
}

// RemoveProxy clears any configured proxy, allowing direct connections
func (c *Client) RemoveProxy() {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.HTTPClient.Transport == nil {
		return
	}

	transport, ok := c.HTTPClient.Transport.(*http.Transport)

	if !ok {
		return // If it's not *http.Transport, it doesn't have a proxy to remove
	}

	transport.Proxy = nil
}
