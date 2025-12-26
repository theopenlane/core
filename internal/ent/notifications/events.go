package notifications

import (
	"errors"
	"fmt"

	"github.com/theopenlane/core/internal/ent/events"
	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/generated/groupmembership"
	"github.com/theopenlane/core/internal/ent/generated/internalpolicy"
	"github.com/theopenlane/core/internal/ent/generated/note"
	"github.com/theopenlane/core/internal/ent/generated/privacy"
	"github.com/theopenlane/core/internal/ent/generated/task"
	"github.com/theopenlane/core/pkg/enums"
	"github.com/theopenlane/core/pkg/events/soiree"
	"github.com/theopenlane/core/pkg/logx"
	"github.com/theopenlane/core/pkg/slateparser"
)

var (
	// ErrFailedToGetClient is returned when the client cannot be retrieved from context
	ErrFailedToGetClient = errors.New("failed to get client from context")
	// ErrEntityIDNotFound is returned when entity ID is not found in props
	ErrEntityIDNotFound = errors.New("entity ID not found in props")
)

type taskFields struct {
	title    string
	entityID string
	ownerID  string
}
type policyFields struct {
	approverID string
	name       string
	entityID   string
	ownerID    string
}

type taskNotificationInput struct {
	assigneeID string
	taskTitle  string
	taskID     string
	ownerID    string
}

type policyNotificationInput struct {
	approverID string
	policyName string
	policyID   string
	ownerID    string
}

// handleTaskMutation processes task mutations and creates notifications when assignee changes or mentions are added
func handleTaskMutation(ctx *soiree.EventContext, payload *events.MutationPayload) error {
	props := ctx.Properties()
	if props == nil {
		return nil
	}

	// Check if assignee_id field changed - only trigger notification if this field was updated
	assigneeIDVal := props.GetKey(task.FieldAssigneeID)
	if assigneeIDVal != nil {
		assigneeID, ok := assigneeIDVal.(string)
		if ok && assigneeID != "" {
			// Get other fields from props and payload, fallback to database query if missing
			fields, err := fetchTaskFields(ctx, props, payload)
			if err != nil {
				logx.FromContext(ctx.Context()).Error().Err(err).Msg("failed to get task fields")
				return err
			}

			input := taskNotificationInput{
				assigneeID: assigneeID,
				taskTitle:  fields.title,
				taskID:     fields.entityID,
				ownerID:    fields.ownerID,
			}

			if err := addTaskAssigneeNotification(ctx, input); err != nil {
				logx.FromContext(ctx.Context()).Error().Err(err).Msg("failed to add task assignee notification")
				return err
			}
		}
	}

	// Check for mentions in task details
	if err := handleObjectMentions(ctx, payload, "Task"); err != nil {
		logx.FromContext(ctx.Context()).Error().Err(err).Msg("failed to handle task mentions")
		return err
	}

	return nil
}

// fetchTaskFields retrieves task fields from payload, props, or queries database if missing
func fetchTaskFields(ctx *soiree.EventContext, props soiree.Properties, payload *events.MutationPayload) (*taskFields, error) {
	fields := &taskFields{}

	extractTaskFromPayload(payload, fields)
	extractTaskFromProps(props, fields)

	if needsTaskDBQuery(fields) {
		if err := queryTaskFromDB(ctx, fields); err != nil {
			return nil, err
		}
	}

	return fields, nil
}

// extractTaskFromPayload extracts task fields from mutation payload
func extractTaskFromPayload(payload *events.MutationPayload, fields *taskFields) {
	if payload == nil {
		return
	}

	if payload.EntityID != "" {
		fields.entityID = payload.EntityID
	}

	taskMut, ok := payload.Mutation.(*generated.TaskMutation)
	if !ok {
		return
	}

	if title, exists := taskMut.Title(); exists {
		fields.title = title
	}

	if ownerID, exists := taskMut.OwnerID(); exists {
		fields.ownerID = ownerID
	}
}

// extractTaskFromProps extracts task fields from properties
func extractTaskFromProps(props soiree.Properties, fields *taskFields) {
	if fields.title == "" {
		if title, ok := props.GetKey(task.FieldTitle).(string); ok {
			fields.title = title
		}
	}

	if fields.entityID == "" {
		if id, ok := props.GetKey(task.FieldID).(string); ok {
			fields.entityID = id
		}
	}

	if fields.ownerID == "" {
		if ownerID, ok := props.GetKey(task.FieldOwnerID).(string); ok {
			fields.ownerID = ownerID
		}
	}
}

// needsTaskDBQuery checks if database query is needed
func needsTaskDBQuery(fields *taskFields) bool {
	return fields.title == "" || fields.entityID == "" || fields.ownerID == ""
}

// queryTaskFromDB queries task from database to fill missing fields
func queryTaskFromDB(ctx *soiree.EventContext, fields *taskFields) error {
	if fields.entityID == "" {
		return ErrEntityIDNotFound
	}

	client, ok := soiree.ClientAs[*generated.Client](ctx)
	if !ok {
		return ErrFailedToGetClient
	}

	allowCtx := privacy.DecisionContext(ctx.Context(), privacy.Allow)
	taskEntity, err := client.Task.Get(allowCtx, fields.entityID)
	if err != nil {
		return fmt.Errorf("failed to query task: %w", err)
	}

	if fields.title == "" {
		fields.title = taskEntity.Title
	}

	if fields.ownerID == "" {
		fields.ownerID = taskEntity.OwnerID
	}

	return nil
}

// handleInternalPolicyMutation processes internal policy mutations and creates notifications when status = NEEDS_APPROVAL or mentions are added
func handleInternalPolicyMutation(ctx *soiree.EventContext, payload *events.MutationPayload) error {
	props := ctx.Properties()
	if props == nil {
		return nil
	}

	// Check if status field changed - only trigger notification if this field was updated
	statusVal := props.GetKey(internalpolicy.FieldStatus)
	if statusVal != nil {
		status, ok := statusVal.(string)
		if ok {
			statusEnum := enums.ToDocumentStatus(status)

			// Check if status is NEEDS_APPROVAL
			if statusEnum == &enums.DocumentNeedsApproval {
				// Get approver_id from payload and props, fallback to database query if missing
				fields, err := fetchPolicyFields(ctx, props, payload)
				if err != nil {
					logx.FromContext(ctx.Context()).Error().Err(err).Msg("failed to get internal policy fields")
					return err
				}

				if fields.approverID == "" {
					logx.FromContext(ctx.Context()).Warn().Msg("approver_id not set for internal policy with NEEDS_APPROVAL status")
				} else {
					input := policyNotificationInput{
						approverID: fields.approverID,
						policyName: fields.name,
						policyID:   fields.entityID,
						ownerID:    fields.ownerID,
					}

					if err := addInternalPolicyNotification(ctx, input); err != nil {
						logx.FromContext(ctx.Context()).Error().Err(err).Msg("failed to add internal policy notification")
						return err
					}
				}
			}
		}
	}

	// Check for mentions in policy details
	if err := handleObjectMentions(ctx, payload, "InternalPolicy"); err != nil {
		logx.FromContext(ctx.Context()).Error().Err(err).Msg("failed to handle internal policy mentions")
		return err
	}

	return nil
}

// handleRiskMutation processes risk mutations and creates notifications for mentions
func handleRiskMutation(ctx *soiree.EventContext, payload *events.MutationPayload) error {
	return handleObjectMentions(ctx, payload, "Risk")
}

// handleProcedureMutation processes procedure mutations and creates notifications for mentions
func handleProcedureMutation(ctx *soiree.EventContext, payload *events.MutationPayload) error {
	return handleObjectMentions(ctx, payload, "Procedure")
}

// fetchPolicyFields retrieves internal policy fields from payload, props, or queries database if missing
func fetchPolicyFields(ctx *soiree.EventContext, props soiree.Properties, payload *events.MutationPayload) (*policyFields, error) {
	fields := &policyFields{}

	extractPolicyFromPayload(payload, fields)
	extractPolicyFromProps(props, fields)

	if needsPolicyDBQuery(fields) {
		if err := queryPolicyFromDB(ctx, fields); err != nil {
			return nil, err
		}
	}

	return fields, nil
}

// extractPolicyFromPayload extracts policy fields from mutation payload
func extractPolicyFromPayload(payload *events.MutationPayload, fields *policyFields) {
	if payload == nil {
		return
	}

	if payload.EntityID != "" {
		fields.entityID = payload.EntityID
	}

	policyMut, ok := payload.Mutation.(*generated.InternalPolicyMutation)
	if !ok {
		return
	}

	if name, exists := policyMut.Name(); exists {
		fields.name = name
	}

	if ownerID, exists := policyMut.OwnerID(); exists {
		fields.ownerID = ownerID
	}

	if approverID, exists := policyMut.ApproverID(); exists {
		fields.approverID = approverID
	}
}

// extractPolicyFromProps extracts policy fields from properties
func extractPolicyFromProps(props soiree.Properties, fields *policyFields) {
	if fields.approverID == "" {
		if approverID, ok := props.GetKey(internalpolicy.FieldApproverID).(string); ok {
			fields.approverID = approverID
		}
	}

	if fields.name == "" {
		if name, ok := props.GetKey(internalpolicy.FieldName).(string); ok {
			fields.name = name
		}
	}

	if fields.entityID == "" {
		if id, ok := props.GetKey(internalpolicy.FieldID).(string); ok {
			fields.entityID = id
		}
	}

	if fields.ownerID == "" {
		if ownerID, ok := props.GetKey(internalpolicy.FieldOwnerID).(string); ok {
			fields.ownerID = ownerID
		}
	}
}

// needsPolicyDBQuery checks if database query is needed
func needsPolicyDBQuery(fields *policyFields) bool {
	return fields.name == "" || fields.entityID == "" || fields.ownerID == "" || fields.approverID == ""
}

// queryPolicyFromDB queries policy from database to fill missing fields
func queryPolicyFromDB(ctx *soiree.EventContext, fields *policyFields) error {
	if fields.entityID == "" {
		return ErrEntityIDNotFound
	}

	client, ok := soiree.ClientAs[*generated.Client](ctx)
	if !ok {
		return ErrFailedToGetClient
	}

	allowCtx := privacy.DecisionContext(ctx.Context(), privacy.Allow)
	policy, err := client.InternalPolicy.Get(allowCtx, fields.entityID)
	if err != nil {
		return fmt.Errorf("failed to query internal policy: %w", err)
	}

	if fields.name == "" {
		fields.name = policy.Name
	}

	if fields.ownerID == "" {
		fields.ownerID = policy.OwnerID
	}

	if fields.approverID == "" && policy.ApproverID != "" {
		fields.approverID = policy.ApproverID
	}

	return nil
}

func addTaskAssigneeNotification(ctx *soiree.EventContext, input taskNotificationInput) error {
	client, ok := soiree.ClientAs[*generated.Client](ctx)
	if !ok {
		return nil
	}

	consoleURL := client.EntConfig.Notifications.ConsoleURL

	// create the data map with the URL
	dataMap := map[string]interface{}{
		"url": fmt.Sprintf("%s/tasks?id=%s", consoleURL, input.taskID),
	}

	topic := "task_assignment"
	notifInput := &generated.CreateNotificationInput{
		NotificationType: enums.NotificationTypeUser,
		Title:            "New task assigned",
		Body:             fmt.Sprintf("Task %s has been assigned to you", input.taskTitle),
		Data:             dataMap,
		OwnerID:          &input.ownerID,
		Topic:            &topic,
		ObjectType:       "Task",
	}

	return newNotificationCreation(ctx, []string{input.assigneeID}, notifInput)
}

func addInternalPolicyNotification(ctx *soiree.EventContext, input policyNotificationInput) error {
	client, ok := soiree.ClientAs[*generated.Client](ctx)
	if !ok {
		return nil
	}

	// set allow context to query the group
	allowCtx := privacy.DecisionContext(ctx.Context(), privacy.Allow)

	users, err :=
		client.GroupMembership.Query().Where(groupmembership.GroupID(input.approverID)).All(allowCtx)
	if err != nil {
		logx.FromContext(ctx.Context()).Error().Err(err).Str("group_id", input.approverID).Msg("failed to get approver group")
		return err
	}

	if len(users) == 0 {
		logx.FromContext(ctx.Context()).Warn().Str("group_id", input.approverID).Msg("no users found in approver group")
		return nil
	}

	// collect user IDs
	userIDs := make([]string, len(users))
	for i, user := range users {
		userIDs[i] = user.ID
	}

	consoleURL := client.EntConfig.Notifications.ConsoleURL

	// create the data map with the URL
	dataMap := map[string]interface{}{
		"url": fmt.Sprintf("%s/policies/%s/view", consoleURL, input.policyID),
	}

	topic := "policy_approval"
	notifInput := &generated.CreateNotificationInput{
		NotificationType: enums.NotificationTypeOrganization,
		Title:            "Policy approval required",
		Body:             fmt.Sprintf("%s needs approval, internalPolicy", input.policyName),
		Data:             dataMap,
		OwnerID:          &input.ownerID,
		Topic:            &topic,
		ObjectType:       "InternalPolicy",
	}

	return newNotificationCreation(ctx, userIDs, notifInput)
}

func newNotificationCreation(ctx *soiree.EventContext, userIDs []string, input *generated.CreateNotificationInput) error {
	client, ok := soiree.ClientAs[*generated.Client](ctx)
	if !ok {
		return nil
	}

	// set allow context
	allowCtx := privacy.DecisionContext(ctx.Context(), privacy.Allow)

	// create notification per user that it should be sent to
	for _, userID := range userIDs {
		mut := client.Notification.Create().SetInput(*input).SetUserID(userID)

		if _, err := mut.Save(allowCtx); err != nil {
			logx.FromContext(ctx.Context()).Error().Err(err).Msg("failed to create notification")
			return err
		}
	}

	return nil
}

type noteFields struct {
	text        string
	textJSON    string
	oldText     string
	oldTextJSON string
	entityID    string
	ownerID     string
	taskID      string
	controlID   string
	procedureID string
	riskID      string
	policyID    string
	evidenceID  string
}

type mentionNotificationInput struct {
	mentionedUserIDs []string
	objectType       string
	objectID         string
	objectName       string
	ownerID          string
	noteID           string
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
		// Convert []any to string for parsing
		if len(textJSON) > 0 {
			fields.textJSON = fmt.Sprintf("%v", textJSON)
		}
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

	if len(noteEntity.TextJSON) > 0 {
		fields.oldTextJSON = fmt.Sprintf("%v", noteEntity.TextJSON)
	}

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
		url = fmt.Sprintf("%s/controls/%s/view", consoleURL, input.objectID)
	case "Procedure":
		url = fmt.Sprintf("%s/procedures/%s/view", consoleURL, input.objectID)
	case "Risk":
		url = fmt.Sprintf("%s/risks/%s/view", consoleURL, input.objectID)
	case "InternalPolicy":
		url = fmt.Sprintf("%s/policies/%s/view", consoleURL, input.objectID)
	case "Evidence":
		url = fmt.Sprintf("%s/evidence/%s/view", consoleURL, input.objectID)
	default:
		url = fmt.Sprintf("%s/comments/%s", consoleURL, input.noteID)
	}

	// Create the data map with the URL and context
	dataMap := map[string]interface{}{
		"url":         url,
		"object_type": input.objectType,
		"object_id":   input.objectID,
		"object_name": input.objectName,
		"note_id":     input.noteID,
	}

	topic := "mention_alert"
	notifInput := &generated.CreateNotificationInput{
		NotificationType: enums.NotificationTypeUser,
		Title:            "You were mentioned",
		Body:             fmt.Sprintf("You were mentioned in a comment on %s: %s", input.objectType, input.objectName),
		Data:             dataMap,
		OwnerID:          &input.ownerID,
		Topic:            &topic,
		ObjectType:       "Note",
	}

	return newNotificationCreation(ctx, input.mentionedUserIDs, notifInput)
}

// handleObjectMentions is a generic handler for checking mentions in object details fields
func handleObjectMentions(ctx *soiree.EventContext, payload *events.MutationPayload, objectType string) error {
	if payload == nil {
		return nil
	}

	client, ok := soiree.ClientAs[*generated.Client](ctx)
	if !ok {
		return ErrFailedToGetClient
	}

	allowCtx := privacy.DecisionContext(ctx.Context(), privacy.Allow)

	var objectID, objectName, ownerID string
	var newDetailsJSON, oldDetailsJSON string

	// Extract details based on object type
	switch objectType {
	case "Task":
		taskMut, ok := payload.Mutation.(*generated.TaskMutation)
		if !ok {
			return nil
		}

		objectID = payload.EntityID
		if details, exists := taskMut.DetailsJSON(); exists && len(details) > 0 {
			newDetailsJSON = fmt.Sprintf("%v", details)
		}

		if title, exists := taskMut.Title(); exists {
			objectName = title
		}

		if owner, exists := taskMut.OwnerID(); exists {
			ownerID = owner
		}

		// Get old details if this is an update
		if payload.Operation == "UpdateOne" || payload.Operation == "Update" {
			if objectID != "" {
				task, err := client.Task.Get(allowCtx, objectID)
				if err == nil {
					if len(task.DetailsJSON) > 0 {
						oldDetailsJSON = fmt.Sprintf("%v", task.DetailsJSON)
					}
					if objectName == "" {
						objectName = task.Title
					}
					if ownerID == "" {
						ownerID = task.OwnerID
					}
				}
			}
		}

	case "Risk":
		riskMut, ok := payload.Mutation.(*generated.RiskMutation)
		if !ok {
			return nil
		}

		objectID = payload.EntityID
		if details, exists := riskMut.DetailsJSON(); exists && len(details) > 0 {
			newDetailsJSON = fmt.Sprintf("%v", details)
		}

		if name, exists := riskMut.Name(); exists {
			objectName = name
		}

		if owner, exists := riskMut.OwnerID(); exists {
			ownerID = owner
		}

		// Get old details if this is an update
		if payload.Operation == "UpdateOne" || payload.Operation == "Update" {
			if objectID != "" {
				risk, err := client.Risk.Get(allowCtx, objectID)
				if err == nil {
					if len(risk.DetailsJSON) > 0 {
						oldDetailsJSON = fmt.Sprintf("%v", risk.DetailsJSON)
					}
					if objectName == "" {
						objectName = risk.Name
					}
					if ownerID == "" {
						ownerID = risk.OwnerID
					}
				}
			}
		}

	case "Procedure":
		procedureMut, ok := payload.Mutation.(*generated.ProcedureMutation)
		if !ok {
			return nil
		}

		objectID = payload.EntityID
		if details, exists := procedureMut.DetailsJSON(); exists && len(details) > 0 {
			newDetailsJSON = fmt.Sprintf("%v", details)
		}

		if name, exists := procedureMut.Name(); exists {
			objectName = name
		}

		if owner, exists := procedureMut.OwnerID(); exists {
			ownerID = owner
		}

		// Get old details if this is an update
		if payload.Operation == "UpdateOne" || payload.Operation == "Update" {
			if objectID != "" {
				procedure, err := client.Procedure.Get(allowCtx, objectID)
				if err == nil {
					if len(procedure.DetailsJSON) > 0 {
						oldDetailsJSON = fmt.Sprintf("%v", procedure.DetailsJSON)
					}
					if objectName == "" {
						objectName = procedure.Name
					}
					if ownerID == "" {
						ownerID = procedure.OwnerID
					}
				}
			}
		}

	case "InternalPolicy":
		policyMut, ok := payload.Mutation.(*generated.InternalPolicyMutation)
		if !ok {
			return nil
		}

		objectID = payload.EntityID
		if details, exists := policyMut.DetailsJSON(); exists && len(details) > 0 {
			newDetailsJSON = fmt.Sprintf("%v", details)
		}

		if name, exists := policyMut.Name(); exists {
			objectName = name
		}

		if owner, exists := policyMut.OwnerID(); exists {
			ownerID = owner
		}

		// Get old details if this is an update
		if payload.Operation == "UpdateOne" || payload.Operation == "Update" {
			if objectID != "" {
				policy, err := client.InternalPolicy.Get(allowCtx, objectID)
				if err == nil {
					if len(policy.DetailsJSON) > 0 {
						oldDetailsJSON = fmt.Sprintf("%v", policy.DetailsJSON)
					}
					if objectName == "" {
						objectName = policy.Name
					}
					if ownerID == "" {
						ownerID = policy.OwnerID
					}
				}
			}
		}

	default:
		return nil
	}

	// If no new details JSON, nothing to check
	if newDetailsJSON == "" {
		return nil
	}

	// Check if the new details contain valid Slate text
	if !slateparser.IsValidSlateText(newDetailsJSON) {
		return nil
	}

	// Check for new mentions
	newMentions := slateparser.GetNewMentions(oldDetailsJSON, newDetailsJSON, objectType, objectID, objectName)
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
		objectID:         objectID,
		objectName:       objectName,
		ownerID:          ownerID,
		noteID:           "", // No note ID for direct object mentions
	}

	if err := addMentionNotification(ctx, input); err != nil {
		logx.FromContext(ctx.Context()).Error().Err(err).Msg("failed to add mention notification for object")
		return err
	}

	return nil
}

// RegisterListeners registers notification listeners with the given eventer
// This is called from hooks package to register the listeners
func RegisterListeners(addListener func(entityType string, handler func(*soiree.EventContext, *events.MutationPayload) error)) {
	addListener(generated.TypeTask, handleTaskMutation)
	addListener(generated.TypeInternalPolicy, handleInternalPolicyMutation)
	addListener(generated.TypeRisk, handleRiskMutation)
	addListener(generated.TypeProcedure, handleProcedureMutation)
	addListener(generated.TypeNote, handleNoteMutation)
}
