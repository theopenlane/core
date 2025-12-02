package hooks

import (
	"context"
	"fmt"

	"entgo.io/ent"

	"github.com/theopenlane/ent/encrypt"
	"github.com/theopenlane/ent/generated"
	"github.com/theopenlane/ent/generated/hook"
)

// HookHush runs on hush create/update mutations to encrypt secret_value
func HookHush() ent.Hook {
	return hook.If(
		func(next ent.Mutator) ent.Mutator {
			return hook.HushFunc(func(ctx context.Context, m *generated.HushMutation) (generated.Value, error) {
				v, ok := m.SecretValue()
				if !ok || v == "" {
					return next.Mutate(ctx, m)
				}

				// Encrypt the secret value using Tink
				encodedValue, err := encrypt.Encrypt([]byte(v))
				if err != nil {
					return nil, fmt.Errorf("failed to encrypt secret value: %w", err)
				}

				m.SetSecretValue(encodedValue)

				// Proceed with mutation
				result, err := next.Mutate(ctx, m)
				if err != nil {
					return nil, err
				}

				// Decrypt the result for immediate use
				if hush, ok := result.(*generated.Hush); ok {
					err = DecryptHush(hush)
				}

				return result, err
			})
		},
		hook.HasFields("secret_value"),
	)
}

// DecryptHush decrypts the secret value in a Hush entity using Tink
func DecryptHush(u *generated.Hush) error {
	if u.SecretValue == "" {
		return nil
	}

	// Decrypt using Tink
	decrypted, err := encrypt.Decrypt(u.SecretValue)
	if err != nil {
		return fmt.Errorf("failed to decrypt secret value: %w", err)
	}

	u.SecretValue = string(decrypted)

	return nil
}
