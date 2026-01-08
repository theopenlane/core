package soiree

func typedEventTopic(name string) TypedTopic[Event] {
	return NewTypedTopic(
		name,
		WithWrap(func(e Event) Event { return e }),
		WithUnwrap(func(e Event) (Event, error) { return e, nil }),
	)
}

func collectEmitErrors(errCh <-chan error) []error {
	var errs []error
	for err := range errCh {
		if err != nil {
			errs = append(errs, err)
		}
	}

	return errs
}
