package corejobs_test

import (
	"context"
	"errors"
	"testing"

	"github.com/cloudflare/cloudflare-go/v6"
	"github.com/cloudflare/cloudflare-go/v6/custom_hostnames"
	"github.com/riverqueue/river"
	"github.com/riverqueue/river/rivertype"
	"github.com/samber/lo"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	rivermocks "github.com/theopenlane/riverboat/pkg/riverqueue/mocks"

	"github.com/theopenlane/core/pkg/corejobs"
	"github.com/theopenlane/core/pkg/openlaneclient"

	cfmocks "github.com/theopenlane/core/pkg/corejobs/internal/cloudflare/mocks"
	olmocks "github.com/theopenlane/core/pkg/corejobs/internal/olclient/mocks"
)

var (
	ErrTest = errors.New("test error")
)

func TestCreateCustomDomainWorker(t *testing.T) {
	t.Parallel()

	customDomainID := "customdomainid123"
	mappableDomainID := "mappabldomain123"
	ownerID := "owner123"
	cnameRecord := "trust.meow.io"
	zoneID := "cfzoneid123"
	cfHostnameID := "cfhostnameid123"
	domainVerificationID := "domainverificationid123"
	txtRecord := "_cfverify.trust.meow.io"
	txtRecordValue := "cfverifyvalue123"

	testCases := []struct {
		name                          string
		customDomainID                string
		expectedCustomHostnames       *custom_hostnames.CustomHostnameNewParams
		expectedCreateDNSVerification *openlaneclient.CreateDNSVerificationInput
		expectedUpdateCustomDomain    *openlaneclient.UpdateCustomDomainInput
		expectedInsertDeleteJobInput  *corejobs.DeleteCustomDomainArgs
		expectedError                 string
		callGetMappableDomain         bool
		callCustomHostnames           bool
		getCustomDomainResponse       *openlaneclient.GetCustomDomainByID
		verificationCreateError       error
		updateCustomDomainError       error
	}{
		{
			name:                  "happy path",
			customDomainID:        customDomainID,
			callGetMappableDomain: true,
			callCustomHostnames:   true,
			expectedCustomHostnames: &custom_hostnames.CustomHostnameNewParams{
				ZoneID:   cloudflare.F(zoneID),
				Hostname: cloudflare.F(cnameRecord),
				SSL: cloudflare.F(custom_hostnames.CustomHostnameNewParamsSSL{
					Method: cloudflare.F(custom_hostnames.DCVMethodHTTP),
					Settings: cloudflare.F(custom_hostnames.CustomHostnameNewParamsSSLSettings{
						MinTLSVersion: cloudflare.F(custom_hostnames.CustomHostnameNewParamsSSLSettingsMinTLSVersion1_0),
					}),
					Type: cloudflare.F(custom_hostnames.DomainValidationTypeDv),
				}),
			},
			expectedCreateDNSVerification: &openlaneclient.CreateDNSVerificationInput{
				CloudflareHostnameID: cfHostnameID,
				DNSTxtRecord:         txtRecord,
				DNSTxtValue:          txtRecordValue,
				OwnerID:              &ownerID,
			},
			expectedUpdateCustomDomain: &openlaneclient.UpdateCustomDomainInput{
				DNSVerificationID: lo.ToPtr(domainVerificationID),
			},
			getCustomDomainResponse: &openlaneclient.GetCustomDomainByID{
				CustomDomain: openlaneclient.GetCustomDomainByID_CustomDomain{
					CnameRecord:       cnameRecord,
					DNSVerificationID: nil,
					ID:                customDomainID,
					MappableDomainID:  mappableDomainID,
					OwnerID:           &ownerID,
				},
			},
		},
		{
			name:          "missing custom domain id",
			expectedError: "custom_domain_id is required for the create_custom_domain job",
		},
		{
			name:                  "dns verification already exists",
			customDomainID:        customDomainID,
			callGetMappableDomain: false,
			callCustomHostnames:   false,
			getCustomDomainResponse: &openlaneclient.GetCustomDomainByID{
				CustomDomain: openlaneclient.GetCustomDomainByID_CustomDomain{
					CnameRecord:       cnameRecord,
					DNSVerificationID: lo.ToPtr(domainVerificationID),
					ID:                customDomainID,
					MappableDomainID:  mappableDomainID,
					OwnerID:           &ownerID,
				},
			},
			expectedError: "custom domain already has a verification member",
		},
		{
			name:                  "deletes on verification create error",
			customDomainID:        customDomainID,
			callGetMappableDomain: true,
			callCustomHostnames:   true,
			expectedCustomHostnames: &custom_hostnames.CustomHostnameNewParams{
				ZoneID:   cloudflare.F(zoneID),
				Hostname: cloudflare.F(cnameRecord),
				SSL: cloudflare.F(custom_hostnames.CustomHostnameNewParamsSSL{
					Method: cloudflare.F(custom_hostnames.DCVMethodHTTP),
					Settings: cloudflare.F(custom_hostnames.CustomHostnameNewParamsSSLSettings{
						MinTLSVersion: cloudflare.F(custom_hostnames.CustomHostnameNewParamsSSLSettingsMinTLSVersion1_0),
					}),
					Type: cloudflare.F(custom_hostnames.DomainValidationTypeDv),
				}),
			},
			expectedCreateDNSVerification: &openlaneclient.CreateDNSVerificationInput{
				CloudflareHostnameID: cfHostnameID,
				DNSTxtRecord:         txtRecord,
				DNSTxtValue:          txtRecordValue,
				OwnerID:              &ownerID,
			},
			getCustomDomainResponse: &openlaneclient.GetCustomDomainByID{
				CustomDomain: openlaneclient.GetCustomDomainByID_CustomDomain{
					CnameRecord:       cnameRecord,
					DNSVerificationID: nil,
					ID:                customDomainID,
					MappableDomainID:  mappableDomainID,
					OwnerID:           &ownerID,
				},
			},
			verificationCreateError: ErrTest,
			expectedInsertDeleteJobInput: &corejobs.DeleteCustomDomainArgs{
				DNSVerificationID:          "",
				CloudflareCustomHostnameID: cfHostnameID,
				CloudflareZoneID:           zoneID,
			},
			expectedError: "test error",
		},
		{
			name:                  "deletes on update custom domain error",
			customDomainID:        customDomainID,
			callGetMappableDomain: true,
			callCustomHostnames:   true,
			expectedCustomHostnames: &custom_hostnames.CustomHostnameNewParams{
				ZoneID:   cloudflare.F(zoneID),
				Hostname: cloudflare.F(cnameRecord),
				SSL: cloudflare.F(custom_hostnames.CustomHostnameNewParamsSSL{
					Method: cloudflare.F(custom_hostnames.DCVMethodHTTP),
					Settings: cloudflare.F(custom_hostnames.CustomHostnameNewParamsSSLSettings{
						MinTLSVersion: cloudflare.F(custom_hostnames.CustomHostnameNewParamsSSLSettingsMinTLSVersion1_0),
					}),
					Type: cloudflare.F(custom_hostnames.DomainValidationTypeDv),
				}),
			},
			expectedCreateDNSVerification: &openlaneclient.CreateDNSVerificationInput{
				CloudflareHostnameID: cfHostnameID,
				DNSTxtRecord:         txtRecord,
				DNSTxtValue:          txtRecordValue,
				OwnerID:              &ownerID,
			},
			getCustomDomainResponse: &openlaneclient.GetCustomDomainByID{
				CustomDomain: openlaneclient.GetCustomDomainByID_CustomDomain{
					CnameRecord:       cnameRecord,
					DNSVerificationID: nil,
					ID:                customDomainID,
					MappableDomainID:  mappableDomainID,
					OwnerID:           &ownerID,
				},
			},
			updateCustomDomainError: ErrTest,
			expectedInsertDeleteJobInput: &corejobs.DeleteCustomDomainArgs{
				DNSVerificationID:          domainVerificationID,
				CloudflareCustomHostnameID: cfHostnameID,
				CloudflareZoneID:           zoneID,
			},
			expectedUpdateCustomDomain: &openlaneclient.UpdateCustomDomainInput{
				DNSVerificationID: lo.ToPtr(domainVerificationID),
			},
			expectedError: "test error",
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctx := context.Background()
			cfMock := cfmocks.NewMockClient(t)
			cfHostnamesMock := cfmocks.NewMockCustomHostnamesService(t)

			if tc.callCustomHostnames {
				cfMock.EXPECT().CustomHostnames().Return(cfHostnamesMock)
			}

			riverMock := rivermocks.NewMockJobClient(t)
			olMock := olmocks.NewMockOpenlaneGraphClient(t)

			if tc.expectedCustomHostnames != nil {
				cfHostnamesMock.EXPECT().New(mock.Anything, *tc.expectedCustomHostnames).Return(&custom_hostnames.CustomHostnameNewResponse{
					ID: cfHostnameID,
					OwnershipVerification: custom_hostnames.CustomHostnameNewResponseOwnershipVerification{
						Name:  txtRecord,
						Value: txtRecordValue,
					},
				}, nil)
			}

			if tc.getCustomDomainResponse != nil {
				olMock.EXPECT().GetCustomDomainByID(mock.Anything, tc.customDomainID).Return(tc.getCustomDomainResponse, nil)
			}

			if tc.callGetMappableDomain {
				olMock.EXPECT().GetMappableDomainByID(mock.Anything, mappableDomainID).Return(&openlaneclient.GetMappableDomainByID{
					MappableDomain: openlaneclient.GetMappableDomainByID_MappableDomain{
						ID:     mappableDomainID,
						ZoneID: zoneID,
					},
				}, nil)
			}

			if tc.expectedCreateDNSVerification != nil {
				olMock.EXPECT().CreateDNSVerification(mock.Anything, *tc.expectedCreateDNSVerification).Return(&openlaneclient.CreateDNSVerification{
					CreateDNSVerification: openlaneclient.CreateDNSVerification_CreateDNSVerification{
						DNSVerification: openlaneclient.CreateDNSVerification_CreateDNSVerification_DNSVerification{
							ID: domainVerificationID,
						},
					},
				}, tc.verificationCreateError)
			}

			if tc.expectedUpdateCustomDomain != nil {
				olMock.EXPECT().UpdateCustomDomain(mock.Anything, tc.customDomainID, *tc.expectedUpdateCustomDomain).Return(&openlaneclient.UpdateCustomDomain{
					UpdateCustomDomain: openlaneclient.UpdateCustomDomain_UpdateCustomDomain{},
				}, tc.updateCustomDomainError)
			}

			if tc.expectedInsertDeleteJobInput != nil {
				riverMock.EXPECT().Insert(mock.Anything, *tc.expectedInsertDeleteJobInput, mock.Anything).Return(&rivertype.JobInsertResult{}, nil)
			}

			worker := &corejobs.CreateCustomDomainWorker{
				Config: corejobs.CustomDomainConfig{
					CloudflareAPIKey: "test",
				},
			}

			worker.WithCloudflareClient(cfMock)
			worker.WithOpenlaneClient(olMock)
			worker.WithRiverClient(riverMock)

			err := worker.Work(ctx, &river.Job[corejobs.CreateCustomDomainArgs]{Args: corejobs.CreateCustomDomainArgs{
				CustomDomainID: tc.customDomainID,
			}})

			if tc.expectedError != "" {
				require.Error(t, err)
				require.Contains(t, err.Error(), tc.expectedError)
			} else {
				require.NoError(t, err)
			}

			if tc.expectedInsertDeleteJobInput == nil {
				riverMock.AssertNotCalled(t, "Insert")
			}

			if tc.expectedCreateDNSVerification == nil {
				olMock.AssertNotCalled(t, "CreateDNSVerification")
			}

			if tc.expectedUpdateCustomDomain == nil {
				olMock.AssertNotCalled(t, "UpdateCustomDomain")
			}

			if tc.expectedCustomHostnames == nil {
				cfHostnamesMock.AssertNotCalled(t, "New")
			}
		})
	}
}
