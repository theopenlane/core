//go:build test

package hooks_test

import (
	"github.com/brianvoe/gofakeit/v7"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/theopenlane/iam/auth"

	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/generated/privacy"
)

func (suite *HookTestSuite) TestHookCustomEnums_DuplicateSystemAndOrgEnum() {
	t := suite.T()

	systemAdmin := suite.seedSystemAdmin()
	require.NotEmpty(t, systemAdmin.Edges.OrgMemberships)

	orgID := systemAdmin.Edges.OrgMemberships[0].OrganizationID

	// system admin context — used to create system-owned enums
	sysCtx := auth.NewTestContextForSystemAdmin(systemAdmin.ID, orgID)
	sysCtx = generated.NewContext(sysCtx, suite.client)

	// non-admin context — used to create org-owned enums and tasks
	userCtx := auth.NewTestContextWithOrgID(systemAdmin.ID, orgID)
	userCtx = generated.NewContext(userCtx, suite.client)

	allowSysCtx := privacy.DecisionContext(sysCtx, privacy.Allow)
	allowUserCtx := privacy.DecisionContext(userCtx, privacy.Allow)

	enumName := "TestEnum-" + gofakeit.UUID()

	// create a system-owned custom enum (no owner_id, system_owned=true)
	sysEnum, err := suite.client.CustomTypeEnum.Create().
		SetName(enumName).
		SetObjectType("task").
		SetField("kind").
		Save(allowSysCtx)
	require.NoError(t, err)
	assert.True(t, sysEnum.SystemOwned)

	// create an org-owned custom enum with the same name/field/object_type
	orgEnum, err := suite.client.CustomTypeEnum.Create().
		SetName(enumName).
		SetObjectType("task").
		SetField("kind").
		SetOwnerID(orgID).
		Save(allowUserCtx)
	require.NoError(t, err)
	assert.False(t, orgEnum.SystemOwned)

	assert.NotEqual(t, sysEnum.ID, orgEnum.ID)

	// create a task referencing the enum by name — this triggers HookCustomEnums
	// before the fix, this would fail with: "generated: custom_type_enum not singular"
	task, err := suite.client.Task.Create().
		SetTitle(gofakeit.AppName()).
		SetTaskKindName(enumName).
		SetOwnerID(orgID).
		Save(userCtx)
	require.NoError(t, err)

	// the hook should resolve to the system-owned enum
	assert.Equal(t, sysEnum.ID, task.TaskKindID)
}
