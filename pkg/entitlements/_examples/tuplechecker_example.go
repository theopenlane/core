package entitlements

import (
	"context"
	"fmt"

	"github.com/redis/go-redis/v9"
)

// Example: Using TupleChecker in a service
func ExampleTupleChecker_usage() {
	ctx := context.Background()
	redisClient := redis.NewClient(&redis.Options{Addr: "localhost:6379", DB: 0})
	fgaClient := &MyFGAClient{} // Replace with your real FGA client

	tc := NewTupleChecker(
		WithRedisClient(redisClient),
		WithFGAClient(fgaClient),
	)

	tuple := FeatureTuple{
		UserID:  "user123",
		Feature: "featureX",
		Context: map[string]any{"org": "acme"},
	}

	// Check if user has feature
	hasFeature, err := tc.CheckFeatureTuple(ctx, tuple)
	if err != nil {
		fmt.Println("error:", err)
		return
	}
	if hasFeature {
		fmt.Println("User has feature!")
	} else {
		fmt.Println("User does not have feature.")
	}

	// Grant feature to user
	err = tc.CreateFeatureTuple(ctx, tuple)
	if err != nil {
		fmt.Println("error creating tuple:", err)
	}
}

// MyFGAClient is a stub for demonstration. Replace with your real implementation.
type MyFGAClient struct{}

func (m *MyFGAClient) CheckTuple(ctx context.Context, tuple FeatureTuple) (bool, error) {
	// Implement FGA check logic here
	return false, nil
}
func (m *MyFGAClient) CreateTuple(ctx context.Context, tuple FeatureTuple) error {
	// Implement FGA create logic here
	return nil
}
