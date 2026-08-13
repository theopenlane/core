package hooks

import (
	"encoding/json"
	"fmt"

	"entgo.io/ent"
)

// mutationEventID is the identifier structure parsed from mutation return values
type mutationEventID struct {
	ID string `json:"id,omitempty"`
}

// mutationEventEntityID parses the mutated entity id from the returned entity value
func mutationEventEntityID(retVal ent.Value) (string, error) {
	out, err := json.Marshal(retVal)
	if err != nil {
		return "", fmt.Errorf("%w: %w", ErrUnableToDetermineEventID, err)
	}

	event := mutationEventID{}
	if err := json.Unmarshal(out, &event); err != nil {
		return "", fmt.Errorf("%w: %w", ErrUnableToDetermineEventID, err)
	}

	return event.ID, nil
}
