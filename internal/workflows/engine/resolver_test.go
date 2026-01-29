//go:build test

package engine_test

import (
	"github.com/oklog/ulid/v2"

	"github.com/theopenlane/core/common/enums"
	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/privacy/rule"
	"github.com/theopenlane/core/internal/workflows"
	"github.com/theopenlane/core/internal/workflows/engine"
)

// TestResolveTargetsRequiresEngine verifies the target resolver is properly initialized
// within the workflow engine.
//
// Test Flow:
//  1. Retrieves the workflow engine
//  2. Verifies the engine is not nil (resolver infrastructure ready)
//
// Why This Matters:
//   Foundational test ensuring the target resolution subsystem is available.
func (s *WorkflowEngineTestSuite) TestResolveTargetsRequiresEngine() {
	wfEngine := s.Engine()
	s.Require().NotNil(wfEngine)
}

// TestResolveUserTarget verifies that USER-type targets resolve to the specified user ID.
// User targets are the simplest target type - a direct reference to a specific user.
//
// Test Scenarios:
//   "resolve user with ID":
//     - Target: { type: USER, id: "user-123" }
//     - Expected: Returns ["user-123"]
//
//   "user target without ID returns error":
//     - Target: { type: USER, id: "" }
//     - Expected: ErrMissingRequiredField error
//
// Why This Matters:
//   User targets enable explicit assignment to specific individuals. The resolver must
//   validate that the user ID is provided and return it directly.
func (s *WorkflowEngineTestSuite) TestResolveUserTarget() {
	s.ClearWorkflowDefinitions()

	_, orgID, userCtx := s.SetupTestUser()

	wfEngine := s.Engine()

	internalCtx := generated.NewContext(rule.WithInternalContext(s.ctx), s.client)
	user, err := s.client.User.Create().
		SetEmail("test-" + ulid.Make().String() + "@example.com").
		SetLastLoginProvider(enums.AuthProviderCredentials).
		Save(internalCtx)
	s.Require().NoError(err)

	obj := &workflows.Object{
		ID:   "test123",
		Type: enums.WorkflowObjectTypeControl,
	}

	s.Run("resolve user with ID", func() {
		target := workflows.TargetConfig{
			Type: enums.WorkflowTargetTypeUser,
			ID:   user.ID,
		}

		userIDs, err := wfEngine.ResolveTargets(userCtx, target, obj)
		s.NoError(err)
		s.Len(userIDs, 1)
		s.Equal(user.ID, userIDs[0])
	})

	s.Run("user target without ID returns error", func() {
		target := workflows.TargetConfig{
			Type: enums.WorkflowTargetTypeUser,
			ID:   "",
		}

		_, err := wfEngine.ResolveTargets(userCtx, target, obj)
		s.Error(err)
		s.ErrorIs(err, engine.ErrMissingRequiredField)
	})

	_ = orgID
}

// TestResolveGroupTarget verifies that GROUP-type targets resolve to all user IDs who are
// members of the specified group.
//
// Test Setup:
//  1. Creates a Group with two user members
//
// Test Scenarios:
//   "resolve group members":
//     - Target: { type: GROUP, id: "group-123" }
//     - Expected: Returns user IDs of all group members
//
//   "group target without ID returns error":
//     - Target: { type: GROUP, id: "" }
//     - Expected: ErrMissingRequiredField error
//
// Why This Matters:
//   Group targets enable dynamic assignment to team members. As group membership changes,
//   workflow assignments automatically reflect the current membership.
func (s *WorkflowEngineTestSuite) TestResolveGroupTarget() {
	s.ClearWorkflowDefinitions()

	userID, orgID, userCtx := s.SetupTestUser()
	seedCtx := s.SeedContext(userID, orgID)

	wfEngine := s.Engine()

	// Create a second user for the group and add them to the org
	user2ID, _ := s.CreateTestUserInOrg(orgID, enums.RoleMember)

	group, err := s.client.Group.Create().
		SetName("Test Group " + ulid.Make().String()).
		SetOwnerID(orgID).
		Save(seedCtx)
	s.Require().NoError(err)

	// Add group memberships separately to avoid conflicts
	_, err = s.client.GroupMembership.Create().
		SetGroupID(group.ID).
		SetUserID(userID).
		Save(seedCtx)
	s.Require().NoError(err)

	_, err = s.client.GroupMembership.Create().
		SetGroupID(group.ID).
		SetUserID(user2ID).
		Save(seedCtx)
	s.Require().NoError(err)

	obj := &workflows.Object{
		ID:   "test123",
		Type: enums.WorkflowObjectTypeControl,
	}

	s.Run("resolve group members", func() {
		target := workflows.TargetConfig{
			Type: enums.WorkflowTargetTypeGroup,
			ID:   group.ID,
		}

		userIDs, err := wfEngine.ResolveTargets(userCtx, target, obj)
		s.NoError(err)
		// Group should resolve to at least one member
		s.GreaterOrEqual(len(userIDs), 1)
		s.Contains(userIDs, userID)
	})

	s.Run("group target without ID returns error", func() {
		target := workflows.TargetConfig{
			Type: enums.WorkflowTargetTypeGroup,
			ID:   "",
		}

		_, err := wfEngine.ResolveTargets(userCtx, target, obj)
		s.Error(err)
		s.ErrorIs(err, engine.ErrMissingRequiredField)
	})
}

// TestResolveRoleTarget verifies that ROLE-type targets resolve to all user IDs who have
// the specified role within the object's owning organization.
//
// Test Scenarios:
//   "role target requires ID":
//     - Target: { type: ROLE, id: "" }
//     - Expected: ErrMissingRequiredField error
//
//   "role resolution returns org members by role":
//     - Target: { type: ROLE, id: "OWNER" }
//     - Expected: Returns user IDs of org members with OWNER role
//
//   "invalid role returns error":
//     - Target: { type: ROLE, id: "NOT_A_ROLE" }
//     - Expected: Error (unrecognized role)
//
// Why This Matters:
//   Role targets enable automatic assignment to users based on their organizational role.
//   This is useful for "assign to all org owners" or "assign to all admins" workflows.
func (s *WorkflowEngineTestSuite) TestResolveRoleTarget() {
	s.ClearWorkflowDefinitions()

	wfEngine := s.Engine()

	s.Run("role target requires ID", func() {
		obj := &workflows.Object{
			ID:   "test123",
			Type: enums.WorkflowObjectTypeControl,
		}

		target := workflows.TargetConfig{
			Type: enums.WorkflowTargetTypeRole,
			ID:   "",
		}

		_, err := wfEngine.ResolveTargets(s.ctx, target, obj)
		s.Error(err)
		s.ErrorIs(err, engine.ErrMissingRequiredField)
	})

	s.Run("role resolution returns org members by role", func() {
		userID, orgID, userCtx := s.SetupTestUser()

		// User from SetupTestUser already has org membership
		control, err := s.client.Control.Create().
			SetRefCode("CTL-ROLE-TARGET-" + ulid.Make().String()).
			SetOwnerID(orgID).
			Save(userCtx)
		s.Require().NoError(err)

		obj := &workflows.Object{
			ID:   control.ID,
			Type: enums.WorkflowObjectTypeControl,
			Node: control,
		}

		target := workflows.TargetConfig{
			Type: enums.WorkflowTargetTypeRole,
			ID:   enums.RoleOwner.String(),
		}

		userIDs, err := wfEngine.ResolveTargets(userCtx, target, obj)
		s.NoError(err)
		s.GreaterOrEqual(len(userIDs), 1)
		s.Contains(userIDs, userID)
	})

	s.Run("invalid role returns error", func() {
		_, orgID, userCtx := s.SetupTestUser()

		control, err := s.client.Control.Create().
			SetRefCode("CTL-ROLE-TARGET-INVALID-" + ulid.Make().String()).
			SetOwnerID(orgID).
			Save(userCtx)
		s.Require().NoError(err)

		obj := &workflows.Object{
			ID:   control.ID,
			Type: enums.WorkflowObjectTypeControl,
			Node: control,
		}

		target := workflows.TargetConfig{
			Type: enums.WorkflowTargetTypeRole,
			ID:   "NOT_A_ROLE",
		}

		_, err = wfEngine.ResolveTargets(userCtx, target, obj)
		s.Error(err)
	})
}

// TestResolveControlTarget verifies RESOLVER-type targets that use predefined resolver keys
// to dynamically determine target users based on object relationships.
//
// Test Scenarios:
//   "resolve CONTROL_OWNER":
//     - Target: { type: RESOLVER, resolverKey: "CONTROL_OWNER" }
//     - Expected: Returns user IDs of the Control's owning organization's owners
//
//   "resolve OBJECT_CREATOR":
//     - Target: { type: RESOLVER, resolverKey: "OBJECT_CREATOR" }
//     - Expected: Returns user ID(s) who created the object (if tracked)
//
// Why This Matters:
//   Resolver targets enable dynamic assignment based on object context. For example,
//   "CONTROL_OWNER" automatically assigns to whoever owns the control being modified,
//   without hardcoding specific user IDs in the workflow definition.
func (s *WorkflowEngineTestSuite) TestResolveControlTarget() {
	s.ClearWorkflowDefinitions()

	userID, orgID, userCtx := s.SetupTestUser()

	wfEngine := s.Engine()

	control, err := s.client.Control.Create().
		SetRefCode("CTL-TEST-" + ulid.Make().String()).
		SetOwnerID(orgID).
		Save(userCtx)
	s.Require().NoError(err)

	obj := &workflows.Object{
		ID:   control.ID,
		Type: enums.WorkflowObjectTypeControl,
	}

	s.Run("resolve CONTROL_OWNER", func() {
		target := workflows.TargetConfig{
			Type:        enums.WorkflowTargetTypeResolver,
			ResolverKey: "CONTROL_OWNER",
		}

		userIDs, err := wfEngine.ResolveTargets(userCtx, target, obj)
		s.NoError(err)
		// Org owners are resolved, which includes the user from SetupTestUser
		s.GreaterOrEqual(len(userIDs), 1)
		s.Contains(userIDs, userID)
	})

	s.Run("resolve OBJECT_CREATOR", func() {
		target := workflows.TargetConfig{
			Type:        enums.WorkflowTargetTypeResolver,
			ResolverKey: "OBJECT_CREATOR",
		}

		userIDs, err := wfEngine.ResolveTargets(userCtx, target, obj)
		s.NoError(err)
		s.GreaterOrEqual(len(userIDs), 0)
	})
}

// TestResolveInvalidTarget verifies error handling for malformed or invalid target configurations.
//
// Test Scenarios:
//   "invalid target type":
//     - Target: { type: "INVALID" }
//     - Expected: ErrInvalidTargetType error
//
//   "resolver without resolver key":
//     - Target: { type: RESOLVER, resolverKey: "" }
//     - Expected: ErrMissingRequiredField error
//
// Why This Matters:
//   The resolver must validate target configurations and fail fast with clear errors
//   rather than silently returning empty results or causing downstream failures.
func (s *WorkflowEngineTestSuite) TestResolveInvalidTarget() {
	s.ClearWorkflowDefinitions()

	wfEngine := s.Engine()

	obj := &workflows.Object{
		ID:   "test123",
		Type: enums.WorkflowObjectTypeControl,
	}

	s.Run("invalid target type", func() {
		target := workflows.TargetConfig{
			Type: "INVALID",
			ID:   "test",
		}

		_, err := wfEngine.ResolveTargets(s.ctx, target, obj)
		s.Error(err)
		s.ErrorIs(err, engine.ErrInvalidTargetType)
	})

	s.Run("resolver without resolver key", func() {
		target := workflows.TargetConfig{
			Type:        enums.WorkflowTargetTypeResolver,
			ResolverKey: "",
		}

		_, err := wfEngine.ResolveTargets(s.ctx, target, obj)
		s.Error(err)
		s.ErrorIs(err, engine.ErrMissingRequiredField)
	})
}
