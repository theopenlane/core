package entitlements

import (
	"bytes"

	"github.com/stretchr/testify/mock"
	"github.com/stripe/stripe-go/v81"
	"github.com/stripe/stripe-go/v81/form"
)

// FakeStripeClient is a fake implementation of the StripeClient interface
type FakeStripeClient struct {
	mock.Mock
}

// CreateAccount mock for CreateAccount method in StripeClient interface
func (t *FakeStripeClient) CreateAccount(accountParams *stripe.AccountParams) (*stripe.Account, error) {
	args := t.Called(accountParams)

	return args.Get(0).(*stripe.Account), args.Error(1)
}

// UpdateAccount mock for UpdateAccount method in StripeClient interface
func (t *FakeStripeClient) UpdateAccount(accountID string, accountParams *stripe.AccountParams) (*stripe.Account, error) {
	args := t.Called(accountID, accountParams)

	return args.Get(0).(*stripe.Account), args.Error(1)
}

// GetAccount mock for GetAccount method in StripeClient interface
func (t *FakeStripeClient) GetAccount(accountID string, accountParams *stripe.AccountParams) (*stripe.Account, error) {
	args := t.Called(accountID, accountParams)

	return args.Get(0).(*stripe.Account), args.Error(1)
}

// CreateCustomer mock for CreateCustomer method in StripeClient interface
func (t *FakeStripeClient) CreateCustomer(customerParams *stripe.CustomerParams) (*stripe.Customer, error) {
	args := t.Called(customerParams)

	return args.Get(0).(*stripe.Customer), args.Error(1)
}

// UpdateCustomer mock for UpdateCustomer method in StripeClient interface
func (t *FakeStripeClient) UpdateCustomer(customerID string, customerParams *stripe.CustomerParams) (*stripe.Customer, error) {
	args := t.Called(customerID, customerParams)

	return args.Get(0).(*stripe.Customer), args.Error(1)
}

// CreateCheckoutSession mock for CreateCheckoutSession method in StripeClient interface
func (t *FakeStripeClient) CreateCheckoutSession(checkoutSessionParams *stripe.CheckoutSessionParams) (*stripe.CheckoutSession, error) {
	args := t.Called(checkoutSessionParams)

	return args.Get(0).(*stripe.CheckoutSession), args.Error(1)
}

// GetSubscription mock for GetSubscription method in StripeClient interface
func (t *FakeStripeClient) GetSubscription(subscriptionID string, params *stripe.SubscriptionParams) (*stripe.Subscription, error) {
	args := t.Called(subscriptionID, params)

	return args.Get(0).(*stripe.Subscription), args.Error(1)
}

// CreateSubscription mock for CreateSubscription method in StripeClient interface
func (t *FakeStripeClient) CreateSubscription(subscriptionParams *stripe.SubscriptionParams) (*stripe.Subscription, error) {
	args := t.Called(subscriptionParams)

	return args.Get(0).(*stripe.Subscription), args.Error(1)
}

// UpdateSubscription mock for UpdateSubscription method in StripeClient interface
func (t *FakeStripeClient) UpdateSubscription(subscriptionID string, subscriptionParams *stripe.SubscriptionParams) (*stripe.Subscription, error) {
	args := t.Called(subscriptionID, subscriptionParams)

	return args.Get(0).(*stripe.Subscription), args.Error(1)
}

// CancelSubscription mock for CancelSubscription method in StripeClient interface
func (t *FakeStripeClient) CancelSubscription(subscriptionID string, cancelParams *stripe.SubscriptionCancelParams) (*stripe.Subscription, error) {
	args := t.Called(subscriptionID, cancelParams)

	return args.Get(0).(*stripe.Subscription), args.Error(1)
}

// CreateBillingPortalSession mock for CreateBillingPortalSession method in StripeClient interface
func (t *FakeStripeClient) CreateBillingPortalSession(billingPortalSParams *stripe.BillingPortalSessionParams) (*stripe.BillingPortalSession, error) {
	args := t.Called(billingPortalSParams)

	return args.Get(0).(*stripe.BillingPortalSession), args.Error(1)
}

// GetPaymentIntent mock for GetPaymentIntent method in StripeClient interface
func (t *FakeStripeClient) GetPaymentIntent(paymentIntentID string, paymentIntentParams *stripe.PaymentIntentParams) (*stripe.PaymentIntent, error) {
	args := t.Called(paymentIntentID, paymentIntentParams)

	return args.Get(0).(*stripe.PaymentIntent), args.Error(1)
}

// ConstructWebhookEvent mock for ConstructWebhookEvent method in StripeClient interface
func (t *FakeStripeClient) ConstructWebhookEvent(reqBody []byte, signature string, webhookKey string) (stripe.Event, error) {
	return t.Called(reqBody, signature, webhookKey).Get(0).(stripe.Event), t.Called(reqBody, signature, webhookKey).Error(1)
}

// NewCheckoutSession mock for NewCheckoutSession method in StripeClient interface
func (t *FakeStripeClient) NewCheckoutSession(
	paymentMethods []string,
	mode, successURL, cancelURL string,
	lineItems []*stripe.CheckoutSessionLineItemParams,
	_ []*stripe.CheckoutSessionDiscountParams) *stripe.CheckoutSession {
	return t.Called(paymentMethods, mode, successURL, cancelURL, lineItems).Get(0).(*stripe.CheckoutSession)
}

// GetActiveEntitlements mock for GetActiveEntitlements method in StripeClient interface
func (t *FakeStripeClient) RetrieveActiveEntitlements(customerID string) ([]string, error) {
	args := t.Called(customerID)

	return args.Get(0).([]string), args.Error(1)
}

// EntitlementsActiveEntitlements mock for EntitlementsActiveEntitlements method in StripeClient interface
func (t *FakeStripeClient) EntitlementsActiveEntitlements(customerID string) ([]string, error) {
	args := t.Called(customerID)

	return args.Get(0).([]string), args.Error(1)
}

// MockEntitlementsActiveEntitlementIter is a mock of the stripe.EntitlementsActiveEntitlementIter
type MockEntitlementsActiveEntitlementIter struct {
	mock.Mock
}

// Next is a mock implementation of the Next method
func (t *MockEntitlementsActiveEntitlementIter) Next() bool {
	args := t.Called()
	return args.Bool(0)
}

// EntitlementsActiveEntitlement is a mock implementation of the EntitlementsActiveEntitlement method
func (t *MockEntitlementsActiveEntitlementIter) EntitlementsActiveEntitlement() *stripe.EntitlementsActiveEntitlement {
	args := t.Called()
	return args.Get(0).(*stripe.EntitlementsActiveEntitlement)
}

// Err is a mock implementation of the Err method
func (t *MockEntitlementsActiveEntitlementIter) Err() error {
	args := t.Called()
	return args.Error(0)
}

// List is a mock implementation of the List method
func (t *FakeStripeClient) List(params *stripe.EntitlementsActiveEntitlementListParams) *MockEntitlementsActiveEntitlementIter {
	args := t.Called(params)
	return args.Get(0).(*MockEntitlementsActiveEntitlementIter)
}

// MockStripeBackend mock for Stripe Backend interface
type MockStripeBackend struct {
	mock.Mock
}

// Call mock for Call method in Stripe Backend interface
func (m *MockStripeBackend) Call(method, path, key string, params stripe.ParamsContainer, v stripe.LastResponseSetter) error {
	args := m.Called(method, path, key, params, v)

	return args.Error(0)
}

// CallStreaming mock for Call method in Stripe Backend interface
func (m *MockStripeBackend) CallStreaming(method, path, key string, params stripe.ParamsContainer, v stripe.StreamingLastResponseSetter) error {
	args := m.Called(method, path, key, params, v)

	return args.Error(0)
}

// CallRaw mock for Call method in Stripe Backend interface
func (m *MockStripeBackend) CallRaw(method, path, key string, body *form.Values, params *stripe.Params, v stripe.LastResponseSetter) error {
	args := m.Called(method, path, key, body, params, v)

	return args.Error(0)
}

// CallMultipart mock for Call method in Stripe Backend interface
func (m *MockStripeBackend) CallMultipart(method, path, key, boundary string, body *bytes.Buffer, params *stripe.Params, v stripe.LastResponseSetter) error {
	args := m.Called(method, path, key, boundary, body, params, v)

	return args.Error(0)
}

// SetMaxNetworkRetries mock for SetMaxNetworkRetries method in Stripe Backend interface
func (m *MockStripeBackend) SetMaxNetworkRetries(maxNetworkRetries int64) {
	m.Called(maxNetworkRetries)
}

func (m *FakeStripeClient) CreateTrialSubscription(cust *stripe.Customer) (*Subscription, error) {
	args := m.Called(cust)
	return args.Get(0).(*Subscription), args.Error(1)
}

func (m *FakeStripeClient) MapStripeSubscription(subs *stripe.Subscription) *Subscription {
	args := m.Called(subs)
	return args.Get(0).(*Subscription)
}
