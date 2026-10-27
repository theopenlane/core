package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	echo "github.com/theopenlane/echox"

	"github.com/go-webauthn/webauthn/protocol"
	"github.com/go-webauthn/webauthn/webauthn"

	"github.com/theopenlane/utils/rout"

	"github.com/theopenlane/iam/auth"
	provider "github.com/theopenlane/iam/providers/webauthn"
	"github.com/theopenlane/iam/sessions"

	"github.com/theopenlane/core/common/enums"
	models "github.com/theopenlane/core/common/openapi"
	"github.com/theopenlane/core/v2/internal/ent/privacy/token"
	"github.com/theopenlane/core/v2/pkg/metrics"
	sso "github.com/theopenlane/core/v2/pkg/ssoutils"
	"github.com/theopenlane/logx"
)

const (
	webauthnRegistration = "WEBAUTHN_REGISTRATION"
	webauthnLogin        = "WEBAUTHN_LOGIN"
)

// BeginWebauthnRegistration is the request to begin adding a passkey to the authenticated user's account
func (h *Handler) BeginWebauthnRegistration(ctx echo.Context) error {
	r, err := BindAndValidate[models.WebauthnRegistrationRequest](ctx)
	if err != nil {
		return h.InvalidInput(ctx, err)
	}

	reqCtx := ctx.Request().Context()

	caller, ok := passkeyRegistrationCaller(reqCtx)
	if !ok {
		return h.Unauthorized(ctx, ErrAuthenticationRequired)
	}

	entUser, _, err := h.getUserByID(reqCtx, caller.SubjectID)
	if err != nil {
		logx.FromContext(reqCtx).Error().Err(err).Msg("unable to get user by id")

		return h.InternalServerError(ctx, ErrProcessingRequest)
	}

	if !strings.EqualFold(r.Email, entUser.Email) {
		return h.BadRequest(ctx, ErrPasskeyEmailMismatch)
	}

	var credentials = make([]webauthn.Credential, 0, len(entUser.Edges.Webauthns))

	for _, credential := range entUser.Edges.Webauthns {
		credentials = append(credentials, webauthn.Credential{
			ID:        credential.CredentialID,
			PublicKey: credential.PublicKey,
		})
	}

	user := &provider.User{
		ID:    entUser.ID,
		Email: entUser.Email,
		Name:  entUser.FirstName + " " + entUser.LastName,
		// set excluded passkeys
		WebauthnCredentials: credentials,
	}

	// options is the object that needs to be returned for the front end to open the creation dialog for the user to create the passkey
	options, session, err := h.WebAuthn.BeginRegistration(user,
		webauthn.WithRegistrationRelyingPartyName("Openlane"),
		webauthn.WithResidentKeyRequirement(protocol.ResidentKeyRequirementRequired),
		webauthn.WithExclusions(user.CredentialExcludeList()),
	)
	if err != nil {
		return err
	}

	// we have to set not just a regular session for the user but also capture the return of the webauthn session
	setSessionMap := map[string]any{}
	setSessionMap[sessions.WebAuthnKey] = session
	setSessionMap[sessions.UsernameKey] = user.Name
	setSessionMap[sessions.UserTypeKey] = webauthnRegistration
	setSessionMap[sessions.EmailKey] = user.Email
	setSessionMap[sessions.UserIDKey] = user.ID

	sessionCtx, err := h.SessionConfig.SaveAndStoreSession(reqCtx, ctx.Response().Writer, setSessionMap, user.ID)
	if err != nil {
		logx.FromContext(reqCtx).Error().Err(err).Msg("unable to save session")

		return h.InternalServerError(ctx, ErrProcessingRequest)
	}

	// return the session value for the UI to use
	// the UI will need to set the cookie because authentication is handled
	// server side
	s, err := sessions.SessionToken(sessionCtx)
	if err != nil {
		logx.FromContext(reqCtx).Error().Err(err).Msg("unable to get session token")

		return h.InternalServerError(ctx, ErrProcessingRequest)
	}

	out := &models.WebauthnBeginRegistrationResponse{
		Reply:              rout.Reply{Success: true},
		CredentialCreation: options,
		Session:            s,
	}

	return h.Success(ctx, out)
}

// FinishWebauthnRegistration is the request to finish a webauthn registration - this is where we get the credential created by the user back
func (h *Handler) FinishWebauthnRegistration(ctx echo.Context) error {
	requestData, err := BindAndValidate[models.WebauthnRegistrationFinishRequest](ctx)
	if err != nil {
		return h.InvalidInput(ctx, err)
	}

	// lookup userID in cache to ensure cookie and tokens match
	session, err := h.SessionConfig.SessionManager.Get(ctx.Request(), h.SessionConfig.CookieConfig.Name)
	if err != nil {
		return h.BadRequest(ctx, err)
	}

	reqCtx := ctx.Request().Context()

	// Get sessionID from cookie and check against redis
	sessionID := h.SessionConfig.SessionManager.GetSessionIDFromCookie(session)

	userID, err := h.SessionConfig.RedisStore.GetSession(reqCtx, sessionID)
	if err != nil {
		return h.BadRequest(ctx, err)
	}

	// get session data from cookie to get the user id stored
	sessionData := h.SessionConfig.SessionManager.GetSessionDataFromCookie(session)

	userIDFromCookie := sessionData.(map[string]any)[sessions.UserIDKey]

	// ensure the user is the same as the one who started the registration
	if userIDFromCookie != userID {
		return h.BadRequest(ctx, ErrInvalidCredentials)
	}

	caller, ok := passkeyRegistrationCaller(reqCtx)
	if !ok || caller.SubjectID != userID {
		return h.Unauthorized(ctx, ErrAuthenticationRequired)
	}

	// get user from the database
	entUser, userCtx, err := h.getUserByID(reqCtx, userID)
	if err != nil {
		logx.FromContext(reqCtx).Error().Err(err).Msg("unable to get user by id")

		return h.InternalServerError(ctx, ErrProcessingRequest)
	}

	data, err := json.Marshal(requestData)
	if err != nil {
		return h.BadRequest(ctx, err)
	}

	response, err := protocol.ParseCredentialCreationResponseBytes(data)
	if err != nil {
		return h.BadRequest(ctx, err)
	}

	userCtx = token.NewContextWithWebauthnCreationContextKey(userCtx)

	// get webauthn session data from the session
	webauthnData := sessionData.(map[string]any)[sessions.WebAuthnKey]

	wd, ok := webauthnData.(webauthn.SessionData)
	if !ok {
		return h.BadRequest(ctx, ErrLoginFailed)
	}

	user := &provider.User{
		ID:    entUser.ID,
		Email: entUser.Email,
		Name:  entUser.FirstName + " " + entUser.LastName,
	}

	// validate the credential
	credential, err := h.WebAuthn.CreateCredential(user, wd, response)
	if err != nil {
		return h.BadRequest(ctx, err)
	}

	// save the credential to the database
	if err := h.addCredentialToUser(userCtx, entUser, *credential); err != nil {
		if IsConstraintError(err) {
			return h.BadRequestWithCode(ctx, ErrDeviceAlreadyRegistered, DeviceRegisteredErrCode)
		}

		logx.FromContext(reqCtx).Error().Err(err).Msg("unable to add credential to user")

		return h.InternalServerError(ctx, ErrProcessingRequest)
	}

	if err := h.setWebauthnAllowed(userCtx, entUser); err != nil {
		logx.FromContext(reqCtx).Error().Err(err).Msg("unable to set webauthn allowed for user")

		return h.InternalServerError(ctx, ErrProcessingRequest)
	}

	out := &models.WebauthnRegistrationResponse{
		Reply:   rout.Reply{Success: true},
		Message: "passkey successfully created",
	}

	return h.Success(ctx, out)
}

// passkeyRegistrationCaller returns the caller when it is a user signed in with their own session
func passkeyRegistrationCaller(ctx context.Context) (*auth.Caller, bool) {
	caller, ok := auth.CallerFromContext(ctx)
	if !ok || caller.AuthenticationType != auth.JWTAuthentication || caller.IsAnonymous() || caller.IsImpersonated() {
		return nil, false
	}

	return caller, true
}

// BeginWebauthnLogin is the request to begin a webauthn login
func (h *Handler) BeginWebauthnLogin(ctx echo.Context) error {
	r, err := BindAndValidate[models.WebauthnLoginRequest](ctx)
	if err != nil {
		metrics.RecordLogin(false)
		return h.InvalidInput(ctx, err)
	}

	credential, session, err := h.WebAuthn.BeginDiscoverableLogin()
	if err != nil {
		metrics.RecordLogin(false)
		return err
	}

	reqCtx := ctx.Request().Context()

	setSessionMap := map[string]any{}
	setSessionMap[sessions.WebAuthnKey] = session
	setSessionMap[sessions.UserTypeKey] = webauthnLogin
	setSessionMap[sessions.EmailKey] = r.Email

	// set the user id to the challenge so we can verify it later
	// this allows to ensure the start and finish is the same user without
	// having the email at the start
	setSessionMap[sessions.UserIDKey] = credential.Response.Challenge.String()

	sessionCtx, err := h.SessionConfig.SaveAndStoreSession(reqCtx, ctx.Response().Writer, setSessionMap, credential.Response.Challenge.String())
	if err != nil {
		logx.FromContext(reqCtx).Error().Err(err).Msg("unable to save session")
		metrics.RecordLogin(false)

		return h.InternalServerError(ctx, ErrProcessingRequest)
	}

	// return the session value for the UI to use
	// the UI will need to set the cookie because authentication is handled
	// server side
	s, err := sessions.SessionToken(sessionCtx)
	if err != nil {
		logx.FromContext(reqCtx).Error().Err(err).Msg("unable to get session token")
		metrics.RecordLogin(false)

		return h.InternalServerError(ctx, ErrProcessingRequest)
	}

	out := &models.WebauthnBeginLoginResponse{
		Reply:               rout.Reply{Success: true},
		CredentialAssertion: credential,
		Session:             s,
	}

	return h.Success(ctx, out)
}

// FinishWebauthnLogin is the request to finish a webauthn login
func (h *Handler) FinishWebauthnLogin(ctx echo.Context) error {
	requestData, err := BindAndValidate[models.WebauthnLoginFinishRequest](ctx)
	if err != nil {
		metrics.RecordLogin(false)
		return h.InvalidInput(ctx, err)
	}

	reqCtx := ctx.Request().Context()

	session, err := h.SessionConfig.SessionManager.Get(ctx.Request(), h.SessionConfig.CookieConfig.Name)
	if err != nil {
		logx.FromContext(reqCtx).Error().Err(err).Msg("unable to get session from cookie")
		metrics.RecordLogin(false)

		return h.BadRequest(ctx, ErrInvalidCredentials)
	}

	sessionData := h.SessionConfig.SessionManager.GetSessionDataFromCookie(session)
	webauthnData := sessionData.(map[string]any)[sessions.WebAuthnKey]

	wd, ok := webauthnData.(webauthn.SessionData)
	if !ok {
		metrics.RecordLogin(false)

		return h.BadRequest(ctx, ErrLoginFailed)
	}

	data, err := json.Marshal(requestData)
	if err != nil {
		metrics.RecordLogin(false)

		return h.BadRequest(ctx, err)
	}

	response, err := protocol.ParseCredentialRequestResponseBytes(data)
	if err != nil {
		logx.FromContext(reqCtx).Error().Err(err).Msg("unable to parse credential request response body")
		metrics.RecordLogin(false)

		return h.BadRequest(ctx, ErrInvalidCredentials)
	}

	if _, err = h.WebAuthn.ValidateDiscoverableLogin(h.userHandler(reqCtx), wd, response); err != nil {
		logx.FromContext(reqCtx).Error().Err(err).Msg("unable to validate webauthn login")
		metrics.RecordLogin(false)

		return h.BadRequest(ctx, ErrInvalidCredentials)
	}

	userID := string(response.Response.UserHandle)

	userIDFromCookie := sessionData.(map[string]any)[sessions.UserIDKey]

	// ensure the user is the same as the one who started the login based on the challenge
	if userIDFromCookie != response.Response.CollectedClientData.Challenge {
		logx.FromContext(reqCtx).Error().Err(err).Msg("challenge ids do not match")
		metrics.RecordLogin(false)

		return h.BadRequest(ctx, ErrInvalidCredentials)
	}

	// get user from the database
	entUser, reqCtx, err := h.getUserByID(reqCtx, userID)
	if err != nil {
		logx.FromContext(reqCtx).Error().Err(err).Msg("unable to get user by id")
		metrics.RecordLogin(false)

		return h.InternalServerError(ctx, ErrProcessingRequest)
	}

	if !entUser.Edges.Setting.EmailConfirmed {
		metrics.RecordLogin(false)
		logx.FromContext(reqCtx).Info().Str("user_id", entUser.ID).Msg("user email not verified, unable to login")

		return h.BadRequest(ctx, fmt.Errorf("%w: please check your email and verify your account", auth.ErrUnverifiedUser))
	}

	orgStatus := h.orgEnforcementsForUser(reqCtx, entUser.Email)
	if orgStatus != nil && orgStatus.Enforced {
		if !h.HasValidSSOSession(ctx, entUser.ID) {
			return ctx.Redirect(http.StatusFound, sso.SSOLogin(ctx.Echo(), orgStatus.OrganizationID))
		}
	}

	// create claims for verified user
	authData, err := h.AuthManager.GenerateUserAuthSession(reqCtx, ctx.Response().Writer, entUser)
	if err != nil {
		logx.FromContext(reqCtx).Error().Err(err).Msg("unable to create new auth session")
		metrics.RecordLogin(false)

		return h.InternalServerError(ctx, ErrProcessingRequest)
	}

	// set the last seen for the user
	if err := h.updateUserLastSeen(reqCtx, userID, enums.AuthProviderCredentials); err != nil {
		logx.FromContext(reqCtx).Error().Err(err).Msg("unable to update last seen")
		metrics.RecordLogin(false)

		return h.InternalServerError(ctx, ErrProcessingRequest)
	}

	out := &models.WebauthnLoginResponse{
		Reply:    rout.Reply{Success: true},
		Message:  "passkey successfully created",
		AuthData: *authData,
	}

	metrics.RecordLogin(true)

	return h.Success(ctx, out)
}

// userHandler returns a webauthn.DiscoverableUserHandler that can be used to look up a user by their userHandle
func (h *Handler) userHandler(ctx context.Context) webauthn.DiscoverableUserHandler {
	return func(_, userHandle []byte) (user webauthn.User, err error) {
		u, _, err := h.getUserByID(ctx, string(userHandle))
		if err != nil {
			return nil, err
		}

		authnUser := &provider.User{
			ID:                  u.ID,
			Email:               u.Email,
			Name:                u.FirstName + " " + u.LastName,
			WebauthnCredentials: []webauthn.Credential{},
		}

		for _, cred := range u.Edges.Webauthns {
			authnCred := webauthn.Credential{
				ID:              cred.CredentialID,
				PublicKey:       cred.PublicKey,
				AttestationType: cred.AttestationType,
				Flags: webauthn.CredentialFlags{
					BackupEligible: cred.BackupEligible,
					BackupState:    cred.BackupState,
					UserPresent:    cred.UserPresent,
					UserVerified:   cred.UserVerified,
				},
			}

			for _, t := range cred.Transports {
				authnCred.Transport = append(authnCred.Transport, protocol.AuthenticatorTransport(t))
			}

			authnUser.WebauthnCredentials = append(authnUser.WebauthnCredentials, authnCred)
		}

		return authnUser, nil
	}
}

func (h *Handler) HasValidSSOSession(ctx echo.Context, userID string) bool {
	sess, err := h.SessionConfig.SessionManager.Get(ctx.Request(), h.SessionConfig.CookieConfig.Name)
	if err != nil {
		return false
	}

	sessionID := h.SessionConfig.SessionManager.GetSessionIDFromCookie(sess)

	storedUser, err := h.SessionConfig.RedisStore.GetSession(ctx.Request().Context(), sessionID)
	if err != nil || storedUser != userID {
		return false
	}

	data := h.SessionConfig.SessionManager.GetSessionDataFromCookie(sess)

	m, ok := data.(map[string]any)
	if !ok {
		return false
	}

	userType, ok := m[sessions.UserTypeKey].(string)

	return ok && userType == enums.AuthProviderOIDC.String()
}
