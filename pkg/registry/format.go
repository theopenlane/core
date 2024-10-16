package registry

import (
	"encoding/base64"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"regexp"
	"time"
)

// FormatFunc validates the customized format in json schema
type FormatFunc func(v interface{}) error

var (
	formatsFuncs = map[string]FormatFunc{
		"urlname":          urlName,
		"httpmethod":       httpMethod,
		"httpmethod-array": httpMethodArray,
		"httpcode":         httpCode,
		"httpcode-array":   httpCodeArray,
		"timerfc3339":      timerfc3339,
		"duration":         duration,
		"ipcidr":           ipcidr,
		"ipcidr-array":     ipcidrArray,
		"hostport":         hostport,
		"regexp":           _regexp,
		"base64":           _base64,
		"url":              _url,
	}

	urlCharsRegexp = regexp.MustCompile(`^[\p{L}0-9\-_\.~]{1,253}$`)
)

func getFormatFunc(format string) (FormatFunc, bool) {
	switch format {
	case "date-time", "email", "hostname", "ipv4", "ipv6", "uri":
		return standardFormat, true

	case "":
		// NOTICE: Empty format does nothing like standard format.
		return standardFormat, true
	}

	if fn, exists := formatsFuncs[format]; exists {
		return fn, true
	}

	return nil, false
}

func standardFormat(v interface{}) error {
	// errors will be reported by standard json schema validation
	return nil
}

var ErrInvalidName = fmt.Errorf("invalid name format")

func urlName(v interface{}) error {
	if urlCharsRegexp.MatchString(v.(string)) {
		return nil
	}

	return fmt.Errorf("%w: %s", ErrInvalidName, v)
}

var ErrInvalidHTTPMethod = fmt.Errorf("invalid http method")

func httpMethod(v interface{}) error {
	switch v.(string) {
	case http.MethodGet,
		http.MethodHead,
		http.MethodPost,
		http.MethodPut,
		http.MethodPatch,
		http.MethodDelete,
		http.MethodConnect,
		http.MethodOptions,
		http.MethodTrace:
		return nil
	default:
		return fmt.Errorf("%w", ErrInvalidHTTPMethod)
	}
}

func httpMethodArray(v interface{}) error {
	for _, method := range v.([]string) {
		err := httpMethod(method)
		if err != nil {
			return err
		}
	}

	return nil
}

var ErrInvalidHTTPCode = fmt.Errorf("invalid http code")

func httpCode(v interface{}) error {
	code := v.(int)
	// Reference: https://tools.ietf.org/html/rfc7231#section-6
	if code < 100 || code >= 600 {
		return fmt.Errorf("%w", ErrInvalidHTTPCode)
	}

	return nil
}

func httpCodeArray(v interface{}) error {
	for _, method := range v.([]int) {
		err := httpCode(method)
		if err != nil {
			return err
		}
	}

	return nil
}

var ErrInvalidRFC3339Time = fmt.Errorf("invalid RFC3339 time")

func timerfc3339(v interface{}) error {
	s := v.(string)

	_, err := time.Parse(time.RFC3339, s)
	if err != nil {
		return fmt.Errorf("%w: %v", ErrInvalidRFC3339Time, err)
	}

	return nil
}

var ErrInvalidDuration = fmt.Errorf("invalid duration")

func duration(v interface{}) error {
	s := v.(string)

	_, err := time.ParseDuration(s)
	if err != nil {
		return fmt.Errorf("%w: %v", ErrInvalidDuration, err)
	}

	return nil
}

var ErrInvalidIPCIDR = fmt.Errorf("invalid ip or cidr")

func ipcidr(v interface{}) error {
	s := v.(string)

	ip := net.ParseIP(s)
	if ip != nil {
		return nil
	}

	_, _, err := net.ParseCIDR(s)
	if err != nil {
		return fmt.Errorf("%w", ErrInvalidIPCIDR)
	}

	return nil
}

func ipcidrArray(v interface{}) error {
	for _, ic := range v.([]string) {
		err := ipcidr(ic)
		if err != nil {
			return err
		}
	}

	return nil
}

var ErrInvalidHostport = fmt.Errorf("invalid hostport")

func hostport(v interface{}) error {
	s := v.(string)

	_, _, err := net.SplitHostPort(s)
	if err != nil {
		return fmt.Errorf("%w: %v", ErrInvalidHostport, err)
	}

	return nil
}

var ErrInvalidRegexp = fmt.Errorf("invalid regular expression")

func _regexp(v interface{}) error {
	s := v.(string)

	_, err := regexp.Compile(s)
	if err != nil {
		return fmt.Errorf("%w: %v", ErrInvalidRegexp, err)
	}

	return nil
}

var ErrInvalidBase64 = fmt.Errorf("invalid base64")

func _base64(v interface{}) error {
	s := v.(string)

	_, err := base64.StdEncoding.DecodeString(s)
	if err != nil {
		return fmt.Errorf("%w: %v", ErrInvalidBase64, err)
	}

	return nil
}

var ErrInvalidURL = fmt.Errorf("invalid url")

func _url(v interface{}) error {
	s := v.(string)

	_, err := url.Parse(s)
	if err != nil {
		return fmt.Errorf("%w: %v", ErrInvalidURL, err)
	}

	return nil
}
