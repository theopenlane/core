package handlers

import (
	"github.com/rs/zerolog/log"
	echo "github.com/theopenlane/echox"

	"github.com/theopenlane/iam/auth"
)

// UserInfo returns the user information for the authenticated user
func (h *Handler) UserInfo(ctx echo.Context) error {
	// setup view context
	reqCtx := ctx.Request().Context()

	userID, err := auth.GetSubjectIDFromContext(reqCtx)
	if err != nil {
		log.Err(err).Msg("unable to get user id from context")

		return h.BadRequest(ctx, err)
	}

	// get user from database by subject
	user, err := h.getUserDetailsByID(reqCtx, userID)
	if err != nil {
		log.Error().Err(err).Msg("unable to get user by subject")

		return h.BadRequest(ctx, err)
	}

	return h.Success(ctx, user)
}
