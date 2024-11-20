package handlers

import (
	"context"

	"github.com/rs/zerolog/log"
	"github.com/stripe/stripe-go/v81"
	echo "github.com/theopenlane/echox"

	"github.com/theopenlane/iam/auth"

	ent "github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/pkg/models"
)

// CheckoutSessionHandler is responsible for handling requests to /checkout/session and creating a stripe checkout session based on the user's organization context
func (h *Handler) CheckoutSessionHandler(ctx echo.Context) error {
	reqCtx := ctx.Request().Context()

	orgID, err := auth.GetOrganizationIDFromContext(reqCtx)
	if err != nil {
		log.Error().Err(err).Msg("unable to get organization id from context")

		return h.BadRequest(ctx, err)
	}

	settings, err := h.getOrgSettingByOrgID(reqCtx, orgID)
	if err != nil {
		log.Error().Err(err).Str("organization_id", orgID).Msg("unable to get organization settings by org id")

		return h.BadRequest(ctx, err)
	}

	cust, err := h.fetchOrCreateStripe(reqCtx, settings)
	if err != nil {
		log.Error().Err(err).Msg("unable to fetch or create stripe customer")

		return h.BadRequest(ctx, err)
	}

	// TODO: determine if customerSession + pricing table is what we want, or if the billingportalsession is the correct URL to return and then redirect the customer
	params := &stripe.CustomerSessionParams{
		Customer: stripe.String(cust.ID),
		Components: &stripe.CustomerSessionComponentsParams{
			PricingTable: &stripe.CustomerSessionComponentsPricingTableParams{
				Enabled: stripe.Bool(true),
			},
		},
	}

	result, err := h.Entitlements.Client.CustomerSessions.New(params)
	if err != nil {
		log.Error().Err(err).Msg("unable to create stripe checkout session")

		return h.BadRequest(ctx, err)
	}

	// set the out attributes we send back to the client only on success
	out := &models.EntitlementsReply{
		ClientSecret: result.ClientSecret,
	}

	return h.Success(ctx, out)
}

// CheckoutSuccessHandler is responsible for handling requests to the `/checkout/success` endpoint
func (h *Handler) CheckoutSuccessHandler(ctx echo.Context) error {
	// TODO[MKA] Determine what is needed of the success handler and implement
	return h.Success(ctx, nil)
}

func (h *Handler) fetchOrCreateStripe(context context.Context, orgsetting *ent.OrganizationSetting) (*stripe.Customer, error) {
	if orgsetting.BillingEmail == "" {
		log.Error().Msgf("billing email is required to be set to create a checkout session")
		return nil, ErrNoBillingEmail
	}

	if orgsetting.StripeID != "" {
		cust, err := h.Entitlements.Client.Customers.Get(orgsetting.StripeID, nil)
		if err != nil {
			log.Error().Err(err).Msg("error fetching stripe customer")
			return nil, err
		}

		if cust.Email != orgsetting.BillingEmail {
			log.Error().Msgf("customer email does not match, updating stripe customer")

			_, err := h.Entitlements.Client.Customers.Update(orgsetting.StripeID, &stripe.CustomerParams{
				Email: &orgsetting.BillingEmail,
			})
			if err != nil {
				log.Error().Err(err).Msg("error updating stripe customer")
				return nil, err
			}
		}

		return cust, nil
	}

	stripeCustomer, err := h.Entitlements.Client.Customers.New(&stripe.CustomerParams{
		Email: &orgsetting.BillingEmail,
	})
	if err != nil {
		log.Error().Err(err).Msg("error creating stripe customer")
		return nil, err
	}

	if err := h.updateOrganizationSettingWithCustomerID(context, orgsetting.ID, stripeCustomer.ID); err != nil {
		log.Error().Err(err).Msg("error updating organization setting with stripe customer id")
		return nil, err
	}

	return stripeCustomer, nil
}
