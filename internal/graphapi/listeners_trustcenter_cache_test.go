//go:build test

package graphapi_test

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"

	"gotest.tools/v3/assert"
	is "gotest.tools/v3/assert/cmp"

	"github.com/theopenlane/entx"

	"github.com/theopenlane/core/common/enums"
	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/generated/privacy"
	"github.com/theopenlane/core/internal/ent/hooks"
	"github.com/theopenlane/core/internal/graphapi"
	"github.com/theopenlane/core/internal/graphapi/testclient"
)

func TestTrustCenterCacheListeners(t *testing.T) {
	var refreshHits atomic.Int64

	refreshServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("fresh") == "1" {
			refreshHits.Add(1)
		}

		w.WriteHeader(http.StatusOK)
	}))
	t.Cleanup(refreshServer.Close)

	// the http scheme puts trustcenterurl.BuildURL in test mode, routing every refresh to the
	// local server so refreshes succeed instead of retrying against unresolvable hosts
	hooks.SetTrustCenterConfig(hooks.TrustCenterConfig{
		CnameTarget:              cnameTargetTest,
		DefaultTrustCenterDomain: strings.TrimPrefix(refreshServer.URL, "http://"),
		CacheRefreshScheme:       "http",
	})
	t.Cleanup(func() {
		hooks.SetTrustCenterConfig(hooks.TrustCenterConfig{
			CnameTarget:              cnameTargetTest,
			DefaultTrustCenterDomain: defaultDomainTest,
		})
	})

	tcOrg := createFreshOrgWithTrustCenter(t)
	trustCenter := tcOrg.trustCenter
	dbCtx := privacy.DecisionContext(setContext(tcOrg.owner.UserCtx, suite.client.db), privacy.Allow)

	// the runtime is created after seeding so only the mutations under test dispatch
	setup, err := graphapi.SetupListenerRuntime(suite.galaRuntime, hooks.TrustCenterCacheListeners())
	assert.NilError(t, err)
	t.Cleanup(setup.Teardown)

	refreshDelta := func(t *testing.T, mutate func()) int64 {
		t.Helper()

		before := refreshHits.Load()
		mutate()
		waitForGala(t, setup.Runtime)

		return refreshHits.Load() - before
	}

	var privateDoc, publicDoc *generated.TrustCenterDoc

	t.Run("not visible doc create skips refresh", func(t *testing.T) {
		delta := refreshDelta(t, func() {
			privateDoc = (&TrustCenterDocBuilder{
				client:        suite.client,
				TrustCenterID: trustCenter.ID,
				Visibility:    enums.TrustCenterDocumentVisibilityNotVisible,
			}).MustNew(tcOrg.owner.UserCtx, t)
		})

		assert.Check(t, is.Equal(enums.TrustCenterDocumentVisibilityNotVisible, privateDoc.Visibility))
		assert.Check(t, is.Equal(int64(0), delta))
	})

	t.Run("doc update without visibility change skips refresh", func(t *testing.T) {
		delta := refreshDelta(t, func() {
			err := suite.client.db.TrustCenterDoc.UpdateOneID(privateDoc.ID).
				SetTitle("Cache Listener Private Doc Renamed").
				Exec(dbCtx)
			assert.NilError(t, err)
		})

		assert.Check(t, is.Equal(int64(0), delta))
	})

	t.Run("publicly visible doc create refreshes", func(t *testing.T) {
		delta := refreshDelta(t, func() {
			publicDoc = (&TrustCenterDocBuilder{
				client:        suite.client,
				TrustCenterID: trustCenter.ID,
				Visibility:    enums.TrustCenterDocumentVisibilityPubliclyVisible,
			}).MustNew(tcOrg.owner.UserCtx, t)
		})

		assert.Check(t, is.Equal(int64(1), delta))
	})

	t.Run("doc visibility change refreshes", func(t *testing.T) {
		// the API path pre-caches the mutation's old row, which the post-mutation
		// visibility-tuple hook requires; a bare ent update trips ent's old-value guard
		delta := refreshDelta(t, func() {
			_, err := suite.client.api.UpdateTrustCenterDoc(tcOrg.owner.UserCtx, publicDoc.ID, testclient.UpdateTrustCenterDocInput{
				Visibility: &enums.TrustCenterDocumentVisibilityProtected,
			}, nil, nil)
			assert.NilError(t, err)
		})

		assert.Check(t, is.Equal(int64(1), delta))
	})

	t.Run("soft deleted doc refreshes", func(t *testing.T) {
		delta := refreshDelta(t, func() {
			err := suite.client.db.TrustCenterDoc.DeleteOneID(privateDoc.ID).Exec(dbCtx)
			assert.NilError(t, err)
		})

		assert.Check(t, is.Equal(int64(1), delta))
	})

	t.Run("note create linked to trust center refreshes", func(t *testing.T) {
		var note *generated.Note

		delta := refreshDelta(t, func() {
			note = (&NoteBuilder{
				client:        suite.client,
				TrustCenterID: trustCenter.ID,
			}).MustNew(tcOrg.owner.UserCtx, t)
		})

		assert.Check(t, is.Equal(int64(1), delta))

		delta = refreshDelta(t, func() {
			err := suite.client.db.Note.UpdateOneID(note.ID).
				SetText("cache listener note update without linkage change").
				Exec(dbCtx)
			assert.NilError(t, err)
		})

		assert.Check(t, is.Equal(int64(0), delta))
	})

	t.Run("trust center entity create refreshes", func(t *testing.T) {
		delta := refreshDelta(t, func() {
			(&TrustCenterEntityBuilder{
				client:        suite.client,
				TrustCenterID: trustCenter.ID,
			}).MustNew(tcOrg.owner.UserCtx, t)
		})

		assert.Check(t, is.Equal(int64(1), delta))
	})

	t.Run("subprocessor mutations refresh linked trust centers", func(t *testing.T) {
		var sub *generated.Subprocessor

		delta := refreshDelta(t, func() {
			sub = (&SubprocessorBuilder{client: suite.client}).MustNew(tcOrg.owner.UserCtx, t)
		})

		assert.Check(t, is.Equal(int64(0), delta))

		delta = refreshDelta(t, func() {
			err := suite.client.db.TrustCenterSubprocessor.Create().
				SetTrustCenterID(trustCenter.ID).
				SetSubprocessorID(sub.ID).
				SetCountries([]string{"US"}).
				Exec(dbCtx)
			assert.NilError(t, err)
		})

		assert.Check(t, is.Equal(int64(1), delta))

		delta = refreshDelta(t, func() {
			err := suite.client.db.Subprocessor.UpdateOneID(sub.ID).
				SetName("cache-listener-renamed-vendor").
				Exec(dbCtx)
			assert.NilError(t, err)
		})

		assert.Check(t, is.Equal(int64(1), delta))

		delta = refreshDelta(t, func() {
			err := suite.client.db.Subprocessor.UpdateOneID(sub.ID).
				SetDescription("update on a field the listener does not gate on").
				Exec(dbCtx)
			assert.NilError(t, err)
		})

		assert.Check(t, is.Equal(int64(0), delta))
	})

	t.Run("soft deleted trust center refreshes via skip soft delete load", func(t *testing.T) {
		// cascaded child soft deletes emit their own events and each resolves back to the
		// still-loadable trust center, so the delta is at least one
		delta := refreshDelta(t, func() {
			err := suite.client.db.TrustCenter.DeleteOneID(trustCenter.ID).Exec(dbCtx)
			assert.NilError(t, err)
		})

		assert.Check(t, delta >= 1, "expected at least one refresh, got %d", delta)
	})

	t.Run("hard deleted trust center acks without refresh", func(t *testing.T) {
		delta := refreshDelta(t, func() {
			err := suite.client.db.TrustCenter.DeleteOneID(trustCenter.ID).Exec(entx.SkipSoftDelete(dbCtx))
			assert.NilError(t, err)
		})

		assert.Check(t, is.Equal(int64(0), delta))
	})

	cleanupOrganizationDataWithContext(tcOrg.owner.UserCtx, t)
}
