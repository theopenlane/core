package hooks

import (
	"context"
	"encoding/base64"

	"entgo.io/ent"
	"github.com/theopenlane/iam/auth"
	"github.com/theopenlane/iam/tokens"

	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/generated/hook"
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

			// generate the token
			publicID, secretBytes, err := tokens.GenerateAPITokenKeyMaterial()
			if err != nil {
				return nil, err
			}

			// construct the token string
			// default prefix is "tolp_"
			// default delimiter is "_"
			// secret is base64 encoded
			secret := base64.RawStdEncoding.EncodeToString(secretBytes)
			tokenStr := "tolp_" + publicID + "_" + secret

			m.SetToken(tokenStr)
			m.SetTokenPublicID(publicID)
			m.SetTokenSecret(secret)

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
