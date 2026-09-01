package graphapi_test

import (
	"context"
	"testing"

	th "github.com/theopenlane/core/v2/internal/graphapi/testharness"

	"github.com/samber/lo"
	"gotest.tools/v3/assert"
	is "gotest.tools/v3/assert/cmp"

	"github.com/theopenlane/core/common/enums"
	"github.com/theopenlane/core/v2/internal/ent/generated"
	"github.com/theopenlane/core/v2/internal/graphapi/testclient"
)

func TestMutationCreateExport(t *testing.T) {
	emptyFilter := "{}"
	testCases := []struct {
		name        string
		request     testclient.CreateExportInput
		client      *testclient.TestClient
		ctx         context.Context
		expectedErr string
	}{
		{
			name: "happy path, create export with regular user",
			request: testclient.CreateExportInput{
				ExportType: enums.ExportTypeInternalPolicy,
				Format:     lo.ToPtr(enums.ExportFormatCsv),
				Fields:     []string{"name", "details", "summary", "updatedAt"},
				Filters:    &emptyFilter,
			},
			client: suite.Client.API,
			ctx:    th.SharedTestUser1.UserCtx,
		},
		{
			name: "happy path with PAT",
			request: testclient.CreateExportInput{
				ExportType: enums.ExportTypeControl,
				OwnerID:    &th.SharedTestUser1.OrganizationID,
				Format:     lo.ToPtr(enums.ExportFormatCsv),
			},
			client: suite.Client.APIWithPAT,
			ctx:    context.Background(),
		},
		{
			name: "happy path with system admin",
			request: testclient.CreateExportInput{
				ExportType: enums.ExportTypeControl,
				OwnerID:    &th.SharedTestUser1.OrganizationID,
				Format:     lo.ToPtr(enums.ExportFormatCsv),
			},
			client: suite.Client.API,
			ctx:    th.SharedSystemAdminUser.UserCtx,
		},
	}

	for _, tc := range testCases {
		t.Run("Create "+tc.name, func(t *testing.T) {
			resp, err := tc.client.CreateExport(tc.ctx, tc.request)
			if tc.expectedErr != "" {
				assert.ErrorContains(t, err, tc.expectedErr)

				return
			}

			assert.NilError(t, err)
			assert.Assert(t, resp != nil)
			assert.Check(t, is.Equal(tc.request.ExportType.String(), resp.CreateExport.Export.ExportType.String()))
			assert.Check(t, is.Equal(enums.ExportStatusPending.String(), resp.CreateExport.Export.Status.String()))

			if tc.request.OwnerID != nil {
				assert.Check(t, is.Equal(*tc.request.OwnerID, *resp.CreateExport.Export.OwnerID))
			}

			if tc.ctx == th.SharedSystemAdminUser.UserCtx {
				export, err := suite.Client.API.GetExportByID(th.SharedSystemAdminUser.UserCtx, resp.CreateExport.Export.ID)
				assert.NilError(t, err)
				assert.Assert(t, export != nil)
				assert.Check(t, is.Equal(resp.CreateExport.Export.ID, export.Export.ID))
			}

			(&th.Cleanup[*generated.ExportDeleteOne]{Client: suite.Client.DB.Export, ID: resp.CreateExport.Export.ID}).MustDelete(th.SharedSystemAdminUser.UserCtx, t)
		})
	}
}

func TestMutationUpdateExport(t *testing.T) {
	createResp, err := suite.Client.API.CreateExport(th.SharedTestUser1.UserCtx, testclient.CreateExportInput{
		ExportType: enums.ExportTypeControl,
		OwnerID:    &th.SharedTestUser1.OrganizationID,
		Format:     lo.ToPtr(enums.ExportFormatCsv),
	})
	assert.NilError(t, err)
	assert.Assert(t, createResp != nil)
	exportID := createResp.CreateExport.Export.ID

	testCases := []struct {
		name        string
		exportID    string
		request     testclient.UpdateExportInput
		client      *testclient.TestClient
		ctx         context.Context
		expectedErr string
	}{
		{
			name:     "happy path, update status with system admin",
			exportID: exportID,
			request: testclient.UpdateExportInput{
				Status: &enums.ExportStatusReady,
			},
			client: suite.Client.API,
			ctx:    th.SharedSystemAdminUser.UserCtx,
		},
		{
			name:     "unauthorized user - regular user cannot update",
			exportID: exportID,
			request: testclient.UpdateExportInput{
				Status: &enums.ExportStatusReady,
			},
			client:      suite.Client.API,
			ctx:         th.SharedTestUser1.UserCtx,
			expectedErr: "export not found",
		},
		{
			name:     "unauthorized user - different org user cannot update",
			exportID: exportID,
			request: testclient.UpdateExportInput{
				Status: &enums.ExportStatusReady,
			},
			client:      suite.Client.API,
			ctx:         th.SharedTestUser2.UserCtx,
			expectedErr: "export not found",
		},
		{
			name:     "export not found",
			exportID: "non-existent-id",
			request: testclient.UpdateExportInput{
				Status: &enums.ExportStatusReady,
			},
			client:      suite.Client.API,
			ctx:         th.SharedSystemAdminUser.UserCtx,
			expectedErr: "export not found",
		},
	}

	for _, tc := range testCases {
		t.Run("Update "+tc.name, func(t *testing.T) {
			resp, err := tc.client.UpdateExport(tc.ctx, tc.exportID, tc.request, nil)
			if tc.expectedErr != "" {
				assert.ErrorContains(t, err, tc.expectedErr)

				return
			}

			assert.NilError(t, err)
			assert.Assert(t, resp != nil)
			assert.Check(t, is.Equal(tc.exportID, resp.UpdateExport.Export.ID))

			if tc.request.Status != nil {
				assert.Check(t, is.Equal(tc.request.Status.String(), resp.UpdateExport.Export.Status.String()))
			}
		})
	}

	(&th.Cleanup[*generated.ExportDeleteOne]{Client: suite.Client.DB.Export, ID: exportID}).MustDelete(th.SharedSystemAdminUser.UserCtx, t)
}

func TestQueryExport(t *testing.T) {
	createResp, err := suite.Client.API.CreateExport(th.SharedTestUser1.UserCtx, testclient.CreateExportInput{
		ExportType: enums.ExportTypeControl,
		OwnerID:    &th.SharedTestUser1.OrganizationID,
		Format:     lo.ToPtr(enums.ExportFormatCsv),
	})
	assert.NilError(t, err)
	assert.Assert(t, createResp != nil)
	exportID := createResp.CreateExport.Export.ID

	testCases := []struct {
		name        string
		exportID    string
		client      *testclient.TestClient
		ctx         context.Context
		expectedErr string
	}{
		{
			name:     "happy path with system admin",
			exportID: exportID,
			client:   suite.Client.API,
			ctx:      th.SharedSystemAdminUser.UserCtx,
		},
		{
			name:     "happy path with PAT and system admin context",
			exportID: exportID,
			client:   suite.Client.APIWithPAT,
			ctx:      context.Background(),
		},
		{
			name:     "regular user can query",
			exportID: exportID,
			client:   suite.Client.API,
			ctx:      th.SharedTestUser1.UserCtx,
		},
		{
			name:        "unauthorized user - different org user cannot query",
			exportID:    exportID,
			client:      suite.Client.API,
			ctx:         th.SharedTestUser2.UserCtx,
			expectedErr: "export not found",
		},
		{
			name:     "view only user can query",
			exportID: exportID,
			client:   suite.Client.API,
			ctx:      th.SharedViewOnlyUser2.UserCtx,
		},
		{
			name:        "export not found",
			exportID:    "non-existent-id",
			client:      suite.Client.API,
			ctx:         th.SharedSystemAdminUser.UserCtx,
			expectedErr: "export not found",
		},
		{
			name:     "system admin can fetch item",
			exportID: exportID,
			client:   suite.Client.API,
			ctx:      th.SharedSystemAdminUser.UserCtx,
		},
	}

	for _, tc := range testCases {
		t.Run("Query "+tc.name, func(t *testing.T) {
			export, err := tc.client.GetExportByID(tc.ctx, tc.exportID)
			if tc.expectedErr != "" {
				assert.ErrorContains(t, err, tc.expectedErr)
				return
			}

			assert.NilError(t, err)
			assert.Assert(t, export != nil)
			assert.Check(t, is.Equal(tc.exportID, export.Export.ID))
			assert.Check(t, is.Equal(enums.ExportTypeControl.String(), export.Export.ExportType.String()))
			assert.Check(t, is.Equal(enums.ExportStatusPending.String(), export.Export.Status.String()))
			assert.Check(t, is.Equal(th.SharedTestUser1.OrganizationID, *export.Export.OwnerID))
		})
	}

	(&th.Cleanup[*generated.ExportDeleteOne]{Client: suite.Client.DB.Export, ID: exportID}).MustDelete(th.SharedSystemAdminUser.UserCtx, t)
}

func TestMutationDeleteExport(t *testing.T) {
	testCases := []struct {
		name        string
		setupExport func() string
		client      *testclient.TestClient
		ctx         context.Context
		expectedErr string
	}{
		{
			name: "happy path with system admin",
			setupExport: func() string {
				resp, err := suite.Client.API.CreateExport(th.SharedTestUser1.UserCtx, testclient.CreateExportInput{
					ExportType: enums.ExportTypeControl,
					OwnerID:    &th.SharedTestUser1.OrganizationID,
					Format:     lo.ToPtr(enums.ExportFormatCsv),
				})
				assert.NilError(t, err)
				return resp.CreateExport.Export.ID
			},
			client: suite.Client.API,
			ctx:    th.SharedSystemAdminUser.UserCtx,
		},
		{
			name: "unauthorized user - regular user cannot delete",
			setupExport: func() string {
				resp, err := suite.Client.API.CreateExport(th.SharedTestUser1.UserCtx, testclient.CreateExportInput{
					Format:     lo.ToPtr(enums.ExportFormatCsv),
					OwnerID:    &th.SharedTestUser1.OrganizationID,
					ExportType: enums.ExportTypeControl,
				})
				assert.NilError(t, err)
				return resp.CreateExport.Export.ID
			},
			client:      suite.Client.API,
			ctx:         th.SharedTestUser1.UserCtx,
			expectedErr: "export not found",
		},
		{
			name: "unauthorized user - different org user cannot delete",
			setupExport: func() string {
				resp, err := suite.Client.API.CreateExport(th.SharedTestUser1.UserCtx, testclient.CreateExportInput{
					Format:     lo.ToPtr(enums.ExportFormatCsv),
					ExportType: enums.ExportTypeControl,
					OwnerID:    &th.SharedTestUser1.OrganizationID,
				})
				assert.NilError(t, err)
				return resp.CreateExport.Export.ID
			},
			client:      suite.Client.API,
			ctx:         th.SharedTestUser2.UserCtx,
			expectedErr: "export not found",
		},
	}

	for _, tc := range testCases {
		t.Run("Delete "+tc.name, func(t *testing.T) {
			exportID := tc.setupExport()

			cleanup := &th.Cleanup[*generated.ExportDeleteOne]{Client: suite.Client.DB.Export, ID: exportID}

			if tc.expectedErr != "" {
				_, err := tc.client.UpdateExport(tc.ctx, exportID, testclient.UpdateExportInput{
					Status: &enums.ExportStatusFailed,
				}, nil)
				assert.ErrorContains(t, err, tc.expectedErr)

				cleanup.MustDelete(th.SharedSystemAdminUser.UserCtx, t)
				return
			}

			cleanup.MustDelete(tc.ctx, t)

			_, err := suite.Client.API.GetExportByID(th.SharedSystemAdminUser.UserCtx, exportID)
			assert.ErrorContains(t, err, "export not found")
		})
	}
}
