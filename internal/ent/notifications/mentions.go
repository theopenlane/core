package notifications

import (
	"context"
	"fmt"

	"github.com/stoewer/go-strcase"

	"github.com/theopenlane/core/common/enums"
	"github.com/theopenlane/core/internal/ent/entityops"
	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/generated/note"
	"github.com/theopenlane/core/internal/ent/generated/orgmembership"
	"github.com/theopenlane/core/pkg/jsonx"
	"github.com/theopenlane/core/pkg/logx"
	"github.com/theopenlane/core/pkg/slateparser"
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
	// ownerID is the ID of the user who created the mention
	ownerID string
	// noteID is the ID of the note associated with the mention, when applicable
	noteID string
	// isComment indicates whether the mention occurred within a comment/note
	isComment bool
}

// mentionTexts resolves the new and pre-mutation rich text scanned for mentions, preferring
// each JSON candidate over its plain-text fallback only when it is valid slate text
func mentionTexts(payload entityops.MutationPayload, spec entityops.MentionSpec) (newText, oldText string) {
	newText, _ = payload.StringValue(spec.DetailsField)

	if raw, ok := payload.Value(spec.DetailsJSONField); ok {
		if candidate := jsonx.Stringify(raw); candidate != "" && slateparser.IsValidSlateText(candidate) {
			newText = candidate
		}
	}

	oldText, _ = payload.OldStringValue(spec.DetailsField)

	if raw, ok := payload.OldValue(spec.DetailsJSONField); ok {
		if candidate := jsonx.Stringify(raw); candidate != "" && slateparser.IsValidSlateText(candidate) {
			oldText = candidate
		}
	}

	return newText, oldText
}

// mentionedUserIDs resolves the user ids newly mentioned between the old and new rich text
// for an object
func mentionedUserIDs(ctx context.Context, client *generated.Client, oldText, newText, objectType, objectID, objectName string) ([]string, error) {
	newMentions := slateparser.GetNewMentions(oldText, newText, objectType, objectID, objectName)
	if len(newMentions) == 0 {
		return nil, nil
	}

	mentionedOrgMemberIDs := slateparser.ExtractMentionedOrgMemberIDs(newMentions)
	if len(mentionedOrgMemberIDs) == 0 {
		return nil, nil
	}

	userIDs, err := client.OrgMembership.Query().
		Where(orgmembership.IDIn(mentionedOrgMemberIDs...)).
		Select(orgmembership.FieldUserID).
		Strings(ctx)
	if err != nil {
		logx.FromContext(ctx).Error().Err(err).Msg("failed to get user IDs from org membership IDs")
		return nil, err
	}

	return userIDs, nil
}

// handleNoteMutation processes note mutations and creates notifications for mentioned users
func handleNoteMutation(inv entityops.Invocation, payload entityops.MutationPayload) error {
	spec, ok := entityops.MentionSpecFor(generated.TypeNote)
	if !ok {
		return nil
	}

	newText, oldText := mentionTexts(payload, spec)

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

	userIDs, err := mentionedUserIDs(inv.Context, inv.Client, oldText, newText, parentType, parentID, parentName)
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
		ownerID:          noteEntity.OwnerID,
		noteID:           noteEntity.ID,
		isComment:        true,
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

	var body string
	if input.isComment {
		body = fmt.Sprintf("You were mentioned in a comment on %s: %s", input.objectType, input.objectName)
	} else {
		body = fmt.Sprintf("You were mentioned in %s: %s", input.objectType, input.objectName)
	}

	notifInput := &generated.CreateNotificationInput{
		NotificationType: enums.NotificationTypeUser,
		Title:            "You were mentioned",
		Body:             body,
		Data:             dataMap,
		OwnerID:          &input.ownerID,
		Topic:            &topic,
		ObjectType:       input.objectType,
	}

	filteredMentionedUserIDs := make([]string, 0, len(input.mentionedUserIDs))
	for _, id := range input.mentionedUserIDs {
		if id != input.ownerID {
			filteredMentionedUserIDs = append(filteredMentionedUserIDs, id)
		}
	}

	if len(filteredMentionedUserIDs) == 0 {
		return nil
	}

	return newNotificationCreation(ctx, client, filteredMentionedUserIDs, notifInput)
}

// objectMentionDetails holds extracted details for mention processing
type objectMentionDetails struct {
	objectID   string
	objectType string
	objectName string
	ownerID    string
}

// handleObjectMentions checks mentions in object details fields for every mention-eligible schema
func handleObjectMentions(inv entityops.Invocation, payload entityops.MutationPayload) error {
	spec, ok := entityops.MentionSpecFor(payload.MutationType)
	if !ok {
		return nil
	}

	details := extractMentionDetails(payload, spec)

	// name and owner are absent from the payload when unchanged by the mutation; fill
	// them from the current row before mention context is built
	if details.objectName == "" || details.ownerID == "" {
		row, err := inv.Schema.Load(inv.Context, inv.Client, payload.EntityID)
		switch {
		case generated.IsNotFound(err):
			return nil
		case err != nil:
			logx.FromContext(inv.Context).Error().Err(err).Msg("failed to load mention object")
			return err
		}

		if details.objectName == "" {
			details.objectName = inv.Schema.DisplayValue(row)
		}

		if details.ownerID == "" {
			fields, err := jsonx.Decode[map[string]any](row)
			if err != nil {
				logx.FromContext(inv.Context).Error().Err(err).Msg("failed to decode mention object row")
				return err
			}

			details.ownerID, _ = fields[spec.OwnerField].(string)
		}
	}

	newText, oldText := mentionTexts(payload, spec)
	if newText == "" {
		return nil
	}

	userIDs, err := mentionedUserIDs(inv.Context, inv.Client, oldText, newText, details.objectType, details.objectID, details.objectName)
	if err != nil {
		return err
	}

	if len(userIDs) == 0 {
		return nil
	}

	input := mentionNotificationInput{
		mentionedUserIDs: userIDs,
		objectType:       details.objectType,
		objectID:         details.objectID,
		objectName:       details.objectName,
		ownerID:          details.ownerID,
		noteID:           "",
		isComment:        false,
	}

	if err := addMentionNotification(inv.Context, inv.Client, input); err != nil {
		logx.FromContext(inv.Context).Error().Err(err).Msg("failed to add mention notification for object")
		return err
	}

	return nil
}

// extractMentionDetails extracts mention identity values from the mutation payload using
// the schema's mention-scan field spec
func extractMentionDetails(payload entityops.MutationPayload, spec entityops.MentionSpec) objectMentionDetails {
	details := objectMentionDetails{
		objectID:   payload.EntityID,
		objectType: payload.MutationType,
	}

	details.objectName, _ = payload.StringValue(spec.NameField)
	details.ownerID, _ = payload.StringValue(spec.OwnerField)

	return details
}
