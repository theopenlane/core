package main

//go:generate_input entc.go ../schema/* ../mixin/* ../templates/*
//go:generate_output ../generated/*
//go:generate go run entc.go
