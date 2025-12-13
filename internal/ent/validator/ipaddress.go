package validator

import (
	"net"

	"github.com/theopenlane/utils/rout"
)

// ValidateIPAddress validates an IP address and returns an error if it is invalid
// It accepts both IPv4 and IPv6 addresses but does not allow loopback or unspecified IPs
func ValidateIPAddress() func(ipAddress string) error {
	return func(ipAddress string) error {
		// Parse the IP address
		ip := net.ParseIP(ipAddress)
		if ip == nil {
			return rout.InvalidField("ip_address")
		}

		// Check if the IP is a loopback address
		if ip.IsLoopback() {
			return rout.InvalidField("ip_address")
		}

		// Check if the IP is an unspecified address
		if ip.IsUnspecified() {
			return rout.InvalidField("ip_address")
		}

		return nil
	}
}
