package objects

import (
	"fmt"
	"net/http"
	"time"
)

type Config struct {
	// Enabled indicates if the store is enabled
	Enabled bool `json:"enabled" koanf:"enabled" default:"true"`
	// Provider is the name of the provider
	Provider          string `json:"provider" koanf:"provider"`
	ConfigurationPath string
	SourcePath        string
	ExecutionTime     time.Time
	Storage           map[string]string `json:"storage" koanf:"storage"`
	DatasetType       string            `json:"dataset_type" koanf:"dataset_type"`
	Kind              string            `json:"kind" koanf:"kind"`
	Path              string            `json:"path" koanf:"path"`
	Container         string            `json:"container" koanf:"container"`
	AccessKey         string            `json:"accesskey" koanf:"accesskey"`
	Region            string            `json:"region" koanf:"region"`
	SecretKey         string            `json:"secretkey" koanf:"secretkey"`
	CredentialsJSON   string            `json:"credentials_json" koanf:"credentials_json"`
	Bucket            string            `json:"bucket" koanf:"bucket"`
	Endpoint          string            `json:"endpoint" koanf:"endpoint"`
	DisableSSL        bool              `json:"disable_ssl" koanf:"disable_ssl"`
	ForcePathStyle    bool              `json:"force_path_style" koanf:"force_path_style"`
	PathStyle         bool              `json:"path_style" koanf:"path_style"`
	EndpointStyle     bool              `json:"endpoint_style" koanf:"endpoint_style"`
}

var (
	// allows all file pass through
	defaultValidationFunc ValidationFunc = func(f File) error {
		return nil
	}

	// defaultNameGeneratorFunc uses the objects-158888-originalname to
	// upload files
	defaultNameGeneratorFunc NameGeneratorFunc = func(s string) string {
		return fmt.Sprintf("objects-%d-%s", time.Now().Unix(), s)
	}

	defaultFileUploadMaxSize int64 = 1024 * 1024 * 5

	defaultErrorResponseHandler ErrResponseHandler = func(err error) http.HandlerFunc {
		return func(w http.ResponseWriter, _ *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Fprintf(w, `{"message" : "could not upload file", "error" : %s}`, err.Error())
		}
	}
)

var OrganizationNameFunc NameGeneratorFunc = func(s string) string {
	return fmt.Sprintf("%s", s)
}
