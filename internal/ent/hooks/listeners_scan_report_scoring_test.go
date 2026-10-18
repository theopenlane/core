package hooks

import (
	"context"
	"testing"

	"github.com/samber/lo"
	"gotest.tools/v3/assert"
	"gotest.tools/v3/assert/cmp"

	"github.com/theopenlane/core/common/enums"
	"github.com/theopenlane/core/common/models"

	"github.com/theopenlane/core/v2/internal/ent/generated"
	"github.com/theopenlane/core/v2/internal/integrations/definitions/gemini"
)

func TestReportScoresRecordOutcomeAndYesNoAnswers(t *testing.T) {
	questions := lo.KeyBy(models.DefaultVendorScoringQuestions, func(question models.VendorScoringQuestionDef) string { return question.Key })

	scores := reportScores(context.Background(), questions, gemini.ReportAssessment{
		RecommendationOutcome: models.ReviewOutcomeLowRisk,
		Recommendation:        "No material concerns.",
		ScoringAnswers: []gemini.ReportScoringAnswer{
			{QuestionKey: "IAM-05.1", Answer: "true", Likelihood: "VERY_LOW", Notes: "Section 4, CC6.3"},
			{QuestionKey: "NOT-A-KEY", Answer: "true", Likelihood: "LOW"},
			{QuestionKey: "IAM-07.1", Answer: "false", Likelihood: "SOMETIMES"},
			{QuestionKey: models.ReviewOutcomeQuestionKey, Answer: models.ReviewOutcomeNotApproved, Likelihood: "HIGH"},
		},
	})

	assert.Assert(t, cmp.Len(scores, 2))

	assert.Check(t, cmp.Equal(scores[0].key, models.ReviewOutcomeQuestionKey))
	assert.Check(t, cmp.Equal(scores[0].answer, models.ReviewOutcomeLowRisk))
	assert.Check(t, cmp.Equal(scores[0].impact, enums.VendorRiskImpactLow))
	assert.Check(t, cmp.Equal(scores[0].likelihood, enums.VendorRiskLikelihoodMedium))
	assert.Check(t, cmp.Equal(scores[0].notes, "No material concerns."))

	assert.Check(t, cmp.Equal(scores[1].key, "IAM-05.1"))
	assert.Check(t, cmp.Equal(scores[1].impact, questions["IAM-05.1"].SuggestedImpact))
	assert.Check(t, cmp.Equal(scores[1].likelihood, enums.VendorRiskLikelihoodVeryLow))
}

func TestReportScoresWithoutOutcomeQuestion(t *testing.T) {
	scores := reportScores(context.Background(), map[string]models.VendorScoringQuestionDef{}, gemini.ReportAssessment{RecommendationOutcome: models.ReviewOutcomeLowRisk})

	assert.Check(t, cmp.Len(scores, 0))
}

func TestIsReportScoringAnswer(t *testing.T) {
	assert.Check(t, isReportScoringAnswer(&generated.VendorRiskScore{Tags: []string{"vendor", reportScoringTag}}))
	assert.Check(t, !isReportScoringAnswer(&generated.VendorRiskScore{Tags: []string{"vendor"}, Notes: lo.ToPtr("Checked by hand")}))
	assert.Check(t, !isReportScoringAnswer(&generated.VendorRiskScore{}))
}
