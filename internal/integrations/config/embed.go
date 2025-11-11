package config

import "embed"

// ProvidersFS embeds the provider configuration files.
//
//go:embed providers/*.json
var ProvidersFS embed.FS
