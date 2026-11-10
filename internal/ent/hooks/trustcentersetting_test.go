package hooks

import (
	"context"
	"testing"

	"entgo.io/ent"
	"gotest.tools/v3/assert"

	"github.com/theopenlane/core/v2/internal/ent/generated"
)

func TestHookTrustCenterSettingSanitize(t *testing.T) {
	tests := []struct {
		name     string
		overview string
		expected string
	}{
		{
			name:     "plain text passes through",
			overview: "Just text",
			expected: "Just text",
		},
		{
			name:     "formatting markup is kept",
			overview: "<p>Hello <strong>world</strong></p>",
			expected: "<p>Hello <strong>world</strong></p>",
		},
		{
			name:     "script element and its content are removed",
			overview: "<p>Hello</p><script>alert('xss')</script>",
			expected: "<p>Hello</p>",
		},
		{
			name:     "event handler attributes are removed",
			overview: `<div onclick="steal()">text</div>`,
			expected: "<div>text</div>",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			mutations := map[string]*generated.TrustCenterSettingMutation{
				"create": generated.NewClient().TrustCenterSetting.Create().SetOverview(tc.overview).Mutation(),
				"update": generated.NewClient().TrustCenterSetting.UpdateOneID("setting-id").SetOverview(tc.overview).Mutation(),
			}

			for op, mutation := range mutations {
				t.Run(op, func(t *testing.T) {
					var got string

					next := ent.MutateFunc(func(_ context.Context, m ent.Mutation) (ent.Value, error) {
						got, _ = m.(*generated.TrustCenterSettingMutation).Overview()

						return nil, nil
					})

					_, err := HookTrustCenterSettingSanitize()(next).Mutate(context.Background(), mutation)
					assert.NilError(t, err)
					assert.Equal(t, tc.expected, got)
				})
			}
		})
	}

	t.Run("unset overview is left untouched", func(t *testing.T) {
		mutation := generated.NewClient().TrustCenterSetting.UpdateOneID("setting-id").SetTitle("title").Mutation()

		next := ent.MutateFunc(func(_ context.Context, m ent.Mutation) (ent.Value, error) {
			_, ok := m.(*generated.TrustCenterSettingMutation).Overview()
			assert.Assert(t, !ok)

			return nil, nil
		})

		_, err := HookTrustCenterSettingSanitize()(next).Mutate(context.Background(), mutation)
		assert.NilError(t, err)
	})
}
