package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"

	echo "github.com/theopenlane/echox"

	"github.com/go-webauthn/webauthn/protocol"
	"github.com/go-webauthn/webauthn/webauthn"

	"github.com/theopenlane/utils/rout"

	"github.com/theopenlane/iam/auth"
	provider "github.com/theopenlane/iam/providers/webauthn"
	"github.com/theopenlane/iam/sessions"

	"github.com/theopenlane/core/internal/ent/privacy/token"
	entval "github.com/theopenlane/core/internal/ent/validator"
	"github.com/theopenlane/core/pkg/enums"
	"github.com/theopenlane/core/pkg/logx"
	"github.com/theopenlane/core/pkg/metrics"
	models "github.com/theopenlane/core/pkg/openapi"
	sso "github.com/theopenlane/core/pkg/ssoutils"
)

const (
	webauthnRegistration = "WEBAUTHN_REGISTRATION"
	webauthnLogin        = "WEBAUTHN_LOGIN"
)

// BeginWebauthnRegistration is the request to begin a webauthn login
func (h *Handler) BeginWebauthnRegistration(ctx echo.Context, openapi *OpenAPIContext) error {
	r, err := BindAndValidateWithAutoRegistry(ctx, h, openapi.Operation, models.ExampleWebauthnBeginRegistrationRequest, models.ExampleWebauthnBeginRegistrationResponse, openapi.Registry)
	if err != nil {
		return h.InvalidInput(ctx, err, openapi)
	}

	if isRegistrationContext(ctx) {
		return nil
	}

	reqCtx := ctx.Request().Context()

	ctxWithToken := token.NewContextWithOauthTooToken(reqCtx, r.Email)

	// to register a new passkey, the user needs to be created + logged in first
	// once the the passkey is added to the user's account, they can use it to login
	// we treat this verify similar to the oauth or basic registration flow
	// user is created first, no credential method is set / they are unable to login until the credential flow is finished
	entUser, err := h.CheckAndCreateUser(ctxWithToken, r.Name, r.Email, enums.AuthProviderCredentials, "")
	if err != nil {
		if errors.Is(err, entval.ErrEmailNotAllowed) {
			logx.FromContext(reqCtx).Error().Err(err).Str("email", r.Email).Msg("email not allowed")

			return h.InvalidInput(ctx, err, openapi)
		}

		return h.InternalServerError(ctx, err, openapi)
	}

	// set context for remaining request based on logged in user
	userCtx := setAuthenticatedContext(ctxWithToken, entUser)

	webAuthns, err := entUser.QueryWebauthns().All(userCtx)
	if err != nil {
		return h.InternalServerError(ctx, err, openapi)
	}

	var credentials = make([]webauthn.Credential, 0, len(webAuthns))

	for _, credential := range webAuthns {
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
	setSessionMap[sessions.UsernameKey] = r.Name
	setSessionMap[sessions.UserTypeKey] = webauthnRegistration
	setSessionMap[sessions.EmailKey] = r.Email
	setSessionMap[sessions.UserIDKey] = user.ID

	sessionCtx, err := h.SessionConfig.SaveAndStoreSession(userCtx, ctx.Response().Writer, setSessionMap, user.ID)
	if err != nil {
		return h.InternalServerError(ctx, err, openapi)
	}

	// return the session value for the UI to use
	// the UI will need to set the cookie because authentication is handled
	// server side
	s, err := sessions.SessionToken(sessionCtx)
	if err != nil {
		return h.InternalServerError(ctx, err, openapi)
	}

	out := &models.WebauthnBeginRegistrationResponse{
		Reply:              rout.Reply{Success: true},
		CredentialCreation: options,
		Session:            s,
	}

	return h.Success(ctx, out, openapi)
}

// FinishWebauthnRegistration is the request to finish a webauthn registration - this is where we get the credential created by the user back
func (h *Handler) FinishWebauthnRegistration(ctx echo.Context, openapi *OpenAPIContext) error {
	requestData, err := BindAndValidateWithAutoRegistry(ctx, h, openapi.Operation, models.ExampleWebauthnRegistrationFinishRequest, models.ExampleWebauthnRegistrationResponse, openapi.Registry)
	if err != nil {
		return h.InvalidInput(ctx, err, openapi)
	}

	// lookup userID in cache to ensure cookie and tokens match
	session, err := h.SessionConfig.SessionManager.Get(ctx.Request(), h.SessionConfig.CookieConfig.Name)
	if err != nil {
		return h.BadRequest(ctx, err, openapi)
	}

	reqCtx := ctx.Request().Context()

	// Get sessionID from cookie and check against redis
	sessionID := h.SessionConfig.SessionManager.GetSessionIDFromCookie(session)

	userID, err := h.SessionConfig.RedisStore.GetSession(reqCtx, sessionID)
	if err != nil {
		return h.BadRequest(ctx, err, openapi)
	}

	// get session data from cookie to get the user id stored
	sessionData := h.SessionConfig.SessionManager.GetSessionDataFromCookie(session)

	userIDFromCookie := sessionData.(map[string]any)[sessions.UserIDKey]

	// ensure the user is the same as the one who started the registration
	if userIDFromCookie != userID {
		return h.BadRequest(ctx, err, openapi)
	}

	// get user from the database
	entUser, userCtx, err := h.getUserByID(reqCtx, userID)
	if err != nil {
		return h.InternalServerError(ctx, err, openapi)
	}

	data, err := json.Marshal(requestData)
	if err != nil {
		return h.BadRequest(ctx, err, openapi)
	}

	response, err := protocol.ParseCredentialCreationResponseBytes(data)
	if err != nil {
		return h.BadRequest(ctx, err, openapi)
	}

	userCtx = token.NewContextWithWebauthnCreationContextKey(userCtx)

	// get webauthn session data from the session
	webauthnData := sessionData.(map[string]any)[sessions.WebAuthnKey]

	wd, ok := webauthnData.(webauthn.SessionData)
	if !ok {
		return h.BadRequest(ctx, auth.ErrNoAuthUser, openapi)
	}

	user := &provider.User{
		ID:    entUser.ID,
		Email: entUser.Email,
		Name:  entUser.FirstName + " " + entUser.LastName,
	}

	// validate the credential
	credential, err := h.WebAuthn.CreateCredential(user, wd, response)
	if err != nil {
		return h.BadRequest(ctx, err, openapi)
	}

	// save the credential to the database
	if err := h.addCredentialToUser(userCtx, entUser, *credential); err != nil {
		if IsConstraintError(err) {
			return h.BadRequestWithCode(ctx, ErrDeviceAlreadyRegistered, DeviceRegisteredErrCode)
		}

		return h.InternalServerError(ctx, err, openapi)
	}

	if err := h.setWebauthnAllowed(userCtx, entUser); err != nil {
		return h.InternalServerError(ctx, err, openapi)
	}

	// create new claims for the user
	auth, err := h.AuthManager.GenerateUserAuthSession(userCtx, ctx.Response().Writer, entUser)
	if err != nil {
		logx.FromContext(reqCtx).Error().Err(err).Msg("unable to create new auth session")

		return h.InternalServerError(ctx, err, openapi)
	}

	out := &models.WebauthnRegistrationResponse{
		Reply:    rout.Reply{Success: true},
		Message:  "passkey successfully created",
		AuthData: *auth,
	}

	return h.Success(ctx, out, openapi)
}

// BeginWebauthnLogin is the request to begin a webauthn login
func (h *Handler) BeginWebauthnLogin(ctx echo.Context, openapi *OpenAPIContext) error {
	r, err := BindAndValidateWithAutoRegistry(ctx, h, openapi.Operation, models.ExampleWebauthnLoginRequest, models.ExampleWebauthnBeginLoginResponse, openapi.Registry)
	if err != nil {
		metrics.RecordLogin(false)
		return h.InvalidInput(ctx, err, openapi)
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

		return h.InternalServerError(ctx, ErrProcessingRequest, openapi)
	}

	// return the session value for the UI to use
	// the UI will need to set the cookie because authentication is handled
	// server side
	s, err := sessions.SessionToken(sessionCtx)
	if err != nil {
		logx.FromContext(reqCtx).Error().Err(err).Msg("unable to get session token")
		metrics.RecordLogin(false)

		return h.InternalServerError(ctx, ErrProcessingRequest, openapi)
	}

	out := &models.WebauthnBeginLoginResponse{
		Reply:               rout.Reply{Success: true},
		CredentialAssertion: credential,
		Session:             s,
	}

	return h.Success(ctx, out, openapi)
}

// FinishWebauthnLogin is the request to finish a webauthn login
func (h *Handler) FinishWebauthnLogin(ctx echo.Context, openapi *OpenAPIContext) error {
	requestData, err := BindAndValidateWithAutoRegistry(ctx, h, openapi.Operation, models.ExampleWebauthnLoginFinishRequest, models.ExampleWebauthnLoginResponse, openapi.Registry)
	if err != nil {
		metrics.RecordLogin(false)
		return h.InvalidInput(ctx, err, openapi)
	}

	if isRegistrationContext(ctx) {
		return nil
	}

	reqCtx := ctx.Request().Context()

	session, err := h.SessionConfig.SessionManager.Get(ctx.Request(), h.SessionConfig.CookieConfig.Name)
	if err != nil {
		logx.FromContext(reqCtx).Error().Err(err).Msg("unable to get session from cookie")
		metrics.RecordLogin(false)

		return h.BadRequest(ctx, ErrInvalidCredentials, openapi)
	}

	sessionData := h.SessionConfig.SessionManager.GetSessionDataFromCookie(session)
	webauthnData := sessionData.(map[string]any)[sessions.WebAuthnKey]

	wd, ok := webauthnData.(webauthn.SessionData)
	if !ok {
		metrics.RecordLogin(false)

		return h.BadRequest(ctx, auth.ErrNoAuthUser, openapi)
	}

	data, err := json.Marshal(requestData)
	if err != nil {
		metrics.RecordLogin(false)

		return h.BadRequest(ctx, err, openapi)
	}

	response, err := protocol.ParseCredentialRequestResponseBytes(data)
	if err != nil {
		logx.FromContext(reqCtx).Error().Err(err).Msg("unable to parse credential request response body")
		metrics.RecordLogin(false)

		return h.BadRequest(ctx, ErrInvalidCredentials, openapi)
	}

	if _, err = h.WebAuthn.ValidateDiscoverableLogin(h.userHandler(reqCtx), wd, response); err != nil {
		logx.FromContext(reqCtx).Error().Err(err).Msg("unable to validate webauthn login")
		metrics.RecordLogin(false)

		return h.BadRequest(ctx, ErrInvalidCredentials, openapi)
	}

	userID := string(response.Response.UserHandle)

	userIDFromCookie := sessionData.(map[string]any)[sessions.UserIDKey]

	// ensure the user is the same as the one who started the login based on the challenge
	if userIDFromCookie != response.Response.CollectedClientData.Challenge {
		logx.FromContext(reqCtx).Error().Err(err).Msg("challenge ids do not match")
		metrics.RecordLogin(false)

		return h.BadRequest(ctx, ErrInvalidCredentials, openapi)
	}

	// get user from the database
	entUser, reqCtx, err := h.getUserByID(reqCtx, userID)
	if err != nil {
		logx.FromContext(reqCtx).Error().Err(err).Msg("unable to get user by id")
		metrics.RecordLogin(false)

		return h.InternalServerError(ctx, ErrProcessingRequest, openapi)
	}

	orgStatus := h.orgEnforcementsForUser(reqCtx, entUser.Email)
	if orgStatus != nil && orgStatus.Enforced {
		if !h.HasValidSSOSession(ctx, entUser.ID) {
			return ctx.Redirect(http.StatusFound, sso.SSOLogin(ctx.Echo(), orgStatus.OrganizationID))
		}
	}

	// create claims for verified user
	auth, err := h.AuthManager.GenerateUserAuthSession(reqCtx, ctx.Response().Writer, entUser)
	if err != nil {
		logx.FromContext(reqCtx).Error().Err(err).Msg("unable to create new auth session")
		metrics.RecordLogin(false)

		return h.InternalServerError(ctx, ErrProcessingRequest, openapi)
	}

	// set the last seen for the user
	if err := h.updateUserLastSeen(reqCtx, userID, enums.AuthProviderCredentials); err != nil {
		logx.FromContext(reqCtx).Error().Err(err).Msg("unable to update last seen")
		metrics.RecordLogin(false)

		return h.InternalServerError(ctx, ErrProcessingRequest, openapi)
	}

	out := &models.WebauthnLoginResponse{
		Reply:    rout.Reply{Success: true},
		Message:  "passkey successfully created",
		AuthData: *auth,
	}

	metrics.RecordLogin(true)

	return h.Success(ctx, out, openapi)
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
