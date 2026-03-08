package graphapi_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/theopenlane/core/internal/ent/generated/privacy"
	"github.com/theopenlane/core/internal/httpserve/authmanager"
	"github.com/theopenlane/iam/auth"
	"github.com/theopenlane/utils/ulids"
	"gotest.tools/v3/assert"
)

// cleanupTrustCenterData removes all trust centers and watermark configs for the test user's organization.
// This ensures the Only() query in hooks works correctly when tests expect a single watermark config.
func cleanupTrustCenterData(t *testing.T) {
	t.Helper()
	ctx := privacy.DecisionContext(setContext(testUser1.UserCtx, suite.client.db), privacy.Allow)

	wcs, err := suite.client.db.TrustCenterWatermarkConfig.Query().All(ctx)
	assert.NilError(t, err)
	for _, wc := range wcs {
		_ = suite.client.db.TrustCenterWatermarkConfig.DeleteOneID(wc.ID).Exec(ctx)
	}

	tcs, err := suite.client.db.TrustCenter.Query().All(ctx)
	assert.NilError(t, err)
	for _, tc := range tcs {
		_ = suite.client.db.TrustCenter.DeleteOneID(tc.ID).Exec(ctx)
	}
}

// cleanupWatermarkConfigs removes all watermark configs for the test user's organization.
func cleanupWatermarkConfigs(t *testing.T) {
	t.Helper()
	ctx := privacy.DecisionContext(setContext(testUser1.UserCtx, suite.client.db), privacy.Allow)

	wcs, _ := suite.client.db.TrustCenterWatermarkConfig.Query().All(ctx)
	for _, wc := range wcs {
		_ = suite.client.db.TrustCenterWatermarkConfig.DeleteOneID(wc.ID).Exec(ctx)
	}
}

// createAnonymousTrustCenterContext creates a context for an anonymous trust center user
func createAnonymousTrustCenterContext(trustCenterID, organizationID string) context.Context {
	anonUserID := fmt.Sprintf("%s%s", authmanager.AnonTrustCenterJWTPrefix, ulids.New().String())
	ctx := auth.WithCaller(context.Background(), auth.NewTrustCenterCaller(organizationID, anonUserID, "Anonymous User", ""))
	return auth.ActiveTrustCenterIDKey.Set(ctx, trustCenterID)
}

// createAnonymousTrustCenterContextWithEmail creates a context for an anonymous trust center user with subject email
func createAnonymousTrustCenterContextWithEmail(trustCenterID, organizationID, email string) (context.Context, *auth.Caller) {
	anonUserID := fmt.Sprintf("%s%s", authmanager.AnonTrustCenterJWTPrefix, ulids.New().String())
	caller := auth.NewTrustCenterCaller(organizationID, anonUserID, "Anonymous User", email)
	ctx := auth.WithCaller(context.Background(), caller)
	ctx = auth.ActiveTrustCenterIDKey.Set(ctx, trustCenterID)
	return ctx, caller
}
