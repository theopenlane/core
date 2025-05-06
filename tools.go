//go:build tools
// +build tools

package tools

import (
	_ "entgo.io/ent"
	_ "github.com/99designs/gqlgen"
	_ "github.com/99designs/gqlgen/graphql/introspection"
	_ "github.com/Yamashou/gqlgenc"
	_ "github.com/oNaiPs/go-generate-fast"
	_ "github.com/openfga/go-sdk"
	_ "github.com/vektra/mockery/v3"
)
