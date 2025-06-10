package openlaneclient

import (
	"context"

	"github.com/theopenlane/httpsling"
)

const (
	csrfHeader = "X-CSRF-Token"
	csrfCookie = "csrf_token"
	csrfPath   = "/livez"
)

// InitCSRF fetches a CSRF token and sets it on the client for subsequent
// requests. It returns an error if the token cannot be obtained.
func (c *OpenlaneClient) InitCSRF(ctx context.Context) (string, error) {
	token, err := c.fetchCSRFToken(ctx)
	if err != nil {
		return "", err
	}

	return token, nil
}

// fetchCSRFToken performs a safe request to retrieve the CSRF cookie value.
func (c *OpenlaneClient) fetchCSRFToken(ctx context.Context) (string, error) {
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

	return c.getCSRFToken()
}

// getCSRFToken retrieves the CSRF token from the cookie jar
// and returns it. If the token is not found or is empty, it returns an error.
// if it doesn't exist, it returns an empty string without an error.
// this is used for cases where CSRF protection is not enabled.
func (c *OpenlaneClient) getCSRFToken() (string, error) {
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
