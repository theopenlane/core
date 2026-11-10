package handlers

import (
	"testing"

	ent "github.com/theopenlane/core/v2/internal/ent/generated"
	"gotest.tools/v3/assert"
	is "gotest.tools/v3/assert/cmp"
)

func TestParseName(t *testing.T) {
	tests := []struct {
		name string
		user string
		want ent.CreateUserInput
	}{
		{
			name: "happy path",
			user: "Matty Anderson",
			want: ent.CreateUserInput{
				FirstName: new("Matty"),
				LastName:  new("Anderson"),
			},
		},
		{
			name: "very long name",
			user: "Matty Anderson Is The Best",
			want: ent.CreateUserInput{
				FirstName: new("Matty"),
				LastName:  new("Anderson Is The Best"),
			},
		},
		{
			name: "single name",
			user: "Matty",
			want: ent.CreateUserInput{
				FirstName: new("Matty"),
			},
		},
		{
			name: "empty name",
			user: "",
			want: ent.CreateUserInput{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := parseName(tt.user)

			assert.Check(t, is.DeepEqual(tt.want, got))
		})
	}
}
