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
		caller, ok := auth.CallerFromContext(ctx)
		if !ok {
			return auth.ErrNoAuthUser
		}

		hasPaymentMethod, err := HasPaymentMethod(ctx, generated.FromContext(ctx), caller)
		if err != nil {
			return err
		}

		if hasPaymentMethod {
			// evaluate next rule
			return privacy.Skip
		}

		return errNoPaymentMethodAttached
	})
}

// HasPaymentMethod reports whether the caller's organization may use paid features: it has a
// payment method in stripe, an active trial, or the check does not apply to this caller or server
func HasPaymentMethod(ctx context.Context, client *generated.Client, caller *auth.Caller) (bool, error) {
	if !utils.PaymentMethodCheckRequired(client) || caller.Has(auth.CapSystemAdmin) {
		return true, nil
	}

	orgSetting, err := client.OrganizationSetting.Query().
		Where(organizationsetting.OrganizationID(caller.OrganizationID)).
		Select(organizationsetting.FieldPaymentMethodAdded).
		Only(ctx)
	if err != nil {
		logx.FromContext(ctx).Err(err).Msg("failed to fetch organization from db")

		return false, err
	}

	if orgSetting.PaymentMethodAdded {
		return true, nil
	}

	emailDomain := strings.SplitAfter(caller.SubjectEmail, "@")[1]

	if slices.Contains(client.EntConfig.Billing.BypassEmailDomains, emailDomain) {
		return true, nil
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

			return true, nil
		}
	}

	return false, nil
}
