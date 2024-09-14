package handlers

import (
	"context"

	"github.com/rs/zerolog/log"
	echo "github.com/theopenlane/echox"

	"github.com/go-webauthn/webauthn/protocol"
	"github.com/go-webauthn/webauthn/webauthn"

	"github.com/theopenlane/utils/rout"

	"github.com/theopenlane/iam/auth"
	provider "github.com/theopenlane/iam/providers/webauthn"
	"github.com/theopenlane/iam/sessions"

	"github.com/theopenlane/core/internal/ent/privacy/token"
	"github.com/theopenlane/core/pkg/enums"
	"github.com/theopenlane/core/pkg/models"
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

	ctxWithToken := token.NewContextWithOauthTooToken(ctx.Request().Context(), r.Email)

	// to register a new passkey, the user needs to be created + logged in first
	// once the the passkey is added to the user's account, they can use it to login
	// we treat this verify similar to the oauth or basic registration flow
	// user is created first, no credential method is set / they are unable to login until the credential flow is finished
	entUser, err := h.CheckAndCreateUser(ctxWithToken, r.Name, r.Email, enums.AuthProvider(webauthnProvider), "")
	if err != nil {
		return h.InternalServerError(ctx, err)
	}

	// set context for remaining request based on logged in user
	userCtx := auth.AddAuthenticatedUserContext(ctx, &auth.AuthenticatedUser{
		SubjectID: entUser.ID,
	})

	// set webauthn allowed
	if err := h.setWebauthnAllowed(userCtx, entUser); err != nil {
		return h.InternalServerError(ctx, err)
	}

	if err := h.addDefaultOrgToUserQuery(userCtx, entUser); err != nil {
		return h.InternalServerError(ctx, err)
	}

	user := &provider.User{
		ID:    entUser.ID,
		Email: entUser.Email,
		Name:  entUser.FirstName + " " + entUser.LastName,
	}

	// options is the object that needs to be returned for the front end to open the creation dialog for the user to create the passkey
	options, session, err := h.WebAuthn.BeginRegistration(user,
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

// FinishWebauthnRegistration is the request to finish a webauthn registration - this is where we get the credential created by the user back
func (h *Handler) FinishWebauthnRegistration(ctx echo.Context) error {
	// lookup userID in cache to ensure cookie and tokens match
	session, err := h.SessionConfig.SessionManager.Get(ctx.Request(), h.SessionConfig.CookieConfig.Name)
	if err != nil {
		return h.BadRequest(ctx, err)
	}

	// Get sessionID from cookie and check against redis
	sessionID := h.SessionConfig.SessionManager.GetSessionIDFromCookie(session)

	userID, err := h.SessionConfig.RedisStore.GetSession(ctx.Request().Context(), sessionID)
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
	entUser, err := h.getUserByID(ctx.Request().Context(), userID, enums.AuthProvider(webauthnProvider))
	if err != nil {
		return h.InternalServerError(ctx, err)
	}

	// set user in the viewer context for the rest of the request
	userCtx := auth.AddAuthenticatedUserContext(ctx, &auth.AuthenticatedUser{
		SubjectID: entUser.ID,
	})

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

	// create new claims for the user
	auth, err := h.AuthManager.GenerateUserAuthSession(ctx, entUser)
	if err != nil {
		log.Error().Err(err).Msg("unable to create new auth session")

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
	credential, session, err := h.WebAuthn.BeginDiscoverableLogin()
	if err != nil {
		return err
	}

	setSessionMap := map[string]any{}
	setSessionMap[sessions.WebAuthnKey] = session
	setSessionMap[sessions.UserTypeKey] = webauthnLogin

	sessionCtx, err := h.SessionConfig.SaveAndStoreSession(ctx.Request().Context(), ctx.Response().Writer, setSessionMap, "")
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

	out := &models.WebauthnBeginLoginResponse{
		Reply:               rout.Reply{Success: true},
		CredentialAssertion: credential,
		Session:             s,
	}

	return h.Success(ctx, out)
}

// FinishWebauthnLogin is the request to finish a webauthn login
func (h *Handler) FinishWebauthnLogin(ctx echo.Context) error {
	session, err := h.SessionConfig.SessionManager.Get(ctx.Request(), h.SessionConfig.CookieConfig.Name)
	if err != nil {
		return h.BadRequest(ctx, err)
	}

	sessionData := h.SessionConfig.SessionManager.GetSessionDataFromCookie(session)
	webauthnData := sessionData.(map[string]any)[sessions.WebAuthnKey]

	wd, ok := webauthnData.(webauthn.SessionData)
	if !ok {
		return h.BadRequest(ctx, ErrNoAuthUser)
	}

	response, err := protocol.ParseCredentialRequestResponseBody(ctx.Request().Body)
	if err != nil {
		return h.BadRequest(ctx, err)
	}

	if _, err = h.WebAuthn.ValidateDiscoverableLogin(h.userHandler(ctx.Request().Context()), wd, response); err != nil {
		return h.BadRequest(ctx, err)
	}

	// get user from the database
	entUser, err := h.getUserByID(ctx.Request().Context(), string(response.Response.UserHandle), enums.AuthProvider(webauthnProvider))
	if err != nil {
		return h.InternalServerError(ctx, err)
	}

	// create claims for verified user
	auth, err := h.AuthManager.GenerateUserAuthSession(ctx, entUser)
	if err != nil {
		log.Error().Err(err).Msg("unable to create new auth session")

		return h.InternalServerError(ctx, err)
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
	return func(rawID, userHandle []byte) (user webauthn.User, err error) {
		u, err := h.getUserByID(ctx, string(userHandle), enums.AuthProvider(webauthnProvider))
		if err != nil {
			return nil, err
		}

		authnUser := &provider.User{
			ID:                  u.ID,
			Email:               u.Email,
			Name:                u.FirstName + " " + u.LastName,
			WebauthnCredentials: []webauthn.Credential{},
		}

		for _, cred := range u.Edges.Webauthn {
			authnCred := webauthn.Credential{
				ID:              cred.CredentialID,
				PublicKey:       cred.PublicKey,
				AttestationType: cred.AttestationType,
			}

			for _, t := range cred.Transports {
				authnCred.Transport = append(authnCred.Transport, protocol.AuthenticatorTransport(t))
			}

			authnUser.WebauthnCredentials = append(authnUser.WebauthnCredentials, authnCred)
		}

		return authnUser, nil
	}
}
