package hooks

import (
	"context"

	"entgo.io/ent"
	"github.com/theopenlane/iam/auth"

	"github.com/theopenlane/core/common/enums"
	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/generated/group"
	"github.com/theopenlane/core/internal/ent/generated/groupmembership"
	"github.com/theopenlane/core/internal/ent/generated/orgmembership"
	"github.com/theopenlane/core/internal/ent/generated/predicate"
	"github.com/theopenlane/core/internal/ent/generated/privacy"
	"github.com/theopenlane/core/pkg/logx"
)

const (
	// AdminsGroup is the group name for all organization admins, super admins, and owners.
	// These users have full read and write access in the organization.
	AdminsGroup = "Admins"
	// ViewersGroup is the group name for all organization members that only have view access in the organization
	ViewersGroup = "Viewers"
	// AllMembersGroup is the group name for all members of the organization, no matter their role
	AllMembersGroup = "All Members"
)

// defaultGroups are the default groups created for an organization that are managed by the system
var defaultGroups = map[string]string{
	AdminsGroup:     "Openlane managed group containing all organization admins, super admins, and owners with full access",
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
	// NewRole is the role of the org member
	NewRole enums.Role
	// OldRow is the old role of the org member
	OldRole enums.Role
	// OrgID is the organization ID of the org member
	OrgID string
}

// updateManagedGroupMembers groups adds or removes the org members to the managed system groups
// all deletes are handled by the cascade delete hook and are not managed here
func updateManagedGroupMembers(ctx context.Context, m *generated.OrgMembershipMutation) error {
	// deletes are handled by the cascade delete hook, exit early
	if isDeleteOp(ctx, m) {
		return nil
	}

	// add managed group bypass capability to the caller for downstream hook checks
	const managedCaps = auth.CapBypassOrgFilter | auth.CapBypassManagedGroup
	var managedCtx context.Context
	if existingCaller, hasCaller := auth.CallerFromContext(ctx); hasCaller {
		managedCtx = auth.WithCaller(ctx, existingCaller.WithCapabilities(managedCaps))
	} else {
		managedCtx = auth.WithCaller(ctx, &auth.Caller{Capabilities: managedCaps})
	}

	orgMemberRole, ok := m.Role()
	if !ok {
		if !m.Op().Is(ent.OpCreate) {
			// role didn't change, nothing to update
			return nil
		}

		// default to member role for new org members
		orgMemberRole = enums.RoleMember
	}

	orgMemberUserID, _ := m.UserID()
	orgMemberOrgID, _ := m.OrganizationID()

	var omIDs []string

	switch m.Op() {
	case ent.OpCreate:
		orgMember := OrgMember{
			UserID:  orgMemberUserID,
			OrgID:   orgMemberOrgID,
			NewRole: orgMemberRole,
		}

		return addToManagedGroups(managedCtx, m, orgMember)
	case ent.OpUpdateOne:
		mID, _ := m.ID()
		omIDs = []string{mID}
	default:
		var err error
		omIDs, err = m.IDs(ctx)
		if err != nil {
			logx.FromContext(ctx).Error().Err(err).Msg("error getting org member updated ids")
			return err
		}
	}

	updatedOrgMembers, err := m.Client().OrgMembership.Query().Where(
		orgmembership.IDIn(omIDs...),
	).Select(orgmembership.FieldID, orgmembership.FieldRole, orgmembership.FieldUserID, orgmembership.FieldOrganizationID).All(ctx)
	if err != nil {
		logx.FromContext(ctx).Error().Err(err).Msg("error getting org membership")
		return err
	}

	orgMembers := make([]OrgMember, 0, len(updatedOrgMembers))

	for _, om := range updatedOrgMembers {
		orgMembers = append(orgMembers, OrgMember{
			UserID:  om.UserID,
			NewRole: orgMemberRole,
			OldRole: om.Role,
			OrgID:   om.OrganizationID,
		})
	}

	return updateManagedGroups(managedCtx, m, orgMembers)
}

// updateManagedGroups updates the managed groups based on the role of the user for update requests
func updateManagedGroups(ctx context.Context, m *generated.OrgMembershipMutation, oms []OrgMember) error {
	for _, om := range oms {
		if om.OldRole != om.NewRole {
			if err := removeFromManagedGroups(ctx, m, om); err != nil {
				return err
			}

			return addToManagedGroups(ctx, m, om)
		}
	}

	// if the role has not changed, do nothing
	return nil
}

// addToManagedGroups adds the user to the system managed groups based on their role on creation or update
func addToManagedGroups(ctx context.Context, m *generated.OrgMembershipMutation, om OrgMember) error {
	switch om.NewRole {
	case enums.RoleMember:
		if err := addMemberToManagedGroup(ctx, m, om, ViewersGroup); err != nil {
			return err
		}
	case enums.RoleAdmin, enums.RoleSuperAdmin, enums.RoleOwner:
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
	switch om.OldRole {
	case enums.RoleMember:
		if err := removeMemberFromManagedGroup(ctx, m, om, ViewersGroup); err != nil {
			return err
		}
	case enums.RoleAdmin, enums.RoleSuperAdmin, enums.RoleOwner:
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
