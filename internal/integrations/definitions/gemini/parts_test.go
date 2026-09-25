package gemini

import (
	"testing"

	"gotest.tools/v3/assert"
)

func TestRequestedPartsEmpty(t *testing.T) {
	parts, err := RequestedParts(nil)

	assert.NilError(t, err)
	assert.Check(t, len(parts) == 0)

	parts, err = RequestedParts(map[string]any{RequestedPartsMetadataKey: []any{}})

	assert.NilError(t, err)
	assert.Check(t, len(parts) == 0)
}

func TestRequestedParts(t *testing.T) {
	parts, err := RequestedParts(map[string]any{RequestedPartsMetadataKey: []any{"controls", "reviews"}})

	assert.NilError(t, err)
	assert.DeepEqual(t, parts, []string{"controls", "reviews"})
}

func TestRequestedPartsUnknown(t *testing.T) {
	_, err := RequestedParts(map[string]any{RequestedPartsMetadataKey: []any{"controls", "invoices"}})

	assert.ErrorIs(t, err, ErrUnknownRequestedPart)
	assert.ErrorContains(t, err, `"invoices"`)
}

func TestSelectPartsDefaultsToConfigured(t *testing.T) {
	request := ReportScanRequest{parts: []string{"controls", "reviews"}}

	selected, err := request.selectParts(nil)

	assert.NilError(t, err)
	assert.DeepEqual(t, selected, []string{"controls", "reviews"})
}

func TestSelectPartsIntersectsWithConfigured(t *testing.T) {
	request := ReportScanRequest{parts: []string{"controls", "reviews", "assets"}}

	selected, err := request.selectParts(map[string]any{RequestedPartsMetadataKey: []any{"reviews", "findings"}})

	assert.NilError(t, err)
	assert.DeepEqual(t, selected, []string{"reviews"})
}

func TestSelectPartsNoneAvailable(t *testing.T) {
	request := ReportScanRequest{parts: []string{"controls"}}

	_, err := request.selectParts(map[string]any{RequestedPartsMetadataKey: []any{"findings"}})

	assert.ErrorIs(t, err, ErrRequestedPartsUnavailable)
}
