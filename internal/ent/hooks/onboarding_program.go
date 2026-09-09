package hooks

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"entgo.io/ent/dialect/sql"
	"github.com/theopenlane/iam/auth"

	"github.com/theopenlane/core/common/enums"

	"github.com/theopenlane/core/v2/internal/controls"
	"github.com/theopenlane/core/v2/internal/ent/generated"
	"github.com/theopenlane/core/v2/internal/ent/generated/privacy"
	"github.com/theopenlane/core/v2/internal/ent/generated/standard"
)

func createProgram(ctx context.Context, client *generated.Client, org *generated.Organization, complianceData map[string]interface{}) error {
	data, err := json.Marshal(complianceData["frameworks"])
	if err != nil {
		return err
	}

	var frameworks []string
	if err := json.Unmarshal(data, &frameworks); err != nil {
		return fmt.Errorf("invalid onboarding frameworks: %w", err)
	}

	if len(frameworks) == 0 {
		return nil
	}

	caller, ok := auth.CallerFromContext(ctx)
	if !ok || caller == nil {
		return auth.ErrNoAuthUser
	}

	newCaller := *caller
	newCaller.OrganizationID = org.ID

	ctx = auth.WithCaller(ctx, &newCaller)
	ctx = privacy.DecisionContext(ctx, privacy.Allow)

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
				standard.IsPublic(true),
				standard.FrameworkNotIn("openlane-standard", "openlane-trust-center"),
			).
			Order(standard.ByVersion(sql.OrderDesc()), standard.ByID()).
			First(ctx)
		if err != nil {
			return fmt.Errorf("resolve onboarding framework %q: %w", framework, err)
		}

		standards = append(standards, std)
		label := std.ShortName
		if label == "" {
			label = std.Name
		}

		labels = append(labels, label)
	}

	builder := client.Program.Create().
		SetOwnerID(org.ID).
		SetName(org.DisplayName + " Compliance Program").
		SetFrameworkName(strings.Join(labels, ", "))

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
			input, _ := controls.CreateCloneControlInput(source, &program.ID, org.ID)
			cloned, err := client.Control.Create().SetInput(input).Save(ctx)
			if err != nil {
				return err
			}

			for _, subcontrol := range source.Edges.Subcontrols {
				input := controls.CreateCloneSubcontrolInput(subcontrol, org.ID, controls.SubcontrolToCreate{RefControl: source})
				input.ControlID = cloned.ID
				if err := client.Subcontrol.Create().SetInput(*input).Exec(ctx); err != nil {
					return err
				}
			}
		}
	}

	return nil
}
