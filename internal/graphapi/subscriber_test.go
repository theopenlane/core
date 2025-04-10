package graphapi_test

import (
	"context"
	"strings"
	"testing"

	"github.com/samber/lo"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/theopenlane/core/pkg/openlaneclient"
)

func (suite *GraphTestSuite) TestQuerySubscriber() {
	t := suite.T()

	subscriber := (&SubscriberBuilder{client: suite.client}).MustNew(testUser1.UserCtx, t)

	subscriber2 := (&SubscriberBuilder{client: suite.client}).MustNew(testUser2.UserCtx, t)

	testCases := []struct {
		name    string
		email   string
		client  *openlaneclient.OpenlaneClient
		ctx     context.Context
		wantErr bool
	}{
		{
			name:    "happy path",
			email:   subscriber.Email,
			client:  suite.client.api,
			ctx:     testUser1.UserCtx,
			wantErr: false,
		},
		{
			name:    "happy path, using api token",
			email:   subscriber.Email,
			client:  suite.client.apiWithToken,
			ctx:     context.Background(),
			wantErr: false,
		},
		{
			name:    "happy path, using personal access token",
			email:   subscriber.Email,
			client:  suite.client.apiWithPAT,
			ctx:     context.Background(),
			wantErr: false,
		},
		{
			name:    "invalid email",
			email:   "beep@boop.com",
			client:  suite.client.api,
			ctx:     testUser1.UserCtx,
			wantErr: true,
		},
		{
			name:    "subscriber for another org",
			email:   subscriber2.Email,
			client:  suite.client.api,
			ctx:     testUser1.UserCtx,
			wantErr: true,
		},
		{
			name:    "subscriber for another org using api token",
			email:   subscriber2.Email,
			client:  suite.client.apiWithToken,
			ctx:     context.Background(),
			wantErr: true,
		},
		{
			name:    "subscriber for another org using personal access token",
			email:   subscriber2.Email,
			client:  suite.client.apiWithPAT,
			ctx:     context.Background(),
			wantErr: true,
		},
	}

	for _, tc := range testCases {
		t.Run("Get "+tc.name, func(t *testing.T) {
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

	(&SubscriberBuilder{client: suite.client}).MustNew(testUser1.UserCtx, t)
	(&SubscriberBuilder{client: suite.client}).MustNew(testUser1.UserCtx, t)

	(&SubscriberBuilder{client: suite.client}).MustNew(testUser2.UserCtx, t)

	testCases := []struct {
		name        string
		client      *openlaneclient.OpenlaneClient
		ctx         context.Context
		numExpected int
	}{
		{
			name:        "happy path, multiple subscribers",
			client:      suite.client.api,
			ctx:         testUser1.UserCtx,
			numExpected: 2,
		},
		{
			name:        "happy path, multiple subscribers using api token",
			client:      suite.client.apiWithToken,
			ctx:         context.Background(),
			numExpected: 2,
		},
		{
			name:        "happy path, multiple subscribers using personal access token",
			client:      suite.client.apiWithPAT,
			ctx:         context.Background(),
			numExpected: 2,
		},
		{
			name:        "happy path, one subscriber",
			client:      suite.client.api,
			ctx:         testUser2.UserCtx,
			numExpected: 1,
		},
	}

	for _, tc := range testCases {
		t.Run("Get "+tc.name, func(t *testing.T) {
			resp, err := tc.client.GetAllSubscribers(tc.ctx)

			require.NoError(t, err)
			require.NotNil(t, resp)
			require.Len(t, resp.Subscribers.Edges, tc.numExpected)
		})
	}
}

func (suite *GraphTestSuite) TestMutationCreateSubscriber() {
	t := suite.T()

	testCases := []struct {
		name             string
		email            string
		ownerID          string
		setUnsubscribed  bool
		client           *openlaneclient.OpenlaneClient
		ctx              context.Context
		wantErr          bool
		expectedAttempts int64
	}{
		{
			name:             "happy path, new subscriber",
			email:            "c.stark@example.com",
			setUnsubscribed:  true, //unsubscribe the subscriber to test for re-creation
			client:           suite.client.api,
			ctx:              testUser1.UserCtx,
			wantErr:          false,
			expectedAttempts: 1,
		},
		{
			name:             "happy path, duplicate subscriber but original was unsubscribed",
			email:            "c.stark@example.com",
			client:           suite.client.api,
			ctx:              testUser1.UserCtx,
			wantErr:          false,
			expectedAttempts: 2,
		},
		{
			name:             "happy path, duplicate subscriber, case insensitive",
			email:            "c.STARK@example.com",
			client:           suite.client.api,
			ctx:              testUser1.UserCtx,
			wantErr:          false,
			expectedAttempts: 3,
		},
		{
			name:             "happy path, new subscriber using api token",
			email:            "e.stark@example.com",
			client:           suite.client.apiWithToken,
			ctx:              context.Background(),
			wantErr:          false,
			expectedAttempts: 1,
		},
		{
			name:             "happy path, new subscriber using personal access token",
			email:            "a.stark@example.com",
			ownerID:          testUser1.OrganizationID,
			client:           suite.client.apiWithPAT,
			ctx:              context.Background(),
			wantErr:          false,
			expectedAttempts: 1,
		},
		{
			name:    "missing email",
			client:  suite.client.api,
			ctx:     testUser1.UserCtx,
			wantErr: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
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
			// Since we convert to lower case already on insertion/update
			assert.Equal(t, strings.ToLower(tc.email), resp.CreateSubscriber.Subscriber.Email)
			assert.False(t, resp.CreateSubscriber.Subscriber.Unsubscribed)

			if tc.setUnsubscribed {
				// Set the subscriber as unsubscribed to test for duplicate email
				resp, err := tc.client.UpdateSubscriber(tc.ctx, resp.CreateSubscriber.Subscriber.Email, openlaneclient.UpdateSubscriberInput{
					Unsubscribed: lo.ToPtr(true),
				})
				require.NoError(t, err)
				require.NotNil(t, resp)

				require.True(t, resp.UpdateSubscriber.Subscriber.Unsubscribed) // ensure the subscriber is unsubscribed now
				require.False(t, resp.UpdateSubscriber.Subscriber.Active)      // ensure the subscriber is inactive now after unsubscribing
				require.Zero(t, resp.UpdateSubscriber.Subscriber.SendAttempts) // reset attempts count to zero
			}
		})
	}
}

func (suite *GraphTestSuite) TestUpdateSubscriber() {
	t := suite.T()

	subscriber := (&SubscriberBuilder{client: suite.client}).MustNew(testUser1.UserCtx, t)
	subscriber2 := (&SubscriberBuilder{client: suite.client}).MustNew(testUser2.UserCtx, t)

	testCases := []struct {
		name        string
		email       string
		updateInput openlaneclient.UpdateSubscriberInput
		client      *openlaneclient.OpenlaneClient
		ctx         context.Context
		wantErr     bool
	}{
		{
			name:  "happy path",
			email: subscriber.Email,
			updateInput: openlaneclient.UpdateSubscriberInput{
				PhoneNumber: lo.ToPtr("+1-555-867-5309"),
			},
			client:  suite.client.api,
			ctx:     testUser1.UserCtx,
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
			name:  "happy path, using api token, set unsubscribed = false",
			email: subscriber.Email,
			updateInput: openlaneclient.UpdateSubscriberInput{
				Unsubscribed: lo.ToPtr(true),
			},
			client:  suite.client.apiWithToken,
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
			ctx:     testUser1.UserCtx,
			wantErr: true,
		},
		{
			name:  "subscriber for another org",
			email: subscriber2.Email,
			updateInput: openlaneclient.UpdateSubscriberInput{
				PhoneNumber: lo.ToPtr("+1-555-867-5309"),
			},
			client:  suite.client.api,
			ctx:     testUser1.UserCtx,
			wantErr: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
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

			if tc.updateInput.Unsubscribed != nil {
				require.Equal(t, *tc.updateInput.Unsubscribed, resp.UpdateSubscriber.Subscriber.Unsubscribed)

				if *tc.updateInput.Unsubscribed {
					// ensure I can create another subscriber with the same email
					resp, err := tc.client.CreateSubscriber(tc.ctx, openlaneclient.CreateSubscriberInput{
						Email: tc.email,
					})
					require.NoError(t, err)
					require.NotNil(t, resp)
				}
			}
		})
	}
}

func (suite *GraphTestSuite) TestDeleteSubscriber() {
	t := suite.T()

	subscriber1 := (&SubscriberBuilder{client: suite.client}).MustNew(testUser1.UserCtx, t)
	subscriber2 := (&SubscriberBuilder{client: suite.client}).MustNew(testUser1.UserCtx, t)
	subscriber3 := (&SubscriberBuilder{client: suite.client}).MustNew(testUser1.UserCtx, t)

	subscriberOtherOrg := (&SubscriberBuilder{client: suite.client}).MustNew(testUser2.UserCtx, t)

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
			ctx:     testUser1.UserCtx,
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
			organizationID: testUser1.OrganizationID,
			client:         suite.client.apiWithPAT,
			ctx:            context.Background(),
			wantErr:        false,
		},
		{
			name:    "invalid email",
			email:   "beep@boop.com",
			client:  suite.client.api,
			ctx:     testUser1.UserCtx,
			wantErr: true,
		},
		{
			name:    "subscriber for another org",
			email:   subscriberOtherOrg.Email,
			client:  suite.client.api,
			ctx:     testUser1.UserCtx,
			wantErr: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
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
