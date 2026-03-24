package controls

import (
	"context"

	"entgo.io/ent/dialect/sql"
	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/generated/privacy"
	"github.com/theopenlane/core/internal/ent/generated/standard"
	"github.com/theopenlane/core/pkg/logx"
	"github.com/theopenlane/iam/auth"
	"github.com/theopenlane/utils/rout"
)

// trustCenterStandardShortName is the short name of the trust center standard
// used to identify controls that should be flagged as trust center controls during clone
var trustCenterStandardShortName = "OTS"
var TrustCenterFrameworkName = "openlane-trust-center"

// isTrustCenterStandard returns true if the standard is the trust center standard
func isTrustCenterStandard(std *generated.Standard) bool {
	return std != nil && std.ShortName == trustCenterStandardShortName
}

var trustCenterStandardFilter = CloneFilterOptions{
	StandardShortName:     &trustCenterStandardShortName,
	StandardFrameworkName: &TrustCenterFrameworkName,
}

func getTrustCenterControls(ctx context.Context, client *generated.Client) ([]*generated.Control, error) {
	if client == nil {
		return nil, nil
	}

	stdWhereFilter := StandardFilter(trustCenterStandardFilter)
	stds, err := client.Standard.Query().
		Where(stdWhereFilter...).
		Select(standard.FieldID, standard.FieldIsPublic).
		Order(standard.OrderOption(standard.ByVersion(sql.OrderDesc()))).
		All(ctx)
	if err != nil || stds == nil || len(stds) == 0 {
		logx.FromContext(ctx).Error().Err(err).Msgf("error getting standard with ID")

		return nil, err
	}

	// get the first standard, this will be the most recent revision if multiple revisions exist
	std := stds[0]

	// if we get the standard back, all controls should be accessible so we can allow context to skip checks
	allowCtx := privacy.DecisionContext(ctx, privacy.Allow)
	where, err := ControlFilterByStandard(allowCtx, trustCenterStandardFilter, std)
	if err != nil {

		return nil, err
	}

	controls, err := client.Control.Query().
		Where(
			where...,
		).
		WithStandard().
		All(allowCtx)
	if err != nil {
		return nil, err
	}

	return controls, nil
}

// CloneTrustCenterControl clones the trust center controls and assumes the the user has the trust center module already
// this is intended to be called from an internal-hook when a trust center is created
func CloneTrustCenterControls(ctx context.Context) error {
	caller, ok := auth.CallerFromContext(ctx)
	if !ok || caller == nil || caller.OrganizationID == "" {
		return rout.NewMissingRequiredFieldError("owner_id")
	}

	orgID := caller.OrganizationID

	controls, err := getTrustCenterControls(ctx, getClientFromContext(ctx))
	if err != nil {
		return err
	}

	_, _, err = CloneControls(ctx, getClientFromContext(ctx), controls, WithOrgID(orgID))
	if err != nil {
		logx.FromContext(ctx).Error().Err(err).Msg("error cloning trust center controls")

		return err
	}

	return nil
}

// getClientFromContext is a helper function to get the generated client from the context and log an error if it is not found
// it will not prevent the function from executing, but it will return nil and log the error for debugging purposes
func getClientFromContext(ctx context.Context) *generated.Client {
	client := generated.FromContext(ctx)
	if client == nil {
		logx.FromContext(ctx).Error().Msg("unable to get client from context")

		return nil
	}

	return client
}
