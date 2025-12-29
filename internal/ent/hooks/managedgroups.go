package hooks

import (
	"context"

	"entgo.io/ent"
	"github.com/theopenlane/iam/auth"
	"github.com/theopenlane/utils/contextx"

	"github.com/theopenlane/core/common/enums"
	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/generated/group"
	"github.com/theopenlane/core/internal/ent/generated/groupmembership"
	"github.com/theopenlane/core/internal/ent/generated/predicate"
	"github.com/theopenlane/core/internal/ent/generated/privacy"
	"github.com/theopenlane/core/pkg/logx"
)

// ManagedContextKey is the context key name for managed group updates
type ManagedContextKey struct{}

const (
	// AdminsGroup is the group name for all organization admins and owner, these users have full read and write access in the organization
	AdminsGroup = "Admins"
	// ViewersGroup is the group name for all organization members that only have view access in the organization
	ViewersGroup = "Viewers"
	// AllMembersGroup is the group name for all members of the organization, no matter their role
	AllMembersGroup = "All Members"
)

// defaultGroups are the default groups created for an organization that are managed by the system
var defaultGroups = map[string]string{
	AdminsGroup:     "Openlane managed group containing all organization admins with full access",
	ViewersGroup:    "Openlane managed group containing all organization members with only view access",
	AllMembersGroup: "Openlane managed group containing all members of the organization",
}

// generateOrganizationGroups creates the default groups for an organization that are managed by Openlane
// this includes the Admins, Viewers, and All Members groups where users are automatically added based on their role
func generateOrganizationGroups(ctx context.Context, m *generated.OrganizationMutation, orgID string) error {
	// skip group creation for personal orgs
	if isPersonal, _ := m.PersonalOrg(); isPersonal {
		logx.FromContext(ctx).Debug().Msg("skipping group creation for personal org")

		return nil
	}

	builders := make([]*generated.GroupCreate, 0, len(defaultGroups))

	for name, desc := range defaultGroups {
		tags := []string{"managed"}

		switch name {
		case AdminsGroup:
			tags = append(tags, "admins")
		case ViewersGroup:
			tags = append(tags, "viewers")
		}

		groupInput := generated.CreateGroupInput{
			Name:        name,
			Description: &desc,
			Tags:        tags,
		}

		builders = append(builders, m.Client().Group.Create().
			SetInput(groupInput).
			SetIsManaged(true).
			SetOwnerID(orgID),
		)
	}

	groups, err := m.Client().Group.CreateBulk(builders...).Save(ctx)
	if err != nil {
		logx.FromContext(ctx).Error().Err(err).Msg("error creating system managed groups")

		return err
	}

	// add group member to managed groups
	for _, g := range groups {
		if g.Name == ViewersGroup {
			continue
		}

		userID, err := auth.GetSubjectIDFromContext(ctx)
		if err != nil {
			logx.FromContext(ctx).Error().Err(err).Msg("error getting user ID from context")
			return err
		}

		input := generated.CreateGroupMembershipInput{
			Role:    &enums.RoleMember,
			UserID:  userID,
			GroupID: g.ID,
		}

		if err := m.Client().GroupMembership.Create().SetInput(input).Exec(ctx); err != nil {
			logx.FromContext(ctx).Error().Err(err).Msg("error adding user to managed group")
			return err
		}
	}

	logx.FromContext(ctx).Debug().Str("organization", orgID).Msg("created system managed groups")

	return nil
}

// OrgMember is a struct to hold the org member details
type OrgMember struct {
	// UserID is the user ID of the org member
	UserID string
	// Role is the role of the org member
	Role enums.Role
	// OrgID is the organization ID of the org member
	OrgID string
}

// updateManagedGroupMembers groups adds or removes the org members to the managed system groups
func updateManagedGroupMembers(ctx context.Context, m *generated.OrgMembershipMutation) error {
	// set a context key to indicate that this is a managed group update
	// and allowed to skip the check for managed groups
	managedCtx := contextx.With(ctx, ManagedContextKey{})

	op := m.Op()

	var (
		orgMember OrgMember
		ok        bool
	)

	orgMember.Role, ok = m.Role()
	if !ok && op.Is(ent.OpCreate) {
		// default to member role for new org members
		orgMember.Role = enums.RoleMember
	}

	orgMember.UserID, _ = m.UserID()
	orgMember.OrgID, _ = m.OrganizationID()

	// if role or user is empty, get the org membership details
	if orgMember.Role == "" || orgMember.UserID == "" {
		mID, _ := m.ID()

		om, err := m.Client().OrgMembership.Get(ctx, mID)
		if err != nil {
			logx.FromContext(ctx).Error().Err(err).Msg("error getting org membership")
			return err
		}

		if orgMember.Role == "" {
			orgMember.Role = om.Role
		}

		if orgMember.UserID == "" {
			orgMember.UserID = om.UserID
		}

		orgMember.OrgID = om.OrganizationID
	}

	switch op {
	case ent.OpCreate:
		return addToManagedGroups(managedCtx, m, orgMember)
	default:
		// deletes are handled by the cascade delete hook
		// which will delete the group memberships when the org membership is deleted
		if !isDeleteOp(ctx, m) {
			return updateManagedGroups(managedCtx, m, orgMember)
		}
	}

	return nil
}

// updateManagedGroups updates the managed groups based on the role of the user for update requests
func updateManagedGroups(ctx context.Context, m *generated.OrgMembershipMutation, om OrgMember) error {
	oldRole, _ := m.OldRole(ctx)

	if oldRole != om.Role {
		removeOm := om
		removeOm.Role = oldRole

		if err := removeFromManagedGroups(ctx, m, removeOm); err != nil {
			return err
		}

		return addToManagedGroups(ctx, m, om)
	}

	// if the role has not changed, do nothing
	return nil
}

// addToManagedGroups adds the user to the system managed groups based on their role on creation or update
func addToManagedGroups(ctx context.Context, m *generated.OrgMembershipMutation, om OrgMember) error {
	switch om.Role {
	case enums.RoleMember:
		if err := addMemberToManagedGroup(ctx, m, om, ViewersGroup); err != nil {
			return err
		}
	case enums.RoleAdmin:
		if err := addMemberToManagedGroup(ctx, m, om, AdminsGroup); err != nil {
			return err
		}
	}

	// add all users to the all users group
	if m.Op() == ent.OpCreate {
		return addMemberToManagedGroup(ctx, m, om, AllMembersGroup)
	}

	return nil
}

// removeFromManagedGroups removes the user from the system managed groups when their role changes
// users that are removed from the organization are handled by the cascade delete
func removeFromManagedGroups(ctx context.Context, m *generated.OrgMembershipMutation, om OrgMember) error {
	switch om.Role {
	case enums.RoleMember:
		if err := removeMemberFromManagedGroup(ctx, m, om, ViewersGroup); err != nil {
			return err
		}
	case enums.RoleAdmin:
		if err := removeMemberFromManagedGroup(ctx, m, om, AdminsGroup); err != nil {
			return err
		}
	}

	return nil
}

// addMemberToManagedGroup adds the user to the system managed groups
func addMemberToManagedGroup(ctx context.Context, m *generated.OrgMembershipMutation, om OrgMember, groupName string) error {
	pred := []predicate.Group{
		group.IsManaged(true), // grab the managed group
		group.Name(groupName),
	}

	if om.OrgID != "" {
		pred = append(pred, group.OwnerID(om.OrgID))
	}

	// allow the request to bypass the privacy check
	allowCtx := privacy.DecisionContext(ctx, privacy.Allow)

	// get the group to update
	group, err := m.Client().Group.Query().Where(
		pred...,
	).Only(allowCtx)
	if err != nil {
		logx.FromContext(ctx).Error().Err(err).Msgf("error getting managed group: %s", groupName)
		return err
	}

	// Check if user is already a member of this group
	existingMembership, err := m.Client().GroupMembership.Query().
		Where(
			groupmembership.UserID(om.UserID),
			groupmembership.GroupID(group.ID),
		).
		Exist(allowCtx)
	if err != nil {
		logx.FromContext(ctx).Error().Err(err).Msg("error checking existing group membership")
		return err
	}

	// If user is already in the group, skip creation
	if existingMembership {
		logx.FromContext(ctx).Debug().Str("user_id", om.UserID).Str("group", groupName).Msg("user already in managed group, skipping")
		return nil
	}

	input := generated.CreateGroupMembershipInput{
		Role:    &enums.RoleMember,
		UserID:  om.UserID,
		GroupID: group.ID,
	}

	if err := m.Client().GroupMembership.Create().SetInput(input).Exec(allowCtx); err != nil {
		logx.FromContext(ctx).Error().Err(err).Msg("error adding user to managed group")
		return err
	}

	logx.FromContext(ctx).Debug().Str("user_id", om.UserID).Str("group", groupName).Msg("user added to managed group")

	return nil
}

func removeMemberFromManagedGroup(ctx context.Context, m *generated.OrgMembershipMutation, om OrgMember, groupName string) error {
	pred := []predicate.Group{
		group.IsManaged(true), // grab the managed group
		group.Name(groupName),
	}

	if om.OrgID != "" {
		pred = append(pred, group.OwnerID(om.OrgID))
	}

	// allow the request to bypass the privacy check
	allowCtx := privacy.DecisionContext(ctx, privacy.Allow)

	groupID, err := m.Client().Group.Query().Where(
		pred...,
	).OnlyID(allowCtx)
	if err != nil {
		logx.FromContext(ctx).Error().Err(err).Msgf("error getting managed group: %s", groupName)

		return err
	}

	groupMembershipID, err := m.Client().GroupMembership.Query().
		Where(groupmembership.GroupID(groupID),
			groupmembership.UserID(om.UserID),
			groupmembership.RoleEQ(enums.RoleMember)).OnlyID(allowCtx)
	if err != nil {
		if generated.IsNotFound(err) {
			logx.FromContext(ctx).Warn().Str("user_id", om.UserID).Str("group", groupName).Msg("user not found in managed group, nothing to delete")

			return nil
		}

		logx.FromContext(ctx).Error().Err(err).Msg("error getting group membership")

		return err
	}

	if err = m.Client().GroupMembership.DeleteOneID(groupMembershipID).Exec(allowCtx); err != nil {
		logx.FromContext(ctx).Error().Err(err).Msg("error removing user from managed group")

		return err
	}

	return nil
}
