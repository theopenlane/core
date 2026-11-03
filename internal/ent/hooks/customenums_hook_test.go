//go:build test

package hooks_test

import (
	"context"
	"testing"

	"github.com/brianvoe/gofakeit/v7"
	"github.com/theopenlane/iam/auth"

	"github.com/theopenlane/core/common/models"
	"github.com/theopenlane/core/v2/internal/ent/generated"
	"github.com/theopenlane/core/v2/pkg/entitlements"

	"gotest.tools/v3/assert"
	is "gotest.tools/v3/assert/cmp"
)

func TestHookCustomEnums_DuplicateSystemAndOrgEnum(t *testing.T) {
	s := suite.setupTest(t)

	systemAdmin := s.seedSystemAdmin()
	assert.Assert(t, len(systemAdmin.Edges.OrgMemberships) > 0)

	orgID := systemAdmin.Edges.OrgMemberships[0].OrganizationID

	err := entitlements.CreateFeatureTuples(context.Background(), &s.client.Authz, orgID,
		[]models.OrgModule{models.CatalogBaseModule, models.CatalogComplianceModule})
	assert.NilError(t, err)

	// system admin context — used to create system-owned enums
	sysCtx := auth.NewTestContextForSystemAdmin(systemAdmin.ID, orgID)
	sysCtx = generated.NewContext(sysCtx, s.client)

	// non-admin context — used to create org-owned enums and tasks
	userCtx := auth.NewTestContextWithOrgID(systemAdmin.ID, orgID)
	userCtx = generated.NewContext(userCtx, s.client)

	internalSysCtx := auth.WithInternalOperationContext(sysCtx)
	internalUserCtx := auth.WithInternalOperationContext(userCtx)

	enumName := "TestEnum-" + gofakeit.UUID()

	// create a system-owned custom enum (no owner_id, system_owned=true)
	sysEnum, err := s.client.CustomTypeEnum.Create().
		SetName(enumName).
		SetObjectType("task").
		SetField("kind").
		Save(internalSysCtx)
	assert.NilError(t, err)
	assert.Check(t, sysEnum.SystemOwned)

	// create an org-owned custom enum with the same name/field/object_type
	orgEnum, err := s.client.CustomTypeEnum.Create().
		SetName(enumName).
		SetObjectType("task").
		SetField("kind").
		SetOwnerID(orgID).
		Save(internalUserCtx)
	assert.NilError(t, err)
	assert.Check(t, !orgEnum.SystemOwned)

	assert.Check(t, sysEnum.ID != orgEnum.ID)

	// create a task referencing the enum by name — this triggers HookCustomEnums
	// before the fix, this would fail with: "generated: custom_type_enum not singular"
	task, err := s.client.Task.Create().
		SetTitle(gofakeit.AppName()).
		SetTaskKindName(enumName).
		SetOwnerID(orgID).
		Save(userCtx)
	assert.NilError(t, err)

	// the hook should resolve to the system-owned enum
	assert.Check(t, is.Equal(sysEnum.ID, task.TaskKindID))
}
