//go:build test

package graphapi_test

import (
	"context"
	"strings"
	"testing"

	"github.com/brianvoe/gofakeit/v7"
	"github.com/samber/lo"
	"gotest.tools/v3/assert"

	"github.com/theopenlane/core/common/enums"
	"github.com/theopenlane/core/common/models"

	"github.com/theopenlane/core/v2/internal/ent/hooks"
	"github.com/theopenlane/core/v2/internal/graphapi"
	"github.com/theopenlane/core/v2/internal/graphapi/testclient"
	th "github.com/theopenlane/core/v2/internal/graphapi/testharness"
)

func TestNDAAutoApprovalRules(t *testing.T) {
	tcOrg := th.CreateFreshOrgWithTrustCenter(t, th.WithNDATemplate())
	freshOrg := th.CreateFreshOrgWithTrustCenter(t, th.WithNDATemplate())

	defer th.CleanupOrganizationDataWithContext(tcOrg.Owner.UserCtx, t)
	defer th.CleanupOrganizationDataWithContext(freshOrg.Owner.UserCtx, t)

	const (
		approvedDomain    = "theopenlane.io"
		blockedDomain     = "microsoft.com"
		org2ContactDomain = "github.com"
		freeDomain        = "gmail.com"
		disposableDomain  = "mailinator.com"
		allowedDomain     = "stripe.com"
		unmatchedDomain   = "gitlab.com"
		companyDomain     = "cloudflare.com"
		unknownDomain     = "atlassian.com"
	)

	approvedEmail := gofakeit.LetterN(12) + "@" + approvedDomain
	blockedEmail := gofakeit.LetterN(12) + "@" + blockedDomain
	gmailEmail := gofakeit.LetterN(12) + "@" + freeDomain
	workEmail := gofakeit.LetterN(12) + "@" + freeDomain
	inactiveContactEmail := gofakeit.LetterN(12) + "@" + approvedDomain
	disposableContactEmail := gofakeit.LetterN(12) + "@" + disposableDomain
	contactInOrg2 := gofakeit.LetterN(12) + "@" + org2ContactDomain
	matchingContactEmail := gofakeit.LetterN(12) + "@" + approvedDomain
	restrictedContactEmail := gofakeit.LetterN(12) + "@" + approvedDomain

	requests := make([]*testclient.CreateTrustCenterNDARequestInput, 0, 3)
	for _, email := range []string{
		gofakeit.LetterN(12) + "@" + approvedDomain,
		gofakeit.LetterN(12) + "@" + freeDomain,
		gofakeit.LetterN(12) + "@" + disposableDomain,
	} {
		requests = append(requests, &testclient.CreateTrustCenterNDARequestInput{
			TrustCenterID: &tcOrg.TrustCenter.ID,
			FirstName:     gofakeit.FirstName(),
			LastName:      gofakeit.LastName(),
			Email:         email,
			Status:        lo.ToPtr(enums.TrustCenterNDARequestStatusApproved),
		})
	}

	_, err := suite.Client.API.CreateBulkTrustCenterNDARequest(tcOrg.Owner.UserCtx, requests)
	assert.NilError(t, err)

	_, err = suite.Client.API.CreateTrustCenterNDARequest(freshOrg.Owner.UserCtx, testclient.CreateTrustCenterNDARequestInput{
		TrustCenterID: &freshOrg.TrustCenter.ID,
		FirstName:     gofakeit.FirstName(),
		LastName:      gofakeit.LastName(),
		Email:         gofakeit.LetterN(12) + "@" + org2ContactDomain,
		Status:        lo.ToPtr(enums.TrustCenterNDARequestStatusApproved),
	})
	assert.NilError(t, err)

	// seed contacts for the first org
	contacts := make([]*testclient.CreateContactInput, 0, 5)
	for _, email := range []string{
		approvedEmail,
		blockedEmail,
		gmailEmail,
		workEmail,
		disposableContactEmail,
		matchingContactEmail,
		restrictedContactEmail,
	} {
		contacts = append(contacts, &testclient.CreateContactInput{
			Email:   lo.ToPtr(email),
			OwnerID: &tcOrg.OrganizationID,
		})
	}
	_, err = suite.Client.API.CreateBulkContact(tcOrg.Owner.UserCtx, contacts)
	assert.NilError(t, err)

	_, err = suite.Client.API.CreateContact(tcOrg.Owner.UserCtx, testclient.CreateContactInput{
		Email:   &inactiveContactEmail,
		OwnerID: &tcOrg.OrganizationID,
		Status:  lo.ToPtr(enums.UserStatusInactive),
	})
	assert.NilError(t, err)

	// create this in the second org, will be needed for cross org isolation tests to make sure
	// contacts do not apporve across orgs
	_, err = suite.Client.API.CreateContact(freshOrg.Owner.UserCtx, testclient.CreateContactInput{
		Email:   &contactInOrg2,
		OwnerID: &freshOrg.OrganizationID,
	})
	assert.NilError(t, err)

	setting, err := tcOrg.TrustCenter.QuerySetting().Only(tcOrg.Owner.UserCtx)
	assert.NilError(t, err)

	setupListener, err := graphapi.SetupListenerRuntime(suite.GalaRuntime, hooks.NDAAutoApprovalListeners())
	assert.NilError(t, err)

	defer setupListener.Teardown()

	// make anonCtx to verify we can still send nda requests via anon
	anonCtx := th.CreateAnonymousTrustCenterContext(tcOrg.TrustCenter.ID, tcOrg.OrganizationID)

	tests := []struct {
		ctx     context.Context
		name    string
		email   string
		options models.TrustCenterNDARequestSetting
		status  enums.TrustCenterNDARequestStatus
	}{
		{
			ctx:   anonCtx,
			name:  "domain whitelist",
			email: gofakeit.LetterN(12) + "@" + allowedDomain,
			options: models.TrustCenterNDARequestSetting{
				WorkEmailOnly:      true,
				UseDomainAllowlist: true,
				DomainAllowlist: []string{
					allowedDomain,
				},
			},
			status: enums.TrustCenterNDARequestStatusApproved,
		},
		{
			ctx:   tcOrg.Owner.UserCtx,
			name:  "role account is rejected by default",
			email: "support@" + allowedDomain,
			options: models.TrustCenterNDARequestSetting{
				WorkEmailOnly: true,
			},
			status: enums.TrustCenterNDARequestStatusDeclined,
		},
		{
			ctx:   tcOrg.Owner.UserCtx,
			name:  "role account is approved when allowed",
			email: "sales@" + allowedDomain,
			options: models.TrustCenterNDARequestSetting{
				WorkEmailOnly:    true,
				AllowRoleAccount: true,
			},
			status: enums.TrustCenterNDARequestStatusApproved,
		},
		{
			ctx:   tcOrg.Owner.UserCtx,
			name:  "approval falls back to needing approval",
			email: gofakeit.LetterN(12) + "@" + unmatchedDomain,
			options: models.TrustCenterNDARequestSetting{
				WorkEmailOnly:           true,
				UseDomainAllowlist:      true,
				DomainAllowlist:         []string{},
				ManualApprovalOnFailure: true,
			},
			status: enums.TrustCenterNDARequestStatusNeedsApproval,
		},
		{
			ctx:   tcOrg.Owner.UserCtx,
			name:  "existing approved domain from a previous nda request",
			email: gofakeit.LetterN(12) + "@" + approvedDomain,
			options: models.TrustCenterNDARequestSetting{
				WorkEmailOnly:                    true,
				ApproveFromExistingRequestDomain: true,
			},
			status: enums.TrustCenterNDARequestStatusApproved,
		},
		{
			ctx:   tcOrg.Owner.UserCtx,
			name:  "approved contact in another another org gets excluded",
			email: gofakeit.LetterN(12) + "@" + org2ContactDomain,
			options: models.TrustCenterNDARequestSetting{
				WorkEmailOnly:                    true,
				ApproveFromExistingRequestDomain: true,
			},
			status: enums.TrustCenterNDARequestStatusDeclined,
		},
		{
			ctx:   tcOrg.Owner.UserCtx,
			name:  "existing approved free domain is allowed",
			email: gofakeit.LetterN(12) + "@" + freeDomain,
			options: models.TrustCenterNDARequestSetting{
				WorkEmailOnly:                    false,
				ApproveFromExistingRequestDomain: true,
			},
			status: enums.TrustCenterNDARequestStatusApproved,
		},
		{
			ctx:   tcOrg.Owner.UserCtx,
			name:  "contact matches so gets approved",
			email: approvedEmail,
			options: models.TrustCenterNDARequestSetting{
				WorkEmailOnly:          true,
				ApproveIfContactExists: true,
			},
			status: enums.TrustCenterNDARequestStatusApproved,
		},
		{
			ctx:   tcOrg.Owner.UserCtx,
			name:  "contact domain matches so gets approved",
			email: gofakeit.LetterN(12) + "@" + approvedDomain,
			options: models.TrustCenterNDARequestSetting{
				WorkEmailOnly:            true,
				ApproveFromContactDomain: true,
			},
			status: enums.TrustCenterNDARequestStatusApproved,
		},
		{
			ctx:   tcOrg.Owner.UserCtx,
			name:  "another organization contact is excluded",
			email: contactInOrg2,
			options: models.TrustCenterNDARequestSetting{
				WorkEmailOnly:            true,
				ApproveIfContactExists:   true,
				ApproveFromContactDomain: true,
			},
			status: enums.TrustCenterNDARequestStatusDeclined,
		},
		{
			ctx:   tcOrg.Owner.UserCtx,
			name:  "free contact domain match is approved",
			email: gofakeit.LetterN(12) + "@" + freeDomain,
			options: models.TrustCenterNDARequestSetting{
				WorkEmailOnly:            false,
				ApproveFromContactDomain: true,
			},
			status: enums.TrustCenterNDARequestStatusApproved,
		},
		{
			ctx:   tcOrg.Owner.UserCtx,
			name:  "allow only work email",
			email: gofakeit.LetterN(12) + "@" + companyDomain,
			options: models.TrustCenterNDARequestSetting{
				WorkEmailOnly: true,
			},
			status: enums.TrustCenterNDARequestStatusApproved,
		},
		{
			ctx:   tcOrg.Owner.UserCtx,
			name:  "allowlisted free domain approves before work email restriction",
			email: gofakeit.LetterN(12) + "@" + freeDomain,
			options: models.TrustCenterNDARequestSetting{
				WorkEmailOnly:      true,
				UseDomainAllowlist: true,
				DomainAllowlist: []string{
					freeDomain,
				},
			},
			status: enums.TrustCenterNDARequestStatusApproved,
		},
		{
			ctx:   tcOrg.Owner.UserCtx,
			name:  "work email setting rejects disposable provider",
			email: gofakeit.LetterN(12) + "@" + disposableDomain,
			options: models.TrustCenterNDARequestSetting{
				WorkEmailOnly: true,
			},
			status: enums.TrustCenterNDARequestStatusDeclined,
		},
		{
			ctx:   tcOrg.Owner.UserCtx,
			name:  "blocklist overrides exact contact match",
			email: blockedEmail,
			options: models.TrustCenterNDARequestSetting{
				WorkEmailOnly:          true,
				ApproveIfContactExists: true,
				UseDomainBlocklist:     true,
				DomainBlocklist: []string{
					blockedDomain,
				},
			},
			status: enums.TrustCenterNDARequestStatusDeclined,
		},

		{
			ctx:   tcOrg.Owner.UserCtx,
			name:  "allowlists should be case insensitive",
			email: gofakeit.LetterN(12) + "@" + strings.ToUpper(allowedDomain),
			options: models.TrustCenterNDARequestSetting{
				WorkEmailOnly:      true,
				UseDomainAllowlist: true,
				DomainAllowlist: []string{
					allowedDomain,
				},
			},
			status: enums.TrustCenterNDARequestStatusApproved,
		},

		{
			ctx:   tcOrg.Owner.UserCtx,
			name:  "empty domain allowlist does not approve anyone",
			email: gofakeit.LetterN(12) + "@" + companyDomain,
			options: models.TrustCenterNDARequestSetting{
				WorkEmailOnly:      true,
				UseDomainAllowlist: true,
				DomainAllowlist:    []string{},
			},
			status: enums.TrustCenterNDARequestStatusDeclined,
		},
		{
			ctx:   tcOrg.Owner.UserCtx,
			name:  "unmatched existing domain",
			email: gofakeit.LetterN(12) + "@" + unknownDomain,
			options: models.TrustCenterNDARequestSetting{
				WorkEmailOnly:                    true,
				ApproveFromExistingRequestDomain: true,
			},
			status: enums.TrustCenterNDARequestStatusDeclined,
		},
		{
			ctx:   tcOrg.Owner.UserCtx,
			name:  "existing approved disposable domain approves before email restrictions",
			email: gofakeit.LetterN(12) + "@" + disposableDomain,
			options: models.TrustCenterNDARequestSetting{
				WorkEmailOnly:                    false,
				ApproveFromExistingRequestDomain: true,
			},
			status: enums.TrustCenterNDARequestStatusApproved,
		},
		{
			ctx:   tcOrg.Owner.UserCtx,
			name:  "disposable contact domain approves before email restrictions",
			email: gofakeit.LetterN(12) + "@" + disposableDomain,
			options: models.TrustCenterNDARequestSetting{
				WorkEmailOnly:            false,
				ApproveFromContactDomain: true,
			},
			status: enums.TrustCenterNDARequestStatusApproved,
		},
		{
			ctx:   tcOrg.Owner.UserCtx,
			name:  "existing contact match approves before work email restriction",
			email: workEmail,
			options: models.TrustCenterNDARequestSetting{
				WorkEmailOnly:          true,
				ApproveIfContactExists: true,
			},
			status: enums.TrustCenterNDARequestStatusApproved,
		},
		{
			ctx:   tcOrg.Owner.UserCtx,
			name:  "organisation contact must be active to be granted access",
			email: inactiveContactEmail,
			options: models.TrustCenterNDARequestSetting{
				WorkEmailOnly:          true,
				ApproveIfContactExists: true,
			},
			status: enums.TrustCenterNDARequestStatusDeclined,
		},
		{
			ctx:   tcOrg.Owner.UserCtx,
			name:  "domain blocklist overrides the allowlist",
			email: gofakeit.LetterN(12) + "@" + allowedDomain,
			options: models.TrustCenterNDARequestSetting{
				WorkEmailOnly:      true,
				UseDomainAllowlist: true,
				DomainAllowlist: []string{
					allowedDomain,
				},
				UseDomainBlocklist: true,
				DomainBlocklist: []string{
					allowedDomain,
				},
			},
			status: enums.TrustCenterNDARequestStatusDeclined,
		},
		{
			ctx:     tcOrg.Owner.UserCtx,
			name:    "disposable email rejected when work restriction is off",
			email:   gofakeit.LetterN(12) + "@" + disposableDomain,
			options: models.TrustCenterNDARequestSetting{},
			status:  enums.TrustCenterNDARequestStatusDeclined,
		},
		{
			ctx:   tcOrg.Owner.UserCtx,
			name:  "disposable email approved when explicitly allowed",
			email: gofakeit.LetterN(12) + "@" + disposableDomain,
			options: models.TrustCenterNDARequestSetting{
				AllowDisposableEmail: true,
			},
			status: enums.TrustCenterNDARequestStatusApproved,
		},
		{
			ctx:   tcOrg.Owner.UserCtx,
			name:  "work email restriction overrides disposable email permission",
			email: gofakeit.LetterN(12) + "@" + disposableDomain,
			options: models.TrustCenterNDARequestSetting{
				WorkEmailOnly:        true,
				AllowDisposableEmail: true,
			},
			status: enums.TrustCenterNDARequestStatusDeclined,
		},
		{
			ctx:   tcOrg.Owner.UserCtx,
			name:  "disposable exact contact approves before email restrictions",
			email: disposableContactEmail,
			options: models.TrustCenterNDARequestSetting{
				ApproveIfContactExists: true,
			},
			status: enums.TrustCenterNDARequestStatusApproved,
		},
		{
			ctx:   tcOrg.Owner.UserCtx,
			name:  "free email approved when work restriction is off",
			email: gofakeit.LetterN(12) + "@" + freeDomain,
			options: models.TrustCenterNDARequestSetting{
				UseDomainAllowlist: true,
				DomainAllowlist:    []string{freeDomain},
			},
			status: enums.TrustCenterNDARequestStatusApproved,
		},
		{
			ctx:   tcOrg.Owner.UserCtx,
			name:  "disposable email permission does not override blocklist",
			email: gofakeit.LetterN(12) + "@" + disposableDomain,
			options: models.TrustCenterNDARequestSetting{
				AllowDisposableEmail: true,
				UseDomainAllowlist:   true,
				DomainAllowlist:      []string{disposableDomain},
				UseDomainBlocklist:   true,
				DomainBlocklist:      []string{disposableDomain},
			},
			status: enums.TrustCenterNDARequestStatusDeclined,
		},
		{
			ctx:   tcOrg.Owner.UserCtx,
			name:  "disposable rejection falls back to manual approval",
			email: gofakeit.LetterN(12) + "@" + disposableDomain,
			options: models.TrustCenterNDARequestSetting{
				ManualApprovalOnFailure: true,
			},
			status: enums.TrustCenterNDARequestStatusNeedsApproval,
		},
		{
			ctx:   tcOrg.Owner.UserCtx,
			name:  "exact contact approved",
			email: matchingContactEmail,
			options: models.TrustCenterNDARequestSetting{
				ApproveIfContactExists: true,
			},
			status: enums.TrustCenterNDARequestStatusApproved,
		},
		{
			ctx:   tcOrg.Owner.UserCtx,
			name:  "existing approved domain allowed",
			email: gofakeit.LetterN(12) + "@" + approvedDomain,
			options: models.TrustCenterNDARequestSetting{
				ApproveFromExistingRequestDomain: true,
			},
			status: enums.TrustCenterNDARequestStatusApproved,
		},
		{
			ctx:   tcOrg.Owner.UserCtx,
			name:  "unmatched existing domain requires manual approval",
			email: gofakeit.LetterN(12) + "@" + unknownDomain,
			options: models.TrustCenterNDARequestSetting{
				AllowDisposableEmail:             true,
				AllowRoleAccount:                 true,
				ApproveFromExistingRequestDomain: true,
				ManualApprovalOnFailure:          true,
			},
			status: enums.TrustCenterNDARequestStatusNeedsApproval,
		},
		{
			ctx:   tcOrg.Owner.UserCtx,
			name:  "contact match approves before work restriction",
			email: restrictedContactEmail,
			options: models.TrustCenterNDARequestSetting{
				AllowDisposableEmail:    true,
				AllowRoleAccount:        true,
				WorkEmailOnly:           true,
				ApproveIfContactExists:  true,
				ManualApprovalOnFailure: true,
			},
			status: enums.TrustCenterNDARequestStatusApproved,
		},
		{
			ctx:   tcOrg.Owner.UserCtx,
			name:  "contact domain match approves",
			email: gofakeit.LetterN(12) + "@" + approvedDomain,
			options: models.TrustCenterNDARequestSetting{
				ApproveFromContactDomain: true,
				ManualApprovalOnFailure:  true,
			},
			status: enums.TrustCenterNDARequestStatusApproved,
		},
	}

	for _, tt := range tests {

		t.Run(tt.name, func(t *testing.T) {

			// update each test run with our new config
			_, err := suite.Client.API.UpdateTrustCenterSetting(tcOrg.Owner.UserCtx, setting.ID, testclient.UpdateTrustCenterSettingInput{
				NdaApprovalRequired: lo.ToPtr(true),
				EnableAutoApproval:  lo.ToPtr(true),
				AutoApprovalRules:   &tt.options,
			}, nil, nil, nil, nil, nil, nil)
			assert.NilError(t, err)

			input := testclient.CreateTrustCenterNDARequestInput{
				TrustCenterID: &tcOrg.TrustCenter.ID,
				FirstName:     "New",
				LastName:      "Requester",
				Email:         tt.email,
			}

			resp, err := suite.Client.API.CreateTrustCenterNDARequest(tt.ctx, input)
			assert.NilError(t, err)

			req := resp.CreateTrustCenterNDARequest.TrustCenterNDARequest

			assert.Equal(t, *req.Status, enums.TrustCenterNDARequestStatusPendingApproval)
			assert.NilError(t, suite.GalaRuntime.WaitIdle(t.Context()))

			result, err := suite.Client.API.GetTrustCenterNDARequestByID(tcOrg.Owner.UserCtx, req.ID)
			assert.NilError(t, err)

			assert.Equal(t, *result.TrustCenterNDARequest.Status, tt.status)
			assert.Assert(t, result.TrustCenterNDARequest.AutoApproved != nil)
			assert.Equal(t, *result.TrustCenterNDARequest.AutoApproved, tt.status == enums.TrustCenterNDARequestStatusApproved)
			if tt.status == enums.TrustCenterNDARequestStatusApproved {
				assert.Check(t, result.TrustCenterNDARequest.ApprovedAt != nil)
			}
		})
	}
}
