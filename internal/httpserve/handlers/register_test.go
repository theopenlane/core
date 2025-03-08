package handlers_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/riverqueue/river/riverdriver/riverpgxv5"
	"github.com/riverqueue/river/rivertest"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/theopenlane/newman"
	"github.com/theopenlane/riverboat/pkg/jobs"

	"github.com/theopenlane/utils/rout"

	"github.com/theopenlane/httpsling"
	"github.com/theopenlane/iam/auth"

	"github.com/theopenlane/core/internal/ent/generated/privacy"
	_ "github.com/theopenlane/core/internal/ent/generated/runtime"
	"github.com/theopenlane/core/internal/ent/generated/usersetting"
	"github.com/theopenlane/core/internal/httpserve/handlers"
	"github.com/theopenlane/core/pkg/enums"
	"github.com/theopenlane/core/pkg/models"
	"github.com/theopenlane/core/pkg/openlaneclient"
)

func (suite *HandlerTestSuite) TestRegisterHandler() {
	t := suite.T()

	// add handler
	suite.e.POST("register", suite.h.RegisterHandler)

	var bonkers = "b!a!n!a!n!a!s!"

	testCases := []struct {
		name               string
		email              string
		firstName          string
		lastName           string
		password           string
		emailExpected      bool
		expectedErrMessage string
		expectedErrorCode  rout.ErrorCode
		expectedStatus     int
		from               string
	}{
		{
			name:           "happy path",
			email:          "bananas@theopenlane.io",
			firstName:      "Princess",
			lastName:       "Fiona",
			password:       bonkers,
			emailExpected:  true,
			expectedStatus: http.StatusCreated,
		},
		{
			name:               "invalid email",
			email:              "bananas.net",
			firstName:          "Princess",
			lastName:           "Fiona",
			password:           bonkers,
			emailExpected:      false,
			expectedErrMessage: "email was invalid",
			expectedStatus:     http.StatusBadRequest,
			expectedErrorCode:  handlers.InvalidInputErrCode,
		},
		{
			name:               "missing email",
			firstName:          "Princess",
			lastName:           "Fiona",
			password:           bonkers,
			emailExpected:      false,
			expectedErrMessage: "missing required field: email",
			expectedStatus:     http.StatusBadRequest,
			expectedErrorCode:  handlers.InvalidInputErrCode,
		},
		{
			name:               "bad password",
			email:              "pancakes@theopenlane.io",
			firstName:          "Princess",
			lastName:           "Fiona",
			password:           "asfghjkl",
			emailExpected:      false,
			expectedErrMessage: "password is too weak",
			expectedStatus:     http.StatusBadRequest,
			expectedErrorCode:  handlers.InvalidInputErrCode,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			registerJSON := models.RegisterRequest{
				FirstName: tc.firstName,
				LastName:  tc.lastName,
				Email:     tc.email,
				Password:  tc.password,
			}

			body, err := json.Marshal(registerJSON)
			if err != nil {
				require.NoError(t, err)
			}

			req := httptest.NewRequest(http.MethodPost, "/register", strings.NewReader(string(body)))
			req.Header.Set(httpsling.HeaderContentType, httpsling.ContentTypeJSONUTF8)

			// Set writer for tests that write on the response
			recorder := httptest.NewRecorder()

			// Using the ServerHTTP on echo will trigger the router and middleware
			suite.e.ServeHTTP(recorder, req)

			res := recorder.Result()
			defer res.Body.Close()

			var out *models.RegisterReply

			// parse request body
			if err := json.NewDecoder(res.Body).Decode(&out); err != nil {
				t.Error("error parsing response", err)
			}

			assert.Equal(t, tc.expectedStatus, recorder.Code)
			assert.Equal(t, tc.expectedErrorCode, out.ErrorCode)

			if tc.expectedStatus == http.StatusCreated {
				assert.Equal(t, out.Email, tc.email)
				assert.NotEmpty(t, out.Message)
				assert.NotEmpty(t, out.ID)

				// setup context to get the data back
				ctx := auth.NewTestContextWithValidUser(out.ID)

				// we haven't set the user's default org yet in the context
				// so allow the request to go through
				ctx = privacy.DecisionContext(ctx, privacy.Allow)

				// get the user and make sure things were created as expected
				u, err := suite.db.UserSetting.Query().Where(usersetting.UserID(out.ID)).WithDefaultOrg().Only(ctx)
				require.NoError(t, err)

				assert.NotEmpty(t, u.Edges.DefaultOrg)
				require.NotEmpty(t, u.Edges.DefaultOrg.ID)

				// setup context
				ctx = auth.NewTestContextWithOrgID(out.ID, u.Edges.DefaultOrg.ID)

				// make sure user is an owner of their personal org
				orgMemberships, err := suite.api.GetOrgMembersByOrgID(ctx, &openlaneclient.OrgMembershipWhereInput{
					OrganizationID: &u.Edges.DefaultOrg.ID,
				})
				require.NoError(t, err)
				require.Len(t, orgMemberships.OrgMemberships.Edges, 1)
				assert.Equal(t, orgMemberships.OrgMemberships.Edges[0].Node.Role, enums.RoleOwner)
			} else {
				assert.Contains(t, out.Error, tc.expectedErrMessage)
			}

			// wait for messages
			if tc.emailExpected {
				job := rivertest.RequireInserted[*riverpgxv5.Driver](context.Background(), t, riverpgxv5.New(suite.db.Job.GetPool()), &jobs.EmailArgs{
					Message: *newman.NewEmailMessageWithOptions(
						newman.WithSubject("Please verify your email address to login to Openlane"),
					),
				}, nil)
				require.NotNil(t, job)
				require.Equal(t, []string{tc.email}, job.Args.Message.To)
			}
		})
	}
}
