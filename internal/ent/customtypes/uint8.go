package customtypes

import (
	"fmt"
	"io"
	"strconv"

	"github.com/99designs/gqlgen/graphql"
)

// Uint8 is a custom type for uint8
type Uint8 uint8

// MarshalGQL implements the graphql.Marshaler interface
func (u Uint8) MarshalGQL(w io.Writer) {
	fmt.Fprint(w, strconv.FormatUint(uint64(u), 10))
}

// UnmarshalGQL implements the graphql.Unmarshaler interface
func (u *Uint8) UnmarshalGQL(v interface{}) error {
	i, err := graphql.UnmarshalUint64(v)
	if err != nil {
		return err
	}

	*u = Uint8(i) // nolint:gosec

	return nil
}
