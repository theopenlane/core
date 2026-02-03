package helpers

// AddPayloadIf attaches payloads to a details map when requested.
func AddPayloadIf(details map[string]any, include bool, key string, payload any) map[string]any {
	if !include {
		return details
	}
	if details == nil {
		details = map[string]any{}
	}
	details[key] = payload
	return details
}
