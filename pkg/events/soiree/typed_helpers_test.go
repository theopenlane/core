package soiree

func typedEventTopic(name string) TypedTopic[Event] {
	return NewTypedTopic(
		name,
		func(e Event) Event { return e },
		func(e Event) (Event, error) { return e, nil },
	)
}
