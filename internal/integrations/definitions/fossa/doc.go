// Package fossa defines the FOSSA software composition analysis integration definition.
//
// FOSSA scans project dependencies and reports issues in three categories: vulnerability,
// licensing, and quality. This definition syncs security vulnerabilities unconditionally and
// OSS license compliance issues only when the installation opts in; quality issues are out of scope.
//
// A fresh FOSSA organization reports no vulnerability issues until a project containing a
// vulnerable dependency has been scanned. To produce one for local testing, scan a project
// pinned to a dependency with a known CVE, for example npm axios@1.15.0 or Go
// github.com/gogo/protobuf@v1.3.1, then confirm a non-zero count from /api/v2/issues/categories.
package fossa
