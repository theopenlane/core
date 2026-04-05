package hooks

import (
	"context"
	"math"
	"strings"

	"entgo.io/ent"
	"github.com/samber/lo"

	"github.com/theopenlane/core/common/enums"
	"github.com/theopenlane/core/common/models"
	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/generated/hook"
	"github.com/theopenlane/core/internal/ent/generated/privacy"
	"github.com/theopenlane/core/internal/ent/generated/vendorriskscore"
	"github.com/theopenlane/core/internal/ent/generated/vendorscoringconfig"
)

// defaultThresholds is the fallback when no config is available
var defaultThresholds = models.RiskThresholdsConfig{}

// HookVendorScoringConfigKeyGen assigns stable keys to custom questions that lack
// a generated CUST-prefix key before the config is persisted
func HookVendorScoringConfigKeyGen() ent.Hook {
	return hook.On(func(next ent.Mutator) ent.Mutator {
		return hook.VendorScoringConfigFunc(func(ctx context.Context, m *generated.VendorScoringConfigMutation) (generated.Value, error) {
			questions, ok := m.Questions()
			if !ok {
				return next.Mutate(ctx, m)
			}

			questions.AssignCustomKeys()
			m.SetQuestions(questions)

			return next.Mutate(ctx, m)
		})
	}, ent.OpCreate|ent.OpUpdateOne|ent.OpUpdate)
}

// HookVendorRiskScoreCompute sets the score field based on impact x likelihood,
// and populates denormalized question fields from the scoring config on create
func HookVendorRiskScoreCompute() ent.Hook {
	return hook.On(func(next ent.Mutator) ent.Mutator {
		return hook.VendorRiskScoreFunc(func(ctx context.Context, m *generated.VendorRiskScoreMutation) (generated.Value, error) {
			if m.Op().Is(ent.OpCreate) {
				if err := computeOnCreate(ctx, m); err != nil {
					return nil, err
				}

				return next.Mutate(ctx, m)
			}

			if !hasScoringFieldUpdate(m) {
				return next.Mutate(ctx, m)
			}

			if err := recomputeScore(ctx, m); err != nil {
				return nil, err
			}

			return next.Mutate(ctx, m)
		})
	}, ent.OpCreate|ent.OpUpdateOne)
}

// HookVendorRiskScoreAggregate recomputes Entity.risk_score and Entity.risk_rating
// after a VendorRiskScore is created, updated, or deleted
func HookVendorRiskScoreAggregate() ent.Hook {
	return hook.On(func(next ent.Mutator) ent.Mutator {
		return hook.VendorRiskScoreFunc(func(ctx context.Context, m *generated.VendorRiskScoreMutation) (generated.Value, error) {
			allowCtx := privacy.DecisionContext(ctx, privacy.Allow)

			// Capture entity_id before deletion since the record will be gone after mutate
			var preDeleteEntityID string
			if m.Op().Is(ent.OpDeleteOne) {
				id, ok := m.ID()
				if ok {
					existing, err := m.Client().VendorRiskScore.Get(allowCtx, id)
					if err != nil && !generated.IsNotFound(err) {
						return nil, err
					}

					if existing != nil {
						preDeleteEntityID = existing.EntityID
					}
				}
			}

			v, err := next.Mutate(ctx, m)
			if err != nil {
				return v, err
			}

			entityID, err := aggregateEntityID(ctx, m, preDeleteEntityID)
			if err != nil {
				return v, err
			}
			if entityID == "" {
				return v, nil
			}

			if err := RecomputeEntityRiskAggregate(allowCtx, m.Client(), entityID); err != nil {
				return v, err
			}

			return v, nil
		})
	}, ent.OpCreate|ent.OpUpdateOne|ent.OpDeleteOne)
}

// maxPenaltyScore is the score assigned to unanswered questions under FULL_QUESTIONNAIRE mode
// (CRITICAL impact=5 x VERY_HIGH likelihood=4)
const maxPenaltyScore = 20.0

// RecomputeEntityRiskAggregate recalculates an entity's aggregate risk_score, risk_rating,
// and risk_score_coverage from all VendorRiskScore records, respecting the scoring mode
// on the associated VendorScoringConfig
func RecomputeEntityRiskAggregate(ctx context.Context, client *generated.Client, entityID string) error {
	scores, err := client.VendorRiskScore.Query().
		Where(vendorriskscore.EntityID(entityID)).
		Select(
			vendorriskscore.FieldScore,
			vendorriskscore.FieldAnswer,
			vendorriskscore.FieldQuestionKey,
			vendorriskscore.FieldVendorScoringConfigID,
		).
		All(ctx)
	if err != nil {
		return err
	}

	if len(scores) == 0 {
		return nil
	}

	// Resolve config for scoring mode and thresholds; fall back to defaults
	scoringMode := enums.VendorScoringModeAnsweredOnly
	thresholds := defaultThresholds

	configID := resolveConfigIDFromScores(scores)
	if configID != "" {
		cfg, err := client.VendorScoringConfig.Get(ctx, configID)
		if err != nil && !generated.IsNotFound(err) {
			return err
		}

		if cfg != nil {
			scoringMode = cfg.ScoringMode
			thresholds = cfg.RiskThresholds
		}
	}

	// MANUAL mode skips all entity-level aggregation so user-supplied values are preserved
	if scoringMode == enums.VendorScoringModeManual {
		return nil
	}

	var total float64
	for _, s := range scores {
		total += s.Score
	}

	coverage := len(lo.Filter(scores, func(score *generated.VendorRiskScore, _ int) bool {
		return hasAnswerValue(score.Answer)
	}))

	// Under FULL_QUESTIONNAIRE mode, add penalty scores for unanswered questions
	// that exist in the config but have no VendorRiskScore record with an answer
	if scoringMode == enums.VendorScoringModeFullQuestionnaire && configID != "" {
		penalty, err := computeUnansweredPenalty(ctx, client, configID, scores)
		if err != nil {
			return err
		}

		total += penalty
	}

	return client.Entity.UpdateOneID(entityID).
		SetRiskScore(int(math.Round(total))).
		SetRiskRating(thresholds.Resolve(total)).
		SetRiskScoreCoverage(coverage).
		Exec(ctx)
}

// computeUnansweredPenalty calculates the total penalty for questions in the config
// that have no answered VendorRiskScore record
func computeUnansweredPenalty(ctx context.Context, client *generated.Client, configID string, scores []*generated.VendorRiskScore) (float64, error) {
	cfg, err := client.VendorScoringConfig.Get(ctx, configID)
	if err != nil {
		return 0, err
	}

	answeredKeys := make(map[string]bool, len(scores))
	for _, s := range scores {
		if hasAnswerValue(s.Answer) {
			answeredKeys[s.QuestionKey] = true
		}
	}

	var penalty float64

	for _, q := range cfg.Questions.All() {
		if !q.Enabled {
			continue
		}

		if answeredKeys[q.Key] {
			continue
		}

		penalty += maxPenaltyScore
	}

	return penalty, nil
}

// resolveConfigIDFromScores returns the first non-empty VendorScoringConfigID from a score set
func resolveConfigIDFromScores(scores []*generated.VendorRiskScore) string {
	for _, s := range scores {
		if s.VendorScoringConfigID != "" {
			return s.VendorScoringConfigID
		}
	}

	return ""
}

// computeOnCreate populates denormalized question fields, resolves the scoring config,
// and computes the initial score for a new VendorRiskScore record
func computeOnCreate(ctx context.Context, m *generated.VendorRiskScoreMutation) error {
	questionKey, ok := m.QuestionKey()
	if !ok {
		return nil
	}

	allowCtx := privacy.DecisionContext(ctx, privacy.Allow)

	// Resolve vendor_scoring_config_id from org context when not explicitly provided
	configID, hasConfigID := m.VendorScoringConfigID()
	if !hasConfigID {
		ownerID, ok := m.OwnerID()
		if ok {
			cfg, err := m.Client().VendorScoringConfig.Query().
				Where(vendorscoringconfig.OwnerID(ownerID)).
				Only(allowCtx)
			if err != nil && !generated.IsNotFound(err) {
				return err
			}

			if cfg != nil {
				configID = cfg.ID
				m.SetVendorScoringConfigID(configID)
			}
		}
	}

	def, err := resolveQuestionDef(allowCtx, m, configID, questionKey)
	if err != nil {
		return err
	}

	if def == nil {
		return ErrVendorScoringQuestionNotFound
	}

	// Populate denormalized fields only if not already provided by the caller
	if _, ok := m.QuestionName(); !ok {
		m.SetQuestionName(def.Name)
	}

	if _, ok := m.QuestionDescription(); !ok && def.Description != "" {
		m.SetQuestionDescription(def.Description)
	}

	if _, ok := m.QuestionCategory(); !ok {
		m.SetQuestionCategory(def.Category)
	}

	if _, ok := m.AnswerType(); !ok {
		m.SetAnswerType(def.AnswerType)
	}

	// Default impact from the question definition when not provided
	if _, ok := m.Impact(); !ok {
		m.SetImpact(def.SuggestedImpact)
	}

	impact, _ := m.Impact()

	likelihood, ok := m.Likelihood()
	if !ok {
		// Likelihood is required by the schema validator; defer to it
		return nil
	}

	answer, hasAnswer := currentAnswer(m)
	answerType, _ := m.AnswerType()

	m.SetScore(computeScore(impact, likelihood, hasAnswer, answer, answerType))

	return nil
}

// recomputeScore recalculates the score on update, fetching unchanged fields from the stored record
func recomputeScore(ctx context.Context, m *generated.VendorRiskScoreMutation) error {
	impact, err := currentImpact(ctx, m)
	if err != nil {
		return err
	}

	likelihood, err := currentLikelihood(ctx, m)
	if err != nil {
		return err
	}

	answer, hasAnswer, err := currentOrOldAnswer(ctx, m)
	if err != nil {
		return err
	}

	answerType, err := currentAnswerType(ctx, m)
	if err != nil {
		return err
	}

	m.SetScore(computeScore(impact, likelihood, hasAnswer, answer, answerType))

	return nil
}

// resolveQuestionDef looks up the VendorScoringQuestionDef for a given question key,
// merging org-custom overrides with system defaults via VendorScoringQuestionsConfig.All
func resolveQuestionDef(ctx context.Context, m *generated.VendorRiskScoreMutation, configID, questionKey string) (*models.VendorScoringQuestionDef, error) {
	var questions models.VendorScoringQuestionsConfig

	if configID != "" {
		cfg, err := m.Client().VendorScoringConfig.Get(ctx, configID)
		if err != nil && !generated.IsNotFound(err) {
			return nil, err
		}

		if cfg != nil {
			questions = cfg.Questions
		}
	}

	def, found := lo.Find(questions.All(), func(d models.VendorScoringQuestionDef) bool {
		return d.Key == questionKey
	})
	if !found {
		return nil, nil
	}

	return &def, nil
}

func hasScoringFieldUpdate(m *generated.VendorRiskScoreMutation) bool {
	_, hasImpact := m.Impact()
	_, hasLikelihood := m.Likelihood()
	_, hasAnswer := m.Answer()
	_, hasAnswerType := m.AnswerType()

	return hasImpact || hasLikelihood || hasAnswer || m.AnswerCleared() || hasAnswerType
}

func aggregateEntityID(ctx context.Context, m *generated.VendorRiskScoreMutation, preDeleteEntityID string) (string, error) {
	if preDeleteEntityID != "" {
		return preDeleteEntityID, nil
	}

	if entityID, ok := m.EntityID(); ok {
		return entityID, nil
	}

	if m.Op().Is(ent.OpUpdateOne) {
		return m.OldEntityID(ctx)
	}

	return "", nil
}

func currentImpact(ctx context.Context, m *generated.VendorRiskScoreMutation) (enums.VendorRiskImpact, error) {
	if impact, ok := m.Impact(); ok {
		return impact, nil
	}

	return m.OldImpact(ctx)
}

func currentLikelihood(ctx context.Context, m *generated.VendorRiskScoreMutation) (enums.VendorRiskLikelihood, error) {
	if likelihood, ok := m.Likelihood(); ok {
		return likelihood, nil
	}

	return m.OldLikelihood(ctx)
}

func currentAnswerType(ctx context.Context, m *generated.VendorRiskScoreMutation) (enums.VendorScoringAnswerType, error) {
	if answerType, ok := m.AnswerType(); ok {
		return answerType, nil
	}

	return m.OldAnswerType(ctx)
}

func currentAnswer(m *generated.VendorRiskScoreMutation) (string, bool) {
	answer, ok := m.Answer()
	if !ok {
		return "", false
	}

	answer = strings.TrimSpace(answer)

	return answer, answer != ""
}

func currentOrOldAnswer(ctx context.Context, m *generated.VendorRiskScoreMutation) (string, bool, error) {
	if answer, ok := currentAnswer(m); ok {
		return answer, true, nil
	}

	if m.AnswerCleared() {
		return "", false, nil
	}

	old, err := m.OldAnswer(ctx)
	if err != nil {
		return "", false, err
	}
	if !hasAnswerValue(old) {
		return "", false, nil
	}

	return strings.TrimSpace(*old), true, nil
}

func hasAnswerValue(answer *string) bool {
	return answer != nil && strings.TrimSpace(*answer) != ""
}

// computeScore returns zero for unanswered questions and for boolean questions
// where answer == "true" (control present, no gap); all other answered cases
// contribute impactNumeric * likelihoodNumeric
func computeScore(impact enums.VendorRiskImpact, likelihood enums.VendorRiskLikelihood, hasAnswer bool, answer string, answerType enums.VendorScoringAnswerType) float64 {
	if !hasAnswer {
		return 0
	}

	if answerType == enums.VendorScoringAnswerTypeBoolean && strings.EqualFold(answer, "true") {
		return 0
	}

	return impactToFloat(impact) * likelihoodToFloat(likelihood)
}

func impactToFloat(v enums.VendorRiskImpact) float64 {
	const (
		impactVeryLow  = 1.0
		impactLow      = 2.0
		impactMedium   = 3.0
		impactHigh     = 4.0
		impactCritical = 5.0
	)

	switch v {
	case enums.VendorRiskImpactVeryLow:
		return impactVeryLow
	case enums.VendorRiskImpactLow:
		return impactLow
	case enums.VendorRiskImpactMedium:
		return impactMedium
	case enums.VendorRiskImpactHigh:
		return impactHigh
	case enums.VendorRiskImpactCritical:
		return impactCritical
	default:
		return 0
	}
}

func likelihoodToFloat(v enums.VendorRiskLikelihood) float64 {
	const (
		likelihoodVeryLow  = 0.5
		likelihoodLow      = 1.0
		likelihoodMedium   = 2.0
		likelihoodHigh     = 3.0
		likelihoodVeryHigh = 4.0
	)

	switch v {
	case enums.VendorRiskLikelihoodVeryLow:
		return likelihoodVeryLow
	case enums.VendorRiskLikelihoodLow:
		return likelihoodLow
	case enums.VendorRiskLikelihoodMedium:
		return likelihoodMedium
	case enums.VendorRiskLikelihoodHigh:
		return likelihoodHigh
	case enums.VendorRiskLikelihoodVeryHigh:
		return likelihoodVeryHigh
	default:
		return 0
	}
}
