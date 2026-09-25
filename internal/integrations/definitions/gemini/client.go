package gemini

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/theopenlane/core/v2/internal/integrations/types"
	"github.com/theopenlane/core/v2/pkg/docextract"
	"github.com/theopenlane/core/v2/pkg/docextract/soc2"
	"github.com/theopenlane/core/v2/pkg/modelarmor"
)

// runtimeOnlyClient rejects installation-bound builds since the Gemini client only exists on the runtime path
func runtimeOnlyClient(_ context.Context, _ types.ClientBuildRequest) (any, error) {
	return nil, ErrRuntimeOnly
}

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

		client, err := docextract.NewClient(ctx,
			docextract.WithBackend(cfg.GenAIBackend()),
			docextract.WithAPIKey(cfg.APIKey),
			docextract.WithProject(cfg.Project),
			docextract.WithLocation(cfg.Location),
			docextract.WithModel(cfg.Model),
			docextract.WithSystemInstruction(cfg.Prompts.SystemInstruction),
		)
		if err != nil {
			return nil, err
		}

		screener, err := newScreener(ctx, cfg)
		if err != nil {
			return nil, err
		}

		return &Client{Client: client, SOC2: soc2.NewKind(cfg.Prompts.SOC2), Screener: screener}, nil
	}
}

// newScreener builds the Model Armor client when a template is configured, authenticating with the
// same application default credentials the Vertex backend uses
func newScreener(ctx context.Context, cfg RuntimeConfig) (*modelarmor.Client, error) {
	if cfg.ModelArmorTemplate == "" {
		return nil, nil
	}

	return modelarmor.New(ctx, cfg.ModelArmorTemplate)
}
