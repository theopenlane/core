package interceptors

type (
	// SkipMode is a bit flag for the Skip annotation.
	SkipMode int
)

const (
	// SkipOnlyQuery skips the interceptor on `Only` queries.
	SkipOnlyQuery SkipMode = 1 << iota
	// SkipAllQuery skips the interceptor on `All` queries.
	SkipAllQuery
	// SkipExistsQuery skips the interceptor on `Exists` queries.
	SkipExistsQuery

	// SkipAll is default mode to skip all.
	SkipAll = SkipOnlyQuery |
		SkipAllQuery |
		SkipExistsQuery
)
