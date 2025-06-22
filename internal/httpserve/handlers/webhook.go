package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/rs/zerolog/log"
	"github.com/stripe/stripe-go/v82"
	"github.com/stripe/stripe-go/v82/webhook"
	echo "github.com/theopenlane/echox"
	"github.com/theopenlane/iam/auth"
	"github.com/theopenlane/iam/fgax"
	"github.com/theopenlane/utils/contextx"

	ent "github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/generated/apitoken"
	"github.com/theopenlane/core/internal/ent/generated/organization"
	"github.com/theopenlane/core/internal/ent/generated/orgmodule"
	"github.com/theopenlane/core/internal/ent/generated/orgsubscription"
	"github.com/theopenlane/core/internal/ent/generated/personalaccesstoken"
	"github.com/theopenlane/core/internal/ent/generated/privacy"
	"github.com/theopenlane/core/pkg/catalog"
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

	ownerID, err := h.syncOrgSubscriptionWithStripe(ctx, s, customer)
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

	_, err = h.syncOrgSubscriptionWithStripe(ctx, s, customer)
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
func (h *Handler) syncOrgSubscriptionWithStripe(ctx context.Context, subscription *stripe.Subscription, customer *stripe.Customer) (*string, error) {
	orgSubscription, err := getOrgSubscription(ctx, subscription)
	if err != nil {
		return nil, err
	}

	stripeOrgSubscription := mapStripeToOrgSubscription(subscription, entitlements.MapStripeCustomer(customer))

	changed := false
	mutation := transaction.FromContext(ctx).OrgSubscription.UpdateOne(orgSubscription)

	if orgSubscription.StripeSubscriptionStatus != stripeOrgSubscription.StripeSubscriptionStatus {
		mutation.SetStripeSubscriptionStatus(stripeOrgSubscription.StripeSubscriptionStatus)
		changed = true
	}

	if stripeOrgSubscription.TrialExpiresAt != nil && orgSubscription.TrialExpiresAt != stripeOrgSubscription.TrialExpiresAt {
		mutation.SetTrialExpiresAt(*stripeOrgSubscription.TrialExpiresAt)
		changed = true
	}

	if stripeOrgSubscription.DaysUntilDue != nil && orgSubscription.DaysUntilDue != stripeOrgSubscription.DaysUntilDue {
		mutation.SetDaysUntilDue(*stripeOrgSubscription.DaysUntilDue)
		changed = true
	}

	if stripeOrgSubscription.PaymentMethodAdded != nil && orgSubscription.PaymentMethodAdded != stripeOrgSubscription.PaymentMethodAdded {
		mutation.SetPaymentMethodAdded(*stripeOrgSubscription.PaymentMethodAdded)
		changed = true
	}

	if orgSubscription.Active != stripeOrgSubscription.Active {
		mutation.SetActive(stripeOrgSubscription.Active)
		changed = true
	}

	if changed {
		if err = mutation.Exec(ctx); err != nil {
			return nil, err
		}
	}

	// update org modules based on subscription items
	oldMods, _ := orgSubscription.QueryModules().All(ctx)
	oldKeys := make([]string, 0, len(oldMods))
	for _, m := range oldMods {
		oldKeys = append(oldKeys, m.Module)
	}
	if len(oldMods) > 0 {
		if _, err := h.DBClient.OrgModule.Delete().Where(orgmodule.SubscriptionID(orgSubscription.ID)).Exec(ctx); err != nil {
			return nil, err
		}
	}

	newKeys := []string{}
	if subscription.Items != nil {
		for _, item := range subscription.Items.Data {
			modKey, _ := h.moduleForPrice(item.Price.ID)
			if modKey == "" {
				continue
			}
			newKeys = append(newKeys, modKey)
			_, err = h.DBClient.OrgModule.Create().
				SetOwnerID(orgSubscription.OwnerID).
				SetSubscriptionID(orgSubscription.ID).
				SetModule(modKey).
				SetStripePriceID(item.Price.ID).
				SetPrice(models.Price{Amount: item.Price.UnitAmountDecimal, Currency: string(item.Price.Currency), Interval: string(item.Price.Recurring.Interval)}).
				SetActive(true).
				Save(ctx)
			if err != nil {
				return nil, err
			}
		}
	}

	if err := syncFeatureTuples(ctx, &h.DBClient.Authz, orgSubscription.OwnerID, oldKeys, newKeys); err != nil {
		log.Error().Err(err).Msg("sync feature tuples")
	}

	return &orgSubscription.OwnerID, nil
}

// mapStripeToOrgSubscription maps a stripe.Subscription and OrganizationCustomer to an ent.OrgSubscription
func mapStripeToOrgSubscription(subscription *stripe.Subscription, customer *entitlements.OrganizationCustomer) *ent.OrgSubscription {
	if subscription == nil {
		return nil
	}

	return &ent.OrgSubscription{
		StripeSubscriptionID:     subscription.ID,
		TrialExpiresAt:           timePtr(time.Unix(subscription.TrialEnd, 0)),
		StripeSubscriptionStatus: string(subscription.Status),
		Active:                   entitlements.IsSubscriptionActive(subscription.Status),
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

// moduleForPrice returns the module key for a given Stripe price ID
func (h *Handler) moduleForPrice(priceID string) (string, catalog.Billing) {
	if h.Catalog == nil {
		return "", catalog.Billing{}
	}
	search := func(fs catalog.FeatureSet) (string, catalog.Billing) {
		for key, mod := range fs {
			for _, p := range mod.Billing.Prices {
				if p.PriceID == priceID {
					return key, catalog.Billing{Prices: []catalog.Price{p}}
				}
			}
		}
		return "", catalog.Billing{}
	}
	if k, b := search(h.Catalog.Modules); k != "" {
		return k, b
	}
	return search(h.Catalog.Addons)
}

// syncFeatureTuples updates openFGA tuples for organization feature access
func syncFeatureTuples(ctx context.Context, client *fgax.Client, orgID string, old, new []string) error {
	if client == nil {
		return nil
	}

	addMap := map[string]struct{}{}
	for _, f := range new {
		addMap[f] = struct{}{}
	}
	for _, f := range old {
		if _, ok := addMap[f]; ok {
			delete(addMap, f)
		}
	}

	delMap := map[string]struct{}{}
	for _, f := range old {
		delMap[f] = struct{}{}
	}
	for _, f := range new {
		delete(delMap, f)
	}

	adds := []fgax.TupleKey{}
	for f := range addMap {
		adds = append(adds, fgax.GetTupleKey(fgax.TupleRequest{
			SubjectID:   orgID,
			SubjectType: "organization",
			ObjectID:    f,
			ObjectType:  "feature",
			Relation:    "enabled",
		}))
	}

	dels := []fgax.TupleKey{}
	for f := range delMap {
		dels = append(dels, fgax.GetTupleKey(fgax.TupleRequest{
			SubjectID:   orgID,
			SubjectType: "organization",
			ObjectID:    f,
			ObjectType:  "feature",
			Relation:    "enabled",
		}))
	}

	if len(adds) == 0 && len(dels) == 0 {
		return nil
	}

	_, err := client.WriteTupleKeys(ctx, adds, dels)
	return err
}
