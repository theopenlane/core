package handlers

import (
	"database/sql"
	"slices"

	models "github.com/theopenlane/core/common/openapi"
	"github.com/theopenlane/core/pkg/logx"
	echo "github.com/theopenlane/echox"
	"github.com/theopenlane/iam/auth"
	"github.com/theopenlane/iam/totp"
	"github.com/theopenlane/utils/rout"
)

// ValidateTOTP validates a user's TOTP code
// this currently only supports TOTP and not OTP codes via email and SMS
func (h *Handler) ValidateTOTP(ctx echo.Context, openapi *OpenAPIContext) error {
	in, err := BindAndValidateWithAutoRegistry(ctx, h, openapi.Operation, models.ExampleTFASuccessRequest, models.ExampleTFASSuccessResponse, openapi.Registry)
	if err != nil {
		return h.InvalidInput(ctx, err, openapi)
	}

	// Skip actual handler logic during OpenAPI registration
	if isRegistrationContext(ctx) {
		return nil
	}

	reqCtx := ctx.Request().Context()

	userID, err := auth.GetSubjectIDFromContext(reqCtx)
	if err != nil {
		logx.FromContext(reqCtx).Err(err).Msg("unable to get user id from context")

		return h.BadRequest(ctx, err, openapi)
	}

	// get user from database by subject
	user, err := h.getUserTFASettings(reqCtx, userID)
	if err != nil {
		logx.FromContext(reqCtx).Error().Err(err).Msg("unable to get user")

		return h.BadRequest(ctx, err, openapi)
	}

	if user.Edges.TfaSettings == nil || len(user.Edges.TfaSettings) != 1 || user.Edges.TfaSettings[0].TfaSecret == nil {
		logx.FromContext(reqCtx).Info().Msg("tfa validation request but user has no TFA settings")

		return h.InvalidInput(ctx, ErrInvalidInput, openapi)
	}

	tfasetting := user.Edges.TfaSettings[0]

	if in.TOTPCode == "" {
		// validate recovery code instead
		recoveryCodeIndex := slices.Index(tfasetting.RecoveryCodes, in.RecoveryCode)
		if recoveryCodeIndex > -1 {
			// remove the recovery code from the list
			tfasetting.RecoveryCodes = append(tfasetting.RecoveryCodes[:recoveryCodeIndex], tfasetting.RecoveryCodes[recoveryCodeIndex+1:]...)

			if err := h.updateRecoveryCodes(reqCtx, tfasetting.ID, tfasetting.RecoveryCodes); err != nil {
				logx.FromContext(reqCtx).Error().Err(err).Msg("unable to update recovery codes")

				return h.BadRequest(ctx, err, openapi)
			}

			return h.Success(ctx, models.TFAReply{
				Reply: rout.Reply{Success: true},
			}, openapi)
		}

		return h.BadRequest(ctx, ErrInvalidRecoveryCode, openapi)
	}

	totpUser := totp.User{
		ID:            userID,
		TFASecret:     *tfasetting.TfaSecret,
		IsTOTPAllowed: tfasetting.TotpAllowed,
		Email:         sql.NullString{String: user.Email, Valid: true},
	}

	if err := h.OTPManager.Manager.ValidateTOTP(reqCtx, &totpUser, in.TOTPCode); err != nil {
		logx.FromContext(reqCtx).Error().Err(err).Msg("unable to validate TOTP code")

		return h.BadRequest(ctx, err, openapi)
	}

	return h.Success(ctx, models.TFAReply{
		Reply: rout.Reply{Success: true},
	}, openapi)
}
