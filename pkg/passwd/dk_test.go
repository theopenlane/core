package passwd_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/theopenlane/core/pkg/passwd"
)

func TestDerivedKey(t *testing.T) {
	testCases := []struct {
		name           string
		passwordCreate string
		passwordVerify string
		verified       bool
	}{
		{
			"happy path, matching",
			"supersafesa$#%asaf!",
			"supersafesa$#%asaf!",
			true,
		},
		{
			"not matching",
			"supersafesa$#%asaf!",
			"notthesamething",
			false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Create a derived key from a password
			password, err := passwd.CreateDerivedKey(tc.passwordCreate)
			require.NoError(t, err)

			// verify key
			verified, err := passwd.VerifyDerivedKey(password, tc.passwordVerify)
			require.NoError(t, err)
			require.Equal(t, tc.verified, verified)
		})
	}
}

func TestDerivedKeyErrors(t *testing.T) {
	testCases := []struct {
		name          string
		dk            string
		password      string
		expectedError string
	}{
		{
			"cannot verify empty derived key or password",
			"",
			"foo",
			"cannot verify empty derived key or password",
		},
		{
			"cannot verify empty derived key or password, take 2",
			"foo",
			"",
			"cannot verify empty derived key or password",
		},
		{
			"cannot parse encoded derived key, does not match regular expression",
			"notarealkey",
			"supersecretpassword",
			"cannot parse encoded derived key, does not match regular expression",
		},
		{
			"could not parse version",
			"$argon2id$v=13212$m=65536,t=1,p=2$FrAEw4rWRDpyIZXR/QSzpg==$chQikgApfQfSaPZ7idk6caqBk79xRalpPUs4Ro/hywM=",
			"supersecretpassword",
			"could not parse version",
		},
		{
			"could not parse time",
			"$argon2id$v=19$m=65536,t=999999999999999999,p=2$FrAEw4rWRDpyIZXR/QSzpg==$chQikgApfQfSaPZ7idk6caqBk79xRalpPUs4Ro/hywM=",
			"supersecretpassword",
			"could not parse time",
		},
		{
			"could not parse memory",
			"$argon2id$v=19$m=999999999999999999,t=1,p=2$FrAEw4rWRDpyIZXR/QSzpg==$chQikgApfQfSaPZ7idk6caqBk79xRalpPUs4Ro/hywM=",
			"supersecretpassword",
			"could not parse memory",
		},
		{
			"could not parse threads",
			"$argon2id$v=19$m=65536,t=1,p=999999999999999999$FrAEw4rWRDpyIZXR/QSzpg==$chQikgApfQfSaPZ7idk6caqBk79xRalpPUs4Ro/hywM=",
			"supersecretpassword",
			"could not parse threads",
		},
		{
			"could not parse salt",
			"$argon2id$v=19$m=65536,t=1,p=2$==FrAEw4rWRDpyIZXR/QSzpg==$chQikgApfQfSaPZ7idk6caqBk79xRalpPUs4Ro/hywM=",
			"supersecretpassword",
			"could not parse salt",
		},
		{
			"could not parse dk",
			"$argon2id$v=19$m=65536,t=1,p=2$FrAEw4rWRDpyIZXR/QSzpg==$==chQikgApfQfSaPZ7idk6caqBk79xRalpPUs4Ro/hywM=",
			"supersecretpassword",
			"could not parse dk",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := passwd.VerifyDerivedKey(tc.dk, tc.password)
			require.ErrorContains(t, err, tc.expectedError)
		})
	}
}
