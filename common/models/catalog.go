package models

// Catalog contains all modules and addons offered by Openlane
type Catalog struct {
	// Version is the version of the catalog, following semantic versioning
	// It is used to track changes and updates to the catalog structure and content.
	// Example: "1.0.0", "2.3.1"
	Version string `json:"version" yaml:"version" jsonschema:"description=Catalog version,example=1.0.0"`
	// SHA is the SHA256 hash of the catalog version string, used to verify integrity
	SHA string `json:"sha" yaml:"sha" jsonschema:"description=SHA of the catalog version"`
	// Modules is a set of purchasable modules available in the catalog
	// Each module has its own set of features, pricing, and audience targeting.
	// Example: "compliance", "reporting", "analytics"
	Modules FeatureSet `json:"modules" yaml:"modules" jsonschema:"description=Set of modules available in the catalog"`
	// Addons is a set of purchasable addons available in the catalog
	Addons FeatureSet `json:"addons" yaml:"addons" jsonschema:"description=Set of addons available in the catalog"`
}

// FeatureSet is a mapping of feature identifiers to metadata
type FeatureSet map[string]Feature

// ItemPrice describes a single price option for a module or addon
type ItemPrice struct {
	Interval   string            `json:"interval" yaml:"interval" jsonschema:"enum=year,enum=month,description=Billing interval for the price,example=month"`
	UnitAmount int64             `json:"unit_amount" yaml:"unit_amount" jsonschema:"description=Amount to be charged per interval,example=1000"`
	Nickname   string            `json:"nickname,omitempty" yaml:"nickname,omitempty" jsonschema:"description=Optional nickname for the price,example=price_compliance_monthly"`
	LookupKey  string            `json:"lookup_key,omitempty" yaml:"lookup_key,omitempty" jsonschema:"description=Optional lookup key for referencing the price,example=price_compliance_monthly,pattern=^[a-z0-9_]+$"`
	Metadata   map[string]string `json:"metadata,omitempty" yaml:"metadata,omitempty" jsonschema:"description=Additional metadata for the price,example={\"tier\":\"premium\"}"`
	PriceID    string            `json:"price_id,omitempty" yaml:"price_id,omitempty" jsonschema:"description=Stripe price ID,example=price_1N2Yw2A1b2c3d4e5"`
}

// Billing contains one or more price options for a module or addon
type Billing struct {
	// Prices is a list of price options for the feature, each with its own billing interval and amount
	Prices []ItemPrice `json:"prices" yaml:"prices" jsonschema:"description=List of price options for this feature"`
}

// Feature defines a purchasable module or addon feature
type Feature struct {
	// DisplayName is the human-readable name for the feature
	DisplayName string `json:"display_name" yaml:"display_name" jsonschema:"description=Human-readable name for the feature,example=Advanced Reporting"`
	// LookupKey is a stable identifier for the feature, used for referencing in Stripe
	// and other systems. It should be lowercase, alphanumeric, and can include underscores or dashes.
	// Example: "compliance", "advanced_reporting"
	// Pattern: ^[a-z0-9_-]+$
	LookupKey string `json:"lookup_key,omitempty" yaml:"lookup_key,omitempty" jsonschema:"description=Stable identifier for the feature,example=compliance,pattern=^[a-z0-9_-]+$"`
	// Description provides additional context about the feature
	Description string `json:"description" yaml:"description,omitempty" jsonschema:"description=Optional description of the feature,example=Provides advanced analytics and reporting capabilities"`
	// MarketingDescription is a longer description of the feature used for marketing material
	MarketingDescription string `json:"marketing_description,omitempty" yaml:"marketing_description,omitempty" jsonschema:"description=Optional long description of the feature used for marketing material,example=Automate evidence collection and task tracking to simplify certification workflows."`
	// Billing contains the pricing information for the feature
	Billing Billing `json:"billing" yaml:"billing" jsonschema:"description=Billing information for the feature"`
	// Audience indicates the intended audience for the feature - it can either be "public", "private", or "beta".
	// - "public" features are available to all users
	// - "private" features are restricted to specific users or organizations
	// - "beta" features are in testing and may not be fully stable
	Audience string `json:"audience" yaml:"audience" jsonschema:"enum=public,enum=private,enum=beta,description=Intended audience for the feature,example=public"`
	// Usage defines the usage limits granted by the feature, such as storage or record counts
	Usage *Usage `json:"usage,omitempty" yaml:"usage,omitempty" jsonschema:"description=Usage limits granted by the feature"`
	// ProductID is the Stripe product ID associated with this feature
	ProductID string `json:"product_id,omitempty" yaml:"product_id,omitempty" jsonschema:"description=Stripe product ID"`
	// PersonalOrg indicates if the feature should be automatically added to personal organizations
	PersonalOrg bool `json:"personal_org,omitempty" yaml:"personal_org,omitempty" jsonschema:"description=Include feature in personal organizations"`
	// IncludeWithTrial indicates if the feature should be automatically included with trial subscriptions
	IncludeWithTrial bool `json:"include_with_trial,omitempty" yaml:"include_with_trial,omitempty" jsonschema:"description=Include feature with trial subscriptions"`
}

// Usage defines usage limits granted by a feature.
type Usage struct {
	// EvidenceStorageGB is the storage limit in GB for evidence related to the feature
	EvidenceStorageGB int64 `json:"evidence_storage_gb,omitempty" yaml:"evidence_storage_gb,omitempty" jsonschema:"description=Storage limit in GB for evidence,example=10"`
	// RecordCount is the maximum number of records allowed for the feature
	RecordCount int64 `json:"record_count,omitempty" yaml:"record_count,omitempty" jsonschema:"description=Maximum number of records allowed,example=1000"`
}
