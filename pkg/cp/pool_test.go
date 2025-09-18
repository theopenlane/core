package cp

import (
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// TestNewClientPool tests the NewClientPool function
func TestNewClientPool(t *testing.T) {
	ttl := 5 * time.Minute
	pool := NewClientPool[string](ttl)

	assert.NotNil(t, pool)
	assert.Equal(t, ttl, pool.ttl)
	assert.NotNil(t, pool.clients)
	assert.Empty(t, pool.clients)
}

// TestGetClientFromPool tests the GetClient method
func TestGetClientFromPool(t *testing.T) {
	type testClient struct {
		ID string
	}

	tests := []struct {
		name          string
		setupPool     func() *ClientPool[testClient]
		key           ClientCacheKey
		expectPresent bool
	}{
		{
			name: "get existing non-expired client",
			setupPool: func() *ClientPool[testClient] {
				pool := NewClientPool[testClient](5 * time.Minute)
				key := ClientCacheKey{
					TenantID:        "tenant1",
					IntegrationType: "storage",
					HushID:          "hush1",
				}
				pool.SetClient(key, testClient{ID: "client1"})
				return pool
			},
			key: ClientCacheKey{
				TenantID:        "tenant1",
				IntegrationType: "storage",
				HushID:          "hush1",
			},
			expectPresent: true,
		},
		{
			name: "get non-existent client",
			setupPool: func() *ClientPool[testClient] {
				return NewClientPool[testClient](5 * time.Minute)
			},
			key: ClientCacheKey{
				TenantID:        "tenant1",
				IntegrationType: "storage",
				HushID:          "hush1",
			},
			expectPresent: false,
		},
		{
			name: "get expired client",
			setupPool: func() *ClientPool[testClient] {
				pool := &ClientPool[testClient]{
					clients: make(map[ClientCacheKey]*ClientEntry[testClient]),
					ttl:     1 * time.Nanosecond,
				}
				key := ClientCacheKey{
					TenantID:        "tenant1",
					IntegrationType: "storage",
					HushID:          "hush1",
				}
				pool.clients[key] = &ClientEntry[testClient]{
					Client:     testClient{ID: "client1"},
					Expiration: time.Now().Add(-1 * time.Hour), // Already expired
				}
				return pool
			},
			key: ClientCacheKey{
				TenantID:        "tenant1",
				IntegrationType: "storage",
				HushID:          "hush1",
			},
			expectPresent: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pool := tt.setupPool()
			result := pool.GetClient(tt.key)

			assert.Equal(t, tt.expectPresent, result.IsPresent())
		})
	}
}

// TestSetClient tests the SetClient method
func TestSetClient(t *testing.T) {
	type testClient struct {
		ID string
	}

	pool := NewClientPool[testClient](5 * time.Minute)

	key := ClientCacheKey{
		TenantID:        "tenant1",
		IntegrationType: "storage",
		HushID:          "hush1",
	}

	client := testClient{ID: "client1"}

	// Set the client
	pool.SetClient(key, client)

	// Verify it was stored
	assert.Len(t, pool.clients, 1)
	entry, exists := pool.clients[key]
	assert.True(t, exists)
	assert.Equal(t, client, entry.Client)
	assert.True(t, entry.Expiration.After(time.Now()))

	// Set a different client with the same key (should overwrite)
	client2 := testClient{ID: "client2"}
	pool.SetClient(key, client2)

	// Verify it was overwritten
	assert.Len(t, pool.clients, 1)
	entry2, exists := pool.clients[key]
	assert.True(t, exists)
	assert.Equal(t, client2, entry2.Client)
}

// TestRemoveClient tests the RemoveClient method
func TestRemoveClient(t *testing.T) {
	type testClient struct {
		ID string
	}

	pool := NewClientPool[testClient](5 * time.Minute)

	key1 := ClientCacheKey{
		TenantID:        "tenant1",
		IntegrationType: "storage",
		HushID:          "hush1",
	}

	key2 := ClientCacheKey{
		TenantID:        "tenant2",
		IntegrationType: "storage",
		HushID:          "hush2",
	}

	// Add two clients
	pool.SetClient(key1, testClient{ID: "client1"})
	pool.SetClient(key2, testClient{ID: "client2"})

	assert.Len(t, pool.clients, 2)

	// Remove one client
	pool.RemoveClient(key1)

	// Verify only one remains
	assert.Len(t, pool.clients, 1)
	_, exists1 := pool.clients[key1]
	assert.False(t, exists1)
	_, exists2 := pool.clients[key2]
	assert.True(t, exists2)

	// Remove non-existent client (should not panic)
	pool.RemoveClient(ClientCacheKey{
		TenantID:        "tenant3",
		IntegrationType: "storage",
		HushID:          "hush3",
	})
	assert.Len(t, pool.clients, 1)
}

// TestCleanExpired tests the CleanExpired method
func TestCleanExpired(t *testing.T) {
	type testClient struct {
		ID string
	}

	pool := &ClientPool[testClient]{
		clients: make(map[ClientCacheKey]*ClientEntry[testClient]),
		ttl:     5 * time.Minute,
	}

	now := time.Now()

	// Add clients with different expiration times
	pool.clients[ClientCacheKey{TenantID: "1"}] = &ClientEntry[testClient]{
		Client:     testClient{ID: "1"},
		Expiration: now.Add(-1 * time.Hour), // Expired
	}

	pool.clients[ClientCacheKey{TenantID: "2"}] = &ClientEntry[testClient]{
		Client:     testClient{ID: "2"},
		Expiration: now.Add(1 * time.Hour), // Not expired
	}

	pool.clients[ClientCacheKey{TenantID: "3"}] = &ClientEntry[testClient]{
		Client:     testClient{ID: "3"},
		Expiration: now.Add(-1 * time.Minute), // Expired
	}

	pool.clients[ClientCacheKey{TenantID: "4"}] = &ClientEntry[testClient]{
		Client:     testClient{ID: "4"},
		Expiration: now.Add(1 * time.Minute), // Not expired
	}

	// Clean expired clients
	removed := pool.CleanExpired()

	// Verify correct number were removed
	assert.Equal(t, 2, removed)
	assert.Len(t, pool.clients, 2)

	// Verify the correct clients remain
	_, exists1 := pool.clients[ClientCacheKey{TenantID: "1"}]
	assert.False(t, exists1)
	_, exists2 := pool.clients[ClientCacheKey{TenantID: "2"}]
	assert.True(t, exists2)
	_, exists3 := pool.clients[ClientCacheKey{TenantID: "3"}]
	assert.False(t, exists3)
	_, exists4 := pool.clients[ClientCacheKey{TenantID: "4"}]
	assert.True(t, exists4)
}

// TestConcurrentAccess tests thread safety of the pool
func TestConcurrentAccess(t *testing.T) {
	type testClient struct {
		ID string
	}

	pool := NewClientPool[testClient](5 * time.Minute)

	var wg sync.WaitGroup
	operations := 100

	// Concurrent writes
	for i := 0; i < operations; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			key := ClientCacheKey{
				TenantID:        "tenant",
				IntegrationType: "storage",
				HushID:          string(rune(idx)),
			}
			pool.SetClient(key, testClient{ID: string(rune(idx))})
		}(i)
	}

	// Concurrent reads
	for i := 0; i < operations; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			key := ClientCacheKey{
				TenantID:        "tenant",
				IntegrationType: "storage",
				HushID:          string(rune(idx)),
			}
			pool.GetClient(key)
		}(i)
	}

	// Concurrent removes
	for i := 0; i < operations/2; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			key := ClientCacheKey{
				TenantID:        "tenant",
				IntegrationType: "storage",
				HushID:          string(rune(idx)),
			}
			pool.RemoveClient(key)
		}(i)
	}

	// Concurrent clean operations
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			pool.CleanExpired()
		}()
	}

	wg.Wait()

	// Just verify the pool is still in a valid state
	assert.NotNil(t, pool.clients)
}

// TestClientCacheKeyEquality tests that ClientCacheKey works correctly as a map key
func TestClientCacheKeyEquality(t *testing.T) {
	key1 := ClientCacheKey{
		TenantID:        "tenant1",
		IntegrationType: "storage",
		HushID:          "hush1",
	}

	key2 := ClientCacheKey{
		TenantID:        "tenant1",
		IntegrationType: "storage",
		HushID:          "hush1",
	}

	key3 := ClientCacheKey{
		TenantID:        "tenant2",
		IntegrationType: "storage",
		HushID:          "hush1",
	}

	// Test that identical keys are equal
	assert.Equal(t, key1, key2)

	// Test that different keys are not equal
	assert.NotEqual(t, key1, key3)

	// Test that they work correctly as map keys
	m := make(map[ClientCacheKey]string)
	m[key1] = "value1"

	assert.Equal(t, "value1", m[key2])
	assert.Empty(t, m[key3])
}

// CustomerS3Client represents an S3 client for a specific customer
type CustomerS3Client struct {
	CustomerID string
	Bucket     string
	Region     string
	AccessKey  string
}

// CustomerStripeClient represents a Stripe client for a specific customer
type CustomerStripeClient struct {
	CustomerID string
	APIKey     string
	AccountID  string
}

// TestMultiCustomerS3PoolIsolation tests that different customers get isolated S3 clients
func TestMultiCustomerS3PoolIsolation(t *testing.T) {
	pool := NewClientPool[CustomerS3Client](10 * time.Minute)

	customers := []struct {
		id     string
		bucket string
		region string
	}{
		{"customer-a", "customer-a-bucket", "us-east-1"},
		{"customer-b", "customer-b-bucket", "eu-west-1"},
		{"customer-c", "customer-c-bucket", "ap-southeast-1"},
	}

	// Store clients for each customer
	var clientResults []struct {
		key    ClientCacheKey
		client CustomerS3Client
	}

	for _, customer := range customers {
		key := ClientCacheKey{
			TenantID:        customer.id,
			IntegrationType: "s3",
			HushID:          "s3-creds-" + customer.id,
		}

		client := CustomerS3Client{
			CustomerID: customer.id,
			Bucket:     customer.bucket,
			Region:     customer.region,
			AccessKey:  "AKIA-" + customer.id,
		}

		pool.SetClient(key, client)
		clientResults = append(clientResults, struct {
			key    ClientCacheKey
			client CustomerS3Client
		}{key, client})
	}

	// Verify all clients are cached separately
	assert.Len(t, pool.clients, 3)

	// Verify each customer gets their own client
	for i := 0; i < len(clientResults); i++ {
		result := pool.GetClient(clientResults[i].key)
		assert.True(t, result.IsPresent())

		client, _ := result.Get()
		assert.Equal(t, clientResults[i].client.CustomerID, client.CustomerID)
		assert.Equal(t, clientResults[i].client.Bucket, client.Bucket)
		assert.Equal(t, clientResults[i].client.Region, client.Region)

		// Verify this client is different from all others
		for j := 0; j < len(clientResults); j++ {
			if i != j {
				otherResult := pool.GetClient(clientResults[j].key)
				assert.True(t, otherResult.IsPresent())
				otherClient, _ := otherResult.Get()
				assert.NotEqual(t, client.CustomerID, otherClient.CustomerID)
			}
		}
	}
}

// TestMultiCustomerStripePoolIsolation tests that different customers get isolated Stripe clients
func TestMultiCustomerStripePoolIsolation(t *testing.T) {
	pool := NewClientPool[CustomerStripeClient](10 * time.Minute)

	customers := []struct {
		id     string
		apiKey string
	}{
		{"customer-a", "definitelynotarealkey123456"},
		{"customer-b", "alsonotarealkey789012"},
		{"customer-c", "defdefdefnotarealkey345678"},
	}

	// Store clients for each customer
	var clientResults []struct {
		key    ClientCacheKey
		client CustomerStripeClient
	}

	for _, customer := range customers {
		key := ClientCacheKey{
			TenantID:        customer.id,
			IntegrationType: "stripe",
			HushID:          "stripe-creds-" + customer.id,
		}

		client := CustomerStripeClient{
			CustomerID: customer.id,
			APIKey:     customer.apiKey,
			AccountID:  "acct_" + customer.id,
		}

		pool.SetClient(key, client)
		clientResults = append(clientResults, struct {
			key    ClientCacheKey
			client CustomerStripeClient
		}{key, client})
	}

	// Verify all clients are cached separately
	assert.Len(t, pool.clients, 3)

	// Verify each customer gets their own client
	for i := 0; i < len(clientResults); i++ {
		result := pool.GetClient(clientResults[i].key)
		assert.True(t, result.IsPresent())

		client, _ := result.Get()
		assert.Equal(t, clientResults[i].client.CustomerID, client.CustomerID)
		assert.Equal(t, clientResults[i].client.APIKey, client.APIKey)
		assert.Equal(t, clientResults[i].client.AccountID, client.AccountID)

		// Verify this client is different from all others
		for j := 0; j < len(clientResults); j++ {
			if i != j {
				otherResult := pool.GetClient(clientResults[j].key)
				assert.True(t, otherResult.IsPresent())
				otherClient, _ := otherResult.Get()
				assert.NotEqual(t, client.CustomerID, otherClient.CustomerID)
				assert.NotEqual(t, client.APIKey, otherClient.APIKey)
			}
		}
	}
}

// TestConcurrentMultiCustomerPoolAccess tests concurrent access to customer-specific clients
func TestConcurrentMultiCustomerPoolAccess(t *testing.T) {
	s3Pool := NewClientPool[CustomerS3Client](10 * time.Minute)
	stripePool := NewClientPool[CustomerStripeClient](10 * time.Minute)

	customers := []struct {
		id     string
		bucket string
		apiKey string
	}{
		{"concurrent-customer-1", "bucket-1", "sk_test_1"},
		{"concurrent-customer-2", "bucket-2", "sk_test_2"},
		{"concurrent-customer-3", "bucket-3", "sk_test_3"},
		{"concurrent-customer-4", "bucket-4", "sk_test_4"},
		{"concurrent-customer-5", "bucket-5", "sk_test_5"},
	}

	var wg sync.WaitGroup
	results := make(map[string]map[string]interface{})
	var resultsMutex sync.Mutex

	// Concurrently store and retrieve clients for different customers
	for _, customer := range customers {
		wg.Add(2) // One for S3, one for Stripe

		// S3 client goroutine
		go func(custID, bucket string) {
			defer wg.Done()

			s3Key := ClientCacheKey{
				TenantID:        custID,
				IntegrationType: "s3",
				HushID:          "s3-" + custID,
			}

			s3Client := CustomerS3Client{
				CustomerID: custID,
				Bucket:     bucket,
				Region:     "us-west-2",
				AccessKey:  "AKIA-" + custID,
			}

			// Store and retrieve multiple times to test caching
			for i := 0; i < 10; i++ {
				s3Pool.SetClient(s3Key, s3Client)

				result := s3Pool.GetClient(s3Key)
				assert.True(t, result.IsPresent())

				retrievedClient, _ := result.Get()
				assert.Equal(t, custID, retrievedClient.CustomerID)
				assert.Equal(t, bucket, retrievedClient.Bucket)

				// Store first result for verification
				if i == 0 {
					resultsMutex.Lock()
					if results[custID] == nil {
						results[custID] = make(map[string]interface{})
					}
					results[custID]["s3_client"] = retrievedClient
					resultsMutex.Unlock()
				}

				// Simulate some work
				time.Sleep(1 * time.Millisecond)
			}
		}(customer.id, customer.bucket)

		// Stripe client goroutine
		go func(custID, apiKey string) {
			defer wg.Done()

			stripeKey := ClientCacheKey{
				TenantID:        custID,
				IntegrationType: "stripe",
				HushID:          "stripe-" + custID,
			}

			stripeClient := CustomerStripeClient{
				CustomerID: custID,
				APIKey:     apiKey,
				AccountID:  "acct_" + custID,
			}

			// Store and retrieve multiple times to test caching
			for i := 0; i < 10; i++ {
				stripePool.SetClient(stripeKey, stripeClient)

				result := stripePool.GetClient(stripeKey)
				assert.True(t, result.IsPresent())

				retrievedClient, _ := result.Get()
				assert.Equal(t, custID, retrievedClient.CustomerID)
				assert.Equal(t, apiKey, retrievedClient.APIKey)

				// Store first result for verification
				if i == 0 {
					resultsMutex.Lock()
					if results[custID] == nil {
						results[custID] = make(map[string]interface{})
					}
					results[custID]["stripe_client"] = retrievedClient
					resultsMutex.Unlock()
				}

				// Simulate some work
				time.Sleep(1 * time.Millisecond)
			}
		}(customer.id, customer.apiKey)
	}

	wg.Wait()

	// Verify all customers have both S3 and Stripe clients
	assert.Len(t, results, len(customers))
	for _, customer := range customers {
		custResults, exists := results[customer.id]
		assert.True(t, exists)
		assert.NotNil(t, custResults["s3_client"])
		assert.NotNil(t, custResults["stripe_client"])

		s3Client := custResults["s3_client"].(CustomerS3Client)
		assert.Equal(t, customer.id, s3Client.CustomerID)
		assert.Equal(t, customer.bucket, s3Client.Bucket)

		stripeClient := custResults["stripe_client"].(CustomerStripeClient)
		assert.Equal(t, customer.id, stripeClient.CustomerID)
		assert.Equal(t, customer.apiKey, stripeClient.APIKey)
	}

	// Verify proper cache usage - each customer should have exactly one S3 and one Stripe client cached
	assert.Len(t, s3Pool.clients, len(customers))
	assert.Len(t, stripePool.clients, len(customers))

	// Verify different customers have different client instances in cache
	customerIDs := make([]string, len(customers))
	for i, customer := range customers {
		customerIDs[i] = customer.id
	}

	for i := 0; i < len(customerIDs); i++ {
		for j := i + 1; j < len(customerIDs); j++ {
			custA, custB := customerIDs[i], customerIDs[j]

			// S3 clients should be different
			s3ClientA := results[custA]["s3_client"].(CustomerS3Client)
			s3ClientB := results[custB]["s3_client"].(CustomerS3Client)
			assert.NotEqual(t, s3ClientA.CustomerID, s3ClientB.CustomerID)
			assert.NotEqual(t, s3ClientA.Bucket, s3ClientB.Bucket)

			// Stripe clients should be different
			stripeClientA := results[custA]["stripe_client"].(CustomerStripeClient)
			stripeClientB := results[custB]["stripe_client"].(CustomerStripeClient)
			assert.NotEqual(t, stripeClientA.CustomerID, stripeClientB.CustomerID)
			assert.NotEqual(t, stripeClientA.APIKey, stripeClientB.APIKey)
		}
	}
}

// TestMultiCustomerClientExpiration tests that clients for different customers expire independently
func TestMultiCustomerClientExpiration(t *testing.T) {
	// Use very short TTL for testing
	pool := NewClientPool[CustomerS3Client](50 * time.Millisecond)

	customers := []struct {
		id     string
		bucket string
	}{
		{"expire-customer-a", "bucket-a"},
		{"expire-customer-b", "bucket-b"},
		{"expire-customer-c", "bucket-c"},
	}

	var keys []ClientCacheKey

	// Set clients for all customers
	for _, customer := range customers {
		key := ClientCacheKey{
			TenantID:        customer.id,
			IntegrationType: "s3",
			HushID:          "s3-" + customer.id,
		}

		client := CustomerS3Client{
			CustomerID: customer.id,
			Bucket:     customer.bucket,
			Region:     "us-east-1",
			AccessKey:  "AKIA-" + customer.id,
		}

		pool.SetClient(key, client)
		keys = append(keys, key)
	}

	// All should be present
	for i, key := range keys {
		result := pool.GetClient(key)
		assert.True(t, result.IsPresent(), "Customer %s client should be present", customers[i].id)
	}

	// Wait for expiration
	time.Sleep(100 * time.Millisecond)

	// All should be expired
	for i, key := range keys {
		result := pool.GetClient(key)
		assert.False(t, result.IsPresent(), "Customer %s client should be expired", customers[i].id)
	}
}
