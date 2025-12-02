package interceptors

type (
	// SkipMode is a bit flag for the Skip annotation.
	SkipMode int
)

const (
	// SkipNone skips no queries.
	SkipNone SkipMode = 0
	// SkipOnlyQuery skips the interceptor on `Only` queries.
	SkipOnlyQuery SkipMode = 1 << iota
	// SkipAllQuery skips the interceptor on `All` queries.
	SkipAllQuery
	// SkipExistsQuery skips the interceptor on `Exists` queries.
	SkipExistsQuery
	// SkipIDsQuery skips the interceptor on `IDs` queries.
	SkipIDsQuery

	// SkipAll is default mode to skip all.
	SkipAll = SkipOnlyQuery |
		SkipAllQuery |
		SkipExistsQuery |
		SkipIDsQuery
)
