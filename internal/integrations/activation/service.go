package activation

import (
	"context"
	"encoding/json"

	"github.com/theopenlane/core/common/models"
	"github.com/theopenlane/core/internal/integrations"
	"github.com/theopenlane/core/internal/integrations/types"
	"github.com/theopenlane/core/pkg/jsonx"
)

// CredentialWriter persists credential payloads produced during activation
type CredentialWriter interface {
	SaveCredential(ctx context.Context, orgID string, provider types.ProviderType, authKind types.AuthKind, credential models.CredentialSet) (models.CredentialSet, error)
}

// HealthValidator validates provider connectivity using a supplied credential payload.
type HealthValidator interface {
	// ValidateProviderHealth executes the provider health operation using the supplied payload.
	ValidateProviderHealth(ctx context.Context, orgID string, provider types.ProviderType, payload models.CredentialSet) (types.OperationResult, error)
}

// CredentialMinter mints credentials from an in-memory credential set without accessing the credential store.
type CredentialMinter interface {
	// MintCredential exchanges or refreshes the supplied in-memory credential set via the provider.
	MintCredential(ctx context.Context, request types.CredentialMintRequest) (models.CredentialSet, error)
}

// Service coordinates non-OAuth provider configuration and health validation
type Service struct {
	// store persists credential payloads after successful configuration.
	store CredentialWriter
	// validator executes provider health checks.
	validator HealthValidator
	// minter mints credentials from an in-memory credential set for pre-persist validation.
	minter CredentialMinter
}

// NewService constructs an activation service from the supplied dependencies
func NewService(store CredentialWriter, validator HealthValidator, minter CredentialMinter) (*Service, error) {
	if store == nil {
		return nil, ErrStoreRequired
	}
	if minter == nil {
		return nil, ErrMinterRequired
	}

	return &Service{
		store:     store,
		validator: validator,
		minter:    minter,
	}, nil
}

// ConfigureRequest carries the information required to persist non-OAuth credentials
type ConfigureRequest struct {
	// OrgID identifies the organization initiating the configuration
	OrgID string
	// Provider specifies which provider to configure
	Provider types.ProviderType
	// AuthKind is the concrete provider auth kind used for persisted credential ownership.
	AuthKind types.AuthKind
	// ProviderData carries provider-specific configuration values
	ProviderData json.RawMessage
	// Validate controls whether a health check should be executed
	Validate bool
}

// ConfigureResult reports the persisted credential and optional health result
type ConfigureResult struct {
	// Credential contains the persisted credential payload
	Credential models.CredentialSet
	// HealthResult captures the optional health check result
	HealthResult *types.OperationResult
}

// Configure validates connectivity with the provider and persists credentials only on success
func (s *Service) Configure(ctx context.Context, req ConfigureRequest) (ConfigureResult, error) {
	if req.OrgID == "" {
		return ConfigureResult{}, integrations.ErrOrgIDRequired
	}
	if req.Provider == types.ProviderUnknown {
		return ConfigureResult{}, types.ErrProviderTypeRequired
	}

	credential := models.CredentialSet{
		ProviderData: jsonx.CloneRawMessage(req.ProviderData),
	}

	if !req.Validate {
		saved, err := s.store.SaveCredential(ctx, req.OrgID, req.Provider, req.AuthKind, credential)
		if err != nil {
			return ConfigureResult{}, err
		}

		return ConfigureResult{Credential: saved}, nil
	}

	if s.validator == nil {
		return ConfigureResult{}, ErrHealthValidatorRequired
	}

	minted, err := s.minter.MintCredential(ctx, types.CredentialMintRequest{
		Provider:   req.Provider,
		OrgID:      req.OrgID,
		Credential: credential,
	})
	if err != nil {
		return ConfigureResult{}, err
	}

	health, err := s.validator.ValidateProviderHealth(ctx, req.OrgID, req.Provider, minted)
	if err != nil {
		return ConfigureResult{}, err
	}

	if health.Status != types.OperationStatusOK {
		return ConfigureResult{HealthResult: &health}, ErrHealthCheckFailed
	}

	// Persist provider-minted fields while preserving submitted provider metadata.
	persisted := types.CloneCredentialSet(minted)
	if len(credential.ProviderData) > 0 && len(persisted.ProviderData) == 0 {
		persisted.ProviderData = jsonx.CloneRawMessage(credential.ProviderData)
	}

	authKind := req.AuthKind.Normalize()
	if authKind == types.AuthKindUnknown {
		authKind = types.InferAuthKind(persisted)
	}

	saved, err := s.store.SaveCredential(ctx, req.OrgID, req.Provider, authKind, persisted)
	if err != nil {
		return ConfigureResult{}, err
	}

	return ConfigureResult{Credential: saved, HealthResult: &health}, nil
}
