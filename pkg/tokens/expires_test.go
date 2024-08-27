package tokens_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/theopenlane/core/pkg/tokens"
)

const (
	accessToken  = "eyJhbGciOiJSUzI1NiIsImtpZCI6IjAxR1g2NDdTOFBDVkJDUEpIWEdKUjI2UE42IiwidHlwIjoiSldUIn0.eyJpc3MiOiJodHRwOi8vMTI3LjAuMC4xIiwiYXVkIjpbImh0dHA6Ly8xMjcuMC4wLjEiXSwiZXhwIjoxNjgwNjE1MzMwLCJuYmYiOjE2ODA2MTE3MzAsImlhdCI6MTY4MDYxMTczMCwianRpIjoiMDFneDY0N3M4cGN2YmNwamh4Z2pzcG04N3AiLCJuYW1lIjoiSm9obiBEb2UiLCJlbWFpbCI6Impkb2VAZXhhbXBsZS5jb20iLCJvcmciOiIxMjMiLCJwcm9qZWN0IjoiYWJjIiwicGVybWlzc2lvbnMiOlsicmVhZDpkYXRhIiwid3JpdGU6ZGF0YSJdfQ.LLb6c2RdACJmoT3IFgJEwfu2_YJMcKgM2bF3ISF41A37gKTOkBaOe-UuTmjgZ7WEcuQ-cVkht0KI_4zqYYctB_WB9481XoNwff5VgFf3xrPdOYxS00YXQnl09RRqt6Fmca8nvd4mXfdO7uvpyNVuCIqNxBPXdSnRhreSoFB1GtFm42sBPAD7vF-MQUmU0c4PTsbiCfhR1_buH0NYEE1QFp3vYcgoiXOJHh9VStmRscqvLB12AQrcs26G9opdTCCORmvR2W3JLJ_hliHyp-d9lhXmCDFyiGkDEhTAUglqwBjqz5SO1UfAThWJO18PvZl4QPhb724oNT82VPh0DMDwfw" // nolint: gosec
	refreshToken = "eyJhbGciOiJSUzI1NiIsImtpZCI6IjAxR1g2NDdTOFBDVkJDUEpIWEdKUjI2UE42IiwidHlwIjoiSldUIn0.eyJpc3MiOiJodHRwOi8vMTI3LjAuMC4xIiwiYXVkIjpbImh0dHA6Ly8xMjcuMC4wLjEiLCJodHRwOi8vMTI3LjAuMC4xL3YxL3JlZnJlc2giXSwiZXhwIjoxNjgwNjE4OTMwLCJuYmYiOjE2ODA2MTQ0MzAsImlhdCI6MTY4MDYxMTczMCwianRpIjoiMDFneDY0N3M4cGN2YmNwamh4Z2pzcG04N3AifQ.CLHmtZwSPFCPoMBX06D_C3h3WuEonUbvbfWLvtmrMmIwnTwQ4hxsaRJo_a4qI-emp1HNg-yu_7c3VNwjkti-d0c7CAGApTaf5eRdGJ5HGUkI8RDHbbMFaOK86nAFnzdPJ2JLmGtLzvpF9eFXFllDhRiAB-2t0uKcOdN7cFghdwyWXIVJIJNjngF_WUFklmLKnqORtj_tA6UJ6NJnZln34eMGftAHbuH8x-xUiRePHnro4ydS43CKNOgRP8biMHiRR2broBz0apIt30TeQShaBSbmGx__LYdm7RKPJNVHAn_3h_PwwKQG567-Aqabg6TSmpwhXCk_RfUyQVGv2b997w"                                                                                                                 // nolint: gosec
)

func TestParse(t *testing.T) {
	accessClaims, err := tokens.ParseUnverified(accessToken)
	require.NoError(t, err, "could not parse access token")

	refreshClaims, err := tokens.ParseUnverified(refreshToken)
	require.NoError(t, err, "could not parse refresh token")

	// We expect the claims and refresh tokens to have the same ID
	require.Equal(t, accessClaims.ID, refreshClaims.ID, "access and refresh token had different IDs or the parse was unsuccessful")

	// Check that an error is returned when parsing a bad token
	_, err = tokens.ParseUnverified("notarealtoken")
	require.Error(t, err, "should not be able to parse a bad token")
}

func TestExpiresAt(t *testing.T) {
	expiration, err := tokens.ExpiresAt(accessToken)
	require.NoError(t, err, "could not parse access token")

	// Expect the time to be fetched correctly from the token
	expected := time.Date(2023, 4, 4, 13, 35, 30, 0, time.UTC)
	require.True(t, expected.Equal(expiration))

	// Check that an error is returned when parsing a bad token
	_, err = tokens.ExpiresAt("notarealtoken")
	require.Error(t, err, "should not be able to parse a bad token")
}

func TestNotBefore(t *testing.T) {
	expiration, err := tokens.NotBefore(refreshToken)
	require.NoError(t, err, "could not parse access token")

	// Expect the time to be fetched correctly from the token
	expected := time.Date(2023, 4, 4, 13, 20, 30, 0, time.UTC)
	require.True(t, expected.Equal(expiration))

	// Check that an error is returned when parsing a bad token
	_, err = tokens.NotBefore("notarealtoken")
	require.Error(t, err, "should not be able to parse a bad token")
}
