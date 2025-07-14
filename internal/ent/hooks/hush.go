package hooks

import (
	"context"
	"encoding/base64"
	"encoding/hex"
	"fmt"

	"entgo.io/ent"
	"gocloud.dev/secrets"

	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/generated/hook"
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

				// Encrypt the secret value
				if m.Secrets != nil {
					encrypted, err := m.Secrets.Encrypt(ctx, []byte(v))
					if err != nil {
						return nil, fmt.Errorf("failed to encrypt secret value: %w", err)
					}

					// Store as base64-encoded encrypted value
					encodedValue := base64.StdEncoding.EncodeToString(encrypted)
					m.SetSecretValue(encodedValue)
				} else {
					// Fallback to AES encryption
					key := GetEncryptionKey()
					encrypted, err := EncryptAESHelper([]byte(v), key)
					if err != nil {
						return nil, fmt.Errorf("failed to encrypt secret value with AES: %w", err)
					}

					encodedValue := base64.StdEncoding.EncodeToString(encrypted)
					m.SetSecretValue(encodedValue)
				}

				// Proceed with mutation
				result, err := next.Mutate(ctx, m)
				if err != nil {
					return nil, err
				}

				// Decrypt the result for immediate use
				if hush, ok := result.(*generated.Hush); ok {
					err = Decrypt(ctx, m.Secrets, hush)
				}

				return result, err
			})
		},
		hook.HasFields("secret_value"),
	)
}

// Decrypt decrypts the secret value
func Decrypt(ctx context.Context, k *secrets.Keeper, u *generated.Hush) error {
	if u.SecretValue == "" {
		return nil
	}

	// Decode from base64
	encrypted, err := base64.StdEncoding.DecodeString(u.SecretValue)
	if err != nil {
		// Try hex decoding for backward compatibility
		encrypted, err = hex.DecodeString(u.SecretValue)
		if err != nil {
			return fmt.Errorf("failed to decode encrypted value: %w", err)
		}
	}

	if k != nil {
		// Use secrets keeper for decryption
		plain, err := k.Decrypt(ctx, encrypted)
		if err != nil {
			return fmt.Errorf("failed to decrypt with secrets keeper: %w", err)
		}
		u.SecretValue = string(plain)
	} else {
		// Fallback to AES decryption
		key := GetEncryptionKey()
		plain, err := DecryptAESHelper(encrypted, key)
		if err != nil {
			return fmt.Errorf("failed to decrypt with AES: %w", err)
		}
		u.SecretValue = string(plain)
	}

	return nil
}
