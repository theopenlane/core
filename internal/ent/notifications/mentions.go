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

// noteFields holds the extracted fields from a note mutation for processing mentions.
// A note can be associated with exactly one parent object (task, control, procedure,
// risk, policy, or evidence). The parent relationship fields are mutually exclusive -
// only one will be populated based on where the note was created.
type noteFields struct {
	// text is the plain text content of the note
	text string
	// textJSON is the JSON-serialized Slate content of the note
	textJSON string
	// oldText is the previous plain text content (for update comparisons)
	oldText string
	// oldTextJSON is the previous JSON-serialized content (for update comparisons)
	oldTextJSON string
	// entityID is the unique identifier of the note itself
	entityID string
	// ownerID is the organization owner of the note
	ownerID string
	// taskID is set when the note is a comment on a task
	taskID string
	// controlID is set when the note is a comment on a control
	controlID string
	// procedureID is set when the note is a comment on a procedure
	procedureID string
	// riskID is set when the note is a comment on a risk
	riskID string
	// policyID is set when the note is a comment on an internal policy
	policyID string
	// evidenceID is set when the note is a comment on evidence
	evidenceID string
}

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

	props := ctx.Envelope.Headers.Properties

	fields, err := fetchNoteFields(ctx.Context, client, props, payload)
	if err != nil {
		logx.FromContext(ctx.Context).Error().Err(err).Msg("failed to get note fields")
		return err
	}

	// Determine which text field to use (prefer text_json if available and valid)
	newText := fields.text
	oldText := fields.oldText

	if fields.textJSON != "" && slateparser.IsValidSlateText(fields.textJSON) {
		newText = fields.textJSON
	}

	if fields.oldTextJSON != "" && slateparser.IsValidSlateText(fields.oldTextJSON) {
		oldText = fields.oldTextJSON
	}

	// If no valid text, nothing to process
	if newText == "" {
		return nil
	}

	parentType, parentID, parentName, err := getParentObjectInfo(ctx.Context, client, fields)
	if err != nil {
		logx.FromContext(ctx.Context).Error().Err(err).Msg("failed to get parent object info")
		return err
	}

	newMentions := slateparser.GetNewMentions(oldText, newText, parentType, parentID, parentName)
	if len(newMentions) == 0 {
		return nil
	}

	mentionedOrgMemberIDs := slateparser.ExtractMentionedOrgMemberIDs(newMentions)
	if len(mentionedOrgMemberIDs) == 0 {
		return nil
	}

	allowCtx := privacy.DecisionContext(ctx.Context, privacy.Allow)

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
		ownerID:          fields.ownerID,
		noteID:           fields.entityID,
		isComment:        true,
	}

	if err := addMentionNotification(ctx.Context, client, input); err != nil {
		logx.FromContext(ctx.Context).Error().Err(err).Msg("failed to add mention notification")
		return err
	}

	return nil
}

// fetchNoteFields retrieves note fields from payload metadata, header properties, or DB fallback
func fetchNoteFields(ctx context.Context, client *generated.Client, props map[string]string, payload eventqueue.MutationGalaPayload) (*noteFields, error) {
	fields := &noteFields{}

	extractNoteFromPayload(payload, fields)
	extractNoteFromProps(props, fields)

	if fields.entityID == "" {
		if entityID, ok := eventqueue.MutationEntityID(payload, props); ok {
			fields.entityID = entityID
		}
	}

	if isUpdateOperation(payload.Operation) && fields.oldText == "" && fields.oldTextJSON == "" {
		if err := queryNoteOldText(ctx, client, fields); err != nil {
			logx.FromContext(ctx).Warn().Err(err).Msg("failed to get old note text, treating as create")
		}
	}

	if needsNoteDBQuery(fields) {
		if err := queryNoteFromDB(ctx, client, fields); err != nil {
			return nil, err
		}
	}

	return fields, nil
}

func needsNoteDBQuery(fields *noteFields) bool {
	if fields == nil {
		return true
	}

	missingParent := fields.taskID == "" &&
		fields.controlID == "" &&
		fields.procedureID == "" &&
		fields.riskID == "" &&
		fields.policyID == "" &&
		fields.evidenceID == ""

	return fields.entityID == "" || fields.ownerID == "" || missingParent
}

// extractNoteFromPayload extracts note fields from mutation payload metadata
func extractNoteFromPayload(payload eventqueue.MutationGalaPayload, fields *noteFields) {
	if fields == nil {
		return
	}

	if payload.EntityID != "" {
		fields.entityID = payload.EntityID
	}

	if text, ok := eventqueue.MutationStringValue(payload, note.FieldText); ok {
		fields.text = text
	}

	if raw, ok := eventqueue.MutationValue(payload, note.FieldTextJSON); ok {
		fields.textJSON = jsonValueToString(raw)
	}

	if ownerID, ok := eventqueue.MutationStringValue(payload, note.FieldOwnerID); ok {
		fields.ownerID = ownerID
	}

	if taskID, ok := eventqueue.MutationStringValue(payload, note.TaskColumn); ok {
		fields.taskID = taskID
	}

	if controlID, ok := eventqueue.MutationStringValue(payload, note.ControlColumn); ok {
		fields.controlID = controlID
	}

	if procedureID, ok := eventqueue.MutationStringValue(payload, note.ProcedureColumn); ok {
		fields.procedureID = procedureID
	}

	if riskID, ok := eventqueue.MutationStringValue(payload, note.RiskColumn); ok {
		fields.riskID = riskID
	}

	if policyID, ok := eventqueue.MutationStringValue(payload, note.InternalPolicyColumn); ok {
		fields.policyID = policyID
	}

	if evidenceID, ok := eventqueue.MutationStringValue(payload, note.EvidenceColumn); ok {
		fields.evidenceID = evidenceID
	}
}

// extractNoteFromProps extracts note fields from mutation properties
func extractNoteFromProps(props map[string]string, fields *noteFields) {
	if fields == nil {
		return
	}

	if fields.text == "" {
		fields.text = eventqueue.MutationStringFromProperties(props, note.FieldText)
	}

	if fields.entityID == "" {
		fields.entityID = eventqueue.MutationStringFromProperties(props, note.FieldID)
	}

	if fields.ownerID == "" {
		fields.ownerID = eventqueue.MutationStringFromProperties(props, note.FieldOwnerID)
	}

	if fields.taskID == "" {
		fields.taskID = eventqueue.MutationStringFromProperties(props, note.TaskColumn)
	}

	if fields.controlID == "" {
		fields.controlID = eventqueue.MutationStringFromProperties(props, note.ControlColumn)
	}

	if fields.procedureID == "" {
		fields.procedureID = eventqueue.MutationStringFromProperties(props, note.ProcedureColumn)
	}

	if fields.riskID == "" {
		fields.riskID = eventqueue.MutationStringFromProperties(props, note.RiskColumn)
	}

	if fields.policyID == "" {
		fields.policyID = eventqueue.MutationStringFromProperties(props, note.InternalPolicyColumn)
	}

	if fields.evidenceID == "" {
		fields.evidenceID = eventqueue.MutationStringFromProperties(props, note.EvidenceColumn)
	}
}

// queryNoteOldText queries the database to get old note text before update
func queryNoteOldText(ctx context.Context, client *generated.Client, fields *noteFields) error {
	if fields == nil || fields.entityID == "" {
		return nil
	}

	if client == nil {
		return ErrFailedToGetClient
	}

	allowCtx := privacy.DecisionContext(ctx, privacy.Allow)
	noteEntity, err := client.Note.Get(allowCtx, fields.entityID)
	if err != nil {
		return fmt.Errorf("failed to query note: %w", err)
	}

	fields.oldText = noteEntity.Text
	fields.oldTextJSON = jsonSliceToString(noteEntity.TextJSON)

	return nil
}

// queryNoteFromDB queries note and relationships to fill missing fields
func queryNoteFromDB(ctx context.Context, client *generated.Client, fields *noteFields) error {
	if fields == nil || fields.entityID == "" {
		return ErrEntityIDNotFound
	}

	if client == nil {
		return ErrFailedToGetClient
	}

	allowCtx := privacy.DecisionContext(ctx, privacy.Allow)
	noteEntity, err := client.Note.Query().
		Where(note.ID(fields.entityID)).
		WithTask().
		WithControl().
		WithProcedure().
		WithRisk().
		WithInternalPolicy().
		WithEvidence().
		Only(allowCtx)
	if err != nil {
		return fmt.Errorf("failed to query note with relationships: %w", err)
	}

	if fields.ownerID == "" {
		fields.ownerID = noteEntity.OwnerID
	}

	if noteEntity.Edges.Task != nil {
		fields.taskID = noteEntity.Edges.Task.ID
	}

	if noteEntity.Edges.Control != nil {
		fields.controlID = noteEntity.Edges.Control.ID
	}

	if noteEntity.Edges.Procedure != nil {
		fields.procedureID = noteEntity.Edges.Procedure.ID
	}

	if noteEntity.Edges.Risk != nil {
		fields.riskID = noteEntity.Edges.Risk.ID
	}

	if noteEntity.Edges.InternalPolicy != nil {
		fields.policyID = noteEntity.Edges.InternalPolicy.ID
	}

	if noteEntity.Edges.Evidence != nil {
		fields.evidenceID = noteEntity.Edges.Evidence.ID
	}

	return nil
}

// getParentObjectInfo determines the parent object type and ID for a note
func getParentObjectInfo(ctx context.Context, client *generated.Client, fields *noteFields) (string, string, string, error) {
	if client == nil {
		return "", "", "", ErrFailedToGetClient
	}

	allowCtx := privacy.DecisionContext(ctx, privacy.Allow)

	if fields.taskID != "" {
		taskEntity, err := client.Task.Get(allowCtx, fields.taskID)
		if err != nil {
			return "", "", "", fmt.Errorf("failed to get task: %w", err)
		}
		return generated.TypeTask, fields.taskID, taskEntity.Title, nil
	}

	if fields.controlID != "" {
		control, err := client.Control.Get(allowCtx, fields.controlID)
		if err != nil {
			return "", "", "", fmt.Errorf("failed to get control: %w", err)
		}
		return generated.TypeControl, fields.controlID, control.Title, nil
	}

	if fields.procedureID != "" {
		proc, err := client.Procedure.Get(allowCtx, fields.procedureID)
		if err != nil {
			return "", "", "", fmt.Errorf("failed to get procedure: %w", err)
		}
		return generated.TypeProcedure, fields.procedureID, proc.Name, nil
	}

	if fields.riskID != "" {
		riskEntity, err := client.Risk.Get(allowCtx, fields.riskID)
		if err != nil {
			return "", "", "", fmt.Errorf("failed to get risk: %w", err)
		}
		return generated.TypeRisk, fields.riskID, riskEntity.Name, nil
	}

	if fields.policyID != "" {
		policy, err := client.InternalPolicy.Get(allowCtx, fields.policyID)
		if err != nil {
			return "", "", "", fmt.Errorf("failed to get internal policy: %w", err)
		}
		return generated.TypeInternalPolicy, fields.policyID, policy.Name, nil
	}

	if fields.evidenceID != "" {
		evidence, err := client.Evidence.Get(allowCtx, fields.evidenceID)
		if err != nil {
			return "", "", "", fmt.Errorf("failed to get evidence: %w", err)
		}
		return generated.TypeEvidence, fields.evidenceID, evidence.Name, nil
	}

	return generated.TypeNote, fields.entityID, "Comment", nil
}

// addMentionNotification creates notifications for all mentioned users
func addMentionNotification(ctx context.Context, client *generated.Client, input mentionNotificationInput) error {
	if client == nil {
		return ErrFailedToGetClient
	}

	consoleURL := client.EntConfig.Notifications.ConsoleURL

	url := getURLPathForObject(consoleURL, input.objectID, input.objectType)

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
	valid      bool
}

// oldDocumentDetails holds old document details fetched from the database
type oldDocumentDetails struct {
	name        string
	ownerID     string
	details     string
	detailsJSON []any
}

// handleObjectMentions checks mentions in object details fields (task/risk/procedure/policy).
func handleObjectMentions(ctx gala.HandlerContext, payload eventqueue.MutationGalaPayload) error {
	ctx, client, ok := eventqueue.ClientFromHandler(ctx)
	if !ok {
		return ErrFailedToGetClient
	}

	allowCtx := privacy.DecisionContext(ctx.Context, privacy.Allow)
	props := ctx.Envelope.Headers.Properties

	var details objectMentionDetails

	switch payload.MutationType {
	case generated.TypeTask:
		details = extractTaskMentionDetails(allowCtx, client, payload, props)
	case generated.TypeRisk:
		details = extractDocumentMentionDetails(
			payload,
			props,
			risk.FieldName,
			risk.FieldDetails,
			risk.FieldDetailsJSON,
			risk.FieldOwnerID,
			func(objectID string) (*oldDocumentDetails, error) {
				riskEntity, err := client.Risk.Get(allowCtx, objectID)
				if err != nil {
					return nil, err
				}

				return &oldDocumentDetails{
					name:        riskEntity.Name,
					ownerID:     riskEntity.OwnerID,
					details:     riskEntity.Details,
					detailsJSON: riskEntity.DetailsJSON,
				}, nil
			},
		)
	case generated.TypeProcedure:
		details = extractDocumentMentionDetails(
			payload,
			props,
			procedure.FieldName,
			procedure.FieldDetails,
			procedure.FieldDetailsJSON,
			procedure.FieldOwnerID,
			func(objectID string) (*oldDocumentDetails, error) {
				proc, err := client.Procedure.Get(allowCtx, objectID)
				if err != nil {
					return nil, err
				}

				return &oldDocumentDetails{
					name:        proc.Name,
					ownerID:     proc.OwnerID,
					details:     proc.Details,
					detailsJSON: proc.DetailsJSON,
				}, nil
			},
		)
	case generated.TypeInternalPolicy:
		details = extractDocumentMentionDetails(
			payload,
			props,
			internalpolicy.FieldName,
			internalpolicy.FieldDetails,
			internalpolicy.FieldDetailsJSON,
			internalpolicy.FieldOwnerID,
			func(objectID string) (*oldDocumentDetails, error) {
				policy, err := client.InternalPolicy.Get(allowCtx, objectID)
				if err != nil {
					return nil, err
				}

				return &oldDocumentDetails{
					name:        policy.Name,
					ownerID:     policy.OwnerID,
					details:     policy.Details,
					detailsJSON: policy.DetailsJSON,
				}, nil
			},
		)
	default:
		return nil
	}

	if !details.valid {
		return nil
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

// extractTaskMentionDetails extracts mention details from task payload metadata.
func extractTaskMentionDetails(allowCtx context.Context, client *generated.Client, payload eventqueue.MutationGalaPayload, props map[string]string) objectMentionDetails {
	objectID, _ := eventqueue.MutationEntityID(payload, props)
	details := objectMentionDetails{
		objectID:   objectID,
		objectType: generated.TypeTask,
		valid:      true,
	}

	if raw, ok := eventqueue.MutationValue(payload, task.FieldDetailsJSON); ok {
		details.newDetailsJSON = jsonValueToString(raw)
	}

	details.newDetails = eventqueue.MutationStringValueOrProperty(payload, props, task.FieldDetails)
	details.objectName = eventqueue.MutationStringValueOrProperty(payload, props, task.FieldTitle)
	details.ownerID = eventqueue.MutationStringValueOrProperty(payload, props, task.FieldOwnerID)

	if isUpdateOperation(payload.Operation) && details.objectID != "" {
		taskEntity, err := client.Task.Get(allowCtx, details.objectID)
		if err == nil && taskEntity != nil {
			details.oldDetailsJSON = jsonSliceToString(taskEntity.DetailsJSON)
			details.oldDetails = taskEntity.Details
			if details.objectName == "" {
				details.objectName = taskEntity.Title
			}
			if details.ownerID == "" {
				details.ownerID = taskEntity.OwnerID
			}
		}
	}

	return details
}

// extractDocumentMentionDetails extracts mention details for Risk/Procedure/InternalPolicy.
func extractDocumentMentionDetails(
	payload eventqueue.MutationGalaPayload,
	props map[string]string,
	nameField,
	detailsField,
	detailsJSONField,
	ownerField string,
	queryFunc func(string) (*oldDocumentDetails, error),
) objectMentionDetails {
	objectID, _ := eventqueue.MutationEntityID(payload, props)
	details := objectMentionDetails{
		objectID:   objectID,
		objectType: payload.MutationType,
		valid:      true,
	}

	if raw, ok := eventqueue.MutationValue(payload, detailsJSONField); ok {
		details.newDetailsJSON = jsonValueToString(raw)
	}

	details.newDetails = eventqueue.MutationStringValueOrProperty(payload, props, detailsField)
	details.objectName = eventqueue.MutationStringValueOrProperty(payload, props, nameField)
	details.ownerID = eventqueue.MutationStringValueOrProperty(payload, props, ownerField)

	if isUpdateOperation(payload.Operation) && details.objectID != "" && queryFunc != nil {
		oldDoc, err := queryFunc(details.objectID)
		if err == nil && oldDoc != nil {
			details.oldDetailsJSON = jsonSliceToString(oldDoc.detailsJSON)
			details.oldDetails = oldDoc.details
			if details.objectName == "" {
				details.objectName = oldDoc.name
			}
			if details.ownerID == "" {
				details.ownerID = oldDoc.ownerID
			}
		}
	}

	return details
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
