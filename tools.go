//go:build tools
// +build tools

package tools

import (
	_ "entgo.io/ent"
	_ "github.com/99designs/gqlgen"
	_ "github.com/99designs/gqlgen/graphql/introspection"
	_ "github.com/Yamashou/gqlgenc"
	_ "github.com/datumforge/fgax/mockery"
	_ "github.com/openfga/go-sdk"
)
