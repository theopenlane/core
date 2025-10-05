package validator

import "testing"

func TestValidateIPAddress(t *testing.T) {
	validate := ValidateIPAddress()

	cases := []struct {
		name    string
		input   string
		wantErr bool
	}{
		{name: "empty allowed", input: "", wantErr: false},
		{name: "valid ipv4", input: "192.168.1.1", wantErr: false},
		{name: "valid ipv6", input: "2001:0db8:85a3:0000:0000:8a2e:0370:7334", wantErr: false},
		{name: "short ipv6", input: "2001:db8::1", wantErr: false},
		{name: "loopback v4", input: "127.0.0.1", wantErr: true},
		{name: "loopback v6", input: "::1", wantErr: true},
		{name: "unspecified v4", input: "0.0.0.0", wantErr: true},
		{name: "unspecified v6", input: "::", wantErr: true},
		{name: "invalid string", input: "not-an-ip", wantErr: true},
		{name: "ipv4 with port", input: "192.168.1.1:80", wantErr: true},
		{name: "ipv6 with brackets and port", input: "[2001:db8::1]:443", wantErr: true},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			err := validate(tc.input)
			if tc.wantErr && err == nil {
				t.Fatalf("expected error for input %q", tc.input)
			}
			if !tc.wantErr && err != nil {
				t.Fatalf("unexpected error for input %q: %v", tc.input, err)
			}
		})
	}
}
