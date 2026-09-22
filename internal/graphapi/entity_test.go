package graphapi_test

import (
	"context"
	"slices"
	"strings"
	"testing"

	th "github.com/theopenlane/core/v2/internal/graphapi/testharness"

	"github.com/samber/lo"
	"gotest.tools/v3/assert"
	is "gotest.tools/v3/assert/cmp"

	"github.com/theopenlane/utils/ulids"

	"github.com/theopenlane/core/common/enums"
	"github.com/theopenlane/core/v2/internal/ent/generated"
	"github.com/theopenlane/core/v2/internal/ent/generated/entitytype"
	"github.com/theopenlane/core/v2/internal/graphapi/testclient"
)

func TestQueryEntity(t *testing.T) {
	entity := (&th.EntityBuilder{Client: suite.Client}).MustNew(th.SharedTestUser1.UserCtx, t)
	anonymousContext := th.CreateAnonymousTrustCenterContext(ulids.New().String(), th.SharedTestUser1.OrganizationID)

	testCases := []struct {
		name     string
		queryID  string
		client   *testclient.TestClient
		ctx      context.Context
		errorMsg string
	}{
		{
			name:    "happy path entity",
			queryID: entity.ID,
			client:  suite.Client.API,
			ctx:     th.SharedTestUser1.UserCtx,
		},
		{
			name:    "happy path entity, using api token",
			queryID: entity.ID,
			client:  suite.Client.APIWithToken,
			ctx:     context.Background(),
		},
		{
			name:    "happy path entity, using personal access token",
			queryID: entity.ID,
			client:  suite.Client.APIWithPAT,
			ctx:     context.Background(),
		},
		{
			name:     "no access",
			queryID:  entity.ID,
			client:   suite.Client.API,
			ctx:      th.SharedTestUser2.UserCtx,
			errorMsg: "entity not found",
		},
		{
			name:     "no access, anonymous user",
			client:   suite.Client.API,
			ctx:      anonymousContext,
			queryID:  entity.ID,
			errorMsg: "entity not found",
		},
	}

	for _, tc := range testCases {
		t.Run("Get "+tc.name, func(t *testing.T) {
			resp, err := tc.client.GetEntityByID(tc.ctx, tc.queryID)

			if tc.errorMsg != "" {
				assert.ErrorContains(t, err, tc.errorMsg)

				return
			}

			assert.NilError(t, err)
			assert.Assert(t, resp != nil)
			assert.Check(t, resp.Entity.ID != "")
		})
	}

	// delete created entity
	(&th.Cleanup[*generated.EntityDeleteOne]{Client: suite.Client.DB.Entity, ID: entity.ID}).MustDelete(th.SharedTestUser1.UserCtx, t)
	// delete the entityType
	(&th.Cleanup[*generated.EntityTypeDeleteOne]{Client: suite.Client.DB.EntityType, ID: entity.EntityTypeID}).MustDelete(th.SharedTestUser1.UserCtx, t)
}

func TestQueryEntities(t *testing.T) {
	entity1 := (&th.EntityBuilder{Client: suite.Client}).MustNew(th.SharedTestUser1.UserCtx, t)
	entity2 := (&th.EntityBuilder{Client: suite.Client}).MustNew(th.SharedTestUser1.UserCtx, t)

	testCases := []struct {
		name            string
		client          *testclient.TestClient
		ctx             context.Context
		expectedResults int
	}{
		{
			name:            "happy path",
			client:          suite.Client.API,
			ctx:             th.SharedTestUser1.UserCtx,
			expectedResults: 2,
		},
		{
			name:            "happy path, using api token",
			client:          suite.Client.APIWithToken,
			ctx:             context.Background(),
			expectedResults: 2,
		},
		{
			name:            "happy path, using pat",
			client:          suite.Client.APIWithPAT,
			ctx:             context.Background(),
			expectedResults: 2,
		},
		{
			name:            "another user, no entities should be returned",
			client:          suite.Client.API,
			ctx:             th.SharedTestUser2.UserCtx,
			expectedResults: 0,
		},
	}

	for _, tc := range testCases {
		t.Run("List "+tc.name, func(t *testing.T) {
			resp, err := tc.client.GetAllEntities(tc.ctx)
			assert.NilError(t, err)
			assert.Assert(t, resp != nil)

			assert.Check(t, is.Len(resp.Entities.Edges, tc.expectedResults))
		})
	}

	(&th.Cleanup[*generated.EntityDeleteOne]{Client: suite.Client.DB.Entity, IDs: []string{entity1.ID, entity2.ID}}).MustDelete(th.SharedTestUser1.UserCtx, t)
	(&th.Cleanup[*generated.EntityTypeDeleteOne]{Client: suite.Client.DB.EntityType, IDs: []string{entity1.EntityTypeID, entity2.EntityTypeID}}).MustDelete(th.SharedTestUser1.UserCtx, t)
}

func TestMutationCreateEntity(t *testing.T) {
	entitiesToDelete := []string{}
	entityTypesToDelete := []string{}
	defaultTier := lo.ToPtr(enums.VendorTierStandard)

	entityType := (&th.EntityTypeBuilder{Client: suite.Client, Name: "superheros"}).MustNew(th.SharedTestUser1.UserCtx, t)
	entityTypeAnotherOrg := (&th.EntityTypeBuilder{Client: suite.Client, Name: "villains"}).MustNew(th.SharedTestUser2.UserCtx, t)

	testCases := []struct {
		name           string
		entityTypeName *string
		request        testclient.CreateEntityInput
		client         *testclient.TestClient
		ctx            context.Context
		expectedErr    string
	}{
		{
			name: "happy path, minimal input",
			request: testclient.CreateEntityInput{
				Name: lo.ToPtr("fraser fir"),
				Tier: defaultTier,
			},
			client: suite.Client.API,
			ctx:    th.SharedTestUser1.UserCtx,
		},
		{
			name: "happy path, all input with entity type",
			request: testclient.CreateEntityInput{
				Name:        lo.ToPtr("mitb"),
				DisplayName: lo.ToPtr("fraser fir"),
				Description: lo.ToPtr("the pine trees of appalachia"),
				Domains:     []string{"https://appalachiatrees.com"},
				Status:      &enums.EntityStatusUnderReview,
				Tier:        defaultTier,
				Note: &testclient.CreateNoteInput{
					Text: "matt is the best",
				},
			},
			entityTypeName: &entityType.Name,
			client:         suite.Client.API,
			ctx:            th.SharedAdminUser.UserCtx,
		},
		{
			name: "not allowed to use another org's entity type",
			request: testclient.CreateEntityInput{
				Name: lo.ToPtr("peter pan"),
				Tier: defaultTier,
			},
			entityTypeName: &entityTypeAnotherOrg.Name,
			client:         suite.Client.API,
			ctx:            th.SharedTestUser1.UserCtx,
			expectedErr:    "invalid or unparsable field: entity_type_name",
		},
		{
			name: "happy path, using api token",
			request: testclient.CreateEntityInput{
				Name: lo.ToPtr("douglas fir"),
				Tier: defaultTier,
			},
			client: suite.Client.APIWithToken,
			ctx:    context.Background(),
		},
		{
			name: "happy path, using pat",
			request: testclient.CreateEntityInput{
				Name:    lo.ToPtr("blue spruce"),
				OwnerID: &th.SharedTestUser1.OrganizationID,
				Tier:    defaultTier,
			},
			client: suite.Client.APIWithPAT,
			ctx:    context.Background(),
		},
		{
			name: "do not create if not allowed",
			request: testclient.CreateEntityInput{
				Name: lo.ToPtr("test-entity"),
				Tier: defaultTier,
			},
			client:      suite.Client.API,
			ctx:         th.SharedViewOnlyUser.UserCtx,
			expectedErr: th.NotAuthorizedErrorMsg,
		},
		{
			name: "missing name, but display name provided",
			request: testclient.CreateEntityInput{
				DisplayName: lo.ToPtr("fraser firs"),
				Tier:        defaultTier,
			},
			client: suite.Client.API,
			ctx:    th.SharedTestUser1.UserCtx,
		},
		{
			name: "name already exists, different casing",
			request: testclient.CreateEntityInput{
				Name: lo.ToPtr("Blue spruce"),
				Tier: defaultTier,
			},
			client:      suite.Client.API,
			ctx:         th.SharedTestUser1.UserCtx,
			expectedErr: "entity already exists",
		},
		{
			name: "invalid domain(s)",
			request: testclient.CreateEntityInput{
				Name:    lo.ToPtr("stone pines"),
				Domains: []string{"appalachiatrees"},
				Tier:    defaultTier,
			},
			client:      suite.Client.API,
			ctx:         th.SharedTestUser1.UserCtx,
			expectedErr: "invalid or unparsable field: domains",
		},
	}

	for _, tc := range testCases {
		t.Run("Create "+tc.name, func(t *testing.T) {
			resp, err := tc.client.CreateEntity(tc.ctx, tc.request, tc.entityTypeName, nil, nil, nil, nil)
			if tc.expectedErr != "" {
				assert.ErrorContains(t, err, tc.expectedErr)

				return
			}

			assert.NilError(t, err)
			assert.Assert(t, resp != nil)

			// Name is set to the Display Name if not provided
			if tc.request.Name == nil {
				assert.Check(t, is.Contains(*resp.CreateEntity.Entity.Name, *tc.request.DisplayName))
			} else {
				assert.Check(t, is.Equal(*tc.request.Name, *resp.CreateEntity.Entity.Name))
			}

			// Display Name is set to the Name if not provided
			if tc.request.DisplayName == nil {
				assert.Check(t, is.Equal(*tc.request.Name, *resp.CreateEntity.Entity.DisplayName))
			} else {
				assert.Check(t, is.Equal(*tc.request.DisplayName, *resp.CreateEntity.Entity.DisplayName))
			}

			if tc.request.Description == nil {
				assert.Check(t, is.Equal(*resp.CreateEntity.Entity.Description, ""))
			} else {
				assert.Check(t, is.Equal(*tc.request.Description, *resp.CreateEntity.Entity.Description))
			}

			if tc.request.Domains != nil {
				assert.Check(t, is.DeepEqual(tc.request.Domains, resp.CreateEntity.Entity.Domains))
			}

			if tc.request.Status != nil {
				assert.Check(t, is.DeepEqual(tc.request.Status, resp.CreateEntity.Entity.Status))
			} else {
				// default status is active
				assert.Check(t, is.Equal(enums.EntityStatusActive, *resp.CreateEntity.Entity.Status))
			}

			if tc.request.Note != nil {
				assert.Check(t, is.Len(resp.CreateEntity.Entity.Notes.Edges, 1))
				assert.Check(t, is.Equal(tc.request.Note.Text, resp.CreateEntity.Entity.Notes.Edges[0].Node.Text))
			}

			if tc.entityTypeName != nil {
				assert.Check(t, resp.CreateEntity.Entity.EntityType != nil)
				assert.Check(t, is.Equal(*tc.entityTypeName, resp.CreateEntity.Entity.EntityType.Name))
			} else {
				assert.Check(t, resp.CreateEntity.Entity.EntityType == nil)
			}

			entitiesToDelete = append(entitiesToDelete, resp.CreateEntity.Entity.ID)

			if resp.CreateEntity.Entity.EntityType != nil {
				entityTypesToDelete = append(entityTypesToDelete, resp.CreateEntity.Entity.EntityType.ID)
			}
		})
	}

	(&th.Cleanup[*generated.EntityDeleteOne]{Client: suite.Client.DB.Entity, IDs: entitiesToDelete}).MustDelete(th.SharedTestUser1.UserCtx, t)
	(&th.Cleanup[*generated.EntityTypeDeleteOne]{Client: suite.Client.DB.EntityType, IDs: entityTypesToDelete}).MustDelete(th.SharedTestUser1.UserCtx, t)
}

func TestMutationCreateEntityEnrichment(t *testing.T) {
	systemCtx := th.SetContext(th.SharedSystemAdminUser.UserCtx, suite.Client.DB)

	name := "Enriched Vendor " + ulids.New().String()
	description := "Seeded subprocessor description"
	logoRemoteURL := "https://example.com/enriched-logo.png"

	subprocessor := (&th.SubprocessorBuilder{
		Client:        suite.Client,
		Name:          name,
		Description:   description,
		LogoRemoteURL: logoRemoteURL,
	}).MustNew(th.SharedSystemAdminUser.UserCtx, t)

	resp, err := suite.Client.API.CreateEntity(th.SharedTestUser1.UserCtx, testclient.CreateEntityInput{
		Name: lo.ToPtr(strings.ToUpper(name)),
		Tier: lo.ToPtr(enums.VendorTierStandard),
	}, nil, nil, nil, nil, nil)

	assert.NilError(t, err)
	assert.Assert(t, resp != nil)

	assert.Check(t, is.Equal(description, *resp.CreateEntity.Entity.Description))
	assert.Check(t, is.Equal(logoRemoteURL, *resp.CreateEntity.Entity.LogoRemoteURL))

	userDescription := "User provided description"
	userLogoURL := "https://example.com/requested-logo.png"

	entityResp, err := suite.Client.API.CreateEntity(th.SharedTestUser1.UserCtx, testclient.CreateEntityInput{
		Name:          lo.ToPtr(name + " entity"),
		DisplayName:   lo.ToPtr(name),
		Description:   lo.ToPtr(userDescription),
		Tier:          lo.ToPtr(enums.VendorTierStandard),
		LogoRemoteURL: lo.ToPtr(userLogoURL),
	}, nil, nil, nil, nil, nil)

	assert.NilError(t, err)
	assert.Assert(t, entityResp != nil)
	assert.Check(t, is.Equal(userDescription, *entityResp.CreateEntity.Entity.Description))
	assert.Check(t, is.Equal(userLogoURL, *entityResp.CreateEntity.Entity.LogoRemoteURL))

	(&th.Cleanup[*generated.EntityDeleteOne]{Client: suite.Client.DB.Entity, IDs: []string{resp.CreateEntity.Entity.ID, entityResp.CreateEntity.Entity.ID}}).MustDelete(th.SharedTestUser1.UserCtx, t)
	(&th.Cleanup[*generated.SubprocessorDeleteOne]{Client: suite.Client.DB.Subprocessor, ID: subprocessor.ID}).MustDelete(systemCtx, t)
}

func TestMutationAdoptEntity(t *testing.T) {
	systemCtx := th.SetContext(th.SharedSystemAdminUser.UserCtx, suite.Client.DB)

	name := "Catalog Vendor " + ulids.New().String()
	displayName := "Catalog Vendor"
	description := "Vetted catalogue description"
	domains := []string{"https://catalog-vendor.example.com"}
	logoRemoteURL := "https://example.com/catalog-logo.png"

	catalogEntity, err := suite.Client.DB.Entity.Create().
		SetName(name).
		SetDisplayName(displayName).
		SetDescription(description).
		SetDomains(domains).
		SetLogoRemoteURL(logoRemoteURL).
		Save(systemCtx)
	assert.NilError(t, err)
	assert.Check(t, catalogEntity.SystemOwned)

	resp, err := suite.Client.API.AdoptEntity(th.SharedTestUser1.UserCtx, catalogEntity.ID)
	assert.NilError(t, err)
	assert.Assert(t, resp != nil)

	adopted := resp.AdoptEntity.Entity
	assert.Check(t, is.Equal(name, *adopted.Name))
	assert.Check(t, is.Equal(displayName, *adopted.DisplayName))
	assert.Check(t, is.Equal(description, *adopted.Description))
	assert.Check(t, is.DeepEqual(domains, adopted.Domains))
	assert.Check(t, is.Equal(logoRemoteURL, *adopted.LogoRemoteURL))
	assert.Check(t, is.Equal(th.SharedTestUser1.OrganizationID, *adopted.OwnerID))
	assert.Assert(t, adopted.CatalogEntity != nil)
	assert.Check(t, is.Equal(catalogEntity.ID, adopted.CatalogEntity.ID))
	assert.Assert(t, adopted.EntityType != nil)
	assert.Check(t, is.Equal("vendor", adopted.EntityType.Name))

	vendorTypeID, err := suite.Client.DB.EntityType.Query().
		Where(entitytype.NameEqualFold("vendor"), entitytype.OwnerID(th.SharedTestUser1.OrganizationID)).
		OnlyID(th.SetContext(th.SharedTestUser1.UserCtx, suite.Client.DB))
	assert.NilError(t, err)
	assert.Check(t, is.Equal(vendorTypeID, adopted.EntityType.ID))

	// adopting again is idempotent per organization
	again, err := suite.Client.API.AdoptEntity(th.SharedTestUser1.UserCtx, catalogEntity.ID)
	assert.NilError(t, err)
	assert.Check(t, is.Equal(adopted.ID, again.AdoptEntity.Entity.ID))

	// a second organization gets its own copy
	other, err := suite.Client.API.AdoptEntity(th.SharedTestUser2.UserCtx, catalogEntity.ID)
	assert.NilError(t, err)
	assert.Check(t, adopted.ID != other.AdoptEntity.Entity.ID)
	assert.Check(t, is.Equal(th.SharedTestUser2.OrganizationID, *other.AdoptEntity.Entity.OwnerID))
	assert.Check(t, is.Equal(catalogEntity.ID, other.AdoptEntity.Entity.CatalogEntity.ID))

	// an organization row is not a catalogue row
	_, err = suite.Client.API.AdoptEntity(th.SharedTestUser1.UserCtx, adopted.ID)
	assert.ErrorContains(t, err, "entity not found")

	(&th.Cleanup[*generated.EntityDeleteOne]{Client: suite.Client.DB.Entity, ID: adopted.ID}).MustDelete(th.SharedTestUser1.UserCtx, t)
	(&th.Cleanup[*generated.EntityDeleteOne]{Client: suite.Client.DB.Entity, ID: other.AdoptEntity.Entity.ID}).MustDelete(th.SharedTestUser2.UserCtx, t)
	(&th.Cleanup[*generated.EntityDeleteOne]{Client: suite.Client.DB.Entity, ID: catalogEntity.ID}).MustDelete(systemCtx, t)
}

func TestMutationUpdateEntity(t *testing.T) {
	entity := (&th.EntityBuilder{Client: suite.Client}).MustNew(th.SharedTestUser1.UserCtx, t)
	numNotes := 0
	numDomains := 0

	testCases := []struct {
		name        string
		request     testclient.UpdateEntityInput
		client      *testclient.TestClient
		ctx         context.Context
		expectedErr string
	}{
		{
			name: "happy path, update display name",
			request: testclient.UpdateEntityInput{
				DisplayName:    lo.ToPtr("blue spruce"),
				ApprovedForUse: lo.ToPtr(true),
				Note: &testclient.CreateNoteInput{
					Text: "the pine tree with blue-green colored needles",
				},
			},
			client: suite.Client.API,
			ctx:    th.SharedTestUser1.UserCtx,
		},
		{
			name: "update description using api token",
			request: testclient.UpdateEntityInput{
				Description: lo.ToPtr("the pine tree with blue-green colored needles"),
				Status:      &enums.EntityStatusDraft,
			},
			client: suite.Client.APIWithToken,
			ctx:    context.Background(),
		},
		{
			name: "update notes, domains using personal access token",
			request: testclient.UpdateEntityInput{
				Note: &testclient.CreateNoteInput{
					Text: "the pine tree with blue-green colored needles",
				},
				Domains: []string{"https://appalachiatrees.com"},
				Status:  &enums.EntityStatusActive,
			},
			client: suite.Client.APIWithPAT,
			ctx:    context.Background(),
		},
		{
			name: "update status and domain",
			request: testclient.UpdateEntityInput{
				Status:         &enums.EntityStatusSuspended,
				AppendDomains:  []string{"example.com"},
				ApprovedForUse: lo.ToPtr(false),
			},
			client: suite.Client.API,
			ctx:    th.SharedTestUser1.UserCtx,
		},
		{
			name: "conflicting status and approved for use, approved should take precedence",
			request: testclient.UpdateEntityInput{
				Status:         &enums.EntityStatusActive,
				ApprovedForUse: lo.ToPtr(false),
			},
			client: suite.Client.API,
			ctx:    th.SharedTestUser1.UserCtx,
		},
		{
			name: "not allowed to update",
			request: testclient.UpdateEntityInput{
				Description: lo.ToPtr("pine trees of the west"),
			},
			client:      suite.Client.API,
			ctx:         th.SharedViewOnlyUser.UserCtx,
			expectedErr: th.NotAuthorizedErrorMsg,
		},
		{
			name: "not allowed to update, no access to entity",
			request: testclient.UpdateEntityInput{
				Description: lo.ToPtr("pine trees of the west"),
			},
			client:      suite.Client.API,
			ctx:         th.SharedTestUser2.UserCtx,
			expectedErr: "entity not found",
		},
	}

	for _, tc := range testCases {
		t.Run("Update "+tc.name, func(t *testing.T) {
			resp, err := tc.client.UpdateEntity(tc.ctx, entity.ID, tc.request, nil, nil, nil, nil)
			if tc.expectedErr != "" {

				assert.ErrorContains(t, err, tc.expectedErr)

				return
			}

			assert.NilError(t, err)
			assert.Assert(t, resp != nil)

			if tc.request.Description != nil {
				assert.Check(t, is.Equal(*tc.request.Description, *resp.UpdateEntity.Entity.Description))
			}

			if tc.request.DisplayName != nil {
				assert.Check(t, is.Equal(*tc.request.DisplayName, *resp.UpdateEntity.Entity.DisplayName))
			}

			if tc.request.Status != nil {
				assert.Check(t, is.Equal(*tc.request.Status, *resp.UpdateEntity.Entity.Status))
			}

			if tc.request.Domains != nil {
				numDomains++

				assert.Check(t, is.Contains(resp.UpdateEntity.Entity.Domains, tc.request.Domains[0]))
				assert.Check(t, is.Len(resp.UpdateEntity.Entity.Domains, numDomains))
			}

			if tc.request.AppendDomains != nil {
				numDomains++

				assert.Check(t, is.Contains(resp.UpdateEntity.Entity.Domains, tc.request.AppendDomains[0]))
				assert.Check(t, is.Len(resp.UpdateEntity.Entity.Domains, numDomains))
			}

			if tc.request.Note != nil {
				numNotes++

				assert.Check(t, is.Len(resp.UpdateEntity.Entity.Notes.Edges, numNotes))
				assert.Check(t, is.Equal(tc.request.Note.Text, resp.UpdateEntity.Entity.Notes.Edges[0].Node.Text))
			}

			// if approed for use is set, it should always respect that over status
			if tc.request.ApprovedForUse != nil {
				assert.Check(t, is.Equal(*tc.request.ApprovedForUse, *resp.UpdateEntity.Entity.ApprovedForUse))
			} else if tc.request.Status != nil {
				// else its done based on status, where active and approved are approved
				status := *tc.request.Status
				if slices.Contains([]enums.EntityStatus{enums.EntityStatusApproved, enums.EntityStatusActive}, status) {
					assert.Check(t, *resp.UpdateEntity.Entity.ApprovedForUse)
				} else {
					assert.Check(t, !*resp.UpdateEntity.Entity.ApprovedForUse)
				}
			}
		})
	}

	(&th.Cleanup[*generated.EntityDeleteOne]{Client: suite.Client.DB.Entity, ID: entity.ID}).MustDelete(th.SharedTestUser1.UserCtx, t)
	(&th.Cleanup[*generated.EntityTypeDeleteOne]{Client: suite.Client.DB.EntityType, ID: entity.EntityTypeID}).MustDelete(th.SharedTestUser1.UserCtx, t)
}

func TestMutationDeleteEntity(t *testing.T) {
	entity1 := (&th.EntityBuilder{Client: suite.Client}).MustNew(th.SharedTestUser1.UserCtx, t)
	entity2 := (&th.EntityBuilder{Client: suite.Client}).MustNew(th.SharedTestUser1.UserCtx, t)
	entity3 := (&th.EntityBuilder{Client: suite.Client}).MustNew(th.SharedTestUser1.UserCtx, t)

	testCases := []struct {
		name        string
		idToDelete  string
		client      *testclient.TestClient
		ctx         context.Context
		expectedErr string
	}{
		{
			name:        "not allowed to delete",
			idToDelete:  entity1.ID,
			client:      suite.Client.API,
			ctx:         th.SharedViewOnlyUser.UserCtx,
			expectedErr: th.NotAuthorizedErrorMsg,
		},
		{
			name:       "happy path, delete entity",
			idToDelete: entity1.ID,
			client:     suite.Client.API,
			ctx:        th.SharedTestUser1.UserCtx,
		},
		{
			name:        "entity already deleted, not found",
			idToDelete:  entity1.ID,
			client:      suite.Client.API,
			ctx:         th.SharedTestUser1.UserCtx,
			expectedErr: "entity not found",
		},
		{
			name:       "happy path, delete entity using api token",
			idToDelete: entity2.ID,
			client:     suite.Client.APIWithToken,
			ctx:        context.Background(),
		},
		{
			name:       "happy path, delete entity using personal access token",
			idToDelete: entity3.ID,
			client:     suite.Client.APIWithPAT,
			ctx:        context.Background(),
		},
		{
			name:        "unknown entity, not found",
			idToDelete:  ulids.New().String(),
			client:      suite.Client.API,
			ctx:         th.SharedTestUser1.UserCtx,
			expectedErr: "entity not found",
		},
	}

	for _, tc := range testCases {
		t.Run("Delete "+tc.name, func(t *testing.T) {
			resp, err := tc.client.DeleteEntity(tc.ctx, tc.idToDelete)
			if tc.expectedErr != "" {
				assert.ErrorContains(t, err, tc.expectedErr)

				return
			}

			assert.NilError(t, err)
			assert.Assert(t, resp != nil)
			assert.Check(t, is.Equal(tc.idToDelete, resp.DeleteEntity.DeletedID))
		})
	}

	(&th.Cleanup[*generated.EntityTypeDeleteOne]{Client: suite.Client.DB.EntityType, IDs: []string{entity1.EntityTypeID, entity2.EntityTypeID, entity3.EntityTypeID}}).MustDelete(th.SharedTestUser1.UserCtx, t)
}
