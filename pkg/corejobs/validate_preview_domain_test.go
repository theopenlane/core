package corejobs_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/cloudflare/cloudflare-go/v6/dns"
	"github.com/cloudflare/cloudflare-go/v6/packages/pagination"
	"github.com/riverqueue/river"
	"github.com/riverqueue/river/rivertype"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/theopenlane/core/pkg/corejobs"
	"github.com/theopenlane/core/pkg/enums"
	"github.com/theopenlane/core/pkg/openlaneclient"

	cfmocks "github.com/theopenlane/core/pkg/corejobs/internal/cloudflare/mocks"
	olmocks "github.com/theopenlane/core/pkg/corejobs/internal/olclient/mocks"
)

func TestValidatePreviewDomainWorker(t *testing.T) {
	t.Parallel()

	trustCenterID := "trustcenterid123"
	zoneID := "cfzoneid123"
	dnsTxtRecord := "_acme-challenge.preview.example.com"
	dnsTxtValue := "verification-value-123"
	recordID := "recordid123"

	testCases := []struct {
		name                      string
		trustCenterID             string
		trustCenterPreviewZoneID  string
		metadata                  []byte
		getTrustCenterResponse    *openlaneclient.GetTrustCenterByID
		getTrustCenterError       error
		listRecordsResponse       *pagination.V4PagePaginationArray[dns.RecordResponse]
		listRecordsError          error
		createRecordResponse      *dns.RecordResponse
		createRecordError         error
		updateTrustCenterResponse *openlaneclient.UpdateTrustCenter
		updateTrustCenterError    error
		expectedError             string
		expectSnooze              bool
		expectMaxSnoozesError     bool
	}{
		{
			name:                     "missing trust center id",
			trustCenterPreviewZoneID: zoneID,
			expectedError:            "trust_center_id is required for the create_preview_domain job",
		},
		{
			name:          "missing trust center preview zone id",
			trustCenterID: trustCenterID,
			expectedError: "trust_center_preview_zone_id is required for the create_preview_domain job",
		},
		{
			name:                     "get trust center error",
			trustCenterID:            trustCenterID,
			trustCenterPreviewZoneID: zoneID,
			getTrustCenterError:      errors.New("trust center not found"),
			expectedError:            "trust center not found",
		},
		{
			name:                     "missing preview domain",
			trustCenterID:            trustCenterID,
			trustCenterPreviewZoneID: zoneID,
			getTrustCenterResponse: &openlaneclient.GetTrustCenterByID{
				TrustCenter: openlaneclient.GetTrustCenterByID_TrustCenter{
					ID:            trustCenterID,
					PreviewDomain: nil,
				},
			},
			expectedError: "preview_domain is required for the create_preview_domain job",
		},
		{
			name:                     "dns verification nil - snooze",
			trustCenterID:            trustCenterID,
			trustCenterPreviewZoneID: zoneID,
			getTrustCenterResponse: &openlaneclient.GetTrustCenterByID{
				TrustCenter: openlaneclient.GetTrustCenterByID_TrustCenter{
					ID: trustCenterID,
					PreviewDomain: &openlaneclient.GetTrustCenterByID_TrustCenter_PreviewDomain{
						CnameRecord:     "preview.example.com",
						DNSVerification: nil,
					},
				},
			},
			expectSnooze: true,
		},
		{
			name:                     "dns verification nil - max snoozes reached",
			trustCenterID:            trustCenterID,
			trustCenterPreviewZoneID: zoneID,
			metadata:                 []byte(`{"snoozes": 30}`),
			getTrustCenterResponse: &openlaneclient.GetTrustCenterByID{
				TrustCenter: openlaneclient.GetTrustCenterByID_TrustCenter{
					ID: trustCenterID,
					PreviewDomain: &openlaneclient.GetTrustCenterByID_TrustCenter_PreviewDomain{
						CnameRecord:     "preview.example.com",
						DNSVerification: nil,
					},
				},
			},
			expectMaxSnoozesError: true,
		},
		{
			name:                     "txt record does not exist - create it",
			trustCenterID:            trustCenterID,
			trustCenterPreviewZoneID: zoneID,
			getTrustCenterResponse: &openlaneclient.GetTrustCenterByID{
				TrustCenter: openlaneclient.GetTrustCenterByID_TrustCenter{
					ID: trustCenterID,
					PreviewDomain: &openlaneclient.GetTrustCenterByID_TrustCenter_PreviewDomain{
						CnameRecord: "preview.example.com",
						DNSVerification: &openlaneclient.GetTrustCenterByID_TrustCenter_PreviewDomain_DNSVerification{
							DNSTxtRecord:          dnsTxtRecord,
							DNSTxtValue:           dnsTxtValue,
							DNSVerificationStatus: enums.DNSVerificationStatusPending,
							AcmeChallengeStatus:   enums.SSLVerificationStatusPendingValidation,
						},
					},
				},
			},
			listRecordsResponse: &pagination.V4PagePaginationArray[dns.RecordResponse]{
				Result: []dns.RecordResponse{},
			},
			createRecordResponse: &dns.RecordResponse{
				ID:      recordID,
				Name:    dnsTxtRecord,
				Content: dnsTxtValue,
			},
		},
		{
			name:                     "txt record exists - verification active - update trust center",
			trustCenterID:            trustCenterID,
			trustCenterPreviewZoneID: zoneID,
			getTrustCenterResponse: &openlaneclient.GetTrustCenterByID{
				TrustCenter: openlaneclient.GetTrustCenterByID_TrustCenter{
					ID: trustCenterID,
					PreviewDomain: &openlaneclient.GetTrustCenterByID_TrustCenter_PreviewDomain{
						CnameRecord: "preview.example.com",
						DNSVerification: &openlaneclient.GetTrustCenterByID_TrustCenter_PreviewDomain_DNSVerification{
							DNSTxtRecord:          dnsTxtRecord,
							DNSTxtValue:           dnsTxtValue,
							DNSVerificationStatus: enums.DNSVerificationStatusActive,
							AcmeChallengeStatus:   enums.SSLVerificationStatusActive,
						},
					},
				},
			},
			listRecordsResponse: &pagination.V4PagePaginationArray[dns.RecordResponse]{
				Result: []dns.RecordResponse{
					{
						ID:      recordID,
						Name:    dnsTxtRecord,
						Content: dnsTxtValue,
						Type:    "TXT",
					},
				},
			},
			updateTrustCenterResponse: &openlaneclient.UpdateTrustCenter{},
		},
		{
			name:                     "txt record exists - verification pending - snooze",
			trustCenterID:            trustCenterID,
			trustCenterPreviewZoneID: zoneID,
			getTrustCenterResponse: &openlaneclient.GetTrustCenterByID{
				TrustCenter: openlaneclient.GetTrustCenterByID_TrustCenter{
					ID: trustCenterID,
					PreviewDomain: &openlaneclient.GetTrustCenterByID_TrustCenter_PreviewDomain{
						CnameRecord: "preview.example.com",
						DNSVerification: &openlaneclient.GetTrustCenterByID_TrustCenter_PreviewDomain_DNSVerification{
							DNSTxtRecord:          dnsTxtRecord,
							DNSTxtValue:           dnsTxtValue,
							DNSVerificationStatus: enums.DNSVerificationStatusPending,
							AcmeChallengeStatus:   enums.SSLVerificationStatusPendingValidation,
						},
					},
				},
			},
			listRecordsResponse: &pagination.V4PagePaginationArray[dns.RecordResponse]{
				Result: []dns.RecordResponse{
					{
						ID:      recordID,
						Name:    dnsTxtRecord,
						Content: dnsTxtValue,
						Type:    "TXT",
					},
				},
			},
			expectSnooze: true,
		},
		{
			name:                     "txt record exists - verification pending - max snoozes reached",
			trustCenterID:            trustCenterID,
			trustCenterPreviewZoneID: zoneID,
			metadata:                 []byte(`{"snoozes": 30}`),
			getTrustCenterResponse: &openlaneclient.GetTrustCenterByID{
				TrustCenter: openlaneclient.GetTrustCenterByID_TrustCenter{
					ID: trustCenterID,
					PreviewDomain: &openlaneclient.GetTrustCenterByID_TrustCenter_PreviewDomain{
						CnameRecord: "preview.example.com",
						DNSVerification: &openlaneclient.GetTrustCenterByID_TrustCenter_PreviewDomain_DNSVerification{
							DNSTxtRecord:          dnsTxtRecord,
							DNSTxtValue:           dnsTxtValue,
							DNSVerificationStatus: enums.DNSVerificationStatusPending,
							AcmeChallengeStatus:   enums.SSLVerificationStatusPendingValidation,
						},
					},
				},
			},
			listRecordsResponse: &pagination.V4PagePaginationArray[dns.RecordResponse]{
				Result: []dns.RecordResponse{
					{
						ID:      recordID,
						Name:    dnsTxtRecord,
						Content: dnsTxtValue,
						Type:    "TXT",
					},
				},
			},
			expectMaxSnoozesError: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctx := context.Background()
			cfMock := cfmocks.NewMockClient(t)
			olMock := olmocks.NewMockOpenlaneGraphClient(t)

			// Setup mocks
			if tc.trustCenterID != "" && tc.trustCenterPreviewZoneID != "" {
				if tc.getTrustCenterError != nil {
					olMock.EXPECT().GetTrustCenterByID(mock.Anything, tc.trustCenterID).Return(nil, tc.getTrustCenterError)
				} else if tc.getTrustCenterResponse != nil {
					olMock.EXPECT().GetTrustCenterByID(mock.Anything, tc.trustCenterID).Return(tc.getTrustCenterResponse, nil)

					// If we have DNS verification, setup cloudflare mocks
					if tc.getTrustCenterResponse.TrustCenter.PreviewDomain != nil &&
						tc.getTrustCenterResponse.TrustCenter.PreviewDomain.DNSVerification != nil {

						cfRecordsMock := cfmocks.NewMockRecordService(t)
						cfMock.EXPECT().Record().Return(cfRecordsMock)

						if tc.listRecordsError != nil {
							cfRecordsMock.EXPECT().List(mock.Anything, mock.MatchedBy(func(params dns.RecordListParams) bool {
								return params.ZoneID.Value == tc.trustCenterPreviewZoneID
							})).Return(nil, tc.listRecordsError)
						} else if tc.listRecordsResponse != nil {
							cfRecordsMock.EXPECT().List(mock.Anything, mock.MatchedBy(func(params dns.RecordListParams) bool {
								return params.ZoneID.Value == tc.trustCenterPreviewZoneID
							})).Return(tc.listRecordsResponse, nil)

							// If TXT record doesn't exist, expect create call
							txtRecordExists := false
							for _, record := range tc.listRecordsResponse.Result {
								if string(record.Type) == "TXT" && record.Content == tc.getTrustCenterResponse.TrustCenter.PreviewDomain.DNSVerification.DNSTxtValue {
									txtRecordExists = true
									break
								}
							}

							if !txtRecordExists && tc.createRecordResponse != nil {
								if tc.createRecordError != nil {
									cfRecordsMock.EXPECT().New(mock.Anything, mock.MatchedBy(func(params dns.RecordNewParams) bool {
										return params.ZoneID.Value == tc.trustCenterPreviewZoneID
									})).Return(nil, tc.createRecordError)
								} else {
									cfRecordsMock.EXPECT().New(mock.Anything, mock.MatchedBy(func(params dns.RecordNewParams) bool {
										return params.ZoneID.Value == tc.trustCenterPreviewZoneID
									})).Return(tc.createRecordResponse, nil)
								}
							}

							// If verification is active, expect update trust center call
							dnsVerification := tc.getTrustCenterResponse.TrustCenter.PreviewDomain.DNSVerification
							if txtRecordExists &&
								dnsVerification.DNSVerificationStatus == enums.DNSVerificationStatusActive &&
								dnsVerification.AcmeChallengeStatus == enums.SSLVerificationStatusActive {

								if tc.updateTrustCenterError != nil {
									olMock.EXPECT().UpdateTrustCenter(mock.Anything, tc.trustCenterID, mock.MatchedBy(func(input openlaneclient.UpdateTrustCenterInput) bool {
										return input.PreviewStatus != nil && *input.PreviewStatus == enums.TrustCenterPreviewStatusReady
									})).Return(nil, tc.updateTrustCenterError)
								} else if tc.updateTrustCenterResponse != nil {
									olMock.EXPECT().UpdateTrustCenter(mock.Anything, tc.trustCenterID, mock.MatchedBy(func(input openlaneclient.UpdateTrustCenterInput) bool {
										return input.PreviewStatus != nil && *input.PreviewStatus == enums.TrustCenterPreviewStatusReady
									})).Return(tc.updateTrustCenterResponse, nil)
								}
							}
						}
					}
				}
			}

			worker := &corejobs.ValidatePreviewDomainWorker{
				Config: corejobs.ValidatePreviewDomainConfig{
					CloudflareAPIKey: "test",
					MaxSnoozes:       30,
					SnoozeDuration:   5 * time.Second,
				},
			}

			worker.WithCloudflareClient(cfMock)
			worker.WithOpenlaneClient(olMock)

			metadata := tc.metadata
			if metadata == nil {
				metadata = []byte(`{}`)
			}

			job := &river.Job[corejobs.ValidatePreviewDomainArgs]{
				JobRow: &rivertype.JobRow{
					Metadata: metadata,
				},
				Args: corejobs.ValidatePreviewDomainArgs{
					TrustCenterID:            tc.trustCenterID,
					TrustCenterPreviewZoneID: tc.trustCenterPreviewZoneID,
				},
			}

			err := worker.Work(ctx, job)

			if tc.expectedError != "" {
				require.Error(t, err)
				require.Contains(t, err.Error(), tc.expectedError)
			} else if tc.expectSnooze {
				require.Error(t, err)
				require.Contains(t, err.Error(), "JobSnoozeError")
			} else if tc.expectMaxSnoozesError {
				require.Error(t, err)
				require.Equal(t, corejobs.ErrMaxSnoozesReached, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}
