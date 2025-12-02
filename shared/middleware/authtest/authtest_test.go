package authtest_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/theopenlane/utils/ulids"

	"github.com/theopenlane/iam/tokens"

	authtest "github.com/theopenlane/shared/middleware/authtest"
)

// This test generates an example token with fake RSA keys for use in examples,
// documentation and other tests that don't need a valid token (since it will expire).
func TestGenerateToken(t *testing.T) {
	// The `t.Skip` function is used to skip the execution of a test case. In this case, the test case is
	// skipped with the message "comment the skip out if you want to generate a token". This is done to
	// prevent the generation of a token during normal test runs, as generating a token may not be
	// necessary for every test execution.
	t.Skip("comment the skip out if you want to generate a token")

	srv, err := authtest.NewServer()
	require.NoError(t, err, "could not start authtest server")

	defer srv.Close()

	userID := ulids.New().String()
	claims := &tokens.Claims{
		UserID: userID,
		OrgID:  "01H6PGFG71N0AFEVTK3NJB71T9",
	}

	accessToken, refreshToken, err := srv.CreateTokenPair(claims)
	require.NoError(t, err, "could not generate access token")

	// Log the tokens then fail the test so the tokens are printed out.
	t.Logf("access token: %s", accessToken)
	t.Logf("refresh token: %s", refreshToken)
	t.FailNow()
}
