package keystore

import (
	"context"
	"errors"

	ent "github.com/theopenlane/core/internal/ent/generated"
	enthush "github.com/theopenlane/core/internal/ent/generated/hush"
	entintegration "github.com/theopenlane/core/internal/ent/generated/integration"
	"github.com/theopenlane/core/internal/ent/generated/privacy"
	newtypes "github.com/theopenlane/core/internal/integrationsv2/types"
	"github.com/theopenlane/iam/auth"
)

// IntegrationV2CredentialResolver adapts keystore persistence to the integrationsv2 credential contract
type IntegrationV2CredentialResolver struct {
	store *Store
}

// NewIntegrationV2CredentialResolver constructs the integrationsv2 credential adapter
func NewIntegrationV2CredentialResolver(store *Store) (*IntegrationV2CredentialResolver, error) {
	if store == nil {
		return nil, ErrStoreNotInitialized
	}

	return &IntegrationV2CredentialResolver{store: store}, nil
}

// LoadCredential resolves persisted credentials for one installation record
func (r *IntegrationV2CredentialResolver) LoadCredential(ctx context.Context, integration *ent.Integration) (newtypes.CredentialSet, bool, error) {
	if integration == nil {
		return newtypes.CredentialSet{}, false, ErrCredentialNotFound
	}

	record, err := r.store.db.Hush.Query().
		Where(enthush.HasIntegrationsWith(entintegration.IDEQ(integration.ID))).
		Only(integrationV2SystemContext(ctx))
	if err != nil {
		if ent.IsNotFound(err) {
			return newtypes.CredentialSet{}, false, nil
		}

		if errors.Is(err, ErrCredentialNotFound) {
			return newtypes.CredentialSet{}, false, nil
		}

		return newtypes.CredentialSet{}, false, err
	}

	return newtypes.CredentialSet(record.CredentialSet), true, nil
}

// integrationV2SystemContext applies privileged auth and privacy decisions for integrationsv2 credential reads
func integrationV2SystemContext(ctx context.Context) context.Context {
	callCtx := privacy.DecisionContext(ctx, privacy.Allow)
	return auth.WithCaller(callCtx, auth.NewKeystoreCaller())
}
