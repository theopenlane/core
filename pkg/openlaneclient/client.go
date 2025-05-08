package openlaneclient

import (
	"context"
	"net/http"
	"net/url"
	"strings"
	"time"

	"golang.org/x/oauth2"

	"github.com/theopenlane/httpsling"
	"github.com/theopenlane/iam/auth"
	"github.com/theopenlane/iam/sessions"
	"github.com/vektah/gqlparser/v2/gqlerror"

	api "github.com/theopenlane/core/pkg/models"
)

const (
	// cookieExpiryMinutes is the duration for which the cookies are valid
	cookieExpiryMinutes = 10 * time.Minute // nolint:revive
)

// OpenlaneClient wraps the Openlane API client methods to form a single client interface
type OpenlaneClient struct {
	OpenlaneRestClient
	OpenlaneGraphClient
}

// A Reauthenticator generates new access and refresh pair given a valid refresh token
type Reauthenticator interface {
	Refresh(context.Context, *api.RefreshRequest) (*api.RefreshReply, error)
}

// NewWithDefaults creates a new API v1 client with default configuration
func NewWithDefaults(opts ...ClientOption) (*OpenlaneClient, error) {
	config := NewDefaultConfig()

	return New(config, opts...)
}

// New creates a new API v1 client that implements the Client interface
func New(config Config, opts ...ClientOption) (*OpenlaneClient, error) {
	// configure rest client
	c, err := NewRestClient(config, opts...)
	if err != nil {
		return nil, err
	}

	api := c.(*APIv1)

	// create the graph client
	// use api.Config instead of config because some fields are updated in NewRestClient
	graphClient := NewClient(
		api.Requester.HTTPClient(),
		graphRequestPath(api.Config),
		&api.Config.Clientv2Options,
		api.Config.Interceptors...,
	)

	return &OpenlaneClient{
		c,
		graphClient,
	}, nil
}

// APIv1 implements the Client interface and provides methods to interact with the API
type APIv1 struct {
	// Config is the configuration for the APIv1 client
	Config Config
	// Requester is the HTTP client for the APIv1 client
	Requester *httpsling.Requester
}

// Config is the configuration for the APIv1 client
func (c *OpenlaneClient) Config() Config {
	api := c.OpenlaneRestClient.(*APIv1)

	return api.Config
}

// HTTPSlingRequester is the http client for the APIv1 client
func (c *OpenlaneClient) HTTPSlingRequester() *httpsling.Requester {
	api := c.OpenlaneRestClient.(*APIv1)

	return api.Requester
}

// AccessToken returns the access token cached on the client or an error if it is not
// available. This method is primarily used for testing but can be used to fetch the
// access token for debugging or inspection if necessary.
func (c *OpenlaneClient) AccessToken() (_ string, err error) {
	var cookies []*http.Cookie

	if cookies, err = c.Cookies(); err != nil {
		return "", err
	}

	for _, cookie := range cookies {
		if cookie.Name == auth.AccessTokenCookie {
			return cookie.Value, nil
		}
	}

	return "", err
}

// RefreshToken returns the refresh token cached on the client or an error if it is not
// available. This method is primarily used for testing but can be used to fetch the
// refresh token for debugging or inspection if necessary.
func (c *OpenlaneClient) RefreshToken() (_ string, err error) {
	var cookies []*http.Cookie

	if cookies, err = c.Cookies(); err != nil {
		return "", err
	}

	for _, cookie := range cookies {
		if cookie.Name == auth.RefreshTokenCookie {
			return cookie.Value, nil
		}
	}

	return "", err
}

// SetAuthTokens is a helper function to set the access and refresh tokens on the
// client cookie jar.
func (c *OpenlaneClient) SetAuthTokens(access, refresh string) error {
	if c.HTTPSlingRequester().CookieJar() == nil {
		return ErrNoCookieJarSet
	}

	// The URL for the cookies
	u := c.Config().BaseURL.ResolveReference(&url.URL{Path: "/"})

	// Set the cookies on the client
	cookies := make([]*http.Cookie, 0, 2) //nolint:mnd
	if access != "" {
		cookies = append(cookies, &http.Cookie{
			Name:     auth.AccessTokenCookie,
			Value:    access,
			Expires:  time.Now().Add(cookieExpiryMinutes),
			HttpOnly: true,
			Secure:   true,
		})
	}

	if refresh != "" {
		cookies = append(cookies, &http.Cookie{
			Name:    auth.RefreshTokenCookie,
			Value:   refresh,
			Expires: time.Now().Add(cookieExpiryMinutes),
			Secure:  true,
		})
	}

	c.HTTPSlingRequester().CookieJar().SetCookies(u, cookies)

	return nil
}

// ClearAuthTokens clears the access and refresh tokens on the client Jar.
func (c *OpenlaneClient) ClearAuthTokens() {
	if cookies, err := c.Cookies(); err == nil {
		// Expire the access and refresh cookies.
		for _, cookie := range cookies {
			switch cookie.Name {
			case auth.AccessTokenCookie:
				cookie.MaxAge = -1
			case auth.RefreshTokenCookie:
				cookie.MaxAge = -1
			}
		}
	}
}

// Returns the cookies set from the previous request(s) on the client Jar.
func (c *OpenlaneClient) Cookies() ([]*http.Cookie, error) {
	if c.HTTPSlingRequester().CookieJar() == nil {
		return nil, ErrNoCookieJarSet
	}

	cookies := c.HTTPSlingRequester().CookieJar().Cookies(c.Config().BaseURL)

	return cookies, nil
}

// GetSessionFromCookieJar parses the cookie jar for the session cookie
func (c *OpenlaneClient) GetSessionFromCookieJar() (sessionID string, err error) {
	cookies, err := c.Cookies()
	if err != nil {
		return "", err
	}

	cookieName := sessions.DefaultCookieName

	// Use the dev cookie when running on localhost
	if strings.Contains(c.Config().BaseURL.Host, "localhost") {
		cookieName = sessions.DevCookieName
	}

	for _, c := range cookies {
		if c.Name == cookieName {
			return c.Value, nil
		}
	}

	return "", nil
}

// GetAuthTokensFromCookieJar gets the access and refresh tokens from the cookie jar
// and returns them as an oauth2.Token if they are set
func (c *OpenlaneClient) GetAuthTokensFromCookieJar() *oauth2.Token {
	token := oauth2.Token{}

	if cookies, err := c.Cookies(); err == nil {
		for _, cookie := range cookies {
			switch cookie.Name {
			case auth.AccessTokenCookie:
				token.AccessToken = cookie.Value
			case auth.RefreshTokenCookie:
				token.RefreshToken = cookie.Value
			}
		}
	}

	// return nil if the tokens are not set
	if token.AccessToken == "" || token.RefreshToken == "" {
		return nil
	}

	return &token
}

// GetErrorMessage returns the error message from the GraphQL error extensions
func GetErrorCode(err error) string {
	if err == nil {
		return ""
	}

	gqlErr, ok := err.(*gqlerror.Error)
	if !ok {
		return ""
	}

	if code, ok := gqlErr.Extensions["code"].(string); ok {
		return code
	}

	return ""
}

// GetErrorMessage returns the error message from the GraphQL error extensions
func GetErrorMessage(err error) string {
	if err == nil {
		return ""
	}

	gqlErr, ok := err.(*gqlerror.Error)
	if !ok {
		return ""
	}

	if message, ok := gqlErr.Extensions["message"].(string); ok {
		return message
	}

	return ""
}
