package notifications

import (
	"github.com/theopenlane/core/common/enums"

	"github.com/theopenlane/core/v2/internal/ent/entityops"
	"github.com/theopenlane/core/v2/internal/ent/generated"
	"github.com/theopenlane/core/v2/internal/ent/generated/task"
)

// handleTaskMutation notifies the proposed assignee using the last active status. status
// may not be included in the mutation but an assignee may change and we need to still notify
// the new assignee
func handleTaskMutation(inv entityops.Invocation, payload entityops.MutationPayload) error {
	if !payload.FieldChanged(task.FieldAssigneeID) {
		return nil
	}

	id, ok := payload.StringValue(task.FieldAssigneeID)
	if !ok {
		return nil
	}

	task, err := inv.Client.Task.Get(inv.Context, payload.EntityID)
	if generated.IsNotFound(err) {
		return nil
	}

	if err != nil {
		return err
	}

	if task.Status != enums.TaskStatusOpen &&
		task.Status != enums.TaskStatusInProgress &&
		task.Status != enums.TaskStatusInReview {
		return nil
	}

	topic := enums.NotificationTopicTaskAssignment
	data := map[string]any{}
	if url := entityops.ConsoleObjectPath(payload.MutationType, payload.EntityID); url != "" {
		data["url"] = url
	}

	input := &generated.CreateNotificationInput{
		NotificationType: enums.NotificationTypeUser,
		Title:            "New task assigned",
		Body:             "Task " + task.Title + " has been assigned to you",
		Data:             data,
		Topic:            &topic,
		ObjectType:       payload.MutationType,
		OwnerID:          &task.OwnerID,
	}

	return entityops.CreateNotifications(inv.Context, inv.Client, []string{id}, input)
}
