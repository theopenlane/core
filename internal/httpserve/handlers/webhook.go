package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"slices"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/rs/zerolog/log"
	"github.com/stripe/stripe-go/v82"
	"github.com/stripe/stripe-go/v82/webhook"
	echo "github.com/theopenlane/echox"
	"github.com/theopenlane/iam/auth"
	"github.com/theopenlane/utils/contextx"

	ent "github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/generated/apitoken"
	"github.com/theopenlane/core/internal/ent/generated/organization"
	"github.com/theopenlane/core/internal/ent/generated/orgsubscription"
	"github.com/theopenlane/core/internal/ent/generated/personalaccesstoken"
	"github.com/theopenlane/core/internal/ent/generated/privacy"
	"github.com/theopenlane/core/pkg/entitlements"
	"github.com/theopenlane/core/pkg/middleware/transaction"
	"github.com/theopenlane/core/pkg/models"
)

const (
	// stripeSignatureHeaderKey is the header key for the stripe signature
	stripeSignatureHeaderKey = "Stripe-Signature"
	// maxBodyBytes is the maximum size of the request body for stripe webhooks
	maxBodyBytes = int64(65536)
)

// supportedEventTypes is a map of supported stripe event types that the webhook receiver can handle
var supportedEventTypes = map[stripe.EventType]bool{
	stripe.EventTypeCustomerSubscriptionUpdated:      true,
	stripe.EventTypeCustomerSubscriptionDeleted:      true,
	stripe.EventTypeCustomerSubscriptionPaused:       true,
	stripe.EventTypeCustomerSubscriptionTrialWillEnd: true,
	stripe.EventTypePaymentMethodAttached:            true,
}

var (
	webhookReceivedCounter = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "webhook_received_total",
			Help: "Total number of webhooks received",
		},
		[]string{"event_type"},
	)

	webhookProcessingLatency = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "webhook_processing_latency_seconds",
			Help:    "Latency of webhook processing in seconds",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"event_type"},
	)

	webhookResponseCounter = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "webhook_response_total",
			Help: "Total number of webhook responses grouped by event type and status code",
		},
		[]string{"event_type", "status_code"},
	)
)

func init() {
	prometheus.MustRegister(webhookReceivedCounter)
	prometheus.MustRegister(webhookProcessingLatency)
	prometheus.MustRegister(webhookResponseCounter)
}

// WebhookReceiverHandler handles incoming stripe webhook events for the supported event types
func (h *Handler) WebhookReceiverHandler(ctx echo.Context) error {
	startTime := time.Now()

	req := ctx.Request()
	res := ctx.Response()

	payload, err := io.ReadAll(http.MaxBytesReader(res.Writer, req.Body, maxBodyBytes))
	if err != nil {
		webhookResponseCounter.WithLabelValues("payload_exceeded", "500").Inc()
		log.Error().Err(err).Msg("failed to read request body")

		return h.InternalServerError(ctx, err)
	}

	event, err := webhook.ConstructEvent(payload, req.Header.Get(stripeSignatureHeaderKey), h.Entitlements.Config.StripeWebhookSecret)
	if err != nil {
		webhookResponseCounter.WithLabelValues("event_signature_failure", "400").Inc()
		log.Error().Err(err).Msg("failed to construct event")

		return h.BadRequest(ctx, err)
	}

	webhookReceivedCounter.WithLabelValues(string(event.Type)).Inc()

	if !supportedEventTypes[event.Type] {
		webhookResponseCounter.WithLabelValues(string(event.Type)+"_discarded", "200").Inc()
		log.Debug().Str("event_type", string(event.Type)).Msg("unsupported event type")

		return h.Success(ctx, "unsupported event type")
	}

	newCtx := privacy.DecisionContext(req.Context(), privacy.Allow)
	newCtx = contextx.With(newCtx, auth.OrgSubscriptionContextKey{})

	exists, err := h.checkForEventID(newCtx, event.ID)
	if err != nil {
		webhookResponseCounter.WithLabelValues(string(event.Type), "500").Inc()
		log.Error().Err(err).Msg("failed to check for event ID")

		return h.InternalServerError(ctx, err)
	}

	if !exists {
		meowevent, err := h.createEvent(newCtx, ent.CreateEventInput{EventID: &event.ID})
		if err != nil {
			webhookResponseCounter.WithLabelValues(string(event.Type), "500").Inc()
			log.Error().Err(err).Msg("failed to create event")

			return h.InternalServerError(ctx, err)
		}

		log.Debug().Msgf("internal event: %v", meowevent)

		if err = h.HandleEvent(newCtx, &event); err != nil {
			webhookResponseCounter.WithLabelValues(string(event.Type), "500").Inc()
			log.Error().Err(err).Msg("failed to handle event")

			return h.InternalServerError(ctx, err)
		}
	}

	duration := time.Since(startTime).Seconds()
	webhookProcessingLatency.WithLabelValues(string(event.Type)).Observe(duration)
	webhookResponseCounter.WithLabelValues(string(event.Type)+"_processed", "200").Inc()

	return h.Success(ctx, nil)
}

// unmarshalEventData is used to unmarshal event data from a stripe.Event object into a specific type T
func unmarshalEventData[T interface{}](e *stripe.Event) (*T, error) {
	var data T

	err := json.Unmarshal(e.Data.Raw, &data)
	if err != nil {
		log.Error().Err(err).Msg("Failed to unmarshal event data")
	}

	return &data, err
}

// HandleEvent unmarshals event data and triggers a corresponding function to be executed based on case match
func (h *Handler) HandleEvent(c context.Context, e *stripe.Event) error {
	switch e.Type {
	case stripe.EventTypeCustomerSubscriptionUpdated:
		subscription, err := unmarshalEventData[stripe.Subscription](e)
		if err != nil {
			return err
		}

		return h.handleSubscriptionUpdated(c, subscription)
	case stripe.EventTypePaymentMethodAttached:
		paymentMethod, err := unmarshalEventData[stripe.PaymentMethod](e)
		if err != nil {
			return err
		}

		return h.handlePaymentMethodAdded(c, paymentMethod)
	case stripe.EventTypeCustomerSubscriptionTrialWillEnd:
		subscription, err := unmarshalEventData[stripe.Subscription](e)
		if err != nil {
			return err
		}

		return h.handleTrialWillEnd(c, subscription)
	case stripe.EventTypeCustomerSubscriptionDeleted, stripe.EventTypeCustomerSubscriptionPaused:
		subscription, err := unmarshalEventData[stripe.Subscription](e)
		if err != nil {
			return err
		}

		return h.handleSubscriptionPaused(c, subscription)
	default:
		log.Warn().Str("event_type", string(e.Type)).Msg("unsupported event type")

		return ErrUnsupportedEventType
	}
}

// invalidateAPITokens invalidates all API tokens for an organization
func (h *Handler) invalidateAPITokens(ctx context.Context, orgID string) error {
	allowCtx := privacy.DecisionContext(ctx, privacy.Allow)              // bypass privacy policy
	allowCtx = contextx.With(allowCtx, auth.OrgSubscriptionContextKey{}) // bypass org owned interceptor

	num, err := h.DBClient.APIToken.Update().Where(apitoken.OwnerID(orgID)).
		SetIsActive(false).
		SetExpiresAt(time.Now().Add(-24 * time.Hour)). // nolint:mnd // set expiration to 24 hours ago
		SetRevokedAt(time.Now()).
		SetRevokedReason("subscription paused or deleted").
		SetRevokedBy("entitlements_engine").
		Save(allowCtx)
	if err != nil {
		return err
	}

	log.Debug().Str("organization_id", orgID).Int("num_tokens_revoked", num).Msg("revoked API tokens")

	return nil
}

// invalidatePersonalAccessTokens invalidates all personal access tokens tokens for an organization
func (h *Handler) invalidatePersonalAccessTokens(ctx context.Context, orgID string) error {
	allowCtx := privacy.DecisionContext(ctx, privacy.Allow)              // bypass privacy policy
	allowCtx = contextx.With(allowCtx, auth.OrgSubscriptionContextKey{}) // bypass org owned interceptor

	num, err := h.DBClient.PersonalAccessToken.Update().
		RemoveOrganizationIDs(orgID).
		Where(personalaccesstoken.HasOrganizationsWith(organization.ID(orgID))).
		Save(allowCtx)
	if err != nil {
		return err
	}

	log.Debug().Str("organization_id", orgID).Int("num_tokens_revoked", num).Msg("revoked personal access tokens tokens")

	return nil
}

// handleSubscriptionPaused handles subscription updated events for paused subscriptions
func (h *Handler) handleSubscriptionPaused(ctx context.Context, s *stripe.Subscription) (err error) {
	if s.Customer == nil {
		log.Error().Msg("subscription has no customer, cannot proceed")

		return ErrSubscriberNotFound
	}

	customer, err := h.Entitlements.GetCustomerByStripeID(ctx, s.Customer.ID)
	if err != nil {
		return err
	}

	ownerID, err := syncOrgSubscriptionWithStripe(ctx, s, customer)
	if err != nil {
		return
	}

	err = h.invalidateAPITokens(ctx, *ownerID)
	if err != nil {
		return
	}

	return h.invalidatePersonalAccessTokens(ctx, *ownerID)
}

// handleSubscriptionUpdated handles subscription updated events
func (h *Handler) handleSubscriptionUpdated(ctx context.Context, s *stripe.Subscription) error {
	if s.Customer == nil {
		log.Error().Msg("subscription has no customer, cannot proceed")

		return ErrSubscriberNotFound
	}

	customer, err := h.Entitlements.GetCustomerByStripeID(ctx, s.Customer.ID)
	if err != nil {
		return err
	}

	_, err = syncOrgSubscriptionWithStripe(ctx, s, customer)
	if err != nil {
		return err
	}

	return nil
}

// handleTrialWillEnd handles trial will end events, currently just calls handleSubscriptionUpdated
func (h *Handler) handleTrialWillEnd(ctx context.Context, s *stripe.Subscription) error {
	return h.handleSubscriptionUpdated(ctx, s)
}

// handlePaymentMethodAdded handles payment intent events for added payment methods
func (h *Handler) handlePaymentMethodAdded(ctx context.Context, paymentMethod *stripe.PaymentMethod) error {
	if paymentMethod.Customer == nil {
		log.Error().Msg("payment method has no customer, cannot proceed")

		return nil
	}

	return transaction.FromContext(ctx).OrgSubscription.Update().
		Where(orgsubscription.StripeCustomerID(paymentMethod.Customer.ID)).
		SetPaymentMethodAdded(true).
		Exec(ctx)
}

// getOrgSubscription retrieves the OrgSubscription from the database based on the Stripe subscription ID
func getOrgSubscription(ctx context.Context, subscription *stripe.Subscription) (*ent.OrgSubscription, error) {
	allowCtx := contextx.With(ctx, auth.OrgSubscriptionContextKey{})

	orgSubscription, err := transaction.FromContext(ctx).OrgSubscription.Query().
		Where(orgsubscription.StripeSubscriptionID(subscription.ID)).Only(allowCtx)
	if err != nil {
		log.Error().Err(err).Msg("failed to find org subscription")
		return nil, err
	}

	return orgSubscription, nil
}

// syncOrgSubscriptionWithStripe updates the internal OrgSubscription record with data from Stripe and
// returns the owner (organization) ID of the OrgSubscription to be used for further operations if needed
func syncOrgSubscriptionWithStripe(ctx context.Context, subscription *stripe.Subscription, customer *stripe.Customer) (*string, error) {
	orgSubscription, err := getOrgSubscription(ctx, subscription)
	if err != nil {
		return nil, err
	}

	// map stripe data to internal OrgSubscription
	stripeOrgSubscription := mapStripeToOrgSubscription(subscription, entitlements.MapStripeCustomer(customer))

	// Check if any fields have changed before saving the updated OrgSubscription
	changed := false
	mutation := transaction.FromContext(ctx).OrgSubscription.UpdateOne(orgSubscription)

	if orgSubscription.StripeSubscriptionStatus != stripeOrgSubscription.StripeSubscriptionStatus {
		mutation.SetStripeSubscriptionStatus(stripeOrgSubscription.StripeSubscriptionStatus)

		changed = true

		log.Debug().Str("subscription_id", orgSubscription.ID).Str("status", orgSubscription.StripeSubscriptionStatus).Msg("stripe subscription status changed")
	}

	if !slices.Equal(orgSubscription.Features, stripeOrgSubscription.Features) {
		mutation.SetFeatures(stripeOrgSubscription.Features)

		changed = true

		log.Debug().Str("subscription_id", orgSubscription.ID).Strs("features", orgSubscription.Features).Msg("features changes")
	}

	if orgSubscription.ProductTier != stripeOrgSubscription.ProductTier {
		mutation.SetProductTier(stripeOrgSubscription.ProductTier)

		changed = true

		log.Debug().Str("subscription_id", orgSubscription.ID).Str("tier", orgSubscription.ProductTier).Msg("tier changed")
	}

	if orgSubscription.ProductPrice != stripeOrgSubscription.ProductPrice {
		productPriceCopy := stripeOrgSubscription.ProductPrice

		productPriceCopy.Amount /= 100 // convert to dollars from cents

		mutation.SetProductPrice(productPriceCopy)

		mutation.SetProductPrice(stripeOrgSubscription.ProductPrice)

		changed = true

		log.Debug().Str("subscription_id", orgSubscription.ID).Str("price", orgSubscription.ProductPrice.String()).Msgf("product price changed")
	}

	if stripeOrgSubscription.TrialExpiresAt != nil && orgSubscription.TrialExpiresAt != stripeOrgSubscription.TrialExpiresAt {
		mutation.SetTrialExpiresAt(*stripeOrgSubscription.TrialExpiresAt)

		changed = true

		log.Debug().Str("subscription_id", orgSubscription.ID).Str("trial_expires_at", stripeOrgSubscription.TrialExpiresAt.String()).Msg("subscription trial expiration changed")
	}

	if !slices.Equal(orgSubscription.FeatureLookupKeys, stripeOrgSubscription.FeatureLookupKeys) {
		mutation.SetFeatureLookupKeys(stripeOrgSubscription.FeatureLookupKeys)

		changed = true

		log.Debug().Str("subscription_id", orgSubscription.ID).Strs("feature_lookup_keys", orgSubscription.FeatureLookupKeys).Msg("feature lookup keys changed")
	}

	if orgSubscription.TrialExpiresAt != nil && orgSubscription.TrialExpiresAt != stripeOrgSubscription.TrialExpiresAt {
		mutation.SetTrialExpiresAt(*stripeOrgSubscription.TrialExpiresAt)

		changed = true

		log.Debug().Str("subscription_id", orgSubscription.ID).Str("expires_at", orgSubscription.TrialExpiresAt.String()).Msg("trial expiration changed")
	}

	if stripeOrgSubscription.DaysUntilDue != nil && orgSubscription.DaysUntilDue != stripeOrgSubscription.DaysUntilDue {
		mutation.SetDaysUntilDue(*stripeOrgSubscription.DaysUntilDue)

		changed = true

		log.Debug().Str("subscription_id", orgSubscription.ID).Str("days_until_due", *stripeOrgSubscription.DaysUntilDue).Msg("days until due changed")
	}

	if stripeOrgSubscription.PaymentMethodAdded != nil && orgSubscription.PaymentMethodAdded != stripeOrgSubscription.PaymentMethodAdded {
		mutation.SetPaymentMethodAdded(*stripeOrgSubscription.PaymentMethodAdded)

		changed = true

		if orgSubscription.PaymentMethodAdded != nil {
			log.Debug().Str("subscription_id", orgSubscription.ID).Bool("payment_method_added", *orgSubscription.PaymentMethodAdded).Msg("payment method added changed")
		} else {
			log.Debug().Str("subscription_id", orgSubscription.ID).Msg("payment method added changed but was previously nil")
		}
	}

	if orgSubscription.Active != stripeOrgSubscription.Active {
		mutation.SetActive(stripeOrgSubscription.Active)

		changed = true

		log.Debug().Str("subscription_id", orgSubscription.ID).Bool("active", orgSubscription.Active).Msg("active status changed")
	}

	if changed {
		// Collect all changes and execute the mutation once
		err = mutation.Exec(ctx)
		if err != nil {
			log.Error().Err(err).Msg("failed to update OrgSubscription")

			return nil, err
		}

		log.Debug().Str("subscription_id", orgSubscription.ID).Msg("OrgSubscription updated successfully")
	}

	return &orgSubscription.OwnerID, nil
}

// mapStripeToOrgSubscription maps a stripe.Subscription and OrganizationCustomer to an ent.OrgSubscription
func mapStripeToOrgSubscription(subscription *stripe.Subscription, customer *entitlements.OrganizationCustomer) *ent.OrgSubscription {
	if subscription == nil {
		return nil
	}

	productName := ""
	productPrice := models.Price{}

	if subscription.Items != nil && len(subscription.Items.Data) == 1 {
		item := subscription.Items.Data[0]
		if item.Price != nil {
			if item.Price.Product != nil {
				productName = item.Price.Product.Name
			}

			productPrice.Amount = subscription.Items.Data[0].Price.UnitAmountDecimal
			productPrice.Currency = string(subscription.Items.Data[0].Price.Currency)
			productPrice.Interval = string(subscription.Items.Data[0].Price.Recurring.Interval)
		}
	}

	return &ent.OrgSubscription{
		StripeSubscriptionID:     subscription.ID,
		TrialExpiresAt:           timePtr(time.Unix(subscription.TrialEnd, 0)),
		StripeSubscriptionStatus: string(subscription.Status),
		Active:                   entitlements.IsSubscriptionActive(subscription.Status),
		ProductTier:              productName,
		ProductPrice:             productPrice,
		Features:                 customer.Features,
		FeatureLookupKeys:        customer.FeatureNames,
		DaysUntilDue:             int64ToStringPtr(subscription.DaysUntilDue),
	}
}

// int64ToStringPtr converts an int64 to a *string
func int64ToStringPtr(i int64) *string {
	s := fmt.Sprintf("%d", i)
	return &s
}

// timePtr returns a pointer to the given time.Time value
func timePtr(t time.Time) *time.Time {
	return &t
}
