// Package observability provides logging and metrics wrappers and consistency for workflows
// there are several functions in this package that are geared towards reducing boilerplate overhead with the main callers
// by pre-setting common fields such as operation origin and trigger event so that inline within the workflow package we don't have crazy verbose
// log and metric statements making the code harder to read
package observability
