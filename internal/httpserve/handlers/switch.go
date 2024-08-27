package handlers

import (
	"net/http"

	"github.com/getkin/kin-openapi/openapi3"
	ph "github.com/posthog/posthog-go"
	echo "github.com/theopenlane/echox"
	"github.com/theopenlane/iam/fgax"

	"github.com/theopenlane/utils/rout"

	"github.com/theopenlane/core/internal/ent/generated/privacy"
	"github.com/theopenlane/core/pkg/models"
	"github.com/theopenlane/iam/auth"
)

// SwitchHandler is responsible for handling requests to the `/switch` endpoint, and changing the user's logged in organization context
func (h *Handler) SwitchHandler(ctx echo.Context) error {
	var in models.SwitchOrganizationRequest
	if err := ctx.Bind(&in); err != nil {
		return h.InvalidInput(ctx, err)
	}

	if err := in.Validate(); err != nil {
		return h.InvalidInput(ctx, err)
	}

	reqCtx := ctx.Request().Context()

	userID, err := auth.GetUserIDFromContext(reqCtx)
	if err != nil {
		h.Logger.Errorw("unable to get user id from context", "error", err)

		return h.BadRequest(ctx, err)
	}

	// get user from database by subject
	user, err := h.getUserDetailsByID(reqCtx, userID)
	if err != nil {
		h.Logger.Errorw("unable to get user by subject", "error", err)

		return h.BadRequest(ctx, err)
	}

	orgID, err := auth.GetOrganizationIDFromContext(reqCtx)
	if err != nil {
		h.Logger.Errorw("unable to get organization id from context", "error", err)

		return h.BadRequest(ctx, err)
	}

	// ensure the user is not already in the target organization
	if orgID == in.TargetOrganizationID {
		return h.BadRequest(ctx, ErrAlreadySwitchedIntoOrg)
	}

	// ensure user is already a member of the destination organization
	req := fgax.AccessCheck{
		SubjectID:   userID,
		SubjectType: auth.UserSubjectType,
		ObjectID:    in.TargetOrganizationID,
	}

	if allow, err := h.DBClient.Authz.CheckOrgReadAccess(reqCtx, req); err != nil || !allow {
		h.Logger.Errorw("user not authorized to access organization", "error", err)

		return h.Unauthorized(ctx, err)
	}

	// get the target organization
	orgGetCtx := privacy.DecisionContext(reqCtx, privacy.Allow)

	org, err := h.getOrgByID(orgGetCtx, in.TargetOrganizationID)
	if err != nil {
		h.Logger.Errorw("unable to get target organization by id", "error", err)

		return h.BadRequest(ctx, err)
	}

	// create new claims for the user
	auth, err := h.AuthManager.GenerateUserAuthSessionWithOrg(ctx, user, org.ID)
	if err != nil {
		h.Logger.Errorw("unable create new auth session", "error", err)

		return h.InternalServerError(ctx, err)
	}

	// track the organization switch event
	props := ph.NewProperties().
		Set("user_id", user.ID).
		Set("email", user.Email).
		Set("target_organization_id", org.ID).
		Set("auth_provider", user.AuthProvider).
		Set("previous_organization_id", orgID)

	h.AnalyticsClient.Event("organization_switched", props)

	// set the out attributes we send back to the client only on success
	out := &models.SwitchOrganizationReply{
		Reply:    rout.Reply{Success: true},
		AuthData: *auth,
	}

	return h.Success(ctx, out)
}

// BindSwitchHandler binds the reset password handler to the OpenAPI schema
func (h *Handler) BindSwitchHandler() *openapi3.Operation {
	switchHandler := openapi3.NewOperation()
	switchHandler.Description = "Switch the user's organization context"
	switchHandler.OperationID = "OrganizationSwitch"

	h.AddRequestBody("SwitchOrganizationRequest", models.ExampleSwitchSuccessRequest, switchHandler)
	h.AddResponse("SwitchOrganizationReply", "success", models.ExampleSwitchSuccessReply, switchHandler, http.StatusOK)
	switchHandler.AddResponse(http.StatusInternalServerError, internalServerError())
	switchHandler.AddResponse(http.StatusBadRequest, badRequest())
	switchHandler.AddResponse(http.StatusUnauthorized, unauthorized())

	return switchHandler
}
