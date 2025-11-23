package corejobs_test

import (
	"context"
	"testing"
	"time"

	"github.com/cloudflare/cloudflare-go/v6/custom_hostnames"
	"github.com/cloudflare/cloudflare-go/v6/dns"
	"github.com/cloudflare/cloudflare-go/v6/packages/pagination"
	"github.com/riverqueue/river"
	"github.com/riverqueue/river/rivertype"
	"github.com/samber/lo"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/theopenlane/core/pkg/corejobs"
	"github.com/theopenlane/core/pkg/enums"
	"github.com/theopenlane/core/pkg/openlaneclient"

	cfmocks "github.com/theopenlane/core/pkg/corejobs/internal/cloudflare/mocks"
	olmocks "github.com/theopenlane/core/pkg/corejobs/internal/olclient/mocks"
)

func TestValidateCustomDomainWorker(t *testing.T) {
	t.Parallel()

	customDomainID := "customdomainid123"
	mappableDomainID := "mappabldomain123"
	ownerID := "owner123"
	cnameRecord := "trust.meow.io"
	zoneID := "cfzoneid123"
	cfHostnameID := "cfhostnameid123"
	domainVerificationID := "domainverificationid123"
	acmeChallengePath := "acmechallengepath123"
	acmeChallengeValue := "acmechallengevalue123"

	testCases := []struct {
		name                          string
		input                         corejobs.ValidateCustomDomainArgs
		getCustomDomainResponse       *openlaneclient.GetCustomDomainByID
		getCustomHostnameResponse     *custom_hostnames.CustomHostnameGetResponse
		expectedUpdateDNSVerification *openlaneclient.UpdateDNSVerificationInput
		expectedError                 string
	}{
		{
			name: "happy path - update DNS verification with ACME challenge",
			input: corejobs.ValidateCustomDomainArgs{
				CustomDomainID: customDomainID,
			},
			getCustomDomainResponse: &openlaneclient.GetCustomDomainByID{
				CustomDomain: openlaneclient.GetCustomDomainByID_CustomDomain{
					CnameRecord:       cnameRecord,
					DNSVerificationID: lo.ToPtr(domainVerificationID),
					ID:                customDomainID,
					MappableDomainID:  mappableDomainID,
					OwnerID:           &ownerID,
					MappableDomain: openlaneclient.GetCustomDomainByID_CustomDomain_MappableDomain{
						ID:     mappableDomainID,
						ZoneID: zoneID,
					},
					DNSVerification: &openlaneclient.GetCustomDomainByID_CustomDomain_DNSVerification{
						ID:                         domainVerificationID,
						CloudflareHostnameID:       cfHostnameID,
						AcmeChallengePath:          nil,
						AcmeChallengeStatus:        enums.SSLVerificationStatusInitializing,
						DNSVerificationStatus:      enums.DNSVerificationStatusPending,
						ExpectedAcmeChallengeValue: nil,
					},
				},
			},
			getCustomHostnameResponse: &custom_hostnames.CustomHostnameGetResponse{
				ID:     cfHostnameID,
				Status: "active",
				SSL: custom_hostnames.CustomHostnameGetResponseSSL{
					Status: "active",
					ValidationRecords: []custom_hostnames.CustomHostnameGetResponseSSLValidationRecord{
						{
							HTTPURL:  "http://trust.meow.io/.well-known/acme-challenge/" + acmeChallengePath,
							HTTPBody: acmeChallengeValue,
						},
					},
				},
			},
			expectedUpdateDNSVerification: &openlaneclient.UpdateDNSVerificationInput{
				AcmeChallengePath:          &acmeChallengePath,
				ExpectedAcmeChallengeValue: &acmeChallengeValue,
				AcmeChallengeStatus:        lo.ToPtr(enums.SSLVerificationStatusActive),
				DNSVerificationStatus:      lo.ToPtr(enums.DNSVerificationStatusActive),
			},
		},
		{
			name: "happy path - no updates needed",
			input: corejobs.ValidateCustomDomainArgs{
				CustomDomainID: customDomainID,
			},
			getCustomDomainResponse: &openlaneclient.GetCustomDomainByID{
				CustomDomain: openlaneclient.GetCustomDomainByID_CustomDomain{
					CnameRecord:       cnameRecord,
					DNSVerificationID: lo.ToPtr(domainVerificationID),
					ID:                customDomainID,
					MappableDomainID:  mappableDomainID,
					OwnerID:           &ownerID,
					DNSVerification: &openlaneclient.GetCustomDomainByID_CustomDomain_DNSVerification{
						ID:                    domainVerificationID,
						CloudflareHostnameID:  cfHostnameID,
						AcmeChallengePath:     &acmeChallengePath,
						AcmeChallengeStatus:   enums.SSLVerificationStatusActive,
						DNSVerificationStatus: enums.DNSVerificationStatusActive,
					},
					MappableDomain: openlaneclient.GetCustomDomainByID_CustomDomain_MappableDomain{
						ID:     mappableDomainID,
						ZoneID: zoneID,
					},
				},
			},
			getCustomHostnameResponse: &custom_hostnames.CustomHostnameGetResponse{
				ID:     cfHostnameID,
				Status: "active",
				SSL: custom_hostnames.CustomHostnameGetResponseSSL{
					Status: "active",
				},
			},
		},
		{
			name: "happy path - with verification errors",
			input: corejobs.ValidateCustomDomainArgs{
				CustomDomainID: customDomainID,
			},
			getCustomDomainResponse: &openlaneclient.GetCustomDomainByID{
				CustomDomain: openlaneclient.GetCustomDomainByID_CustomDomain{
					CnameRecord:       cnameRecord,
					DNSVerificationID: lo.ToPtr(domainVerificationID),
					ID:                customDomainID,
					MappableDomainID:  mappableDomainID,
					OwnerID:           &ownerID,
					DNSVerification: &openlaneclient.GetCustomDomainByID_CustomDomain_DNSVerification{
						ID:                    domainVerificationID,
						CloudflareHostnameID:  cfHostnameID,
						AcmeChallengeStatus:   enums.SSLVerificationStatusInitializing,
						DNSVerificationStatus: enums.DNSVerificationStatusPending,
					},
					MappableDomain: openlaneclient.GetCustomDomainByID_CustomDomain_MappableDomain{
						ID:     mappableDomainID,
						ZoneID: zoneID,
					},
				},
			},
			getCustomHostnameResponse: &custom_hostnames.CustomHostnameGetResponse{
				ID:                 cfHostnameID,
				Status:             "pending",
				VerificationErrors: []string{"DNS record not found"},
				SSL: custom_hostnames.CustomHostnameGetResponseSSL{
					Status: custom_hostnames.CustomHostnameGetResponseSSLStatusValidationTimedOut,
					ValidationErrors: []custom_hostnames.CustomHostnameGetResponseSSLValidationError{
						{
							Message: "Certificate not issued yet",
						},
					},
				},
			},
			expectedUpdateDNSVerification: &openlaneclient.UpdateDNSVerificationInput{
				DNSVerificationStatus:       lo.ToPtr(enums.DNSVerificationStatusPending),
				DNSVerificationStatusReason: lo.ToPtr("DNS record not found"),
				AcmeChallengeStatus:         lo.ToPtr(enums.SSLVerificationStatusValidationTimedOut),
				AcmeChallengeStatusReason:   lo.ToPtr(", Certificate not issued yet"),
			},
		},
		{
			name: "no DNS verification ID",
			input: corejobs.ValidateCustomDomainArgs{
				CustomDomainID: customDomainID,
			},
			getCustomDomainResponse: &openlaneclient.GetCustomDomainByID{
				CustomDomain: openlaneclient.GetCustomDomainByID_CustomDomain{
					CnameRecord:       cnameRecord,
					DNSVerificationID: nil,
					ID:                customDomainID,
					MappableDomainID:  mappableDomainID,
					OwnerID:           &ownerID,
					MappableDomain: openlaneclient.GetCustomDomainByID_CustomDomain_MappableDomain{
						ID:     mappableDomainID,
						ZoneID: zoneID,
					},
				},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			cfMock := cfmocks.NewMockClient(t)
			cfHostnamesMock := cfmocks.NewMockCustomHostnamesService(t)
			olMock := olmocks.NewMockOpenlaneGraphClient(t)

			// Setup mocks
			if tc.getCustomDomainResponse != nil {
				olMock.EXPECT().GetCustomDomainByID(mock.Anything, tc.input.CustomDomainID).Return(tc.getCustomDomainResponse, nil)
			}

			if tc.getCustomDomainResponse != nil && tc.getCustomDomainResponse.CustomDomain.DNSVerificationID != nil {
				cfMock.EXPECT().CustomHostnames().Return(cfHostnamesMock)
				cfHostnamesMock.EXPECT().Get(mock.Anything, tc.getCustomDomainResponse.CustomDomain.DNSVerification.CloudflareHostnameID, mock.Anything).Return(tc.getCustomHostnameResponse, nil)

				if tc.expectedUpdateDNSVerification != nil {
					olMock.EXPECT().UpdateDNSVerification(mock.Anything, *tc.getCustomDomainResponse.CustomDomain.DNSVerificationID, mock.MatchedBy(func(input openlaneclient.UpdateDNSVerificationInput) bool {
						// Match only the fields we care about
						if input.AcmeChallengePath != nil && tc.expectedUpdateDNSVerification.AcmeChallengePath != nil {
							require.Equal(t, *tc.expectedUpdateDNSVerification.AcmeChallengePath, *input.AcmeChallengePath)
						}
						if input.ExpectedAcmeChallengeValue != nil && tc.expectedUpdateDNSVerification.ExpectedAcmeChallengeValue != nil {
							require.Equal(t, *tc.expectedUpdateDNSVerification.ExpectedAcmeChallengeValue, *input.ExpectedAcmeChallengeValue)
						}
						if input.AcmeChallengeStatus != nil && tc.expectedUpdateDNSVerification.AcmeChallengeStatus != nil {
							require.Equal(t, *tc.expectedUpdateDNSVerification.AcmeChallengeStatus, *input.AcmeChallengeStatus)
						}
						if input.DNSVerificationStatus != nil && tc.expectedUpdateDNSVerification.DNSVerificationStatus != nil {
							require.Equal(t, *tc.expectedUpdateDNSVerification.DNSVerificationStatus, *input.DNSVerificationStatus)
						}
						if input.DNSVerificationStatusReason != nil && tc.expectedUpdateDNSVerification.DNSVerificationStatusReason != nil {
							require.Equal(t, *tc.expectedUpdateDNSVerification.DNSVerificationStatusReason, *input.DNSVerificationStatusReason)
						}
						if input.AcmeChallengeStatusReason != nil && tc.expectedUpdateDNSVerification.AcmeChallengeStatusReason != nil {
							require.Equal(t, *tc.expectedUpdateDNSVerification.AcmeChallengeStatusReason, *input.AcmeChallengeStatusReason)
						}
						return true
					})).Return(&openlaneclient.UpdateDNSVerification{}, nil)
				}
			}

			worker := &corejobs.ValidateCustomDomainWorker{
				Config: corejobs.CustomDomainConfig{
					CloudflareAPIKey: "test",
				},
			}

			worker.WithCloudflareClient(cfMock)
			worker.WithOpenlaneClient(olMock)

			ctx := context.Background()
			err := worker.Work(ctx, &river.Job[corejobs.ValidateCustomDomainArgs]{Args: tc.input})

			if tc.expectedError != "" {
				require.Error(t, err)
				require.Contains(t, err.Error(), tc.expectedError)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestValidateCustomDomainWorkerAllDomains(t *testing.T) {
	cfMock := cfmocks.NewMockClient(t)
	cfHostnamesMock := cfmocks.NewMockCustomHostnamesService(t)
	olMock := olmocks.NewMockOpenlaneGraphClient(t)
	mappableDomainID := "mappableDomainID123"
	zoneID1 := "zoneID1"

	olMock.EXPECT().GetAllCustomDomains(mock.Anything).Return(&openlaneclient.GetAllCustomDomains{
		CustomDomains: openlaneclient.GetAllCustomDomains_CustomDomains{
			Edges: []*openlaneclient.GetAllCustomDomains_CustomDomains_Edges{
				// Domain 1: Already active, no updates needed
				{
					Node: &openlaneclient.GetAllCustomDomains_CustomDomains_Edges_Node{
						CnameRecord: "trust.meow.io",
						DNSVerification: &openlaneclient.GetAllCustomDomains_CustomDomains_Edges_Node_DNSVerification{
							ID:                    "dnsVerificationID1",
							CloudflareHostnameID:  "cloudflareHostnameID1",
							AcmeChallengePath:     lo.ToPtr("acmeChallengePath1"),
							AcmeChallengeStatus:   enums.SSLVerificationStatusActive,
							DNSVerificationStatus: enums.DNSVerificationStatusActive,
						},
						DNSVerificationID: lo.ToPtr("dnsVerificationID1"),
						ID:                "1",
						OwnerID:           lo.ToPtr("ownerID1"),
						MappableDomain: openlaneclient.GetAllCustomDomains_CustomDomains_Edges_Node_MappableDomain{
							ID:     mappableDomainID,
							ZoneID: zoneID1,
						},
					},
				},
				// Domain 2: Needs DNS verification update (pending -> active)
				{
					Node: &openlaneclient.GetAllCustomDomains_CustomDomains_Edges_Node{
						CnameRecord: "app.example.com",
						DNSVerification: &openlaneclient.GetAllCustomDomains_CustomDomains_Edges_Node_DNSVerification{
							ID:                    "dnsVerificationID2",
							CloudflareHostnameID:  "cloudflareHostnameID2",
							AcmeChallengePath:     nil,
							AcmeChallengeStatus:   enums.SSLVerificationStatusInitializing,
							DNSVerificationStatus: enums.DNSVerificationStatusPending,
						},
						DNSVerificationID: lo.ToPtr("dnsVerificationID2"),
						ID:                "2",
						OwnerID:           lo.ToPtr("ownerID2"),
						MappableDomain: openlaneclient.GetAllCustomDomains_CustomDomains_Edges_Node_MappableDomain{
							ID:     mappableDomainID,
							ZoneID: zoneID1,
						},
					},
				},
				// Domain 3: No DNS verification ID - should be skipped
				{
					Node: &openlaneclient.GetAllCustomDomains_CustomDomains_Edges_Node{
						CnameRecord:       "portal.test.com",
						DNSVerification:   nil,
						DNSVerificationID: nil,
						ID:                "3",
						OwnerID:           lo.ToPtr("ownerID3"),
						MappableDomain: openlaneclient.GetAllCustomDomains_CustomDomains_Edges_Node_MappableDomain{
							ID:     mappableDomainID,
							ZoneID: zoneID1,
						},
					},
				},
				// Domain 4: Needs ACME challenge update (initializing -> active with new challenge data)
				{
					Node: &openlaneclient.GetAllCustomDomains_CustomDomains_Edges_Node{
						CnameRecord: "secure.domain.org",
						DNSVerification: &openlaneclient.GetAllCustomDomains_CustomDomains_Edges_Node_DNSVerification{
							ID:                         "dnsVerificationID4",
							CloudflareHostnameID:       "cloudflareHostnameID4",
							AcmeChallengePath:          nil,
							AcmeChallengeStatus:        enums.SSLVerificationStatusInitializing,
							DNSVerificationStatus:      enums.DNSVerificationStatusPending,
							ExpectedAcmeChallengeValue: nil,
						},
						DNSVerificationID: lo.ToPtr("dnsVerificationID4"),
						ID:                "4",
						OwnerID:           lo.ToPtr("ownerID4"),
						MappableDomain: openlaneclient.GetAllCustomDomains_CustomDomains_Edges_Node_MappableDomain{
							ID:     mappableDomainID,
							ZoneID: zoneID1,
						},
					},
				},
			},
			PageInfo:   openlaneclient.GetAllCustomDomains_CustomDomains_PageInfo{},
			TotalCount: 4,
		},
	}, nil)

	// Domain 1: Already active, no updates needed
	cfMock.EXPECT().CustomHostnames().Return(cfHostnamesMock)
	cfHostnamesMock.EXPECT().Get(mock.Anything, "cloudflareHostnameID1", mock.Anything).Return(
		&custom_hostnames.CustomHostnameGetResponse{
			ID:     "cloudflareHostnameID1",
			Status: "active",
			SSL: custom_hostnames.CustomHostnameGetResponseSSL{
				Status: "active",
			},
		},
		nil,
	)

	// Domain 2: Needs DNS verification update (pending -> active)
	cfMock.EXPECT().CustomHostnames().Return(cfHostnamesMock)
	cfHostnamesMock.EXPECT().Get(mock.Anything, "cloudflareHostnameID2", mock.Anything).Return(
		&custom_hostnames.CustomHostnameGetResponse{
			ID:     "cloudflareHostnameID2",
			Status: "active",
			SSL: custom_hostnames.CustomHostnameGetResponseSSL{
				Status: "active",
			},
		},
		nil,
	)
	// Expect update for domain 2
	olMock.EXPECT().UpdateDNSVerification(mock.Anything, "dnsVerificationID2", mock.MatchedBy(func(input openlaneclient.UpdateDNSVerificationInput) bool {
		return input.DNSVerificationStatus != nil && *input.DNSVerificationStatus == enums.DNSVerificationStatusActive &&
			input.AcmeChallengeStatus != nil && *input.AcmeChallengeStatus == enums.SSLVerificationStatusActive
	})).Return(&openlaneclient.UpdateDNSVerification{}, nil)

	// Domain 4: Needs ACME challenge update (initializing -> active with new challenge data)
	cfMock.EXPECT().CustomHostnames().Return(cfHostnamesMock)
	cfHostnamesMock.EXPECT().Get(mock.Anything, "cloudflareHostnameID4", mock.Anything).Return(
		&custom_hostnames.CustomHostnameGetResponse{
			ID:     "cloudflareHostnameID4",
			Status: "active",
			SSL: custom_hostnames.CustomHostnameGetResponseSSL{
				Status: "active",
				ValidationRecords: []custom_hostnames.CustomHostnameGetResponseSSLValidationRecord{
					{
						HTTPURL:  "http://secure.domain.org/.well-known/acme-challenge/newChallengePath4",
						HTTPBody: "newChallengeValue4",
					},
				},
			},
		},
		nil,
	)
	// Expect update for domain 4 with new ACME challenge data
	olMock.EXPECT().UpdateDNSVerification(mock.Anything, "dnsVerificationID4", mock.MatchedBy(func(input openlaneclient.UpdateDNSVerificationInput) bool {
		return input.AcmeChallengePath != nil && *input.AcmeChallengePath == "newChallengePath4" &&
			input.ExpectedAcmeChallengeValue != nil && *input.ExpectedAcmeChallengeValue == "newChallengeValue4" &&
			input.AcmeChallengeStatus != nil && *input.AcmeChallengeStatus == enums.SSLVerificationStatusActive &&
			input.DNSVerificationStatus != nil && *input.DNSVerificationStatus == enums.DNSVerificationStatusActive
	})).Return(&openlaneclient.UpdateDNSVerification{}, nil)

	worker := &corejobs.ValidateCustomDomainWorker{
		Config: corejobs.CustomDomainConfig{
			CloudflareAPIKey: "test",
		},
	}

	worker.WithCloudflareClient(cfMock)
	worker.WithOpenlaneClient(olMock)

	ctx := context.Background()
	err := worker.Work(ctx, &river.Job[corejobs.ValidateCustomDomainArgs]{Args: corejobs.ValidateCustomDomainArgs{
		CustomDomainID: "",
	}})

	require.NoError(t, err)

}

func TestValidatePreviewDomainWorker_UpdateTxtRecord(t *testing.T) {
	t.Parallel()

	trustCenterID := "trustcenterid123"
	zoneID := "cfzoneid123"
	dnsTxtRecord := "_acme-challenge.preview.example.com"
	dnsTxtValue := "verification-value-123"
	oldTxtValue := "old-verification-value"
	recordID := "recordid123"

	testCases := []struct {
		name                      string
		trustCenterID             string
		trustCenterPreviewZoneID  string
		getTrustCenterResponse    *openlaneclient.GetTrustCenterByID
		listRecordsResponse       *pagination.V4PagePaginationArray[dns.RecordResponse]
		editRecordResponse        *dns.RecordResponse
		updateTrustCenterResponse *openlaneclient.UpdateTrustCenter
		expectEdit                bool
		expectSnooze              bool
	}{
		{
			name:                     "txt record exists with different content - update it",
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
						Content: oldTxtValue, // Different content
						Type:    "TXT",
					},
				},
			},
			editRecordResponse: &dns.RecordResponse{
				ID:      recordID,
				Name:    dnsTxtRecord,
				Content: dnsTxtValue,
				Type:    "TXT",
			},
			expectEdit:   true,
			expectSnooze: true,
		},
		{
			name:                     "txt record exists with different content - verification active - update record and trust center",
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
						Content: oldTxtValue, // Different content
						Type:    "TXT",
					},
				},
			},
			editRecordResponse: &dns.RecordResponse{
				ID:      recordID,
				Name:    dnsTxtRecord,
				Content: dnsTxtValue,
				Type:    "TXT",
			},
			updateTrustCenterResponse: &openlaneclient.UpdateTrustCenter{},
			expectEdit:                true,
			expectSnooze:              false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctx := context.Background()
			cfMock := cfmocks.NewMockClient(t)
			olMock := olmocks.NewMockOpenlaneGraphClient(t)

			// Setup mocks
			olMock.EXPECT().GetTrustCenterByID(mock.Anything, tc.trustCenterID).Return(tc.getTrustCenterResponse, nil)

			cfRecordsMock := cfmocks.NewMockRecordService(t)
			cfMock.EXPECT().Record().Return(cfRecordsMock)

			cfRecordsMock.EXPECT().List(mock.Anything, mock.MatchedBy(func(params dns.RecordListParams) bool {
				return params.ZoneID.Value == tc.trustCenterPreviewZoneID
			})).Return(tc.listRecordsResponse, nil)

			if tc.expectEdit {
				cfRecordsMock.EXPECT().Edit(mock.Anything, recordID, mock.MatchedBy(func(params dns.RecordEditParams) bool {
					return params.ZoneID.Value == tc.trustCenterPreviewZoneID
				})).Return(tc.editRecordResponse, nil)
			}

			// If verification is active, expect update trust center call
			if tc.updateTrustCenterResponse != nil {
				olMock.EXPECT().UpdateTrustCenter(mock.Anything, tc.trustCenterID, mock.MatchedBy(func(input openlaneclient.UpdateTrustCenterInput) bool {
					return input.PreviewStatus != nil && *input.PreviewStatus == enums.TrustCenterPreviewStatusReady
				})).Return(tc.updateTrustCenterResponse, nil)
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

			job := &river.Job[corejobs.ValidatePreviewDomainArgs]{
				JobRow: &rivertype.JobRow{
					Metadata: []byte(`{}`),
				},
				Args: corejobs.ValidatePreviewDomainArgs{
					TrustCenterID:            tc.trustCenterID,
					TrustCenterPreviewZoneID: tc.trustCenterPreviewZoneID,
				},
			}

			err := worker.Work(ctx, job)

			if tc.expectSnooze {
				require.Error(t, err)
				require.Contains(t, err.Error(), "JobSnoozeError")
			} else {
				require.NoError(t, err)
			}
		})
	}
}
