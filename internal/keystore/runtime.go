package keystore

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/theopenlane/httpsling"
	"golang.org/x/oauth2"
)

const runtimeHTTPStatusSuccess = 2

var defaultHTTPRequester = httpsling.MustNew(httpsling.WithDoer(http.DefaultClient))

func sendHTTPRequest(ctx context.Context, opts ...httpsling.Option) (*http.Response, error) {
	requester, err := defaultHTTPRequester.With(opts...)
	if err != nil {
		return nil, err
	}
	return requester.SendWithContext(ctx)
}

type IntegrationUserInfo struct {
	ID       string
	Email    string
	Username string
	Raw      json.RawMessage
}

type ProviderRuntime struct {
	Spec        ProviderSpec
	OAuthConfig *oauth2.Config // populated when Spec.AuthType is oauth2/oidc
	Validator   Validator
}

type Validator interface {
	Validate(ctx context.Context, accessToken string, rt *ProviderRuntime) (*IntegrationUserInfo, error)
}

// Registry keyed by provider name (e.g., "github", "slack", "oidc_generic")
type Registry map[string]*ProviderRuntime

// BuildRegistry constructs the registry from config + built-in validators.
func BuildRegistry(specs map[string]ProviderSpec) (Registry, error) {
	reg := Registry{}
	for name, sp := range specs {
		rt := &ProviderRuntime{Spec: sp}
		switch sp.AuthType {
		case AuthTypeOAuth2, AuthTypeOIDC:
			endpoint := oauth2.Endpoint{
				AuthURL:  sp.OAuth.AuthURL,
				TokenURL: sp.OAuth.TokenURL,
			}
			rt.OAuthConfig = &oauth2.Config{
				ClientID:     sp.OAuth.ClientID,
				ClientSecret: sp.OAuth.ClientSecret,
				Endpoint:     endpoint,
				Scopes:       sp.OAuth.Scopes,
				RedirectURL:  sp.OAuth.RedirectURI,
			}
			rt.Validator = DefaultUserInfoValidator{}
		case AuthTypeAPIKey, AuthTypeWorkloadIdentity, AuthTypeGitHubApp:
			rt.Validator = NoopValidator{}
		default:
			rt.Validator = DefaultUserInfoValidator{}
		}

		// Optional: override per provider
		switch strings.ToLower(sp.Name) {
		case "github":
			rt.Validator = GitHubValidator{}
		case "slack":
			rt.Validator = SlackValidator{}
			// add others as needed
		}
		reg[name] = rt
	}
	return reg, nil
}

// ----- Default validator (driven by UserInfoSpec) -----

type DefaultUserInfoValidator struct{}

func (v DefaultUserInfoValidator) Validate(ctx context.Context, accessToken string, rt *ProviderRuntime) (*IntegrationUserInfo, error) {
	ui := rt.Spec.UserInfo
	if ui == nil {
		return nil, errNoUserInfoSpec
	}

	headerName := ui.AuthHeader
	if headerName == "" {
		headerName = httpsling.HeaderAuthorization
	}

	var authHeaderValue string
	switch ui.AuthStyle {
	case AuthHeaderStyleBearer:
		authHeaderValue = httpsling.BearerAuthHeader + accessToken
	case AuthHeaderStyleToken:
		authHeaderValue = "token " + accessToken
	case AuthHeaderStyleNone:
		// no header
	default:
		authHeaderValue = fmt.Sprintf("%s %s", httpsling.BearerAuthHeader, accessToken)
	}

	requestOpts := []httpsling.Option{
		httpsling.Method(string(ui.Method)),
		httpsling.URL(ui.URL),
	}
	if authHeaderValue != "" {
		requestOpts = append(requestOpts, httpsling.Header(headerName, authHeaderValue))
	}

	resp, err := sendHTTPRequest(ctx, requestOpts...)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		b, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("%w: status %d: %s", errUserInfoHTTPFailure, resp.StatusCode, string(b))
	}
	body, _ := io.ReadAll(resp.Body)

	var m map[string]any
	if err := json.Unmarshal(body, &m); err != nil {
		return nil, err
	}
	get := func(path JSONFieldPath) string {
		if path == "" {
			return ""
		}
		parts := strings.Split(string(path), ".")
		var cur any = m
		for _, p := range parts {
			if mm, ok := cur.(map[string]any); ok {
				cur = mm[p]
			} else {
				return ""
			}
		}
		switch v := cur.(type) {
		case string:
			return v
		case float64:
			return strings.TrimSuffix(strings.TrimSuffix(fmt.Sprintf("%.f", v), ".0"), ".00")
		default:
			b, _ := json.Marshal(v)
			return string(b)
		}
	}

	info := &IntegrationUserInfo{
		ID:       get(ui.IDPath),
		Email:    get(ui.EmailPath),
		Username: get(ui.LoginPath),
		Raw:      json.RawMessage(body),
	}

	// Optional secondary email fetch (e.g., GitHub-style)
	if info.Email == "" && ui.SecondaryEmailURL != "" {
		em, _ := fetchPrimaryEmail(ctx, accessToken, ui)
		if em != "" {
			info.Email = em
		}
	}
	return info, nil
}

func fetchPrimaryEmail(ctx context.Context, accessToken string, ui *UserInfoSpec) (string, error) {
	headerName := ui.AuthHeader
	if headerName == "" {
		headerName = httpsling.HeaderAuthorization
	}
	switch ui.AuthStyle {
	case AuthHeaderStyleBearer:
		return fetchEmailWithRequest(ctx, ui.SecondaryEmailURL, headerName, httpsling.BearerAuthHeader+accessToken)
	case AuthHeaderStyleToken:
		return fetchEmailWithRequest(ctx, ui.SecondaryEmailURL, headerName, "token "+accessToken)
	}
	return fetchEmailWithRequest(ctx, ui.SecondaryEmailURL, "", "")
}

func fetchEmailWithRequest(ctx context.Context, url, headerName, headerValue string) (string, error) {
	opts := []httpsling.Option{
		httpsling.Get(url),
	}
	if headerName != "" && headerValue != "" {
		opts = append(opts, httpsling.Header(headerName, headerValue))
	}
	resp, err := sendHTTPRequest(ctx, opts...)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode/100 != runtimeHTTPStatusSuccess {
		return "", fmt.Errorf("%w: status %d", errSecondaryEmailFailure, resp.StatusCode)
	}
	var arr []struct {
		Email    string `json:"email"`
		Primary  bool   `json:"primary"`
		Verified bool   `json:"verified"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&arr); err != nil {
		return "", err
	}
	for _, e := range arr {
		if e.Primary && e.Verified {
			return e.Email, nil
		}
	}
	if len(arr) > 0 {
		return arr[0].Email, nil
	}
	return "", nil
}

// NoopValidator returns an empty IntegrationUserInfo. Useful for non-OAuth flows.
type NoopValidator struct{}

func (NoopValidator) Validate(_ context.Context, _ string, _ *ProviderRuntime) (*IntegrationUserInfo, error) {
	return &IntegrationUserInfo{}, nil
}

// ----- Provider-specific validators (GitHub/Slack), optional -----

type GitHubValidator struct{}

func (GitHubValidator) Validate(ctx context.Context, accessToken string, rt *ProviderRuntime) (*IntegrationUserInfo, error) {
	// Reuse DefaultUserInfoValidator with token style "token"
	orig := rt.Spec.UserInfo
	if orig == nil {
		return nil, errMissingUserInfoSpec
	}
	ui := *orig
	ui.AuthStyle = AuthHeaderStyleToken
	rt2 := *rt
	rt2.Spec.UserInfo = &ui
	return DefaultUserInfoValidator{}.Validate(ctx, accessToken, &rt2)
}

type SlackValidator struct{}

func (SlackValidator) Validate(ctx context.Context, accessToken string, rt *ProviderRuntime) (*IntegrationUserInfo, error) {
	// Slack classic 'users.identity' or OIDC 'userinfo'
	return DefaultUserInfoValidator{}.Validate(ctx, accessToken, rt)
}
