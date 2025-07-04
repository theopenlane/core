package handlers

import (
	"errors"
	"net/http"
	"time"

	"github.com/getkin/kin-openapi/openapi3"
	echo "github.com/theopenlane/echox"
	"github.com/zitadel/oidc/pkg/client/rp"

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
func (h *Handler) SSOTokenCallbackHandler(ctx echo.Context) error {
	var in models.SSOTokenCallbackRequest
	if err := ctx.Bind(&in); err != nil {
		return h.InvalidInput(ctx, err)
	}

	if err := in.Validate(); err != nil {
		return h.BadRequest(ctx, err)
	}

	req := ctx.Request()
	reqCtx := ctx.Request().Context()

	// read cookies set during the authorize step
	tokenIDCookie, err := sessions.GetCookie(req, "token_id")
	if err != nil {
		return h.BadRequest(ctx, ErrMissingField)
	}

	tokenTypeCookie, err := sessions.GetCookie(req, "token_type")
	if err != nil {
		return h.BadRequest(ctx, ErrMissingField)
	}

	orgCookie, err := sessions.GetCookie(req, "organization_id")
	if err != nil {
		return h.BadRequest(ctx, ErrMissingField)
	}

	stateCookie, err := sessions.GetCookie(req, "state")
	if err != nil || in.State != stateCookie.Value {
		return h.BadRequest(ctx, ErrStateMismatch)
	}

	nonceCookie, err := sessions.GetCookie(req, "nonce")
	if err != nil {
		return h.BadRequest(ctx, ErrNonceMissing)
	}

	rpCfg, err := h.oidcConfig(reqCtx, orgCookie.Value)
	if err != nil {
		return h.BadRequest(ctx, err)
	}

	nonceCtx := contextx.With(reqCtx, nonce(nonceCookie.Value))
	if _, err = rp.CodeExchange(nonceCtx, in.Code, rpCfg); err != nil {
		return h.BadRequest(ctx, err)
	}

	allowCtx := privacy.DecisionContext(reqCtx, privacy.Allow)
	now := time.Now()

	switch tokenTypeCookie.Value {
	case "api":
		tkn, err := h.DBClient.APIToken.Get(allowCtx, tokenIDCookie.Value)
		if err != nil {
			return h.BadRequest(ctx, err)
		}

		if tkn.SSOAuthorizations == nil {
			tkn.SSOAuthorizations = models.SSOAuthorizationMap{}
		}

		tkn.SSOAuthorizations[orgCookie.Value] = now

		_, err = h.DBClient.APIToken.UpdateOne(tkn).SetSSOAuthorizations(tkn.SSOAuthorizations).Save(allowCtx)
		if err != nil {
			return h.InternalServerError(ctx, err)
		}
	case "personal":
		tkn, err := h.DBClient.PersonalAccessToken.Get(allowCtx, tokenIDCookie.Value)
		if err != nil {
			return h.BadRequest(ctx, err)
		}

		if tkn.SSOAuthorizations == nil {
			tkn.SSOAuthorizations = models.SSOAuthorizationMap{}
		}

		tkn.SSOAuthorizations[orgCookie.Value] = now

		_, err = h.DBClient.PersonalAccessToken.UpdateOne(tkn).SetSSOAuthorizations(tkn.SSOAuthorizations).Save(allowCtx)
		if err != nil {
			return h.InternalServerError(ctx, err)
		}
	default:
		return h.BadRequest(ctx, errInvalidTokenType)
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

	return h.Success(ctx, out)
}

// BindSSOTokenCallbackHandler binds the SSO token callback endpoint to the OpenAPI schema.
func (h *Handler) BindSSOTokenCallbackHandler() *openapi3.Operation {
	op := openapi3.NewOperation()
	op.Description = "Complete token SSO authorization"
	op.OperationID = "SSOTokenCallback"
	op.Tags = []string{"authentication"}
	op.Security = &openapi3.SecurityRequirements{}

	h.AddQueryParameter("code", op)
	h.AddQueryParameter("state", op)
	h.AddResponse("SSOTokenAuthorizeReply", "success", models.ExampleSSOTokenAuthorizeReply, op, http.StatusOK)
	op.AddResponse(http.StatusBadRequest, badRequest())
	op.AddResponse(http.StatusInternalServerError, internalServerError())

	return op
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
