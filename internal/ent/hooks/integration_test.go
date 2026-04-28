//go:build test

package hooks_test

import (
	"github.com/stretchr/testify/require"

	"github.com/theopenlane/core/common/enums"
	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/generated/privacy"
	emaildef "github.com/theopenlane/core/internal/integrations/definitions/email"
	"github.com/theopenlane/iam/auth"
)

func (suite *HookTestSuite) TestIntegrationCampaignEmailUniquePerOrg() {
	t := suite.T()

	user := suite.seedUser()
	require.NotEmpty(t, user.Edges.OrgMemberships)

	orgID := user.Edges.OrgMemberships[0].OrganizationID
	ctx := auth.NewTestContextWithOrgID(user.ID, orgID)
	ctx = generated.NewContext(ctx, suite.client)
	ctx = privacy.DecisionContext(ctx, privacy.Allow)

	first, err := suite.client.Integration.Create().
		SetOwnerID(orgID).
		SetName("Email One").
		SetDefinitionID(emaildef.DefinitionID.ID()).
		SetStatus(enums.IntegrationStatusConnected).
		SetCampaignEmail(true).
		Save(ctx)
	require.NoError(t, err)

	second, err := suite.client.Integration.Create().
		SetOwnerID(orgID).
		SetName("Email Two").
		SetDefinitionID(emaildef.DefinitionID.ID()).
		SetStatus(enums.IntegrationStatusConnected).
		SetCampaignEmail(true).
		Save(ctx)
	require.NoError(t, err)

	first, err = suite.client.Integration.Get(ctx, first.ID)
	require.NoError(t, err)
	second, err = suite.client.Integration.Get(ctx, second.ID)
	require.NoError(t, err)

	require.False(t, first.CampaignEmail)
	require.True(t, second.CampaignEmail)

	first, err = suite.client.Integration.UpdateOneID(first.ID).
		SetCampaignEmail(true).
		Save(ctx)
	require.NoError(t, err)
	second, err = suite.client.Integration.Get(ctx, second.ID)
	require.NoError(t, err)

	require.True(t, first.CampaignEmail)
	require.False(t, second.CampaignEmail)

	resolved := emaildef.ResolveCampaignEmailIntegration(ctx, suite.client, orgID, "")
	require.Equal(t, first.ID, resolved)

	first, err = suite.client.Integration.UpdateOneID(first.ID).
		SetStatus(enums.IntegrationStatusPending).
		Save(ctx)
	require.NoError(t, err)

	resolved = emaildef.ResolveCampaignEmailIntegration(ctx, suite.client, orgID, first.ID)
	require.Equal(t, second.ID, resolved)
	require.True(t, first.CampaignEmail)
}
