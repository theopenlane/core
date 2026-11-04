package auth_test

import (
	"context"
	"reflect"
	"testing"
	"time"

	"gotest.tools/v3/assert"
	is "gotest.tools/v3/assert/cmp"

	"github.com/theopenlane/iam/tokens"

	"github.com/theopenlane/core/v2/pkg/middleware/auth"
)

func TestDefaultAuthOptions(t *testing.T) {
	// Should be able to create a default auth options with no extra input.
	conf := auth.NewAuthOptions()
	assert.Assert(t, !reflect.ValueOf(conf).IsZero(), "a zero valued configuration was returned")
	assert.Assert(t, is.Equal(auth.DefaultAuthOptions.KeysURL, conf.KeysURL))
	assert.Assert(t, is.Equal(auth.DefaultAuthOptions.Audience, conf.Audience))
	assert.Assert(t, is.Equal(auth.DefaultAuthOptions.Issuer, conf.Issuer))
	assert.Assert(t, is.Equal(auth.DefaultAuthOptions.MinRefreshInterval, conf.MinRefreshInterval))
	assert.Assert(t, conf.Context != nil, "no context was created")
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
	assert.Assert(t, !reflect.ValueOf(conf).IsZero(), "a zero valued configuration was returned")
	assert.Assert(t, is.Equal("http://localhost:8088/.well-known/jwks.json", conf.KeysURL))
	assert.Assert(t, is.Equal("http://localhost:3000", conf.Audience))
	assert.Assert(t, is.Equal("http://localhost:8088", conf.Issuer))
	assert.Assert(t, is.Equal(67*time.Minute, conf.MinRefreshInterval))
	assert.ErrorIs(t, conf.Context.Err(), context.Canceled)
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

	assert.Assert(t, &opts != &conf, "expected a new configuration object to be created")
	assert.Check(t, is.Equal(opts.KeysURL, conf.KeysURL))
	assert.Check(t, is.Equal(opts.Audience, conf.Audience))
	assert.Check(t, is.Equal(opts.Issuer, conf.Issuer))
	assert.Check(t, is.Equal(opts.MinRefreshInterval, conf.MinRefreshInterval))
	assert.Check(t, is.Equal(opts.Context, conf.Context))

	// Ensure the context is the same on the configuration
	cancel()
	assert.ErrorIs(t, conf.Context.Err(), context.Canceled)
}

func TestAuthOptionsValidator(t *testing.T) {
	validator := &tokens.MockValidator{}
	conf := auth.NewAuthOptions(auth.WithValidator(validator))
	assert.Assert(t, !reflect.ValueOf(conf).IsZero(), "a zero valued configuration was returned")

	actual, err := conf.Validator()
	assert.NilError(t, err, "could not create default validator")
	assert.Assert(t, is.Equal(validator, actual), "conf did not return the same validator")
}
