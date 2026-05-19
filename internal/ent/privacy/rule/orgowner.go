package rule

import (
	"context"
	"strings"

	"errors"

	"entgo.io/ent"
	"entgo.io/ent/dialect/sql"
	"entgo.io/ent/privacy"
	"github.com/gertd/go-pluralize"
	"github.com/stoewer/go-strcase"
	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/generated/organization"
	"github.com/theopenlane/core/internal/ent/generated/orgmembership"
	"github.com/theopenlane/core/internal/ent/generated/trustcenter"
	"github.com/theopenlane/core/internal/ent/privacy/utils"
	"github.com/theopenlane/core/pkg/logx"
	"github.com/theopenlane/iam/auth"
	"github.com/theopenlane/iam/fgax"
)

var skipperType = map[string]struct{}{
	"Onboarding":             {},
	"User":                   {},
	"UserSetting":            {},
	"EmailVerificationToken": {},
	"PasswordResetToken":     {},
	"TFASetting":             {},
	"Webauthn":               {},
}

// DenyIfNotInOrganization runs to ensure the object being updated is part of the user's authorized organization
func DenyIfNotInOrganization() privacy.MutationRule {
	return privacy.MutationRuleFunc(func(ctx context.Context, m ent.Mutation) error {
		actor, ok := auth.CallerFromContext(ctx)
		if !ok || actor == nil || actor.IsAnonymous() {
			logx.FromContext(ctx).Error().Msg("unable to get caller from context")

			return auth.ErrNoAuthUser
		}

		orgID := actor.OrganizationID

		if _, ok := skipperType[m.Type()]; ok {
			return privacy.Skip
		}

		// History happens automatically, there are no external mutations to create history records
		if strings.Contains(m.Type(), "History") {
			return privacy.Allow
		}

		orgMut, ok := m.(*generated.OrganizationMutation)
		if ok {
			if m.Op().Is(ent.OpCreate) {
				// check if the parent org is set first
				parentOrgID, ok := orgMut.ParentOrganizationID()
				if ok && parentOrgID != "" {
					if err := checkOrgAccess(ctx, fgax.CanView, parentOrgID); errors.Is(err, privacy.Allow) {
						return nil
					}

					return privacy.Denyf("user does not have access to the parent organization")
				}

				// if there is no parent org, allow
				return privacy.Skip
			}

			objID, ok := orgMut.ID()
			if !ok || orgID == "" {
				return privacy.Skip
			}

			if err := EnsureObjectInOrganization(ctx, m, m.Type(), objID, orgID); errors.Is(err, privacy.Deny) {
				return err
			}

			return privacy.Skip
		}

		membershipMutation, ok := m.(*generated.OrgMembershipMutation)
		if ok {
			membershipID, ok := membershipMutation.ID()
			if !ok || membershipID == "" {
				return privacy.Skip
			}

			orgMembership, err := membershipMutation.Client().OrgMembership.Get(ctx, membershipID)
			if err != nil {
				return privacy.Skipf("unable to get org membership: %v", err)
			}

			if err := EnsureObjectInOrganization(ctx, m, orgmembership.Label, orgMembership.ID, orgID); errors.Is(err, privacy.Deny) {
				return err
			}

			return privacy.Skip
		}

		groupMemberMutation, ok := m.(*generated.GroupMembershipMutation)
		if ok {
			groupMemberID, ok := groupMemberMutation.ID()
			if !ok || groupMemberID == "" {
				return privacy.Skip
			}

			groupMembership, err := groupMemberMutation.Client().GroupMembership.Get(ctx, groupMemberID)
			if err != nil {
				return privacy.Skipf("unable to get group membership: %v", err)
			}

			if err := EnsureObjectInOrganization(ctx, m, "group", groupMembership.GroupID, orgID); errors.Is(err, privacy.Deny) {
				return err
			}

			return privacy.Skip
		}

		if m.Op().Is(ent.OpCreate) {
			return privacy.Skip
		}

		mut, ok := m.(utils.GenericMutation)
		if !ok {
			return privacy.Skip
		}

		id, ok := mut.ID()
		if !ok {
			return privacy.Skip
		}

		_, okOrg := m.(utils.OrgOwnedMutation)
		if okOrg {
			// ensure the object being mutated is in the organization specified in the owner_id field
			if err := EnsureObjectInOrganization(ctx, m, m.Type(), id, orgID); errors.Is(err, privacy.Deny) {
				return err
			}

			return privacy.Skip
		}

		// ensure the object being mutated is in the organization specified in the owner_id field
		if err := EnsureTrustCenterInOrganization(ctx, m, orgID); errors.Is(err, privacy.Deny) {
			return err
		}

		return privacy.Skip
	})
}

// EnsureObjectInOrganization checks if the object is in the organization
func EnsureObjectInOrganization(ctx context.Context, m ent.Mutation, objectType string, objectID, orgID string) error {
	// also ensure the id is part of the organization
	mut, ok := m.(utils.GenericMutation)
	if !ok {
		return privacy.Denyf("unable to determine access")
	}

	// check view access to the organization instead if the object is an organization
	if strings.EqualFold(objectType, organization.Label) {
		if objectID != "" && orgID != objectID {
			return privacy.Denyf("user does not have access to the requested organization")
		}

		if err := CheckCurrentOrgAccess(ctx, m, fgax.CanView); errors.Is(err, privacy.Allow) {
			return nil
		}

		return privacy.Denyf("user does not have access to the requested organization")
	}

	if strings.EqualFold(objectType, orgmembership.Label) {
		if err := CheckCurrentOrgAccess(ctx, m, fgax.CanView); errors.Is(err, privacy.Allow) {
			return nil
		}

		return privacy.Denyf("user does not have access to the requested organization")

	}

	// check if the object is in the organization
	pluralObjectType := pluralize.NewClient().Plural(objectType)
	tableName := strcase.SnakeCase(pluralObjectType)

	// files are not org owned, they rely on parents
	if tableName == "files" {
		// files can be skipped from this, they still do a parent
		// organization unlike other objects
		return privacy.Skip
	}

	query := "SELECT EXISTS (SELECT 1 FROM " + tableName + " WHERE id = $1 and (owner_id = $2))"

	var rows sql.Rows
	if err := mut.Client().Driver().Query(ctx, query, []any{objectID, orgID}, &rows); err != nil {
		logx.FromContext(ctx).Error().Err(err).Str("object", tableName).Msg("failed to check for object in organization")

		return err
	}

	defer rows.Close()

	if rows.Next() {
		var exists bool
		if err := rows.Scan(&exists); err == nil && exists {
			return nil
		}
	}

	// fall back to deny if the object is not in the organization
	return privacy.Denyf("requested object not in organization")
}

// EnsureTrustCenterInOrganization checks if the object is in the organization
func EnsureTrustCenterInOrganization(ctx context.Context, m ent.Mutation, orgID string) error {
	trustCenterID := getTrustCenterIDFromMutation(ctx, m)
	if trustCenterID == "" {
		return privacy.Skip
	}

	// the function above already checks this is a trust
	// center mutation, no need to check again
	tcMutation, _ := m.(utils.TrustCenterMutation)

	ownerID, err := tcMutation.Client().TrustCenter.Query().Where(
		trustcenter.ID(trustCenterID),
	).Select(trustcenter.FieldOwnerID).String(ctx)
	if err != nil {
		return privacy.Denyf("requested object not in organization")
	}

	if orgID == ownerID {
		return privacy.Skip
	}

	// fall back to deny if the object is not in the organization
	return privacy.Denyf("requested object not in organization")
}
