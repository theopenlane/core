// Package version carries the build identity stamped into the core server binary
package version

// Version is the release the binary was built from, stamped at build time with
// -X github.com/theopenlane/core/v2/pkg/version.Version. It is empty in unstamped builds,
// which callers read as "not a release", e.g. local development
var Version = ""
