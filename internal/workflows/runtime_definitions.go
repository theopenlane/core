package workflows

import (
	"fmt"
	"strings"

	"github.com/samber/lo"

	"github.com/theopenlane/core/common/enums"
	"github.com/theopenlane/core/common/models"
	"github.com/theopenlane/core/internal/ent/generated"
)

// runtimeIDPrefix is prepended to the key when synthesizing ent IDs for runtime definitions
const runtimeIDPrefix = "runtime:"

// RuntimeWorkflowDefinition holds a system workflow definition that lives
// in memory and participates in trigger matching without DB persistence
type RuntimeWorkflowDefinition struct {
	// Key is the stable identifier for this definition
	Key string
	// Name is the display name
	Name string
	// Description is optional description text
	Description string
	// WorkflowKind is the workflow classification
	WorkflowKind enums.WorkflowKind
	// SchemaType is the object type this workflow applies to
	SchemaType string
	// ApprovalSubmissionMode controls how proposals are submitted
	ApprovalSubmissionMode enums.WorkflowApprovalSubmissionMode
	// Definition is the workflow definition document
	Definition models.WorkflowDefinitionDocument
	// TriggerOperations is derived from the definition at registration time
	TriggerOperations []string
	// TriggerFields is derived from the definition at registration time
	TriggerFields []string
}

// ToEntDefinition converts a runtime definition to a synthetic *generated.WorkflowDefinition
// for compatibility with the existing engine execution path
func (d *RuntimeWorkflowDefinition) ToEntDefinition() *generated.WorkflowDefinition {
	key := d.Key

	return &generated.WorkflowDefinition{
		ID:                     runtimeIDPrefix + d.Key,
		SystemOwned:            true,
		SystemInternalID:       &key,
		Name:                   d.Name,
		Description:            d.Description,
		WorkflowKind:           d.WorkflowKind,
		SchemaType:             d.SchemaType,
		DefinitionJSON:         d.Definition,
		TriggerOperations:      d.TriggerOperations,
		TriggerFields:          d.TriggerFields,
		Active:                 true,
		Draft:                  false,
		ApprovalSubmissionMode: d.ApprovalSubmissionMode,
	}
}

// RuntimeDefinitionRegistry holds runtime workflow definitions registered at startup
type RuntimeDefinitionRegistry struct {
	definitions []*RuntimeWorkflowDefinition
	byKey       map[string]*RuntimeWorkflowDefinition
}

// NewRuntimeDefinitionRegistry creates an empty registry
func NewRuntimeDefinitionRegistry() *RuntimeDefinitionRegistry {
	return &RuntimeDefinitionRegistry{
		byKey: make(map[string]*RuntimeWorkflowDefinition),
	}
}

// Register adds a runtime definition to the registry. It derives trigger prefilter
// fields from the definition document using DeriveTriggerPrefilter and returns an
// error if the key is empty or already registered
func (r *RuntimeDefinitionRegistry) Register(def RuntimeWorkflowDefinition) error {
	key := strings.TrimSpace(def.Key)
	if key == "" {
		return ErrRuntimeDefinitionKeyRequired
	}

	if _, exists := r.byKey[key]; exists {
		return fmt.Errorf("%w: %s", ErrRuntimeDefinitionDuplicateKey, key)
	}

	operations, fields := DeriveTriggerPrefilter(def.Definition)
	def.Key = key
	def.TriggerOperations = operations
	def.TriggerFields = fields

	r.definitions = append(r.definitions, &def)
	r.byKey[key] = &def

	return nil
}

// Match returns runtime definitions matching the given trigger criteria.
// It applies the same coarse prefiltering as the DB query: schema type match,
// operation match (if trigger operations are set), and field overlap
// (if trigger fields are set)
func (r *RuntimeDefinitionRegistry) Match(schemaType string, eventType string, changedFields []string, changedEdges []string) []*RuntimeWorkflowDefinition {
	allChanges := lo.Flatten([][]string{changedFields, changedEdges})

	return lo.Filter(r.definitions, func(def *RuntimeWorkflowDefinition, _ int) bool {
		// Schema type must match exactly
		if def.SchemaType != schemaType {
			return false
		}

		// Operation prefilter: if the definition has trigger operations, the event must be among them
		if len(def.TriggerOperations) > 0 && !lo.Contains(def.TriggerOperations, eventType) {
			return false
		}

		// Field prefilter: if the definition has trigger fields, at least one changed field/edge must overlap.
		// Empty trigger fields means the definition matches any field change (same as DB query logic).
		if len(def.TriggerFields) > 0 && len(allChanges) > 0 {
			if len(lo.Intersect(def.TriggerFields, allChanges)) == 0 {
				return false
			}
		}

		return true
	})
}
