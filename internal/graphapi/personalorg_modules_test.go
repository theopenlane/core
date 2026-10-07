package graphapi_test

import (
	"context"
	"testing"

	"github.com/theopenlane/iam/auth"
	"gotest.tools/v3/assert"
	is "gotest.tools/v3/assert/cmp"

	"github.com/theopenlane/core/common/enums"
	"github.com/theopenlane/core/common/models"
	"github.com/theopenlane/core/v2/internal/ent/interceptors"
	"github.com/theopenlane/core/v2/internal/graphapi/testclient"
	th "github.com/theopenlane/core/v2/internal/graphapi/testharness"
)

func personalOrgOwnerCtx(t *testing.T) context.Context {
	t.Helper()

	user := suite.UserBuilder(context.Background(), t)

	return th.SetUserContext(auth.NewTestContextWithOrgID(user.ID, user.PersonalOrgID, auth.WithOrganizationRole(auth.OwnerRole)), suite.Client.DB)
}

func TestPersonalOrgCannotCreateModuleGatedObjects(t *testing.T) {
	t.Parallel()

	ctx := personalOrgOwnerCtx(t)

	t.Run("task", func(t *testing.T) {
		_, err := suite.Client.API.CreateTask(ctx, testclient.CreateTaskInput{Title: "personal org task"})
		assert.ErrorContains(t, err, th.NotFoundErrorMsg)
	})

	t.Run("workflow definition", func(t *testing.T) {
		_, err := suite.Client.API.CreateWorkflowDefinition(ctx, testclient.CreateWorkflowDefinitionInput{
			Name:         "personal org workflow",
			WorkflowKind: enums.WorkflowKindNotification,
			SchemaType:   "Control",
			DefinitionJSON: &models.WorkflowDefinitionDocument{
				Triggers: []models.WorkflowTrigger{
					{Operation: "UPDATE", ObjectType: enums.WorkflowObjectTypeControl},
				},
				Actions: []models.WorkflowAction{
					{Key: "notify", Type: string(enums.WorkflowActionTypeNotification)},
				},
			},
		})
		assert.ErrorContains(t, err, th.NotFoundErrorMsg)
	})

	t.Run("tag definition", func(t *testing.T) {
		_, err := suite.Client.API.CreateTagDefinition(ctx, testclient.CreateTagDefinitionInput{Name: "personal-org-tag"})
		assert.ErrorContains(t, err, th.NotFoundErrorMsg)
	})

	t.Run("custom type enum", func(t *testing.T) {
		_, err := suite.Client.API.CreateCustomTypeEnum(ctx, testclient.CreateCustomTypeEnumInput{Name: "personal-org-enum", ObjectType: "task"})
		assert.ErrorContains(t, err, th.NotFoundErrorMsg)
	})

	t.Run("api token", func(t *testing.T) {
		_, err := suite.Client.API.CreateAPIToken(ctx, testclient.CreateAPITokenInput{Name: "personal-org-token"})
		assert.ErrorContains(t, err, th.NotFoundErrorMsg)
	})

	t.Run("personal access token", func(t *testing.T) {
		_, err := suite.Client.API.CreatePersonalAccessToken(ctx, testclient.CreatePersonalAccessTokenInput{Name: "personal-org-pat"})
		assert.ErrorContains(t, err, th.NotFoundErrorMsg)
	})
}

func TestPersonalOrgCannotQueryModuleGatedObjects(t *testing.T) {
	t.Parallel()

	ctx := personalOrgOwnerCtx(t)
	featureErr := interceptors.ErrFeatureNotEnabled.Error()

	t.Run("tasks", func(t *testing.T) {
		_, err := suite.Client.API.GetAllTasks(ctx, nil, nil, nil, nil, nil)
		assert.ErrorContains(t, err, featureErr)
	})

	t.Run("integrations", func(t *testing.T) {
		_, err := suite.Client.API.GetAllIntegrations(ctx)
		assert.ErrorContains(t, err, featureErr)
	})

	t.Run("tag definitions", func(t *testing.T) {
		_, err := suite.Client.API.GetAllTagDefinitions(ctx)
		assert.ErrorContains(t, err, featureErr)
	})

	t.Run("api tokens", func(t *testing.T) {
		_, err := suite.Client.API.GetAllAPITokens(ctx)
		assert.ErrorContains(t, err, featureErr)
	})

	t.Run("exports", func(t *testing.T) {
		_, err := suite.Client.API.GetAllExports(ctx)
		assert.ErrorContains(t, err, featureErr)
	})
}

func TestPersonalOrgCanQueryBaseObjects(t *testing.T) {
	t.Parallel()

	ctx := personalOrgOwnerCtx(t)

	resp, err := suite.Client.API.GetAllOrganizationSettings(ctx)
	assert.NilError(t, err)
	assert.Check(t, is.Len(resp.OrganizationSettings.Edges, 1))
}
