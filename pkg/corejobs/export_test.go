package corejobs_test

import (
	"context"
	"testing"

	"github.com/99designs/gqlgen/graphql"
	"github.com/riverqueue/river"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/theopenlane/core/pkg/corejobs"
	"github.com/theopenlane/core/pkg/enums"
	"github.com/theopenlane/core/pkg/openlaneclient"

	olmocks "github.com/theopenlane/core/pkg/corejobs/internal/olclient/mocks"
)

func TestExportContentWorker(t *testing.T) {
	t.Parallel()

	exportID := "export123"
	controlID1 := "control123"
	controlID2 := "control456"

	testCases := []struct {
		name                     string
		input                    corejobs.ExportContentArgs
		config                   corejobs.ExportWorkerConfig
		getExportByIDResponse    *openlaneclient.GetExportByID
		getExportByIDError       error
		getAllControlsResponse   *openlaneclient.GetAllControls
		getAllControlsError      error
		updateExportError        error
		updateExportErrorSecond  error
		expectedError            string
		expectGetExportByID      bool
		expectGetAllControls     bool
		expectUpdateExport       bool
		expectUpdateExportFiles  bool
		expectUpdateExportSecond bool
		expectedExportStatus     *enums.ExportStatus
	}{
		{
			name: "happy path - export controls",
			input: corejobs.ExportContentArgs{
				ExportID: exportID,
			},
			config: corejobs.ExportWorkerConfig{
				OpenlaneAPIHost:  "https://api.example.com",
				OpenlaneAPIToken: "test-token",
			},
			getExportByIDResponse: &openlaneclient.GetExportByID{
				Export: openlaneclient.GetExportByID_Export{
					ID:         exportID,
					ExportType: enums.ExportTypeControl,
				},
			},
			getAllControlsResponse: &openlaneclient.GetAllControls{
				Controls: openlaneclient.GetAllControls_Controls{
					Edges: []*openlaneclient.GetAllControls_Controls_Edges{
						{
							Node: &openlaneclient.GetAllControls_Controls_Edges_Node{
								ID: controlID1,
							},
						},
						{
							Node: &openlaneclient.GetAllControls_Controls_Edges_Node{
								ID: controlID2,
							},
						},
					},
				},
			},
			expectGetExportByID:     true,
			expectGetAllControls:    true,
			expectUpdateExport:      true,
			expectUpdateExportFiles: true,
			expectedExportStatus:    &enums.ExportStatusReady,
		},
		{
			name: "missing export ID",
			input: corejobs.ExportContentArgs{
				ExportID: "",
			},
			config: corejobs.ExportWorkerConfig{
				OpenlaneAPIHost:  "https://api.example.com",
				OpenlaneAPIToken: "test-token",
			},
			expectedError:       "export_id is required for the export_content job",
			expectGetExportByID: false,
		},
		{
			name: "error getting export",
			input: corejobs.ExportContentArgs{
				ExportID: exportID,
			},
			config: corejobs.ExportWorkerConfig{
				OpenlaneAPIHost:  "https://api.example.com",
				OpenlaneAPIToken: "test-token",
			},
			getExportByIDError:   assert.AnError,
			expectGetExportByID:  true,
			expectUpdateExport:   true,
			expectedExportStatus: &enums.ExportStatusFailed,
		},
		{
			name: "unsupported export type",
			input: corejobs.ExportContentArgs{
				ExportID: exportID,
			},
			config: corejobs.ExportWorkerConfig{
				OpenlaneAPIHost:  "https://api.example.com",
				OpenlaneAPIToken: "test-token",
			},
			getExportByIDResponse: &openlaneclient.GetExportByID{
				Export: openlaneclient.GetExportByID_Export{
					ID:         exportID,
					ExportType: enums.ExportTypeInvalid, // Unsupported type
				},
			},
			expectGetExportByID:  true,
			expectUpdateExport:   true,
			expectedExportStatus: &enums.ExportStatusFailed,
		},
		{
			name: "error getting controls",
			input: corejobs.ExportContentArgs{
				ExportID: exportID,
			},
			config: corejobs.ExportWorkerConfig{
				OpenlaneAPIHost:  "https://api.example.com",
				OpenlaneAPIToken: "test-token",
			},
			getExportByIDResponse: &openlaneclient.GetExportByID{
				Export: openlaneclient.GetExportByID_Export{
					ID:         exportID,
					ExportType: enums.ExportTypeControl,
				},
			},
			getAllControlsError:  assert.AnError,
			expectGetExportByID:  true,
			expectGetAllControls: true,
			expectUpdateExport:   true,
			expectedExportStatus: &enums.ExportStatusFailed,
		},
		{
			name: "no controls found",
			input: corejobs.ExportContentArgs{
				ExportID: exportID,
			},
			config: corejobs.ExportWorkerConfig{
				OpenlaneAPIHost:  "https://api.example.com",
				OpenlaneAPIToken: "test-token",
			},
			getExportByIDResponse: &openlaneclient.GetExportByID{
				Export: openlaneclient.GetExportByID_Export{
					ID:         exportID,
					ExportType: enums.ExportTypeControl,
				},
			},
			getAllControlsResponse: &openlaneclient.GetAllControls{
				Controls: openlaneclient.GetAllControls_Controls{
					Edges: []*openlaneclient.GetAllControls_Controls_Edges{},
				},
			},
			expectGetExportByID:  true,
			expectGetAllControls: true,
			expectUpdateExport:   true,
			expectedExportStatus: &enums.ExportStatusFailed,
		},
		{
			name: "error updating export with file",
			input: corejobs.ExportContentArgs{
				ExportID: exportID,
			},
			config: corejobs.ExportWorkerConfig{
				OpenlaneAPIHost:  "https://api.example.com",
				OpenlaneAPIToken: "test-token",
			},
			getExportByIDResponse: &openlaneclient.GetExportByID{
				Export: openlaneclient.GetExportByID_Export{
					ID:         exportID,
					ExportType: enums.ExportTypeControl,
				},
			},
			getAllControlsResponse: &openlaneclient.GetAllControls{
				Controls: openlaneclient.GetAllControls_Controls{
					Edges: []*openlaneclient.GetAllControls_Controls_Edges{
						{
							Node: &openlaneclient.GetAllControls_Controls_Edges_Node{
								ID: controlID1,
							},
						},
					},
				},
			},
			updateExportError:        assert.AnError,
			expectGetExportByID:      true,
			expectGetAllControls:     true,
			expectUpdateExport:       true,
			expectUpdateExportFiles:  true,
			expectUpdateExportSecond: true,
			expectedExportStatus:     &enums.ExportStatusFailed,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			olMock := olmocks.NewMockOpenlaneGraphClient(t)

			if tc.expectGetExportByID {
				olMock.EXPECT().GetExportByID(mock.MatchedBy(func(ctx context.Context) bool {
					return ctx != nil
				}), tc.input.ExportID).Return(tc.getExportByIDResponse, tc.getExportByIDError)
			}

			if tc.expectGetAllControls {
				olMock.EXPECT().GetAllControls(mock.MatchedBy(func(ctx context.Context) bool {
					return ctx != nil
				})).Return(tc.getAllControlsResponse, tc.getAllControlsError)
			}

			if tc.expectUpdateExport {
				if tc.expectUpdateExportFiles {

					// success
					//
					olMock.EXPECT().UpdateExport(mock.MatchedBy(func(ctx context.Context) bool {
						return ctx != nil
					}), tc.input.ExportID, mock.MatchedBy(func(input openlaneclient.UpdateExportInput) bool {
						return input.Status != nil && *input.Status == enums.ExportStatusReady
					}), mock.MatchedBy(func(files []*graphql.Upload) bool {
						return len(files) == 1 && files[0].ContentType == "text/csv"
					})).Return(&openlaneclient.UpdateExport{}, tc.updateExportError)

					if tc.expectUpdateExportSecond {
						olMock.EXPECT().UpdateExport(mock.MatchedBy(func(ctx context.Context) bool {
							return ctx != nil
						}), tc.input.ExportID, mock.MatchedBy(func(input openlaneclient.UpdateExportInput) bool {
							return input.Status != nil && *input.Status == *tc.expectedExportStatus
						}), ([]*graphql.Upload)(nil)).Return(&openlaneclient.UpdateExport{}, tc.updateExportErrorSecond)
					}
				} else {

					olMock.EXPECT().UpdateExport(mock.MatchedBy(func(ctx context.Context) bool {
						return ctx != nil
					}), tc.input.ExportID, mock.MatchedBy(func(input openlaneclient.UpdateExportInput) bool {
						return input.Status != nil && *input.Status == *tc.expectedExportStatus
					}), ([]*graphql.Upload)(nil)).Return(&openlaneclient.UpdateExport{}, tc.updateExportError)

				}
			}

			worker := &corejobs.ExportContentWorker{
				Config: tc.config,
			}

			worker.WithOpenlaneClient(olMock)

			ctx := context.Background()
			err := worker.Work(ctx, &river.Job[corejobs.ExportContentArgs]{Args: tc.input})

			if tc.expectedError != "" {
				require.Error(t, err)
				require.Contains(t, err.Error(), tc.expectedError)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestExportContentWorker_WithOpenlaneClient(t *testing.T) {
	t.Parallel()

	olMock := olmocks.NewMockOpenlaneGraphClient(t)
	worker := &corejobs.ExportContentWorker{}

	result := worker.WithOpenlaneClient(olMock)

	require.Equal(t, worker, result, "WithOpenlaneClient should return the same worker instance")
}

func TestExportContentArgs_Kind(t *testing.T) {
	t.Parallel()

	args := corejobs.ExportContentArgs{}
	require.Equal(t, "export_content", args.Kind())
}
