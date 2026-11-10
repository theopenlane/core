package hooks

import (
	"context"
	"fmt"
	"strings"
	"time"

	"entgo.io/ent/dialect/sql"
	"github.com/theopenlane/iam/auth"

	"github.com/theopenlane/core/common/enums"

	"github.com/theopenlane/core/v2/internal/controls"
	"github.com/theopenlane/core/v2/internal/ent/entityops"
	"github.com/theopenlane/core/v2/internal/ent/generated"
	"github.com/theopenlane/core/v2/internal/ent/generated/onboarding"
	"github.com/theopenlane/core/v2/internal/ent/generated/standard"
	"github.com/theopenlane/core/v2/internal/integrations/definitions/objectstore"
	intruntime "github.com/theopenlane/core/v2/internal/integrations/runtime"
	"github.com/theopenlane/core/v2/internal/workflows"
	"github.com/theopenlane/core/v2/pkg/gala"
	"github.com/theopenlane/logx"
	"github.com/theopenlane/utils/jsonx"
)

// onboardingDemoRow is the part of the onboarding row the demo data seed reads
type onboardingDemoRow struct {
	// DemoDataset is the demo dataset to seed
	DemoDataset objectstore.DemoDataset `json:"demo_dataset"`
}

func init() { registerListeners(OnboardingProgramListeners) }

// OnboardingProgramListeners sets up gala to process onboarding requests such as creating programs
// and cloning them in the background
func OnboardingProgramListeners() []gala.Registration {
	return []gala.Registration{
		entityops.MutationListener{
			Schema:     entityops.SchemaOnboarding,
			Operations: []string{entityops.OpCreate},
			Caller:     onboardingListenerCaller,
			Handle:     handleOnboardingProgram,
		},
		entityops.MutationListener{
			Schema:     entityops.SchemaOnboarding,
			Operations: []string{entityops.OpCreate},
			RowMatch:   []entityops.FieldMatch{{Field: onboarding.FieldSeedDemoData, In: []string{"true"}}},
			Caller:     onboardingListenerCaller,
			Handle:     entityops.RequireDep(handleDemoDataSeed),
		},
	}
}

// onboardingListenerCaller adds the capability needed to load the onboarding record
func onboardingListenerCaller(restored *auth.Caller, _ entityops.MutationPayload) *auth.Caller {
	return restored.WithCapabilities(auth.CapInternalOperation)
}

func handleOnboardingProgram(inv entityops.Invocation, _ entityops.MutationPayload) error {
	// cannot use MutationPayload because the organization_id needed is actually stored by the hook
	// so it will not be available here
	record, ok, err := entityops.LoadEntity(inv.Context, inv.EntityID, inv.Client.Onboarding.Get)
	if err != nil || !ok {
		return err
	}

	if len(record.Compliance) == 0 {
		return nil
	}

	_, err = workflows.WithTx(inv.Context, inv.Client, nil, func(ctx context.Context, tx *generated.Tx) (struct{}, error) {
		return struct{}{}, createProgram(ctx, tx.Client(), record.OrganizationID, record.Compliance)
	})

	return err
}

// handleDemoDataSeed queues the demo dataset seed for the new organization when the user chose demo data
func handleDemoDataSeed(inv entityops.Invocation, _ entityops.MutationPayload, rt *intruntime.Runtime) error {
	if _, ok := rt.Registry().RuntimeClient(objectstore.DefinitionID.ID()); !ok {
		logx.FromContext(inv.Context).Warn().Str("onboarding_id", inv.EntityID).Msg("objectstore runtime client is not configured, skipping onboarding demo data")

		return nil
	}

	row, err := jsonx.Decode[onboardingDemoRow](inv.Row)
	if err != nil {
		return err
	}

	req, err := objectstore.SeedDemoRequest(inv.Context, row.DemoDataset, enums.DemoDataProfileDemo)
	if err != nil {
		return err
	}

	_, err = rt.Dispatch(inv.Context, req)

	return err
}

func generateProgramName(standards []*generated.Standard, year int) string {
	if len(standards) == 1 {
		name := standards[0].ShortName
		if name == "" {
			name = standards[0].Name
		}

		return fmt.Sprintf("%s Program %d", name, year)
	}

	return fmt.Sprintf("Compliance Program %d", year)
}

func createProgram(ctx context.Context, client *generated.Client, orgID string, complianceData map[string]any) error {
	standards, labels, err := resolveOnboardingStandards(ctx, client, complianceData)
	if err != nil || len(labels) == 0 {
		return err
	}

	currentYear := time.Now().Year()

	frameworks := strings.Join(labels, ", ")

	description := fmt.Sprintf("Track %s compliance activities, evidence, and audit readiness for %d.", frameworks, currentYear)

	builder := client.Program.Create().
		SetName(generateProgramName(standards, currentYear)).
		SetDescription(description).
		SetProgramKindName(enums.ProgramTypeFramework.String()).
		SetFrameworkName(frameworks)

	if auditor, ok := complianceData["auditor_name"].(string); ok && auditor != "" {
		builder.SetAuditor(auditor)
	}

	if email, ok := complianceData["auditor_email"].(string); ok && email != "" {
		builder.SetAuditorEmail(email)
	}

	program, err := builder.Save(ctx)
	if err != nil {
		return err
	}

	// can't use worker pool here since we are in a tx, and the program was created in this tx
	cloner := controls.Cloner{DB: client, OrgID: orgID, SkipProgramAccess: true}

	for _, std := range standards {
		filters := controls.CloneFilterOptions{StandardID: &std.ID}
		if std.Framework == "soc2" {
			filters.Categories = []string{"Security"}
		}

		if _, err := cloner.FromStandard(ctx, filters, &program.ID); err != nil {
			return err
		}
	}

	return nil
}

func getOnboardingFrameworks(complianceData map[string]any) ([]string, error) {
	var frameworks []string
	if err := jsonx.RoundTrip(complianceData["frameworks"], &frameworks); err != nil {
		return nil, fmt.Errorf("invalid onboarding frameworks: %w", err)
	}

	return frameworks, nil
}

func resolveOnboardingStandards(ctx context.Context, client *generated.Client, complianceData map[string]any) ([]*generated.Standard, []string, error) {
	frameworks, err := getOnboardingFrameworks(complianceData)
	if err != nil || len(frameworks) == 0 {
		return nil, nil, err
	}

	standards := make([]*generated.Standard, 0, len(frameworks))

	labels := make([]string, 0, len(frameworks))
	seen := make(map[string]bool, len(frameworks))

	for _, framework := range frameworks {
		if seen[framework] {
			continue
		}

		seen[framework] = true

		if framework == "other" {
			labels = append(labels, "Other")
			continue
		}

		std, err := client.Standard.Query().
			Where(
				standard.FrameworkEQ(framework),
				standard.StatusEQ(enums.StandardActive),
				standard.SystemOwned(true),
			).
			Order(
				standard.ByVersion(sql.OrderDesc()),
				standard.ByID(),
			).
			First(ctx)
		if err != nil {
			return nil, nil, fmt.Errorf("resolve onboarding framework %q: %w", framework, err)
		}

		standards = append(standards, std)
		label := std.ShortName
		if label == "" {
			label = std.Name
		}

		labels = append(labels, label)
	}

	return standards, labels, nil
}
