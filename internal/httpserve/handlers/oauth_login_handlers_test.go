package handlers_test

import (
	"testing"
	"time"

	"github.com/samber/lo"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/theopenlane/echox/middleware/echocontext"

	ent "github.com/theopenlane/ent/generated"
	"github.com/theopenlane/ent/generated/privacy"
	"github.com/theopenlane/shared/enums"
	"github.com/theopenlane/shared/middleware/transaction"
)

func (suite *HandlerTestSuite) TestHandlerCheckAndCreateUser() {
	t := suite.T()

	// add login handler
	// Create operation for LoginHandler
	operation := suite.createImpersonationOperation("LoginHandler", "Login handler")
	suite.registerTestHandler("POST", "login", operation, suite.h.LoginHandler)

	ec := echocontext.NewTestEchoContext().Request().Context()

	// set privacy allow in order to allow the creation of the users without
	// authentication in the tests
	ctx := privacy.DecisionContext(ec, privacy.Allow)

	type args struct {
		name     string
		email    string
		provider enums.AuthProvider
		image    string
	}

	tests := []struct {
		name    string
		args    args
		want    *ent.User
		writes  bool
		wantErr bool
	}{
		{
			name: "happy path, github",
			args: args{
				name:     "Wanda Maximoff",
				email:    "wmaximoff@marvel.com",
				provider: enums.AuthProviderGitHub,
				image:    "https://example.com/images/photo.jpg",
			},
			want: &ent.User{
				FirstName:         "Wanda",
				LastName:          "Maximoff",
				Email:             "wmaximoff@marvel.com",
				AuthProvider:      enums.AuthProviderGitHub,
				LastLoginProvider: enums.AuthProviderGitHub,
				AvatarRemoteURL:   lo.ToPtr("https://example.com/images/photo.jpg"),
			},
			writes: true,
		},
		{
			name: "happy path, same email, different provider, should not fail",
			args: args{
				name:     "Wanda Maximoff",
				email:    "wmaximoff@marvel.com",
				provider: enums.AuthProviderGoogle,
				image:    "https://example.com/images/photo.jpg",
			},
			want: &ent.User{
				FirstName:         "Wanda",
				LastName:          "Maximoff",
				Email:             "wmaximoff@marvel.com",
				AuthProvider:      enums.AuthProviderGitHub,
				LastLoginProvider: enums.AuthProviderGoogle,
				AvatarRemoteURL:   lo.ToPtr("https://example.com/images/photo.jpg"),
			},
			writes:  true,
			wantErr: false,
		},
		{
			name: "user already exists, should not fail, just update last seen",
			args: args{
				name:     "Wanda Maximoff",
				email:    "wmaximoff@marvel.com",
				provider: enums.AuthProviderGitHub,
				image:    "https://example.com/images/photo.jpg",
			},
			want: &ent.User{
				FirstName:         "Wanda",
				LastName:          "Maximoff",
				Email:             "wmaximoff@marvel.com",
				AuthProvider:      enums.AuthProviderGitHub,
				LastLoginProvider: enums.AuthProviderGitHub,
				AvatarRemoteURL:   lo.ToPtr("https://example.com/images/photo.jpg"),
			},
			writes: false,
		},
		{
			name: "no image, avatar URL not nil ",
			args: args{
				name:     "Wand Maxim",
				email:    "wmaximoff1@marvel.com",
				provider: enums.AuthProviderGitHub,
				image:    "",
			},
			want: &ent.User{
				FirstName:         "Wand",
				LastName:          "Maxim",
				Email:             "wmaximoff1@marvel.com",
				AuthProvider:      enums.AuthProviderGitHub,
				LastLoginProvider: enums.AuthProviderGitHub,
			},
			writes: true,
		},
		{
			name: "no image, update last seen",
			args: args{
				name:     "Wand Maxim",
				email:    "wmaximoff1@marvel.com",
				provider: enums.AuthProviderGitHub,
				image:    "",
			},
			want: &ent.User{
				FirstName:         "Wand",
				LastName:          "Maxim",
				Email:             "wmaximoff1@marvel.com",
				AuthProvider:      enums.AuthProviderGitHub,
				LastLoginProvider: enums.AuthProviderGitHub,
			},
			writes: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			now := time.Now()

			// start transaction because the query expects a transaction in the context
			tx, err := suite.h.DBClient.Tx(ctx)
			require.NoError(t, err)

			// commit transaction after test finishes
			defer tx.Commit() //nolint:errcheck

			// set transaction in the context
			ctx = transaction.NewContext(ctx, tx)

			got, err := suite.h.CheckAndCreateUser(ctx, tt.args.name, tt.args.email, tt.args.provider, tt.args.image)
			if tt.wantErr {
				require.Error(t, err)
				assert.Nil(t, got)

				return
			}

			// check if user was created
			require.NoError(t, err)
			require.NotNil(t, got)

			// verify fields
			assert.Equal(t, tt.want.FirstName, got.FirstName)
			assert.Equal(t, tt.want.LastName, got.LastName)
			assert.Equal(t, tt.want.Email, got.Email)
			assert.Equal(t, tt.want.AuthProvider, got.AuthProvider)
			assert.Equal(t, tt.want.LastLoginProvider, got.LastLoginProvider)
			assert.WithinDuration(t, now, *got.LastSeen, time.Second*5)

			if tt.args.image == "" {
				assert.NotEmpty(t, got.AvatarRemoteURL)
			} else {
				assert.Equal(t, *tt.want.AvatarRemoteURL, *got.AvatarRemoteURL)
			}
		})
	}
}
