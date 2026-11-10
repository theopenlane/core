package hooks_test

import (
	"context"
	"fmt"
	"slices"
	"sync"
	"testing"
	"time"

	"github.com/brianvoe/gofakeit/v7"
	"github.com/theopenlane/core/common/enums"
	"github.com/theopenlane/iam/auth"

	"github.com/theopenlane/core/v2/internal/ent/entityops"
	"github.com/theopenlane/core/v2/internal/ent/generated"
	"github.com/theopenlane/core/v2/internal/ent/generated/organization"
	"github.com/theopenlane/core/v2/internal/ent/generated/program"
	"github.com/theopenlane/core/v2/internal/ent/generated/sladefinition"
	"github.com/theopenlane/core/v2/internal/ent/generated/tagdefinition"
	"github.com/theopenlane/core/v2/internal/ent/generated/task"
	"github.com/theopenlane/core/v2/internal/ent/taskrules"
	"github.com/theopenlane/core/v2/pkg/gala"

	"gotest.tools/v3/assert"
	is "gotest.tools/v3/assert/cmp"
)

func TestHookOnboarding(t *testing.T) {
	s := suite.setupTest(t)

	user := s.seedUser()

	userCtx := auth.NewTestContextWithOrgID(user.ID, user.Edges.OrgMemberships[0].OrganizationID)

	// add the client to the context for hooks
	userCtx = generated.NewContext(userCtx, s.client)

	name := "MITB"

	testCases := []struct {
		name        string
		input       generated.CreateOnboardingInput
		expectedErr string
	}{
		{
			name: "valid onboarding, full data",
			input: generated.CreateOnboardingInput{
				CompanyName: name,
				Domains:     []string{gofakeit.DomainName(), gofakeit.DomainName()},
				CompanyDetails: map[string]any{
					"sector":       "Technology",
					"company_size": "100-500",
				},
				UserDetails: map[string]any{
					"name":       gofakeit.Name(),
					"job_title":  gofakeit.JobTitle(),
					"department": gofakeit.JobDescriptor(),
				},
				Compliance: map[string]any{
					"existing_policies": true,
					"existing_controls": false,
					"risk_assessment":   true,
				},
			},
		},
		{
			name: "valid onboarding, same name, no details",
			input: generated.CreateOnboardingInput{
				CompanyName: name,
			},
		},
		{
			name: "valid onboarding, same name again, with domains",
			input: generated.CreateOnboardingInput{
				CompanyName: name,
				Domains:     []string{gofakeit.DomainName(), gofakeit.DomainName()},
			},
		},
		{
			name: "invalid onboarding, empty name",
			input: generated.CreateOnboardingInput{
				Domains: []string{gofakeit.DomainName(), gofakeit.DomainName()},
			},
			expectedErr: "company name is required",
		},
	}

	for i, tc := range testCases {
		t.Run("Create "+tc.name, func(t *testing.T) {
			// setup the allow context
			ctx := auth.WithInternalOperationContext(userCtx)

			// add the client to the context
			ctx = generated.NewContext(ctx, s.client)

			onboarding, err := s.client.Onboarding.Create().SetInput(tc.input).Save(ctx)
			if tc.expectedErr != "" {

				assert.Assert(t, err != nil)
				assert.Check(t, is.Nil(onboarding))

				return
			}

			assert.NilError(t, err)
			assert.Check(t, onboarding != nil)

			assert.Check(t, onboarding.ID != "")
			assert.Check(t, onboarding.OrganizationID != "")
			assert.Check(t, is.Equal(name, onboarding.CompanyName))

			org, err := s.client.Organization.Query().
				Where(organization.IDEQ(onboarding.OrganizationID)).
				WithSetting().
				Only(ctx)
			assert.NilError(t, err)

			assert.Check(t, is.Equal(tc.input.CompanyName, org.DisplayName))

			if i == 0 {
				assert.Check(t, is.Equal(tc.input.CompanyName, org.Name))
			} else {
				assert.Check(t, tc.input.CompanyName != org.Name)
			}

			assert.Check(t, is.DeepEqual(slices.Sorted(slices.Values(tc.input.Domains)), slices.Sorted(slices.Values(org.Edges.Setting.Domains))))

			newOrgCtx := generated.NewContext(auth.NewTestContextWithOrgID(user.ID, onboarding.OrganizationID), s.client)
			newOrgCtx = auth.WithInternalOperationContext(newOrgCtx)

			managedTagExists, err := s.client.TagDefinition.Query().
				Where(
					tagdefinition.OwnerID(onboarding.OrganizationID),
					tagdefinition.NameEqualFold("managed"),
				).
				Exist(newOrgCtx)
			assert.NilError(t, err)
			assert.Check(t, managedTagExists)

			slaCount, err := s.client.SLADefinition.Query().
				Where(sladefinition.OwnerID(onboarding.OrganizationID)).
				Count(newOrgCtx)
			assert.NilError(t, err)
			assert.Check(t, is.Equal(4, slaCount))
		})
	}
}

func TestOnboardingProgramControlsLoadableByListeners(t *testing.T) {
	s := suite.setupTest(t)

	admin := s.seedSystemAdmin()
	sysCtx := generated.NewContext(auth.NewTestContextForSystemAdmin(admin.ID, admin.Edges.OrgMemberships[0].OrganizationID), s.client)
	framework := gofakeit.UUID()

	std, err := s.client.Standard.Create().SetName(framework).SetShortName(framework).
		SetFramework(framework).SetIsPublic(true).SetSystemOwned(true).
		SetStatus(enums.StandardActive).Save(sysCtx)
	assert.NilError(t, err)

	const controlCount = 8

	for i := range controlCount {
		_, err = s.client.Control.Create().SetStandardID(std.ID).
			SetSystemOwned(true).SetRefCode(fmt.Sprintf("LOAD-%d", i)).Save(sysCtx)
		assert.NilError(t, err)
	}

	var (
		mu        sync.Mutex
		delivered int
		missing   []string
	)

	listenerIDs, err := gala.Register(s.galaRuntime, gala.Definition[entityops.MutationPayload]{
		Topic:      entityops.MutationTopic(entityops.MutationConcernWorkflow, generated.TypeControl),
		Name:       "test.control.loadable",
		Operations: []string{entityops.OpCreate},
		Caller: func(restored *auth.Caller, _ entityops.MutationPayload) *auth.Caller {
			return restored
		},
		Handle: func(hc gala.HandlerContext, payload entityops.MutationPayload) error {
			_, ok, loadErr := entityops.LoadEntity(generated.NewContext(hc.Context, s.client), payload.EntityID, s.client.Control.Get)

			mu.Lock()
			defer mu.Unlock()

			delivered++

			if loadErr != nil || !ok {
				missing = append(missing, payload.EntityID)
			}

			return nil
		},
	})
	assert.NilError(t, err)

	t.Cleanup(func() {
		assert.NilError(t, s.galaRuntime.RemoveListeners(context.Background(), listenerIDs...))
	})

	user := s.seedUser()
	ctx := generated.NewContext(auth.NewTestContextWithOrgID(user.ID, user.Edges.OrgMemberships[0].OrganizationID), s.client)
	ctx = auth.WithInternalOperationContext(ctx)

	_, err = s.client.Onboarding.Create().SetInput(generated.CreateOnboardingInput{
		CompanyName: "Onboarding Co " + gofakeit.LetterN(8),
		Compliance:  map[string]any{"frameworks": []string{framework}},
	}).Save(ctx)
	assert.NilError(t, err)

	s.waitForEvents()

	mu.Lock()
	defer mu.Unlock()

	assert.Check(t, is.Equal(controlCount, delivered))
	assert.Check(t, is.Len(missing, 0))
}

func TestOnboardingProgramFrameworkSelections(t *testing.T) {
	s := suite.setupTest(t)

	admin := s.seedSystemAdmin()
	sysCtx := generated.NewContext(auth.NewTestContextForSystemAdmin(admin.ID, admin.Edges.OrgMemberships[0].OrganizationID), s.client)
	framework := gofakeit.UUID()

	std, err := s.client.Standard.Create().SetName(framework).SetShortName(framework).
		SetFramework(framework).SetIsPublic(true).SetSystemOwned(true).
		SetStatus(enums.StandardActive).Save(sysCtx)
	assert.NilError(t, err)

	_, err = s.client.Control.Create().SetStandardID(std.ID).
		SetSystemOwned(true).SetRefCode("TEST-1").Save(sysCtx)
	assert.NilError(t, err)

	for _, tc := range []struct {
		name         string
		compliance   map[string]any
		programCount int
		controlCount int
	}{
		{name: "no compliance"},
		{name: "no frameworks", compliance: map[string]any{"existing_controls": true}},
		{name: "empty frameworks", compliance: map[string]any{"frameworks": []string{}}},
		{
			name:         "single framework",
			compliance:   map[string]any{"frameworks": []string{framework}},
			programCount: 1,
			controlCount: 1,
		},
		{
			name:         "other framework",
			compliance:   map[string]any{"frameworks": []string{"other"}},
			programCount: 1,
		},
	} {

		t.Run(tc.name, func(t *testing.T) {
			s := s.withT(t)

			user := s.seedUser()
			ctx := generated.NewContext(auth.NewTestContextWithOrgID(user.ID, user.Edges.OrgMemberships[0].OrganizationID), s.client)
			ctx = auth.WithInternalOperationContext(ctx)

			onboarding, err := s.client.Onboarding.Create().SetInput(generated.CreateOnboardingInput{
				CompanyName: "Onboarding Co " + gofakeit.LetterN(8),
				Compliance:  tc.compliance,
			}).Save(ctx)

			assert.NilError(t, err)

			s.waitForEvents()

			programs, err := s.client.Program.Query().Where(program.OwnerIDEQ(onboarding.OrganizationID)).WithControls().All(ctx)
			assert.NilError(t, err)
			assert.Assert(t, is.Len(programs, tc.programCount))

			if tc.programCount > 0 {
				assert.Check(t, is.Len(programs[0].Edges.Controls, tc.controlCount))
				year := time.Now().Year()
				expectedName := fmt.Sprintf("Compliance Program %d", year)
				if tc.name == "single framework" {
					expectedName = fmt.Sprintf("%s Program %d", framework, year)
				}
				assert.Check(t, is.Equal(expectedName, programs[0].Name))
				assert.Check(t, is.Equal(fmt.Sprintf("Track %s compliance activities, evidence, and audit readiness for %d.", programs[0].FrameworkName, year), programs[0].Description))
			}

			taskCount, err := s.client.Task.Query().Where(
				task.OwnerIDEQ(onboarding.OrganizationID),
				task.SourceKeyEQ("onboarding-"+taskrules.RuleFrameworkGeneric),
			).Count(ctx)
			assert.NilError(t, err)

			if tc.programCount == 0 {
				assert.Check(t, is.Equal(1, taskCount))
			} else {
				assert.Check(t, is.Equal(0, taskCount))
			}
		})
	}
}
