package posthog

import "github.com/posthog/posthog-go"

var _ posthog.Logger = (*noopLogger)(nil)

type noopLogger struct{}

func (n *noopLogger) Logf(_ string, _ ...interface{}) {
	// noop logger noop's the logs
}
func (n noopLogger) Errorf(_ string, _ ...interface{}) {
	// noop logger noop's the logs
}
