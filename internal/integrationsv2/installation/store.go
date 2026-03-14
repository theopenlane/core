package installation

import (
	"context"

	"github.com/theopenlane/core/common/enums"
	ent "github.com/theopenlane/core/internal/ent/generated"
	entintegration "github.com/theopenlane/core/internal/ent/generated/integration"
	"github.com/theopenlane/core/internal/integrationsv2/types"
)

// CreateParams describes one persisted installation record
type CreateParams struct {
	// OwnerID is the organization identifier for org-owned installations
	OwnerID string
	// SystemOwned indicates the installation is platform-owned
	SystemOwned bool
	// Name is the installation label
	Name string
	// Description is the optional installation description
	Description string
	// Definition is the registered definition metadata captured on the installation
	Definition types.DefinitionSpec
	// Status is the lifecycle state persisted on the installation
	Status enums.IntegrationStatus
	// Metadata is the operator-facing installation metadata payload
	Metadata map[string]any
	// DefinitionMetadataSnapshot is the definition metadata snapshot stored on the installation
	DefinitionMetadataSnapshot map[string]any
}

// Store persists installation rows in Ent
type Store struct {
	db *ent.Client
}

// NewStore constructs the installation store
func NewStore(db *ent.Client) (*Store, error) {
	if db == nil {
		return nil, ErrDBClientRequired
	}

	return &Store{db: db}, nil
}

// Create inserts one installation row
func (s *Store) Create(ctx context.Context, params CreateParams) (*ent.Integration, error) {
	if params.Definition.ID == "" {
		return nil, ErrDefinitionIDRequired
	}

	if params.Name == "" {
		return nil, ErrInstallationNameRequired
	}

	if !params.SystemOwned && params.OwnerID == "" {
		return nil, ErrOwnerIDRequired
	}

	status := params.Status
	if status == "" {
		status = enums.IntegrationStatusPending
	}

	builder := s.db.Integration.Create().
		SetName(params.Name).
		SetDescription(params.Description).
		SetDefinitionID(string(params.Definition.ID)).
		SetDefinitionVersion(params.Definition.Version).
		SetDefinitionSlug(params.Definition.Slug).
		SetFamily(params.Definition.Family).
		SetStatus(status).
		SetProviderMetadataSnapshot(cloneMap(params.DefinitionMetadataSnapshot)).
		SetMetadata(cloneMap(params.Metadata))

	if params.SystemOwned {
		builder = builder.SetSystemOwned(true)
	}

	if params.OwnerID != "" {
		builder = builder.SetOwnerID(params.OwnerID)
	}

	return builder.Save(ctx)
}

// Get resolves one installation by id
func (s *Store) Get(ctx context.Context, installationID string) (*ent.Integration, error) {
	if installationID == "" {
		return nil, ErrInstallationIDRequired
	}

	return s.db.Integration.Get(ctx, installationID)
}

// ListByOwner resolves all installations for one owner in stable update order
func (s *Store) ListByOwner(ctx context.Context, ownerID string) ([]*ent.Integration, error) {
	if ownerID == "" {
		return nil, ErrOwnerIDRequired
	}

	return s.db.Integration.Query().
		Where(entintegration.OwnerIDEQ(ownerID)).
		Order(ent.Desc(entintegration.FieldUpdatedAt), ent.Desc(entintegration.FieldCreatedAt)).
		All(ctx)
}

// UpdateStatus updates one installation lifecycle status
func (s *Store) UpdateStatus(ctx context.Context, installationID string, status enums.IntegrationStatus) (*ent.Integration, error) {
	if installationID == "" {
		return nil, ErrInstallationIDRequired
	}

	if status == "" {
		status = enums.IntegrationStatusPending
	}

	return s.db.Integration.UpdateOneID(installationID).
		SetStatus(status).
		Save(ctx)
}

// Delete removes one installation row
func (s *Store) Delete(ctx context.Context, installationID string) error {
	if installationID == "" {
		return ErrInstallationIDRequired
	}

	return s.db.Integration.DeleteOneID(installationID).Exec(ctx)
}

// cloneMap copies one JSON-style map so callers do not retain shared state
func cloneMap(value map[string]any) map[string]any {
	if len(value) == 0 {
		return map[string]any{}
	}

	out := make(map[string]any, len(value))
	for key, item := range value {
		out[key] = item
	}

	return out
}
