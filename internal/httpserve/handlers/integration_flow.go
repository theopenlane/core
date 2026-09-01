package handlers

import (
	"maps"
	"net/http"
	"net/url"
	"slices"
	"strings"

	echo "github.com/theopenlane/echox"

	"github.com/theopenlane/iam/auth"
	"github.com/theopenlane/iam/sessions"
	"github.com/theopenlane/utils/rout"

	openapi "github.com/theopenlane/core/common/openapi"
	"github.com/theopenlane/core/v2/internal/integrations/types"
	"github.com/theopenlane/core/v2/internal/keymaker"
	"github.com/theopenlane/core/v2/pkg/jsonx"
	"github.com/theopenlane/core/v2/pkg/logx"
)

// StartIntegrationAuth initiates the auth flow for an integration definition
func (h *Handler) StartIntegrationAuth(ctx echo.Context) error {
	in, err := BindAndValidate[IntegrationAuthStartRequest](ctx)
	if err != nil {
		return h.InvalidInput(ctx, err)
	}

	if h.IntegrationsRuntime == nil {
		return h.BadRequest(ctx, ErrIntegrationsNotEnabled)
	}

	requestCtx := ctx.Request().Context()

	caller, ok := auth.CallerFromContext(requestCtx)
	if !ok || caller == nil {
		return h.Unauthorized(ctx, auth.ErrNoAuthUser)
	}

	def, ok := h.IntegrationsRuntime.Registry().Definition(in.DefinitionID)
	if !ok || !def.Active {
		return h.BadRequest(ctx, ErrInvalidProvider)
	}

	credentialRef := types.NewCredentialSlotID(in.CredentialRef)

	connection, err := def.ConnectionRegistration(credentialRef)
	if err != nil {
		return h.BadRequest(ctx, ErrUnsupportedAuthType)
	}

	if connection.Auth == nil || connection.Auth.Start == nil {
		return h.BadRequest(ctx, ErrUnsupportedAuthType)
	}

	// if integrationID is empty, we assume this is a new installation and proceed to create a record that the auth flow can reference; if it is provided we will attempt to resolve and reuse the existing installation record for the auth flow
	installationRec, _, err := h.IntegrationsRuntime.EnsureInstallation(requestCtx, caller.OrganizationID, in.IntegrationID, def)
	if err != nil {
		logx.FromContext(requestCtx).Error().Err(err).Interface("request", in).Msg("failed to resolve integration")

		return h.BadRequest(ctx, ErrIntegrationNotFound)
	}

	// required user input has to be satisfied before we hand out provider scopes, otherwise the
	// install authorizes successfully and then fails every sync it runs
	effectiveInput := in.UserInput
	if jsonx.IsEmptyRawMessage(effectiveInput) {
		effectiveInput = installationRec.Config.ClientConfig
	}

	if err := h.IntegrationsRuntime.ValidateUserInput(requestCtx, def, effectiveInput); err != nil {
		logx.FromContext(requestCtx).Warn().Err(err).Str("definition_id", def.ID).Msg("integration user input incomplete, refusing to start auth flow")

		return h.BadRequest(ctx, ErrIntegrationUserInputRequired)
	}

	// if we got optional config with the input, persist it
	if !jsonx.IsEmptyRawMessage(in.UserInput) {
		if err := h.IntegrationsRuntime.Reconcile(requestCtx, installationRec, in.UserInput, types.CredentialSlotID{}, nil, nil); err != nil {
			logx.FromContext(requestCtx).Error().Err(err).Interface("request", in).Msg("failed to reconcile user input")

			return h.InternalServerError(ctx, ErrProcessingRequest)
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

		return h.BadRequest(ctx, ErrIntegrationNotFound)
	}

	cfg := *h.SessionConfig.CookieConfig
	cookies := map[string]string{
		"state":           begin.State,
		"organization_id": caller.OrganizationID,
	}

	// ConsoleURL is the full base URL for the frontend (e.g. https://console.theopenlane.io).
	// Accept either form with or without a trailing slash.
	redirectTo := strings.TrimRight(h.ConsoleURL, "/") + h.IntegrationsConfig.ConsoleIntegrationPath + "/" + def.ID
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
func (h *Handler) HandleIntegrationAuthCallback(ctx echo.Context) error {
	if h.IntegrationsRuntime == nil {
		return h.BadRequest(ctx, ErrIntegrationsNotEnabled)
	}

	reqCtx := ctx.Request().Context()

	stateCookie, err := sessions.GetCookie(ctx.Request(), "state")
	if err != nil {
		logx.FromContext(reqCtx).Error().Err(err).Msg("state cookie not found")

		return h.BadRequest(ctx, ErrInvalidState)
	}

	orgCookie, err := sessions.GetCookie(ctx.Request(), "organization_id")
	if err != nil {
		logx.FromContext(reqCtx).Error().Err(err).Msg("organization_id cookie not found")

		return h.BadRequest(ctx, ErrMissingOrganizationContext)
	}

	callbackInput := normalizeIntegrationAuthCallbackInput(ctx.Request())

	reqCtx = auth.WithCaller(reqCtx, auth.NewWebhookCaller(orgCookie.Value))

	_, err = h.IntegrationsRuntime.CompleteAuth(reqCtx, keymaker.CompleteRequest{
		State:    stateCookie.Value,
		Callback: callbackInput,
	})

	redirectTo := h.ConsoleURL
	if redirectCookie, cookieErr := sessions.GetCookie(ctx.Request(), "redirect_to"); cookieErr == nil {
		redirectTo = redirectCookie.Value
	}

	h.clearAuthFlowCookies(ctx.Response().Writer, "state", "organization_id", "redirect_to")

	if err != nil {
		logx.FromContext(reqCtx).Warn().Err(err).Msg("integration auth callback failed")

		return h.Redirect(ctx, appendAuthFailure(redirectTo))
	}

	return h.Redirect(ctx, redirectTo)
}

const (
	// integrationAuthErrorParam is the query param carrying the auth failure code back to the console
	integrationAuthErrorParam = "error"
	// integrationAuthErrorValue is the auth failure code sent back to the console
	integrationAuthErrorValue = "auth_failed"
)

// appendAuthFailure returns the redirect target with the auth failure error param applied
func appendAuthFailure(redirectTo string) string {
	target, err := url.Parse(redirectTo)
	if err != nil {
		return redirectTo
	}

	q := target.Query()
	q.Set(integrationAuthErrorParam, integrationAuthErrorValue)
	target.RawQuery = q.Encode()

	return target.String()
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
