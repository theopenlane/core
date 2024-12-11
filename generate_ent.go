package main

//go:generate echo "------> Generating code - running entc.go... <------"
//go:generate_input internal/ent/entc.go internal/ent/schema/*.go internal/ent/mixin/*.go internal/ent/templates/*.tmpl
//go:generate_output internal/ent/generated
//go:generate go run ./internal/ent/entc.go

//go:generate echo "------> Tidying up... <------"
//go:generate go mod tidy
//go:generate echo "------> Code generation process completed! <------"
