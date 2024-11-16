package utils

import "github.com/google/uuid"

func Filter[T any](s []T, f func(T) bool) []T {
	result := make([]T, 0)
	for _, v := range s {
		if f(v) {
			result = append(result, v)
		}
	}
	return result
}

func ContainsFunc[T any](s []T, f func(T) bool) bool {
	for _, v := range s {
		if f(v) {
			return true
		}
	}
	return false
}

// IsValidUUID returns true if passed string in uuid format
// defined by `github.com/google/uuid`.Parse
// else return false
func IsValidUUID(key string) bool {
	_, err := uuid.Parse(key)
	return err == nil
}
