package aikido

import (
	"context"
	"fmt"
	"net/http"
	"testing"

	"gotest.tools/v3/assert"
	is "gotest.tools/v3/assert/cmp"

	"github.com/theopenlane/core/v2/internal/ent/entityops"
)

func TestCollectAssets(t *testing.T) {
	responses := map[string]string{
		pathCodeRepositories:                 `[{"id":42,"name":"example-repo","provider":"github"}]`,
		pathContainers:                       `[{"id":5,"name":"api-image","provider":"aws"}]`,
		pathClouds:                           `[{"id":7,"name":"AWS Cloud environment","provider":"aws"}]`,
		pathDomains:                          `[{"id":9,"domain":"app.example.com","kind":"front_end"}]`,
		pathVirtualMachines:                  `[{"id":3,"name":"vm1"}]`,
		pathDevices:                          `[{"id":1,"name":"Yuri's MacBook Pro"}]`,
		fmt.Sprintf(pathSubdomainsFormat, "9"): `[{"id":"123e4567-e89b-12d3-a456-426614174000","name":"api.example.com"},{"id":"123e4567-e89b-12d3-a456-426614174001","name":"old.example.com","is_excluded":true}]`,
	}

	client := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		body, ok := responses[r.URL.Path]
		assert.Check(t, ok, "unexpected path %s", r.URL.Path)

		if r.URL.Path == pathContainers {
			assert.Check(t, is.Equal("true", r.URL.Query().Get("include_labels")))
		}

		writeJSON(w, body)
	})

	payloads, err := CollectAssets(context.Background(), client)
	assert.NilError(t, err)

	envelopes := payloadSet(payloads, entityops.SchemaAsset.Name)
	assert.Check(t, is.Len(envelopes, 7))

	got := map[string]string{}
	for _, envelope := range envelopes {
		got[envelope.Variant] = envelope.Resource
	}

	assert.Check(t, is.DeepEqual(map[string]string{
		assetVariantCodeRepository: "example-repo",
		assetVariantContainer:      "api-image",
		assetVariantCloud:          "AWS Cloud environment",
		assetVariantDomain:         "app.example.com",
		assetVariantVirtualMachine: "vm1",
		assetVariantDevice:         "Yuri's MacBook Pro",
		assetVariantSubdomain:      "api.example.com",
	}, got))
}

func TestCollectAssetsSkipsOptionalKinds(t *testing.T) {
	client := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == pathVirtualMachines || r.URL.Path == pathDevices {
			w.WriteHeader(http.StatusForbidden)

			return
		}

		writeJSON(w, `[{"id":1,"name":"example"}]`)
	})

	payloads, err := CollectAssets(context.Background(), client)
	assert.NilError(t, err)

	variants := map[string]int{}
	for _, envelope := range payloadSet(payloads, entityops.SchemaAsset.Name) {
		variants[envelope.Variant]++
	}

	assert.Check(t, is.Equal(0, variants[assetVariantVirtualMachine]))
	assert.Check(t, is.Equal(0, variants[assetVariantDevice]))
	assert.Check(t, is.Equal(1, variants[assetVariantCodeRepository]))
}

func TestCollectAssetsPropagatesAPIErrors(t *testing.T) {
	client := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == pathClouds {
			w.WriteHeader(http.StatusForbidden)

			return
		}

		writeJSON(w, `[]`)
	})

	_, err := CollectAssets(context.Background(), client)
	assert.Check(t, is.ErrorIs(err, ErrAssetsFetchFailed))
}
