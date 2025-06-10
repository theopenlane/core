package openlaneclient

import (
	"context"
	"net/http"

	"github.com/theopenlane/httpsling"
)

const (
	csrfHeader = "X-CSRF-Token"
	csrfCookie = "csrf_token"
	csrfPath   = "/livez"
)

// InitCSRF fetches a CSRF token and sets it on the client for subsequent
// requests. It returns an error if the token cannot be obtained.
func (c *OpenlaneClient) InitCSRF(ctx context.Context) error {
	token, err := c.FetchCSRFToken(ctx)
	if err != nil {
		return err
	}

	c.SetCSRFToken(token)

	return nil
}

// FetchCSRFToken performs a safe request to retrieve the CSRF cookie value.
func (c *OpenlaneClient) FetchCSRFToken(ctx context.Context) (string, error) {
	if c.HTTPSlingRequester().CookieJar() == nil {
		return "", ErrNoCookieJarSet
	}

	// make a GET request to acquire the CSRF cookie
	resp, err := c.HTTPSlingRequester().ReceiveWithContext(ctx, nil, httpsling.Get(csrfPath))
	if err != nil {
		return "", err
	}

	if resp != nil {
		resp.Body.Close()
	}

	cookies, err := c.Cookies()
	if err != nil {
		return "", err
	}

	for _, ck := range cookies {
		if ck.Name == csrfCookie {
			if ck.Value == "" {
				return "", ErrEmptyCSRFToken
			}

			return ck.Value, nil
		}
	}

	// do not return an error, if CSRF protection is not enabled
	// there may not be a CSRF cookie set
	return "", nil
}

// SetCSRFToken sets the CSRF header on the client for all subsequent requests.
func (c *OpenlaneClient) SetCSRFToken(token string) {
	if token == "" {
		return
	}

	if c.HTTPSlingRequester().Header == nil {
		c.HTTPSlingRequester().Header = make(http.Header)
	}

	c.HTTPSlingRequester().Header.Set(csrfHeader, token)
}
