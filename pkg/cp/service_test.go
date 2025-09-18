package cp

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// mockBuilder is a test implementation of ClientBuilder
type mockBuilder struct {
	credentials   map[string]string
	config        map[string]any
	buildFunc     func(ctx context.Context) (string, error)
	clientType    string
	returnError   bool
	returnedValue string
}

func (m *mockBuilder) WithCredentials(credentials map[string]string) ClientBuilder[string] {
	m.credentials = credentials
	return m
}

func (m *mockBuilder) WithConfig(config map[string]any) ClientBuilder[string] {
	m.config = config
	return m
}

func (m *mockBuilder) Build(ctx context.Context) (string, error) {
	if m.returnError {
		return "", errors.New("build failed")
	}
	if m.buildFunc != nil {
		return m.buildFunc(ctx)
	}
	return m.returnedValue, nil
}

func (m *mockBuilder) ClientType() ProviderType {
	return ProviderType(m.clientType)
}

// TestNewClientService tests the NewClientService function
func TestNewClientService(t *testing.T) {
	pool := NewClientPool[string](5 * time.Minute)
	service := NewClientService(pool)

	assert.NotNil(t, service)
	assert.Equal(t, pool, service.pool)
	assert.NotNil(t, service.builders)
	assert.Empty(t, service.builders)
}

// TestRegisterBuilder tests the RegisterBuilder method
func TestRegisterBuilder(t *testing.T) {
	pool := NewClientPool[string](5 * time.Minute)
	service := NewClientService(pool)

	builder1 := &mockBuilder{clientType: "storage"}
	builder2 := &mockBuilder{clientType: "slack"}

	// Register builders
	service.RegisterBuilder("storage", builder1)
	service.RegisterBuilder(ProviderType("slack"), builder2)

	// Verify they were registered
	assert.Len(t, service.builders, 2)
	assert.Equal(t, builder1, service.builders["storage"])
	assert.Equal(t, builder2, service.builders["slack"])

	// Overwrite a builder
	builder3 := &mockBuilder{clientType: "storage"}
	service.RegisterBuilder("storage", builder3)

	// Verify it was overwritten
	assert.Len(t, service.builders, 2)
	assert.Equal(t, builder3, service.builders["storage"])
}

// TestGetClientFromService tests the GetClient method
func TestGetClientFromService(t *testing.T) {
	tests := []struct {
		name          string
		setupService  func() *ClientService[string]
		key           ClientCacheKey
		clientType    string
		credentials   map[string]string
		config        map[string]any
		expectPresent bool
		expectedValue string
		expectCached  bool
	}{
		{
			name: "get from cache",
			setupService: func() *ClientService[string] {
				pool := NewClientPool[string](5 * time.Minute)
				service := NewClientService(pool)

				// Pre-populate cache
				key := ClientCacheKey{
					TenantID:        "tenant1",
					IntegrationType: "storage",
					HushID:          "hush1",
				}
				pool.SetClient(key, "cached-client")

				return service
			},
			key: ClientCacheKey{
				TenantID:        "tenant1",
				IntegrationType: "storage",
				HushID:          "hush1",
			},
			clientType:    "storage",
			expectPresent: true,
			expectedValue: "cached-client",
			expectCached:  true,
		},
		{
			name: "build new client",
			setupService: func() *ClientService[string] {
				pool := NewClientPool[string](5 * time.Minute)
				service := NewClientService(pool)

				builder := &mockBuilder{
					returnedValue: "new-client",
					clientType:    "storage",
				}
				service.RegisterBuilder(ProviderType("storage"), builder)

				return service
			},
			key: ClientCacheKey{
				TenantID:        "tenant1",
				IntegrationType: "storage",
				HushID:          "hush1",
			},
			clientType: "storage",
			credentials: map[string]string{
				"key": "value",
			},
			config: map[string]any{
				"option": "value",
			},
			expectPresent: true,
			expectedValue: "new-client",
			expectCached:  false,
		},
		{
			name: "builder not found",
			setupService: func() *ClientService[string] {
				pool := NewClientPool[string](5 * time.Minute)
				return NewClientService(pool)
			},
			key: ClientCacheKey{
				TenantID:        "tenant1",
				IntegrationType: "unknown",
				HushID:          "hush1",
			},
			clientType:    "unknown",
			expectPresent: false,
		},
		{
			name: "builder returns error",
			setupService: func() *ClientService[string] {
				pool := NewClientPool[string](5 * time.Minute)
				service := NewClientService(pool)

				builder := &mockBuilder{
					returnError: true,
					clientType:  "storage",
				}
				service.RegisterBuilder(ProviderType("storage"), builder)

				return service
			},
			key: ClientCacheKey{
				TenantID:        "tenant1",
				IntegrationType: "storage",
				HushID:          "hush1",
			},
			clientType:    "storage",
			expectPresent: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			service := tt.setupService()
			result := service.GetClient(context.Background(), tt.key, ProviderType(tt.clientType), tt.credentials, tt.config)

			assert.Equal(t, tt.expectPresent, result.IsPresent())

			if tt.expectPresent {
				value, _ := result.Get()
				assert.Equal(t, tt.expectedValue, value)

				// Verify caching behavior
				if !tt.expectCached {
					// Should now be in cache
					cached := service.pool.GetClient(tt.key)
					assert.True(t, cached.IsPresent())
					cachedValue, _ := cached.Get()
					assert.Equal(t, tt.expectedValue, cachedValue)
				}
			}
		})
	}
}

// TestGetClientWithCredentials tests that credentials are passed to builder
func TestGetClientWithCredentials(t *testing.T) {
	pool := NewClientPool[string](5 * time.Minute)
	service := NewClientService(pool)

	expectedCredentials := map[string]string{
		"access_key": "key123",
		"secret_key": "secret456",
	}

	expectedConfig := map[string]any{
		"region":  "us-west-2",
		"timeout": 30,
	}

	var capturedBuilder *mockBuilder
	capturedBuilder = &mockBuilder{
		clientType: "storage",
		buildFunc: func(ctx context.Context) (string, error) {
			// Verify credentials and config were set
			assert.Equal(t, expectedCredentials, capturedBuilder.credentials)
			assert.Equal(t, expectedConfig, capturedBuilder.config)
			return "configured-client", nil
		},
	}
	builder := capturedBuilder

	service.RegisterBuilder(ProviderType("storage"), builder)

	key := ClientCacheKey{
		TenantID:        "tenant1",
		IntegrationType: "storage",
		HushID:          "hush1",
	}

	result := service.GetClient(context.Background(), key, ProviderType("storage"), expectedCredentials, expectedConfig)
	assert.True(t, result.IsPresent())
}

// TestPool tests the Pool method
func TestPool(t *testing.T) {
	pool := NewClientPool[string](5 * time.Minute)
	service := NewClientService(pool)

	assert.Same(t, pool, service.Pool())
}

// TestConcurrentServiceAccess tests thread safety of the service
func TestConcurrentServiceAccess(t *testing.T) {
	pool := NewClientPool[string](5 * time.Minute)
	service := NewClientService(pool)

	// Register multiple builders
	for i := 0; i < 10; i++ {
		clientType := string(rune('a' + i))
		builder := &mockBuilder{
			clientType:    clientType,
			returnedValue: "client-" + clientType,
		}
		service.RegisterBuilder(ProviderType(clientType), builder)
	}

	done := make(chan bool)

	// Concurrent builder registration
	go func() {
		for i := 0; i < 100; i++ {
			clientType := string(rune((i % 10) + 'a'))
			builder := &mockBuilder{
				clientType:    clientType,
				returnedValue: "updated-" + clientType,
			}
			service.RegisterBuilder(ProviderType(clientType), builder)
		}
		done <- true
	}()

	// Concurrent client retrieval
	go func() {
		for i := 0; i < 100; i++ {
			clientType := string(rune((i % 10) + 'a'))
			key := ClientCacheKey{
				TenantID:        "tenant",
				IntegrationType: clientType,
				HushID:          string(rune(i)),
			}
			service.GetClient(context.Background(), key, ProviderType(clientType), nil, nil)
		}
		done <- true
	}()

	// Wait for all goroutines
	for range 4 {
		<-done
	}
}

// TestGetClientContextPropagation tests that context is properly propagated to builders
func TestGetClientContextPropagation(t *testing.T) {
	pool := NewClientPool[string](5 * time.Minute)
	service := NewClientService(pool)

	contextKey := struct{}{}
	expectedValue := "context-value"

	builder := &mockBuilder{
		clientType: "storage",
		buildFunc: func(ctx context.Context) (string, error) {
			// Verify context value is present
			value := ctx.Value(contextKey)
			assert.Equal(t, expectedValue, value)
			return "client-with-context", nil
		},
	}

	service.RegisterBuilder(ProviderType("storage"), builder)

	ctx := context.WithValue(context.Background(), contextKey, expectedValue)
	key := ClientCacheKey{
		TenantID:        "tenant1",
		IntegrationType: "storage",
		HushID:          "hush1",
	}

	result := service.GetClient(ctx, key, ProviderType("storage"), nil, nil)
	assert.True(t, result.IsPresent())
}

// callTrackingBuilder tracks the order of method calls
type callTrackingBuilder struct {
	callOrder   *[]string
	credentials map[string]string
	config      map[string]any
}

func (b *callTrackingBuilder) WithCredentials(credentials map[string]string) ClientBuilder[string] {
	*b.callOrder = append(*b.callOrder, "WithCredentials")
	b.credentials = credentials
	return b
}

func (b *callTrackingBuilder) WithConfig(config map[string]any) ClientBuilder[string] {
	*b.callOrder = append(*b.callOrder, "WithConfig")
	b.config = config
	return b
}

func (b *callTrackingBuilder) Build(ctx context.Context) (string, error) {
	*b.callOrder = append(*b.callOrder, "Build")
	return "test-client", nil
}

func (b *callTrackingBuilder) ClientType() ProviderType {
	return ProviderType("storage")
}

// TestBuilderInterface tests that the builder interface is used correctly
func TestBuilderInterface(t *testing.T) {
	// This test ensures the builder interface methods are called in the correct order
	pool := NewClientPool[string](5 * time.Minute)
	service := NewClientService(pool)

	callOrder := []string{}
	builder := &callTrackingBuilder{
		callOrder: &callOrder,
	}

	service.RegisterBuilder(ProviderType("storage"), builder)

	key := ClientCacheKey{
		TenantID:        "tenant1",
		IntegrationType: "storage",
		HushID:          "hush1",
	}

	credentials := map[string]string{"key": "value"}
	config := map[string]any{"option": "value"}

	result := service.GetClient(context.Background(), key, ProviderType("storage"), credentials, config)
	assert.True(t, result.IsPresent())

	// Verify methods were called in correct order
	assert.Equal(t, []string{"WithCredentials", "WithConfig", "Build"}, callOrder)
}

// TestServiceWithComplexClientType tests the service with a more complex client type
func TestServiceWithComplexClientType(t *testing.T) {
	pool := NewClientPool[complexClient](5 * time.Minute)
	service := NewClientService(pool)

	builder := &mockcomplexBuilder{}
	service.RegisterBuilder(ProviderType("complex"), builder)

	key := ClientCacheKey{
		TenantID:        "tenant1",
		IntegrationType: "complex",
		HushID:          "hush1",
	}

	result := service.GetClient(context.Background(), key, ProviderType("complex"), map[string]string{"key": "value"}, map[string]any{"option": "value"})
	assert.True(t, result.IsPresent())

	client := result.MustGet()
	assert.Equal(t, "complex-123", client.ID)
	assert.Equal(t, "advanced", client.Provider)
}

// mockcomplexBuilder properly implements ClientBuilder[complexClient]
type mockcomplexBuilder struct {
	credentials map[string]string
	config      map[string]any
}

type complexClient struct {
	ID       string
	Provider string
	Config   map[string]any
}

func (m *mockcomplexBuilder) WithCredentials(credentials map[string]string) ClientBuilder[complexClient] {
	m.credentials = credentials
	return m
}

func (m *mockcomplexBuilder) WithConfig(config map[string]any) ClientBuilder[complexClient] {
	m.config = config
	return m
}

func (m *mockcomplexBuilder) Build(ctx context.Context) (complexClient, error) {
	return complexClient{
		ID:       "complex-123",
		Provider: "advanced",
		Config:   m.config,
	}, nil
}

func (m *mockcomplexBuilder) ClientType() ProviderType {
	return ProviderType("complex")
}

// TestEmptyCredentialsAndConfig tests that nil credentials and config are handled
func TestEmptyCredentialsAndConfig(t *testing.T) {
	pool := NewClientPool[string](5 * time.Minute)
	service := NewClientService(pool)

	var capturedBuilder *mockBuilder
	capturedBuilder = &mockBuilder{
		clientType: "storage",
		buildFunc: func(ctx context.Context) (string, error) {
			// Verify nil was passed through correctly
			assert.Nil(t, capturedBuilder.credentials)
			assert.Nil(t, capturedBuilder.config)
			return "client-no-config", nil
		},
	}
	builder := capturedBuilder

	service.RegisterBuilder(ProviderType("storage"), builder)

	key := ClientCacheKey{
		TenantID:        "tenant1",
		IntegrationType: "storage",
		HushID:          "hush1",
	}

	result := service.GetClient(context.Background(), key, ProviderType("storage"), nil, nil)
	assert.True(t, result.IsPresent())
}
