package tokens_test

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	jwt "github.com/golang-jwt/jwt/v5"
	"github.com/lestrrat-go/jwx/v2/jwk"

	"github.com/theopenlane/core/pkg/tokens"
)

func (s *TokenTestSuite) TestJWKSValidator() {
	// This is a long running test, skip if in short mode
	if testing.Short() {
		s.T().Skip("skipping long running test in short mode")
	}

	// NOTE: this test requires the jwks.json fixture to use the same keys as the
	// testdata keys loaded from the PEM file fixtures.
	// Create access and refresh tokens to validate.
	require := s.Require()
	tm, err := tokens.New(s.conf)
	require.NoError(err, "could not initialize token manager")

	claims := &tokens.Claims{
		RegisteredClaims: jwt.RegisteredClaims{
			Subject: "01H6PGFB4T34D4WWEXQMAGJNMK",
		},
		UserID: "Rusty Shackleford",
		OrgID:  "01H6PGFG71N0AFEVTK3NJB71T9",
	}

	atks, rtks, err := tm.CreateTokenPair(claims)
	require.NoError(err, "could not create token pair")
	time.Sleep(500 * time.Millisecond)

	// Create a validator from a JWKS key set
	jwks, err := jwk.ReadFile("testdata/jwks.json")
	require.NoError(err, "could not read jwks from file")

	validator := tokens.NewJWKSValidator(jwks, "http://localhost:3000", "http://localhost:3001")

	parsedClaims, err := validator.Verify(atks)
	require.NoError(err, "could not validate access token")
	require.Equal(claims, parsedClaims, "parsed claims not returned correctly")

	_, err = validator.Parse(rtks)
	require.NoError(err, "could not parse refresh token")
}

func (s *TokenTestSuite) TestCachedJWKSValidator() {
	// This is a long running test, skip if in short mode
	if testing.Short() {
		s.T().Skip("skipping long running test in short mode")
	}

	// Create a test server that initially serves the partial_jwks.json file then
	// serves the jwks.json file from then on out.
	requests := 0
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var (
			err  error
			path string
			f    *os.File
		)

		if requests == 0 {
			path = "testdata/partial_jwks.json"
		} else {
			path = "testdata/jwks.json"
		}

		if f, err = os.Open(path); err != nil {
			w.Header().Add("Content-Type", "text/plain")
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(err.Error())) // nolint: errcheck

			return
		}

		requests++

		w.Header().Add("Content-Type", "application/json")
		io.Copy(w, f) // nolint: errcheck
	}))

	defer srv.Close()

	// NOTE: this test requires the jwks.json fixture to use the same keys as the
	// testdata keys loaded from the PEM file fixtures.
	// Create access and refresh tokens to validate.
	require := s.Require()
	tm, err := tokens.New(s.conf)
	require.NoError(err, "could not initialize token manager")

	claims := &tokens.Claims{
		RegisteredClaims: jwt.RegisteredClaims{
			Subject: "01H6PGFB4T34D4WWEXQMAGJNMK",
		},
		UserID: "Rusty Shackleford",
		OrgID:  "01H6PGFG71N0AFEVTK3NJB71T9",
	}

	atks, _, err := tm.CreateTokenPair(claims)
	require.NoError(err, "could not create token pair")
	time.Sleep(500 * time.Millisecond)

	// Create a new cached validator for testing
	cache := jwk.NewCache(context.Background())
	cache.Register(srv.URL, jwk.WithMinRefreshInterval(1*time.Minute)) // nolint: errcheck
	validator, err := tokens.NewCachedJWKSValidator(context.Background(), cache, srv.URL, "http://localhost:3000", "http://localhost:3001")
	require.NoError(err, "could not create new cached JWKS validator")

	// The first attempt to validate the access token should fail since the
	// partial_jwks.json fixture does not have the keys that signed the token.
	_, err = validator.Verify(atks)
	require.EqualError(err, "token is unverifiable: error while executing keyfunc: unknown signing key")

	// After refreshing the cache, the access token should be able to be verified.
	err = validator.Refresh(context.Background())
	require.NoError(err, "could not refresh cache")

	actualClaims, err := validator.Verify(atks)
	require.NoError(err, "should have been able to verify the access token")
	require.Equal(claims, actualClaims, "expected the correct claims to be returned")
}
