package httpsling

import "net/http"

// AuthType represents the type of authentication method
type AuthType string

var (
	// BearerAuthType is the Bearer token authentication type using the Authorization header
	BearerAuthType AuthType = "Bearer"
	// BasicAuthType is the Basic authentication type using a username and password
	BasicAuthType AuthType = "Basic"
)

// AuthMethod defines the interface for applying authentication strategies to httpsling
type AuthMethod interface {
	// Apply adds the authentication method to the request
	Apply(req *http.Request)
	// Valid checks if the authentication method is valid
	Valid() bool
}

// BasicAuth represents HTTP Basic Authentication credentials
type BasicAuth struct {
	Username string
	Password string
}

// CustomAuth allows for custom Authorization header values
type CustomAuth struct {
	Header string
}

// BearerAuth represents an OAuth 2.0 Bearer token
type BearerAuth struct {
	Token string
}

// Apply adds the Basic Auth credentials to the request
func (b BasicAuth) Apply(req *http.Request) {
	req.SetBasicAuth(b.Username, b.Password)
}

// Valid checks if the Basic Auth credentials are present
func (b BasicAuth) Valid() bool {
	return b.Username != "" && b.Password != ""
}

// Apply adds the Bearer token to the request's Authorization header
func (b BearerAuth) Apply(req *http.Request) {
	if b.Valid() {
		req.Header.Set(HeaderAuthorization, "Bearer "+b.Token)
	}
}

// Valid checks if the Bearer token is present
func (b BearerAuth) Valid() bool {
	return b.Token != ""
}

// Apply sets a custom Authorization header value
func (c CustomAuth) Apply(req *http.Request) {
	if c.Valid() {
		req.Header.Set(HeaderAuthorization, c.Header)
	}
}

// Valid checks if the custom Authorization header value is present
func (c CustomAuth) Valid() bool {
	return c.Header != ""
}
