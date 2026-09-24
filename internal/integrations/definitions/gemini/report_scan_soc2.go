package gemini

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/theopenlane/core/v2/pkg/docextract/soc2"
)

// This file holds what the report scan saga needs to know about SOC 2 reports. What a SOC 2
// report contains lives in pkg/docextract/soc2; what stays here is how this integration batches
// and requests those sections

const (
	// ReviewBatchSize is how many controls one reviews job covers
	ReviewBatchSize = 10
	// FindingBatchSize is how many controls one findings job covers; most controls have none so batches can be larger
	FindingBatchSize = 25
)

// controlScopedBatchSizes are the sections scoped by control ref codes, with how many controls
// one job of each covers
var controlScopedBatchSizes = map[string]int{
	soc2.ReviewsPart:  ReviewBatchSize,
	soc2.FindingsPart: FindingBatchSize,
}

// RequestedParts reads the sections a scan asked for from its metadata, validating every name;
// an absent or empty list means every configured section
func RequestedParts(metadata map[string]any) ([]string, error) {
	raw, _ := metadata[RequestedPartsMetadataKey].([]any)

	names := make([]string, 0, len(raw))

	for _, value := range raw {
		name, _ := value.(string)
		if _, ok := soc2.PartByName(name); !ok {
			return nil, fmt.Errorf("%w: %q, expected one of %s", ErrUnknownRequestedPart, name, strings.Join(soc2.PartNames(), ", "))
		}

		names = append(names, name)
	}

	return names, nil
}

// thinSectionError returns an error when the section holds fewer items than a report of this kind
// should contain, so the pass can be retried while attempts remain
func thinSectionError(envelope ReportScanPartEnvelope, section json.RawMessage) error {
	var items []json.RawMessage
	if err := json.Unmarshal(section, &items); err != nil {
		return nil
	}

	scoped := 0
	if envelope.isBatch() {
		scoped = len(envelope.RefCodes)
	}

	minimum, thin := soc2.SectionTooThin(envelope.Part, len(items), scoped)
	if !thin {
		return nil
	}

	return fmt.Errorf("%w: %s returned %d item(s), expected at least %d", ErrSectionTooThin, envelope.Part, len(items), minimum)
}
