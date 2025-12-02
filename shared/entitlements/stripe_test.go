package entitlements_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stripe/stripe-go/v83"

	"github.com/theopenlane/shared/entitlements"
	"github.com/theopenlane/shared/entitlements/mocks"
)

func TestNew(t *testing.T) {
	c := assert.New(t)

	t.Setenv("STRIPE_SECRET_KEY", "secret_key")

	stripeService, err := entitlements.NewStripeClient(entitlements.WithAPIKey("secret_key"))
	c.NoError(err)
	c.IsType(stripeService, stripeService)
}

func TestNewErrMissingAPIKey(t *testing.T) {
	c := assert.New(t)

	stripeService, err := entitlements.NewStripeClient()
	c.Nil(stripeService)
	c.ErrorIs(err, entitlements.ErrMissingAPIKey)
}

func TestWithConfig(t *testing.T) {
	config := entitlements.Config{
		PrivateStripeKey: "private_key",
	}

	option := entitlements.WithConfig(config)
	client := &entitlements.StripeClient{}

	option(client)

	if client.Config.PrivateStripeKey != config.PrivateStripeKey {
		t.Errorf("expected config %v, got %v", config, client.Config)
	}
}

func TestNewConfig(t *testing.T) {
	tests := []struct {
		name string
		opts []entitlements.ConfigOpts
		want *entitlements.Config
	}{
		{
			name: "custom config",
			opts: []entitlements.ConfigOpts{
				entitlements.WithEnabled(true),
				entitlements.WithPrivateStripeKey("private_key"),
				entitlements.WithStripeWebhookSecret("webhook_secret"),
				entitlements.WithStripeWebhookURL("https://custom.webhook.url"),
				entitlements.WithStripeBillingPortalSuccessURL("https://custom.billing.success.url"),
				entitlements.WithStripeCancellationReturnURL("https://custom.cancellation.return.url"),
				entitlements.WithStripeWebhookEvents([]string{"invoice.paid"}),
			},
			want: &entitlements.Config{
				Enabled:                       true,
				PrivateStripeKey:              "private_key",
				StripeWebhookSecret:           "webhook_secret",
				StripeWebhookURL:              "https://custom.webhook.url",
				StripeBillingPortalSuccessURL: "https://custom.billing.success.url",
				StripeCancellationReturnURL:   "https://custom.cancellation.return.url",
				StripeWebhookEvents:           []string{"invoice.paid"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := entitlements.NewConfig(tt.opts...)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestCreateCustomer(t *testing.T) {
	c := assert.New(t)

	expectedCustomer := &stripe.Customer{
		ID:    "cus_123",
		Email: "test@example.com",
		Name:  "org_123",
		Phone: "1234567890",
		Address: &stripe.Address{
			Line1:      "123 Main St",
			City:       "Anytown",
			State:      "CA",
			PostalCode: "12345",
			Country:    "US",
		},
		Metadata: map[string]string{
			"organization_id":          "org_123",
			"organization_settings_id": "settings_123",
			"organization_name":        "Test Organization",
		},
	}

	orgCustomer := &entitlements.OrganizationCustomer{
		OrganizationID:         "org_123",
		OrganizationSettingsID: "settings_123",
		OrganizationName:       "Test Organization",
		ContactInfo: entitlements.ContactInfo{
			Email: "test@example.com",
			Phone: "1234567890",
		},
	}

	stripeBackendMock := new(mocks.MockStripeBackend)
	stripeTestBackends := &stripe.Backends{
		API:     stripeBackendMock,
		Connect: stripeBackendMock,
		Uploads: stripeBackendMock,
	}

	stripeBackendMock.On("Call", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Run(func(args mock.Arguments) {
		mockCustomerResult := args.Get(4).(*stripe.Customer)

		*mockCustomerResult = *expectedCustomer
	}).Return(nil)

	mockStripeClient := stripe.NewClient("sk_test", stripe.WithBackends(stripeTestBackends))

	service := entitlements.StripeClient{
		Client: mockStripeClient,
	}

	ctx := context.Background()

	customer, err := service.CreateCustomer(ctx, orgCustomer)
	c.NoError(err)
	c.Equal(expectedCustomer, customer)
}

func TestUpdateCustomer(t *testing.T) {
	c := assert.New(t)

	expectedCustomer := &stripe.Customer{
		ID:    "cus_123",
		Email: "updated@example.com",
		Phone: "0987654321",
	}

	updateParams := &stripe.CustomerUpdateParams{
		Email: stripe.String("updated@example.com"),
		Phone: stripe.String("0987654321"),
	}

	stripeBackendMock := new(mocks.MockStripeBackend)
	stripeTestBackends := &stripe.Backends{
		API:     stripeBackendMock,
		Connect: stripeBackendMock,
		Uploads: stripeBackendMock,
	}

	stripeBackendMock.On("Call", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Run(func(args mock.Arguments) {
		mockCustomerResult := args.Get(4).(*stripe.Customer)

		*mockCustomerResult = *expectedCustomer
	}).Return(nil)

	mockStripeClient, err := entitlements.NewStripeClient(
		entitlements.WithAPIKey("sk_test"),
		entitlements.WithBackends(stripeTestBackends),
	)
	c.NoError(err)

	service := entitlements.StripeClient{
		Client: mockStripeClient.Client,
	}

	customer, err := service.UpdateCustomer(context.Background(), "cus_123", updateParams)
	c.NoError(err)
	c.Equal(expectedCustomer, customer)
}

func TestDeleteCustomer(t *testing.T) {
	c := assert.New(t)

	expectedCustomer := &stripe.Customer{
		ID: "cus_123",
	}

	stripeBackendMock := new(mocks.MockStripeBackend)
	stripeTestBackends := &stripe.Backends{
		API:     stripeBackendMock,
		Connect: stripeBackendMock,
		Uploads: stripeBackendMock,
	}

	stripeBackendMock.On("Call", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Run(func(args mock.Arguments) {
		mockCustomerResult := args.Get(4).(*stripe.Customer)

		*mockCustomerResult = *expectedCustomer
	}).Return(nil)

	mockStripeClient, err := entitlements.NewStripeClient(
		entitlements.WithAPIKey("sk_test"),
		entitlements.WithBackends(stripeTestBackends),
	)
	c.NoError(err)

	service := entitlements.StripeClient{
		Client: mockStripeClient.Client,
	}

	err = service.DeleteCustomer(context.Background(), "cus_123")
	c.NoError(err)
}

func TestCreateSubscription(t *testing.T) {
	c := assert.New(t)

	expectedSubscription := &stripe.Subscription{
		ID: "sub_123",
	}

	subscriptionParams := &stripe.SubscriptionCreateParams{
		Customer: stripe.String("cus_123"),
		Items: []*stripe.SubscriptionCreateItemParams{
			{
				Price: stripe.String("price_123"),
			},
		},
	}

	stripeBackendMock := new(mocks.MockStripeBackend)
	stripeTestBackends := &stripe.Backends{
		API:     stripeBackendMock,
		Connect: stripeBackendMock,
		Uploads: stripeBackendMock,
	}

	stripeBackendMock.On("Call", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Run(func(args mock.Arguments) {
		mockSubscriptionResult := args.Get(4).(*stripe.Subscription)

		*mockSubscriptionResult = *expectedSubscription
	}).Return(nil)

	mockStripeClient := stripe.NewClient("sk_test", stripe.WithBackends(stripeTestBackends))

	service := entitlements.StripeClient{
		Client: mockStripeClient,
	}

	subscription, err := service.CreateSubscription(context.Background(), subscriptionParams)
	c.NoError(err)
	c.Equal(expectedSubscription, subscription)
}

func TestUpdateSubscription(t *testing.T) {
	c := assert.New(t)

	expectedSubscription := &stripe.Subscription{
		ID: "sub_123",
	}

	updateParams := &stripe.SubscriptionUpdateParams{
		Items: []*stripe.SubscriptionUpdateItemParams{
			{
				Price: stripe.String("price_456"),
			},
		},
	}

	stripeBackendMock := new(mocks.MockStripeBackend)
	stripeTestBackends := &stripe.Backends{
		API:     stripeBackendMock,
		Connect: stripeBackendMock,
		Uploads: stripeBackendMock,
	}

	stripeBackendMock.On("Call", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Run(func(args mock.Arguments) {
		mockSubscriptionResult := args.Get(4).(*stripe.Subscription)

		*mockSubscriptionResult = *expectedSubscription
	}).Return(nil)

	mockStripeClient := stripe.NewClient("sk_test", stripe.WithBackends(stripeTestBackends))

	service := entitlements.StripeClient{
		Client: mockStripeClient,
	}

	subscription, err := service.UpdateSubscription(context.Background(), "sub_123", updateParams)
	c.NoError(err)
	c.Equal(expectedSubscription, subscription)
}

func TestCancelSubscription(t *testing.T) {
	c := assert.New(t)

	expectedSubscription := &stripe.Subscription{
		ID: "sub_123",
	}

	cancelParams := &stripe.SubscriptionCancelParams{}

	stripeBackendMock := new(mocks.MockStripeBackend)
	stripeTestBackends := &stripe.Backends{
		API:     stripeBackendMock,
		Connect: stripeBackendMock,
		Uploads: stripeBackendMock,
	}

	stripeBackendMock.On("Call", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Run(func(args mock.Arguments) {
		mockSubscriptionResult := args.Get(4).(*stripe.Subscription)

		*mockSubscriptionResult = *expectedSubscription
	}).Return(nil)

	mockStripeClient := stripe.NewClient("sk_test", stripe.WithBackends(stripeTestBackends))

	service := entitlements.StripeClient{
		Client: mockStripeClient,
	}

	subscription, err := service.CancelSubscription(context.Background(), "sub_123", cancelParams)
	c.NoError(err)
	c.Equal(expectedSubscription, subscription)
}

func TestMapStripeSubscription(t *testing.T) {
	c := assert.New(t)

	stripeSubscription := &stripe.Subscription{
		ID: "sub_123",
		Items: &stripe.SubscriptionItemList{
			Data: []*stripe.SubscriptionItem{
				{
					Price: &stripe.Price{
						ID:         "price_123",
						UnitAmount: 1000,
						Recurring:  &stripe.PriceRecurring{Interval: "month"},
						Currency:   "usd",
						Product:    &stripe.Product{ID: "prod_123"},
					},
				},
			},
		},
		TrialEnd: 1620000000,
		Status:   "active",
		Customer: &stripe.Customer{ID: "cus_123"},
		Metadata: map[string]string{
			"organization_id": "org_123",
		},
	}

	stripeSubscriptionSchedule := &stripe.SubscriptionSchedule{
		ID: "sub_sched_123",
	}

	expectedSubscription := &entitlements.Subscription{
		ID: "sub_123",
		Prices: []entitlements.Price{
			{
				ID:          "price_123",
				Price:       10.00,
				ProductID:   "prod_123",
				ProductName: "Test Product",
				Interval:    "month",
				Currency:    "usd",
			},
		},
		TrialEnd:                     1620000000,
		ProductID:                    "prod_123",
		Status:                       "active",
		StripeCustomerID:             "cus_123",
		OrganizationID:               "org_123",
		StripeSubscriptionScheduleID: "sub_sched_123",
	}

	stripeBackendMock := new(mocks.MockStripeBackend)
	stripeTestBackends := &stripe.Backends{
		API:     stripeBackendMock,
		Connect: stripeBackendMock,
		Uploads: stripeBackendMock,
	}

	stripeBackendMock.On("Call", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Run(func(args mock.Arguments) {
		mockProductResult := args.Get(4).(*stripe.Product)

		*mockProductResult = stripe.Product{
			ID:   "prod_123",
			Name: "Test Product",
		}
	}).Return(nil)

	mockStripeClient := stripe.NewClient("sk_test", stripe.WithBackends(stripeTestBackends))

	service := entitlements.StripeClient{
		Client: mockStripeClient,
	}

	subscription := service.MapStripeSubscription(context.Background(), stripeSubscription, stripeSubscriptionSchedule)
	c.Equal(expectedSubscription, subscription)
}

func TestCreateSubscriptionWithOptions_MultipleItems(t *testing.T) {
	c := assert.New(t)

	expectedSubscription := &stripe.Subscription{
		ID: "sub_multi",
	}

	subscriptionParams := &stripe.SubscriptionCreateParams{
		Customer: stripe.String("cus_multi"),
	}
	items := []*stripe.SubscriptionCreateItemParams{
		{Price: stripe.String("price_1")},
		{Price: stripe.String("price_2")},
	}

	stripeBackendMock := new(mocks.MockStripeBackend)
	stripeTestBackends := &stripe.Backends{
		API:     stripeBackendMock,
		Connect: stripeBackendMock,
		Uploads: stripeBackendMock,
	}

	stripeBackendMock.On("Call", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Run(func(args mock.Arguments) {
		mockSubscriptionResult := args.Get(4).(*stripe.Subscription)
		*mockSubscriptionResult = *expectedSubscription
	}).Return(nil)

	mockStripeClient := stripe.NewClient("sk_test", stripe.WithBackends(stripeTestBackends))

	service := entitlements.StripeClient{
		Client: mockStripeClient,
	}

	subscription, err := service.CreateSubscriptionWithOptions(context.Background(), subscriptionParams, entitlements.WithSubscriptionItems(items...))
	c.NoError(err)
	c.Equal(expectedSubscription, subscription)
}

func TestUpdateSubscriptionWithOptions_AddNewItemsIfNotExist(t *testing.T) {
	c := assert.New(t)

	existingItems := []*stripe.SubscriptionItem{
		{ID: "item_1", Price: &stripe.Price{ID: "price_1"}},
	}
	newItems := []*stripe.SubscriptionUpdateItemParams{
		{Price: stripe.String("price_2")}, // new
		{Price: stripe.String("price_1")}, // already exists
	}

	updateParams := &stripe.SubscriptionUpdateParams{}
	entitlements.AddNewItemsIfNotExist(existingItems, updateParams, newItems...)

	// Only price_2 should be added
	c.Len(updateParams.Items, 1)
	c.Equal("price_2", *updateParams.Items[0].Price)

	expectedSubscription := &stripe.Subscription{ID: "sub_update"}

	stripeBackendMock := new(mocks.MockStripeBackend)
	stripeTestBackends := &stripe.Backends{
		API:     stripeBackendMock,
		Connect: stripeBackendMock,
		Uploads: stripeBackendMock,
	}
	stripeBackendMock.On("Call", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Run(func(args mock.Arguments) {
		mockSubscriptionResult := args.Get(4).(*stripe.Subscription)
		*mockSubscriptionResult = *expectedSubscription
	}).Return(nil)

	mockStripeClient := stripe.NewClient("sk_test", stripe.WithBackends(stripeTestBackends))

	service := entitlements.StripeClient{
		Client: mockStripeClient,
	}

	updatedSub, err := service.UpdateSubscriptionWithOptions(context.Background(), "sub_update", updateParams, entitlements.WithUpdateSubscriptionItems(updateParams.Items...))
	c.NoError(err)
	c.Equal(expectedSubscription, updatedSub)
}

func TestCreateWebhookEndpoint(t *testing.T) {
	c := assert.New(t)

	expectedWebhook := &stripe.WebhookEndpoint{
		ID:     "we_123",
		Secret: "whsec_test",
	}

	stripeBackendMock := new(mocks.MockStripeBackend)
	stripeTestBackends := &stripe.Backends{
		API:     stripeBackendMock,
		Connect: stripeBackendMock,
		Uploads: stripeBackendMock,
	}

	stripeBackendMock.On("Call", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Run(func(args mock.Arguments) {
		result := args.Get(4).(*stripe.WebhookEndpoint)
		*result = *expectedWebhook
	}).Return(nil)

	mockStripeClient := stripe.NewClient("sk_test", stripe.WithBackends(stripeTestBackends))

	service := entitlements.StripeClient{
		Client: mockStripeClient,
	}

	webhook, err := service.CreateWebhookEndpoint(context.Background(), "https://example.com/webhook", entitlements.SupportedEventTypeStrings(), stripe.APIVersion, false)
	c.NoError(err)
	c.Equal(expectedWebhook, webhook)
}

func TestCreateWebhookEndpointDefaultEvents(t *testing.T) {
	c := assert.New(t)

	expectedWebhook := &stripe.WebhookEndpoint{
		ID:     "we_123",
		Secret: "whsec_test",
	}

	stripeBackendMock := new(mocks.MockStripeBackend)
	stripeTestBackends := &stripe.Backends{
		API:     stripeBackendMock,
		Connect: stripeBackendMock,
		Uploads: stripeBackendMock,
	}

	stripeBackendMock.On("Call", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Run(func(args mock.Arguments) {
		result := args.Get(4).(*stripe.WebhookEndpoint)
		*result = *expectedWebhook
	}).Return(nil)

	mockStripeClient := stripe.NewClient("sk_test", stripe.WithBackends(stripeTestBackends))

	service := entitlements.StripeClient{
		Client: mockStripeClient,
	}

	webhook, err := service.CreateWebhookEndpoint(context.Background(), "https://example.com/webhook", nil, "", false)
	c.NoError(err)
	c.Equal(expectedWebhook, webhook)
}
