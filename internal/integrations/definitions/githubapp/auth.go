package githubapp

import (
	"context"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	gh "github.com/google/go-github/v83/github"
	"golang.org/x/oauth2"

	"github.com/theopenlane/core/internal/integrations/types"
	"github.com/theopenlane/core/pkg/jsonx"
)

const (
	defaultJWTExpiry           = 9 * time.Minute
	jwtIssuedAtBackdateSeconds = 30 * time.Second
)

// App executes the GitHub App install flow and webhook verification logic
type App struct {
	// Config holds the operator-supplied GitHub App settings
	Config Config
}

// ProviderData is the provider-specific data stored in the credential ProviderData field
type ProviderData struct {
	// AppID is the GitHub App identifier used to mint installation tokens
	AppID string `json:"appId"`
	// InstallationID is the installation selected for this credential
	InstallationID string `json:"installationId"`
}

// MintInstallationCredential exchanges one installation ID for a GitHub App installation token.
func MintInstallationCredential(ctx context.Context, cfg Config, installationID string) (types.CredentialSet, error) {
	if installationID == "" {
		return types.CredentialSet{}, ErrInstallationIDMissing
	}

	a := App{Config: cfg}

	if a.Config.AppID == "" {
		return types.CredentialSet{}, ErrAppIDMissing
	}

	jwtToken, err := a.appJWT(a.Config.PrivateKey)
	if err != nil {
		return types.CredentialSet{}, err
	}

	installationToken, err := a.installationToken(ctx, installationID, jwtToken)
	if err != nil {
		return types.CredentialSet{}, err
	}

	providerData, err := jsonx.ToRawMessage(ProviderData{
		AppID:          a.Config.AppID,
		InstallationID: installationID,
	})
	if err != nil {
		return types.CredentialSet{}, ErrAuthProviderDataEncode
	}

	credential := types.CredentialSet{
		OAuthAccessToken:  installationToken.AccessToken,
		OAuthRefreshToken: installationToken.RefreshToken,
		OAuthTokenType:    installationToken.TokenType,
		ProviderData:      providerData,
	}

	if !installationToken.Expiry.IsZero() {
		expiry := installationToken.Expiry.UTC()
		credential.OAuthExpiry = &expiry
	}

	return credential, nil
}

// appJWT signs a short-lived JWT for GitHub App authentication
func (a App) appJWT(privateKey string) (string, error) {
	key, err := parseRSAPrivateKey(privateKey)
	if err != nil {
		return "", err
	}

	now := time.Now()
	claims := jwt.RegisteredClaims{
		Issuer:    a.Config.AppID,
		IssuedAt:  jwt.NewNumericDate(now.Add(-jwtIssuedAtBackdateSeconds)),
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
func (a App) installationToken(ctx context.Context, installationID string, jwtToken string) (*oauth2.Token, error) {
	if installationID == "" {
		return nil, ErrInstallationIDMissing
	}

	installationIDInt, err := strconv.ParseInt(installationID, 10, 64)
	if err != nil {
		return nil, ErrInstallationTokenRequestFailed
	}

	client := a.installationTokenClient(ctx, jwtToken)
	installationToken, _, err := client.Apps.CreateInstallationToken(ctx, installationIDInt, &gh.InstallationTokenOptions{})
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
func (a App) installationTokenClient(ctx context.Context, jwtToken string) *gh.Client {
	source := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: jwtToken})
	httpClient := oauth2.NewClient(ctx, source)
	client := gh.NewClient(httpClient)

	if a.Config.APIURL == "" {
		return client
	}

	apiURL, err := url.Parse(strings.TrimRight(a.Config.APIURL, "/") + "/api/v3/")
	if err != nil {
		return client
	}

	uploadURL, err := url.Parse(strings.TrimRight(a.Config.APIURL, "/") + "/api/uploads/")
	if err != nil {
		return client
	}

	client.BaseURL = apiURL
	client.UploadURL = uploadURL

	return client
}
