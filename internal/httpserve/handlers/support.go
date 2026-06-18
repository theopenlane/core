package handlers

import (
	"context"
	"crypto/subtle"
	"strings"
	"time"

	echo "github.com/theopenlane/echox"
	"github.com/theopenlane/iam/auth"
	"github.com/theopenlane/iam/sessions"
	"github.com/theopenlane/iam/tokens"
	"github.com/theopenlane/utils/rout"
	"github.com/zitadel/oidc/v3/pkg/client/rp"
	"github.com/zitadel/oidc/v3/pkg/oidc"
	"golang.org/x/oauth2"

	apimodels "github.com/theopenlane/core/common/openapi"
	ent "github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/generated/privacy"
	"github.com/theopenlane/core/pkg/logx"
	sso "github.com/theopenlane/core/pkg/ssoutils"
)

// support access cookie names carry first factor state into the second factor identity provider step
const (
	supportPendingCookie  = "support_pending"
	supportStateCookie    = "support_state"
	supportNonceCookie    = "support_nonce"
	supportOrgCookie      = "support_org"
	supportReasonCookie   = "support_reason"
)

// supportSessionDefaultHours is the default support access session length when none is requested
const supportSessionDefaultHours = 1

// supportFirstFactor handles the first factor of the Openlane support login when the support identity
// email is used at the login endpoint. It validates the shared password against configuration (never
// the database), confirms the target organization consents to support access, and returns the second
// factor identity provider redirect. It is only reached from LoginHandler after the support email is
// detected, so the support email can never authenticate against a database user
func (h *Handler) supportFirstFactor(ctx echo.Context, openapi *OpenAPIContext, req *apimodels.LoginRequest) error {
	reqCtx := ctx.Request().Context()

	cfg := h.SupportAccessConfig

	// validate the shared password against configuration, never the database
	if cfg.Password == "" || subtle.ConstantTimeCompare([]byte(req.Password), []byte(cfg.Password)) != 1 {
		return h.Unauthorized(ctx, ErrSupportInvalidCredentials, openapi)
	}

	if req.TargetOrganizationID == "" || len(req.Reason) < apimodels.MinImpersonationReasonLength {
		return h.InvalidInput(ctx, ErrSupportLoginRequiresOrgAndReason, openapi)
	}

	// the target organization must have consented to support access
	allowCtx := privacy.DecisionContext(reqCtx, privacy.Allow)

	setting, err := h.getOrganizationSettingByOrgID(allowCtx, req.TargetOrganizationID)
	if err != nil {
		if ent.IsNotFound(err) {
			return h.NotFound(ctx, ErrNotFound, openapi)
		}

		logx.FromContext(reqCtx).Error().Err(err).Msg("error fetching organization setting")

		return h.InternalServerError(ctx, ErrProcessingRequest, openapi)
	}

	if !setting.AllowSupportAccess {
		return h.Forbidden(ctx, ErrSupportAccessNotConsented, openapi)
	}

	authURL, err := h.generateSupportAuthURL(ctx, req.TargetOrganizationID, req.Reason)
	if err != nil {
		logx.FromContext(reqCtx).Error().Err(err).Msg("unable to generate support second factor auth URL")

		return h.BadRequest(ctx, err, openapi)
	}

	return h.Success(ctx, apimodels.LoginReply{
		Reply:       rout.Reply{Success: true},
		RedirectURI: authURL,
	}, openapi)
}

// SupportCallbackHandler is the second factor of the Openlane support access flow. It completes the
// configured identity provider exchange, requires the individual's email to be within the configured
// domain, and mints a support session token attributed to the individual who completed it
func (h *Handler) SupportCallbackHandler(ctx echo.Context, openapi *OpenAPIContext) error {
	in, err := BindAndValidateQueryParamsWithResponse(ctx, openapi.Operation, apimodels.ExampleSupportCallbackRequest, apimodels.ExampleSupportAccessReply, openapi.Registry)
	if err != nil {
		return h.InvalidInput(ctx, err, openapi)
	}

	if isRegistrationContext(ctx) {
		return nil
	}

	reqCtx := ctx.Request().Context()

	cfg := h.SupportAccessConfig
	if !cfg.Enabled {
		return h.BadRequest(ctx, ErrSupportAccessNotEnabled, openapi)
	}

	// confirm the first factor was completed in this browser session
	if _, err := sessions.GetCookie(ctx.Request(), supportPendingCookie); err != nil {
		return h.BadRequest(ctx, ErrSupportFirstFactorRequired, openapi)
	}

	stateCookie, err := sessions.GetCookie(ctx.Request(), supportStateCookie)
	if err != nil || in.State == "" || in.State != stateCookie.Value {
		return h.BadRequest(ctx, ErrStateMismatch, openapi)
	}

	nonceCookie, err := sessions.GetCookie(ctx.Request(), supportNonceCookie)
	if err != nil {
		return h.BadRequest(ctx, ErrNonceMissing, openapi)
	}

	orgCookie, err := sessions.GetCookie(ctx.Request(), supportOrgCookie)
	if err != nil {
		return h.BadRequest(ctx, ErrMissingField, openapi)
	}

	rpCfg, err := h.supportOIDCConfig(reqCtx)
	if err != nil {
		return h.BadRequest(ctx, err, openapi)
	}

	// exchange the code for the identity provider tokens, verifying the nonce
	nonceCtx := ssoNonceContextKey.Set(reqCtx, nonce(nonceCookie.Value))

	idTokens, err := rp.CodeExchange[*oidc.IDTokenClaims](nonceCtx, in.Code, rpCfg)
	if err != nil {
		return h.BadRequest(ctx, err, openapi)
	}

	individualEmail := idTokens.IDTokenClaims.Email

	// domain restriction: only individuals within the configured domain may complete support access
	if !emailInDomain(individualEmail, cfg.AllowedDomain) {
		logx.FromContext(reqCtx).Warn().Str("email", individualEmail).Msg("support second factor email domain not allowed")

		return h.Forbidden(ctx, ErrSupportDomainNotAllowed, openapi)
	}

	// re-confirm the organization still consents to support access
	allowCtx := privacy.DecisionContext(reqCtx, privacy.Allow)

	setting, err := h.getOrganizationSettingByOrgID(allowCtx, orgCookie.Value)
	if err != nil || !setting.AllowSupportAccess {
		return h.Forbidden(ctx, ErrSupportAccessNotConsented, openapi)
	}

	reason := ""
	if c, cErr := sessions.GetCookie(ctx.Request(), supportReasonCookie); cErr == nil {
		reason = c.Value
	}

	duration := time.Duration(supportSessionDefaultHours) * time.Hour

	if h.TokenManager == nil {
		logx.FromContext(reqCtx).Error().Msg("token manager not configured")

		return h.InternalServerError(ctx, ErrProcessingRequest, openapi)
	}

	// mint the support session token targeting the virtual support identity, attributed to the individual
	token, err := h.TokenManager.CreateImpersonationToken(reqCtx, tokens.CreateImpersonationTokenOptions{
		ImpersonatorID:    individualEmail,
		ImpersonatorEmail: individualEmail,
		TargetUserID:      cfg.SubjectID,
		TargetUserEmail:   cfg.Email,
		OrganizationID:    orgCookie.Value,
		Type:              string(auth.SupportImpersonation),
		Reason:            reason,
		Duration:          duration,
		Scopes:            []string{"read", "write"},
	})
	if err != nil {
		logx.FromContext(reqCtx).Error().Err(err).Msg("error creating support access token")

		return h.InternalServerError(ctx, ErrProcessingRequest, openapi)
	}

	sessionID, err := h.extractSessionIDFromToken(token)
	if err != nil {
		logx.FromContext(reqCtx).Error().Err(err).Msg("failed to extract session ID from support access token")

		return h.InternalServerError(ctx, ErrFailedToExtractSessionID, openapi)
	}

	auditLog := &auth.ImpersonationAuditLog{
		Type:              auth.SupportImpersonation,
		ImpersonatorID:    individualEmail,
		ImpersonatorEmail: individualEmail,
		TargetUserID:      cfg.SubjectID,
		TargetUserEmail:   cfg.Email,
		Action:            "start",
		Reason:            reason,
		Timestamp:         time.Now(),
		IPAddress:         ctx.RealIP(),
		UserAgent:         ctx.Request().UserAgent(),
		OrganizationID:    orgCookie.Value,
		Scopes:            []string{"read", "write"},
	}

	if err := h.logImpersonationEvent(reqCtx, "start", auditLog); err != nil {
		logx.FromContext(reqCtx).Error().Err(err).Msg("failed to log support access event")
	}

	sessions.RemoveCookies(ctx.Response().Writer, sessions.CookieConfig{Path: "/"},
		supportPendingCookie, supportStateCookie, supportNonceCookie, supportOrgCookie, supportReasonCookie)

	return h.Success(ctx, apimodels.SupportAccessReply{
		Reply:          rout.Reply{Success: true},
		Token:          token,
		ExpiresAt:      time.Now().Add(duration),
		SessionID:      sessionID,
		OrganizationID: orgCookie.Value,
		Impersonator:   individualEmail,
		Message:        "Support access session started successfully",
	}, openapi)
}

// supportOIDCConfig builds the relying party for the configured support second factor identity provider
func (h *Handler) supportOIDCConfig(ctx context.Context) (rp.RelyingParty, error) {
	cfg := h.SupportAccessConfig
	if cfg.ClientID == "" || cfg.ClientSecret == "" || (cfg.IssuerURL == "" && cfg.DiscoveryEndpoint == "") {
		return nil, ErrSupportIDPNotConfigured
	}

	issuer := cfg.IssuerURL
	if issuer == "" {
		issuer = strings.TrimSuffix(cfg.DiscoveryEndpoint, "/.well-known/openid-configuration")
	}

	verifierOpt := rp.WithVerifierOpts(rp.WithNonce(func(ctx context.Context) string {
		if n, ok := ssoNonceContextKey.Get(ctx); ok {
			return string(n)
		}

		return ""
	}))

	opts := []rpConfigOption{withRPOptions(verifierOpt)}
	if cfg.DiscoveryEndpoint != "" {
		opts = append(opts, withDiscovery(cfg.DiscoveryEndpoint))
	}

	return newRelyingParty(ctx, issuer, cfg.ClientID, cfg.ClientSecret, cfg.RedirectURL, opts...)
}

// generateSupportAuthURL stores the first factor state in cookies and returns the second factor
// identity provider authorization URL
func (h *Handler) generateSupportAuthURL(ctx echo.Context, orgID, reason string) (string, error) {
	rpCfg, err := h.supportOIDCConfig(ctx.Request().Context())
	if err != nil {
		return "", err
	}

	cookieCfg := *h.SessionConfig.CookieConfig

	state, err := auth.GenerateOAuthState(stateLength)
	if err != nil {
		return "", err
	}

	nonceVal, err := auth.GenerateOAuthState(stateLength)
	if err != nil {
		return "", err
	}

	sessions.SetCookies(ctx.Response().Writer, cookieCfg, map[string]string{
		supportPendingCookie:  authenticatedUserSSOCookieValue,
		supportOrgCookie:      orgID,
		supportReasonCookie:   reason,
		supportStateCookie:    state,
		supportNonceCookie:    nonceVal,
	})

	return rpCfg.OAuthConfig().AuthCodeURL(state, oauth2.SetAuthURLParam("nonce", nonceVal)), nil
}

// emailInDomain reports whether email belongs to domain, case-insensitively
func emailInDomain(email, domain string) bool {
	if domain == "" {
		return false
	}

	d := sso.EmailDomain(email)

	return d != "" && strings.EqualFold(d, domain)
}
