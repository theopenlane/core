package auth

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"

	"golang.org/x/oauth2"

	"github.com/theopenlane/core/internal/integrations/types"
)

type testCredential struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

var errDecodeTestCredential = errors.New("test: decode credential failed")

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
