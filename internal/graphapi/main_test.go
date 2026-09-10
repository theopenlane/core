package graphapi_test

import (
	"context"
	"flag"
	"os"
	"testing"

	th "github.com/theopenlane/core/v2/internal/graphapi/testharness"
)

// graphTestSuite embeds the shared harness so this package can hang its own helper
// methods off suite while the harness itself lives in testharness
type graphTestSuite struct {
	*th.GraphTestSuite
}

var suite = &graphTestSuite{th.Suite}

func TestMain(m *testing.M) {
	flag.Parse()

	// Create a new testing.T instance
	// Note: this is only to seed data; you should not use this instance for actual tests
	// this also cannot be used with a t.FailNow(), you must os.Exit when using this t
	t := &testing.T{}

	suite.SetupSuite(t)
	suite.SetupTestData(context.Background(), t)

	exitCode := m.Run()

	suite.TearDownSuite(t)

	os.Exit(exitCode)
}
