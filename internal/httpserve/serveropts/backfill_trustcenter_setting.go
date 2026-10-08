package serveropts

import (
	"context"

	"github.com/theopenlane/core/common/models"
	"github.com/theopenlane/core/v2/internal/ent/generated/trustcentersetting"
)

func backfillTrustCenterSetting(ctx context.Context, deps backfillDeps) error {
	const batchSize = 100

	var afterCursor string

	for {
		settings, err := deps.Client.TrustCenterSetting.Query().
			Where(
				trustcentersetting.IDGT(afterCursor),
				trustcentersetting.AutoApprovalRulesIsNil(),
			).
			Order(trustcentersetting.ByID()).
			Limit(batchSize).
			All(ctx)
		if err != nil {
			return err
		}

		if len(settings) == 0 {
			return nil
		}

		ids := make([]string, 0, len(settings))
		for _, setting := range settings {
			ids = append(ids, setting.ID)
		}

		if err := deps.Client.TrustCenterSetting.Update().
			Where(
				trustcentersetting.IDIn(ids...),
				trustcentersetting.AutoApprovalRulesIsNil(),
			).
			SetEnableAutoApproval(false).
			SetAutoApprovalRules(models.TrustCenterNDARequestSetting{
				ManualApprovalOnFailure: true,
			}).
			Exec(ctx); err != nil {
			return err
		}

		afterCursor = settings[len(settings)-1].ID
		if len(settings) < batchSize {
			return nil
		}
	}
}
