package handlers

import (
	"context"
	"net/http"

	"github.com/rs/zerolog"
	echo "github.com/theopenlane/echox"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/go-webauthn/webauthn/protocol"
	"github.com/go-webauthn/webauthn/webauthn"

	"github.com/theopenlane/utils/rout"

	provider "github.com/theopenlane/iam/providers/webauthn"
	"github.com/theopenlane/iam/sessions"

	"github.com/theopenlane/core/internal/ent/privacy/token"
	"github.com/theopenlane/core/pkg/enums"
	"github.com/theopenlane/core/pkg/models"
	sso "github.com/theopenlane/core/pkg/ssoutils"
)

const (
	webauthnProvider     = "WEBAUTHN"
	webauthnRegistration = "WEBAUTHN_REGISTRATION"
	webauthnLogin        = "WEBAUTHN_LOGIN"
)

// BeginWebauthnRegistration is the request to begin a webauthn login
func (h *Handler) BeginWebauthnRegistration(ctx echo.Context) error {
	var r models.WebauthnRegistrationRequest
	if err := ctx.Bind(&r); err != nil {
		return h.InvalidInput(ctx, err)
	}

	reqCtx := ctx.Request().Context()

	ctxWithToken := token.NewContextWithOauthTooToken(reqCtx, r.Email)

	// to register a new passkey, the user needs to be created + logged in first
	// once the the passkey is added to the user's account, they can use it to login
	// we treat this verify similar to the oauth or basic registration flow
	// user is created first, no credential method is set / they are unable to login until the credential flow is finished
	entUser, err := h.CheckAndCreateUser(ctxWithToken, r.Name, r.Email, enums.AuthProviderCredentials, "")
	if err != nil {
		return h.InternalServerError(ctx, err)
	}

	// set context for remaining request based on logged in user
	userCtx := setAuthenticatedContext(ctxWithToken, entUser)

	webAuthns, err := entUser.QueryWebauthns().All(userCtx)
	if err != nil {
		return h.InternalServerError(ctx, err)
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
		return h.InternalServerError(ctx, err)
	}

	// return the session value for the UI to use
	// the UI will need to set the cookie because authentication is handled
	// server side
	s, err := sessions.SessionToken(sessionCtx)
	if err != nil {
		return h.InternalServerError(ctx, err)
	}

	out := &models.WebauthnBeginRegistrationResponse{
		Reply:              rout.Reply{Success: true},
		CredentialCreation: options,
		Session:            s,
	}

	return h.Success(ctx, out)
}

func (h *Handler) BindWebauthnRegistration() *openapi3.Operation {
	webauthnReg := openapi3.NewOperation()
	webauthnReg.Description = "Complete WebAuthn credential registration for a user that is currently logged in using a different Openlane authentication method. This API must be called from the backend using the user access token returned upon successful authentication. If successful, the credential will be registered for the user that corresponds to the authorization token."
	webauthnReg.Tags = []string{"passkeys"}
	webauthnReg.OperationID = "WebauthnRegistration"
	webauthnReg.Security = &openapi3.SecurityRequirements{}

	h.AddRequestBody("WebauthnRegistrationRequest", models.ExampleWebauthnBeginRegistrationRequest, webauthnReg)
	h.AddResponse("WebauthnBeginRegistrationResponse", "success", models.ExampleWebauthnBeginRegistrationResponse, webauthnReg, http.StatusOK)
	webauthnReg.AddResponse(http.StatusInternalServerError, internalServerError())
	webauthnReg.AddResponse(http.StatusBadRequest, invalidInput())

	return webauthnReg
}

// FinishWebauthnRegistration is the request to finish a webauthn registration - this is where we get the credential created by the user back
func (h *Handler) FinishWebauthnRegistration(ctx echo.Context) error {
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
		return h.BadRequest(ctx, err)
	}

	// get user from the database
	entUser, userCtx, err := h.getUserByID(reqCtx, userID)
	if err != nil {
		return h.InternalServerError(ctx, err)
	}

	// follows https://www.w3.org/TR/webauthn/#sctn-registering-a-new-credential
	response, err := protocol.ParseCredentialCreationResponseBody(ctx.Request().Body)
	if err != nil {
		return h.BadRequest(ctx, err)
	}

	// get webauthn session data from the session
	webauthnData := sessionData.(map[string]any)[sessions.WebAuthnKey]

	wd, ok := webauthnData.(webauthn.SessionData)
	if !ok {
		return h.BadRequest(ctx, ErrNoAuthUser)
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

		return h.InternalServerError(ctx, err)
	}

	if err := h.setWebauthnAllowed(userCtx, entUser); err != nil {
		return h.InternalServerError(ctx, err)
	}

	// create new claims for the user
	auth, err := h.AuthManager.GenerateUserAuthSession(userCtx, ctx.Response().Writer, entUser)
	if err != nil {
		zerolog.Ctx(reqCtx).Error().Err(err).Msg("unable to create new auth session")

		return h.InternalServerError(ctx, err)
	}

	out := &models.WebauthnRegistrationResponse{
		Reply:    rout.Reply{Success: true},
		Message:  "passkey successfully created",
		AuthData: *auth,
	}

	return h.Success(ctx, out)
}

// BeginWebauthnLogin is the request to begin a webauthn login
func (h *Handler) BeginWebauthnLogin(ctx echo.Context) error {
	var r models.WebauthnLoginRequest
	if err := ctx.Bind(&r); err != nil {
		return h.InvalidInput(ctx, err)
	}

	credential, session, err := h.WebAuthn.BeginDiscoverableLogin()
	if err != nil {
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
		zerolog.Ctx(reqCtx).Error().Err(err).Msg("unable to save session")

		return h.InternalServerError(ctx, ErrProcessingRequest)
	}

	// return the session value for the UI to use
	// the UI will need to set the cookie because authentication is handled
	// server side
	s, err := sessions.SessionToken(sessionCtx)
	if err != nil {
		zerolog.Ctx(reqCtx).Error().Err(err).Msg("unable to get session token")

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
	reqCtx := ctx.Request().Context()

	session, err := h.SessionConfig.SessionManager.Get(ctx.Request(), h.SessionConfig.CookieConfig.Name)
	if err != nil {
		zerolog.Ctx(reqCtx).Error().Err(err).Msg("unable to get session from cookie")

		return h.BadRequest(ctx, ErrInvalidCredentials)
	}

	sessionData := h.SessionConfig.SessionManager.GetSessionDataFromCookie(session)
	webauthnData := sessionData.(map[string]any)[sessions.WebAuthnKey]

	wd, ok := webauthnData.(webauthn.SessionData)
	if !ok {
		return h.BadRequest(ctx, ErrNoAuthUser)
	}

	response, err := protocol.ParseCredentialRequestResponseBody(ctx.Request().Body)
	if err != nil {
		zerolog.Ctx(reqCtx).Error().Err(err).Msg("unable to parse credential request response body")

		return h.BadRequest(ctx, ErrInvalidCredentials)
	}

	if _, err = h.WebAuthn.ValidateDiscoverableLogin(h.userHandler(reqCtx), wd, response); err != nil {
		zerolog.Ctx(reqCtx).Error().Err(err).Msg("unable to validate webauthn login")

		return h.BadRequest(ctx, ErrInvalidCredentials)
	}

	userID := string(response.Response.UserHandle)

	userIDFromCookie := sessionData.(map[string]any)[sessions.UserIDKey]

	// ensure the user is the same as the one who started the login based on the challenge
	if userIDFromCookie != response.Response.CollectedClientData.Challenge {
		zerolog.Ctx(reqCtx).Error().Err(err).Msg("challenge ids do not match")

		return h.BadRequest(ctx, ErrInvalidCredentials)
	}

	// get user from the database
	entUser, reqCtx, err := h.getUserByID(reqCtx, userID)
	if err != nil {
		zerolog.Ctx(reqCtx).Error().Err(err).Msg("unable to get user by id")

		return h.InternalServerError(ctx, ErrProcessingRequest)
	}

	if orgID, enforced := h.ssoOrgForUser(reqCtx, entUser.Email); enforced {
		if !h.HasValidSSOSession(ctx, entUser.ID) {
			return ctx.Redirect(http.StatusFound, sso.SSOLogin(ctx.Echo(), orgID))
		}
	}

	// create claims for verified user
	auth, err := h.AuthManager.GenerateUserAuthSession(reqCtx, ctx.Response().Writer, entUser)
	if err != nil {
		zerolog.Ctx(reqCtx).Error().Err(err).Msg("unable to create new auth session")

		return h.InternalServerError(ctx, ErrProcessingRequest)
	}

	// set the last seen for the user
	if err := h.updateUserLastSeen(reqCtx, userID, enums.AuthProviderCredentials); err != nil {
		zerolog.Ctx(reqCtx).Error().Err(err).Msg("unable to update last seen")

		return h.InternalServerError(ctx, ErrProcessingRequest)
	}

	out := &models.WebauthnLoginResponse{
		Reply:    rout.Reply{Success: true},
		Message:  "passkey successfully created",
		AuthData: *auth,
	}

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
