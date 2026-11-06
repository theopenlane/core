package handlers

import (
	"testing"
	"time"

	"github.com/brianvoe/gofakeit/v7"

	"github.com/theopenlane/core/common/enums"
	ent "github.com/theopenlane/core/v2/internal/ent/generated"
	"gotest.tools/v3/assert"
	is "gotest.tools/v3/assert/cmp"
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
				LastSeen:        new(time.Now().UTC()),
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
				FirstName:       new("meow"),
				Email:           email,
				AuthProvider:    &enums.AuthProviderGitHub,
				LastSeen:        new(time.Now().UTC()),
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
				LastSeen:        new(time.Now().UTC()),
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
				LastSeen:        new(time.Now().UTC()),
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
				FirstName:       new("Wanda"),
				LastName:        new("Maximoff"),
				Email:           "wmaximoff@marvel.com",
				AuthProvider:    &enums.AuthProviderWebauthn,
				LastSeen:        new(time.Now().UTC()),
				AvatarRemoteURL: nil,
			},
			image: "",
		},
	}
	for _, tc := range testCases {
		t.Run(tc.testName, func(t *testing.T) {
			input := createUserInput(tc.name, tc.email, tc.provider, tc.image)
			assert.Check(t, is.DeepEqual(tc.expected.FirstName, input.FirstName))
			assert.Check(t, is.DeepEqual(tc.expected.LastName, input.LastName))
			assert.Check(t, is.DeepEqual(tc.expected.Email, input.Email))
			assert.Check(t, is.DeepEqual(tc.expected.AuthProvider, input.AuthProvider))
			assert.Check(t, input.LastSeen.Sub(*tc.expected.LastSeen).Abs() <= 2*time.Minute) // allow for a reasonable drift while tests are running
			assert.Check(t, is.DeepEqual(tc.expected.AvatarRemoteURL, input.AvatarRemoteURL))
		})
	}
}
