//go:build test

package graphapi_test

import (
	"context"
	"testing"

	"gotest.tools/v3/assert"
	is "gotest.tools/v3/assert/cmp"

	"github.com/theopenlane/core/common/enums"
	"github.com/theopenlane/core/common/models"
	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/hooks"
	"github.com/theopenlane/core/internal/graphapi"
	"github.com/theopenlane/core/internal/graphapi/testclient"
)

const (
	answeredQuestionScore = 6
	unansweredPenalty     = 20
	sentinelRiskScore     = 9999
	sentinelRiskRating    = "SENTINEL"
)

func TestVendorScoringConfigListenerRecompute(t *testing.T) {
	scoringUser := suite.userBuilder(context.Background(), t)
	ctx := setContext(scoringUser.UserCtx, suite.client.db)

	setup, err := graphapi.SetupListenerRuntime(ctx, suite.client.db, suite.tf.URI, hooks.VendorScoringListeners())
	assert.NilError(t, err)
	defer setup.Teardown()

	entity := (&EntityBuilder{client: suite.client, Tier: enums.VendorTierStandard}).MustNew(scoringUser.UserCtx, t)

	configResp, err := suite.client.api.CreateVendorScoringConfig(scoringUser.UserCtx, testclient.CreateVendorScoringConfigInput{
		OwnerID: &scoringUser.OrganizationID,
	})
	assert.NilError(t, err)

	configID := configResp.CreateVendorScoringConfig.VendorScoringConfig.ID

	question := mustVendorQuestion(t, "IAM-05.1")
	falseAnswer := "false"

	_, err = suite.client.api.CreateVendorRiskScore(scoringUser.UserCtx, testclient.CreateVendorRiskScoreInput{
		OwnerID:               &scoringUser.OrganizationID,
		VendorScoringConfigID: &configID,
		EntityID:              entity.ID,
		QuestionKey:           question.Key,
		QuestionName:          question.Name,
		QuestionCategory:      question.Category,
		Impact:                enums.VendorRiskImpactMedium,
		Likelihood:            enums.VendorRiskLikelihoodMedium,
		Answer:                &falseAnswer,
	})
	assert.NilError(t, err)

	baseline, err := suite.client.db.Entity.Get(ctx, entity.ID)
	assert.NilError(t, err)
	assert.Check(t, is.Equal(answeredQuestionScore, baseline.RiskScore))
	assert.Check(t, is.Equal("MEDIUM", baseline.RiskRating))
	assert.Check(t, is.Equal(1, baseline.RiskScoreCoverage))

	t.Run("update without scoring_mode or risk_thresholds does not recompute", func(t *testing.T) {
		err := suite.client.db.Entity.UpdateOneID(entity.ID).
			SetRiskScore(sentinelRiskScore).
			SetRiskRating(sentinelRiskRating).
			Exec(ctx)
		assert.NilError(t, err)

		// disabled so it never contributes to the FULL_QUESTIONNAIRE penalty arithmetic later
		_, err = suite.client.api.UpdateVendorScoringConfig(scoringUser.UserCtx, configID, testclient.UpdateVendorScoringConfigInput{
			Questions: &models.VendorScoringQuestionsConfig{
				Custom: []models.VendorScoringQuestionDef{{
					Name:            "Disabled custom probe",
					Category:        enums.VendorScoringCategorySecurityPractices,
					AnswerType:      enums.VendorScoringAnswerTypeBoolean,
					SuggestedImpact: enums.VendorRiskImpactLow,
					Enabled:         false,
				}},
			},
		})
		assert.NilError(t, err)

		setup.Runtime.WaitIdle()

		unchanged, err := suite.client.db.Entity.Get(ctx, entity.ID)
		assert.NilError(t, err)
		assert.Check(t, is.Equal(sentinelRiskScore, unchanged.RiskScore))
		assert.Check(t, is.Equal(sentinelRiskRating, unchanged.RiskRating))
	})

	t.Run("risk_thresholds update recomputes every scored entity", func(t *testing.T) {
		widened := models.RiskThresholdsConfig{
			Custom: []models.RiskThreshold{
				{Rating: enums.VendorRiskRatingLow, MaxScore: 8},
			},
		}

		_, err := suite.client.api.UpdateVendorScoringConfig(scoringUser.UserCtx, configID, testclient.UpdateVendorScoringConfigInput{
			RiskThresholds: &widened,
		})
		assert.NilError(t, err)

		recomputed, err := listenerPoll(func() (*generated.Entity, error) {
			return suite.client.db.Entity.Get(ctx, entity.ID)
		}, func(e *generated.Entity) bool {
			return e.RiskRating == "LOW"
		})
		assert.NilError(t, err)
		assert.Check(t, is.Equal(answeredQuestionScore, recomputed.RiskScore))
		assert.Check(t, is.Equal(1, recomputed.RiskScoreCoverage))
	})

	t.Run("scoring_mode update recomputes with unanswered penalties", func(t *testing.T) {
		fullMode := enums.VendorScoringModeFullQuestionnaire

		_, err := suite.client.api.UpdateVendorScoringConfig(scoringUser.UserCtx, configID, testclient.UpdateVendorScoringConfigInput{
			ScoringMode: &fullMode,
		})
		assert.NilError(t, err)

		expected := answeredQuestionScore + (len(models.DefaultVendorScoringQuestions)-1)*unansweredPenalty

		recomputed, err := listenerPoll(func() (*generated.Entity, error) {
			return suite.client.db.Entity.Get(ctx, entity.ID)
		}, func(e *generated.Entity) bool {
			return e.RiskScore == expected
		})
		assert.NilError(t, err)
		assert.Check(t, is.Equal(expected, recomputed.RiskScore))
		assert.Check(t, is.Equal(1, recomputed.RiskScoreCoverage))
	})

	cleanupOrganizationDataWithContext(scoringUser.UserCtx, t)
}
