package hooks

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"entgo.io/ent/dialect/sql"
	"github.com/theopenlane/iam/auth"

	"github.com/theopenlane/core/common/enums"

	"github.com/theopenlane/core/v2/internal/controls"
	"github.com/theopenlane/core/v2/internal/ent/entityops"
	"github.com/theopenlane/core/v2/internal/ent/generated"
	"github.com/theopenlane/core/v2/internal/ent/generated/standard"
	"github.com/theopenlane/core/v2/internal/workflows"
	"github.com/theopenlane/core/v2/pkg/gala"
)

func init() { registerListeners(OnboardingProgramListeners) }

// OnboardingProgramListeners sets up gala to process onboarding requests such as creating programs
// and cloning them in the background
func OnboardingProgramListeners() []gala.Registration {
	return []gala.Registration{
		entityops.MutationListener{
			Schema:     entityops.SchemaOnboarding,
			Operations: []string{entityops.OpCreate},
			Caller: func(restored *auth.Caller, _ entityops.MutationPayload) *auth.Caller {
				return restored.WithCapabilities(auth.CapInternalOperation)
			},
			Handle: handleOnboardingProgram,
		},
	}
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

	caller := *inv.Caller
	caller.OrganizationID = record.OrganizationID
	ctx := auth.WithCaller(inv.Context, &caller)

	_, err = workflows.WithTx(ctx, inv.Client, nil, func(tx *generated.Tx) (struct{}, error) {
		return struct{}{}, createProgram(ctx, tx.Client(), record.OrganizationID, record.Compliance)
	})

	return err
}

func generateProgramName(standards []*generated.Standard, year int) string {
	if len(standards) == 1 {
		return fmt.Sprintf("%s Program %d", standards[0].ShortName, year)
	}

	return fmt.Sprintf("Compliance Program %d", year)
}

func createProgram(ctx context.Context, client *generated.Client, orgID string, complianceData map[string]interface{}) error {
	standards, labels, err := resolveOnboardingStandards(ctx, client, complianceData)
	if err != nil || len(labels) == 0 {
		return err
	}

	currentYear := time.Now().Year()

	frameworks := strings.Join(labels, ", ")

	description := fmt.Sprintf("Track %s compliance activities, evidence, and audit readiness for %d.", frameworks, currentYear)

	builder := client.Program.Create().
		SetOwnerID(orgID).
		SetName(generateProgramName(standards, currentYear)).
		SetDescription(description).
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

	for _, std := range standards {
		filters := controls.CloneFilterOptions{StandardID: &std.ID}
		if std.Framework == "soc2" {
			filters.Categories = []string{"Security"}
		}

		where, err := controls.ControlFilterByStandard(ctx, filters, std)
		if err != nil {
			return err
		}

		sources, err := client.Control.Query().Where(where...).WithStandard().WithSubcontrols().All(ctx)
		if err != nil {
			return err
		}

		// can't use worker pool here since we are in a tx
		for _, source := range sources {
			input, _ := controls.CreateCloneControlInput(source, &program.ID, orgID)
			cloned, err := client.Control.Create().SetInput(input).Save(ctx)
			if err != nil {
				return err
			}

			for _, subcontrol := range source.Edges.Subcontrols {
				input := controls.CreateCloneSubcontrolInput(subcontrol, orgID, controls.SubcontrolToCreate{RefControl: source})
				input.ControlID = cloned.ID
				if err := client.Subcontrol.Create().SetInput(*input).Exec(ctx); err != nil {
					return err
				}
			}
		}
	}

	return nil
}

func getOnboardingFrameworks(complianceData map[string]interface{}) ([]string, error) {
	data, err := json.Marshal(complianceData["frameworks"])
	if err != nil {
		return nil, err
	}

	var frameworks []string
	if err := json.Unmarshal(data, &frameworks); err != nil {
		return nil, fmt.Errorf("invalid onboarding frameworks: %w", err)
	}

	return frameworks, nil
}

func resolveOnboardingStandards(ctx context.Context, client *generated.Client, complianceData map[string]interface{}) ([]*generated.Standard, []string, error) {
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
