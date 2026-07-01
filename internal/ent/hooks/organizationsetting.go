package hooks

import (
	"context"
	"fmt"
	"strings"
	"time"

	"entgo.io/ent"

	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/generated/hook"
	"github.com/theopenlane/core/internal/ent/generated/organization"
	"github.com/theopenlane/core/internal/ent/generated/organizationsetting"
	"github.com/theopenlane/core/internal/ent/generated/privacy"
	emaildef "github.com/theopenlane/core/internal/integrations/definitions/email"
	"github.com/theopenlane/core/pkg/logx"
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

			allowedDomains := []string{}

			var client *generated.Client

			switch m := m.(type) {
			case *generated.OrganizationSettingMutation:
				client = m.Client()
				allowedDomains, _ = m.AllowedEmailDomains()
			case *generated.OrganizationMutation:
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

			allowedEmailDomains, okSet := m.AllowedEmailDomains()
			if okSet && m.EmailVerifier.IncludesFreeDomain(allowedEmailDomains) {
				logx.FromContext(ctx).Warn().Strs("domains", allowedEmailDomains).Msg("organization allowed email domains include free email domains")

				return nil, fmt.Errorf("%w: allowed email domains cannot include free email domains", ErrInvalidInput)
			}

			appendedDomains, ok := m.AppendedAllowedEmailDomains()
			if ok && m.EmailVerifier.IncludesFreeDomain(appendedDomains) {
				logx.FromContext(ctx).Warn().Strs("domains", appendedDomains).Msg("organization allowed email domains include free email domains")

				return nil, fmt.Errorf("%w: allowed email domains cannot include free email domains", ErrInvalidInput)
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

// HookBillingEmailChange is triggered when the billing_email field is updated on an organization setting.
func HookBillingEmailChange() ent.Hook {
	return hook.If(func(next ent.Mutator) ent.Mutator {
		return hook.OrganizationSettingFunc(func(ctx context.Context, m *generated.OrganizationSettingMutation) (ent.Value, error) {
			oldEmail, err := m.OldBillingEmail(ctx)
			if err != nil {
				logx.FromContext(ctx).Error().Err(err).Msg("failed to get old billing email")
				return nil, err
			}

			newEmail, _ := m.BillingEmail()

			retVal, err := next.Mutate(ctx, m)
			if err != nil {
				return nil, err
			}

			// only send the email if the billing email actually changed
			// or it is not the first time the billing email is being updated
			if oldEmail == "" || strings.EqualFold(newEmail, oldEmail) {
				return retVal, nil
			}

			orgID, err := getOrgIDFromSettingMutation(ctx, m, retVal)
			if err != nil {
				logx.FromContext(ctx).Error().Err(err).Msg("failed to get organization ID for billing email change notification")
				return retVal, nil
			}

			orgName, err := organizationDisplayNameByID(ctx, m.Client(), orgID)
			if err != nil {
				logx.FromContext(ctx).Error().Err(err).Msg("failed to get organization for billing email change notification")
				return retVal, nil
			}

			if err := sendBillingEmailChangeNotifications(ctx, m.Client(), orgName, oldEmail, newEmail); err != nil {
				logx.FromContext(ctx).Error().Err(err).Msg("failed to send billing email change notifications")
			}

			return retVal, nil
		})
	},
		hook.And(
			hook.HasFields("billing_email"),
			hook.HasOp(ent.OpUpdateOne),
		),
	)
}

func sendBillingEmailChangeNotifications(ctx context.Context, client *generated.Client, orgName, previousEmail, newEmail string) error {
	changedAt := time.Now().UTC()

	for _, currentEmail := range []string{previousEmail, newEmail} {
		if currentEmail == "" {
			continue
		}

		if err := sendSystemEmail(ctx, client, emaildef.BillingEmailChangedOp.Name(), emaildef.BillingEmailChangedEmail{
			RecipientInfo:   emaildef.RecipientInfo{Email: currentEmail},
			OrgName:         orgName,
			OldBillingEmail: previousEmail,
			NewBillingEmail: newEmail,
			ChangedAt:       changedAt,
		}); err != nil {
			logx.FromContext(ctx).Error().Err(err).Msg("failed to send billing email change notification")
		}
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
