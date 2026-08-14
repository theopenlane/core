//go:build test

package graphapi_test

import (
	"context"
	"strings"
	"testing"

	"gotest.tools/v3/assert"
	is "gotest.tools/v3/assert/cmp"

	"github.com/theopenlane/utils/ulids"

	"github.com/theopenlane/core/internal/ent/generated/organization"
	"github.com/theopenlane/core/internal/ent/generated/privacy"
	"github.com/theopenlane/core/internal/ent/hooks"
	"github.com/theopenlane/core/internal/graphapi"
	"github.com/theopenlane/core/internal/graphapi/testclient"
)

func TestOrganizationAvatarListener(t *testing.T) {
	setup, err := graphapi.SetupListenerRuntime(context.Background(), suite.client.db, suite.tf.URI, hooks.OrganizationAvatarListeners())
	assert.NilError(t, err)
	defer setup.Teardown()

	user := suite.userBuilder(context.Background(), t)
	allowCtx := privacy.DecisionContext(setContext(user.UserCtx, suite.client.db), privacy.Allow)

	t.Run("create without domains keeps default avatar", func(t *testing.T) {
		org := (&OrganizationBuilder{client: suite.client}).MustNew(user.UserCtx, t)
		assert.Assert(t, org.AvatarRemoteURL != nil)

		setup.Runtime.WaitIdle()

		reloaded, err := suite.client.db.Organization.Query().
			Where(organization.IDEQ(org.ID)).
			WithSetting().
			Only(allowCtx)
		assert.NilError(t, err)
		assert.Assert(t, reloaded.Edges.Setting != nil)
		assert.Check(t, is.Len(reloaded.Edges.Setting.Domains, 0))
		assert.Assert(t, reloaded.AvatarRemoteURL != nil)
		assert.Check(t, is.Equal(*org.AvatarRemoteURL, *reloaded.AvatarRemoteURL))
	})

	t.Run("unreachable domain acks without avatar update", func(t *testing.T) {
		domain := "unreachable-" + strings.ToLower(ulids.New().String()) + ".invalid"

		resp, err := suite.client.api.CreateOrganization(user.UserCtx, testclient.CreateOrganizationInput{
			Name: "avatar-listener-" + ulids.New().String(),
			CreateOrgSettings: &testclient.CreateOrganizationSettingInput{
				Domains: []string{domain},
			},
		}, nil, nil)
		assert.NilError(t, err)

		created := resp.CreateOrganization.Organization
		assert.Assert(t, created.AvatarRemoteURL != nil)

		setup.Runtime.WaitIdle()

		reloaded, err := suite.client.db.Organization.Query().
			Where(organization.IDEQ(created.ID)).
			WithSetting().
			Only(allowCtx)
		assert.NilError(t, err)
		assert.Assert(t, reloaded.Edges.Setting != nil)
		assert.Check(t, is.DeepEqual([]string{domain}, reloaded.Edges.Setting.Domains))
		assert.Assert(t, reloaded.AvatarRemoteURL != nil)
		assert.Check(t, is.Equal(*created.AvatarRemoteURL, *reloaded.AvatarRemoteURL))
	})
}
