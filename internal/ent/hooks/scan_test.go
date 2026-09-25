package hooks

import (
	"context"
	"testing"

	"github.com/theopenlane/iam/auth"
	"gotest.tools/v3/assert"

	"github.com/theopenlane/core/common/enums"
)

func TestScanOriginFromContext(t *testing.T) {
	cases := []struct {
		name   string
		caller *auth.Caller
		want   enums.ScanOrigin
	}{
		{name: "integration", caller: auth.NewIntegrationCaller("org"), want: enums.ScanOriginIntegration},
		{name: "internal", caller: &auth.Caller{OrganizationID: "org", Capabilities: auth.CapInternalOperation}, want: enums.ScanOriginSystem},
		{name: "session", caller: &auth.Caller{SubjectID: "user", OrganizationID: "org", AuthenticationType: auth.JWTAuthentication}, want: enums.ScanOriginUser},
		{name: "pat", caller: &auth.Caller{SubjectID: "user", OrganizationID: "org", AuthenticationType: auth.PATAuthentication}, want: enums.ScanOriginAPI},
		{name: "api token", caller: &auth.Caller{SubjectID: "svc", OrganizationID: "org", AuthenticationType: auth.APITokenAuthentication}, want: enums.ScanOriginAPI},
		{name: "internal user session", caller: &auth.Caller{SubjectID: "user", AuthenticationType: auth.JWTAuthentication, Capabilities: auth.CapInternalOperation}, want: enums.ScanOriginSystem},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			origin, err := scanOriginFromContext(auth.WithCaller(context.Background(), tc.caller))

			assert.NilError(t, err)
			assert.Check(t, origin == tc.want)
		})
	}
}

func TestScanOriginFromContextUnresolved(t *testing.T) {
	_, err := scanOriginFromContext(context.Background())

	assert.ErrorIs(t, err, ErrScanOriginUnresolved)

	_, err = scanOriginFromContext(auth.WithCaller(context.Background(), &auth.Caller{SubjectID: "anon"}))

	assert.ErrorIs(t, err, ErrScanOriginUnresolved)
}
