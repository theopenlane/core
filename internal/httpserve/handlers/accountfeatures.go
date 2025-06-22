package handlers

import (
	"net/http"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/rs/zerolog"
	echo "github.com/theopenlane/echox"

	"github.com/theopenlane/utils/rout"

	"github.com/theopenlane/iam/auth"
	"github.com/theopenlane/iam/fgax"

	"github.com/theopenlane/core/internal/ent/generated/organization"
	"github.com/theopenlane/core/pkg/models"

	sliceutil "github.com/theopenlane/utils/slice"
)

// AccountFeaturesHandler lists all features the authenticated user has access to in relation to an organization
func (h *Handler) AccountFeaturesHandler(ctx echo.Context) error {
	var in models.AccountFeaturesRequest
	if err := ctx.Bind(&in); err != nil {
		return h.InvalidInput(ctx, err)
	}

	reqCtx := ctx.Request().Context()

	au, err := auth.GetAuthenticatedUserFromContext(reqCtx)
	if err != nil {
		zerolog.Ctx(reqCtx).Error().Err(err).Msg("error getting authenticated user")

		return h.InternalServerError(ctx, err)
	}

	in.ID, err = h.getOrganizationID(in.ID, au)
	if err != nil {
		return h.BadRequest(ctx, err)
	}

	// validate the input
	if err := in.Validate(); err != nil {
		return h.BadRequest(ctx, err)
	}

	var features []string
	if h.FeatureCache != nil {
		features, err = h.FeatureCache.Get(reqCtx, in.ID)
		if err != nil {
			zerolog.Ctx(reqCtx).Error().Err(err).Msg("feature cache get")
		}
	}
	if len(features) == 0 {
		req := fgax.ListRequest{
			SubjectID:   in.ID,
			SubjectType: organization.Table,
			ObjectType:  "feature",
			Relation:    "enabled",
		}

		resp, err := h.DBClient.Authz.ListObjectsRequest(reqCtx, req)
		if err != nil {
			zerolog.Ctx(reqCtx).Error().Err(err).Msg("error listing features from FGA")

			return h.InternalServerError(ctx, err)
		}

		features = make([]string, 0, len(resp.Objects))
		for _, obj := range resp.Objects {
			ent, parseErr := fgax.ParseEntity(obj)
			if parseErr != nil {
				continue
			}

			features = append(features, ent.Identifier)
		}

		if h.FeatureCache != nil {
			_ = h.FeatureCache.Set(reqCtx, in.ID, features)
		}
	}

	return h.Success(ctx, models.AccountFeaturesReply{
		Reply:          rout.Reply{Success: true},
		Features:       features,
		OrganizationID: in.ID,
	})
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

// BindAccountFeatures returns the OpenAPI3 operation for accepting an account features organization request
func (h *Handler) BindAccountFeatures() *openapi3.Operation {
	orgFeatures := openapi3.NewOperation()
	orgFeatures.Description = "List features a subject has in relation to the authenticated organization"
	orgFeatures.Tags = []string{"account"}
	orgFeatures.OperationID = "AccountFeatures"
	orgFeatures.Security = AllSecurityRequirements()

	orgFeatures.AddResponse(http.StatusInternalServerError, internalServerError())
	orgFeatures.AddResponse(http.StatusBadRequest, badRequest())
	orgFeatures.AddResponse(http.StatusBadRequest, invalidInput())

	return orgFeatures
}

// BindAccountFeatures returns the OpenAPI3 operation for accepting an account features organization request
func (h *Handler) BindAccountFeaturesByID() *openapi3.Operation {
	orgFeatures := openapi3.NewOperation()
	orgFeatures.Description = "List the features a subject has in relation to the organization ID provided"
	orgFeatures.Tags = []string{"account"}
	orgFeatures.OperationID = "AccountFeaturesByID"
	orgFeatures.Security = AllSecurityRequirements()

	h.AddResponse("AccountFeaturesReply", "success", models.ExampleAccountFeaturesReply, orgFeatures, http.StatusOK)
	orgFeatures.AddResponse(http.StatusInternalServerError, internalServerError())
	orgFeatures.AddResponse(http.StatusBadRequest, badRequest())
	orgFeatures.AddResponse(http.StatusUnauthorized, unauthorized())

	return orgFeatures
}
