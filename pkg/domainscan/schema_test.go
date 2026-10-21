package domainscan

import (
	"testing"

	"gotest.tools/v3/assert"

	"github.com/theopenlane/utils/jsonx"
)

func TestDomainScanReportSchemaReflects(t *testing.T) {
	assert.Assert(t, len(DomainScanReportSchema) > 0)
	assert.Equal(t, jsonx.SchemaID(DomainScanReportSchema), "Report")
}
