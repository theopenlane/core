package graphapi_test

import (
	"testing"

	"github.com/99designs/gqlgen/graphql"
	"github.com/samber/lo"
	"gotest.tools/v3/assert"
	is "gotest.tools/v3/assert/cmp"

	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/generated/privacy"
	"github.com/theopenlane/core/internal/ent/generated/trustcenterwatermarkconfig"
	"github.com/theopenlane/core/internal/graphapi/testclient"
	"github.com/theopenlane/core/pkg/objects/storage"
)

// TestTrustCenterDocGlobalConfigOverride tests the global watermark config override logic
func TestTrustCenterDocGlobalConfigOverride(t *testing.T) {
	trustCenter := (&TrustCenterBuilder{client: suite.client}).MustNew(testUser1.UserCtx, t)

	dbCtx := setContext(testUser1.UserCtx, suite.client.db)
	allowCtx := privacy.DecisionContext(dbCtx, privacy.Allow)

	watermarkConfig, err := suite.client.db.TrustCenterWatermarkConfig.Query().
		Where(trustcenterwatermarkconfig.TrustCenterID(trustCenter.ID)).
		Only(allowCtx)
	assert.NilError(t, err)

	createPDFUpload := func() *graphql.Upload {
		pdfFile, err := storage.NewUploadFile("testdata/uploads/hello.pdf")
		assert.NilError(t, err)
		return &graphql.Upload{
			File:        pdfFile.RawFile,
			Filename:    pdfFile.OriginalName,
			Size:        pdfFile.Size,
			ContentType: pdfFile.ContentType,
		}
	}

	t.Run("when global config is false, documents can be set to false or true", func(t *testing.T) {
		_, err := suite.client.db.TrustCenterWatermarkConfig.UpdateOne(watermarkConfig).
			SetIsEnabled(false).
			Save(allowCtx)
		assert.NilError(t, err)

		file := createPDFUpload()
		expectUpload(t, suite.client.mockProvider, []graphql.Upload{*file})

		createInput := testclient.CreateTrustCenterDocInput{
			Title:               "Test Document",
			Category:            "Policy",
			TrustCenterID:       &trustCenter.ID,
			WatermarkingEnabled: lo.ToPtr(true),
		}

		createResp, err := suite.client.api.CreateTrustCenterDoc(testUser1.UserCtx, createInput, *file)
		assert.NilError(t, err)
		assert.Assert(t, createResp != nil)

		docID := createResp.CreateTrustCenterDoc.TrustCenterDoc.ID
		dbDoc, err := suite.client.db.TrustCenterDoc.Get(dbCtx, docID)
		assert.NilError(t, err)
		assert.Check(t, is.Equal(true, dbDoc.WatermarkingEnabled))

		updateInput := testclient.UpdateTrustCenterDocInput{
			WatermarkingEnabled: lo.ToPtr(false),
		}

		updateResp, err := suite.client.api.UpdateTrustCenterDoc(testUser1.UserCtx, docID, updateInput, nil, nil)
		assert.NilError(t, err)
		assert.Assert(t, updateResp != nil)

		dbDoc, err = suite.client.db.TrustCenterDoc.Get(dbCtx, docID)
		assert.NilError(t, err)
		assert.Check(t, is.Equal(false, dbDoc.WatermarkingEnabled), "with global config false, document watermarking can be disabled")

		(&Cleanup[*generated.TrustCenterDocDeleteOne]{client: suite.client.db.TrustCenterDoc, ID: docID}).MustDelete(testUser1.UserCtx, t)
	})

	t.Run("when global config is true, documents cannot be set to false", func(t *testing.T) {
		_, err := suite.client.db.TrustCenterWatermarkConfig.UpdateOne(watermarkConfig).
			SetIsEnabled(true).
			Save(allowCtx)
		assert.NilError(t, err)

		file := createPDFUpload()
		expectUpload(t, suite.client.mockProvider, []graphql.Upload{*file})

		createInput := testclient.CreateTrustCenterDocInput{
			Title:               "Test Document with Global Config Enabled",
			Category:            "Policy",
			TrustCenterID:       &trustCenter.ID,
			WatermarkingEnabled: lo.ToPtr(true),
		}

		createResp, err := suite.client.api.CreateTrustCenterDoc(testUser1.UserCtx, createInput, *file)
		assert.NilError(t, err)
		assert.Assert(t, createResp != nil)

		docID := createResp.CreateTrustCenterDoc.TrustCenterDoc.ID
		dbDoc, err := suite.client.db.TrustCenterDoc.Get(dbCtx, docID)
		assert.NilError(t, err)
		assert.Check(t, is.Equal(true, dbDoc.WatermarkingEnabled))

		updateInput := testclient.UpdateTrustCenterDocInput{
			WatermarkingEnabled: lo.ToPtr(false),
		}

		updateResp, err := suite.client.api.UpdateTrustCenterDoc(testUser1.UserCtx, docID, updateInput, nil, nil)
		assert.NilError(t, err)
		assert.Assert(t, updateResp != nil)

		dbDoc, err = suite.client.db.TrustCenterDoc.Get(dbCtx, docID)
		assert.NilError(t, err)
		assert.Check(t, is.Equal(true, dbDoc.WatermarkingEnabled), "with global config enabled, document watermarking is forced to true")

		(&Cleanup[*generated.TrustCenterDocDeleteOne]{client: suite.client.db.TrustCenterDoc, ID: docID}).MustDelete(testUser1.UserCtx, t)
	})

	t.Run("when global config is true, documents can still be set to true", func(t *testing.T) {
		_, err := suite.client.db.TrustCenterWatermarkConfig.UpdateOne(watermarkConfig).
			SetIsEnabled(true).
			Save(allowCtx)
		assert.NilError(t, err)

		file := createPDFUpload()
		expectUpload(t, suite.client.mockProvider, []graphql.Upload{*file})

		createInput := testclient.CreateTrustCenterDocInput{
			Title:               "Test Document Explicitly True",
			Category:            "Policy",
			TrustCenterID:       &trustCenter.ID,
			WatermarkingEnabled: lo.ToPtr(false),
		}

		createResp, err := suite.client.api.CreateTrustCenterDoc(testUser1.UserCtx, createInput, *file)
		assert.NilError(t, err)
		assert.Assert(t, createResp != nil)

		docID := createResp.CreateTrustCenterDoc.TrustCenterDoc.ID
		dbDoc, err := suite.client.db.TrustCenterDoc.Get(dbCtx, docID)
		assert.NilError(t, err)
		assert.Check(t, is.Equal(true, dbDoc.WatermarkingEnabled), "global config overrides document-level false on create")

		updateInput := testclient.UpdateTrustCenterDocInput{
			WatermarkingEnabled: lo.ToPtr(true),
		}

		updateResp, err := suite.client.api.UpdateTrustCenterDoc(testUser1.UserCtx, docID, updateInput, nil, nil)
		assert.NilError(t, err)
		assert.Assert(t, updateResp != nil)

		dbDoc, err = suite.client.db.TrustCenterDoc.Get(dbCtx, docID)
		assert.NilError(t, err)
		assert.Check(t, is.Equal(true, dbDoc.WatermarkingEnabled), "document watermarking can be set to true when global config is true")

		(&Cleanup[*generated.TrustCenterDocDeleteOne]{client: suite.client.db.TrustCenterDoc, ID: docID}).MustDelete(testUser1.UserCtx, t)
	})

	(&Cleanup[*generated.TrustCenterWatermarkConfigDeleteOne]{client: suite.client.db.TrustCenterWatermarkConfig, ID: watermarkConfig.ID}).MustDelete(testUser1.UserCtx, t)
	(&Cleanup[*generated.TrustCenterDeleteOne]{client: suite.client.db.TrustCenter, ID: trustCenter.ID}).MustDelete(testUser1.UserCtx, t)
}
