package graphapi_test

import (
	"context"
	"testing"

	"github.com/samber/lo"
	"gotest.tools/v3/assert"
	is "gotest.tools/v3/assert/cmp"

	"github.com/theopenlane/core/common/enums"
	"github.com/theopenlane/core/common/models"
	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/graphapi/testclient"
)

func TestVendorScoringEntityManualRiskFieldsPersistWithoutScores(t *testing.T) {
	scoringUser := suite.userBuilder(context.Background(), t)
	entity := newVendorScoringEntity(scoringUser.UserCtx, t)

	manualRiskScore := int64(17)
	manualRiskRating := "HIGH"

	_, err := suite.client.api.UpdateEntity(scoringUser.UserCtx, entity.ID, testclient.UpdateEntityInput{
		RiskScore:  &manualRiskScore,
		RiskRating: &manualRiskRating,
	})
	assert.NilError(t, err)

	configResp, err := suite.client.api.CreateVendorScoringConfig(scoringUser.UserCtx, testclient.CreateVendorScoringConfigInput{
		OwnerID: &scoringUser.OrganizationID,
	})
	assert.NilError(t, err)
	assert.Assert(t, configResp != nil)

	entityResp, err := suite.client.api.GetEntityByID(scoringUser.UserCtx, entity.ID)
	assert.NilError(t, err)
	assert.Assert(t, entityResp != nil)
	assert.Assert(t, entityResp.Entity.RiskScore != nil)
	assert.Assert(t, entityResp.Entity.RiskRating != nil)
	assert.Check(t, is.Equal(manualRiskScore, *entityResp.Entity.RiskScore))
	assert.Check(t, is.Equal(manualRiskRating, *entityResp.Entity.RiskRating))

	(&Cleanup[*generated.VendorScoringConfigDeleteOne]{
		client: suite.client.db.VendorScoringConfig,
		ID:     configResp.CreateVendorScoringConfig.VendorScoringConfig.ID,
	}).MustDelete(scoringUser.UserCtx, t)
	(&Cleanup[*generated.EntityDeleteOne]{client: suite.client.db.Entity, ID: entity.ID}).MustDelete(scoringUser.UserCtx, t)
	(&Cleanup[*generated.EntityTypeDeleteOne]{client: suite.client.db.EntityType, ID: entity.EntityTypeID}).MustDelete(scoringUser.UserCtx, t)
}

func TestVendorScoringEntityManualRiskFieldsOverriddenByScores(t *testing.T) {
	scoringUser := suite.userBuilder(context.Background(), t)
	entity := newVendorScoringEntity(scoringUser.UserCtx, t)

	// Set manual risk fields
	manualRiskScore := int64(17)
	manualRiskRating := "HIGH"

	_, err := suite.client.api.UpdateEntity(scoringUser.UserCtx, entity.ID, testclient.UpdateEntityInput{
		RiskScore:  &manualRiskScore,
		RiskRating: &manualRiskRating,
	})
	assert.NilError(t, err)

	assertEntityRiskState(t, scoringUser.UserCtx, entity.ID, 17, "HIGH", 0)

	// Submit a scored question — should overwrite manual fields
	question := mustVendorQuestion(t, "IAM-05.1")
	falseAnswer := "false"

	scoreResp, err := suite.client.api.CreateVendorRiskScore(scoringUser.UserCtx, testclient.CreateVendorRiskScoreInput{
		OwnerID:          &scoringUser.OrganizationID,
		EntityID:         entity.ID,
		QuestionKey:      question.Key,
		QuestionName:     question.Name,
		QuestionCategory: question.Category,
		Impact:           enums.VendorRiskImpactMedium,
		Likelihood:       enums.VendorRiskLikelihoodMedium,
		Answer:           &falseAnswer,
	})
	assert.NilError(t, err)

	// MEDIUM impact (3) * MEDIUM likelihood (2) = 6
	assert.Check(t, is.Equal(6.0, scoreResp.CreateVendorRiskScore.VendorRiskScore.Score))
	assertEntityRiskState(t, scoringUser.UserCtx, entity.ID, 6, "MEDIUM", 1)

	(&Cleanup[*generated.VendorRiskScoreDeleteOne]{client: suite.client.db.VendorRiskScore, ID: scoreResp.CreateVendorRiskScore.VendorRiskScore.ID}).MustDelete(scoringUser.UserCtx, t)
	(&Cleanup[*generated.EntityDeleteOne]{client: suite.client.db.Entity, ID: entity.ID}).MustDelete(scoringUser.UserCtx, t)
	(&Cleanup[*generated.EntityTypeDeleteOne]{client: suite.client.db.EntityType, ID: entity.EntityTypeID}).MustDelete(scoringUser.UserCtx, t)
}

func TestVendorScoringConfigCustomQuestionRoundTrip(t *testing.T) {
	scoringUser := suite.userBuilder(context.Background(), t)
	entity := newVendorScoringEntity(scoringUser.UserCtx, t)

	customQuestion := models.VendorScoringQuestionDef{
		Key:             "CUSTOM-SOC2-001",
		Name:            "Does the vendor maintain a current SOC 2 Type II report?",
		Description:     "Custom org-defined question used to validate config-backed scoring inputs.",
		Category:        enums.VendorScoringCategoryRegulatoryCompliance,
		AnswerType:      enums.VendorScoringAnswerTypeBoolean,
		SuggestedImpact: enums.VendorRiskImpactHigh,
		Enabled:         true,
	}

	configResp, err := suite.client.api.CreateVendorScoringConfig(scoringUser.UserCtx, testclient.CreateVendorScoringConfigInput{
		OwnerID: &scoringUser.OrganizationID,
		Questions: &models.VendorScoringQuestionsConfig{
			Custom: []models.VendorScoringQuestionDef{customQuestion},
		},
	})
	assert.NilError(t, err)
	assert.Assert(t, configResp != nil)

	configID := configResp.CreateVendorScoringConfig.VendorScoringConfig.ID

	fetchedConfigResp, err := suite.client.api.GetVendorScoringConfigByID(scoringUser.UserCtx, configID)
	assert.NilError(t, err)
	assert.Assert(t, fetchedConfigResp != nil)

	fetchedQuestion, found := lo.Find(fetchedConfigResp.VendorScoringConfig.Questions.Custom, func(question models.VendorScoringQuestionDef) bool {
		return question.Name == customQuestion.Name
	})
	assert.Assert(t, found)
	assert.Check(t, is.Equal(customQuestion.Name, fetchedQuestion.Name))
	assert.Check(t, is.Equal(customQuestion.Description, fetchedQuestion.Description))
	assert.Check(t, is.Equal(customQuestion.Category, fetchedQuestion.Category))
	assert.Check(t, is.Equal(customQuestion.AnswerType, fetchedQuestion.AnswerType))
	assert.Check(t, is.Equal(customQuestion.SuggestedImpact, fetchedQuestion.SuggestedImpact))

	answer := "false"
	scoreResp, err := suite.client.api.CreateVendorRiskScore(scoringUser.UserCtx, testclient.CreateVendorRiskScoreInput{
		OwnerID:               &scoringUser.OrganizationID,
		VendorScoringConfigID: &configID,
		EntityID:              entity.ID,
		QuestionKey:           fetchedQuestion.Key,
		QuestionName:          fetchedQuestion.Name,
		QuestionCategory:      fetchedQuestion.Category,
		Impact:                fetchedQuestion.SuggestedImpact,
		Likelihood:            enums.VendorRiskLikelihoodHigh,
		Answer:                &answer,
	})
	assert.NilError(t, err)
	assert.Assert(t, scoreResp != nil)
	assert.Assert(t, scoreResp.CreateVendorRiskScore.VendorRiskScore.QuestionDescription != nil)
	assert.Check(t, is.Equal(customQuestion.Description, *scoreResp.CreateVendorRiskScore.VendorRiskScore.QuestionDescription))
	assert.Check(t, is.Equal(customQuestion.AnswerType, scoreResp.CreateVendorRiskScore.VendorRiskScore.AnswerType))
	assert.Assert(t, scoreResp.CreateVendorRiskScore.VendorRiskScore.VendorScoringConfigID != nil)
	assert.Check(t, is.Equal(configID, *scoreResp.CreateVendorRiskScore.VendorRiskScore.VendorScoringConfigID))

	// HIGH impact (4) * HIGH likelihood (3) = 12
	assert.Check(t, is.Equal(12.0, scoreResp.CreateVendorRiskScore.VendorRiskScore.Score))

	scoreID := scoreResp.CreateVendorRiskScore.VendorRiskScore.ID

	fetchedScoreResp, err := suite.client.api.GetVendorRiskScoreByID(scoringUser.UserCtx, scoreID)
	assert.NilError(t, err)
	assert.Assert(t, fetchedScoreResp != nil)
	assert.Assert(t, fetchedScoreResp.VendorRiskScore.QuestionDescription != nil)
	assert.Check(t, is.Equal(customQuestion.Description, *fetchedScoreResp.VendorRiskScore.QuestionDescription))
	assert.Check(t, is.Equal(customQuestion.AnswerType, fetchedScoreResp.VendorRiskScore.AnswerType))
	assert.Assert(t, fetchedScoreResp.VendorRiskScore.VendorScoringConfigID != nil)
	assert.Check(t, is.Equal(configID, *fetchedScoreResp.VendorRiskScore.VendorScoringConfigID))

	assertEntityRiskState(t, scoringUser.UserCtx, entity.ID, 12, "HIGH", 1)

	(&Cleanup[*generated.VendorRiskScoreDeleteOne]{client: suite.client.db.VendorRiskScore, ID: scoreID}).MustDelete(scoringUser.UserCtx, t)
	(&Cleanup[*generated.VendorScoringConfigDeleteOne]{client: suite.client.db.VendorScoringConfig, ID: configID}).MustDelete(scoringUser.UserCtx, t)
	(&Cleanup[*generated.EntityDeleteOne]{client: suite.client.db.Entity, ID: entity.ID}).MustDelete(scoringUser.UserCtx, t)
	(&Cleanup[*generated.EntityTypeDeleteOne]{client: suite.client.db.EntityType, ID: entity.EntityTypeID}).MustDelete(scoringUser.UserCtx, t)
}

func TestVendorScoringCustomQuestionOverridesDefaultKey(t *testing.T) {
	scoringUser := suite.userBuilder(context.Background(), t)
	entity := newVendorScoringEntity(scoringUser.UserCtx, t)

	// Override the default IAM-05.1 question with custom wording and different suggested impact
	defaultQuestion := mustVendorQuestion(t, "IAM-05.1")
	overriddenQuestion := models.VendorScoringQuestionDef{
		Key:             defaultQuestion.Key,
		Name:            "Custom override: Does the vendor enforce least privilege?",
		Description:     "Org-specific rewording of the default IAM-05.1 question.",
		Category:        defaultQuestion.Category,
		AnswerType:      defaultQuestion.AnswerType,
		SuggestedImpact: enums.VendorRiskImpactCritical,
		Enabled:         true,
	}

	configResp, err := suite.client.api.CreateVendorScoringConfig(scoringUser.UserCtx, testclient.CreateVendorScoringConfigInput{
		OwnerID: &scoringUser.OrganizationID,
		Questions: &models.VendorScoringQuestionsConfig{
			Custom: []models.VendorScoringQuestionDef{overriddenQuestion},
		},
	})
	assert.NilError(t, err)

	configID := configResp.CreateVendorScoringConfig.VendorScoringConfig.ID

	// Submit a score referencing the overridden key — hook should resolve custom definition
	falseAnswer := "false"
	scoreResp, err := suite.client.api.CreateVendorRiskScore(scoringUser.UserCtx, testclient.CreateVendorRiskScoreInput{
		OwnerID:               &scoringUser.OrganizationID,
		VendorScoringConfigID: &configID,
		EntityID:              entity.ID,
		QuestionKey:           overriddenQuestion.Key,
		QuestionName:          overriddenQuestion.Name,
		QuestionCategory:      overriddenQuestion.Category,
		Impact:                overriddenQuestion.SuggestedImpact,
		Likelihood:            enums.VendorRiskLikelihoodHigh,
		Answer:                &falseAnswer,
	})
	assert.NilError(t, err)

	// CRITICAL impact (5) * HIGH likelihood (3) = 15
	assert.Check(t, is.Equal(15.0, scoreResp.CreateVendorRiskScore.VendorRiskScore.Score))

	// Verify the hook resolved the custom description, not the default
	assert.Assert(t, scoreResp.CreateVendorRiskScore.VendorRiskScore.QuestionDescription != nil)
	assert.Check(t, is.Equal(overriddenQuestion.Description, *scoreResp.CreateVendorRiskScore.VendorRiskScore.QuestionDescription))

	assertEntityRiskState(t, scoringUser.UserCtx, entity.ID, 15, "HIGH", 1)

	(&Cleanup[*generated.VendorRiskScoreDeleteOne]{client: suite.client.db.VendorRiskScore, ID: scoreResp.CreateVendorRiskScore.VendorRiskScore.ID}).MustDelete(scoringUser.UserCtx, t)
	(&Cleanup[*generated.VendorScoringConfigDeleteOne]{client: suite.client.db.VendorScoringConfig, ID: configID}).MustDelete(scoringUser.UserCtx, t)
	(&Cleanup[*generated.EntityDeleteOne]{client: suite.client.db.Entity, ID: entity.ID}).MustDelete(scoringUser.UserCtx, t)
	(&Cleanup[*generated.EntityTypeDeleteOne]{client: suite.client.db.EntityType, ID: entity.EntityTypeID}).MustDelete(scoringUser.UserCtx, t)
}

func TestVendorRiskScoreComputedValues(t *testing.T) {
	scoringUser := suite.userBuilder(context.Background(), t)
	entity := newVendorScoringEntity(scoringUser.UserCtx, t)

	question := mustVendorQuestion(t, "CEK-03.1")
	falseAnswer := "false"
	trueAnswer := "true"

	// Submit with answer "false" (boolean, control gap) — should score impact * likelihood
	scoreResp, err := suite.client.api.CreateVendorRiskScore(scoringUser.UserCtx, testclient.CreateVendorRiskScoreInput{
		OwnerID:          &scoringUser.OrganizationID,
		EntityID:         entity.ID,
		QuestionKey:      question.Key,
		QuestionName:     question.Name,
		QuestionCategory: question.Category,
		Impact:           enums.VendorRiskImpactCritical,
		Likelihood:       enums.VendorRiskLikelihoodVeryHigh,
		Answer:           &falseAnswer,
	})
	assert.NilError(t, err)

	// CRITICAL (5) * VERY_HIGH (4) = 20
	assert.Check(t, is.Equal(20.0, scoreResp.CreateVendorRiskScore.VendorRiskScore.Score))
	assertEntityRiskState(t, scoringUser.UserCtx, entity.ID, 20, "CRITICAL", 1)

	scoreID := scoreResp.CreateVendorRiskScore.VendorRiskScore.ID

	// Update answer to "true" (boolean, control present) — score drops to 0
	updatedResp, err := suite.client.api.UpdateVendorRiskScore(scoringUser.UserCtx, scoreID, testclient.UpdateVendorRiskScoreInput{
		Answer: &trueAnswer,
	})
	assert.NilError(t, err)
	assert.Check(t, is.Equal(0.0, updatedResp.UpdateVendorRiskScore.VendorRiskScore.Score))
	assertEntityRiskState(t, scoringUser.UserCtx, entity.ID, 0, "NONE", 1)

	// Update impact — should recompute with new impact but answer "true" still zeroes it
	updatedImpactResp, err := suite.client.api.UpdateVendorRiskScore(scoringUser.UserCtx, scoreID, testclient.UpdateVendorRiskScoreInput{
		Impact: lo.ToPtr(enums.VendorRiskImpactLow),
	})
	assert.NilError(t, err)
	assert.Check(t, is.Equal(0.0, updatedImpactResp.UpdateVendorRiskScore.VendorRiskScore.Score))

	// Flip answer back to "false" — now LOW (2) * VERY_HIGH (4) = 8
	updatedBackResp, err := suite.client.api.UpdateVendorRiskScore(scoringUser.UserCtx, scoreID, testclient.UpdateVendorRiskScoreInput{
		Answer: &falseAnswer,
	})
	assert.NilError(t, err)
	assert.Check(t, is.Equal(8.0, updatedBackResp.UpdateVendorRiskScore.VendorRiskScore.Score))
	assertEntityRiskState(t, scoringUser.UserCtx, entity.ID, 8, "MEDIUM", 1)

	// Clear answer — score drops to 0, coverage drops to 0
	clearedResp, err := suite.client.api.UpdateVendorRiskScore(scoringUser.UserCtx, scoreID, testclient.UpdateVendorRiskScoreInput{
		ClearAnswer: lo.ToPtr(true),
	})
	assert.NilError(t, err)
	assert.Check(t, is.Equal(0.0, clearedResp.UpdateVendorRiskScore.VendorRiskScore.Score))
	assert.Assert(t, clearedResp.UpdateVendorRiskScore.VendorRiskScore.Answer == nil)
	assertEntityRiskState(t, scoringUser.UserCtx, entity.ID, 0, "NONE", 0)

	(&Cleanup[*generated.VendorRiskScoreDeleteOne]{client: suite.client.db.VendorRiskScore, ID: scoreID}).MustDelete(scoringUser.UserCtx, t)
	(&Cleanup[*generated.EntityDeleteOne]{client: suite.client.db.Entity, ID: entity.ID}).MustDelete(scoringUser.UserCtx, t)
	(&Cleanup[*generated.EntityTypeDeleteOne]{client: suite.client.db.EntityType, ID: entity.EntityTypeID}).MustDelete(scoringUser.UserCtx, t)
}

func TestVendorRiskScorePartialSubmissions(t *testing.T) {
	scoringUser := suite.userBuilder(context.Background(), t)
	entity := newVendorScoringEntity(scoringUser.UserCtx, t)

	questionOne := mustVendorQuestion(t, "IAM-05.1")
	questionTwo := mustVendorQuestion(t, "DSP-16.1")

	falseAnswer := "false"

	firstScoreResp, err := suite.client.api.CreateVendorRiskScore(scoringUser.UserCtx, testclient.CreateVendorRiskScoreInput{
		OwnerID:          &scoringUser.OrganizationID,
		EntityID:         entity.ID,
		QuestionKey:      questionOne.Key,
		QuestionName:     questionOne.Name,
		QuestionCategory: questionOne.Category,
		Impact:           enums.VendorRiskImpactMedium,
		Likelihood:       enums.VendorRiskLikelihoodHigh,
		Answer:           &falseAnswer,
	})
	assert.NilError(t, err)
	assert.Assert(t, firstScoreResp != nil)
	assert.Check(t, is.Equal(9.0, firstScoreResp.CreateVendorRiskScore.VendorRiskScore.Score))

	secondScoreResp, err := suite.client.api.CreateVendorRiskScore(scoringUser.UserCtx, testclient.CreateVendorRiskScoreInput{
		OwnerID:          &scoringUser.OrganizationID,
		EntityID:         entity.ID,
		QuestionKey:      questionTwo.Key,
		QuestionName:     questionTwo.Name,
		QuestionCategory: questionTwo.Category,
		Impact:           enums.VendorRiskImpactMedium,
		Likelihood:       enums.VendorRiskLikelihoodHigh,
	})
	assert.NilError(t, err)
	assert.Assert(t, secondScoreResp != nil)
	assert.Check(t, is.Equal(0.0, secondScoreResp.CreateVendorRiskScore.VendorRiskScore.Score))

	assertEntityRiskState(t, scoringUser.UserCtx, entity.ID, 9, "MEDIUM", 1)

	firstScoreID := firstScoreResp.CreateVendorRiskScore.VendorRiskScore.ID
	secondScoreID := secondScoreResp.CreateVendorRiskScore.VendorRiskScore.ID

	updatedFirstScoreResp, err := suite.client.api.UpdateVendorRiskScore(scoringUser.UserCtx, firstScoreID, testclient.UpdateVendorRiskScoreInput{
		ClearAnswer: lo.ToPtr(true),
	})
	assert.NilError(t, err)
	assert.Assert(t, updatedFirstScoreResp != nil)
	assert.Check(t, is.Equal(0.0, updatedFirstScoreResp.UpdateVendorRiskScore.VendorRiskScore.Score))
	assert.Assert(t, updatedFirstScoreResp.UpdateVendorRiskScore.VendorRiskScore.Answer == nil)

	assertEntityRiskState(t, scoringUser.UserCtx, entity.ID, 0, "NONE", 0)

	updatedSecondScoreResp, err := suite.client.api.UpdateVendorRiskScore(scoringUser.UserCtx, secondScoreID, testclient.UpdateVendorRiskScoreInput{
		Answer: &falseAnswer,
	})
	assert.NilError(t, err)
	assert.Assert(t, updatedSecondScoreResp != nil)
	assert.Check(t, is.Equal(9.0, updatedSecondScoreResp.UpdateVendorRiskScore.VendorRiskScore.Score))
	assert.Assert(t, updatedSecondScoreResp.UpdateVendorRiskScore.VendorRiskScore.Answer != nil)
	assert.Check(t, is.Equal(falseAnswer, *updatedSecondScoreResp.UpdateVendorRiskScore.VendorRiskScore.Answer))

	assertEntityRiskState(t, scoringUser.UserCtx, entity.ID, 9, "MEDIUM", 1)

	(&Cleanup[*generated.VendorRiskScoreDeleteOne]{client: suite.client.db.VendorRiskScore, IDs: []string{firstScoreID, secondScoreID}}).MustDelete(scoringUser.UserCtx, t)
	(&Cleanup[*generated.EntityDeleteOne]{client: suite.client.db.Entity, ID: entity.ID}).MustDelete(scoringUser.UserCtx, t)
	(&Cleanup[*generated.EntityTypeDeleteOne]{client: suite.client.db.EntityType, ID: entity.EntityTypeID}).MustDelete(scoringUser.UserCtx, t)
}

func TestVendorRiskScoreAllDefaultQuestions(t *testing.T) {
	scoringUser := suite.userBuilder(context.Background(), t)
	entity := newVendorScoringEntity(scoringUser.UserCtx, t)

	falseAnswer := "false"
	var scoreIDs []string

	var expectedTotal float64

	for _, question := range models.DefaultVendorScoringQuestions {
		resp, err := suite.client.api.CreateVendorRiskScore(scoringUser.UserCtx, testclient.CreateVendorRiskScoreInput{
			OwnerID:          &scoringUser.OrganizationID,
			EntityID:         entity.ID,
			QuestionKey:      question.Key,
			QuestionName:     question.Name,
			QuestionCategory: question.Category,
			Impact:           question.SuggestedImpact,
			Likelihood:       enums.VendorRiskLikelihoodMedium,
			Answer:           &falseAnswer,
		})
		assert.NilError(t, err, "question %s", question.Key)
		assert.Assert(t, resp.CreateVendorRiskScore.VendorRiskScore.Score > 0, "question %s should have non-zero score", question.Key)

		expectedTotal += resp.CreateVendorRiskScore.VendorRiskScore.Score
		scoreIDs = append(scoreIDs, resp.CreateVendorRiskScore.VendorRiskScore.ID)
	}

	// Verify entity aggregation across all default questions
	entityResp, err := suite.client.api.GetEntityByID(scoringUser.UserCtx, entity.ID)
	assert.NilError(t, err)
	assert.Assert(t, entityResp.Entity.RiskScore != nil)
	assert.Assert(t, entityResp.Entity.RiskScoreCoverage != nil)
	assert.Check(t, is.Equal(int64(len(models.DefaultVendorScoringQuestions)), *entityResp.Entity.RiskScoreCoverage))
	assert.Check(t, is.Equal(int64(expectedTotal), *entityResp.Entity.RiskScore))

	(&Cleanup[*generated.VendorRiskScoreDeleteOne]{client: suite.client.db.VendorRiskScore, IDs: scoreIDs}).MustDelete(scoringUser.UserCtx, t)
	(&Cleanup[*generated.EntityDeleteOne]{client: suite.client.db.Entity, ID: entity.ID}).MustDelete(scoringUser.UserCtx, t)
	(&Cleanup[*generated.EntityTypeDeleteOne]{client: suite.client.db.EntityType, ID: entity.EntityTypeID}).MustDelete(scoringUser.UserCtx, t)
}

func TestVendorRiskScoreDefaultPlusCustomQuestions(t *testing.T) {
	scoringUser := suite.userBuilder(context.Background(), t)
	entity := newVendorScoringEntity(scoringUser.UserCtx, t)

	customQuestion := models.VendorScoringQuestionDef{
		Key:             "CUSTOM-PENTEST-001",
		Name:            "Has the vendor completed a penetration test in the last 12 months?",
		Category:        enums.VendorScoringCategorySecurityPractices,
		AnswerType:      enums.VendorScoringAnswerTypeBoolean,
		SuggestedImpact: enums.VendorRiskImpactHigh,
		Enabled:         true,
	}

	configResp, err := suite.client.api.CreateVendorScoringConfig(scoringUser.UserCtx, testclient.CreateVendorScoringConfigInput{
		OwnerID: &scoringUser.OrganizationID,
		Questions: &models.VendorScoringQuestionsConfig{
			Custom: []models.VendorScoringQuestionDef{customQuestion},
		},
	})
	assert.NilError(t, err)

	configID := configResp.CreateVendorScoringConfig.VendorScoringConfig.ID

	// Fetch the config to get the assigned custom key
	fetchedConfig, err := suite.client.api.GetVendorScoringConfigByID(scoringUser.UserCtx, configID)
	assert.NilError(t, err)

	assignedCustomQ, found := lo.Find(fetchedConfig.VendorScoringConfig.Questions.Custom, func(q models.VendorScoringQuestionDef) bool {
		return q.Name == customQuestion.Name
	})
	assert.Assert(t, found)

	falseAnswer := "false"

	// Submit one default question
	defaultQ := mustVendorQuestion(t, "IAM-14.1")
	defaultResp, err := suite.client.api.CreateVendorRiskScore(scoringUser.UserCtx, testclient.CreateVendorRiskScoreInput{
		OwnerID:               &scoringUser.OrganizationID,
		VendorScoringConfigID: &configID,
		EntityID:              entity.ID,
		QuestionKey:           defaultQ.Key,
		QuestionName:          defaultQ.Name,
		QuestionCategory:      defaultQ.Category,
		Impact:                enums.VendorRiskImpactCritical,
		Likelihood:            enums.VendorRiskLikelihoodHigh,
		Answer:                &falseAnswer,
	})
	assert.NilError(t, err)

	// CRITICAL (5) * HIGH (3) = 15
	assert.Check(t, is.Equal(15.0, defaultResp.CreateVendorRiskScore.VendorRiskScore.Score))
	assertEntityRiskState(t, scoringUser.UserCtx, entity.ID, 15, "HIGH", 1)

	// Submit the custom question using the assigned key
	customResp, err := suite.client.api.CreateVendorRiskScore(scoringUser.UserCtx, testclient.CreateVendorRiskScoreInput{
		OwnerID:               &scoringUser.OrganizationID,
		VendorScoringConfigID: &configID,
		EntityID:              entity.ID,
		QuestionKey:           assignedCustomQ.Key,
		QuestionName:          assignedCustomQ.Name,
		QuestionCategory:      assignedCustomQ.Category,
		Impact:                enums.VendorRiskImpactHigh,
		Likelihood:            enums.VendorRiskLikelihoodMedium,
		Answer:                &falseAnswer,
	})
	assert.NilError(t, err)

	// HIGH (4) * MEDIUM (2) = 8
	assert.Check(t, is.Equal(8.0, customResp.CreateVendorRiskScore.VendorRiskScore.Score))

	// Aggregate: 15 + 8 = 23, coverage = 2
	assertEntityRiskState(t, scoringUser.UserCtx, entity.ID, 23, "CRITICAL", 2)

	(&Cleanup[*generated.VendorRiskScoreDeleteOne]{
		client: suite.client.db.VendorRiskScore,
		IDs:    []string{defaultResp.CreateVendorRiskScore.VendorRiskScore.ID, customResp.CreateVendorRiskScore.VendorRiskScore.ID},
	}).MustDelete(scoringUser.UserCtx, t)
	(&Cleanup[*generated.VendorScoringConfigDeleteOne]{client: suite.client.db.VendorScoringConfig, ID: configID}).MustDelete(scoringUser.UserCtx, t)
	(&Cleanup[*generated.EntityDeleteOne]{client: suite.client.db.Entity, ID: entity.ID}).MustDelete(scoringUser.UserCtx, t)
	(&Cleanup[*generated.EntityTypeDeleteOne]{client: suite.client.db.EntityType, ID: entity.EntityTypeID}).MustDelete(scoringUser.UserCtx, t)
}

func TestVendorRiskScoreMultipleSubmissionsAggregateAcrossAssessments(t *testing.T) {
	scoringUser := suite.userBuilder(context.Background(), t)
	entity := newVendorScoringEntity(scoringUser.UserCtx, t)
	assessment := (&AssessmentBuilder{client: suite.client}).MustNew(scoringUser.UserCtx, t)

	responseOne := (&AssessmentResponseBuilder{
		client:       suite.client,
		AssessmentID: assessment.ID,
		OwnerID:      scoringUser.OrganizationID,
	}).MustNew(scoringUser.UserCtx, t)
	responseTwo := (&AssessmentResponseBuilder{
		client:       suite.client,
		AssessmentID: assessment.ID,
		OwnerID:      scoringUser.OrganizationID,
	}).MustNew(scoringUser.UserCtx, t)

	question := mustVendorQuestion(t, "IAM-05.1")
	falseAnswer := "false"

	firstScoreResp, err := suite.client.api.CreateVendorRiskScore(scoringUser.UserCtx, testclient.CreateVendorRiskScoreInput{
		OwnerID:              &scoringUser.OrganizationID,
		EntityID:             entity.ID,
		AssessmentResponseID: &responseOne.ID,
		QuestionKey:          question.Key,
		QuestionName:         question.Name,
		QuestionCategory:     question.Category,
		Impact:               enums.VendorRiskImpactMedium,
		Likelihood:           enums.VendorRiskLikelihoodHigh,
		Answer:               &falseAnswer,
	})
	assert.NilError(t, err)
	assert.Assert(t, firstScoreResp != nil)

	secondScoreResp, err := suite.client.api.CreateVendorRiskScore(scoringUser.UserCtx, testclient.CreateVendorRiskScoreInput{
		OwnerID:              &scoringUser.OrganizationID,
		EntityID:             entity.ID,
		AssessmentResponseID: &responseTwo.ID,
		QuestionKey:          question.Key,
		QuestionName:         question.Name,
		QuestionCategory:     question.Category,
		Impact:               enums.VendorRiskImpactMedium,
		Likelihood:           enums.VendorRiskLikelihoodHigh,
		Answer:               &falseAnswer,
	})
	assert.NilError(t, err)
	assert.Assert(t, secondScoreResp != nil)

	where := &testclient.VendorRiskScoreWhereInput{
		EntityID:    &entity.ID,
		QuestionKey: &question.Key,
	}

	scoreListResp, err := suite.client.api.GetVendorRiskScores(scoringUser.UserCtx, nil, nil, nil, nil, nil, where)
	assert.NilError(t, err)
	assert.Assert(t, scoreListResp != nil)
	assert.Check(t, is.Len(scoreListResp.VendorRiskScores.Edges, 2))

	assessmentResponseIDs := lo.Map(scoreListResp.VendorRiskScores.Edges, func(edge *testclient.GetVendorRiskScores_VendorRiskScores_Edges, _ int) string {
		return lo.FromPtr(edge.Node.AssessmentResponseID)
	})
	assert.Check(t, lo.Contains(assessmentResponseIDs, responseOne.ID))
	assert.Check(t, lo.Contains(assessmentResponseIDs, responseTwo.ID))

	assertEntityRiskState(t, scoringUser.UserCtx, entity.ID, 18, "CRITICAL", 2)

	(&Cleanup[*generated.VendorRiskScoreDeleteOne]{
		client: suite.client.db.VendorRiskScore,
		IDs: []string{
			firstScoreResp.CreateVendorRiskScore.VendorRiskScore.ID,
			secondScoreResp.CreateVendorRiskScore.VendorRiskScore.ID,
		},
	}).MustDelete(scoringUser.UserCtx, t)
	(&Cleanup[*generated.AssessmentResponseDeleteOne]{client: suite.client.db.AssessmentResponse, IDs: []string{responseOne.ID, responseTwo.ID}}).MustDelete(scoringUser.UserCtx, t)
	(&Cleanup[*generated.AssessmentDeleteOne]{client: suite.client.db.Assessment, ID: assessment.ID}).MustDelete(scoringUser.UserCtx, t)
	(&Cleanup[*generated.TemplateDeleteOne]{client: suite.client.db.Template, ID: assessment.TemplateID}).MustDelete(scoringUser.UserCtx, t)
	(&Cleanup[*generated.EntityDeleteOne]{client: suite.client.db.Entity, ID: entity.ID}).MustDelete(scoringUser.UserCtx, t)
	(&Cleanup[*generated.EntityTypeDeleteOne]{client: suite.client.db.EntityType, ID: entity.EntityTypeID}).MustDelete(scoringUser.UserCtx, t)
}

func TestVendorScoringConfigScoringModeDefault(t *testing.T) {
	scoringUser := suite.userBuilder(context.Background(), t)

	configResp, err := suite.client.api.CreateVendorScoringConfig(scoringUser.UserCtx, testclient.CreateVendorScoringConfigInput{
		OwnerID: &scoringUser.OrganizationID,
	})
	assert.NilError(t, err)
	assert.Check(t, is.Equal(enums.VendorScoringModeAnsweredOnly, configResp.CreateVendorScoringConfig.VendorScoringConfig.ScoringMode))

	configID := configResp.CreateVendorScoringConfig.VendorScoringConfig.ID

	// Verify default thresholds are returned via All() merge
	fetched, err := suite.client.api.GetVendorScoringConfigByID(scoringUser.UserCtx, configID)
	assert.NilError(t, err)
	assert.Check(t, is.Len(fetched.VendorScoringConfig.RiskThresholds.Custom, 0))
	assert.Check(t, is.Len(fetched.VendorScoringConfig.RiskThresholds.All(), len(models.DefaultRiskThresholds)))

	(&Cleanup[*generated.VendorScoringConfigDeleteOne]{client: suite.client.db.VendorScoringConfig, ID: configID}).MustDelete(scoringUser.UserCtx, t)
}

func TestVendorScoringConfigCustomThresholds(t *testing.T) {
	scoringUser := suite.userBuilder(context.Background(), t)
	entity := newVendorScoringEntity(scoringUser.UserCtx, t)

	// Create config with custom thresholds: tighten LOW to maxScore 3 instead of 5
	customThresholds := models.RiskThresholdsConfig{
		Custom: []models.RiskThreshold{
			{Rating: enums.VendorRiskRatingLow, MaxScore: 3},
		},
	}

	configResp, err := suite.client.api.CreateVendorScoringConfig(scoringUser.UserCtx, testclient.CreateVendorScoringConfigInput{
		OwnerID:        &scoringUser.OrganizationID,
		RiskThresholds: &customThresholds,
	})
	assert.NilError(t, err)

	configID := configResp.CreateVendorScoringConfig.VendorScoringConfig.ID

	// Verify the custom threshold merged correctly
	fetched, err := suite.client.api.GetVendorScoringConfigByID(scoringUser.UserCtx, configID)
	assert.NilError(t, err)

	allThresholds := fetched.VendorScoringConfig.RiskThresholds.All()
	lowThreshold, found := lo.Find(allThresholds, func(t models.RiskThreshold) bool {
		return t.Rating == enums.VendorRiskRatingLow
	})
	assert.Assert(t, found)
	assert.Check(t, is.Equal(3.0, lowThreshold.MaxScore))

	// Submit a score of 4 (MEDIUM impact=3 * LOW likelihood=1 = 3... no, let's use values that give us 4)
	// VeryLow impact=1 * VeryHigh likelihood=4 = 4 — but we need answer=false for it to score
	// Actually: Low impact=2 * Medium likelihood=2 = 4
	question := mustVendorQuestion(t, "IAM-05.1")
	falseAnswer := "false"

	scoreResp, err := suite.client.api.CreateVendorRiskScore(scoringUser.UserCtx, testclient.CreateVendorRiskScoreInput{
		OwnerID:               &scoringUser.OrganizationID,
		VendorScoringConfigID: &configID,
		EntityID:              entity.ID,
		QuestionKey:           question.Key,
		QuestionName:          question.Name,
		QuestionCategory:      question.Category,
		Impact:                enums.VendorRiskImpactLow,
		Likelihood:            enums.VendorRiskLikelihoodMedium,
		Answer:                &falseAnswer,
	})
	assert.NilError(t, err)

	// LOW (2) * MEDIUM (2) = 4
	assert.Check(t, is.Equal(4.0, scoreResp.CreateVendorRiskScore.VendorRiskScore.Score))

	// With default thresholds, 4 <= 5 would be LOW, but with custom threshold LOW maxScore=3,
	// 4 > 3 (LOW) and <= 11 (MEDIUM), so rating should be MEDIUM
	assertEntityRiskState(t, scoringUser.UserCtx, entity.ID, 4, "MEDIUM", 1)

	(&Cleanup[*generated.VendorRiskScoreDeleteOne]{client: suite.client.db.VendorRiskScore, ID: scoreResp.CreateVendorRiskScore.VendorRiskScore.ID}).MustDelete(scoringUser.UserCtx, t)
	(&Cleanup[*generated.VendorScoringConfigDeleteOne]{client: suite.client.db.VendorScoringConfig, ID: configID}).MustDelete(scoringUser.UserCtx, t)
	(&Cleanup[*generated.EntityDeleteOne]{client: suite.client.db.Entity, ID: entity.ID}).MustDelete(scoringUser.UserCtx, t)
	(&Cleanup[*generated.EntityTypeDeleteOne]{client: suite.client.db.EntityType, ID: entity.EntityTypeID}).MustDelete(scoringUser.UserCtx, t)
}

func TestVendorScoringConfigManualModeSkipsAggregation(t *testing.T) {
	scoringUser := suite.userBuilder(context.Background(), t)
	entity := newVendorScoringEntity(scoringUser.UserCtx, t)

	// Set manual risk fields first
	manualRiskScore := int64(42)
	manualRiskRating := "CUSTOM_MANUAL"

	_, err := suite.client.api.UpdateEntity(scoringUser.UserCtx, entity.ID, testclient.UpdateEntityInput{
		RiskScore:  &manualRiskScore,
		RiskRating: &manualRiskRating,
	})
	assert.NilError(t, err)

	// Create config in MANUAL mode
	manualMode := enums.VendorScoringModeManual

	configResp, err := suite.client.api.CreateVendorScoringConfig(scoringUser.UserCtx, testclient.CreateVendorScoringConfigInput{
		OwnerID:     &scoringUser.OrganizationID,
		ScoringMode: &manualMode,
	})
	assert.NilError(t, err)

	configID := configResp.CreateVendorScoringConfig.VendorScoringConfig.ID
	assert.Check(t, is.Equal(enums.VendorScoringModeManual, configResp.CreateVendorScoringConfig.VendorScoringConfig.ScoringMode))

	// Submit a scored question — aggregate hook should NOT overwrite entity risk fields
	question := mustVendorQuestion(t, "CEK-03.1")
	falseAnswer := "false"

	scoreResp, err := suite.client.api.CreateVendorRiskScore(scoringUser.UserCtx, testclient.CreateVendorRiskScoreInput{
		OwnerID:               &scoringUser.OrganizationID,
		VendorScoringConfigID: &configID,
		EntityID:              entity.ID,
		QuestionKey:           question.Key,
		QuestionName:          question.Name,
		QuestionCategory:      question.Category,
		Impact:                enums.VendorRiskImpactCritical,
		Likelihood:            enums.VendorRiskLikelihoodVeryHigh,
		Answer:                &falseAnswer,
	})
	assert.NilError(t, err)

	// Per-record score computation still runs
	assert.Check(t, is.Equal(20.0, scoreResp.CreateVendorRiskScore.VendorRiskScore.Score))

	// Entity risk fields should remain at the manual values
	entityResp, err := suite.client.api.GetEntityByID(scoringUser.UserCtx, entity.ID)
	assert.NilError(t, err)
	assert.Assert(t, entityResp.Entity.RiskScore != nil)
	assert.Assert(t, entityResp.Entity.RiskRating != nil)
	assert.Check(t, is.Equal(manualRiskScore, *entityResp.Entity.RiskScore))
	assert.Check(t, is.Equal(manualRiskRating, *entityResp.Entity.RiskRating))

	// Update the score — entity should still not be touched
	updatedResp, err := suite.client.api.UpdateVendorRiskScore(scoringUser.UserCtx, scoreResp.CreateVendorRiskScore.VendorRiskScore.ID, testclient.UpdateVendorRiskScoreInput{
		Impact: lo.ToPtr(enums.VendorRiskImpactLow),
	})
	assert.NilError(t, err)
	assert.Check(t, is.Equal(8.0, updatedResp.UpdateVendorRiskScore.VendorRiskScore.Score))

	// Entity risk fields still untouched
	entityResp2, err := suite.client.api.GetEntityByID(scoringUser.UserCtx, entity.ID)
	assert.NilError(t, err)
	assert.Check(t, is.Equal(manualRiskScore, *entityResp2.Entity.RiskScore))
	assert.Check(t, is.Equal(manualRiskRating, *entityResp2.Entity.RiskRating))

	(&Cleanup[*generated.VendorRiskScoreDeleteOne]{client: suite.client.db.VendorRiskScore, ID: scoreResp.CreateVendorRiskScore.VendorRiskScore.ID}).MustDelete(scoringUser.UserCtx, t)
	(&Cleanup[*generated.VendorScoringConfigDeleteOne]{client: suite.client.db.VendorScoringConfig, ID: configID}).MustDelete(scoringUser.UserCtx, t)
	(&Cleanup[*generated.EntityDeleteOne]{client: suite.client.db.Entity, ID: entity.ID}).MustDelete(scoringUser.UserCtx, t)
	(&Cleanup[*generated.EntityTypeDeleteOne]{client: suite.client.db.EntityType, ID: entity.EntityTypeID}).MustDelete(scoringUser.UserCtx, t)
}

func TestVendorScoringConfigFullQuestionnaireMode(t *testing.T) {
	scoringUser := suite.userBuilder(context.Background(), t)
	entity := newVendorScoringEntity(scoringUser.UserCtx, t)

	fullMode := enums.VendorScoringModeFullQuestionnaire

	configResp, err := suite.client.api.CreateVendorScoringConfig(scoringUser.UserCtx, testclient.CreateVendorScoringConfigInput{
		OwnerID:     &scoringUser.OrganizationID,
		ScoringMode: &fullMode,
	})
	assert.NilError(t, err)

	configID := configResp.CreateVendorScoringConfig.VendorScoringConfig.ID

	// Submit one answered question out of 37 defaults
	question := mustVendorQuestion(t, "IAM-05.1")
	falseAnswer := "false"

	scoreResp, err := suite.client.api.CreateVendorRiskScore(scoringUser.UserCtx, testclient.CreateVendorRiskScoreInput{
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

	// MEDIUM (3) * MEDIUM (2) = 6 for the answered question
	assert.Check(t, is.Equal(6.0, scoreResp.CreateVendorRiskScore.VendorRiskScore.Score))

	// Under FULL_QUESTIONNAIRE, the 36 unanswered enabled questions each contribute 20 (max penalty)
	// Total = 6 + (36 * 20) = 726
	entityResp, err := suite.client.api.GetEntityByID(scoringUser.UserCtx, entity.ID)
	assert.NilError(t, err)
	assert.Assert(t, entityResp.Entity.RiskScore != nil)
	assert.Assert(t, entityResp.Entity.RiskScoreCoverage != nil)

	unansweredCount := int64(len(models.DefaultVendorScoringQuestions) - 1)
	expectedTotal := int64(6) + (unansweredCount * 20)
	assert.Check(t, is.Equal(expectedTotal, *entityResp.Entity.RiskScore))
	assert.Check(t, is.Equal(int64(1), *entityResp.Entity.RiskScoreCoverage))
	assert.Check(t, is.Equal("CRITICAL", *entityResp.Entity.RiskRating))

	(&Cleanup[*generated.VendorRiskScoreDeleteOne]{client: suite.client.db.VendorRiskScore, ID: scoreResp.CreateVendorRiskScore.VendorRiskScore.ID}).MustDelete(scoringUser.UserCtx, t)
	(&Cleanup[*generated.VendorScoringConfigDeleteOne]{client: suite.client.db.VendorScoringConfig, ID: configID}).MustDelete(scoringUser.UserCtx, t)
	(&Cleanup[*generated.EntityDeleteOne]{client: suite.client.db.Entity, ID: entity.ID}).MustDelete(scoringUser.UserCtx, t)
	(&Cleanup[*generated.EntityTypeDeleteOne]{client: suite.client.db.EntityType, ID: entity.EntityTypeID}).MustDelete(scoringUser.UserCtx, t)
}

func TestVendorScoringConfigSwitchModeRecomputes(t *testing.T) {
	scoringUser := suite.userBuilder(context.Background(), t)
	entity := newVendorScoringEntity(scoringUser.UserCtx, t)

	// Start with ANSWERED_ONLY (default)
	configResp, err := suite.client.api.CreateVendorScoringConfig(scoringUser.UserCtx, testclient.CreateVendorScoringConfigInput{
		OwnerID: &scoringUser.OrganizationID,
	})
	assert.NilError(t, err)

	configID := configResp.CreateVendorScoringConfig.VendorScoringConfig.ID

	// Submit one answered question
	question := mustVendorQuestion(t, "IAM-05.1")
	falseAnswer := "false"

	scoreResp, err := suite.client.api.CreateVendorRiskScore(scoringUser.UserCtx, testclient.CreateVendorRiskScoreInput{
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
	assert.Check(t, is.Equal(6.0, scoreResp.CreateVendorRiskScore.VendorRiskScore.Score))

	// Under ANSWERED_ONLY, entity risk score = 6
	assertEntityRiskState(t, scoringUser.UserCtx, entity.ID, 6, "MEDIUM", 1)

	// Switch to MANUAL mode — entity fields should stop being updated
	manualMode := enums.VendorScoringModeManual
	_, err = suite.client.api.UpdateVendorScoringConfig(scoringUser.UserCtx, configID, testclient.UpdateVendorScoringConfigInput{
		ScoringMode: &manualMode,
	})
	assert.NilError(t, err)

	// Note: the gala listener fires async, so in a synchronous test context the
	// recomputation may not have run yet. The key assertion is that subsequent
	// score mutations don't overwrite entity fields.
	// Manually set entity risk fields to prove they won't be overwritten
	manualScore := int64(99)
	manualRating := "MANUAL_TEST"

	_, err = suite.client.api.UpdateEntity(scoringUser.UserCtx, entity.ID, testclient.UpdateEntityInput{
		RiskScore:  &manualScore,
		RiskRating: &manualRating,
	})
	assert.NilError(t, err)

	// Update the score — aggregate hook should skip in MANUAL mode
	_, err = suite.client.api.UpdateVendorRiskScore(scoringUser.UserCtx, scoreResp.CreateVendorRiskScore.VendorRiskScore.ID, testclient.UpdateVendorRiskScoreInput{
		Impact: lo.ToPtr(enums.VendorRiskImpactCritical),
	})
	assert.NilError(t, err)

	// Entity fields remain at manual values
	entityResp, err := suite.client.api.GetEntityByID(scoringUser.UserCtx, entity.ID)
	assert.NilError(t, err)
	assert.Check(t, is.Equal(manualScore, *entityResp.Entity.RiskScore))
	assert.Check(t, is.Equal(manualRating, *entityResp.Entity.RiskRating))

	// Switch back to ANSWERED_ONLY — next mutation should recompute
	answeredOnly := enums.VendorScoringModeAnsweredOnly
	_, err = suite.client.api.UpdateVendorScoringConfig(scoringUser.UserCtx, configID, testclient.UpdateVendorScoringConfigInput{
		ScoringMode: &answeredOnly,
	})
	assert.NilError(t, err)

	// Trigger recomputation by updating the score again
	_, err = suite.client.api.UpdateVendorRiskScore(scoringUser.UserCtx, scoreResp.CreateVendorRiskScore.VendorRiskScore.ID, testclient.UpdateVendorRiskScoreInput{
		Impact: lo.ToPtr(enums.VendorRiskImpactHigh),
	})
	assert.NilError(t, err)

	// HIGH (4) * MEDIUM (2) = 8; entity should now reflect the computed aggregate
	assertEntityRiskState(t, scoringUser.UserCtx, entity.ID, 8, "MEDIUM", 1)

	(&Cleanup[*generated.VendorRiskScoreDeleteOne]{client: suite.client.db.VendorRiskScore, ID: scoreResp.CreateVendorRiskScore.VendorRiskScore.ID}).MustDelete(scoringUser.UserCtx, t)
	(&Cleanup[*generated.VendorScoringConfigDeleteOne]{client: suite.client.db.VendorScoringConfig, ID: configID}).MustDelete(scoringUser.UserCtx, t)
	(&Cleanup[*generated.EntityDeleteOne]{client: suite.client.db.Entity, ID: entity.ID}).MustDelete(scoringUser.UserCtx, t)
	(&Cleanup[*generated.EntityTypeDeleteOne]{client: suite.client.db.EntityType, ID: entity.EntityTypeID}).MustDelete(scoringUser.UserCtx, t)
}

func TestVendorScoringConfigThresholdUpdateChangesRating(t *testing.T) {
	scoringUser := suite.userBuilder(context.Background(), t)
	entity := newVendorScoringEntity(scoringUser.UserCtx, t)

	configResp, err := suite.client.api.CreateVendorScoringConfig(scoringUser.UserCtx, testclient.CreateVendorScoringConfigInput{
		OwnerID: &scoringUser.OrganizationID,
	})
	assert.NilError(t, err)

	configID := configResp.CreateVendorScoringConfig.VendorScoringConfig.ID

	// Submit a score of 6 — with default thresholds this is MEDIUM (6 > 5, <= 11)
	question := mustVendorQuestion(t, "IAM-05.1")
	falseAnswer := "false"

	scoreResp, err := suite.client.api.CreateVendorRiskScore(scoringUser.UserCtx, testclient.CreateVendorRiskScoreInput{
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
	assert.Check(t, is.Equal(6.0, scoreResp.CreateVendorRiskScore.VendorRiskScore.Score))
	assertEntityRiskState(t, scoringUser.UserCtx, entity.ID, 6, "MEDIUM", 1)

	// Now widen LOW threshold to maxScore 8 — score of 6 should become LOW
	// This triggers through the aggregate hook on the next score mutation since
	// the gala listener is async. We update thresholds then trigger a score update.
	widenedThresholds := models.RiskThresholdsConfig{
		Custom: []models.RiskThreshold{
			{Rating: enums.VendorRiskRatingLow, MaxScore: 8},
		},
	}

	_, err = suite.client.api.UpdateVendorScoringConfig(scoringUser.UserCtx, configID, testclient.UpdateVendorScoringConfigInput{
		RiskThresholds: &widenedThresholds,
	})
	assert.NilError(t, err)

	// Trigger recomputation via a no-op score update (change impact then change it back)
	_, err = suite.client.api.UpdateVendorRiskScore(scoringUser.UserCtx, scoreResp.CreateVendorRiskScore.VendorRiskScore.ID, testclient.UpdateVendorRiskScoreInput{
		Impact: lo.ToPtr(enums.VendorRiskImpactMedium),
	})
	assert.NilError(t, err)

	// Score is still 6 but now 6 <= 8 (widened LOW), so rating should be LOW
	assertEntityRiskState(t, scoringUser.UserCtx, entity.ID, 6, "LOW", 1)

	(&Cleanup[*generated.VendorRiskScoreDeleteOne]{client: suite.client.db.VendorRiskScore, ID: scoreResp.CreateVendorRiskScore.VendorRiskScore.ID}).MustDelete(scoringUser.UserCtx, t)
	(&Cleanup[*generated.VendorScoringConfigDeleteOne]{client: suite.client.db.VendorScoringConfig, ID: configID}).MustDelete(scoringUser.UserCtx, t)
	(&Cleanup[*generated.EntityDeleteOne]{client: suite.client.db.Entity, ID: entity.ID}).MustDelete(scoringUser.UserCtx, t)
	(&Cleanup[*generated.EntityTypeDeleteOne]{client: suite.client.db.EntityType, ID: entity.EntityTypeID}).MustDelete(scoringUser.UserCtx, t)
}

func TestVendorScoringConfigCustomKeyGeneration(t *testing.T) {
	scoringUser := suite.userBuilder(context.Background(), t)

	// Create config with custom questions that have no CUST- prefix keys — hook should assign them
	customQuestions := models.VendorScoringQuestionsConfig{
		Custom: []models.VendorScoringQuestionDef{
			{
				Name:            "Custom security question one",
				Category:        enums.VendorScoringCategorySecurityPractices,
				AnswerType:      enums.VendorScoringAnswerTypeBoolean,
				SuggestedImpact: enums.VendorRiskImpactMedium,
				Enabled:         true,
			},
			{
				Name:            "Custom data access question",
				Category:        enums.VendorScoringCategoryDataAccess,
				AnswerType:      enums.VendorScoringAnswerTypeBoolean,
				SuggestedImpact: enums.VendorRiskImpactHigh,
				Enabled:         true,
			},
			{
				Name:            "Custom security question two",
				Category:        enums.VendorScoringCategorySecurityPractices,
				AnswerType:      enums.VendorScoringAnswerTypeBoolean,
				SuggestedImpact: enums.VendorRiskImpactLow,
				Enabled:         true,
			},
		},
	}

	configResp, err := suite.client.api.CreateVendorScoringConfig(scoringUser.UserCtx, testclient.CreateVendorScoringConfigInput{
		OwnerID:   &scoringUser.OrganizationID,
		Questions: &customQuestions,
	})
	assert.NilError(t, err)

	configID := configResp.CreateVendorScoringConfig.VendorScoringConfig.ID

	fetched, err := suite.client.api.GetVendorScoringConfigByID(scoringUser.UserCtx, configID)
	assert.NilError(t, err)

	// Verify keys were generated with proper format
	assert.Check(t, is.Len(fetched.VendorScoringConfig.Questions.Custom, 3))

	spQuestions := lo.Filter(fetched.VendorScoringConfig.Questions.Custom, func(q models.VendorScoringQuestionDef, _ int) bool {
		return q.Category == enums.VendorScoringCategorySecurityPractices
	})
	assert.Check(t, is.Len(spQuestions, 2))
	assert.Check(t, is.Equal("CUST-SP-01.01", spQuestions[0].Key))
	assert.Check(t, is.Equal("CUST-SP-02.01", spQuestions[1].Key))

	daQuestions := lo.Filter(fetched.VendorScoringConfig.Questions.Custom, func(q models.VendorScoringQuestionDef, _ int) bool {
		return q.Category == enums.VendorScoringCategoryDataAccess
	})
	assert.Check(t, is.Len(daQuestions, 1))
	assert.Check(t, is.Equal("CUST-DA-01.01", daQuestions[0].Key))

	// Verify keys are stable on re-save — update with no question changes
	_, err = suite.client.api.UpdateVendorScoringConfig(scoringUser.UserCtx, configID, testclient.UpdateVendorScoringConfigInput{
		Questions: &fetched.VendorScoringConfig.Questions,
	})
	assert.NilError(t, err)

	reFetched, err := suite.client.api.GetVendorScoringConfigByID(scoringUser.UserCtx, configID)
	assert.NilError(t, err)

	for i, q := range reFetched.VendorScoringConfig.Questions.Custom {
		assert.Check(t, is.Equal(fetched.VendorScoringConfig.Questions.Custom[i].Key, q.Key))
	}

	(&Cleanup[*generated.VendorScoringConfigDeleteOne]{client: suite.client.db.VendorScoringConfig, ID: configID}).MustDelete(scoringUser.UserCtx, t)
}

func assertEntityRiskState(t *testing.T, ctx context.Context, entityID string, wantRiskScore int64, wantRiskRating string, wantCoverage int64) {
	t.Helper()

	entityResp, err := suite.client.api.GetEntityByID(ctx, entityID)
	assert.NilError(t, err)
	assert.Assert(t, entityResp != nil)
	assert.Assert(t, entityResp.Entity.RiskScore != nil)
	assert.Assert(t, entityResp.Entity.RiskRating != nil)
	assert.Assert(t, entityResp.Entity.RiskScoreCoverage != nil)
	assert.Check(t, is.Equal(wantRiskScore, *entityResp.Entity.RiskScore))
	assert.Check(t, is.Equal(wantRiskRating, *entityResp.Entity.RiskRating))
	assert.Check(t, is.Equal(wantCoverage, *entityResp.Entity.RiskScoreCoverage))
}

func mustVendorQuestion(t *testing.T, key string) models.VendorScoringQuestionDef {
	t.Helper()

	question, found := lo.Find(models.DefaultVendorScoringQuestions, func(question models.VendorScoringQuestionDef) bool {
		return question.Key == key
	})
	if !found {
		t.Fatalf("vendor scoring question %q not found", key)
	}

	return question
}

func newVendorScoringEntity(ctx context.Context, t *testing.T) *generated.Entity {
	t.Helper()

	return (&EntityBuilder{
		client: suite.client,
		Tier:   enums.VendorTierStandard,
	}).MustNew(ctx, t)
}
