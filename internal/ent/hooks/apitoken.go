package hooks

import (
	"context"
	"fmt"
	"time"

	"entgo.io/ent"
	"github.com/samber/lo"
	"github.com/theopenlane/iam/fgax"

	"github.com/theopenlane/iam/auth"

	fgamodel "github.com/theopenlane/core/fga/model"
	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/generated/hook"
	"github.com/theopenlane/core/pkg/logx"
)

func validateExpirationTime(m mutationWithExpirationTime) error {
	t, ok := m.ExpiresAt()
	if !ok {
		return nil
	}

	if t.Before(time.Now()) {
		return ErrPastTimeNotAllowed
	}

	return nil
}

type mutationWithExpirationTime interface {
	ExpiresAt() (time.Time, bool)
}

// HookCreateAPIToken runs on api token mutations and sets the owner id
func HookCreateAPIToken() ent.Hook {
	return hook.On(func(next ent.Mutator) ent.Mutator {
		return hook.APITokenFunc(func(ctx context.Context, m *generated.APITokenMutation) (generated.Value, error) {
			orgID, err := auth.GetOrganizationIDFromContext(ctx)
			if err != nil {
				return nil, err
			}

			// set organization on the token
			m.SetOwnerID(orgID)

			if err := validateExpirationTime(m); err != nil {
				return nil, err
			}

			retVal, err := next.Mutate(ctx, m)
			if err != nil {
				return nil, err
			}

			token, ok := retVal.(*generated.APIToken)
			if !ok {
				return retVal, err
			}

			// create the relationship tuples in fga for the token
			tuples, err := createScopeTuples(ctx, token.Scopes, orgID, token.ID)
			if err != nil {
				return retVal, err
			}

			// add self relation tuple for the token
			req := fgax.TupleRequest{
				SubjectID:   token.ID,
				SubjectType: auth.ServiceSubjectType,
				ObjectID:    token.ID,
				ObjectType:  auth.ServiceSubjectType,
				Relation:    fgax.SelfRelation,
			}

			tuples = append(tuples, fgax.GetTupleKey(req))

			// create the relationship tuples if we have any
			if len(tuples) > 0 {
				if err := batchWriteTuples(ctx, m.Authz, tuples, nil); err != nil {
					logx.FromContext(ctx).Error().Err(err).Msg("failed to create relationship tuple")

					return nil, err
				}
			}

			return retVal, err
		})
	}, ent.OpCreate)
}

// HookUpdateAPIToken runs on api token update and redacts the token
func HookUpdateAPIToken() ent.Hook {
	return hook.On(func(next ent.Mutator) ent.Mutator {
		return hook.APITokenFunc(func(ctx context.Context, m *generated.APITokenMutation) (generated.Value, error) {
			var oldScopes []string
			var scopesModified bool

			// Only query old scopes if scopes are being modified and this is an UpdateOne operation
			if _, scopesModified = m.Scopes(); scopesModified && m.Op().Is(ent.OpUpdateOne) {
				var err error
				oldScopes, err = m.OldScopes(ctx)
				if err != nil {
					return nil, err
				}
			}

			retVal, err := next.Mutate(ctx, m)
			if err != nil {
				return nil, err
			}

			// redact the token
			at, ok := retVal.(*generated.APIToken)
			if !ok {
				return retVal, nil
			}

			at.Token = redacted

			// Only update scope tuples if scopes were modified
			if scopesModified {
				scopeSet, err := fgamodel.DefaultServiceScopeSet()
				if err != nil {
					return nil, fmt.Errorf("failed to load available token scopes from model: %w", err)
				}

				addedScopes, removedScopes := diffScopes(oldScopes, at.Scopes)

				addTuples, err := scopeTuples(ctx, addedScopes, at.OwnerID, at.ID, scopeSet)
				if err != nil {
					return nil, err
				}

				removeTuples, err := scopeTuples(ctx, removedScopes, at.OwnerID, at.ID, scopeSet)
				if err != nil {
					return nil, err
				}

				if len(addTuples) > 0 || len(removeTuples) > 0 {
					if err := batchWriteTuples(ctx, m.Authz, addTuples, removeTuples); err != nil {
						logx.FromContext(ctx).Error().Err(err).Msg("failed to update api token scope tuples")

						return nil, err
					}
				}
			}

			return at, nil
		})
	}, ent.OpUpdate|ent.OpUpdateOne)
}

// createScopeTuples creates the relationship tuples for the token
func createScopeTuples(ctx context.Context, scopes []string, orgID, tokenID string) ([]fgax.TupleKey, error) {
	scopeSet, err := fgamodel.DefaultServiceScopeSet()
	if err != nil {
		return nil, fmt.Errorf("failed to load available token scopes from model: %w", err)
	}

	return scopeTuples(ctx, scopes, orgID, tokenID, scopeSet)
}

// scopeTuples creates relationship tuples for the given scopes
func scopeTuples(ctx context.Context, scopes []string, orgID, tokenID string, scopeSet map[string]struct{}) ([]fgax.TupleKey, error) {
	var tuples []fgax.TupleKey

	for _, scope := range scopes {
		relation := fgamodel.NormalizeScope(scope)

		if relation == "" {
			logx.FromContext(ctx).Warn().Str("scope", scope).Msg("ignoring empty scope on api token")

			continue
		}

		if _, ok := scopeSet[relation]; !ok {
			return nil, fmt.Errorf("%w: %q (%s)", ErrInvalidScope, scope, relation)
		}

		req := fgax.TupleRequest{
			SubjectID:   tokenID,
			SubjectType: auth.ServiceSubjectType,
			ObjectID:    orgID,
			ObjectType:  generated.TypeOrganization,
			Relation:    relation,
		}

		tuples = append(tuples, fgax.GetTupleKey(req))
	}

	return tuples, nil
}

// diffScopes returns the added and removed scopes between two scope slices
func diffScopes(oldScopes, newScopes []string) (added []string, removed []string) {
	// lo for the win
	added, _ = lo.Difference(newScopes, oldScopes)
	removed, _ = lo.Difference(oldScopes, newScopes)

	return
}

const (
	// maxFGATuplesPerBatch is the maximum number of tuples that can be written to FGA in a single batch
	maxFGATuplesPerBatch = 100
)

// batchWriteTuples writes tuples to FGA in batches of maxFGATuplesPerBatch to avoid exceeding OpenFGA's limit
func batchWriteTuples(ctx context.Context, authz fgax.Client, addTuples, removeTuples []fgax.TupleKey) error {
	// Process additions in batches
	for i := 0; i < len(addTuples); i += maxFGATuplesPerBatch {
		end := i + maxFGATuplesPerBatch
		if end > len(addTuples) {
			end = len(addTuples)
		}

		batch := addTuples[i:end]

		if _, err := authz.WriteTupleKeys(ctx, batch, nil); err != nil {
			return err
		}
	}

	// Process removals in batches
	for i := 0; i < len(removeTuples); i += maxFGATuplesPerBatch {
		end := i + maxFGATuplesPerBatch
		if end > len(removeTuples) {
			end = len(removeTuples)
		}

		batch := removeTuples[i:end]

		if _, err := authz.WriteTupleKeys(ctx, nil, batch); err != nil {
			return err
		}
	}

	return nil
}
