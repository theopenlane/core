//go:build test

package testutils

import (
	"os"
	"testing"

	"github.com/rs/zerolog/log"
)

// RequireNoError exits the test binary when setup or teardown outside of a test fails
func RequireNoError(t *testing.T, err error) {
	t.Helper()
	if err != nil {
		log.Error().Err(err).Msg("fatal error during test setup or teardown")

		os.Exit(1)
	}
}

// FailNow logs the messages and exits the test binary from setup or teardown outside of a test
func FailNow(t *testing.T, msgs ...string) {
	t.Helper()
	logMsg := log.Error()

	for _, m := range msgs {
		logMsg.Str("msg", m)
	}

	logMsg.Msg("fatal error during test setup or teardown")

	os.Exit(1)
}
