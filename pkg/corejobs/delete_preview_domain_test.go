package corejobs_test

import (
	"context"
	"errors"
	"testing"

	"github.com/cloudflare/cloudflare-go/v6/dns"
	"github.com/cloudflare/cloudflare-go/v6/packages/pagination"
	"github.com/riverqueue/river"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/theopenlane/core/pkg/corejobs"
	cfmocks "github.com/theopenlane/core/pkg/corejobs/internal/cloudflare/mocks"
	olmocks "github.com/theopenlane/core/pkg/corejobs/internal/olclient/mocks"
	"github.com/theopenlane/core/pkg/openlaneclient"
)

func TestDeletePreviewDomainWorker_Work(t *testing.T) {
	t.Parallel()

	customDomainID := "customdomain123"
	zoneID := "zone123"
	cnameRecord := "preview.example.com"
	recordID := "record123"
	mockErr := errors.New("mock error")

	testCases := []struct {
		name                     string
		input                    corejobs.DeletePreviewDomainArgs
		getCustomDomainResponse  *openlaneclient.GetCustomDomainByID
		getCustomDomainError     error
		listRecordsResponse      *pagination.V4PagePaginationArray[dns.RecordResponse]
		listRecordsError         error
		deleteRecordError        error
		deleteCustomDomainError  error
		expectedError            string
		expectGetCustomDomain    bool
		expectListRecords        bool
		expectDeleteRecord       bool
		expectDeleteCustomDomain bool
	}{
		{
			name: "happy path - delete preview domain with CNAME record",
			input: corejobs.DeletePreviewDomainArgs{
				CustomDomainID:           customDomainID,
				TrustCenterPreviewZoneID: zoneID,
			},
			getCustomDomainResponse: &openlaneclient.GetCustomDomainByID{
				CustomDomain: openlaneclient.GetCustomDomainByID_CustomDomain{
					ID:          customDomainID,
					CnameRecord: cnameRecord,
				},
			},
			listRecordsResponse: &pagination.V4PagePaginationArray[dns.RecordResponse]{
				Result: []dns.RecordResponse{
					{
						ID:   recordID,
						Type: "CNAME",
					},
				},
			},
			expectGetCustomDomain:    true,
			expectListRecords:        true,
			expectDeleteRecord:       true,
			expectDeleteCustomDomain: true,
		},
		{
			name: "happy path - no CNAME records found",
			input: corejobs.DeletePreviewDomainArgs{
				CustomDomainID:           customDomainID,
				TrustCenterPreviewZoneID: zoneID,
			},
			getCustomDomainResponse: &openlaneclient.GetCustomDomainByID{
				CustomDomain: openlaneclient.GetCustomDomainByID_CustomDomain{
					ID:          customDomainID,
					CnameRecord: cnameRecord,
				},
			},
			listRecordsResponse: &pagination.V4PagePaginationArray[dns.RecordResponse]{
				Result: []dns.RecordResponse{},
			},
			expectGetCustomDomain:    true,
			expectListRecords:        true,
			expectDeleteRecord:       false,
			expectDeleteCustomDomain: true,
		},
		{
			name: "happy path - non-CNAME records ignored",
			input: corejobs.DeletePreviewDomainArgs{
				CustomDomainID:           customDomainID,
				TrustCenterPreviewZoneID: zoneID,
			},
			getCustomDomainResponse: &openlaneclient.GetCustomDomainByID{
				CustomDomain: openlaneclient.GetCustomDomainByID_CustomDomain{
					ID:          customDomainID,
					CnameRecord: cnameRecord,
				},
			},
			listRecordsResponse: &pagination.V4PagePaginationArray[dns.RecordResponse]{
				Result: []dns.RecordResponse{
					{
						ID:   "record-a",
						Type: "A",
					},
					{
						ID:   "record-txt",
						Type: "TXT",
					},
				},
			},
			expectGetCustomDomain:    true,
			expectListRecords:        true,
			expectDeleteRecord:       false,
			expectDeleteCustomDomain: true,
		},
		{
			name: "happy path - multiple CNAME records deleted",
			input: corejobs.DeletePreviewDomainArgs{
				CustomDomainID:           customDomainID,
				TrustCenterPreviewZoneID: zoneID,
			},
			getCustomDomainResponse: &openlaneclient.GetCustomDomainByID{
				CustomDomain: openlaneclient.GetCustomDomainByID_CustomDomain{
					ID:          customDomainID,
					CnameRecord: cnameRecord,
				},
			},
			listRecordsResponse: &pagination.V4PagePaginationArray[dns.RecordResponse]{
				Result: []dns.RecordResponse{
					{
						ID:   recordID,
						Type: "CNAME",
					},
					{
						ID:   "record-cname-2",
						Type: "CNAME",
					},
				},
			},
			expectGetCustomDomain:    true,
			expectListRecords:        true,
			expectDeleteRecord:       true,
			expectDeleteCustomDomain: true,
		},
		{
			name: "missing custom_domain_id",
			input: corejobs.DeletePreviewDomainArgs{
				CustomDomainID:           "",
				TrustCenterPreviewZoneID: zoneID,
			},
			expectedError:            "custom_domain_id is required",
			expectGetCustomDomain:    false,
			expectListRecords:        false,
			expectDeleteRecord:       false,
			expectDeleteCustomDomain: false,
		},
		{
			name: "missing trust_center_preview_zone_id",
			input: corejobs.DeletePreviewDomainArgs{
				CustomDomainID:           customDomainID,
				TrustCenterPreviewZoneID: "",
			},
			expectedError:            "trust_center_preview_zone_id is required",
			expectGetCustomDomain:    false,
			expectListRecords:        false,
			expectDeleteRecord:       false,
			expectDeleteCustomDomain: false,
		},
		{
			name: "error getting custom domain",
			input: corejobs.DeletePreviewDomainArgs{
				CustomDomainID:           customDomainID,
				TrustCenterPreviewZoneID: zoneID,
			},
			getCustomDomainError:     mockErr,
			expectedError:            "mock error",
			expectGetCustomDomain:    true,
			expectListRecords:        false,
			expectDeleteRecord:       false,
			expectDeleteCustomDomain: false,
		},
		{
			name: "error listing DNS records",
			input: corejobs.DeletePreviewDomainArgs{
				CustomDomainID:           customDomainID,
				TrustCenterPreviewZoneID: zoneID,
			},
			getCustomDomainResponse: &openlaneclient.GetCustomDomainByID{
				CustomDomain: openlaneclient.GetCustomDomainByID_CustomDomain{
					ID:          customDomainID,
					CnameRecord: cnameRecord,
				},
			},
			listRecordsError:         mockErr,
			expectedError:            "mock error",
			expectGetCustomDomain:    true,
			expectListRecords:        true,
			expectDeleteRecord:       false,
			expectDeleteCustomDomain: false,
		},
		{
			name: "error deleting DNS record",
			input: corejobs.DeletePreviewDomainArgs{
				CustomDomainID:           customDomainID,
				TrustCenterPreviewZoneID: zoneID,
			},
			getCustomDomainResponse: &openlaneclient.GetCustomDomainByID{
				CustomDomain: openlaneclient.GetCustomDomainByID_CustomDomain{
					ID:          customDomainID,
					CnameRecord: cnameRecord,
				},
			},
			listRecordsResponse: &pagination.V4PagePaginationArray[dns.RecordResponse]{
				Result: []dns.RecordResponse{
					{
						ID:   recordID,
						Type: "CNAME",
					},
				},
			},
			deleteRecordError:        mockErr,
			expectedError:            "mock error",
			expectGetCustomDomain:    true,
			expectListRecords:        true,
			expectDeleteRecord:       true,
			expectDeleteCustomDomain: false,
		},
		{
			name: "error deleting custom domain",
			input: corejobs.DeletePreviewDomainArgs{
				CustomDomainID:           customDomainID,
				TrustCenterPreviewZoneID: zoneID,
			},
			getCustomDomainResponse: &openlaneclient.GetCustomDomainByID{
				CustomDomain: openlaneclient.GetCustomDomainByID_CustomDomain{
					ID:          customDomainID,
					CnameRecord: cnameRecord,
				},
			},
			listRecordsResponse: &pagination.V4PagePaginationArray[dns.RecordResponse]{
				Result: []dns.RecordResponse{
					{
						ID:   recordID,
						Type: "CNAME",
					},
				},
			},
			deleteCustomDomainError:  mockErr,
			expectedError:            "mock error",
			expectGetCustomDomain:    true,
			expectListRecords:        true,
			expectDeleteRecord:       true,
			expectDeleteCustomDomain: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			ctx := context.Background()
			cfMock := cfmocks.NewMockClient(t)
			cfRecordMock := cfmocks.NewMockRecordService(t)
			olMock := olmocks.NewMockOpenlaneGraphClient(t)

			// Setup Cloudflare Record service mock
			if tc.expectListRecords || tc.expectDeleteRecord {
				cfMock.EXPECT().Record().Return(cfRecordMock)
			}

			// Setup GetCustomDomainByID mock
			if tc.expectGetCustomDomain {
				olMock.EXPECT().GetCustomDomainByID(
					mock.MatchedBy(func(ctx context.Context) bool {
						return ctx != nil
					}),
					tc.input.CustomDomainID,
				).Return(tc.getCustomDomainResponse, tc.getCustomDomainError)
			}

			// Setup List DNS records mock
			if tc.expectListRecords {
				cfRecordMock.EXPECT().List(
					mock.MatchedBy(func(ctx context.Context) bool {
						return ctx != nil
					}),
					mock.Anything,
				).Return(tc.listRecordsResponse, tc.listRecordsError)
			}

			// Setup Delete DNS record mock
			if tc.expectDeleteRecord {
				// Count how many CNAME records need to be deleted
				cnameCount := 0
				if tc.listRecordsResponse != nil {
					for _, record := range tc.listRecordsResponse.Result {
						if string(record.Type) == "CNAME" {
							cnameCount++
						}
					}
				}

				// Expect Delete to be called for each CNAME record
				for i := 0; i < cnameCount; i++ {
					cfRecordMock.EXPECT().Delete(
						mock.MatchedBy(func(ctx context.Context) bool {
							return ctx != nil
						}),
						mock.Anything, // Accept any record ID
						mock.Anything,
					).Return(&dns.RecordDeleteResponse{}, tc.deleteRecordError)
				}
			}

			// Setup DeleteCustomDomain mock
			if tc.expectDeleteCustomDomain {
				olMock.EXPECT().DeleteCustomDomain(
					mock.MatchedBy(func(ctx context.Context) bool {
						return ctx != nil
					}),
					tc.input.CustomDomainID,
				).Return(&openlaneclient.DeleteCustomDomain{
					DeleteCustomDomain: openlaneclient.DeleteCustomDomain_DeleteCustomDomain{
						DeletedID: customDomainID,
					},
				}, tc.deleteCustomDomainError)
			}

			// Create worker with config
			worker := &corejobs.DeletePreviewDomainWorker{
				Config: corejobs.DeletePreviewDomainConfig{
					OpenlaneConfig: corejobs.OpenlaneConfig{
						OpenlaneAPIHost:  "https://api.example.com",
						OpenlaneAPIToken: "tola_test-token",
					},
					Enabled:          true,
					CloudflareAPIKey: "test-cf-key",
				},
			}

			// Inject mock clients
			worker.WithCloudflareClient(cfMock)
			worker.WithOpenlaneClient(olMock)

			// Execute
			err := worker.Work(ctx, &river.Job[corejobs.DeletePreviewDomainArgs]{
				Args: tc.input,
			})

			// Assert
			if tc.expectedError != "" {
				require.Error(t, err)
				require.Contains(t, err.Error(), tc.expectedError)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestDeletePreviewDomainWorker_WithCloudflareClient(t *testing.T) {
	t.Parallel()

	cfMock := cfmocks.NewMockClient(t)
	worker := &corejobs.DeletePreviewDomainWorker{}

	result := worker.WithCloudflareClient(cfMock)

	require.Equal(t, worker, result, "WithCloudflareClient should return the same worker instance")
}

func TestDeletePreviewDomainWorker_WithOpenlaneClient(t *testing.T) {
	t.Parallel()

	olMock := olmocks.NewMockOpenlaneGraphClient(t)
	worker := &corejobs.DeletePreviewDomainWorker{}

	result := worker.WithOpenlaneClient(olMock)

	require.Equal(t, worker, result, "WithOpenlaneClient should return the same worker instance")
}

func TestDeletePreviewDomainArgs_Kind(t *testing.T) {
	t.Parallel()

	args := corejobs.DeletePreviewDomainArgs{}
	require.Equal(t, "delete_preview_domain", args.Kind())
}
