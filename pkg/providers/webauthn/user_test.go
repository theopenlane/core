package webauthn_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/go-webauthn/webauthn/protocol"
	gowebauthn "github.com/go-webauthn/webauthn/webauthn"

	"github.com/theopenlane/core/pkg/providers/webauthn"
)

func TestUserWebAuthnID(t *testing.T) {
	// Create a user instance
	user := &webauthn.User{
		ID: "exampleID",
	}

	// Call the WebAuthnID method
	webAuthnID := user.WebAuthnID()

	// Check if the returned value is correct
	expectedWebAuthnID := []byte("exampleID")
	assert.Equal(t, expectedWebAuthnID, webAuthnID)
}

func TestUserWebAuthnName(t *testing.T) {
	// Create a user instance
	user := &webauthn.User{
		Name: "example",
	}

	// Call the WebAuthnID method
	webAuthnName := user.WebAuthnName()

	// Check if the returned value is correct
	assert.Equal(t, "example", webAuthnName)
}

func TestWebAuthnDisplayName(t *testing.T) {
	testCases := []struct {
		testName    string
		name        string
		displayName string
		expected    string
	}{
		{
			testName:    "display name is set",
			name:        "Noah Kahan",
			displayName: "Noah",
			expected:    "Noah",
		},
		{
			testName:    "display name is empty",
			name:        "Noah Kahan",
			displayName: "",
			expected:    "Noah Kahan",
		},
	}
	for _, tc := range testCases {
		t.Run(tc.testName, func(t *testing.T) {
			// Create a user instance
			user := &webauthn.User{
				DisplayName: tc.displayName,
				Name:        tc.name,
			}

			result := user.WebAuthnDisplayName()
			assert.Equal(t, tc.expected, result)
		})
	}
}

func TestWebAuthnCredentials(t *testing.T) {
	// Create a user instance
	user := &webauthn.User{
		WebauthnCredentials: []gowebauthn.Credential{
			{
				ID: []byte("exampleID"),
			},
		},
	}

	// Call the WebAuthnCredentials method
	creds := user.WebAuthnCredentials()

	// Check if the returned value is correct
	assert.NotEmpty(t, creds)
	assert.Equal(t, user.WebauthnCredentials, creds)
}

func TestUserCredentialExcludeList(t *testing.T) {
	// Create a user instance
	user := &webauthn.User{
		WebauthnCredentials: []gowebauthn.Credential{
			{
				ID: []byte("exampleID"),
			},
		},
	}

	// Call the CredentialExcludeList method
	excludeList := user.CredentialExcludeList()

	// Check if the returned value is correct
	expectedExcludeList := []protocol.CredentialDescriptor{
		{
			Type:         protocol.PublicKeyCredentialType,
			CredentialID: []byte("exampleID"),
		},
	}
	assert.Equal(t, expectedExcludeList, excludeList)
}
