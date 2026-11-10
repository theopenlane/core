package storage

import (
	"testing"

	"gotest.tools/v3/assert"
	is "gotest.tools/v3/assert/cmp"
)

func TestScheme(t *testing.T) {
	assert.Check(t, is.Equal("s3://", Scheme(S3Provider)))
	assert.Check(t, is.Equal("r2://", Scheme(R2Provider)))
	assert.Check(t, is.Equal("gs://", Scheme(GCSProvider)))
	assert.Check(t, is.Equal("file://", Scheme(DiskProvider)))
	assert.Check(t, is.Equal("database://", Scheme(DatabaseProvider)))
	assert.Check(t, is.Len(Scheme(ProviderType("ftp")), 0))
}

func TestNewProviderOptions(t *testing.T) {
	t.Run("no options", func(t *testing.T) {
		opts := NewProviderOptions()
		assert.Assert(t, opts != nil)
		assert.Check(t, is.Len(opts.Bucket, 0))
		assert.Check(t, is.Len(opts.Region, 0))
		assert.Check(t, is.Len(opts.Endpoint, 0))
	})

	t.Run("with multiple options", func(t *testing.T) {
		opts := NewProviderOptions(
			WithBucket("test-bucket"),
			WithRegion("us-west-2"),
			WithEndpoint("https://s3.example.com"),
		)
		assert.Assert(t, opts != nil)
		assert.Check(t, is.Equal("test-bucket", opts.Bucket))
		assert.Check(t, is.Equal("us-west-2", opts.Region))
		assert.Check(t, is.Equal("https://s3.example.com", opts.Endpoint))
	})
}

func TestProviderOptions_Apply(t *testing.T) {
	t.Run("apply options", func(t *testing.T) {
		opts := &ProviderOptions{}
		opts.Apply(
			WithBucket("my-bucket"),
			WithRegion("eu-west-1"),
		)
		assert.Check(t, is.Equal("my-bucket", opts.Bucket))
		assert.Check(t, is.Equal("eu-west-1", opts.Region))
	})

	t.Run("apply nil option", func(t *testing.T) {
		opts := &ProviderOptions{}
		opts.Apply(nil)
		assert.Check(t, is.Len(opts.Bucket, 0))
	})
}

func TestProviderOptions_Clone(t *testing.T) {
	t.Run("clone with values", func(t *testing.T) {
		original := &ProviderOptions{
			Bucket:   "test-bucket",
			Region:   "us-east-1",
			Endpoint: "https://endpoint.com",
			LocalURL: "http://localhost:8080",
			Credentials: ProviderCredentials{
				AccessKeyID:     "key",
				SecretAccessKey: "secret",
			},
		}

		cloned := original.Clone()
		assert.Assert(t, cloned != nil)
		assert.Check(t, is.Equal(original.Bucket, cloned.Bucket))
		assert.Check(t, is.Equal(original.Region, cloned.Region))
		assert.Check(t, is.Equal(original.Endpoint, cloned.Endpoint))
		assert.Check(t, is.Equal(original.LocalURL, cloned.LocalURL))
		assert.Check(t, is.DeepEqual(original.Credentials, cloned.Credentials))

		cloned.Bucket = "modified"
		assert.Check(t, original.Bucket != cloned.Bucket)
	})

	t.Run("clone with extras", func(t *testing.T) {
		original := &ProviderOptions{
			Bucket: "test-bucket",
		}
		original.Apply(WithExtra("key1", "value1"), WithExtra("key2", 42))

		cloned := original.Clone()
		assert.Assert(t, cloned != nil)

		val, ok := cloned.Extra("key1")
		assert.Check(t, ok)
		assert.Check(t, is.Equal("value1", val))

		val, ok = cloned.Extra("key2")
		assert.Check(t, ok)
		assert.Check(t, is.Equal(42, val))
	})

	t.Run("clone nil", func(t *testing.T) {
		var opts *ProviderOptions
		cloned := opts.Clone()
		assert.Check(t, is.Nil(cloned))
	})
}

func TestWithCredentials(t *testing.T) {
	creds := ProviderCredentials{
		AccessKeyID:     "access-key",
		SecretAccessKey: "secret-key",
		AccountID:       "account-456",
	}

	opts := NewProviderOptions(WithCredentials(creds))
	assert.Check(t, is.DeepEqual(creds, opts.Credentials))
	assert.Check(t, is.Equal("access-key", opts.Credentials.AccessKeyID))
	assert.Check(t, is.Equal("secret-key", opts.Credentials.SecretAccessKey))
}

func TestWithBucket(t *testing.T) {
	opts := NewProviderOptions(WithBucket("my-bucket"))
	assert.Check(t, is.Equal("my-bucket", opts.Bucket))
}

func TestWithRegion(t *testing.T) {
	opts := NewProviderOptions(WithRegion("ap-south-1"))
	assert.Check(t, is.Equal("ap-south-1", opts.Region))
}

func TestWithEndpoint(t *testing.T) {
	opts := NewProviderOptions(WithEndpoint("https://custom.endpoint.com"))
	assert.Check(t, is.Equal("https://custom.endpoint.com", opts.Endpoint))
}

func TestWithLocalURL(t *testing.T) {
	opts := NewProviderOptions(WithLocalURL("http://localhost:9000"))
	assert.Check(t, is.Equal("http://localhost:9000", opts.LocalURL))
}

func TestWithExtra(t *testing.T) {
	t.Run("add single extra", func(t *testing.T) {
		opts := NewProviderOptions(WithExtra("timeout", 30))
		val, ok := opts.Extra("timeout")
		assert.Check(t, ok)
		assert.Check(t, is.Equal(30, val))
	})

	t.Run("add multiple extras", func(t *testing.T) {
		opts := NewProviderOptions(
			WithExtra("timeout", 30),
			WithExtra("retry", true),
			WithExtra("name", "test-provider"),
		)

		val, ok := opts.Extra("timeout")
		assert.Check(t, ok)
		assert.Check(t, is.Equal(30, val))

		val, ok = opts.Extra("retry")
		assert.Check(t, ok)
		assert.Check(t, is.Equal(true, val))

		val, ok = opts.Extra("name")
		assert.Check(t, ok)
		assert.Check(t, is.Equal("test-provider", val))
	})

	t.Run("overwrite extra", func(t *testing.T) {
		opts := NewProviderOptions(
			WithExtra("key", "old-value"),
			WithExtra("key", "new-value"),
		)

		val, ok := opts.Extra("key")
		assert.Check(t, ok)
		assert.Check(t, is.Equal("new-value", val))
	})
}

func TestProviderOptions_Extra(t *testing.T) {
	t.Run("existing key", func(t *testing.T) {
		opts := NewProviderOptions(WithExtra("key", "value"))
		val, ok := opts.Extra("key")
		assert.Check(t, ok)
		assert.Check(t, is.Equal("value", val))
	})

	t.Run("non-existent key", func(t *testing.T) {
		opts := NewProviderOptions()
		val, ok := opts.Extra("nonexistent")
		assert.Check(t, !ok)
		assert.Check(t, is.Nil(val))
	})

	t.Run("nil options", func(t *testing.T) {
		var opts *ProviderOptions
		val, ok := opts.Extra("key")
		assert.Check(t, !ok)
		assert.Check(t, is.Nil(val))
	})

	t.Run("no extras", func(t *testing.T) {
		opts := &ProviderOptions{Bucket: "test"}
		val, ok := opts.Extra("key")
		assert.Check(t, !ok)
		assert.Check(t, is.Nil(val))
	})
}

func TestProviderOptions_BucketFor(t *testing.T) {
	opts := NewProviderOptions(WithBucket("own-bucket"))

	t.Run("explicit bucket wins", func(t *testing.T) {
		assert.Check(t, is.Equal("other-bucket", opts.BucketFor("other-bucket")))
	})

	t.Run("empty bucket falls back to the provider bucket", func(t *testing.T) {
		assert.Check(t, is.Equal("own-bucket", opts.BucketFor("")))
	})

	t.Run("no provider bucket", func(t *testing.T) {
		assert.Check(t, is.Len(NewProviderOptions().BucketFor(""), 0))
	})
}

func TestProviderOptions_CombinedOptions(t *testing.T) {
	t.Run("all options together", func(t *testing.T) {
		creds := ProviderCredentials{
			AccessKeyID:     "key",
			SecretAccessKey: "secret",
		}

		opts := NewProviderOptions(
			WithCredentials(creds),
			WithBucket("bucket"),
			WithRegion("region"),
			WithEndpoint("endpoint"),
			WithLocalURL("http://local"),
			WithExtra("custom", "value"),
		)

		assert.Check(t, is.Equal("key", opts.Credentials.AccessKeyID))
		assert.Check(t, is.Equal("bucket", opts.Bucket))
		assert.Check(t, is.Equal("region", opts.Region))
		assert.Check(t, is.Equal("endpoint", opts.Endpoint))
		assert.Check(t, is.Equal("http://local", opts.LocalURL))

		val, ok := opts.Extra("custom")
		assert.Check(t, ok)
		assert.Check(t, is.Equal("value", val))
	})
}

func TestProvidersByType(t *testing.T) {
	providers := Providers{
		S3:       S3Config{ProviderCommon: ProviderCommon{Enabled: true, Bucket: "s3-bucket"}, Region: "us-east-1"},
		R2:       R2Config{ProviderCommon: ProviderCommon{Bucket: "r2-bucket"}},
		GCS:      GCSConfig{ProviderCommon: ProviderCommon{Enabled: true, Bucket: "gcs-bucket"}, ProjectID: "project"},
		Disk:     DiskConfig{ProviderCommon: ProviderCommon{EnsureAvailable: true, Bucket: "/data"}},
		Database: DatabaseConfig{ProviderCommon: ProviderCommon{Backup: &BackupConfig{Enabled: true}}},
	}

	byType := providers.ByType()

	assert.Check(t, is.Len(byType, 5))
	assert.Check(t, is.DeepEqual(providers.S3.ProviderCommon, byType[S3Provider]))
	assert.Check(t, is.DeepEqual(providers.R2.ProviderCommon, byType[R2Provider]))
	assert.Check(t, is.DeepEqual(providers.GCS.ProviderCommon, byType[GCSProvider]))
	assert.Check(t, is.DeepEqual(providers.Disk.ProviderCommon, byType[DiskProvider]))
	assert.Check(t, is.DeepEqual(providers.Database.ProviderCommon, byType[DatabaseProvider]))
}

func TestBackupDestination(t *testing.T) {
	t.Run("disabled provider", func(t *testing.T) {
		cfg := ProviderCommon{Backup: &BackupConfig{Enabled: true, Provider: R2Provider}}
		_, ok := cfg.BackupDestination(S3Provider)
		assert.Check(t, !ok)
	})

	t.Run("no backup", func(t *testing.T) {
		cfg := ProviderCommon{Enabled: true}
		_, ok := cfg.BackupDestination(S3Provider)
		assert.Check(t, !ok)
	})

	t.Run("backup disabled", func(t *testing.T) {
		cfg := ProviderCommon{Enabled: true, Backup: &BackupConfig{Provider: R2Provider}}
		_, ok := cfg.BackupDestination(S3Provider)
		assert.Check(t, !ok)
	})

	t.Run("backup to source", func(t *testing.T) {
		cfg := ProviderCommon{Enabled: true, Backup: &BackupConfig{Enabled: true}}
		destination, ok := cfg.BackupDestination(S3Provider)
		assert.Check(t, ok)
		assert.Check(t, is.Equal(S3Provider, destination))
	})

	t.Run("backup to named provider", func(t *testing.T) {
		cfg := ProviderCommon{Enabled: true, Backup: &BackupConfig{Enabled: true, Provider: R2Provider}}
		destination, ok := cfg.BackupDestination(S3Provider)
		assert.Check(t, ok)
		assert.Check(t, is.Equal(R2Provider, destination))
	})
}

func TestAccessKeyCredentialsProviderCredentials(t *testing.T) {
	creds := AccessKeyCredentials{AccessKeyID: "key", SecretAccessKey: "secret"}

	assert.Check(t, is.DeepEqual(ProviderCredentials{AccessKeyID: "key", SecretAccessKey: "secret"}, creds.ProviderCredentials()))
}

func TestR2CredentialsProviderCredentials(t *testing.T) {
	creds := R2Credentials{
		AccessKeyID: "key", SecretAccessKey: "secret",
		AccountID: "account",
	}

	assert.Check(t, is.DeepEqual(ProviderCredentials{AccessKeyID: "key", SecretAccessKey: "secret", AccountID: "account"}, creds.ProviderCredentials()))
}
