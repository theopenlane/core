package hooks

import (
	"context"
	"database/sql"

	"entgo.io/ent"

	"github.com/theopenlane/gqlgen-plugins/graphutils"
	"github.com/theopenlane/iam/auth"
	"github.com/theopenlane/iam/totp"

	"github.com/theopenlane/core/pkg/logx"
	"github.com/theopenlane/ent/generated"
	"github.com/theopenlane/ent/generated/hook"
	"github.com/theopenlane/ent/generated/user"
	"github.com/theopenlane/ent/generated/usersetting"
)

// HookEnableTFA is a hook that generates the tfa secrets if the totp setting is set to allowed
func HookEnableTFA() ent.Hook {
	return hook.If(func(next ent.Mutator) ent.Mutator {
		return hook.TFASettingFunc(func(ctx context.Context, m *generated.TFASettingMutation) (generated.Value, error) {
			// if the user has TOTP enabled, generate a secret
			totpAllowed, _ := m.TotpAllowed()
			if !totpAllowed {
				return next.Mutate(ctx, m)
			}

			// check if the user has a TFA secret
			if m.Op() != ent.OpCreate {
				id, _ := m.ID() // get the ID of the TFA setting which will always be present on update

				existingSetting, err := m.Client().TFASetting.Get(ctx, id)
				if err != nil {
					return nil, err
				}

				if existingSetting.TfaSecret != nil {
					return next.Mutate(ctx, m)
				}
			}

			// generate the TFA secret
			u, err := constructTOTPUser(ctx, m)
			if err != nil {
				return nil, err
			}

			u.TFASecret, err = m.TOTP.Manager.TOTPSecret(u)
			if err != nil {
				logx.FromContext(ctx).Error().Err(err).Msg("unable to generate TOTP secret")

				return nil, err
			}

			// set the TFA secret
			m.SetTfaSecret(u.TFASecret)

			return next.Mutate(ctx, m)
		})
	},
		hook.And(
			hook.HasOp(ent.OpCreate|ent.OpUpdate|ent.OpUpdateOne),
			hook.HasFields("totp_allowed"),
		),
	)
}

// HookVerifyTFA is a hook that will generate recovery codes and enable TFA for a user
// if the TFA has been verified
func HookVerifyTFA() ent.Hook {
	return hook.On(func(next ent.Mutator) ent.Mutator {
		return hook.TFASettingFunc(func(ctx context.Context, m *generated.TFASettingMutation) (generated.Value, error) {
			// once verified, create recovery codes
			verified, ok := m.Verified()

			// if recovery codes are cleared, generate new ones
			regenBackupCodes := false

			requestInput := graphutils.GetMapInputVariableByName(ctx, "input")
			if requestInput != nil {
				i := *requestInput
				regenBackupCodes, _ = i["regenBackupCodes"].(bool)
			}

			if (ok && verified) || regenBackupCodes {
				codes := m.TOTP.Manager.GenerateRecoveryCodes()
				m.SetRecoveryCodes(codes)

				if verified {
					if err := setUserTFASetting(ctx, m, true); err != nil {
						return nil, err
					}
				}
			}

			totpAllowed, ok := m.TotpAllowed()
			if ok && !totpAllowed {
				// if TOTP is not allowed, clear the TFA settings
				m.SetVerified(false)
				m.ClearRecoveryCodes()
				m.ClearTfaSecret()

				// disable TFA on the user settings
				if err := setUserTFASetting(ctx, m, false); err != nil {
					return nil, err
				}
			}

			return next.Mutate(ctx, m)
		})
	}, ent.OpUpdate|ent.OpUpdateOne)
}

// enableTFA is a function that enables TFA for a user on their user settings
// once the TFA has been verified
func setUserTFASetting(ctx context.Context, m *generated.TFASettingMutation, enabled bool) error {
	userID, ok := m.OwnerID()
	if !ok {
		var err error

		userID, err = auth.GetSubjectIDFromContext(ctx)
		if err != nil {
			return err
		}
	}

	// update user settings
	if err := m.Client().UserSetting.Update().
		Where(usersetting.UserID(userID)).
		SetIsTfaEnabled(enabled). // set tfa enabled
		Exec(ctx); err != nil {
		return err
	}

	return nil
}

// constructTOTPUser constructs a TOTP user object from the mutation
func constructTOTPUser(ctx context.Context, m *generated.TFASettingMutation) (*totp.User, error) {
	userID, ok := m.OwnerID()
	if !ok {
		var err error

		userID, err = auth.GetSubjectIDFromContext(ctx)
		if err != nil {
			return nil, err
		}
	}

	u := &totp.User{
		ID: userID,
	}

	// get the user object
	userDetails, err := m.Client().User.Query().
		WithSetting().
		Where(user.ID(userID)).
		Only(ctx)
	if err != nil {
		return nil, err
	}

	// set the TFA settings
	u.IsEmailOTPAllowed, _ = m.EmailOtpAllowed()
	u.IsPhoneOTPAllowed, _ = m.PhoneOtpAllowed()
	u.IsTOTPAllowed, _ = m.TotpAllowed()

	// setup account name fields
	isValid := userDetails.Email != ""

	u.Email = sql.NullString{
		String: userDetails.Email,
		Valid:  isValid,
	}

	phoneNumber := userDetails.Edges.Setting.PhoneNumber

	isValid = true
	if phoneNumber == nil {
		isValid = false
	}

	if phoneNumber != nil {
		u.Phone = sql.NullString{
			String: *phoneNumber,
			Valid:  isValid,
		}
	}

	return u, nil
}
