package handlers

import (
	"context"
	"errors"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/stoewer/go-strcase"
	echo "github.com/theopenlane/echox"
	"github.com/theopenlane/iam/auth"
	"github.com/theopenlane/iam/fgax"
	iamoauth "github.com/theopenlane/iam/oauthserver"
	"github.com/theopenlane/logx"

	"github.com/theopenlane/core/common/enums"
	models "github.com/theopenlane/core/common/openapi"
	fgamodel "github.com/theopenlane/core/v2/fga/model"
	ent "github.com/theopenlane/core/v2/internal/ent/generated"
	"github.com/theopenlane/core/v2/internal/ent/generated/oauthgrant"
	"github.com/theopenlane/core/v2/internal/ent/generated/oauthtoken"
	"github.com/theopenlane/core/v2/internal/ent/generated/privacy"
	"github.com/theopenlane/core/v2/internal/ent/privacy/rule"
	"github.com/theopenlane/core/v2/internal/oauthserver"
	"github.com/theopenlane/core/v2/internal/oauthserver/store"
)

// oauthConsolePath is the console page that renders the consent screen
const oauthConsolePath = "/oauth/consent"

// installOrgAppRelation is the permission to install an organization app, the same as creating an API token
var installOrgAppRelation = rule.CanCreatePrefix + strcase.SnakeCase(ent.TypeAPIToken)

// oauthInstall is a validated consent decision
type oauthInstall struct {
	kind   enums.OAuthGrantKind
	access enums.OAuthAccessLevel
	scopes []string
}

// oauthIssuer is the authorization server issuer, such as https://api.theopenlane.io
func (h *Handler) oauthIssuer() string {
	return strings.TrimRight(h.TokenManager.Config().Issuer, "/")
}

// OAuthServerMetadataHandler serves RFC 8414 authorization server metadata
func (h *Handler) OAuthServerMetadataHandler(ctx echo.Context) error {
	scopes, err := fgamodel.ScopeList()
	if err != nil {
		logx.FromContext(ctx.Request().Context()).Error().Err(err).Msg("oauth metadata: unable to list scopes")
	}

	return ctx.JSON(http.StatusOK, iamoauth.NewMetadata(h.oauthIssuer(), scopes))
}

// OAuthAuthorizeHandler validates an authorization request, stores it, and sends the browser to the consent screen
// errors are never redirected, since clients are not pre-registered and a redirect would make this an open redirect
func (h *Handler) OAuthAuthorizeHandler(ctx echo.Context) error {
	reqCtx := ctx.Request().Context()
	q := ctx.Request().URL.Query()

	if err := iamoauth.SingleValued(q, iamoauth.AuthorizeParams...); err != nil {
		return oauthAuthorizeError(ctx, iamoauth.ErrCodeInvalidRequest, err.Error())
	}

	client, err := h.OAuthClients.Resolve(reqCtx, q.Get(iamoauth.ParamClientID))
	if err != nil {
		logx.FromContext(reqCtx).Info().Err(err).Str("client_id", q.Get(iamoauth.ParamClientID)).Msg("oauth authorize: invalid client")

		return oauthAuthorizeError(ctx, iamoauth.ErrCodeInvalidClient, err.Error())
	}

	redirectURI := q.Get(iamoauth.ParamRedirectURI)
	if err := client.ValidateRedirectURI(redirectURI); err != nil {
		return oauthAuthorizeError(ctx, iamoauth.ErrCodeInvalidRequest, err.Error())
	}

	if q.Get(iamoauth.ParamResponseType) != iamoauth.ResponseTypeCode {
		return oauthAuthorizeError(ctx, iamoauth.ErrCodeUnsupportedResponse, "only response_type=code is supported")
	}

	challenge := q.Get(iamoauth.ParamCodeChallenge)
	if !iamoauth.ValidChallenge(challenge, q.Get(iamoauth.ParamCodeChallengeMethod)) {
		return oauthAuthorizeError(ctx, iamoauth.ErrCodeInvalidRequest, "PKCE with code_challenge_method=S256 is required")
	}

	scopes := iamoauth.ParseScope(q.Get(iamoauth.ParamScope))

	if _, _, err := oauthserver.ClassifyScopes(scopes); err != nil {
		return oauthAuthorizeError(ctx, iamoauth.ErrCodeInvalidScope, err.Error())
	}

	resource, err := iamoauth.ResolveResource(h.OAuthServerConfig.AllowedResources, q[iamoauth.ParamResource])
	if err != nil {
		return oauthAuthorizeError(ctx, iamoauth.ErrCodeInvalidTarget, err.Error())
	}

	raw, create, err := h.newOAuthToken(iamoauth.KindRequest)
	if err != nil {
		return oauthAuthorizeError(ctx, iamoauth.ErrCodeServerError, "unable to create authorization request")
	}

	if err := create.
		SetClientID(client.ID).
		SetRedirectURI(redirectURI).
		SetCodeChallenge(challenge).
		SetState(q.Get(iamoauth.ParamState)).
		SetResource(resource).
		SetScopes(scopes).
		Exec(auth.WithInternalOperationContext(reqCtx)); err != nil {
		logx.FromContext(reqCtx).Error().Err(err).Msg("oauth authorize: unable to store request")

		return oauthAuthorizeError(ctx, iamoauth.ErrCodeServerError, "unable to create authorization request")
	}

	consent := strings.TrimRight(h.ConsoleURL, "/") + oauthConsolePath + "?" + url.Values{"request_id": {raw}}.Encode()

	return ctx.Redirect(http.StatusFound, consent)
}

// oauthAuthorizeError writes an authorize error as JSON, never as a redirect
func oauthAuthorizeError(ctx echo.Context, code, description string) error {
	return ctx.JSON(http.StatusBadRequest, iamoauth.NewError(code, description))
}

// OAuthAuthorizationRequestHandler returns a pending authorization request so the console can render the consent screen
func (h *Handler) OAuthAuthorizationRequestHandler(ctx echo.Context) error {
	in, err := BindAndValidate[models.OAuthAuthorizationRequestRequest](ctx)
	if err != nil {
		return h.InvalidInput(ctx, err)
	}

	reqCtx := ctx.Request().Context()

	req, err := h.pendingOAuthRequest(reqCtx, in.RequestID)
	if err != nil {
		return h.NotFound(ctx, err)
	}

	client, err := h.OAuthClients.Resolve(reqCtx, req.ClientID)
	if err != nil {
		return h.BadRequest(ctx, err)
	}

	out := models.OAuthAuthorizationRequestReply{
		RequestID:  in.RequestID,
		ClientID:   client.ID,
		ClientName: client.Name,
		ClientURI:  client.URI,
		LogoURI:    client.LogoURI,
		Scopes:     req.Scopes,
		Resource:   req.Resource,
		ExpiresAt:  req.ExpiresAt,
	}

	if kind, decided, err := oauthserver.ClassifyScopes(req.Scopes); err == nil {
		if decided {
			out.Kind = kind
		}

		if kind != enums.OAuthGrantKindOrganization {
			out.MaxAccessLevel = oauthserver.MaxAccess(req.Scopes)
			out.DefaultScopes = h.defaultPersonalScopes(reqCtx, req.Scopes, req.Resource)
		}
	}

	return h.Success(ctx, out)
}

// defaultPersonalScopes returns the scopes to pre-select for a personal connection, what the client named,
// or when it only asked for broad access, the object scopes its resource advertises so the user starts from least privilege
func (h *Handler) defaultPersonalScopes(ctx context.Context, requested []string, resource string) []string {
	fallback := requested
	if len(fallback) == 0 {
		fallback = []string{oauthserver.ScopeRead}
	}

	if len(oauthserver.ObjectScopes(requested)) > 0 || resource == "" {
		return fallback
	}

	advertised, err := h.OAuthClients.ResourceScopes(ctx, resource)
	if err != nil {
		logx.FromContext(ctx).Debug().Err(err).Str("resource", resource).Msg("oauth: no resource scopes to suggest")

		return fallback
	}

	var defaults []string

	for _, s := range advertised {
		if _, err := oauthserver.ValidatePersonalGrant(requested, []string{s}); err == nil {
			defaults = append(defaults, s)
		}
	}

	if len(oauthserver.ObjectScopes(defaults)) == 0 {
		return fallback
	}

	return defaults
}

// OAuthConsentHandler records the user's decision, installing the application in the session's organization on approval
// and returns the client redirect for the console to follow
func (h *Handler) OAuthConsentHandler(ctx echo.Context) error {
	in, err := BindAndValidate[models.OAuthConsentRequest](ctx)
	if err != nil {
		return h.InvalidInput(ctx, err)
	}

	reqCtx := ctx.Request().Context()

	caller, ok := auth.CallerFromContext(reqCtx)
	if !ok {
		return h.Unauthorized(ctx, auth.ErrNoAuthUser)
	}

	// only a person in the console can approve an application, not a token or an impersonating support user
	if caller.AuthenticationType != auth.JWTAuthentication {
		return h.Forbidden(ctx, ErrOAuthUserRequired)
	}

	if _, impersonated := auth.IsImpersonatedCallerContext(reqCtx); impersonated {
		return h.Forbidden(ctx, ErrOAuthImpersonation)
	}

	// a session belongs to one organization, so this can't target an organization the user hasn't signed in to
	orgID := caller.OrganizationID
	if orgID == "" {
		return h.BadRequest(ctx, ErrOAuthNoOrganization)
	}

	req, err := h.pendingOAuthRequest(reqCtx, in.RequestID)
	if err != nil {
		return h.NotFound(ctx, err)
	}

	if !iamoauth.ValidRedirect(req.RedirectURI) {
		return h.BadRequest(ctx, ErrOAuthBadRedirect)
	}

	internalCtx := auth.WithInternalOperationContext(reqCtx)

	if !in.Approve {
		if consumed, err := store.Consume(internalCtx, h.DBClient, req.ID); err != nil || !consumed {
			return h.NotFound(ctx, ErrOAuthRequestNotFound)
		}

		return h.Success(ctx, models.OAuthConsentReply{RedirectURL: h.oauthRedirect(req.RedirectURI, url.Values{
			iamoauth.ParamError: {iamoauth.ErrCodeAccessDenied},
			iamoauth.ParamState: {req.State},
		})})
	}

	install, err := validateOAuthInstall(req.Scopes, in)
	if err != nil {
		return h.BadRequest(ctx, err)
	}

	if err := h.checkOAuthInstallAccess(reqCtx, caller, install.kind); err != nil {
		return h.Forbidden(ctx, err)
	}

	client, err := h.OAuthClients.Resolve(reqCtx, req.ClientID)
	if err != nil {
		return h.BadRequest(ctx, err)
	}

	// requests are single use, consumed only once the decision is known to be valid
	if consumed, err := store.Consume(internalCtx, h.DBClient, req.ID); err != nil || !consumed {
		return h.NotFound(ctx, ErrOAuthRequestNotFound)
	}

	grant, err := h.upsertOAuthGrant(internalCtx, orgID, caller.SubjectID, client, install)
	if err != nil {
		logx.FromContext(reqCtx).Error().Err(err).Msg("oauth consent: unable to install application")

		return h.InternalServerError(ctx, ErrProcessingRequest)
	}

	code, create, err := h.newOAuthToken(iamoauth.KindCode)
	if err != nil {
		return h.InternalServerError(ctx, ErrProcessingRequest)
	}

	if err := create.
		SetClientID(client.ID).
		SetRedirectURI(req.RedirectURI).
		SetCodeChallenge(req.CodeChallenge).
		SetResource(req.Resource).
		SetScopes(grant.Scopes).
		SetUserID(caller.SubjectID).
		SetOauthGrantID(grant.ID).
		Exec(internalCtx); err != nil {
		logx.FromContext(reqCtx).Error().Err(err).Msg("oauth consent: unable to store authorization code")

		return h.InternalServerError(ctx, ErrProcessingRequest)
	}

	return h.Success(ctx, models.OAuthConsentReply{RedirectURL: h.oauthRedirect(req.RedirectURI, url.Values{
		iamoauth.ParamCode:  {code},
		iamoauth.ParamState: {req.State},
	})})
}

// checkOAuthInstallAccess checks the user may install the kind of application and that the organization can pay for it
func (h *Handler) checkOAuthInstallAccess(ctx context.Context, caller *auth.Caller, kind enums.OAuthGrantKind) error {
	relation := fgax.CanView
	if kind == enums.OAuthGrantKindOrganization {
		relation = installOrgAppRelation
	}

	allowed, err := h.DBClient.Authz.CheckOrgAccess(ctx, fgax.AccessCheck{
		SubjectID:   caller.SubjectID,
		SubjectType: auth.UserSubjectType,
		ObjectID:    caller.OrganizationID,
		Relation:    relation,
	})
	if err != nil {
		logx.FromContext(ctx).Error().Err(err).Msg("oauth consent: access check failed")

		return ErrProcessingRequest
	}

	if !allowed {
		return ErrOAuthNoPermission
	}

	// applications are credentials like API tokens and personal access tokens, which require a payment method
	if err := rule.RequirePaymentMethod()(ctx, nil); err != nil && !errors.Is(err, privacy.Skip) {
		return err
	}

	return nil
}

// validateOAuthInstall checks the user's decision against what the application requested
func validateOAuthInstall(requested []string, in *models.OAuthConsentRequest) (*oauthInstall, error) {
	reqKind, decided, err := oauthserver.ClassifyScopes(requested)
	if err != nil {
		return nil, err
	}

	if in.Kind != enums.OAuthGrantKindUser && in.Kind != enums.OAuthGrantKindOrganization {
		return nil, ErrOAuthInvalidKind
	}

	if decided && in.Kind != reqKind {
		return nil, ErrOAuthKindMismatch
	}

	scopes := iamoauth.ParseScope(strings.Join(in.Scopes, " "))

	if in.Kind == enums.OAuthGrantKindOrganization {
		if err := oauthserver.ValidateOrganizationGrant(requested, scopes); err != nil {
			return nil, err
		}

		return &oauthInstall{kind: in.Kind, scopes: scopes}, nil
	}

	if len(scopes) == 0 && in.AccessLevel != "" {
		if in.AccessLevel != enums.OAuthAccessLevelRead && in.AccessLevel != enums.OAuthAccessLevelWrite {
			return nil, ErrOAuthInvalidAccess
		}

		scopes = oauthserver.AccessScopes(in.AccessLevel)
	}

	access, err := oauthserver.ValidatePersonalGrant(requested, scopes)
	if err != nil {
		return nil, err
	}

	return &oauthInstall{kind: in.Kind, access: access, scopes: scopes}, nil
}

// upsertOAuthGrant installs the application or updates its active install
// a revoked install is never re-enabled, it is soft deleted as history and replaced by a new grant
func (h *Handler) upsertOAuthGrant(ctx context.Context, orgID, userID string, client *iamoauth.Client, in *oauthInstall) (*ent.OauthGrant, error) {
	grantUserID := ""
	if in.kind == enums.OAuthGrantKindUser {
		grantUserID = userID
	}

	existing, err := h.DBClient.OauthGrant.Query().
		Where(
			oauthgrant.ClientID(client.ID),
			oauthgrant.OwnerID(orgID),
			oauthgrant.KindEQ(in.kind),
			oauthgrant.UserID(grantUserID),
		).
		Only(ctx)

	switch {
	case err == nil && existing.IsActive:
		upd := h.DBClient.OauthGrant.UpdateOne(existing).
			SetClientName(client.Name).
			SetClientURI(client.URI).
			SetLogoURI(client.LogoURI).
			SetScopes(in.scopes)

		if in.kind == enums.OAuthGrantKindUser {
			upd.SetAccessLevel(in.access)
		}

		return upd.Save(ctx)
	case err == nil:
		if err := h.DBClient.OauthGrant.DeleteOne(existing).Exec(ctx); err != nil {
			return nil, err
		}

		return h.createOAuthGrant(ctx, orgID, userID, client, in)
	case ent.IsNotFound(err):
		return h.createOAuthGrant(ctx, orgID, userID, client, in)
	default:
		return nil, err
	}
}

// createOAuthGrant installs the application as a new grant
func (h *Handler) createOAuthGrant(ctx context.Context, orgID, userID string, client *iamoauth.Client, in *oauthInstall) (*ent.OauthGrant, error) {
	create := h.DBClient.OauthGrant.Create().
		SetOwnerID(orgID).
		SetKind(in.kind).
		SetClientID(client.ID).
		SetClientName(client.Name).
		SetClientURI(client.URI).
		SetLogoURI(client.LogoURI).
		SetScopes(in.scopes).
		SetInstalledByUserID(userID)

	if in.kind == enums.OAuthGrantKindUser {
		create.SetUserID(userID).SetAccessLevel(in.access)
	}

	return create.Save(ctx)
}

// OAuthTokenHandler is the token endpoint that exchanges authorization codes and rotates refresh tokens for public PKCE clients
func (h *Handler) OAuthTokenHandler(ctx echo.Context) error {
	iamoauth.NoStore(ctx.Response().Header())

	form, oerr := iamoauth.ParseForm(ctx.Request())
	if oerr != nil {
		return oauthJSONError(ctx, oerr)
	}

	if err := iamoauth.SingleValued(form, iamoauth.TokenParams...); err != nil {
		return oauthJSONError(ctx, iamoauth.NewError(iamoauth.ErrCodeInvalidRequest, err.Error()))
	}

	internalCtx := auth.WithInternalOperationContext(ctx.Request().Context())

	var resp *iamoauth.TokenResponse

	switch form.Get(iamoauth.ParamGrantType) {
	case iamoauth.GrantTypeAuthorizationCode:
		resp, oerr = h.exchangeAuthorizationCode(internalCtx, form)
	case iamoauth.GrantTypeRefreshToken:
		resp, oerr = h.exchangeRefreshToken(internalCtx, form)
	default:
		oerr = iamoauth.NewError(iamoauth.ErrCodeUnsupportedGrantType, "grant_type must be authorization_code or refresh_token")
	}

	if oerr != nil {
		return oauthJSONError(ctx, oerr)
	}

	return ctx.JSON(http.StatusOK, resp)
}

// exchangeAuthorizationCode redeems an authorization code for an access and refresh token
func (h *Handler) exchangeAuthorizationCode(ctx context.Context, form url.Values) (*iamoauth.TokenResponse, *iamoauth.OAuthError) {
	const invalidCode = "authorization code is invalid, expired, or already used"

	code, oerr := h.redeemOAuthToken(ctx, form, form.Get(iamoauth.ParamCode), iamoauth.KindCode, invalidCode)
	if oerr != nil {
		return nil, oerr
	}

	if code.RedirectURI != form.Get(iamoauth.ParamRedirectURI) {
		return nil, rejectOAuthGrant(ctx, form, "redirect_uri does not match the code", invalidCode)
	}

	if !iamoauth.VerifyPKCE(form.Get(iamoauth.ParamCodeVerifier), code.CodeChallenge) {
		return nil, iamoauth.NewError(iamoauth.ErrCodeInvalidGrant, "code_verifier does not match code_challenge")
	}

	if consumed, err := store.Consume(ctx, h.DBClient, code.ID); err != nil || !consumed {
		return nil, rejectOAuthGrant(ctx, form, "code consumed concurrently", invalidCode)
	}

	return h.issueOAuthTokens(ctx, code)
}

// exchangeRefreshToken rotates a refresh token into a new access and refresh token
func (h *Handler) exchangeRefreshToken(ctx context.Context, form url.Values) (*iamoauth.TokenResponse, *iamoauth.OAuthError) {
	const invalidRefresh = "refresh token is invalid, expired, or revoked"

	rt, oerr := h.redeemOAuthToken(ctx, form, form.Get(iamoauth.ParamRefreshToken), iamoauth.KindRefresh, invalidRefresh)
	if oerr != nil {
		return nil, oerr
	}

	if consumed, err := store.Consume(ctx, h.DBClient, rt.ID); err != nil || !consumed {
		return nil, rejectOAuthGrant(ctx, form, "refresh token consumed concurrently", invalidRefresh)
	}

	return h.issueOAuthTokens(ctx, rt)
}

// redeemOAuthToken loads a code or refresh token presented at the token endpoint and checks it belongs to the client and resource
// a replayed code or reused refresh token means it leaked, so everything issued under its grant is revoked
func (h *Handler) redeemOAuthToken(ctx context.Context, form url.Values, raw string, kind iamoauth.TokenKind, invalid string) (*ent.OauthToken, *iamoauth.OAuthError) {
	// the caller is the client's backend, so there is no organization in context to read the org-owned grant
	token, err := store.Redeem(auth.WithInternalReadCrossOrgContext(ctx), h.DBClient, h.OAuthTokens, raw, kind)
	if errors.Is(err, store.ErrTokenConsumed) {
		h.revokeOAuthGrantTokens(ctx, token.OauthGrantID)
	}

	if err != nil {
		return nil, rejectOAuthGrant(ctx, form, err.Error(), invalid)
	}

	if token.ClientID != form.Get(iamoauth.ParamClientID) {
		return nil, rejectOAuthGrant(ctx, form, "client_id does not match the token", invalid)
	}

	if r := form.Get(iamoauth.ParamResource); r != "" && r != token.Resource {
		return nil, iamoauth.NewError(iamoauth.ErrCodeInvalidTarget, iamoauth.ErrInvalidTarget.Error())
	}

	return token, nil
}

// rejectOAuthGrant logs why a code or refresh token was rejected and returns the invalid_grant error the client sees
func rejectOAuthGrant(ctx context.Context, form url.Values, reason, description string) *iamoauth.OAuthError {
	logx.FromContext(ctx).Info().Str("reason", reason).Str("client_id", form.Get(iamoauth.ParamClientID)).Str("grant_type", form.Get(iamoauth.ParamGrantType)).Msg("oauth token: grant rejected")

	return iamoauth.NewError(iamoauth.ErrCodeInvalidGrant, description)
}

// issueOAuthTokens mints an access and refresh token pair under the grant of a redeemed token
// scopes come from the grant, so a narrower re-consent takes effect on refresh
func (h *Handler) issueOAuthTokens(ctx context.Context, redeemed *ent.OauthToken) (*iamoauth.TokenResponse, *iamoauth.OAuthError) {
	serverErr := iamoauth.NewError(iamoauth.ErrCodeServerError, "unable to issue tokens")
	grant := redeemed.Edges.OauthGrant

	access, accessCreate, err := h.newOAuthToken(iamoauth.KindAccess)
	if err != nil {
		return nil, serverErr
	}

	refresh, refreshCreate, err := h.newOAuthToken(iamoauth.KindRefresh)
	if err != nil {
		return nil, serverErr
	}

	builders := []*ent.OauthTokenCreate{accessCreate, refreshCreate}
	for _, b := range builders {
		b.SetClientID(redeemed.ClientID).SetResource(redeemed.Resource).SetScopes(grant.Scopes).SetUserID(redeemed.UserID).SetOauthGrantID(grant.ID)
	}

	if err := h.DBClient.OauthToken.CreateBulk(builders...).Exec(ctx); err != nil {
		logx.FromContext(ctx).Error().Err(err).Msg("oauth token: unable to store tokens")

		return nil, serverErr
	}

	return &iamoauth.TokenResponse{
		AccessToken:  access,
		TokenType:    iamoauth.TokenTypeBearer,
		ExpiresIn:    int64(iamoauth.AccessTokenTTL.Seconds()),
		RefreshToken: refresh,
		Scope:        iamoauth.FormatScope(grant.Scopes),
	}, nil
}

// newOAuthToken issues a token of kind and returns its value with a create builder holding its ID, hash, and expiry
func (h *Handler) newOAuthToken(kind iamoauth.TokenKind) (string, *ent.OauthTokenCreate, error) {
	tok, err := h.OAuthTokens.Issue(kind)
	if err != nil {
		return "", nil, err
	}

	return tok.Value, h.DBClient.OauthToken.Create().
		SetID(tok.ID.String()).
		SetKind(kind).
		SetTokenHash(tok.Hash).
		SetKeyVersion(tok.KeyVersion).
		SetExpiresAt(tok.ExpiresAt), nil
}

// OAuthRevokeHandler is the RFC 7009 revocation endpoint, it always returns 200 so callers can't probe for valid tokens
func (h *Handler) OAuthRevokeHandler(ctx echo.Context) error {
	reqCtx := ctx.Request().Context()

	form, oerr := iamoauth.ParseForm(ctx.Request())
	if oerr != nil {
		return oauthJSONError(ctx, oerr)
	}

	raw := form.Get(iamoauth.ParamToken)

	kind := iamoauth.KindOf(raw)
	if kind != iamoauth.KindAccess && kind != iamoauth.KindRefresh {
		return ctx.String(http.StatusOK, "")
	}

	// a client may only revoke its own tokens
	token, err := store.Load(auth.WithInternalReadCrossOrgContext(reqCtx), h.DBClient.OauthToken.Query(), h.OAuthTokens, raw, kind)
	if err != nil || token.ClientID != form.Get(iamoauth.ParamClientID) {
		return ctx.String(http.StatusOK, "")
	}

	if _, err := store.Consume(auth.WithInternalOperationContext(reqCtx), h.DBClient, token.ID); err != nil {
		logx.FromContext(reqCtx).Error().Err(err).Msg("oauth revoke: unable to revoke token")
	}

	return ctx.String(http.StatusOK, "")
}

// pendingOAuthRequest loads an unconsumed, unexpired authorization request
func (h *Handler) pendingOAuthRequest(ctx context.Context, raw string) (*ent.OauthToken, error) {
	query := h.DBClient.OauthToken.Query().Where(oauthtoken.ConsumedAtIsNil(), oauthtoken.ExpiresAtGT(time.Now()))

	req, err := store.Load(auth.WithInternalReadContext(ctx), query, h.OAuthTokens, raw, iamoauth.KindRequest)
	if err != nil {
		return nil, ErrOAuthRequestNotFound
	}

	return req, nil
}

// revokeOAuthGrantTokens consumes every outstanding token under a grant
func (h *Handler) revokeOAuthGrantTokens(ctx context.Context, grantID string) {
	if grantID == "" {
		return
	}

	if err := store.RevokeGrantTokens(ctx, h.DBClient, grantID); err != nil {
		logx.FromContext(ctx).Error().Err(err).Str("grant_id", grantID).Msg("oauth: unable to revoke grant tokens")
	}
}

// oauthRedirect adds params and the issuer to a validated client redirect URI
func (h *Handler) oauthRedirect(redirectURI string, params url.Values) string {
	redirect, err := iamoauth.RedirectURL(redirectURI, h.oauthIssuer(), params)
	if err != nil {
		return strings.TrimRight(h.ConsoleURL, "/")
	}

	return redirect
}

// oauthJSONError writes an RFC 6749 error response with its status
func oauthJSONError(ctx echo.Context, e *iamoauth.OAuthError) error {
	return ctx.JSON(e.Status(), e)
}
