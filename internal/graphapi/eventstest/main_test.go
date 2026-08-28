package eventstest_test

import (
	"context"
	"flag"
	"os"
	"testing"

	th "github.com/theopenlane/core/v2/internal/graphapi/testharness"
)

var suite = th.Suite

func TestMain(m *testing.M) {
	flag.Parse()

	t := &testing.T{}

	suite.SetupSuite(t)
	suite.SetupTestData(context.Background(), t)

	exitCode := m.Run()

	suite.TearDownSuite(t)

	os.Exit(exitCode)
}
