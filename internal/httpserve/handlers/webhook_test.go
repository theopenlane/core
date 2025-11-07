package handlers_test

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stripe/stripe-go/v83"
	"github.com/stripe/stripe-go/v83/webhook"
	echo "github.com/theopenlane/echox"

	entEvent "github.com/theopenlane/core/internal/ent/generated/event"
	"github.com/theopenlane/core/internal/ent/generated/orgsubscription"
	"github.com/theopenlane/core/internal/ent/generated/privacy"
	models "github.com/theopenlane/core/pkg/openapi"
)

type webhookBinder struct {
	echo.Binder
}

// Bind is the echo binder override that preserves the request body for repeated reads
func (b *webhookBinder) Bind(c echo.Context, i interface{}) error {
	inner := b.Binder
	if inner == nil {
		inner = new(echo.DefaultBinder)
	}

	var bodyCopy []byte
	if req := c.Request(); req != nil && req.Body != nil {
		data, err := io.ReadAll(req.Body)
		if err != nil {
			return err
		}
		bodyCopy = data
		req.Body = io.NopCloser(bytes.NewReader(bodyCopy))
	}

	if err := inner.Bind(c, i); err != nil {
		return err
	}

	if bodyCopy != nil {
		c.Request().Body = io.NopCloser(bytes.NewReader(bodyCopy))
	}

	if req, ok := i.(*models.StripeWebhookRequest); ok && req.APIVersion == "" {
		req.APIVersion = c.QueryParam("api_version")
	}

	return nil
}

func (suite *HandlerTestSuite) TestWebhookReceiverHandler() {
	t := suite.T()

	allowCtx := privacy.DecisionContext(testUser1.UserCtx, privacy.Allow)

	ensureSecret := func(version string) {
		if version == "" {
			return
		}
		if suite.h.Entitlements.Config.StripeWebhookSecrets == nil {
			suite.h.Entitlements.Config.StripeWebhookSecrets = map[string]string{}
		}
		suite.h.Entitlements.Config.StripeWebhookSecrets[version] = webhookSecret
	}

	suite.db.Organization.UpdateOneID(testUser1.OrganizationID).
		SetStripeCustomerID("cus_test_customer").
		ExecX(allowCtx)

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

	strPtr := func(s string) *string { return &s }
	boolPtr := func(b bool) *bool { return &b }

	currentAPIVersion := stripe.APIVersion
	parts := strings.SplitN(currentAPIVersion, ".", 2)
	releaseSuffix := ""
	if len(parts) == 2 {
		releaseSuffix = "." + parts[1]
	}

	parseDate := func(version string) time.Time {
		datePart := strings.SplitN(version, ".", 2)[0]
		parsed, perr := time.Parse("2006-01-02", datePart)
		if perr != nil {
			return time.Now()
		}
		return parsed
	}

	baseDate := parseDate(currentAPIVersion)
	previousDate := baseDate.AddDate(0, 0, -14).Format("2006-01-02")
	olderDate := baseDate.AddDate(0, 0, -21).Format("2006-01-02")

	versionWithDate := func(date string) string {
		return date + releaseSuffix
	}

	discardAPIVersion := versionWithDate(previousDate)
	mismatchAPIVersion := versionWithDate(olderDate)

	testCases := []struct {
		name                    string
		payload                 *stripe.Event
		expectedStatus          int
		queryParams             string
		configAPIVersion        *string
		configDiscardAPIVersion *string
		signatureOverride       string
		expectRevocation        bool
		expectProcessed         *bool
	}{
		{
			name: "valid payload - paused subscription",
			payload: &stripe.Event{
				ID:         "evt_test_webhook_paused",
				Object:     "event",
				Type:       stripe.EventTypeCustomerSubscriptionPaused,
				APIVersion: currentAPIVersion,
				Data: &stripe.EventData{
					Raw: json.RawMessage(jsonDataUpdate),
				},
			},
			expectedStatus:          http.StatusOK,
			expectRevocation:        true,
			configAPIVersion:        strPtr(currentAPIVersion),
			configDiscardAPIVersion: strPtr(discardAPIVersion),
		},
		{
			name: "invalid payload, missing api_version",
			payload: &stripe.Event{
				ID:     "evt_test_webhook_missing_api_version",
				Object: "event",
				Type:   stripe.EventTypeCustomerSubscriptionUpdated,
				Data: &stripe.EventData{
					Raw: json.RawMessage(jsonDataUpdate),
				},
			},
			expectedStatus:          http.StatusBadRequest,
			configAPIVersion:        strPtr(currentAPIVersion),
			configDiscardAPIVersion: strPtr(discardAPIVersion),
			expectProcessed:         boolPtr(false),
		},
		{
			name: "valid payload - subscription updated",
			payload: &stripe.Event{
				ID:         "evt_test_webhook_subscription_updated",
				Object:     "event",
				Type:       stripe.EventTypeCustomerSubscriptionUpdated,
				APIVersion: currentAPIVersion,
				Data: &stripe.EventData{
					Raw: json.RawMessage(jsonDataUpdate),
				},
			},
			expectedStatus:          http.StatusOK,
			configAPIVersion:        strPtr(currentAPIVersion),
			configDiscardAPIVersion: strPtr(discardAPIVersion),
		},
		{
			name: "valid payload - trial will end",
			payload: &stripe.Event{
				ID:         "evt_test_webhook_trial",
				Object:     "event",
				Type:       stripe.EventTypeCustomerSubscriptionTrialWillEnd,
				APIVersion: currentAPIVersion,
				Data: &stripe.EventData{
					Raw: json.RawMessage(jsonDataUpdate),
				},
			},
			expectedStatus:          http.StatusOK,
			configAPIVersion:        strPtr(currentAPIVersion),
			configDiscardAPIVersion: strPtr(discardAPIVersion),
		},
		{
			name: "valid payload - payment method attached",
			payload: &stripe.Event{
				ID:         "evt_test_webhook_payment_method",
				Object:     "event",
				Type:       stripe.EventTypePaymentMethodAttached,
				APIVersion: currentAPIVersion,
				Data: &stripe.EventData{
					Raw: json.RawMessage(jsonPaymentUpdate),
				},
			},
			expectedStatus:          http.StatusOK,
			configAPIVersion:        strPtr(currentAPIVersion),
			configDiscardAPIVersion: strPtr(discardAPIVersion),
		},
		{
			name: "unsupported event type",
			payload: &stripe.Event{
				ID:         "evt_test_webhook_unsupported",
				Object:     "event",
				Type:       stripe.EventTypeCustomerUpdated,
				APIVersion: currentAPIVersion,
				Data: &stripe.EventData{
					Raw: json.RawMessage(`{"id":"cus_test","object":"customer"}`),
				},
			},
			expectedStatus:          http.StatusOK,
			configAPIVersion:        strPtr(currentAPIVersion),
			configDiscardAPIVersion: strPtr(discardAPIVersion),
			expectProcessed:         boolPtr(false),
		},
		{
			name:                    "webhook with discard API version is ignored",
			queryParams:             "api_version=" + discardAPIVersion,
			configAPIVersion:        strPtr(currentAPIVersion),
			configDiscardAPIVersion: strPtr(discardAPIVersion),
			payload: &stripe.Event{
				ID:         "evt_test_webhook_discard",
				Object:     "event",
				Type:       stripe.EventTypeCustomerSubscriptionUpdated,
				APIVersion: discardAPIVersion,
				Data: &stripe.EventData{
					Raw: json.RawMessage(jsonDataUpdate),
				},
			},
			expectedStatus:  http.StatusOK,
			expectProcessed: boolPtr(false),
		},
		{
			name:                    "webhook with mismatched API version is ignored",
			queryParams:             "api_version=" + mismatchAPIVersion,
			configAPIVersion:        strPtr(currentAPIVersion),
			configDiscardAPIVersion: strPtr(discardAPIVersion),
			payload: &stripe.Event{
				ID:         "evt_test_webhook_mismatch",
				Object:     "event",
				Type:       stripe.EventTypeCustomerSubscriptionUpdated,
				APIVersion: mismatchAPIVersion,
				Data: &stripe.EventData{
					Raw: json.RawMessage(jsonDataUpdate),
				},
			},
			expectedStatus:  http.StatusOK,
			expectProcessed: boolPtr(false),
		},
		{
			name:                    "webhook with matching API version is processed",
			queryParams:             "api_version=" + currentAPIVersion,
			configAPIVersion:        strPtr(currentAPIVersion),
			configDiscardAPIVersion: strPtr(discardAPIVersion),
			payload: &stripe.Event{
				ID:         "evt_test_webhook_match",
				Object:     "event",
				Type:       stripe.EventTypeCustomerSubscriptionUpdated,
				APIVersion: currentAPIVersion,
				Data: &stripe.EventData{
					Raw: json.RawMessage(jsonDataUpdate),
				},
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:                    "webhook with API version query param is processed when config unset",
			queryParams:             "api_version=" + currentAPIVersion,
			configAPIVersion:        strPtr(""),
			configDiscardAPIVersion: strPtr(""),
			payload: &stripe.Event{
				ID:         "evt_test_webhook_param_config_empty",
				Object:     "event",
				Type:       stripe.EventTypeCustomerSubscriptionUpdated,
				APIVersion: currentAPIVersion,
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
				APIVersion: currentAPIVersion,
				Data: &stripe.EventData{
					Raw: json.RawMessage(jsonDataUpdate),
				},
			},
			expectedStatus:          http.StatusOK,
			configAPIVersion:        strPtr(""),
			configDiscardAPIVersion: strPtr(""),
		},
		{
			name: "webhook without API version query param is processed when config set",
			payload: &stripe.Event{
				ID:         "evt_test_webhook_no_param_configured",
				Object:     "event",
				Type:       stripe.EventTypeCustomerSubscriptionUpdated,
				APIVersion: currentAPIVersion,
				Data: &stripe.EventData{
					Raw: json.RawMessage(jsonDataUpdate),
				},
			},
			expectedStatus:          http.StatusOK,
			configAPIVersion:        strPtr(currentAPIVersion),
			configDiscardAPIVersion: strPtr(discardAPIVersion),
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
			expectedStatus:          http.StatusBadRequest,
			signatureOverride:       "invalid",
			configAPIVersion:        strPtr(currentAPIVersion),
			configDiscardAPIVersion: strPtr(discardAPIVersion),
			expectProcessed:         boolPtr(false),
		},
	}

	for _, tc := range testCases {
		tc := tc

		t.Run(tc.name, func(t *testing.T) {
			suite.stripeMockBackend.ExpectedCalls = suite.stripeMockBackend.ExpectedCalls[:0]
			suite.orgSubscriptionMocks()

			suite.db.OrgSubscription.Update().
				Where(orgsubscription.StripeSubscriptionID(seedStripeSubscriptionID)).
				SetActive(true).
				SetStripeSubscriptionStatus("active").
				ExecX(allowCtx)

			var (
				apiTokenID string
				patID      string
			)

			if tc.expectRevocation {
				apiToken := suite.db.APIToken.Create().
					SetOwnerID(testUser1.OrganizationID).
					SetName("test_token").
					SaveX(allowCtx)
				apiTokenID = apiToken.ID

				pat := suite.db.PersonalAccessToken.Create().
					SetOwnerID(testUser1.ID).
					AddOrganizationIDs(testUser1.OrganizationID).
					SetName("test_token").
					SaveX(allowCtx)
				patID = pat.ID

				t.Cleanup(func() {
					_ = suite.db.APIToken.DeleteOneID(apiTokenID).Exec(allowCtx)
					_ = suite.db.PersonalAccessToken.DeleteOneID(patID).Exec(allowCtx)
				})
			}

			origAPIVersion := suite.h.Entitlements.Config.StripeWebhookAPIVersion
			origDiscardVersion := suite.h.Entitlements.Config.StripeWebhookDiscardAPIVersion

			t.Cleanup(func() {
				suite.h.Entitlements.Config.StripeWebhookAPIVersion = origAPIVersion
				suite.h.Entitlements.Config.StripeWebhookDiscardAPIVersion = origDiscardVersion
			})

			if tc.configAPIVersion != nil {
				suite.h.Entitlements.Config.StripeWebhookAPIVersion = *tc.configAPIVersion
			} else {
				suite.h.Entitlements.Config.StripeWebhookAPIVersion = origAPIVersion
			}

			if tc.configDiscardAPIVersion != nil {
				suite.h.Entitlements.Config.StripeWebhookDiscardAPIVersion = *tc.configDiscardAPIVersion
			} else {
				suite.h.Entitlements.Config.StripeWebhookDiscardAPIVersion = origDiscardVersion
			}

			if tc.configAPIVersion != nil {
				ensureSecret(*tc.configAPIVersion)
			}
			if tc.configDiscardAPIVersion != nil {
				ensureSecret(*tc.configDiscardAPIVersion)
			}
			ensureSecret(string(tc.payload.APIVersion))
			if tc.queryParams != "" {
				values, err := url.ParseQuery(tc.queryParams)
				require.NoError(t, err)
				ensureSecret(values.Get("api_version"))
			}

			payloadBytes, err := json.Marshal(tc.payload)
			require.NoError(t, err)

			signedPayload := webhook.GenerateTestSignedPayload(&webhook.UnsignedPayload{
				Payload: payloadBytes,
				Secret:  webhookSecret,
			})

			requestURL := "/webhook"
			if tc.queryParams != "" {
				requestURL += "?" + tc.queryParams
			}

			req := httptest.NewRequest(http.MethodPost, requestURL, bytes.NewReader(signedPayload.Payload))
			req.ContentLength = int64(len(signedPayload.Payload))
			req.Header.Set("Content-Type", "application/json")

			signature := signedPayload.Header
			if tc.signatureOverride != "" {
				signature = tc.signatureOverride
			}
			req.Header.Set("Stripe-Signature", signature)

			recorder := httptest.NewRecorder()
			suite.e.ServeHTTP(recorder, req)

			res := recorder.Result()
			bodyBytes, readErr := io.ReadAll(res.Body)
			require.NoError(t, readErr)

			if recorder.Code != tc.expectedStatus {
				t.Fatalf("expected status %d got %d, response: %s", tc.expectedStatus, recorder.Code, string(bodyBytes))
			}

			require.NoError(t, res.Body.Close())

			if tc.expectRevocation && tc.expectedStatus == http.StatusOK {
				sub, err := suite.db.OrgSubscription.Query().
					Where(orgsubscription.StripeSubscriptionID(seedStripeSubscriptionID)).
					Only(testUser1.UserCtx)
				require.NoError(t, err)
				assert.False(t, sub.Active)

				apiToken, err := suite.db.APIToken.Get(testUser1.UserCtx, apiTokenID)
				require.NoError(t, err)
				require.False(t, apiToken.IsActive)
				require.NotEmpty(t, apiToken.RevokedBy)
				require.NotEmpty(t, apiToken.RevokedAt)
				require.NotEmpty(t, apiToken.RevokedReason)
				assert.Less(t, *apiToken.ExpiresAt, time.Now())

				pat, err := suite.db.PersonalAccessToken.Get(testUser1.UserCtx, patID)
				require.NoError(t, err)
				require.Len(t, pat.Edges.Organizations, 0)
			}

			if tc.expectProcessed != nil {
				exists, err := suite.db.Event.Query().
					Where(entEvent.IDEQ(tc.payload.ID)).
					Exist(testUser1.UserCtx)
				require.NoError(t, err)
				assert.Equal(t, *tc.expectProcessed, exists)
			}
		})
	}
}
