package entitlements

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

// TupleCheckerOption allows functional options for TupleChecker.
type TupleCheckerOption func(*TupleChecker)

// TupleChecker provides methods to create/check tuples with Redis cache fallback to FGA.
type TupleChecker struct {
	redisClient *redis.Client
	fgaChecker  FGAClient
	cacheTTL    time.Duration
}

// FGAClient is an interface for FGA tuple checking/creation/deletion
type FGAClient interface {
	CheckTuple(ctx context.Context, tuple FeatureTuple) (bool, error)
	CreateTuple(ctx context.Context, tuple FeatureTuple) error
	DeleteTuple(ctx context.Context, tuple FeatureTuple) error
}

// FeatureTuple represents a generic tuple for feature checks
type FeatureTuple struct {
	// UserID is the identifier for the user or entity being checked
	UserID string
	// Feature is the name of the feature being checked - can be any feature
	Feature string
	// Context is a map of additional context for the tuple, can be used for feature flags or other metadata
	Context map[string]any
}

// WithCacheTTL sets the cache TTL for TupleChecker
func WithCacheTTL(ttl time.Duration) TupleCheckerOption {
	return func(tc *TupleChecker) {
		tc.cacheTTL = ttl
	}
}

// WithRedisClient sets a custom Redis client
func WithRedisClient(client *redis.Client) TupleCheckerOption {
	return func(tc *TupleChecker) {
		tc.redisClient = client
	}
}

// WithFGAClient sets a custom FGA client
func WithFGAClient(client FGAClient) TupleCheckerOption {
	return func(tc *TupleChecker) {
		tc.fgaChecker = client
	}
}

// NewTupleChecker creates a new TupleChecker with options
func NewTupleChecker(opts ...TupleCheckerOption) *TupleChecker {
	tc := &TupleChecker{
		cacheTTL: 5 * time.Minute, // nolint:mnd
	}
	for _, opt := range opts {
		opt(tc)
	}

	return tc
}

// CheckFeatureTuple checks if a tuple exists, using Redis cache first, then FGA
func (tc *TupleChecker) CheckFeatureTuple(ctx context.Context, tuple FeatureTuple) (bool, error) {
	if tc.redisClient == nil || tc.fgaChecker == nil {
		return false, fmt.Errorf("%w", ErrTupleCheckerNotConfigured)
	}

	key := tc.cacheKey(tuple)

	val, err := tc.redisClient.Get(ctx, key).Result()
	if err == nil {
		return val == "1", nil
	}

	if !errors.Is(err, redis.Nil) {
		return false, err
	}
	// Not in cache, check FGA
	ok, err := tc.fgaChecker.CheckTuple(ctx, tuple)
	if err != nil {
		return false, err
	}
	// Cache result
	cacheVal := "0"
	if ok {
		cacheVal = "1"
	}

	// this suppresses the error if the set fails, as we don't want to fail the check if caching fails, because FGA is the source of truth
	_ = tc.redisClient.Set(ctx, key, cacheVal, tc.cacheTTL).Err()

	return ok, nil
}

// CreateFeatureTuple creates a tuple in FGA and updates the cache
func (tc *TupleChecker) CreateFeatureTuple(ctx context.Context, tuple FeatureTuple) error {
	if tc.redisClient == nil || tc.fgaChecker == nil {
		return ErrTupleCheckerNotConfigured
	}

	if err := tc.fgaChecker.CreateTuple(ctx, tuple); err != nil {
		return err
	}

	key := tc.cacheKey(tuple)

	return tc.redisClient.Set(ctx, key, "1", tc.cacheTTL).Err()
}

// DeleteFeatureTuple deletes a tuple in FGA and removes it from the cache
func (tc *TupleChecker) DeleteFeatureTuple(ctx context.Context, tuple FeatureTuple) error {
	if tc.redisClient == nil || tc.fgaChecker == nil {
		return ErrTupleCheckerNotConfigured
	}

	if err := tc.fgaChecker.DeleteTuple(ctx, tuple); err != nil {
		return err
	}

	key := tc.cacheKey(tuple)
	if err := tc.redisClient.Del(ctx, key).Err(); err != nil {
		return err
	}

	return nil
}

func (tc *TupleChecker) cacheKey(tuple FeatureTuple) string {
	ctxBytes, _ := json.Marshal(tuple.Context)

	return fmt.Sprintf("feature:%s:user:%s:ctx:%s", tuple.Feature, tuple.UserID, string(ctxBytes))
}
