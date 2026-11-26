package rule

import (
	"context"
	"errors"
	"slices"
	"strings"

	"entgo.io/ent"
	"github.com/samber/lo"

	"github.com/stripe/stripe-go/v86"
	"github.com/theopenlane/iam/auth"

	"github.com/theopenlane/logx"

	"github.com/theopenlane/core/v2/internal/ent/generated"
	"github.com/theopenlane/core/v2/internal/ent/generated/organizationsetting"
	"github.com/theopenlane/core/v2/internal/ent/generated/orgsubscription"
	"github.com/theopenlane/core/v2/internal/ent/generated/privacy"
	"github.com/theopenlane/core/v2/internal/ent/privacy/utils"
)

var (
	errNoPaymentMethodAttached = errors.New("A valid payment method is required to create tokens. Contact your organization admin to add one in billing.") //nolint:staticcheck,revive
)

type paymentRequiredMutation interface {
	OrganizationsIDs() []string
}

// RequirePaymentMethod makes sure the organization has a payment method ( card or any other)
// added to stripe already
func RequirePaymentMethod() privacy.MutationRuleFunc {
	return privacy.MutationRuleFunc(func(ctx context.Context, m ent.Mutation) error {
		client := generated.FromContext(ctx)

		caller, ok := auth.CallerFromContext(ctx)
		if !ok {
			return auth.ErrNoAuthUser
		}

		if !utils.PaymentMethodCheckRequired(client) || caller.HasAny(auth.CapSystemAdmin|auth.CapOrgSupport) ||
			auth.IsAPITokenAuthentication(ctx) {
			return privacy.Skip
		}

		// early return
		_, domain, hasEmailAt := strings.Cut(caller.SubjectEmail, "@")
		if hasEmailAt && slices.Contains(client.EntConfig.Billing.BypassEmailDomains, domain) {
			return privacy.Skip
		}

		orgIDs := []string{caller.OrganizationID}

		if orgMutation, ok := m.(paymentRequiredMutation); ok {
			orgIDs = append(orgIDs, orgMutation.OrganizationsIDs()...)
		}

		orgIDs = lo.Uniq(orgIDs)

		queryCtx := auth.WithInternalReadCrossOrgContext(ctx)

		// get all orgs with a payment method. get the set difference between all the orgs provided in the mutation
		// and these paid ones then we can check those to see if they have an active sub.
		orgsWithPaymentMethod, err := client.OrganizationSetting.Query().
			Where(
				organizationsetting.OrganizationIDIn(orgIDs...),
				organizationsetting.PaymentMethodAdded(true),
			).
			Select(organizationsetting.FieldOrganizationID).
			Strings(queryCtx)
		if err != nil {
			logx.FromContext(ctx).Err(err).Msg("failed to fetch organization from db")

			return err
		}

		missingOrgIDs := lo.Without(orgIDs, orgsWithPaymentMethod...)
		if len(missingOrgIDs) == 0 {
			return privacy.Skip
		}

		if client.EntitlementManager == nil {
			return errNoPaymentMethodAttached
		}

		activeOrgs, err := client.OrgSubscription.Query().
			Where(
				orgsubscription.OwnerIDIn(missingOrgIDs...),
				orgsubscription.StripeSubscriptionStatusEQ(string(stripe.SubscriptionStatusActive)),
			).
			Count(queryCtx)
		if err != nil || activeOrgs < len(missingOrgIDs) {
			return errNoPaymentMethodAttached
		}

		logx.FromContext(ctx).Info().Msg("org has no payment method, but their trial is active, skipping requiring payment method")

		return privacy.Skip
	})
}
