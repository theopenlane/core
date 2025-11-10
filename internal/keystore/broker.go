package keystore

import (
	"context"
	"crypto/rsa"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"time"

	iamcredentials "cloud.google.com/go/iam/credentials/apiv1"
	credentialspb "cloud.google.com/go/iam/credentials/apiv1/credentialspb"
	"github.com/aws/aws-sdk-go-v2/aws"
	sts "github.com/aws/aws-sdk-go-v2/service/sts"
	"github.com/golang-jwt/jwt/v5"
	"github.com/googleapis/gax-go/v2"
	"github.com/samber/lo"
	"github.com/theopenlane/httpsling"
	"golang.org/x/oauth2"
	"google.golang.org/api/option"
	"google.golang.org/protobuf/types/known/durationpb"
)

const (
	stsEndpoint              = "https://sts.googleapis.com/v1/token"
	githubAPIBase            = "https://api.github.com"
	defaultCacheBuffer       = 30 * time.Second
	defaultGitHubTokenTTL    = 10 * time.Minute
	defaultWIFTokenTTL       = time.Hour
	httpStatusCodeSuccess    = 2
	defaultJWTExpiryDuration = 10 * time.Minute
	defaultAWSRegion         = "us-east-1"
)

var (
	errBrokerNotConfigured           = errors.New("integration broker not configured")
	errRegistryNotConfigured         = errors.New("integration registry not configured")
	errProviderNotRegistered         = errors.New("provider not registered")
	errRefreshTokenUnavailable       = errors.New("refresh token not available")
	errOAuthConfigMissing            = errors.New("oauth configuration missing for provider")
	errWorkloadIdentityIssuerNil     = errors.New("workload identity subject token issuer not configured")
	errAzureFederationNotImplemented = errors.New("azure federated credential flow not implemented")
)

// Broker exchanges persisted credentials for short-lived access tokens.
type Broker struct {
	store                  tokenStore
	registry               *Registry
	WorkloadIdentityIssuer WorkloadIdentityIssuer

	mu    sync.RWMutex
	cache map[cacheKey]cachedCredential

	httpClient HTTPClient
	iamFactory iamClientFactory
	stsFactory stsClientFactory
	now        func() time.Time
}

type cacheKey struct {
	OrgID    string
	Provider string
}

type cachedCredential struct {
	creds   *Credentials
	expires time.Time
}

type tokenStore interface {
	LoadTokens(ctx context.Context, orgID, provider string) (*TokenBundle, error)
	RecordMint(ctx context.Context, payload OAuthTokens) error
}

type storeAdapter struct {
	inner *Store
}

func (s storeAdapter) LoadTokens(ctx context.Context, orgID, provider string) (*TokenBundle, error) {
	return s.inner.LoadTokens(ctx, orgID, provider)
}

func (s storeAdapter) RecordMint(ctx context.Context, payload OAuthTokens) error {
	_, err := s.inner.UpsertOAuthTokens(ctx, payload)
	return err
}

// HTTPClient defines the subset of http.Client used by the keystore. It enables tests to stub network calls.
type HTTPClient interface {
	Do(req *http.Request) (*http.Response, error)
}

type iamCredentialsClient interface {
	GenerateAccessToken(ctx context.Context, req *credentialspb.GenerateAccessTokenRequest, opts ...gax.CallOption) (*credentialspb.GenerateAccessTokenResponse, error)
	Close() error
}

type iamClientFactory func(ctx context.Context, ts oauth2.TokenSource) (iamCredentialsClient, error)

type stsClient interface {
	AssumeRoleWithWebIdentity(ctx context.Context, params *sts.AssumeRoleWithWebIdentityInput, optFns ...func(*sts.Options)) (*sts.AssumeRoleWithWebIdentityOutput, error)
}

type stsClientFactory func(region string) (stsClient, error)

func defaultIAMFactory(ctx context.Context, ts oauth2.TokenSource) (iamCredentialsClient, error) {
	return iamcredentials.NewIamCredentialsClient(ctx, option.WithTokenSource(ts))
}

func defaultSTSFactory(region string) (stsClient, error) {
	cfg := aws.Config{
		Region:      region,
		Credentials: aws.AnonymousCredentials{},
	}
	return sts.NewFromConfig(cfg), nil
}

// WorkloadIdentitySubjectToken captures the subject token used with Google STS.
type WorkloadIdentitySubjectToken struct {
	Token string
	Type  string
}

// WorkloadIdentityIssuer issues subject tokens for Workload Identity Federation exchanges.
type WorkloadIdentityIssuer interface {
	IssueSubjectToken(ctx context.Context, orgID string, spec *ProviderSpec, attributes map[string]string) (*WorkloadIdentitySubjectToken, error)
}

// WorkloadIdentityIssuerFunc turns a function into a WorkloadIdentityIssuer.
type WorkloadIdentityIssuerFunc func(context.Context, string, *ProviderSpec, map[string]string) (*WorkloadIdentitySubjectToken, error)

func (f WorkloadIdentityIssuerFunc) IssueSubjectToken(ctx context.Context, orgID string, spec *ProviderSpec, attrs map[string]string) (*WorkloadIdentitySubjectToken, error) {
	return f(ctx, orgID, spec, attrs)
}

// New returns a broker backed by the provided store and registry.
func New(store *Store, registry *Registry) *Broker {
	var adapted tokenStore
	if store != nil {
		adapted = storeAdapter{inner: store}
	}
	return newBroker(adapted, registry)
}

func newBroker(store tokenStore, registry *Registry) *Broker {
	b := &Broker{
		store:      store,
		registry:   registry,
		cache:      make(map[cacheKey]cachedCredential),
		httpClient: http.DefaultClient,
		iamFactory: defaultIAMFactory,
		stsFactory: defaultSTSFactory,
		now:        time.Now,
	}
	return b
}

// WithWorkloadIdentityIssuer attaches a subject token issuer for WIF flows.
func (b *Broker) WithWorkloadIdentityIssuer(issuer WorkloadIdentityIssuer) *Broker {
	b.WorkloadIdentityIssuer = issuer
	return b
}

// WithHTTPClient overrides the HTTP client used for outbound calls (STS, GitHub, etc.).
func (b *Broker) WithHTTPClient(client HTTPClient) *Broker {
	if client == nil {
		b.httpClient = http.DefaultClient
		return b
	}
	b.httpClient = client
	return b
}

// WithIAMClientFactory overrides the factory used to construct IAM credentials clients.
func (b *Broker) WithIAMClientFactory(factory iamClientFactory) *Broker {
	if factory == nil {
		b.iamFactory = defaultIAMFactory
		return b
	}
	b.iamFactory = factory
	return b
}

// WithSTSClientFactory overrides the factory used to construct AWS STS clients.
func (b *Broker) WithSTSClientFactory(factory stsClientFactory) *Broker {
	if factory == nil {
		b.stsFactory = defaultSTSFactory
		return b
	}
	b.stsFactory = factory
	return b
}

// WithClock overrides the clock used by the broker (useful for deterministic tests).
func (b *Broker) WithClock(now func() time.Time) *Broker {
	if now == nil {
		b.now = time.Now
		return b
	}
	b.now = now
	return b
}

// Credentials represents the brokered access token result.
type Credentials struct {
	Provider         string            `json:"provider"`
	AccessToken      string            `json:"accessToken"`
	RefreshToken     string            `json:"refreshToken,omitempty"`
	ExpiresAt        *time.Time        `json:"expiresAt,omitempty"`
	ProviderUserID   string            `json:"providerUserId,omitempty"`
	ProviderUsername string            `json:"providerUsername,omitempty"`
	ProviderEmail    string            `json:"providerEmail,omitempty"`
	Metadata         map[string]string `json:"metadata,omitempty"`
}

// MintOAuthToken fetches or mints credentials for the given provider.
func (b *Broker) MintOAuthToken(ctx context.Context, orgID, provider string) (*Credentials, error) {
	if b.store == nil || b.registry == nil {
		return nil, errBrokerNotConfigured
	}

	providerKey := strings.ToLower(provider)

	if creds := b.getCachedCredentials(orgID, providerKey); creds != nil {
		return creds, nil
	}

	rt, err := b.runtimeForProvider(providerKey)
	if err != nil {
		return nil, err
	}

	bundle, err := b.store.LoadTokens(ctx, orgID, providerKey)
	if err != nil {
		return nil, err
	}

	var creds *Credentials

	switch rt.Spec.AuthType {
	case AuthTypeOAuth2, AuthTypeOIDC:
		creds, err = b.refreshOAuth(ctx, orgID, providerKey, rt, bundle)
	case AuthTypeWorkloadIdentity:
		creds, err = b.mintWorkloadIdentity(ctx, orgID, providerKey, rt, bundle)
	case AuthTypeGitHubApp:
		creds, err = b.mintGitHubAppInstallationToken(ctx, orgID, providerKey, rt, bundle)
	case AuthTypeAPIKey:
		err = errAPIKeyMintingUnsupported
	case AuthTypeAWSSTS:
		creds, err = b.mintAWSFederation(ctx, orgID, providerKey, rt, bundle)
	case AuthTypeAzureFederated:
		creds, err = b.mintAzureFederation(ctx, orgID, providerKey, rt, bundle)
	default:
		err = fmt.Errorf("%w: %s", errUnsupportedAuthType, rt.Spec.AuthType)
	}

	if err != nil {
		return nil, err
	}

	if creds == nil {
		return nil, errNoCredentialsProduced
	}

	creds.Provider = providerKey
	b.setCachedCredentials(orgID, providerKey, creds)
	return creds, nil
}

func (b *Broker) refreshOAuth(ctx context.Context, orgID, provider string, rt *ProviderRuntime, bundle *TokenBundle) (*Credentials, error) {
	if rt.OAuthConfig == nil {
		return nil, errOAuthConfigMissing
	}
	if bundle.RefreshToken == "" {
		return nil, errRefreshTokenUnavailable
	}

	token := &oauth2.Token{
		AccessToken:  bundle.AccessToken,
		RefreshToken: bundle.RefreshToken,
	}
	if bundle.ExpiresAt != nil {
		token.Expiry = *bundle.ExpiresAt
	}

	freshToken, err := rt.OAuthConfig.TokenSource(ctx, token).Token()
	if err != nil {
		return nil, err
	}

	userInfo, err := rt.Validator.Validate(ctx, freshToken.AccessToken, rt)
	if err != nil {
		return nil, err
	}

	payload := OAuthTokens{
		OrgID:             orgID,
		Provider:          provider,
		Username:          userInfo.Username,
		UserID:            userInfo.ID,
		Email:             userInfo.Email,
		AccessToken:       freshToken.AccessToken,
		RefreshToken:      freshToken.RefreshToken,
		StoreRefreshToken: shouldStoreRefresh(rt),
		Attributes: map[string]string{
			ProviderUserIDField:   userInfo.ID,
			ProviderUsernameField: userInfo.Username,
			ProviderEmailField:    userInfo.Email,
		},
	}
	if !freshToken.Expiry.IsZero() {
		exp := freshToken.Expiry
		payload.ExpiresAt = &exp
	}

	additionalAttributes := lo.SliceToMap(
		lo.Filter(lo.Entries(bundle.Attributes), func(entry lo.Entry[string, string], _ int) bool {
			if entry.Value == "" {
				return false
			}
			_, exists := payload.Attributes[entry.Key]
			return !exists
		}),
		func(entry lo.Entry[string, string]) (string, string) {
			return entry.Key, entry.Value
		},
	)
	payload.Attributes = lo.Assign(payload.Attributes, additionalAttributes)
	if len(bundle.Scopes) > 0 {
		payload.Scopes = append([]string(nil), bundle.Scopes...)
	}
	if len(bundle.Metadata) > 0 {
		payload.Metadata = lo.Assign(map[string]any{}, bundle.Metadata)
	}

	if err := b.store.RecordMint(ctx, payload); err != nil {
		return nil, err
	}

	creds := &Credentials{
		AccessToken:      freshToken.AccessToken,
		RefreshToken:     freshToken.RefreshToken,
		ExpiresAt:        payload.ExpiresAt,
		ProviderUserID:   userInfo.ID,
		ProviderUsername: userInfo.Username,
		ProviderEmail:    userInfo.Email,
		Metadata:         lo.Assign(map[string]string{}, payload.Attributes),
	}

	return creds, nil
}

func (b *Broker) mintGitHubAppInstallationToken(ctx context.Context, _ string, _ string, rt *ProviderRuntime, bundle *TokenBundle) (*Credentials, error) {
	privateKey := bundle.Attributes["private_key"]
	installationID := bundle.Attributes["installation_id"]
	appID := bundle.Attributes["app_id"]

	if privateKey == "" || installationID == "" || appID == "" {
		return nil, errMissingGitHubAppCredentials
	}

	appIDInt, err := strconv.ParseInt(appID, 10, 64)
	if err != nil {
		return nil, fmt.Errorf("invalid app_id: %w", err)
	}
	installationIDInt, err := strconv.ParseInt(installationID, 10, 64)
	if err != nil {
		return nil, fmt.Errorf("invalid installation_id: %w", err)
	}

	jwtToken, err := signGitHubAppJWT(appIDInt, privateKey)
	if err != nil {
		return nil, err
	}

	baseURL := githubAPIBase
	if rt.Spec.GitHubApp != nil && rt.Spec.GitHubApp.BaseURL != "" {
		baseURL = strings.TrimSuffix(rt.Spec.GitHubApp.BaseURL, "/")
	}

	requester, err := b.newRequester(httpsling.URL(baseURL))
	if err != nil {
		return nil, err
	}

	path := fmt.Sprintf("/app/installations/%d/access_tokens", installationIDInt)
	resp, err := requester.SendWithContext(
		ctx,
		httpsling.Post(path),
		httpsling.HeadersFromMap(map[string]string{
			"Authorization": "Bearer " + jwtToken,
			"Accept":        "application/vnd.github+json",
		}),
	)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode/100 != httpStatusCodeSuccess {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("%w: %s", errGitHubInstallationTokenExchange, strings.TrimSpace(string(body)))
	}

	var payload struct {
		Token     string    `json:"token"`
		ExpiresAt time.Time `json:"expires_at"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return nil, err
	}

	expires := payload.ExpiresAt
	if expires.IsZero() {
		expires = b.now().Add(defaultGitHubTokenTTL)
	}

	creds := &Credentials{
		AccessToken: payload.Token,
		ExpiresAt:   &expires,
		Metadata: map[string]string{
			"installation_id": installationID,
			"app_id":          appID,
		},
	}

	return creds, nil
}

func (b *Broker) mintWorkloadIdentity(ctx context.Context, orgID string, _ string, rt *ProviderRuntime, bundle *TokenBundle) (*Credentials, error) {
	if b.WorkloadIdentityIssuer == nil {
		return nil, errWorkloadIdentityIssuerNil
	}

	attrs := bundle.Attributes

	audience := attrs["audience"]
	if audience == "" && rt.Spec.WorkloadIdentity != nil {
		audience = rt.Spec.WorkloadIdentity.Audience
	}
	if audience == "" {
		return nil, errWorkloadIdentityAudienceMissing
	}

	subjectToken, err := b.WorkloadIdentityIssuer.IssueSubjectToken(ctx, orgID, &rt.Spec, attrs)
	if err != nil {
		return nil, err
	}

	subjectTokenType := "urn:ietf:params:oauth:token-type:id_token" //nolint:gosec
	if rt.Spec.WorkloadIdentity != nil && rt.Spec.WorkloadIdentity.SubjectTokenType != "" {
		subjectTokenType = rt.Spec.WorkloadIdentity.SubjectTokenType
	}
	if subjectToken != nil && subjectToken.Type != "" {
		subjectTokenType = subjectToken.Type
	}

	values := url.Values{}
	values.Set("grant_type", "urn:ietf:params:oauth:grant-type:token-exchange")
	values.Set("audience", audience)
	values.Set("subject_token", subjectToken.Token)
	values.Set("subject_token_type", subjectTokenType)

	scopes := extractScopes(rt, attrs)
	if len(scopes) > 0 {
		values.Set("scope", strings.Join(scopes, " "))
	}

	requester, err := b.newRequester()
	if err != nil {
		return nil, err
	}

	resp, err := requester.SendWithContext(
		ctx,
		httpsling.Post(stsEndpoint),
		httpsling.Form(),
		httpsling.Body(values),
	)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode/100 != httpStatusCodeSuccess {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("%w: %s", errSTSTokenExchange, strings.TrimSpace(string(body)))
	}

	var stsResponse struct {
		AccessToken string `json:"access_token"`
		ExpiresIn   int64  `json:"expires_in"`
		TokenType   string `json:"token_type"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&stsResponse); err != nil {
		return nil, err
	}

	// Impersonate the target service account using the exchanged token.
	serviceAccount := attrs["serviceAccountEmail"]
	if serviceAccount == "" && rt.Spec.WorkloadIdentity != nil {
		serviceAccount = rt.Spec.WorkloadIdentity.TargetServiceAccount
	}
	if serviceAccount == "" {
		return nil, errTargetServiceAccountEmailMissing
	}

	tokenLifetime := defaultWIFTokenTTL
	if rt.Spec.WorkloadIdentity != nil && rt.Spec.WorkloadIdentity.TokenLifetime > 0 {
		tokenLifetime = rt.Spec.WorkloadIdentity.TokenLifetime
	}
	if lifetimeString := attrs["tokenLifetime"]; lifetimeString != "" {
		if dur, err := time.ParseDuration(lifetimeString); err == nil {
			tokenLifetime = dur
		}
	}

	oauthToken := &oauth2.Token{
		AccessToken: stsResponse.AccessToken,
		TokenType:   stsResponse.TokenType,
	}
	if stsResponse.ExpiresIn > 0 {
		oauthToken.Expiry = b.now().Add(time.Duration(stsResponse.ExpiresIn) * time.Second)
	}

	ts := oauth2.StaticTokenSource(oauthToken)
	iamClient, err := b.iamFactory(ctx, ts)
	if err != nil {
		return nil, err
	}
	defer iamClient.Close()

	generateReq := &credentialspb.GenerateAccessTokenRequest{
		Name:  fmt.Sprintf("projects/-/serviceAccounts/%s", serviceAccount),
		Scope: scopes,
	}
	if tokenLifetime > 0 {
		generateReq.Lifetime = durationpb.New(tokenLifetime)
	}

	tokenResp, err := iamClient.GenerateAccessToken(ctx, generateReq)
	if err != nil {
		return nil, err
	}

	expireTime := tokenResp.ExpireTime.AsTime()
	metadata := map[string]string{
		"audience":            audience,
		"serviceAccountEmail": serviceAccount,
	}
	for key, value := range attrs {
		if value == "" {
			continue
		}
		metadata[key] = value
	}
	creds := &Credentials{
		AccessToken: tokenResp.AccessToken,
		ExpiresAt:   &expireTime,
		Metadata:    metadata,
	}

	return creds, nil
}

func (b *Broker) mintAWSFederation(ctx context.Context, orgID string, provider string, rt *ProviderRuntime, bundle *TokenBundle) (*Credentials, error) {
	if bundle == nil {
		return nil, fmt.Errorf("%s: %w", provider, errAWSFederationConfigurationMissing)
	}

	attrs := map[string]string{}
	if bundle.Attributes != nil {
		attrs = bundle.Attributes
	}

	if accessKey := strings.TrimSpace(attrs["access_key_id"]); accessKey != "" {
		secretKey := strings.TrimSpace(attrs["secret_access_key"])
		if secretKey == "" {
			return nil, fmt.Errorf("%s: %w", provider, errAWSFederationConfigurationMissing)
		}
		sessionToken := strings.TrimSpace(attrs["session_token"])
		region := resolveAWSRegion(attrs, rt)
		metadata := map[string]string{
			"access_key_id":     accessKey,
			"secret_access_key": secretKey,
			"region":            region,
		}
		if sessionToken != "" {
			metadata["session_token"] = sessionToken
		}
		if role := strings.TrimSpace(attrs["role_arn"]); role != "" {
			metadata["role_arn"] = role
		}
		if sessionName := strings.TrimSpace(attrs["session_name"]); sessionName != "" {
			metadata["session_name"] = sessionName
		}
		return &Credentials{
			AccessToken: sessionToken,
			Metadata:    metadata,
		}, nil
	}

	roleARN := resolveAWSRoleARN(attrs, rt)
	if roleARN == "" {
		return nil, fmt.Errorf("%s: %w", provider, errAWSFederationConfigurationMissing)
	}

	webIdentity := strings.TrimSpace(attrs["web_identity_token"])
	if webIdentity == "" {
		webIdentity = strings.TrimSpace(bundle.AccessToken)
	}
	if webIdentity == "" {
		return nil, fmt.Errorf("%s: %w", provider, errAWSWebIdentityTokenMissing)
	}

	sessionName := strings.TrimSpace(attrs["session_name"])
	if sessionName == "" && rt != nil && rt.Spec.AWSSTS != nil && rt.Spec.AWSSTS.SessionName != "" {
		sessionName = rt.Spec.AWSSTS.SessionName
	}
	if sessionName == "" {
		sessionName = fmt.Sprintf("openlane-%s", orgID)
	}

	durationSeconds := int32(3600)
	if rt != nil && rt.Spec.AWSSTS != nil && rt.Spec.AWSSTS.Duration > 0 {
		durationSeconds = int32(rt.Spec.AWSSTS.Duration / time.Second)
	}
	if raw := strings.TrimSpace(attrs["session_duration"]); raw != "" {
		if val, err := parseDurationSeconds(raw); err == nil && val > 0 {
			durationSeconds = val
		}
	}

	region := resolveAWSRegion(attrs, rt)
	factory := b.stsFactory
	if factory == nil {
		factory = defaultSTSFactory
	}
	client, err := factory(region)
	if err != nil {
		return nil, err
	}

	input := &sts.AssumeRoleWithWebIdentityInput{
		RoleArn:          aws.String(roleARN),
		RoleSessionName:  aws.String(sessionName),
		WebIdentityToken: aws.String(webIdentity),
		DurationSeconds:  aws.Int32(durationSeconds),
	}
	result, err := client.AssumeRoleWithWebIdentity(ctx, input)
	if err != nil {
		return nil, err
	}
	if result.Credentials == nil {
		return nil, errNoCredentialsProduced
	}

	creds := result.Credentials
	expiresAt := creds.Expiration

	metadata := map[string]string{
		"access_key_id":     strings.TrimSpace(lo.FromPtr(creds.AccessKeyId)),
		"secret_access_key": strings.TrimSpace(lo.FromPtr(creds.SecretAccessKey)),
		"session_token":     strings.TrimSpace(lo.FromPtr(creds.SessionToken)),
		"region":            region,
		"role_arn":          roleARN,
		"session_name":      sessionName,
	}
	if result.AssumedRoleUser != nil && result.AssumedRoleUser.Arn != nil {
		metadata["assumed_role_arn"] = strings.TrimSpace(*result.AssumedRoleUser.Arn)
	}
	for key, value := range attrs {
		if value == "" {
			continue
		}
		if _, exists := metadata[key]; exists {
			continue
		}
		metadata[key] = value
	}

	credsOut := &Credentials{
		AccessToken: strings.TrimSpace(lo.FromPtr(creds.SessionToken)),
		Metadata:    metadata,
	}
	if expiresAt != nil && !expiresAt.IsZero() {
		ts := expiresAt.UTC()
		credsOut.ExpiresAt = &ts
	}
	return credsOut, nil
}

func resolveAWSRegion(attrs map[string]string, rt *ProviderRuntime) string {
	if attrs != nil {
		if region := strings.TrimSpace(attrs["region"]); region != "" {
			return region
		}
	}
	if rt != nil && rt.Spec.AWSSTS != nil && strings.TrimSpace(rt.Spec.AWSSTS.Region) != "" {
		return strings.TrimSpace(rt.Spec.AWSSTS.Region)
	}
	return defaultAWSRegion
}

func resolveAWSRoleARN(attrs map[string]string, rt *ProviderRuntime) string {
	if attrs != nil {
		if arn := strings.TrimSpace(attrs["role_arn"]); arn != "" {
			return arn
		}
	}
	if rt != nil && rt.Spec.AWSSTS != nil && strings.TrimSpace(rt.Spec.AWSSTS.RoleARN) != "" {
		return strings.TrimSpace(rt.Spec.AWSSTS.RoleARN)
	}
	return ""
}

func parseDurationSeconds(value string) (int32, error) {
	v := strings.TrimSpace(value)
	if v == "" {
		return 0, errors.New("empty duration")
	}
	if strings.ContainsAny(v, "h") || strings.ContainsAny(v, "m") || strings.ContainsAny(v, "s") {
		dur, err := time.ParseDuration(v)
		if err != nil {
			return 0, err
		}
		return int32(dur / time.Second), nil
	}
	secs, err := strconv.ParseInt(v, 10, 32)
	if err != nil {
		return 0, err
	}
	return int32(secs), nil
}
func (b *Broker) mintAzureFederation(_ context.Context, _ string, provider string, _ *ProviderRuntime, _ *TokenBundle) (*Credentials, error) {
	return nil, fmt.Errorf("%s: %w", provider, errAzureFederationNotImplemented)
}

func (b *Broker) runtimeForProvider(provider string) (*ProviderRuntime, error) {
	if b.registry == nil {
		return nil, errRegistryNotConfigured
	}
	rt, ok := (*b.registry)[provider]
	if !ok || rt == nil {
		return nil, errProviderNotRegistered
	}
	return rt, nil
}

func (b *Broker) getCachedCredentials(orgID, provider string) *Credentials {
	b.mu.RLock()
	entry, ok := b.cache[cacheKey{OrgID: orgID, Provider: provider}]
	b.mu.RUnlock()
	if !ok {
		return nil
	}

	now := b.now()
	if entry.expires.IsZero() || now.Before(entry.expires) {
		return entry.creds
	}

	b.mu.Lock()
	delete(b.cache, cacheKey{OrgID: orgID, Provider: provider})
	b.mu.Unlock()
	return nil
}

func (b *Broker) setCachedCredentials(orgID, provider string, creds *Credentials) {
	if creds == nil {
		return
	}

	expiry := time.Time{}
	if creds.ExpiresAt != nil {
		expiry = creds.ExpiresAt.Add(-defaultCacheBuffer)
	}

	b.mu.Lock()
	b.cache[cacheKey{OrgID: orgID, Provider: provider}] = cachedCredential{creds: creds, expires: expiry}
	b.mu.Unlock()
}

func (b *Broker) newRequester(opts ...httpsling.Option) (*httpsling.Requester, error) {
	options := make([]httpsling.Option, 0, len(opts)+1)

	if client := b.httpClient; client != nil {
		doer := httpsling.DoerFunc(func(req *http.Request) (*http.Response, error) {
			return client.Do(req)
		})
		options = append(options, httpsling.WithDoer(doer))
	}

	options = append(options, opts...)

	return httpsling.New(options...)
}

func shouldStoreRefresh(rt *ProviderRuntime) bool {
	if rt == nil || rt.Spec.Persistence == nil {
		return true
	}
	return rt.Spec.Persistence.StoreRefreshToken
}

func signGitHubAppJWT(appID int64, privateKeyPEM string) (string, error) {
	block, _ := pem.Decode([]byte(privateKeyPEM))
	if block == nil {
		return "", errInvalidGitHubAppPrivateKey
	}
	key, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		// Try PKCS8
		pkcs8Key, pkcs8Err := x509.ParsePKCS8PrivateKey(block.Bytes)
		if pkcs8Err != nil {
			return "", fmt.Errorf("parse private key: %w", err)
		}
		var ok bool
		key, ok = pkcs8Key.(*rsa.PrivateKey)
		if !ok {
			return "", errPrivateKeyNotRSA
		}
	}

	now := time.Now().UTC()
	claims := jwt.MapClaims{
		"iat": now.Add(-time.Minute).Unix(),
		"exp": now.Add(defaultJWTExpiryDuration).Unix(),
		"iss": appID,
	}

	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	return token.SignedString(key)
}

func extractScopes(rt *ProviderRuntime, attrs map[string]string) []string {
	if raw := attrs["scopes"]; raw != "" {
		parts := strings.Split(raw, ",")
		scopes := make([]string, 0, len(parts))
		for _, p := range parts {
			if trimmed := strings.TrimSpace(p); trimmed != "" {
				scopes = append(scopes, trimmed)
			}
		}
		if len(scopes) > 0 {
			return scopes
		}
	}

	if rt != nil && rt.Spec.WorkloadIdentity != nil && len(rt.Spec.WorkloadIdentity.Scopes) > 0 {
		return append([]string{}, rt.Spec.WorkloadIdentity.Scopes...)
	}

	return []string{"https://www.googleapis.com/auth/cloud-platform"}
}
