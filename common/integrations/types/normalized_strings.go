package types

import (
	"encoding"
	"strings"
)

// TrimmedString stores a string normalized by trimming leading/trailing whitespace
type TrimmedString string

// LowerString stores a trimmed, lower-cased string
type LowerString string

// UpperString stores a trimmed, upper-cased string
type UpperString string

var (
	_ encoding.TextUnmarshaler = (*TrimmedString)(nil)
	_ encoding.TextUnmarshaler = (*LowerString)(nil)
	_ encoding.TextUnmarshaler = (*UpperString)(nil)
)

// UnmarshalText parses text into a trimmed string
func (s *TrimmedString) UnmarshalText(text []byte) error {
	*s = TrimmedString(strings.TrimSpace(string(text)))

	return nil
}

// UnmarshalText parses text into a trimmed, lower-cased string
func (s *LowerString) UnmarshalText(text []byte) error {
	*s = LowerString(strings.ToLower(strings.TrimSpace(string(text))))

	return nil
}

// UnmarshalText parses text into a trimmed, upper-cased string
func (s *UpperString) UnmarshalText(text []byte) error {
	*s = UpperString(strings.ToUpper(strings.TrimSpace(string(text))))

	return nil
}

// String returns the trimmed string value
func (s TrimmedString) String() string { return string(s) }

// String returns the lower-cased string value
func (s LowerString) String() string { return string(s) }

// String returns the upper-cased string value
func (s UpperString) String() string { return string(s) }
