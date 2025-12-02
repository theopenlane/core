package models

import (
	"errors"
	"net"
)

// ValidateIP takes in an ip address and checks if it is usable for a job runner node
func ValidateIP(s string) error {
	ip := net.ParseIP(s)
	if ip == nil {
		return errors.New("invalid ip address") // nolint: err113
	}

	if ip.IsLoopback() || ip.IsUnspecified() {
		return errors.New("you cannot use a loopback address or unspecified IP like 0.0.0.0 and others") // nolint: err113
	}

	return nil
}
