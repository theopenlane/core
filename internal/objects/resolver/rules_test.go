package resolver

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/knadh/koanf/parsers/yaml"
	"github.com/knadh/koanf/providers/file"
	"github.com/knadh/koanf/v2"
	"gotest.tools/v3/assert"
	is "gotest.tools/v3/assert/cmp"

	"github.com/theopenlane/core/common/enums"
	"github.com/theopenlane/core/common/models"
	"github.com/theopenlane/core/v2/internal/objects"
	"github.com/theopenlane/core/v2/pkg/objects/storage"
	"github.com/theopenlane/eddy"
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
			Disk: storage.DiskConfig{ProviderCommon: storage.ProviderCommon{Enabled: false}},
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
	assert.Check(t, option.IsPresent(), "expected dev mode rule to resolve")

	result := option.MustGet()
	assert.Check(t, is.Equal(diskBuilder, result.Builder), "expected disk builder for dev mode")
	assert.Check(t, result.Config != nil)
	assert.Check(t, is.Equal(objects.DefaultDevStorageBucket, result.Config.Bucket))
	extra, ok := result.Config.Extra("dev_mode")
	assert.Check(t, ok)
	assert.Check(t, is.Equal(true, extra))
}

func TestKnownProviderRule(t *testing.T) {
	ctx := objects.WithKnownProviderHint(context.Background(), storage.DiskProvider)
	resolver := eddy.NewResolver[storage.Provider, storage.ProviderCredentials, *storage.ProviderOptions]()

	diskBuilder := &stubBuilder{providerType: "disk"}
	config := storage.ProviderConfig{
		Providers: storage.Providers{
			Disk: storage.DiskConfig{
				ProviderCommon: storage.ProviderCommon{
					Enabled: true,
					Bucket:  "/mnt/storage",
				},
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
	assert.Check(t, option.IsPresent(), "expected known provider rule to resolve")

	result := option.MustGet()
	assert.Check(t, is.Equal(diskBuilder, result.Builder))
	assert.Check(t, is.Equal("/mnt/storage", result.Config.Bucket))
	assert.Check(t, is.Equal("http://local", result.Config.LocalURL))
}

func TestModuleRules(t *testing.T) {
	ctx := objects.WithModuleHint(context.Background(), models.CatalogTrustCenterModule)
	resolver := eddy.NewResolver[storage.Provider, storage.ProviderCredentials, *storage.ProviderOptions]()

	r2Builder := &stubBuilder{providerType: "r2"}
	config := storage.ProviderConfig{
		Providers: storage.Providers{
			R2: storage.R2Config{ProviderCommon: storage.ProviderCommon{
				Enabled: true,
				Bucket:  "tc-bucket",
			}},
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
	assert.Check(t, option.IsPresent(), "expected module rule to resolve")

	result := option.MustGet()
	assert.Check(t, is.Equal(r2Builder, result.Builder))
	assert.Check(t, is.Equal("tc-bucket", result.Config.Bucket))
}

func TestTemplateKindRule(t *testing.T) {
	ctx := objects.WithTemplateKindHint(context.Background(), enums.TemplateKindTrustCenterNda)
	resolver := eddy.NewResolver[storage.Provider, storage.ProviderCredentials, *storage.ProviderOptions]()

	r2Builder := &stubBuilder{providerType: "r2"}
	config := storage.ProviderConfig{
		Providers: storage.Providers{
			R2: storage.R2Config{ProviderCommon: storage.ProviderCommon{
				Enabled: true,
				Bucket:  "nda-bucket",
			}},
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
	assert.Check(t, option.IsPresent(), "expected template kind rule to resolve")

	result := option.MustGet()
	assert.Check(t, is.Equal(r2Builder, result.Builder))
	assert.Check(t, is.Equal("nda-bucket", result.Config.Bucket))
}

func TestDefaultRuleSelectsFirstEnabledProvider(t *testing.T) {
	resolver := eddy.NewResolver[storage.Provider, storage.ProviderCredentials, *storage.ProviderOptions]()

	s3Builder := &stubBuilder{providerType: "s3"}
	r2Builder := &stubBuilder{providerType: "r2"}

	config := storage.ProviderConfig{
		Providers: storage.Providers{
			S3: storage.S3Config{ProviderCommon: storage.ProviderCommon{
				Enabled: false,
			}},
			R2: storage.R2Config{ProviderCommon: storage.ProviderCommon{
				Enabled: true,
				Bucket:  "r2-bucket",
			}},
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
	assert.Check(t, option.IsPresent(), "expected default rule to resolve")

	result := option.MustGet()
	assert.Check(t, is.Equal(r2Builder, result.Builder), "expected first enabled provider to be used")
	assert.Check(t, is.Equal("r2-bucket", result.Config.Bucket))
}

func TestDefaultRuleUsesS3WhenEnabled(t *testing.T) {
	resolver := eddy.NewResolver[storage.Provider, storage.ProviderCredentials, *storage.ProviderOptions]()

	s3Builder := &stubBuilder{providerType: "s3"}
	config := storage.ProviderConfig{
		Providers: storage.Providers{
			S3: storage.S3Config{ProviderCommon: storage.ProviderCommon{
				Enabled: true,
				Bucket:  "default-bucket",
			}},
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
	assert.Check(t, option.IsPresent(), "expected default rule to resolve")

	result := option.MustGet()
	assert.Check(t, is.Equal(s3Builder, result.Builder))
	assert.Check(t, is.Equal("default-bucket", result.Config.Bucket))
}

func TestProviderEnabledChecksConfig(t *testing.T) {
	rc := &ruleCoordinator{
		config: storage.ProviderConfig{
			Providers: storage.Providers{
				S3:   storage.S3Config{ProviderCommon: storage.ProviderCommon{Enabled: true}},
				Disk: storage.DiskConfig{ProviderCommon: storage.ProviderCommon{Enabled: false}},
			},
		},
	}

	assert.Check(t, rc.providerEnabled(storage.S3Provider))
	assert.Check(t, !rc.providerEnabled(storage.DiskProvider))
}

func TestResolveProviderWithUnsupportedBuilder(t *testing.T) {
	resolver := eddy.NewResolver[storage.Provider, storage.ProviderCredentials, *storage.ProviderOptions]()
	rc := newRuleCoordinator(
		resolver,
		WithProviderConfig(storage.ProviderConfig{}),
		WithProviderBuilders(providerBuilders{}),
	)

	_, err := rc.resolveProviderWithBuilder(storage.ProviderType("unsupported"))
	assert.Check(t, err != nil)
	assert.Check(t, is.ErrorIs(err, errUnsupportedProvider))
}

func TestResolveProviderFromConfigCopiesOptions(t *testing.T) {
	resolver := eddy.NewResolver[storage.Provider, storage.ProviderCredentials, *storage.ProviderOptions]()
	rc := newRuleCoordinator(
		resolver,
		WithProviderConfig(storage.ProviderConfig{
			Providers: storage.Providers{
				S3: storage.S3Config{
					ProviderCommon: storage.ProviderCommon{
						Enabled: true,
						Bucket:  "bucket",
					},
					Region: "us-east-1",
				},
			},
		}),
		WithProviderBuilders(providerBuilders{
			s3: &stubBuilder{providerType: "s3"},
		}),
	)

	resolved, err := rc.resolveProvider(storage.S3Provider)
	assert.Check(t, err)
	assert.Check(t, is.Equal("bucket", resolved.Config.Bucket))
	assert.Check(t, is.Equal("us-east-1", resolved.Config.Region))
}

func TestProviderResolveFromConfigDisabled(t *testing.T) {
	_, err := resolveProviderFromConfig(storage.S3Provider, storage.ProviderConfig{
		Providers: storage.Providers{
			S3: storage.S3Config{ProviderCommon: storage.ProviderCommon{Enabled: false}},
		},
	}, serviceOptions{})
	assert.Check(t, err != nil)
}

func TestHandleDevModeOptionClone(t *testing.T) {
	resolver := eddy.NewResolver[storage.Provider, storage.ProviderCredentials, *storage.ProviderOptions]()
	diskBuilder := &stubBuilder{providerType: "disk"}
	rc := newRuleCoordinator(
		resolver,
		WithProviderConfig(storage.ProviderConfig{
			DevMode: true,
			Providers: storage.Providers{
				Disk: storage.DiskConfig{ProviderCommon: storage.ProviderCommon{Enabled: false}},
			},
		}),
		WithProviderBuilders(providerBuilders{
			disk: diskBuilder,
		}),
	)

	assert.Check(t, rc.handleDevMode())

	option := resolver.Resolve(context.Background())
	assert.Check(t, option.IsPresent())

	result := option.MustGet()
	assert.Check(t, result.Config != nil)
	// ensure options cloned on each invocation
	result.Config.Apply(storage.WithExtra("mutated", true))

	option = resolver.Resolve(context.Background())
	assert.Check(t, option.IsPresent())
	_, ok := option.MustGet().Config.Extra("mutated")
	assert.Check(t, !ok)
}

func TestDefaultRuleSkipsDisabledProviders(t *testing.T) {
	resolver := eddy.NewResolver[storage.Provider, storage.ProviderCredentials, *storage.ProviderOptions]()
	rc := newRuleCoordinator(
		resolver,
		WithProviderConfig(storage.ProviderConfig{
			Providers: storage.Providers{
				S3: storage.S3Config{ProviderCommon: storage.ProviderCommon{Enabled: true}},
				R2: storage.R2Config{ProviderCommon: storage.ProviderCommon{Enabled: false}},
			},
		}),
		WithProviderBuilders(providerBuilders{
			s3: &stubBuilder{providerType: "s3"},
			r2: &stubBuilder{providerType: "r2"},
		}),
	)

	rc.addDefaultProviderRule()

	option := resolver.Resolve(context.Background())
	assert.Check(t, option.IsPresent())
	assert.Check(t, is.Equal("s3", option.MustGet().Builder.ProviderType()))
}

type oldProvidersStructWithCloudflareR2 struct {
	S3           storage.S3Config       `json:"s3" koanf:"s3"`
	CloudflareR2 storage.R2Config       `json:"cloudflarer2" koanf:"cloudflarer2"`
	Disk         storage.DiskConfig     `json:"disk" koanf:"disk"`
	Database     storage.DatabaseConfig `json:"database" koanf:"database"`
}

type oldProviderConfigWithCloudflareR2 struct {
	Enabled     bool                               `json:"enabled" koanf:"enabled"`
	Keys        []string                           `json:"keys" koanf:"keys"`
	MaxSizeMB   int64                              `json:"maxsizemb" koanf:"maxsizemb"`
	MaxMemoryMB int64                              `json:"maxmemorymb" koanf:"maxmemorymb"`
	DevMode     bool                               `json:"devmode" koanf:"devmode"`
	Providers   oldProvidersStructWithCloudflareR2 `json:"providers" koanf:"providers"`
}

func TestModuleRuleR2ProviderConstantMatchesR2ConfigField(t *testing.T) {
	ctx := objects.WithModuleHint(context.Background(), models.CatalogTrustCenterModule)
	resolver := eddy.NewResolver[storage.Provider, storage.ProviderCredentials, *storage.ProviderOptions]()

	s3Builder := &stubBuilder{providerType: "s3"}

	yamlConfig := `
enabled: true
providers:
  s3:
    enabled: true
    bucket: "opln"
    region: "us-east-2"
  cloudflarer2:
    enabled: true
    bucket: "ol-trust-center"
`

	path := filepath.Join(t.TempDir(), "config.yaml")
	assert.NilError(t, os.WriteFile(path, []byte(yamlConfig), 0o600))

	k := koanf.New(".")
	assert.NilError(t, k.Load(file.Provider(path), yaml.Parser()))

	var oldStyleConfig oldProviderConfigWithCloudflareR2
	assert.NilError(t, k.Unmarshal("", &oldStyleConfig))

	assert.Check(t, oldStyleConfig.Providers.CloudflareR2.Enabled, "CloudflareR2 config populated from YAML")
	assert.Check(t, is.Equal("ol-trust-center", oldStyleConfig.Providers.CloudflareR2.Bucket))

	actualConfig := storage.ProviderConfig{
		Enabled: true,
		Providers: storage.Providers{
			S3: oldStyleConfig.Providers.S3,
			R2: oldStyleConfig.Providers.CloudflareR2,
		},
	}

	cloudflareR2BuilderReportingCloudflareR2Type := &stubBuilder{providerType: "cloudflarer2"}

	rc := newRuleCoordinator(
		resolver,
		WithProviderConfig(actualConfig),
		WithProviderBuilders(providerBuilders{
			s3:   s3Builder,
			r2:   cloudflareR2BuilderReportingCloudflareR2Type,
			disk: &stubBuilder{providerType: "disk"},
			db:   &stubBuilder{providerType: "db"},
		}),
		WithRuntimeOptions(serviceOptions{}),
	)

	rc.addKnownProviderRule()
	rc.addModuleRule(models.CatalogTrustCenterModule, storage.R2Provider)
	rc.addDefaultProviderRule()

	assert.Check(t, actualConfig.Providers.R2.Enabled, "R2 config is enabled from CloudflareR2 YAML")
	assert.Check(t, is.Equal("ol-trust-center", actualConfig.Providers.R2.Bucket), "R2 bucket populated from cloudflarer2 config")

	option := resolver.Resolve(ctx)
	assert.Check(t, option.IsPresent())

	result := option.MustGet()
	assert.Check(t, is.Equal(cloudflareR2BuilderReportingCloudflareR2Type, result.Builder), "module rule with storage.R2Provider ('r2') matches config.Providers.R2, resolver builds provider successfully")
	assert.Check(t, is.Equal("ol-trust-center", result.Config.Bucket), "trust center documents go to R2 as expected")
	assert.Check(t, is.Equal("cloudflarer2", result.Builder.ProviderType()), "but provider reports 'cloudflarer2' type - validateProviderType would catch this mismatch")
}

// backupRuleResolver wires the rule chain for a config and returns the resolver
func backupRuleResolver(config storage.ProviderConfig, builders providerBuilders) *eddy.Resolver[storage.Provider, storage.ProviderCredentials, *storage.ProviderOptions] {
	resolver := eddy.NewResolver[storage.Provider, storage.ProviderCredentials, *storage.ProviderOptions]()

	configureProviderRules(
		resolver,
		WithProviderConfig(config),
		WithProviderBuilders(builders),
		WithRuntimeOptions(serviceOptions{}),
	)

	return resolver
}

func TestBackupRuleResolvesDifferentProvider(t *testing.T) {
	r2Builder := &stubBuilder{providerType: "r2"}
	config := storage.ProviderConfig{
		Providers: storage.Providers{
			S3: storage.S3Config{
				ProviderCommon: storage.ProviderCommon{
					Enabled: true,
					Bucket:  "live-bucket",
					Backup: &storage.BackupConfig{
						Enabled:  true,
						Provider: storage.R2Provider,
					},
				},
				Region: "us-east-1",
			},
			R2: storage.R2Config{ProviderCommon: storage.ProviderCommon{
				Enabled: true,
				Bucket:  "r2-bucket",
			}},
		},
	}

	resolver := backupRuleResolver(config, providerBuilders{
		s3:   &stubBuilder{providerType: "s3"},
		r2:   r2Builder,
		disk: &stubBuilder{providerType: "disk"},
		db:   &stubBuilder{providerType: "db"},
	})

	ctx := objects.WithBackupSourceHint(context.Background(), storage.S3Provider)

	option := resolver.Resolve(ctx)
	assert.Check(t, option.IsPresent(), "expected backup rule to resolve")

	result := option.MustGet()
	assert.Check(t, is.Equal(r2Builder, result.Builder))
	assert.Check(t, is.Equal("r2-bucket-backup", result.Config.Bucket))

	marker, ok := result.Config.Extra(storage.BackupTargetExtraKey)
	assert.Check(t, ok, "expected the resolved backup to carry its marker")
	assert.Check(t, is.Equal(true, marker))
}

func TestBackupRuleResolvesDifferentRegionOnSameProvider(t *testing.T) {
	s3Builder := &stubBuilder{providerType: "s3"}
	config := storage.ProviderConfig{
		Providers: storage.Providers{
			S3: storage.S3Config{
				ProviderCommon: storage.ProviderCommon{
					Enabled: true,
					Bucket:  "live-bucket",
					Backup: &storage.BackupConfig{
						Enabled: true,
						Region:  "us-west-2",
					},
				},
				Region: "us-east-1",
			},
		},
	}

	resolver := backupRuleResolver(config, providerBuilders{
		s3:   s3Builder,
		r2:   &stubBuilder{providerType: "r2"},
		disk: &stubBuilder{providerType: "disk"},
		db:   &stubBuilder{providerType: "db"},
	})

	ctx := objects.WithBackupSourceHint(context.Background(), storage.S3Provider)

	option := resolver.Resolve(ctx)
	assert.Check(t, option.IsPresent(), "expected backup rule to resolve")

	result := option.MustGet()
	assert.Check(t, is.Equal(s3Builder, result.Builder))
	assert.Check(t, is.Equal("live-bucket-backup", result.Config.Bucket))
	assert.Check(t, is.Equal("us-west-2", result.Config.Region))

	// the live provider keeps its own region, so the two clients never share a configuration
	liveOption := resolver.Resolve(objects.WithKnownProviderHint(context.Background(), storage.S3Provider))
	assert.Check(t, liveOption.IsPresent())
	assert.Check(t, is.Equal("us-east-1", liveOption.MustGet().Config.Region))
	assert.Check(t, is.Equal("live-bucket", liveOption.MustGet().Config.Bucket))
}

func TestBackupRuleAbsentWithoutBackupConfig(t *testing.T) {
	config := storage.ProviderConfig{
		Providers: storage.Providers{
			S3: storage.S3Config{ProviderCommon: storage.ProviderCommon{Enabled: true, Bucket: "live-bucket"}},
		},
	}

	resolver := backupRuleResolver(config, providerBuilders{
		s3:   &stubBuilder{providerType: "s3"},
		r2:   &stubBuilder{providerType: "r2"},
		disk: &stubBuilder{providerType: "disk"},
		db:   &stubBuilder{providerType: "db"},
	})

	ctx := objects.WithBackupSourceHint(context.Background(), storage.S3Provider)

	// the default rule still answers, but without the marker, so a caller never mistakes the live
	// provider for a backup destination
	result := resolver.Resolve(ctx).MustGet()
	_, ok := result.Config.Extra(storage.BackupTargetExtraKey)
	assert.Check(t, !ok)
	assert.Check(t, is.Equal("live-bucket", result.Config.Bucket))
}

func TestBackupRuleSuffixesDefaultDiskBucket(t *testing.T) {
	diskBuilder := &stubBuilder{providerType: "disk"}
	config := storage.ProviderConfig{
		Providers: storage.Providers{
			Disk: storage.DiskConfig{ProviderCommon: storage.ProviderCommon{
				Enabled: true,
				Backup:  &storage.BackupConfig{Enabled: true},
			}},
		},
	}

	resolver := backupRuleResolver(config, providerBuilders{
		s3:   &stubBuilder{providerType: "s3"},
		r2:   &stubBuilder{providerType: "r2"},
		disk: diskBuilder,
		db:   &stubBuilder{providerType: "db"},
	})

	ctx := objects.WithBackupSourceHint(context.Background(), storage.DiskProvider)

	option := resolver.Resolve(ctx)
	assert.Check(t, option.IsPresent(), "expected backup rule to resolve")

	result := option.MustGet()
	assert.Check(t, is.Equal(diskBuilder, result.Builder))
	assert.Check(t, is.Equal(objects.DefaultDevStorageBucket+storage.BackupBucketSuffix, result.Config.Bucket))
}

func TestDatabaseOptionsForceProxyPresign(t *testing.T) {
	ctx := objects.WithKnownProviderHint(context.Background(), storage.DatabaseProvider)
	resolver := eddy.NewResolver[storage.Provider, storage.ProviderCredentials, *storage.ProviderOptions]()

	dbBuilder := &stubBuilder{providerType: "db"}
	config := storage.ProviderConfig{
		Providers: storage.Providers{
			Database: storage.DatabaseConfig{ProviderCommon: storage.ProviderCommon{Enabled: true}},
		},
	}

	configureProviderRules(
		resolver,
		WithProviderConfig(config),
		WithProviderBuilders(providerBuilders{
			s3:   &stubBuilder{providerType: "s3"},
			r2:   &stubBuilder{providerType: "r2"},
			disk: &stubBuilder{providerType: "disk"},
			db:   dbBuilder,
		}),
		WithRuntimeOptions(serviceOptions{}),
	)

	option := resolver.Resolve(ctx)
	assert.Check(t, option.IsPresent(), "expected known provider rule to resolve")

	result := option.MustGet()
	assert.Check(t, is.Equal(dbBuilder, result.Builder))
	assert.Check(t, result.Config.ProxyPresignEnabled)
}
