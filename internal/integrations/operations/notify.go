package operations

import (
	"context"
	"fmt"

	"github.com/samber/lo"

	"github.com/theopenlane/iam/auth"

	"github.com/theopenlane/core/common/enums"

	"github.com/theopenlane/core/v2/internal/ent/entityops"
	ent "github.com/theopenlane/core/v2/internal/ent/generated"
	"github.com/theopenlane/logx"
)

const (
	// importCompletedObjectType is the notification object type for a finished import
	importCompletedObjectType = "import.completed"
	// deleteCompletedObjectType is the notification object type for a finished source delete
	deleteCompletedObjectType = "delete.completed"
)

// NotifyOutcome sends the notification describing how a background operation finished to the caller's organization
// the title and body are sent as written, the type defaults to organization and the topic to import complete
func NotifyOutcome(ctx context.Context, objectType string, content entityops.NotificationContent) error {
	orgID, err := auth.GetOrganizationIDFromContext(ctx)
	if err != nil {
		return err
	}

	// notifications are only created from inside the server, not as the caller the operation runs as
	ctx = auth.WithOrgInternalCaller(ctx, orgID)

	create := ent.FromContext(ctx).Notification.Create().
		SetNotificationType(lo.CoalesceOrEmpty(content.Type, enums.NotificationTypeOrganization)).
		SetObjectType(objectType).
		SetTitle(content.TitleTemplate).
		SetBody(content.BodyTemplate).
		SetData(content.Data).
		SetTopic(lo.CoalesceOrEmpty(content.Topic, enums.NotificationTopicImportComplete))

	if len(content.Channels) > 0 {
		create = create.SetChannels(content.Channels)
	}

	return create.Exec(ctx)
}

// NotifyImportOutcome sends the organization notification summarizing an import
// subject names what was imported, and detail, such as how to retry, is appended to the body when set
// url is the console path the notification links to when set
func NotifyImportOutcome(ctx context.Context, subject string, result IngestResult, ingestErr error, detail, url string) error {
	if ingestErr != nil {
		logx.FromContext(ctx).Error().Err(ingestErr).Str("subject", subject).Msg("import failed")
	}

	return NotifyOutcome(ctx, importCompletedObjectType, importOutcome(subject, result, ingestErr, detail, url))
}

// importOutcome builds the notification summarizing an import
func importOutcome(subject string, result IngestResult, ingestErr error, detail, url string) entityops.NotificationContent {
	title := subject + " import completed"
	body := fmt.Sprintf("Imported %d record(s)", result.Persisted)

	if ingestErr != nil {
		title = subject + " import failed"
		body = "The import could not be completed"
	}

	if ingestErr == nil && result.Skipped > 0 {
		body += fmt.Sprintf(", %d record(s) matched existing records and were left in place", result.Skipped)
	}

	if ingestErr == nil && result.Failed > 0 {
		body += fmt.Sprintf(", %d record(s) could not be imported", result.Failed)
	}

	if detail != "" {
		body += ", " + detail
	}

	data := map[string]any{"succeeded": result.Succeeded, "persisted": result.Persisted, "skipped": result.Skipped, "failed": result.Failed, "failure_summary": result.FailureSummary()}

	if url != "" {
		data["url"] = url
	}

	return entityops.NotificationContent{
		TitleTemplate: title,
		BodyTemplate:  body,
		Data:          data,
	}
}

// NotifyDeleteOutcome sends the organization notification summarizing a source delete
// subject names what was removed, deleted is the count per schema and deleteErr is the error that stopped it
func NotifyDeleteOutcome(ctx context.Context, subject string, deleted map[string]int, deleteErr error) error {
	if deleteErr != nil {
		logx.FromContext(ctx).Error().Err(deleteErr).Str("subject", subject).Msg("delete failed")
	}

	return NotifyOutcome(ctx, deleteCompletedObjectType, deleteOutcome(subject, deleted, deleteErr))
}

// deleteOutcome builds the notification summarizing a source delete
func deleteOutcome(subject string, deleted map[string]int, deleteErr error) entityops.NotificationContent {
	total := lo.Sum(lo.Values(deleted))

	title := subject + " removed"
	body := fmt.Sprintf("Removed %d record(s)", total)

	if deleteErr != nil {
		title = subject + " removal failed"
		body = fmt.Sprintf("Removed %d record(s) before the removal stopped, run it again to remove the rest", total)
	}

	return entityops.NotificationContent{
		TitleTemplate: title,
		BodyTemplate:  body,
		Data:          map[string]any{"deleted": deleted},
	}
}
