package runtime

import (
	"context"

	"github.com/samber/do/v2"

	ent "github.com/theopenlane/core/internal/ent/generated"
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

// LoadCredential resolves persisted credentials for one installation
func (r *Runtime) LoadCredential(ctx context.Context, installation *ent.Integration) (types.CredentialSet, bool, error) {
	return do.MustInvoke[*keystore.Store](r.injector).LoadCredential(ctx, installation)
}

// SaveCredential persists credentials for one installation and evicts cached clients for that installation
func (r *Runtime) SaveCredential(ctx context.Context, installation *ent.Integration, credential types.CredentialSet) error {
	if installation == nil {
		return ErrInstallationRequired
	}

	return do.MustInvoke[*keystore.Store](r.injector).SaveCredential(ctx, installation, credential)
}

// SaveInstallationCredential persists credentials for one installation identifier and evicts cached clients
func (r *Runtime) SaveInstallationCredential(ctx context.Context, installationID string, credential types.CredentialSet) error {
	return do.MustInvoke[*keystore.Store](r.injector).SaveInstallationCredential(ctx, installationID, credential)
}

// DeleteCredential removes credentials for one installation identifier and evicts cached clients
func (r *Runtime) DeleteCredential(ctx context.Context, installationID string) error {
	return do.MustInvoke[*keystore.Store](r.injector).DeleteCredential(ctx, installationID)
}
