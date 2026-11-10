//go:build test

package hooks_test

import (
	"strings"
	"testing"

	"github.com/brianvoe/gofakeit/v7"
	"github.com/samber/lo"
	"github.com/theopenlane/iam/auth"
	"gotest.tools/v3/assert"
	is "gotest.tools/v3/assert/cmp"

	"github.com/theopenlane/core/v2/internal/ent/generated"
	"github.com/theopenlane/core/v2/internal/ent/generated/tagdefinition"
)

func TestHookTags_SlugCollisionInTransaction(t *testing.T) {
	s := suite.setupTest(t)

	user := s.seedUser()
	orgID := user.Edges.OrgMemberships[0].OrganizationID

	ctx := generated.NewContext(auth.NewTestContextWithOrgID(user.ID, orgID), s.client)
	ctx = auth.WithInternalOperationContext(ctx)

	suffix := strings.ToLower(gofakeit.LetterN(6))
	existingTag := "aws-marketplace-" + suffix
	collidingTag := "AWS Marketplace " + suffix
	newTag := "conditional approval " + suffix

	_, err := s.client.TagDefinition.Create().
		SetName(existingTag).
		SetOwnerID(orgID).
		Save(ctx)
	assert.NilError(t, err)

	tx, err := s.client.Tx(ctx)
	assert.NilError(t, err)

	task, err := tx.Task.Create().
		SetTitle(gofakeit.AppName()).
		SetOwnerID(orgID).
		SetTags([]string{collidingTag, newTag}).
		Save(ctx)
	assert.NilError(t, err)
	assert.NilError(t, tx.Commit())

	t.Cleanup(func() {
		_ = s.client.Task.DeleteOneID(task.ID).Exec(ctx)
		_, _ = s.client.TagDefinition.Delete().Where(tagdefinition.OwnerID(orgID)).Exec(ctx)
	})

	tags, err := s.client.TagDefinition.Query().
		Where(tagdefinition.OwnerID(orgID)).
		All(ctx)
	assert.NilError(t, err)

	names := lo.Map(tags, func(tag *generated.TagDefinition, _ int) string { return tag.Name })

	assert.Check(t, is.Len(tags, 2))
	assert.Check(t, is.Contains(names, existingTag))
	assert.Check(t, is.Contains(names, strings.ToLower(newTag)))
}
