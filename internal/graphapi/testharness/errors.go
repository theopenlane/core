//go:build test

package testharness

import (
	"testing"

	"github.com/gqlgo/gqlgenc/clientv2"
	"github.com/theopenlane/core/v2/internal/graphapi/testclient"
	coreutils "github.com/theopenlane/core/v2/internal/testutils"
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

var (
	// RequireNoError exits the test binary when setup or teardown outside of a test fails
	RequireNoError = coreutils.RequireNoError
	// FailNow logs the messages and exits the test binary from setup or teardown outside of a test
	FailNow = coreutils.FailNow
)

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
