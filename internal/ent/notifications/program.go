package notifications

import (
	"context"
	"encoding/json"

	"github.com/theopenlane/core/common/enums"

	"github.com/theopenlane/core/v2/internal/ent/entityops"
	"github.com/theopenlane/core/v2/internal/ent/generated"
	"github.com/theopenlane/core/v2/internal/ent/generated/programmembership"
	"github.com/theopenlane/core/v2/pkg/logx"
)

func programAuditorRecipients(inv entityops.Invocation, payload entityops.MutationPayload, row json.RawMessage) ([]string, error) {
	ownerID := entityops.FieldValue(row, "owner_id")
	ids, err := getProgramAuditorUserIDs(inv.Context, inv.Client, payload.EntityID)
	if err != nil {
		return nil, err
	}

	if len(ids) == 0 {
		ids, err = OrgUserIDsByRole(inv.Context, inv.Client, ownerID, enums.RoleAuditor)
		if err != nil {
			return nil, err
		}
	}

	if len(ids) == 0 {
		logx.FromContext(inv.Context).Warn().Str("org_id", ownerID).Msg("no auditors found for program ready notification")
	}

	return ids, nil
}

func getProgramAuditorUserIDs(ctx context.Context, client *generated.Client, programID string) ([]string, error) {
	var userIDs []string
	err := client.ProgramMembership.Query().
		Where(
			programmembership.ProgramID(programID),
			programmembership.RoleEQ(enums.RoleAuditor),
		).
		Select(programmembership.FieldUserID).
		Scan(ctx, &userIDs)
	if err != nil {
		return nil, err
	}

	return userIDs, nil
}
