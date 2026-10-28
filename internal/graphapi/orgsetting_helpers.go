package graphapi

import (
	"context"

	"github.com/theopenlane/core/v2/internal/ent/generated/organization"
	"github.com/theopenlane/core/v2/internal/graphapi/common"
)

// orgSettingID returns the organization's own setting id, rejecting a requested setting id that belongs to another org
func orgSettingID(ctx context.Context, orgID string, requested *string) (string, error) {
	settingID, err := withTransactionalMutation(ctx).Organization.Query().
		Where(organization.ID(orgID)).
		QuerySetting().
		OnlyID(ctx)
	if err != nil {
		return "", parseRequestError(ctx, err, common.Action{Action: common.ActionUpdate, Object: "organization"})
	}

	if requested != nil && *requested != settingID {
		return "", newPermissionDeniedError()
	}

	return settingID, nil
}
