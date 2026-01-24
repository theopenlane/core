package graphapi_test

import (
	"context"
	"testing"

	"gotest.tools/v3/assert"
	is "gotest.tools/v3/assert/cmp"

	"github.com/samber/lo"
	"github.com/theopenlane/utils/ulids"

	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/graphapi/testclient"
)

func TestQueryCustomTypeEnum(t *testing.T) {

	// create an customTypeEnum to be queried using testUser1
	customTypeEnum := (&CustomTypeEnumBuilder{client: suite.client}).MustNew(testUser1.UserCtx, t)

	systemOwnedEnum := (&CustomTypeEnumBuilder{
		client:      suite.client,
		Name:        "Preventative",
		ObjectType:  "control",
		Description: "A system owned enum",
		Color:       "#ff0000",
	}).MustNew(systemAdminUser.UserCtx, t)

	// add test cases for querying the CustomTypeEnum
	testCases := []struct {
		name     string
		queryID  string
		client   *testclient.TestClient
		ctx      context.Context
		errorMsg string
	}{
		{
			name:    "happy path",
			queryID: customTypeEnum.ID,
			client:  suite.client.api,
			ctx:     testUser1.UserCtx,
		},
		{
			name:    "happy path, system owned",
			queryID: systemOwnedEnum.ID,
			client:  suite.client.api,
			ctx:     testUser1.UserCtx,
		},
		{
			name:    "happy path, read only user",
			queryID: customTypeEnum.ID,
			client:  suite.client.api,
			ctx:     viewOnlyUser.UserCtx,
		},
		{
			name:    "happy path using personal access token",
			queryID: customTypeEnum.ID,
			client:  suite.client.apiWithPAT,
			ctx:     context.Background(),
		},
		{
			name:     "not found, invalid ID",
			queryID:  "invalid",
			client:   suite.client.api,
			ctx:      testUser1.UserCtx,
			errorMsg: notFoundErrorMsg,
		},
		{
			name:     "not found, using not authorized user",
			queryID:  customTypeEnum.ID,
			client:   suite.client.api,
			ctx:      testUser2.UserCtx,
			errorMsg: notFoundErrorMsg,
		},
	}

	for _, tc := range testCases {
		t.Run("Get "+tc.name, func(t *testing.T) {
			resp, err := tc.client.GetCustomTypeEnumByID(tc.ctx, tc.queryID)

			if tc.errorMsg != "" {
				assert.ErrorContains(t, err, tc.errorMsg)

				return
			}

			assert.NilError(t, err)
			assert.Assert(t, resp != nil)

			assert.Check(t, is.Equal(tc.queryID, resp.CustomTypeEnum.ID))

			// add additional assertions for the object
			assert.Check(t, resp.CustomTypeEnum.Name != "")
			assert.Check(t, resp.CustomTypeEnum.Color != nil)

			if resp.CustomTypeEnum.ID == customTypeEnum.ID {
				assert.Check(t, *resp.CustomTypeEnum.SystemOwned == false)
			} else if resp.CustomTypeEnum.ID == systemOwnedEnum.ID {
				assert.Check(t, *resp.CustomTypeEnum.SystemOwned == true)
			}
		})
	}

	(&Cleanup[*generated.CustomTypeEnumDeleteOne]{client: suite.client.db.CustomTypeEnum, ID: customTypeEnum.ID}).MustDelete(testUser1.UserCtx, t)
	(&Cleanup[*generated.CustomTypeEnumDeleteOne]{client: suite.client.db.CustomTypeEnum, ID: systemOwnedEnum.ID}).MustDelete(systemAdminUser.UserCtx, t)
}

func TestMutationCreateCustomTypeEnum(t *testing.T) {
	testCases := []struct {
		name        string
		request     testclient.CreateCustomTypeEnumInput
		client      *testclient.TestClient
		ctx         context.Context
		expectedErr string
	}{
		{
			name: "happy path, minimal input",
			request: testclient.CreateCustomTypeEnumInput{
				Name:       "Preventative",
				ObjectType: "control",
			},
			client: suite.client.api,
			ctx:    testUser1.UserCtx,
		},
		{
			name: "happy path, all input",
			request: testclient.CreateCustomTypeEnumInput{
				Name:        "Detective",
				ObjectType:  "control",
				Description: lo.ToPtr("A detective control is designed to detect threats instead of proactively preventing them."),
				Color:       lo.ToPtr("#00ff00"),
			},
			client: suite.client.api,
			ctx:    testUser1.UserCtx,
		},
		{
			name: "happy path, using pat",
			request: testclient.CreateCustomTypeEnumInput{
				Name:       "Evidence",
				ObjectType: "task",
			},
			client: suite.client.apiWithPAT,
			ctx:    context.Background(),
		},
		{
			name: "happy path, using api token",
			request: testclient.CreateCustomTypeEnumInput{
				Name:       "JustDoIt",
				ObjectType: "task",
			},
			client: suite.client.apiWithToken,
			ctx:    context.Background(),
		},
		{
			name: "user not authorized, not enough permissions",
			request: testclient.CreateCustomTypeEnumInput{
				Name:       "Operational",
				ObjectType: "risk",
			},
			client:      suite.client.api,
			ctx:         viewOnlyUser.UserCtx,
			expectedErr: notAuthorizedErrorMsg,
		},
		{
			name: "missing required field, object type",
			request: testclient.CreateCustomTypeEnumInput{
				Name: "missing type",
			},
			client:      suite.client.api,
			ctx:         testUser1.UserCtx,
			expectedErr: "invalid global enum field",
		},
		{
			name: "missing required field, name",
			request: testclient.CreateCustomTypeEnumInput{
				ObjectType: "task",
			},
			client:      suite.client.api,
			ctx:         testUser1.UserCtx,
			expectedErr: "value is less than the required length",
		},
	}

	for _, tc := range testCases {
		t.Run("Create "+tc.name, func(t *testing.T) {
			resp, err := tc.client.CreateCustomTypeEnum(tc.ctx, tc.request)
			if tc.expectedErr != "" {
				assert.ErrorContains(t, err, tc.expectedErr)

				return
			}

			assert.NilError(t, err)
			assert.Assert(t, resp != nil)

			// check required fields
			assert.Check(t, resp.CreateCustomTypeEnum.CustomTypeEnum.Name == tc.request.Name)
			assert.Check(t, resp.CreateCustomTypeEnum.CustomTypeEnum.ObjectType == tc.request.ObjectType)

			if tc.request.Description != nil {
				assert.Check(t, resp.CreateCustomTypeEnum.CustomTypeEnum.Description != nil)
				assert.Check(t, *resp.CreateCustomTypeEnum.CustomTypeEnum.Description == *tc.request.Description)
			} else {
				assert.Check(t, *resp.CreateCustomTypeEnum.CustomTypeEnum.Description == "")
			}

			// we set a color by default if none is provided
			assert.Check(t, resp.CreateCustomTypeEnum.CustomTypeEnum.Color != nil)
			if tc.request.Color != nil {
				assert.Check(t, *resp.CreateCustomTypeEnum.CustomTypeEnum.Color == *tc.request.Color)
			}

			// cleanup each object created
			(&Cleanup[*generated.CustomTypeEnumDeleteOne]{client: suite.client.db.CustomTypeEnum, ID: resp.CreateCustomTypeEnum.CustomTypeEnum.ID}).MustDelete(testUser1.UserCtx, t)
		})
	}
}

func TestMutationUpdateCustomTypeEnum(t *testing.T) {
	customTypeEnum := (&CustomTypeEnumBuilder{client: suite.client}).MustNew(testUser1.UserCtx, t)
	systemTypeEnum := (&CustomTypeEnumBuilder{
		client:      suite.client,
		Name:        "SystemEnum",
		ObjectType:  "control",
		Description: "A system owned enum",
		Color:       "#123456",
	}).MustNew(systemAdminUser.UserCtx, t)

	testCases := []struct {
		name        string
		request     testclient.UpdateCustomTypeEnumInput
		requestID   string
		client      *testclient.TestClient
		ctx         context.Context
		expectedErr string
	}{
		{
			name: "happy path, update field",
			request: testclient.UpdateCustomTypeEnumInput{
				Description: lo.ToPtr("Updated description"),
			},
			requestID: customTypeEnum.ID,
			client:    suite.client.api,
			ctx:       testUser1.UserCtx,
		},
		{
			name: "not authorized, update system owned enum",
			request: testclient.UpdateCustomTypeEnumInput{
				Description: lo.ToPtr("Updated description"),
			},
			requestID:   systemTypeEnum.ID,
			client:      suite.client.api,
			ctx:         testUser1.UserCtx,
			expectedErr: notAuthorizedErrorMsg,
		},
		{
			name: "happy path, update multiple fields",
			request: testclient.UpdateCustomTypeEnumInput{
				Color:       lo.ToPtr("#ffffff"),
				Description: lo.ToPtr("a description of a custom type enum"),
			},
			requestID: customTypeEnum.ID,
			client:    suite.client.apiWithPAT,
			ctx:       context.Background(),
		},
		{
			name: "update not allowed, not enough permissions",
			request: testclient.UpdateCustomTypeEnumInput{
				Description: lo.ToPtr("Updated description"),
			},
			requestID:   customTypeEnum.ID,
			client:      suite.client.api,
			ctx:         viewOnlyUser.UserCtx,
			expectedErr: notAuthorizedErrorMsg,
		},
		{
			name: "update not allowed, no permissions",
			request: testclient.UpdateCustomTypeEnumInput{
				Description: lo.ToPtr("Updated description"),
			},
			requestID:   customTypeEnum.ID,
			client:      suite.client.api,
			ctx:         testUser2.UserCtx,
			expectedErr: notFoundErrorMsg,
		},
	}

	for _, tc := range testCases {
		t.Run("Update "+tc.name, func(t *testing.T) {
			resp, err := tc.client.UpdateCustomTypeEnum(tc.ctx, tc.requestID, tc.request)
			if tc.expectedErr != "" {
				assert.ErrorContains(t, err, tc.expectedErr)

				return
			}

			assert.NilError(t, err)
			assert.Assert(t, resp != nil)

			if tc.request.Description != nil {
				assert.Check(t, resp.UpdateCustomTypeEnum.CustomTypeEnum.Description != nil)
				assert.Check(t, *resp.UpdateCustomTypeEnum.CustomTypeEnum.Description == *tc.request.Description)
			}

			if tc.request.Color != nil {
				assert.Check(t, resp.UpdateCustomTypeEnum.CustomTypeEnum.Color != nil)
				assert.Check(t, *resp.UpdateCustomTypeEnum.CustomTypeEnum.Color == *tc.request.Color)
			}
		})
	}

	(&Cleanup[*generated.CustomTypeEnumDeleteOne]{client: suite.client.db.CustomTypeEnum, ID: customTypeEnum.ID}).MustDelete(testUser1.UserCtx, t)
}

func TestMutationDeleteCustomTypeEnum(t *testing.T) {
	// create objects to be deleted
	customTypeEnum1 := (&CustomTypeEnumBuilder{client: suite.client}).MustNew(testUser1.UserCtx, t)
	customTypeEnum2 := (&CustomTypeEnumBuilder{client: suite.client}).MustNew(testUser1.UserCtx, t)
	customTypeEnum3 := (&CustomTypeEnumBuilder{client: suite.client}).MustNew(testUser1.UserCtx, t)

	testCases := []struct {
		name        string
		idToDelete  string
		client      *testclient.TestClient
		ctx         context.Context
		expectedErr string
	}{
		{
			name:        "not found, delete",
			idToDelete:  customTypeEnum1.ID,
			client:      suite.client.api,
			ctx:         testUser2.UserCtx,
			expectedErr: notFoundErrorMsg,
		},
		{
			name:        "not authorized, delete",
			idToDelete:  customTypeEnum1.ID,
			client:      suite.client.api,
			ctx:         viewOnlyUser.UserCtx,
			expectedErr: notAuthorizedErrorMsg,
		},
		{
			name:       "happy path, delete",
			idToDelete: customTypeEnum1.ID,
			client:     suite.client.api,
			ctx:        testUser1.UserCtx,
		},
		{
			name:        "already deleted, not found",
			idToDelete:  customTypeEnum1.ID,
			client:      suite.client.api,
			ctx:         testUser1.UserCtx,
			expectedErr: "not found",
		},
		{
			name:       "happy path, delete using personal access token",
			idToDelete: customTypeEnum2.ID,
			client:     suite.client.apiWithPAT,
			ctx:        context.Background(),
		},
		{
			name:       "happy path, delete using api token",
			idToDelete: customTypeEnum3.ID,
			client:     suite.client.apiWithToken,
			ctx:        context.Background(),
		},
		{
			name:        "unknown id, not found",
			idToDelete:  ulids.New().String(),
			client:      suite.client.api,
			ctx:         testUser1.UserCtx,
			expectedErr: notFoundErrorMsg,
		},
	}

	for _, tc := range testCases {
		t.Run("Delete "+tc.name, func(t *testing.T) {
			resp, err := tc.client.DeleteCustomTypeEnum(tc.ctx, tc.idToDelete)
			if tc.expectedErr != "" {
				assert.ErrorContains(t, err, tc.expectedErr)

				return
			}

			assert.NilError(t, err)
			assert.Assert(t, resp != nil)
			assert.Check(t, is.Equal(tc.idToDelete, resp.DeleteCustomTypeEnum.DeletedID))
		})
	}
}

func TestMutationDeleteCustomTypeEnumInUse(t *testing.T) {
	// create a control enum
	controlEnum := (&CustomTypeEnumBuilder{
		client:     suite.client,
		Name:       "Preventative",
		ObjectType: "control",
	}).MustNew(testUser1.UserCtx, t)

	// create a task enum
	taskEnum := (&CustomTypeEnumBuilder{
		client:     suite.client,
		Name:       "Evidence",
		ObjectType: "task",
	}).MustNew(testUser1.UserCtx, t)

	controlResp, err := suite.client.api.CreateControl(testUser1.UserCtx, testclient.CreateControlInput{
		RefCode:         "TEST-1",
		ControlKindName: lo.ToPtr(controlEnum.Name),
	})
	assert.NilError(t, err)
	controlID := controlResp.CreateControl.Control.ID

	taskResp, err := suite.client.api.CreateTask(testUser1.UserCtx, testclient.CreateTaskInput{
		Title:        "Test Task",
		TaskKindName: lo.ToPtr(taskEnum.Name),
	})
	assert.NilError(t, err)
	taskID := taskResp.CreateTask.Task.ID

	subcontrolResp, err := suite.client.api.CreateSubcontrol(testUser1.UserCtx, testclient.CreateSubcontrolInput{
		RefCode:            "SUB-1",
		ControlID:          controlID,
		SubcontrolKindName: lo.ToPtr(controlEnum.Name),
	})
	assert.NilError(t, err)
	subcontrolID := subcontrolResp.CreateSubcontrol.Subcontrol.ID

	t.Run("delete enum in use by control", func(t *testing.T) {
		_, err := suite.client.api.DeleteCustomTypeEnum(testUser1.UserCtx, controlEnum.ID)
		assert.ErrorContains(t, err, "enum is in use")
	})

	t.Run("delete enum in use by task", func(t *testing.T) {
		_, err := suite.client.api.DeleteCustomTypeEnum(testUser1.UserCtx, taskEnum.ID)
		assert.ErrorContains(t, err, "enum is in use")
	})

	t.Run("delete enum in use by control and subcontrol", func(t *testing.T) {
		_, err := suite.client.api.DeleteCustomTypeEnum(testUser1.UserCtx, controlEnum.ID)
		assert.ErrorContains(t, err, "enum is in use")
	})

	// clean up the objects using the enums
	(&Cleanup[*generated.ControlDeleteOne]{client: suite.client.db.Control, ID: controlID}).MustDelete(testUser1.UserCtx, t)
	(&Cleanup[*generated.TaskDeleteOne]{client: suite.client.db.Task, ID: taskID}).MustDelete(testUser1.UserCtx, t)
	(&Cleanup[*generated.SubcontrolDeleteOne]{client: suite.client.db.Subcontrol, ID: subcontrolID}).MustDelete(testUser1.UserCtx, t)

	t.Run("enum deletion works if no object using it", func(t *testing.T) {
		resp, err := suite.client.api.DeleteCustomTypeEnum(testUser1.UserCtx, controlEnum.ID)
		assert.NilError(t, err)
		assert.Assert(t, resp != nil)
		assert.Check(t, is.Equal(controlEnum.ID, resp.DeleteCustomTypeEnum.DeletedID))
	})

	(&Cleanup[*generated.CustomTypeEnumDeleteOne]{client: suite.client.db.CustomTypeEnum, ID: taskEnum.ID}).MustDelete(testUser1.UserCtx, t)
}

func TestMutationCreateGlobalCustomTypeEnum(t *testing.T) {
	testCases := []struct {
		name        string
		request     testclient.CreateCustomTypeEnumInput
		client      *testclient.TestClient
		ctx         context.Context
		expectedErr string
	}{
		{
			name: "happy path, global enum with empty object type",
			request: testclient.CreateCustomTypeEnumInput{
				Name:       "Production",
				ObjectType: "",
				Field:      lo.ToPtr("environment"),
			},
			client: suite.client.api,
			ctx:    testUser1.UserCtx,
		},
		{
			name: "happy path, global scope enum",
			request: testclient.CreateCustomTypeEnumInput{
				Name:        "PCI-DSS",
				ObjectType:  "",
				Field:       lo.ToPtr("scope"),
				Description: lo.ToPtr("Payment Card Industry Data Security Standard scope"),
			},
			client: suite.client.api,
			ctx:    testUser1.UserCtx,
		},
	}

	for _, tc := range testCases {
		t.Run("Create "+tc.name, func(t *testing.T) {
			resp, err := tc.client.CreateCustomTypeEnum(tc.ctx, tc.request)
			if tc.expectedErr != "" {
				assert.ErrorContains(t, err, tc.expectedErr)

				return
			}

			assert.NilError(t, err)
			assert.Assert(t, resp != nil)

			assert.Check(t, resp.CreateCustomTypeEnum.CustomTypeEnum.Name == tc.request.Name)
			assert.Check(t, resp.CreateCustomTypeEnum.CustomTypeEnum.ObjectType == tc.request.ObjectType)

			if tc.request.Field != nil {
				assert.Check(t, resp.CreateCustomTypeEnum.CustomTypeEnum.Field == *tc.request.Field)
			}

			(&Cleanup[*generated.CustomTypeEnumDeleteOne]{client: suite.client.db.CustomTypeEnum, ID: resp.CreateCustomTypeEnum.CustomTypeEnum.ID}).MustDelete(testUser1.UserCtx, t)
		})
	}
}

func TestMutationDeleteGlobalCustomTypeEnumInUse(t *testing.T) {
	// create a global environment enum
	globalEnvEnum := (&CustomTypeEnumBuilder{
		client:     suite.client,
		Name:       "Staging",
		ObjectType: "",
		Field:      "environment",
	}).MustNew(testUser1.UserCtx, t)

	// create a task using the global environment enum
	taskResp, err := suite.client.api.CreateTask(testUser1.UserCtx, testclient.CreateTaskInput{
		Title:           "Test Task with Global Env",
		EnvironmentName: lo.ToPtr(globalEnvEnum.Name),
	})
	assert.NilError(t, err)
	taskID := taskResp.CreateTask.Task.ID

	t.Run("delete global enum in use by task", func(t *testing.T) {
		_, err := suite.client.api.DeleteCustomTypeEnum(testUser1.UserCtx, globalEnvEnum.ID)
		assert.ErrorContains(t, err, "enum is in use")
	})

	// cleanup task
	(&Cleanup[*generated.TaskDeleteOne]{client: suite.client.db.Task, ID: taskID}).MustDelete(testUser1.UserCtx, t)

	t.Run("global enum deletion works after removing references", func(t *testing.T) {
		resp, err := suite.client.api.DeleteCustomTypeEnum(testUser1.UserCtx, globalEnvEnum.ID)
		assert.NilError(t, err)
		assert.Assert(t, resp != nil)
		assert.Check(t, is.Equal(globalEnvEnum.ID, resp.DeleteCustomTypeEnum.DeletedID))
	})
}

func TestMutationDeleteGlobalEnumInUseByMultipleTables(t *testing.T) {
	// create a global scope enum
	globalScopeEnum := (&CustomTypeEnumBuilder{
		client:     suite.client,
		Name:       "SOC2-Type2",
		ObjectType: "",
		Field:      "scope",
	}).MustNew(testUser1.UserCtx, t)

	// create a task using the global scope enum
	taskResp, err := suite.client.api.CreateTask(testUser1.UserCtx, testclient.CreateTaskInput{
		Title:     "Task with Scope",
		ScopeName: lo.ToPtr(globalScopeEnum.Name),
	})
	assert.NilError(t, err)
	taskID := taskResp.CreateTask.Task.ID

	// create a control using the same global scope enum
	controlResp, err := suite.client.api.CreateControl(testUser1.UserCtx, testclient.CreateControlInput{
		RefCode:   "SCOPE-TEST-1",
		ScopeName: lo.ToPtr(globalScopeEnum.Name),
	})
	assert.NilError(t, err)
	controlID := controlResp.CreateControl.Control.ID

	t.Run("delete global enum in use by multiple tables fails", func(t *testing.T) {
		_, err := suite.client.api.DeleteCustomTypeEnum(testUser1.UserCtx, globalScopeEnum.ID)
		assert.ErrorContains(t, err, "enum is in use")
	})

	// remove task reference
	(&Cleanup[*generated.TaskDeleteOne]{client: suite.client.db.Task, ID: taskID}).MustDelete(testUser1.UserCtx, t)

	t.Run("delete global enum still fails with one reference remaining", func(t *testing.T) {
		_, err := suite.client.api.DeleteCustomTypeEnum(testUser1.UserCtx, globalScopeEnum.ID)
		assert.ErrorContains(t, err, "enum is in use")
	})

	// remove control reference
	(&Cleanup[*generated.ControlDeleteOne]{client: suite.client.db.Control, ID: controlID}).MustDelete(testUser1.UserCtx, t)

	t.Run("global enum deletion works after all references removed", func(t *testing.T) {
		resp, err := suite.client.api.DeleteCustomTypeEnum(testUser1.UserCtx, globalScopeEnum.ID)
		assert.NilError(t, err)
		assert.Assert(t, resp != nil)
		assert.Check(t, is.Equal(globalScopeEnum.ID, resp.DeleteCustomTypeEnum.DeletedID))
	})
}

func TestGlobalEnumLookup(t *testing.T) {
	// create a global environment enum
	globalEnvEnum := (&CustomTypeEnumBuilder{
		client:     suite.client,
		Name:       "GlobalDev",
		ObjectType: "",
		Field:      "environment",
	}).MustNew(testUser1.UserCtx, t)

	t.Run("task uses global enum", func(t *testing.T) {
		taskResp, err := suite.client.api.CreateTask(testUser1.UserCtx, testclient.CreateTaskInput{
			Title:           "Task Using Global Env",
			EnvironmentName: lo.ToPtr(globalEnvEnum.Name),
		})
		assert.NilError(t, err)
		assert.Assert(t, taskResp != nil)

		(&Cleanup[*generated.TaskDeleteOne]{client: suite.client.db.Task, ID: taskResp.CreateTask.Task.ID}).MustDelete(testUser1.UserCtx, t)
	})

	t.Run("invalid enum name fails", func(t *testing.T) {
		_, err := suite.client.api.CreateTask(testUser1.UserCtx, testclient.CreateTaskInput{
			Title:           "Task With Invalid Env",
			EnvironmentName: lo.ToPtr("NonExistentEnvironment"),
		})
		assert.ErrorContains(t, err, "does not exist")
	})

	(&Cleanup[*generated.CustomTypeEnumDeleteOne]{client: suite.client.db.CustomTypeEnum, ID: globalEnvEnum.ID}).MustDelete(testUser1.UserCtx, t)
}

func TestGlobalEnumUsedAcrossMultipleSchemas(t *testing.T) {
	// create a global environment enum
	globalEnvEnum := (&CustomTypeEnumBuilder{
		client:     suite.client,
		Name:       "Production",
		ObjectType: "",
		Field:      "environment",
	}).MustNew(testUser1.UserCtx, t)

	var taskID, controlID, riskID, evidenceID string

	t.Run("task can use global environment enum", func(t *testing.T) {
		resp, err := suite.client.api.CreateTask(testUser1.UserCtx, testclient.CreateTaskInput{
			Title:           "Task in Production",
			EnvironmentName: lo.ToPtr(globalEnvEnum.Name),
		})
		assert.NilError(t, err)
		assert.Assert(t, resp != nil)
		taskID = resp.CreateTask.Task.ID
	})

	t.Run("control can use same global environment enum", func(t *testing.T) {
		resp, err := suite.client.api.CreateControl(testUser1.UserCtx, testclient.CreateControlInput{
			RefCode:         "PROD-CTRL-1",
			EnvironmentName: lo.ToPtr(globalEnvEnum.Name),
		})
		assert.NilError(t, err)
		assert.Assert(t, resp != nil)
		controlID = resp.CreateControl.Control.ID
	})

	t.Run("risk can use same global environment enum", func(t *testing.T) {
		resp, err := suite.client.api.CreateRisk(testUser1.UserCtx, testclient.CreateRiskInput{
			Name:            "Production Risk",
			EnvironmentName: lo.ToPtr(globalEnvEnum.Name),
		})
		assert.NilError(t, err)
		assert.Assert(t, resp != nil)
		riskID = resp.CreateRisk.Risk.ID
	})

	t.Run("evidence can use same global environment enum", func(t *testing.T) {
		resp, err := suite.client.api.CreateEvidence(testUser1.UserCtx, testclient.CreateEvidenceInput{
			Name:            "Production Evidence",
			EnvironmentName: lo.ToPtr(globalEnvEnum.Name),
		}, nil)
		assert.NilError(t, err)
		assert.Assert(t, resp != nil)
		evidenceID = resp.CreateEvidence.Evidence.ID
	})

	t.Run("cannot delete global enum while in use by multiple schemas", func(t *testing.T) {
		_, err := suite.client.api.DeleteCustomTypeEnum(testUser1.UserCtx, globalEnvEnum.ID)
		assert.ErrorContains(t, err, "enum is in use")
	})

	// cleanup in reverse order
	(&Cleanup[*generated.EvidenceDeleteOne]{client: suite.client.db.Evidence, ID: evidenceID}).MustDelete(testUser1.UserCtx, t)
	(&Cleanup[*generated.RiskDeleteOne]{client: suite.client.db.Risk, ID: riskID}).MustDelete(testUser1.UserCtx, t)
	(&Cleanup[*generated.ControlDeleteOne]{client: suite.client.db.Control, ID: controlID}).MustDelete(testUser1.UserCtx, t)
	(&Cleanup[*generated.TaskDeleteOne]{client: suite.client.db.Task, ID: taskID}).MustDelete(testUser1.UserCtx, t)

	t.Run("can delete global enum after all references removed", func(t *testing.T) {
		resp, err := suite.client.api.DeleteCustomTypeEnum(testUser1.UserCtx, globalEnvEnum.ID)
		assert.NilError(t, err)
		assert.Assert(t, resp != nil)
		assert.Check(t, is.Equal(globalEnvEnum.ID, resp.DeleteCustomTypeEnum.DeletedID))
	})
}
