package enums

import (
	"io"
	"strings"
)

// marshalGQL writes the quoted string representation for any string-based enum
func marshalGQL[T ~string](r T, w io.Writer) {
	// GQL enums are serialized as quoted strings
	_, _ = w.Write([]byte(`"` + string(r) + `"`))
}

// unmarshalGQL type-asserts a GraphQL value and converts it to the target enum type
func unmarshalGQL[T ~string](r *T, v any) error {
	str, ok := v.(string)
	if !ok {
		return ErrInvalidType
	}

	// convert the raw string to the concrete enum type and assign to the receiver
	*r = T(str)

	return nil
}

// parse finds a matching value in the slice using case-insensitive comparison and
// returns a pointer into the slice; Returns the fallback pointer if no match is found
func parse[T ~string](input string, values []T, fallback *T) *T {
	upper := strings.ToUpper(input)
	for i := range values {
		if string(values[i]) == upper {
			// return pointer into the slice to avoid allocating a new value
			return &values[i]
		}
	}

	return fallback
}

// stringValues converts a typed enum slice to []string
func stringValues[T ~string](vals []T) []string {
	out := make([]string, len(vals))
	for i, v := range vals {
		out[i] = string(v)
	}

	return out
}
