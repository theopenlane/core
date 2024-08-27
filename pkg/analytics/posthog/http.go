package posthog

import (
	"net/http"

	"github.com/posthog/posthog-go"
)

// GetPackageToPosthogProperties can be used to parse out an http header and map the headers to posthog properties array
func GetPackageToPosthogProperties(header *http.Header) posthog.Properties {
	properties := posthog.NewProperties().
		Set("browser", header.Get("browser")).
		Set("lang", header.Get("lang")).
		Set("lang_version", header.Get("lang_version")).
		Set("machine", header.Get("machine")).
		Set("os", header.Get("os")).
		Set("package_version", header.Get("package_version")).
		Set("processor", header.Get("processor")).
		Set("publisher", header.Get("publisher")).
		Set("release", header.Get("release")).
		Set("sdk_runtime", header.Get("sdk_runtime")).
		Set("system", header.Get("system")).
		Set("login_method", header.Get("user_agent"))

	return properties
}
