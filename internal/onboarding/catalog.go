package onboarding

import (
	"context"
	"sort"

	"github.com/theopenlane/core/common/enums"
	"github.com/theopenlane/core/common/models"
	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/generated/standard"
	"github.com/theopenlane/core/pkg/catalog"
	"github.com/theopenlane/core/pkg/catalog/gencatalog"
)

// frameworkOrder pins the display order of these frameworks ahead of the rest
var frameworkOrder = map[string]int{
	"SOC 2":     0, //nolint:mnd
	"ISO 42001": 1, //nolint:mnd
	"ISO 27001": 2, //nolint:mnd
	"HIPAA":     3, //nolint:mnd
	"PCI DSS":   4, //nolint:mnd
}

// Catalog returns the onboarding questionnaire with its dynamic cards and framework options populated
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
			standard.FrameworkNotIn("openlane-standard", "openlane-trust-center"),
			standard.SystemOwned(true),
			standard.IsPublic(true),
		).
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
		})
	}

	rank := func(label string) int {
		if r, ok := frameworkOrder[label]; ok {
			return r
		}

		return len(frameworkOrder)
	}

	sort.SliceStable(options, func(i, j int) bool {
		return rank(options[i].Label) < rank(options[j].Label)
	})

	options = append(options, models.QuestionOption{
		Value:    "other",
		Label:    "Other",
		Priority: len(options) + 1,
	})

	return options, nil
}
