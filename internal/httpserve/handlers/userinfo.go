package handlers

import (
	echo "github.com/theopenlane/echox"
	"github.com/theopenlane/httpsling"

	"github.com/theopenlane/core/pkg/logx"

	"github.com/theopenlane/iam/auth"
)

// UserInfo returns the user information for the authenticated user
func (h *Handler) UserInfo(ctx echo.Context) error {
	ctx.Response().Header().Set(httpsling.HeaderContentType, httpsling.ContentTypeJSONUTF8)

	// setup view context
	reqCtx := ctx.Request().Context()

	caller, ok := auth.CallerFromContext(reqCtx)
	if !ok || caller == nil || caller.SubjectID == "" {
		logx.FromContext(reqCtx).Error().Msg("unable to get user id from context")

		return h.BadRequest(ctx, auth.ErrNoAuthUser)
	}

	userID := caller.SubjectID

	// get user from database by subject
	user, err := h.getUserDetailsByID(reqCtx, userID)
	if err != nil {
		logx.FromContext(reqCtx).Error().Err(err).Msg("unable to get user by subject")

		return h.BadRequest(ctx, err)
	}

	return h.Success(ctx, user)
}
