package soiree

import "strings"

// PropertyEventID is the reserved properties key used to identify events across retries/replays
const PropertyEventID = "soiree.event_id"

// EventID returns the stable idempotency key for an event when present
func EventID(event Event) string {
	if event == nil {
		return ""
	}

	props := event.Properties()
	if props == nil {
		return ""
	}

	value, ok := props[PropertyEventID].(string)
	if !ok {
		return ""
	}

	return strings.TrimSpace(value)
}
