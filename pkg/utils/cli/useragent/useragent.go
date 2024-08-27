package useragent

import (
	"fmt"
	"runtime"

	"github.com/theopenlane/core/internal/constants"
)

// GetUserAgent returns the user agent string for the CLI
func GetUserAgent() string {
	product := "openlane-cli"
	productVersion := constants.CLIVersion

	userAgent := fmt.Sprintf("%s/%s (%s) %s (%s)",
		product, productVersion, runtime.GOOS, runtime.GOARCH, runtime.Version())

	return userAgent
}
