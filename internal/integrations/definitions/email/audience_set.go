package email

import (
	"context"

	"github.com/theopenlane/core/v2/internal/audiences"
	"github.com/theopenlane/core/v2/internal/ent/generated"
	"github.com/theopenlane/core/v2/internal/ent/generated/campaigntarget"
)

type campaignTargetSet struct {
	seen map[string]struct{}
}

func (s *campaignTargetSet) add(recipient audiences.ResolvedRecipient) bool {
	key := canonicalizeEmail(recipient.Email)
	if key == "" {
		return false
	}

	if _, ok := s.seen[key]; ok {
		return false
	}

	s.seen[key] = struct{}{}
	return true
}

// load fetches any already created target so we do not send out the same campaign email
// twice to a user
func (s *campaignTargetSet) load(ctx context.Context, db *generated.Client, campaignID string) error {
	var lastKnownID string

	for {
		query := db.CampaignTarget.Query().
			Where(campaigntarget.CampaignIDEQ(campaignID)).
			Select(campaigntarget.FieldID, campaigntarget.FieldEmail).
			Order(campaigntarget.ByID()).
			Limit(audienceTargetBatchSize)

		if lastKnownID != "" {
			query.Where(campaigntarget.IDGT(lastKnownID))
		}

		campaignTargets, err := query.All(ctx)
		if err != nil {
			return err
		}

		for _, target := range campaignTargets {
			lastKnownID = target.ID

			key := canonicalizeEmail(target.Email)
			if key != "" {
				s.seen[key] = struct{}{}
			}
		}

		if len(campaignTargets) < audienceTargetBatchSize {
			break
		}
	}

	return nil
}
