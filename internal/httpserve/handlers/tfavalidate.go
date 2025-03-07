package handlers

import (
	"database/sql"
	"net/http"
	"slices"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/rs/zerolog/log"
	echo "github.com/theopenlane/echox"
	"github.com/theopenlane/iam/auth"
	"github.com/theopenlane/iam/totp"
	"github.com/theopenlane/utils/rout"

	"github.com/theopenlane/core/pkg/models"
)

// ValidateTOTP validates a user's TOTP code
// this currently only supports TOTP and not OTP codes via email and SMS
func (h *Handler) ValidateTOTP(ctx echo.Context) error {
	var in models.TFARequest
	if err := ctx.Bind(&in); err != nil {
		return h.InvalidInput(ctx, err)
	}

	if err := in.Validate(); err != nil {
		return h.InvalidInput(ctx, err)
	}

	reqCtx := ctx.Request().Context()

	userID, err := auth.GetSubjectIDFromContext(reqCtx)
	if err != nil {
		log.Err(err).Msg("unable to get user id from context")

		return h.BadRequest(ctx, err)
	}

	// get user from database by subject
	user, err := h.getUserTFASettings(reqCtx, userID)
	if err != nil {
		log.Error().Err(err).Msg("unable to get user")

		return h.BadRequest(ctx, err)
	}

	if user.Edges.TfaSettings == nil || len(user.Edges.TfaSettings) != 1 || user.Edges.TfaSettings[0].TfaSecret == nil {
		log.Warn().Msg("tfa validation request but user has no TFA settings")

		return h.InvalidInput(ctx, err)
	}

	tfasetting := user.Edges.TfaSettings[0]

	if in.TOTPCode == "" {
		// validate recovery code instead
		recoveryCodeIndex := slices.Index(tfasetting.RecoveryCodes, in.RecoveryCode)
		if recoveryCodeIndex > -1 {
			// remove the recovery code from the list
			tfasetting.RecoveryCodes = append(tfasetting.RecoveryCodes[:recoveryCodeIndex], tfasetting.RecoveryCodes[recoveryCodeIndex+1:]...)

			if err := h.updateRecoveryCodes(reqCtx, tfasetting.ID, tfasetting.RecoveryCodes); err != nil {
				log.Error().Err(err).Msg("unable to update recovery codes")

				return h.BadRequest(ctx, err)
			}

			return h.Success(ctx, models.TFAReply{
				Reply: rout.Reply{Success: true},
			})
		}

		return h.BadRequest(ctx, ErrInvalidRecoveryCode)
	}

	totpUser := totp.User{
		ID:            userID,
		TFASecret:     *tfasetting.TfaSecret,
		IsTOTPAllowed: tfasetting.TotpAllowed,
		Email:         sql.NullString{String: user.Email, Valid: true},
	}

	if err := h.OTPManager.TOTPManager.ValidateTOTP(reqCtx, &totpUser, in.TOTPCode); err != nil {
		log.Error().Err(err).Msg("unable to validate TOTP code")

		return h.BadRequest(ctx, err)
	}

	return h.Success(ctx, models.TFAReply{
		Reply: rout.Reply{Success: true},
	})
}

// BindTFAHandler binds the tfavalidate request to the OpenAPI schema
func (h *Handler) BindTFAHandler() *openapi3.Operation {
	tfavalidate := openapi3.NewOperation()
	tfavalidate.Description = "Validate a user's TOTP code"
	tfavalidate.Tags = []string{"tfa"}
	tfavalidate.OperationID = "TFAValidation"
	tfavalidate.Security = AllSecurityRequirements()

	h.AddRequestBody("TFARequest", models.ExampleTFASuccessRequest, tfavalidate)
	h.AddResponse("TFAReply", "success", models.ExampleTFASSuccessResponse, tfavalidate, http.StatusOK)
	tfavalidate.AddResponse(http.StatusBadRequest, badRequest())
	tfavalidate.AddResponse(http.StatusBadRequest, invalidInput())

	return tfavalidate
}
