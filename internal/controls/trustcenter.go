package controls

import (
	"context"
	"errors"

	"github.com/theopenlane/iam/auth"
	"github.com/theopenlane/utils/rout"

	"github.com/theopenlane/core/v2/internal/ent/generated"
	"github.com/theopenlane/core/v2/internal/ent/generated/standard"
	"github.com/theopenlane/core/v2/internal/graphapi/common"
	"github.com/theopenlane/logx"
)

var (
	// trustCenterStandardShortName is the short name of the trust center standard
	// used to identify controls that should be flagged as trust center controls during clone
	trustCenterStandardShortName = "OTS"

	// TrustCenterFrameworkName is the name of the framework the trust center standard belongs to, used in conjunction with trustCenterStandardShortName to identify controls that should be flagged as trust center controls during clone
	TrustCenterFrameworkName = "openlane-trust-center"

	// StandardTrustCenterFrameworkName is the name of the fremework the default standards go to
	StandardTrustCenterFrameworkName = "openlane-standard"

	// trustCenterStandardFilter is the filter used to identify controls that should be flagged as trust center controls during clone
	trustCenterStandardFilter = CloneFilterOptions{
		StandardShortName:     &trustCenterStandardShortName,
		StandardFrameworkName: &TrustCenterFrameworkName,
	}
)

// isTrustCenterStandard returns true if the standard is the trust center standard
func isTrustCenterStandard(std *generated.Standard) bool {
	return std != nil && std.ShortName == trustCenterStandardShortName
}

// getTrustCenterControls retrieves the trust center controls
func getTrustCenterControls(ctx context.Context, client *generated.Client) ([]*generated.Control, error) {
	if client == nil {
		return nil, nil
	}

	// ensure we get the public, system owned standard
	std, err := findStandard(ctx, client, trustCenterStandardFilter, standard.IsPublic(true), standard.SystemOwned(true))
	if err != nil {
		var notFound common.NotFoundError
		if errors.As(err, &notFound) {
			logx.FromContext(ctx).Error().Msg("trust center standard not found")

			return nil, nil
		}

		return nil, err
	}

	// if we get the standard back, all controls should be accessible so we can run as an internal read to skip checks
	internalCtx := auth.WithInternalReadContext(ctx)
	where, err := ControlFilterByStandard(internalCtx, trustCenterStandardFilter, std)
	if err != nil {

		return nil, err
	}

	controls, err := client.Control.Query().
		Where(
			where...,
		).
		WithStandard().
		All(internalCtx)
	if err != nil {
		return nil, err
	}

	return controls, nil
}

// CloneTrustCenterControl clones the trust center controls and assumes the the user has the trust center module already
// this is intended to be called from an internal-hook when a trust center is created
func CloneTrustCenterControls(ctx context.Context, m *generated.TrustCenterMutation) error {
	orgID, err := auth.GetOrganizationIDFromContext(ctx)
	if err != nil {
		return rout.NewMissingRequiredFieldError("owner_id")
	}

	controls, err := getTrustCenterControls(ctx, m.Client())
	if err != nil {
		return err
	}

	if len(controls) == 0 {
		logx.FromContext(ctx).Info().Msg("no trust center controls found to clone")

		return nil
	}

	// trust center controls do no have subcontrols so we can ignore the returned subcontrols to create
	_, _, err = CloneControls(ctx, m.Client(), controls, WithOrgID(orgID), WithPool(m.Client().Pool))
	if err != nil {
		logx.FromContext(ctx).Error().Err(err).Msg("error cloning trust center controls")

		return err
	}

	return nil
}
