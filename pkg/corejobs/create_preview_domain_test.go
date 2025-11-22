package corejobs_test

import (
	"context"
	"errors"
	"testing"

	"github.com/cloudflare/cloudflare-go/v6/dns"
	"github.com/cloudflare/cloudflare-go/v6/zones"
	"github.com/riverqueue/river"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/theopenlane/core/pkg/corejobs"
	"github.com/theopenlane/core/pkg/enums"
	"github.com/theopenlane/core/pkg/openlaneclient"

	cfmocks "github.com/theopenlane/core/pkg/corejobs/internal/cloudflare/mocks"
	olmocks "github.com/theopenlane/core/pkg/corejobs/internal/olclient/mocks"
)

func TestCreatePreviewDomainWorker(t *testing.T) {
	t.Parallel()

	trustCenterID := "trustcenterid123"
	trustCenterSlug := "TestTrustCenter"
	ownerID := "owner123"
	zoneID := "cfzoneid123"
	zoneName := "preview.example.com"
	cnameTarget := "cname.example.com"
	mappableDomainID := "mappabledomainid123"
	customDomainID := "customdomainid123"
	recordID := "recordid123"

	testCases := []struct {
		name                     string
		trustCenterID            string
		trustCenterPreviewZoneID string
		trustCenterCnameTarget   string
		expectedError            string
		getTrustCenterError      error
		getZoneError             error
		getMappableDomainsError  error
		createCustomDomainError  error
		updateTrustCenterError   error
		createRecordError        error
	}{
		{
			name:                     "happy path",
			trustCenterID:            trustCenterID,
			trustCenterPreviewZoneID: zoneID,
			trustCenterCnameTarget:   cnameTarget,
		},
		{
			name:          "missing trust center id",
			expectedError: "trust_center_id is required for the create_preview_domain job",
		},
		{
			name:          "missing trust center preview zone id",
			trustCenterID: trustCenterID,
			expectedError: "trust_center_preview_zone_id is required for the create_preview_domain job",
		},
		{
			name:                     "missing trust center cname target",
			trustCenterID:            trustCenterID,
			trustCenterPreviewZoneID: zoneID,
			expectedError:            "trust_center_cname_target is required for the create_preview_domain job",
		},
		{
			name:                     "get trust center error",
			trustCenterID:            trustCenterID,
			trustCenterPreviewZoneID: zoneID,
			trustCenterCnameTarget:   cnameTarget,
			getTrustCenterError:      errors.New("trust center not found"),
			expectedError:            "trust center not found",
		},
		{
			name:                     "get zone error",
			trustCenterID:            trustCenterID,
			trustCenterPreviewZoneID: zoneID,
			trustCenterCnameTarget:   cnameTarget,
			getZoneError:             errors.New("zone not found"),
			expectedError:            "zone not found",
		},
		{
			name:                     "get mappable domains error",
			trustCenterID:            trustCenterID,
			trustCenterPreviewZoneID: zoneID,
			trustCenterCnameTarget:   cnameTarget,
			getMappableDomainsError:  errors.New("mappable domain error"),
			expectedError:            "mappable domain error",
		},
		{
			name:                     "no mappable domain found",
			trustCenterID:            trustCenterID,
			trustCenterPreviewZoneID: zoneID,
			trustCenterCnameTarget:   cnameTarget,
			getMappableDomainsError:  nil, // This will be handled specially in the mock setup
			expectedError:            "no mappable domain found",
		},
		{
			name:                     "create custom domain error",
			trustCenterID:            trustCenterID,
			trustCenterPreviewZoneID: zoneID,
			trustCenterCnameTarget:   cnameTarget,
			createCustomDomainError:  errors.New("custom domain creation failed"),
			expectedError:            "custom domain creation failed",
		},
		{
			name:                     "update trust center error",
			trustCenterID:            trustCenterID,
			trustCenterPreviewZoneID: zoneID,
			trustCenterCnameTarget:   cnameTarget,
			updateTrustCenterError:   errors.New("trust center update failed"),
			expectedError:            "trust center update failed",
		},
		{
			name:                     "create record error",
			trustCenterID:            trustCenterID,
			trustCenterPreviewZoneID: zoneID,
			trustCenterCnameTarget:   cnameTarget,
			createRecordError:        errors.New("record creation failed"),
			expectedError:            "record creation failed",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctx := context.Background()
			cfMock := cfmocks.NewMockClient(t)
			olMock := olmocks.NewMockOpenlaneGraphClient(t)

			// Setup mocks for happy path and error cases
			if tc.trustCenterID != "" && tc.trustCenterPreviewZoneID != "" && tc.trustCenterCnameTarget != "" {
				olMock.EXPECT().GetTrustCenterByID(mock.Anything, tc.trustCenterID).Return(&openlaneclient.GetTrustCenterByID{
					TrustCenter: openlaneclient.GetTrustCenterByID_TrustCenter{
						ID:      trustCenterID,
						Slug:    &trustCenterSlug,
						OwnerID: &ownerID,
					},
				}, tc.getTrustCenterError)

				if tc.getTrustCenterError == nil {
					cfZonesMock := cfmocks.NewMockZonesService(t)
					cfMock.EXPECT().Zones().Return(cfZonesMock)
					cfZonesMock.EXPECT().Get(mock.Anything, mock.MatchedBy(func(params zones.ZoneGetParams) bool {
						return params.ZoneID.Value == tc.trustCenterPreviewZoneID
					})).Return(&zones.Zone{
						ID:   zoneID,
						Name: zoneName,
					}, tc.getZoneError)

					if tc.getZoneError == nil {
						// Determine the edges based on test case
						var edges []*openlaneclient.GetMappableDomains_MappableDomains_Edges
						if tc.name != "no mappable domain found" {
							edges = []*openlaneclient.GetMappableDomains_MappableDomains_Edges{
								{
									Node: &openlaneclient.GetMappableDomains_MappableDomains_Edges_Node{
										ID: mappableDomainID,
									},
								},
							}
						}

						olMock.EXPECT().GetMappableDomains(mock.Anything, mock.Anything, mock.Anything, mock.MatchedBy(func(input *openlaneclient.MappableDomainWhereInput) bool {
							return input != nil && input.Name != nil && *input.Name == tc.trustCenterCnameTarget
						})).Return(&openlaneclient.GetMappableDomains{
							MappableDomains: openlaneclient.GetMappableDomains_MappableDomains{
								Edges: edges,
							},
						}, tc.getMappableDomainsError)

						if tc.getMappableDomainsError == nil && len(edges) > 0 {
							olMock.EXPECT().CreateCustomDomain(mock.Anything, mock.MatchedBy(func(input openlaneclient.CreateCustomDomainInput) bool {
								return input.MappableDomainID == mappableDomainID && input.OwnerID == &ownerID
							})).Return(&openlaneclient.CreateCustomDomain{
								CreateCustomDomain: openlaneclient.CreateCustomDomain_CreateCustomDomain{
									CustomDomain: openlaneclient.CreateCustomDomain_CreateCustomDomain_CustomDomain{
										ID: customDomainID,
									},
								},
							}, tc.createCustomDomainError)

							if tc.createCustomDomainError == nil {
								olMock.EXPECT().UpdateTrustCenter(mock.Anything, tc.trustCenterID, mock.MatchedBy(func(input openlaneclient.UpdateTrustCenterInput) bool {
									return input.PreviewDomainID != nil && *input.PreviewDomainID == customDomainID &&
										input.PreviewStatus != nil && *input.PreviewStatus == enums.TrustCenterPreviewStatusProvisioning
								})).Return(&openlaneclient.UpdateTrustCenter{
									UpdateTrustCenter: openlaneclient.UpdateTrustCenter_UpdateTrustCenter{},
								}, tc.updateTrustCenterError)

								if tc.updateTrustCenterError == nil {
									cfRecordsMock := cfmocks.NewMockRecordService(t)
									cfMock.EXPECT().Record().Return(cfRecordsMock)
									cfRecordsMock.EXPECT().New(mock.Anything, mock.MatchedBy(func(params dns.RecordNewParams) bool {
										return params.ZoneID.Value == tc.trustCenterPreviewZoneID
									})).Return(&dns.RecordResponse{
										ID:      recordID,
										Name:    "preview-domain",
										Content: cnameTarget,
									}, tc.createRecordError)
								}
							}
						}
					}
				}
			}

			worker := &corejobs.CreatePreviewDomainWorker{
				Config: corejobs.PreviewDomainConfig{
					CloudflareAPIKey: "test",
				},
			}

			worker.WithCloudflareClient(cfMock)
			worker.WithOpenlaneClient(olMock)

			err := worker.Work(ctx, &river.Job[corejobs.CreatePreviewDomainArgs]{
				Args: corejobs.CreatePreviewDomainArgs{
					TrustCenterID:            tc.trustCenterID,
					TrustCenterPreviewZoneID: tc.trustCenterPreviewZoneID,
					TrustCenterCnameTarget:   tc.trustCenterCnameTarget,
				},
			})

			if tc.expectedError != "" {
				require.Error(t, err)
				require.Contains(t, err.Error(), tc.expectedError)
			} else {
				require.NoError(t, err)
			}
		})
	}
}
