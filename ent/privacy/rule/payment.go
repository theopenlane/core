package rule

import (
	"context"
	"errors"
	"slices"
	"strings"

	"entgo.io/ent"

	"github.com/theopenlane/iam/auth"

	"github.com/theopenlane/core/pkg/logx"
	"github.com/theopenlane/ent/generated"
	"github.com/theopenlane/ent/generated/organizationsetting"
	"github.com/theopenlane/ent/generated/privacy"
	"github.com/theopenlane/ent/privacy/utils"
)

var (
	errNoPaymentMethodAttached = errors.New("A valid payment method is required to create tokens. Contact your organization admin to add one in billing.") //nolint:staticcheck,revive
)

// RequirePaymentMethod makes sure the organization has a payment method ( card or any other)
// added to stripe already
func RequirePaymentMethod() privacy.MutationRuleFunc {
	return privacy.MutationRuleFunc(func(ctx context.Context, _ ent.Mutation) error {

		client := generated.FromContext(ctx)

		if !utils.PaymentMethodCheckRequired(client) || auth.IsSystemAdminFromContext(ctx) {
			return privacy.Skip
		}

		au, ok := auth.AuthenticatedUserFromContext(ctx)
		if !ok {
			return auth.ErrNoAuthUser
		}

		orgSetting, err := client.OrganizationSetting.Query().
			Where(organizationsetting.OrganizationID(au.OrganizationID)).
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

		emailDomain := strings.SplitAfter(au.SubjectEmail, "@")[1]

		if slices.Contains(client.EntConfig.Billing.BypassEmailDomains, emailDomain) {
			return privacy.Skip
		}

		return errNoPaymentMethodAttached
	})
}
