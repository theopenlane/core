package hooks

import (
	"context"
	"database/sql"

	"entgo.io/ent"
	"github.com/99designs/gqlgen/graphql"

	"github.com/theopenlane/iam/totp"

	"github.com/theopenlane/iam/auth"

	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/generated/hook"
	"github.com/theopenlane/core/internal/ent/generated/usersetting"
)

func HookEnableTFA() ent.Hook {
	return hook.On(func(next ent.Mutator) ent.Mutator {
		return hook.TFASettingFunc(func(ctx context.Context, mutation *generated.TFASettingMutation) (generated.Value, error) {
			// once verified, create recovery codes
			verified, ok := mutation.Verified()

			// if recovery codes are cleared, generate new ones
			gtx := graphql.GetOperationContext(ctx)
			regenBackupCodes, _ := gtx.Variables["input"].(map[string]interface{})["regenBackupCodes"].(bool)

			if (ok && verified) || regenBackupCodes {
				u, err := constructTOTPUser(ctx, mutation)
				if err != nil {
					return nil, err
				}

				u.TFASecret, err = mutation.TOTP.TOTPManager.TOTPSecret(u)
				if err != nil {
					return nil, err
				}

				codes := mutation.TOTP.TOTPManager.GenerateRecoveryCodes()
				mutation.SetRecoveryCodes(codes)

				if verified {
					// update user settings
					_, err := mutation.Client().UserSetting.Update().
						Where(usersetting.UserID(u.ID)).
						SetIsTfaEnabled(true). // set tfa enabled to true
						Save(ctx)
					if err != nil {
						return nil, err
					}
				}
			}

			return next.Mutate(ctx, mutation)
		})
	}, ent.OpUpdate|ent.OpUpdateOne)
}

// constructTOTPUser constructs a TOTP user object from the mutation
func constructTOTPUser(ctx context.Context, mutation *generated.TFASettingMutation) (*totp.User, error) {
	userID, ok := mutation.OwnerID()
	if !ok {
		var err error

		userID, err = auth.GetUserIDFromContext(ctx)
		if err != nil {
			return nil, err
		}
	}

	u := &totp.User{
		ID: userID,
	}

	// get the user object
	user, err := mutation.Client().User.Get(ctx, userID)
	if err != nil {
		return nil, err
	}

	// get the full setting object
	setting, err := user.Setting(ctx)
	if err != nil {
		return nil, err
	}

	// set the TFA settings
	u.IsEmailOTPAllowed, _ = mutation.EmailOtpAllowed()
	u.IsPhoneOTPAllowed, _ = mutation.PhoneOtpAllowed()
	u.IsTOTPAllowed, _ = mutation.TotpAllowed()

	// setup account name fields
	u.Email = sql.NullString{
		String: user.Email,
	}

	phoneNumber := setting.PhoneNumber
	if phoneNumber != nil {
		u.Phone = sql.NullString{
			String: *setting.PhoneNumber,
		}
	}

	return u, nil
}
