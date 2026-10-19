//go:build test

package microsoft

import (
	"context"
	"net/http"
	"strings"
	"testing"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore/policy"
	"gotest.tools/v3/assert"
)

func TestMockProviderIssuesToken(t *testing.T) {
	mock := NewMockMicrosoftProvider()
	t.Cleanup(mock.Close)

	cred, err := NewTokenCredential(mock.AppConfig(), MockTenantID)
	assert.NilError(t, err)

	token, err := cred.GetToken(context.Background(), policy.TokenRequestOptions{Scopes: []string{GraphScope}})
	assert.NilError(t, err)
	assert.Check(t, token.Token == MockAccessToken)
	assert.Check(t, len(mock.TokenTenants()) == 1 && mock.TokenTenants()[0] == MockTenantID)
}

func TestMockProviderTokenFailure(t *testing.T) {
	mock := NewMockMicrosoftProvider()
	t.Cleanup(mock.Close)

	mock.SetTokenStatus(http.StatusUnauthorized)

	cred, err := NewTokenCredential(mock.AppConfig(), MockTenantID)
	assert.NilError(t, err)

	_, err = cred.GetToken(context.Background(), policy.TokenRequestOptions{Scopes: []string{GraphScope}})
	assert.Assert(t, err != nil)
}

func TestListSubscriptionIDsPagesAndFiltersStates(t *testing.T) {
	mock := NewMockMicrosoftProvider()
	t.Cleanup(mock.Close)

	mock.SetSubscriptionPageSize(2)
	mock.SetSubscriptions(
		MockSubscription{ID: "sub-1", State: "Enabled"},
		MockSubscription{ID: "sub-2", State: "Disabled"},
		MockSubscription{ID: "sub-3", State: "Warned"},
		MockSubscription{ID: "sub-4", State: "Deleted"},
		MockSubscription{ID: "sub-5", State: "PastDue"},
	)

	cfg := mock.AppConfig()

	cred, err := NewTokenCredential(cfg, MockTenantID)
	assert.NilError(t, err)

	client, err := NewARMClient(cfg, cred)
	assert.NilError(t, err)

	ids, err := client.ListSubscriptionIDs(context.Background(), MockTenantID)
	assert.NilError(t, err)
	assert.Check(t, len(ids) == 3)
	assert.Check(t, ids[0] == "sub-1" && ids[1] == "sub-3" && ids[2] == "sub-5")
}

func TestListSubscriptionIDsFiltersTenant(t *testing.T) {
	mock := NewMockMicrosoftProvider()
	t.Cleanup(mock.Close)

	otherTenant := "00000000-0000-0000-0000-000000000002"

	mock.SetSubscriptions(
		MockSubscription{ID: "sub-1", State: "Enabled"},
		MockSubscription{ID: "sub-2", TenantID: otherTenant, State: "Enabled"},
		MockSubscription{ID: "sub-3", TenantID: strings.ToUpper(MockTenantID), State: "Enabled"},
	)

	cfg := mock.AppConfig()

	cred, err := NewTokenCredential(cfg, MockTenantID)
	assert.NilError(t, err)

	client, err := NewARMClient(cfg, cred)
	assert.NilError(t, err)

	ids, err := client.ListSubscriptionIDs(context.Background(), MockTenantID)
	assert.NilError(t, err)
	assert.Check(t, len(ids) == 2)
	assert.Check(t, ids[0] == "sub-1" && ids[1] == "sub-3")

	ids, err = client.ListSubscriptionIDs(context.Background(), otherTenant)
	assert.NilError(t, err)
	assert.Check(t, len(ids) == 1 && ids[0] == "sub-2")

	_, err = client.ListSubscriptionIDs(context.Background(), "")
	assert.ErrorIs(t, err, ErrTenantIDRequired)
}

func TestListSubscriptionIDsFailure(t *testing.T) {
	mock := NewMockMicrosoftProvider()
	t.Cleanup(mock.Close)

	mock.SetSubscriptionStatus(http.StatusForbidden)

	cfg := mock.AppConfig()

	cred, err := NewTokenCredential(cfg, MockTenantID)
	assert.NilError(t, err)

	client, err := NewARMClient(cfg, cred)
	assert.NilError(t, err)

	_, err = client.ListSubscriptionIDs(context.Background(), MockTenantID)
	assert.ErrorIs(t, err, ErrSubscriptionListFailed)
}

func TestMockProviderRejectsWrongBearer(t *testing.T) {
	mock := NewMockMicrosoftProvider()
	t.Cleanup(mock.Close)

	resp, err := mock.Server.Client().Get(mock.Server.URL + "/subscriptions")
	assert.NilError(t, err)
	t.Cleanup(func() { _ = resp.Body.Close() })

	assert.Check(t, resp.StatusCode == http.StatusUnauthorized)
}

func TestGraphClientUsesMockProvider(t *testing.T) {
	mock := NewMockMicrosoftProvider()
	t.Cleanup(mock.Close)

	mock.HandleGraph(http.MethodGet, "/users", func(w http.ResponseWriter, _ *http.Request) {
		WriteJSON(w, http.StatusOK, map[string]any{"value": []map[string]any{{"id": "user-1", "displayName": "Mock User"}}})
	})

	cfg := mock.AppConfig()

	graph, err := cfg.NewGraphClient(MockTokenCredential{})
	assert.NilError(t, err)

	users, err := graph.Users().Get(context.Background(), nil)
	assert.NilError(t, err)
	assert.Check(t, len(users.GetValue()) == 1)
	assert.Check(t, *users.GetValue()[0].GetId() == "user-1")
}
