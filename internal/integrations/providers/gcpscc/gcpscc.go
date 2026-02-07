package gcpscc

import (
	"context"
	"encoding/json"
	"fmt"
	"maps"
	"strings"
	"time"

	cloudscc "cloud.google.com/go/securitycenter/apiv2"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/impersonate"
	"google.golang.org/api/option"
	stsv1 "google.golang.org/api/sts/v1"

	"github.com/theopenlane/core/common/integrations/auth"
	"github.com/theopenlane/core/common/integrations/config"
	"github.com/theopenlane/core/common/integrations/operations"
	"github.com/theopenlane/core/common/integrations/types"
	"github.com/theopenlane/core/common/models"
	"github.com/theopenlane/core/internal/integrations/providers"
	"github.com/theopenlane/core/pkg/logx"
)

// TypeGCPSCC identifies the GCP Security Command Center provider.
const TypeGCPSCC = types.ProviderType("gcp_scc")

const (
	workloadGrantType        = "urn:ietf:params:oauth:grant-type:token-exchange"
	requestedTokenType       = "urn:ietf:params:oauth:token-type:access_token" //nolint:gosec
	defaultScope             = "https://www.googleapis.com/auth/cloud-platform"
	defaultSubjectTokenType  = "urn:ietf:params:oauth:token-type:id_token" //nolint:gosec
	subjectTokenAttr         = "subject_token"
	subjectTokenTypeAttr     = "subject_token_type"
	minImpersonationLifetime = time.Minute
	maxImpersonationLifetime = time.Hour
	defaultImpersonationLife = time.Hour
)

// ClientSecurityCenter identifies the SCC client descriptor.
const ClientSecurityCenter types.ClientName = "securitycenter.v2"

const securityCenterDescription = "Google Cloud Security Command Center v2 client"

var (
	errSubjectTokenRequired   = ErrSubjectTokenRequired
	errProjectIDRequired      = ErrProjectIDRequired
	errAudienceRequired       = ErrAudienceRequired
	errServiceAccountRequired = ErrServiceAccountRequired
	errSourceIDRequired       = ErrSourceIDRequired
	errCredentialMetadata     = ErrCredentialMetadataRequired
	errAccessTokenMissing     = ErrAccessTokenMissing
)

var _ types.ClientProvider = (*Provider)(nil)

// Provider implements the GCP SCC integration using workload identity federation.
type Provider struct {
	providers.BaseProvider
	spec      config.ProviderSpec
	defaults  workloadDefaults
	stsClient func(ctx context.Context) (*stsv1.Service, error)
}

type workloadDefaults struct {
	scopes            []string
	tokenLifetime     time.Duration
	audience          string
	targetServiceAcct string
	subjectTokenType  string
}

// Builder returns the GCP SCC provider builder.
func Builder() providers.Builder {
	return providers.BuilderFunc{
		ProviderType: TypeGCPSCC,
		BuildFunc: func(_ context.Context, spec config.ProviderSpec) (providers.Provider, error) {
			if spec.GoogleWorkloadIdentity == nil {
				return nil, providers.ErrSpecWorkloadIdentityRequired
			}

			return newProvider(spec), nil
		},
	}
}

// newProvider constructs the GCP SCC provider instance
func newProvider(spec config.ProviderSpec) *Provider {
	defaultScopes := append([]string(nil), spec.GoogleWorkloadIdentity.Scopes...)
	if len(defaultScopes) == 0 {
		defaultScopes = []string{defaultScope}
	}

	lifetime := spec.GoogleWorkloadIdentity.TokenLifetime
	if lifetime <= 0 {
		lifetime = defaultImpersonationLife
	}

	subjectTokenType := strings.TrimSpace(spec.GoogleWorkloadIdentity.SubjectTokenType)
	if subjectTokenType == "" {
		subjectTokenType = defaultSubjectTokenType
	}
	audience := strings.TrimSpace(spec.GoogleWorkloadIdentity.Audience)
	targetServiceAcct := strings.TrimSpace(spec.GoogleWorkloadIdentity.TargetServiceAccount)

	defaults := workloadDefaults{
		scopes:            defaultScopes,
		tokenLifetime:     clampLifetime(lifetime),
		audience:          audience,
		targetServiceAcct: targetServiceAcct,
		subjectTokenType:  subjectTokenType,
	}

	caps := types.ProviderCapabilities{
		SupportsRefreshTokens: false,
		SupportsClientPooling: true,
		SupportsMetadataForm:  len(spec.CredentialsSchema) > 0,
	}

	return &Provider{
		BaseProvider: providers.NewBaseProvider(TypeGCPSCC, caps, nil, nil),
		spec:         spec,
		defaults:     defaults,
		stsClient: func(ctx context.Context) (*stsv1.Service, error) {
			return stsv1.NewService(ctx)
		},
	}
}

// BeginAuth is not applicable for workload identity flows; callers should rely on declarative metadata.
func (p *Provider) BeginAuth(_ context.Context, _ types.AuthContext) (types.AuthSession, error) {
	return nil, ErrBeginAuthNotSupported
}

// Mint exchanges stored workload identity metadata for short-lived Google credentials and persists the updated payload.
func (p *Provider) Mint(ctx context.Context, subject types.CredentialSubject) (types.CredentialPayload, error) {
	meta, err := metadataFromPayload(subject.Credential)
	if err != nil {
		return types.CredentialPayload{}, err
	}

	meta = meta.withDefaults(p.defaults)
	if err := meta.validate(); err != nil {
		return types.CredentialPayload{}, err
	}

	if meta.ServiceAccountKey != "" {
		if _, err := serviceAccountCredentials(ctx, string(meta.ServiceAccountKey), meta.Scopes); err != nil {
			return types.CredentialPayload{}, err
		}

		providerData := meta.persist(subject.Credential.Data.ProviderData)

		builder := types.NewCredentialBuilder(TypeGCPSCC).With(
			types.WithCredentialKind(types.CredentialKindMetadata),
			types.WithCredentialSet(models.CredentialSet{
				ProviderData: providerData,
			}),
		)

		return builder.Build()
	}

	var subjectToken, tokenType string
	subjectToken, tokenType, err = resolveSubjectToken(subject, meta, p.defaults)
	if err != nil {
		return types.CredentialPayload{}, err
	}
	meta.SubjectToken = types.TrimmedString(subjectToken)

	accessToken, err := p.mintWorkloadToken(ctx, meta, subjectToken, tokenType)
	if err != nil {
		return types.CredentialPayload{}, err
	}

	providerData := meta.persist(subject.Credential.Data.ProviderData)

	builder := types.NewCredentialBuilder(TypeGCPSCC).With(
		types.WithCredentialKind(types.CredentialKindWorkload),
		types.WithOAuthToken(accessToken),
		types.WithCredentialSet(models.CredentialSet{
			ProviderData: providerData,
		}),
	)

	return builder.Build()
}

// ClientDescriptors returns the client builders exposed by this provider.
func (p *Provider) ClientDescriptors() []types.ClientDescriptor {
	return []types.ClientDescriptor{
		{
			Provider:    TypeGCPSCC,
			Name:        ClientSecurityCenter,
			Description: securityCenterDescription,
			Build:       buildSecurityCenterClient,
			ConfigSchema: map[string]any{
				"type": "object",
				"properties": map[string]any{
					"findingFilter": map[string]any{
						"type":        "string",
						"description": "Optional SCC findings filter overriding the stored metadata.",
					},
					"sourceId": map[string]any{
						"type":        "string",
						"description": "Optional SCC source override (e.g. organizations/123/sources/456).",
					},
				},
			},
		},
	}
}

// mintWorkloadToken exchanges subject tokens for an impersonated access token
func (p *Provider) mintWorkloadToken(ctx context.Context, meta credentialMetadata, subjectToken, subjectTokenType string) (*oauth2.Token, error) {
	stsSvc, err := p.stsClient(ctx)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrSTSInit, err)
	}

	scope := strings.Join(meta.Scopes, " ")
	if scope == "" {
		scope = defaultScope
	}

	optionsValue, err := buildSTSOptions(meta)
	if err != nil {
		return nil, err
	}

	req := &stsv1.GoogleIdentityStsV1ExchangeTokenRequest{
		Audience:           string(meta.Audience),
		GrantType:          workloadGrantType,
		RequestedTokenType: requestedTokenType,
		Scope:              scope,
		SubjectToken:       subjectToken,
		SubjectTokenType:   subjectTokenType,
		Options:            optionsValue,
	}

	resp, err := stsSvc.V1.Token(req).Context(ctx).Do()
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrWorkloadIdentityExchange, err)
	}

	baseToken := &oauth2.Token{
		AccessToken: resp.AccessToken,
		TokenType:   resp.TokenType,
		Expiry:      time.Now().Add(time.Duration(resp.ExpiresIn) * time.Second),
	}

	if resp.ExpiresIn == 0 {
		baseToken.Expiry = time.Time{}
	}

	staticSource := oauth2.StaticTokenSource(baseToken)

	lifetime := clampLifetime(meta.tokenLifetime())
	cfg := impersonate.CredentialsConfig{
		TargetPrincipal: string(meta.ServiceAccountEmail),
		Scopes:          meta.Scopes,
		Lifetime:        lifetime,
	}

	ts, err := impersonate.CredentialsTokenSource(ctx, cfg, option.WithTokenSource(staticSource))
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrImpersonateServiceAccount, err)
	}

	token, err := ts.Token()
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrImpersonatedTokenFetch, err)
	}

	return token, nil
}

// buildSecurityCenterClient builds the SCC client using stored credentials
func buildSecurityCenterClient(ctx context.Context, payload types.CredentialPayload, _ map[string]any) (any, error) {
	meta, err := metadataFromPayload(payload)
	if err != nil {
		return nil, err
	}

	clientOpts, err := securityCenterClientOptions(ctx, meta, payload.Token)
	if err != nil {
		return nil, err
	}

	opts := append([]option.ClientOption{}, clientOpts...)
	if meta.ProjectID != "" {
		opts = append(opts, option.WithQuotaProject(string(meta.ProjectID)))
	}

	client, err := cloudscc.NewClient(ctx, opts...)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrSecurityCenterClientCreate, err)
	}

	return client, nil
}

// securityCenterClientOptions builds client options based on available credentials
func securityCenterClientOptions(ctx context.Context, meta credentialMetadata, token *oauth2.Token) ([]option.ClientOption, error) {
	logger := logx.FromContext(ctx).With().
		Str("provider", string(TypeGCPSCC)).
		Str("projectId", string(meta.ProjectID)).
		Logger()

	if meta.ServiceAccountKey != "" {
		creds, err := serviceAccountCredentials(ctx, string(meta.ServiceAccountKey), meta.Scopes)
		if err != nil {
			logger.Error().Err(err).Msg("gcpscc: failed to parse service account key for client")
			return nil, err
		}

		logger.Debug().Msg("gcpscc: building client with service account key")
		return []option.ClientOption{
			option.WithCredentials(creds),
		}, nil
	}

	if token != nil {
		accessEmpty := strings.TrimSpace(token.AccessToken) == ""
		logger = logger.With().
			Bool("tokenEmpty", accessEmpty).
			Time("tokenExpiry", token.Expiry).
			Str("tokenType", token.Type()).
			Logger()
		logger.Debug().Msg("gcpscc: building client with minted oauth token")
		return []option.ClientOption{
			option.WithTokenSource(oauth2.StaticTokenSource(token)),
		}, nil
	}

	logger.Error().Msg("gcpscc: service account key and oauth token missing for client build")
	return nil, errAccessTokenMissing
}

// serviceAccountCredentials parses and validates a service account key
func serviceAccountCredentials(ctx context.Context, rawKey string, scopes []string) (*google.Credentials, error) {
	key := auth.NormalizeServiceAccountKey(rawKey)
	if key == "" {
		return nil, ErrServiceAccountKeyInvalid
	}

	scopeList := scopes
	if len(scopeList) == 0 {
		scopeList = []string{defaultScope}
	}

	creds, err := google.CredentialsFromJSONWithType(ctx, []byte(key), google.ServiceAccount, scopeList...)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrServiceAccountKeyInvalid, err)
	}

	return creds, nil
}

// credentialMetadata captures the persisted SCC metadata supplied during activation.
type credentialMetadata struct {
	ProjectID                types.TrimmedString `json:"projectId"`
	OrganizationID           types.TrimmedString `json:"organizationId"`
	WorkloadIdentityProvider types.TrimmedString `json:"workloadIdentityProvider"`
	Audience                 types.TrimmedString `json:"audience"`
	ServiceAccountEmail      types.TrimmedString `json:"serviceAccountEmail"`
	SourceID                 types.TrimmedString `json:"sourceId"`
	Scopes                   []string            `json:"scopes"`
	TokenLifetime            types.TrimmedString `json:"tokenLifetime"`
	AudienceHint             types.TrimmedString `json:"audienceHint"`
	WorkloadPoolProject      types.TrimmedString `json:"workloadPoolProject"`
	FindingFilter            types.TrimmedString `json:"findingFilter"`
	SubjectToken             types.TrimmedString `json:"subjectToken"`
	ServiceAccountKey        types.TrimmedString `json:"serviceAccountKey"`
}

// withDefaults applies provider defaults to missing metadata values
func (m credentialMetadata) withDefaults(defaults workloadDefaults) credentialMetadata {
	result := m
	if len(result.Scopes) == 0 {
		result.Scopes = append([]string(nil), defaults.scopes...)
	}
	if result.ServiceAccountEmail == "" {
		result.ServiceAccountEmail = types.TrimmedString(defaults.targetServiceAcct)
	}
	if result.Audience == "" {
		result.Audience = types.TrimmedString(defaults.audience)
	}
	return result
}

// validate ensures required metadata values are present
func (m credentialMetadata) validate() error {
	if m.ProjectID == "" {
		return errProjectIDRequired
	}

	if m.ServiceAccountKey != "" {
		return nil
	}

	switch {
	case m.Audience == "":
		return errAudienceRequired
	case m.ServiceAccountEmail == "":
		return errServiceAccountRequired
	}
	return nil
}

// persist merges metadata into the existing provider data map
func (m credentialMetadata) persist(existing map[string]any) map[string]any {
	out := map[string]any{}
	if len(existing) > 0 {
		out = maps.Clone(existing)
	}

	setIfNotEmpty(out, "projectId", m.ProjectID)
	setIfNotEmpty(out, "organizationId", m.OrganizationID)
	setIfNotEmpty(out, "workloadIdentityProvider", m.WorkloadIdentityProvider)
	setIfNotEmpty(out, "audience", m.Audience)
	setIfNotEmpty(out, "serviceAccountEmail", m.ServiceAccountEmail)
	setIfNotEmpty(out, "sourceId", m.SourceID)
	setIfNotEmpty(out, "audienceHint", m.AudienceHint)
	setIfNotEmpty(out, "workloadPoolProject", m.WorkloadPoolProject)
	setIfNotEmpty(out, "findingFilter", m.FindingFilter)
	setIfNotEmpty(out, "serviceAccountKey", m.ServiceAccountKey)
	if len(m.Scopes) > 0 {
		out["scopes"] = append([]string(nil), m.Scopes...)
	}
	if m.TokenLifetime != "" {
		out["tokenLifetime"] = string(m.TokenLifetime)
	}

	// Persist subject token only when explicitly supplied to allow autonomous refresh.
	if m.SubjectToken != "" {
		out["subjectToken"] = string(m.SubjectToken)
	}

	return out
}

// tokenLifetime parses the configured token lifetime
func (m credentialMetadata) tokenLifetime() time.Duration {
	if m.TokenLifetime == "" {
		return 0
	}
	dur, err := time.ParseDuration(string(m.TokenLifetime))
	if err != nil {
		return 0
	}
	return dur
}

// metadataFromPayload decodes provider metadata from the credential payload
func metadataFromPayload(payload types.CredentialPayload) (credentialMetadata, error) {
	if payload.Data.ProviderData == nil {
		return credentialMetadata{}, errCredentialMetadata
	}

	var meta credentialMetadata
	if err := operations.DecodeConfig(payload.Data.ProviderData, &meta); err != nil {
		return credentialMetadata{}, fmt.Errorf("%w: %w", ErrMetadataDecode, err)
	}

	return meta.normalize(), nil
}

// resolveSubjectToken selects the subject token and token type for STS exchange
func resolveSubjectToken(subject types.CredentialSubject, meta credentialMetadata, defaults workloadDefaults) (token string, tokenType string, err error) {
	if attr := strings.TrimSpace(subject.Attributes[subjectTokenAttr]); attr != "" {
		token = attr
	} else {
		token = string(meta.SubjectToken)
	}

	if token == "" {
		return "", "", errSubjectTokenRequired
	}

	if attrType := strings.TrimSpace(subject.Attributes[subjectTokenTypeAttr]); attrType != "" {
		tokenType = attrType
	} else {
		tokenType = defaults.subjectTokenType
	}

	return token, tokenType, nil
}

// buildSTSOptions encodes STS options for workload pool billing
func buildSTSOptions(meta credentialMetadata) (string, error) {
	if meta.WorkloadPoolProject == "" {
		return "", nil
	}

	payload := map[string]string{
		"userProject": string(meta.WorkloadPoolProject),
	}

	encoded, err := json.Marshal(payload)
	if err != nil {
		return "", fmt.Errorf("%w: %w", ErrSTSOptionsEncode, err)
	}

	return string(encoded), nil
}

// clampLifetime bounds token lifetimes to supported ranges
func clampLifetime(value time.Duration) time.Duration {
	if value <= 0 {
		return defaultImpersonationLife
	}
	if value < minImpersonationLifetime {
		return minImpersonationLifetime
	}
	if value > maxImpersonationLifetime {
		return maxImpersonationLifetime
	}
	return value
}

// setIfNotEmpty writes a key when the value is non-empty
func setIfNotEmpty[T ~string](target map[string]any, key string, value T) {
	if value == "" {
		return
	}
	target[key] = string(value)
}

// normalize cleans up metadata values for persistence
func (m credentialMetadata) normalize() credentialMetadata {
	normalized := m
	normalized.ServiceAccountKey = types.TrimmedString(auth.NormalizeServiceAccountKey(string(normalized.ServiceAccountKey)))
	normalized.Scopes = normalizeScopes(normalized.Scopes)
	return normalized
}

// normalizeScopes trims, flattens, and de-duplicates scope values
func normalizeScopes(scopes []string) []string {
	if len(scopes) == 0 {
		return nil
	}

	out := make([]string, 0, len(scopes))
	for _, scope := range scopes {
		scope = strings.TrimSpace(scope)
		if scope == "" {
			continue
		}

		if strings.HasPrefix(scope, "[") && strings.HasSuffix(scope, "]") {
			var decoded []string
			if err := json.Unmarshal([]byte(scope), &decoded); err == nil {
				out = append(out, normalizeScopes(decoded)...)
				continue
			}
		}

		out = append(out, scope)
	}

	if len(out) == 0 {
		return nil
	}

	seen := make(map[string]struct{}, len(out))
	uniq := out[:0]
	for _, scope := range out {
		if _, ok := seen[scope]; ok {
			continue
		}
		seen[scope] = struct{}{}
		uniq = append(uniq, scope)
	}

	return uniq
}
