package handlers_test

import (
	"context"
	"time"

	"github.com/brianvoe/gofakeit/v7"
	"github.com/stretchr/testify/require"
	ent "github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/generated/orgsubscription"
	"github.com/theopenlane/core/internal/ent/generated/privacy"
	"github.com/theopenlane/core/pkg/enums"
	"github.com/theopenlane/iam/auth"
	"github.com/theopenlane/utils/contextx"
	"github.com/theopenlane/utils/ulids"
)

var (
	testUser1     testUserDetails
	dummyFeatures = []string{"feature1", "feature2"}
)

// testUserDetails is a struct that holds the details of a test user
type testUserDetails struct {
	// ID is the ID of the user
	ID string
	// UserInfo contains all the details of the user
	UserInfo *ent.User
	// PersonalOrgID is the ID of the personal organization of the user
	PersonalOrgID string
	// OrganizationID is the ID of the organization of the user
	OrganizationID string
	// SubscriptionID is the ID of the subscription of the user
	SubscriptionID string
	// UserCtx is the context of the user that can be used for the test requests that require authentication
	UserCtx context.Context
}

// userBuilder creates a new test user and returns the details
// this includes a test user and an organization the user is the owner of
func (suite *HandlerTestSuite) userBuilder(ctx context.Context) testUserDetails {
	return suite.userBuilderWithInput(ctx, nil)
}

type userInput struct {
	email         string
	password      string
	confirmedUser bool
}

func (suite *HandlerTestSuite) userBuilderWithInput(ctx context.Context, input *userInput) testUserDetails {
	t := suite.T()

	testUser := testUserDetails{}

	ctx = privacy.DecisionContext(ctx, privacy.Allow)

	// create a test user
	var err error
	var userSetting *ent.UserSetting

	if input != nil && input.confirmedUser {
		userSetting, err = suite.db.UserSetting.Create().
			SetEmailConfirmed(true).
			SetIsTfaEnabled(true).
			Save(ctx)
		require.NoError(t, err)
	}

	email := gofakeit.Email()
	if input != nil && input.email != "" {
		email = input.email
	}

	builder := suite.db.User.Create().
		SetEmail(email).
		SetFirstName(gofakeit.FirstName()).
		SetLastName(gofakeit.LastName()).
		SetLastLoginProvider(enums.AuthProviderCredentials).
		SetLastSeen(time.Now())

	if input != nil && input.password != "" {
		builder.SetPassword(input.password)
	}

	if userSetting != nil {
		builder.SetSetting(userSetting)
	}

	createCtx := auth.NewTestContextWithOrgID(ulids.New().String(), ulids.New().String())
	createCtx = contextx.With(createCtx, auth.OrganizationCreationContextKey{})
	createCtx = privacy.DecisionContext(createCtx, privacy.Allow)
	createCtx = ent.NewContext(createCtx, suite.db)

	testUser.UserInfo, err = builder.Save(ctx)
	require.NoError(t, err)

	testUser.ID = testUser.UserInfo.ID

	// get the personal org for the user
	testPersonalOrg, err := testUser.UserInfo.Edges.Setting.DefaultOrg(ctx)
	require.NoError(t, err)

	testUser.PersonalOrgID = testPersonalOrg.ID

	// setup user context with the personal org
	userCtx := auth.NewTestContextWithOrgID(testUser.ID, testUser.PersonalOrgID)

	// set privacy allow in order to allow the creation of the users without
	// authentication in the tests seeds
	userCtx = privacy.DecisionContext(userCtx, privacy.Allow)

	// add client to context, required for hooks that expect the client to be in the context
	userCtx = ent.NewContext(userCtx, suite.db)

	// create a non-personal test organization
	orgSetting, err := suite.db.OrganizationSetting.Create().
		SetBillingEmail(testUser.UserInfo.Email).
		Save(userCtx)
	require.NoError(t, err)

	testOrg, err := suite.db.Organization.Create().
		SetName(gofakeit.AdjectiveDescriptive() + " " + gofakeit.Noun()).
		SetSettingID(orgSetting.ID).
		Save(userCtx)

	require.NoError(t, err)

	testUser.OrganizationID = testOrg.ID

	// add dummy subscription to the organization
	err = suite.db.OrgSubscription.Update().Where(orgsubscription.OwnerID(testOrg.ID)).
		SetFeatureLookupKeys(dummyFeatures).
		Exec(userCtx)
	require.NoError(t, err)

	// setup user context with the org (and not the personal org)
	testUser.UserCtx = auth.NewTestContextWithOrgID(testUser.ID, testUser.OrganizationID)
	testUser.UserCtx = privacy.DecisionContext(testUser.UserCtx, privacy.Allow)
	testUser.UserCtx = ent.NewContext(testUser.UserCtx, suite.db)

	return testUser
}

// setupTestData creates test users and sets up the clients with the necessary tokens
func (suite *HandlerTestSuite) setupTestData(ctx context.Context) {
	// create test users
	testUser1 = suite.userBuilder(ctx)
}
