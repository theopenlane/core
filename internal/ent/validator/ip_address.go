package validator

import (
	"net"

	"github.com/theopenlane/utils/rout"
)

// ValidateIPAddress returns a validator function for ent fields that ensures the
// input is a valid IPv4 or IPv6 address and is neither loopback nor unspecified.
// Empty values pass validation to allow optional fields; use NotEmpty() on the
// schema field when required.
func ValidateIPAddress() func(s string) error {
	return func(s string) error {
		if s == "" {
			return nil
		}

		ip := net.ParseIP(s)
		if ip == nil {
			return rout.InvalidField("ip_address")
		}

		// Reject loopback and unspecified addresses
		if ip.IsLoopback() || ip.IsUnspecified() {
			return rout.InvalidField("ip_address")
		}

		return nil
	}
}
