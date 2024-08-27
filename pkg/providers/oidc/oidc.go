package oidc

import (
	"context"
	"net/http"

	"github.com/coreos/go-oidc/v3/oidc"
	"golang.org/x/oauth2"

	echo "github.com/theopenlane/echox"
)

type User struct {
	OAuth2Token *oauth2.Token
	IDToken     *IDToken
}

type UserResponse struct {
	AccessToken string
	IDToken     string
	Name        string
	Email       string
	Picture     string
}

type IDToken struct {
	RawToken string
	Claims   *Claims
}

type Claims struct {
	Email         string `json:"email"`
	EmailVerified bool   `json:"email_verified"`
	Name          string `json:"name"`
	Picture       string `json:"picture"`
}

func ExchangeCode(ctx context.Context, r *http.Request, config *oauth2.Config, provider *oidc.Provider) (*User, error) {
	state, err := r.Cookie("state")
	if err != nil {
		return nil, echo.NewHTTPError(http.StatusBadRequest, "State cookie is not set in request")
	}

	if r.URL.Query().Get("state") != state.Value {
		return nil, echo.NewHTTPError(http.StatusBadRequest, "State cookie did not match")
	}

	oauth2Token, err := config.Exchange(ctx, r.URL.Query().Get("code"))
	if err != nil {
		return nil, echo.NewHTTPError(http.StatusInternalServerError, "Unable to exchange token: "+err.Error())
	}

	rawIDToken, ok := oauth2Token.Extra("id_token").(string)

	if !ok {
		return nil, echo.NewHTTPError(http.StatusInternalServerError, "No id_token field in oauth2 token.")
	}

	oidcConfig := &oidc.Config{
		ClientID: config.ClientID,
	}

	verifier := provider.Verifier(oidcConfig)
	idToken, err := verifier.Verify(ctx, rawIDToken)

	if err != nil {
		return nil, echo.NewHTTPError(http.StatusInternalServerError, "Unable to verify ID Token: "+err.Error())
	}

	nonce, err := r.Cookie("nonce")
	if err != nil {
		return nil, echo.NewHTTPError(http.StatusBadRequest, "Nonce is not provided")
	}

	if idToken.Nonce != nonce.Value {
		return nil, echo.NewHTTPError(http.StatusBadRequest, "Nonce did not match")
	}

	var claims Claims

	if err := idToken.Claims(&claims); err != nil {
		return nil, echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	user := User{
		OAuth2Token: oauth2Token,
		IDToken: &IDToken{
			RawToken: rawIDToken,
			Claims:   &claims,
		},
	}

	return &user, nil
}
