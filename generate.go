package main

//go:generate echo "------> Generating code - running entc.go... <------"
//go:generate go run -mod=mod ./internal/ent/entc.go
//go:generate echo "------> Generating code - running gqlgen... <------"
//go:generate go run ./internal/graphapi/generate/generate.go
//go:generate echo "------> Generating code - running gen_schema.go... <------"
//go:generate go run -mod=mod ./gen_schema.go
//go:generate echo "------> Generating code - running gqlgenc... <------"
//go:generate go run -mod=mod github.com/Yamashou/gqlgenc generate --configdir schema
//go:generate echo "------> Tidying up... <------"
//go:generate go mod tidy
//go:generate echo "------> Code generation process completed! <------"
