package models

import (
	"errors"
	"io"
	"net"
	"net/url"
	"strings"

	"github.com/theopenlane/core/pkg/enums"
)

var (
	ErrInvalidURL               = errors.New("invalid url provided")
	ErrLocalHostNotAllowed      = errors.New("cannot use localhost url")
	ErrNoLoopbackAddressAllowed = errors.New("no loopback address acceptable")
	ErrUnsupportedJobConfig     = errors.New("we do not support this job type")
	ErrHTTPSOnlyURL             = errors.New("you can only check ssl of a domain with https")
)

type SSLJobConfig struct {
	URL string `json:"url"`
}

type JobConfiguration struct {
	SSL SSLJobConfig `json:"ssl"`
}

// MarshalGQL implement the Marshaler interface for gqlgen
func (job JobConfiguration) MarshalGQL(w io.Writer) {
	marshalGQLJSON(w, job)
}

// UnmarshalGQL implement the Unmarshaler interface for gqlgen
func (job *JobConfiguration) UnmarshalGQL(v interface{}) error {
	return unmarshalGQLJSON(v, job)
}

func (job *JobConfiguration) Validate(typ enums.JobType) error {
	switch typ {
	case enums.JobTypeSsl:
		return job.SSL.Validate()

	default:
		return ErrUnsupportedJobConfig
	}
}

func (s SSLJobConfig) Validate() error {
	_, err := ValidateURL(s.URL)
	return err
}

func ValidateURL(s string) (string, error) {
	if s == "" {
		return "", ErrInvalidURL
	}

	u, err := url.Parse(s)
	if err != nil {
		return "", ErrInvalidURL
	}

	hostname := strings.ToLower(u.Hostname())

	if hostname == "localhost" || hostname == "127.0.0.1" {
		return "", ErrLocalHostNotAllowed
	}

	if ip := net.ParseIP(hostname); ip != nil && ip.IsLoopback() {
		return "", ErrNoLoopbackAddressAllowed
	}

	if u.Scheme != "https" {
		return "", ErrHTTPSOnlyURL
	}

	return s, nil
}
