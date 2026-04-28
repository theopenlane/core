package awssecurityhub

// Config holds operator-level credentials for the AWS Security Hub definition
type Config struct {
	// AccessKeyID is the AWS access key ID for Openlane's source identity used when assuming cross-account roles
	AccessKeyID string `json:"accessKeyId" koanf:"accesskeyid" sensitive:"true"`
	// SecretAccessKey is the AWS secret access key for Openlane's source identity
	SecretAccessKey string `json:"secretAccessKey" koanf:"secretaccesskey" sensitive:"true"`
	// ARN is the Openlane ARN that is used as the Principal that is allowed to use the assume role
	ARN string `json:"arn" koanf:"arn"`
}
