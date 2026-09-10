//go:build test

package testharness

import (
	"encoding/json"
	"time"

	"github.com/stretchr/testify/mock"
	"github.com/stripe/stripe-go/v86"
	"github.com/theopenlane/core/v2/pkg/entitlements"
	"github.com/theopenlane/core/v2/pkg/entitlements/mocks"
	"github.com/theopenlane/utils/ulids"
)

// MockStripeClient creates a new stripe client with mock backend
func (suite *GraphTestSuite) MockStripeClient() (*entitlements.StripeClient, error) {
	suite.StripeMockBackend = new(mocks.MockStripeBackend)
	stripeTestBackends := &stripe.Backends{
		API:     suite.StripeMockBackend,
		Connect: suite.StripeMockBackend,
		Uploads: suite.StripeMockBackend,
	}

	suite.OrgSubscriptionMocks()

	return entitlements.NewStripeClient(entitlements.WithAPIKey("sk_test_testing"),
		entitlements.WithConfig(entitlements.Config{
			Enabled:             true,
			StripeWebhookSecret: webhookSecret,
		},
		),
		entitlements.WithBackends(stripeTestBackends),
	)
}

var MockItems = []*stripe.SubscriptionItem{
	{
		Price: &stripe.Price{
			Product: &stripe.Product{
				ID: "prod_test_product",
			},
			ID: "price_test_price",
			Recurring: &stripe.PriceRecurring{
				Interval: "month",
			},
			Currency:   "usd",
			UnitAmount: 1000,
		},
	},
}

// MockCustomer for webhook tests
var MockCustomer = &stripe.Customer{
	ID: "cus_test_customer",
	Subscriptions: &stripe.SubscriptionList{
		Data: []*stripe.Subscription{
			{
				Customer: &stripe.Customer{
					ID: "cus_test_customer",
				},
				ID: seedStripeSubscriptionID,
				Items: &stripe.SubscriptionItemList{
					Data: MockItems,
				},
			},
		},
	},
}

var MockSubscription = &stripe.Subscription{
	ID:     "sub_test_subscription",
	Status: "active",
	Items: &stripe.SubscriptionItemList{
		Data: MockItems,
	},
	Metadata: map[string]string{
		"organization_id": ulids.New().String(),
	},
	Customer: &stripe.Customer{
		ID: "cus_test_customer",
	},
	TrialEnd:     time.Now().Add(7 * 24 * time.Hour).Unix(), // 7 days from now
	DaysUntilDue: 15,
}

var MockProduct = &stripe.Product{
	ID:   "prod_test_product",
	Name: "Test Product",
}

// OrgSubscriptionMocks mocks the stripe calls for org subscription during the webhook tests
func (suite *GraphTestSuite) OrgSubscriptionMocks() {
	// setup mocks for get customer by id
	suite.StripeMockBackend.On("Call", mock.Anything, mock.Anything, mock.Anything, mock.AnythingOfType("*stripe.CustomerRetrieveParams"), mock.AnythingOfType("*stripe.Customer")).Run(func(args mock.Arguments) {
		mockCustomerSearchResult := args.Get(4).(*stripe.Customer)

		*mockCustomerSearchResult = *MockCustomer

	}).Return(nil)

	// setup mocks for creating customer params
	suite.StripeMockBackend.On("Call", mock.Anything, mock.Anything, mock.Anything, mock.AnythingOfType("*stripe.CustomerCreateParams"), mock.AnythingOfType("*stripe.Customer")).Run(func(args mock.Arguments) {
		mockCustomerSearchResult := args.Get(4).(*stripe.Customer)

		*mockCustomerSearchResult = *MockCustomer

	}).Return(nil)

	// mock customer search
	suite.StripeMockBackend.On("CallRaw", mock.Anything, mock.Anything, mock.Anything, mock.AnythingOfType("*stripe.Params"), mock.AnythingOfType("*stripe.v1SearchPage[*github.com/stripe/stripe-go/v86.Customer]")).Run(func(args mock.Arguments) {
		out := args.Get(4) // this is *v1SearchPage[*stripe.Customer] now, but unexported

		// Build a payload that matches Stripe search response shape
		payload := map[string]any{
			"object":   "search_result",
			"data":     []*stripe.Customer{MockCustomer},
			"has_more": false,
		}

		b, _ := json.Marshal(payload)
		_ = json.Unmarshal(b, out)
	}).Return(nil)

	// mock for subscription create params
	suite.StripeMockBackend.On("Call", mock.Anything, mock.Anything, mock.Anything, mock.AnythingOfType("*stripe.SubscriptionCreateParams"), mock.AnythingOfType("*stripe.Subscription")).Run(func(args mock.Arguments) {
		mockSubscriptionSearchResult := args.Get(4).(*stripe.Subscription)

		*mockSubscriptionSearchResult = *MockSubscription

	}).Return(nil)

	// mock for product retrieve params
	suite.StripeMockBackend.On("Call", mock.Anything, mock.Anything, mock.Anything, mock.AnythingOfType("*stripe.ProductRetrieveParams"), mock.AnythingOfType("*stripe.Product")).Run(func(args mock.Arguments) {
		mockProductRetrieveResult := args.Get(4).(*stripe.Product)

		*mockProductRetrieveResult = *MockProduct

	}).Return(nil)

	// mock for product params
	suite.StripeMockBackend.On("Call", mock.Anything, mock.Anything, mock.Anything, mock.AnythingOfType("*stripe.SubscriptionRetrieveParams"), mock.AnythingOfType("*stripe.Product")).Run(func(args mock.Arguments) {
		mockSubscriptionRetrieveResult := args.Get(4).(*stripe.Subscription)

		*mockSubscriptionRetrieveResult = *MockSubscription

	}).Return(nil)

	// setup mocks for getting entitlements
	suite.StripeMockBackend.On("CallRaw", mock.Anything, mock.Anything, mock.Anything, mock.AnythingOfType("*stripe.Params"), mock.AnythingOfType("*stripe.EntitlementsActiveEntitlementList")).Run(func(args mock.Arguments) {
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

	// setup mocks for subscription schedule creation
	suite.StripeMockBackend.On("Call", mock.Anything, mock.Anything, mock.Anything, mock.AnythingOfType("*stripe.SubscriptionScheduleCreateParams"), mock.AnythingOfType("*stripe.SubscriptionSchedule")).Run(func(args mock.Arguments) {
		mockSubscriptionScheduleResult := args.Get(4).(*stripe.SubscriptionSchedule)

		*mockSubscriptionScheduleResult = stripe.SubscriptionSchedule{
			ID:     "sched_test_schedule",
			Status: "active",
		}

	}).Return(nil)

	// setup mocks for customer update params
	suite.StripeMockBackend.On("Call", mock.Anything, mock.Anything, mock.Anything, mock.AnythingOfType("*stripe.CustomerUpdateParams"), mock.AnythingOfType("*stripe.Customer")).Run(func(args mock.Arguments) {
		mockCustomerUpdateResult := args.Get(4).(*stripe.Customer)

		*mockCustomerUpdateResult = *MockCustomer

	}).Return(nil)
}
