package hooks

import (
	"context"

	"entgo.io/ent"

	"github.com/theopenlane/iam/auth"

	"github.com/theopenlane/ent/generated"
	"github.com/theopenlane/ent/generated/hook"
)

const (
	redacted = "*****************************"
)

// HookCreatePersonalAccessToken runs on access token mutations and sets the owner id
func HookCreatePersonalAccessToken() ent.Hook {
	return hook.On(func(next ent.Mutator) ent.Mutator {
		return hook.PersonalAccessTokenFunc(func(ctx context.Context, m *generated.PersonalAccessTokenMutation) (generated.Value, error) {
			userID, err := auth.GetSubjectIDFromContext(ctx)
			if err != nil {
				return nil, err
			}

			if err := validateExpirationTime(m); err != nil {
				return nil, err
			}

			// set user on the token
			m.SetOwnerID(userID)

			return next.Mutate(ctx, m)
		})
	}, ent.OpCreate)
}

// HookUpdatePersonalAccessToken runs on access token update and redacts the token
func HookUpdatePersonalAccessToken() ent.Hook {
	return hook.On(func(next ent.Mutator) ent.Mutator {
		return hook.PersonalAccessTokenFunc(func(ctx context.Context, m *generated.PersonalAccessTokenMutation) (generated.Value, error) {
			// do not allow user to be changed
			_, ok := m.OwnerID()
			if ok {
				m.ClearOwner()
			}

			retVal, err := next.Mutate(ctx, m)
			if err != nil {
				return nil, err
			}

			// redact the token
			pat, ok := retVal.(*generated.PersonalAccessToken)
			if !ok {
				return retVal, nil
			}

			pat.Token = redacted

			return pat, nil
		})
	}, ent.OpUpdate|ent.OpUpdateOne)
}
