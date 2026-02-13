package eventqueue

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/riverqueue/river"
	"github.com/theopenlane/iam/auth"
	"github.com/theopenlane/utils/ulids"

	"github.com/theopenlane/core/common/jobspec"
	"github.com/theopenlane/core/internal/ent/events"
	"github.com/theopenlane/core/pkg/events/soiree"
)

const (
	// MutationDispatchJobKind is the River job kind used for durable mutation event dispatch.
	MutationDispatchJobKind = "mutation_dispatch"
)

// MutationDispatchArgs captures a JSON-safe mutation envelope for async dispatch.
type MutationDispatchArgs struct {
	// Topic is the target soiree topic.
	Topic string `json:"topic"`
	// Operation is the mutation operation (e.g. CREATE/UPDATE/DELETE).
	Operation string `json:"operation"`
	// EntityID is the mutated entity identifier.
	EntityID string `json:"entity_id"`
	// MutationType is the ent schema type that emitted the mutation.
	MutationType string `json:"mutation_type"`
	// ChangedFields captures updated/cleared fields.
	ChangedFields []string `json:"changed_fields,omitempty"`
	// ClearedFields captures fields explicitly cleared in the mutation.
	ClearedFields []string `json:"cleared_fields,omitempty"`
	// ChangedEdges captures changed edge names for workflow-eligible edges.
	ChangedEdges []string `json:"changed_edges,omitempty"`
	// AddedIDs captures edge IDs added by edge name.
	AddedIDs map[string][]string `json:"added_ids,omitempty"`
	// RemovedIDs captures edge IDs removed by edge name.
	RemovedIDs map[string][]string `json:"removed_ids,omitempty"`
	// ProposedChanges captures field-level proposed values (including nil for clears).
	ProposedChanges map[string]any `json:"proposed_changes,omitempty"`
	// EventID is the idempotency key associated with this event.
	EventID string `json:"event_id"`
	// Properties contains JSON-safe event properties serialized as strings.
	Properties map[string]string `json:"properties,omitempty"`
	// Auth carries auth context metadata for listener execution.
	Auth *MutationAuthContext `json:"auth,omitempty"`
	// OccurredAt records when the mutation event envelope was created.
	OccurredAt time.Time `json:"occurred_at"`
}

// MutationAuthContext is a portable snapshot of auth metadata.
type MutationAuthContext struct {
	// SubjectID is the authenticated principal identifier.
	SubjectID string `json:"subject_id,omitempty"`
	// SubjectName is the display name of the authenticated principal.
	SubjectName string `json:"subject_name,omitempty"`
	// SubjectEmail is the email of the authenticated principal.
	SubjectEmail string `json:"subject_email,omitempty"`
	// OrganizationID is the active organization scope.
	OrganizationID string `json:"organization_id,omitempty"`
	// OrganizationName is the active organization display name.
	OrganizationName string `json:"organization_name,omitempty"`
	// OrganizationIDs is the set of organizations available in scope.
	OrganizationIDs []string `json:"organization_ids,omitempty"`
	// AuthenticationType identifies the authentication mechanism.
	AuthenticationType string `json:"authentication_type,omitempty"`
	// OrganizationRole captures the caller's role in the active organization.
	OrganizationRole string `json:"organization_role,omitempty"`
	// IsSystemAdmin indicates whether the caller has system-admin privileges.
	IsSystemAdmin bool `json:"is_system_admin,omitempty"`
}

// Kind satisfies river.JobArgs.
func (MutationDispatchArgs) Kind() string { return MutationDispatchJobKind }

// InsertOpts sets default queue behavior for mutation dispatch jobs.
func (MutationDispatchArgs) InsertOpts() river.InsertOpts {
	return river.InsertOpts{Queue: jobspec.QueueDefault}
}

// ToAuthenticatedUser converts the auth snapshot into a runtime auth user context.
func (c *MutationAuthContext) ToAuthenticatedUser() *auth.AuthenticatedUser {
	return &auth.AuthenticatedUser{
		SubjectID:          c.SubjectID,
		SubjectName:        c.SubjectName,
		SubjectEmail:       c.SubjectEmail,
		OrganizationID:     c.OrganizationID,
		OrganizationName:   c.OrganizationName,
		OrganizationIDs:    append([]string(nil), c.OrganizationIDs...),
		AuthenticationType: auth.AuthenticationType(c.AuthenticationType),
		IsSystemAdmin:      c.IsSystemAdmin,
		OrganizationRole:   auth.OrganizationRoleType(c.OrganizationRole),
	}
}

// NewMutationDispatchArgs converts an in-memory mutation payload into a durable River job envelope.
func NewMutationDispatchArgs(ctx context.Context, topic string, payload *events.MutationPayload, props soiree.Properties) MutationDispatchArgs {
	eventProps := make(map[string]string)
	for key, value := range props {
		normalized, ok := normalizePropertyValue(value)
		if !ok {
			continue
		}

		eventProps[key] = normalized
	}

	eventID := strings.TrimSpace(eventProps[soiree.PropertyEventID])
	if eventID == "" {
		eventID = ulids.New().String()
		eventProps[soiree.PropertyEventID] = eventID
	}

	entityID := ""
	operation := ""
	mutationType := strings.TrimSpace(topic)
	changedFields := []string(nil)
	clearedFields := []string(nil)
	changedEdges := []string(nil)
	addedIDs := map[string][]string(nil)
	removedIDs := map[string][]string(nil)
	proposedChanges := map[string]any(nil)

	if payload != nil {
		entityID = payload.EntityID
		operation = payload.Operation
		mutationType = strings.TrimSpace(payload.MutationType)
		changedFields = append([]string(nil), payload.ChangedFields...)
		clearedFields = append([]string(nil), payload.ClearedFields...)
		changedEdges = append([]string(nil), payload.ChangedEdges...)
		addedIDs = events.CloneStringSliceMap(payload.AddedIDs)
		removedIDs = events.CloneStringSliceMap(payload.RemovedIDs)
		proposedChanges = events.CloneAnyMap(payload.ProposedChanges)

		if mutationType == "" && payload.Mutation != nil {
			mutationType = payload.Mutation.Type()
		}
	}

	if entityID != "" {
		eventProps["ID"] = entityID
	}

	return MutationDispatchArgs{
		Topic:           strings.TrimSpace(topic),
		Operation:       operation,
		EntityID:        entityID,
		MutationType:    mutationType,
		ChangedFields:   changedFields,
		ClearedFields:   clearedFields,
		ChangedEdges:    changedEdges,
		AddedIDs:        addedIDs,
		RemovedIDs:      removedIDs,
		ProposedChanges: proposedChanges,
		EventID:         eventID,
		Properties:      eventProps,
		Auth:            authContextFromContext(ctx),
		OccurredAt:      time.Now().UTC(),
	}
}

// authContextFromContext captures auth metadata from context for transport in mutation jobs.
func authContextFromContext(ctx context.Context) *MutationAuthContext {
	au, err := auth.GetAuthenticatedUserFromContext(ctx)
	if err != nil || au == nil {
		return nil
	}

	return &MutationAuthContext{
		SubjectID:          au.SubjectID,
		SubjectName:        au.SubjectName,
		SubjectEmail:       au.SubjectEmail,
		OrganizationID:     au.OrganizationID,
		OrganizationName:   au.OrganizationName,
		OrganizationIDs:    append([]string(nil), au.OrganizationIDs...),
		AuthenticationType: string(au.AuthenticationType),
		OrganizationRole:   string(au.OrganizationRole),
		IsSystemAdmin:      au.IsSystemAdmin,
	}
}

// normalizePropertyValue converts event properties into JSON-safe string values.
func normalizePropertyValue(value any) (string, bool) {
	if value == nil {
		return "", false
	}

	switch typed := value.(type) {
	case string:
		return typed, true
	case []byte:
		return string(typed), true
	case bool:
		return fmt.Sprintf("%t", typed), true
	case int, int8, int16, int32, int64:
		return fmt.Sprintf("%d", typed), true
	case uint, uint8, uint16, uint32, uint64:
		return fmt.Sprintf("%d", typed), true
	case float32:
		return fmt.Sprintf("%f", typed), true
	case float64:
		return fmt.Sprintf("%f", typed), true
	case time.Time:
		return typed.UTC().Format(time.RFC3339Nano), true
	case fmt.Stringer:
		return typed.String(), true
	}

	raw, err := json.Marshal(value)
	if err != nil {
		return "", false
	}

	return string(raw), true
}
