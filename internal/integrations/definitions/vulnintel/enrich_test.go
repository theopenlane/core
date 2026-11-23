package vulnintel

import (
	"testing"
	"time"

	"github.com/ossf/osv-schema/bindings/go/osvschema"
	"google.golang.org/protobuf/encoding/protojson"
	"gotest.tools/v3/assert"
	is "gotest.tools/v3/assert/cmp"

	"github.com/theopenlane/core/common/models"
	"github.com/theopenlane/core/v2/internal/ent/generated"
)

func testAdvisory(t *testing.T) osvRecord {
	t.Helper()

	var vuln osvschema.Vulnerability
	assert.NilError(t, protojson.Unmarshal([]byte(osvAdvisory), &vuln))

	return recordFromOSV(&vuln)
}

func TestAdvisoryUpdateFillsEmptyFields(t *testing.T) {
	update := generated.NewClient().Vulnerability.UpdateOneID("vuln-1")
	vuln := &generated.Vulnerability{ID: "vuln-1"}

	update, changed := advisoryUpdate(update, vuln, testAdvisory(t))
	assert.Assert(t, changed)

	mutation := update.Mutation()

	summary, _ := mutation.Summary()
	description, _ := mutation.Description()
	vector, _ := mutation.Vector()
	published, _ := mutation.PublishedAt()
	references, _ := mutation.References()

	assert.Check(t, is.Equal("Regular expression denial of service in zod", summary))
	assert.Check(t, is.Equal("The emoji validator backtracks exponentially.", description))
	assert.Check(t, is.Equal("CVSS:3.1/AV:N/AC:L/PR:N/UI:N/S:U/C:N/I:N/A:L", vector))
	assert.Check(t, is.Equal(models.DateTime(time.Date(2026, 10, 21, 17, 59, 27, 0, time.UTC)), published))
	assert.Check(t, is.DeepEqual([]string{
		"https://osv.dev/vulnerability/GHSA-253c-mchw-3w2r",
		"https://github.com/advisories/GHSA-253c-mchw-3w2r",
	}, references))
}

func TestAdvisoryUpdateKeepsProviderFields(t *testing.T) {
	published := models.DateTime(time.Date(2026, 1, 27, 0, 0, 0, 0, time.UTC))
	vuln := &generated.Vulnerability{
		ID:                  "vuln-1",
		Summary:             "Attacker can trigger DOS-attack via regex",
		Description:         "Aikido tldr",
		Vector:              "CVSS:3.1/AV:L",
		FirstPatchedVersion: "3.25.1",
		PublishedAt:         &published,
		References:          []string{"https://intel.aikido.dev/cve/GHSA-253c-mchw-3w2r"},
	}

	_, changed := advisoryUpdate(generated.NewClient().Vulnerability.UpdateOneID("vuln-1"), vuln, testAdvisory(t))
	assert.Check(t, !changed)
}

func TestAdvisoryID(t *testing.T) {
	id, ok := advisoryID(" GHSA-253c-mchw-3w2r ")
	assert.Check(t, ok)
	assert.Check(t, is.Equal("GHSA-253c-mchw-3w2r", id))

	for _, rawID := range []string{"", "../x", "a/b", "CVE..1", ".hidden", "CVE 2026"} {
		_, ok := advisoryID(rawID)
		assert.Check(t, !ok, rawID)
	}
}
