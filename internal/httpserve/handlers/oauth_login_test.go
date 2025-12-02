package handlers

import (
	"testing"

	"github.com/samber/lo"
	"github.com/stretchr/testify/assert"

	ent "github.com/theopenlane/ent/generated"
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
				FirstName: lo.ToPtr("Matty"),
				LastName:  lo.ToPtr("Anderson"),
			},
		},
		{
			name: "very long name",
			user: "Matty Anderson Is The Best",
			want: ent.CreateUserInput{
				FirstName: lo.ToPtr("Matty"),
				LastName:  lo.ToPtr("Anderson Is The Best"),
			},
		},
		{
			name: "single name",
			user: "Matty",
			want: ent.CreateUserInput{
				FirstName: lo.ToPtr("Matty"),
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

			assert.Equal(t, tt.want, got)
		})
	}
}
