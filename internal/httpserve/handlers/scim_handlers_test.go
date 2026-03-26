package handlers_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"

	echo "github.com/theopenlane/echox"

	"github.com/theopenlane/core/common/enums"
	"github.com/theopenlane/core/internal/ent/generated/directorymembership"
	"github.com/theopenlane/core/internal/ent/generated/integration"
	"github.com/theopenlane/core/internal/ent/generated/privacy"
	"github.com/theopenlane/core/internal/httpserve/route"
	definitionscim "github.com/theopenlane/core/internal/integrations/definitions/scim"
	"github.com/theopenlane/core/internal/integrations/registry"
	"github.com/theopenlane/utils/ulids"
)

const (
	scimUserSchema  = "urn:ietf:params:scim:schemas:core:2.0:User"
	scimGroupSchema = "urn:ietf:params:scim:schemas:core:2.0:Group"
	scimPatchSchema = "urn:ietf:params:scim:api:messages:2.0:PatchOp"
)

func (suite *HandlerTestSuite) registerSCIMRoutesWithAuth() func() {
	suite.T().Helper()

	restore := suite.withDefinitionRuntime(suite.T(), []registry.Builder{definitionscim.Builder()})
	suite.h.AuthMiddleware = []echo.MiddlewareFunc{suite.sharedAuthMiddleware}
	suite.router.Handler = suite.h

	err := route.RegisterRoutes(suite.router)
	suite.Require().NoError(err)

	return restore
}

func (suite *HandlerTestSuite) createSCIMIntegration(ctx context.Context, orgID, name string) string {
	suite.T().Helper()

	scimInteg, err := suite.db.Integration.Create().
		SetOwnerID(orgID).
		SetName(name).
		SetDefinitionID(definitionscim.DefinitionID.ID()).
		Save(ctx)
	suite.Require().NoError(err)

	return scimInteg.ID
}

func (suite *HandlerTestSuite) createOrgAPIToken(ctx context.Context, orgID string) string {
	suite.T().Helper()

	token, err := suite.db.APIToken.Create().
		SetOwnerID(orgID).
		SetName("scim-test-token").
		SetScopes([]string{"read", "write"}).
		Save(ctx)
	suite.Require().NoError(err)

	return token.Token
}

func (suite *HandlerTestSuite) newSCIMRequest(method, path, token string, payload []byte) *http.Request {
	suite.T().Helper()

	req := httptest.NewRequest(method, path, bytes.NewReader(payload))
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/scim+json")
	req.Header.Set("Accept", "application/scim+json")

	return req
}

func (suite *HandlerTestSuite) getStringField(data map[string]any, key string) string {
	value, ok := data[key]
	suite.Require().True(ok, "missing key %s", key)

	str, ok := value.(string)
	suite.Require().True(ok, "key %s is not a string", key)

	return str
}

func (suite *HandlerTestSuite) TestSCIMUserHandlerCreateUsesBearerAPITokenAndDirectoryIngest() {
	restore := suite.registerSCIMRoutesWithAuth()
	defer restore()

	ctx := context.Background()
	testUser := suite.userBuilderWithInput(ctx, &userInput{
		email: ulids.New().String() + "@example.com",
	})

	allowCtx := privacy.DecisionContext(testUser.UserCtx, privacy.Allow)
	scimIntegrationID := suite.createSCIMIntegration(allowCtx, testUser.OrganizationID, "SCIM Directory A")
	apiToken := suite.createOrgAPIToken(allowCtx, testUser.OrganizationID)

	email := fmt.Sprintf("directory-user-%s@example.com", strings.ToLower(ulids.New().String()))
	payload, err := json.Marshal(map[string]any{
		"schemas":  []string{scimUserSchema},
		"userName": email,
		"name": map[string]any{
			"givenName":  "Directory",
			"familyName": "User",
		},
		"displayName": "Directory User",
	})
	suite.Require().NoError(err)

	req := suite.newSCIMRequest(http.MethodPost, fmt.Sprintf("/scim/%s/v2/Users", scimIntegrationID), apiToken, payload)
	rec := httptest.NewRecorder()

	suite.e.ServeHTTP(rec, req)
	suite.Require().Equal(http.StatusCreated, rec.Code, rec.Body.String())

	var resp map[string]any
	suite.Require().NoError(json.Unmarshal(rec.Body.Bytes(), &resp))

	accountID := suite.getStringField(resp, "id")
	account, err := suite.db.DirectoryAccount.Get(allowCtx, accountID)
	suite.Require().NoError(err)
	suite.Equal(scimIntegrationID, account.IntegrationID)
	suite.Equal(email, account.ExternalID)
	suite.Require().NotNil(account.CanonicalEmail)
	suite.Equal(email, *account.CanonicalEmail)
	suite.Equal("Directory User", account.DisplayName)
	suite.Equal(enums.DirectoryAccountStatusActive, account.Status)
}

func (suite *HandlerTestSuite) TestSCIMUserHandlerPatchActiveToggleUpdatesDirectoryAccount() {
	restore := suite.registerSCIMRoutesWithAuth()
	defer restore()

	ctx := context.Background()
	testUser := suite.userBuilderWithInput(ctx, &userInput{
		email: ulids.New().String() + "@example.com",
	})

	allowCtx := privacy.DecisionContext(testUser.UserCtx, privacy.Allow)
	scimIntegrationID := suite.createSCIMIntegration(allowCtx, testUser.OrganizationID, "SCIM Directory")
	apiToken := suite.createOrgAPIToken(allowCtx, testUser.OrganizationID)

	email := fmt.Sprintf("directory-toggle-%s@example.com", strings.ToLower(ulids.New().String()))
	createPayload, err := json.Marshal(map[string]any{
		"schemas":  []string{scimUserSchema},
		"userName": email,
		"name": map[string]any{
			"givenName":  "Toggle",
			"familyName": "Target",
		},
	})
	suite.Require().NoError(err)

	createReq := suite.newSCIMRequest(http.MethodPost, fmt.Sprintf("/scim/%s/v2/Users", scimIntegrationID), apiToken, createPayload)
	createRec := httptest.NewRecorder()
	suite.e.ServeHTTP(createRec, createReq)
	suite.Require().Equal(http.StatusCreated, createRec.Code, createRec.Body.String())

	var created map[string]any
	suite.Require().NoError(json.Unmarshal(createRec.Body.Bytes(), &created))
	accountID := suite.getStringField(created, "id")

	patchPayload, err := json.Marshal(map[string]any{
		"schemas": []string{scimPatchSchema},
		"Operations": []map[string]any{
			{
				"op":    "replace",
				"value": map[string]any{"active": false},
			},
		},
	})
	suite.Require().NoError(err)

	patchReq := suite.newSCIMRequest(http.MethodPatch, fmt.Sprintf("/scim/%s/v2/Users/%s", scimIntegrationID, accountID), apiToken, patchPayload)
	patchRec := httptest.NewRecorder()
	suite.e.ServeHTTP(patchRec, patchReq)
	suite.Require().Equal(http.StatusOK, patchRec.Code, patchRec.Body.String())

	var deactivated map[string]any
	suite.Require().NoError(json.Unmarshal(patchRec.Body.Bytes(), &deactivated))
	active, ok := deactivated["active"].(bool)
	suite.Require().True(ok)
	suite.False(active)

	account, err := suite.db.DirectoryAccount.Get(allowCtx, accountID)
	suite.Require().NoError(err)
	suite.Equal(enums.DirectoryAccountStatusInactive, account.Status)

	patchPayload, err = json.Marshal(map[string]any{
		"schemas": []string{scimPatchSchema},
		"Operations": []map[string]any{
			{
				"op":    "replace",
				"value": map[string]any{"active": true},
			},
		},
	})
	suite.Require().NoError(err)

	patchReq = suite.newSCIMRequest(http.MethodPatch, fmt.Sprintf("/scim/%s/v2/Users/%s", scimIntegrationID, accountID), apiToken, patchPayload)
	patchRec = httptest.NewRecorder()
	suite.e.ServeHTTP(patchRec, patchReq)
	suite.Require().Equal(http.StatusOK, patchRec.Code, patchRec.Body.String())

	var reactivated map[string]any
	suite.Require().NoError(json.Unmarshal(patchRec.Body.Bytes(), &reactivated))
	active, ok = reactivated["active"].(bool)
	suite.Require().True(ok)
	suite.True(active)

	account, err = suite.db.DirectoryAccount.Get(allowCtx, accountID)
	suite.Require().NoError(err)
	suite.Equal(enums.DirectoryAccountStatusActive, account.Status)
}

func (suite *HandlerTestSuite) TestSCIMGroupHandlerCreateDeduplicatesMembers() {
	restore := suite.registerSCIMRoutesWithAuth()
	defer restore()

	ctx := context.Background()
	testUser := suite.userBuilderWithInput(ctx, &userInput{
		email: ulids.New().String() + "@example.com",
	})

	allowCtx := privacy.DecisionContext(testUser.UserCtx, privacy.Allow)
	scimIntegrationID := suite.createSCIMIntegration(allowCtx, testUser.OrganizationID, "SCIM Directory")
	apiToken := suite.createOrgAPIToken(allowCtx, testUser.OrganizationID)

	memberPayload, err := json.Marshal(map[string]any{
		"schemas":  []string{scimUserSchema},
		"userName": fmt.Sprintf("member-%s@example.com", strings.ToLower(ulids.New().String())),
		"name": map[string]any{
			"givenName":  "Member",
			"familyName": "Target",
		},
	})
	suite.Require().NoError(err)

	memberReq := suite.newSCIMRequest(http.MethodPost, fmt.Sprintf("/scim/%s/v2/Users", scimIntegrationID), apiToken, memberPayload)
	memberRec := httptest.NewRecorder()
	suite.e.ServeHTTP(memberRec, memberReq)
	suite.Require().Equal(http.StatusCreated, memberRec.Code, memberRec.Body.String())

	var createdUser map[string]any
	suite.Require().NoError(json.Unmarshal(memberRec.Body.Bytes(), &createdUser))
	accountID := suite.getStringField(createdUser, "id")

	groupPayload, err := json.Marshal(map[string]any{
		"schemas":     []string{scimGroupSchema},
		"displayName": "SCIM Directory Team",
		"members": []map[string]any{
			{"value": accountID},
			{"value": accountID},
		},
	})
	suite.Require().NoError(err)

	groupReq := suite.newSCIMRequest(http.MethodPost, fmt.Sprintf("/scim/%s/v2/Groups", scimIntegrationID), apiToken, groupPayload)
	groupRec := httptest.NewRecorder()
	suite.e.ServeHTTP(groupRec, groupReq)
	suite.Require().Equal(http.StatusCreated, groupRec.Code, groupRec.Body.String())

	var resp map[string]any
	suite.Require().NoError(json.Unmarshal(groupRec.Body.Bytes(), &resp))

	groupID := suite.getStringField(resp, "id")
	group, err := suite.db.DirectoryGroup.Get(allowCtx, groupID)
	suite.Require().NoError(err)
	suite.Equal(scimIntegrationID, group.IntegrationID)
	suite.Equal("SCIM Directory Team", group.ExternalID)
	suite.Equal("SCIM Directory Team", group.DisplayName)

	membersValue, ok := resp["members"].([]any)
	suite.Require().True(ok)
	suite.Len(membersValue, 1)

	membershipRows, err := suite.db.DirectoryMembership.Query().
		Where(directorymembership.DirectoryGroupID(groupID)).
		All(allowCtx)
	suite.Require().NoError(err)
	suite.Len(membershipRows, 1)
	suite.Equal(scimIntegrationID, membershipRows[0].IntegrationID)
	suite.Equal(accountID, membershipRows[0].DirectoryAccountID)
}

func (suite *HandlerTestSuite) TestSCIMRouteScopesDirectoryRecordsByIntegration() {
	restore := suite.registerSCIMRoutesWithAuth()
	defer restore()

	ctx := context.Background()
	testUser := suite.userBuilderWithInput(ctx, &userInput{
		email: ulids.New().String() + "@example.com",
	})

	allowCtx := privacy.DecisionContext(testUser.UserCtx, privacy.Allow)
	integrationA := suite.createSCIMIntegration(allowCtx, testUser.OrganizationID, "SCIM Directory A")
	integrationB := suite.createSCIMIntegration(allowCtx, testUser.OrganizationID, "SCIM Directory B")
	apiToken := suite.createOrgAPIToken(allowCtx, testUser.OrganizationID)

	userEmailA := fmt.Sprintf("scoped-a-%s@example.com", strings.ToLower(ulids.New().String()))
	createPayloadA, err := json.Marshal(map[string]any{
		"schemas":  []string{scimUserSchema},
		"userName": userEmailA,
		"name": map[string]any{
			"givenName":  "Scoped",
			"familyName": "Alpha",
		},
		"displayName": "Scoped Alpha",
	})
	suite.Require().NoError(err)

	createReqA := suite.newSCIMRequest(http.MethodPost, fmt.Sprintf("/scim/%s/v2/Users", integrationA), apiToken, createPayloadA)
	createRecA := httptest.NewRecorder()
	suite.e.ServeHTTP(createRecA, createReqA)
	suite.Require().Equal(http.StatusCreated, createRecA.Code, createRecA.Body.String())

	var createdUserA map[string]any
	suite.Require().NoError(json.Unmarshal(createRecA.Body.Bytes(), &createdUserA))
	accountIDA := suite.getStringField(createdUserA, "id")

	userEmailB := fmt.Sprintf("scoped-b-%s@example.com", strings.ToLower(ulids.New().String()))
	createPayloadB, err := json.Marshal(map[string]any{
		"schemas":  []string{scimUserSchema},
		"userName": userEmailB,
		"name": map[string]any{
			"givenName":  "Scoped",
			"familyName": "Beta",
		},
		"displayName": "Scoped Beta",
	})
	suite.Require().NoError(err)

	createReqB := suite.newSCIMRequest(http.MethodPost, fmt.Sprintf("/scim/%s/v2/Users", integrationB), apiToken, createPayloadB)
	createRecB := httptest.NewRecorder()
	suite.e.ServeHTTP(createRecB, createReqB)
	suite.Require().Equal(http.StatusCreated, createRecB.Code, createRecB.Body.String())

	var createdUserB map[string]any
	suite.Require().NoError(json.Unmarshal(createRecB.Body.Bytes(), &createdUserB))
	accountIDB := suite.getStringField(createdUserB, "id")

	accountA, err := suite.db.DirectoryAccount.Get(allowCtx, accountIDA)
	suite.Require().NoError(err)
	suite.Equal(integrationA, accountA.IntegrationID)
	suite.Equal(userEmailA, accountA.ExternalID)
	suite.Require().NotNil(accountA.CanonicalEmail)
	suite.Equal(userEmailA, *accountA.CanonicalEmail)
	suite.Equal("Scoped Alpha", accountA.DisplayName)
	suite.Equal(enums.DirectoryAccountStatusActive, accountA.Status)

	accountB, err := suite.db.DirectoryAccount.Get(allowCtx, accountIDB)
	suite.Require().NoError(err)
	suite.Equal(integrationB, accountB.IntegrationID)
	suite.Equal(userEmailB, accountB.ExternalID)
	suite.Require().NotNil(accountB.CanonicalEmail)
	suite.Equal(userEmailB, *accountB.CanonicalEmail)
	suite.Equal("Scoped Beta", accountB.DisplayName)
	suite.Equal(enums.DirectoryAccountStatusActive, accountB.Status)

	groupPayloadA, err := json.Marshal(map[string]any{
		"schemas":     []string{scimGroupSchema},
		"displayName": "Scoped Team Alpha",
		"members": []map[string]any{
			{"value": accountIDA},
		},
	})
	suite.Require().NoError(err)

	groupReqA := suite.newSCIMRequest(http.MethodPost, fmt.Sprintf("/scim/%s/v2/Groups", integrationA), apiToken, groupPayloadA)
	groupRecA := httptest.NewRecorder()
	suite.e.ServeHTTP(groupRecA, groupReqA)
	suite.Require().Equal(http.StatusCreated, groupRecA.Code, groupRecA.Body.String())

	var createdGroupA map[string]any
	suite.Require().NoError(json.Unmarshal(groupRecA.Body.Bytes(), &createdGroupA))
	groupIDA := suite.getStringField(createdGroupA, "id")

	groupPayloadB, err := json.Marshal(map[string]any{
		"schemas":     []string{scimGroupSchema},
		"displayName": "Scoped Team Beta",
		"members": []map[string]any{
			{"value": accountIDB},
		},
	})
	suite.Require().NoError(err)

	groupReqB := suite.newSCIMRequest(http.MethodPost, fmt.Sprintf("/scim/%s/v2/Groups", integrationB), apiToken, groupPayloadB)
	groupRecB := httptest.NewRecorder()
	suite.e.ServeHTTP(groupRecB, groupReqB)
	suite.Require().Equal(http.StatusCreated, groupRecB.Code, groupRecB.Body.String())

	var createdGroupB map[string]any
	suite.Require().NoError(json.Unmarshal(groupRecB.Body.Bytes(), &createdGroupB))
	groupIDB := suite.getStringField(createdGroupB, "id")

	groupA, err := suite.db.DirectoryGroup.Get(allowCtx, groupIDA)
	suite.Require().NoError(err)
	suite.Equal(integrationA, groupA.IntegrationID)
	suite.Equal("Scoped Team Alpha", groupA.ExternalID)
	suite.Equal("Scoped Team Alpha", groupA.DisplayName)
	suite.Equal(enums.DirectoryGroupStatusActive, groupA.Status)

	groupB, err := suite.db.DirectoryGroup.Get(allowCtx, groupIDB)
	suite.Require().NoError(err)
	suite.Equal(integrationB, groupB.IntegrationID)
	suite.Equal("Scoped Team Beta", groupB.ExternalID)
	suite.Equal("Scoped Team Beta", groupB.DisplayName)
	suite.Equal(enums.DirectoryGroupStatusActive, groupB.Status)

	membershipRowsA, err := suite.db.DirectoryMembership.Query().
		Where(directorymembership.DirectoryGroupID(groupIDA)).
		All(allowCtx)
	suite.Require().NoError(err)
	suite.Len(membershipRowsA, 1)
	suite.Equal(integrationA, membershipRowsA[0].IntegrationID)
	suite.Equal(accountIDA, membershipRowsA[0].DirectoryAccountID)

	membershipRowsB, err := suite.db.DirectoryMembership.Query().
		Where(directorymembership.DirectoryGroupID(groupIDB)).
		All(allowCtx)
	suite.Require().NoError(err)
	suite.Len(membershipRowsB, 1)
	suite.Equal(integrationB, membershipRowsB[0].IntegrationID)
	suite.Equal(accountIDB, membershipRowsB[0].DirectoryAccountID)

	getUserReqA := suite.newSCIMRequest(http.MethodGet, fmt.Sprintf("/scim/%s/v2/Users/%s", integrationA, accountIDA), apiToken, nil)
	getUserRecA := httptest.NewRecorder()
	suite.e.ServeHTTP(getUserRecA, getUserReqA)
	suite.Require().Equal(http.StatusOK, getUserRecA.Code, getUserRecA.Body.String())

	var fetchedUserA map[string]any
	suite.Require().NoError(json.Unmarshal(getUserRecA.Body.Bytes(), &fetchedUserA))
	suite.Equal(userEmailA, suite.getStringField(fetchedUserA, "userName"))
	suite.Equal("Scoped Alpha", suite.getStringField(fetchedUserA, "displayName"))
	nameA, ok := fetchedUserA["name"].(map[string]any)
	suite.Require().True(ok)
	suite.Equal("Scoped", suite.getStringField(nameA, "givenName"))
	suite.Equal("Alpha", suite.getStringField(nameA, "familyName"))
	activeA, ok := fetchedUserA["active"].(bool)
	suite.Require().True(ok)
	suite.True(activeA)

	getUserReqB := suite.newSCIMRequest(http.MethodGet, fmt.Sprintf("/scim/%s/v2/Users/%s", integrationB, accountIDB), apiToken, nil)
	getUserRecB := httptest.NewRecorder()
	suite.e.ServeHTTP(getUserRecB, getUserReqB)
	suite.Require().Equal(http.StatusOK, getUserRecB.Code, getUserRecB.Body.String())

	var fetchedUserB map[string]any
	suite.Require().NoError(json.Unmarshal(getUserRecB.Body.Bytes(), &fetchedUserB))
	suite.Equal(userEmailB, suite.getStringField(fetchedUserB, "userName"))
	suite.Equal("Scoped Beta", suite.getStringField(fetchedUserB, "displayName"))
	nameB, ok := fetchedUserB["name"].(map[string]any)
	suite.Require().True(ok)
	suite.Equal("Scoped", suite.getStringField(nameB, "givenName"))
	suite.Equal("Beta", suite.getStringField(nameB, "familyName"))
	activeB, ok := fetchedUserB["active"].(bool)
	suite.Require().True(ok)
	suite.True(activeB)

	getGroupReqA := suite.newSCIMRequest(http.MethodGet, fmt.Sprintf("/scim/%s/v2/Groups/%s", integrationA, groupIDA), apiToken, nil)
	getGroupRecA := httptest.NewRecorder()
	suite.e.ServeHTTP(getGroupRecA, getGroupReqA)
	suite.Require().Equal(http.StatusOK, getGroupRecA.Code, getGroupRecA.Body.String())

	var fetchedGroupA map[string]any
	suite.Require().NoError(json.Unmarshal(getGroupRecA.Body.Bytes(), &fetchedGroupA))
	suite.Equal("Scoped Team Alpha", suite.getStringField(fetchedGroupA, "displayName"))
	groupMembersA, ok := fetchedGroupA["members"].([]any)
	suite.Require().True(ok)
	suite.Len(groupMembersA, 1)
	groupMemberA, ok := groupMembersA[0].(map[string]any)
	suite.Require().True(ok)
	suite.Equal(accountIDA, suite.getStringField(groupMemberA, "value"))

	getGroupReqB := suite.newSCIMRequest(http.MethodGet, fmt.Sprintf("/scim/%s/v2/Groups/%s", integrationB, groupIDB), apiToken, nil)
	getGroupRecB := httptest.NewRecorder()
	suite.e.ServeHTTP(getGroupRecB, getGroupReqB)
	suite.Require().Equal(http.StatusOK, getGroupRecB.Code, getGroupRecB.Body.String())

	var fetchedGroupB map[string]any
	suite.Require().NoError(json.Unmarshal(getGroupRecB.Body.Bytes(), &fetchedGroupB))
	suite.Equal("Scoped Team Beta", suite.getStringField(fetchedGroupB, "displayName"))
	groupMembersB, ok := fetchedGroupB["members"].([]any)
	suite.Require().True(ok)
	suite.Len(groupMembersB, 1)
	groupMemberB, ok := groupMembersB[0].(map[string]any)
	suite.Require().True(ok)
	suite.Equal(accountIDB, suite.getStringField(groupMemberB, "value"))

	getReq := suite.newSCIMRequest(http.MethodGet, fmt.Sprintf("/scim/%s/v2/Users/%s", integrationB, accountIDA), apiToken, nil)
	getRec := httptest.NewRecorder()
	suite.e.ServeHTTP(getRec, getReq)
	suite.Require().Equal(http.StatusNotFound, getRec.Code, getRec.Body.String())

	crossGroupReq := suite.newSCIMRequest(http.MethodGet, fmt.Sprintf("/scim/%s/v2/Groups/%s", integrationA, groupIDB), apiToken, nil)
	crossGroupRec := httptest.NewRecorder()
	suite.e.ServeHTTP(crossGroupRec, crossGroupReq)
	suite.Require().Equal(http.StatusNotFound, crossGroupRec.Code, crossGroupRec.Body.String())

	listUsersReqA := suite.newSCIMRequest(http.MethodGet, fmt.Sprintf("/scim/%s/v2/Users", integrationA), apiToken, nil)
	listUsersRecA := httptest.NewRecorder()
	suite.e.ServeHTTP(listUsersRecA, listUsersReqA)
	suite.Require().Equal(http.StatusOK, listUsersRecA.Code, listUsersRecA.Body.String())

	var userPageA map[string]any
	suite.Require().NoError(json.Unmarshal(listUsersRecA.Body.Bytes(), &userPageA))
	totalResults, ok := userPageA["totalResults"].(float64)
	suite.Require().True(ok)
	suite.Equal(float64(1), totalResults)
	resourcesA, ok := userPageA["Resources"].([]any)
	suite.Require().True(ok)
	suite.Len(resourcesA, 1)

	listUsersReqB := suite.newSCIMRequest(http.MethodGet, fmt.Sprintf("/scim/%s/v2/Users", integrationB), apiToken, nil)
	listUsersRecB := httptest.NewRecorder()
	suite.e.ServeHTTP(listUsersRecB, listUsersReqB)
	suite.Require().Equal(http.StatusOK, listUsersRecB.Code, listUsersRecB.Body.String())

	var userPageB map[string]any
	suite.Require().NoError(json.Unmarshal(listUsersRecB.Body.Bytes(), &userPageB))
	totalResults, ok = userPageB["totalResults"].(float64)
	suite.Require().True(ok)
	suite.Equal(float64(1), totalResults)
	resourcesB, ok := userPageB["Resources"].([]any)
	suite.Require().True(ok)
	suite.Len(resourcesB, 1)

	listGroupsReqA := suite.newSCIMRequest(http.MethodGet, fmt.Sprintf("/scim/%s/v2/Groups", integrationA), apiToken, nil)
	listGroupsRecA := httptest.NewRecorder()
	suite.e.ServeHTTP(listGroupsRecA, listGroupsReqA)
	suite.Require().Equal(http.StatusOK, listGroupsRecA.Code, listGroupsRecA.Body.String())

	var groupPageA map[string]any
	suite.Require().NoError(json.Unmarshal(listGroupsRecA.Body.Bytes(), &groupPageA))
	totalResults, ok = groupPageA["totalResults"].(float64)
	suite.Require().True(ok)
	suite.Equal(float64(1), totalResults)
	groupResourcesA, ok := groupPageA["Resources"].([]any)
	suite.Require().True(ok)
	suite.Len(groupResourcesA, 1)

	listGroupsReqB := suite.newSCIMRequest(http.MethodGet, fmt.Sprintf("/scim/%s/v2/Groups", integrationB), apiToken, nil)
	listGroupsRecB := httptest.NewRecorder()
	suite.e.ServeHTTP(listGroupsRecB, listGroupsReqB)
	suite.Require().Equal(http.StatusOK, listGroupsRecB.Code, listGroupsRecB.Body.String())

	var groupPageB map[string]any
	suite.Require().NoError(json.Unmarshal(listGroupsRecB.Body.Bytes(), &groupPageB))
	totalResults, ok = groupPageB["totalResults"].(float64)
	suite.Require().True(ok)
	suite.Equal(float64(1), totalResults)
	groupResourcesB, ok := groupPageB["Resources"].([]any)
	suite.Require().True(ok)
	suite.Len(groupResourcesB, 1)
}

func (suite *HandlerTestSuite) TestSCIMRouteRejectsNonSCIMIntegration() {
	restore := suite.registerSCIMRoutesWithAuth()
	defer restore()

	ctx := context.Background()
	testUser := suite.userBuilderWithInput(ctx, &userInput{
		email: ulids.New().String() + "@example.com",
	})

	allowCtx := privacy.DecisionContext(testUser.UserCtx, privacy.Allow)
	nonSCIM, err := suite.db.Integration.Create().
		SetOwnerID(testUser.OrganizationID).
		SetName("Not SCIM").
		SetKind("github").
		Save(allowCtx)
	suite.Require().NoError(err)

	apiToken := suite.createOrgAPIToken(allowCtx, testUser.OrganizationID)
	req := suite.newSCIMRequest(http.MethodGet, fmt.Sprintf("/scim/%s/v2/Users", nonSCIM.ID), apiToken, nil)
	rec := httptest.NewRecorder()

	suite.e.ServeHTTP(rec, req)
	suite.Require().Equal(http.StatusNotFound, rec.Code, rec.Body.String())

	exists, err := suite.db.Integration.Query().
		Where(integration.ID(nonSCIM.ID)).
		Exist(allowCtx)
	suite.Require().NoError(err)
	suite.True(exists)
}
