package graphsubscriptions

// Notification is an interface that represents a notification that can be published
// This avoids import cycles with ent/generated
// The ent/generated.Notification type has ID and UserID fields that satisfy this interface
type Notification interface{}
