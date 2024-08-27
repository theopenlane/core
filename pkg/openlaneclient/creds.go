package openlaneclient

import (
	"net/http"
	"strings"

	"github.com/theopenlane/utils/rout"

	"github.com/theopenlane/core/pkg/sessions"
	"github.com/theopenlane/httpsling"
)

// Credentials provides a basic interface for loading credentials
type Credentials interface {
	AccessToken() (string, error)
	GetSession() string
}

// Authorization contains the bearer token and optional session cookie
type Authorization struct {
	// BearerToken is the bearer token to be used in the authorization header
	// this can be the access token, api token, or personal access token
	BearerToken string
	// Session is the session cookie to be used in the request
	// this is required for requests using the access token
	Session string
}

func NewAuthorization(creds Credentials) (Authorization, error) {
	token, err := creds.AccessToken()
	if err != nil {
		return Authorization{}, err
	}

	session := creds.GetSession()

	return Authorization{
		BearerToken: token,
		Session:     session,
	}, nil
}

// Token implements the credentials interface and performs limited validation
func (a Authorization) AccessToken() (string, error) {
	if string(a.BearerToken) == "" {
		return "", rout.ErrInvalidCredentials
	}

	return a.BearerToken, nil
}

// Session implements the credentials interface
// session is not always required so no validation is provided
func (a Authorization) GetSession() string {
	return a.Session
}

// SetAuthorizationHeader sets the authorization header on the request if
// not already set
func (a Authorization) SetAuthorizationHeader(req *http.Request) {
	h := req.Header.Get(httpsling.HeaderAuthorization)
	if h == "" {
		// ignore error as we are setting the header
		token, _ := a.AccessToken()

		auth := httpsling.BearerAuth{
			Token: token,
		}

		auth.Apply(req)
	}
}

// SetSessionCookie sets the session cookie on the request if the session is not empty
func (a Authorization) SetSessionCookie(req *http.Request) {
	if a.Session != "" {
		if strings.Contains(req.Host, "localhost") {
			req.AddCookie(sessions.NewDevSessionCookie(a.GetSession()))
		} else {
			req.AddCookie(sessions.NewSessionCookie(a.GetSession()))
		}
	}
}
