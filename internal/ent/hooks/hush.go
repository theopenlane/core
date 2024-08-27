package hooks

import (
	"context"
	"encoding/hex"
	"fmt"

	"entgo.io/ent"
	"gocloud.dev/secrets"

	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/generated/hook"
)

// HookHush runs on invite create mutations
func HookHush() ent.Hook {
	return hook.If(
		func(next ent.Mutator) ent.Mutator {
			return hook.HushFunc(func(ctx context.Context, m *generated.HushMutation) (generated.Value, error) {
				v, ok := m.SecretValue()
				if !ok || v == "" {
					return nil, fmt.Errorf("unexpected 'secret_name' value") // nolint:err113
				}

				c, err := m.Secrets.Encrypt(ctx, []byte(v))
				if err != nil {
					return nil, err
				}

				m.SetName(hex.EncodeToString(c))
				u, err := next.Mutate(ctx, m)

				if err != nil {
					return nil, err
				}

				if u, ok := u.(*generated.Hush); ok {
					err = Decrypt(ctx, m.Secrets, u)
				}

				return u, err
			})
		},
		hook.HasFields("secret_value"),
	)
}

// Decrypt decrypts the secret value
func Decrypt(ctx context.Context, k *secrets.Keeper, u *generated.Hush) error {
	b, err := hex.DecodeString(u.SecretValue)
	if err != nil {
		return err
	}

	plain, err := k.Decrypt(ctx, b)
	if err != nil {
		return err
	}

	u.Name = string(plain)

	return nil
}
