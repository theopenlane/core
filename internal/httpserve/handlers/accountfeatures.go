package handlers

import (
	"sort"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	echo "github.com/theopenlane/echox"

	"github.com/theopenlane/utils/rout"

	"github.com/theopenlane/iam/auth"

	"github.com/theopenlane/core/internal/ent/privacy/rule"
	"github.com/theopenlane/core/pkg/models"

	sliceutil "github.com/theopenlane/utils/slice"
)

// AccountFeaturesHandler lists all features the authenticated user has access to in relation to an organization
func (h *Handler) AccountFeaturesHandler(ctx echo.Context, openapi *OpenAPIContext) error {
	in, err := BindAndValidateQueryParamsWithResponse(ctx, openapi.Operation, models.ExampleAccountFeaturesRequest, models.ExampleAccountFeaturesReply, openapi.Registry)
	if err != nil {
		return h.InvalidInput(ctx, err, openapi)
	}

	reqCtx := ctx.Request().Context()

	au, err := auth.GetAuthenticatedUserFromContext(reqCtx)
	if err != nil {
		zerolog.Ctx(reqCtx).Error().Err(err).Msg("error getting authenticated user")

		return h.InternalServerError(ctx, err, openapi)
	}

	in.ID, err = h.getOrganizationID(in.ID, au)
	if err != nil {
		return h.BadRequest(ctx, err, openapi)
	}

	var features []string

	if in.ID != "" {
		features, err = rule.GetFeaturesForSpecificOrganization(reqCtx, in.ID)
	} else {
		features, err = rule.GetOrgFeatures(reqCtx)
	}

	if err != nil {
		log.Error().Err(err).Msg("error getting features")

		return h.InternalServerError(ctx, err, openapi)
	}

	// sort for consistency
	sort.Strings(features)

	return h.Success(ctx, models.AccountFeaturesReply{
		Reply:          rout.Reply{Success: true},
		Features:       features,
		OrganizationID: in.ID,
	}, openapi)
}

// getOrganizationID returns the organization ID to use for the request based on the input and authenticated user
func (h *Handler) getOrganizationID(id string, au *auth.AuthenticatedUser) (string, error) {
	// if an ID is provided, check if the authenticated user has access to it
	if id != "" {
		if !sliceutil.Contains(au.OrganizationIDs, id) {
			return "", ErrInvalidInput
		}

		return id, nil
	}

	// if no ID is provided, default to the authenticated organization
	if au.OrganizationID != "" {
		return au.OrganizationID, nil
	}

	// if it is still empty, and the personal access token only has one organization use that
	if len(au.OrganizationIDs) == 1 {
		return au.OrganizationIDs[0], nil
	}

	return "", nil
}
