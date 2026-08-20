//go:build test

package graphapi_test

import (
	"context"
	"strings"
	"sync"
	"sync/atomic"
	"testing"

	"github.com/stretchr/testify/mock"
	"github.com/theopenlane/entx"
	"github.com/theopenlane/utils/ulids"
	"gotest.tools/v3/assert"
	is "gotest.tools/v3/assert/cmp"

	"github.com/theopenlane/core/internal/ent/generated/organization"
	"github.com/theopenlane/core/internal/ent/generated/organizationsetting"
	"github.com/theopenlane/core/internal/ent/generated/privacy"
)

var (
	stripeSubscriptionCancelCalls atomic.Int64
	stripeSubscriptionCancelOnce  sync.Once
)

// ensureStripeSubscriptionCancelMock registers the one stripe backend expectation the suite
// setup omits so FindAndDeactivateCustomerSubscription completes against the mock instead of
// parking a retrying job on the shared runtime
func ensureStripeSubscriptionCancelMock() {
	stripeSubscriptionCancelOnce.Do(func() {
		suite.stripeMockBackend.On("Call", mock.Anything, mock.Anything, mock.Anything, mock.AnythingOfType("*stripe.SubscriptionCancelParams"), mock.AnythingOfType("*stripe.Subscription")).
			Run(func(mock.Arguments) { stripeSubscriptionCancelCalls.Add(1) }).
			Return(nil)
	})
}

// fakeStripeCustomerID mints a unique customer id so a stamped org never collides with the
// suite's shared cus_test_customer unique constraint
func fakeStripeCustomerID() string {
	return "cus_listener_" + strings.ToLower(ulids.New().String())
}

func TestEntitlementListenerOrganizationCreated(t *testing.T) {
	user := suite.userBuilder(context.Background(), t)
	allowCtx := privacy.DecisionContext(setContext(user.UserCtx, suite.client.db), privacy.Allow)

	// the suite's search mock resolves every org to the one shared cus_test_customer, which
	// the first seeded org claims; reconcile for every later org hits the unique constraint
	// and must terminally ack — WaitForEvents returning proves no parked retry
	suite.WaitForEvents()

	org, err := suite.client.db.Organization.Get(allowCtx, user.OrganizationID)
	assert.NilError(t, err)
	assert.Check(t, org.StripeCustomerID == nil)

	personalOrg, err := suite.client.db.Organization.Get(allowCtx, user.PersonalOrgID)
	assert.NilError(t, err)
	assert.Check(t, personalOrg.StripeCustomerID == nil)
}

func TestEntitlementListenerOrganizationDeleted(t *testing.T) {
	ensureStripeSubscriptionCancelMock()

	user := suite.userBuilder(context.Background(), t)
	allowCtx := privacy.DecisionContext(setContext(user.UserCtx, suite.client.db), privacy.Allow)

	suite.WaitForEvents()

	customerID := fakeStripeCustomerID()
	assert.NilError(t, suite.client.db.Organization.UpdateOneID(user.OrganizationID).SetStripeCustomerID(customerID).Exec(allowCtx))

	before := stripeSubscriptionCancelCalls.Load()

	resp, err := suite.client.api.DeleteOrganization(user.UserCtx, user.OrganizationID)
	assert.NilError(t, err)
	assert.Check(t, is.Equal(user.OrganizationID, resp.DeleteOrganization.DeletedID))

	waitForCondition(t, func() bool {
		return stripeSubscriptionCancelCalls.Load() > before
	}, "organization delete should deactivate the stripe subscription")

	purgedCtx := entx.SkipSoftDelete(allowCtx)

	waitForCondition(t, func() bool {
		exists, err := suite.client.db.Organization.Query().Where(organization.ID(user.OrganizationID)).Exist(purgedCtx)
		return err == nil && !exists
	}, "cascade should still purge the organization")
}

func TestEntitlementListenerOrganizationDeletedWithoutCustomer(t *testing.T) {
	user := suite.userBuilder(context.Background(), t)
	allowCtx := privacy.DecisionContext(setContext(user.UserCtx, suite.client.db), privacy.Allow)

	suite.WaitForEvents()

	assert.NilError(t, suite.client.db.Organization.UpdateOneID(user.OrganizationID).ClearStripeCustomerID().Exec(allowCtx))

	resp, err := suite.client.api.DeleteOrganization(user.UserCtx, user.OrganizationID)
	assert.NilError(t, err)
	assert.Check(t, is.Equal(user.OrganizationID, resp.DeleteOrganization.DeletedID))

	suite.WaitForEvents()

	purgedCtx := entx.SkipSoftDelete(allowCtx)

	waitForCondition(t, func() bool {
		exists, err := suite.client.db.Organization.Query().Where(organization.ID(user.OrganizationID)).Exist(purgedCtx)
		return err == nil && !exists
	}, "delete without a stripe customer should skip deactivation and still purge the organization")
}

func TestEntitlementListenerBillingUpdate(t *testing.T) {
	user := suite.userBuilder(context.Background(), t)
	allowCtx := privacy.DecisionContext(setContext(user.UserCtx, suite.client.db), privacy.Allow)

	suite.WaitForEvents()

	customerID := fakeStripeCustomerID()
	assert.NilError(t, suite.client.db.Organization.UpdateOneID(user.OrganizationID).SetStripeCustomerID(customerID).Exec(allowCtx))

	setting, err := suite.client.db.OrganizationSetting.Query().
		Where(organizationsetting.OrganizationID(user.OrganizationID)).
		Only(allowCtx)
	assert.NilError(t, err)

	assert.NilError(t, suite.client.db.OrganizationSetting.UpdateOneID(setting.ID).SetBillingEmail("billing@theopenlane.io").Exec(allowCtx))

	// the billing listener pushes the change through the mocked CustomerUpdate then
	// reconciles; the drain proves neither call parked a retrying job
	suite.WaitForEvents()

	org, err := suite.client.db.Organization.Get(allowCtx, user.OrganizationID)
	assert.NilError(t, err)
	assert.Assert(t, org.StripeCustomerID != nil)
	assert.Check(t, is.Equal(customerID, *org.StripeCustomerID))
}
