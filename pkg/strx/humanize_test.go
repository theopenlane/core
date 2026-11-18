package strx

import (
	"testing"

	"gotest.tools/v3/assert"
	is "gotest.tools/v3/assert/cmp"
)

func TestHumanizeSnake(t *testing.T) {
	testCases := []struct {
		input    string
		expected string
	}{
		{input: "", expected: ""},
		{input: "WEAK_SSL_POLICY", expected: "Weak SSL policy"},
		{input: "BINARY_AUTHORIZATION_DISABLED", expected: "Binary authorization disabled"},
		{input: "SERVICE_ACCOUNT_KEY_NOT_ROTATED", expected: "Service account key not rotated"},
		{input: "GKE_RUNTIME_OS_VULNERABILITY", expected: "GKE runtime OS vulnerability"},
		{input: "OS_LOGIN_DISABLED", expected: "OS login disabled"},
		{input: "PUBLIC_BUCKET__ACL", expected: "Public bucket ACL"},
		{input: "weak_ssl_policy", expected: "Weak SSL policy"},
		{input: "SOFTWARE", expected: "Software"},
		{input: "CVE_CVSS_SCORE_HIGH", expected: "CVE CVSS score high"},
		{input: "EC2_IMDSV2_NOT_REQUIRED", expected: "EC2 IMDSv2 not required"},
		{input: "OAUTH_CLIENT_ALLOWS_IPV6", expected: "OAuth client allows IPv6"},
		{input: "S3_BUCKET_PUBLIC_ACCESS", expected: "S3 bucket public access"},
		{input: "SSRF_IN_GITHUB_WEBHOOK", expected: "SSRF in GitHub webhook"},
		{input: "Persistence: New API Method", expected: "Persistence: New API Method"},
		{input: "camelCase", expected: "camelCase"},
		{input: "ÉCHEC_DU_SCAN", expected: "Échec du scan"},
	}

	for _, tc := range testCases {
		t.Run(tc.input, func(t *testing.T) {
			assert.Check(t, is.Equal(tc.expected, HumanizeSnake(tc.input)))
		})
	}
}
