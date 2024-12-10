package main

//go:generate echo "------> Generating code - running gqlgen... <------"
//go:generate_input  internal/graphapi/generate/generate.go internal/graphapi/*.go
//go:generate_output internal/graphapi/*.resolvers.go internal/graphapi/gen_models.go internal/graphapi/gen_server.go pkg/openlaneclient/models.go
//go:generate go run ./internal/graphapi/generate/generate.go

//go:generate echo "------> Generating code - running gen_schema.go... <------"
//go:generate_input gen_schema.go internal/graphapi/*.go
//go:generate_output schema.graphql
//go:generate go run ./gen_schema.go

//go:generate echo "------> Generating code - running gqlgenc... <------"
//go:generate_input query/*.graphql schema/*.graphql internal/graphapi/*.resolvers.go internal/graphapi/gen_models.go internal/graphapi/gen_server.go pkg/openlaneclient/models.go
//go:generate_output pkg/openlaneclient/graphqlclient.go
//go:generate go run github.com/Yamashou/gqlgenc generate --configdir schema

//go:generate echo "------> Tidying up... <------"
//go:generate go mod tidy
//go:generate echo "------> Code generation process completed! <------"
