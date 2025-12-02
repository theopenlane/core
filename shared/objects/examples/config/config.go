//go:build examples

package config

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/knadh/koanf/parsers/yaml"
	"github.com/knadh/koanf/providers/env/v2"
	"github.com/knadh/koanf/providers/file"
	"github.com/knadh/koanf/v2"
	"github.com/mcuadros/go-defaults"
	"github.com/rs/zerolog/log"
)

var (
	defaultConfigFilePath = "./config/.config.yaml"
)

// Config contains the configuration for object storage examples
type Config struct {
	// Openlane contains Openlane-specific configuration
	Openlane OpenlaneConfig `json:"openlane" koanf:"openlane"`
	// Simple contains simple disk storage example configuration
	Simple SimpleConfig `json:"simple" koanf:"simple"`
	// SimpleS3 contains simple S3 storage example configuration
	SimpleS3 SimpleS3Config `json:"simpleS3" koanf:"simples3"`
}

// OpenlaneConfig contains Openlane API and authentication configuration
type OpenlaneConfig struct {
	// BaseURL is the Openlane API base URL
	BaseURL string `json:"baseUrl" koanf:"baseurl" jsonschema:"required" default:"http://localhost:17608"`
	// Email is the registered user email
	Email string `json:"email" koanf:"email" default:""`
	// Password is the user password
	Password string `json:"password" koanf:"password" default:"" sensitive:"true"`
	// Token is the registration/verification token
	Token string `json:"token" koanf:"token" default:"" sensitive:"true"`
	// OrganizationID is the organization ID
	OrganizationID string `json:"organizationId" koanf:"organizationid" default:""`
	// PAT is the Personal Access Token
	PAT string `json:"pat" koanf:"pat" default:"" sensitive:"true"`
}

// SimpleConfig contains configuration for the simple disk storage example
type SimpleConfig struct {
	// Dir is the directory to use for disk storage
	Dir string `json:"dir" koanf:"dir" default:"./tmp/storage"`
	// LocalURL is the local URL for presigned links
	LocalURL string `json:"localUrl" koanf:"localurl" default:"http://localhost:17608/v1/files"`
	// Keep determines whether to keep the storage directory after completion
	Keep bool `json:"keep" koanf:"keep" default:"false"`
}

// SimpleS3Config contains configuration for the simple S3 storage example
type SimpleS3Config struct {
	// Endpoint is the S3 or MinIO endpoint URL
	Endpoint string `json:"endpoint" koanf:"endpoint" default:"http://127.0.0.1:9000"`
	// AccessKey is the access key ID
	AccessKey string `json:"accessKey" koanf:"accesskey" default:"minioadmin" sensitive:"true"`
	// SecretKey is the secret access key
	SecretKey string `json:"secretKey" koanf:"secretkey" default:"minioadmin" sensitive:"true"`
	// Region is the AWS region
	Region string `json:"region" koanf:"region" default:"us-east-1"`
	// Bucket is the bucket name
	Bucket string `json:"bucket" koanf:"bucket" default:"core-simple-s3"`
	// Source is the path to the file to upload
	Source string `json:"source" koanf:"source" default:"assets/sample-data.txt"`
	// Object is the object key in the bucket
	Object string `json:"object" koanf:"object" default:"examples/simple-s3/sample-data.txt"`
	// Download is the destination path for downloaded files
	Download string `json:"download" koanf:"download" default:"output/downloaded-sample.txt"`
	// PathStyle determines whether to use path-style addressing
	PathStyle bool `json:"pathstyle" koanf:"pathstyle" default:"true"`
}

// New creates a Config with default values applied
func New() *Config {
	cfg := &Config{}
	defaults.SetDefaults(cfg)
	return cfg
}

// Load loads configuration from a YAML file and environment variables
func Load(cfgFile *string) (*Config, error) {
	k := koanf.New(".")

	var configPath string
	if cfgFile == nil || *cfgFile == "" {
		examplesDir := getExamplesDir()
		configPath = filepath.Join(examplesDir, defaultConfigFilePath)
	} else {
		configPath = *cfgFile
	}

	if _, err := os.Stat(configPath); err != nil {
		if os.IsNotExist(err) {
			log.Warn().Err(err).Msg("config file not found, proceeding with default configuration")
		}
	}

	conf := New()

	if err := k.Load(file.Provider(configPath), yaml.Parser()); err != nil {
		log.Warn().Err(err).Msg("failed to load config file - ensure the .config.yaml is present and valid or use environment variables to set the configuration")
	}

	if err := k.Unmarshal("", &conf); err != nil {
		log.Error().Err(err).Msg("failed to unmarshal config file")
		return nil, err
	}

	if err := k.Load(env.Provider(".", env.Opt{
		Prefix: "OBJECTS_EXAMPLES_",
		TransformFunc: func(key, v string) (string, any) {
			key = strings.ReplaceAll(strings.ToLower(strings.TrimPrefix(key, "OBJECTS_EXAMPLES_")), "_", ".")

			if strings.Contains(v, ",") {
				return key, strings.Split(v, ",")
			}

			return key, v
		},
	}), nil); err != nil {
		log.Warn().Err(err).Msg("failed to load env vars, some settings may not be applied")
	}

	if err := k.Unmarshal("", &conf); err != nil {
		log.Error().Err(err).Msg("failed to unmarshal env vars")
		return nil, err
	}

	return conf, nil
}

// SaveOpenlaneConfig saves the Openlane configuration to the config file
func SaveOpenlaneConfig(cfg *OpenlaneConfig) error {
	examplesDir := getExamplesDir()
	configPath := filepath.Join(examplesDir, defaultConfigFilePath)

	fullConfig, err := Load(&configPath)
	if err != nil {
		fullConfig = New()
	}

	fullConfig.Openlane = *cfg

	k := koanf.New(".")
	if err := k.Load(file.Provider(configPath), yaml.Parser()); err != nil {
		log.Debug().Err(err).Msg("no existing config file, creating new one")
	}

	data := make(map[string]interface{})
	if err := k.Unmarshal("", &data); err != nil {
		data = make(map[string]interface{})
	}

	openlaneData := map[string]interface{}{
		"baseurl":        cfg.BaseURL,
		"email":          cfg.Email,
		"password":       cfg.Password,
		"token":          cfg.Token,
		"organizationid": cfg.OrganizationID,
		"pat":            cfg.PAT,
	}

	data["openlane"] = openlaneData

	yamlData, err := yaml.Parser().Marshal(data)
	if err != nil {
		return err
	}

	configDir := filepath.Dir(configPath)
	if err := os.MkdirAll(configDir, 0o700); err != nil {
		return err
	}

	if err := os.WriteFile(configPath, yamlData, 0o600); err != nil {
		return err
	}

	return nil
}

// ConfigExists checks if a configuration file exists
func ConfigExists() bool {
	examplesDir := getExamplesDir()
	configPath := filepath.Join(examplesDir, defaultConfigFilePath)
	_, err := os.Stat(configPath)
	return err == nil
}

// DeleteConfig removes the configuration file
func DeleteConfig() error {
	examplesDir := getExamplesDir()
	configPath := filepath.Join(examplesDir, defaultConfigFilePath)
	if err := os.Remove(configPath); err != nil && !os.IsNotExist(err) {
		return err
	}
	return nil
}

func getExamplesDir() string {
	cwd, err := os.Getwd()
	if err != nil {
		return "."
	}

	if strings.Contains(cwd, "pkg/objects/examples") {
		parts := strings.Split(cwd, "pkg/objects/examples")
		if len(parts) > 0 {
			return filepath.Join(parts[0], "pkg/objects/examples")
		}
	}

	return "."
}
