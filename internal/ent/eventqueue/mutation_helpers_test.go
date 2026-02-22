package eventqueue

import (
	"testing"

	"github.com/samber/do/v2"
	"github.com/stretchr/testify/require"

	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/pkg/gala"
)

// TestMutationStringValueOrProperty verifies payload-first string fallback semantics
func TestMutationStringValueOrProperty(t *testing.T) {
	t.Parallel()

	const field = "email"

	t.Run("prefers proposed value", func(t *testing.T) {
		payload := MutationGalaPayload{
			ProposedChanges: map[string]any{
				field: " proposed@example.com ",
			},
		}

		got := MutationStringValueOrProperty(payload, map[string]string{field: "header@example.com"}, field)
		require.Equal(t, "proposed@example.com", got)
	})

	t.Run("falls back to properties when proposed value is missing", func(t *testing.T) {
		got := MutationStringValueOrProperty(MutationGalaPayload{}, map[string]string{field: "header@example.com"}, field)
		require.Equal(t, "header@example.com", got)
	})

	t.Run("falls back to properties when proposed value is blank", func(t *testing.T) {
		payload := MutationGalaPayload{
			ProposedChanges: map[string]any{
				field: "   ",
			},
		}

		got := MutationStringValueOrProperty(payload, map[string]string{field: "header@example.com"}, field)
		require.Equal(t, "header@example.com", got)
	})
}

// TestMutationStringValuePreferPayload verifies strict payload precedence semantics
func TestMutationStringValuePreferPayload(t *testing.T) {
	t.Parallel()

	const field = "email"

	t.Run("prefers proposed value when present", func(t *testing.T) {
		payload := MutationGalaPayload{
			ProposedChanges: map[string]any{
				field: " proposed@example.com ",
			},
		}

		got := MutationStringValuePreferPayload(payload, map[string]string{field: "header@example.com"}, field)
		require.Equal(t, "proposed@example.com", got)
	})

	t.Run("falls back to properties when proposed value is missing", func(t *testing.T) {
		got := MutationStringValuePreferPayload(MutationGalaPayload{}, map[string]string{field: "header@example.com"}, field)
		require.Equal(t, "header@example.com", got)
	})

	t.Run("does not fall back when proposed value is blank", func(t *testing.T) {
		payload := MutationGalaPayload{
			ProposedChanges: map[string]any{
				field: "   ",
			},
		}

		got := MutationStringValuePreferPayload(payload, map[string]string{field: "header@example.com"}, field)
		require.Empty(t, got)
	})

	t.Run("preserves non-string proposed value conversion semantics", func(t *testing.T) {
		payload := MutationGalaPayload{
			ProposedChanges: map[string]any{
				field: []any{"invalid"},
			},
		}

		got := MutationStringValuePreferPayload(payload, map[string]string{field: "header@example.com"}, field)
		require.Equal(t, "[invalid]", got)
	})
}

// TestClientFromHandler verifies client resolution from handler injectors
func TestClientFromHandler(t *testing.T) {
	t.Parallel()

	t.Run("returns injected client", func(t *testing.T) {
		injector := do.New()
		client := &generated.Client{}
		do.ProvideValue(injector, client)

		got, ok := ClientFromHandler(gala.HandlerContext{Injector: injector})
		require.True(t, ok)
		require.Same(t, client, got)
	})

	t.Run("returns false without injected client", func(t *testing.T) {
		got, ok := ClientFromHandler(gala.HandlerContext{Injector: do.New()})
		require.False(t, ok)
		require.Nil(t, got)
	})
}
