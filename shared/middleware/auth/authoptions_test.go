package auth_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/theopenlane/iam/tokens"
	"github.com/theopenlane/shared/middleware/auth"
)

func TestDefaultAuthOptions(t *testing.T) {
	// Should be able to create a default auth options with no extra input.
	conf := auth.NewAuthOptions()
	require.NotZero(t, conf, "a zero valued configuration was returned")
	require.Equal(t, auth.DefaultAuthOptions.KeysURL, conf.KeysURL)
	require.Equal(t, auth.DefaultAuthOptions.Audience, conf.Audience)
	require.Equal(t, auth.DefaultAuthOptions.Issuer, conf.Issuer)
	require.Equal(t, auth.DefaultAuthOptions.MinRefreshInterval, conf.MinRefreshInterval)
	require.NotNil(t, conf.Context, "no context was created")
}

func TestAuthOptions(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	conf := auth.NewAuthOptions(
		auth.WithJWKSEndpoint("http://localhost:8088/.well-known/jwks.json"),
		auth.WithAudience("http://localhost:3000"),
		auth.WithIssuer("http://localhost:8088"),
		auth.WithMinRefreshInterval(67*time.Minute),
		auth.WithContext(ctx),
	)

	cancel()
	require.NotZero(t, conf, "a zero valued configuration was returned")
	require.Equal(t, "http://localhost:8088/.well-known/jwks.json", conf.KeysURL)
	require.Equal(t, "http://localhost:3000", conf.Audience)
	require.Equal(t, "http://localhost:8088", conf.Issuer)
	require.Equal(t, 67*time.Minute, conf.MinRefreshInterval)
	require.ErrorIs(t, conf.Context.Err(), context.Canceled)
}

func TestAuthOptionsOverride(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	opts := auth.Options{
		KeysURL:            "http://localhost:8088/.well-known/jwks.json",
		Audience:           "http://localhost:3000",
		Issuer:             "http://localhost:8088",
		MinRefreshInterval: 42 * time.Minute,
		Context:            ctx,
	}

	conf := auth.NewAuthOptions(
		auth.WithAuthOptions(opts),
	)

	require.NotSame(t, &opts, &conf, "expected a new configuration object to be created")
	assert.Equal(t, opts.KeysURL, conf.KeysURL)
	assert.Equal(t, opts.Audience, conf.Audience)
	assert.Equal(t, opts.Issuer, conf.Issuer)
	assert.Equal(t, opts.MinRefreshInterval, conf.MinRefreshInterval)
	assert.Equal(t, opts.Context, conf.Context)

	// Ensure the context is the same on the configuration
	cancel()
	require.ErrorIs(t, conf.Context.Err(), context.Canceled)
}

func TestAuthOptionsValidator(t *testing.T) {
	validator := &tokens.MockValidator{}
	conf := auth.NewAuthOptions(auth.WithValidator(validator))
	require.NotZero(t, conf, "a zero valued configuration was returned")

	actual, err := conf.Validator()
	require.NoError(t, err, "could not create default validator")
	require.Same(t, validator, actual, "conf did not return the same validator")
}
