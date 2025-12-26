package handlers

import (
	"testing"
	"time"

	"github.com/brianvoe/gofakeit/v7"
	"github.com/samber/lo"
	"github.com/stretchr/testify/assert"

	ent "github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/common/enums"
)

func TestCreateUserInput(t *testing.T) {
	name := "Walter White"
	email := "ww@theopenlane.io"
	firstName := "Walter"
	lastName := "White"
	image := gofakeit.URL()

	testCases := []struct {
		testName string
		name     string
		email    string
		provider enums.AuthProvider
		expected ent.CreateUserInput
		image    string
	}{
		{
			testName: "oauth provider - github",
			name:     name,
			email:    email,
			provider: enums.AuthProviderGitHub,
			expected: ent.CreateUserInput{
				FirstName:       &firstName,
				LastName:        &lastName,
				Email:           email,
				AuthProvider:    &enums.AuthProviderGitHub,
				LastSeen:        lo.ToPtr(time.Now().UTC()),
				AvatarRemoteURL: &image,
			},
			image: image,
		},
		{
			testName: "oauth provider - github, only first name",
			name:     "meow",
			email:    email,
			provider: enums.AuthProviderGitHub,
			expected: ent.CreateUserInput{
				FirstName:       lo.ToPtr("meow"),
				Email:           email,
				AuthProvider:    &enums.AuthProviderGitHub,
				LastSeen:        lo.ToPtr(time.Now().UTC()),
				AvatarRemoteURL: &image,
			},
			image: image,
		},
		{
			testName: "oauth provider - google",
			name:     name,
			email:    email,
			provider: enums.AuthProviderGoogle,
			expected: ent.CreateUserInput{
				FirstName:       &firstName,
				LastName:        &lastName,
				Email:           email,
				AuthProvider:    &enums.AuthProviderGoogle,
				LastSeen:        lo.ToPtr(time.Now().UTC()),
				AvatarRemoteURL: &image,
			},
			image: image,
		},
		{
			testName: "webauthn provider",
			name:     name,
			email:    email,
			provider: enums.AuthProviderWebauthn,
			expected: ent.CreateUserInput{
				FirstName:       &firstName,
				LastName:        &lastName,
				Email:           email,
				AuthProvider:    &enums.AuthProviderWebauthn,
				LastSeen:        lo.ToPtr(time.Now().UTC()),
				AvatarRemoteURL: nil,
			},
			image: "",
		},
		{
			testName: "no image",
			name:     "Wanda Maximoff",
			email:    "wmaximoff@marvel.com",
			provider: enums.AuthProviderWebauthn,
			expected: ent.CreateUserInput{
				FirstName:       lo.ToPtr("Wanda"),
				LastName:        lo.ToPtr("Maximoff"),
				Email:           "wmaximoff@marvel.com",
				AuthProvider:    &enums.AuthProviderWebauthn,
				LastSeen:        lo.ToPtr(time.Now().UTC()),
				AvatarRemoteURL: nil,
			},
			image: "",
		},
	}
	for _, tc := range testCases {
		t.Run(tc.testName, func(t *testing.T) {
			input := createUserInput(tc.name, tc.email, tc.provider, tc.image)
			assert.Equal(t, tc.expected.FirstName, input.FirstName)
			assert.Equal(t, tc.expected.LastName, input.LastName)
			assert.Equal(t, tc.expected.Email, input.Email)
			assert.Equal(t, tc.expected.AuthProvider, input.AuthProvider)
			assert.WithinDuration(t, *tc.expected.LastSeen, *input.LastSeen, 2*time.Minute) // allow for a reasonable drift while tests are running
			assert.Equal(t, tc.expected.AvatarRemoteURL, input.AvatarRemoteURL)
		})
	}
}
