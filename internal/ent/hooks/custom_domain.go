package hooks

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"io"

	"entgo.io/ent"
	"golang.org/x/crypto/bcrypt"

	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/generated/hook"
)

const (
	domainValidationSecretLength = 16
)

// HookCustomDomain runs on create mutations
func HookCustomDomain() ent.Hook {
	return hook.If(
		func(next ent.Mutator) ent.Mutator {
			return hook.CustomDomainFunc(func(ctx context.Context, m *generated.CustomDomainMutation) (generated.Value, error) {
				// TODO(acookin): add cloudflare validation
				return next.Mutate(ctx, m)
			})
		},
		hook.HasOp(ent.OpCreate),
	)
}

// GenerateDomainValidationSecret creates a random string of specified length
func GenerateDomainValidationSecret() (string, error) {
	bytes := make([]byte, domainValidationSecretLength)

	if _, err := io.ReadFull(rand.Reader, bytes); err != nil {
		return "", err
	}

	encodedString := base64.URLEncoding.EncodeToString(bytes)[:domainValidationSecretLength]

	return encodedString, nil
}

// VerifyDomainValidationSecret verifies if the provided secretString matches the hashed value
func VerifyDomainValidationSecret(hashedSecret, secretString string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hashedSecret), []byte(secretString))
	return err == nil
}
