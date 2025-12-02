package hooks

import (
	"context"
	"fmt"
	"slices"

	"entgo.io/ent"
	"github.com/theopenlane/iam/fgax"

	"github.com/theopenlane/core/pkg/logx"
	"github.com/theopenlane/ent/generated"
	"github.com/theopenlane/ent/generated/hook"
	"github.com/theopenlane/ent/generated/organization"
	"github.com/theopenlane/ent/generated/organizationsetting"
	"github.com/theopenlane/ent/generated/privacy"
	"github.com/theopenlane/ent/privacy/utils"
)

// HookOrganizationCreatePolicy is used on organization and organization setting creation mutations
// if the allowed email domains are set, it will create a conditional tuple that restricts access
// to the organization based on the email domain
func HookOrganizationCreatePolicy() ent.Hook {
	return hook.If(func(next ent.Mutator) ent.Mutator {
		return ent.MutateFunc(func(ctx context.Context, m ent.Mutation) (ent.Value, error) {
			retVal, err := next.Mutate(ctx, m)
			if err != nil {
				return nil, err
			}

			// setup vars before switch
			orgID := ""

			allowedDomains := []string{}

			var client *generated.Client

			switch m := m.(type) {
			case *generated.OrganizationSettingMutation:
				client = m.Client()
				allowedDomains, _ = m.AllowedEmailDomains()

				orgID, err = getOrgIDFromSettingMutation(ctx, m, retVal)
				if err != nil {
					// skip if its a not found error
					// a setting can be created without an organization
					if generated.IsNotFound(err) {
						return retVal, nil
					}

					return nil, err
				}
			case *generated.OrganizationMutation:
				orgID, err = GetObjectIDFromEntValue(retVal)
				if err != nil {
					return nil, err
				}

				settingID, ok := m.SettingID()
				if !ok || settingID == "" {
					return retVal, nil
				}

				client = m.Client()

				setting, err := client.OrganizationSetting.Query().
					Where(organizationsetting.ID(settingID)).
					Select("allowed_email_domains").Only(ctx)
				if err != nil {
					return nil, err
				}

				allowedDomains = setting.AllowedEmailDomains
			}

			// ensure we didn't get a nil slice from the database, fga doesn't like that
			if allowedDomains == nil {
				allowedDomains = []string{}
			}

			if client.EmailVerifier.IncludesFreeDomain(allowedDomains) {
				logx.FromContext(ctx).Warn().Strs("domains", allowedDomains).Msg("organization allowed email domains include free email domains")

				return nil, fmt.Errorf("%w: allowed email domains cannot include free email domains", ErrInvalidInput)
			}

			if err := updateOrgConditionalTuples(ctx, m, orgID, allowedDomains); err != nil {
				return nil, err
			}

			return retVal, nil
		})
	},
		hook.HasOp(ent.OpCreate),
	)
}

// HookOrganizationUpdatePolicy is used on organization setting mutations where the allowed email domains are set in the request
// it will update the conditional tuple that restricts access to the organization based on the email domain
func HookOrganizationUpdatePolicy() ent.Hook {
	return hook.If(func(next ent.Mutator) ent.Mutator {
		return hook.OrganizationSettingFunc(func(ctx context.Context, m *generated.OrganizationSettingMutation) (ent.Value, error) {
			retVal, err := next.Mutate(ctx, m)
			if err != nil {
				return nil, err
			}

			orgID, err := getOrgIDFromSettingMutation(ctx, m, retVal)
			if err != nil {
				return nil, err
			}

			allowedEmailDomains, okSet := m.AllowedEmailDomains()
			if m.EmailVerifier.IncludesFreeDomain(allowedEmailDomains) {
				logx.FromContext(ctx).Warn().Strs("domains", allowedEmailDomains).Msg("organization allowed email domains include free email domains")

				return nil, fmt.Errorf("%w: allowed email domains cannot include free email domains", ErrInvalidInput)
			}

			okClear := m.AllowedEmailDomainsCleared()

			appendedDomains, okAppend := m.AppendedAllowedEmailDomains()

			if m.EmailVerifier.IncludesFreeDomain(appendedDomains) {
				logx.FromContext(ctx).Warn().Strs("domains", appendedDomains).Msg("organization allowed email domains include free email domains")

				return nil, fmt.Errorf("%w: allowed email domains cannot include free email domains", ErrInvalidInput)
			}

			var domainUpdates []string

			switch {
			case okSet:
				domainUpdates = allowedEmailDomains
			case okClear:
				domainUpdates = []string{}
			case okAppend:
				originalDomains, err := m.OldAllowedEmailDomains(ctx)
				if err != nil {
					return nil, err
				}

				domainUpdates = slices.Concat(originalDomains, appendedDomains)
			default:
				// we shouldn't get here because the hook is only called when the allowed email domains are set
				// but if we do, just return
				return retVal, nil
			}

			// update the conditional tuples with the new set of domains
			if err := updateOrgConditionalTuples(ctx, m, orgID, domainUpdates); err != nil {
				return nil, err
			}

			return retVal, nil
		})
	},
		hook.And(
			hook.Or(
				hook.HasFields("allowed_email_domains"),
				hook.HasAddedFields("allowed_email_domains"),
				hook.HasClearedFields("allowed_email_domains"),
			),
			hook.HasOp(ent.OpUpdateOne|ent.OpUpdate),
		),
	)
}

// updateOrgConditionalTuples will update (or create) a conditional tuple for the organization
// that restricts access based on the email domain
// the tuple will look like the following, where the allowed_domains are the email domains that are allowed
// if the list is empty, then all domains are allowed
//
// user: organization:openlane#member
// relation: access
// object: organization:openlane
// condition:
//
//	name: email_domain_allowed
//	context:
//	  allowed_domains: []
func updateOrgConditionalTuples(ctx context.Context, m ent.Mutation, orgID string, allowedEmailDomains []string) error {
	// create the tuple request, this is a self-referential tuple so the object and subject are the same
	tk := fgax.TupleRequest{
		ObjectID:         orgID,
		ObjectType:       generated.TypeOrganization,
		SubjectID:        orgID,
		SubjectType:      generated.TypeOrganization,
		SubjectRelation:  fgax.MemberRelation,
		Relation:         utils.OrgAccessCheckRelation,
		ConditionName:    utils.OrgEmailConditionName,
		ConditionContext: utils.NewOrganizationConditionContext(allowedEmailDomains),
	}

	if _, err := utils.AuthzClient(ctx, m).UpdateConditionalTupleKey(ctx, fgax.GetTupleKey(tk)); err != nil {
		logx.FromContext(ctx).Error().Err(err).Msg("failed to create org access restriction tuple")

		return err
	}

	return nil
}

// getGroupIDFromSettingMutation returns the group ID(s) from the group setting mutation or return value
func getOrgIDFromSettingMutation(ctx context.Context, m *generated.OrganizationSettingMutation, retVal any) (string, error) {
	// if we have it just return it
	orgID, ok := m.OrganizationID()
	if ok && orgID != "" {
		return orgID, nil
	}

	// otherwise get from the settings
	var (
		err       error
		settingID string
	)

	switch m.Op() {
	case ent.OpCreate:
		settingID, err = GetObjectIDFromEntValue(retVal)
		if err != nil {
			return "", err
		}
	case ent.OpUpdateOne, ent.OpUpdate:
		settingID, _ = m.ID()
	}

	// allow the retrieval, which may happen before the tuples are created
	allowCtx := privacy.DecisionContext(ctx, privacy.Allow)

	return m.Client().Organization.Query().
		Where(organization.HasSettingWith(organizationsetting.ID(settingID))).
		OnlyID(allowCtx)
}
