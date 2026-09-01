package notifications

import (
	"context"
	"fmt"

	"github.com/samber/lo"
	"github.com/stoewer/go-strcase"

	"github.com/theopenlane/core/common/enums"
	"github.com/theopenlane/core/v2/internal/ent/entityops"
	"github.com/theopenlane/core/v2/internal/ent/generated"
	"github.com/theopenlane/core/v2/internal/ent/generated/note"
	"github.com/theopenlane/logx"
)

// mentionNotificationInput carries all data required to create notifications
// for users mentioned in an object (for example, a note or task).
type mentionNotificationInput struct {
	// mentionedUserIDs contains the IDs of users that were mentioned and should receive a notification
	mentionedUserIDs []string
	// objectType describes the type of entity where the mention occurred
	objectType string
	// objectID is the identifier of the entity where the mention occurred
	objectID string
	// objectName is a human-readable name/title used in notification content
	objectName string
	// objectLabel is the human-readable schema label used in notification content
	objectLabel string
	// ownerID is the ID of the user who created the mention
	ownerID string
	// noteID is the ID of the note associated with the mention, when applicable
	noteID string
}

// handleNoteMutation processes note mutations and creates notifications for mentioned users
func handleNoteMutation(inv entityops.Invocation, payload entityops.MutationPayload) error {
	spec, ok := entityops.MentionSpecFor(generated.TypeNote)
	if !ok {
		return nil
	}

	newText, oldText := entityops.MentionTexts(payload, spec)

	// If no valid text, nothing to process
	if newText == "" {
		return nil
	}

	noteEntity, err := inv.Client.Note.Query().
		Where(note.ID(payload.EntityID)).
		WithTask().
		WithControl().
		WithProcedure().
		WithRisk().
		WithInternalPolicy().
		WithEvidence().
		Only(inv.Context)
	switch {
	case generated.IsNotFound(err):
		return nil
	case err != nil:
		logx.FromContext(inv.Context).Error().Err(err).Msg("failed to query note with relationships")
		return err
	}

	parentType, parentID, parentName := noteParent(noteEntity)

	parentLabel := ""
	if parentSchema, ok := entityops.LookupSchema(parentType); ok {
		parentLabel = parentSchema.Label()
	}

	userIDs, err := entityops.MentionedUsers(inv.Context, inv.Client, oldText, newText, parentType, parentID, parentName)
	if err != nil {
		return err
	}

	if len(userIDs) == 0 {
		return nil
	}

	input := mentionNotificationInput{
		objectType:       parentType,
		mentionedUserIDs: userIDs,
		objectID:         parentID,
		objectName:       parentName,
		objectLabel:      parentLabel,
		ownerID:          noteEntity.OwnerID,
		noteID:           noteEntity.ID,
	}

	if err := addMentionNotification(inv.Context, inv.Client, input); err != nil {
		logx.FromContext(inv.Context).Error().Err(err).Msg("failed to add mention notification")
		return err
	}

	return nil
}

// noteParent resolves the parent object type, ID, and display name from the note's
// loaded edges; a note is associated with at most one parent object
func noteParent(noteEntity *generated.Note) (string, string, string) {
	edges := noteEntity.Edges

	switch {
	case edges.Task != nil:
		return generated.TypeTask, edges.Task.ID, edges.Task.Title
	case edges.Control != nil:
		return generated.TypeControl, edges.Control.ID, edges.Control.Title
	case edges.Procedure != nil:
		return generated.TypeProcedure, edges.Procedure.ID, edges.Procedure.Name
	case edges.Risk != nil:
		return generated.TypeRisk, edges.Risk.ID, edges.Risk.Name
	case edges.InternalPolicy != nil:
		return generated.TypeInternalPolicy, edges.InternalPolicy.ID, edges.InternalPolicy.Name
	case edges.Evidence != nil:
		return generated.TypeEvidence, edges.Evidence.ID, edges.Evidence.Name
	default:
		return generated.TypeNote, noteEntity.ID, "Comment"
	}
}

// addMentionNotification creates notifications for all mentioned users
func addMentionNotification(ctx context.Context, client *generated.Client, input mentionNotificationInput) error {
	url := entityops.ConsoleObjectPath(input.objectType, input.objectID)

	dataMap := map[string]any{
		"object_type": strcase.UpperSnakeCase(input.objectType),
		"object_id":   input.objectID,
		"object_name": input.objectName,
		"note_id":     input.noteID,
	}

	if url != "" {
		dataMap["url"] = url
	}

	topic := enums.NotificationTopicMention

	body := fmt.Sprintf("You were mentioned in a comment on %s: %s", input.objectLabel, input.objectName)

	notifInput := &generated.CreateNotificationInput{
		NotificationType: enums.NotificationTypeUser,
		Title:            "You were mentioned",
		Body:             body,
		Data:             dataMap,
		OwnerID:          &input.ownerID,
		Topic:            &topic,
		ObjectType:       input.objectType,
	}

	filteredMentionedUserIDs := lo.Filter(input.mentionedUserIDs, func(id string, _ int) bool {
		return id != input.ownerID
	})

	if len(filteredMentionedUserIDs) == 0 {
		return nil
	}

	return entityops.CreateNotifications(ctx, client, filteredMentionedUserIDs, notifInput)
}
