package interceptors

import (
	"context"
	"testing"

	"entgo.io/ent"
	"github.com/99designs/gqlgen/graphql"
	"github.com/stretchr/testify/assert"
)

func TestFilterType(t *testing.T) {
	tests := []struct {
		name string
		ctx  context.Context
		want string
	}{
		{
			name: "updateGroup",
			ctx: ent.NewQueryContext(graphql.WithRootFieldContext(context.Background(), &graphql.RootFieldContext{
				Object: "updateGroup",
			}), &ent.QueryContext{
				Type: "",
			},
			),
			want: "org",
		},
		{
			name: "task",
			ctx: ent.NewQueryContext(graphql.WithRootFieldContext(context.Background(), &graphql.RootFieldContext{
				Object: "task",
			}), &ent.QueryContext{
				Type: "",
			},
			),
			want: "org",
		},
		{
			name: "createGroup",
			ctx: ent.NewQueryContext(graphql.WithRootFieldContext(context.Background(), &graphql.RootFieldContext{
				Object: "createGroup",
			}), &ent.QueryContext{
				Type: "",
			},
			),
			want: "org",
		},
		{
			name: "OrgMembership",
			ctx: ent.NewQueryContext(graphql.WithRootFieldContext(context.Background(), &graphql.RootFieldContext{
				Object: "",
			}), &ent.QueryContext{
				Type: "OrgMembership",
			},
			),
			want: "org",
		},
		{
			name: "OrgMembership",
			ctx: ent.NewQueryContext(graphql.WithRootFieldContext(context.Background(), &graphql.RootFieldContext{
				Object: "",
			}), &ent.QueryContext{
				Type: "GroupMembership",
			},
			),
			want: "org",
		},
		{
			name: "Group",
			ctx: ent.NewQueryContext(graphql.WithRootFieldContext(context.Background(), &graphql.RootFieldContext{
				Object: "",
			}), &ent.QueryContext{
				Type: "Group",
			},
			),
			want: "org",
		},
		{
			name: "Organization",
			ctx: ent.NewQueryContext(graphql.WithRootFieldContext(context.Background(), &graphql.RootFieldContext{
				Object: "",
			}), &ent.QueryContext{
				Type: "Organization",
			},
			),
			want: "",
		},
		{
			name: "User",
			ctx: ent.NewQueryContext(graphql.WithRootFieldContext(context.Background(), &graphql.RootFieldContext{
				Object: "",
			}), &ent.QueryContext{
				Type: "User",
			},
			),
			want: "user",
		},
		{
			name: "UserSetting",
			ctx: ent.NewQueryContext(graphql.WithRootFieldContext(context.Background(), &graphql.RootFieldContext{
				Object: "",
			}), &ent.QueryContext{
				Type: "UserSetting",
			},
			),
			want: "user",
		},
		{
			name: "nil graphql context",
			ctx: ent.NewQueryContext(context.Background(), &ent.QueryContext{
				Type: "UserSetting",
			}),
			want: "user",
		},
		{
			name: "nil ent context", // shouldn't happen but just in case
			ctx:  context.Background(),
			want: "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ft := userFilterType(tt.ctx)

			assert.Equal(t, tt.want, ft)
		})
	}
}

func TestEmailDomainConditionContext(t *testing.T) {
	tests := []struct {
		name  string
		email string
		want  *map[string]any
	}{
		{
			name:  "valid email",
			email: "person@example.com",
			want:  &map[string]any{"email_domain": "example.com"},
		},
		{
			name:  "normalizes domain",
			email: "person@ EXAMPLE.COM ",
			want:  &map[string]any{"email_domain": "example.com"},
		},
		{
			name:  "missing separator",
			email: "person.example.com",
		},
		{
			name:  "missing domain",
			email: "person@",
		},
		{
			name: "empty email",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, emailDomainConditionContext(tt.email))
		})
	}
}
