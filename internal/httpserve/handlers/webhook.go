package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/rs/zerolog/log"
	"github.com/stripe/stripe-go/v86"
	"github.com/stripe/stripe-go/v86/webhook"
	echo "github.com/theopenlane/echox"
	"github.com/theopenlane/entx"
	"github.com/theopenlane/iam/auth"

	models "github.com/theopenlane/core/common/openapi"
	ent "github.com/theopenlane/core/v2/internal/ent/generated"
	"github.com/theopenlane/core/v2/internal/ent/generated/apitoken"
	"github.com/theopenlane/core/v2/internal/ent/generated/organization"
	"github.com/theopenlane/core/v2/internal/ent/generated/organizationsetting"
	"github.com/theopenlane/core/v2/internal/ent/generated/orgsubscription"
	"github.com/theopenlane/core/v2/internal/ent/generated/personalaccesstoken"
	"github.com/theopenlane/core/v2/internal/ent/generated/privacy"
	catalogprices "github.com/theopenlane/core/v2/internal/entitlements"
	em "github.com/theopenlane/core/v2/internal/entitlements/entmapping"
	"github.com/theopenlane/core/v2/pkg/entitlements"
	"github.com/theopenlane/core/v2/pkg/logx"
	"github.com/theopenlane/core/v2/pkg/middleware/transaction"
)

const (
	// stripeSignatureHeaderKey is the header key for the stripe signature
	stripeSignatureHeaderKey = "Stripe-Signature"
	// maxBodyBytes is the maximum size of the request body for stripe webhooks
	maxBodyBytes = int64(65536)
)

// supportedEventTypes is a map of supported stripe event types that the webhook receiver can handle
var supportedEventTypes = func() map[stripe.EventType]bool {
	m := make(map[stripe.EventType]bool)
	for _, evt := range entitlements.SupportedEventTypes {
		m[evt] = true
	}

	return m
}()

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

	errPayloadEmpty = errors.New("payload is empty")

	errMissingSecret = errors.New("webhook secret is missing")
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

	reqCtx := req.Context()

	webhookReq := &models.StripeWebhookRequest{APIVersion: ctx.QueryParam("api_version")}

	if webhookReq.APIVersion != "" && h.Entitlements.Config.StripeWebhookDiscardAPIVersion != "" {
		if webhookReq.APIVersion == h.Entitlements.Config.StripeWebhookDiscardAPIVersion {
			webhookResponseCounter.WithLabelValues("api_version_discarded", "200").Inc()
			logx.FromContext(reqCtx).Debug().Str("api_version", webhookReq.APIVersion).Msg("webhook with discard API version received, ignoring")

			return h.Success(ctx, "webhook ignored - API version being discarded")
		}
	}

	if webhookReq.APIVersion != "" && h.Entitlements.Config.StripeWebhookAPIVersion != "" {
		if webhookReq.APIVersion != h.Entitlements.Config.StripeWebhookAPIVersion {
			webhookResponseCounter.WithLabelValues("api_version_mismatch", "200").Inc()
			logx.FromContext(reqCtx).Warn().Str("api_version", webhookReq.APIVersion).Str("expected_version", h.Entitlements.Config.StripeWebhookAPIVersion).Msg("webhook with unexpected API version")

			return h.Success(ctx, "webhook ignored - API version mismatch")
		}
	}

	payload, err := io.ReadAll(http.MaxBytesReader(res.Writer, req.Body, maxBodyBytes))
	if err != nil {
		webhookResponseCounter.WithLabelValues("payload_exceeded", "500").Inc()
		logx.FromContext(reqCtx).Error().Err(err).Msg("failed to read request body")

		return h.InternalServerError(ctx, ErrProcessingRequest)
	}

	logx.FromContext(reqCtx).Info().Msgf("version: %s", webhookReq.APIVersion)

	if payload == nil {
		webhookResponseCounter.WithLabelValues("empty_payload", "400").Inc()
		logx.FromContext(reqCtx).Error().Msg("empty payload received")

		return h.BadRequest(ctx, errPayloadEmpty)
	}

	webhookSecret := h.Entitlements.Config.GetWebhookSecretForVersion(webhookReq.APIVersion)

	if webhookSecret == "" {
		webhookResponseCounter.WithLabelValues("missing_webhook_secret", "500").Inc()
		logx.FromContext(reqCtx).Error().Str("api_version", webhookReq.APIVersion).Msg("missing webhook secret for API version")

		return h.InternalServerError(ctx, errMissingSecret)
	}

	event, err := webhook.ConstructEvent(payload, req.Header.Get(stripeSignatureHeaderKey), webhookSecret)
	if err != nil {
		webhookResponseCounter.WithLabelValues("event_signature_failure", "400").Inc()
		logx.FromContext(reqCtx).Error().Err(err).Str("api_version", webhookReq.APIVersion).Msg("failed to construct event")

		return h.BadRequest(ctx, err)
	}

	webhookReceivedCounter.WithLabelValues(string(event.Type)).Inc()

	if !supportedEventTypes[event.Type] {
		webhookResponseCounter.WithLabelValues(string(event.Type)+"_discarded", "200").Inc()
		logx.FromContext(reqCtx).Debug().Str("event_type", string(event.Type)).Msg("unsupported event type")

		return h.Success(ctx, "unsupported event type")
	}

	newCtx := privacy.DecisionContext(req.Context(), privacy.Allow)
	newCtx = auth.WithCaller(newCtx, auth.NewWebhookCaller(""))

	exists, err := h.checkForEventID(newCtx, event.ID)
	if err != nil {
		webhookResponseCounter.WithLabelValues(string(event.Type), "500").Inc()
		logx.FromContext(reqCtx).Error().Err(err).Msg("failed to check for event ID")

		return h.InternalServerError(ctx, ErrProcessingRequest)
	}

	if !exists {
		meowevent, err := h.createEvent(newCtx, ent.CreateEventInput{EventID: &event.ID})
		if err != nil {
			webhookResponseCounter.WithLabelValues(string(event.Type), "500").Inc()
			logx.FromContext(reqCtx).Error().Err(err).Msg("failed to create event")

			return h.InternalServerError(ctx, ErrProcessingRequest)
		}

		logx.FromContext(reqCtx).Debug().Msgf("internal event: %v", meowevent)

		if err = h.HandleEvent(newCtx, &event); err != nil {
			webhookResponseCounter.WithLabelValues(string(event.Type), "500").Inc()
			logx.FromContext(reqCtx).Error().Str("event", string(event.Type)).Err(err).Msg("failed to handle event")

			return h.InternalServerError(ctx, ErrProcessingRequest)
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

		return nil, err
	}

	return &data, nil
}

// HandleEvent unmarshals event data and triggers a corresponding function to be executed based on case match
func (h *Handler) HandleEvent(c context.Context, e *stripe.Event) error {
	switch e.Type {
	case stripe.EventTypeCustomerSubscriptionCreated, stripe.EventTypeCustomerSubscriptionUpdated:
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
	case stripe.EventTypeInvoicePaymentFailed:
		invoice, err := unmarshalEventData[stripe.Invoice](e)
		if err != nil {
			return err
		}

		return h.handleInvoicePaymentFailed(c, invoice)
	case stripe.EventTypeCustomerSubscriptionDeleted, stripe.EventTypeCustomerSubscriptionPaused:
		subscription, err := unmarshalEventData[stripe.Subscription](e)
		if err != nil {
			return err
		}

		err = h.handleSubscriptionPaused(c, subscription)
		if ent.IsNotFound(err) {
			// if org subscription not found, just log and move on
			logx.FromContext(c).Info().Str("subscription_id", subscription.ID).Msg("org subscription not found for paused/deleted subscription, skipping further processing")

			return nil
		}

		return err
	default:
		logx.FromContext(c).Warn().Str("event_type", string(e.Type)).Msg("unsupported event type")

		return ErrUnsupportedEventType
	}
}

// invalidateAPITokens invalidates all API tokens for an organization
func (h *Handler) invalidateAPITokens(ctx context.Context, orgID string) error {
	allowCtx := privacy.DecisionContext(ctx, privacy.Allow)
	allowCtx = auth.WithCaller(allowCtx, auth.NewWebhookCaller(orgID))

	num, err := h.DBClient.APIToken.Update().Where(apitoken.OwnerID(orgID)).
		SetIsActive(false).
		// set expiration to 24 hours
		SetExpiresAt(time.Now().Add(-24 * time.Hour)). //nolint:mnd
		SetRevokedAt(time.Now()).
		SetRevokedReason("subscription paused or deleted").
		SetRevokedBy("entitlements_engine").
		Save(allowCtx)
	if err != nil {
		return err
	}

	logx.FromContext(ctx).Debug().Str("organization_id", orgID).Int("num_tokens_revoked", num).Msg("revoked API tokens")

	return nil
}

// invalidatePersonalAccessTokens invalidates all personal access tokens tokens for an organization
func (h *Handler) invalidatePersonalAccessTokens(ctx context.Context, orgID string) error {
	allowCtx := privacy.DecisionContext(ctx, privacy.Allow)
	allowCtx = auth.WithCaller(allowCtx, auth.NewWebhookCaller(orgID))

	num, err := h.DBClient.PersonalAccessToken.Update().
		RemoveOrganizationIDs(orgID).
		Where(personalaccesstoken.HasOrganizationsWith(organization.ID(orgID))).
		Save(allowCtx)
	if err != nil {
		return err
	}

	logx.FromContext(ctx).Debug().Str("organization_id", orgID).Int("num_tokens_revoked", num).Msg("revoked personal access tokens tokens")

	return nil
}

// handleSubscriptionPaused handles subscription updated events for paused subscriptions
func (h *Handler) handleSubscriptionPaused(ctx context.Context, s *stripe.Subscription) (err error) {
	ownerID, err := h.syncOrgSubscriptionWithStripe(ctx, s)
	if err != nil {
		return
	}

	if err = h.removeAllModules(ctx, s); err != nil {
		return
	}

	if ownerID == nil {
		return
	}

	err = h.invalidateAPITokens(ctx, *ownerID)
	if err != nil {
		return
	}

	return h.invalidatePersonalAccessTokens(ctx, *ownerID)
}

// maxRetryAttempts is what the stripe smart retry setup is set to
// this is set in stripe to max of 8 retries over 2 weeks for credit cards
const maxRetryAttempts = 8

// handleInvoicePaymentFailed drops an organization to the free modules once Stripe has stopped
// retrying a failed payment. Stripe schedules a retry after every failed attempt, so an invoice with
// no next attempt left means the retry window is over. Cancelling at that point takes the whole
// organization down and is painful to unwind, keeping the subscription alive on the free modules
// leaves it recoverable by adding a payment method
func (h *Handler) handleInvoicePaymentFailed(ctx context.Context, invoice *stripe.Invoice) error {
	if invoice == nil {
		return nil
	}

	// stripe is configured to a max number of retries, as long as we are under that keep the modules as-is
	if invoice.AttemptCount < maxRetryAttempts {
		logx.FromContext(ctx).Debug().Str("invoice_id", invoice.ID).
			Int64("attempt_count", invoice.AttemptCount).
			Time("next_payment_attempt", time.Unix(invoice.NextPaymentAttempt, 0)).
			Msg("payment failed but stripe has retries left, leaving the modules alone")

		return nil
	}

	if invoice.Parent == nil || invoice.Parent.SubscriptionDetails == nil || invoice.Parent.SubscriptionDetails.Subscription == nil {
		// not a subscription invoice, nothing to downgrade
		return nil
	}

	subscriptionID := invoice.Parent.SubscriptionDetails.Subscription.ID
	if subscriptionID == "" {
		return nil
	}

	// the invoice carries a thin subscription, the items are needed to work out what to remove
	sub, err := h.Entitlements.GetSubscriptionByID(ctx, subscriptionID)
	if err != nil {
		logx.FromContext(ctx).Error().Err(err).Str("subscription_id", subscriptionID).
			Msg("failed to load subscription for payment failure downgrade")

		return err
	}

	freePriceIDs := catalogprices.FreeMonthlyPriceIDs(h.DBClient.EntConfig.Modules.UseSandbox)

	downgraded, err := h.Entitlements.DowngradeToFreeModules(ctx, sub, freePriceIDs)
	if err != nil {
		logx.FromContext(ctx).Error().Err(err).Str("subscription_id", subscriptionID).
			Msg("failed to downgrade subscription to the free modules")

		return err
	}

	logx.FromContext(ctx).Info().Str("subscription_id", subscriptionID).
		Msg("payment retries exhausted, downgraded subscription to the free modules")

	// sync so the modules and their feature tuples match what the subscription now carries
	return h.handleSubscriptionUpdated(ctx, downgraded)
}

// handleSubscriptionUpdated handles subscription updated events
func (h *Handler) handleSubscriptionUpdated(ctx context.Context, s *stripe.Subscription) error {
	_, err := h.syncOrgSubscriptionWithStripe(ctx, s)
	if err != nil {
		logx.FromContext(ctx).Error().Err(err).Msg("failed to sync org subscription with stripe")

		return err
	}

	return h.syncSubscriptionItemsWithStripe(ctx, s, s.Items.Data, s.Status)
}

// handleTrialWillEnd handles trial will end events, currently just calls handleSubscriptionUpdated
func (h *Handler) handleTrialWillEnd(ctx context.Context, s *stripe.Subscription) error {
	return h.handleSubscriptionUpdated(ctx, s)
}

// handlePaymentMethodAdded handles payment intent events for added payment methods
func (h *Handler) handlePaymentMethodAdded(ctx context.Context, paymentMethod *stripe.PaymentMethod) error {
	if paymentMethod.Customer == nil {
		logx.FromContext(ctx).Error().Msg("payment method has no customer, cannot proceed")

		return nil
	}

	allowCtx := privacy.DecisionContext(ctx, privacy.Allow)

	org, err := transaction.FromContext(ctx).Organization.Query().
		Where(organization.StripeCustomerID(paymentMethod.Customer.ID)).
		Only(allowCtx)
	if err != nil {
		logx.FromContext(ctx).Error().Err(err).Msg("could not fetch organization by stripe customer id")
		return err
	}

	return transaction.FromContext(ctx).OrganizationSetting.Update().
		Where(organizationsetting.OrganizationID(org.ID)).
		SetPaymentMethodAdded(true).
		ClearPendingDeletionAt().
		Exec(ctx)
}

// findOrgSubscriptionByCustomer resolves an OrgSubscription through the stripe customer on the
// subscription. The customer outlives individual subscriptions, so this is what still connects a
// replacement subscription to its organization once the original was cancelled. Returns nil when the
// customer is missing or maps to no organization
func findOrgSubscriptionByCustomer(ctx context.Context, subscription *stripe.Subscription) *ent.OrgSubscription {
	if subscription == nil || subscription.Customer == nil || subscription.Customer.ID == "" {
		return nil
	}

	allowCtx := auth.WithCaller(ctx, auth.NewWebhookCaller(""))

	org, err := transaction.FromContext(ctx).Organization.Query().
		Where(organization.StripeCustomerID(subscription.Customer.ID)).Only(allowCtx)
	if err != nil {
		logx.FromContext(ctx).Debug().Err(err).Str("stripe_customer_id", subscription.Customer.ID).
			Msg("no organization found for stripe customer")

		return nil
	}

	// an organization can accumulate more than one subscription record over its life, the most
	// recent is the one a new stripe subscription belongs to
	orgSub, err := transaction.FromContext(ctx).OrgSubscription.Query().
		Where(orgsubscription.OwnerID(org.ID), orgsubscription.DeletedAtIsNil()).
		Order(ent.Desc(orgsubscription.FieldCreatedAt)).
		First(allowCtx)
	if err != nil {
		logx.FromContext(ctx).Debug().Err(err).Str("organization_id", org.ID).
			Msg("no org subscription found for organization")

		return nil
	}

	return orgSub
}

// adoptStripeSubscriptionID points an OrgSubscription at the given stripe subscription. Without this
// an organization whose subscription was replaced keeps referring to the old, cancelled one
func adoptStripeSubscriptionID(ctx context.Context, orgSub *ent.OrgSubscription, subscriptionID string) {
	if orgSub == nil || subscriptionID == "" || orgSub.StripeSubscriptionID == subscriptionID {
		return
	}

	allowCtx := auth.WithCaller(ctx, auth.NewWebhookCaller(""))

	if err := transaction.FromContext(ctx).OrgSubscription.UpdateOne(orgSub).
		SetStripeSubscriptionID(subscriptionID).
		Exec(allowCtx); err != nil {
		logx.FromContext(ctx).Error().Err(err).Str("subscription_id", subscriptionID).
			Msg("failed to update org subscription with stripe subscription id")

		return
	}

	logx.FromContext(ctx).Info().Str("subscription_id", subscriptionID).Str("org_subscription_id", orgSub.ID).
		Msg("org subscription now points at the stripe subscription")

	orgSub.StripeSubscriptionID = subscriptionID
}

// getOrgSubscription retrieves the OrgSubscription from the database based on the Stripe subscription ID
func getOrgSubscription(ctx context.Context, subscription *stripe.Subscription) (*ent.OrgSubscription, error) {
	allowCtx := auth.WithCaller(ctx, auth.NewWebhookCaller(""))

	orgSubscription, err := transaction.FromContext(ctx).OrgSubscription.Query().
		Where(orgsubscription.StripeSubscriptionID(subscription.ID)).Only(allowCtx)
	if err != nil {
		if ent.IsNotFound(err) {
			// try by the metadata field as a fallback if customer is provided
			if subscription != nil && subscription.Metadata != nil {
				orgSubscription := &ent.OrgSubscription{}

				// first try org_subscription_id
				if orgSubID := entitlements.GetOrganizationSubscriptionIDFromMetadata(subscription.Metadata); orgSubID != "" {
					orgSubscription, _ = transaction.FromContext(ctx).OrgSubscription.Query().
						Where(orgsubscription.ID(orgSubID), orgsubscription.DeletedAtIsNil()).Only(allowCtx)
					if orgSubscription == nil {
						// fallback to organization_id
						if orgID := entitlements.GetOrganizationIDFromMetadata(subscription.Metadata); orgID != "" {
							orgSubscription, err = transaction.FromContext(ctx).OrgSubscription.Query().
								Where(orgsubscription.OwnerID(orgID), orgsubscription.DeletedAtIsNil()).Only(allowCtx)
						}
					}
				}

				if orgSubscription != nil && orgSubscription.ID != "" {
					adoptStripeSubscriptionID(ctx, orgSubscription, subscription.ID)

					// but still return the org subscription to update the rest of the fields
					return orgSubscription, nil
				}
			}

			// nothing in the metadata matched, which is the normal case for a subscription created
			// by hand in the stripe dashboard. The customer is the only remaining link back to the
			// organization, and this is what repoints an organization at a replacement subscription
			// after its previous one was cancelled
			if orgSub := findOrgSubscriptionByCustomer(ctx, subscription); orgSub != nil {
				adoptStripeSubscriptionID(ctx, orgSub, subscription.ID)

				return orgSub, nil
			}

			// if we got here we could not find the org subscription
			// first check to see if the org was deleted already
			allowCtx = entx.SkipSoftDelete(allowCtx)
			if orgSubID := entitlements.GetOrganizationSubscriptionIDFromMetadata(subscription.Metadata); orgSubID != "" {
				orgSubscription, _ = transaction.FromContext(ctx).OrgSubscription.Query().
					Where(orgsubscription.ID(orgSubID)).Only(allowCtx)
				if orgSubscription == nil {
					// fallback to organization_id
					if orgID := entitlements.GetOrganizationIDFromMetadata(subscription.Metadata); orgID != "" {
						orgSubscription, err = transaction.FromContext(ctx).OrgSubscription.Query().
							Where(orgsubscription.OwnerID(orgID)).Only(allowCtx)
					}
				}
			}

			if orgSubscription == nil {
				logx.FromContext(ctx).Warn().Str("subscription_id", subscription.ID).Msg("org subscription never existed in system")
			} else {
				logx.FromContext(ctx).Info().Str("subscription_id", subscription.ID).Msg("org subscription found but was already deleted")
			}

		}

		logx.FromContext(ctx).Warn().Err(err).Msg("failed to find org subscription")

		return nil, err
	}

	return orgSubscription, nil
}

// syncOrgSubscriptionWithStripe updates the internal OrgSubscription record with data from Stripe and
// returns the owner (organization) ID of the OrgSubscription to be used for further operations if needed
func (h *Handler) syncOrgSubscriptionWithStripe(ctx context.Context, subscription *stripe.Subscription) (*string, error) {
	orgSubscription, err := getOrgSubscription(ctx, subscription)

	// getOrgSubscription exhausts all possible routes to find the org so this is fine
	if ent.IsNotFound(err) {
		// no org subscription found, nothing to sync
		logx.FromContext(ctx).Info().Str("subscription_id", subscription.ID).Msg("no org subscription found to sync with stripe subscription")

		return nil, nil
	}

	if err != nil {
		return nil, err
	}

	// map stripe data to internal OrgSubscription
	stripeOrgSubscription := em.StripeSubscriptionToOrgSubscription(subscription)

	// Check if any fields have changed before saving the updated OrgSubscription
	changed := false
	mutation := transaction.FromContext(ctx).OrgSubscription.UpdateOne(orgSubscription)

	if orgSubscription.StripeSubscriptionStatus != stripeOrgSubscription.StripeSubscriptionStatus {
		mutation.SetStripeSubscriptionStatus(stripeOrgSubscription.StripeSubscriptionStatus)

		changed = true

		logx.FromContext(ctx).Debug().Str("subscription_id", orgSubscription.ID).Str("status", orgSubscription.StripeSubscriptionStatus).Msg("stripe subscription status changed")
	}

	if stripeOrgSubscription.TrialExpiresAt != nil && orgSubscription.TrialExpiresAt != stripeOrgSubscription.TrialExpiresAt {
		mutation.SetTrialExpiresAt(*stripeOrgSubscription.TrialExpiresAt)

		changed = true

		logx.FromContext(ctx).Debug().Str("subscription_id", orgSubscription.ID).Str("trial_expires_at", stripeOrgSubscription.TrialExpiresAt.String()).Msg("subscription trial expiration changed")
	}

	if orgSubscription.TrialExpiresAt != nil && orgSubscription.TrialExpiresAt != stripeOrgSubscription.TrialExpiresAt {
		mutation.SetTrialExpiresAt(*stripeOrgSubscription.TrialExpiresAt)

		changed = true

		logx.FromContext(ctx).Debug().Str("subscription_id", orgSubscription.ID).Str("expires_at", orgSubscription.TrialExpiresAt.String()).Msg("trial expiration changed")
	}

	if stripeOrgSubscription.DaysUntilDue != nil && orgSubscription.DaysUntilDue != stripeOrgSubscription.DaysUntilDue {
		mutation.SetDaysUntilDue(*stripeOrgSubscription.DaysUntilDue)

		changed = true

		logx.FromContext(ctx).Debug().Str("subscription_id", orgSubscription.ID).Str("days_until_due", *stripeOrgSubscription.DaysUntilDue).Msg("days until due changed")
	}

	if orgSubscription.Active != stripeOrgSubscription.Active {
		mutation.SetActive(stripeOrgSubscription.Active)

		changed = true

		logx.FromContext(ctx).Debug().Str("subscription_id", orgSubscription.ID).Bool("active", orgSubscription.Active).Msg("active status changed")
	}

	if stripeOrgSubscription.Active {
		err := transaction.FromContext(ctx).OrganizationSetting.Update().
			Where(organizationsetting.OrganizationID(orgSubscription.OwnerID)).
			ClearPendingDeletionAt().
			Exec(ctx)
		if err != nil {
			return nil, err
		}
	}

	if changed {
		// Collect all changes and execute the mutation once
		err = mutation.Exec(ctx)
		if err != nil {
			logx.FromContext(ctx).Error().Err(err).Msg("failed to update OrgSubscription")

			return nil, err
		}

		logx.FromContext(ctx).Debug().Str("subscription_id", orgSubscription.ID).Msg("OrgSubscription updated successfully")

		if err := h.clearFeatureCache(ctx, orgSubscription.OwnerID); err != nil {
			logx.FromContext(ctx).Error().Err(err).Msg("failed to ensure feature tuples")
		}
	}

	return &orgSubscription.OwnerID, nil
}

// clearFeatureCache clears the feature cache for an organization in Redis
func (h *Handler) clearFeatureCache(ctx context.Context, orgID string) error {
	if h.RedisClient != nil {
		key := "features:" + orgID
		pipe := h.RedisClient.TxPipeline()
		pipe.Del(ctx, key)
		_, _ = pipe.Exec(ctx)
	}

	return nil
}
