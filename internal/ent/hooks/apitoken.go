package hooks

import (
	"context"
	"fmt"
	"time"

	"entgo.io/ent"
	"github.com/samber/lo"
	"github.com/theopenlane/iam/fgax"

	"github.com/theopenlane/iam/auth"

	"github.com/theopenlane/core/common/models"
	fgamodel "github.com/theopenlane/core/v2/fga/model"
	"github.com/theopenlane/core/v2/internal/ent/generated"
	"github.com/theopenlane/core/v2/internal/ent/generated/hook"
	"github.com/theopenlane/logx"
)

func validateTimeNotInPast[T time.Time | models.DateTime](t T, ok bool) error {
	if !ok {
		return nil
	}

	timeValue := time.Time(t)

	if timeValue.Before(time.Now()) {
		return ErrPastTimeNotAllowed
	}

	return nil
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

			if err := validateTimeNotInPast(m.ExpiresAt()); err != nil {
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

			if err := writeServiceSubjectTuples(ctx, m.Authz, token.Scopes, orgID, token.ID); err != nil {
				return nil, err
			}

			return retVal, err
		})
	}, ent.OpCreate)
}

// HookUpdateAPIToken runs on api token update and redacts the token
func HookUpdateAPIToken() ent.Hook {
	return hook.On(func(next ent.Mutator) ent.Mutator {
		return hook.APITokenFunc(func(ctx context.Context, m *generated.APITokenMutation) (generated.Value, error) {
			oldScopes, scopesModified, err := oldScopesIfModified(ctx, m)
			if err != nil {
				return nil, err
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
				if err := syncScopeTuples(ctx, m.Authz, oldScopes, at.Scopes, at.OwnerID, at.ID); err != nil {
					return nil, err
				}
			}

			return at, nil
		})
	}, ent.OpUpdateOne)
}

// scopedMutation is a mutation of a service subject that holds scopes, an api token or an oauth grant
type scopedMutation interface {
	Op() ent.Op
	Scopes() ([]string, bool)
	AppendedScopes() ([]string, bool)
	OldScopes(ctx context.Context) ([]string, error)
}

// oldScopesIfModified returns the scopes before an UpdateOne mutation that changes them, and whether they change
func oldScopesIfModified(ctx context.Context, m scopedMutation) ([]string, bool, error) {
	_, modified := m.Scopes()
	if !modified {
		_, modified = m.AppendedScopes()
	}

	if !modified || !m.Op().Is(ent.OpUpdateOne) {
		return nil, modified, nil
	}

	oldScopes, err := m.OldScopes(ctx)

	return oldScopes, true, err
}

// writeServiceSubjectTuples writes the scope tuples and self tuple of a new service subject
func writeServiceSubjectTuples(ctx context.Context, authz fgax.Client, scopes []string, orgID, subjectID string) error {
	tuples, err := createScopeTuples(ctx, scopes, orgID, subjectID)
	if err != nil {
		return err
	}

	tuples = append(tuples, fgax.GetTupleKey(fgax.TupleRequest{
		SubjectID:   subjectID,
		SubjectType: auth.ServiceSubjectType,
		ObjectID:    subjectID,
		ObjectType:  auth.ServiceSubjectType,
		Relation:    fgax.SelfRelation,
	}))

	if _, err := authz.WriteTupleKeys(ctx, tuples, nil); err != nil {
		logx.FromContext(ctx).Error().Err(err).Str("subject_id", subjectID).Msg("failed to create service subject tuples")

		return ErrInternalServerError
	}

	return nil
}

// syncScopeTuples writes and deletes the scope tuples of a service subject (an api token
// or an organization application install) when its scopes change from oldScopes to newScopes
func syncScopeTuples(ctx context.Context, authz fgax.Client, oldScopes, newScopes []string, orgID, subjectID string) error {
	scopeSet, err := fgamodel.DefaultServiceScopeSet()
	if err != nil {
		return fmt.Errorf("failed to load available token scopes from model: %w", err)
	}

	added, removed := diffScopes(oldScopes, newScopes)

	addTuples, err := scopeTuples(ctx, added, orgID, subjectID, scopeSet)
	if err != nil {
		return err
	}

	removeTuples, err := scopeTuples(ctx, removed, orgID, subjectID, scopeSet)
	if err != nil {
		return err
	}

	if len(addTuples) == 0 && len(removeTuples) == 0 {
		return nil
	}

	if _, err := authz.WriteTupleKeys(ctx, addTuples, removeTuples); err != nil {
		logx.FromContext(ctx).Error().Err(err).Str("subject_id", subjectID).Msg("failed to update scope tuples")

		return ErrInternalServerError
	}

	return nil
}

// / createScopeTuples creates the relationship tuples for the token
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
