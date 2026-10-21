package gemini

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/theopenlane/precis/docextract"
	geminiprovider "github.com/theopenlane/precis/docextract/gemini"
	"github.com/theopenlane/precis/docextract/gemini/modelarmor"
	"github.com/theopenlane/precis/docextract/kinds/soc2"

	"github.com/theopenlane/core/v2/internal/docextract/soc2schema"
)

// runtimeClientBuilder returns a build function that constructs the Gemini client for the
// runtime (system) path using the operator-owned API key
func runtimeClientBuilder() func(context.Context, json.RawMessage) (any, error) {
	return func(ctx context.Context, config json.RawMessage) (any, error) {
		var cfg RuntimeConfig
		if err := json.Unmarshal(config, &cfg); err != nil {
			return nil, fmt.Errorf("%w: %w", ErrRuntimeConfigDecode, err)
		}

		if !cfg.Provisioned() {
			return nil, ErrRuntimeConfigInvalid
		}

		opts := []geminiprovider.Option{geminiprovider.WithBackend(cfg.GenAIBackend())}

		// an empty model keeps the provider default
		if cfg.Model != "" {
			opts = append(opts, geminiprovider.WithModel(cfg.Model))
		}

		provider, err := geminiprovider.New(ctx, geminiprovider.Config{
			APIKey:   cfg.APIKey,
			Project:  cfg.Project,
			Location: cfg.Location,
		}, opts...)
		if err != nil {
			return nil, err
		}

		screener, err := newScreener(ctx, cfg)
		if err != nil {
			return nil, err
		}

		client, err := docextract.NewClient(docextract.Config{
			Provider:          provider,
			Screener:          screener,
			SystemInstruction: cfg.Prompts.SystemInstruction,
		})
		if err != nil {
			return nil, err
		}

		return &Client{Client: client, SOC2: soc2.NewKind(cfg.Prompts.SOC2, soc2schema.Schemas)}, nil
	}
}

// newScreener builds the Model Armor screener when a template is configured, authenticating with the
// same application default credentials the Vertex backend uses; nil leaves screening off
func newScreener(ctx context.Context, cfg RuntimeConfig) (docextract.Screener, error) {
	if cfg.ModelArmorTemplate == "" {
		return nil, nil
	}

	// a failed build returns a nil interface rather than a typed nil client
	screener, err := modelarmor.New(ctx, cfg.ModelArmorTemplate)
	if err != nil {
		return nil, err
	}

	return screener, nil
}
