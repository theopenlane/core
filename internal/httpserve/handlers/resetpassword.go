package handlers

import (
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/oklog/ulid/v2"
	echo "github.com/theopenlane/echox"

	"github.com/theopenlane/utils/rout"

	"github.com/theopenlane/iam/tokens"

	"github.com/theopenlane/utils/passwd"

	models "github.com/theopenlane/core/common/openapi"
	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/privacy/token"
	"github.com/theopenlane/core/internal/integrations/definitions/email"
	"github.com/theopenlane/core/pkg/logx"
)

// ResetPassword allows the user (after requesting a password reset) to
// set a new password - the password reset token needs to be set in the request
// and not expired. If the request is successful, a confirmation of the reset is sent
// to the user and a 204 no content is returned
func (h *Handler) ResetPassword(ctx echo.Context) error {
	req, err := BindAndValidate[models.ResetPasswordRequest](ctx)
	if err != nil {
		return h.InvalidInput(ctx, err)
	}

	reqCtx := ctx.Request().Context()

	// setup viewer context
	ctxWithToken := token.NewContextWithResetToken(reqCtx, req.Token)

	// lookup user from db based on provided token
	entUser, err := h.getUserByResetToken(ctxWithToken, req.Token)
	if err != nil {
		logx.FromContext(reqCtx).Error().Err(err).Msg("error retrieving user token")

		if generated.IsNotFound(err) {
			return h.BadRequest(ctx, ErrPassWordResetTokenInvalid)
		}

		return h.InternalServerError(ctx, ErrUnableToVerifyEmail)
	}

	// ent user to &User for funcs
	user := &User{
		ID:    entUser.ID,
		Email: entUser.Email,
	}

	// set tokens for request
	if err := user.setResetTokens(entUser, req.Token); err != nil {
		logx.FromContext(reqCtx).Error().Err(err).Msg("unable to set reset tokens for request")

		return h.BadRequest(ctx, err)
	}

	// Construct the user token from the database fields
	// type ulid to string
	uid, err := ulid.Parse(entUser.ID)
	if err != nil {
		return err
	}

	// construct token from db fields
	token := &tokens.ResetToken{
		UserID: uid,
	}

	if token.ExpiresAt, err = user.GetPasswordResetExpires(); err != nil {
		logx.FromContext(reqCtx).Error().Err(err).Msg("unable to parse expiration")

		return h.InternalServerError(ctx, ErrUnableToVerifyEmail)
	}

	// Verify the token is valid with the stored secret
	if err = token.Verify(user.GetPasswordResetToken(), user.PasswordResetSecret); err != nil {
		if errors.Is(err, tokens.ErrTokenExpired) {
			errMsg := "reset token is expired, please request a new token using forgot-password"

			return h.BadRequest(ctx, fmt.Errorf("%w: %s", ErrPassWordResetTokenInvalid, errMsg))
		}

		return h.BadRequest(ctx, err)
	}

	// make sure its not the same password as current
	// a user that previously authenticated with oauth and resets their password
	// won't have a password originally so this will be nil
	if entUser.Password != nil {
		valid, err := passwd.VerifyDerivedKey(*entUser.Password, req.Password)
		if err != nil || valid {
			return h.BadRequest(ctx, ErrNonUniquePassword)
		}
	}

	// set context for remaining request based on logged in user
	userCtx := setAuthenticatedContext(ctxWithToken, entUser)

	if err := h.updateUserPassword(userCtx, entUser.ID, req.Password); err != nil {
		logx.FromContext(reqCtx).Error().Err(err).Msg("error updating user password")

		return h.BadRequest(ctx, err)
	}

	if err := h.expireAllResetTokensUserByEmail(userCtx, user.Email); err != nil {
		logx.FromContext(reqCtx).Error().Err(err).Msg("error expiring existing tokens")

		return h.BadRequest(ctx, err)
	}

	if err := h.sendEmail(userCtx, email.ResetSuccessOp.Name(), email.PasswordResetSuccessRequest{
		RecipientInfo: email.RecipientInfo{
			Email:     entUser.Email,
			FirstName: entUser.FirstName,
			LastName:  entUser.LastName,
		},
	}); err != nil {
		logx.FromContext(reqCtx).Error().Err(err).Msg("error sending password reset success email")

		return h.InternalServerError(ctx, ErrProcessingRequest)
	}

	out := &models.ResetPasswordResponse{
		Reply:   rout.Reply{Success: true},
		Message: "password has been re-set successfully",
	}

	return h.Success(ctx, out)
}

// setResetTokens sets the fields for the password reset
func (u *User) setResetTokens(user *generated.User, reqToken string) error {
	tokens := user.Edges.PasswordResetTokens
	for _, t := range tokens {
		if t.Token == reqToken {
			u.PasswordResetToken = sql.NullString{String: t.Token, Valid: true}
			u.PasswordResetSecret = *t.Secret
			u.PasswordResetExpires = sql.NullString{String: t.TTL.Format(time.RFC3339Nano), Valid: true}

			return nil
		}
	}

	// This should only happen on a race condition with two request
	// otherwise, since we get the user by the token, it should always
	// be there
	return ErrPassWordResetTokenInvalid
}
