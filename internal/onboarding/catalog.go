package onboarding

import (
	"context"

	"entgo.io/ent/dialect/sql"

	"github.com/theopenlane/core/common/enums"
	"github.com/theopenlane/core/common/models"
	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/generated/privacy"
	"github.com/theopenlane/core/internal/ent/generated/standard"
	"github.com/theopenlane/core/pkg/catalog/gencatalog"
)

func Catalog(ctx context.Context, client *generated.Client) (models.Questionnaire, error) {
	questionnaire := defaultQuestionnaire

	opts, err := getFrameworkOptions(ctx, client)
	if err != nil {
		return models.Questionnaire{}, err
	}

	for stepIndex := range questionnaire.Steps {
		if questionnaire.Steps[stepIndex].DynamicModules {
			questionnaire.Steps[stepIndex].Modules = getModules(client)
		}

		for questionIndex := range questionnaire.Steps[stepIndex].Questions {
			question := &questionnaire.Steps[stepIndex].Questions[questionIndex]
			if question.DynamicOptions {
				question.Options = opts
			}
		}
	}

	return questionnaire, nil
}

func getModules(client *generated.Client) []models.Module {

	cat := gencatalog.DefaultCatalog

	if client != nil && client.EntConfig != nil && client.EntConfig.Modules.UseSandbox {
		cat = gencatalog.DefaultSandboxCatalog
	}

	modules := make([]models.Module, 0, len(models.TrialModules))

	for _, orgModule := range models.TrialModules {
		if orgModule == models.CatalogBaseModule {
			continue
		}

		module, ok := cat.Modules[orgModule.String()]
		if !ok {
			continue
		}

		description := module.MarketingDescription
		if description == "" {
			description = module.Description
		}

		key := module.LookupKey
		if key == "" {
			key = orgModule.String()
		}

		modules = append(modules, models.Module{
			Key:         key,
			Title:       module.DisplayName,
			Description: description,
		})
	}

	return modules
}

func getFrameworkOptions(ctx context.Context, client *generated.Client) ([]models.QuestionOption, error) {
	if client == nil {
		return nil, nil
	}

	allowCtx := privacy.DecisionContext(ctx, privacy.Allow)

	standards, err := client.Standard.Query().
		Where(
			standard.StatusEQ(enums.StandardActive),
			standard.Or(
				standard.SystemOwned(true),
				standard.IsPublic(true),
				standard.FreeToUse(true),
			),
		).
		Order(standard.ByPriority(sql.OrderDesc())).
		All(allowCtx)
	if err != nil {
		return nil, err
	}

	options := make([]models.QuestionOption, 0, len(standards))

	for _, std := range standards {
		label := std.ShortName
		if label == "" {
			label = std.Name
		}

		options = append(options, models.QuestionOption{
			Value:       std.ID,
			Label:       label,
			Description: std.Description,
			LogoURL:     std.GoverningBodyLogoURL,
			Priority:    std.Priority,
		})
	}

	options = append(options, models.QuestionOption{
		Value:    "other",
		Label:    "Other",
		Priority: len(options) + 1,
	})

	return options, nil
}
