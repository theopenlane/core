package hooks

import (
	"context"
	"fmt"
	"slices"
	"strings"
	"time"

	"entgo.io/ent"
	"github.com/rs/zerolog/log"
	"github.com/theopenlane/core/common/enums"
	"github.com/theopenlane/core/common/models"
	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/generated/actionplan"
	"github.com/theopenlane/core/internal/ent/generated/hook"
	"github.com/theopenlane/core/internal/ent/generated/remediation"
	"github.com/theopenlane/core/internal/ent/generated/review"
	"github.com/theopenlane/core/internal/ent/generated/risk"
	"github.com/theopenlane/core/pkg/logx"
)

// HookRisks sets fields on the risk based on changes to fields
func HookRisks() ent.Hook {
	return hook.On(func(next ent.Mutator) ent.Mutator {
		return hook.RiskFunc(func(ctx context.Context,
			m *generated.RiskMutation,
		) (generated.Value, error) {
			// skip if delete op, we don't care about setting any fields
			if isDeleteOp(ctx, m) {
				return next.Mutate(ctx, m)
			}

			// set last reviewed at timestamp if a completed review is added
			setLastReviewedAt(ctx, m)

			if !setStatusBasedOnRemediation(ctx, m) {
				// if action plan edge added, and status is OPEN, set status to IN_PROGRESS
				if !setStatusBasedOnActionPlan(ctx, m) && m.Op().Is(ent.OpCreate) {
					// default status to IDENTIFIED on create if not set
					m.SetStatus(enums.RiskIdentified)
				}

			}

			// set mitigated at timestamp if status is set to MITIGATED
			postUpdatesMitigationTimestamp, shouldClear := setMitigatedTimestamp(ctx, m)

			// set next review date based on review edges
			reviewUpdates := setNextReviewDate(ctx, m)

			// set impact based on score
			scoreUpdates := setImpactFromScore(ctx, m)
			if len(scoreUpdates) > 0 {
				return next.Mutate(ctx, m)
			}

			// default the due date based on sla config, allow to be updated manually
			if err := setDueDateBasedOnSLAConfig(ctx, m); err != nil {
				logx.FromContext(ctx).Error().Err(err).Msg("failed to set due date based on SLA config")
				return nil, err
			}

			retVal, err := next.Mutate(ctx, m)
			if err != nil {
				return retVal, err
			}

			if err := updateMitigationTimestamp(ctx, m, postUpdatesMitigationTimestamp, shouldClear); err != nil {
				logx.FromContext(ctx).Error().Err(err).Msg("failed to update mitigated at timestamp for related risks after risk update")
				return retVal, err
			}

			if err := updateNextReviewDate(ctx, m, reviewUpdates); err != nil {
				logx.FromContext(ctx).Error().Err(err).Msg("failed to update next review due at timestamp for related risks after risk update")
				return retVal, err
			}

			if err := updateLevelFromScore(ctx, m, scoreUpdates); err != nil {
				logx.FromContext(ctx).Error().Err(err).Msg("failed to update impact level based on residual risk score for related risks after risk update")
				return retVal, err
			}

			return retVal, nil
		})
	}, ent.OpCreate|ent.OpUpdateOne|ent.OpUpdate)
}

// setLastReviewedAt sets the last reviewed at timestamp to now if a review edge is added and the review status is completed
func setLastReviewedAt(ctx context.Context, m *generated.RiskMutation) {
	// if review edge is added, set the last reviewed at timestamp to now
	edges := m.AddedEdges()
	if len(edges) == 0 {
		return
	}

	if !slices.ContainsFunc(edges, func(e string) bool { return strings.EqualFold(e, "reviews") }) {
		return
	}

	addedReviewEdges := m.AddedIDs("reviews")

	addedIDs := make([]string, len(addedReviewEdges))
	for i, id := range addedReviewEdges {
		addedIDs[i] = id.(string)
	}

	reviews, err := m.Client().Review.Query().Where(review.IDIn(addedIDs...)).All(ctx)
	if err != nil {
		logx.FromContext(ctx).Error().Err(err).Msg("failed to query reviews for added review edges in risk mutation")
		return
	}

	for _, r := range reviews {
		if r.Status.String() == enums.ReviewStatusCompleted.String() {
			m.SetLastReviewedAt(models.DateTime(time.Now()))
			return
		}
	}
}

// setStatusBasedOnActionPlan checks if an action plan edge is added, and if the risk status is OPEN, sets it to IN_PROGRESS
// returns true if the status was updated, false otherwise
func setStatusBasedOnActionPlan(ctx context.Context, m *generated.RiskMutation) bool {
	if !m.Op().Is(ent.OpUpdateOne) {
		return false
	}

	status, ok := m.Status()
	if ok && status != enums.RiskOpen {
		return false
	}

	oldStatus, err := m.OldStatus(ctx)
	if err != nil {
		logx.FromContext(ctx).Error().Err(err).Msg("failed to get old status from risk mutation")
		return false
	}

	if oldStatus != enums.RiskOpen {
		return false
	}

	edges := m.AddedEdges()
	if len(edges) == 0 {
		return false
	}

	if !slices.ContainsFunc(edges, func(e string) bool { return strings.EqualFold(e, "action_plans") }) {
		return false
	}

	addedActionPlanEdges := m.AddedIDs("action_plans")

	addedIDs := make([]string, len(addedActionPlanEdges))
	for i, id := range addedActionPlanEdges {
		addedIDs[i] = id.(string)
	}

	actionPlans, err := m.Client().ActionPlan.Query().Where(actionplan.IDIn(addedIDs...)).All(ctx)
	if err != nil {
		logx.FromContext(ctx).Error().Err(err).Msg("failed to query action plans for added action plan edges in risk mutation")
		return false
	}

	if len(actionPlans) > 0 {
		m.SetStatus(enums.RiskInProgress)
		return true
	}

	return false
}

// setStatusBasedOnRemediation checks if a remediation edge is added or updated to completed, and if the risk status is IDENTIFIED, OPEN or IN_PROGRESS, sets it to MITIGATED
// returns true if the status was updated, false otherwise
func setStatusBasedOnRemediation(ctx context.Context, m *generated.RiskMutation) bool {
	// if the status is already set, no need to update, return true to indicate that we handled the status update
	_, ok := m.Status()
	if ok {
		return true
	}

	if !m.Op().Is(ent.OpUpdateOne) {
		return false
	}

	oldStatus, err := m.OldStatus(ctx)
	if err != nil {
		logx.FromContext(ctx).Error().Err(err).Msg("failed to get old status from risk mutation")
		return false
	}

	statusToCheck := []enums.RiskStatus{enums.RiskIdentified, enums.RiskOpen, enums.RiskInProgress}
	if !slices.Contains(statusToCheck, oldStatus) {
		return false
	}

	edges := m.AddedEdges()
	if len(edges) == 0 {
		return false
	}

	log.Error().Interface("edges", edges).Msg("edges added in risk mutation")

	if !slices.ContainsFunc(edges, func(e string) bool { return strings.EqualFold(e, "remediations") }) {
		return false
	}

	addedRemediationEdges := m.AddedIDs("remediations")

	addedIDs := make([]string, len(addedRemediationEdges))
	for i, id := range addedRemediationEdges {
		addedIDs[i] = id.(string)
	}

	remediations, err := m.Client().Remediation.Query().Where(remediation.IDIn(addedIDs...)).All(ctx)
	if err != nil {
		logx.FromContext(ctx).Error().Err(err).Msg("failed to query remediations for added remediation edges in risk mutation")
		return false
	}

	for _, r := range remediations {
		if r.Status.String() == enums.RemediationStatusCompleted.String() {
			m.SetStatus(enums.RiskMitigated)
			return true
		}
	}

	return false
}

// setDueDateBasedOnSLAConfig sets the next review due date based on the SLA config for the risk, if the due date is not already set or being updated
func setDueDateBasedOnSLAConfig(ctx context.Context, m *generated.RiskMutation) error {
	slaConfig, err := m.Client().SLADefinition.Query().All(ctx)
	if err != nil {
		return err
	}

	impact, ok := m.Impact()
	if !ok {
		return nil
	}

	for _, sla := range slaConfig {
		if riskToSeverityLevel(impact) == sla.SecurityLevel {
			dueDate := time.Now().Add(time.Duration(sla.SLADays) * 24 * time.Hour)
			m.SetDueDate(models.DateTime(dueDate))

			return nil
		}
	}

	return nil
}

// riskToSeverityLevel maps a risk impact level to a normalized severity level for SLA config matching
func riskToSeverityLevel(impact enums.RiskImpact) enums.SecurityLevel {
	switch impact {
	case enums.RiskImpactCritical:
		return enums.SecurityLevelCritical
	case enums.RiskImpactHigh:
		return enums.SecurityLevelHigh
	case enums.RiskImpactModerate:
		return enums.SecurityLevelMedium
	default:
		return enums.SecurityLevelLow
	}
}

// setImpactFromScore sets the impact level based on the score
// if the residual risk score is set, it will use that to set the impact level, otherwise it will use the inherent risk score
// if the impact level is manually set, it will not override it
func setImpactFromScore(ctx context.Context, m *generated.RiskMutation) []*generated.Risk {
	// if residual risk is set, update severity level based on the residual risk score, instead of the inherent risk score
	// do not overwrite an impact manually set
	impact, ok := m.Impact()
	if ok && impact != enums.RiskImpactInvalid {
		return nil
	}

	residualRiskScore, residualOK := m.ResidualScore()
	score, scoreOk := m.Score()
	cleared := m.ResidualScoreCleared()

	if !residualOK && !scoreOk && !cleared {
		if m.Op().Is(ent.OpCreate) {
			m.SetImpact(enums.RiskImpactLow)
		}

		return nil
	}

	if residualOK && residualRiskScore != 0 {
		level := impactLevelFromScore(residualRiskScore)
		m.SetImpact(level)

		return nil
	}

	if scoreOk && score != 0 {
		level := impactLevelFromScore(score)
		m.SetImpact(level)

		return nil
	}

	if !cleared {
		return nil
	}

	switch m.Op() {
	case ent.OpUpdateOne:
		score, err := m.OldScore(ctx)
		if err != nil {
			logx.FromContext(ctx).Error().Err(err).Msg("failed to get old score from risk mutation")
			return nil
		}

		level := impactLevelFromScore(score)
		m.SetImpact(level)
	case ent.OpUpdate:
		updates, _ := determineUpdateAll(ctx, m, risk.FieldScore, "", false)

		return updates
	}

	return nil
}

const (
	riskScoreCriticalMin     = 17
	riskScoreHighlyLikelyMin = 10
	riskScoreHighlyLikelyMax = 16
	riskScoreLikelyMin       = 5
	riskScoreLikelyMax       = 9
)

// impactLevelFromScore maps a CVSS v4.0 score to a normalized impact level
func impactLevelFromScore(score int) enums.RiskImpact {
	switch {
	case score >= riskScoreCriticalMin:
		return enums.RiskImpactCritical
	case score >= riskScoreHighlyLikelyMin && score <= riskScoreHighlyLikelyMax:
		return enums.RiskImpactHigh
	case score >= riskScoreLikelyMin && score <= riskScoreLikelyMax:
		return enums.RiskImpactModerate
	default:
		return enums.RiskImpactLow
	}
}

// setMitigatedTimestamp sets the mitigated at timestamp if the risk is being marked as mitigated, and clears it if it's being unmarked as mitigated
func setMitigatedTimestamp(ctx context.Context, m *generated.RiskMutation) (updates []*generated.Risk, shouldClear bool) {
	status, ok := m.Status()
	if !ok {
		return nil, false
	}

	setMitigatedTimestamp := func() {
		m.SetMitigatedAt(models.DateTime(time.Now()))
	}

	switch status {
	case enums.RiskMitigated:
		switch m.Op() {
		case ent.OpCreate:
			m.SetMitigatedAt(models.DateTime(time.Now()))
			return nil, false
		case ent.OpUpdate, ent.OpUpdateOne:
			// look to see if all the old risks have the same old value, if the old value is
			// not mitigated, then we can update all, this allows us to follow the same path for
			// update and updateOne
			updates, updateAll := determineUpdateAll(ctx, m, risk.FieldStatus, enums.RiskMitigated.String(), false)

			if updateAll {
				setMitigatedTimestamp()
			}

			// return the old values if we need to do a secondary update
			// for each risk that is being updated to mitigated, if it was not mitigated before, we need to set the mitigated at timestamp
			if !updateAll && len(updates) > 0 {
				return updates, false
			}
		}
	default:
		switch m.Op() {
		case ent.OpUpdateOne, ent.OpUpdate:
			updates, updateAll := determineUpdateAll(ctx, m, risk.FieldStatus, enums.RiskMitigated.String(), true)

			if updateAll {
				m.ClearMitigatedAt()
			}

			// return the old values if we need to do a secondary update
			if !updateAll && len(updates) > 0 {
				return updates, true
			}
		}
	}

	return nil, false
}

// setNextReviewDate sets the next review date based on the review frequency and last reviewed at timestamp
func setNextReviewDate(ctx context.Context, m *generated.RiskMutation) (updates []*generated.Risk) {
	// if review frequency is changed, set the next review date based on the last reviewed at timestamp + frequency
	reviewFrequency, ok := m.ReviewFrequency()
	if !ok {
		return
	}

	setNextReviewDate := func(frequency enums.Frequency, lastReviewAt models.DateTime) {
		if lastReviewAt.IsZero() {
			lastReviewAt = models.DateTime(time.Now())
		}

		duration := getDurationForFrequency(frequency)

		nextReviewDueAt := time.Time(lastReviewAt).Add(duration)

		m.SetNextReviewDueAt(models.DateTime(nextReviewDueAt))
	}

	switch m.Op() {
	case ent.OpCreate:
		// get the last reviewed at timestamp, if it's not set, use the current time
		lastReviewedAt, _ := m.LastReviewedAt()

		setNextReviewDate(reviewFrequency, lastReviewedAt)

		return nil
	case ent.OpUpdateOne:
		oldFrequency, err := m.OldReviewFrequency(ctx)
		if err != nil {
			logx.FromContext(ctx).Error().Err(err).Msg("failed to get old review frequency")
			return nil
		}

		if oldFrequency != reviewFrequency {
			// get the last reviewed at timestamp, if it's not set, use the current time
			lastReviewedAt, _ := m.LastReviewedAt()

			setNextReviewDate(reviewFrequency, lastReviewedAt)

		}

		return nil
	case ent.OpUpdate:
		updates, _ := determineUpdateAll(ctx, m, risk.FieldReviewFrequency, reviewFrequency.String(), false)

		if len(updates) > 0 {
			return updates
		}
	}

	return nil
}

// updateMitigationTimestamp updates the mitigated at timestamp for related risks after a risk update, if the status is being updated to mitigated or unmitigated
func updateMitigationTimestamp(ctx context.Context, m *generated.RiskMutation, updates []*generated.Risk, shouldClear bool) error {
	if len(updates) == 0 {
		return nil
	}

	ids := make([]string, len(updates))
	for i, riskToUpdate := range updates {
		ids[i] = riskToUpdate.ID
	}

	base := m.Client().Risk.Update().Where(risk.IDIn(ids...))

	if shouldClear {
		return base.ClearMitigatedAt().Exec(ctx)
	}

	return base.SetMitigatedAt(models.DateTime(time.Now())).Exec(ctx)
}

// updateNextReviewDate updates the next review due at timestamp for related risks after a risk update, if the review frequency is being updated
func updateNextReviewDate(ctx context.Context, m *generated.RiskMutation, updates []*generated.Risk) error {
	if len(updates) == 0 {
		return nil
	}

	for _, riskToUpdate := range updates {
		duration := getDurationForFrequency(riskToUpdate.ReviewFrequency)
		nextReviewDueAt := time.Time(*riskToUpdate.LastReviewedAt).Add(duration)
		err := m.Client().Risk.Update().Where(risk.IDEQ(riskToUpdate.ID)).
			SetNextReviewDueAt(models.DateTime(nextReviewDueAt)).Exec(ctx)
		if err != nil {
			logx.FromContext(ctx).Error().Err(err).Msg("failed to update next review due at timestamp for related risks after risk update")
			return err
		}

	}

	return nil
}

// updateLevelFromScore updates the impact level based on the score for related risks after a risk update, if the score or residual score is being updated
func updateLevelFromScore(ctx context.Context, m *generated.RiskMutation, updates []*generated.Risk) error {
	if len(updates) == 0 {
		return nil
	}

	for _, riskToUpdate := range updates {
		score := riskToUpdate.Score
		if riskToUpdate.ResidualScore != 0 {
			score = riskToUpdate.ResidualScore
		}

		level := impactLevelFromScore(score)

		err := m.Client().Risk.Update().Where(risk.IDEQ(riskToUpdate.ID)).
			SetImpact(level).Exec(ctx)
		if err != nil {
			logx.FromContext(ctx).Error().Err(err).Msg("failed to update impact level based on residual risk score for related risks after risk update")
			return err
		}

	}

	return nil
}

// getDurationForFrequency returns the duration for a given frequency, defaulting to yearly if the frequency is not recognized
func getDurationForFrequency(frequency enums.Frequency) time.Duration {
	switch frequency {
	case enums.FrequencyYearly:
		return 365 * 24 * time.Hour
	case enums.FrequencyBiAnnually:
		return 180 * 24 * time.Hour
	case enums.FrequencyQuarterly:
		return 90 * 24 * time.Hour
	case enums.FrequencyMonthly:
		return 30 * 24 * time.Hour
	default:
		return 365 * 24 * time.Hour
	}
}

// determineUpdateAll determines if an update operation is updating all risks to the same value for a given field, and returns the old values if not
func determineUpdateAll(ctx context.Context, m *generated.RiskMutation, field string, compareValue string, shouldEqual bool) ([]*generated.Risk, bool) {
	oldVals, err := fetchOldRiskValue(ctx, m, field)
	if err != nil {
		logx.FromContext(ctx).Error().Err(err).Msg("failed to fetch old risk values for update")
		return nil, false
	}

	uniqueValues := []string{}

	for _, val := range oldVals {
		var value string
		switch field {
		case risk.FieldStatus:
			value = val.Status.String()
		case risk.FieldReviewFrequency:
			value = val.ReviewFrequency.String()
		case risk.FieldScore:
			value = fmt.Sprintf("%d", val.Score)
		case risk.FieldResidualScore:
			value = fmt.Sprintf("%d", val.ResidualScore)
		default:
			return nil, false
		}

		if !slices.Contains(uniqueValues, value) {
			uniqueValues = append(uniqueValues, value)
		}
	}

	if len(uniqueValues) == 1 {
		// if they should equal to do the update, and the unique value is equal to the compare value, then we can update all
		if shouldEqual && uniqueValues[0] == compareValue {
			return nil, true
		} else if !shouldEqual && uniqueValues[0] != compareValue {
			return nil, true
		}

		// otherwise we are saying there is nothing to update
		return nil, false
	}

	// if they do not need to equal, and they are all something other than the compare value, then we can update all
	if !shouldEqual && !slices.Contains(uniqueValues, compareValue) {
		return nil, true
	}

	// otherwise we should just return the ones that need to be updated
	var updates []*generated.Risk

	for _, r := range oldVals {
		val, err := r.Value(field)
		if err != nil {
			logx.FromContext(ctx).Error().Err(err).Msg("failed to get old risk value for update")
			return nil, false
		}

		if shouldEqual && !strings.EqualFold(val.(string), compareValue) {
			updates = append(updates, r)
		} else if !shouldEqual && strings.EqualFold(val.(string), compareValue) {
			updates = append(updates, r)
		}
	}

	return updates, false
}

// fetchOldRiskValue fetches the old values for a given field for all risks being updated in an update operation
var fetchOldRiskValue = func(ctx context.Context, m *generated.RiskMutation, field string) ([]*generated.Risk, error) {
	ids, err := m.IDs(ctx)
	if err != nil {
		logx.FromContext(ctx).Error().Err(err).Msg("failed to get risk ids from mutation")
		return nil, err
	}

	oldValues, err := m.Client().Risk.Query().Where(risk.IDIn(ids...)).Select(risk.FieldID, field).All(ctx)
	if err != nil {
		logx.FromContext(ctx).Error().Err(err).Msg("failed to get old risk status")
		return nil, err
	}

	return oldValues, nil
}
