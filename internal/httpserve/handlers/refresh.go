package handlers

import (
	"github.com/rs/zerolog/log"
	echo "github.com/theopenlane/echox"

	"github.com/theopenlane/utils/rout"

	"github.com/theopenlane/iam/auth"

	ent "github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/pkg/models"
)

// RefreshHandler allows users to refresh their access token using their refresh token
func (h *Handler) RefreshHandler(ctx echo.Context, openapi *OpenAPIContext) error {
	req, err := BindAndValidateWithAutoRegistry(ctx, h, openapi.Operation, models.ExampleRefreshRequest, models.ExampleRefreshSuccessResponse, openapi.Registry)
	if err != nil {
		return h.InvalidInput(ctx, err, openapi)
	}

	// Skip actual handler logic during OpenAPI registration
	if isRegistrationContext(ctx) {
		return nil
	}

	// verify the refresh token
	claims, err := h.TokenManager.Verify(req.RefreshToken)
	if err != nil {
		log.Error().Err(err).Msg("error verifying token")

		return h.BadRequest(ctx, err, openapi)
	}

	// check user in the database, sub == claims subject and ensure only one record is returned
	user, err := h.getUserDetailsByID(ctx.Request().Context(), claims.Subject)
	if err != nil {
		if ent.IsNotFound(err) {
			return h.NotFound(ctx, ErrNoAuthUser, openapi)
		}

		return h.InternalServerError(ctx, ErrProcessingRequest, openapi)
	}

	// ensure the user is still active
	if user.Edges.Setting.Status != "ACTIVE" {
		return h.NotFound(ctx, ErrNoAuthUser, openapi)
	}

	// UserID is not on the refresh token, so we need to set it now
	claims.UserID = user.ID

	accessToken, refreshToken, err := h.TokenManager.CreateTokenPair(claims)
	if err != nil {
		log.Error().Err(err).Msg("error creating token pair")

		return h.InternalServerError(ctx, ErrProcessingRequest, openapi)
	}

	// set cookies on request with the access and refresh token
	auth.SetAuthCookies(ctx.Response().Writer, accessToken, refreshToken, *h.SessionConfig.CookieConfig)

	// set sessions in response
	if _, err = h.SessionConfig.CreateAndStoreSession(ctx.Request().Context(), ctx.Response().Writer, user.ID); err != nil {
		log.Error().Err(err).Msg("error storing session")

		return err
	}

	out := &models.RefreshReply{
		Reply:   rout.Reply{Success: true},
		Message: "success",
		AuthData: models.AuthData{
			AccessToken:  accessToken,
			RefreshToken: refreshToken,
		},
	}

	return h.Success(ctx, out, openapi)
}
