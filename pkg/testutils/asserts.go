package testutils

import (
	"io"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

// AssertSuccessNotCalled is a success http.Handler that fails if called
func AssertSuccessNotCalled(t *testing.T) http.Handler {
	funk := func(w http.ResponseWriter, req *http.Request) {
		assert.Fail(t, "unexpected call to success handler")
	}

	return http.HandlerFunc(funk)
}

// AssertFailureNotCalled is a failure http.Handler that fails if called
func AssertFailureNotCalled(t *testing.T) http.Handler {
	funk := func(w http.ResponseWriter, req *http.Request) {
		assert.Fail(t, "unexpected call to failure handler")
	}

	return http.HandlerFunc(funk)
}

// AssertBodyString asserts that a Request Body matches the expected string
func AssertBodyString(t *testing.T, rc io.ReadCloser, expected string) {
	defer rc.Close()

	if b, err := io.ReadAll(rc); err == nil {
		if string(b) != expected {
			t.Errorf("expected %q, got %q", expected, string(b))
		}
	} else {
		t.Errorf("error reading body")
	}
}
