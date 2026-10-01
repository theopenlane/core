//go:build test

package hooks_test

import (
	"strings"

	"github.com/brianvoe/gofakeit/v7"
	"github.com/samber/lo"
	"github.com/theopenlane/iam/auth"
	"gotest.tools/v3/assert"
	is "gotest.tools/v3/assert/cmp"

	"github.com/theopenlane/core/v2/internal/ent/generated"
	"github.com/theopenlane/core/v2/internal/ent/generated/privacy"
	"github.com/theopenlane/core/v2/internal/ent/generated/tagdefinition"
)

func (suite *HookTestSuite) TestHookTags_SlugCollisionInTransaction() {
	t := suite.T()

	user := suite.seedUser()
	orgID := user.Edges.OrgMemberships[0].OrganizationID

	ctx := generated.NewContext(auth.NewTestContextWithOrgID(user.ID, orgID), suite.client)
	ctx = privacy.DecisionContext(ctx, privacy.Allow)

	suffix := strings.ToLower(gofakeit.LetterN(6))
	existingTag := "aws-marketplace-" + suffix
	collidingTag := "AWS Marketplace " + suffix
	newTag := "conditional approval " + suffix

	_, err := suite.client.TagDefinition.Create().
		SetName(existingTag).
		SetOwnerID(orgID).
		Save(ctx)
	assert.NilError(t, err)

	tx, err := suite.client.Tx(ctx)
	assert.NilError(t, err)

	task, err := tx.Task.Create().
		SetTitle(gofakeit.AppName()).
		SetOwnerID(orgID).
		SetTags([]string{collidingTag, newTag}).
		Save(ctx)
	assert.NilError(t, err)
	assert.NilError(t, tx.Commit())

	t.Cleanup(func() {
		_ = suite.client.Task.DeleteOneID(task.ID).Exec(ctx)
		_, _ = suite.client.TagDefinition.Delete().Where(tagdefinition.OwnerID(orgID)).Exec(ctx)
	})

	tags, err := suite.client.TagDefinition.Query().
		Where(tagdefinition.OwnerID(orgID)).
		All(ctx)
	assert.NilError(t, err)

	names := lo.Map(tags, func(tag *generated.TagDefinition, _ int) string { return tag.Name })

	assert.Check(t, is.Len(tags, 2))
	assert.Check(t, is.Contains(names, existingTag))
	assert.Check(t, is.Contains(names, strings.ToLower(newTag)))
}
