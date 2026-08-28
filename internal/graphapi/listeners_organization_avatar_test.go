//go:build test

package graphapi_test

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	"gotest.tools/v3/assert"
	is "gotest.tools/v3/assert/cmp"

	"github.com/theopenlane/httpsling"
	"github.com/theopenlane/utils/ulids"

	"github.com/theopenlane/core/v2/internal/ent/generated/organization"
	"github.com/theopenlane/core/v2/internal/ent/generated/privacy"
	"github.com/theopenlane/core/v2/internal/ent/hooks"
	"github.com/theopenlane/core/v2/internal/graphapi"
	"github.com/theopenlane/core/v2/internal/graphapi/testclient"
	"github.com/theopenlane/core/v2/pkg/urlx"
)

var errAvatarTestUnreachable = errors.New("avatar test transport unreachable")

func TestOrganizationAvatarListener(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/":
			_, _ = fmt.Fprint(w, `<html><head><link rel="icon" href="/logo.png" sizes="64x64"></head></html>`)
		case "/logo.png", "/favicon.ico":
			w.Header().Set("Content-Type", "image/png")
			_, _ = w.Write([]byte("test-image"))
		default:
			http.NotFound(w, r)
		}
	}))
	t.Cleanup(server.Close)

	serverURL, err := url.Parse(server.URL)
	assert.NilError(t, err)

	client := server.Client()
	client.Timeout = time.Second
	client.Transport = avatarTestTransport{
		base:   client.Transport,
		target: serverURL,
	}

	requester, err := urlx.NewRequester(httpsling.WithHTTPClient(client))
	assert.NilError(t, err)

	setup, err := graphapi.SetupListenerRuntime(suite.galaRuntime, hooks.OrganizationAvatarListeners(
		hooks.WithOrganizationAvatarRequester(requester),
	))
	assert.NilError(t, err)
	defer setup.Teardown()

	user := suite.userBuilder(context.Background(), t)
	allowCtx := privacy.DecisionContext(setContext(user.UserCtx, suite.client.db), privacy.Allow)

	t.Run("create without domains keeps default avatar", func(t *testing.T) {
		org := (&OrganizationBuilder{client: suite.client}).MustNew(user.UserCtx, t)
		assert.Assert(t, org.AvatarRemoteURL != nil)

		assert.NilError(t, setup.Runtime.WaitIdle(t.Context()))

		reloaded, err := suite.client.db.Organization.Query().
			Where(organization.IDEQ(org.ID)).
			WithSetting().
			Only(allowCtx)
		assert.NilError(t, err)
		assert.Assert(t, reloaded.Edges.Setting != nil)
		assert.Check(t, is.Len(reloaded.Edges.Setting.Domains, 0))
		assert.Assert(t, reloaded.AvatarRemoteURL != nil)
		assert.Check(t, is.Equal(*org.AvatarRemoteURL, *reloaded.AvatarRemoteURL))
	})

	t.Run("local avatar discovery updates the organization", func(t *testing.T) {
		domain := "avatar-" + strings.ToLower(ulids.New().String()) + ".test"

		resp, err := suite.client.api.CreateOrganization(user.UserCtx, testclient.CreateOrganizationInput{
			Name: "avatar-listener-" + ulids.New().String(),
			CreateOrgSettings: &testclient.CreateOrganizationSettingInput{
				Domains: []string{domain},
			},
		}, nil, nil)
		assert.NilError(t, err)

		created := resp.CreateOrganization.Organization
		assert.Assert(t, created.AvatarRemoteURL != nil)

		assert.NilError(t, setup.Runtime.WaitIdle(t.Context()))

		reloaded, err := suite.client.db.Organization.Query().
			Where(organization.IDEQ(created.ID)).
			WithSetting().
			Only(allowCtx)
		assert.NilError(t, err)
		assert.Assert(t, reloaded.Edges.Setting != nil)
		assert.Check(t, is.DeepEqual([]string{domain}, reloaded.Edges.Setting.Domains))
		assert.Assert(t, reloaded.AvatarRemoteURL != nil)
		assert.Check(t, is.Equal("https://"+domain+"/logo.png", *reloaded.AvatarRemoteURL))
	})

	t.Run("local transport failure acks without avatar update", func(t *testing.T) {
		domain := "unreachable-" + strings.ToLower(ulids.New().String()) + ".test"

		resp, err := suite.client.api.CreateOrganization(user.UserCtx, testclient.CreateOrganizationInput{
			Name: "avatar-listener-unreachable-" + ulids.New().String(),
			CreateOrgSettings: &testclient.CreateOrganizationSettingInput{
				Domains: []string{domain},
			},
		}, nil, nil)
		assert.NilError(t, err)

		created := resp.CreateOrganization.Organization
		assert.Assert(t, created.AvatarRemoteURL != nil)

		assert.NilError(t, setup.Runtime.WaitIdle(t.Context()))

		reloaded, err := suite.client.db.Organization.Query().
			Where(organization.IDEQ(created.ID)).
			WithSetting().
			Only(allowCtx)
		assert.NilError(t, err)
		assert.Assert(t, reloaded.AvatarRemoteURL != nil)
		assert.Check(t, is.Equal(*created.AvatarRemoteURL, *reloaded.AvatarRemoteURL))
	})
}

type avatarTestTransport struct {
	base   http.RoundTripper
	target *url.URL
}

func (t avatarTestTransport) RoundTrip(request *http.Request) (*http.Response, error) {
	if strings.HasPrefix(request.URL.Hostname(), "unreachable-") {
		return nil, errAvatarTestUnreachable
	}

	originalRequest := request.Clone(request.Context())
	rewrittenRequest := request.Clone(request.Context())
	rewrittenURL := *request.URL
	rewrittenURL.Scheme = t.target.Scheme
	rewrittenURL.Host = t.target.Host
	rewrittenRequest.URL = &rewrittenURL
	rewrittenRequest.Host = t.target.Host

	response, err := t.base.RoundTrip(rewrittenRequest)
	if response != nil {
		response.Request = originalRequest
	}

	return response, err
}
