package hooks

import (
	"context"
	"fmt"

	"entgo.io/ent"

	"github.com/theopenlane/iam/auth"

	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/generated/hook"
	"github.com/theopenlane/utils/keygen"

	"github.com/theopenlane/utils/passwd"
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

			// generate raw token
			secret := keygen.Secret()
			publicID := keygen.AlphaNumeric(16)
			prefix := "tolp"

			rawToken := fmt.Sprintf("%s_%s_%s", prefix, publicID, secret)

			// hash the secret for storage
			hash, err := passwd.CreateDerivedKey(secret)
			if err != nil {
				return nil, err
			}

			// set the hashed token for storage (backwards compatibility)
			m.SetToken(hash)

			// set the new token_hash (argon2 of secret)
			m.SetTokenHash(hash)

			// set the token fingerprint for lookup
			tokenFP := keygen.GenerateSHA256Hmac("", []byte(publicID))
			m.SetTokenFingerprint(tokenFP)

			// set user on the token
			m.SetOwnerID(userID)

			retVal, err := next.Mutate(ctx, m)
			if err != nil {
				return nil, err
			}

			// set the token field on the returned object to the raw token so the caller can see it
			pat, ok := retVal.(*generated.PersonalAccessToken)
			if !ok {
				return retVal, nil
			}

			pat.Token = rawToken

			return pat, nil
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
