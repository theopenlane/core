package githubapp

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/hex"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	gh "github.com/google/go-github/v84/github"
	"golang.org/x/oauth2"

	"github.com/theopenlane/core/internal/integrations/types"
	"github.com/theopenlane/core/pkg/jsonx"
)

const (
	// defaultJWTExpiry is the default expiry duration for GitHub App JWTs (max 10 minutes per GitHub docs)
	defaultJWTExpiry = 9 * time.Minute
	// jwtIssuedAtBackdate is the amount of time to backdate the JWT iat claim to account for clock skew
	jwtIssuedAtBackdate = 30 * time.Second
	// stateTokenBytes is the number of random bytes used for CSRF state tokens
	stateTokenBytes = 16
	// installURLTemplate is the GitHub App installation URL pattern used to construct the install redirect
	installURLTemplate = "https://github.com/apps/%s/installations/new"
)

// disconnectDetails holds the metadata returned when initiating a GitHub App disconnect
type disconnectDetails struct {
	// InstallationID is the GitHub App installation ID being disconnected
	InstallationID string `json:"installationId,omitempty"`
}

// statePayload is the opaque state token round-tripped through the auth flow
type statePayload struct {
	// Token is the CSRF state token
	Token string `json:"token"`
}

// startAppInstall generates the GitHub App installation URL and CSRF state token
func startAppInstall(cfg Config) (types.AuthStartResult, error) {
	if cfg.AppSlug == "" {
		return types.AuthStartResult{}, ErrAppSlugMissing
	}

	stateToken, err := generateStateToken()
	if err != nil {
		return types.AuthStartResult{}, err
	}

	stateJSON, err := jsonx.ToRawMessage(statePayload{Token: stateToken})
	if err != nil {
		return types.AuthStartResult{}, ErrAuthStateEncode
	}

	installURL := fmt.Sprintf(installURLTemplate, url.PathEscape(cfg.AppSlug)) + "?state=" + url.QueryEscape(stateToken)

	return types.AuthStartResult{
		URL:   installURL,
		State: stateJSON,
	}, nil
}

// completeAppInstall validates the callback state, exchanges the installation ID for a token, and returns the auth result
func completeAppInstall(ctx context.Context, cfg Config, state json.RawMessage, input types.AuthCallbackInput) (types.AuthCompleteResult, error) {
	var savedState statePayload
	if err := jsonx.UnmarshalIfPresent(state, &savedState); err != nil {
		return types.AuthCompleteResult{}, ErrAuthStateDecode
	}

	callbackState := input.First("state")
	if savedState.Token != "" && callbackState != "" && savedState.Token != callbackState {
		return types.AuthCompleteResult{}, ErrAuthStateMismatch
	}

	raw := input.First("installation_id")
	if raw == "" {
		return types.AuthCompleteResult{}, ErrInstallationIDMissing
	}

	integrationID, err := strconv.ParseInt(raw, 10, 64)
	if err != nil || integrationID == 0 {
		return types.AuthCompleteResult{}, ErrInstallationIDMissing
	}

	cred, err := mintCredential(ctx, cfg, integrationID)
	if err != nil {
		return types.AuthCompleteResult{}, err
	}

	installInput, err := jsonx.ToRawMessage(InstallationMetadata{
		InstallationID: strconv.FormatInt(integrationID, 10),
	})
	if err != nil {
		return types.AuthCompleteResult{}, ErrInstallationMetadataEncode
	}

	return types.AuthCompleteResult{
		Credential:        cred,
		InstallationInput: installInput,
	}, nil
}

// disconnectInstallationID extracts the installation ID from the credential or installation metadata
func disconnectInstallationID(req types.DisconnectRequest) (int64, error) {
	cred, ok, err := gitHubAppCredential.Resolve(req.Credentials)
	if err != nil {
		return 0, ErrCredentialDecode
	}

	if ok && cred.InstallationID != 0 {
		return cred.InstallationID, nil
	}

	var metadata InstallationMetadata
	if err := jsonx.UnmarshalIfPresent(req.Integration.InstallationMetadata.Attributes, &metadata); err != nil {
		return 0, ErrInstallationMetadataDecode
	}

	if metadata.InstallationID == "" {
		return 0, ErrInstallationIDMissing
	}

	integrationID, err := strconv.ParseInt(metadata.InstallationID, 10, 64)
	if err != nil || integrationID == 0 {
		return 0, ErrInstallationIDMissing
	}

	return integrationID, nil
}

// mintCredential mints an installation token and marshals it into a CredentialSet
func mintCredential(ctx context.Context, cfg Config, integrationID int64) (types.CredentialSet, error) {
	if cfg.AppID == "" {
		return types.CredentialSet{}, ErrAppIDMissing
	}

	jwtToken, err := appJWT(cfg)
	if err != nil {
		return types.CredentialSet{}, err
	}

	token, err := installationToken(ctx, cfg, integrationID, jwtToken)
	if err != nil {
		return types.CredentialSet{}, err
	}

	appIDInt, err := strconv.ParseInt(cfg.AppID, 10, 64)
	if err != nil {
		return types.CredentialSet{}, ErrAppIDMissing
	}

	cred := githubAppCredential{
		AppID:          appIDInt,
		InstallationID: integrationID,
		AccessToken:    token.AccessToken,
	}

	if !token.Expiry.IsZero() {
		expiry := token.Expiry.UTC()
		cred.Expiry = &expiry
	}

	data, err := jsonx.ToRawMessage(cred)
	if err != nil {
		return types.CredentialSet{}, ErrCredentialEncode
	}

	return types.CredentialSet{Data: data}, nil
}

// appJWT signs a short-lived JWT for GitHub App authentication
func appJWT(cfg Config) (string, error) {
	key, err := parseRSAPrivateKey(cfg.PrivateKey)
	if err != nil {
		return "", err
	}

	now := time.Now()
	claims := jwt.RegisteredClaims{
		Issuer:    cfg.AppID,
		IssuedAt:  jwt.NewNumericDate(now.Add(-jwtIssuedAtBackdate)),
		ExpiresAt: jwt.NewNumericDate(now.Add(defaultJWTExpiry)),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	signed, err := token.SignedString(key)
	if err != nil {
		return "", ErrJWTSigningFailed
	}

	return signed, nil
}

// parseRSAPrivateKey parses PKCS#1 or PKCS#8 encoded RSA private keys
func parseRSAPrivateKey(privateKey string) (*rsa.PrivateKey, error) {
	block, _ := pem.Decode([]byte(privateKey))
	if block == nil {
		return nil, ErrPrivateKeyInvalid
	}

	if key, err := x509.ParsePKCS1PrivateKey(block.Bytes); err == nil {
		return key, nil
	}

	key, err := x509.ParsePKCS8PrivateKey(block.Bytes)
	if err != nil {
		return nil, ErrPrivateKeyInvalid
	}

	rsaKey, ok := key.(*rsa.PrivateKey)
	if !ok {
		return nil, ErrPrivateKeyInvalid
	}

	return rsaKey, nil
}

// installationToken exchanges an app JWT for an installation access token
func installationToken(ctx context.Context, cfg Config, integrationID int64, jwtToken string) (*oauth2.Token, error) {
	if integrationID == 0 {
		return nil, ErrInstallationIDMissing
	}

	client := installationTokenClient(ctx, cfg, jwtToken)
	installationToken, _, err := client.Apps.CreateInstallationToken(ctx, integrationID, &gh.InstallationTokenOptions{})
	if err != nil {
		return nil, ErrInstallationTokenRequestFailed
	}

	if installationToken.GetToken() == "" {
		return nil, ErrInstallationTokenEmpty
	}

	token := &oauth2.Token{
		AccessToken: installationToken.GetToken(),
		TokenType:   "Bearer",
	}

	expiresAt := installationToken.GetExpiresAt().Time
	if !expiresAt.IsZero() {
		token.Expiry = expiresAt
	}

	return token, nil
}

// installationTokenClient builds the GitHub API client used for installation token requests
func installationTokenClient(ctx context.Context, cfg Config, jwtToken string) *gh.Client {
	source := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: jwtToken})
	httpClient := oauth2.NewClient(ctx, source)
	client := gh.NewClient(httpClient)

	if cfg.APIURL == "" {
		return client
	}

	apiURL, err := url.Parse(strings.TrimRight(cfg.APIURL, "/") + "/api/v3/")
	if err != nil {
		return client
	}

	uploadURL, err := url.Parse(strings.TrimRight(cfg.APIURL, "/") + "/api/uploads/")
	if err != nil {
		return client
	}

	client.BaseURL = apiURL
	client.UploadURL = uploadURL

	return client
}

// generateStateToken produces a cryptographically random hex string for CSRF state
func generateStateToken() (string, error) {
	b := make([]byte, stateTokenBytes)
	if _, err := rand.Read(b); err != nil {
		return "", ErrAuthStateGenerate
	}

	return hex.EncodeToString(b), nil
}
