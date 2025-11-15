package corejobs_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
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

func mockGraphQLServer(t *testing.T, responses map[string]interface{}) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var requestBody struct {
			Query string `json:"query"`
		}

		err := json.NewDecoder(r.Body).Decode(&requestBody)
		require.NoError(t, err)

		var response map[string]interface{}
		if strings.Contains(requestBody.Query, "controls") {
			response = map[string]interface{}{
				"controls": responses["controls"],
			}
		} else if strings.Contains(requestBody.Query, "evidences") {
			response = map[string]interface{}{
				"evidences": responses["evidences"],
			}
		} else {
			response = responses
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"data": response,
		})
	}))
}

func TestExportContentWorker(t *testing.T) {
	t.Parallel()

	exportID := "export123"
	controlID1 := "control123"
	controlID2 := "control456"
	ownerID := "owner123"

	testCases := []struct {
		name                     string
		input                    corejobs.ExportContentArgs
		getExportByIDResponse    *openlaneclient.GetExportByID
		getExportByIDError       error
		graphQLResponses         map[string]interface{}
		updateExportError        error
		updateExportErrorSecond  error
		expectedError            string
		expectGetExportByID      bool
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
			getExportByIDResponse: &openlaneclient.GetExportByID{
				Export: openlaneclient.GetExportByID_Export{
					ID:         exportID,
					ExportType: enums.ExportTypeControl,
					OwnerID:    &ownerID,
					Fields:     []string{"id", "name", "description"},
				},
			},
			graphQLResponses: map[string]interface{}{
				"controls": map[string]interface{}{
					"edges": []interface{}{
						map[string]interface{}{
							"node": map[string]interface{}{
								"id":          controlID1,
								"name":        "Control 1",
								"description": "First control",
							},
						},
						map[string]interface{}{
							"node": map[string]interface{}{
								"id":          controlID2,
								"name":        "Control 2",
								"description": "Second control",
							},
						},
					},
					"pageInfo": map[string]interface{}{
						"hasNextPage": false,
						"endCursor":   nil,
					},
				},
			},
			expectGetExportByID:     true,
			expectUpdateExport:      true,
			expectUpdateExportFiles: true,
			expectedExportStatus:    &enums.ExportStatusReady,
		},
		{
			name: "missing export ID",
			input: corejobs.ExportContentArgs{
				ExportID: "",
			},
			expectedError:       "export_id is required for the export_content job",
			expectGetExportByID: false,
		},
		{
			name: "error getting export",
			input: corejobs.ExportContentArgs{
				ExportID: exportID,
			},
			getExportByIDError:   assert.AnError,
			expectGetExportByID:  true,
			expectUpdateExport:   true,
			expectedExportStatus: &enums.ExportStatusFailed,
		},
		{
			name: "no data found for export",
			input: corejobs.ExportContentArgs{
				ExportID: exportID,
			},
			getExportByIDResponse: &openlaneclient.GetExportByID{
				Export: openlaneclient.GetExportByID_Export{
					ID:         exportID,
					ExportType: enums.ExportTypeControl,
					OwnerID:    &ownerID,
					Fields:     []string{"id", "name"},
				},
			},
			graphQLResponses: map[string]interface{}{
				"controls": map[string]interface{}{
					"edges": []interface{}{},
					"pageInfo": map[string]interface{}{
						"hasNextPage": false,
						"endCursor":   nil,
					},
				},
			},
			expectGetExportByID:  true,
			expectUpdateExport:   true,
			expectedExportStatus: &enums.ExportStatusNodata,
		},
		{
			name: "error updating export with file",
			input: corejobs.ExportContentArgs{
				ExportID: exportID,
			},
			getExportByIDResponse: &openlaneclient.GetExportByID{
				Export: openlaneclient.GetExportByID_Export{
					ID:         exportID,
					ExportType: enums.ExportTypeControl,
					OwnerID:    &ownerID,
					Fields:     []string{"id", "name"},
				},
			},
			graphQLResponses: map[string]interface{}{
				"controls": map[string]interface{}{
					"edges": []interface{}{
						map[string]interface{}{
							"node": map[string]interface{}{
								"id":   controlID1,
								"name": "Control 1",
							},
						},
					},
					"pageInfo": map[string]interface{}{
						"hasNextPage": false,
						"endCursor":   nil,
					},
				},
			},
			updateExportError:        assert.AnError,
			expectGetExportByID:      true,
			expectUpdateExport:       true,
			expectUpdateExportFiles:  true,
			expectUpdateExportSecond: true,
			expectedExportStatus:     &enums.ExportStatusFailed,
		},
		{
			name: "export evidence type",
			input: corejobs.ExportContentArgs{
				ExportID: exportID,
			},
			getExportByIDResponse: &openlaneclient.GetExportByID{
				Export: openlaneclient.GetExportByID_Export{
					ID:         exportID,
					ExportType: enums.ExportTypeEvidence,
					OwnerID:    &ownerID,
					Fields:     []string{"id", "name", "type"},
				},
			},
			graphQLResponses: map[string]interface{}{
				"evidences": map[string]interface{}{
					"edges": []interface{}{
						map[string]interface{}{
							"node": map[string]interface{}{
								"id":   "evidence123",
								"name": "Evidence 1",
								"type": "Document",
							},
						},
					},
					"pageInfo": map[string]interface{}{
						"hasNextPage": false,
						"endCursor":   nil,
					},
				},
			},
			expectGetExportByID:     true,
			expectUpdateExport:      true,
			expectUpdateExportFiles: true,
			expectedExportStatus:    &enums.ExportStatusReady,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			var mockServer *httptest.Server
			if tc.graphQLResponses != nil {
				mockServer = mockGraphQLServer(t, tc.graphQLResponses)
				defer mockServer.Close()
			}

			olMock := olmocks.NewMockOpenlaneGraphClient(t)

			if tc.expectGetExportByID {
				olMock.EXPECT().GetExportByID(mock.MatchedBy(func(ctx context.Context) bool {
					return ctx != nil
				}), tc.input.ExportID).Return(tc.getExportByIDResponse, tc.getExportByIDError)
			}

			if tc.expectUpdateExport {
				if tc.expectUpdateExportFiles {
					olMock.EXPECT().UpdateExport(mock.MatchedBy(func(ctx context.Context) bool {
						return ctx != nil
					}), tc.input.ExportID, mock.MatchedBy(func(input openlaneclient.UpdateExportInput) bool {
						return input.Status != nil && *input.Status == enums.ExportStatusReady
					}), mock.MatchedBy(func(files []*graphql.Upload) bool {
						return len(files) == 1 && files[0].ContentType == "text/csv"
					}), mock.Anything,
					).Return(&openlaneclient.UpdateExport{}, tc.updateExportError)

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
						if input.Status == nil {
							return false
						}
						expectedStatus := *tc.expectedExportStatus
						actualStatus := *input.Status
						return actualStatus == expectedStatus
					}), ([]*graphql.Upload)(nil)).Return(&openlaneclient.UpdateExport{}, tc.updateExportError)
				}
			}

			config := corejobs.ExportWorkerConfig{
				OpenlaneConfig: corejobs.OpenlaneConfig{
					OpenlaneAPIHost:  "https://api.example.com",
					OpenlaneAPIToken: "tola_test-token",
				},
			}
			if mockServer != nil {
				config.OpenlaneAPIHost = mockServer.URL
			}

			worker := &corejobs.ExportContentWorker{
				Config: config,
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

func TestExportContentWorker_GraphQLErrorResponse(t *testing.T) {
	t.Parallel()

	exportID := "export123"
	ownerID := "owner123"

	// mock server that returns GraphQL errors
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"errors": []interface{}{
				map[string]interface{}{
					"message": "Field 'controls' not found",
				},
			},
		})
	}))
	defer mockServer.Close()

	olMock := olmocks.NewMockOpenlaneGraphClient(t)

	// mock successful GetExportByID
	olMock.EXPECT().GetExportByID(mock.MatchedBy(func(ctx context.Context) bool {
		return ctx != nil
	}), exportID).Return(&openlaneclient.GetExportByID{
		Export: openlaneclient.GetExportByID_Export{
			ID:         exportID,
			ExportType: enums.ExportTypeControl,
			OwnerID:    &ownerID,
			Fields:     []string{"id", "name"},
		},
	}, nil)

	// mock failed UpdateExport call
	olMock.EXPECT().UpdateExport(mock.MatchedBy(func(ctx context.Context) bool {
		return ctx != nil
	}), exportID, mock.MatchedBy(func(input openlaneclient.UpdateExportInput) bool {
		return input.Status != nil && *input.Status == enums.ExportStatusFailed
	}), ([]*graphql.Upload)(nil)).Return(&openlaneclient.UpdateExport{}, nil)

	worker := &corejobs.ExportContentWorker{
		Config: corejobs.ExportWorkerConfig{
			OpenlaneConfig: corejobs.OpenlaneConfig{
				OpenlaneAPIHost:  mockServer.URL,
				OpenlaneAPIToken: "tola_test-token",
			},
		},
	}

	worker.WithOpenlaneClient(olMock)

	ctx := context.Background()
	err := worker.Work(ctx, &river.Job[corejobs.ExportContentArgs]{
		Args: corejobs.ExportContentArgs{ExportID: exportID},
	})

	require.NoError(t, err)
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

func TestCreateFieldsStr(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name     string
		fields   []string
		expected string
	}{
		{
			name:     "empty fields returns id",
			fields:   []string{},
			expected: "id\n        ",
		},
		{
			name:     "nil fields returns id",
			fields:   nil,
			expected: "id\n        ",
		},
		{
			name:     "simple fields",
			fields:   []string{"id", "name", "description"},
			expected: "id\n        name\n        description\n        ",
		},
		{
			name:     "single nested field",
			fields:   []string{"owner.name"},
			expected: "owner\n        {  name  } \n        ",
		},
		{
			name:     "nested field with plural parent",
			fields:   []string{"tasks.title"},
			expected: "tasks\n        { edges { node { title  }  }  } \n        ",
		},
		{
			name:     "deeply nested field",
			fields:   []string{"owner.tasks.title"},
			expected: "owner\n        {  tasks { edges { node { title  }  }  }  } \n        ",
		},
		{
			name:     "mixed simple and nested fields",
			fields:   []string{"id", "name", "owner.name", "description"},
			expected: "id\n        name\n        owner\n        {  name  } \n        description\n        ",
		},
		{
			name:     "multiple nested fields",
			fields:   []string{"owner.name", "tasks.title"},
			expected: "owner\n        {  name  } \n        tasks\n        { edges { node { title  }  }  } \n        ",
		},
		{
			name:     "nested field with multiple levels and plurals",
			fields:   []string{"controls.tasks.assignee.name"},
			expected: "controls\n        { edges { node { tasks { edges { node { assignee {  name  }  }  }  }  }  }  } \n        ",
		},
		{
			name:     "single field",
			fields:   []string{"id"},
			expected: "id\n        ",
		},
		{
			name:     "field with singular to singular nesting",
			fields:   []string{"owner.profile.email"},
			expected: "owner\n        {  profile {  email  }  } \n        ",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := corejobs.CreateFieldsStr(tc.fields)
			require.Equal(t, tc.expected, result)
		})
	}
}
