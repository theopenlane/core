package handlers_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/generated/privacy"
	"github.com/theopenlane/core/pkg/enums"
	"github.com/theopenlane/core/pkg/models"
)

func ptr[T any](v T) *T { return &v }

func (suite *HandlerTestSuite) TestWebfingerHandler() {
	t := suite.T()

	suite.e.GET(".well-known/webfinger", suite.h.WebfingerHandler)

	ctx := context.Background()
	ctx = privacy.DecisionContext(ctx, privacy.Allow)

	setting := suite.db.OrganizationSetting.Create().SetInput(generated.CreateOrganizationSettingInput{
		IdentityProviderLoginEnforced: ptr(true),
		IdentityProvider:              ptr(enums.SSOProviderOkta),
		OidcDiscoveryEndpoint:         ptr("http://example.com"),
	}).SaveX(ctx)

	org := suite.db.Organization.Create().SetInput(generated.CreateOrganizationInput{
		Name:      "testorg",
		SettingID: &setting.ID,
	}).SaveX(ctx)

	req := httptest.NewRequest(http.MethodGet, "/.well-known/webfinger?resource=org:"+org.ID, nil)
	rec := httptest.NewRecorder()

	suite.e.ServeHTTP(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)
	var out models.SSOStatusReply
	require.NoError(t, json.NewDecoder(rec.Body).Decode(&out))
	assert.True(t, out.Enforced)
}
