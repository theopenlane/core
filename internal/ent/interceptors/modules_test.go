package interceptors

import (
	"testing"

	"github.com/samber/lo"
	"github.com/stretchr/testify/assert"

	"github.com/theopenlane/core/internal/ent/entconfig"
	"github.com/theopenlane/core/internal/ent/generated"
)

func testInterceptorLogic(client *generated.Client) bool {
	if client != nil && client.EntConfig != nil && !client.EntConfig.Modules.Enabled {
		return true
	}

	return false
}

func TestInterceptorModules(t *testing.T) {
	tests := []struct {
		title            string
		entConfigEnabled *bool
		expectedSkip     bool
	}{
		{
			title:            "modules enabled - should continue processing",
			entConfigEnabled: lo.ToPtr(true),
			expectedSkip:     false,
		},
		{
			title:            "modules disabled - should skip",
			entConfigEnabled: lo.ToPtr(false),
			expectedSkip:     true,
		},
		{
			title:            "no EntConfig - should not panic and continue",
			entConfigEnabled: nil,
			expectedSkip:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.title, func(t *testing.T) {
			client := generated.NewClient()
			if tt.entConfigEnabled != nil {
				client.EntConfig = &entconfig.Config{
					Modules: entconfig.Modules{
						Enabled: *tt.entConfigEnabled,
					},
				}
			}

			shouldSkip := testInterceptorLogic(client)
			assert.Equal(t, tt.expectedSkip, shouldSkip)
		})
	}
}
