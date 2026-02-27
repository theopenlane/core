package graphapi

import (
	"context"
	"fmt"
	"time"

	"entgo.io/ent/dialect/sql"
	"github.com/rs/zerolog/log"
	"github.com/samber/lo"

	"github.com/theopenlane/core/internal/ent/generated/privacy"
	historygenerated "github.com/theopenlane/core/internal/ent/historygenerated"
	"github.com/theopenlane/core/internal/ent/historygenerated/controlhistory"
	"github.com/theopenlane/core/internal/ent/historygenerated/standardhistory"
	"github.com/theopenlane/core/internal/graphapi/common"
	"github.com/theopenlane/core/internal/graphapi/model"
	"github.com/theopenlane/core/pkg/jsonx"
)

var controlDiffFields = []string{
	"title",
	"description",
	"description_json",
	"aliases",
	"category",
	"subcategory",
	"mapped_categories",
	"assessment_objectives",
	"assessment_methods",
	"control_questions",
	"implementation_guidance",
	"example_evidence",
	"references",
	"testing_procedures",
	"evidence_requests",
}

func (r *Resolver) controlDiff(ctx context.Context, input model.ControlDiffInput) (*model.ControlDiffPayload, error) {
	historyClient := r.db.HistoryClient
	if historyClient == nil {
		return nil, fmt.Errorf("history client not configured") //nolint:err113
	}

	oldRevisionTime, err := getTimestampOfRevision(ctx, historyClient, input.StandardID, input.OldRevision)
	if err != nil {
		return nil, err
	}

	newRevisionTime, err := getTimestampOfRevision(ctx, historyClient, input.StandardID, input.NewRevision)
	if err != nil {
		return nil, err
	}

	revisionDiffCutoff, err := getNextTimestampOfRevison(ctx, historyClient, input.StandardID, input.NewRevision, newRevisionTime)
	if err != nil {
		return nil, err
	}

	log.Debug().Str("standard_id", input.StandardID).
		Time("old_revision_time", oldRevisionTime).Time("new_revision_time", newRevisionTime).
		Time("revision_diff_cutoff", revisionDiffCutoff).
		Msg("control diffing timestamps")

	allowCtx := privacy.DecisionContext(ctx, privacy.Allow)

	oldSnapshots, err := getControlHistories(allowCtx, historyClient, input.StandardID, oldRevisionTime)
	if err != nil {
		return nil, err
	}

	newSnapshots, err := getControlHistories(allowCtx, historyClient, input.StandardID, revisionDiffCutoff)
	if err != nil {
		return nil, err
	}

	oldByRef := getFirstControlByRefCode(oldSnapshots)
	newByRef := getFirstControlByRefCode(newSnapshots)

	controlChanges, err := detectAndBuildControlChanges(oldByRef, newByRef)
	if err != nil {
		return nil, err
	}

	return &model.ControlDiffPayload{
		StandardID:  input.StandardID,
		OldRevision: input.OldRevision,
		NewRevision: input.NewRevision,
		Changes:     controlChanges,
	}, nil
}

func getTimestampOfRevision(ctx context.Context, hc *historygenerated.Client, standardID, revision string) (time.Time, error) {
	history, err := hc.StandardHistory.Query().
		Where(
			standardhistory.Ref(standardID),
			standardhistory.Revision(revision),
		).
		Order(standardhistory.ByHistoryTime(sql.OrderDesc())).
		First(ctx)
	if err != nil {
		return time.Time{}, parseRequestError(ctx, err, common.Action{Action: common.ActionGet, Object: "standard history"})
	}

	return history.HistoryTime, nil
}

// getNextTimestampOfRevison finds the timestamp of the next standard revision change
// after afterTime. if the current revision is still the latest, returns time.Now().
func getNextTimestampOfRevison(
	ctx context.Context, hc *historygenerated.Client, standardID, currentRevision string, pastTime time.Time) (time.Time, error) {
	history, err := hc.StandardHistory.Query().
		Where(
			standardhistory.Ref(standardID),
			standardhistory.HistoryTimeGT(pastTime),
			standardhistory.RevisionNEQ(currentRevision),
		).
		Order(standardhistory.ByHistoryTime(sql.OrderAsc())).
		First(ctx)
	if err != nil {
		if historygenerated.IsNotFound(err) {
			return time.Now(), nil
		}

		return time.Time{}, parseRequestError(ctx, err, common.Action{Action: common.ActionGet, Object: "standard history"})
	}

	return history.HistoryTime, nil
}

func getControlHistories(ctx context.Context, hc *historygenerated.Client, standardID string, timestamp time.Time) ([]*historygenerated.ControlHistory, error) {
	controlHistories, err := hc.ControlHistory.Query().
		Where(
			controlhistory.StandardIDEQ(standardID),
			controlhistory.HistoryTimeLTE(timestamp),
		).
		Order(
			controlhistory.ByHistoryTime(sql.OrderDesc()),
			controlhistory.ByID(sql.OrderDesc()),
		).
		All(ctx)
	if err != nil {
		return nil, parseRequestError(ctx, err, common.Action{Action: common.ActionGet, Object: "control history"})
	}

	return lo.UniqBy(controlHistories, func(ch *historygenerated.ControlHistory) string {
		return ch.Ref
	}), nil
}

func detectAndBuildControlChanges(oldControlsByRefCode, newControlsByRefCode map[string]*historygenerated.ControlHistory) ([]*model.ControlChange, error) {
	var changes []*model.ControlChange

	for code, control := range newControlsByRefCode {
		oldRec, ok := oldControlsByRefCode[code]
		if !ok {
			continue
		}

		diffs, err := diffControlHistories(oldRec, control)
		if err != nil {
			return nil, err
		}

		if len(diffs) == 0 {
			continue
		}

		changes = append(changes, &model.ControlChange{
			RefCode: code,
			Title:   control.Title,
			Diffs:   diffs,
		})
	}

	if changes == nil {
		changes = []*model.ControlChange{}
	}

	return changes, nil
}

func getFirstControlByRefCode(snapshots []*historygenerated.ControlHistory) map[string]*historygenerated.ControlHistory {
	out := make(map[string]*historygenerated.ControlHistory, len(snapshots))

	for _, s := range snapshots {
		if _, exists := out[s.RefCode]; !exists {
			out[s.RefCode] = s
		}
	}

	return out
}

func diffControlHistories(old, newControls *historygenerated.ControlHistory) ([]*model.ControlFieldDiff, error) {
	oldMap, err := jsonx.ToMap(old)
	if err != nil {
		return nil, fmt.Errorf("failed to serialize old control history: %w", err)
	}

	newMap, err := jsonx.ToMap(newControls)
	if err != nil {
		return nil, fmt.Errorf("failed to serialize new control history: %w", err)
	}

	diffs := lo.FilterMap(controlDiffFields, func(field string, _ int) (*model.ControlFieldDiff, bool) {
		oldVal := oldMap[field]
		newVal := newMap[field]

		if controlDiffValuesEqual(oldVal, newVal) {
			return nil, false
		}

		diff := &model.ControlFieldDiff{
			Field:    field,
			OldValue: oldVal,
			NewValue: newVal,
		}

		diffText := workflowProposalDiff(oldVal, newVal)
		if diffText != "" {
			diff.Diff = &diffText
		}

		return diff, true
	})

	return diffs, nil
}

func controlDiffValuesEqual(a, b any) bool {
	if a == nil && b == nil {
		return true
	}

	if a == nil || b == nil {
		return false
	}

	aStr, aOK := workflowProposalDiffString(a)
	bStr, bOK := workflowProposalDiffString(b)

	if !aOK || !bOK {
		return false
	}

	return aStr == bStr
}
