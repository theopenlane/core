package graphapi_test

import (
	"context"
	"slices"
	"strings"
	"testing"

	"github.com/samber/lo"
	"gotest.tools/v3/assert"
	is "gotest.tools/v3/assert/cmp"

	"github.com/theopenlane/core/internal/graphapi/testclient"
	"github.com/theopenlane/ent/generated"
	"github.com/theopenlane/shared/gqlerrors"
)

func TestQuerySubscriber(t *testing.T) {
	subscriber := (&SubscriberBuilder{client: suite.client}).MustNew(testUser1.UserCtx, t)
	subscriber2 := (&SubscriberBuilder{client: suite.client}).MustNew(testUser2.UserCtx, t)

	testCases := []struct {
		name    string
		email   string
		client  *testclient.TestClient
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

				errors := parseClientError(t, err)
				for _, e := range errors {
					assertErrorCode(t, e, gqlerrors.NotFoundErrorCode)
					assertErrorMessage(t, e, "subscriber not found")
				}

				return
			}

			assert.NilError(t, err)
			assert.Assert(t, resp != nil)
			assert.Check(t, resp.Subscriber.ID != "")
		})
	}

	// cleanup
	(&Cleanup[*generated.SubscriberDeleteOne]{client: suite.client.db.Subscriber, ID: subscriber.ID}).MustDelete(testUser1.UserCtx, t)
	(&Cleanup[*generated.SubscriberDeleteOne]{client: suite.client.db.Subscriber, ID: subscriber2.ID}).MustDelete(testUser2.UserCtx, t)
}

func TestQuerySubscribers(t *testing.T) {
	s1 := (&SubscriberBuilder{client: suite.client}).MustNew(testUser1.UserCtx, t)
	s2 := (&SubscriberBuilder{client: suite.client}).MustNew(testUser1.UserCtx, t)
	s3 := (&SubscriberBuilder{client: suite.client}).MustNew(testUser2.UserCtx, t)

	testCases := []struct {
		name        string
		client      *testclient.TestClient
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
			resp, err := tc.client.GetAllSubscribers(tc.ctx, nil, nil, nil, nil, nil)

			assert.NilError(t, err)
			assert.Assert(t, resp != nil)
			assert.Check(t, is.Len(resp.Subscribers.Edges, tc.numExpected))
		})
	}

	// cleanup
	(&Cleanup[*generated.SubscriberDeleteOne]{client: suite.client.db.Subscriber, IDs: []string{s1.ID, s2.ID}}).MustDelete(testUser1.UserCtx, t)
	(&Cleanup[*generated.SubscriberDeleteOne]{client: suite.client.db.Subscriber, IDs: []string{s3.ID}}).MustDelete(testUser2.UserCtx, t)
}

func TestMutationCreateBulkSubscribers(t *testing.T) {

	testCases := []struct {
		name    string
		emails  []string
		client  *testclient.TestClient
		ctx     context.Context
		wantErr bool
	}{
		{
			name:    "happy path, multiple subscribers",
			emails:  []string{"e.stark@example.com", "y.stark@example.com"},
			client:  suite.client.apiWithToken,
			ctx:     context.Background(),
			wantErr: false,
		},
		{
			name:    "happy path, one subscriber for bulk endpoint",
			emails:  []string{"rr.stark@example.com"},
			client:  suite.client.apiWithToken,
			ctx:     context.Background(),
			wantErr: false,
		},
		{
			name:    "happy path, no provided email",
			emails:  []string{},
			client:  suite.client.apiWithToken,
			ctx:     context.Background(),
			wantErr: false,
		},
		{
			name:    "bad path, invalid emails provided",
			emails:  []string{"not_a_valid_email"},
			client:  suite.client.apiWithToken,
			ctx:     context.Background(),
			wantErr: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			input := []*testclient.CreateSubscriberInput{}

			for _, v := range tc.emails {
				input = append(input, &testclient.CreateSubscriberInput{
					Email: v,
				})
			}

			resp, err := tc.client.CreateBulkSubscriber(tc.ctx, input)

			if tc.wantErr {

				return
			}

			assert.NilError(t, err)
			assert.Assert(t, resp != nil)

			for k, v := range tc.emails {
				assert.Check(t, is.Equal(strings.ToLower(v), resp.CreateBulkSubscriber.Subscribers[k].Email))
				assert.Check(t, !resp.CreateBulkSubscriber.Subscribers[k].Unsubscribed)
				assert.Check(t, !resp.CreateBulkSubscriber.Subscribers[k].Active)
			}

			// cleanup
			for _, v := range resp.CreateBulkSubscriber.Subscribers {
				(&Cleanup[*generated.SubscriberDeleteOne]{client: suite.client.db.Subscriber, ID: v.ID}).MustDelete(testUser1.UserCtx, t)
			}
		})
	}
}

func TestMutationCreateSubscriber_Tokens(t *testing.T) {

	testCases := []struct {
		name             string
		email            string
		ownerID          string
		setUnsubscribed  bool
		client           *testclient.TestClient
		ctx              context.Context
		wantErr          bool
		expectedAttempts int
	}{
		{
			name:    "happy path, new subscriber using api token",
			email:   "e.stark@example.com",
			client:  suite.client.apiWithToken,
			ctx:     context.Background(),
			wantErr: false,
		},
		{
			name:    "happy path, new subscriber using personal access token",
			email:   "a.stark@example.com",
			ownerID: testUser1.OrganizationID,
			client:  suite.client.apiWithPAT,
			ctx:     context.Background(),
			wantErr: false,
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
			input := testclient.CreateSubscriberInput{
				Email: tc.email,
			}

			if tc.ownerID != "" {
				input.OwnerID = &tc.ownerID
			}

			resp, err := tc.client.CreateSubscriber(tc.ctx, input)

			if tc.wantErr {
				assert.ErrorContains(t, err, "email is required but not provided")

				return
			}

			assert.NilError(t, err)
			assert.Assert(t, resp != nil)

			// Assert matching fields
			// Since we convert to lower case already on insertion/update
			assert.Check(t, is.Equal(strings.ToLower(tc.email), resp.CreateSubscriber.Subscriber.Email))
			assert.Check(t, !resp.CreateSubscriber.Subscriber.Unsubscribed)

			// cleanup
			(&Cleanup[*generated.SubscriberDeleteOne]{client: suite.client.db.Subscriber, ID: resp.CreateSubscriber.Subscriber.ID}).MustDelete(testUser1.UserCtx, t)
		})
	}
}

func TestMutationCreateSubscriber_SendAttempts(t *testing.T) {

	createdSubscriberEmails := []string{}
	testCases := []struct {
		name              string
		email             string
		ownerID           string
		setUnsubscribed   bool
		client            *testclient.TestClient
		ctx               context.Context
		wantErr           bool
		expectedMessage   string
		expectedErrorCode string
		expectedAttempts  int64
	}{
		{
			name:             "happy path, new subscriber",
			email:            "c.stark@example.com",
			setUnsubscribed:  true, //unsubscribe the subscriber to test for re-creation
			client:           suite.client.api,
			ctx:              testUser1.UserCtx,
			wantErr:          false,
			expectedAttempts: 0, // since we unsubscribe, it should reset
		},
		{
			name:             "happy path, duplicate subscriber but original was unsubscribed",
			email:            "c.stark@example.com",
			client:           suite.client.api,
			ctx:              testUser1.UserCtx,
			wantErr:          false,
			expectedAttempts: 1,
		},
		{
			name:             "happy path, duplicate subscriber, case insensitive",
			email:            "c.STARK@example.com",
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
			name:             "happy path, duplicate subscriber, case insensitive",
			email:            "c.STARK@example.com",
			client:           suite.client.api,
			ctx:              testUser1.UserCtx,
			wantErr:          false,
			expectedAttempts: 4,
		},
		{
			name:             "happy path, duplicate subscriber, case insensitive",
			email:            "c.STARK@example.com",
			client:           suite.client.api,
			ctx:              testUser1.UserCtx,
			wantErr:          false,
			expectedAttempts: 5,
		},
		{
			name:              "happy path, duplicate subscriber, case insensitive, max attempts",
			email:             "c.STARK@example.com",
			client:            suite.client.api,
			ctx:               testUser1.UserCtx,
			wantErr:           true,
			expectedErrorCode: gqlerrors.MaxAttemptsErrorCode,
			expectedMessage:   "max attempts reached for this email, please reach out to support",
			expectedAttempts:  5,
		},
		{
			name:              "missing email",
			client:            suite.client.api,
			ctx:               testUser1.UserCtx,
			expectedErrorCode: gqlerrors.BadRequestErrorCode,
			expectedMessage:   "subscriber email is required, please provide a valid email",
			wantErr:           true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			input := testclient.CreateSubscriberInput{
				Email: tc.email,
			}

			if tc.ownerID != "" {
				input.OwnerID = &tc.ownerID
			}

			resp, err := tc.client.CreateSubscriber(tc.ctx, input)

			if tc.wantErr {
				errors := parseClientError(t, err)
				for _, e := range errors {
					assertErrorCode(t, e, tc.expectedErrorCode)
					assertErrorMessage(t, e, tc.expectedMessage)
				}

				return
			}

			assert.NilError(t, err)
			assert.Assert(t, resp != nil)

			// Assert matching fields
			// Since we convert to lower case already on insertion/update
			assert.Check(t, is.Equal(strings.ToLower(tc.email), resp.CreateSubscriber.Subscriber.Email))
			assert.Check(t, !resp.CreateSubscriber.Subscriber.Unsubscribed)

			if tc.setUnsubscribed {
				// Set the subscriber as unsubscribed to test for duplicate email
				resp, err := tc.client.UpdateSubscriber(tc.ctx, resp.CreateSubscriber.Subscriber.Email, testclient.UpdateSubscriberInput{
					Unsubscribed: lo.ToPtr(true),
				})
				assert.NilError(t, err)
				assert.Assert(t, resp != nil)

				assert.Check(t, resp.UpdateSubscriber.Subscriber.Unsubscribed) // ensure the subscriber is unsubscribed now
				assert.Check(t, !resp.UpdateSubscriber.Subscriber.Active)      // ensure the subscriber is inactive now after unsubscribing
			}

			// add the list to cleanup later
			if !slices.Contains(createdSubscriberEmails, strings.ToLower(tc.email)) {
				createdSubscriberEmails = append(createdSubscriberEmails, strings.ToLower(tc.email))
			}

			sub, err := tc.client.GetSubscriberByEmail(tc.ctx, strings.ToLower(tc.email))
			assert.NilError(t, err)

			if tc.setUnsubscribed {
				assert.Equal(t, sub.Subscriber.SendAttempts, int64(0)) // reset attempts count to zero
				return
			}

			assert.Equal(t, tc.expectedAttempts, sub.Subscriber.SendAttempts)
		})
	}

	// cleanup
	for _, v := range createdSubscriberEmails {
		_, err := suite.client.api.DeleteSubscriber(testUser1.UserCtx, v, nil)
		assert.NilError(t, err)
	}
}

func TestUpdateSubscriber(t *testing.T) {
	subscriber := (&SubscriberBuilder{client: suite.client}).MustNew(testUser1.UserCtx, t)
	subscriber2 := (&SubscriberBuilder{client: suite.client}).MustNew(testUser2.UserCtx, t)

	testCases := []struct {
		name        string
		email       string
		updateInput testclient.UpdateSubscriberInput
		client      *testclient.TestClient
		ctx         context.Context
		wantErr     bool
	}{
		{
			name:  "happy path",
			email: subscriber.Email,
			updateInput: testclient.UpdateSubscriberInput{
				PhoneNumber: lo.ToPtr("+1-555-867-5309"),
			},
			client:  suite.client.api,
			ctx:     testUser1.UserCtx,
			wantErr: false,
		},
		{
			name:  "happy path, using api token",
			email: subscriber.Email,
			updateInput: testclient.UpdateSubscriberInput{
				PhoneNumber: lo.ToPtr("+1-555-867-5310"),
			},
			client:  suite.client.apiWithToken,
			ctx:     context.Background(),
			wantErr: false,
		},
		{
			name:  "happy path, using personal access token",
			email: subscriber.Email,
			updateInput: testclient.UpdateSubscriberInput{
				PhoneNumber: lo.ToPtr("+1-555-867-5311"),
			},
			client:  suite.client.apiWithPAT,
			ctx:     context.Background(),
			wantErr: false,
		},
		{
			name:  "happy path, using api token, set unsubscribed = false",
			email: subscriber.Email,
			updateInput: testclient.UpdateSubscriberInput{
				Unsubscribed: lo.ToPtr(true),
			},
			client:  suite.client.apiWithToken,
			ctx:     context.Background(),
			wantErr: false,
		},
		{
			name:  "invalid email",
			email: "beep@boop.com",
			updateInput: testclient.UpdateSubscriberInput{
				PhoneNumber: lo.ToPtr("+1-555-867-5309"),
			},
			client:  suite.client.api,
			ctx:     testUser1.UserCtx,
			wantErr: true,
		},
		{
			name:  "subscriber for another org",
			email: subscriber2.Email,
			updateInput: testclient.UpdateSubscriberInput{
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
				assert.ErrorContains(t, err, "subscriber not found")
				return
			}

			assert.NilError(t, err)
			assert.Assert(t, resp != nil)
			assert.Equal(t, tc.email, resp.UpdateSubscriber.Subscriber.Email)

			if tc.updateInput.PhoneNumber != nil {
				assert.Check(t, is.Equal(*tc.updateInput.PhoneNumber, *resp.UpdateSubscriber.Subscriber.PhoneNumber))
			}

			if tc.updateInput.Unsubscribed != nil {
				assert.Check(t, *tc.updateInput.Unsubscribed, resp.UpdateSubscriber.Subscriber.Unsubscribed)

				if *tc.updateInput.Unsubscribed {
					// ensure I can create another subscriber with the same email
					resp, err := tc.client.CreateSubscriber(tc.ctx, testclient.CreateSubscriberInput{
						Email: tc.email,
					})
					assert.NilError(t, err)
					assert.Assert(t, resp != nil)
				}
			}
		})
	}

	// cleanup
	(&Cleanup[*generated.SubscriberDeleteOne]{client: suite.client.db.Subscriber, ID: subscriber.ID}).MustDelete(testUser1.UserCtx, t)
	(&Cleanup[*generated.SubscriberDeleteOne]{client: suite.client.db.Subscriber, ID: subscriber2.ID}).MustDelete(testUser2.UserCtx, t)
}

func TestDeleteSubscriber(t *testing.T) {
	subscriber1 := (&SubscriberBuilder{client: suite.client}).MustNew(testUser1.UserCtx, t)
	subscriber2 := (&SubscriberBuilder{client: suite.client}).MustNew(testUser1.UserCtx, t)
	subscriber3 := (&SubscriberBuilder{client: suite.client}).MustNew(testUser1.UserCtx, t)

	subscriberOtherOrg := (&SubscriberBuilder{client: suite.client}).MustNew(testUser2.UserCtx, t)

	testCases := []struct {
		name           string
		email          string
		organizationID string
		client         *testclient.TestClient
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
				assert.ErrorContains(t, err, notFoundErrorMsg)

				return
			}

			assert.NilError(t, err)
			assert.Assert(t, resp != nil)
			assert.Equal(t, tc.email, resp.DeleteSubscriber.Email)
		})
	}

	// cleanup
	(&Cleanup[*generated.SubscriberDeleteOne]{client: suite.client.db.Subscriber, ID: subscriberOtherOrg.ID}).MustDelete(testUser2.UserCtx, t)
}

func TestActiveSubscriber(t *testing.T) {

	testCases := []struct {
		name       string
		email      string
		ownerID    string
		client     *testclient.TestClient
		ctx        context.Context
		wantErr    bool
		markActive bool
	}{
		{
			name:       "happy path, active subscriber",
			email:      "c.stark@example.com",
			client:     suite.client.api,
			ctx:        testUser1.UserCtx,
			wantErr:    false,
			markActive: true,
		},
		{
			name:       "happy path, resubscribing",
			email:      "aa.stark@example.com",
			client:     suite.client.api,
			ctx:        testUser1.UserCtx,
			wantErr:    false,
			markActive: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			input := testclient.CreateSubscriberInput{
				Email: tc.email,
			}

			if tc.ownerID != "" {
				input.OwnerID = &tc.ownerID
			}

			resp, err := tc.client.CreateSubscriber(tc.ctx, input)

			if tc.wantErr {
				assert.Assert(t, is.Nil(resp))

				return
			}

			assert.Assert(t, resp != nil)

			if tc.markActive {
				ctx := setContext(tc.ctx, suite.client.db)

				// update the subscriber and mark active
				_, err = suite.client.db.Subscriber.
					UpdateOneID(resp.CreateSubscriber.Subscriber.ID).
					SetActive(true).
					Save(ctx)

				assert.NilError(t, err)
			}

			_, err = tc.client.CreateSubscriber(tc.ctx, input)
			if tc.markActive {
				// if we marked the user as active, this should fail
				assert.ErrorContains(t, err, "subscriber already exists")

				return
			}

			assert.NilError(t, err)

			// cleanup
			(&Cleanup[*generated.SubscriberDeleteOne]{client: suite.client.db.Subscriber, ID: resp.CreateSubscriber.Subscriber.ID}).MustDelete(tc.ctx, t)
		})
	}
}
