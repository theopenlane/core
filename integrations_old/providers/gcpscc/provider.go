package gcpscc

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"strings"
	"time"

	cloudscc "cloud.google.com/go/securitycenter/apiv2"
	securitycenterpb "cloud.google.com/go/securitycenter/apiv2/securitycenterpb"
	"github.com/samber/lo"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/impersonate"
	"google.golang.org/api/iterator"
	"google.golang.org/api/option"
	stsv1 "google.golang.org/api/sts/v1"
	"google.golang.org/protobuf/encoding/protojson"

	"github.com/theopenlane/core/internal/integrations/providerkit"
	"github.com/theopenlane/core/internal/integrations/providers"
	"github.com/theopenlane/core/internal/integrations/spec"
	"github.com/theopenlane/core/internal/integrations/types"
	"github.com/theopenlane/core/pkg/jsonx"
	"github.com/theopenlane/core/pkg/logx"
)

// TypeGCPSCC identifies the GCP Security Command Center provider
const TypeGCPSCC = types.ProviderType("gcpscc")

// ClientSecurityCenter identifies the SCC client descriptor
const ClientSecurityCenter types.ClientName = "securitycenter.v2"

const securityCenterDescription = "Google Cloud Security Command Center v2 client"

const (
	workloadGrantType       = "urn:ietf:params:oauth:grant-type:token-exchange"
	requestedTokenType      = "urn:ietf:params:oauth:token-type:access_token" //nolint:gosec
	defaultScope            = "https://www.googleapis.com/auth/cloud-platform"
	defaultSubjectTokenType = "urn:ietf:params:oauth:token-type:id_token" //nolint:gosec
	projectScopeAll         = "all"
	projectScopeSpecific    = "specific"
	subjectTokenAttr        = "subject_token"
	subjectTokenTypeAttr    = "subject_token_type"
	sccAlertTypeFinding     = "finding"
	findingsPageSize        = 100
	findingsMaxPageSize     = 1000
	settingsPageSize        = 10
	sampleConfigsCapacity   = 5
)

var (
	minImpersonationLifetime = time.Minute
	maxImpersonationLifetime = time.Hour
	defaultImpersonationLife = time.Hour
)

// Operation names published by the GCP SCC provider
const (
	OperationHealthDefault   types.OperationName = types.OperationHealthDefault
	OperationCollectFindings types.OperationName = "findings.collect"
	OperationScanSettings    types.OperationName = "settings.scan"
)

var _ types.ClientProvider = (*Provider)(nil)
var _ types.OperationProvider = (*Provider)(nil)

// gcpWorkloadIdentityConfig holds workload identity defaults for the GCP SCC provider.
type gcpWorkloadIdentityConfig struct {
	// Audience is the STS audience for the workload identity pool
	Audience string
	// TargetServiceAccount is the service account to impersonate
	TargetServiceAccount string
	// SubjectTokenType overrides the default subject token type
	SubjectTokenType string
	// Scopes lists default OAuth scopes for access tokens
	Scopes []string
}

// gcpSCCCredentialsSchema is the JSON Schema for GCP SCC credentials.
var gcpSCCCredentialsSchema = []byte(`{"type":"object","additionalProperties":false,"allOf":[{"anyOf":[{"required":["organizationId"]},{"required":["projectId"]}]},{"anyOf":[{"required":["serviceAccountKey"]},{"required":["workloadIdentityProvider","audience","serviceAccountEmail","subjectToken"]}]},{"if":{"properties":{"projectScope":{"const":"specific"}},"required":["projectScope"]},"then":{"required":["projectIds"]}}],"properties":{"projectId":{"type":"string","title":"GCP Project ID","description":"Primary GCP project used for single-project SCC access and quota project defaults."},"workloadIdentityProvider":{"type":"string","title":"Workload Identity Provider","description":"Fully-qualified provider resource (e.g. projects/123/locations/global/workloadIdentityPools/pool/providers/provider)."},"audience":{"type":"string","title":"STS Audience","description":"Audience used when exchanging the subject token with Google STS."},"serviceAccountEmail":{"type":"string","title":"Target Service Account","description":"Service account email to impersonate for SCC calls."},"organizationId":{"type":"string","title":"Organization ID","description":"GCP organization used as the SCC parent for organization-wide collection."},"projectScope":{"type":"string","title":"Project Scope","description":"Collect across all accessible projects or restrict to a specific set of project IDs.","enum":["all","specific"],"default":"all"},"projectIds":{"type":"array","title":"Project IDs","description":"Required when Project Scope is specific. These project IDs are used to fan out source collection.","items":{"type":"string"},"minItems":1},"sourceId":{"type":"string","title":"SCC Source ID","description":"Optional SCC source identifier (e.g. organizations/123/sources/456) to scope findings."},"sourceIds":{"type":"array","title":"SCC Source IDs","description":"Optional list of SCC source identifiers. Bare source IDs are expanded against selected parents.","items":{"type":"string"}},"scopes":{"type":"array","title":"Additional OAuth Scopes","description":"Optional extra OAuth scopes requested when impersonating the target service account.","items":{"type":"string"},"default":["https://www.googleapis.com/auth/cloud-platform"]},"tokenLifetime":{"type":"string","title":"Access Token Lifetime","description":"Optional access token lifetime (Go duration, e.g. 3600s).","default":"3600s"},"audienceHint":{"type":"string","title":"Audience Override","description":"Optional audience override if multiple STS exchanges are supported."},"workloadPoolProject":{"type":"string","title":"Workload Identity Project","description":"Optional project that owns the workload identity pool when different from projectId."},"findingFilter":{"type":"string","title":"Findings Filter","description":"Optional CEL filter applied when querying SCC findings."},"subjectToken":{"type":"string","title":"Subject Token","description":"External identity token (JWT/SAML/OIDC) used as the workload identity subject when STS exchanges occur.","minLength":1},"serviceAccountKey":{"type":"string","title":"Service Account Key JSON","description":"Paste a Google Cloud service-account JSON key to mint access tokens directly (testing shortcut).","minLength":1}}}`)

type workloadDefaults struct {
	scopes            []string
	audience          string
	targetServiceAcct string
	subjectTokenType  string
}

// Provider implements the GCP SCC integration using workload identity federation
type Provider struct {
	// BaseProvider holds shared provider metadata
	providers.BaseProvider
	defaults  workloadDefaults
	stsClient func(ctx context.Context) (*stsv1.Service, error)
}

// Builder returns the GCP SCC provider builder with the supplied operator config applied.
func Builder(cfg Config) providers.Builder {
	return providers.BuilderFunc{
		ProviderType: TypeGCPSCC,
		SpecFunc:     gcpSCCSpec,
		BuildFunc: func(_ context.Context, s spec.ProviderSpec) (types.Provider, error) {
			if err := providerkit.ValidateAuthType(s, types.AuthKindWorkloadIdentity, ErrAuthTypeMismatch); err != nil {
				return nil, err
			}

			wiConfig := gcpWorkloadIdentityConfig{
				Audience:             cfg.Audience,
				TargetServiceAccount: cfg.ServiceAccount,
				SubjectTokenType:     cfg.SubjectTokenType,
				Scopes:               cfg.Scopes,
			}

			return newProvider(s, wiConfig), nil
		},
	}
}

// gcpSCCSpec returns the static provider specification for the GCP Security Command Center provider.
func gcpSCCSpec() spec.ProviderSpec {
	return spec.ProviderSpec{
		Name:        "gcpscc",
		DisplayName: "Google Cloud SCC",
		Category:    "cloud",
		AuthType:    types.AuthKindWorkloadIdentity,
		Active:      lo.ToPtr(true),
		Visible:     lo.ToPtr(true),
		LogoURL:     "",
		DocsURL:     "https://docs.theopenlane.io/docs/platform/integrations/gcp_scc/overview",
		Persistence: &spec.PersistenceSpec{
			StoreRefreshToken: false,
		},
		Labels: map[string]string{
			"vendor":  "google",
			"product": "security-command-center",
		},
		CredentialsSchema: gcpSCCCredentialsSchema,
		Description:       "Collect Google Cloud Security Command Center findings and settings using workload identity or service-account based access.",
	}
}

// newProvider constructs the GCP SCC provider instance
func newProvider(s spec.ProviderSpec, wiConfig gcpWorkloadIdentityConfig) *Provider {
	defaultScopes := append([]string(nil), wiConfig.Scopes...)
	if len(defaultScopes) == 0 {
		defaultScopes = []string{defaultScope}
	}

	subjectTokenType := wiConfig.SubjectTokenType
	if subjectTokenType == "" {
		subjectTokenType = defaultSubjectTokenType
	}

	defaults := workloadDefaults{
		scopes:            defaultScopes,
		audience:          wiConfig.Audience,
		targetServiceAcct: wiConfig.TargetServiceAccount,
		subjectTokenType:  subjectTokenType,
	}

	caps := types.ProviderCapabilities{
		SupportsRefreshTokens: false,
		SupportsClientPooling: true,
		SupportsMetadataForm:  len(s.CredentialsSchema) > 0,
	}

	return &Provider{
		BaseProvider: providers.NewBaseProvider(TypeGCPSCC, caps, nil, nil),
		defaults:     defaults,
		stsClient: func(ctx context.Context) (*stsv1.Service, error) {
			return stsv1.NewService(ctx)
		},
	}
}

// BeginAuth is not applicable for workload identity flows
func (p *Provider) BeginAuth(_ context.Context, _ types.AuthContext) (types.AuthSession, error) {
	return nil, ErrBeginAuthNotSupported
}

// Mint exchanges stored workload identity metadata for short-lived Google credentials
func (p *Provider) Mint(ctx context.Context, subject types.CredentialMintRequest) (types.CredentialSet, error) {
	meta, err := metadataFromPayload(subject.Credential)
	if err != nil {
		return types.CredentialSet{}, err
	}

	meta = meta.withDefaults(p.defaults)

	if meta.ServiceAccountKey != "" {
		if _, err := serviceAccountCredentials(ctx, meta.ServiceAccountKey, meta.Scopes); err != nil {
			return types.CredentialSet{}, err
		}

		providerData, err := jsonx.ToRawMessage(meta.providerData())
		if err != nil {
			return types.CredentialSet{}, err
		}

		return types.CredentialSet{
			ProviderData: providerData,
		}, nil
	}

	subjectToken, tokenType, err := resolveSubjectToken(subject, meta, p.defaults)
	if err != nil {
		return types.CredentialSet{}, err
	}

	meta.SubjectToken = subjectToken

	accessToken, err := p.mintWorkloadToken(ctx, meta, subjectToken, tokenType)
	if err != nil {
		return types.CredentialSet{}, err
	}

	providerData, err := jsonx.ToRawMessage(meta.providerData())
	if err != nil {
		return types.CredentialSet{}, err
	}

	credential := types.CredentialSet{
		ProviderData:      providerData,
		OAuthAccessToken:  accessToken.AccessToken,
		OAuthRefreshToken: accessToken.RefreshToken,
		OAuthTokenType:    accessToken.TokenType,
	}
	if !accessToken.Expiry.IsZero() {
		exp := accessToken.Expiry.UTC()
		credential.OAuthExpiry = &exp
	}

	return credential, nil
}

// ClientDescriptors returns the client builders exposed by this provider
func (p *Provider) ClientDescriptors() []types.ClientDescriptor {
	return []types.ClientDescriptor{
		{
			Provider:     TypeGCPSCC,
			Name:         ClientSecurityCenter,
			Description:  securityCenterDescription,
			Build:        buildSecurityCenterClient,
			ConfigSchema: securityCenterClientConfigSchema,
		},
	}
}

// Operations returns the provider operations published by GCP SCC
func (p *Provider) Operations() []types.OperationDescriptor {
	return []types.OperationDescriptor{
		{
			Provider:    TypeGCPSCC,
			Name:        OperationHealthDefault,
			Kind:        types.OperationKindHealth,
			Description: "Validate Security Command Center access by listing sources.",
			Client:      ClientSecurityCenter,
			Run:         runSecurityCenterHealthOperation,
		},
		{
			Provider:     TypeGCPSCC,
			Name:         OperationCollectFindings,
			Kind:         types.OperationKindCollectFindings,
			Description:  "Collect Security Command Center findings using the configured source/filter.",
			Client:       ClientSecurityCenter,
			ConfigSchema: securityCenterFindingsConfigSchema,
			Run:          runSecurityCenterFindingsOperation,
		},
		{
			Provider:    TypeGCPSCC,
			Name:        OperationScanSettings,
			Kind:        types.OperationKindScanSettings,
			Description: "Inspect Security Command Center organization settings.",
			Client:      ClientSecurityCenter,
			Run:         runSecurityCenterSettingsOperation,
		},
	}
}

// mintWorkloadToken exchanges subject tokens for an impersonated access token
func (p *Provider) mintWorkloadToken(ctx context.Context, meta credentialMetadata, subjectToken, subjectTokenType string) (*oauth2.Token, error) {
	stsSvc, err := p.stsClient(ctx)
	if err != nil {
		return nil, ErrSTSInit
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
		Audience:           meta.Audience.String(),
		GrantType:          workloadGrantType,
		RequestedTokenType: requestedTokenType,
		Scope:              scope,
		SubjectToken:       subjectToken,
		SubjectTokenType:   subjectTokenType,
		Options:            optionsValue,
	}

	resp, err := stsSvc.V1.Token(req).Context(ctx).Do()
	if err != nil {
		return nil, ErrWorkloadIdentityExchange
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
		TargetPrincipal: meta.ServiceAccountEmail.String(),
		Scopes:          meta.Scopes,
		Lifetime:        lifetime,
	}

	ts, err := impersonate.CredentialsTokenSource(ctx, cfg, option.WithTokenSource(staticSource))
	if err != nil {
		return nil, ErrImpersonateServiceAccount
	}

	token, err := ts.Token()
	if err != nil {
		return nil, ErrImpersonatedTokenFetch
	}

	return token, nil
}

type securityCenterClientConfig struct {
	// FindingFilter overrides the finding filter in provider metadata
	FindingFilter string `json:"findingFilter,omitempty" jsonschema:"description=Optional SCC findings filter overriding the stored metadata."`
	// SourceID sets one SCC source (for example organizations/123/sources/456)
	SourceID string `json:"sourceId,omitempty" jsonschema:"description=Optional SCC source override (e.g. organizations/123/sources/456)."`
	// SourceIDs sets multiple SCC sources for fan-out collection
	SourceIDs []string `json:"sourceIds,omitempty" jsonschema:"description=Optional SCC source overrides for fan-out collection."`
}

var securityCenterClientConfigSchema = providerkit.SchemaFrom[securityCenterClientConfig]()

// buildSecurityCenterClient builds the SCC client using stored credentials
func buildSecurityCenterClient(ctx context.Context, credential types.CredentialSet, _ json.RawMessage) (types.ClientInstance, error) {
	meta, err := metadataFromPayload(credential)
	if err != nil {
		return types.EmptyClientInstance(), err
	}

	clientOpts, err := securityCenterClientOptions(ctx, meta, oauthTokenFromCredential(credential))
	if err != nil {
		return types.EmptyClientInstance(), err
	}

	opts := append([]option.ClientOption{}, clientOpts...)
	if meta.ProjectID != "" {
		opts = append(opts, option.WithQuotaProject(meta.ProjectID.String()))
	}

	client, err := cloudscc.NewClient(ctx, opts...)
	if err != nil {
		return types.EmptyClientInstance(), ErrSecurityCenterClientCreate
	}

	return types.NewClientInstance(client), nil
}

// securityCenterClientOptions builds client options based on available credentials
func securityCenterClientOptions(ctx context.Context, meta credentialMetadata, token *oauth2.Token) ([]option.ClientOption, error) {
	logger := logx.FromContext(ctx).With().
		Str("provider", string(TypeGCPSCC)).
		Str("projectId", meta.ProjectID.String()).
		Logger()

	if meta.ServiceAccountKey != "" {
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
		accessEmpty := token.AccessToken == ""
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
	return nil, ErrAccessTokenMissing
}

// serviceAccountCredentials parses and validates a service account key
func serviceAccountCredentials[T ~string](ctx context.Context, rawKey T, scopes []string) (*google.Credentials, error) {
	key := normalizeServiceAccountKey(string(rawKey))
	if key == "" {
		return nil, ErrServiceAccountKeyInvalid
	}

	scopeList := scopes
	if len(scopeList) == 0 {
		scopeList = []string{defaultScope}
	}

	creds, err := google.CredentialsFromJSONWithType(ctx, []byte(key), google.ServiceAccount, scopeList...)
	if err != nil {
		return nil, ErrServiceAccountKeyInvalid
	}

	return creds, nil
}

// normalizeServiceAccountKey trims and unwraps JSON-encoded service account keys
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

// credentialMetadata captures the persisted SCC metadata supplied during activation
type credentialMetadata struct {
	// ProjectID is the GCP project identifier
	ProjectID types.TrimmedString `json:"projectId,omitempty"`
	// OrganizationID is the GCP organization identifier
	OrganizationID types.TrimmedString `json:"organizationId,omitempty"`
	// ProjectScope controls whether collection runs across all accessible projects or a specific set
	ProjectScope types.LowerString `json:"projectScope,omitempty"`
	// ProjectIDs lists explicit project IDs used when projectScope is specific
	ProjectIDs []string `json:"projectIds,omitempty"`
	// WorkloadIdentityProvider is the workload identity provider resource name
	WorkloadIdentityProvider types.TrimmedString `json:"workloadIdentityProvider,omitempty"`
	// Audience is the STS audience used for token exchange
	Audience types.TrimmedString `json:"audience,omitempty"`
	// ServiceAccountEmail is the target service account to impersonate
	ServiceAccountEmail types.TrimmedString `json:"serviceAccountEmail,omitempty"`
	// SourceID is the SCC source identifier used for findings
	SourceID types.TrimmedString `json:"sourceId,omitempty"`
	// SourceIDs optionally lists multiple SCC source identifiers for fan-out collection
	SourceIDs []string `json:"sourceIds,omitempty"`
	// Scopes lists OAuth scopes requested for access tokens
	Scopes []string `json:"scopes,omitempty"`
	// TokenLifetime configures the impersonation token lifetime
	TokenLifetime types.TrimmedString `json:"tokenLifetime,omitempty"`
	// AudienceHint provides an optional audience hint for metadata
	AudienceHint types.TrimmedString `json:"audienceHint,omitempty"`
	// WorkloadPoolProject identifies the billing project for STS exchanges
	WorkloadPoolProject types.TrimmedString `json:"workloadPoolProject,omitempty"`
	// FindingFilter holds the default SCC findings filter
	FindingFilter types.TrimmedString `json:"findingFilter,omitempty"`
	// SubjectToken stores the subject token for STS exchange
	SubjectToken string `json:"subjectToken,omitempty"`
	// ServiceAccountKey stores the service account key JSON
	ServiceAccountKey string `json:"serviceAccountKey,omitempty"`
}

type sccProviderData struct {
	ProjectID                string   `json:"projectId,omitempty"`
	OrganizationID           string   `json:"organizationId,omitempty"`
	ProjectScope             string   `json:"projectScope,omitempty"`
	ProjectIDs               []string `json:"projectIds,omitempty"`
	WorkloadIdentityProvider string   `json:"workloadIdentityProvider,omitempty"`
	Audience                 string   `json:"audience,omitempty"`
	ServiceAccountEmail      string   `json:"serviceAccountEmail,omitempty"`
	SourceID                 string   `json:"sourceId,omitempty"`
	SourceIDs                []string `json:"sourceIds,omitempty"`
	Scopes                   []string `json:"scopes,omitempty"`
	TokenLifetime            string   `json:"tokenLifetime,omitempty"`
	AudienceHint             string   `json:"audienceHint,omitempty"`
	WorkloadPoolProject      string   `json:"workloadPoolProject,omitempty"`
	FindingFilter            string   `json:"findingFilter,omitempty"`
}

func (m credentialMetadata) providerData() sccProviderData {
	return sccProviderData{
		ProjectID:                m.ProjectID.String(),
		OrganizationID:           m.OrganizationID.String(),
		ProjectScope:             m.ProjectScope.String(),
		ProjectIDs:               m.ProjectIDs,
		WorkloadIdentityProvider: m.WorkloadIdentityProvider.String(),
		Audience:                 m.Audience.String(),
		ServiceAccountEmail:      m.ServiceAccountEmail.String(),
		SourceID:                 m.SourceID.String(),
		SourceIDs:                m.SourceIDs,
		Scopes:                   m.Scopes,
		TokenLifetime:            m.TokenLifetime.String(),
		AudienceHint:             m.AudienceHint.String(),
		WorkloadPoolProject:      m.WorkloadPoolProject.String(),
		FindingFilter:            m.FindingFilter.String(),
	}
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

// tokenLifetime parses the configured token lifetime
func (m credentialMetadata) tokenLifetime() time.Duration {
	if m.TokenLifetime == "" {
		return 0
	}

	dur, err := time.ParseDuration(m.TokenLifetime.String())
	if err != nil {
		return 0
	}

	return dur
}

// metadataFromPayload decodes provider metadata from the credential set
func metadataFromPayload(payload types.CredentialSet) (credentialMetadata, error) {
	if len(payload.ProviderData) == 0 {
		return credentialMetadata{}, ErrCredentialMetadataRequired
	}

	var meta credentialMetadata
	if err := jsonx.UnmarshalIfPresent(payload.ProviderData, &meta); err != nil {
		return credentialMetadata{}, ErrMetadataDecode
	}

	return meta.applyDefaults(), nil
}

func oauthTokenFromCredential(payload types.CredentialSet) *oauth2.Token {
	if payload.OAuthAccessToken == "" && payload.OAuthRefreshToken == "" {
		return nil
	}

	token := &oauth2.Token{
		AccessToken:  payload.OAuthAccessToken,
		RefreshToken: payload.OAuthRefreshToken,
		TokenType:    payload.OAuthTokenType,
	}
	if payload.OAuthExpiry != nil {
		token.Expiry = payload.OAuthExpiry.UTC()
	}

	return token
}

// resolveSubjectToken selects the subject token and token type for STS exchange
func resolveSubjectToken(subject types.CredentialMintRequest, meta credentialMetadata, defaults workloadDefaults) (token string, tokenType string, err error) {
	if attr := subject.Attributes[subjectTokenAttr]; attr != "" {
		token = attr
	} else {
		token = meta.SubjectToken
	}

	if token == "" {
		return "", "", ErrSubjectTokenRequired
	}

	if attrType := subject.Attributes[subjectTokenTypeAttr]; attrType != "" {
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
		"userProject": meta.WorkloadPoolProject.String(),
	}

	encoded, err := json.Marshal(payload)
	if err != nil {
		return "", ErrSTSOptionsEncode
	}

	return string(encoded), nil
}

// clampLifetime bounds token lifetimes to supported ranges
func clampLifetime(value time.Duration) time.Duration {
	if value <= 0 {
		return defaultImpersonationLife
	}

	return lo.Clamp(value, minImpersonationLifetime, maxImpersonationLifetime)
}

// applyDefaults fills in fallback values, deduplicates slice fields, and normalizes the service account key
func (m credentialMetadata) applyDefaults() credentialMetadata {
	normalized := m
	if normalized.ProjectScope == "" {
		normalized.ProjectScope = types.LowerString(projectScopeAll)
	}

	normalized.ProjectIDs = types.NormalizeStringSlice(normalized.ProjectIDs)
	normalized.SourceIDs = types.NormalizeStringSlice(normalized.SourceIDs)
	normalized.ServiceAccountKey = normalizeServiceAccountKey(normalized.ServiceAccountKey)
	normalized.Scopes = types.NormalizeStringSlice(normalized.Scopes)

	return normalized
}

type securityCenterFindingsConfig struct {
	// Pagination controls page sizing for SCC findings
	providerkit.Pagination
	// PayloadOptions controls payload inclusion for findings
	providerkit.PayloadOptions

	// Filter overrides the stored findings filter
	Filter types.TrimmedString `json:"filter"`
	// SourceID overrides the stored SCC source identifier
	SourceID types.TrimmedString `json:"sourceId"`
	// SourceIDs overrides stored SCC source identifiers for fan-out collection
	SourceIDs []string `json:"sourceIds"`
	// MaxFindings caps the number of findings returned
	MaxFindings int `json:"max_findings"`
}

type securityCenterFindingsSchema struct {
	// SourceID overrides the SCC source identifier
	SourceID types.TrimmedString `json:"sourceId,omitempty" jsonschema:"description=Optional SCC source override (full resource name or bare source ID)."`
	// SourceIDs overrides SCC source identifiers for fan-out collection
	SourceIDs []string `json:"sourceIds,omitempty" jsonschema:"description=Optional SCC source overrides for fan-out collection. Bare source IDs expand against selected parents."`
	// Filter overrides the stored SCC findings filter
	Filter types.TrimmedString `json:"filter,omitempty" jsonschema:"description=Optional SCC findings filter overriding stored metadata."`
	// PageSize overrides the findings page size
	PageSize int `json:"page_size,omitempty" jsonschema:"description=Optional page size override (max 1000)."`
	// MaxFindings caps the total findings returned
	MaxFindings int `json:"max_findings,omitempty" jsonschema:"description=Optional cap on total findings returned."`
	// IncludePayloads controls whether raw payloads are returned
	IncludePayloads bool `json:"include_payloads,omitempty" jsonschema:"description=Return raw finding payloads in the response (defaults to false)."`
}

type securityCenterHealthDetails struct {
	Parents []string `json:"parents"`
}

type securityCenterFindingSample struct {
	Name     string `json:"name"`
	Category string `json:"category"`
	State    string `json:"state"`
	Severity string `json:"severity"`
	Source   string `json:"source"`
}

type securityCenterFindingsDetails struct {
	Sources          []string                      `json:"sources"`
	SourceCount      int                           `json:"sourceCount"`
	Filter           string                        `json:"filter"`
	TotalFindings    int                           `json:"totalFindings"`
	FindingsBySource map[string]int                `json:"findingsBySource"`
	SeverityCounts   map[string]int                `json:"severity_counts"`
	StateCounts      map[string]int                `json:"state_counts"`
	Samples          []securityCenterFindingSample `json:"samples"`
	Alerts           []types.AlertEnvelope         `json:"alerts,omitempty"`
}

type securityCenterNotificationConfigSample struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	PubSubTopic string `json:"pubsubTopic"`
	Parent      string `json:"parent"`
}

type securityCenterSettingsDetails struct {
	Parents                   []string                                 `json:"parents"`
	NotificationConfigCount   int                                      `json:"notificationConfigCount"`
	SampleNotificationConfigs []securityCenterNotificationConfigSample `json:"sampleNotificationConfigs"`
}

type securityCenterFailureDetails struct {
	Parent  string   `json:"parent,omitempty"`
	Parents []string `json:"parents,omitempty"`
	Source  string   `json:"source,omitempty"`
	Sources []string `json:"sources,omitempty"`
	Filter  string   `json:"filter,omitempty"`
}

var securityCenterFindingsConfigSchema = providerkit.SchemaFrom[securityCenterFindingsSchema]()

// runSecurityCenterHealthOperation checks SCC reachability for the org or project
func runSecurityCenterHealthOperation(ctx context.Context, input types.OperationInput) (types.OperationResult, error) {
	meta, err := metadataFromPayload(input.Credential)
	if err != nil {
		return types.OperationResult{}, err
	}

	client, ok := types.ClientInstanceAs[*cloudscc.Client](input.Client)
	if !ok {
		return types.OperationResult{}, ErrSecurityCenterClientRequired
	}

	parents, err := resolveSecurityCenterParents(meta)
	if err != nil {
		return types.OperationResult{}, err
	}

	for _, parent := range parents {
		req := &securitycenterpb.ListSourcesRequest{
			Parent:   parent,
			PageSize: 1,
		}

		it := client.ListSources(ctx, req)
		_, err = it.Next()
		if errors.Is(err, iterator.Done) {
			err = nil
		}

		if err != nil {
			return providerkit.OperationFailure("Security Command Center list sources failed", err, securityCenterFailureDetails{
				Parent:  parent,
				Parents: parents,
			})
		}
	}

	return providerkit.OperationSuccess(fmt.Sprintf("Security Command Center reachable for %d parent(s)", len(parents)), securityCenterHealthDetails{
		Parents: parents,
	}), nil
}

// runSecurityCenterFindingsOperation collects findings from SCC
func runSecurityCenterFindingsOperation(ctx context.Context, input types.OperationInput) (types.OperationResult, error) {
	meta, err := metadataFromPayload(input.Credential)
	if err != nil {
		return types.OperationResult{}, err
	}

	config, err := decodeSecurityCenterFindingsConfig(input.Config)
	if err != nil {
		return types.OperationResult{}, err
	}

	client, ok := types.ClientInstanceAs[*cloudscc.Client](input.Client)
	if !ok {
		return types.OperationResult{}, ErrSecurityCenterClientRequired
	}

	sources, err := resolveSecurityCenterSources(meta, config)
	if err != nil {
		return types.OperationResult{}, err
	}

	filter := lo.CoalesceOrEmpty(config.Filter, meta.FindingFilter).String()

	pageSize := config.EffectivePageSize(findingsPageSize)
	if pageSize <= 0 {
		pageSize = findingsPageSize
	}

	if pageSize > findingsMaxPageSize {
		pageSize = findingsMaxPageSize
	}

	maxFindings := config.MaxFindings

	total := 0
	samples := make([]securityCenterFindingSample, 0, providerkit.DefaultSampleSize)
	envelopes := make([]types.AlertEnvelope, 0)
	severityCounts := map[string]int{}
	stateCounts := map[string]int{}
	sourceCounts := map[string]int{}
	marshaler := protojson.MarshalOptions{UseProtoNames: true}

collectLoop:
	for _, sourceName := range sources {
		req := &securitycenterpb.ListFindingsRequest{
			Parent:   sourceName,
			Filter:   filter,
			PageSize: int32(min(pageSize, math.MaxInt32)), //nolint:gosec // bounds checked via min
		}

		it := client.ListFindings(ctx, req)

		for {
			result, err := it.Next()
			if errors.Is(err, iterator.Done) {
				break
			}

			if err != nil {
				return providerkit.OperationFailure("Security Command Center list findings failed", err, securityCenterFailureDetails{
					Sources: sources,
					Filter:  filter,
					Source:  sourceName,
				})
			}

			finding := result.GetFinding()
			if finding == nil {
				continue
			}

			if maxFindings > 0 && total >= maxFindings {
				break collectLoop
			}

			payload, err := marshaler.Marshal(finding)
			if err != nil {
				return providerkit.OperationFailure("Security Command Center finding serialization failed", err, securityCenterFailureDetails{
					Sources: sources,
					Source:  sourceName,
				})
			}

			resourceName := finding.GetResourceName()
			envelopes = append(envelopes, types.AlertEnvelope{
				AlertType: sccAlertTypeFinding,
				Resource:  resourceName,
				Payload:   payload,
			})
			total++
			sourceCounts[sourceName]++

			if severity := finding.GetSeverity().String(); severity != "" {
				key := strings.ToLower(severity)
				if key != "severity_unspecified" {
					severityCounts[key]++
				}
			}

			if state := finding.GetState().String(); state != "" {
				key := strings.ToLower(state)
				if key != "state_unspecified" {
					stateCounts[key]++
				}
			}

			if len(samples) < cap(samples) {
				samples = append(samples, securityCenterFindingSample{
					Name:     finding.GetName(),
					Category: finding.GetCategory(),
					State:    finding.GetState().String(),
					Severity: finding.GetSeverity().String(),
					Source:   sourceName,
				})
			}
		}
	}

	details := securityCenterFindingsDetails{
		Sources:          sources,
		SourceCount:      len(sources),
		Filter:           filter,
		TotalFindings:    total,
		FindingsBySource: sourceCounts,
		SeverityCounts:   severityCounts,
		StateCounts:      stateCounts,
		Samples:          samples,
	}
	if config.IncludePayloads {
		details.Alerts = envelopes
	}

	return providerkit.OperationSuccess(fmt.Sprintf("Collected %d findings from %d source(s)", total, len(sources)), details), nil
}

// runSecurityCenterSettingsOperation lists SCC notification configs
func runSecurityCenterSettingsOperation(ctx context.Context, input types.OperationInput) (types.OperationResult, error) {
	meta, err := metadataFromPayload(input.Credential)
	if err != nil {
		return types.OperationResult{}, err
	}

	client, ok := types.ClientInstanceAs[*cloudscc.Client](input.Client)
	if !ok {
		return types.OperationResult{}, ErrSecurityCenterClientRequired
	}

	parents, err := resolveSecurityCenterParents(meta)
	if err != nil {
		return types.OperationResult{}, err
	}

	configs := make([]securityCenterNotificationConfigSample, 0, sampleConfigsCapacity)
	count := 0

	for _, parent := range parents {
		req := &securitycenterpb.ListNotificationConfigsRequest{
			Parent:   parent,
			PageSize: settingsPageSize,
		}

		it := client.ListNotificationConfigs(ctx, req)
		for {
			cfg, err := it.Next()
			if errors.Is(err, iterator.Done) {
				break
			}

			if err != nil {
				return providerkit.OperationFailure("Security Command Center notification config scan failed", err, securityCenterFailureDetails{
					Parents: parents,
					Parent:  parent,
				})
			}

			count++
			if len(configs) < cap(configs) {
				configs = append(configs, securityCenterNotificationConfigSample{
					Name:        cfg.GetName(),
					Description: cfg.GetDescription(),
					PubSubTopic: cfg.GetPubsubTopic(),
					Parent:      parent,
				})
			}
		}
	}

	return providerkit.OperationSuccess(fmt.Sprintf("Discovered %d notification configs across %d parent(s)", count, len(parents)), securityCenterSettingsDetails{
		Parents:                   parents,
		NotificationConfigCount:   count,
		SampleNotificationConfigs: configs,
	}), nil
}

// resolveSecurityCenterParents chooses the SCC parent resources used for health/settings checks
func resolveSecurityCenterParents(meta credentialMetadata) ([]string, error) {
	if meta.OrganizationID != "" && meta.ProjectScope != projectScopeSpecific {
		return []string{fmt.Sprintf("organizations/%s", meta.OrganizationID)}, nil
	}

	if meta.ProjectScope == projectScopeSpecific {
		parentList := lo.FilterMap(meta.ProjectIDs, func(projectID string, _ int) (string, bool) {
			value := strings.TrimSpace(projectID)
			if value == "" {
				return "", false
			}

			return fmt.Sprintf("projects/%s", value), true
		})
		parentList = lo.Uniq(parentList)

		if len(parentList) == 0 {
			return nil, ErrProjectIDRequired
		}

		return parentList, nil
	}

	if meta.ProjectID != "" {
		return []string{fmt.Sprintf("projects/%s", meta.ProjectID)}, nil
	}

	if meta.OrganizationID != "" {
		return []string{fmt.Sprintf("organizations/%s", meta.OrganizationID)}, nil
	}

	return nil, ErrProjectIDRequired
}

// resolveSecurityCenterSources resolves source resource names from config and metadata
func resolveSecurityCenterSources(meta credentialMetadata, config securityCenterFindingsConfig) ([]string, error) {
	raw := make([]string, 0, 1+len(meta.SourceIDs))

	if config.SourceID != "" {
		raw = append(raw, config.SourceID.String())
	}

	raw = append(raw, config.SourceIDs...)

	if len(raw) == 0 {
		raw = append(raw, meta.SourceIDs...)
		if len(raw) == 0 && meta.SourceID != "" {
			raw = append(raw, meta.SourceID.String())
		}
	}

	if len(raw) == 0 {
		return nil, ErrSourceIDRequired
	}

	parents, err := resolveSecurityCenterParents(meta)
	if err != nil {
		return nil, err
	}

	out := lo.Uniq(lo.FlatMap(raw, func(source string, _ int) []string {
		source = strings.TrimSpace(source)
		if source == "" {
			return nil
		}

		switch {
		case strings.HasPrefix(source, "organizations/"), strings.HasPrefix(source, "projects/"):
			return []string{source}
		default:
			return lo.Map(parents, func(parent string, _ int) string {
				return fmt.Sprintf("%s/sources/%s", parent, source)
			})
		}
	}))

	if len(out) == 0 {
		return nil, ErrSourceIDRequired
	}

	return out, nil
}

// decodeSecurityCenterFindingsConfig decodes operation config into a typed struct
func decodeSecurityCenterFindingsConfig(config json.RawMessage) (securityCenterFindingsConfig, error) {
	var decoded securityCenterFindingsConfig
	if err := jsonx.UnmarshalIfPresent(config, &decoded); err != nil {
		return decoded, err
	}

	decoded.SourceIDs = types.NormalizeStringSlice(decoded.SourceIDs)

	return decoded, nil
}
