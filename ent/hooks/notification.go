package hooks

import (
	"context"
	"fmt"

	"entgo.io/ent"

	"github.com/theopenlane/core/pkg/enums"
	"github.com/theopenlane/ent/generated"
	"github.com/theopenlane/ent/generated/hook"
)

// HookNotification runs on notification mutations to validate channels
func HookNotification() ent.Hook {
	return hook.If(func(next ent.Mutator) ent.Mutator {
		return hook.NotificationFunc(func(ctx context.Context, m *generated.NotificationMutation) (generated.Value, error) {
			// Validate channels using m.Channels()
			if channels, ok := m.Channels(); ok {
				if err := isValidChannels(channels); err != nil {
					return nil, err
				}
			}

			// Validate appended channels using m.AppendedChannels()
			if appendedChannels, ok := m.AppendedChannels(); ok {
				if err := isValidChannels(appendedChannels); err != nil {
					return nil, err
				}
			}

			return next.Mutate(ctx, m)
		})
	},
		hook.And(
			hook.Or(
				hook.HasFields("channels"),
				hook.HasAddedFields("channels"),
			),
			hook.HasOp(ent.OpCreate|ent.OpUpdateOne|ent.OpUpdate),
		),
	)
}

// isValidChannels validates that all channels in the slice are valid
func isValidChannels(channels []enums.Channel) error {
	if len(channels) == 0 {
		return nil
	}

	validChannels := enums.Channel("").Values()
	validMap := make(map[string]bool)
	for _, v := range validChannels {
		validMap[v] = true
	}

	for _, ch := range channels {
		if !validMap[string(ch)] {
			return fmt.Errorf("%w: %s", ErrInvalidChannel, ch)
		}
	}

	return nil
}
