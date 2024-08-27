package oauth2

import (
	"net/http"

	"golang.org/x/oauth2"

	"github.com/theopenlane/core/pkg/keygen"
	"github.com/theopenlane/core/pkg/sessions"
)

// StateHandler checks for a state cookie, if found, adds to context; if missing, a
// random generated value is added to the context and to a (short-lived) state cookie
// issued to the requester - this implements OAuth 2 RFC 6749 10.12 CSRF Protection
func StateHandler(config sessions.CookieConfig, success http.Handler) http.Handler {
	funk := func(w http.ResponseWriter, req *http.Request) {
		ctx := req.Context()

		queryParams := req.URL.Query()
		if queryParams.Get("state") != "" {
			ctx = WithState(ctx, queryParams.Get("state"))
		} else {
			val := keygen.GenerateRandomString(32) //nolint:mnd
			http.SetCookie(w, sessions.NewCookie(config.Name, val, &config))
			ctx = WithState(ctx, val)
		}

		// set redirect
		redirect, err := req.Cookie("redirect_to")
		if err == nil && redirect.Value != "" {
			ctx = WithRedirectURL(ctx, redirect.Value)
		} else {
			_ = req.ParseForm() //nolint: errcheck

			redirect := req.Form.Get("redirect_uri")
			if redirect != "" {
				http.SetCookie(w, sessions.NewCookie("redirect_to", redirect, &config))
				ctx = WithRedirectURL(ctx, redirect)
			}
		}

		success.ServeHTTP(w, req.WithContext(ctx))
	}

	return http.HandlerFunc(funk)
}

// LoginHandler reads the state value from the context and redirects requests to the AuthURL with that state value
func LoginHandler(config *oauth2.Config, failure http.Handler) http.Handler {
	if failure == nil {
		failure = DefaultFailureHandler
	}

	fn := func(w http.ResponseWriter, req *http.Request) {
		ctx := req.Context()
		state, err := StateFromContext(ctx)

		if err != nil {
			ctx = WithError(ctx, err)
			failure.ServeHTTP(w, req.WithContext(ctx))

			return
		}

		authURL := config.AuthCodeURL(state)
		http.Redirect(w, req, authURL, http.StatusFound)
	}

	return http.HandlerFunc(fn)
}

// CallbackHandler parses the auth code + state and compares it to the state value from the context
func CallbackHandler(config *oauth2.Config, success, failure http.Handler) http.Handler {
	if failure == nil {
		failure = DefaultFailureHandler
	}

	funk := func(w http.ResponseWriter, req *http.Request) {
		ctx := req.Context()

		authCode, state, err := parseCallback(req)
		if err != nil {
			ctx = WithError(ctx, err)
			failure.ServeHTTP(w, req.WithContext(ctx))

			return
		}

		ownerState, err := StateFromContext(ctx)
		if err != nil {
			ctx = WithError(ctx, err)
			failure.ServeHTTP(w, req.WithContext(ctx))

			return
		}

		if state != ownerState || state == "" {
			ctx = WithError(ctx, ErrInvalidState)
			failure.ServeHTTP(w, req.WithContext(ctx))

			return
		}

		token, err := config.Exchange(ctx, authCode)
		if err != nil {
			ctx = WithError(ctx, err)
			failure.ServeHTTP(w, req.WithContext(ctx))

			return
		}

		ctx = WithToken(ctx, token)
		success.ServeHTTP(w, req.WithContext(ctx))
	}

	return http.HandlerFunc(funk)
}

// parseCallback parses code and state parameters from the http.Request and returns them
func parseCallback(req *http.Request) (authCode, state string, err error) {
	if err = req.ParseForm(); err != nil {
		return "", "", err
	}

	authCode = req.Form.Get("code")
	state = req.Form.Get("state")

	if authCode == "" || state == "" {
		return "", "", ErrMissingCodeOrState
	}

	return authCode, state, nil
}
