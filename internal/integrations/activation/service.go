package activation

import (
	"context"
	"maps"

	"github.com/theopenlane/core/common/models"
	"github.com/theopenlane/core/internal/integrations"
	"github.com/theopenlane/core/internal/integrations/types"
)

// CredentialWriter persists credential payloads produced during activation
type CredentialWriter interface {
	SaveCredential(ctx context.Context, orgID string, payload types.CredentialPayload) (types.CredentialPayload, error)
}

// HealthValidator validates provider connectivity using a supplied credential payload.
type HealthValidator interface {
	// ValidateProviderHealth executes the provider health operation using the supplied payload.
	ValidateProviderHealth(ctx context.Context, orgID string, provider types.ProviderType, payload types.CredentialPayload) (types.OperationResult, error)
}

// PayloadMinter mints credentials from an in-memory payload without accessing the credential store
type PayloadMinter interface {
	// MintPayload exchanges or refreshes the supplied in-memory credential payload via the provider
	MintPayload(ctx context.Context, subject types.CredentialSubject) (types.CredentialPayload, error)
}

// Service coordinates non-OAuth provider configuration and health validation
type Service struct {
	// store persists credential payloads after successful configuration.
	store CredentialWriter
	// validator executes provider health checks.
	validator HealthValidator
	// minter mints credentials from an in-memory payload for pre-persist validation.
	minter PayloadMinter
}

// NewService constructs an activation service from the supplied dependencies
func NewService(store CredentialWriter, validator HealthValidator, minter PayloadMinter) (*Service, error) {
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

// Configure validates connectivity with the provider and persists credentials only on success
func (s *Service) Configure(ctx context.Context, req ConfigureRequest) (ConfigureResult, error) {
	if req.OrgID == "" {
		return ConfigureResult{}, integrations.ErrOrgIDRequired
	}
	if req.Provider == types.ProviderUnknown {
		return ConfigureResult{}, types.ErrProviderTypeRequired
	}

	payload, err := types.NewCredentialBuilder(req.Provider).
		With(
			types.WithCredentialKind(types.CredentialKindMetadata),
			types.WithCredentialSet(models.CredentialSet{
				ProviderData: maps.Clone(req.ProviderData),
			}),
		).
		Build()
	if err != nil {
		return ConfigureResult{}, err
	}

	if !req.Validate {
		saved, err := s.store.SaveCredential(ctx, req.OrgID, payload)
		if err != nil {
			return ConfigureResult{}, err
		}

		return ConfigureResult{Credential: saved}, nil
	}

	if s.validator == nil {
		return ConfigureResult{}, ErrHealthValidatorRequired
	}

	minted, err := s.minter.MintPayload(ctx, types.CredentialSubject{
		Provider:   req.Provider,
		OrgID:      req.OrgID,
		Credential: payload,
	})
	if err != nil {
		return ConfigureResult{}, err
	}
	if minted.Provider == types.ProviderUnknown {
		minted.Provider = req.Provider
	}

	health, err := s.validator.ValidateProviderHealth(ctx, req.OrgID, req.Provider, minted)
	if err != nil {
		return ConfigureResult{}, err
	}

	if health.Status != types.OperationStatusOK {
		return ConfigureResult{HealthResult: &health}, ErrHealthCheckFailed
	}

	// Persist provider-minted fields while preserving submitted provider metadata.
	persisted := minted
	if persisted.Kind == "" {
		persisted.Kind = payload.Kind
	}
	if len(payload.Data.ProviderData) > 0 {
		if len(persisted.Data.ProviderData) == 0 {
			persisted.Data.ProviderData = maps.Clone(payload.Data.ProviderData)
		} else {
			for key, value := range payload.Data.ProviderData {
				if _, exists := persisted.Data.ProviderData[key]; !exists {
					persisted.Data.ProviderData[key] = value
				}
			}
		}
	}

	saved, err := s.store.SaveCredential(ctx, req.OrgID, persisted)
	if err != nil {
		return ConfigureResult{}, err
	}

	return ConfigureResult{Credential: saved, HealthResult: &health}, nil
}
