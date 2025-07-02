package handlers

import (
	"errors"
	"net/http"

	"github.com/getkin/kin-openapi/openapi3"
	echo "github.com/theopenlane/echox"

	"github.com/rs/zerolog/log"

	"github.com/theopenlane/iam/sessions"
	"github.com/theopenlane/utils/ulids"
	"golang.org/x/oauth2"

	"github.com/theopenlane/core/internal/ent/generated/privacy"
	"github.com/theopenlane/core/pkg/models"
)

// SSOTokenAuthorizeHandler marks a token as authorized for SSO for an organization.
func (h *Handler) SSOTokenAuthorizeHandler(ctx echo.Context) error {
	var in models.SSOTokenAuthorizeRequest
	if err := ctx.Bind(&in); err != nil {
		return h.InvalidInput(ctx, err)
	}
	if err := in.Validate(); err != nil {
		return h.InvalidInput(ctx, err)
	}

	reqCtx := privacy.DecisionContext(ctx.Request().Context(), privacy.Allow)

	switch in.TokenType {
	case "api":
		if _, err := h.DBClient.APIToken.Get(reqCtx, in.TokenID); err != nil {
			log.Error().Err(err).Msg("unable to find api token for SSO")
			return h.BadRequest(ctx, err)
		}
	case "personal":
		if _, err := h.DBClient.PersonalAccessToken.Get(reqCtx, in.TokenID); err != nil {
			log.Error().Err(err).Msg("unable to find personal access token")
			return h.BadRequest(ctx, err)
		}
	default:
		return h.BadRequest(ctx, errInvalidTokenType)
	}

	rpCfg, err := h.oidcConfig(reqCtx, in.OrganizationID)
	if err != nil {
		return h.BadRequest(ctx, err)
	}

	state := ulids.New().String()
	nonce := ulids.New().String()

	sessions.SetCookie(ctx.Response().Writer, in.TokenID, "token_id", sessions.CookieConfig{Path: "/", HTTPOnly: true, SameSite: http.SameSiteLaxMode})
	sessions.SetCookie(ctx.Response().Writer, in.TokenType, "token_type", sessions.CookieConfig{Path: "/", HTTPOnly: true, SameSite: http.SameSiteLaxMode})
	sessions.SetCookie(ctx.Response().Writer, in.OrganizationID, "organization_id", sessions.CookieConfig{Path: "/", HTTPOnly: true, SameSite: http.SameSiteLaxMode})
	sessions.SetCookie(ctx.Response().Writer, state, "state", sessions.CookieConfig{Path: "/", HTTPOnly: true, SameSite: http.SameSiteLaxMode})
	sessions.SetCookie(ctx.Response().Writer, nonce, "nonce", sessions.CookieConfig{Path: "/", HTTPOnly: true, SameSite: http.SameSiteLaxMode})

	authURL := rpCfg.OAuthConfig().AuthCodeURL(state, oauth2.SetAuthURLParam("nonce", nonce))

	return ctx.Redirect(http.StatusFound, authURL)
}

// BindSSOTokenAuthorizeHandler binds the SSO token authorization endpoint to the OpenAPI schema.
func (h *Handler) BindSSOTokenAuthorizeHandler() *openapi3.Operation {
	op := openapi3.NewOperation()
	op.Description = "Authorize a token for SSO with an organization"
	op.OperationID = "SSOTokenAuthorize"
	op.Tags = []string{"authentication"}
	op.Security = &openapi3.SecurityRequirements{}

	h.AddQueryParameter("organization_id", op)
	h.AddQueryParameter("token_id", op)
	h.AddQueryParameter("token_type", op)
	op.AddResponse(http.StatusFound, openapi3.NewResponse().WithDescription("Redirect"))
	op.AddResponse(http.StatusBadRequest, badRequest())
	op.AddResponse(http.StatusInternalServerError, internalServerError())

	return op
}

var errInvalidTokenType = errors.New("invalid token type")
