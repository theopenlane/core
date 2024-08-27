package hooks

import (
	"context"

	"entgo.io/ent"

	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/generated/hook"
	"github.com/theopenlane/iam/auth"
)

const (
	redacted = "*****************************"
)

// HookCreatePersonalAccessToken runs on access token mutations and sets the owner id
func HookCreatePersonalAccessToken() ent.Hook {
	return hook.On(func(next ent.Mutator) ent.Mutator {
		return hook.PersonalAccessTokenFunc(func(ctx context.Context, mutation *generated.PersonalAccessTokenMutation) (generated.Value, error) {
			userID, err := auth.GetUserIDFromContext(ctx)
			if err != nil {
				return nil, err
			}

			// set user on the token
			mutation.SetOwnerID(userID)

			return next.Mutate(ctx, mutation)
		})
	}, ent.OpCreate)
}

// HookUpdatePersonalAccessToken runs on access token update and redacts the token
func HookUpdatePersonalAccessToken() ent.Hook {
	return hook.On(func(next ent.Mutator) ent.Mutator {
		return hook.PersonalAccessTokenFunc(func(ctx context.Context, mutation *generated.PersonalAccessTokenMutation) (generated.Value, error) {
			// do not allow user to be changed
			_, ok := mutation.OwnerID()
			if ok {
				mutation.ClearOwner()
			}

			retVal, err := next.Mutate(ctx, mutation)
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
