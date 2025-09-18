package storage_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/theopenlane/core/pkg/objects/storage"
)

func TestNewObjectService(t *testing.T) {
	service := storage.NewObjectService()

	assert.NotNil(t, service)
	assert.Equal(t, int64(storage.DefaultMaxFileSize), service.MaxSize())
	assert.Equal(t, int64(storage.DefaultMaxMemory), service.MaxMemory())
	assert.Equal(t, []string{storage.DefaultUploadFileKey}, service.Keys())
	assert.False(t, service.IgnoreNonExistentKeys())
	assert.NotNil(t, service.Skipper())
	assert.NotNil(t, service.ErrorResponseHandler())
}

func TestObjectServiceAccessors(t *testing.T) {
	service := storage.NewObjectService()

	// Test all accessor methods
	assert.Equal(t, int64(storage.DefaultMaxFileSize), service.MaxSize())
	assert.Equal(t, int64(storage.DefaultMaxMemory), service.MaxMemory())
	assert.Equal(t, []string{storage.DefaultUploadFileKey}, service.Keys())
	assert.False(t, service.IgnoreNonExistentKeys())
	assert.NotNil(t, service.Skipper())
	assert.NotNil(t, service.ErrorResponseHandler())

	// Test skipper function
	skipper := service.Skipper()
	assert.False(t, skipper(nil)) // DefaultSkipper always returns false

	// Test error response handler
	handler := service.ErrorResponseHandler()
	assert.NotNil(t, handler)
}
