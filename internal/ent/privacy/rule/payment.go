package rule

import (
	"context"
	"errors"

	"entgo.io/ent"

	"github.com/rs/zerolog/log"
	"github.com/theopenlane/iam/auth"

	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/generated/organization"
	"github.com/theopenlane/core/internal/ent/generated/privacy"
	"github.com/theopenlane/core/internal/ent/privacy/utils"
)

var (
	errNoPaymentMethodAttached = errors.New("you do not have a payment method attached. please add one in billing")
)

// RequirePaymentMethod makes sure the organization has a payment mehod ( card or any other)
// added to stripe already
func RequirePaymentMethod() privacy.MutationRuleFunc {
	return privacy.MutationRuleFunc(func(ctx context.Context, _ ent.Mutation) error {

		client := generated.FromContext(ctx)

		if !utils.PaymentMethodCheckRequired(client) {
			return privacy.Skip
		}

		orgID, err := auth.GetOrganizationIDFromContext(ctx)
		if err != nil {
			return err
		}

		orgSetting, err := client.OrganizationSetting.Query().Where(
				organizationsetting.HasOrganizationWith(organization.ID("")),
			).Select(organizationsetting.FieldPaymentMethodAdded).Only(ctx)
		if err != nil {
			log.Err(err).Msg("failed to fetch organization from db")
			return err
		}

		if orgSetting.PaymentMethodAdded {
			// evaluate next rule
			return privacy.Skip
		}

		return errNoPaymentMethodAttached
	})
}
