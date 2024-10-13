package storage

import "strings"

// toPointer allows you to take the address of any literal
func toPointer[T any](s T) *T {
	return &s
}

// isStringEmpty checks if a string is empty
func isStringEmpty(s string) bool { return len(strings.TrimSpace(s)) == 0 }
