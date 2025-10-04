package objects

// Mutation represents any ent mutation that can provide ID and Type
type Mutation interface {
	ID() (string, error)
	Type() string
}

// GenericMutationAdapter adapts existing ent GenericMutation interface to our Mutation interface
type GenericMutationAdapter[T any] struct {
	mutation T
	idFunc   func(T) (string, bool)
	typeFunc func(T) string
}

// NewGenericMutationAdapter creates an adapter for existing ent mutations
func NewGenericMutationAdapter[T any](mutation T, idFunc func(T) (string, bool), typeFunc func(T) string) Mutation {
	return &GenericMutationAdapter[T]{
		mutation: mutation,
		idFunc:   idFunc,
		typeFunc: typeFunc,
	}
}

// ID implements the Mutation interface
func (a *GenericMutationAdapter[T]) ID() (string, error) {
	id, exists := a.idFunc(a.mutation)
	if !exists {
		return "", ErrMutationIDNotFound
	}

	return id, nil
}

// Type implements the Mutation interface
func (a *GenericMutationAdapter[T]) Type() string {
	return a.typeFunc(a.mutation)
}
