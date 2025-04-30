package entitlements_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"github.com/stripe/stripe-go/v82"
	"github.com/stripe/stripe-go/v82/client"

	"github.com/theopenlane/core/pkg/entitlements"
	"github.com/theopenlane/core/pkg/entitlements/mocks"
)

func TestNew(t *testing.T) {
	c := require.New(t)

	t.Setenv("STRIPE_SECRET_KEY", "secret_key")

	stripeService, err := entitlements.NewStripeClient(entitlements.WithAPIKey("secret_key"))
	c.NoError(err)
	c.IsType(stripeService, stripeService)
}

func TestNewErrMissingAPIKey(t *testing.T) {
	c := require.New(t)

	stripeService, err := entitlements.NewStripeClient()
	c.Nil(stripeService)
	c.ErrorIs(err, entitlements.ErrMissingAPIKey)
}

func TestWithConfig(t *testing.T) {
	config := entitlements.Config{
		PublicStripeKey:  "public_key",
		PrivateStripeKey: "private_key",
	}

	option := entitlements.WithConfig(config)
	client := &entitlements.StripeClient{}

	option(client)

	if client.Config.PublicStripeKey != config.PublicStripeKey ||
		client.Config.PrivateStripeKey != config.PrivateStripeKey {
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
				entitlements.WithPublicStripeKey("public_key"),
				entitlements.WithPrivateStripeKey("private_key"),
				entitlements.WithStripeWebhookSecret("webhook_secret"),
				entitlements.WithTrialSubscriptionPriceID("trial_price_id"),
				entitlements.WithPersonalOrgSubscriptionPriceID("personal_price_id"),
				entitlements.WithStripeWebhookURL("https://custom.webhook.url"),
				entitlements.WithStripeBillingPortalSuccessURL("https://custom.billing.success.url"),
				entitlements.WithStripeCancellationReturnURL("https://custom.cancellation.return.url"),
			},
			want: &entitlements.Config{
				Enabled:                        true,
				PublicStripeKey:                "public_key",
				PrivateStripeKey:               "private_key",
				StripeWebhookSecret:            "webhook_secret",
				TrialSubscriptionPriceID:       "trial_price_id",
				PersonalOrgSubscriptionPriceID: "personal_price_id",
				StripeWebhookURL:               "https://custom.webhook.url",
				StripeBillingPortalSuccessURL:  "https://custom.billing.success.url",
				StripeCancellationReturnURL:    "https://custom.cancellation.return.url",
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
	c := require.New(t)

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

	mockStripeClient := client.New("sk_test", stripeTestBackends)

	service := entitlements.StripeClient{
		Client: mockStripeClient,
	}

	customer, err := service.CreateCustomer(orgCustomer)
	c.NoError(err)
	c.Equal(expectedCustomer, customer)
}

func TestUpdateCustomer(t *testing.T) {
	c := require.New(t)

	expectedCustomer := &stripe.Customer{
		ID:    "cus_123",
		Email: "updated@example.com",
		Phone: "0987654321",
	}

	updateParams := &stripe.CustomerParams{
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

	customer, err := service.UpdateCustomer("cus_123", updateParams)
	c.NoError(err)
	c.Equal(expectedCustomer, customer)
}

func TestDeleteCustomer(t *testing.T) {
	c := require.New(t)

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

	err = service.DeleteCustomer("cus_123")
	c.NoError(err)
}

func TestCreateSubscription(t *testing.T) {
	c := require.New(t)

	expectedSubscription := &stripe.Subscription{
		ID: "sub_123",
	}

	subscriptionParams := &stripe.SubscriptionParams{
		Customer: stripe.String("cus_123"),
		Items: []*stripe.SubscriptionItemsParams{
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

	mockStripeClient := client.New("sk_test", stripeTestBackends)

	service := entitlements.StripeClient{
		Client: mockStripeClient,
	}

	subscription, err := service.CreateSubscription(subscriptionParams)
	c.NoError(err)
	c.Equal(expectedSubscription, subscription)
}

func TestUpdateSubscription(t *testing.T) {
	c := require.New(t)

	expectedSubscription := &stripe.Subscription{
		ID: "sub_123",
	}

	updateParams := &stripe.SubscriptionParams{
		Items: []*stripe.SubscriptionItemsParams{
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

	mockStripeClient := client.New("sk_test", stripeTestBackends)

	service := entitlements.StripeClient{
		Client: mockStripeClient,
	}

	subscription, err := service.UpdateSubscription("sub_123", updateParams)
	c.NoError(err)
	c.Equal(expectedSubscription, subscription)
}

func TestCancelSubscription(t *testing.T) {
	c := require.New(t)

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

	mockStripeClient := client.New("sk_test", stripeTestBackends)

	service := entitlements.StripeClient{
		Client: mockStripeClient,
	}

	subscription, err := service.CancelSubscription("sub_123", cancelParams)
	c.NoError(err)
	c.Equal(expectedSubscription, subscription)
}

func TestMapStripeSubscription(t *testing.T) {
	c := require.New(t)

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
		TrialEnd:         1620000000,
		ProductID:        "prod_123",
		Status:           "active",
		StripeCustomerID: "cus_123",
		OrganizationID:   "org_123",
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

	mockStripeClient := client.New("sk_test", stripeTestBackends)

	service := entitlements.StripeClient{
		Client: mockStripeClient,
	}

	subscription := service.MapStripeSubscription(stripeSubscription)
	c.Equal(expectedSubscription, subscription)
}
