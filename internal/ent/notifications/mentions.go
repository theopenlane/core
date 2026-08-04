package notifications

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/stoewer/go-strcase"

	"github.com/theopenlane/core/common/enums"
	"github.com/theopenlane/core/internal/ent/eventqueue"
	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/generated/internalpolicy"
	"github.com/theopenlane/core/internal/ent/generated/note"
	"github.com/theopenlane/core/internal/ent/generated/orgmembership"
	"github.com/theopenlane/core/internal/ent/generated/privacy"
	"github.com/theopenlane/core/internal/ent/generated/procedure"
	"github.com/theopenlane/core/internal/ent/generated/risk"
	"github.com/theopenlane/core/internal/ent/generated/task"
	"github.com/theopenlane/core/pkg/gala"
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

// handleNoteMutation processes note mutations and creates notifications for mentioned users
func handleNoteMutation(ctx gala.HandlerContext, payload eventqueue.MutationGalaPayload) error {
	ctx, client, ok := eventqueue.ClientFromHandler(ctx)
	if !ok {
		return ErrFailedToGetClient
	}

	// Determine which text field to use (prefer text_json if available and valid)
	newText, _ := eventqueue.MutationStringValue(payload, note.FieldText)

	if raw, ok := eventqueue.MutationValue(payload, note.FieldTextJSON); ok {
		if textJSON := jsonValueToString(raw); textJSON != "" && slateparser.IsValidSlateText(textJSON) {
			newText = textJSON
		}
	}

	// If no valid text, nothing to process
	if newText == "" {
		return nil
	}

	allowCtx := privacy.DecisionContext(ctx.Context, privacy.Allow)

	noteEntity, err := client.Note.Query().
		Where(note.ID(payload.EntityID)).
		WithTask().
		WithControl().
		WithProcedure().
		WithRisk().
		WithInternalPolicy().
		WithEvidence().
		Only(allowCtx)
	switch {
	case generated.IsNotFound(err):
		return nil
	case err != nil:
		logx.FromContext(ctx.Context).Error().Err(err).Msg("failed to query note with relationships")
		return err
	}

	oldText, _ := eventqueue.MutationOldStringValue(payload, note.FieldText)

	if raw, ok := eventqueue.MutationOldValue(payload, note.FieldTextJSON); ok {
		if oldTextJSON := jsonValueToString(raw); oldTextJSON != "" && slateparser.IsValidSlateText(oldTextJSON) {
			oldText = oldTextJSON
		}
	}

	parentType, parentID, parentName := noteParent(noteEntity)

	newMentions := slateparser.GetNewMentions(oldText, newText, parentType, parentID, parentName)
	if len(newMentions) == 0 {
		return nil
	}

	mentionedOrgMemberIDs := slateparser.ExtractMentionedOrgMemberIDs(newMentions)
	if len(mentionedOrgMemberIDs) == 0 {
		return nil
	}

	userIDs, err := client.OrgMembership.Query().
		Where(orgmembership.IDIn(mentionedOrgMemberIDs...)).
		Select(orgmembership.FieldUserID).
		Strings(allowCtx)
	if err != nil {
		logx.FromContext(ctx.Context).Error().Err(err).Msg("failed to get user IDs from org membership IDs")
		return err
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

	if err := addMentionNotification(ctx.Context, client, input); err != nil {
		logx.FromContext(ctx.Context).Error().Err(err).Msg("failed to add mention notification")
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
	url := getURLPathForObject(input.objectID, input.objectType)

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
	objectID       string
	objectType     string
	objectName     string
	ownerID        string
	newDetailsJSON string
	oldDetailsJSON string
	// newDetails and oldDetails are plain text fallbacks when JSON fields are empty.
	newDetails string
	oldDetails string
}

// handleObjectMentions checks mentions in object details fields (task/risk/procedure/policy).
func handleObjectMentions(ctx gala.HandlerContext, payload eventqueue.MutationGalaPayload) error {
	ctx, client, ok := eventqueue.ClientFromHandler(ctx)
	if !ok {
		return ErrFailedToGetClient
	}

	allowCtx := privacy.DecisionContext(ctx.Context, privacy.Allow)

	var details objectMentionDetails

	switch payload.MutationType {
	case generated.TypeTask:
		details = extractMentionDetails(payload, task.FieldTitle, task.FieldDetails, task.FieldDetailsJSON, task.FieldOwnerID)
	case generated.TypeRisk:
		details = extractMentionDetails(payload, risk.FieldName, risk.FieldDetails, risk.FieldDetailsJSON, risk.FieldOwnerID)
	case generated.TypeProcedure:
		details = extractMentionDetails(payload, procedure.FieldName, procedure.FieldDetails, procedure.FieldDetailsJSON, procedure.FieldOwnerID)
	case generated.TypeInternalPolicy:
		details = extractMentionDetails(payload, internalpolicy.FieldName, internalpolicy.FieldDetails, internalpolicy.FieldDetailsJSON, internalpolicy.FieldOwnerID)
	default:
		return nil
	}

	// name and owner are absent from the payload when unchanged by the mutation; fill
	// them from the current row before mention context is built
	if details.objectName == "" || details.ownerID == "" {
		name, ownerID, err := objectNameAndOwner(allowCtx, client, payload.MutationType, payload.EntityID)
		switch {
		case generated.IsNotFound(err):
			return nil
		case err != nil:
			logx.FromContext(ctx.Context).Error().Err(err).Msg("failed to query mention object")
			return err
		}

		if details.objectName == "" {
			details.objectName = name
		}

		if details.ownerID == "" {
			details.ownerID = ownerID
		}
	}

	// Prefer JSON; fall back to plain text when needed.
	newText := details.newDetailsJSON
	oldText := details.oldDetailsJSON

	if newText == "" {
		newText = details.newDetails
		oldText = details.oldDetails
	}

	if newText == "" || !slateparser.IsValidSlateText(newText) {
		return nil
	}

	newMentions := slateparser.GetNewMentions(oldText, newText, details.objectType, details.objectID, details.objectName)
	if len(newMentions) == 0 {
		return nil
	}

	mentionedOrgMemberIDs := slateparser.ExtractMentionedOrgMemberIDs(newMentions)
	if len(mentionedOrgMemberIDs) == 0 {
		return nil
	}

	userIDs, err := client.OrgMembership.Query().
		Where(orgmembership.IDIn(mentionedOrgMemberIDs...)).
		Select(orgmembership.FieldUserID).
		Strings(allowCtx)
	if err != nil {
		logx.FromContext(ctx.Context).Error().Err(err).Msg("failed to get user IDs from org membership IDs")
		return err
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

	if err := addMentionNotification(ctx.Context, client, input); err != nil {
		logx.FromContext(ctx.Context).Error().Err(err).Msg("failed to add mention notification for object")
		return err
	}

	return nil
}

// extractMentionDetails extracts new and pre-update mention detail values from the mutation payload
func extractMentionDetails(payload eventqueue.MutationGalaPayload, nameField, detailsField, detailsJSONField, ownerField string) objectMentionDetails {
	details := objectMentionDetails{
		objectID:   payload.EntityID,
		objectType: payload.MutationType,
	}

	if raw, ok := eventqueue.MutationValue(payload, detailsJSONField); ok {
		details.newDetailsJSON = jsonValueToString(raw)
	}

	if raw, ok := eventqueue.MutationOldValue(payload, detailsJSONField); ok {
		details.oldDetailsJSON = jsonValueToString(raw)
	}

	details.newDetails, _ = eventqueue.MutationStringValue(payload, detailsField)
	details.oldDetails, _ = eventqueue.MutationOldStringValue(payload, detailsField)
	details.objectName, _ = eventqueue.MutationStringValue(payload, nameField)
	details.ownerID, _ = eventqueue.MutationStringValue(payload, ownerField)

	return details
}

// objectNameAndOwner returns the display name and owner of a mention-eligible object from its current row
func objectNameAndOwner(ctx context.Context, client *generated.Client, mutationType, entityID string) (string, string, error) {
	switch mutationType {
	case generated.TypeTask:
		taskEntity, err := client.Task.Get(ctx, entityID)
		if err != nil {
			return "", "", err
		}

		return taskEntity.Title, taskEntity.OwnerID, nil
	case generated.TypeRisk:
		riskEntity, err := client.Risk.Get(ctx, entityID)
		if err != nil {
			return "", "", err
		}

		return riskEntity.Name, riskEntity.OwnerID, nil
	case generated.TypeProcedure:
		proc, err := client.Procedure.Get(ctx, entityID)
		if err != nil {
			return "", "", err
		}

		return proc.Name, proc.OwnerID, nil
	case generated.TypeInternalPolicy:
		policy, err := client.InternalPolicy.Get(ctx, entityID)
		if err != nil {
			return "", "", err
		}

		return policy.Name, policy.OwnerID, nil
	default:
		return "", "", nil
	}
}

// jsonSliceToString converts a []any slice to a JSON string for parsing.
// Returns an empty string if the slice is empty or serialization fails.
func jsonSliceToString(data []any) string {
	if len(data) == 0 {
		return ""
	}

	bytes, err := json.Marshal(data)
	if err != nil {
		return ""
	}

	return string(bytes)
}

func jsonValueToString(raw any) string {
	if raw == nil {
		return ""
	}

	switch value := raw.(type) {
	case string:
		return value
	case []any:
		return jsonSliceToString(value)
	case []string:
		if len(value) == 0 {
			return ""
		}
	default:
	}

	bytes, err := json.Marshal(raw)
	if err != nil || len(bytes) == 0 || string(bytes) == "null" {
		return ""
	}

	return string(bytes)
}
