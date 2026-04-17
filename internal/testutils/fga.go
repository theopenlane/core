//go:build test

package testutils

// GetDefaultFGAEnvs returns env variables required by OpenFGA tests.
func GetDefaultFGAEnvs() map[string]string {
	return map[string]string{
		"OPENFGA_PLAYGROUND_ENABLED":                  "true",
		"OPENFGA_PLAYGROUND_ADDR":                     "0.0.0.0:3000",
		"OPENFGA_MAX_CHECKS_PER_BATCH_CHECK":          "100",
		"OPENFGA_CHECK_ITERATOR_CACHE_ENABLED":        "false",
		"OPENFGA_LIST_OBJECTS_ITERATOR_CACHE_ENABLED": "false",
	}
}
