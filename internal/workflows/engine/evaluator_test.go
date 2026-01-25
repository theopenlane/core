//go:build test

package engine_test

import (
	"github.com/oklog/ulid/v2"

	"github.com/theopenlane/core/common/enums"
	"github.com/theopenlane/core/common/models"
	"github.com/theopenlane/core/internal/workflows"
)

// TestWorkflowEngineEvaluator verifies evaluator setup
func (s *WorkflowEngineTestSuite) TestWorkflowEngineEvaluator() {
	wfEngine := s.NewTestEngine(nil)
	s.Require().NotNil(wfEngine)
}

// TestEvaluateConditions verifies condition evaluation
func (s *WorkflowEngineTestSuite) TestEvaluateConditions() {
	_, orgID, userCtx := s.SetupTestUser()

	wfEngine := s.NewTestEngine(nil)
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
		s.Run(tc.name, func() {
			def.DefinitionJSON.Conditions = tc.conditions
			result, err := wfEngine.EvaluateConditions(s.ctx, def, tc.obj, "UPDATE", tc.changedFields, nil, nil, nil)

			if tc.wantErr {
				s.Error(err)
			} else {
				s.NoError(err)
				s.Equal(tc.expected, result)
			}
		})
	}
}

// TestFindMatchingDefinitions verifies definition matching
func (s *WorkflowEngineTestSuite) TestFindMatchingDefinitions() {
	_, orgID, userCtx := s.SetupTestUser()

	wfEngine := s.NewTestEngine(nil)

	obj := &workflows.Object{ID: "test123", Type: enums.WorkflowObjectTypeControl}
	unknownObj := &workflows.Object{ID: "test123", Type: enums.WorkflowObjectType("NonexistentType")}

	s.Run("no matching definitions", func() {
		defs, err := wfEngine.FindMatchingDefinitions(userCtx, "NonexistentType", "UPDATE", []string{"status"}, nil, nil, nil, unknownObj)
		s.NoError(err)
		s.Empty(defs)
	})

	s.Run("finds matching definition", func() {
		s.ClearWorkflowDefinitions()
		def := s.CreateTestWorkflowDefinition(userCtx, orgID)
		def.DefinitionJSON.Triggers = []models.WorkflowTrigger{
			{Operation: "UPDATE", Fields: []string{"status"}},
		}
		def = s.UpdateWorkflowDefinitionWithPrefilter(def, def.DefinitionJSON)

		defs, err := wfEngine.FindMatchingDefinitions(userCtx, def.SchemaType, "UPDATE", []string{"status"}, nil, nil, nil, obj)
		s.NoError(err)
		s.Len(defs, 1)
		s.Equal(def.ID, defs[0].ID)
	})

	s.Run("filters out inactive definitions", func() {
		s.ClearWorkflowDefinitions()
		def := s.CreateTestWorkflowDefinition(userCtx, orgID)
		def.DefinitionJSON.Triggers = []models.WorkflowTrigger{
			{Operation: "UPDATE", Fields: []string{"status"}},
		}
		s.UpdateWorkflowDefinitionInactive(def, def.DefinitionJSON)

		defs, err := wfEngine.FindMatchingDefinitions(userCtx, def.SchemaType, "UPDATE", []string{"status"}, nil, nil, nil, obj)
		s.NoError(err)
		s.Empty(defs)
	})
}

// TestFindMatchingDefinitionsSelectors verifies selector matching
func (s *WorkflowEngineTestSuite) TestFindMatchingDefinitionsSelectors() {
	userID, orgID, userCtx := s.SetupTestUser()
	seedCtx := s.SeedContext(userID, orgID)

	wfEngine := s.NewTestEngine(nil)

	tag, err := s.client.TagDefinition.Create().
		SetName("PCI-" + ulid.Make().String()).
		SetOwnerID(orgID).
		Save(seedCtx)
	s.Require().NoError(err)

	group, err := s.client.Group.Create().
		SetName("Workflow Group " + ulid.Make().String()).
		SetOwnerID(orgID).
		Save(seedCtx)
	s.Require().NoError(err)

	control, err := s.client.Control.Create().
		SetRefCode("CTL-SELECTOR-" + ulid.Make().String()).
		SetOwnerID(orgID).
		SetTags([]string{tag.Name}).
		Save(seedCtx)
	s.Require().NoError(err)

	_, err = s.client.Control.UpdateOneID(control.ID).
		AddEditorIDs(group.ID).
		Save(seedCtx)
	s.Require().NoError(err)

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

	defs, err := wfEngine.FindMatchingDefinitions(userCtx, def.SchemaType, "UPDATE", []string{"status"}, nil, nil, nil, obj)
	s.NoError(err)
	s.Len(defs, 1)

	otherObj := &workflows.Object{ID: "missing", Type: enums.WorkflowObjectTypeProcedure}
	defs, err = wfEngine.FindMatchingDefinitions(userCtx, def.SchemaType, "UPDATE", []string{"status"}, nil, nil, nil, otherObj)
	s.NoError(err)
	s.Empty(defs)
}

// TestFindMatchingDefinitionsEdgeTriggers verifies edge trigger matching
func (s *WorkflowEngineTestSuite) TestFindMatchingDefinitionsEdgeTriggers() {
	_, orgID, userCtx := s.SetupTestUser()
	wfEngine := s.NewTestEngine(nil)

	control, err := s.client.Control.Create().
		SetRefCode("CTL-EDGE-" + ulid.Make().String()).
		SetOwnerID(orgID).
		Save(userCtx)
	s.Require().NoError(err)

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

	defs, err := wfEngine.FindMatchingDefinitions(userCtx, def.SchemaType, "UPDATE", nil, []string{"evidence"}, addedIDs, nil, obj)
	s.NoError(err)
	s.Len(defs, 1)

	defs, err = wfEngine.FindMatchingDefinitions(userCtx, def.SchemaType, "UPDATE", nil, []string{"other_edge"}, addedIDs, nil, obj)
	s.NoError(err)
	s.Empty(defs)
}

// TestPrefilterBehavior verifies prefilter behavior
func (s *WorkflowEngineTestSuite) TestPrefilterBehavior() {
	_, orgID, userCtx := s.SetupTestUser()
	wfEngine := s.NewTestEngine(nil)

	obj := &workflows.Object{ID: "test123", Type: enums.WorkflowObjectTypeControl}

	s.Run("prefilter by operation - UPDATE only", func() {
		s.ClearWorkflowDefinitions()

		def := s.CreateTestWorkflowDefinitionWithPrefilter(userCtx, orgID,
			[]models.WorkflowTrigger{
				{Operation: "UPDATE", ObjectType: enums.WorkflowObjectTypeControl, Fields: []string{"status"}},
			},
			[]string{"UPDATE"},
			[]string{"status"},
		)

		defs, err := wfEngine.FindMatchingDefinitions(userCtx, def.SchemaType, "UPDATE", []string{"status"}, nil, nil, nil, obj)
		s.NoError(err)
		s.Len(defs, 1)
		s.Equal(def.ID, defs[0].ID)

		defs, err = wfEngine.FindMatchingDefinitions(userCtx, def.SchemaType, "CREATE", []string{"status"}, nil, nil, nil, obj)
		s.NoError(err)
		s.Empty(defs)
	})

	s.Run("prefilter by operation - multiple operations", func() {
		s.ClearWorkflowDefinitions()

		def := s.CreateTestWorkflowDefinitionWithPrefilter(userCtx, orgID,
			[]models.WorkflowTrigger{
				{Operation: "UPDATE", ObjectType: enums.WorkflowObjectTypeControl, Fields: []string{"status"}},
				{Operation: "CREATE", ObjectType: enums.WorkflowObjectTypeControl, Fields: []string{"name"}},
			},
			[]string{"CREATE", "UPDATE"},
			[]string{"name", "status"},
		)

		defs, err := wfEngine.FindMatchingDefinitions(userCtx, def.SchemaType, "UPDATE", []string{"status"}, nil, nil, nil, obj)
		s.NoError(err)
		s.Len(defs, 1)

		defs, err = wfEngine.FindMatchingDefinitions(userCtx, def.SchemaType, "CREATE", []string{"name"}, nil, nil, nil, obj)
		s.NoError(err)
		s.Len(defs, 1)

		defs, err = wfEngine.FindMatchingDefinitions(userCtx, def.SchemaType, "DELETE", []string{"id"}, nil, nil, nil, obj)
		s.NoError(err)
		s.Empty(defs)
	})

	s.Run("prefilter by fields - single field", func() {
		s.ClearWorkflowDefinitions()

		def := s.CreateTestWorkflowDefinitionWithPrefilter(userCtx, orgID,
			[]models.WorkflowTrigger{
				{Operation: "UPDATE", ObjectType: enums.WorkflowObjectTypeControl, Fields: []string{"status"}},
			},
			[]string{"UPDATE"},
			[]string{"status"},
		)

		defs, err := wfEngine.FindMatchingDefinitions(userCtx, def.SchemaType, "UPDATE", []string{"status"}, nil, nil, nil, obj)
		s.NoError(err)
		s.Len(defs, 1)

		defs, err = wfEngine.FindMatchingDefinitions(userCtx, def.SchemaType, "UPDATE", []string{"name"}, nil, nil, nil, obj)
		s.NoError(err)
		s.Empty(defs)
	})

	s.Run("prefilter by fields - multiple fields OR semantics", func() {
		s.ClearWorkflowDefinitions()

		def := s.CreateTestWorkflowDefinitionWithPrefilter(userCtx, orgID,
			[]models.WorkflowTrigger{
				{Operation: "UPDATE", ObjectType: enums.WorkflowObjectTypeControl, Fields: []string{"status", "priority"}},
			},
			[]string{"UPDATE"},
			[]string{"priority", "status"},
		)

		defs, err := wfEngine.FindMatchingDefinitions(userCtx, def.SchemaType, "UPDATE", []string{"status"}, nil, nil, nil, obj)
		s.NoError(err)
		s.Len(defs, 1)

		defs, err = wfEngine.FindMatchingDefinitions(userCtx, def.SchemaType, "UPDATE", []string{"priority"}, nil, nil, nil, obj)
		s.NoError(err)
		s.Len(defs, 1)

		defs, err = wfEngine.FindMatchingDefinitions(userCtx, def.SchemaType, "UPDATE", []string{"status", "priority"}, nil, nil, nil, obj)
		s.NoError(err)
		s.Len(defs, 1)

		defs, err = wfEngine.FindMatchingDefinitions(userCtx, def.SchemaType, "UPDATE", []string{"description"}, nil, nil, nil, obj)
		s.NoError(err)
		s.Empty(defs)
	})

	s.Run("prefilter with null trigger_fields matches any field", func() {
		s.ClearWorkflowDefinitions()

		def := s.CreateTestWorkflowDefinitionWithPrefilter(userCtx, orgID,
			[]models.WorkflowTrigger{
				{Operation: "UPDATE", ObjectType: enums.WorkflowObjectTypeControl},
			},
			[]string{"UPDATE"},
			nil,
		)

		defs, err := wfEngine.FindMatchingDefinitions(userCtx, def.SchemaType, "UPDATE", []string{"status"}, nil, nil, nil, obj)
		s.NoError(err)
		s.Len(defs, 1)

		defs, err = wfEngine.FindMatchingDefinitions(userCtx, def.SchemaType, "UPDATE", []string{"any_field"}, nil, nil, nil, obj)
		s.NoError(err)
		s.Len(defs, 1)
	})

	s.Run("prefilter correctly derives from definition JSON", func() {
		s.ClearWorkflowDefinitions()

		def := s.CreateTestWorkflowDefinition(userCtx, orgID)
		def.DefinitionJSON.Triggers = []models.WorkflowTrigger{
			{Operation: "CREATE", ObjectType: enums.WorkflowObjectTypeControl, Fields: []string{"name", "description"}},
			{Operation: "UPDATE", ObjectType: enums.WorkflowObjectTypeControl, Fields: []string{"status"}},
		}

		updated := s.UpdateWorkflowDefinitionWithPrefilter(def, def.DefinitionJSON)

		s.ElementsMatch([]string{"CREATE", "UPDATE"}, updated.TriggerOperations)
		s.ElementsMatch([]string{"description", "name", "status"}, updated.TriggerFields)

		defs, err := wfEngine.FindMatchingDefinitions(userCtx, def.SchemaType, "CREATE", []string{"name"}, nil, nil, nil, obj)
		s.NoError(err)
		s.Len(defs, 1)

		defs, err = wfEngine.FindMatchingDefinitions(userCtx, def.SchemaType, "UPDATE", []string{"status"}, nil, nil, nil, obj)
		s.NoError(err)
		s.Len(defs, 1)

		defs, err = wfEngine.FindMatchingDefinitions(userCtx, def.SchemaType, "DELETE", []string{"id"}, nil, nil, nil, obj)
		s.NoError(err)
		s.Empty(defs)
	})
}
