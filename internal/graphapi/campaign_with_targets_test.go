package graphapi

import (
	"testing"

	"github.com/samber/lo"
	"gotest.tools/v3/assert"
	"gotest.tools/v3/assert/cmp"

	"github.com/theopenlane/core/internal/ent/generated"
)

// TestSetRecipientCountIfNeeded tests the recipient count helper.
func TestSetRecipientCountIfNeeded(t *testing.T) {
	t.Run("sets count when nil", func(t *testing.T) {
		input := &generated.CreateCampaignInput{}
		setRecipientCountIfNeeded(input, 5)
		assert.Check(t, cmp.Equal(5, lo.FromPtrOr(input.RecipientCount, 0)))
	})

	t.Run("preserves existing count", func(t *testing.T) {
		existing := 10
		input := &generated.CreateCampaignInput{RecipientCount: &existing}
		setRecipientCountIfNeeded(input, 5)
		assert.Check(t, cmp.Equal(10, lo.FromPtrOr(input.RecipientCount, 0)))
	})
}

// TestFilterValidTargetsWithCompact tests that lo.Compact correctly filters nil targets.
func TestFilterValidTargetsWithCompact(t *testing.T) {
	t.Run("removes nil targets", func(t *testing.T) {
		targets := []*generated.CreateCampaignTargetInput{
			{Email: "a@test.com"},
			nil,
			{Email: "b@test.com"},
			nil,
			nil,
			{Email: "c@test.com"},
		}

		valid := lo.Compact(targets)
		emails := lo.Map(valid, func(target *generated.CreateCampaignTargetInput, _ int) string {
			if target == nil {
				return ""
			}
			return target.Email
		})
		assert.Check(t, cmp.Equal(3, len(emails)))
		assert.Check(t, lo.Contains(emails, "a@test.com"))
		assert.Check(t, lo.Contains(emails, "b@test.com"))
		assert.Check(t, lo.Contains(emails, "c@test.com"))
	})

	t.Run("handles all nil targets", func(t *testing.T) {
		targets := []*generated.CreateCampaignTargetInput{nil, nil, nil}
		valid := lo.Compact(targets)
		assert.Check(t, cmp.Equal(0, len(valid)))
	})

	t.Run("handles empty slice", func(t *testing.T) {
		var targets []*generated.CreateCampaignTargetInput
		valid := lo.Compact(targets)
		assert.Check(t, cmp.Equal(0, len(valid)))
	})

	t.Run("handles no nil targets", func(t *testing.T) {
		targets := []*generated.CreateCampaignTargetInput{
			{Email: "a@test.com"},
			{Email: "b@test.com"},
		}

		valid := lo.Compact(targets)
		assert.Check(t, cmp.Equal(2, len(valid)))
	})
}
