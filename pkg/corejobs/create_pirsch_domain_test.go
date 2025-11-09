package corejobs_test

import (
	"context"
	"testing"

	"github.com/riverqueue/river"
	"github.com/riverqueue/river/rivertype"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	rivermocks "github.com/theopenlane/riverboat/pkg/riverqueue/mocks"

	"github.com/theopenlane/core/pkg/corejobs"
	"github.com/theopenlane/core/pkg/corejobs/internal/pirsch"
	"github.com/theopenlane/core/pkg/openlaneclient"

	olmocks "github.com/theopenlane/core/pkg/corejobs/internal/olclient/mocks"
	pirschmocks "github.com/theopenlane/core/pkg/corejobs/internal/pirsch/mocks"
)

func TestCreatePirschDomainWorker(t *testing.T) {
	t.Parallel()

	trustCenterID := "trustcenter123"
	ownerID := "owner123"
	customDomainHostname := "trust.example.com"
	organizationDisplayName := "Acme Corp"
	pirschDomainID := "pirsch123"
	pirschIdentificationCode := "abc123xyz"

	testCases := []struct {
		name                           string
		trustCenterID                  string
		expectedGetTrustCenterByID     bool
		trustCenterResponse            *openlaneclient.GetTrustCenterByID
		trustCenterError               error
		expectedGetOrganizationByID    bool
		organizationResponse           *openlaneclient.GetOrganizationByID
		organizationError              error
		expectedCreateDomainRequest    *pirsch.CreateDomainRequest
		createDomainResponse           *pirsch.Domain
		createDomainError              error
		expectedUpdateTrustCenterInput *openlaneclient.UpdateTrustCenterInput
		updateTrustCenterError         error
		expectedInsertDeleteJob        *corejobs.DeletePirschDomainArgs
		insertDeleteJobError           error
		expectedError                  string
	}{
		{
			name:                       "happy path",
			trustCenterID:              trustCenterID,
			expectedGetTrustCenterByID: true,
			trustCenterResponse: &openlaneclient.GetTrustCenterByID{
				TrustCenter: openlaneclient.GetTrustCenterByID_TrustCenter{
					ID:      trustCenterID,
					OwnerID: &ownerID,
					CustomDomain: &openlaneclient.GetTrustCenterByID_TrustCenter_CustomDomain{
						CnameRecord: customDomainHostname,
					},
				},
			},
			expectedGetOrganizationByID: true,
			organizationResponse: &openlaneclient.GetOrganizationByID{
				Organization: openlaneclient.GetOrganizationByID_Organization{
					ID:          ownerID,
					DisplayName: organizationDisplayName,
				},
			},
			expectedCreateDomainRequest: &pirsch.CreateDomainRequest{
				Hostname:                    "example.com",
				Subdomain:                   "trust",
				Timezone:                    "UTC",
				DisplayName:                 organizationDisplayName,
				Public:                      false,
				GroupByTitle:                false,
				ActiveVisitorsSeconds:       300,
				DisableScripts:              false,
				TrafficSpikeThreshold:       0,
				TrafficWarningThresholdDays: 0,
			},
			createDomainResponse: &pirsch.Domain{
				ID:                 pirschDomainID,
				Hostname:           "example.com",
				Subdomain:          "trust",
				IdentificationCode: pirschIdentificationCode,
			},
			expectedUpdateTrustCenterInput: &openlaneclient.UpdateTrustCenterInput{
				PirschDomainID:           &pirschDomainID,
				PirschIdentificationCode: &pirschIdentificationCode,
			},
		},
		{
			name:          "missing trust center id",
			trustCenterID: "",
			expectedError: "trust_center_id is required for the create_pirsch_domain job",
		},
		{
			name:                       "trust center not found",
			trustCenterID:              trustCenterID,
			expectedGetTrustCenterByID: true,
			trustCenterError:           ErrTest,
			expectedError:              "test error",
		},
		{
			name:                       "trust center without custom domain",
			trustCenterID:              trustCenterID,
			expectedGetTrustCenterByID: true,
			trustCenterResponse: &openlaneclient.GetTrustCenterByID{
				TrustCenter: openlaneclient.GetTrustCenterByID_TrustCenter{
					ID:           trustCenterID,
					OwnerID:      &ownerID,
					CustomDomain: nil,
				},
			},
			expectedError: "trust center does not have a custom domain",
		},
		{
			name:                       "invalid hostname format - no dot",
			trustCenterID:              trustCenterID,
			expectedGetTrustCenterByID: true,
			trustCenterResponse: &openlaneclient.GetTrustCenterByID{
				TrustCenter: openlaneclient.GetTrustCenterByID_TrustCenter{
					ID:      trustCenterID,
					OwnerID: &ownerID,
					CustomDomain: &openlaneclient.GetTrustCenterByID_TrustCenter_CustomDomain{
						CnameRecord: "invalidhostname",
					},
				},
			},
			expectedError: "invalid custom domain hostname",
		},
		{
			name:                       "organization not found",
			trustCenterID:              trustCenterID,
			expectedGetTrustCenterByID: true,
			trustCenterResponse: &openlaneclient.GetTrustCenterByID{
				TrustCenter: openlaneclient.GetTrustCenterByID_TrustCenter{
					ID:      trustCenterID,
					OwnerID: &ownerID,
					CustomDomain: &openlaneclient.GetTrustCenterByID_TrustCenter_CustomDomain{
						CnameRecord: customDomainHostname,
					},
				},
			},
			expectedGetOrganizationByID: true,
			organizationError:           ErrTest,
			expectedError:               "test error",
		},
		{
			name:                       "pirsch domain creation fails",
			trustCenterID:              trustCenterID,
			expectedGetTrustCenterByID: true,
			trustCenterResponse: &openlaneclient.GetTrustCenterByID{
				TrustCenter: openlaneclient.GetTrustCenterByID_TrustCenter{
					ID:      trustCenterID,
					OwnerID: &ownerID,
					CustomDomain: &openlaneclient.GetTrustCenterByID_TrustCenter_CustomDomain{
						CnameRecord: customDomainHostname,
					},
				},
			},
			expectedGetOrganizationByID: true,
			organizationResponse: &openlaneclient.GetOrganizationByID{
				Organization: openlaneclient.GetOrganizationByID_Organization{
					ID:          ownerID,
					DisplayName: organizationDisplayName,
				},
			},
			expectedCreateDomainRequest: &pirsch.CreateDomainRequest{
				Hostname:                    "example.com",
				Subdomain:                   "trust",
				Timezone:                    "UTC",
				DisplayName:                 organizationDisplayName,
				Public:                      false,
				GroupByTitle:                false,
				ActiveVisitorsSeconds:       300,
				DisableScripts:              false,
				TrafficSpikeThreshold:       0,
				TrafficWarningThresholdDays: 0,
			},
			createDomainError: ErrTest,
			expectedError:     "test error",
		},
		{
			name:                       "trust center update fails - cleanup succeeds",
			trustCenterID:              trustCenterID,
			expectedGetTrustCenterByID: true,
			trustCenterResponse: &openlaneclient.GetTrustCenterByID{
				TrustCenter: openlaneclient.GetTrustCenterByID_TrustCenter{
					ID:      trustCenterID,
					OwnerID: &ownerID,
					CustomDomain: &openlaneclient.GetTrustCenterByID_TrustCenter_CustomDomain{
						CnameRecord: customDomainHostname,
					},
				},
			},
			expectedGetOrganizationByID: true,
			organizationResponse: &openlaneclient.GetOrganizationByID{
				Organization: openlaneclient.GetOrganizationByID_Organization{
					ID:          ownerID,
					DisplayName: organizationDisplayName,
				},
			},
			expectedCreateDomainRequest: &pirsch.CreateDomainRequest{
				Hostname:                    "example.com",
				Subdomain:                   "trust",
				Timezone:                    "UTC",
				DisplayName:                 organizationDisplayName,
				Public:                      false,
				GroupByTitle:                false,
				ActiveVisitorsSeconds:       300,
				DisableScripts:              false,
				TrafficSpikeThreshold:       0,
				TrafficWarningThresholdDays: 0,
			},
			createDomainResponse: &pirsch.Domain{
				ID:                 pirschDomainID,
				Hostname:           "example.com",
				Subdomain:          "trust",
				IdentificationCode: pirschIdentificationCode,
			},
			expectedUpdateTrustCenterInput: &openlaneclient.UpdateTrustCenterInput{
				PirschDomainID:           &pirschDomainID,
				PirschIdentificationCode: &pirschIdentificationCode,
			},
			updateTrustCenterError: ErrTest,
			expectedInsertDeleteJob: &corejobs.DeletePirschDomainArgs{
				PirschDomainID: pirschDomainID,
			},
			expectedError: "test error",
		},
		{
			name:                       "trust center update fails - cleanup also fails",
			trustCenterID:              trustCenterID,
			expectedGetTrustCenterByID: true,
			trustCenterResponse: &openlaneclient.GetTrustCenterByID{
				TrustCenter: openlaneclient.GetTrustCenterByID_TrustCenter{
					ID:      trustCenterID,
					OwnerID: &ownerID,
					CustomDomain: &openlaneclient.GetTrustCenterByID_TrustCenter_CustomDomain{
						CnameRecord: customDomainHostname,
					},
				},
			},
			expectedGetOrganizationByID: true,
			organizationResponse: &openlaneclient.GetOrganizationByID{
				Organization: openlaneclient.GetOrganizationByID_Organization{
					ID:          ownerID,
					DisplayName: organizationDisplayName,
				},
			},
			expectedCreateDomainRequest: &pirsch.CreateDomainRequest{
				Hostname:                    "example.com",
				Subdomain:                   "trust",
				Timezone:                    "UTC",
				DisplayName:                 organizationDisplayName,
				Public:                      false,
				GroupByTitle:                false,
				ActiveVisitorsSeconds:       300,
				DisableScripts:              false,
				TrafficSpikeThreshold:       0,
				TrafficWarningThresholdDays: 0,
			},
			createDomainResponse: &pirsch.Domain{
				ID:                 pirschDomainID,
				Hostname:           "example.com",
				Subdomain:          "trust",
				IdentificationCode: pirschIdentificationCode,
			},
			expectedUpdateTrustCenterInput: &openlaneclient.UpdateTrustCenterInput{
				PirschDomainID:           &pirschDomainID,
				PirschIdentificationCode: &pirschIdentificationCode,
			},
			updateTrustCenterError: ErrTest,
			expectedInsertDeleteJob: &corejobs.DeletePirschDomainArgs{
				PirschDomainID: pirschDomainID,
			},
			insertDeleteJobError: ErrTest,
			expectedError:        "test error",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctx := context.Background()
			pirschMock := pirschmocks.NewMockClient(t)
			olMock := olmocks.NewMockOpenlaneGraphClient(t)
			riverMock := rivermocks.NewMockJobClient(t)

			if tc.expectedGetTrustCenterByID {
				olMock.EXPECT().GetTrustCenterByID(mock.Anything, tc.trustCenterID).Return(tc.trustCenterResponse, tc.trustCenterError)
			}

			if tc.expectedGetOrganizationByID {
				olMock.EXPECT().GetOrganizationByID(mock.Anything, ownerID).Return(tc.organizationResponse, tc.organizationError)
			}

			if tc.expectedCreateDomainRequest != nil {
				pirschMock.EXPECT().CreateDomain(mock.Anything, *tc.expectedCreateDomainRequest).Return(tc.createDomainResponse, tc.createDomainError)
			}

			if tc.expectedUpdateTrustCenterInput != nil {
				olMock.EXPECT().UpdateTrustCenter(mock.Anything, trustCenterID, *tc.expectedUpdateTrustCenterInput).Return(&openlaneclient.UpdateTrustCenter{
					UpdateTrustCenter: openlaneclient.UpdateTrustCenter_UpdateTrustCenter{},
				}, tc.updateTrustCenterError)
			}

			if tc.expectedInsertDeleteJob != nil {
				riverMock.EXPECT().Insert(mock.Anything, *tc.expectedInsertDeleteJob, mock.Anything).Return(&rivertype.JobInsertResult{}, tc.insertDeleteJobError)
			}

			worker := &corejobs.CreatePirschDomainWorker{
				Config: corejobs.PirschDomainConfig{
					PirschClientID:     "test_client_id",
					PirschClientSecret: "test_client_secret",
				},
			}

			worker.WithPirschClient(pirschMock)
			worker.WithOpenlaneClient(olMock)
			worker.WithRiverClient(riverMock)

			err := worker.Work(ctx, &river.Job[corejobs.CreatePirschDomainArgs]{Args: corejobs.CreatePirschDomainArgs{
				TrustCenterID: tc.trustCenterID,
			}})

			if tc.expectedError != "" {
				require.Error(t, err)
				require.Contains(t, err.Error(), tc.expectedError)
			} else {
				require.NoError(t, err)
			}

			if !tc.expectedGetTrustCenterByID {
				olMock.AssertNotCalled(t, "GetTrustCenterByID")
			}

			if !tc.expectedGetOrganizationByID {
				olMock.AssertNotCalled(t, "GetOrganizationByID")
			}

			if tc.expectedCreateDomainRequest == nil {
				pirschMock.AssertNotCalled(t, "CreateDomain")
			}

			if tc.expectedUpdateTrustCenterInput == nil {
				olMock.AssertNotCalled(t, "UpdateTrustCenter")
			}

			if tc.expectedInsertDeleteJob == nil {
				riverMock.AssertNotCalled(t, "Insert")
			}
		})
	}
}
