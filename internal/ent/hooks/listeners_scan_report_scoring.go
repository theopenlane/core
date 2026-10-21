package hooks

import (
	"context"
	"strings"

	"github.com/samber/lo"

	"github.com/theopenlane/core/common/enums"
	"github.com/theopenlane/core/common/models"

	"github.com/theopenlane/core/v2/internal/ent/generated"
	"github.com/theopenlane/core/v2/internal/ent/generated/vendorriskscore"
	"github.com/theopenlane/core/v2/internal/integrations/definitions/gemini"
	"github.com/theopenlane/core/v2/pkg/logx"
)

// reportScoringTag marks a vendor scoring answer a report review wrote, so a later review may replace it while an
// answer a person gave is kept
const reportScoringTag = "soc2-report-review"

// isReportScoringAnswer reports whether a report review wrote the answer
func isReportScoringAnswer(score *generated.VendorRiskScore) bool {
	return lo.Contains(score.Tags, reportScoringTag)
}

// reportScore is one vendor scoring answer a report review records
type reportScore struct {
	// key is the scoring question answered
	key string
	// answer is the answer in the question's format
	answer string
	// impact is the impact the answer is scored with
	impact enums.VendorRiskImpact
	// likelihood is the likelihood the answer is scored with
	likelihood enums.VendorRiskLikelihood
	// notes is the evidence behind the answer
	notes string
}

// applyReportScoringAnswers records the review's outcome and the report's yes or no answers on each vendor: a
// question without an answer gets one, an answer an earlier report review wrote is replaced, and an answer a person
// gave is kept
func applyReportScoringAnswers(ctx context.Context, client *generated.Client, ownerID string, entityIDs []string, assessment gemini.ReportAssessment) error {
	questions, err := gemini.EnabledScoringQuestions(ctx, client, ownerID)
	if err != nil {
		return err
	}

	scores := reportScores(ctx, lo.KeyBy(questions, func(question models.VendorScoringQuestionDef) string { return question.Key }), assessment)
	if len(scores) == 0 {
		return nil
	}

	keys := lo.Map(scores, func(score reportScore, _ int) string { return score.key })

	for _, entityID := range entityIDs {
		existing, err := client.VendorRiskScore.Query().
			Where(vendorriskscore.EntityID(entityID), vendorriskscore.QuestionKeyIn(keys...), vendorriskscore.AssessmentResponseIDIsNil()).
			All(ctx)
		if err != nil {
			return err
		}

		existingByKey := lo.KeyBy(existing, func(score *generated.VendorRiskScore) string { return score.QuestionKey })

		for _, score := range scores {
			if err := writeReportScore(ctx, client, ownerID, entityID, existingByKey[score.key], score); err != nil {
				return err
			}
		}
	}

	return nil
}

// reportScores builds the answers a review records: the outcome scored inside its rating band, then each yes or no
// answer to an enabled question at the question's suggested impact, skipping any the model answered unusably
func reportScores(ctx context.Context, questions map[string]models.VendorScoringQuestionDef, assessment gemini.ReportAssessment) []reportScore {
	var scores []reportScore

	if outcome, ok := models.ReviewOutcomeProfiles[assessment.RecommendationOutcome]; ok {
		if _, enabled := questions[models.ReviewOutcomeQuestionKey]; enabled {
			scores = append(scores, reportScore{
				key:        models.ReviewOutcomeQuestionKey,
				answer:     assessment.RecommendationOutcome,
				impact:     outcome.Impact,
				likelihood: outcome.Likelihood,
				notes:      strings.TrimSpace(assessment.Recommendation),
			})
		}
	}

	for _, answer := range assessment.ScoringAnswers {
		question, known := questions[answer.QuestionKey]
		likelihood := enums.ToVendorRiskLikelihood(answer.Likelihood)

		if !known || question.AnswerType != enums.VendorScoringAnswerTypeBoolean || *likelihood == enums.VendorRiskLikelihoodInvalid || answer.Answer == "" {
			logx.FromContext(ctx).Debug().Str("question_key", answer.QuestionKey).Str("likelihood", answer.Likelihood).Msg("report scan: skipped an unusable vendor scoring answer")

			continue
		}

		scores = append(scores, reportScore{key: question.Key, answer: answer.Answer, impact: question.SuggestedImpact, likelihood: *likelihood, notes: answer.Notes})
	}

	return scores
}

// writeReportScore records one answer on one vendor, creating it when the question is unanswered and replacing it
// only when an earlier report review wrote it
func writeReportScore(ctx context.Context, client *generated.Client, ownerID, entityID string, current *generated.VendorRiskScore, score reportScore) error {
	notes := strings.TrimSpace(score.notes)

	switch {
	case current == nil:
		return client.VendorRiskScore.Create().
			SetOwnerID(ownerID).
			SetEntityID(entityID).
			SetQuestionKey(score.key).
			SetImpact(score.impact).
			SetLikelihood(score.likelihood).
			SetAnswer(score.answer).
			SetNillableNotes(lo.EmptyableToPtr(notes)).
			SetTags([]string{reportScoringTag}).
			Exec(ctx)
	case isReportScoringAnswer(current):
		update := client.VendorRiskScore.UpdateOneID(current.ID).
			SetImpact(score.impact).
			SetLikelihood(score.likelihood).
			SetAnswer(score.answer)

		// evidence from an earlier review would not back the new answer
		if notes == "" {
			update.ClearNotes()
		} else {
			update.SetNotes(notes)
		}

		return update.Exec(ctx)
	default:
		return nil
	}
}
