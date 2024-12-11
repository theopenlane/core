package main

//go:generate_input  internal/graphapi/generate/generate.go gqlgen.yml github.com/theopenlane/core/internal/ent/generated/* schema/*
//go:generate_output internal/graphapi/* pkg/openlaneclient/models.go
//go:generate go run ./internal/graphapi/generate/generate.go

//go:generate_input gen_schema.go internal/graphapi/gen_server.go
//go:generate_output schema.graphql
//go:generate go run ./gen_schema.go

//go:generate_input query/* schema/* gqlgenc.yml
//go:generate_output pkg/openlaneclient/graphqlclient.go
//go:generate go run github.com/Yamashou/gqlgenc generate --configdir schema
