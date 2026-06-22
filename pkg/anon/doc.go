// Package anon identifies anonymous callers (trust center and questionnaire) from a
// request context. It depends only on the iam auth package so it can be imported by any
// layer - interceptors, schema mixins, privacy policy, hooks, and schemautil - without
// creating an import cycle
package anon
