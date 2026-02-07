package activation

import (
	"context"
	"strings"

	"github.com/samber/lo"

	"github.com/theopenlane/core/common/integrations/types"
	"github.com/theopenlane/core/common/models"
	"github.com/theopenlane/core/internal/keymaker"
	"github.com/theopenlane/core/internal/keystore"
)

const defaultHealthOperation types.OperationName = "health.default"

// CredentialWriter persists credential payloads produced during activation
type CredentialWriter interface {
	SaveCredential(ctx context.Context, orgID string, payload types.CredentialPayload) (types.CredentialPayload, error)
}

// Service coordinates activation flows for OAuth and non-OAuth providers
type Service struct {
	keymaker   *keymaker.Service
	store      CredentialWriter
	operations *keystore.OperationManager
}

// NewService constructs an activation service from the supplied dependencies
func NewService(keymakerSvc *keymaker.Service, store CredentialWriter, operations *keystore.OperationManager) (*Service, error) {
	if store == nil {
		return nil, ErrStoreRequired
	}
	if keymakerSvc == nil {
		return nil, ErrKeymakerRequired
	}

	return &Service{
		keymaker:   keymakerSvc,
		store:      store,
		operations: operations,
	}, nil
}

// BeginOAuthRequest starts an OAuth/OIDC activation flow
type BeginOAuthRequest struct {
	// OrgID identifies the organization initiating the flow
	OrgID string
	// IntegrationID optionally identifies the integration record being activated
	IntegrationID string
	// Provider specifies which provider to authorize
	Provider types.ProviderType
	// RedirectURI overrides the default callback URL when needed
	RedirectURI string
	// Scopes optionally override the provider default scopes
	Scopes []string
	// Metadata carries optional provider-specific metadata
	Metadata map[string]any
	// LabelOverrides customizes UI labels presented to the user
	LabelOverrides map[string]string
	// State optionally supplies a pre-generated OAuth state value
	State string
}

// BeginOAuthResponse returns the authorization URL/state pair
type BeginOAuthResponse struct {
	// Provider identifies which provider issued the authorization URL
	Provider types.ProviderType
	// State carries the CSRF state value for the flow
	State string
	// AuthURL is the URL the user should visit to authorize
	AuthURL string
}

// BeginOAuth starts an OAuth/OIDC transaction with the requested provider
func (s *Service) BeginOAuth(ctx context.Context, req BeginOAuthRequest) (BeginOAuthResponse, error) {
	begin, err := s.keymaker.BeginAuthorization(ctx, keymaker.BeginRequest{
		OrgID:          req.OrgID,
		IntegrationID:  req.IntegrationID,
		Provider:       req.Provider,
		RedirectURI:    req.RedirectURI,
		Scopes:         append([]string(nil), req.Scopes...),
		Metadata:       lo.Assign(map[string]any{}, req.Metadata),
		LabelOverrides: lo.Assign(map[string]string{}, req.LabelOverrides),
		State:          strings.TrimSpace(req.State),
	})
	if err != nil {
		return BeginOAuthResponse{}, err
	}

	return BeginOAuthResponse{
		Provider: begin.Provider,
		State:    begin.State,
		AuthURL:  begin.AuthURL,
	}, nil
}

// CompleteOAuthRequest finalizes an OAuth/OIDC activation flow
type CompleteOAuthRequest struct {
	// State is the CSRF state value returned by the provider
	State string
	// Code is the authorization code returned by the provider
	Code string
}

// CompleteOAuthResult reports the persisted credential and related identifiers
type CompleteOAuthResult struct {
	// Provider identifies which provider issued the credential
	Provider types.ProviderType
	// OrgID identifies the organization that owns the credential
	OrgID string
	// IntegrationID identifies the integration record containing the credential
	IntegrationID string
	// Credential contains the persisted credential payload
	Credential types.CredentialPayload
}

// CompleteOAuth finalizes an OAuth/OIDC transaction and persists credentials
func (s *Service) CompleteOAuth(ctx context.Context, req CompleteOAuthRequest) (CompleteOAuthResult, error) {
	result, err := s.keymaker.CompleteAuthorization(ctx, keymaker.CompleteRequest{
		State: req.State,
		Code:  req.Code,
	})
	if err != nil {
		return CompleteOAuthResult{}, err
	}

	return CompleteOAuthResult{
		Provider:      result.Provider,
		OrgID:         result.OrgID,
		IntegrationID: result.IntegrationID,
		Credential:    result.Credential,
	}, nil
}

// ConfigureRequest carries the information required to persist non-OAuth credentials
type ConfigureRequest struct {
	// OrgID identifies the organization initiating the configuration
	OrgID string
	// Provider specifies which provider to configure
	Provider types.ProviderType
	// ProviderData carries provider-specific configuration values
	ProviderData map[string]any
	// Validate controls whether a health check should be executed
	Validate bool
}

// ConfigureResult reports the persisted credential and optional health result
type ConfigureResult struct {
	// Credential contains the persisted credential payload
	Credential types.CredentialPayload
	// HealthResult captures the optional health check result
	HealthResult *types.OperationResult
}

// Configure persists non-OAuth credentials and optionally runs a health check
func (s *Service) Configure(ctx context.Context, req ConfigureRequest) (ConfigureResult, error) {
	if req.OrgID == "" {
		return ConfigureResult{}, keystore.ErrOrgIDRequired
	}
	if req.Provider == types.ProviderUnknown {
		return ConfigureResult{}, types.ErrProviderTypeRequired
	}

	payload, err := types.NewCredentialBuilder(req.Provider).
		With(
			types.WithCredentialKind(types.CredentialKindMetadata),
			types.WithCredentialSet(models.CredentialSet{
				ProviderData: lo.Assign(map[string]any{}, req.ProviderData),
			}),
		).
		Build()
	if err != nil {
		return ConfigureResult{}, err
	}

	saved, err := s.store.SaveCredential(ctx, req.OrgID, payload)
	if err != nil {
		return ConfigureResult{}, err
	}

	result := ConfigureResult{Credential: saved}
	if !req.Validate {
		return result, nil
	}

	if s.operations == nil {
		return ConfigureResult{}, ErrOperationsRequired
	}

	health, err := s.operations.Run(ctx, types.OperationRequest{
		OrgID:    req.OrgID,
		Provider: req.Provider,
		Name:     defaultHealthOperation,
		Force:    true,
	})
	if err != nil {
		return ConfigureResult{}, err
	}

	result.HealthResult = &health
	if health.Status != types.OperationStatusOK {
		return result, ErrHealthCheckFailed
	}

	return result, nil
}
