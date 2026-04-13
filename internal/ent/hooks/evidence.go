package hooks

import (
	"context"
	"time"

	"entgo.io/ent"

	"github.com/theopenlane/core/common/enums"
	"github.com/theopenlane/core/common/models"
	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/generated/evidence"
	"github.com/theopenlane/core/internal/ent/generated/file"
	"github.com/theopenlane/core/internal/ent/generated/hook"
	"github.com/theopenlane/core/internal/ent/privacy/utils"
	"github.com/theopenlane/core/pkg/objects"
)

// HookEvidenceReviewDate runs on evidence mutations and calculate the next review time based
// on creation date + review frequency
func HookEvidenceReviewDate() ent.Hook {
	return hook.On(func(next ent.Mutator) ent.Mutator {
		return hook.EvidenceFunc(func(ctx context.Context, m *generated.EvidenceMutation) (generated.Value, error) {
			creationDate, creationDateMutated := m.CreationDate()
			frequency, frequencyMutated := m.ReviewFrequency()

			if !creationDateMutated && !frequencyMutated {
				return next.Mutate(ctx, m)
			}

			switch {
			case m.Op().Is(ent.OpCreate):
				if !creationDateMutated {
					creationDate = evidence.DefaultCreationDate()
				}

				if !frequencyMutated {
					frequency = evidence.DefaultReviewFrequency
				}
			case m.Op().Is(ent.OpUpdateOne):
				if !creationDateMutated {
					oldCreationDate, err := m.OldCreationDate(ctx)
					if err != nil {
						return nil, err
					}

					if oldCreationDate != nil {
						creationDate = *oldCreationDate
					}
				}

				if !frequencyMutated {
					oldFrequency, err := m.OldReviewFrequency(ctx)
					if err != nil {
						return nil, err
					}

					frequency = oldFrequency
				}
			case m.Op().Is(ent.OpUpdate):
				if creationDateMutated && frequencyMutated {
					break
				}

				ids, err := m.IDs(ctx)
				if err != nil {
					return nil, err
				}

				oldEvidences, err := m.Client().Evidence.Query().
					Where(evidence.IDIn(ids...)).
					Select(evidence.FieldID, evidence.FieldCreationDate, evidence.FieldReviewFrequency).
					All(ctx)
				if err != nil {
					return nil, err
				}

				v, err := next.Mutate(ctx, m)
				if err != nil {
					return nil, err
				}

				for _, ev := range oldEvidences {
					evidenceCreationDate := creationDate
					if !creationDateMutated && ev.CreationDate != nil {
						evidenceCreationDate = *ev.CreationDate
					}

					evidenceFrequency := frequency
					if !frequencyMutated {
						evidenceFrequency = ev.ReviewFrequency
					}

					update := m.Client().Evidence.UpdateOneID(ev.ID)
					if evidenceFrequency == enums.FrequencyNone || time.Time(evidenceCreationDate).IsZero() {
						update.ClearRenewalDate()
					} else {
						update.SetRenewalDate(getNextReviewDate(evidenceFrequency, models.DateTime(evidenceCreationDate)))
					}

					err = update.Exec(ctx)
					if err != nil {
						return v, err
					}
				}

				return v, nil
			default:
				return next.Mutate(ctx, m)
			}

			if frequency == enums.FrequencyNone || time.Time(creationDate).IsZero() {
				m.ClearRenewalDate()
				return next.Mutate(ctx, m)
			}

			m.SetRenewalDate(getNextReviewDate(frequency, models.DateTime(creationDate)))

			return next.Mutate(ctx, m)
		})
	}, ent.OpCreate|ent.OpUpdateOne|ent.OpUpdate)
}

// HookEvidenceFiles runs on evidence mutations to check for uploaded files
func HookEvidenceFiles() ent.Hook {
	return hook.On(func(next ent.Mutator) ent.Mutator {
		return hook.EvidenceFunc(func(ctx context.Context, m *generated.EvidenceMutation) (generated.Value, error) {
			if !isDeleteOp(ctx, m) {
				// validate creation date if only
				// - it is a create operation
				// - it was provided in an update operation
				creationDate, ok := m.CreationDate()
				op := m.Op()

				if op == ent.OpCreate && !ok {
					return nil, ErrZeroTimeNotAllowed
				}

				if ok && time.Time(creationDate).After(time.Now()) {
					return nil, ErrFutureTimeNotAllowed
				}

				if err := validateTimeNotInPast(m.RenewalDate()); err != nil {
					return nil, err
				}

				hasURL := checkEvidenceHasURL(ctx, m)
				hasFiles := checkEvidenceHasFiles(ctx, m)

				// we should always take the sent status; we just want to set missing artifact
				// if its created or updated and has not file or url and status isn't sent explicitly
				_, ok = m.Status()
				if !hasURL && !hasFiles && !ok {
					m.SetStatus(enums.EvidenceStatusMissingArtifact)
				}

				// if being updated, and the old status is MISSING_ARTIFACT, but contains a file
				// and url, we need to reset the state though if the status is not passed in the mutation
				// Else we default to submitted
				if m.Op().Is(ent.OpUpdateOne) {
					oldStatus, err := m.OldStatus(ctx)
					if err != nil {
						return nil, err
					}

					if oldStatus == enums.EvidenceStatusMissingArtifact && (hasURL || hasFiles) && !ok {
						m.SetStatus(enums.EvidenceStatusSubmitted)
					}
				}

				if m.Op().Is(ent.OpCreate) {
					_, ok = m.Status()
					if !ok {
						m.SetStatus(enums.EvidenceStatusSubmitted)
					}
				}
			}

			// check for uploaded files (e.g. avatar image)
			fileIDs := objects.GetFileIDsFromContext(ctx)
			if len(fileIDs) > 0 {
				var err error

				ctx, err = checkEvidenceFiles(ctx, m)
				if err != nil {
					return nil, err
				}

				m.AddFileIDs(fileIDs...)
			}

			return next.Mutate(ctx, m)
		})
	}, ent.OpCreate|ent.OpUpdateOne|ent.OpUpdate)
}

// checkEvidenceFiles checks if a evidence files are provided and sets the local file ID(s)
func checkEvidenceFiles[T utils.GenericMutation](ctx context.Context, m T) (context.Context, error) {
	key := "evidenceFiles"

	return objects.ProcessFilesForMutation(ctx, m, key)
}

// checkEvidenceHasFiles checks if evidence has any attached files
func checkEvidenceHasFiles(ctx context.Context, m *generated.EvidenceMutation) bool {
	if len(objects.GetFileIDsFromContext(ctx)) > 0 {
		return true
	}

	if len(m.FilesIDs()) > 0 {
		return true
	}

	if m.Op().Is(ent.OpCreate) {
		return false
	}

	id, ok := m.ID()
	if !ok || id == "" {
		return false
	}

	currentFileCount, err := m.Client().Evidence.Query().
		Where(evidence.ID(id)).
		QueryFiles().
		Select(file.FieldID).
		Count(ctx)
	if err != nil {
		return false
	}

	fileIDs := m.RemovedFilesIDs()
	return len(fileIDs) < currentFileCount
}

func checkEvidenceHasURL(ctx context.Context, m *generated.EvidenceMutation) bool {
	if m.FieldCleared(evidence.FieldURL) {
		return false
	}

	url, ok := m.URL()
	if ok {
		return url != ""
	}

	if m.Op().Is(ent.OpCreate) {
		return false
	}

	id, ok := m.ID()
	if !ok || id == "" {
		return false
	}

	evidence, err := m.Client().Evidence.Query().
		Where(evidence.ID(id)).
		Select(evidence.FieldURL).
		Only(ctx)
	if err != nil {
		return false
	}

	return evidence.URL != ""
}
