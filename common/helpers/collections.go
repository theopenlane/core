package helpers

// FoldMap reduces a map into an accumulator created by init
// The reducer mutates the accumulator in-place for efficiency
func FoldMap[K comparable, V any, R any](in map[K]V, init func(capacity int) R, reducer func(acc *R, key K, value V)) R {
	acc := init(len(in))
	for key, value := range in {
		reducer(&acc, key, value)
	}

	return acc
}
