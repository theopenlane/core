package rule

import (
	"context"
	"errors"
	"slices"
	"strings"

	"entgo.io/ent"

	"github.com/stripe/stripe-go/v86"
	"github.com/theopenlane/iam/auth"

	"github.com/theopenlane/core/v2/internal/ent/generated"
	"github.com/theopenlane/core/v2/internal/ent/generated/organizationsetting"
	"github.com/theopenlane/core/v2/internal/ent/generated/orgsubscription"
	"github.com/theopenlane/core/v2/internal/ent/generated/privacy"
	"github.com/theopenlane/core/v2/internal/ent/privacy/utils"
	"github.com/theopenlane/core/v2/pkg/logx"
)

var (
	errNoPaymentMethodAttached = errors.New("A valid payment method is required to create tokens. Contact your organization admin to add one in billing.") //nolint:staticcheck,revive
)

// RequirePaymentMethod makes sure the organization has a payment method ( card or any other)
// added to stripe already
func RequirePaymentMethod() privacy.MutationRuleFunc {
	return privacy.MutationRuleFunc(func(ctx context.Context, _ ent.Mutation) error {
		client := generated.FromContext(ctx)

		caller, ok := auth.CallerFromContext(ctx)
		if !ok {
			return auth.ErrNoAuthUser
		}

		if !utils.PaymentMethodCheckRequired(client) || caller.Has(auth.CapSystemAdmin) {
			return privacy.Skip
		}

		orgSetting, err := client.OrganizationSetting.Query().
			Where(organizationsetting.OrganizationID(caller.OrganizationID)).
			Select(organizationsetting.FieldPaymentMethodAdded).
			Only(ctx)
		if err != nil {
			logx.FromContext(ctx).Err(err).Msg("failed to fetch organization from db")

			return err
		}

		if orgSetting.PaymentMethodAdded {
			// evaluate next rule
			return privacy.Skip
		}

		emailDomain := strings.SplitAfter(caller.SubjectEmail, "@")[1]

		if slices.Contains(client.EntConfig.Billing.BypassEmailDomains, emailDomain) {
			return privacy.Skip
		}

		// if enititlements enabled, additional check for active sub instead of requireming payment method
		if client.EntitlementManager != nil {
			// fallback to check on org sub status, an active status
			orgSubscription, err := client.OrgSubscription.Query().
				Where(orgsubscription.OwnerID(caller.OrganizationID)).
				Select(orgsubscription.FieldStripeSubscriptionStatus).
				Only(ctx)

			if err == nil && orgSubscription.StripeSubscriptionStatus == string(stripe.SubscriptionStatusActive) {
				logx.FromContext(ctx).Info().Msg("org has no payment method, but their trial is active, skipping requiring payment method")

				return privacy.Skip
			}
		}

		return errNoPaymentMethodAttached
	})
}
