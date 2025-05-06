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
				secretString, err := GenerateDomainValidationSecret()
				if err != nil {
					return nil, err
				}

				// Hash the secret with bcrypt to store in the DB
				hashedSecret, err := bcrypt.GenerateFromPassword([]byte(secretString), bcrypt.DefaultCost)
				if err != nil {
					return nil, err
				}

				m.SetTxtRecordValue(string(hashedSecret))

				u, err := next.Mutate(ctx, m)

				if err != nil {
					return nil, err
				}

				if u, ok := u.(*generated.CustomDomain); ok {
					// Return the original secretString to the caller, not the hashed value
					// this will be readable only in the Create mutation response, and not retrievable later
					u.TxtRecordValue = secretString
				}

				return u, err
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
