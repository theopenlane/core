package handlers_test

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"github.com/stripe/stripe-go/v81"
	"github.com/stripe/stripe-go/v81/webhook"
	"github.com/theopenlane/core/internal/ent/generated/orgsubscription"
	"github.com/theopenlane/core/internal/ent/generated/privacy"
	"github.com/theopenlane/core/pkg/entitlements"
	"github.com/theopenlane/core/pkg/entitlements/mocks"
)

var (
	seedStripeSubscriptionID = "sub_test_subscription"
)

func (suite *HandlerTestSuite) TestWebhookReceiverHandler() {
	t := suite.T()

	// manually create an org subscription for the org and set it as active since this does not happen automatically
	// in tests
	suite.db.OrgSubscription.Create().
		SetStripeSubscriptionStatus("active").
		SetStripeSubscriptionID("PENDING_UPDATE").
		SetOwnerID(testUser1.OrganizationID).
		SetStripeSubscriptionID(seedStripeSubscriptionID).
		ExecX(testUser1.UserCtx)

	// setup mock entitlements client
	entitlements, err := suite.mockStripeClient()
	require.NoError(t, err)

	suite.h.DBClient.EntitlementManager = entitlements
	suite.h.Entitlements = entitlements

	// add handler
	suite.e.POST("/webhook", suite.h.WebhookReceiverHandler)

	// setup payloads based on the mock customer
	// update subscription payload
	dataUpdate := mockCustomer
	dataUpdate.Subscriptions.Data[0].Items.Data[0].Price.UnitAmount = 900

	sub := dataUpdate.Subscriptions.Data[0]
	jsonDataUpdate, err := json.Marshal(sub)
	require.NoError(t, err)

	// payment update payload
	paymentUpdate :=
		&stripe.PaymentMethod{
			ID: "pm_test_payment",
		}
	jsonPaymentUpdate, err := json.Marshal(paymentUpdate)
	require.NoError(t, err)

	// create api token and personal access token to ensure they are revoked when subscription is paused
	allowCtx := privacy.DecisionContext(testUser1.UserCtx, privacy.Allow)
	apiToken := suite.db.APIToken.Create().
		SetOwnerID(testUser1.OrganizationID).
		SetName(
			"test_token",
		).SaveX(allowCtx)

	pat := suite.db.PersonalAccessToken.Create().
		SetOwnerID(testUser1.ID).AddOrganizationIDs(testUser1.OrganizationID).
		SetName(
			"test_token",
		).SaveX(allowCtx)

	testCases := []struct {
		name           string
		payload        *stripe.Event
		expectedStatus int
	}{
		{
			name: "valid payload - paused subscription",
			payload: &stripe.Event{
				ID:         "evt_test_webhook",
				Type:       stripe.EventTypeCustomerSubscriptionPaused,
				APIVersion: stripe.APIVersion,
				Data: &stripe.EventData{
					Raw: json.RawMessage(jsonDataUpdate),
				},
			},
			expectedStatus: http.StatusOK,
		},
		{
			name: "invalid payload, missing api_version",
			payload: &stripe.Event{
				ID:   "evt_test_webhook",
				Type: stripe.EventTypeCustomerSubscriptionUpdated,
				Data: &stripe.EventData{
					Raw: json.RawMessage(jsonDataUpdate),
				},
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "valid payload - subscription updated",
			payload: &stripe.Event{
				ID:         "evt_test_webhook",
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
				ID:         "evt_test_webhook",
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
				ID:         "evt_test_webhook",
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
				ID:         "evt_test_webhook",
				Type:       stripe.EventTypeCustomerUpdated,
				APIVersion: stripe.APIVersion,
			},
			expectedStatus: http.StatusOK,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// clear expected calls before each test
			suite.stripeMockBackend.ExpectedCalls = suite.stripeMockBackend.ExpectedCalls[:0]
			suite.orgSubscriptionMocks()

			payloadBytes, err := json.Marshal(tc.payload)
			require.NoError(t, err)

			signedPayload := webhook.GenerateTestSignedPayload(&webhook.UnsignedPayload{Payload: payloadBytes, Secret: webhookSecret})

			req := httptest.NewRequest(http.MethodPost, "/webhook", io.NopCloser(strings.NewReader(string(signedPayload.Payload))))

			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("Stripe-Signature", signedPayload.Header)

			recorder := httptest.NewRecorder()

			suite.e.ServeHTTP(recorder, req)

			res := recorder.Result()
			defer res.Body.Close()

			require.Equal(t, tc.expectedStatus, recorder.Code)

			if tc.payload.Type == stripe.EventTypeCustomerSubscriptionPaused {
				// check if the subscription is paused
				sub, err := suite.db.OrgSubscription.Query().
					Where(orgsubscription.StripeSubscriptionID(seedStripeSubscriptionID)).
					Only(testUser1.UserCtx)

				require.NoError(t, err)
				assert.False(t, sub.Active)

				// check if token is expired
				apiToken, err := suite.db.APIToken.Get(testUser1.UserCtx, apiToken.ID)
				require.NoError(t, err)
				require.False(t, apiToken.IsActive)
				require.NotEmpty(t, apiToken.RevokedBy)
				require.NotEmpty(t, apiToken.RevokedAt)
				require.NotEmpty(t, apiToken.RevokedReason)
				assert.Less(t, *apiToken.ExpiresAt, time.Now())

				// check if personal access token is expired
				pat, err := suite.db.PersonalAccessToken.Get(testUser1.UserCtx, pat.ID)
				require.NoError(t, err)
				require.Len(t, pat.Edges.Organizations, 0)
			}
		})
	}
}

// mockStripeClient creates a new stripe client with mock backend
func (suite *HandlerTestSuite) mockStripeClient() (*entitlements.StripeClient, error) {
	suite.stripeMockBackend = new(mocks.MockStripeBackend)
	stripeTestBackends := &stripe.Backends{
		API:     suite.stripeMockBackend,
		Connect: suite.stripeMockBackend,
		Uploads: suite.stripeMockBackend,
	}

	suite.orgSubscriptionMocks()

	return entitlements.NewStripeClient(entitlements.WithAPIKey("sk_test_testing"),
		entitlements.WithConfig(entitlements.Config{
			StripeWebhookSecret: webhookSecret,
		},
		),
		entitlements.WithBackends(stripeTestBackends),
	)
}

// mockCustomer for webhook tests
var mockCustomer = &stripe.Customer{
	ID: "cus_test_customer",
	Subscriptions: &stripe.SubscriptionList{
		Data: []*stripe.Subscription{
			{
				Customer: &stripe.Customer{
					ID: "cus_test_customer",
				},
				ID: seedStripeSubscriptionID,
				Items: &stripe.SubscriptionItemList{
					Data: []*stripe.SubscriptionItem{
						{
							Price: &stripe.Price{
								UnitAmount: 1000,
								ID:         "price_test_price",
								Currency:   "usd",
								Recurring: &stripe.PriceRecurring{
									Interval: "month",
								},
							},
						},
					},
				},
			},
		},
	},
}

// orgSubscriptionMocks mocks the stripe calls for org subscription during the webhook tests
func (suite *HandlerTestSuite) orgSubscriptionMocks() {
	// setup mocks for search
	suite.stripeMockBackend.On("CallRaw", mock.Anything, mock.Anything, mock.Anything, mock.AnythingOfType("*stripe.Params"), mock.AnythingOfType("*stripe.CustomerSearchResult")).Run(func(args mock.Arguments) {
		mockCustomerSearchResult := args.Get(4).(*stripe.CustomerSearchResult)

		data := []*stripe.Customer{}
		data = append(data, mockCustomer)
		*mockCustomerSearchResult = stripe.CustomerSearchResult{
			Data: data,
		}

	}).Return(nil)

	// setup mocks for get customer by id
	suite.stripeMockBackend.On("Call", mock.Anything, mock.Anything, mock.Anything, mock.AnythingOfType("*stripe.CustomerParams"), mock.AnythingOfType("*stripe.Customer")).Run(func(args mock.Arguments) {
		mockCustomerSearchResult := args.Get(4).(*stripe.Customer)

		*mockCustomerSearchResult = *mockCustomer

	}).Return(nil)

	// setup mocks for getting entitlements
	suite.stripeMockBackend.On("CallRaw", mock.Anything, mock.Anything, mock.Anything, mock.AnythingOfType("*stripe.Params"), mock.AnythingOfType("*stripe.EntitlementsActiveEntitlementList")).Run(func(args mock.Arguments) {
		mockCustomerSearchResult := args.Get(4).(*stripe.EntitlementsActiveEntitlementList)

		*mockCustomerSearchResult = stripe.EntitlementsActiveEntitlementList{
			Data: []*stripe.EntitlementsActiveEntitlement{
				{
					Feature: &stripe.EntitlementsFeature{
						ID:        "feat_test_feature",
						LookupKey: "test_feature",
					},
				},
			},
		}

	}).Return(nil)
}
