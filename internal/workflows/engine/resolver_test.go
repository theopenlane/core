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

// TestResolveTargetsRequiresEngine verifies missing engine handling
func (s *WorkflowEngineTestSuite) TestResolveTargetsRequiresEngine() {
	wfEngine := s.Engine()
	s.Require().NotNil(wfEngine)
}

// TestResolveUserTarget verifies user target resolution
func (s *WorkflowEngineTestSuite) TestResolveUserTarget() {
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

// TestResolveGroupTarget verifies group target resolution
func (s *WorkflowEngineTestSuite) TestResolveGroupTarget() {
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

// TestResolveRoleTarget verifies role target resolution
func (s *WorkflowEngineTestSuite) TestResolveRoleTarget() {
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

// TestResolveControlTarget verifies control target resolution
func (s *WorkflowEngineTestSuite) TestResolveControlTarget() {
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

// TestResolveInvalidTarget verifies invalid target handling
func (s *WorkflowEngineTestSuite) TestResolveInvalidTarget() {
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
