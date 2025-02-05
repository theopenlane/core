package main

//go:generate_input gen_gqlgen.go .gqlgen.yml ../schema/*
//go:generate_output ../generated/* ../model/gen_models.go ../*.resolvers.go ../../../pkg/openlaneclient/models.go
//go:generate go run gen_gqlgen.go

//go:generate_input gen_schema.go ../generated/*.generated.go ../model/gen_models.go
//go:generate_output ../clientschema/schema.graphql
//go:generate go run gen_schema.go

//go:generate_input ../query/*.graphql ../schema/*.graphql ../clientschema/schema.graphql .gqlgenc.yml
//go:generate_output ../../pkg/openlaneclient/graphqlclient.go
//go:generate go run gen_client.go
