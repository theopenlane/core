package hooks

import (
	"context"
	"errors"
	"time"

	"entgo.io/ent"
	"github.com/theopenlane/iam/auth"
	"github.com/theopenlane/logx"

	"github.com/theopenlane/core/v2/internal/ent/generated"
	"github.com/theopenlane/core/v2/internal/ent/generated/hook"
	"github.com/theopenlane/core/v2/internal/oauthserver"
	"github.com/theopenlane/core/v2/internal/oauthserver/store"
)

// ErrOAuthGrantRevoked is returned when anything tries to update a revoked grant
var ErrOAuthGrantRevoked = errors.New("a revoked application can't be changed, approve it again to reconnect")

// HookCreateOauthGrant writes the service subject tuples of the grant's object scopes, like an API token
func HookCreateOauthGrant() ent.Hook {
	return hook.On(func(next ent.Mutator) ent.Mutator {
		return hook.OauthGrantFunc(func(ctx context.Context, m *generated.OauthGrantMutation) (generated.Value, error) {
			retVal, err := next.Mutate(ctx, m)
			if err != nil {
				return nil, err
			}

			grant, ok := retVal.(*generated.OauthGrant)
			if !ok {
				return retVal, nil
			}

			if err := writeServiceSubjectTuples(ctx, m.Authz, oauthserver.ObjectScopes(grant.Scopes), grant.OwnerID, grant.ID); err != nil {
				return nil, err
			}

			return retVal, nil
		})
	}, ent.OpCreate)
}

// HookUpdateOauthGrant records and enforces one-way revocation and keeps the grant's scope tuples in sync
func HookUpdateOauthGrant() ent.Hook {
	return hook.On(func(next ent.Mutator) ent.Mutator {
		return hook.OauthGrantFunc(func(ctx context.Context, m *generated.OauthGrantMutation) (generated.Value, error) {
			oldActive, err := m.OldIsActive(ctx)
			if err != nil {
				return nil, err
			}

			// revocation is terminal, approving the application again creates a new grant
			if !oldActive {
				return nil, ErrOAuthGrantRevoked
			}

			active, activeSet := m.IsActive()
			revoking := activeSet && !active

			if revoking {
				setRevokedFields(ctx, m)
			}

			oldScopes, scopesModified, err := oldScopesIfModified(ctx, m)
			if err != nil {
				return nil, err
			}

			retVal, err := next.Mutate(ctx, m)
			if err != nil {
				return nil, err
			}

			grant, ok := retVal.(*generated.OauthGrant)
			if !ok {
				return retVal, nil
			}

			if revoking {
				if err := revokeOAuthGrantTokens(ctx, m.Client(), grant.ID); err != nil {
					return nil, err
				}
			}

			if scopesModified {
				if err := syncScopeTuples(ctx, m.Authz, oauthserver.ObjectScopes(oldScopes), oauthserver.ObjectScopes(grant.Scopes), grant.OwnerID, grant.ID); err != nil {
					return nil, err
				}
			}

			return retVal, nil
		})
	}, ent.OpUpdateOne)
}

// HookDeleteOauthGrant ends all tokens of deleted grants
func HookDeleteOauthGrant() ent.Hook {
	return hook.On(func(next ent.Mutator) ent.Mutator {
		return hook.OauthGrantFunc(func(ctx context.Context, m *generated.OauthGrantMutation) (generated.Value, error) {
			if !isDeleteOp(ctx, m) {
				return next.Mutate(ctx, m)
			}

			ids, err := getMutationIDs(ctx, m)
			if err != nil {
				return nil, err
			}

			retVal, err := next.Mutate(ctx, m)
			if err != nil {
				return nil, err
			}

			for _, id := range ids {
				if err := revokeOAuthGrantTokens(ctx, m.Client(), id); err != nil {
					return nil, err
				}
			}

			return retVal, nil
		})
	}, ent.OpDelete|ent.OpDeleteOne|ent.OpUpdate|ent.OpUpdateOne)
}

// setRevokedFields records when and by whom a grant is revoked, unless the caller set them
func setRevokedFields(ctx context.Context, m *generated.OauthGrantMutation) {
	if _, set := m.RevokedAt(); !set {
		m.SetRevokedAt(time.Now())
	}

	if caller, ok := auth.CallerFromContext(ctx); ok && caller.SubjectID != "" {
		if _, set := m.RevokedBy(); !set {
			m.SetRevokedBy(caller.SubjectID)
		}
	}
}

// revokeOAuthGrantTokens ends every outstanding token under a grant so none can be used again
func revokeOAuthGrantTokens(ctx context.Context, client *generated.Client, grantID string) error {
	if err := store.RevokeGrantTokens(auth.WithInternalOperationContext(ctx), client, grantID); err != nil {
		logx.FromContext(ctx).Error().Err(err).Str("grant_id", grantID).Msg("failed to revoke oauth grant tokens")

		return ErrInternalServerError
	}

	return nil
}
