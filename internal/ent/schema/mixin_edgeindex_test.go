package schema

import (
	"testing"

	"gotest.tools/v3/assert"
)

// indexFields returns the field and storage key of every index the mixin produces
func indexFields(m EdgeIndexMixin) map[string]string {
	got := map[string]string{}

	for _, idx := range m.Indexes() {
		desc := idx.Descriptor()
		got[desc.Fields[0]] = desc.StorageKey
	}

	return got
}

func TestEdgeIndexMixinIndexesEdgeFields(t *testing.T) {
	// user_id is held on the user_settings table by the inverse edge to User
	got := indexFields(newEdgeIndexMixin(UserSetting{}))

	assert.Equal(t, got["user_id"], "user_setting_user_id_idx")

	// default_org has no field, ent names the column itself so there is nothing to index on
	_, ok := got["default_org"]
	assert.Assert(t, !ok)
}

func TestEdgeIndexMixinSkipsCoveredFields(t *testing.T) {
	// campaign_target declares its own indexes, the partial ones must still get a full index
	// because postgres cannot use a partial index to enforce a foreign key
	covered := coveredFields(CampaignTarget{})

	for field, isCovered := range covered {
		if isCovered {
			got := indexFields(newEdgeIndexMixin(CampaignTarget{}))
			_, ok := got[field]
			assert.Assert(t, !ok, "field %q is already fully indexed and must not be duplicated", field)
		}
	}
}
