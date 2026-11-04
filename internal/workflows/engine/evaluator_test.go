//go:build test

package engine_test

import (
	"context"
	"slices"
	"testing"

	"github.com/oklog/ulid/v2"
	"github.com/theopenlane/iam/auth"

	"github.com/theopenlane/core/common/enums"
	"github.com/theopenlane/core/common/models"
	"github.com/theopenlane/core/v2/internal/ent/generated"
	"github.com/theopenlane/core/v2/internal/workflows"
	"gotest.tools/v3/assert"
	is "gotest.tools/v3/assert/cmp"
)

// TestWorkflowEngineEvaluator verifies the workflow engine's CEL evaluator is properly initialized.
//
// Test Flow:
//  1. Retrieves the workflow engine
//  2. Verifies the engine is not nil (evaluator infrastructure ready)
//
// Why This Matters:
//
//	Foundational test ensuring the CEL expression evaluator is available for condition
//	and "when" clause evaluation.
func TestWorkflowEngineEvaluator(t *testing.T) {
	s := suite.setupParallelTest(t)
	wfEngine := s.Engine()
	assert.Assert(t, wfEngine != nil)
}

// TestEvaluateConditions verifies the CEL expression evaluator correctly evaluates workflow
// conditions, which determine whether a workflow should proceed after trigger matching.
//
// Test Cases:
//
//	"no conditions always passes":
//	  - Empty conditions list -> Returns true (workflow proceeds)
//
//	"single condition passes":
//	  - Condition: "true" -> Returns true
//
//	"single condition fails":
//	  - Condition: "false" -> Returns false (workflow blocked)
//
//	"all conditions pass" (AND semantics):
//	  - Conditions: ["true", "'status' in changed_fields"]
//	  - Both evaluate to true -> Returns true
//
//	"first condition fails":
//	  - Conditions: ["false", "true"]
//	  - First is false -> Returns false (short-circuit)
//
//	"invalid condition expression":
//	  - Condition: "invalid syntax"
//	  - Expected: Error returned (CEL parsing failure)
//
// Why This Matters:
//
//	Conditions are the second layer of workflow filtering (after triggers). They enable
//	complex business rules using CEL expressions with access to object state and trigger context.
func TestEvaluateConditions(t *testing.T) {
	s := suite.setupParallelTest(t)
	_, orgID, userCtx := s.SetupTestUser()

	wfEngine := s.Engine()
	def := s.CreateTestWorkflowDefinition(userCtx, orgID)

	testCases := []struct {
		name          string
		conditions    []models.WorkflowCondition
		obj           *workflows.Object
		changedFields []string
		expected      bool
		wantErr       bool
	}{
		{
			name:       "no conditions always passes",
			conditions: []models.WorkflowCondition{},
			obj: &workflows.Object{
				ID:   "test123",
				Type: enums.WorkflowObjectTypeControl,
			},
			changedFields: []string{"status"},
			expected:      true,
			wantErr:       false,
		},
		{
			name: "single condition passes",
			conditions: []models.WorkflowCondition{
				{Expression: "true"},
			},
			obj: &workflows.Object{
				ID:   "test123",
				Type: enums.WorkflowObjectTypeControl,
			},
			changedFields: []string{"status"},
			expected:      true,
			wantErr:       false,
		},
		{
			name: "single condition fails",
			conditions: []models.WorkflowCondition{
				{Expression: "false"},
			},
			obj: &workflows.Object{
				ID:   "test123",
				Type: enums.WorkflowObjectTypeControl,
			},
			changedFields: []string{"status"},
			expected:      false,
			wantErr:       false,
		},
		{
			name: "all conditions pass",
			conditions: []models.WorkflowCondition{
				{Expression: "true"},
				{Expression: "'status' in changed_fields"},
			},
			obj: &workflows.Object{
				ID:   "test123",
				Type: enums.WorkflowObjectTypeControl,
			},
			changedFields: []string{"status"},
			expected:      true,
			wantErr:       false,
		},
		{
			name: "first condition fails",
			conditions: []models.WorkflowCondition{
				{Expression: "false"},
				{Expression: "true"},
			},
			obj: &workflows.Object{
				ID:   "test123",
				Type: enums.WorkflowObjectTypeControl,
			},
			changedFields: []string{"status"},
			expected:      false,
			wantErr:       false,
		},
		{
			name: "invalid condition expression",
			conditions: []models.WorkflowCondition{
				{Expression: "invalid syntax"},
			},
			obj: &workflows.Object{
				ID:   "test123",
				Type: enums.WorkflowObjectTypeControl,
			},
			changedFields: []string{"status"},
			expected:      false,
			wantErr:       true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			s := s.withT(t)
			def.DefinitionJSON.Conditions = tc.conditions
			result, err := wfEngine.EvaluateConditions(s.ctx, def, tc.obj, "UPDATE", tc.changedFields, nil, nil, nil, nil)

			if tc.wantErr {
				assert.Check(t, err != nil)
			} else {
				assert.Check(t, err)
				assert.Check(t, is.DeepEqual(tc.expected, result))
			}
		})
	}
}

// TestFindMatchingDefinitions verifies the workflow engine correctly finds workflow definitions
// that match a given mutation based on schema type, operation, and field changes.
//
// Test Scenarios:
//
//	"no matching definitions":
//	  - Queries for a non-existent schema type
//	  - Expected: Empty result
//
//	"finds matching definition":
//	  - Creates a definition for Control with UPDATE trigger on "status" field
//	  - Queries for Control UPDATE with changed_fields = ["status"]
//	  - Expected: The definition is returned
//
//	"filters out inactive definitions":
//	  - Creates a definition but sets active = false
//	  - Queries for matching criteria
//	  - Expected: Empty result (inactive definitions excluded)
//
// Why This Matters:
//
//	FindMatchingDefinitions is the core lookup function called during mutation hooks.
//	It must correctly filter by schema type, active status, operation, and fields to
//	find only the relevant workflow definitions.
func TestFindMatchingDefinitions(t *testing.T) {
	s := suite.setupParallelTest(t)
	_, orgID, userCtx := s.SetupTestUser()

	wfEngine := s.Engine()

	control, err := s.client.Control.Create().
		SetRefCode("CTL-MATCH-" + ulid.Make().String()).
		SetOwnerID(orgID).
		Save(userCtx)
	assert.NilError(t, err)

	obj := &workflows.Object{ID: control.ID, Type: enums.WorkflowObjectTypeControl}
	unknownObj := &workflows.Object{ID: control.ID, Type: enums.WorkflowObjectType("NonexistentType")}

	t.Run("no matching definitions", func(t *testing.T) {
		defs, err := wfEngine.FindMatchingDefinitions(userCtx, "NonexistentType", "UPDATE", []string{"status"}, nil, nil, nil, nil, obj)
		assert.Check(t, err)
		assert.Check(t, is.Len(defs, 0))
	})

	t.Run("unsupported object type", func(t *testing.T) {
		defs, err := wfEngine.FindMatchingDefinitions(userCtx, "NonexistentType", "UPDATE", []string{"status"}, nil, nil, nil, nil, unknownObj)
		assert.Check(t, is.ErrorIs(err, workflows.ErrUnsupportedObjectType))
		assert.Check(t, is.Len(defs, 0))
	})

	t.Run("finds matching definition", func(t *testing.T) {
		s := s.withT(t)
		def := s.CreateTestWorkflowDefinition(userCtx, orgID)
		t.Cleanup(func() { s.ClearWorkflowDefinitionsForOrg(orgID) })

		def.DefinitionJSON.Triggers = []models.WorkflowTrigger{
			{Operation: "UPDATE", Fields: []string{"status"}},
		}
		def = s.UpdateWorkflowDefinitionWithPrefilter(def, def.DefinitionJSON)

		defs, err := wfEngine.FindMatchingDefinitions(userCtx, def.SchemaType, "UPDATE", []string{"status"}, nil, nil, nil, nil, obj)
		assert.Check(t, err)
		assert.Check(t, is.Len(defs, 1))
		assert.Check(t, is.DeepEqual(def.ID, defs[0].ID))
	})

	t.Run("filters out inactive definitions", func(t *testing.T) {
		s := s.withT(t)
		def := s.CreateTestWorkflowDefinition(userCtx, orgID)
		t.Cleanup(func() { s.ClearWorkflowDefinitionsForOrg(orgID) })

		def.DefinitionJSON.Triggers = []models.WorkflowTrigger{
			{Operation: "UPDATE", Fields: []string{"status"}},
		}
		s.UpdateWorkflowDefinitionInactive(def, def.DefinitionJSON)

		defs, err := wfEngine.FindMatchingDefinitions(userCtx, def.SchemaType, "UPDATE", []string{"status"}, nil, nil, nil, nil, obj)
		assert.Check(t, err)
		assert.Check(t, is.Len(defs, 0))
	})
}

// TestFindMatchingDefinitionsAuthContextPermutations verifies auth context permutations used by
// workflow matching, including PAT/API token contexts without a selected OrganizationID.
func TestFindMatchingDefinitionsAuthContextPermutations(t *testing.T) {
	s := suite.setupParallelTest(t)
	userID, orgID, userCtx := s.SetupTestUser()

	wfEngine := s.Engine()
	def := s.CreateTestWorkflowDefinitionWithPrefilter(
		userCtx,
		orgID,
		[]models.WorkflowTrigger{
			{Operation: "UPDATE", Fields: []string{"status"}},
		},
		[]string{"UPDATE"},
		[]string{"status"},
	)
	t.Cleanup(func() { s.ClearWorkflowDefinitionsForOrg(orgID) })

	control, err := s.client.Control.Create().
		SetRefCode("CTL-AUTH-" + ulid.Make().String()).
		SetOwnerID(orgID).
		Save(userCtx)
	assert.NilError(t, err)

	obj := &workflows.Object{ID: control.ID, Type: enums.WorkflowObjectTypeControl}
	otherOrgID := ulid.Make().String()
	apiTokenSubjectID := ulid.Make().String()

	testCases := []struct {
		name   string
		caller *auth.Caller
	}{
		{
			name: "jwt with selected organization",
			caller: &auth.Caller{
				SubjectID:          userID,
				OrganizationID:     orgID,
				OrganizationIDs:    []string{orgID},
				AuthenticationType: auth.JWTAuthentication,
			},
		},
		{
			name: "pat with selected organization",
			caller: &auth.Caller{
				SubjectID:          userID,
				OrganizationID:     orgID,
				OrganizationIDs:    []string{orgID},
				AuthenticationType: auth.PATAuthentication,
			},
		},
		{
			name: "pat with single authorized organization and no selected organization",
			caller: &auth.Caller{
				SubjectID:          userID,
				OrganizationIDs:    []string{orgID},
				AuthenticationType: auth.PATAuthentication,
			},
		},
		{
			name: "pat with multiple authorized organizations and no selected organization",
			caller: &auth.Caller{
				SubjectID:          userID,
				OrganizationIDs:    []string{orgID, otherOrgID},
				AuthenticationType: auth.PATAuthentication,
			},
		},
		{
			name: "pat with empty authorized organizations and no selected organization",
			caller: &auth.Caller{
				SubjectID:          userID,
				OrganizationIDs:    []string{},
				AuthenticationType: auth.PATAuthentication,
			},
		},
		{
			name: "api token with selected organization",
			caller: &auth.Caller{
				SubjectID:          apiTokenSubjectID,
				OrganizationID:     orgID,
				OrganizationIDs:    []string{orgID},
				AuthenticationType: auth.APITokenAuthentication,
			},
		},
		{
			name: "api token with single authorized organization and no selected organization",
			caller: &auth.Caller{
				SubjectID:          apiTokenSubjectID,
				OrganizationIDs:    []string{orgID},
				AuthenticationType: auth.APITokenAuthentication,
			},
		},
		{
			name:   "missing authenticated user",
			caller: nil,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			s := s.withT(t)
			ctx := context.Background()
			if tc.caller != nil {
				ctx = auth.WithCaller(ctx, tc.caller)
			}
			ctx = generated.NewContext(ctx, s.client)

			defs, err := wfEngine.FindMatchingDefinitions(ctx, def.SchemaType, "UPDATE", []string{"status"}, nil, nil, nil, nil, obj)
			assert.Check(t, err)
			assert.Check(t, is.Len(defs, 1))
			assert.Check(t, is.DeepEqual(def.ID, defs[0].ID))
		})
	}
}

// TestFindMatchingDefinitionsSelectors verifies that workflow definitions with selectors
// (tag IDs, group IDs, object types) only match objects that satisfy those selector criteria.
//
// Test Setup:
//  1. Creates a TagDefinition and a Group
//  2. Creates a Control associated with both the tag and group
//  3. Creates a workflow definition with selectors requiring both tag and group
//
// Test Cases:
//   - Object with matching tag AND group -> Definition matches
//   - Object of different type (Procedure vs Control) -> Definition does NOT match
//
// Why This Matters:
//
//	Selectors enable scoping workflows to specific subsets of objects. This is essential
//	for multi-tenant environments where different teams have different approval requirements.
func TestFindMatchingDefinitionsSelectors(t *testing.T) {
	s := suite.setupParallelTest(t)
	userID, orgID, userCtx := s.SetupTestUser()
	seedCtx := s.SeedContext(userID, orgID)

	wfEngine := s.Engine()

	tag, err := s.client.TagDefinition.Create().
		SetName("PCI-" + ulid.Make().String()).
		SetOwnerID(orgID).
		Save(seedCtx)
	assert.NilError(t, err)

	group, err := s.client.Group.Create().
		SetName("Workflow Group " + ulid.Make().String()).
		SetOwnerID(orgID).
		Save(seedCtx)
	assert.NilError(t, err)

	control, err := s.client.Control.Create().
		SetRefCode("CTL-SELECTOR-" + ulid.Make().String()).
		SetOwnerID(orgID).
		SetTags([]string{tag.Name}).
		Save(seedCtx)
	assert.NilError(t, err)

	_, err = s.client.Control.UpdateOneID(control.ID).
		AddEditorIDs(group.ID).
		Save(seedCtx)
	assert.NilError(t, err)

	obj := &workflows.Object{ID: control.ID, Type: enums.WorkflowObjectTypeControl}

	triggers := []models.WorkflowTrigger{
		{
			Operation: "UPDATE",
			Fields:    []string{"status"},
			Selector: models.WorkflowSelector{
				TagIDs:      []string{tag.ID},
				GroupIDs:    []string{group.ID},
				ObjectTypes: []enums.WorkflowObjectType{enums.WorkflowObjectTypeControl},
			},
		},
	}

	def := s.CreateTestWorkflowDefinitionWithPrefilter(userCtx, orgID, triggers, []string{"UPDATE"}, []string{"status"})

	defs, err := wfEngine.FindMatchingDefinitions(userCtx, def.SchemaType, "UPDATE", []string{"status"}, nil, nil, nil, nil, obj)
	assert.Check(t, err)
	assert.Check(t, is.Len(defs, 1))

	procedure, err := s.client.Procedure.Create().
		SetName("Procedure " + ulid.Make().String()).
		SetOwnerID(orgID).
		Save(seedCtx)
	assert.NilError(t, err)

	otherObj := &workflows.Object{ID: procedure.ID, Type: enums.WorkflowObjectTypeProcedure}
	defs, err = wfEngine.FindMatchingDefinitions(userCtx, def.SchemaType, "UPDATE", []string{"status"}, nil, nil, nil, nil, otherObj)
	assert.Check(t, err)
	assert.Check(t, is.Len(defs, 0))
}

// TestFindMatchingDefinitionsEdgeTriggers verifies that workflows can trigger on edge
// (relationship) changes with CEL expressions that evaluate added/removed edge IDs.
//
// Workflow Definition (Plain English):
//
//	"Trigger on UPDATE when 'evidence' edge changes AND exactly 1 evidence was added"
//	Trigger: edges = ["evidence"]
//	Expression: 'evidence' in changed_edges && size(added_ids['evidence']) == 1
//
// Test Cases:
//
//   - changed_edges = ["evidence"], added_ids["evidence"] = ["evidence-1"]
//     -> Definition matches (edge changed, exactly 1 added)
//
//   - changed_edges = ["other_edge"]
//     -> Definition does NOT match ('evidence' not in changed_edges)
//
// Why This Matters:
//
//	Edge-based triggers enable workflows that respond to relationship changes, not just
//	field value changes. The CEL expression support allows complex edge-based conditions.
func TestFindMatchingDefinitionsEdgeTriggers(t *testing.T) {
	s := suite.setupParallelTest(t)
	_, orgID, userCtx := s.SetupTestUser()
	wfEngine := s.Engine()

	control, err := s.client.Control.Create().
		SetRefCode("CTL-EDGE-" + ulid.Make().String()).
		SetOwnerID(orgID).
		Save(userCtx)
	assert.NilError(t, err)

	obj := &workflows.Object{ID: control.ID, Type: enums.WorkflowObjectTypeControl}

	triggers := []models.WorkflowTrigger{
		{
			Operation:  "UPDATE",
			Edges:      []string{"evidence"},
			Expression: "'evidence' in changed_edges && size(added_ids['evidence']) == 1",
		},
	}

	def := s.CreateTestWorkflowDefinitionWithPrefilter(userCtx, orgID, triggers, []string{"UPDATE"}, []string{"evidence"})

	addedIDs := map[string][]string{"evidence": {"evidence-1"}}

	defs, err := wfEngine.FindMatchingDefinitions(userCtx, def.SchemaType, "UPDATE", nil, []string{"evidence"}, addedIDs, nil, nil, obj)
	assert.Check(t, err)
	assert.Check(t, is.Len(defs, 1))

	defs, err = wfEngine.FindMatchingDefinitions(userCtx, def.SchemaType, "UPDATE", nil, []string{"other_edge"}, addedIDs, nil, nil, obj)
	assert.Check(t, err)
	assert.Check(t, is.Len(defs, 0))
}

// TestPrefilterBehavior verifies the prefilter optimization that allows efficient database
// queries by storing denormalized trigger_operations and trigger_fields on workflow definitions.
// The prefilter enables filtering definitions at the database level before evaluating CEL expressions.
//
// Test Scenarios:
//
//	"prefilter by operation - UPDATE only":
//	  - Definition triggers on UPDATE
//	  - Query with UPDATE -> Matches
//	  - Query with CREATE -> Does NOT match
//
//	"prefilter by operation - multiple operations":
//	  - Definition triggers on UPDATE and CREATE
//	  - Query with UPDATE -> Matches
//	  - Query with CREATE -> Matches
//	  - Query with DELETE -> Does NOT match
//
//	"prefilter by fields - single field":
//	  - Definition triggers on "status" field changes
//	  - Query with changed_fields = ["status"] -> Matches
//	  - Query with changed_fields = ["name"] -> Does NOT match
//
//	"prefilter by fields - multiple fields OR semantics":
//	  - Definition triggers on "status" OR "priority" field changes
//	  - Query with ["status"] -> Matches
//	  - Query with ["priority"] -> Matches
//	  - Query with ["description"] -> Does NOT match
//
//	"prefilter with null trigger_fields matches any field":
//	  - Definition with no field restrictions (null trigger_fields)
//	  - Query with any changed_fields -> Matches
//
//	"prefilter correctly derives from definition JSON":
//	  - Verifies DeriveTriggerPrefilter extracts correct values from workflow definition document
//
// Why This Matters:
//
//	Prefiltering avoids loading all workflow definitions and evaluating their CEL expressions
//	for every mutation. The database can efficiently filter based on indexed columns.
func TestPrefilterBehavior(t *testing.T) {
	s := suite.setupParallelTest(t)
	_, orgID, userCtx := s.SetupTestUser()
	wfEngine := s.Engine()

	control, err := s.client.Control.Create().
		SetRefCode("CTL-PREFILTER-" + ulid.Make().String()).
		SetOwnerID(orgID).
		Save(userCtx)
	assert.NilError(t, err)

	obj := &workflows.Object{ID: control.ID, Type: enums.WorkflowObjectTypeControl}

	t.Run("prefilter by operation - UPDATE only", func(t *testing.T) {
		s := s.withT(t)
		def := s.CreateTestWorkflowDefinitionWithPrefilter(userCtx, orgID,
			[]models.WorkflowTrigger{
				{Operation: "UPDATE", ObjectType: enums.WorkflowObjectTypeControl, Fields: []string{"status"}},
			},
			[]string{"UPDATE"},
			[]string{"status"},
		)
		t.Cleanup(func() { s.ClearWorkflowDefinitionsForOrg(orgID) })

		defs, err := wfEngine.FindMatchingDefinitions(userCtx, def.SchemaType, "UPDATE", []string{"status"}, nil, nil, nil, nil, obj)
		assert.Check(t, err)
		assert.Check(t, is.Len(defs, 1))
		assert.Check(t, is.DeepEqual(def.ID, defs[0].ID))

		defs, err = wfEngine.FindMatchingDefinitions(userCtx, def.SchemaType, "CREATE", []string{"status"}, nil, nil, nil, nil, obj)
		assert.Check(t, err)
		assert.Check(t, is.Len(defs, 0))
	})

	t.Run("prefilter by operation - multiple operations", func(t *testing.T) {
		s := s.withT(t)
		def := s.CreateTestWorkflowDefinitionWithPrefilter(userCtx, orgID,
			[]models.WorkflowTrigger{
				{Operation: "UPDATE", ObjectType: enums.WorkflowObjectTypeControl, Fields: []string{"status"}},
				{Operation: "CREATE", ObjectType: enums.WorkflowObjectTypeControl, Fields: []string{"name"}},
			},
			[]string{"CREATE", "UPDATE"},
			[]string{"name", "status"},
		)
		t.Cleanup(func() { s.ClearWorkflowDefinitionsForOrg(orgID) })

		defs, err := wfEngine.FindMatchingDefinitions(userCtx, def.SchemaType, "UPDATE", []string{"status"}, nil, nil, nil, nil, obj)
		assert.Check(t, err)
		assert.Check(t, is.Len(defs, 1))

		defs, err = wfEngine.FindMatchingDefinitions(userCtx, def.SchemaType, "CREATE", []string{"name"}, nil, nil, nil, nil, obj)
		assert.Check(t, err)
		assert.Check(t, is.Len(defs, 1))

		defs, err = wfEngine.FindMatchingDefinitions(userCtx, def.SchemaType, "DELETE", []string{"id"}, nil, nil, nil, nil, obj)
		assert.Check(t, err)
		assert.Check(t, is.Len(defs, 0))
	})

	t.Run("prefilter by fields - single field", func(t *testing.T) {
		s := s.withT(t)
		def := s.CreateTestWorkflowDefinitionWithPrefilter(userCtx, orgID,
			[]models.WorkflowTrigger{
				{Operation: "UPDATE", ObjectType: enums.WorkflowObjectTypeControl, Fields: []string{"status"}},
			},
			[]string{"UPDATE"},
			[]string{"status"},
		)
		t.Cleanup(func() { s.ClearWorkflowDefinitionsForOrg(orgID) })

		defs, err := wfEngine.FindMatchingDefinitions(userCtx, def.SchemaType, "UPDATE", []string{"status"}, nil, nil, nil, nil, obj)
		assert.Check(t, err)
		assert.Check(t, is.Len(defs, 1))

		defs, err = wfEngine.FindMatchingDefinitions(userCtx, def.SchemaType, "UPDATE", []string{"name"}, nil, nil, nil, nil, obj)
		assert.Check(t, err)
		assert.Check(t, is.Len(defs, 0))
	})

	t.Run("prefilter by fields - multiple fields OR semantics", func(t *testing.T) {
		s := s.withT(t)
		def := s.CreateTestWorkflowDefinitionWithPrefilter(userCtx, orgID,
			[]models.WorkflowTrigger{
				{Operation: "UPDATE", ObjectType: enums.WorkflowObjectTypeControl, Fields: []string{"status", "priority"}},
			},
			[]string{"UPDATE"},
			[]string{"priority", "status"},
		)
		t.Cleanup(func() { s.ClearWorkflowDefinitionsForOrg(orgID) })

		defs, err := wfEngine.FindMatchingDefinitions(userCtx, def.SchemaType, "UPDATE", []string{"status"}, nil, nil, nil, nil, obj)
		assert.Check(t, err)
		assert.Check(t, is.Len(defs, 1))

		defs, err = wfEngine.FindMatchingDefinitions(userCtx, def.SchemaType, "UPDATE", []string{"priority"}, nil, nil, nil, nil, obj)
		assert.Check(t, err)
		assert.Check(t, is.Len(defs, 1))

		defs, err = wfEngine.FindMatchingDefinitions(userCtx, def.SchemaType, "UPDATE", []string{"status", "priority"}, nil, nil, nil, nil, obj)
		assert.Check(t, err)
		assert.Check(t, is.Len(defs, 1))

		defs, err = wfEngine.FindMatchingDefinitions(userCtx, def.SchemaType, "UPDATE", []string{"description"}, nil, nil, nil, nil, obj)
		assert.Check(t, err)
		assert.Check(t, is.Len(defs, 0))
	})

	t.Run("prefilter with null trigger_fields matches any field", func(t *testing.T) {
		s := s.withT(t)
		def := s.CreateTestWorkflowDefinitionWithPrefilter(userCtx, orgID,
			[]models.WorkflowTrigger{
				{Operation: "UPDATE", ObjectType: enums.WorkflowObjectTypeControl},
			},
			[]string{"UPDATE"},
			nil,
		)
		t.Cleanup(func() { s.ClearWorkflowDefinitionsForOrg(orgID) })

		defs, err := wfEngine.FindMatchingDefinitions(userCtx, def.SchemaType, "UPDATE", []string{"status"}, nil, nil, nil, nil, obj)
		assert.Check(t, err)
		assert.Check(t, is.Len(defs, 1))

		defs, err = wfEngine.FindMatchingDefinitions(userCtx, def.SchemaType, "UPDATE", []string{"any_field"}, nil, nil, nil, nil, obj)
		assert.Check(t, err)
		assert.Check(t, is.Len(defs, 1))
	})

	t.Run("prefilter correctly derives from definition JSON", func(t *testing.T) {
		s := s.withT(t)
		def := s.CreateTestWorkflowDefinition(userCtx, orgID)
		t.Cleanup(func() { s.ClearWorkflowDefinitionsForOrg(orgID) })

		def.DefinitionJSON.Triggers = []models.WorkflowTrigger{
			{Operation: "CREATE", ObjectType: enums.WorkflowObjectTypeControl, Fields: []string{"name", "description"}},
			{Operation: "UPDATE", ObjectType: enums.WorkflowObjectTypeControl, Fields: []string{"status"}},
		}

		updated := s.UpdateWorkflowDefinitionWithPrefilter(def, def.DefinitionJSON)

		assert.Check(t, is.DeepEqual([]string{"CREATE", "UPDATE"}, slices.Sorted(slices.Values(updated.TriggerOperations))))
		assert.Check(t, is.DeepEqual([]string{"description", "name", "status"}, slices.Sorted(slices.Values(updated.TriggerFields))))

		defs, err := wfEngine.FindMatchingDefinitions(userCtx, def.SchemaType, "CREATE", []string{"name"}, nil, nil, nil, nil, obj)
		assert.Check(t, err)
		assert.Check(t, is.Len(defs, 1))

		defs, err = wfEngine.FindMatchingDefinitions(userCtx, def.SchemaType, "UPDATE", []string{"status"}, nil, nil, nil, nil, obj)
		assert.Check(t, err)
		assert.Check(t, is.Len(defs, 1))

		defs, err = wfEngine.FindMatchingDefinitions(userCtx, def.SchemaType, "DELETE", []string{"id"}, nil, nil, nil, nil, obj)
		assert.Check(t, err)
		assert.Check(t, is.Len(defs, 0))
	})
}
