package runtime

import (
	"context"
	"encoding/json"

	"github.com/samber/do/v2"

	ent "github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/integrations/types"
	"github.com/theopenlane/core/pkg/jsonx"
)

// ResolveInstallationMetadata derives installation metadata for one installation when the definition declares it.
func (r *Runtime) ResolveInstallationMetadata(ctx context.Context, installation *ent.Integration, definition types.Definition, credential types.CredentialSet, input json.RawMessage) (types.IntegrationInstallationMetadata, bool, error) {
	if installation == nil {
		return types.IntegrationInstallationMetadata{}, false, ErrInstallationRequired
	}

	if definition.Installation == nil || definition.Installation.Resolve == nil {
		return types.IntegrationInstallationMetadata{}, false, nil
	}

	metadata, err := definition.Installation.Resolve(ctx, types.InstallationRequest{
		Installation: installation,
		Credential:   credential,
		Config:       installation.Config,
		Input:        input,
	})
	if err != nil {
		return types.IntegrationInstallationMetadata{}, false, err
	}

	if len(metadata.Attributes) == 0 {
		return types.IntegrationInstallationMetadata{}, false, nil
	}

	if len(definition.Installation.Schema) > 0 {
		result, err := jsonx.ValidateSchema(definition.Installation.Schema, metadata.Attributes)
		if err != nil {
			return types.IntegrationInstallationMetadata{}, false, err
		}

		if !result.Valid() {
			return types.IntegrationInstallationMetadata{}, false, ErrInstallationMetadataInvalid
		}
	}

	return metadata, true, nil
}

// SaveInstallationMetadata persists installation metadata for one installation.
func (r *Runtime) SaveInstallationMetadata(ctx context.Context, installation *ent.Integration, metadata types.IntegrationInstallationMetadata) error {
	if installation == nil {
		return ErrInstallationRequired
	}

	if len(metadata.Attributes) == 0 {
		return nil
	}

	if err := do.MustInvoke[*ent.Client](r.injector).Integration.UpdateOneID(installation.ID).SetInstallationMetadata(metadata).Exec(ctx); err != nil {
		return err
	}

	installation.InstallationMetadata = metadata

	return nil
}

// ResolveAndSaveInstallationMetadata derives and persists installation metadata for one installation.
func (r *Runtime) ResolveAndSaveInstallationMetadata(ctx context.Context, installation *ent.Integration, definition types.Definition, credential types.CredentialSet, input json.RawMessage) (types.IntegrationInstallationMetadata, bool, error) {
	metadata, ok, err := r.ResolveInstallationMetadata(ctx, installation, definition, credential, input)
	if err != nil || !ok {
		return metadata, ok, err
	}

	if err := r.SaveInstallationMetadata(ctx, installation, metadata); err != nil {
		return types.IntegrationInstallationMetadata{}, false, err
	}

	return metadata, true, nil
}
