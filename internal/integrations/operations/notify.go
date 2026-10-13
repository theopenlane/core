package operations

import (
	"context"
	"fmt"

	"github.com/theopenlane/core/common/enums"

	ent "github.com/theopenlane/core/v2/internal/ent/generated"
	"github.com/theopenlane/core/v2/internal/integrations/types"
	"github.com/theopenlane/core/v2/pkg/logx"
)

// NotifyImportOutcome sends the organization notification summarizing an import, telling the user to import
// again when records failed; subject names what was imported and detail is appended to the body when set
func NotifyImportOutcome(ctx context.Context, client *ent.Client, organizationID, subject string, outcome types.IngestOutcome, detail string) error {
	title := subject + " import completed"
	body := fmt.Sprintf("Imported %d record(s)", outcome.Succeeded)

	switch {
	case outcome.Err != nil:
		logx.FromContext(ctx).Error().Err(outcome.Err).Str("subject", subject).Msg("import failed")

		title = subject + " import failed"
		body = "The import could not be completed; import the review again to retry"
	case outcome.Failed > 0:
		body += fmt.Sprintf("; %d record(s) could not be imported, import the review again to retry", outcome.Failed)
	}

	if detail != "" {
		body += "; " + detail
	}

	_, err := client.Notification.Create().
		SetOwnerID(organizationID).
		SetNotificationType(enums.NotificationTypeOrganization).
		SetObjectType("scan.import.completed").
		SetTitle(title).
		SetBody(body).
		SetData(map[string]any{"succeeded": outcome.Succeeded, "failed": outcome.Failed, "failure_summary": outcome.FailureSummary}).
		SetTopic(enums.NotificationTopicImportComplete).
		Save(ctx)

	return err
}
