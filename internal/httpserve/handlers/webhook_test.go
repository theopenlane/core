package handlers_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stripe/stripe-go/v83"
	"github.com/stripe/stripe-go/v83/webhook"
	echo "github.com/theopenlane/echox"

	"github.com/theopenlane/core/internal/ent/generated/orgsubscription"
	"github.com/theopenlane/core/internal/ent/generated/privacy"
	models "github.com/theopenlane/core/pkg/openapi"
)

type webhookTestResources struct {
	apiTokenID string
	patID      string
}

type webhookBinder struct {
	echo.Binder
}

func (b *webhookBinder) Bind(c echo.Context, i interface{}) error {
	inner := b.Binder
	if inner == nil {
		inner = new(echo.DefaultBinder)
	}

	if err := inner.Bind(c, i); err != nil {
		return err
	}

	if req, ok := i.(*models.StripeWebhookRequest); ok && req.APIVersion == "" {
		req.APIVersion = c.QueryParam("api_version")
	}

	return nil
}

func (suite *HandlerTestSuite) executeWebhookRequest(t *testing.T, payload *stripe.Event, queryParams, signatureOverride string) *httptest.ResponseRecorder {
	suite.stripeMockBackend.ExpectedCalls = suite.stripeMockBackend.ExpectedCalls[:0]
	suite.orgSubscriptionMocks()

	payloadBytes, err := json.Marshal(payload)
	require.NoError(t, err)

	signedPayload := webhook.GenerateTestSignedPayload(&webhook.UnsignedPayload{
		Payload: payloadBytes,
		Secret:  webhookSecret,
	})

	reqURL := "/webhook"
	if queryParams != "" {
		reqURL += "?" + queryParams
	}

	req := httptest.NewRequest(http.MethodPost, reqURL, bytes.NewReader(signedPayload.Payload))
	req.Header.Set("Content-Type", "application/json")

	signature := signedPayload.Header
	if signatureOverride != "" {
		signature = signatureOverride
	}

	req.Header.Set("Stripe-Signature", signature)

	recorder := httptest.NewRecorder()
	suite.e.ServeHTTP(recorder, req)

	return recorder
}

func (suite *HandlerTestSuite) TestWebhookReceiverHandler() {
	t := suite.T()

	// Pre-seed an active subscription so handler logic finds expected state
	suite.db.OrgSubscription.Create().
		SetStripeSubscriptionStatus("active").
		SetStripeSubscriptionID("PENDING_UPDATE").
		SetOwnerID(testUser1.OrganizationID).
		SetStripeSubscriptionID(seedStripeSubscriptionID).
		ExecX(testUser1.UserCtx)

	operation := suite.createImpersonationOperation("WebhookReceiverHandler", "Webhook receiver handler")
	suite.registerTestHandler(http.MethodPost, "/webhook", operation, suite.h.WebhookReceiverHandler)

	originalBinder := suite.e.Binder
	suite.e.Binder = &webhookBinder{Binder: originalBinder}
	t.Cleanup(func() {
		suite.e.Binder = originalBinder
	})

	dataUpdate := mockCustomer
	dataUpdate.Subscriptions.Data[0].Items.Data[0].Price.UnitAmount = 900

	sub := dataUpdate.Subscriptions.Data[0]
	jsonDataUpdate, err := json.Marshal(sub)
	require.NoError(t, err)

	paymentUpdate := &stripe.PaymentMethod{ID: "pm_test_payment"}
	jsonPaymentUpdate, err := json.Marshal(paymentUpdate)
	require.NoError(t, err)

	testCases := []struct {
		name                    string
		payload                 *stripe.Event
		queryParams             string
		configAPIVersion        string
		configDiscardAPIVersion string
		expectedStatus          int
		signatureOverride       string
		setup                   func(t *testing.T) *webhookTestResources
		validate                func(t *testing.T, resources *webhookTestResources)
	}{
		{
			name: "valid payload - paused subscription",
			payload: &stripe.Event{
				ID:         "evt_test_webhook_paused",
				Object:     "event",
				Type:       stripe.EventTypeCustomerSubscriptionPaused,
				APIVersion: stripe.APIVersion,
				Data: &stripe.EventData{
					Raw: json.RawMessage(jsonDataUpdate),
				},
			},
			expectedStatus: http.StatusOK,
			setup: func(t *testing.T) *webhookTestResources {
				allowCtx := privacy.DecisionContext(testUser1.UserCtx, privacy.Allow)

				apiToken := suite.db.APIToken.Create().
					SetOwnerID(testUser1.OrganizationID).
					SetName("test_token").
					SaveX(allowCtx)

				pat := suite.db.PersonalAccessToken.Create().
					SetOwnerID(testUser1.ID).
					AddOrganizationIDs(testUser1.OrganizationID).
					SetName("test_token").
					SaveX(allowCtx)

				return &webhookTestResources{
					apiTokenID: apiToken.ID,
					patID:      pat.ID,
				}
			},
			validate: func(t *testing.T, resources *webhookTestResources) {
				sub, err := suite.db.OrgSubscription.Query().
					Where(orgsubscription.StripeSubscriptionID(seedStripeSubscriptionID)).
					Only(testUser1.UserCtx)

				require.NoError(t, err)
				assert.False(t, sub.Active)

				apiToken, err := suite.db.APIToken.Get(testUser1.UserCtx, resources.apiTokenID)
				require.NoError(t, err)
				require.False(t, apiToken.IsActive)
				require.NotEmpty(t, apiToken.RevokedBy)
				require.NotEmpty(t, apiToken.RevokedAt)
				require.NotEmpty(t, apiToken.RevokedReason)
				assert.Less(t, *apiToken.ExpiresAt, time.Now())

				pat, err := suite.db.PersonalAccessToken.Get(testUser1.UserCtx, resources.patID)
				require.NoError(t, err)
				require.Len(t, pat.Edges.Organizations, 0)
			},
		},
		{
			name: "valid payload - subscription updated",
			payload: &stripe.Event{
				ID:         "evt_test_webhook_subscription_updated",
				Object:     "event",
				Type:       stripe.EventTypeCustomerSubscriptionUpdated,
				APIVersion: stripe.APIVersion,
				Data: &stripe.EventData{
					Raw: json.RawMessage(jsonDataUpdate),
				},
			},
			expectedStatus: http.StatusOK,
		},
		{
			name: "valid payload - trial will end",
			payload: &stripe.Event{
				ID:         "evt_test_webhook_trial",
				Object:     "event",
				Type:       stripe.EventTypeCustomerSubscriptionTrialWillEnd,
				APIVersion: stripe.APIVersion,
				Data: &stripe.EventData{
					Raw: json.RawMessage(jsonDataUpdate),
				},
			},
			expectedStatus: http.StatusOK,
		},
		{
			name: "valid payload - payment method attached",
			payload: &stripe.Event{
				ID:         "evt_test_webhook_payment_method",
				Object:     "event",
				Type:       stripe.EventTypePaymentMethodAttached,
				APIVersion: stripe.APIVersion,
				Data: &stripe.EventData{
					Raw: json.RawMessage(jsonPaymentUpdate),
				},
			},
			expectedStatus: http.StatusOK,
		},
		{
			name: "unsupported event type",
			payload: &stripe.Event{
				ID:         "evt_test_webhook_unsupported",
				Object:     "event",
				Type:       stripe.EventTypeCustomerUpdated,
				APIVersion: stripe.APIVersion,
				Data: &stripe.EventData{
					Raw: json.RawMessage(`{"id":"cus_test","object":"customer"}`),
				},
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:                    "webhook with discard API version is ignored",
			queryParams:             "api_version=2024-10-28.acacia",
			configAPIVersion:        "2024-11-20.acacia",
			configDiscardAPIVersion: "2024-10-28.acacia",
			payload: &stripe.Event{
				ID:         "evt_test_webhook_discard",
				Object:     "event",
				Type:       stripe.EventTypeCustomerSubscriptionUpdated,
				APIVersion: stripe.APIVersion,
				Data: &stripe.EventData{
					Raw: json.RawMessage(jsonDataUpdate),
				},
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:             "webhook with mismatched API version is ignored",
			queryParams:      "api_version=2024-09-15.acacia",
			configAPIVersion: "2024-11-20.acacia",
			payload: &stripe.Event{
				ID:         "evt_test_webhook_mismatch",
				Object:     "event",
				Type:       stripe.EventTypeCustomerSubscriptionUpdated,
				APIVersion: stripe.APIVersion,
				Data: &stripe.EventData{
					Raw: json.RawMessage(jsonDataUpdate),
				},
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:                    "webhook with matching API version is processed",
			queryParams:             "api_version=2024-11-20.acacia",
			configAPIVersion:        "2024-11-20.acacia",
			configDiscardAPIVersion: "2024-10-28.acacia",
			payload: &stripe.Event{
				ID:         "evt_test_webhook_match",
				Object:     "event",
				Type:       stripe.EventTypeCustomerSubscriptionUpdated,
				APIVersion: stripe.APIVersion,
				Data: &stripe.EventData{
					Raw: json.RawMessage(jsonDataUpdate),
				},
			},
			expectedStatus: http.StatusOK,
		},
		{
			name: "webhook without API version query param is processed when no config",
			payload: &stripe.Event{
				ID:         "evt_test_webhook_no_param",
				Object:     "event",
				Type:       stripe.EventTypeCustomerSubscriptionUpdated,
				APIVersion: stripe.APIVersion,
				Data: &stripe.EventData{
					Raw: json.RawMessage(jsonDataUpdate),
				},
			},
			expectedStatus: http.StatusOK,
		},
		{
			name: "invalid signature returns bad request",
			payload: &stripe.Event{
				ID:         "evt_test_webhook_invalid_signature",
				Object:     "event",
				Type:       stripe.EventTypeCustomerSubscriptionUpdated,
				APIVersion: stripe.APIVersion,
				Data: &stripe.EventData{
					Raw: json.RawMessage(jsonDataUpdate),
				},
			},
			expectedStatus:    http.StatusBadRequest,
			signatureOverride: "invalid",
		},
	}

	for _, tc := range testCases {
		tc := tc

		t.Run(tc.name, func(t *testing.T) {
			t.Helper()

			origAPIVersion := suite.h.Entitlements.Config.StripeWebhookAPIVersion
			origDiscardVersion := suite.h.Entitlements.Config.StripeWebhookDiscardAPIVersion

			t.Cleanup(func() {
				suite.h.Entitlements.Config.StripeWebhookAPIVersion = origAPIVersion
				suite.h.Entitlements.Config.StripeWebhookDiscardAPIVersion = origDiscardVersion
			})

			suite.h.Entitlements.Config.StripeWebhookAPIVersion = tc.configAPIVersion
			suite.h.Entitlements.Config.StripeWebhookDiscardAPIVersion = tc.configDiscardAPIVersion

			var resources *webhookTestResources
			if tc.setup != nil {
				resources = tc.setup(t)
			}

			recorder := suite.executeWebhookRequest(t, tc.payload, tc.queryParams, tc.signatureOverride)
			require.Equal(t, tc.expectedStatus, recorder.Code)

			res := recorder.Result()
			require.NoError(t, res.Body.Close())

			if tc.validate != nil {
				tc.validate(t, resources)
			}
		})
	}
}
