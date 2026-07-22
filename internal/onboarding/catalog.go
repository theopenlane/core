package onboarding

import (
	"context"

	"entgo.io/ent/dialect/sql"

	"github.com/theopenlane/core/common/enums"
	"github.com/theopenlane/core/common/models"
	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/generated/standard"
	"github.com/theopenlane/core/pkg/catalog"
	"github.com/theopenlane/core/pkg/catalog/gencatalog"
)

func Catalog(ctx context.Context, client *generated.Client) (models.Questionnaire, error) {
	questionnaire := defaultQuestionnaire

	opts, err := getFrameworkOptions(ctx, client)
	if err != nil {
		return models.Questionnaire{}, err
	}

	for stepIndex := range questionnaire.Steps {
		if questionnaire.Steps[stepIndex].DynamicCards {
			questionnaire.Steps[stepIndex].Cards = getCards(client)
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

// getCards returns onboarding cards backed by catalog modules.
func getCards(client *generated.Client) []models.Card {
	useSandbox := client != nil && client.EntConfig != nil && client.EntConfig.Modules.UseSandbox
	cat := catalog.FilterByAudience(false, false, gencatalog.GetDefaultCatalog(useSandbox))

	cards := make([]models.Card, 0, len(models.TrialModules))

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

		cards = append(cards, models.Card{
			Key:         key,
			Title:       module.DisplayName,
			Description: description,
		})
	}

	return cards
}

func getFrameworkOptions(ctx context.Context, client *generated.Client) ([]models.QuestionOption, error) {
	if client == nil {
		return nil, nil
	}

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
		All(ctx)
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
			Value:       std.Framework,
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
