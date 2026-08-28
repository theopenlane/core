//go:build test

package eventstest_test

import (
	"fmt"
	"strings"
	"testing"

	th "github.com/theopenlane/core/v2/internal/graphapi/testharness"

	"github.com/theopenlane/utils/ulids"
	"gotest.tools/v3/assert"
	is "gotest.tools/v3/assert/cmp"

	ent "github.com/theopenlane/core/v2/internal/ent/generated"
	"github.com/theopenlane/core/v2/internal/ent/hooks"
	"github.com/theopenlane/core/v2/internal/graphapi"
)

func uniqueSubscriberEmail(prefix string) string {
	return fmt.Sprintf("%s-%s@theopenlane.io", prefix, strings.ToLower(ulids.New().String()))
}

func TestSubscriberLinkListener(t *testing.T) {
	org := suite.SeedFreshMinimalOrgUsers(t, false)
	ownerCtx := org.Owner.UserCtx
	allowCtx := th.SetContext(ownerCtx, suite.Client.DB)

	setup, err := graphapi.SetupListenerRuntime(suite.GalaRuntime, hooks.SubscriberLinkListeners())
	assert.NilError(t, err)
	defer setup.Teardown()

	reload := func(t *testing.T, id string) *ent.Subscriber {
		t.Helper()

		sub, err := suite.Client.DB.Subscriber.Get(allowCtx, id)
		assert.NilError(t, err)

		return sub
	}

	t.Run("links matching contact case insensitively", func(t *testing.T) {
		email := uniqueSubscriberEmail("link-contact")
		linked := (&th.ContactBuilder{Client: suite.Client, Email: email}).MustNew(ownerCtx, t)

		sub := (&th.SubscriberBuilder{Client: suite.Client, Email: strings.ToUpper(email)}).MustNew(ownerCtx, t)

		th.WaitForGala(t, setup.Runtime)

		waitForCondition(t, func() bool {
			return reload(t, sub.ID).ContactID == linked.ID
		}, "subscriber should link to the matching contact")

		assert.Check(t, is.Equal("", reload(t, sub.ID).UserID))
	})

	t.Run("links matching org member user", func(t *testing.T) {
		sub := (&th.SubscriberBuilder{Client: suite.Client, Email: org.Member.UserInfo.Email}).MustNew(ownerCtx, t)

		th.WaitForGala(t, setup.Runtime)

		waitForCondition(t, func() bool {
			return reload(t, sub.ID).UserID == org.Member.ID
		}, "subscriber should link to the matching org member")

		assert.Check(t, is.Equal("", reload(t, sub.ID).ContactID))
	})

	t.Run("links contact and user together", func(t *testing.T) {
		adminEmail := org.Admin.UserInfo.Email
		linked := (&th.ContactBuilder{Client: suite.Client, Email: adminEmail}).MustNew(ownerCtx, t)

		sub := (&th.SubscriberBuilder{Client: suite.Client, Email: adminEmail}).MustNew(ownerCtx, t)

		th.WaitForGala(t, setup.Runtime)

		waitForCondition(t, func() bool {
			s := reload(t, sub.ID)
			return s.ContactID == linked.ID && s.UserID == org.Admin.ID
		}, "subscriber should link to both the contact and the org member")
	})

	t.Run("no match and cross org contact stay unlinked", func(t *testing.T) {
		noMatch := (&th.SubscriberBuilder{Client: suite.Client, Email: uniqueSubscriberEmail("no-match")}).MustNew(ownerCtx, t)

		crossEmail := uniqueSubscriberEmail("cross-org")
		(&th.ContactBuilder{Client: suite.Client, Email: crossEmail}).MustNew(th.SharedTestUser2.UserCtx, t)
		crossOrg := (&th.SubscriberBuilder{Client: suite.Client, Email: crossEmail}).MustNew(ownerCtx, t)

		// a linked sentinel created after the negative cases proves the queue drained past them
		sentinelEmail := uniqueSubscriberEmail("sentinel")
		sentinelContact := (&th.ContactBuilder{Client: suite.Client, Email: sentinelEmail}).MustNew(ownerCtx, t)
		sentinel := (&th.SubscriberBuilder{Client: suite.Client, Email: sentinelEmail}).MustNew(ownerCtx, t)

		th.WaitForGala(t, setup.Runtime)

		waitForCondition(t, func() bool {
			return reload(t, sentinel.ID).ContactID == sentinelContact.ID
		}, "sentinel subscriber should link once the queue drains")

		unlinked := reload(t, noMatch.ID)
		assert.Check(t, is.Equal("", unlinked.ContactID))
		assert.Check(t, is.Equal("", unlinked.UserID))

		unlinked = reload(t, crossOrg.ID)
		assert.Check(t, is.Equal("", unlinked.ContactID))
		assert.Check(t, is.Equal("", unlinked.UserID))
	})
}
