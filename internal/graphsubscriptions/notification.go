package graphsubscriptions

// Notification is an interface that represents a notification that can be published
// This avoids import cycles with ent/generated
// The ent/generated.Notification type has ID and UserID fields that satisfy this interface
type Notification interface{}

// RawNotification wraps a notification received over Redis that hasn't been unmarshaled into its concrete type yet
type RawNotification struct {
	Payload []byte
}
