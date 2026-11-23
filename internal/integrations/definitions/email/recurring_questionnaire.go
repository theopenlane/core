package email

import (
	"context"
	"maps"
	"time"

	"github.com/theopenlane/iam/auth"

	"github.com/theopenlane/core/common/enums"

	"github.com/theopenlane/core/v2/internal/ent/generated"
	"github.com/theopenlane/core/v2/internal/ent/generated/assessmentpolicy"
	"github.com/theopenlane/core/v2/internal/ent/generated/campaign"
	"github.com/theopenlane/core/v2/internal/ent/generated/campaigntarget"
	"github.com/theopenlane/core/v2/internal/workflows"
)

const recurringQuestionnaireMetadataKey = "recurringQuestionnaire"

func createRecurringQuestionnaire(ctx context.Context, db *generated.Client, camp *generated.Campaign, occurrence string) error {
	previousRound, _ := camp.Metadata[recurringQuestionnaireMetadataKey].(map[string]any)
	assessmentID, _ := previousRound["assessmentID"].(string)
	scheduledAt, _ := previousRound["scheduledAt"].(string)

	if assessmentID != "" && assessmentID == camp.AssessmentID && scheduledAt == occurrence {
		return nil
	}

	ctx = auth.EnsureIntegrationCaller(ctx, camp.OwnerID)

	_, err := workflows.WithTx(ctx, db, nil, func(ctx context.Context, tx *generated.Tx) (struct{}, error) {
		assessmentObj, err := createRecurringQuestionnaireAssessment(ctx, tx.Client(), camp)
		if err != nil {
			return struct{}{}, err
		}

		metadata := make(map[string]any, len(camp.Metadata)+1)
		maps.Copy(metadata, camp.Metadata)
		metadata[recurringQuestionnaireMetadataKey] = map[string]any{
			"assessmentID": assessmentObj.ID,
			"scheduledAt":  occurrence,
		}

		previousCampaignFilter := campaign.AssessmentIDIsNil()
		if camp.AssessmentID != "" {
			previousCampaignFilter = campaign.AssessmentIDEQ(camp.AssessmentID)
		}

		err = tx.Campaign.UpdateOneID(camp.ID).
			Where(campaign.NextRunAtEQ(*camp.NextRunAt), previousCampaignFilter).
			SetAssessmentID(assessmentObj.ID).
			SetMetadata(metadata).
			Exec(ctx)
		if err != nil {
			return struct{}{}, err
		}

		err = tx.CampaignTarget.Update().
			Where(campaigntarget.CampaignIDEQ(camp.ID)).
			SetStatus(enums.AssessmentResponseStatusNotStarted).
			ClearSentAt().
			ClearCompletedAt().
			Exec(ctx)

		return struct{}{}, err
	})

	return err
}

func createRecurringQuestionnaireAssessment(ctx context.Context, db *generated.Client, camp *generated.Campaign) (*generated.Assessment, error) {
	create := db.Assessment.Create().
		SetName(generateQuestionnaireTitle(camp))

	if camp.AssessmentID == "" && camp.TemplateID == "" {
		return nil, ErrCampaignMissingAssessment
	}

	if camp.AssessmentID == "" {
		return create.SetOwnerID(camp.OwnerID).
			SetTemplateID(camp.TemplateID).
			Save(ctx)
	}

	source, err := db.Assessment.Get(ctx, camp.AssessmentID)
	if err != nil {
		return nil, err
	}

	cloned, err := create.SetOwnerID(source.OwnerID).
		SetAssessmentType(source.AssessmentType).
		SetResponseDueDuration(source.ResponseDueDuration).
		SetJsonconfig(source.Jsonconfig).
		SetUischema(source.Uischema).
		Save(ctx)
	if err != nil {
		return nil, err
	}

	if err := cloneAssessmentPolicies(ctx, db, source.ID, cloned); err != nil {
		return nil, err
	}

	return cloned, nil
}

func cloneAssessmentPolicies(ctx context.Context, db *generated.Client, sourceID string, cloned *generated.Assessment) error {
	policies, err := db.AssessmentPolicy.Query().
		Where(assessmentpolicy.AssessmentIDEQ(sourceID)).
		Select(assessmentpolicy.FieldInternalPolicyID, assessmentpolicy.FieldPolicyRevision).
		All(ctx)
	if err != nil || len(policies) == 0 {
		return err
	}

	createBulk := make([]*generated.AssessmentPolicyCreate, 0, len(policies))

	for _, policy := range policies {
		createBulk = append(createBulk, db.AssessmentPolicy.Create().
			SetAssessmentID(cloned.ID).
			SetInternalPolicyID(policy.InternalPolicyID).
			SetPolicyRevision(policy.PolicyRevision).
			SetOwnerID(cloned.OwnerID))
	}

	return db.AssessmentPolicy.CreateBulk(createBulk...).Exec(ctx)
}

func generateQuestionnaireTitle(camp *generated.Campaign) string {
	defaultLayout := "01-02-2006"

	switch camp.RecurrenceFrequency {
	case enums.FrequencyYearly:
		defaultLayout = "2006"
	case enums.FrequencyQuarterly, enums.FrequencyMonthly:
		defaultLayout = "01-2006"
	}

	return camp.Name + " - " + time.Time(*camp.NextRunAt).Format(defaultLayout)
}
