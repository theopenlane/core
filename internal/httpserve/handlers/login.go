package handlers

import (
	"net/http"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/rs/zerolog/log"
	echo "github.com/theopenlane/echox"

	"github.com/theopenlane/utils/rout"

	"github.com/theopenlane/iam/auth"

	"github.com/theopenlane/utils/passwd"

	"github.com/theopenlane/core/pkg/enums"
	"github.com/theopenlane/core/pkg/models"
)

// LoginHandler validates the user credentials and returns a valid cookie
// this handler only supports username password login
func (h *Handler) LoginHandler(ctx echo.Context) error {
	var in models.LoginRequest
	if err := ctx.Bind(&in); err != nil {
		return h.InvalidInput(ctx, err)
	}

	if err := in.Validate(); err != nil {
		return h.InvalidInput(ctx, err)
	}

	reqCtx := ctx.Request().Context()

	// check user in the database, username == email and ensure only one record is returned
	user, err := h.getUserByEmail(reqCtx, in.Username, enums.AuthProviderCredentials)
	if err != nil {
		return h.BadRequest(ctx, auth.ErrNoAuthUser)
	}

	if user.Edges.Setting.Status != enums.UserStatusActive {
		return h.BadRequest(ctx, auth.ErrNoAuthUser)
	}

	// verify the password is correct
	valid, err := passwd.VerifyDerivedKey(*user.Password, in.Password)
	if err != nil || !valid {
		return h.BadRequest(ctx, rout.ErrInvalidCredentials)
	}

	if !user.Edges.Setting.EmailConfirmed {
		return h.BadRequest(ctx, auth.ErrUnverifiedUser)
	}

	// set context for remaining request based on logged in user
	userCtx := setAuthenticatedContext(reqCtx, user)

	// create new claims for the user
	auth, err := h.AuthManager.GenerateUserAuthSession(userCtx, ctx.Response().Writer, user)
	if err != nil {
		log.Error().Err(err).Msg("unable to create new auth session")

		return h.InternalServerError(ctx, err)
	}

	if err := h.updateUserLastSeen(userCtx, user.ID); err != nil {
		log.Error().Err(err).Msg("unable to update last seen")

		return h.InternalServerError(ctx, err)
	}

	out := models.LoginReply{
		Reply:      rout.Reply{Success: true},
		TFAEnabled: user.Edges.Setting.IsTfaEnabled,
		Message:    "success",
		AuthData:   *auth,
	}

	return h.Success(ctx, out)
}

// BindLoginHandler binds the login request to the OpenAPI schema
func (h *Handler) BindLoginHandler() *openapi3.Operation {
	login := openapi3.NewOperation()
	login.Description = "Login is oriented towards human users who use their email and password for authentication. Login verifies the password submitted for the user is correct by looking up the user by email and using the argon2 derived key verification process to confirm the password matches. Upon authentication an access token and a refresh token with the authorized claims of the user are returned. The user can use the access token to authenticate to our systems. The access token has an expiration and the refresh token can be used with the refresh endpoint to get a new access token without the user having to log in again. The refresh token overlaps with the access token to provide a seamless authentication experience and the user can refresh their access token so long as the refresh token is valid"
	login.Tags = []string{"authentication"}
	login.OperationID = "LoginHandler"
	login.Security = BasicSecurity()

	h.AddRequestBody("LoginRequest", models.ExampleLoginSuccessRequest, login)
	h.AddResponse("LoginReply", "success", models.ExampleLoginSuccessResponse, login, http.StatusOK)
	login.AddResponse(http.StatusInternalServerError, internalServerError())
	login.AddResponse(http.StatusBadRequest, badRequest())
	login.AddResponse(http.StatusBadRequest, invalidInput())

	return login
}
