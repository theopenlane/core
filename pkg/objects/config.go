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
	Provider string `json:"provider" koanf:"provider"`
	// AccessKey is the access key for the storage provider
	AccessKey string `json:"accessKey" koanf:"accessKey"`
	// Region is the region for the storage provider
	Region string `json:"region" koanf:"region"`
	// SecretKey is the secret key for the storage provider
	SecretKey string `json:"secretKey" koanf:"secretKey"`
	// CredentialsJSON is the credentials JSON for the storage provider
	CredentialsJSON string `json:"credentialsJSON" koanf:"credentialsJSON"`
	// Bucket is the bucket name for the storage provider
	Bucket string `json:"bucket" koanf:"bucket"`
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

	defaultFileUploadMaxSize int64 = 10 << 20
	defaultMaxMemorySize     int64 = 32 << 20

	defaultErrorResponseHandler ErrResponseHandler = func(err error, statusCode int) http.HandlerFunc {
		return func(w http.ResponseWriter, _ *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(statusCode)
			fmt.Fprintf(w, `{"message" : "could not upload file", "error" : "%s"}`, err.Error())
		}
	}
)

var OrganizationNameFunc NameGeneratorFunc = func(s string) string {
	return s
}
