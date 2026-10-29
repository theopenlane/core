package operations

import (
	"context"
	"fmt"

	"github.com/theopenlane/core/common/enums"

	ent "github.com/theopenlane/core/v2/internal/ent/generated"
	"github.com/theopenlane/core/v2/internal/integrations/types"
	"github.com/theopenlane/logx"
)

// importCompletedObjectType is the notification object type for a finished import
const importCompletedObjectType = "import.completed"

// Outcome is the organization notification describing how a background operation finished
type Outcome struct {
	// ObjectType classifies the notification
	ObjectType string
	// Title is the notification title
	Title string
	// Body is the notification body
	Body string
	// Data is structured detail attached to the notification
	Data map[string]any
}

// NotifyOutcome sends the organization notification describing how a background operation finished
func NotifyOutcome(ctx context.Context, client *ent.Client, organizationID string, outcome Outcome) error {
	_, err := client.Notification.Create().
		SetOwnerID(organizationID).
		SetNotificationType(enums.NotificationTypeOrganization).
		SetObjectType(outcome.ObjectType).
		SetTitle(outcome.Title).
		SetBody(outcome.Body).
		SetData(outcome.Data).
		SetTopic(enums.NotificationTopicImportComplete).
		Save(ctx)

	return err
}

// NotifyImportOutcome sends the organization notification summarizing an import
// subject names what was imported, and detail, such as how to retry, is appended to the body when set
func NotifyImportOutcome(ctx context.Context, client *ent.Client, organizationID, subject string, outcome types.IngestOutcome, detail string) error {
	title := subject + " import completed"
	body := fmt.Sprintf("Imported %d record(s)", outcome.Succeeded)

	switch {
	case outcome.Err != nil:
		logx.FromContext(ctx).Error().Err(outcome.Err).Str("subject", subject).Msg("import failed")

		title = subject + " import failed"
		body = "The import could not be completed"
	case outcome.Failed > 0:
		body += fmt.Sprintf(", %d record(s) could not be imported", outcome.Failed)
	}

	if detail != "" {
		body += ", " + detail
	}

	return NotifyOutcome(ctx, client, organizationID, Outcome{
		ObjectType: importCompletedObjectType,
		Title:      title,
		Body:       body,
		Data:       map[string]any{"succeeded": outcome.Succeeded, "failed": outcome.Failed, "failure_summary": outcome.FailureSummary},
	})
}
