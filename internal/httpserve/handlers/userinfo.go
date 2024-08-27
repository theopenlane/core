package handlers

import (
	echo "github.com/theopenlane/echox"

	"github.com/theopenlane/core/pkg/auth"
)

// UserInfo returns the user information for the authenticated user
func (h *Handler) UserInfo(ctx echo.Context) error {
	// setup view context
	reqCtx := ctx.Request().Context()

	userID, err := auth.GetUserIDFromContext(reqCtx)
	if err != nil {
		h.Logger.Errorw("unable to get user id from context", "error", err)

		return h.BadRequest(ctx, err)
	}

	// get user from database by subject
	user, err := h.getUserDetailsByID(reqCtx, userID)
	if err != nil {
		h.Logger.Errorw("unable to get user by subject", "error", err)

		return h.BadRequest(ctx, err)
	}

	return h.Success(ctx, user)
}
