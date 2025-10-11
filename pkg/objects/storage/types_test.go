package storage

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewProviderOptions(t *testing.T) {
	t.Run("no options", func(t *testing.T) {
		opts := NewProviderOptions()
		require.NotNil(t, opts)
		assert.Empty(t, opts.Bucket)
		assert.Empty(t, opts.Region)
		assert.Empty(t, opts.Endpoint)
	})

	t.Run("with multiple options", func(t *testing.T) {
		opts := NewProviderOptions(
			WithBucket("test-bucket"),
			WithRegion("us-west-2"),
			WithEndpoint("https://s3.example.com"),
		)
		require.NotNil(t, opts)
		assert.Equal(t, "test-bucket", opts.Bucket)
		assert.Equal(t, "us-west-2", opts.Region)
		assert.Equal(t, "https://s3.example.com", opts.Endpoint)
	})
}

func TestProviderOptions_Apply(t *testing.T) {
	t.Run("apply options", func(t *testing.T) {
		opts := &ProviderOptions{}
		opts.Apply(
			WithBucket("my-bucket"),
			WithRegion("eu-west-1"),
		)
		assert.Equal(t, "my-bucket", opts.Bucket)
		assert.Equal(t, "eu-west-1", opts.Region)
	})

	t.Run("apply nil option", func(t *testing.T) {
		opts := &ProviderOptions{}
		opts.Apply(nil)
		assert.Empty(t, opts.Bucket)
	})
}

func TestProviderOptions_Clone(t *testing.T) {
	t.Run("clone with values", func(t *testing.T) {
		original := &ProviderOptions{
			Bucket:   "test-bucket",
			Region:   "us-east-1",
			Endpoint: "https://endpoint.com",
			BasePath: "/uploads",
			LocalURL: "http://localhost:8080",
			Credentials: ProviderCredentials{
				AccessKeyID:     "key",
				SecretAccessKey: "secret",
			},
		}

		cloned := original.Clone()
		require.NotNil(t, cloned)
		assert.Equal(t, original.Bucket, cloned.Bucket)
		assert.Equal(t, original.Region, cloned.Region)
		assert.Equal(t, original.Endpoint, cloned.Endpoint)
		assert.Equal(t, original.BasePath, cloned.BasePath)
		assert.Equal(t, original.LocalURL, cloned.LocalURL)
		assert.Equal(t, original.Credentials, cloned.Credentials)

		cloned.Bucket = "modified"
		assert.NotEqual(t, original.Bucket, cloned.Bucket)
	})

	t.Run("clone with extras", func(t *testing.T) {
		original := &ProviderOptions{
			Bucket: "test-bucket",
		}
		original.Apply(WithExtra("key1", "value1"), WithExtra("key2", 42))

		cloned := original.Clone()
		require.NotNil(t, cloned)

		val, ok := cloned.Extra("key1")
		assert.True(t, ok)
		assert.Equal(t, "value1", val)

		val, ok = cloned.Extra("key2")
		assert.True(t, ok)
		assert.Equal(t, 42, val)
	})

	t.Run("clone nil", func(t *testing.T) {
		var opts *ProviderOptions
		cloned := opts.Clone()
		assert.Nil(t, cloned)
	})
}

func TestWithCredentials(t *testing.T) {
	creds := ProviderCredentials{
		AccessKeyID:     "access-key",
		SecretAccessKey: "secret-key",
		ProjectID:       "project-123",
		AccountID:       "account-456",
		APIToken:        "token-789",
	}

	opts := NewProviderOptions(WithCredentials(creds))
	assert.Equal(t, creds, opts.Credentials)
	assert.Equal(t, "access-key", opts.Credentials.AccessKeyID)
	assert.Equal(t, "secret-key", opts.Credentials.SecretAccessKey)
}

func TestWithBucket(t *testing.T) {
	opts := NewProviderOptions(WithBucket("my-bucket"))
	assert.Equal(t, "my-bucket", opts.Bucket)
}

func TestWithRegion(t *testing.T) {
	opts := NewProviderOptions(WithRegion("ap-south-1"))
	assert.Equal(t, "ap-south-1", opts.Region)
}

func TestWithEndpoint(t *testing.T) {
	opts := NewProviderOptions(WithEndpoint("https://custom.endpoint.com"))
	assert.Equal(t, "https://custom.endpoint.com", opts.Endpoint)
}

func TestWithBasePath(t *testing.T) {
	opts := NewProviderOptions(WithBasePath("/var/uploads"))
	assert.Equal(t, "/var/uploads", opts.BasePath)
}

func TestWithLocalURL(t *testing.T) {
	opts := NewProviderOptions(WithLocalURL("http://localhost:9000"))
	assert.Equal(t, "http://localhost:9000", opts.LocalURL)
}

func TestWithExtra(t *testing.T) {
	t.Run("add single extra", func(t *testing.T) {
		opts := NewProviderOptions(WithExtra("timeout", 30))
		val, ok := opts.Extra("timeout")
		assert.True(t, ok)
		assert.Equal(t, 30, val)
	})

	t.Run("add multiple extras", func(t *testing.T) {
		opts := NewProviderOptions(
			WithExtra("timeout", 30),
			WithExtra("retry", true),
			WithExtra("name", "test-provider"),
		)

		val, ok := opts.Extra("timeout")
		assert.True(t, ok)
		assert.Equal(t, 30, val)

		val, ok = opts.Extra("retry")
		assert.True(t, ok)
		assert.Equal(t, true, val)

		val, ok = opts.Extra("name")
		assert.True(t, ok)
		assert.Equal(t, "test-provider", val)
	})

	t.Run("overwrite extra", func(t *testing.T) {
		opts := NewProviderOptions(
			WithExtra("key", "old-value"),
			WithExtra("key", "new-value"),
		)

		val, ok := opts.Extra("key")
		assert.True(t, ok)
		assert.Equal(t, "new-value", val)
	})
}

func TestProviderOptions_Extra(t *testing.T) {
	t.Run("existing key", func(t *testing.T) {
		opts := NewProviderOptions(WithExtra("key", "value"))
		val, ok := opts.Extra("key")
		assert.True(t, ok)
		assert.Equal(t, "value", val)
	})

	t.Run("non-existent key", func(t *testing.T) {
		opts := NewProviderOptions()
		val, ok := opts.Extra("nonexistent")
		assert.False(t, ok)
		assert.Nil(t, val)
	})

	t.Run("nil options", func(t *testing.T) {
		var opts *ProviderOptions
		val, ok := opts.Extra("key")
		assert.False(t, ok)
		assert.Nil(t, val)
	})

	t.Run("no extras", func(t *testing.T) {
		opts := &ProviderOptions{Bucket: "test"}
		val, ok := opts.Extra("key")
		assert.False(t, ok)
		assert.Nil(t, val)
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
			WithBasePath("/path"),
			WithLocalURL("http://local"),
			WithExtra("custom", "value"),
		)

		assert.Equal(t, "key", opts.Credentials.AccessKeyID)
		assert.Equal(t, "bucket", opts.Bucket)
		assert.Equal(t, "region", opts.Region)
		assert.Equal(t, "endpoint", opts.Endpoint)
		assert.Equal(t, "/path", opts.BasePath)
		assert.Equal(t, "http://local", opts.LocalURL)

		val, ok := opts.Extra("custom")
		assert.True(t, ok)
		assert.Equal(t, "value", val)
	})
}
