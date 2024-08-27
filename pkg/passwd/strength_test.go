package passwd_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/theopenlane/core/pkg/passwd"
)

func TestStrength(t *testing.T) {
	testCases := []struct {
		password string
		expected passwd.PasswordStrength
	}{
		{"password", passwd.Weak},
		{"PASSWORD", passwd.Weak},
		{"  password ", passwd.Weak},
		{"foo", passwd.Weak},
		{"akexilaxzp", passwd.Poor},
		{"aklxiwoalsddaiwwa", passwd.Fair},
		{"alda13k932qda2", passwd.Moderate},
		{"alda#13k9-32qda2", passwd.Strong},
		{"a?Lda13k932Qd**A2", passwd.Excellent},
	}

	for i, tc := range testCases {
		require.Equal(t, tc.expected, passwd.Strength(tc.password), "test case %d did not match expectations", i)
	}
}
