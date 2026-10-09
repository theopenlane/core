package validator_test

import (
	"testing"

	"gotest.tools/v3/assert"

	"github.com/theopenlane/core/v2/internal/ent/validator"
)

func TestValidatePublicURL(t *testing.T) {
	t.Parallel()

	tests := []struct {
		url     string
		wantErr bool
	}{
		{url: "https://example.com/hook"},
		{url: "http://metadata.google.internal/computeMetadata/v1/", wantErr: true},
		{url: "http://169.254.169.254/latest/meta-data/", wantErr: true},
		{url: "http://127.0.0.1:8080/", wantErr: true},
		{url: "http://10.0.0.5/", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.url, func(t *testing.T) {
			t.Parallel()

			err := validator.ValidatePublicURL()(tt.url)
			if !tt.wantErr {
				assert.NilError(t, err)
				return
			}

			assert.ErrorIs(t, err, validator.ErrURLNotPublic)
		})
	}
}

func TestValidateOutboundHeaders(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		headers map[string]string
		wantErr bool
	}{
		{name: "no headers"},
		{name: "custom header", headers: map[string]string{"X-Custom": "value"}},
		{name: "gcp metadata flavor", headers: map[string]string{"Metadata-Flavor": "Google"}, wantErr: true},
		{name: "aws metadata token with spacing", headers: map[string]string{" X-AWS-EC2-Metadata-Token ": "token"}, wantErr: true},
		{name: "aws metadata token ttl", headers: map[string]string{"X-aws-ec2-metadata-token-ttl-seconds": "21600"}, wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			err := validator.ValidateOutboundHeaders()(tt.headers)
			if !tt.wantErr {
				assert.NilError(t, err)
				return
			}

			assert.ErrorIs(t, err, validator.ErrHeaderNotAllowed)
		})
	}
}
