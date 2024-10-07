package storage

import "strings"

// StringPointer allows you to take the address of a string literal
func StringPointer(s string) *string {
	return &s
}

func IsStringEmpty(s string) bool { return len(strings.TrimSpace(s)) == 0 }
