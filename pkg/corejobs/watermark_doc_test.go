package corejobs_test

import (
	"context"
	"net/http"
	"net/http/httptest"
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

// mockFileServer creates a mock HTTP server that serves files for testing
func mockFileServer(_ *testing.T, responses map[string][]byte) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		url := r.URL.String()
		if data, exists := responses[url]; exists {
			w.Header().Set("Content-Type", "application/pdf")
			w.WriteHeader(http.StatusOK)
			w.Write(data)
		} else {
			w.WriteHeader(http.StatusNotFound)
		}
	}))
}

func TestWatermarkDocArgs_Kind(t *testing.T) {
	t.Parallel()

	args := corejobs.WatermarkDocArgs{}
	require.Equal(t, "watermark_doc", args.Kind())
}

func TestWatermarkDocWorker_WithOpenlaneClient(t *testing.T) {
	t.Parallel()

	olMock := olmocks.NewMockOpenlaneGraphClient(t)
	worker := &corejobs.WatermarkDocWorker{}

	result := worker.WithOpenlaneClient(olMock)

	require.Equal(t, worker, result, "WithOpenlaneClient should return the same worker instance")
}

func TestWatermarkDocWorker_Work(t *testing.T) {
	t.Parallel()

	docID := "doc123"
	trustCenterID := "tc123"
	fileID := "file123"
	originalFileID := "original123"
	watermarkConfigID := "wm123"

	// Sample PDF content for testing - this is a minimal valid PDF
	samplePDFContent := []byte(`%PDF-1.4
1 0 obj
<<
/Type /Catalog
/Pages 2 0 R
>>
endobj
2 0 obj
<<
/Type /Pages
/Kids [3 0 R]
/Count 1
>>
endobj
3 0 obj
<<
/Type /Page
/Parent 2 0 R
/MediaBox [0 0 612 792]
/Contents 4 0 R
>>
endobj
4 0 obj
<<
/Length 44
>>
stream
BT
/F1 12 Tf
72 720 Td
(Hello World) Tj
ET
endstream
endobj
xref
0 5
0000000000 65535 f
0000000009 00000 n
0000000058 00000 n
0000000115 00000 n
0000000204 00000 n
trailer
<<
/Size 5
/Root 1 0 R
>>
startxref
297
%%EOF`)
	// Invalid PDF content that will cause watermarking to fail
	invalidPDFContent := []byte(`%PDF-1.4
This is not a valid PDF file content.
It lacks proper PDF structure and will cause parsing errors.
%%EOF`)
	sampleImageContent := []byte("fake-image-content")

	testCases := []struct {
		name                            string
		input                           corejobs.WatermarkDocArgs
		mockFileServerResponses         map[string][]byte
		getTrustCenterDocResponse       *openlaneclient.GetTrustCenterDocByID
		getTrustCenterDocError          error
		getWatermarkConfigsResponse     *openlaneclient.GetTrustCenterWatermarkConfigs
		getWatermarkConfigsError        error
		updateTrustCenterDocResponses   []*openlaneclient.UpdateTrustCenterDoc
		updateTrustCenterDocErrors      []error
		expectedError                   string
		expectGetTrustCenterDoc         bool
		expectGetWatermarkConfigs       bool
		expectUpdateTrustCenterDocCalls int
		expectedFinalWatermarkStatus    *enums.WatermarkStatus
	}{
		{
			name: "text watermark - success",
			input: corejobs.WatermarkDocArgs{
				TrustCenterDocumentID: docID,
			},
			mockFileServerResponses: map[string][]byte{
				"/original-file": samplePDFContent,
			},
			getTrustCenterDocResponse: &openlaneclient.GetTrustCenterDocByID{
				TrustCenterDoc: openlaneclient.GetTrustCenterDocByID_TrustCenterDoc{
					ID:            docID,
					FileID:        &fileID,
					TrustCenterID: &trustCenterID,
					OriginalFile: &openlaneclient.GetTrustCenterDocByID_TrustCenterDoc_OriginalFile{
						ID:               originalFileID,
						ProvidedFileName: "test.pdf",
						PresignedURL:     stringPtr("/original-file"),
					},
					File: &openlaneclient.GetTrustCenterDocByID_TrustCenterDoc_File{
						PresignedURL: stringPtr("/original-file"),
					},
				},
			},
			getWatermarkConfigsResponse: &openlaneclient.GetTrustCenterWatermarkConfigs{
				TrustCenterWatermarkConfigs: openlaneclient.GetTrustCenterWatermarkConfigs_TrustCenterWatermarkConfigs{
					Edges: []*openlaneclient.GetTrustCenterWatermarkConfigs_TrustCenterWatermarkConfigs_Edges{
						{
							Node: &openlaneclient.GetTrustCenterWatermarkConfigs_TrustCenterWatermarkConfigs_Edges_Node{
								ID:            watermarkConfigID,
								TrustCenterID: &trustCenterID,
								Text:          stringPtr("CONFIDENTIAL"),
								FontSize:      float64Ptr(12.0),
								Opacity:       float64Ptr(0.5),
								Rotation:      float64Ptr(45.0),
								Color:         stringPtr("red"),
							},
						},
					},
				},
			},
			updateTrustCenterDocResponses: []*openlaneclient.UpdateTrustCenterDoc{
				{}, // First call to set status to IN_PROGRESS
				{}, // Second call to set status to SUCCESS with file upload
			},
			expectGetTrustCenterDoc:         true,
			expectGetWatermarkConfigs:       true,
			expectUpdateTrustCenterDocCalls: 2,
			expectedFinalWatermarkStatus:    &enums.WatermarkStatusSuccess,
		},
		{
			name: "image watermark - fails due to invalid image",
			input: corejobs.WatermarkDocArgs{
				TrustCenterDocumentID: docID,
			},
			mockFileServerResponses: map[string][]byte{
				"/original-file":   samplePDFContent,
				"/watermark-image": sampleImageContent, // This is not a valid image
			},
			getTrustCenterDocResponse: &openlaneclient.GetTrustCenterDocByID{
				TrustCenterDoc: openlaneclient.GetTrustCenterDocByID_TrustCenterDoc{
					ID:            docID,
					FileID:        &fileID,
					TrustCenterID: &trustCenterID,
					OriginalFile: &openlaneclient.GetTrustCenterDocByID_TrustCenterDoc_OriginalFile{
						ID:               originalFileID,
						ProvidedFileName: "test.pdf",
						PresignedURL:     stringPtr("/original-file"),
					},
					File: &openlaneclient.GetTrustCenterDocByID_TrustCenterDoc_File{
						PresignedURL: stringPtr("/original-file"),
					},
				},
			},
			getWatermarkConfigsResponse: &openlaneclient.GetTrustCenterWatermarkConfigs{
				TrustCenterWatermarkConfigs: openlaneclient.GetTrustCenterWatermarkConfigs_TrustCenterWatermarkConfigs{
					Edges: []*openlaneclient.GetTrustCenterWatermarkConfigs_TrustCenterWatermarkConfigs_Edges{
						{
							Node: &openlaneclient.GetTrustCenterWatermarkConfigs_TrustCenterWatermarkConfigs_Edges_Node{
								ID:            watermarkConfigID,
								TrustCenterID: &trustCenterID,
								File: &openlaneclient.GetTrustCenterWatermarkConfigs_TrustCenterWatermarkConfigs_Edges_Node_File{
									PresignedURL: stringPtr("/watermark-image"),
								},
								Opacity:  float64Ptr(0.3),
								Rotation: float64Ptr(0.0),
							},
						},
					},
				},
			},
			updateTrustCenterDocResponses: []*openlaneclient.UpdateTrustCenterDoc{
				{}, // First call to set status to IN_PROGRESS
				{}, // Second call to set status to FAILED
			},
			expectGetTrustCenterDoc:         true,
			expectGetWatermarkConfigs:       true,
			expectUpdateTrustCenterDocCalls: 2,
			expectedError:                   "failed to apply image watermark",
			expectedFinalWatermarkStatus:    &enums.WatermarkStatusFailed,
		},
		{
			name: "text watermark - watermarking fails due to invalid PDF",
			input: corejobs.WatermarkDocArgs{
				TrustCenterDocumentID: docID,
			},
			mockFileServerResponses: map[string][]byte{
				"/original-file": invalidPDFContent,
			},
			getTrustCenterDocResponse: &openlaneclient.GetTrustCenterDocByID{
				TrustCenterDoc: openlaneclient.GetTrustCenterDocByID_TrustCenterDoc{
					ID:            docID,
					FileID:        &fileID,
					TrustCenterID: &trustCenterID,
					OriginalFile: &openlaneclient.GetTrustCenterDocByID_TrustCenterDoc_OriginalFile{
						ID:               originalFileID,
						ProvidedFileName: "test.pdf",
						PresignedURL:     stringPtr("/original-file"),
					},
					File: &openlaneclient.GetTrustCenterDocByID_TrustCenterDoc_File{
						PresignedURL: stringPtr("/original-file"),
					},
				},
			},
			getWatermarkConfigsResponse: &openlaneclient.GetTrustCenterWatermarkConfigs{
				TrustCenterWatermarkConfigs: openlaneclient.GetTrustCenterWatermarkConfigs_TrustCenterWatermarkConfigs{
					Edges: []*openlaneclient.GetTrustCenterWatermarkConfigs_TrustCenterWatermarkConfigs_Edges{
						{
							Node: &openlaneclient.GetTrustCenterWatermarkConfigs_TrustCenterWatermarkConfigs_Edges_Node{
								ID:            watermarkConfigID,
								TrustCenterID: &trustCenterID,
								Text:          stringPtr("CONFIDENTIAL"),
								FontSize:      float64Ptr(12.0),
								Opacity:       float64Ptr(0.5),
								Rotation:      float64Ptr(45.0),
								Color:         stringPtr("red"),
							},
						},
					},
				},
			},
			updateTrustCenterDocResponses: []*openlaneclient.UpdateTrustCenterDoc{
				{}, // First call to set status to IN_PROGRESS
				{}, // Second call to set status to FAILED (from setWatermarkStatus helper)
			},
			expectGetTrustCenterDoc:         true,
			expectGetWatermarkConfigs:       true,
			expectUpdateTrustCenterDocCalls: 2,
			expectedError:                   "failed to apply text watermark",
			expectedFinalWatermarkStatus:    &enums.WatermarkStatusFailed,
		},
		{
			name: "image watermark - watermarking fails due to invalid PDF",
			input: corejobs.WatermarkDocArgs{
				TrustCenterDocumentID: docID,
			},
			mockFileServerResponses: map[string][]byte{
				"/original-file":   invalidPDFContent,
				"/watermark-image": sampleImageContent,
			},
			getTrustCenterDocResponse: &openlaneclient.GetTrustCenterDocByID{
				TrustCenterDoc: openlaneclient.GetTrustCenterDocByID_TrustCenterDoc{
					ID:            docID,
					FileID:        &fileID,
					TrustCenterID: &trustCenterID,
					OriginalFile: &openlaneclient.GetTrustCenterDocByID_TrustCenterDoc_OriginalFile{
						ID:               originalFileID,
						ProvidedFileName: "test.pdf",
						PresignedURL:     stringPtr("/original-file"),
					},
					File: &openlaneclient.GetTrustCenterDocByID_TrustCenterDoc_File{
						PresignedURL: stringPtr("/original-file"),
					},
				},
			},
			getWatermarkConfigsResponse: &openlaneclient.GetTrustCenterWatermarkConfigs{
				TrustCenterWatermarkConfigs: openlaneclient.GetTrustCenterWatermarkConfigs_TrustCenterWatermarkConfigs{
					Edges: []*openlaneclient.GetTrustCenterWatermarkConfigs_TrustCenterWatermarkConfigs_Edges{
						{
							Node: &openlaneclient.GetTrustCenterWatermarkConfigs_TrustCenterWatermarkConfigs_Edges_Node{
								ID:            watermarkConfigID,
								TrustCenterID: &trustCenterID,
								File: &openlaneclient.GetTrustCenterWatermarkConfigs_TrustCenterWatermarkConfigs_Edges_Node_File{
									PresignedURL: stringPtr("/watermark-image"),
								},
								Opacity:  float64Ptr(0.3),
								Rotation: float64Ptr(0.0),
							},
						},
					},
				},
			},
			updateTrustCenterDocResponses: []*openlaneclient.UpdateTrustCenterDoc{
				{}, // First call to set status to IN_PROGRESS
				{}, // Second call to set status to FAILED (from setWatermarkStatus helper)
			},
			expectGetTrustCenterDoc:         true,
			expectGetWatermarkConfigs:       true,
			expectUpdateTrustCenterDocCalls: 2,
			expectedError:                   "failed to apply image watermark",
			expectedFinalWatermarkStatus:    &enums.WatermarkStatusFailed,
		},
		{
			name: "error - failed to get trust center document",
			input: corejobs.WatermarkDocArgs{
				TrustCenterDocumentID: docID,
			},
			getTrustCenterDocError: assert.AnError,
			updateTrustCenterDocResponses: []*openlaneclient.UpdateTrustCenterDoc{
				{}, // First call to set status to IN_PROGRESS
				{}, // Second call to set status to FAILED
			},
			expectGetTrustCenterDoc:         true,
			expectUpdateTrustCenterDocCalls: 2,
			expectedError:                   "failed to fetch trust center document",
			expectedFinalWatermarkStatus:    &enums.WatermarkStatusFailed,
		},
		{
			name: "error - document has no original file",
			input: corejobs.WatermarkDocArgs{
				TrustCenterDocumentID: docID,
			},
			getTrustCenterDocResponse: &openlaneclient.GetTrustCenterDocByID{
				TrustCenterDoc: openlaneclient.GetTrustCenterDocByID_TrustCenterDoc{
					ID:            docID,
					FileID:        &fileID,
					TrustCenterID: &trustCenterID,
					OriginalFile:  nil, // No original file
					File: &openlaneclient.GetTrustCenterDocByID_TrustCenterDoc_File{
						PresignedURL: stringPtr("/original-file"),
					},
				},
			},
			updateTrustCenterDocResponses: []*openlaneclient.UpdateTrustCenterDoc{
				{}, // First call to set status to IN_PROGRESS
				{}, // Second call to set status to FAILED
			},
			expectGetTrustCenterDoc:         true,
			expectUpdateTrustCenterDocCalls: 2,
			expectedError:                   "trust center document has no file or presigned URL",
			expectedFinalWatermarkStatus:    &enums.WatermarkStatusFailed,
		},
		{
			name: "error - document has no trust center ID",
			input: corejobs.WatermarkDocArgs{
				TrustCenterDocumentID: docID,
			},
			getTrustCenterDocResponse: &openlaneclient.GetTrustCenterDocByID{
				TrustCenterDoc: openlaneclient.GetTrustCenterDocByID_TrustCenterDoc{
					ID:            docID,
					FileID:        &fileID,
					TrustCenterID: nil, // No trust center ID
					OriginalFile: &openlaneclient.GetTrustCenterDocByID_TrustCenterDoc_OriginalFile{
						ID:               originalFileID,
						ProvidedFileName: "test.pdf",
						PresignedURL:     stringPtr("/original-file"),
					},
					File: &openlaneclient.GetTrustCenterDocByID_TrustCenterDoc_File{
						PresignedURL: stringPtr("/original-file"),
					},
				},
			},
			updateTrustCenterDocResponses: []*openlaneclient.UpdateTrustCenterDoc{
				{}, // First call to set status to IN_PROGRESS
				{}, // Second call to set status to FAILED
			},
			expectGetTrustCenterDoc:         true,
			expectUpdateTrustCenterDocCalls: 2,
			expectedError:                   "trust center document has no trust center ID",
			expectedFinalWatermarkStatus:    &enums.WatermarkStatusFailed,
		},
		{
			name: "error - no watermark config found",
			input: corejobs.WatermarkDocArgs{
				TrustCenterDocumentID: docID,
			},
			getTrustCenterDocResponse: &openlaneclient.GetTrustCenterDocByID{
				TrustCenterDoc: openlaneclient.GetTrustCenterDocByID_TrustCenterDoc{
					ID:            docID,
					FileID:        &fileID,
					TrustCenterID: &trustCenterID,
					OriginalFile: &openlaneclient.GetTrustCenterDocByID_TrustCenterDoc_OriginalFile{
						ID:               originalFileID,
						ProvidedFileName: "test.pdf",
						PresignedURL:     stringPtr("/original-file"),
					},
					File: &openlaneclient.GetTrustCenterDocByID_TrustCenterDoc_File{
						PresignedURL: stringPtr("/original-file"),
					},
				},
			},
			getWatermarkConfigsResponse: &openlaneclient.GetTrustCenterWatermarkConfigs{
				TrustCenterWatermarkConfigs: openlaneclient.GetTrustCenterWatermarkConfigs_TrustCenterWatermarkConfigs{
					Edges: []*openlaneclient.GetTrustCenterWatermarkConfigs_TrustCenterWatermarkConfigs_Edges{},
				},
			},
			updateTrustCenterDocResponses: []*openlaneclient.UpdateTrustCenterDoc{
				{}, // First call to set status to IN_PROGRESS
				{}, // Second call to set status to FAILED
			},
			expectGetTrustCenterDoc:         true,
			expectGetWatermarkConfigs:       true,
			expectUpdateTrustCenterDocCalls: 2,
			expectedError:                   "no watermark config found for trust center",
			expectedFinalWatermarkStatus:    &enums.WatermarkStatusFailed,
		},
		{
			name: "error - watermark config has neither text nor image",
			input: corejobs.WatermarkDocArgs{
				TrustCenterDocumentID: docID,
			},
			getTrustCenterDocResponse: &openlaneclient.GetTrustCenterDocByID{
				TrustCenterDoc: openlaneclient.GetTrustCenterDocByID_TrustCenterDoc{
					ID:            docID,
					FileID:        &fileID,
					TrustCenterID: &trustCenterID,
					OriginalFile: &openlaneclient.GetTrustCenterDocByID_TrustCenterDoc_OriginalFile{
						ID:               originalFileID,
						ProvidedFileName: "test.pdf",
						PresignedURL:     stringPtr("/original-file"),
					},
					File: &openlaneclient.GetTrustCenterDocByID_TrustCenterDoc_File{
						PresignedURL: stringPtr("/original-file"),
					},
				},
			},
			mockFileServerResponses: map[string][]byte{
				"/original-file": samplePDFContent,
			},
			getWatermarkConfigsResponse: &openlaneclient.GetTrustCenterWatermarkConfigs{
				TrustCenterWatermarkConfigs: openlaneclient.GetTrustCenterWatermarkConfigs_TrustCenterWatermarkConfigs{
					Edges: []*openlaneclient.GetTrustCenterWatermarkConfigs_TrustCenterWatermarkConfigs_Edges{
						{
							Node: &openlaneclient.GetTrustCenterWatermarkConfigs_TrustCenterWatermarkConfigs_Edges_Node{
								ID:            watermarkConfigID,
								TrustCenterID: &trustCenterID,
								// No text or file specified
							},
						},
					},
				},
			},
			updateTrustCenterDocResponses: []*openlaneclient.UpdateTrustCenterDoc{
				{}, // First call to set status to IN_PROGRESS
				{}, // Second call to set status to FAILED
			},
			expectGetTrustCenterDoc:         true,
			expectGetWatermarkConfigs:       true,
			expectUpdateTrustCenterDocCalls: 2,
			expectedError:                   "watermark config has neither text nor image",
			expectedFinalWatermarkStatus:    &enums.WatermarkStatusFailed,
		},
		{
			name: "error - failed to download original file",
			input: corejobs.WatermarkDocArgs{
				TrustCenterDocumentID: docID,
			},
			getTrustCenterDocResponse: &openlaneclient.GetTrustCenterDocByID{
				TrustCenterDoc: openlaneclient.GetTrustCenterDocByID_TrustCenterDoc{
					ID:            docID,
					FileID:        &fileID,
					TrustCenterID: &trustCenterID,
					OriginalFile: &openlaneclient.GetTrustCenterDocByID_TrustCenterDoc_OriginalFile{
						ID:               originalFileID,
						ProvidedFileName: "test.pdf",
						PresignedURL:     stringPtr("/nonexistent-file"), // File doesn't exist
					},
					File: &openlaneclient.GetTrustCenterDocByID_TrustCenterDoc_File{
						PresignedURL: stringPtr("/nonexistent-file"),
					},
				},
			},
			getWatermarkConfigsResponse: &openlaneclient.GetTrustCenterWatermarkConfigs{
				TrustCenterWatermarkConfigs: openlaneclient.GetTrustCenterWatermarkConfigs_TrustCenterWatermarkConfigs{
					Edges: []*openlaneclient.GetTrustCenterWatermarkConfigs_TrustCenterWatermarkConfigs_Edges{
						{
							Node: &openlaneclient.GetTrustCenterWatermarkConfigs_TrustCenterWatermarkConfigs_Edges_Node{
								ID:            watermarkConfigID,
								TrustCenterID: &trustCenterID,
								Text:          stringPtr("CONFIDENTIAL"),
								FontSize:      float64Ptr(12.0),
								Opacity:       float64Ptr(0.5),
								Rotation:      float64Ptr(45.0),
								Color:         stringPtr("red"),
							},
						},
					},
				},
			},
			updateTrustCenterDocResponses: []*openlaneclient.UpdateTrustCenterDoc{
				{}, // First call to set status to IN_PROGRESS
				{}, // Second call to set status to FAILED
			},
			expectGetTrustCenterDoc:         true,
			expectGetWatermarkConfigs:       true,
			expectUpdateTrustCenterDocCalls: 2,
			expectedError:                   "failed to download original document",
			expectedFinalWatermarkStatus:    &enums.WatermarkStatusFailed,
		},
		{
			name: "error - failed to get watermark configs",
			input: corejobs.WatermarkDocArgs{
				TrustCenterDocumentID: docID,
			},
			getTrustCenterDocResponse: &openlaneclient.GetTrustCenterDocByID{
				TrustCenterDoc: openlaneclient.GetTrustCenterDocByID_TrustCenterDoc{
					ID:            docID,
					FileID:        &fileID,
					TrustCenterID: &trustCenterID,
					OriginalFile: &openlaneclient.GetTrustCenterDocByID_TrustCenterDoc_OriginalFile{
						ID:               originalFileID,
						ProvidedFileName: "test.pdf",
						PresignedURL:     stringPtr("/original-file"),
					},
					File: &openlaneclient.GetTrustCenterDocByID_TrustCenterDoc_File{
						PresignedURL: stringPtr("/original-file"),
					},
				},
			},
			getWatermarkConfigsError: assert.AnError,
			updateTrustCenterDocResponses: []*openlaneclient.UpdateTrustCenterDoc{
				{}, // First call to set status to IN_PROGRESS
				{}, // Second call to set status to FAILED
			},
			expectGetTrustCenterDoc:         true,
			expectGetWatermarkConfigs:       true,
			expectUpdateTrustCenterDocCalls: 2,
			expectedError:                   "failed to fetch trust center watermark config",
			expectedFinalWatermarkStatus:    &enums.WatermarkStatusFailed,
		},
		{
			name: "error - failed to download watermark image",
			input: corejobs.WatermarkDocArgs{
				TrustCenterDocumentID: docID,
			},
			mockFileServerResponses: map[string][]byte{
				"/original-file": samplePDFContent,
				// Note: /watermark-image is intentionally missing to cause download failure
			},
			getTrustCenterDocResponse: &openlaneclient.GetTrustCenterDocByID{
				TrustCenterDoc: openlaneclient.GetTrustCenterDocByID_TrustCenterDoc{
					ID:            docID,
					FileID:        &fileID,
					TrustCenterID: &trustCenterID,
					OriginalFile: &openlaneclient.GetTrustCenterDocByID_TrustCenterDoc_OriginalFile{
						ID:               originalFileID,
						ProvidedFileName: "test.pdf",
						PresignedURL:     stringPtr("/original-file"),
					},
					File: &openlaneclient.GetTrustCenterDocByID_TrustCenterDoc_File{
						PresignedURL: stringPtr("/original-file"),
					},
				},
			},
			getWatermarkConfigsResponse: &openlaneclient.GetTrustCenterWatermarkConfigs{
				TrustCenterWatermarkConfigs: openlaneclient.GetTrustCenterWatermarkConfigs_TrustCenterWatermarkConfigs{
					Edges: []*openlaneclient.GetTrustCenterWatermarkConfigs_TrustCenterWatermarkConfigs_Edges{
						{
							Node: &openlaneclient.GetTrustCenterWatermarkConfigs_TrustCenterWatermarkConfigs_Edges_Node{
								ID:            watermarkConfigID,
								TrustCenterID: &trustCenterID,
								File: &openlaneclient.GetTrustCenterWatermarkConfigs_TrustCenterWatermarkConfigs_Edges_Node_File{
									PresignedURL: stringPtr("/watermark-image"), // This will fail to download
								},
								Opacity:  float64Ptr(0.3),
								Rotation: float64Ptr(0.0),
							},
						},
					},
				},
			},
			updateTrustCenterDocResponses: []*openlaneclient.UpdateTrustCenterDoc{
				{}, // First call to set status to IN_PROGRESS
				{}, // Second call to set status to FAILED
			},
			expectGetTrustCenterDoc:         true,
			expectGetWatermarkConfigs:       true,
			expectUpdateTrustCenterDocCalls: 2,
			expectedError:                   "failed to download watermark image",
			expectedFinalWatermarkStatus:    &enums.WatermarkStatusFailed,
		},
		{
			name: "error - failed to update document status to in progress",
			input: corejobs.WatermarkDocArgs{
				TrustCenterDocumentID: docID,
			},
			updateTrustCenterDocErrors:      []error{assert.AnError}, // First call fails
			expectUpdateTrustCenterDocCalls: 1,
			expectedError:                   "failed to set watermark status to in progress",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Setup mock file server if needed
			var mockServer *httptest.Server
			if tc.mockFileServerResponses != nil {
				mockServer = mockFileServer(t, tc.mockFileServerResponses)
				defer mockServer.Close()

				// Update URLs in responses to use mock server
				if tc.getTrustCenterDocResponse != nil && tc.getTrustCenterDocResponse.TrustCenterDoc.OriginalFile != nil {
					if tc.getTrustCenterDocResponse.TrustCenterDoc.OriginalFile.PresignedURL != nil {
						originalURL := *tc.getTrustCenterDocResponse.TrustCenterDoc.OriginalFile.PresignedURL
						newURL := mockServer.URL + originalURL
						tc.getTrustCenterDocResponse.TrustCenterDoc.OriginalFile.PresignedURL = &newURL
					}
				}
				if tc.getTrustCenterDocResponse != nil {
					if tc.getTrustCenterDocResponse.TrustCenterDoc.File.PresignedURL != nil {
						originalURL := *tc.getTrustCenterDocResponse.TrustCenterDoc.File.PresignedURL
						newURL := mockServer.URL + originalURL
						tc.getTrustCenterDocResponse.TrustCenterDoc.File.PresignedURL = &newURL
					}
				}
				if tc.getWatermarkConfigsResponse != nil {
					for i := range tc.getWatermarkConfigsResponse.TrustCenterWatermarkConfigs.Edges {
						edge := tc.getWatermarkConfigsResponse.TrustCenterWatermarkConfigs.Edges[i]
						if edge.Node != nil && edge.Node.File != nil && edge.Node.File.PresignedURL != nil {
							originalURL := *edge.Node.File.PresignedURL
							newURL := mockServer.URL + originalURL
							edge.Node.File.PresignedURL = &newURL
						}
					}
				}
			}

			olMock := olmocks.NewMockOpenlaneGraphClient(t)

			// Setup expectations
			if tc.expectGetTrustCenterDoc {
				olMock.EXPECT().GetTrustCenterDocByID(
					mock.MatchedBy(func(ctx context.Context) bool { return ctx != nil }),
					tc.input.TrustCenterDocumentID,
				).Return(tc.getTrustCenterDocResponse, tc.getTrustCenterDocError)
			}

			if tc.expectGetWatermarkConfigs {
				olMock.EXPECT().GetTrustCenterWatermarkConfigs(
					mock.MatchedBy(func(ctx context.Context) bool { return ctx != nil }),
					(*int64)(nil),
					(*int64)(nil),
					mock.MatchedBy(func(where *openlaneclient.TrustCenterWatermarkConfigWhereInput) bool {
						return where != nil && where.TrustCenterID != nil
					}),
				).Return(tc.getWatermarkConfigsResponse, tc.getWatermarkConfigsError)
			}

			// Setup UpdateTrustCenterDoc expectations
			for i := 0; i < tc.expectUpdateTrustCenterDocCalls; i++ {
				var expectedStatus enums.WatermarkStatus

				if i == 0 {
					// First call should set status to IN_PROGRESS
					expectedStatus = enums.WatermarkStatusInProgress
				} else if i == tc.expectUpdateTrustCenterDocCalls-1 {
					// Last call should set the final status
					if tc.expectedFinalWatermarkStatus != nil {
						expectedStatus = *tc.expectedFinalWatermarkStatus
					}
				}

				var err error
				if i < len(tc.updateTrustCenterDocErrors) {
					err = tc.updateTrustCenterDocErrors[i]
				}

				var response *openlaneclient.UpdateTrustCenterDoc
				if i < len(tc.updateTrustCenterDocResponses) {
					response = tc.updateTrustCenterDocResponses[i]
				} else {
					response = &openlaneclient.UpdateTrustCenterDoc{}
				}

				if expectedStatus == enums.WatermarkStatusSuccess {
					// Success call should include a file upload
					olMock.EXPECT().UpdateTrustCenterDoc(
						mock.MatchedBy(func(ctx context.Context) bool { return ctx != nil }),
						tc.input.TrustCenterDocumentID,
						mock.MatchedBy(func(input openlaneclient.UpdateTrustCenterDocInput) bool {
							return input.WatermarkStatus != nil && *input.WatermarkStatus == expectedStatus
						}),
						(*graphql.Upload)(nil), // trustCenterDocFile
						mock.MatchedBy(func(file *graphql.Upload) bool { return file != nil }), // watermarkedTrustCenterDocFile
					).Return(response, err)
				} else {
					// Other calls should not include a file upload
					olMock.EXPECT().UpdateTrustCenterDoc(
						mock.MatchedBy(func(ctx context.Context) bool { return ctx != nil }),
						tc.input.TrustCenterDocumentID,
						mock.MatchedBy(func(input openlaneclient.UpdateTrustCenterDocInput) bool {
							return input.WatermarkStatus != nil && *input.WatermarkStatus == expectedStatus
						}),
						(*graphql.Upload)(nil), // trustCenterDocFile
						(*graphql.Upload)(nil), // watermarkedTrustCenterDocFile
					).Return(response, err)
				}
			}

			// Create worker
			worker := &corejobs.WatermarkDocWorker{
				Config: corejobs.WatermarkWorkerConfig{
					OpenlaneConfig: corejobs.OpenlaneConfig{
						OpenlaneAPIHost:  "https://api.example.com",
						OpenlaneAPIToken: "test-token",
					},
					Enabled: true,
				},
			}
			worker.WithOpenlaneClient(olMock)

			// Execute test
			ctx := context.Background()
			err := worker.Work(ctx, &river.Job[corejobs.WatermarkDocArgs]{Args: tc.input})

			// Verify results
			if tc.expectedError != "" {
				require.Error(t, err)
				require.Contains(t, err.Error(), tc.expectedError)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

// Helper functions for creating pointers
func stringPtr(s string) *string {
	return &s
}

func float64Ptr(f float64) *float64 {
	return &f
}
