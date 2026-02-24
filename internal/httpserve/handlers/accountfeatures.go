package handlers

import (
	"sort"

	"github.com/samber/lo"
	echo "github.com/theopenlane/echox"

	"github.com/theopenlane/utils/rout"

	"github.com/theopenlane/iam/auth"

	models "github.com/theopenlane/core/common/openapi"
	"github.com/theopenlane/core/internal/ent/privacy/rule"
	"github.com/theopenlane/core/pkg/logx"
)

// AccountFeaturesHandler lists all features the authenticated user has access to in relation to an organization
func (h *Handler) AccountFeaturesHandler(ctx echo.Context, openapi *OpenAPIContext) error {
	in, err := BindAndValidateQueryParamsWithResponse(ctx, openapi.Operation, models.ExampleAccountFeaturesRequest, models.ExampleAccountFeaturesReply, openapi.Registry)
	if err != nil {
		return h.InvalidInput(ctx, err, openapi)
	}

	if isRegistrationContext(ctx) {
		return nil
	}

	reqCtx := ctx.Request().Context()

	caller, ok := auth.CallerFromContext(reqCtx)
	if !ok || caller == nil {
		logx.FromContext(reqCtx).Error().Msg("error getting authenticated user")

		return h.InternalServerError(ctx, ErrProcessingRequest, openapi)
	}

	orgID, orgIDErr := h.getOrganizationID(in.ID, caller)
	if orgIDErr != nil {
		return h.BadRequest(ctx, orgIDErr, openapi)
	}

	in.ID = orgID

	var (
		features []string
		featErr  error
	)

	if in.ID != "" {
		features, featErr = rule.GetFeaturesForSpecificOrganization(reqCtx, in.ID)
	} else {
		features, featErr = rule.GetOrgFeatures(reqCtx)
	}

	if featErr != nil {
		logx.FromContext(reqCtx).Error().Err(featErr).Msg("error getting features")

		return h.InternalServerError(ctx, ErrProcessingRequest, openapi)
	}

	// sort for consistency
	sort.Strings(features)

	return h.Success(ctx, models.AccountFeaturesReply{
		Reply:          rout.Reply{Success: true},
		Features:       features,
		OrganizationID: in.ID,
	}, openapi)
}

// getOrganizationID returns the organization ID to use for the request based on the input and caller
func (h *Handler) getOrganizationID(id string, caller *auth.Caller) (string, error) {
	orgIDs := caller.OrgIDs()

	// if an ID is provided, check if the authenticated user has access to it
	if id != "" {
		if !lo.Contains(orgIDs, id) {
			return "", ErrInvalidInput
		}

		return id, nil
	}

	// if no ID is provided, default to the authenticated organization
	if caller.OrganizationID != "" {
		return caller.OrganizationID, nil
	}

	// if it is still empty, and the personal access token only has one organization use that
	if len(orgIDs) == 1 {
		return orgIDs[0], nil
	}

	return "", nil
}
