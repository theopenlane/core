package graphapi_test

import (
	"context"
	"testing"

	"github.com/samber/lo"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/theopenlane/core/pkg/openlaneclient"
	"github.com/theopenlane/iam/auth"
	mock_fga "github.com/theopenlane/iam/fgax/mockery"
)

func (suite *GraphTestSuite) TestQuerySubscriber() {
	t := suite.T()

	// setup user context
	reqCtx, err := userContext()
	require.NoError(t, err)

	subscriber := (&SubscriberBuilder{client: suite.client}).MustNew(reqCtx, t)

	// setup valid user context with another org
	org := (&OrganizationBuilder{client: suite.client}).MustNew(reqCtx, t)

	reqCtx2, err := auth.NewTestContextWithOrgID(testUser.ID, org.ID)
	require.NoError(t, err)

	sub := (&SubscriberBuilder{client: suite.client}).MustNew(reqCtx2, t)

	testCases := []struct {
		name        string
		email       string
		client      *openlaneclient.OpenlaneClient
		ctx         context.Context
		shouldCheck bool
		wantErr     bool
	}{
		{
			name:        "happy path",
			email:       subscriber.Email,
			client:      suite.client.api,
			ctx:         reqCtx,
			shouldCheck: true,
			wantErr:     false,
		},
		{
			name:        "happy path, using api token",
			email:       subscriber.Email,
			client:      suite.client.apiWithToken,
			ctx:         context.Background(),
			shouldCheck: true,
			wantErr:     false,
		},
		{
			name:        "happy path, using personal access token",
			email:       subscriber.Email,
			client:      suite.client.apiWithPAT,
			ctx:         context.Background(),
			shouldCheck: true,
			wantErr:     false,
		},
		{
			name:        "invalid email",
			email:       "beep@boop.com",
			client:      suite.client.api,
			ctx:         reqCtx,
			shouldCheck: false,
			wantErr:     true,
		},
		{
			name:        "subscriber for another org",
			email:       sub.Email,
			client:      suite.client.api,
			ctx:         reqCtx,
			shouldCheck: false,
			wantErr:     true,
		},
		{
			name:        "subscriber for another org using api token",
			email:       sub.Email,
			client:      suite.client.apiWithToken,
			ctx:         context.Background(),
			shouldCheck: false,
			wantErr:     true,
		},
		{
			name:        "subscriber for another org using personal access token",
			email:       sub.Email,
			client:      suite.client.apiWithPAT,
			ctx:         context.Background(),
			shouldCheck: false,
			wantErr:     true,
		},
	}

	for _, tc := range testCases {
		t.Run("Get "+tc.name, func(t *testing.T) {
			defer mock_fga.ClearMocks(suite.client.fga)

			if tc.shouldCheck {
				mock_fga.CheckAny(t, suite.client.fga, true)
			}

			resp, err := tc.client.GetSubscriberByEmail(tc.ctx, tc.email)

			if tc.wantErr {
				require.Error(t, err)
				assert.Nil(t, resp)

				return
			}

			require.NoError(t, err)
			require.NotNil(t, resp)
			require.NotNil(t, resp.Subscriber)
		})
	}
}

func (suite *GraphTestSuite) TestQuerySubscribers() {
	t := suite.T()

	// setup user context
	reqCtx, err := userContext()
	require.NoError(t, err)

	_ = (&SubscriberBuilder{client: suite.client}).MustNew(reqCtx, t)
	_ = (&SubscriberBuilder{client: suite.client}).MustNew(reqCtx, t)

	// setup valid user context with another org

	org := (&OrganizationBuilder{client: suite.client}).MustNew(reqCtx, t)
	reqCtx2, err := auth.NewTestContextWithOrgID(testUser.ID, org.ID)
	require.NoError(t, err)

	_ = (&SubscriberBuilder{client: suite.client}).MustNew(reqCtx2, t)

	testCases := []struct {
		name        string
		client      *openlaneclient.OpenlaneClient
		ctx         context.Context
		numExpected int
		check       bool
	}{
		{
			name:        "happy path, multiple subscribers",
			client:      suite.client.api,
			ctx:         reqCtx,
			check:       false,
			numExpected: 2,
		},
		{
			name:        "happy path, multiple subscribers using api token",
			client:      suite.client.apiWithToken,
			ctx:         context.Background(),
			check:       false,
			numExpected: 2,
		},
		{
			name:        "happy path, multiple subscribers using personal access token",
			client:      suite.client.apiWithPAT,
			ctx:         context.Background(),
			check:       false,
			numExpected: 2,
		},
		{
			name:        "happy path, one subscriber",
			client:      suite.client.api,
			ctx:         reqCtx2,
			check:       true,
			numExpected: 1,
		},
	}

	for _, tc := range testCases {
		t.Run("Get "+tc.name, func(t *testing.T) {
			defer mock_fga.ClearMocks(suite.client.fga)

			// TODO (sfunk): this is because of 1 vs multiple returned, look at the filter check
			if tc.check {
				mock_fga.CheckAny(t, suite.client.fga, true)
			}

			resp, err := tc.client.GetAllSubscribers(tc.ctx)

			require.NoError(t, err)
			require.NotNil(t, resp)
			require.Len(t, resp.Subscribers.Edges, tc.numExpected)
		})
	}
}

func (suite *GraphTestSuite) TestMutationCreateSubscriber() {
	t := suite.T()

	// setup user context
	reqCtx, err := userContext()
	require.NoError(t, err)

	testCases := []struct {
		name        string
		email       string
		ownerID     string
		listObjects bool
		client      *openlaneclient.OpenlaneClient
		ctx         context.Context
		wantErr     bool
	}{
		{
			name:        "happy path, new subscriber",
			email:       "c.stark@example.com",
			client:      suite.client.api,
			listObjects: true,
			ctx:         reqCtx,
			wantErr:     false,
		},
		{
			name:        "happy path, new subscriber using api token",
			email:       "e.stark@example.com",
			client:      suite.client.apiWithToken,
			listObjects: false,
			ctx:         context.Background(),
			wantErr:     false,
		},
		{
			name:        "happy path, new subscriber using personal access token",
			email:       "a.stark@example.com",
			ownerID:     testOrgID,
			client:      suite.client.apiWithPAT,
			listObjects: true,
			ctx:         context.Background(),
			wantErr:     false,
		},
		{
			name:        "missing email",
			client:      suite.client.api,
			listObjects: false,
			ctx:         reqCtx,
			wantErr:     true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			defer mock_fga.ClearMocks(suite.client.fga)

			mock_fga.CheckAny(t, suite.client.fga, true)

			if tc.listObjects {
				mock_fga.ListAny(t, suite.client.fga, []string{"organization:" + testOrgID})
			}

			input := openlaneclient.CreateSubscriberInput{
				Email: tc.email,
			}

			if tc.ownerID != "" {
				input.OwnerID = &tc.ownerID
			}

			resp, err := tc.client.CreateSubscriber(tc.ctx, input)

			if tc.wantErr {
				require.Error(t, err)
				assert.Nil(t, resp)

				return
			}

			require.NoError(t, err)
			require.NotNil(t, resp)

			// Assert matching fields
			assert.Equal(t, tc.email, resp.CreateSubscriber.Subscriber.Email)

		})
	}
}

func (suite *GraphTestSuite) TestUpdateSubscriber() {
	t := suite.T()

	// setup user context
	reqCtx, err := userContext()
	require.NoError(t, err)

	subscriber := (&SubscriberBuilder{client: suite.client}).MustNew(reqCtx, t)

	// setup valid user context with another org
	org := (&OrganizationBuilder{client: suite.client}).MustNew(reqCtx, t)

	reqCtx2, err := auth.NewTestContextWithOrgID(testUser.ID, org.ID)
	require.NoError(t, err)

	sub := (&SubscriberBuilder{client: suite.client}).MustNew(reqCtx2, t)

	testCases := []struct {
		name        string
		email       string
		updateInput openlaneclient.UpdateSubscriberInput
		client      *openlaneclient.OpenlaneClient
		ctx         context.Context
		shouldCheck bool
		wantErr     bool
	}{
		{
			name:  "happy path",
			email: subscriber.Email,
			updateInput: openlaneclient.UpdateSubscriberInput{
				PhoneNumber: lo.ToPtr("+1-555-867-5309"),
			},
			client:  suite.client.api,
			ctx:     reqCtx,
			wantErr: false,
		},
		{
			name:  "happy path, using api token",
			email: subscriber.Email,
			updateInput: openlaneclient.UpdateSubscriberInput{
				PhoneNumber: lo.ToPtr("+1-555-867-5310"),
			},
			client:  suite.client.apiWithToken,
			ctx:     context.Background(),
			wantErr: false,
		},
		{
			name:  "happy path, using personal access token",
			email: subscriber.Email,
			updateInput: openlaneclient.UpdateSubscriberInput{
				PhoneNumber: lo.ToPtr("+1-555-867-5311"),
			},
			client:  suite.client.apiWithPAT,
			ctx:     context.Background(),
			wantErr: false,
		},
		{
			name:  "invalid email",
			email: "beep@boop.com",
			updateInput: openlaneclient.UpdateSubscriberInput{
				PhoneNumber: lo.ToPtr("+1-555-867-5309"),
			},
			client:  suite.client.api,
			ctx:     reqCtx,
			wantErr: true,
		},
		{
			name:  "subscriber for another org",
			email: sub.Email,
			updateInput: openlaneclient.UpdateSubscriberInput{
				PhoneNumber: lo.ToPtr("+1-555-867-5309"),
			},
			client:  suite.client.api,
			ctx:     reqCtx,
			wantErr: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			defer mock_fga.ClearMocks(suite.client.fga)

			if !tc.wantErr {
				mock_fga.CheckAny(t, suite.client.fga, true)
			}

			resp, err := tc.client.UpdateSubscriber(tc.ctx, tc.email, tc.updateInput)

			if tc.wantErr {
				require.Error(t, err)
				assert.Nil(t, resp)

				return
			}

			require.NoError(t, err)
			require.NotNil(t, resp)
			require.Equal(t, tc.email, resp.UpdateSubscriber.Subscriber.Email)

			if tc.updateInput.PhoneNumber != nil {
				require.Equal(t, tc.updateInput.PhoneNumber, resp.UpdateSubscriber.Subscriber.PhoneNumber)
			}
		})
	}
}

func (suite *GraphTestSuite) TestDeleteSubscriber() {
	t := suite.T()

	// setup user context
	reqCtx, err := userContext()
	require.NoError(t, err)

	subscriber1 := (&SubscriberBuilder{client: suite.client}).MustNew(reqCtx, t)
	subscriber2 := (&SubscriberBuilder{client: suite.client}).MustNew(reqCtx, t)
	subscriber3 := (&SubscriberBuilder{client: suite.client}).MustNew(reqCtx, t)

	// setup valid user context with another org
	org := (&OrganizationBuilder{client: suite.client}).MustNew(reqCtx, t)

	reqCtx2, err := auth.NewTestContextWithOrgID(testUser.ID, org.ID)
	require.NoError(t, err)

	sub := (&SubscriberBuilder{client: suite.client}).MustNew(reqCtx2, t)

	testCases := []struct {
		name           string
		email          string
		organizationID string
		client         *openlaneclient.OpenlaneClient
		ctx            context.Context
		wantErr        bool
	}{
		{
			name:    "happy path",
			email:   subscriber1.Email,
			client:  suite.client.api,
			ctx:     reqCtx,
			wantErr: false,
		},
		{
			name:    "happy path, using api token",
			email:   subscriber2.Email,
			client:  suite.client.apiWithToken,
			ctx:     context.Background(),
			wantErr: false,
		},
		{
			name:           "happy path, using personal access token",
			email:          subscriber3.Email,
			organizationID: testOrgID,
			client:         suite.client.apiWithPAT,
			ctx:            context.Background(),
			wantErr:        false,
		},
		{
			name:    "invalid email",
			email:   "beep@boop.com",
			client:  suite.client.api,
			ctx:     reqCtx,
			wantErr: true,
		},
		{
			name:    "subscriber for another org",
			email:   sub.Email,
			client:  suite.client.api,
			ctx:     reqCtx,
			wantErr: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			defer mock_fga.ClearMocks(suite.client.fga)

			mock_fga.CheckAny(t, suite.client.fga, true)

			resp, err := tc.client.DeleteSubscriber(tc.ctx, tc.email, &tc.organizationID)

			if tc.wantErr {
				require.Error(t, err)
				assert.Nil(t, resp)

				return
			}

			require.NoError(t, err)
			require.NotNil(t, resp)
			require.Equal(t, tc.email, resp.DeleteSubscriber.Email)
		})
	}
}
