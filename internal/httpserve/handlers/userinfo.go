package handlers

import (
	echo "github.com/theopenlane/echox"
	"github.com/theopenlane/shared/logx"

	"github.com/theopenlane/iam/auth"
)

// UserInfo returns the user information for the authenticated user
func (h *Handler) UserInfo(ctx echo.Context, openapi *OpenAPIContext) error {
	if isRegistrationContext(ctx) {
		return nil
	}

	// setup view context
	reqCtx := ctx.Request().Context()

	userID, err := auth.GetSubjectIDFromContext(reqCtx)
	if err != nil {
		logx.FromContext(reqCtx).Err(err).Msg("unable to get user id from context")

		return h.BadRequest(ctx, err, openapi)
	}

	// get user from database by subject
	user, err := h.getUserDetailsByID(reqCtx, userID)
	if err != nil {
		logx.FromContext(reqCtx).Error().Err(err).Msg("unable to get user by subject")

		return h.BadRequest(ctx, err, openapi)
	}

	return h.Success(ctx, user)
}
