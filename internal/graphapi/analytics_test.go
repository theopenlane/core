package graphapi_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	ent "github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/graphapi"
)

func (suite *GraphTestSuite) TestTrackedEvent() {
	t := suite.T()

	testCases := []struct {
		mutation ent.Mutation
		expected bool
	}{
		{
			mutation: suite.client.db.User.Create().Mutation(),
			expected: true,
		},
		{
			mutation: suite.client.db.Organization.Create().Mutation(),
			expected: true,
		},
		{
			mutation: suite.client.db.User.Update().Mutation(),
			expected: false,
		},
		{
			mutation: suite.client.db.EmailVerificationToken.Create().Mutation(),
			expected: false,
		},
	}

	for _, tc := range testCases {
		actual := graphapi.TrackedEvent(tc.mutation)

		assert.Equal(t, tc.expected, actual)
	}
}

func (suite *GraphTestSuite) TestCreateEvent() {
	t := suite.T()

	reqCtx, err := userContext()
	require.NoError(t, err)

	testCases := []struct {
		name          string
		mutation      ent.Mutation
		v             ent.Value
		expectedEvent string
		expectedProps map[string]interface{}
	}{
		{
			name:     "user create event",
			mutation: suite.client.db.User.Create().Mutation(),
			v: ent.User{
				ID:        "123",
				FirstName: "Jack",
				LastName:  "Pearson",
				Email:     "jpearson@us.com",
			},
			expectedEvent: "user.created",
			expectedProps: map[string]interface{}{
				"email":      "jpearson@us.com",
				"first_name": "Jack",
				"last_name":  "Pearson",
				"user_id":    "123",
			},
		},
		{
			name:     "org create event",
			mutation: suite.client.db.Organization.Create().Mutation(),
			v: ent.Organization{
				ID:   "123",
				Name: "Meow",
			},
			expectedEvent: "organization.created",
			expectedProps: map[string]interface{}{
				"organization_name": "Meow",
				"organization_id":   "123",
			},
		},
		{
			name:     "missing id, no logs",
			mutation: suite.client.db.User.Update().Mutation(),
			v: ent.User{
				FirstName: "Jack",
				LastName:  "Pearson",
				Email:     "jpearson@us.com",
			},
			expectedEvent: "",
			expectedProps: map[string]interface{}{},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			out := suite.captureOutput(func() {
				graphapi.CreateEvent(reqCtx, suite.client.db, tc.mutation, tc.v)
			})

			assert.Contains(t, out, tc.expectedEvent)
			for k, v := range tc.expectedProps {
				assert.Contains(t, out, k)
				assert.Contains(t, out, v)
			}
		})
	}
}
