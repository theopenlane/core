package handlers

import (
	"errors"
	"net/http"
	"time"

	echo "github.com/theopenlane/echox"
	"github.com/zitadel/oidc/v3/pkg/client/rp"
	"github.com/zitadel/oidc/v3/pkg/oidc"

	"github.com/rs/zerolog/log"

	"github.com/theopenlane/iam/sessions"
	"github.com/theopenlane/utils/contextx"
	"github.com/theopenlane/utils/rout"
	"github.com/theopenlane/utils/ulids"
	"golang.org/x/oauth2"

	"github.com/theopenlane/core/internal/ent/generated/privacy"
	"github.com/theopenlane/core/pkg/models"
)

// SSOTokenAuthorizeHandler marks a token as authorized for SSO for an organization
func (h *Handler) SSOTokenAuthorizeHandler(ctx echo.Context, openapi *OpenAPIContext) error {
	in, err := BindAndValidateQueryParams(ctx, openapi.Operation, models.ExampleSSOTokenAuthorizeRequest, openapi.Registry)
	if err != nil {
		return h.InvalidInput(ctx, err, openapi)
	}

	reqCtx := privacy.DecisionContext(ctx.Request().Context(), privacy.Allow)

	switch in.TokenType {
	case "api":
		if _, err := h.DBClient.APIToken.Get(reqCtx, in.TokenID); err != nil {
			log.Error().Err(err).Msg("unable to find api token for SSO")

			return h.BadRequest(ctx, err, openapi)
		}
	case "personal":
		if _, err := h.DBClient.PersonalAccessToken.Get(reqCtx, in.TokenID); err != nil {
			log.Error().Err(err).Msg("unable to find personal access token")

			return h.BadRequest(ctx, err, openapi)
		}
	default:
		return h.BadRequest(ctx, errInvalidTokenType, openapi)
	}

	rpCfg, err := h.oidcConfig(reqCtx, in.OrganizationID)
	if err != nil {
		return h.BadRequest(ctx, err, openapi)
	}

	state := ulids.New().String()
	nonce := ulids.New().String()
	writer := ctx.Response().Writer
	cc := sessions.CookieConfig{
		Path:     "/",
		HTTPOnly: true,
		SameSite: http.SameSiteLaxMode,
		Secure:   true,
	}

	sessions.SetCookie(writer, in.TokenID, "token_id", cc)
	sessions.SetCookie(writer, in.TokenType, "token_type", cc)
	sessions.SetCookie(writer, in.OrganizationID, "organization_id", cc)
	sessions.SetCookie(writer, state, "state", cc)
	sessions.SetCookie(writer, nonce, "nonce", cc)

	authURL := rpCfg.OAuthConfig().AuthCodeURL(state, oauth2.SetAuthURLParam("nonce", nonce))

	return ctx.Redirect(http.StatusFound, authURL)
}

// SSOTokenCallbackHandler completes the SSO authorization flow for a token.
// It validates the state and nonce, exchanges the code if required and updates
// the token's SSO authorizations for the organization.
func (h *Handler) SSOTokenCallbackHandler(ctx echo.Context, openapi *OpenAPIContext) error {
	in, err := BindAndValidateQueryParams[models.SSOTokenCallbackRequest](ctx, openapi.Operation, models.ExampleSSOTokenCallbackRequest, openapi.Registry)
	if err != nil {
		return h.InvalidInput(ctx, err, openapi)
	}

	req := ctx.Request()
	reqCtx := ctx.Request().Context()

	// read cookies set during the authorize step
	tokenIDCookie, err := sessions.GetCookie(req, "token_id")
	if err != nil {
		return h.BadRequest(ctx, ErrMissingField, openapi)
	}

	tokenTypeCookie, err := sessions.GetCookie(req, "token_type")
	if err != nil {
		return h.BadRequest(ctx, ErrMissingField, openapi)
	}

	orgCookie, err := sessions.GetCookie(req, "organization_id")
	if err != nil {
		return h.BadRequest(ctx, ErrMissingField, openapi)
	}

	stateCookie, err := sessions.GetCookie(req, "state")
	if err != nil || in.State != stateCookie.Value {
		return h.BadRequest(ctx, ErrStateMismatch, openapi)
	}

	nonceCookie, err := sessions.GetCookie(req, "nonce")
	if err != nil {
		return h.BadRequest(ctx, ErrNonceMissing, openapi)
	}

	rpCfg, err := h.oidcConfig(reqCtx, orgCookie.Value)
	if err != nil {
		return h.BadRequest(ctx, err, openapi)
	}

	nonceCtx := contextx.With(reqCtx, nonce(nonceCookie.Value))
	if _, err = rp.CodeExchange[*oidc.IDTokenClaims](nonceCtx, in.Code, rpCfg); err != nil {
		return h.BadRequest(ctx, err, openapi)
	}

	allowCtx := privacy.DecisionContext(reqCtx, privacy.Allow)
	now := time.Now()

	switch tokenTypeCookie.Value {
	case "api":
		tkn, err := h.DBClient.APIToken.Get(allowCtx, tokenIDCookie.Value)
		if err != nil {
			return h.BadRequest(ctx, err, openapi)
		}

		if tkn.SSOAuthorizations == nil {
			tkn.SSOAuthorizations = models.SSOAuthorizationMap{}
		}

		tkn.SSOAuthorizations[orgCookie.Value] = now

		_, err = h.DBClient.APIToken.UpdateOne(tkn).SetSSOAuthorizations(tkn.SSOAuthorizations).Save(allowCtx)
		if err != nil {
			return h.InternalServerError(ctx, err, openapi)
		}
	case "personal":
		tkn, err := h.DBClient.PersonalAccessToken.Get(allowCtx, tokenIDCookie.Value)
		if err != nil {
			return h.BadRequest(ctx, err, openapi)
		}

		if tkn.SSOAuthorizations == nil {
			tkn.SSOAuthorizations = models.SSOAuthorizationMap{}
		}

		tkn.SSOAuthorizations[orgCookie.Value] = now

		_, err = h.DBClient.PersonalAccessToken.UpdateOne(tkn).SetSSOAuthorizations(tkn.SSOAuthorizations).Save(allowCtx)
		if err != nil {
			return h.InternalServerError(ctx, err, openapi)
		}
	default:
		return h.BadRequest(ctx, errInvalidTokenType, openapi)
	}

	// cleanup cookies
	for _, name := range []string{"token_id", "token_type", "organization_id", "state", "nonce"} {
		sessions.RemoveCookie(ctx.Response().Writer, name, sessions.CookieConfig{Path: "/"})
	}

	out := models.SSOTokenAuthorizeReply{
		Reply:          rout.Reply{Success: true},
		OrganizationID: orgCookie.Value,
		TokenID:        tokenIDCookie.Value,
		Message:        "authorized",
	}

	return h.Success(ctx, out, openapi)
}

var errInvalidTokenType = errors.New("invalid token type")
