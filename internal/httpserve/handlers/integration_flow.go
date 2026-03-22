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

	requestCtx := ctx.Request().Context()

	caller, ok := auth.CallerFromContext(requestCtx)
	if !ok || caller == nil {
		return h.Unauthorized(ctx, auth.ErrNoAuthUser, openapiCtx)
	}

	def, ok := h.IntegrationsRuntime.Registry().Definition(in.DefinitionID)
	if !ok || !def.Active {
		return h.BadRequest(ctx, ErrInvalidProvider, openapiCtx)
	}

	connection, err := def.ConnectionRegistration(in.CredentialRef)
	if err != nil {
		return h.BadRequest(ctx, ErrUnsupportedAuthType, openapiCtx)
	}

	if connection.Auth == nil || connection.Auth.Start == nil {
		return h.BadRequest(ctx, ErrUnsupportedAuthType, openapiCtx)
	}

	installationRec, _, err := h.IntegrationsRuntime.EnsureInstallation(requestCtx, caller.OrganizationID, in.InstallationID, def)
	if err != nil {
		logx.FromContext(requestCtx).Error().Err(err).Interface("request", in).Msg("failed to resolve installation")
		return h.BadRequest(ctx, ErrIntegrationNotFound, openapiCtx)
	}

	if !jsonx.IsEmptyRawMessage(in.UserInput) {
		if err := h.IntegrationsRuntime.Reconcile(requestCtx, installationRec, in.UserInput, types.CredentialRef{}, nil, nil); err != nil {
			logx.FromContext(requestCtx).Error().Err(err).Interface("request", in).Msg("failed to reconcile user input")
			return h.InternalServerError(ctx, ErrProcessingRequest, openapiCtx)
		}
	}

	begin, err := h.IntegrationsRuntime.BeginAuth(requestCtx, keymaker.BeginRequest{
		DefinitionID:   def.ID,
		InstallationID: installationRec.ID,
		CredentialRef:  in.CredentialRef,
	})
	if err != nil {
		logx.FromContext(requestCtx).Error().Err(err).Interface("request", in).Msg("failed to begin auth flow")
		return h.BadRequest(ctx, ErrIntegrationNotFound, openapiCtx)
	}

	cfg := *h.SessionConfig.CookieConfig
	sessions.SetCookies(ctx.Response().Writer, cfg, map[string]string{
		"state":           begin.State,
		"organization_id": caller.OrganizationID,
	})
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

	callbackInput := normalizeIntegrationAuthCallbackInput(ctx.Request())
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

	caller, callerOk := auth.CallerFromContext(reqCtx)
	if !callerOk || caller == nil {
		return h.Unauthorized(ctx, auth.ErrNoAuthUser, openapiCtx)
	}

	if caller.OrganizationID != orgCookie.Value {
		logx.FromContext(reqCtx).Error().Msg("callback organization mismatch")
		return h.BadRequest(ctx, ErrInvalidOrganizationContext, openapiCtx)
	}

	if _, err = h.IntegrationsRuntime.CompleteAuth(reqCtx, keymaker.CompleteRequest{
		State:    stateCookie.Value,
		Callback: callbackInput,
	}); err != nil {
		return h.BadRequest(ctx, err, openapiCtx)
	}

	cfg := *h.SessionConfig.CookieConfig
	sessions.RemoveCookies(ctx.Response().Writer, cfg, "state", "organization_id")

	return h.Success(ctx, rout.Reply{Success: true})
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

// RefreshIntegrationTokenHandler handles requests to refresh an installation's auth credential
func (h *Handler) RefreshIntegrationTokenHandler(ctx echo.Context, openapiCtx *OpenAPIContext) error {
	in, err := BindAndValidateWithAutoRegistry(ctx, h, openapiCtx.Operation, RefreshInstallationCredentialRequest{}, IntegrationTokenResponse{}, openapiCtx.Registry)
	if err != nil {
		return h.InvalidInput(ctx, err, openapiCtx)
	}

	reqCtx := ctx.Request().Context()

	caller, ok := auth.CallerFromContext(reqCtx)
	if !ok || caller == nil {
		return h.Unauthorized(ctx, auth.ErrNoAuthUser, openapiCtx)
	}

	rec, err := h.IntegrationsRuntime.ResolveInstallation(reqCtx, caller.OrganizationID, in.InstallationID, "")
	if err != nil {
		logx.FromContext(reqCtx).Error().Err(err).Interface("request", in).Msg("failed to resolve installation")
		return h.BadRequest(ctx, ErrIntegrationNotFound, openapiCtx)
	}

	def, ok := h.IntegrationsRuntime.Registry().Definition(rec.DefinitionID)
	if !ok {
		return h.BadRequest(ctx, ErrUnsupportedAuthType, openapiCtx)
	}

	connection, err := h.IntegrationsRuntime.ResolvePersistedConnection(rec)
	if err != nil || connection.Auth == nil || connection.Auth.Refresh == nil {
		return h.BadRequest(ctx, ErrUnsupportedAuthType, openapiCtx)
	}

	current, credOk, err := h.IntegrationsRuntime.LoadCredential(reqCtx, rec, connection.Auth.CredentialRef)
	if err != nil {
		logx.FromContext(reqCtx).Error().Err(err).Interface("request", in).Msg("failed to load credential")
		return h.InternalServerError(ctx, ErrProcessingRequest, openapiCtx)
	}

	if !credOk {
		return h.BadRequest(ctx, ErrIntegrationNotFound, openapiCtx)
	}

	refreshed, err := connection.Auth.Refresh(reqCtx, current)
	if err != nil {
		logx.FromContext(reqCtx).Error().Err(err).Interface("request", in).Msg("credential refresh failed")
		return h.InternalServerError(ctx, ErrProcessingRequest, openapiCtx)
	}

	if err := h.IntegrationsRuntime.Reconcile(reqCtx, rec, nil, connection.Auth.CredentialRef, &refreshed, nil); err != nil {
		logx.FromContext(reqCtx).Error().Err(err).Interface("request", in).Msg("failed to reconcile refreshed credential")
		return h.InternalServerError(ctx, ErrProcessingRequest, openapiCtx)
	}

	tokenView, err := connection.Auth.TokenView(reqCtx, refreshed)
	if err != nil {
		logx.FromContext(reqCtx).Error().Err(err).Interface("request", in).Msg("failed to get token view")
		return h.InternalServerError(ctx, ErrProcessingRequest, openapiCtx)
	}

	resp := IntegrationTokenResponse{
		Reply:    rout.Reply{Success: true},
		Provider: def.Slug,
	}

	if tokenView != nil {
		resp.AccessToken = tokenView.AccessToken
		resp.ExpiresAt = tokenView.ExpiresAt
	}

	return h.Success(ctx, resp)
}
