package handlers

import (
	"errors"
	"time"

	"github.com/theopenlane/core/common/models"
	apimodels "github.com/theopenlane/core/common/openapi"
	"github.com/theopenlane/core/internal/ent/generated/privacy"
	"github.com/theopenlane/core/pkg/logx"
	echo "github.com/theopenlane/echox"
	"github.com/theopenlane/iam/sessions"
	"github.com/theopenlane/utils/rout"
	"github.com/zitadel/oidc/v3/pkg/client/rp"
	"github.com/zitadel/oidc/v3/pkg/oidc"
)

var errInvalidTokenType = errors.New("invalid token type")

// SSOTokenAuthorizeHandler marks a token as authorized for SSO for an organization
func (h *Handler) SSOTokenAuthorizeHandler(ctx echo.Context, openapi *OpenAPIContext) error {
	in, err := BindAndValidateQueryParamsWithResponse(ctx, openapi.Operation, apimodels.ExampleSSOTokenAuthorizeRequest, apimodels.SSOTokenAuthorizeReply{}, openapi.Registry)
	if err != nil {
		return h.InvalidInput(ctx, err, openapi)
	}

	// Skip actual handler logic during OpenAPI registration
	if isRegistrationContext(ctx) {
		return nil
	}

	reqCtx := privacy.DecisionContext(ctx.Request().Context(), privacy.Allow)

	switch in.TokenType {
	case "api":
		if _, err := h.DBClient.APIToken.Get(reqCtx, in.TokenID); err != nil {
			logx.FromContext(reqCtx).Error().Err(err).Msg("unable to find api token for SSO")

			return h.BadRequest(ctx, err, openapi)
		}
	case "personal":
		if _, err := h.DBClient.PersonalAccessToken.Get(reqCtx, in.TokenID); err != nil {
			logx.FromContext(reqCtx).Error().Err(err).Msg("unable to find personal access token")

			return h.BadRequest(ctx, err, openapi)
		}
	default:
		return h.BadRequest(ctx, errInvalidTokenType, openapi)
	}

	authURL, err := h.generateSSOAuthURL(ctx, in.OrganizationID)
	if err != nil {
		return h.BadRequest(ctx, err, openapi)
	}

	// set token-specific cookies for the token SSO flow
	cfg := *h.SessionConfig.CookieConfig

	sessions.SetCookies(ctx.Response().Writer, cfg, map[string]string{
		"token_id":                     in.TokenID,
		"token_type":                   in.TokenType,
		authenticatedUserSSOCookieName: authenticatedUserSSOCookieValue,
	})

	out := apimodels.SSOLoginReply{
		Reply:       rout.Reply{Success: true},
		RedirectURI: authURL,
	}

	return h.Success(ctx, out, openapi)
}

// SSOTokenCallbackHandler completes the SSO authorization flow for a token.
// It validates the state and nonce, exchanges the code if required and updates
// the token's SSO authorizations for the organization.
func (h *Handler) SSOTokenCallbackHandler(ctx echo.Context, openapi *OpenAPIContext) error {
	in, err := BindAndValidateQueryParamsWithResponse(ctx, openapi.Operation, apimodels.ExampleSSOTokenCallbackRequest, apimodels.SSOTokenAuthorizeReply{}, openapi.Registry)
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

	nonceCtx := ssoNonceContextKey.Set(reqCtx, nonce(nonceCookie.Value))
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
			logx.FromContext(reqCtx).Error().Err(err).Msg("error updating api token")

			return h.InternalServerError(ctx, ErrProcessingRequest, openapi)
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
			logx.FromContext(reqCtx).Error().Err(err).Msg("error updating personal access token")

			return h.InternalServerError(ctx, ErrProcessingRequest, openapi)
		}
	default:
		return h.BadRequest(ctx, errInvalidTokenType, openapi)
	}

	// cleanup cookies
	sessions.RemoveCookies(ctx.Response().Writer, sessions.CookieConfig{Path: "/"}, "token_id", "token_type", "organization_id", "state", "nonce")

	out := apimodels.SSOTokenAuthorizeReply{
		Reply:          rout.Reply{Success: true},
		OrganizationID: orgCookie.Value,
		TokenID:        tokenIDCookie.Value,
		Message:        "authorized",
	}

	return h.Success(ctx, out, openapi)
}
