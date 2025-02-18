package hooks

import (
	"context"
	"slices"

	"entgo.io/ent"
	"github.com/rs/zerolog/log"
	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/generated/hook"
	"github.com/theopenlane/core/internal/ent/generated/organization"
	"github.com/theopenlane/core/internal/ent/generated/organizationsetting"
	"github.com/theopenlane/core/internal/ent/generated/privacy"
	"github.com/theopenlane/core/internal/ent/privacy/utils"
	"github.com/theopenlane/iam/fgax"
)

func HookOrganizationCreatePolicy() ent.Hook {
	return hook.If(func(next ent.Mutator) ent.Mutator {
		return ent.MutateFunc(func(ctx context.Context, m ent.Mutation) (ent.Value, error) {
			retVal, err := next.Mutate(ctx, m)
			if err != nil {
				return nil, err
			}

			orgID := ""

			allowedDomains := []string{}
			switch m := m.(type) {
			case *generated.OrganizationSettingMutation:
				allowedDomains, _ = m.AllowedEmailDomains()
				orgID, err = getOrgIDFromSettingMutation(ctx, m, retVal)
				if err != nil {
					//  skip if its a not found error
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
				if err != nil {
					return nil, err
				}

				if !ok || settingID == "" {
					return retVal, nil
				}

				setting, err := m.Client().OrganizationSetting.Query().
					Where(organizationsetting.ID(settingID)).
					Select("allowed_email_domains").Only(ctx)
				if err != nil {
					return nil, err
				}

				allowedDomains = setting.AllowedEmailDomains
			}

			if allowedDomains == nil {
				allowedDomains = []string{}
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

			domainUpdates := []string{}
			allowedEmailDomains, okSet := m.AllowedEmailDomains()
			okClear := m.AllowedEmailDomainsCleared()
			appendedDomains, okAppend := m.AppendedAllowedEmailDomains()

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
				return retVal, nil
			}

			// we should always have an orgID on update
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

	tk := fgax.TupleRequest{
		ObjectID:        orgID,
		ObjectType:      generated.TypeOrganization,
		SubjectType:     generated.TypeOrganization,
		SubjectID:       orgID,
		SubjectRelation: fgax.MemberRelation,
		Relation:        "access",
		ConditionName:   "email_domains_allowed",
		ConditionContext: &map[string]any{
			"allowed_domains": allowedEmailDomains,
		},
	}

	if _, err := utils.AuthzClient(ctx, m).UpdateConditionalTupleKey(ctx, fgax.GetTupleKey(tk)); err != nil {
		log.Error().Err(err).Msg("failed to create org access restriction tuple")

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

	// allow the retrieval, which  may happen before the tuples are created
	allowCtx := privacy.DecisionContext(ctx, privacy.Allow)
	return m.Client().Organization.Query().
		Where(organization.HasSettingWith(organizationsetting.ID(settingID))).
		OnlyID(allowCtx)
}
