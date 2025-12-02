package schema

import (
	"context"
	"testing"

	"entgo.io/ent"
	"github.com/stretchr/testify/assert"

	"github.com/theopenlane/ent/interceptors"
)

func Test_skipQueryModeCheck(t *testing.T) {
	type args struct {
		mode interceptors.SkipMode
		op   string
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "SkipNone returns false",
			args: args{
				mode: interceptors.SkipNone,
				op:   interceptors.AllOperation,
			},
			want: false,
		},
		{
			name: "SkipAll returns true",
			args: args{
				mode: interceptors.SkipAll,
				op:   interceptors.AllOperation,
			},
			want: true,
		},
		{
			name: "SkipAllQuery matches AllOperation",
			args: args{
				mode: interceptors.SkipAllQuery,
				op:   interceptors.AllOperation,
			},
			want: true,
		},
		{
			name: "SkipOnlyQuery matches OnlyOperation",
			args: args{
				mode: interceptors.SkipOnlyQuery,
				op:   interceptors.OnlyOperation,
			},
			want: true,
		},
		{
			name: "SkipExistsQuery matches ExistOperation",
			args: args{
				mode: interceptors.SkipExistsQuery,
				op:   interceptors.ExistOperation,
			},
			want: true,
		},
		{
			name: "SkipIDsQuery matches IDsOperation",
			args: args{
				mode: interceptors.SkipIDsQuery,
				op:   interceptors.IDsOperation,
			},
			want: true,
		},
		{
			name: "SkipAllQuery does not match OnlyOperation",
			args: args{
				mode: interceptors.SkipAllQuery,
				op:   interceptors.OnlyOperation,
			},
			want: false,
		},
		{
			name: "Unknown operation returns false",
			args: args{
				mode: interceptors.SkipAllQuery,
				op:   "999", // unknown op
			},
			want: false,
		},
		{
			name: "Multiple flags: SkipAllQuery|SkipOnlyQuery matches OnlyOperation",
			args: args{
				mode: interceptors.SkipAllQuery | interceptors.SkipOnlyQuery,
				op:   interceptors.OnlyOperation,
			},
			want: true,
		},
		{
			name: "Multiple flags: SkipAllQuery|SkipOnlyQuery matches AllOperation",
			args: args{
				mode: interceptors.SkipAllQuery | interceptors.SkipOnlyQuery,
				op:   interceptors.AllOperation,
			},
			want: true,
		},
		{
			name: "Multiple flags: SkipAllQuery|SkipOnlyQuery does not match IDsOperation",
			args: args{
				mode: interceptors.SkipAllQuery | interceptors.SkipOnlyQuery,
				op:   interceptors.IDsOperation,
			},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()

			queryCtx := &ent.QueryContext{Op: tt.args.op}
			ctx = ent.NewQueryContext(ctx, queryCtx)
			got := skipQueryModeCheck(ctx, tt.args.mode)
			assert.Equal(t, tt.want, got)
		})
	}
}
