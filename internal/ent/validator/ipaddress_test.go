package validator_test

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/theopenlane/core/internal/ent/validator"
)

func TestValidateIPAddress(t *testing.T) {
	validIPAddresses := []string{
		// Valid IPv4 addresses
		"192.168.1.1",
		"10.0.0.1",
		"172.16.0.1",
		"8.8.8.8",
		"1.1.1.1",
		"255.255.255.255",
		"192.0.2.1",
		// Valid IPv6 addresses
		"2001:0db8:85a3:0000:0000:8a2e:0370:7334",
		"2001:db8:85a3::8a2e:370:7334",
		"2001:db8::1",
		"fe80::1",
		"2001:4860:4860::8888",
		"2606:4700:4700::1111",
		"::1234:5678",
		"2001:db8:85a3:0:0:8a2e:370:7334",
		"2001:0db8:0001:0000:0000:0ab9:C0A8:0102",
		"::ffff:192.0.2.1", // IPv4-mapped IPv6
	}

	for _, ip := range validIPAddresses {
		t.Run(fmt.Sprintf("valid ip: %s", ip), func(t *testing.T) {
			funcCheck := validator.ValidateIPAddress()
			err := funcCheck(ip)
			assert.NoError(t, err)
		})
	}

	invalidIPAddresses := []string{
		// Invalid format
		"256.256.256.256",
		"192.168.1",
		"192.168.1.1.1",
		"not-an-ip",
		"",
		"hello.world",
		"192.168.1.a",
		"gggg::1",
		// Loopback addresses (not allowed)
		"127.0.0.1",
		"127.0.0.2",
		"127.255.255.255",
		"::1", // IPv6 loopback
		// Unspecified addresses (not allowed)
		"0.0.0.0",
		"::", // IPv6 unspecified
	}

	for _, ip := range invalidIPAddresses {
		t.Run(fmt.Sprintf("invalid ip: %s", ip), func(t *testing.T) {
			funcCheck := validator.ValidateIPAddress()
			err := funcCheck(ip)
			assert.Error(t, err)
		})
	}
}
