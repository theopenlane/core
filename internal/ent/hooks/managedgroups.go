package hooks

import (
	"context"

	"entgo.io/ent"
	"github.com/rs/zerolog/log"
	"github.com/theopenlane/entx"
	"github.com/theopenlane/utils/contextx"

	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/generated/group"
	"github.com/theopenlane/core/internal/ent/generated/groupmembership"
	"github.com/theopenlane/core/internal/ent/generated/predicate"
	"github.com/theopenlane/core/internal/ent/generated/privacy"
	"github.com/theopenlane/core/pkg/enums"
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
		log.Debug().Msg("skipping group creation for personal org")

		return nil
	}

	builders := make([]*generated.GroupCreate, 0, len(defaultGroups))

	for name, desc := range defaultGroups {
		groupInput := generated.CreateGroupInput{
			Name:        name,
			Description: &desc,
		}

		builders = append(builders, m.Client().Group.Create().
			SetInput(groupInput).
			SetIsManaged(true).
			SetOwnerID(orgID),
		)
	}

	if err := m.Client().Group.CreateBulk(builders...).Exec(ctx); err != nil {
		log.Error().Err(err).Msg("error creating system managed groups")

		return err
	}

	log.Debug().Str("organization", orgID).Msg("created system managed groups")

	return nil
}

type OrgMember struct {
	UserID string
	Role   enums.Role
	OrgID  string
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
	if !ok && m.Op().Is(ent.OpCreate) {
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
			log.Error().Err(err).Msg("error getting org membership")
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
	case ent.OpDelete, ent.OpDeleteOne:
		return removeFromManagedGroups(managedCtx, m, orgMember)
	case ent.OpUpdate, ent.OpUpdateOne:
		if entx.CheckIsSoftDelete(managedCtx) {
			return removeFromManagedGroups(managedCtx, m, orgMember)
		}

		return updateManagedGroups(managedCtx, m, orgMember)
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

// removeFromManagedGroups removes the user from the system managed groups when they are removed from the organization or their role changes
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

	// remove from the all users group if they are removed from the organization
	if entx.CheckIsSoftDelete(ctx) {
		return removeMemberFromManagedGroup(ctx, m, om, AllMembersGroup)
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
		log.Error().Err(err).Msgf("error getting managed group: %s", groupName)
		return err
	}

	input := generated.CreateGroupMembershipInput{
		Role:    &enums.RoleMember,
		UserID:  om.UserID,
		GroupID: group.ID,
	}

	if err := m.Client().GroupMembership.Create().SetInput(input).Exec(allowCtx); err != nil {
		log.Error().Err(err).Msg("error adding user to managed group")
		return err
	}

	log.Debug().Str("user_id", om.UserID).Str("group", groupName).Msg("user added to managed group")

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
		log.Error().Err(err).Msgf("error getting managed group: %s", groupName)

		return err
	}

	groupMembershipID, err := m.Client().GroupMembership.Query().
		Where(groupmembership.GroupID(groupID),
			groupmembership.UserID(om.UserID),
			groupmembership.RoleEQ(enums.RoleMember)).OnlyID(allowCtx)
	if err != nil {
		if generated.IsNotFound(err) {
			log.Warn().Str("user_id", om.UserID).Str("group", groupName).Msg("user not found in managed group, nothing to delete")

			return nil
		}

		log.Error().Err(err).Msg("error getting group membership")

		return err
	}

	if err = m.Client().GroupMembership.DeleteOneID(groupMembershipID).Exec(allowCtx); err != nil {
		log.Error().Err(err).Msg("error removing user from managed group")

		return err
	}

	return nil
}
