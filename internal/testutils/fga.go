//go:build test

package testutils

import "os"

// FGACacheEnvVar enables the production openfga cache settings in the test container when set to true
const FGACacheEnvVar = "OPENLANE_TEST_FGA_CACHE"

// GetDefaultFGAEnvs returns env variables required by OpenFGA tests.
func GetDefaultFGAEnvs() map[string]string {
	envs := map[string]string{
		"OPENFGA_PLAYGROUND_ENABLED":                  "true",
		"OPENFGA_PLAYGROUND_ADDR":                     "0.0.0.0:3000",
		"OPENFGA_MAX_CHECKS_PER_BATCH_CHECK":          "100",
		"OPENFGA_CHECK_ITERATOR_CACHE_ENABLED":        "false",
		"OPENFGA_LIST_OBJECTS_ITERATOR_CACHE_ENABLED": "false",
		"OPENFGA_MAX_TYPES_PER_AUTHORIZATION_MODEL":   "1000",
	}

	if os.Getenv(FGACacheEnvVar) == "true" {
		for k, v := range fgaCacheEnvs() {
			envs[k] = v
		}
	}

	return envs
}

// fgaCacheEnvs mirrors the openfga cache settings used in production deployments
func fgaCacheEnvs() map[string]string {
	return map[string]string{
		"OPENFGA_CHECK_CACHE_LIMIT":                       "50000",
		"OPENFGA_CHECK_ITERATOR_CACHE_ENABLED":            "true",
		"OPENFGA_CHECK_ITERATOR_CACHE_MAX_RESULTS":        "50000",
		"OPENFGA_CHECK_ITERATOR_CACHE_TTL":                "60s",
		"OPENFGA_CACHE_CONTROLLER_ENABLED":                "true",
		"OPENFGA_CACHE_CONTROLLER_TTL":                    "60s",
		"OPENFGA_LIST_OBJECTS_ITERATOR_CACHE_ENABLED":     "true",
		"OPENFGA_LIST_OBJECTS_ITERATOR_CACHE_MAX_RESULTS": "10000",
		"OPENFGA_LIST_OBJECTS_ITERATOR_CACHE_TTL":         "60s",
	}
}
