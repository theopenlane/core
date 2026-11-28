package mocks

import (
	"bytes"

	"github.com/stretchr/testify/mock"
	"github.com/stripe/stripe-go/v87"
)

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
// v82 introduced breaking changes and modified the behavior of CallRaw https://github.com/stripe/stripe-go/releases/tag/v82.0.0
func (m *MockStripeBackend) CallRaw(method, path, key string, body []byte, params *stripe.Params, v stripe.LastResponseSetter) error {
	args := m.Called(method, path, key, params, v)

	return args.Error(0)
}

// CallMultipart mock for Call method in Stripe Backend interface
func (m *MockStripeBackend) CallMultipart(method, path, key, boundary string, body *bytes.Buffer, params *stripe.Params, v stripe.LastResponseSetter) error {
	args := m.Called(method, path, key, params, v)

	return args.Error(0)
}

// CallMultipart mock for Call method in Stripe Backend interface
func (m *MockStripeBackend) SetMaxNetworkRetries(maxNetworkRetries int64) {
	m.Called(maxNetworkRetries)
}
