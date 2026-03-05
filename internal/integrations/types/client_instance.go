package types //nolint:revive

// ClientInstance wraps a provider client instance for operation execution
type ClientInstance struct {
	raw any
}

// NewClientInstance creates a client instance wrapper from a concrete client value
func NewClientInstance(raw any) ClientInstance {
	return ClientInstance{raw: raw}
}

// EmptyClientInstance returns a zero client wrapper
func EmptyClientInstance() ClientInstance {
	return ClientInstance{}
}

// ClientInstanceAs unwraps a wrapped client value as a concrete type
func ClientInstanceAs[T any](client ClientInstance) (T, bool) {
	value, ok := client.raw.(T)
	if ok {
		return value, true
	}

	var zero T
	return zero, false
}
