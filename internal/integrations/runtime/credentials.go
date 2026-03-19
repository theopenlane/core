package runtime

import (
	"context"

	"github.com/samber/do/v2"
	"github.com/samber/lo"

	"github.com/theopenlane/core/common/enums"
	ent "github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/generated/privacy"
	"github.com/theopenlane/core/internal/integrations/types"
	"github.com/theopenlane/core/internal/keymaker"
	"github.com/theopenlane/core/internal/keystore"
)

// BeginAuth starts one definition auth flow through the runtime-managed keymaker service
func (r *Runtime) BeginAuth(ctx context.Context, req keymaker.BeginRequest) (keymaker.BeginResponse, error) {
	return do.MustInvoke[*keymaker.Service](r.injector).BeginAuth(ctx, req)
}

// CompleteAuth completes one definition auth flow through the runtime-managed keymaker service
func (r *Runtime) CompleteAuth(ctx context.Context, req keymaker.CompleteRequest) (keymaker.CompleteResult, error) {
	return do.MustInvoke[*keymaker.Service](r.injector).CompleteAuth(ctx, req)
}

// PersistAuthCompletion saves one completed auth result, including provider state and credentials.
func (r *Runtime) PersistAuthCompletion(ctx context.Context, installationID string, credentialRef types.CredentialRef, definition types.Definition, result types.AuthCompleteResult) error {
	if installationID == "" {
		return ErrInstallationIDRequired
	}

	systemCtx := privacy.DecisionContext(ctx, privacy.Allow)
	db := do.MustInvoke[*ent.Client](r.injector)
	store := do.MustInvoke[*keystore.Store](r.injector)
	credentialRegistration, err := resolveCredentialRegistration(definition, credentialRef)
	if err != nil {
		return err
	}

	installation, err := db.Integration.Get(systemCtx, installationID)
	if err != nil {
		if ent.IsNotFound(err) {
			return ErrInstallationNotFound
		}

		return err
	}

	if definition.Installation != nil {
		metadata, ok, err := definition.Installation.Resolve(systemCtx, types.InstallationRequest{
			Installation: installation,
			Credential:   result.Credential,
			Config:       installation.Config,
		})
		if err != nil {
			return err
		}

		if ok {
			if err := SaveInstallationMetadata(systemCtx, installation, metadata); err != nil {
				return err
			}
		}
	}

	if err := store.SaveCredential(systemCtx, installation, credentialRegistration.Ref, result.Credential); err != nil {
		return err
	}

	if err := db.Integration.UpdateOneID(installation.ID).
		SetStatus(enums.IntegrationStatusConnected).
		Exec(systemCtx); err != nil {
		return err
	}

	installation.Status = enums.IntegrationStatusConnected
	if err := r.SyncWebhooks(systemCtx, installation, ""); err != nil {
		return err
	}

	return nil
}

// LoadCredential resolves one persisted credential slot for one installation.
func (r *Runtime) LoadCredential(ctx context.Context, installation *ent.Integration, credentialRef types.CredentialRef) (types.CredentialSet, bool, error) {
	return do.MustInvoke[*keystore.Store](r.injector).LoadCredential(ctx, installation, credentialRef)
}

// LoadCredentials resolves the requested credential slots for one installation.
func (r *Runtime) LoadCredentials(ctx context.Context, installation *ent.Integration, credentialRefs []types.CredentialRef) (types.CredentialBindings, error) {
	return do.MustInvoke[*keystore.Store](r.injector).LoadCredentials(ctx, installation, credentialRefs)
}

// SaveCredential persists one credential slot for one installation and evicts cached clients for that installation.
func (r *Runtime) SaveCredential(ctx context.Context, installation *ent.Integration, credentialRef types.CredentialRef, credential types.CredentialSet) error {
	if installation == nil {
		return ErrInstallationRequired
	}

	return do.MustInvoke[*keystore.Store](r.injector).SaveCredential(ctx, installation, credentialRef, credential)
}

// SaveInstallationCredential persists one credential slot for one installation identifier and evicts cached clients.
func (r *Runtime) SaveInstallationCredential(ctx context.Context, installationID string, credentialRef types.CredentialRef, credential types.CredentialSet) error {
	return do.MustInvoke[*keystore.Store](r.injector).SaveInstallationCredential(ctx, installationID, credentialRef, credential)
}

// DeleteCredential removes credentials for one installation identifier and evicts cached clients
func (r *Runtime) DeleteCredential(ctx context.Context, installationID string) error {
	return do.MustInvoke[*keystore.Store](r.injector).DeleteCredential(ctx, installationID)
}

func resolveCredentialRegistration(definition types.Definition, credentialRef types.CredentialRef) (types.CredentialRegistration, error) {
	if !credentialRef.Valid() {
		return types.CredentialRegistration{}, ErrCredentialNotFound
	}

	registration, ok := lo.Find(definition.CredentialRegistrations, func(registration types.CredentialRegistration) bool {
		return registration.Ref.String() == credentialRef.String()
	})
	if ok {
		return registration, nil
	}

	return types.CredentialRegistration{}, ErrCredentialNotFound
}
