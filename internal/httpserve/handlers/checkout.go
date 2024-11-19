package handlers

import (
	"github.com/rs/zerolog/log"
	"github.com/stripe/stripe-go/v81"
	echo "github.com/theopenlane/echox"

	"github.com/theopenlane/iam/auth"

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

	log.Warn().Msg("obtained organization ID from context")

	settings, err := h.getOrgSettingByOrgID(reqCtx, orgID)
	if err != nil {
		log.Error().Err(err).Msg("unable to get organization settings by org id")

		return h.BadRequest(ctx, err)
	}

	log.Warn().Msg("obtained organization settings by org ID")

	cust, err := h.fetchOrCreateStripe(reqCtx, settings)
	if err != nil {
		log.Error().Err(err).Msg("unable to fetch or create stripe customer")

		return h.BadRequest(ctx, err)
	}

	log.Warn().Msg("fetched or created stripe customer")

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

	log.Warn().Msg("created stripe checkout session")
	log.Warn().Msgf("sending back client secret %s", result.ClientSecret)

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
