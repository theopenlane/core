package objectstore

// Config holds operator-level credentials for the object storage definition
type Config struct {
	// AccessKeyID is the AWS access key ID for Openlane's source identity used when assuming customer roles
	AccessKeyID string `json:"accessKeyId" koanf:"accesskeyid" sensitive:"true"`
	// SecretAccessKey is the AWS secret access key for Openlane's source identity
	SecretAccessKey string `json:"secretAccessKey" koanf:"secretaccesskey" sensitive:"true"`
	// ARN is the Openlane principal customers allow in the trust policy of the role Openlane assumes
	ARN string `json:"arn" koanf:"arn"`
	// RuntimeOnly hides the definition from the catalog and refuses organization installs, leaving only the runtime integration
	RuntimeOnly bool `json:"runtimeOnly" koanf:"runtimeonly"`
}
