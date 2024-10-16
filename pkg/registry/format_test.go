package registry

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestUrlName(t *testing.T) {
	tests := []struct {
		name    string
		input   interface{}
		wantErr bool
	}{
		{
			name:    "Valid URL name",
			input:   "valid-url-name",
			wantErr: false,
		},
		{
			name:    "Invalid URL name with special characters",
			input:   "invalid-url-name!",
			wantErr: true,
		},
		{
			name:    "Invalid URL name with spaces",
			input:   "invalid url name",
			wantErr: true,
		},
		{
			name:    "Empty URL name",
			input:   "",
			wantErr: true,
		},
		{
			name:    "Non-string input",
			input:   12345,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := urlName(tt.input)
			if tt.wantErr {
				require.Error(t, err)

				return
			}

			require.NoError(t, err)
		})
	}
}
func TestHttpMethod(t *testing.T) {
	tests := []struct {
		name    string
		input   interface{}
		wantErr bool
	}{
		{
			name:    "Valid HTTP GET method",
			input:   http.MethodGet,
			wantErr: false,
		},
		{
			name:    "Valid HTTP POST method",
			input:   http.MethodPost,
			wantErr: false,
		},
		{
			name:    "Invalid HTTP method",
			input:   "INVALID",
			wantErr: true,
		},
		{
			name:    "Non-string input",
			input:   12345,
			wantErr: true,
		},
		{
			name:    "Empty string input",
			input:   "",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := httpMethod(tt.input)
			if tt.wantErr {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
		})
	}
}
func TestHttpMethodArray(t *testing.T) {
	tests := []struct {
		name    string
		input   interface{}
		wantErr bool
	}{
		{
			name:    "Valid HTTP methods array",
			input:   []string{http.MethodGet, http.MethodPost, http.MethodPut},
			wantErr: false,
		},
		{
			name:    "Invalid HTTP methods array with one invalid method",
			input:   []string{http.MethodGet, "INVALID", http.MethodPost},
			wantErr: true,
		},
		{
			name:    "Empty HTTP methods array",
			input:   []string{},
			wantErr: false,
		},
		{
			name:    "Non-array input",
			input:   "not-an-array",
			wantErr: true,
		},
		{
			name:    "Array with non-string elements",
			input:   []interface{}{http.MethodGet, 12345, http.MethodPost},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := httpMethodArray(tt.input)
			if tt.wantErr {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
		})
	}
}
func TestHttpCode(t *testing.T) {
	tests := []struct {
		name    string
		input   interface{}
		wantErr bool
	}{
		{
			name:    "Valid HTTP code 200",
			input:   200,
			wantErr: false,
		},
		{
			name:    "Valid HTTP code 404",
			input:   404,
			wantErr: false,
		},
		{
			name:    "Invalid HTTP code below range",
			input:   99,
			wantErr: true,
		},
		{
			name:    "Invalid HTTP code above range",
			input:   600,
			wantErr: true,
		},
		{
			name:    "Non-integer input",
			input:   "not-an-integer",
			wantErr: true,
		},
		{
			name:    "Empty input",
			input:   nil,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := httpCode(tt.input)
			if tt.wantErr {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
		})
	}
}
func TestTimerfc3339(t *testing.T) {
	tests := []struct {
		name    string
		input   interface{}
		wantErr bool
	}{
		{
			name:    "Valid RFC3339 time",
			input:   "2023-10-01T15:04:05Z",
			wantErr: false,
		},
		{
			name:    "Invalid RFC3339 time",
			input:   "2023-10-01 15:04:05",
			wantErr: true,
		},
		{
			name:    "Empty string input",
			input:   "",
			wantErr: true,
		},
		{
			name:    "Non-string input",
			input:   12345,
			wantErr: true,
		},
		{
			name:    "Nil input",
			input:   nil,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := timerfc3339(tt.input)
			if tt.wantErr {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
		})
	}
}
func TestHttpCodeArray(t *testing.T) {
	tests := []struct {
		name    string
		input   interface{}
		wantErr bool
	}{
		{
			name:    "Valid HTTP codes array",
			input:   []int{200, 404, 500},
			wantErr: false,
		},
		{
			name:    "Invalid HTTP codes array with one invalid code",
			input:   []int{200, 99, 500},
			wantErr: true,
		},
		{
			name:    "Empty HTTP codes array",
			input:   []int{},
			wantErr: false,
		},
		{
			name:    "Non-array input",
			input:   "not-an-array",
			wantErr: true,
		},
		{
			name:    "Array with non-integer elements",
			input:   []interface{}{200, "not-an-integer", 500},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := httpCodeArray(tt.input)
			if tt.wantErr {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
		})
	}
}
func TestDuration(t *testing.T) {
	tests := []struct {
		name    string
		input   interface{}
		wantErr bool
	}{
		{
			name:    "Valid duration",
			input:   "2h45m",
			wantErr: false,
		},
		{
			name:    "Invalid duration format",
			input:   "2 hours",
			wantErr: true,
		},
		{
			name:    "Empty string input",
			input:   "",
			wantErr: true,
		},
		{
			name:    "Non-string input",
			input:   12345,
			wantErr: true,
		},
		{
			name:    "Nil input",
			input:   nil,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := duration(tt.input)
			if tt.wantErr {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
		})
	}
}
func TestIpcidr(t *testing.T) {
	tests := []struct {
		name    string
		input   interface{}
		wantErr bool
	}{
		{
			name:    "Valid IP address",
			input:   "192.168.1.1",
			wantErr: false,
		},
		{
			name:    "Valid CIDR notation",
			input:   "192.168.1.0/24",
			wantErr: false,
		},
		{
			name:    "Invalid IP address",
			input:   "999.999.999.999",
			wantErr: true,
		},
		{
			name:    "Invalid CIDR notation",
			input:   "192.168.1.0/33",
			wantErr: true,
		},
		{
			name:    "Empty string input",
			input:   "",
			wantErr: true,
		},
		{
			name:    "Non-string input",
			input:   12345,
			wantErr: true,
		},
		{
			name:    "Nil input",
			input:   nil,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ipcidr(tt.input)
			if tt.wantErr {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
		})
	}
}
func TestIpcidrArray(t *testing.T) {
	tests := []struct {
		name    string
		input   interface{}
		wantErr bool
	}{
		{
			name:    "Valid IP and CIDR array",
			input:   []string{"192.168.1.1", "192.168.1.0/24"},
			wantErr: false,
		},
		{
			name:    "Invalid IP in array",
			input:   []string{"192.168.1.1", "999.999.999.999"},
			wantErr: true,
		},
		{
			name:    "Invalid CIDR in array",
			input:   []string{"192.168.1.1", "192.168.1.0/33"},
			wantErr: true,
		},
		{
			name:    "Empty array",
			input:   []string{},
			wantErr: false,
		},
		{
			name:    "Non-array input",
			input:   "not-an-array",
			wantErr: true,
		},
		{
			name:    "Array with non-string elements",
			input:   []interface{}{"192.168.1.1", 12345},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ipcidrArray(tt.input)
			if tt.wantErr {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
		})
	}
}
func TestHostport(t *testing.T) {
	tests := []struct {
		name    string
		input   interface{}
		wantErr bool
	}{
		{
			name:    "Valid hostport",
			input:   "localhost:8080",
			wantErr: false,
		},
		{
			name:    "Valid hostport with IP",
			input:   "192.168.1.1:8080",
			wantErr: false,
		},
		{
			name:    "Invalid hostport with missing port",
			input:   "localhost",
			wantErr: true,
		},
		{
			name:    "Empty string input",
			input:   "",
			wantErr: true,
		},
		{
			name:    "Non-string input",
			input:   12345,
			wantErr: true,
		},
		{
			name:    "Nil input",
			input:   nil,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := hostport(tt.input)
			if tt.wantErr {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
		})
	}
}
func TestRegexp(t *testing.T) {
	tests := []struct {
		name    string
		input   interface{}
		wantErr bool
	}{
		{
			name:    "Valid regular expression",
			input:   "^[a-zA-Z0-9]+$",
			wantErr: false,
		},
		{
			name:    "Invalid regular expression",
			input:   "[a-zA-Z0-9+",
			wantErr: true,
		},
		{
			name:    "Empty string input",
			input:   "",
			wantErr: false,
		},
		{
			name:    "Non-string input",
			input:   12345,
			wantErr: true,
		},
		{
			name:    "Nil input",
			input:   nil,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := _regexp(tt.input)
			if tt.wantErr {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
		})
	}
}
func TestBase64(t *testing.T) {
	tests := []struct {
		name    string
		input   interface{}
		wantErr bool
	}{
		{
			name:    "Valid base64 string",
			input:   "aGVsbG8gd29ybGQ=",
			wantErr: false,
		},
		{
			name:    "Invalid base64 string",
			input:   "invalid-base64",
			wantErr: true,
		},
		{
			name:    "Empty string input",
			input:   "",
			wantErr: false,
		},
		{
			name:    "Non-string input",
			input:   12345,
			wantErr: true,
		},
		{
			name:    "Nil input",
			input:   nil,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := _base64(tt.input)
			if tt.wantErr {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
		})
	}
}
func TestUrl(t *testing.T) {
	tests := []struct {
		name    string
		input   interface{}
		wantErr bool
	}{
		{
			name:    "Valid URL",
			input:   "https://www.example.com",
			wantErr: false,
		},
		{
			name:    "Invalid URL",
			input:   "invalid-url",
			wantErr: true,
		},
		{
			name:    "scheme only URL",
			input:   "http://",
			wantErr: true,
		},
		{
			name:    "Empty string input",
			input:   "",
			wantErr: true,
		},
		{
			name:    "Non-string input",
			input:   12345,
			wantErr: true,
		},
		{
			name:    "Nil input",
			input:   nil,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := _url(tt.input)
			if tt.wantErr {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
		})
	}
}
