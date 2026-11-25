package vulnintel

// defaultMaxPerRun is the default number of OSV lookups one enrichment makes
const defaultMaxPerRun = 500

// Config configures vulnerability enrichment
type Config struct {
	// MaxPerRun caps the OSV lookups one enrichment makes, the rest wait for a later run while cached advisories keep applying
	MaxPerRun int `json:"maxperrun" koanf:"maxperrun" default:"500" jsonschema:"default=500,description=Maximum OSV advisory lookups per enrichment run"`
}

// maxPerRun returns the configured lookup cap, the default when unset
func (c Config) maxPerRun() int {
	if c.MaxPerRun <= 0 {
		return defaultMaxPerRun
	}

	return c.MaxPerRun
}
