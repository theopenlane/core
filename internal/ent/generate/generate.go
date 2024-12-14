package main

//go:generate_input entc.go ../schema/* ../mixin/* ../templates/* ../../graphapi/schema/*.graphql ../../graphapi/query/*.graphql
//go:generate_output ../generated/* ../../graphapi/schema/*.graphql ../../graphapi/query/*.graphql
//go:generate go run entc.go
