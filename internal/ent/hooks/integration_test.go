//go:build test

package hooks_test

import (
	"context"
	"testing"

	"github.com/theopenlane/core/common/enums"
	"github.com/theopenlane/core/v2/internal/ent/generated"
	emaildef "github.com/theopenlane/core/v2/internal/integrations/definitions/email"
	"github.com/theopenlane/core/v2/internal/integrations/registry"
	intruntime "github.com/theopenlane/core/v2/internal/integrations/runtime"
	"github.com/theopenlane/core/v2/internal/keystore"
	"github.com/theopenlane/core/v2/pkg/gala"
	"github.com/theopenlane/iam/auth"

	"gotest.tools/v3/assert"
	is "gotest.tools/v3/assert/cmp"
)

func TestIntegrationCampaignEmailUniquePerOrg(t *testing.T) {
	s := suite.setupTest(t)

	user := s.seedUser()
	assert.Assert(t, len(user.Edges.OrgMemberships) > 0)

	orgID := user.Edges.OrgMemberships[0].OrganizationID
	ctx := auth.NewTestContextWithOrgID(user.ID, orgID)
	ctx = generated.NewContext(ctx, s.client)
	ctx = auth.WithInternalOperationContext(ctx)

	first, err := s.client.Integration.Create().
		SetOwnerID(orgID).
		SetName("Email One").
		SetDefinitionID(emaildef.DefinitionID.ID()).
		SetStatus(enums.IntegrationStatusConnected).
		SetCampaignEmail(true).
		Save(ctx)
	assert.NilError(t, err)

	second, err := s.client.Integration.Create().
		SetOwnerID(orgID).
		SetName("Email Two").
		SetDefinitionID(emaildef.DefinitionID.ID()).
		SetStatus(enums.IntegrationStatusConnected).
		SetCampaignEmail(true).
		Save(ctx)
	assert.NilError(t, err)

	first, err = s.client.Integration.Get(ctx, first.ID)
	assert.NilError(t, err)
	second, err = s.client.Integration.Get(ctx, second.ID)
	assert.NilError(t, err)

	assert.Assert(t, !first.CampaignEmail)
	assert.Assert(t, second.CampaignEmail)

	first, err = s.client.Integration.UpdateOneID(first.ID).
		SetCampaignEmail(true).
		Save(ctx)
	assert.NilError(t, err)
	second, err = s.client.Integration.Get(ctx, second.ID)
	assert.NilError(t, err)

	assert.Assert(t, first.CampaignEmail)
	assert.Assert(t, !second.CampaignEmail)

	galaRuntime, err := gala.NewGala(context.Background(), gala.Config{
		DispatchMode: gala.DispatchModeInMemory,
		WorkerCount:  1,
	})
	assert.NilError(t, err)
	t.Cleanup(func() { galaRuntime.Close() })

	rt, err := intruntime.New(intruntime.Config{
		DB:       s.client,
		Gala:     galaRuntime,
		Registry: registry.New(),
		Keystore: &keystore.Store{},
	})
	assert.NilError(t, err)

	resolved, err := rt.ResolveOwnerIntegration(ctx, emaildef.DefinitionID.ID(), orgID, func(inst *generated.Integration) bool {
		return inst.CampaignEmail
	})
	assert.NilError(t, err)
	assert.Assert(t, is.Equal(first.ID, resolved))

	first, err = s.client.Integration.UpdateOneID(first.ID).
		SetStatus(enums.IntegrationStatusPending).
		Save(ctx)
	assert.NilError(t, err)

	resolved, err = rt.ResolveOwnerIntegration(ctx, emaildef.DefinitionID.ID(), orgID, func(inst *generated.Integration) bool {
		return inst.CampaignEmail
	})
	assert.NilError(t, err)
	assert.Assert(t, is.Equal(second.ID, resolved))
	assert.Assert(t, first.CampaignEmail)
}
