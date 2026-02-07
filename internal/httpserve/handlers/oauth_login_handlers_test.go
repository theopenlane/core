package handlers_test

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/samber/lo"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/theopenlane/echox/middleware/echocontext"

	"github.com/theopenlane/core/common/enums"
	ent "github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/generated/privacy"
	"github.com/theopenlane/core/pkg/middleware/transaction"
	"github.com/theopenlane/utils/ulids"
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
	baseCtx := privacy.DecisionContext(ec, privacy.Allow)
	baseCtx = ent.NewContext(baseCtx, suite.db)

	// Seed an existing user for tests that require update paths.
	createExistingUser := func(ctx context.Context, name, email string, provider enums.AuthProvider, image string) {
		t.Helper()

		parts := strings.SplitN(name, " ", 2)
		firstName := parts[0]
		lastName := ""
		if len(parts) > 1 {
			lastName = parts[1]
		}

		userSetting := suite.db.UserSetting.Create().
			SetEmailConfirmed(true).
			SaveX(ctx)

		builder := suite.db.User.Create().
			SetFirstName(firstName).
			SetLastName(lastName).
			SetEmail(email).
			SetSetting(userSetting).
			SetAuthProvider(provider).
			SetLastLoginProvider(provider).
			SetLastSeen(time.Now())

		if image != "" {
			builder.SetAvatarRemoteURL(image)
		}

		_, err := builder.Save(ctx)
		require.NoError(t, err)
	}

	type args struct {
		name     string
		email    string
		provider enums.AuthProvider
		image    string
	}

	tests := []struct {
		name    string
		args    args
		seed    func(ctx context.Context, email string)
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
			seed: func(ctx context.Context, email string) {
				createExistingUser(ctx, "Wanda Maximoff", email, enums.AuthProviderGitHub, "https://example.com/images/photo.jpg")
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
			seed: func(ctx context.Context, email string) {
				createExistingUser(ctx, "Wanda Maximoff", email, enums.AuthProviderGitHub, "https://example.com/images/photo.jpg")
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
			seed: func(ctx context.Context, email string) {
				createExistingUser(ctx, "Wand Maxim", email, enums.AuthProviderGitHub, "https://example.com/images/photo.jpg")
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
			ctx := baseCtx

			// Ensure unique emails per subtest to avoid cross-test contamination.
			args := tt.args
			want := *tt.want

			// ensure unique email per subtest to avoid cross-test contamination
			args.email = strings.ToLower(ulids.New().String()) + "@marvel.com"
			want.Email = args.email

			if tt.seed != nil {
				tt.seed(ctx, args.email)
			}

			now := time.Now()

			// start transaction because the query expects a transaction in the context
			tx, err := suite.h.DBClient.Tx(ctx)
			require.NoError(t, err)

			// commit transaction after test finishes
			defer tx.Commit() //nolint:errcheck

			// set transaction in the context
			ctx = transaction.NewContext(ctx, tx)

			got, err := suite.h.CheckAndCreateUser(ctx, args.name, args.email, args.provider, args.image)
			if tt.wantErr {
				require.Error(t, err)
				assert.Nil(t, got)

				return
			}

			// check if user was created
			require.NoError(t, err)
			require.NotNil(t, got)

			// verify fields
			assert.Equal(t, want.FirstName, got.FirstName)
			assert.Equal(t, want.LastName, got.LastName)
			assert.Equal(t, want.Email, got.Email)
			assert.Equal(t, want.AuthProvider, got.AuthProvider)
			assert.Equal(t, want.LastLoginProvider, got.LastLoginProvider)
			assert.WithinDuration(t, now, *got.LastSeen, time.Second*5)

			if args.image == "" {
				assert.NotEmpty(t, got.AvatarRemoteURL)
			} else {
				assert.Equal(t, *want.AvatarRemoteURL, *got.AvatarRemoteURL)
			}
		})
	}
}
