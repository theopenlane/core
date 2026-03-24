package auth

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"

	"golang.org/x/oauth2"
	"github.com/zitadel/oidc/v3/pkg/client/rp"
	"github.com/zitadel/oidc/v3/pkg/oidc"

	"github.com/theopenlane/core/internal/integrations/types"
)

type testCredential struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

var errDecodeTestCredential = errors.New("test: decode credential failed")

// unmarshalableCredential always fails json.Marshal
type unmarshalableCredential struct {
	Bad chan struct{} `json:"bad"`
}

var errEncodeTestCredential = errors.New("test: encode credential failed")

func TestOAuthRegistrationRefreshNilNoOp(t *testing.T) {
	t.Parallel()

	reg := OAuthRegistration(OAuthRegistrationOptions[testCredential]{
		CredentialRef: types.NewCredentialRef[testCredential]("test"),
		Config:        testOAuthCfg,
		Material: func(mat OAuthMaterial) (testCredential, error) {
			return testCredential{AccessToken: mat.AccessToken}, nil
		},
	})

	stored := types.CredentialSet{Data: json.RawMessage(`{"access_token":"stored"}`)}

	result, err := reg.Refresh(context.Background(), stored)
	if err != nil {
		t.Fatalf("Refresh() error = %v", err)
	}

	if string(result.Data) != string(stored.Data) {
		t.Fatalf("Refresh() = %s, want %s", result.Data, stored.Data)
	}
}

func TestOAuthRegistrationRefreshCallsRefreshFunc(t *testing.T) {
	t.Parallel()

	reg := OAuthRegistration(OAuthRegistrationOptions[testCredential]{
		CredentialRef: types.NewCredentialRef[testCredential]("test"),
		Config:        testOAuthCfg,
		Material: func(mat OAuthMaterial) (testCredential, error) {
			return testCredential{AccessToken: mat.AccessToken}, nil
		},
		Refresh: func(_ context.Context, _ OAuthConfig, cred testCredential) (testCredential, error) {
			return testCredential{AccessToken: "refreshed-" + cred.AccessToken}, nil
		},
	})

	stored := types.CredentialSet{Data: json.RawMessage(`{"access_token":"original","refresh_token":"rt"}`)}

	result, err := reg.Refresh(context.Background(), stored)
	if err != nil {
		t.Fatalf("Refresh() error = %v", err)
	}

	var out testCredential
	if err := json.Unmarshal(result.Data, &out); err != nil {
		t.Fatalf("Unmarshal() error = %v", err)
	}

	if out.AccessToken != "refreshed-original" {
		t.Fatalf("Refresh() access_token = %q, want %q", out.AccessToken, "refreshed-original")
	}
}

func TestOAuthRegistrationRefreshBadStoredCredential(t *testing.T) {
	t.Parallel()

	reg := OAuthRegistration(OAuthRegistrationOptions[testCredential]{
		CredentialRef:         types.NewCredentialRef[testCredential]("test"),
		Config:                testOAuthCfg,
		DecodeCredentialError: errDecodeTestCredential,
		Material: func(mat OAuthMaterial) (testCredential, error) {
			return testCredential{}, nil
		},
		Refresh: func(_ context.Context, _ OAuthConfig, cred testCredential) (testCredential, error) {
			return cred, nil
		},
	})

	stored := types.CredentialSet{Data: json.RawMessage(`{invalid`)}

	_, err := reg.Refresh(context.Background(), stored)
	if !errors.Is(err, errDecodeTestCredential) {
		t.Fatalf("Refresh() error = %v, want %v", err, errDecodeTestCredential)
	}
}

func TestOAuthRegistrationTokenViewReturnsToken(t *testing.T) {
	t.Parallel()

	expiry := time.Now().Add(time.Hour).UTC()

	reg := OAuthRegistration(OAuthRegistrationOptions[testCredential]{
		CredentialRef: types.NewCredentialRef[testCredential]("test"),
		Config:        testOAuthCfg,
		Material: func(mat OAuthMaterial) (testCredential, error) {
			return testCredential{}, nil
		},
		TokenView: func(cred testCredential) (*types.TokenView, error) {
			return &types.TokenView{
				AccessToken: cred.AccessToken,
				ExpiresAt:   &expiry,
			}, nil
		},
	})

	stored := types.CredentialSet{Data: json.RawMessage(`{"access_token":"at-123"}`)}
	view, err := reg.TokenView(context.Background(), stored)
	if err != nil {
		t.Fatalf("TokenView() error = %v", err)
	}

	if view == nil {
		t.Fatal("TokenView() returned nil")
	}

	if view.AccessToken != "at-123" {
		t.Fatalf("TokenView() token = %q, want %q", view.AccessToken, "at-123")
	}
}

func TestOAuthRegistrationTokenViewNilFunc(t *testing.T) {
	t.Parallel()

	reg := OAuthRegistration(OAuthRegistrationOptions[testCredential]{
		CredentialRef: types.NewCredentialRef[testCredential]("test"),
		Config:        testOAuthCfg,
		Material: func(mat OAuthMaterial) (testCredential, error) {
			return testCredential{}, nil
		},
	})

	stored := types.CredentialSet{Data: json.RawMessage(`{"access_token":"at-123"}`)}
	view, err := reg.TokenView(context.Background(), stored)
	if err != nil {
		t.Fatalf("TokenView() error = %v", err)
	}

	if view != nil {
		t.Fatalf("TokenView() = %v, want nil", view)
	}
}

func TestBuildOAuthMaterialToken(t *testing.T) {
	t.Parallel()

	expiry := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
	token := &oauth2.Token{
		AccessToken:  "access",
		RefreshToken: "refresh",
		Expiry:       expiry,
	}

	mat, err := buildOAuthMaterial(token, nil)
	if err != nil {
		t.Fatalf("buildOAuthMaterial() error = %v", err)
	}

	if mat.AccessToken != "access" {
		t.Fatalf("AccessToken = %q, want %q", mat.AccessToken, "access")
	}

	if mat.RefreshToken != "refresh" {
		t.Fatalf("RefreshToken = %q, want %q", mat.RefreshToken, "refresh")
	}

	if mat.Expiry == nil || !mat.Expiry.Equal(expiry) {
		t.Fatalf("Expiry = %v, want %v", mat.Expiry, expiry)
	}

	if mat.Claims != nil {
		t.Fatalf("Claims = %v, want nil", mat.Claims)
	}
}

func TestBuildOAuthMaterialNilToken(t *testing.T) {
	t.Parallel()

	mat, err := buildOAuthMaterial(nil, nil)
	if err != nil {
		t.Fatalf("buildOAuthMaterial() error = %v", err)
	}

	if mat.AccessToken != "" {
		t.Fatalf("AccessToken = %q, want empty", mat.AccessToken)
	}

	if mat.Expiry != nil {
		t.Fatalf("Expiry = %v, want nil", mat.Expiry)
	}
}

func TestBuildOAuthMaterialZeroExpiry(t *testing.T) {
	t.Parallel()

	token := &oauth2.Token{
		AccessToken: "access",
	}

	mat, err := buildOAuthMaterial(token, nil)
	if err != nil {
		t.Fatalf("buildOAuthMaterial() error = %v", err)
	}

	if mat.Expiry != nil {
		t.Fatalf("Expiry = %v, want nil for zero time", mat.Expiry)
	}
}

func TestBuildAuthURLOpts(t *testing.T) {
	t.Parallel()

	opts := buildAuthURLOpts(map[string]string{"prompt": "consent", "access_type": "offline"})
	if len(opts) != 2 {
		t.Fatalf("buildAuthURLOpts() returned %d opts, want 2", len(opts))
	}
}

func TestBuildAuthURLOptsNil(t *testing.T) {
	t.Parallel()

	opts := buildAuthURLOpts(nil)
	if opts != nil {
		t.Fatalf("buildAuthURLOpts(nil) = %v, want nil", opts)
	}
}

func TestBuildCodeExchangeOpts(t *testing.T) {
	t.Parallel()

	opts := buildCodeExchangeOpts(map[string]string{"grant_type": "authorization_code"})
	if len(opts) != 1 {
		t.Fatalf("buildCodeExchangeOpts() returned %d opts, want 1", len(opts))
	}
}

func TestBuildOAuthMaterialWithClaims(t *testing.T) {
	t.Parallel()

	claims := &oidc.IDTokenClaims{}
	claims.Subject = "user-42"

	token := &oauth2.Token{
		AccessToken: "at",
	}

	mat, err := buildOAuthMaterial(token, claims)
	if err != nil {
		t.Fatalf("buildOAuthMaterial() error = %v", err)
	}

	if mat.Claims == nil {
		t.Fatal("buildOAuthMaterial() Claims is nil, want non-nil")
	}

	sub, ok := mat.Claims["sub"]
	if !ok {
		t.Fatal("buildOAuthMaterial() Claims missing 'sub' key")
	}

	if sub != "user-42" {
		t.Fatalf("buildOAuthMaterial() Claims[sub] = %v, want %q", sub, "user-42")
	}
}

func TestAsAuthCodeOptionReturnsOption(t *testing.T) {
	t.Parallel()

	param := oauth2.SetAuthURLParam("prompt", "consent")
	opt := asAuthCodeOption[rp.AuthURLOpt](param)

	options := opt()
	if len(options) != 1 {
		t.Fatalf("asAuthCodeOption() returned %d options, want 1", len(options))
	}
}

func TestOAuthRegistrationStartDelegates(t *testing.T) {
	t.Parallel()

	reg := OAuthRegistration(OAuthRegistrationOptions[testCredential]{
		CredentialRef: types.NewCredentialRef[testCredential]("test"),
		Config:        testOAuthCfg,
		Material: func(mat OAuthMaterial) (testCredential, error) {
			return testCredential{AccessToken: mat.AccessToken}, nil
		},
	})

	result, err := reg.Start(context.Background(), nil)
	if err != nil {
		t.Fatalf("Start() error = %v", err)
	}

	if result.URL == "" {
		t.Fatal("Start() URL is empty")
	}
}

func TestOAuthRegistrationCompleteCodeExchangeError(t *testing.T) {
	t.Parallel()

	reg := OAuthRegistration(OAuthRegistrationOptions[testCredential]{
		CredentialRef: types.NewCredentialRef[testCredential]("test"),
		Config:        testOAuthCfg,
		Material: func(mat OAuthMaterial) (testCredential, error) {
			return testCredential{AccessToken: mat.AccessToken}, nil
		},
	})

	// Valid state but code exchange will fail (no real OAuth server)
	state := json.RawMessage(`{}`)
	input := types.AuthCallbackInput{
		Query: []types.AuthCallbackValue{
			{Name: "code", Values: []string{"auth-code"}},
		},
	}

	_, err := reg.Complete(context.Background(), state, input)
	if err == nil {
		t.Fatal("Complete() expected error from code exchange, got nil")
	}
}

func TestOAuthRegistrationRefreshFuncError(t *testing.T) {
	t.Parallel()

	errRefreshFailed := errors.New("test: refresh failed")

	reg := OAuthRegistration(OAuthRegistrationOptions[testCredential]{
		CredentialRef: types.NewCredentialRef[testCredential]("test"),
		Config:        testOAuthCfg,
		Material: func(mat OAuthMaterial) (testCredential, error) {
			return testCredential{}, nil
		},
		Refresh: func(context.Context, OAuthConfig, testCredential) (testCredential, error) {
			return testCredential{}, errRefreshFailed
		},
	})

	stored := types.CredentialSet{Data: json.RawMessage(`{"access_token":"at"}`)}

	_, err := reg.Refresh(context.Background(), stored)
	if !errors.Is(err, errRefreshFailed) {
		t.Fatalf("Refresh() error = %v, want %v", err, errRefreshFailed)
	}
}

func TestOAuthRegistrationRefreshBadCredentialNoCustomError(t *testing.T) {
	t.Parallel()

	reg := OAuthRegistration(OAuthRegistrationOptions[testCredential]{
		CredentialRef: types.NewCredentialRef[testCredential]("test"),
		Config:        testOAuthCfg,
		Material: func(mat OAuthMaterial) (testCredential, error) {
			return testCredential{}, nil
		},
		Refresh: func(context.Context, OAuthConfig, testCredential) (testCredential, error) {
			return testCredential{}, nil
		},
	})

	stored := types.CredentialSet{Data: json.RawMessage(`{invalid`)}

	_, err := reg.Refresh(context.Background(), stored)
	if err == nil {
		t.Fatal("Refresh() expected error for invalid credential, got nil")
	}
}

func TestOAuthRegistrationTokenViewBadCredentialCustomError(t *testing.T) {
	t.Parallel()

	reg := OAuthRegistration(OAuthRegistrationOptions[testCredential]{
		CredentialRef:         types.NewCredentialRef[testCredential]("test"),
		Config:                testOAuthCfg,
		DecodeCredentialError: errDecodeTestCredential,
		Material: func(mat OAuthMaterial) (testCredential, error) {
			return testCredential{}, nil
		},
		TokenView: func(cred testCredential) (*types.TokenView, error) {
			return &types.TokenView{AccessToken: cred.AccessToken}, nil
		},
	})

	stored := types.CredentialSet{Data: json.RawMessage(`{invalid`)}

	_, err := reg.TokenView(context.Background(), stored)
	if !errors.Is(err, errDecodeTestCredential) {
		t.Fatalf("TokenView() error = %v, want %v", err, errDecodeTestCredential)
	}
}

func TestOAuthRegistrationTokenViewBadCredentialNoCustomError(t *testing.T) {
	t.Parallel()

	reg := OAuthRegistration(OAuthRegistrationOptions[testCredential]{
		CredentialRef: types.NewCredentialRef[testCredential]("test"),
		Config:        testOAuthCfg,
		Material: func(mat OAuthMaterial) (testCredential, error) {
			return testCredential{}, nil
		},
		TokenView: func(cred testCredential) (*types.TokenView, error) {
			return &types.TokenView{AccessToken: cred.AccessToken}, nil
		},
	})

	stored := types.CredentialSet{Data: json.RawMessage(`{invalid`)}

	_, err := reg.TokenView(context.Background(), stored)
	if err == nil {
		t.Fatal("TokenView() expected error for invalid credential, got nil")
	}
}

func TestOAuthRegistrationRefreshEncodeErrorCustom(t *testing.T) {
	t.Parallel()

	reg := OAuthRegistration(OAuthRegistrationOptions[unmarshalableCredential]{
		CredentialRef:          types.NewCredentialRef[unmarshalableCredential]("test"),
		Config:                 testOAuthCfg,
		EncodeCredentialError:  errEncodeTestCredential,
		Material: func(OAuthMaterial) (unmarshalableCredential, error) {
			return unmarshalableCredential{}, nil
		},
		Refresh: func(context.Context, OAuthConfig, unmarshalableCredential) (unmarshalableCredential, error) {
			return unmarshalableCredential{}, nil
		},
	})

	// Valid JSON that decodes to unmarshalableCredential (bad field ignored), but re-encoding fails
	stored := types.CredentialSet{Data: json.RawMessage(`{}`)}

	_, err := reg.Refresh(context.Background(), stored)
	if !errors.Is(err, errEncodeTestCredential) {
		t.Fatalf("Refresh() error = %v, want %v", err, errEncodeTestCredential)
	}
}

func TestOAuthRegistrationRefreshEncodeErrorNoCustom(t *testing.T) {
	t.Parallel()

	reg := OAuthRegistration(OAuthRegistrationOptions[unmarshalableCredential]{
		CredentialRef: types.NewCredentialRef[unmarshalableCredential]("test"),
		Config:        testOAuthCfg,
		Material: func(OAuthMaterial) (unmarshalableCredential, error) {
			return unmarshalableCredential{}, nil
		},
		Refresh: func(context.Context, OAuthConfig, unmarshalableCredential) (unmarshalableCredential, error) {
			return unmarshalableCredential{}, nil
		},
	})

	stored := types.CredentialSet{Data: json.RawMessage(`{}`)}

	_, err := reg.Refresh(context.Background(), stored)
	if err == nil {
		t.Fatal("Refresh() expected encode error, got nil")
	}
}

func TestStartOAuthWithAuthParams(t *testing.T) {
	t.Parallel()

	cfg := OAuthConfig{
		ClientID:     "client-id",
		ClientSecret: "client-secret",
		AuthURL:      "https://example.com/oauth/authorize",
		TokenURL:     "https://example.com/oauth/token",
		RedirectURL:  "https://app.example.com/callback",
		AuthParams:   map[string]string{"prompt": "consent", "access_type": "offline"},
	}

	result, err := StartOAuth(context.Background(), cfg)
	if err != nil {
		t.Fatalf("StartOAuth() error = %v", err)
	}

	if result.URL == "" {
		t.Fatal("StartOAuth() URL is empty")
	}
}

func TestCompleteOAuthEmptyStartStateHitsCodeExchange(t *testing.T) {
	t.Parallel()

	// Empty start state skips CSRF validation but still attempts code exchange
	state := json.RawMessage(`{}`)
	input := types.AuthCallbackInput{
		Query: []types.AuthCallbackValue{
			{Name: "code", Values: []string{"auth-code"}},
		},
	}

	_, err := CompleteOAuth(context.Background(), testOAuthCfg, state, input)
	if !errors.Is(err, ErrOAuthCodeExchange) {
		t.Fatalf("CompleteOAuth() error = %v, want %v", err, ErrOAuthCodeExchange)
	}
}

func TestCompleteOAuthMatchingStateHitsCodeExchange(t *testing.T) {
	t.Parallel()

	state := json.RawMessage(`{"state":"same"}`)
	input := types.AuthCallbackInput{
		Query: []types.AuthCallbackValue{
			{Name: "code", Values: []string{"auth-code"}},
			{Name: "state", Values: []string{"same"}},
		},
	}

	_, err := CompleteOAuth(context.Background(), testOAuthCfg, state, input)
	if !errors.Is(err, ErrOAuthCodeExchange) {
		t.Fatalf("CompleteOAuth() error = %v, want %v", err, ErrOAuthCodeExchange)
	}
}
