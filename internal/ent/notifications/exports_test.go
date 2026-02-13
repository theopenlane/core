package notifications

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/theopenlane/core/common/enums"
	"github.com/theopenlane/core/internal/ent/events"
	"github.com/theopenlane/core/internal/ent/generated/export"
)

func TestExtractExportFromPayloadMetadata(t *testing.T) {
	fields := &exportFields{}
	payload := &events.MutationPayload{
		EntityID: "export-123",
		ProposedChanges: map[string]any{
			export.FieldOwnerID:      "org-1",
			export.FieldRequestorID:  "user-2",
			export.FieldExportType:   "TASK",
			export.FieldStatus:       "READY",
			export.FieldErrorMessage: "none",
		},
	}

	extractExportFromPayload(payload, fields)

	assert.Equal(t, "export-123", fields.entityID)
	assert.Equal(t, "org-1", fields.ownerID)
	assert.Equal(t, "user-2", fields.requestorID)
	assert.Equal(t, enums.ExportTypeTask, fields.exportType)
	assert.Equal(t, enums.ExportStatus("READY"), fields.status)
	assert.Equal(t, "none", fields.errorMessage)
}

func TestParseExportEnums(t *testing.T) {
	exportType, ok := parseExportType("TASK")
	assert.True(t, ok)
	assert.Equal(t, enums.ExportTypeTask, exportType)

	status, ok := parseExportStatus(enums.ExportStatus("FAILED"))
	assert.True(t, ok)
	assert.Equal(t, enums.ExportStatusFailed, status)

	exportType, ok = parseExportType("UNKNOWN_EXPORT_TYPE")
	assert.False(t, ok)
	assert.Equal(t, enums.ExportType(""), exportType)

	status, ok = parseExportStatus("UNKNOWN_EXPORT_STATUS")
	assert.False(t, ok)
	assert.Equal(t, enums.ExportStatus(""), status)
}
