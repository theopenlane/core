package gcpscc

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"maps"
	"strings"
	"time"

	cloudscc "cloud.google.com/go/securitycenter/apiv2"

	"github.com/go-viper/mapstructure/v2"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/impersonate"
	"google.golang.org/api/option"
	stsv1 "google.golang.org/api/sts/v1"

	"github.com/theopenlane/shared/integrations/config"
	"github.com/theopenlane/shared/integrations/providers"
	"github.com/theopenlane/shared/integrations/types"
	"github.com/theopenlane/shared/logx"
	"github.com/theopenlane/shared/models"
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
	errSubjectTokenRequired   = errors.New("gcpscc: subject token required")
	errProjectIDRequired      = errors.New("gcpscc: projectId required")
	errAudienceRequired       = errors.New("gcpscc: audience required")
	errServiceAccountRequired = errors.New("gcpscc: service account email required")
	errSourceIDRequired       = errors.New("gcpscc: sourceId required")
	errCredentialMetadata     = errors.New("gcpscc: provider metadata required")
	errAccessTokenMissing     = errors.New("gcpscc: oauth token missing from credential payload")
	errServiceAccountKey      = errors.New("gcpscc: service account key invalid")
)

var _ types.ClientProvider = (*Provider)(nil)

// Provider implements the GCP SCC integration using workload identity federation.
type Provider struct {
	spec      config.ProviderSpec
	caps      types.ProviderCapabilities
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

func newProvider(spec config.ProviderSpec) *Provider {
	defaultScopes := append([]string(nil), spec.GoogleWorkloadIdentity.Scopes...)
	if len(defaultScopes) == 0 {
		defaultScopes = []string{defaultScope}
	}

	lifetime := spec.GoogleWorkloadIdentity.TokenLifetime
	if lifetime <= 0 {
		lifetime = defaultImpersonationLife
	}

	subjectTokenType := spec.GoogleWorkloadIdentity.SubjectTokenType
	if strings.TrimSpace(subjectTokenType) == "" {
		subjectTokenType = defaultSubjectTokenType
	}

	defaults := workloadDefaults{
		scopes:            defaultScopes,
		tokenLifetime:     clampLifetime(lifetime),
		audience:          spec.GoogleWorkloadIdentity.Audience,
		targetServiceAcct: spec.GoogleWorkloadIdentity.TargetServiceAccount,
		subjectTokenType:  subjectTokenType,
	}

	return &Provider{
		spec: spec,
		caps: types.ProviderCapabilities{
			SupportsRefreshTokens: false,
			SupportsClientPooling: true,
			SupportsMetadataForm:  len(spec.CredentialsSchema) > 0,
		},
		defaults: defaults,
		stsClient: func(ctx context.Context) (*stsv1.Service, error) {
			return stsv1.NewService(ctx)
		},
	}
}

// Type returns the provider identifier.
func (p *Provider) Type() types.ProviderType {
	return TypeGCPSCC
}

// Capabilities returns the provider capabilities.
func (p *Provider) Capabilities() types.ProviderCapabilities {
	return p.caps
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

	if key := strings.TrimSpace(meta.ServiceAccountKey); key != "" {
		if _, err := serviceAccountCredentials(ctx, key, meta.Scopes); err != nil {
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
	meta.SubjectToken = subjectToken

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

func (p *Provider) mintWorkloadToken(ctx context.Context, meta credentialMetadata, subjectToken, subjectTokenType string) (*oauth2.Token, error) {
	stsSvc, err := p.stsClient(ctx)
	if err != nil {
		return nil, fmt.Errorf("gcpscc: init sts service: %w", err)
	}

	scope := strings.Join(meta.Scopes, " ")
	if strings.TrimSpace(scope) == "" {
		scope = defaultScope
	}

	optionsValue, err := buildSTSOptions(meta)
	if err != nil {
		return nil, err
	}

	req := &stsv1.GoogleIdentityStsV1ExchangeTokenRequest{
		Audience:           meta.Audience,
		GrantType:          workloadGrantType,
		RequestedTokenType: requestedTokenType,
		Scope:              scope,
		SubjectToken:       subjectToken,
		SubjectTokenType:   subjectTokenType,
		Options:            optionsValue,
	}

	resp, err := stsSvc.V1.Token(req).Context(ctx).Do()
	if err != nil {
		return nil, fmt.Errorf("gcpscc: exchange workload identity token: %w", err)
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
		TargetPrincipal: meta.ServiceAccountEmail,
		Scopes:          meta.Scopes,
		Lifetime:        lifetime,
	}

	ts, err := impersonate.CredentialsTokenSource(ctx, cfg, option.WithTokenSource(staticSource))
	if err != nil {
		return nil, fmt.Errorf("gcpscc: impersonate service account: %w", err)
	}

	token, err := ts.Token()
	if err != nil {
		return nil, fmt.Errorf("gcpscc: fetch impersonated token: %w", err)
	}

	return token, nil
}

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
		opts = append(opts, option.WithQuotaProject(meta.ProjectID))
	}

	client, err := cloudscc.NewClient(ctx, opts...)
	if err != nil {
		return nil, fmt.Errorf("gcpscc: create security center client: %w", err)
	}

	return client, nil
}

func securityCenterClientOptions(ctx context.Context, meta credentialMetadata, token *oauth2.Token) ([]option.ClientOption, error) {
	logger := logx.FromContext(ctx).With().
		Str("provider", string(TypeGCPSCC)).
		Str("projectId", meta.ProjectID).
		Logger()

	if strings.TrimSpace(meta.ServiceAccountKey) != "" {
		creds, err := serviceAccountCredentials(ctx, meta.ServiceAccountKey, meta.Scopes)
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

func normalizeServiceAccountKey(value string) string {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return ""
	}

	var decoded string
	if err := json.Unmarshal([]byte(trimmed), &decoded); err == nil {
		return strings.TrimSpace(decoded)
	}

	return trimmed
}

func serviceAccountCredentials(ctx context.Context, rawKey string, scopes []string) (*google.Credentials, error) {
	key := normalizeServiceAccountKey(rawKey)
	if key == "" {
		return nil, errServiceAccountKey
	}

	scopeList := scopes
	if len(scopeList) == 0 {
		scopeList = []string{defaultScope}
	}

	creds, err := google.CredentialsFromJSON(ctx, []byte(key), scopeList...)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", errServiceAccountKey, err)
	}

	return creds, nil
}

// credentialMetadata captures the persisted SCC metadata supplied during activation.
type credentialMetadata struct {
	ProjectID                string   `mapstructure:"projectId"`
	OrganizationID           string   `mapstructure:"organizationId"`
	WorkloadIdentityProvider string   `mapstructure:"workloadIdentityProvider"`
	Audience                 string   `mapstructure:"audience"`
	ServiceAccountEmail      string   `mapstructure:"serviceAccountEmail"`
	SourceID                 string   `mapstructure:"sourceId"`
	Scopes                   []string `mapstructure:"scopes"`
	TokenLifetime            string   `mapstructure:"tokenLifetime"`
	AudienceHint             string   `mapstructure:"audienceHint"`
	WorkloadPoolProject      string   `mapstructure:"workloadPoolProject"`
	FindingFilter            string   `mapstructure:"findingFilter"`
	SubjectToken             string   `mapstructure:"subjectToken"`
	ServiceAccountKey        string   `mapstructure:"serviceAccountKey"`
}

func (m credentialMetadata) withDefaults(defaults workloadDefaults) credentialMetadata {
	result := m
	if len(result.Scopes) == 0 {
		result.Scopes = append([]string(nil), defaults.scopes...)
	}
	if strings.TrimSpace(result.ServiceAccountEmail) == "" {
		result.ServiceAccountEmail = defaults.targetServiceAcct
	}
	if strings.TrimSpace(result.Audience) == "" {
		result.Audience = defaults.audience
	}
	return result
}

func (m credentialMetadata) validate() error {
	if strings.TrimSpace(m.ProjectID) == "" {
		return errProjectIDRequired
	}

	if strings.TrimSpace(m.ServiceAccountKey) != "" {
		return nil
	}

	switch {
	case strings.TrimSpace(m.Audience) == "":
		return errAudienceRequired
	case strings.TrimSpace(m.ServiceAccountEmail) == "":
		return errServiceAccountRequired
	}
	return nil
}

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
	if strings.TrimSpace(m.TokenLifetime) != "" {
		out["tokenLifetime"] = m.TokenLifetime
	}

	// Persist subject token only when explicitly supplied to allow autonomous refresh.
	if strings.TrimSpace(m.SubjectToken) != "" {
		out["subjectToken"] = m.SubjectToken
	}

	return out
}

func (m credentialMetadata) tokenLifetime() time.Duration {
	if strings.TrimSpace(m.TokenLifetime) == "" {
		return 0
	}
	dur, err := time.ParseDuration(m.TokenLifetime)
	if err != nil {
		return 0
	}
	return dur
}

func metadataFromPayload(payload types.CredentialPayload) (credentialMetadata, error) {
	if payload.Data.ProviderData == nil {
		return credentialMetadata{}, errCredentialMetadata
	}

	var meta credentialMetadata
	decoder, err := mapstructure.NewDecoder(&mapstructure.DecoderConfig{
		TagName:          "mapstructure",
		Result:           &meta,
		WeaklyTypedInput: true,
	})
	if err != nil {
		return credentialMetadata{}, err
	}

	if err := decoder.Decode(payload.Data.ProviderData); err != nil {
		return credentialMetadata{}, fmt.Errorf("gcpscc: decode metadata: %w", err)
	}

	return meta.normalize(), nil
}

func resolveSubjectToken(subject types.CredentialSubject, meta credentialMetadata, defaults workloadDefaults) (token string, tokenType string, err error) {
	if attr := strings.TrimSpace(subject.Attributes[subjectTokenAttr]); attr != "" {
		token = attr
	} else {
		token = strings.TrimSpace(meta.SubjectToken)
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

func buildSTSOptions(meta credentialMetadata) (string, error) {
	if strings.TrimSpace(meta.WorkloadPoolProject) == "" {
		return "", nil
	}

	payload := map[string]string{
		"userProject": meta.WorkloadPoolProject,
	}

	encoded, err := json.Marshal(payload)
	if err != nil {
		return "", fmt.Errorf("gcpscc: encode sts options: %w", err)
	}

	return string(encoded), nil
}

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

func setIfNotEmpty(target map[string]any, key, value string) {
	if strings.TrimSpace(value) == "" {
		return
	}
	target[key] = value
}

func (m credentialMetadata) normalize() credentialMetadata {
	normalized := m
	normalized.ServiceAccountKey = normalizeServiceAccountKey(normalized.ServiceAccountKey)
	normalized.Scopes = normalizeScopes(normalized.Scopes)
	return normalized
}

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
