package handlers

import (
	echo "github.com/theopenlane/echox"

	"github.com/theopenlane/utils/rout"

	"github.com/theopenlane/iam/auth"

	ent "github.com/theopenlane/ent/generated"
	"github.com/theopenlane/ent/privacy/rule"
	"github.com/theopenlane/shared/logx"
	models "github.com/theopenlane/shared/openapi"
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

	reqCtx := ctx.Request().Context()

	// verify the refresh token
	claims, err := h.TokenManager.Verify(req.RefreshToken)
	if err != nil {
		logx.FromContext(reqCtx).Error().Err(err).Msg("error verifying token")

		return h.BadRequest(ctx, ErrUnableToVerifyToken, openapi)
	}

	// check user in the database, sub == claims subject and ensure only one record is returned
	user, err := h.getUserDetailsByID(reqCtx, claims.Subject)
	if err != nil {
		if ent.IsNotFound(err) {
			logx.Ctx(reqCtx).Info().Str("userID", user.ID).Msg("user not found during token refresh")
			return h.NotFound(ctx, ErrProcessingRequest, openapi)
		}

		return h.InternalServerError(ctx, ErrProcessingRequest, openapi)
	}

	// ensure the user is still active
	if user.Edges.Setting.Status != "ACTIVE" {
		logx.Ctx(reqCtx).Info().Str("userID", user.ID).Msg("user not active during token refresh")

		return h.NotFound(ctx, ErrProcessingRequest, openapi)
	}

	// get modules on refresh
	modules, err := rule.GetFeaturesForSpecificOrganization(reqCtx, claims.OrgID)
	if err != nil {
		logx.FromContext(reqCtx).Error().Err(err).Msg("error obtaining org features for claims, skipping modules in JWT")
	}

	claims.Modules = modules

	// UserID is not on the refresh token, so we need to set it now
	claims.UserID = user.ID

	accessToken, refreshToken, err := h.TokenManager.CreateTokenPair(claims)
	if err != nil {
		logx.FromContext(reqCtx).Error().Err(err).Msg("error creating token pair")

		return h.InternalServerError(ctx, ErrProcessingRequest, openapi)
	}

	// set cookies on request with the access and refresh token
	auth.SetAuthCookies(ctx.Response().Writer, accessToken, refreshToken, *h.SessionConfig.CookieConfig)

	// set sessions in response
	if _, err = h.SessionConfig.CreateAndStoreSession(reqCtx, ctx.Response().Writer, user.ID); err != nil {
		logx.FromContext(reqCtx).Error().Err(err).Msg("error storing session")

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
