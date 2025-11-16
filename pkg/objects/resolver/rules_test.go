package resolver

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/theopenlane/core/internal/objects"
	"github.com/theopenlane/core/pkg/models"
	"github.com/theopenlane/core/pkg/objects/storage"
	"github.com/theopenlane/eddy"
	"github.com/theopenlane/utils/contextx"
)

type stubBuilder struct {
	providerType string
	lastConfig   *storage.ProviderOptions
	lastOutput   storage.ProviderCredentials
}

func (b *stubBuilder) Build(_ context.Context, output storage.ProviderCredentials, cfg *storage.ProviderOptions) (storage.Provider, error) {
	b.lastOutput = output
	if cfg != nil {
		copy := cfg.Clone()
		b.lastConfig = copy
	} else {
		b.lastConfig = nil
	}
	return nil, nil
}

func (b *stubBuilder) ProviderType() string {
	return b.providerType
}

func TestConfigureProviderRulesDevMode(t *testing.T) {
	resolver := eddy.NewResolver[storage.Provider, storage.ProviderCredentials, *storage.ProviderOptions]()

	diskBuilder := &stubBuilder{providerType: "disk"}
	config := storage.ProviderConfig{
		DevMode: true,
		Providers: storage.Providers{
			// we build the disk provider even if disabled to ensure dev mode works - so test ensures the builder still constructs
			Disk: storage.ProviderConfigs{Enabled: false},
		},
	}

	configureProviderRules(
		resolver,
		WithProviderConfig(config),
		WithProviderBuilders(providerBuilders{
			s3:   &stubBuilder{providerType: "s3"},
			r2:   &stubBuilder{providerType: "r2"},
			disk: diskBuilder,
			db:   &stubBuilder{providerType: "db"},
		}),
		WithRuntimeOptions(serviceOptions{}),
	)

	option := resolver.Resolve(context.Background())
	require.True(t, option.IsPresent(), "expected dev mode rule to resolve")

	result := option.MustGet()
	require.Equal(t, diskBuilder, result.Builder, "expected disk builder for dev mode")
	require.NotNil(t, result.Config)
	require.Equal(t, objects.DefaultDevStorageBucket, result.Config.Bucket)
	require.Equal(t, objects.DefaultDevStorageBucket, result.Config.BasePath)
	extra, ok := result.Config.Extra("dev_mode")
	require.True(t, ok)
	require.Equal(t, true, extra)
}

func TestKnownProviderRule(t *testing.T) {
	ctx := contextx.With(context.Background(), objects.KnownProviderHint(storage.DiskProvider))
	resolver := eddy.NewResolver[storage.Provider, storage.ProviderCredentials, *storage.ProviderOptions]()

	diskBuilder := &stubBuilder{providerType: "disk"}
	config := storage.ProviderConfig{
		Providers: storage.Providers{
			Disk: storage.ProviderConfigs{
				Enabled:  true,
				Bucket:   "/mnt/storage",
				Endpoint: "http://local",
			},
		},
	}

	configureProviderRules(
		resolver,
		WithProviderConfig(config),
		WithProviderBuilders(providerBuilders{
			s3:   &stubBuilder{providerType: "s3"},
			r2:   &stubBuilder{providerType: "r2"},
			disk: diskBuilder,
			db:   &stubBuilder{providerType: "db"},
		}),
		WithRuntimeOptions(serviceOptions{}),
	)

	option := resolver.Resolve(ctx)
	require.True(t, option.IsPresent(), "expected known provider rule to resolve")

	result := option.MustGet()
	require.Equal(t, diskBuilder, result.Builder)
	require.Equal(t, "/mnt/storage", result.Config.Bucket)
	require.Equal(t, "/mnt/storage", result.Config.BasePath)
	require.Equal(t, "http://local", result.Config.LocalURL)
}

func TestModuleRules(t *testing.T) {
	ctx := contextx.With(context.Background(), objects.ModuleHint(models.CatalogTrustCenterModule))
	resolver := eddy.NewResolver[storage.Provider, storage.ProviderCredentials, *storage.ProviderOptions]()

	r2Builder := &stubBuilder{providerType: "r2"}
	config := storage.ProviderConfig{
		Providers: storage.Providers{
			CloudflareR2: storage.ProviderConfigs{
				Enabled: true,
				Bucket:  "tc-bucket",
			},
		},
	}

	configureProviderRules(
		resolver,
		WithProviderConfig(config),
		WithProviderBuilders(providerBuilders{
			s3:   &stubBuilder{providerType: "s3"},
			r2:   r2Builder,
			disk: &stubBuilder{providerType: "disk"},
			db:   &stubBuilder{providerType: "db"},
		}),
		WithRuntimeOptions(serviceOptions{}),
	)

	option := resolver.Resolve(ctx)
	require.True(t, option.IsPresent(), "expected module rule to resolve")

	result := option.MustGet()
	require.Equal(t, r2Builder, result.Builder)
	require.Equal(t, "tc-bucket", result.Config.Bucket)
}

func TestDefaultRuleSelectsFirstEnabledProvider(t *testing.T) {
	resolver := eddy.NewResolver[storage.Provider, storage.ProviderCredentials, *storage.ProviderOptions]()

	s3Builder := &stubBuilder{providerType: "s3"}
	r2Builder := &stubBuilder{providerType: "r2"}

	config := storage.ProviderConfig{
		Providers: storage.Providers{
			S3: storage.ProviderConfigs{
				Enabled: false,
			},
			CloudflareR2: storage.ProviderConfigs{
				Enabled: true,
				Bucket:  "r2-bucket",
			},
		},
	}

	configureProviderRules(
		resolver,
		WithProviderConfig(config),
		WithProviderBuilders(providerBuilders{
			s3:   s3Builder,
			r2:   r2Builder,
			disk: &stubBuilder{providerType: "disk"},
			db:   &stubBuilder{providerType: "db"},
		}),
		WithRuntimeOptions(serviceOptions{}),
	)

	option := resolver.Resolve(context.Background())
	require.True(t, option.IsPresent(), "expected default rule to resolve")

	result := option.MustGet()
	require.Equal(t, r2Builder, result.Builder, "expected first enabled provider to be used")
	require.Equal(t, "r2-bucket", result.Config.Bucket)
}

func TestDefaultRuleUsesS3WhenEnabled(t *testing.T) {
	resolver := eddy.NewResolver[storage.Provider, storage.ProviderCredentials, *storage.ProviderOptions]()

	s3Builder := &stubBuilder{providerType: "s3"}
	config := storage.ProviderConfig{
		Providers: storage.Providers{
			S3: storage.ProviderConfigs{
				Enabled: true,
				Bucket:  "default-bucket",
			},
		},
	}

	configureProviderRules(
		resolver,
		WithProviderConfig(config),
		WithProviderBuilders(providerBuilders{
			s3:   s3Builder,
			r2:   &stubBuilder{providerType: "r2"},
			disk: &stubBuilder{providerType: "disk"},
			db:   &stubBuilder{providerType: "db"},
		}),
		WithRuntimeOptions(serviceOptions{}),
	)

	option := resolver.Resolve(context.Background())
	require.True(t, option.IsPresent(), "expected default rule to resolve")

	result := option.MustGet()
	require.Equal(t, s3Builder, result.Builder)
	require.Equal(t, "default-bucket", result.Config.Bucket)
}

func TestProviderEnabledChecksConfig(t *testing.T) {
	rc := &ruleCoordinator{
		config: storage.ProviderConfig{
			Providers: storage.Providers{
				S3:   storage.ProviderConfigs{Enabled: true},
				Disk: storage.ProviderConfigs{Enabled: false},
			},
		},
	}

	require.True(t, rc.providerEnabled(storage.S3Provider))
	require.False(t, rc.providerEnabled(storage.DiskProvider))
}

func TestResolveProviderWithUnsupportedBuilder(t *testing.T) {
	resolver := eddy.NewResolver[storage.Provider, storage.ProviderCredentials, *storage.ProviderOptions]()
	rc := newRuleCoordinator(
		resolver,
		WithProviderConfig(storage.ProviderConfig{}),
		WithProviderBuilders(providerBuilders{}),
	)

	_, err := rc.resolveProviderWithBuilder(storage.ProviderType("unsupported"))
	require.Error(t, err)
	require.ErrorIs(t, err, errUnsupportedProvider)
}

func TestResolveProviderFromConfigCopiesOptions(t *testing.T) {
	resolver := eddy.NewResolver[storage.Provider, storage.ProviderCredentials, *storage.ProviderOptions]()
	rc := newRuleCoordinator(
		resolver,
		WithProviderConfig(storage.ProviderConfig{
			Providers: storage.Providers{
				S3: storage.ProviderConfigs{
					Enabled: true,
					Bucket:  "bucket",
					Region:  "us-east-1",
				},
			},
		}),
		WithProviderBuilders(providerBuilders{
			s3: &stubBuilder{providerType: "s3"},
		}),
	)

	resolved, err := rc.resolveProvider(storage.S3Provider)
	require.NoError(t, err)
	require.Equal(t, "bucket", resolved.Config.Bucket)
	require.Equal(t, "us-east-1", resolved.Config.Region)
}

func TestProviderResolveFromConfigDisabled(t *testing.T) {
	_, err := resolveProviderFromConfig(storage.S3Provider, storage.ProviderConfig{
		Providers: storage.Providers{
			S3: storage.ProviderConfigs{Enabled: false},
		},
	}, serviceOptions{})
	require.Error(t, err)
}

func TestHandleDevModeOptionClone(t *testing.T) {
	resolver := eddy.NewResolver[storage.Provider, storage.ProviderCredentials, *storage.ProviderOptions]()
	diskBuilder := &stubBuilder{providerType: "disk"}
	rc := newRuleCoordinator(
		resolver,
		WithProviderConfig(storage.ProviderConfig{
			DevMode: true,
			Providers: storage.Providers{
				Disk: storage.ProviderConfigs{Enabled: false},
			},
		}),
		WithProviderBuilders(providerBuilders{
			disk: diskBuilder,
		}),
	)

	require.True(t, rc.handleDevMode())

	option := resolver.Resolve(context.Background())
	require.True(t, option.IsPresent())

	result := option.MustGet()
	require.NotNil(t, result.Config)
	// ensure options cloned on each invocation
	result.Config.Apply(storage.WithExtra("mutated", true))

	option = resolver.Resolve(context.Background())
	require.True(t, option.IsPresent())
	_, ok := option.MustGet().Config.Extra("mutated")
	require.False(t, ok)
}

func TestDefaultRuleSkipsDisabledProviders(t *testing.T) {
	resolver := eddy.NewResolver[storage.Provider, storage.ProviderCredentials, *storage.ProviderOptions]()
	rc := newRuleCoordinator(
		resolver,
		WithProviderConfig(storage.ProviderConfig{
			Providers: storage.Providers{
				S3:           storage.ProviderConfigs{Enabled: false},
				CloudflareR2: storage.ProviderConfigs{Enabled: true},
			},
		}),
		WithProviderBuilders(providerBuilders{
			s3: &stubBuilder{providerType: "s3"},
			r2: &stubBuilder{providerType: "r2"},
		}),
	)

	rc.addDefaultProviderRule()

	option := resolver.Resolve(context.Background())
	require.True(t, option.IsPresent())
	require.Equal(t, "r2", option.MustGet().Builder.ProviderType())
}
