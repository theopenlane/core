package main

//go:generate_input entc.go ../schema/* ../mixin/* ../templates/* ../../graphapi/schema/*.graphql ../../graphapi/query/*.graphql
//go:generate_output ../generated/* ../../graphapi/schema/*.graphql ../../graphapi/query/*.graphql
//go:generate go run entc.go

//go:generate_input ../../graphapi/generate/clean_notification.go ../../graphapi/schema/ent.graphql
//go:generate_output ../../graphapi/schema/ent.graphql
//go:generate go run ../../graphapi/generate/clean_notification.go
