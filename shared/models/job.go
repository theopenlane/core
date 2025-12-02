package models

import (
	"encoding/json"
	"errors"
	"io"
	"net"
	"net/url"
	"strings"
)

var (
	// ErrInvalidURL defines an invalid url
	ErrInvalidURL = errors.New("invalid url provided")
	// ErrLocalHostNotAllowed defines an error where a user tries to run ssl checks on a localhost address
	ErrLocalHostNotAllowed = errors.New("cannot use localhost url")
	// ErrNoLoopbackAddressAllowed defines an error when a user tries to use loopback address
	ErrNoLoopbackAddressAllowed = errors.New("no loopback address acceptable")
	// ErrUnsupportedJobConfig defines an error for a job type we do not support at the moment
	ErrUnsupportedJobConfig = errors.New("we do not support this job type")
	// ErrHTTPSOnlyURL defines an error where a non https url is being used for a ssl check
	ErrHTTPSOnlyURL = errors.New("you can only check ssl of a domain with https")

	errNilJobConfiguration = errors.New("JobConfiguration: UnmarshalJSON on nil pointer")
)

// JobConfiguration allows users configure the parameters that will be
// templated into their scripts that runs in the automated jobs
type JobConfiguration json.RawMessage

// MarshalJSON implements the json.Marshaler interface
func (job JobConfiguration) MarshalJSON() ([]byte, error) {
	if job == nil {
		return []byte("null"), nil
	}

	return []byte(job), nil
}

// UnmarshalJSON implements the json.Unmarshaler interface
func (job *JobConfiguration) UnmarshalJSON(data []byte) error {
	if job == nil {
		return errNilJobConfiguration
	}

	*job = append((*job)[0:0], data...)

	return nil
}

// MarshalGQL implement the Marshaler interface for gqlgen
func (job JobConfiguration) MarshalGQL(w io.Writer) {
	marshalGQLJSON(w, job)
}

// UnmarshalGQL implement the Unmarshaler interface for gqlgen
func (job *JobConfiguration) UnmarshalGQL(v interface{}) error {
	return unmarshalGQLJSON(v, job)
}

// ValidateDownloadURL takes in url and makes sure it is a valid url
// - it must be https
// - it must not be localhost
// - it must not be a loopback address to our machine
func ValidateDownloadURLnloadURL() func(s string) error {
	return validateURL
}

func validateURL(s string) error {
	if s == "" {
		return ErrInvalidURL
	}

	u, err := url.Parse(s)
	if err != nil {
		return ErrInvalidURL
	}

	hostname := strings.ToLower(u.Hostname())

	if hostname == "localhost" || hostname == "127.0.0.1" {
		return ErrLocalHostNotAllowed
	}

	if ip := net.ParseIP(hostname); ip != nil && ip.IsLoopback() {
		return ErrNoLoopbackAddressAllowed
	}

	if u.Scheme != "https" {
		return ErrHTTPSOnlyURL
	}

	return nil
}
