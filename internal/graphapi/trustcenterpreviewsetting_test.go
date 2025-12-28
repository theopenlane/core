package graphapi_test

import (
	"testing"

	"gotest.tools/v3/assert"
	is "gotest.tools/v3/assert/cmp"

	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/generated/trustcenter"
	"github.com/theopenlane/core/internal/graphapi/testclient"
)

const updateTrustCenterPreviewSettingDocument = `mutation UpdateTrustCenterPreviewSetting ($input: UpdateTrustCenterSettingInput!) {
	updateTrustCenterPreviewSetting(input: $input) {
		trustCenterSetting {
			id
		}
	}
}`

type updateTrustCenterPreviewSettingResponse struct {
	UpdateTrustCenterPreviewSetting struct {
		TrustCenterSetting struct {
			ID string `json:"id"`
		} `json:"trustCenterSetting"`
	} `json:"updateTrustCenterPreviewSetting"`
}

func TestCreateTrustCenterPreviewSettingLinksTrustCenter(t *testing.T) {
	trustCenter := (&TrustCenterBuilder{client: suite.client}).MustNew(testUser1.UserCtx, t)

	resp, err := suite.client.api.CreateTrustCenterPreviewSetting(testUser1.UserCtx, testclient.CreateTrustCenterPreviewSettingInput{
		TrustCenterID: trustCenter.ID,
	})
	assert.NilError(t, err)
	assert.Assert(t, resp != nil)

	previewSettingID := resp.CreateTrustCenterPreviewSetting.TrustCenterSetting.ID
	assert.Check(t, previewSettingID != "")

	dbCtx := setContext(testUser1.UserCtx, suite.client.db)
	updatedTrustCenter, err := suite.client.db.TrustCenter.Query().
		Where(trustcenter.IDEQ(trustCenter.ID)).
		WithPreviewSetting().
		Only(dbCtx)
	assert.NilError(t, err)
	assert.Assert(t, updatedTrustCenter.Edges.PreviewSetting != nil)
	assert.Check(t, is.Equal(previewSettingID, updatedTrustCenter.Edges.PreviewSetting.ID))

	(&Cleanup[*generated.TrustCenterDeleteOne]{client: suite.client.db.TrustCenter, ID: trustCenter.ID}).MustDelete(testUser1.UserCtx, t)
}

func TestUpdateTrustCenterPreviewSettingCreatesAndLinksSetting(t *testing.T) {
	trustCenter := (&TrustCenterBuilder{client: suite.client}).MustNew(testUser1.UserCtx, t)

	dbCtx := setContext(testUser1.UserCtx, suite.client.db)
	tcWithPreview, err := suite.client.db.TrustCenter.Query().
		Where(trustcenter.IDEQ(trustCenter.ID)).
		WithPreviewSetting().
		Only(dbCtx)
	assert.NilError(t, err)

	if tcWithPreview.Edges.PreviewSetting != nil {
		err = suite.client.db.TrustCenterSetting.DeleteOneID(tcWithPreview.Edges.PreviewSetting.ID).Exec(dbCtx)
		assert.NilError(t, err)
	}

	_, err = suite.client.db.TrustCenter.UpdateOneID(trustCenter.ID).
		ClearPreviewSetting().
		Save(dbCtx)
	assert.NilError(t, err)

	graphClient, ok := suite.client.api.TestGraphClient.(*testclient.Client)
	assert.Assert(t, ok, "unexpected test graph client type")

	var gqlResp updateTrustCenterPreviewSettingResponse
	err = graphClient.Client.Post(
		testUser1.UserCtx,
		"UpdateTrustCenterPreviewSetting",
		updateTrustCenterPreviewSettingDocument,
		&gqlResp,
		map[string]any{
			"input": testclient.UpdateTrustCenterSettingInput{},
		},
	)
	assert.NilError(t, err)
	assert.Check(t, gqlResp.UpdateTrustCenterPreviewSetting.TrustCenterSetting.ID != "")

	updatedTrustCenter, err := suite.client.db.TrustCenter.Query().
		Where(trustcenter.IDEQ(trustCenter.ID)).
		WithPreviewSetting().
		Only(dbCtx)
	assert.NilError(t, err)
	assert.Assert(t, updatedTrustCenter.Edges.PreviewSetting != nil)
	assert.Check(t, is.Equal(gqlResp.UpdateTrustCenterPreviewSetting.TrustCenterSetting.ID, updatedTrustCenter.Edges.PreviewSetting.ID))

	(&Cleanup[*generated.TrustCenterDeleteOne]{client: suite.client.db.TrustCenter, ID: trustCenter.ID}).MustDelete(testUser1.UserCtx, t)
}
