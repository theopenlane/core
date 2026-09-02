package serveropts

import (
	"context"
	"crypto"
	"crypto/x509"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/theopenlane/core/v2/pkg/logx"
)

const (
	archivePrefix    = "openlane:jwks:retired:"
	archiveScanCount = 100

	// archiveRetentionFactor pads the retention window past the longest token lifetime so
	// clock skew and lagging replicas cannot retire a key that is still verifying tokens
	archiveRetentionFactor = 2
)

// keyArchive persists the public half of every signing key the server has seen so tokens
// stay verifiable after the private key behind them is replaced on disk
type keyArchive interface {
	// Record stores each public key, refreshing its retention window
	Record(ctx context.Context, keys map[string]crypto.PublicKey) error
	// Load returns every archived public key still inside its retention window
	Load(ctx context.Context) (map[string]crypto.PublicKey, error)
}

// noopKeyArchive is used when no shared store is configured, leaving verification to
// depend on whatever keys are currently on disk
type noopKeyArchive struct{}

// Record discards the supplied keys
func (noopKeyArchive) Record(_ context.Context, _ map[string]crypto.PublicKey) error {
	return nil
}

// Load returns no archived keys
func (noopKeyArchive) Load(_ context.Context) (map[string]crypto.PublicKey, error) {
	return nil, nil
}

// redisKeyArchive archives public keys in redis. Being shared across replicas is the point:
// a pod that starts after a rotation still serves the keys it never saw on disk
type redisKeyArchive struct {
	client    *redis.Client
	retention time.Duration
}

// newRedisKeyArchive returns an archive that retains public keys for the given duration
func newRedisKeyArchive(client *redis.Client, retention time.Duration) *redisKeyArchive {
	return &redisKeyArchive{client: client, retention: retention}
}

// Record stores each public key under its kid
func (a *redisKeyArchive) Record(ctx context.Context, keys map[string]crypto.PublicKey) error {
	pipe := a.client.Pipeline()

	for kid, publicKey := range keys {
		der, err := x509.MarshalPKIXPublicKey(publicKey)
		if err != nil {
			logx.FromContext(ctx).Error().Err(err).Msg("failed encoding public key for archive")

			return fmt.Errorf("failed encoding public key %q for archive: %w", kid, err)
		}

		// refreshing the TTL means a key expires only once it has been absent for the
		// full retention window, not once it stops signing
		pipe.Set(ctx, archivePrefix+kid, der, a.retention)
	}

	if _, err := pipe.Exec(ctx); err != nil {
		logx.FromContext(ctx).Error().Err(err).Msg("failed archiving public keys")

		return fmt.Errorf("failed archiving public keys: %w", err)
	}

	return nil
}

// Load returns every archived public key still inside its retention window
func (a *redisKeyArchive) Load(ctx context.Context) (map[string]crypto.PublicKey, error) {
	keys := make(map[string]crypto.PublicKey)

	iter := a.client.Scan(ctx, 0, archivePrefix+"*", archiveScanCount).Iterator()

	for iter.Next(ctx) {
		name := iter.Val()

		der, err := a.client.Get(ctx, name).Bytes()
		if err != nil {
			// expiry between the scan and the read is routine, not an error
			if errors.Is(err, redis.Nil) {
				continue
			}

			logx.FromContext(ctx).Error().Err(err).Msgf("failed reading archived public key %q", name)

			return nil, fmt.Errorf("failed reading archived public key %q: %w", name, err)
		}

		publicKey, err := x509.ParsePKIXPublicKey(der)
		if err != nil {
			logx.FromContext(ctx).Error().Err(err).Msgf("failed decoding archived public key %q", name)

			return nil, fmt.Errorf("failed decoding archived public key %q: %w", name, err)
		}

		keys[strings.TrimPrefix(name, archivePrefix)] = publicKey
	}

	if err := iter.Err(); err != nil {
		logx.FromContext(ctx).Error().Err(err).Msg("failed scanning archived public keys")

		return nil, fmt.Errorf("failed scanning archived public keys: %w", err)
	}

	return keys, nil
}
