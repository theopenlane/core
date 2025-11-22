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
				if _, err := m.Authz.WriteTupleKeys(ctx, tuples, nil); err != nil {
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

			// create the relationship tuples in fga for the token
			newScopes, err := getNewScopes(ctx, m)
			if err != nil {
				return at, err
			}

			tuples, err := createScopeTuples(ctx, newScopes, at.OwnerID, at.ID)
			if err != nil {
				return retVal, err
			}

			// create the relationship tuples if we have any
			if len(tuples) > 0 {
				if _, err := m.Authz.WriteTupleKeys(ctx, tuples, nil); err != nil {
					logx.FromContext(ctx).Error().Err(err).Msg("failed to create relationship tuple")

					return nil, err
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

	var tuples []fgax.TupleKey

	for _, scope := range scopes {
		relation := fgamodel.NormalizeScope(scope)

		if relation == "" {
			logx.FromContext(ctx).Warn().Str("scope", scope).Msg("ignoring empty scope on api token")

			continue
		}

		if _, ok := scopeSet[relation]; !ok {
			return nil, fmt.Errorf("scope %q (%s) is not assignable to service subjects", scope, relation)
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

// getNewScopes returns the new scopes that were added to the token during an update
// NOTE: there is an AppendedScopes on the mutation, but this is not populated
// so calculating the new scopes for now
func getNewScopes(ctx context.Context, m *generated.APITokenMutation) ([]string, error) {
	scopes, ok := m.Scopes()
	if !ok {
		return nil, nil
	}

	oldScopes, err := m.OldScopes(ctx)
	if err != nil {
		return nil, err
	}

	var newScopes []string

	for _, scope := range scopes {
		if !lo.Contains(oldScopes, scope) {
			newScopes = append(newScopes, scope)
		}
	}

	return newScopes, nil
}
