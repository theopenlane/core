package handlers

import (
	"net/http"

	"github.com/getkin/kin-openapi/openapi3"
	echo "github.com/theopenlane/echox"
	"github.com/theopenlane/iam/auth"
	"github.com/theopenlane/utils/rout"

	"github.com/theopenlane/core/pkg/enums"
	"github.com/theopenlane/core/pkg/models"
)

// AvailableAuthType validates the username/email and returns the
// auth types available to the user
func (h *Handler) AvailableAuthTypeHandler(ctx echo.Context) error {
	var in models.AvailableAuthTypeLoginRequest
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

	if !user.Edges.Setting.EmailConfirmed {
		return h.BadRequest(ctx, auth.ErrUnverifiedUser)
	}

	availableAuthMethods := []enums.AuthProvider{
		enums.AuthProviderCredentials, // available by default
	}

	if user.Edges.Setting.IsWebauthnAllowed {
		availableAuthMethods = append(availableAuthMethods, enums.AuthProviderWebauthn)
	}

	out := models.AvailableAuthTypeReply{
		Reply:   rout.Reply{Success: true},
		Methods: availableAuthMethods,
	}

	return h.Success(ctx, out)
}

// BindAvailableAuthTypeHandler binds the available auth methods request to the OpenAPI schema
func (h *Handler) BindAvailableAuthTypeHandler() *openapi3.Operation {
	availableAuthReq := openapi3.NewOperation()
	availableAuthReq.Description = `AvailableAuthType is oriented towards human users who have their emails but 
	would like to get authenticated via a self selected option of possible auth choices 
	( currently password or passkey )`
	availableAuthReq.Tags = []string{"authentication"}
	availableAuthReq.OperationID = "AvailableAuthTypeHandler"
	availableAuthReq.Security = BasicSecurity()

	h.AddRequestBody("AvailableAuthTypeLoginRequest", models.ExampleAvailableAuthTypeRequest, availableAuthReq)
	h.AddResponse("LoginReply", "success", models.ExampleAvailableAuthTypeSuccessResponse, availableAuthReq, http.StatusOK)
	availableAuthReq.AddResponse(http.StatusInternalServerError, internalServerError())
	availableAuthReq.AddResponse(http.StatusBadRequest, badRequest())
	availableAuthReq.AddResponse(http.StatusBadRequest, invalidInput())

	return availableAuthReq
}
