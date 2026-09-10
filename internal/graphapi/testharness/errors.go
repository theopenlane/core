//go:build test

package testharness

import (
	"os"
	"testing"

	"github.com/gqlgo/gqlgenc/clientv2"
	"github.com/rs/zerolog/log"
	"github.com/theopenlane/core/v2/internal/graphapi/testclient"
	"github.com/vektah/gqlparser/v2/gqlerror"
	"gotest.tools/v3/assert"
)

// AssertErrorCode checks if the error code matches the expected code
func AssertErrorCode(t *testing.T, err *gqlerror.Error, code string) {
	t.Helper()

	assert.Equal(t, code, testclient.GetErrorCode(err))
}

// AssertErrorMessage checks if the error message matches the expected message
func AssertErrorMessage(t *testing.T, err *gqlerror.Error, msg string) {
	t.Helper()

	assert.Equal(t, msg, testclient.GetErrorMessage(err))
}

func RequireNoError(t *testing.T, err error) {
	t.Helper()
	if err != nil {
		log.Error().Err(err).Msg("fatal error during test setup or teardown")

		os.Exit(1)
	}
}

func FailNow(t *testing.T, msgs ...string) {
	t.Helper()
	logMsg := log.Error()

	for _, m := range msgs {
		logMsg.Str("msg", m)
	}

	logMsg.Msg("fatal error during test setup or teardown")

	os.Exit(1)
}

// ParseClientError parses the error response from the client and returns a slice of gqlerror.Error
func ParseClientError(t *testing.T, err error) []*gqlerror.Error {
	t.Helper()

	if err == nil {
		return nil
	}

	errResp, ok := err.(*clientv2.ErrorResponse)
	assert.Check(t, ok)
	assert.Check(t, errResp.HasErrors())

	gqlErrors := []*gqlerror.Error{}

	errors := errResp.GqlErrors.Unwrap()

	for _, e := range errors {
		customErr, ok := e.(*gqlerror.Error)
		assert.Check(t, ok)
		gqlErrors = append(gqlErrors, customErr)
	}

	return gqlErrors
}
