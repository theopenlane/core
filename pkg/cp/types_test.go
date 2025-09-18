package cp

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// TestClientCacheKey tests the ClientCacheKey struct
func TestClientCacheKey(t *testing.T) {
	// Test that ClientCacheKey can be used as a map key
	cache := make(map[ClientCacheKey]string)

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

	// Test setting and getting with identical keys
	cache[key1] = "value1"
	assert.Equal(t, "value1", cache[key2])

	// Test that different keys don't conflict
	cache[key3] = "value3"
	assert.Equal(t, "value1", cache[key1])
	assert.Equal(t, "value3", cache[key3])

	// Test all field combinations for uniqueness
	testKeys := []ClientCacheKey{
		{TenantID: "a", IntegrationType: "b", HushID: "c"},
		{TenantID: "a", IntegrationType: "b", HushID: "d"},
		{TenantID: "a", IntegrationType: "e", HushID: "c"},
		{TenantID: "f", IntegrationType: "b", HushID: "c"},
	}

	for i, k := range testKeys {
		cache[k] = string(rune('A' + i))
	}

	for i, k := range testKeys {
		assert.Equal(t, string(rune('A'+i)), cache[k])
	}
}

// TestClientEntry tests the ClientEntry struct
func TestClientEntry(t *testing.T) {
	type testClient struct {
		ID   string
		Name string
	}

	client := testClient{
		ID:   "123",
		Name: "test-client",
	}

	expiration := time.Now().Add(5 * time.Minute)

	entry := ClientEntry[testClient]{
		Client:     client,
		Expiration: expiration,
	}

	assert.Equal(t, client, entry.Client)
	assert.Equal(t, expiration, entry.Expiration)

	// Test that entry can hold different client types
	stringEntry := ClientEntry[string]{
		Client:     "string-client",
		Expiration: expiration,
	}

	assert.Equal(t, "string-client", stringEntry.Client)

	// Test zero value
	var zeroEntry ClientEntry[testClient]
	assert.Zero(t, zeroEntry.Client)
	assert.Zero(t, zeroEntry.Expiration)
}

// TestClientPool tests the ClientPool struct initialization
func TestClientPoolStruct(t *testing.T) {
	pool := &ClientPool[string]{
		clients: make(map[ClientCacheKey]*ClientEntry[string]),
		ttl:     5 * time.Minute,
	}

	assert.NotNil(t, pool.clients)
	assert.Equal(t, 5*time.Minute, pool.ttl)
	assert.Empty(t, pool.clients)
}

// TestClientService tests the ClientService struct initialization
func TestClientServiceStruct(t *testing.T) {
	pool := &ClientPool[string]{
		clients: make(map[ClientCacheKey]*ClientEntry[string]),
		ttl:     5 * time.Minute,
	}

	service := &ClientService[string]{
		pool:     pool,
		builders: make(map[ProviderType]ClientBuilder[string]),
	}

	assert.Equal(t, pool, service.pool)
	assert.NotNil(t, service.builders)
	assert.Empty(t, service.builders)
}

// TestClientBuilderInterface tests that the ClientBuilder interface is properly defined
func TestClientBuilderInterface(t *testing.T) {
	// Test with a simple string client
	var builder ClientBuilder[string]

	// Create a test implementation
	testBuilder := &testClientBuilder{
		clientType: "test",
	}

	// Ensure it implements the interface
	builder = testBuilder

	// Test the interface methods
	assert.Equal(t, ProviderType("test"), builder.ClientType())

	// Test method chaining
	builder = builder.WithCredentials(map[string]string{"key": "value"})
	assert.NotNil(t, builder)

	builder = builder.WithConfig(map[string]any{"option": "value"})
	assert.NotNil(t, builder)

	// Test Build
	client, err := builder.Build(context.Background())
	assert.NoError(t, err)
	assert.Equal(t, "test-client", client)
}

// testClientBuilder is a test implementation of ClientBuilder
type testClientBuilder struct {
	clientType  string
	credentials map[string]string
	config      map[string]any
}

func (t *testClientBuilder) WithCredentials(credentials map[string]string) ClientBuilder[string] {
	t.credentials = credentials
	return t
}

func (t *testClientBuilder) WithConfig(config map[string]any) ClientBuilder[string] {
	t.config = config
	return t
}

func (t *testClientBuilder) Build(ctx context.Context) (string, error) {
	return "test-client", nil
}

func (t *testClientBuilder) ClientType() ProviderType {
	return ProviderType(t.clientType)
}

// TestTypesWithDifferentClientTypes verifies the generic system works with various types
func TestTypesWithDifferentClientTypes(t *testing.T) {
	// Test with struct type
	type structClient struct {
		Field string
	}

	structPool := &ClientPool[structClient]{
		clients: make(map[ClientCacheKey]*ClientEntry[structClient]),
		ttl:     5 * time.Minute,
	}
	assert.NotNil(t, structPool)

	// Test with interface type
	type interfaceClient interface {
		Method() string
	}

	interfacePool := &ClientPool[interfaceClient]{
		clients: make(map[ClientCacheKey]*ClientEntry[interfaceClient]),
		ttl:     5 * time.Minute,
	}
	assert.NotNil(t, interfacePool)

	// Test with pointer type
	pointerPool := &ClientPool[*structClient]{
		clients: make(map[ClientCacheKey]*ClientEntry[*structClient]),
		ttl:     5 * time.Minute,
	}
	assert.NotNil(t, pointerPool)

	// Test with slice type
	slicePool := &ClientPool[[]string]{
		clients: make(map[ClientCacheKey]*ClientEntry[[]string]),
		ttl:     5 * time.Minute,
	}
	assert.NotNil(t, slicePool)

	// Test with map type
	mapPool := &ClientPool[map[string]any]{
		clients: make(map[ClientCacheKey]*ClientEntry[map[string]any]),
		ttl:     5 * time.Minute,
	}
	assert.NotNil(t, mapPool)
}


// TestTypeConstraints verifies that the generic constraints work as expected
func TestTypeConstraints(t *testing.T) {
	// These should compile and work with any type due to [T any] constraint

	// Primitive types
	_ = NewClientPool[int](time.Minute)
	_ = NewClientPool[string](time.Minute)
	_ = NewClientPool[bool](time.Minute)
	_ = NewClientPool[float64](time.Minute)

	// Complex types
	_ = NewClientPool[struct{ Field string }](time.Minute)
	_ = NewClientPool[interface{ Method() }](time.Minute)
	_ = NewClientPool[chan int](time.Minute)
	_ = NewClientPool[func() error](time.Minute)

	// Verify that nil can be stored for pointer types
	type testStruct struct {
		Value string
	}

	pool := NewClientPool[*testStruct](time.Minute)
	key := ClientCacheKey{TenantID: "1"}

	var nilClient *testStruct
	pool.SetClient(key, nilClient)

	result := pool.GetClient(key)
	assert.True(t, result.IsPresent())
	value, _ := result.Get()
	assert.Nil(t, value)
}
