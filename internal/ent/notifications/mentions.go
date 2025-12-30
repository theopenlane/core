package notifications

import (
	"context"
	"fmt"

	"github.com/theopenlane/core/internal/ent/events"
	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/generated/note"
	"github.com/theopenlane/core/internal/ent/generated/privacy"
	"github.com/theopenlane/core/pkg/enums"
	"github.com/theopenlane/core/pkg/events/soiree"
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
	// mentionedUserIDs contains the IDs of users that were mentioned and
	// should receive a notification.
	mentionedUserIDs []string
	// objectType describes the type of entity where the mention occurred
	// (for example, "note", "task", or another domain object).
	objectType string
	// objectID is the identifier of the entity where the mention occurred.
	objectID string
	// objectName is a human-readable name or title of the mentioned object,
	// used in notification content.
	objectName string
	// ownerID is the ID of the user who created the mention.
	ownerID string
	// noteID is the ID of the note associated with the mention, when
	// applicable. It is optional and may be empty when the mention is not
	// tied to a specific note or when that context is unavailable.
	noteID string
}

// handleNoteMutation processes note mutations and creates notifications for mentioned users
func handleNoteMutation(ctx *soiree.EventContext, payload *events.MutationPayload) error {
	props := ctx.Properties()
	if props == nil {
		return nil
	}

	// Get note fields from props and payload
	fields, err := fetchNoteFields(ctx, props, payload)
	if err != nil {
		logx.FromContext(ctx.Context()).Error().Err(err).Msg("failed to get note fields")
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

	// Determine parent object type and ID
	parentType, parentID, parentName, err := getParentObjectInfo(ctx, fields)
	if err != nil {
		logx.FromContext(ctx.Context()).Error().Err(err).Msg("failed to get parent object info")
		return err
	}

	// Check for new mentions
	newMentions := slateparser.GetNewMentions(oldText, newText, parentType, parentID, parentName)
	if len(newMentions) == 0 {
		return nil
	}

	// Extract unique user IDs from mentions
	mentionedUserIDs := slateparser.ExtractMentionedUserIDs(newMentions)
	if len(mentionedUserIDs) == 0 {
		return nil
	}

	input := mentionNotificationInput{
		mentionedUserIDs: mentionedUserIDs,
		objectType:       parentType,
		objectID:         parentID,
		objectName:       parentName,
		ownerID:          fields.ownerID,
		noteID:           fields.entityID,
	}

	if err := addMentionNotification(ctx, input); err != nil {
		logx.FromContext(ctx.Context()).Error().Err(err).Msg("failed to add mention notification")
		return err
	}

	return nil
}

// fetchNoteFields retrieves note fields from payload, props, or queries database if missing
func fetchNoteFields(ctx *soiree.EventContext, props soiree.Properties, payload *events.MutationPayload) (*noteFields, error) {
	fields := &noteFields{}

	extractNoteFromPayload(payload, fields)
	extractNoteFromProps(props, fields)

	// Query database if we need the old text for comparison (on update operations)
	if payload.Operation == "UpdateOne" || payload.Operation == "Update" {
		if fields.oldText == "" && fields.oldTextJSON == "" {
			if err := queryNoteOldText(ctx, fields); err != nil {
				logx.FromContext(ctx.Context()).Warn().Err(err).Msg("failed to get old note text, treating as create")
			}
		}
	}

	// If we don't have the entity ID yet, we need it to query parent relationships
	if fields.entityID == "" || fields.ownerID == "" {
		if err := queryNoteFromDB(ctx, fields); err != nil {
			return nil, err
		}
	}

	return fields, nil
}

// extractNoteFromPayload extracts note fields from mutation payload
func extractNoteFromPayload(payload *events.MutationPayload, fields *noteFields) {
	if payload == nil {
		return
	}

	if payload.EntityID != "" {
		fields.entityID = payload.EntityID
	}

	noteMut, ok := payload.Mutation.(*generated.NoteMutation)
	if !ok {
		return
	}

	if text, exists := noteMut.Text(); exists {
		fields.text = text
	}

	if textJSON, exists := noteMut.TextJSON(); exists {
		fields.textJSON = jsonSliceToString(textJSON)
	}

	if ownerID, exists := noteMut.OwnerID(); exists {
		fields.ownerID = ownerID
	}

	// Get parent object IDs if they exist
	if taskID, exists := noteMut.TaskID(); exists {
		fields.taskID = taskID
	}

	if controlID, exists := noteMut.ControlID(); exists {
		fields.controlID = controlID
	}

	if procedureID, exists := noteMut.ProcedureID(); exists {
		fields.procedureID = procedureID
	}

	if riskID, exists := noteMut.RiskID(); exists {
		fields.riskID = riskID
	}

	if policyID, exists := noteMut.InternalPolicyID(); exists {
		fields.policyID = policyID
	}

	if evidenceID, exists := noteMut.EvidenceID(); exists {
		fields.evidenceID = evidenceID
	}
}

// extractNoteFromProps extracts note fields from properties
func extractNoteFromProps(props soiree.Properties, fields *noteFields) {
	if fields.text == "" {
		if text, ok := props.GetKey(note.FieldText).(string); ok {
			fields.text = text
		}
	}

	if fields.entityID == "" {
		if id, ok := props.GetKey(note.FieldID).(string); ok {
			fields.entityID = id
		}
	}

	if fields.ownerID == "" {
		if ownerID, ok := props.GetKey(note.FieldOwnerID).(string); ok {
			fields.ownerID = ownerID
		}
	}
}

// queryNoteOldText queries the database to get the old text before update
func queryNoteOldText(ctx *soiree.EventContext, fields *noteFields) error {
	if fields.entityID == "" {
		return nil
	}

	client, ok := soiree.ClientAs[*generated.Client](ctx)
	if !ok {
		return ErrFailedToGetClient
	}

	allowCtx := privacy.DecisionContext(ctx.Context(), privacy.Allow)
	noteEntity, err := client.Note.Get(allowCtx, fields.entityID)
	if err != nil {
		return fmt.Errorf("failed to query note: %w", err)
	}

	fields.oldText = noteEntity.Text

	fields.oldTextJSON = jsonSliceToString(noteEntity.TextJSON)

	return nil
}

// queryNoteFromDB queries note from database to fill missing fields
func queryNoteFromDB(ctx *soiree.EventContext, fields *noteFields) error {
	if fields.entityID == "" {
		return ErrEntityIDNotFound
	}

	client, ok := soiree.ClientAs[*generated.Client](ctx)
	if !ok {
		return ErrFailedToGetClient
	}

	allowCtx := privacy.DecisionContext(ctx.Context(), privacy.Allow)
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

	// Get parent object IDs from edges
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
func getParentObjectInfo(ctx *soiree.EventContext, fields *noteFields) (string, string, string, error) {
	client, ok := soiree.ClientAs[*generated.Client](ctx)
	if !ok {
		return "", "", "", ErrFailedToGetClient
	}

	allowCtx := privacy.DecisionContext(ctx.Context(), privacy.Allow)

	// Check each possible parent type and return the first one found
	if fields.taskID != "" {
		task, err := client.Task.Get(allowCtx, fields.taskID)
		if err != nil {
			return "", "", "", fmt.Errorf("failed to get task: %w", err)
		}
		return "Task", fields.taskID, task.Title, nil
	}

	if fields.controlID != "" {
		control, err := client.Control.Get(allowCtx, fields.controlID)
		if err != nil {
			return "", "", "", fmt.Errorf("failed to get control: %w", err)
		}
		return "Control", fields.controlID, control.Title, nil
	}

	if fields.procedureID != "" {
		procedure, err := client.Procedure.Get(allowCtx, fields.procedureID)
		if err != nil {
			return "", "", "", fmt.Errorf("failed to get procedure: %w", err)
		}
		return "Procedure", fields.procedureID, procedure.Name, nil
	}

	if fields.riskID != "" {
		risk, err := client.Risk.Get(allowCtx, fields.riskID)
		if err != nil {
			return "", "", "", fmt.Errorf("failed to get risk: %w", err)
		}
		return "Risk", fields.riskID, risk.Name, nil
	}

	if fields.policyID != "" {
		policy, err := client.InternalPolicy.Get(allowCtx, fields.policyID)
		if err != nil {
			return "", "", "", fmt.Errorf("failed to get internal policy: %w", err)
		}
		return "InternalPolicy", fields.policyID, policy.Name, nil
	}

	if fields.evidenceID != "" {
		evidence, err := client.Evidence.Get(allowCtx, fields.evidenceID)
		if err != nil {
			return "", "", "", fmt.Errorf("failed to get evidence: %w", err)
		}
		return "Evidence", fields.evidenceID, evidence.Name, nil
	}

	// If no parent found, return Note as the object type
	return "Note", fields.entityID, "Comment", nil
}

// addMentionNotification creates notifications for all mentioned users
func addMentionNotification(ctx *soiree.EventContext, input mentionNotificationInput) error {
	client, ok := soiree.ClientAs[*generated.Client](ctx)
	if !ok {
		return nil
	}

	consoleURL := client.EntConfig.Notifications.ConsoleURL

	// Build URL based on object type
	var url string
	switch input.objectType {
	case "Task":
		url = fmt.Sprintf("%s/tasks?id=%s", consoleURL, input.objectID)
	case "Control":
		url = fmt.Sprintf("%s/controls/%s", consoleURL, input.objectID)
	case "Procedure":
		url = fmt.Sprintf("%s/procedures/%s", consoleURL, input.objectID)
	case "Risk":
		url = fmt.Sprintf("%s/risks/%s", consoleURL, input.objectID)
	case "InternalPolicy":
		url = fmt.Sprintf("%s/policies/%s", consoleURL, input.objectID)
	case "Evidence":
		url = fmt.Sprintf("%s/evidence/?id=%s", consoleURL, input.objectID)
		// default: no URL if we don't have a known parent object type
	}

	// Create the data map with context
	dataMap := map[string]any{
		"object_type": input.objectType,
		"object_id":   input.objectID,
		"object_name": input.objectName,
		"note_id":     input.noteID,
	}

	// Only include URL if we have a valid one
	if url != "" {
		dataMap["url"] = url
	}

	topic := "mention_alert"
	notifInput := &generated.CreateNotificationInput{
		NotificationType: enums.NotificationTypeUser,
		Title:            "You were mentioned",
		Body:             fmt.Sprintf("You were mentioned in a comment on %s: %s", input.objectType, input.objectName),
		Data:             dataMap,
		OwnerID:          &input.ownerID,
		Topic:            &topic,
		ObjectType:       input.objectType,
	}

	// Filter out the ownerID to avoid sending self-mention notifications
	filteredMentionedUserIDs := make([]string, 0, len(input.mentionedUserIDs))
	for _, id := range input.mentionedUserIDs {
		if id != input.ownerID {
			filteredMentionedUserIDs = append(filteredMentionedUserIDs, id)
		}
	}

	if len(filteredMentionedUserIDs) == 0 {
		return nil
	}

	return newNotificationCreation(ctx, filteredMentionedUserIDs, notifInput)
}

// objectMentionDetails holds the extracted details for mention processing
type objectMentionDetails struct {
	objectID       string
	objectName     string
	ownerID        string
	newDetailsJSON string
	oldDetailsJSON string
	valid          bool
}

// documentMutation is an interface for mutations that have Name, DetailsJSON, and OwnerID fields.
// This covers Risk, Procedure, and InternalPolicy mutations which share the same field structure.
type documentMutation interface {
	Name() (r string, exists bool)
	DetailsJSON() (r []any, exists bool)
	OwnerID() (r string, exists bool)
}

// oldDocumentDetails holds the old document details fetched from the database
type oldDocumentDetails struct {
	name        string
	ownerID     string
	detailsJSON []any
}

// handleObjectMentions is a generic handler for checking mentions in object details fields.
// It uses type switching on the mutation to handle different object types.
func handleObjectMentions(ctx *soiree.EventContext, payload *events.MutationPayload, objectType string) error {
	if payload == nil {
		return nil
	}

	client, ok := soiree.ClientAs[*generated.Client](ctx)
	if !ok {
		return ErrFailedToGetClient
	}

	allowCtx := privacy.DecisionContext(ctx.Context(), privacy.Allow)

	var details objectMentionDetails

	// Use type switch for cleaner, more type-safe handling
	switch mut := payload.Mutation.(type) {
	case *generated.TaskMutation:
		// Task is special - it has Title instead of Name
		details = extractTaskMentionDetails(client, allowCtx, payload, mut)
		objectType = "Task"
	case *generated.RiskMutation:
		details = extractDocumentMentionDetails(payload, mut, func() (*oldDocumentDetails, error) {
			risk, err := client.Risk.Get(allowCtx, payload.EntityID)
			if err != nil {
				return nil, err
			}
			return &oldDocumentDetails{name: risk.Name, ownerID: risk.OwnerID, detailsJSON: risk.DetailsJSON}, nil
		})
		objectType = "Risk"
	case *generated.ProcedureMutation:
		details = extractDocumentMentionDetails(payload, mut, func() (*oldDocumentDetails, error) {
			proc, err := client.Procedure.Get(allowCtx, payload.EntityID)
			if err != nil {
				return nil, err
			}
			return &oldDocumentDetails{name: proc.Name, ownerID: proc.OwnerID, detailsJSON: proc.DetailsJSON}, nil
		})
		objectType = "Procedure"
	case *generated.InternalPolicyMutation:
		details = extractDocumentMentionDetails(payload, mut, func() (*oldDocumentDetails, error) {
			policy, err := client.InternalPolicy.Get(allowCtx, payload.EntityID)
			if err != nil {
				return nil, err
			}
			return &oldDocumentDetails{name: policy.Name, ownerID: policy.OwnerID, detailsJSON: policy.DetailsJSON}, nil
		})
		objectType = "InternalPolicy"
	default:
		return nil
	}

	if !details.valid {
		return nil
	}

	// If no new details JSON, nothing to check
	if details.newDetailsJSON == "" {
		return nil
	}

	// Check if the new details contain valid Slate text
	if !slateparser.IsValidSlateText(details.newDetailsJSON) {
		return nil
	}

	// Check for new mentions
	newMentions := slateparser.GetNewMentions(details.oldDetailsJSON, details.newDetailsJSON, objectType, details.objectID, details.objectName)
	if len(newMentions) == 0 {
		return nil
	}

	// Extract unique user IDs from mentions
	mentionedUserIDs := slateparser.ExtractMentionedUserIDs(newMentions)
	if len(mentionedUserIDs) == 0 {
		return nil
	}

	// Create mention notifications
	input := mentionNotificationInput{
		mentionedUserIDs: mentionedUserIDs,
		objectType:       objectType,
		objectID:         details.objectID,
		objectName:       details.objectName,
		ownerID:          details.ownerID,
		noteID:           "", // No note ID for direct object mentions
	}

	if err := addMentionNotification(ctx, input); err != nil {
		logx.FromContext(ctx.Context()).Error().Err(err).Msg("failed to add mention notification for object")
		return err
	}

	return nil
}

// extractTaskMentionDetails extracts mention details from a task mutation.
// Task is handled separately because it has Title instead of Name.
func extractTaskMentionDetails(client *generated.Client, allowCtx context.Context, payload *events.MutationPayload, taskMut *generated.TaskMutation) objectMentionDetails {
	details := objectMentionDetails{
		objectID: payload.EntityID,
		valid:    true,
	}

	if detailsJSON, exists := taskMut.DetailsJSON(); exists {
		details.newDetailsJSON = jsonSliceToString(detailsJSON)
	}

	if title, exists := taskMut.Title(); exists {
		details.objectName = title
	}

	if owner, exists := taskMut.OwnerID(); exists {
		details.ownerID = owner
	}

	// Get old details if this is an update
	if payload.Operation == "UpdateOne" || payload.Operation == "Update" {
		if details.objectID != "" {
			task, err := client.Task.Get(allowCtx, details.objectID)
			if err == nil {
				details.oldDetailsJSON = jsonSliceToString(task.DetailsJSON)
				if details.objectName == "" {
					details.objectName = task.Title
				}
				if details.ownerID == "" {
					details.ownerID = task.OwnerID
				}
			}
		}
	}

	return details
}

// extractDocumentMentionDetails is a generic function that extracts mention details
// from any mutation implementing the documentMutation interface (Risk, Procedure, InternalPolicy).
func extractDocumentMentionDetails[T documentMutation](
	payload *events.MutationPayload,
	mut T,
	queryFunc func() (*oldDocumentDetails, error),
) objectMentionDetails {
	details := objectMentionDetails{
		objectID: payload.EntityID,
		valid:    true,
	}

	if detailsJSON, exists := mut.DetailsJSON(); exists {
		details.newDetailsJSON = jsonSliceToString(detailsJSON)
	}

	if name, exists := mut.Name(); exists {
		details.objectName = name
	}

	if owner, exists := mut.OwnerID(); exists {
		details.ownerID = owner
	}

	// Get old details if this is an update
	if payload.Operation == "UpdateOne" || payload.Operation == "Update" {
		if details.objectID != "" {
			oldDoc, err := queryFunc()
			if err == nil {
				details.oldDetailsJSON = jsonSliceToString(oldDoc.detailsJSON)
				if details.objectName == "" {
					details.objectName = oldDoc.name
				}
				if details.ownerID == "" {
					details.ownerID = oldDoc.ownerID
				}
			}
		}
	}

	return details
}
