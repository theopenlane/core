package gemini

import (
	"context"
	"fmt"
	"strings"

	"github.com/samber/lo"

	"github.com/theopenlane/core/common/enums"
	"github.com/theopenlane/core/common/models"

	"github.com/theopenlane/core/v2/internal/ent/generated"
	"github.com/theopenlane/core/v2/internal/ent/generated/vendorscoringconfig"
)

// The prompt values the assessment part is rendered with
const (
	// customerVar names the organization the vendor's report is reviewed for
	customerVar = "Customer"
	// scoringQuestionsVar lists the vendor scoring questions the assessment answers
	scoringQuestionsVar = "ScoringQuestions"
)

// EnabledScoringQuestions returns the vendor scoring questions enabled for the organization, its custom questions merged
// over the system defaults
func EnabledScoringQuestions(ctx context.Context, client *generated.Client, organizationID string) ([]models.VendorScoringQuestionDef, error) {
	cfg, err := client.VendorScoringConfig.Query().Where(vendorscoringconfig.OwnerID(organizationID)).Only(ctx)
	if err != nil && !generated.IsNotFound(err) {
		return nil, err
	}

	var questions models.VendorScoringQuestionsConfig
	if cfg != nil {
		questions = cfg.Questions
	}

	return lo.Filter(questions.All(), func(question models.VendorScoringQuestionDef, _ int) bool { return question.Enabled }), nil
}

// ScoringQuestions returns the enabled yes or no vendor scoring questions a report answers, since vendor scoring
// counts any other answer as risk whatever it says
func ScoringQuestions(ctx context.Context, client *generated.Client, organizationID string) ([]models.VendorScoringQuestionDef, error) {
	questions, err := EnabledScoringQuestions(ctx, client, organizationID)
	if err != nil {
		return nil, err
	}

	return lo.Filter(questions, func(question models.VendorScoringQuestionDef, _ int) bool {
		return question.AnswerType == enums.VendorScoringAnswerTypeBoolean
	}), nil
}

// assessmentVars returns the prompt values the assessment is written with: the organization it is reviewed for
// and the scoring questions it answers
func assessmentVars(ctx context.Context, client *generated.Client, organizationID string) (map[string]string, error) {
	org, err := client.Organization.Get(ctx, organizationID)
	if err != nil {
		return nil, err
	}

	questions, err := ScoringQuestions(ctx, client, organizationID)
	if err != nil {
		return nil, err
	}

	return map[string]string{
		customerVar:         lo.CoalesceOrEmpty(org.DisplayName, org.Name),
		scoringQuestionsVar: strings.Join(lo.Map(questions, scoringQuestionLine), "\n"),
	}, nil
}

// scoringQuestionLine renders one yes or no scoring question for the prompt with its key and wording
func scoringQuestionLine(question models.VendorScoringQuestionDef, _ int) string {
	return strings.TrimSpace(fmt.Sprintf("- %s: %s. %s", question.Key, question.Name, question.Description))
}
