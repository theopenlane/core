package main

//go:generate_input internal/ent/entc.go internal/ent/schema/* internal/ent/mixin/* internal/ent/templates/*
//go:generate_output internal/ent/generated/*
//go:generate go run ./internal/ent/entc.go
