package models_test

import (
	"testing"

	"github.com/theopenlane/core/pkg/models"
	"gotest.tools/v3/assert"
)

func TestValidateIP(t *testing.T) {
	tt := []struct {
		name     string
		ip       string
		hasError bool
	}{
		{
			name:     "empty string not allowed",
			ip:       "",
			hasError: true,
		},
		{
			name:     "127.0.0.1 not allowed",
			ip:       "127.0.0.1",
			hasError: true,
		},
		{
			name:     "0.0.0.0 not allowed",
			ip:       "0.0.0.0",
			hasError: true,
		},
		{
			name:     "valid ip",
			ip:       "192.168.0.1",
			hasError: false,
		},
	}

	for _, v := range tt {
		t.Run(v.name, func(t *testing.T) {
			err := models.ValidateIP(v.ip)
			if v.hasError {
				assert.Assert(t, err != nil)
				return
			}

			assert.NilError(t, err)
		})
	}
}
