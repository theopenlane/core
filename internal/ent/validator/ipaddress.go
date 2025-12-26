package validator

import (
	"github.com/theopenlane/common/models"
	"github.com/theopenlane/utils/rout"
)

// ValidateIPAddress validates an IP address and returns an error if it is invalid
// It accepts both IPv4 and IPv6 addresses but does not allow loopback or unspecified IPs
func ValidateIPAddress() func(ipAddress string) error {
	return func(ipAddress string) error {
		if err := models.ValidateIP(ipAddress); err != nil {
			return rout.InvalidField("ip_address")
		}
		return nil
	}
}
