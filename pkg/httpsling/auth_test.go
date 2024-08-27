package httpsling

import (
	"context"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBasicAuthApply(t *testing.T) {
	req, _ := http.NewRequestWithContext(context.Background(), http.MethodGet, "http://hotdogs.com", nil)

	auth := BasicAuth{
		Username: "user",
		Password: "pass",
	}
	auth.Apply(req)

	assert.Equal(t, "Basic dXNlcjpwYXNz", req.Header.Get("Authorization"))
}

func TestBasicAuthValid(t *testing.T) {
	auth := BasicAuth{
		Username: "user",
		Password: "pass",
	}

	assert.True(t, auth.Valid())
}

func TestBearerAuthApply(t *testing.T) {
	req, _ := http.NewRequestWithContext(context.Background(), http.MethodGet, "http://sarahisgreat.com", nil)

	auth := BearerAuth{
		Token: "token",
	}
	auth.Apply(req)

	assert.Equal(t, "Bearer token", req.Header.Get("Authorization"))
}

func TestBearerAuthValid(t *testing.T) {
	auth := BearerAuth{
		Token: "token",
	}

	assert.True(t, auth.Valid())
}

func TestCustomAuthApply(t *testing.T) {
	req, _ := http.NewRequestWithContext(context.Background(), http.MethodGet, "http://meow.com", nil)

	auth := CustomAuth{
		Header: "CustomValue",
	}
	auth.Apply(req)

	assert.Equal(t, "CustomValue", req.Header.Get("Authorization"))
}

func TestCustomAuthValid(t *testing.T) {
	auth := CustomAuth{
		Header: "CustomValue",
	}

	assert.True(t, auth.Valid())
}
