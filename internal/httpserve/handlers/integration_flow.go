package handlers

import (
	"maps"
	"net/http"
	"slices"

	echo "github.com/theopenlane/echox"

	"github.com/theopenlane/iam/auth"
	"github.com/theopenlane/iam/sessions"
	"github.com/theopenlane/utils/rout"

	openapi "github.com/theopenlane/core/common/openapi"
	"github.com/theopenlane/core/internal/integrations/types"
	"github.com/theopenlane/core/internal/keymaker"
	"github.com/theopenlane/core/pkg/jsonx"
	"github.com/theopenlane/core/pkg/logx"
)

// StartIntegrationAuth initiates the auth flow for an integration definition
func (h *Handler) StartIntegrationAuth(ctx echo.Context, openapiCtx *OpenAPIContext) error {
	in, err := BindAndValidateWithAutoRegistry(ctx, h, openapiCtx.Operation, ExampleIntegrationAuthStartRequest, openapi.OAuthFlowResponse{}, openapiCtx.Registry)
	if err != nil {
		return h.InvalidInput(ctx, err, openapiCtx)
	}

	if isRegistrationContext(ctx) {
		return nil
	}

	requestCtx := ctx.Request().Context()

	caller, ok := auth.CallerFromContext(requestCtx)
	if !ok || caller == nil {
		return h.Unauthorized(ctx, auth.ErrNoAuthUser, openapiCtx)
	}

	def, ok := h.IntegrationsRuntime.Registry().Definition(in.DefinitionID)
	if !ok || !def.Active {
		return h.BadRequest(ctx, ErrInvalidProvider, openapiCtx)
	}

	credentialRef := types.NewCredentialSlotID(in.CredentialRef)

	connection, err := def.ConnectionRegistration(credentialRef)
	if err != nil {
		return h.BadRequest(ctx, ErrUnsupportedAuthType, openapiCtx)
	}

	if connection.Auth == nil || connection.Auth.Start == nil {
		return h.BadRequest(ctx, ErrUnsupportedAuthType, openapiCtx)
	}

	// if integrationID is empty, we assume this is a new installation and proceed to create a record that the auth flow can reference; if it is provided we will attempt to resolve and reuse the existing installation record for the auth flow
	installationRec, _, err := h.IntegrationsRuntime.EnsureInstallation(requestCtx, caller.OrganizationID, in.IntegrationID, def)
	if err != nil {
		logx.FromContext(requestCtx).Error().Err(err).Interface("request", in).Msg("failed to resolve integration")

		return h.BadRequest(ctx, ErrIntegrationNotFound, openapiCtx)
	}

	// if we got optional config with the input, persist it
	if !jsonx.IsEmptyRawMessage(in.UserInput) {
		if err := h.IntegrationsRuntime.Reconcile(requestCtx, installationRec, in.UserInput, types.CredentialSlotID{}, nil, nil); err != nil {
			logx.FromContext(requestCtx).Error().Err(err).Interface("request", in).Msg("failed to reconcile user input")

			return h.InternalServerError(ctx, ErrProcessingRequest, openapiCtx)
		}
	}

	// we should basically never be trying to start auth flow without an integration record at this point
	begin, err := h.IntegrationsRuntime.BeginAuth(requestCtx, keymaker.BeginRequest{
		DefinitionID:   def.ID,
		InstallationID: installationRec.ID,
		CredentialRef:  credentialRef,
	})
	if err != nil {
		logx.FromContext(requestCtx).Error().Err(err).Interface("request", in).Msg("failed to begin auth flow")

		return h.BadRequest(ctx, ErrIntegrationNotFound, openapiCtx)
	}

	cfg := *h.SessionConfig.CookieConfig
	cookies := map[string]string{
		"state":           begin.State,
		"organization_id": caller.OrganizationID,
	}

	// ConsoleURL is the full base URL for the frontend (e.g. https://console.theopenlane.io).
	// The redirect path is derived from the definition ID so the browser lands on the integration detail page.
	redirectTo := h.ConsoleURL + "organization-settings/integrations/" + def.ID
	cookies["redirect_to"] = redirectTo

	sessions.SetCookies(ctx.Response().Writer, cfg, cookies)

	sessions.CopyCookiesFromRequest(ctx.Request(), ctx.Response().Writer, cfg, auth.AccessTokenCookie, auth.RefreshTokenCookie)

	return h.Success(ctx, openapi.OAuthFlowResponse{
		Reply:   rout.Reply{Success: true},
		AuthURL: begin.AuthURL,
		State:   begin.State,
	})
}

// HandleIntegrationAuthCallback processes the auth callback and delegates credential persistence to keymaker
func (h *Handler) HandleIntegrationAuthCallback(ctx echo.Context, openapiCtx *OpenAPIContext) error {
	if isRegistrationContext(ctx) {
		return nil
	}

	reqCtx := ctx.Request().Context()

	stateCookie, err := sessions.GetCookie(ctx.Request(), "state")
	if err != nil {
		logx.FromContext(reqCtx).Error().Err(err).Msg("state cookie not found")

		return h.BadRequest(ctx, ErrInvalidState, openapiCtx)
	}

	orgCookie, err := sessions.GetCookie(ctx.Request(), "organization_id")
	if err != nil {
		logx.FromContext(reqCtx).Error().Err(err).Msg("organization_id cookie not found")

		return h.BadRequest(ctx, ErrMissingOrganizationContext, openapiCtx)
	}

	callbackInput := normalizeIntegrationAuthCallbackInput(ctx.Request())

	reqCtx = auth.WithCaller(reqCtx, auth.NewWebhookCaller(orgCookie.Value))

	_, err = h.IntegrationsRuntime.CompleteAuth(reqCtx, keymaker.CompleteRequest{
		State:    stateCookie.Value,
		Callback: callbackInput,
	})
	if err != nil {
		return h.BadRequest(ctx, err, openapiCtx)
	}

	redirectTo := h.ConsoleURL
	if redirectCookie, cookieErr := sessions.GetCookie(ctx.Request(), "redirect_to"); cookieErr == nil {
		redirectTo = redirectCookie.Value
	}

	cfg := *h.SessionConfig.CookieConfig
	sessions.RemoveCookies(ctx.Response().Writer, cfg, "state", "organization_id", "redirect_to")

	return h.Redirect(ctx, redirectTo, openapiCtx)
}

// normalizeIntegrationAuthCallbackInput snapshots query params from the callback request
func normalizeIntegrationAuthCallbackInput(req *http.Request) types.AuthCallbackInput {
	params := req.URL.Query()
	input := types.AuthCallbackInput{
		Query: make([]types.AuthCallbackValue, 0, len(params)),
	}

	for _, key := range slices.Sorted(maps.Keys(params)) {
		values := params[key]
		copied := make([]string, len(values))
		copy(copied, values)
		input.Query = append(input.Query, types.AuthCallbackValue{
			Name:   key,
			Values: copied,
		})
	}

	return input
}
