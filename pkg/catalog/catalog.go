package catalog

import (
	"bytes"
	"context"
	"crypto/sha256"
	_ "embed"
	"encoding/hex"
	"fmt"
	"io/fs"
	"maps"
	"os"

	"github.com/goccy/go-yaml"
	"github.com/pmezard/go-difflib/difflib"
	"github.com/rs/zerolog/log"

	"github.com/stripe/stripe-go/v84"
	"github.com/xeipuuv/gojsonschema"

	"github.com/theopenlane/core/pkg/entitlements"
	"github.com/theopenlane/core/pkg/models"
)

//go:embed genjsonschema/catalog.schema.json
var schemaBytes []byte

// ManagedByKey is the metadata key applied to Stripe resources created via the catalog
const ManagedByKey = "managed_by"

// ManagedByValue identifies objects managed by the catalog automation
const ManagedByValue = "module-manager"

// Price describes a single price option for a module or addon
type Price struct {
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
	Prices []Price `json:"prices" yaml:"prices" jsonschema:"description=List of price options for this feature"`
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

// FeatureSet is a mapping of feature identifiers to metadata
type FeatureSet map[string]Feature

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

// computeSHA returns the hex encoded SHA256 of the provided version string.
func computeSHA(ver string) string {
	sum := sha256.Sum256([]byte(ver))
	return hex.EncodeToString(sum[:])
}

// IsCurrent reports whether the catalog SHA matches its version.
func (c *Catalog) IsCurrent() bool {
	if c == nil {
		return true
	}

	return c.SHA == computeSHA(c.Version)
}

// Visible returns modules and addons filtered by audience
func (c *Catalog) Visible(audience string) *Catalog {
	if c == nil {
		return &Catalog{}
	}

	filter := func(in FeatureSet) FeatureSet {
		if audience == "" {
			return in
		}

		out := FeatureSet{}

		for k, v := range in {
			if v.Audience == audience || v.Audience == "public" {
				out[k] = v
			}
		}

		return out
	}

	return &Catalog{
		Modules: filter(c.Modules),
		Addons:  filter(c.Addons),
	}
}

// LoadCatalog reads and parses a Catalog definition from disk.
func LoadCatalog(path string) (*Catalog, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if !os.IsNotExist(err) {
			return nil, err
		}

		return nil, fs.ErrNotExist
	}

	schema := gojsonschema.NewBytesLoader(schemaBytes)

	jsonData, err := yaml.YAMLToJSON(data)
	if err != nil {
		return nil, ErrYamlToJSONConversion
	}

	doc := gojsonschema.NewBytesLoader(jsonData)

	// this effectively "lints" the catalog to ensure it conforms against the schema
	res, err := gojsonschema.Validate(schema, doc)
	if err != nil {
		return nil, err
	}

	if !res.Valid() {
		log.Debug().Msg("Catalog validation failed - ensure you have generated the latest schema file if you have modified the catalog structs")

		// if there are errors with many fields its easiest to see them this way
		for _, e := range res.Errors() {
			log.Debug().Msgf("Validation error: %s", e)
		}

		return nil, ErrCatalogValidationFailed
	}

	var c Catalog

	if err := yaml.Unmarshal(data, &c); err != nil {
		return nil, err
	}

	return &c, nil
}

// ValidatePrices ensures every feature's price attributes match a Stripe price.
// Matching considers unit amount, interval, nickname, lookup key and metadata.
func (c *Catalog) ValidatePrices(ctx context.Context, sc *entitlements.StripeClient) error {
	prodMap, err := c.getProductMap(ctx, sc)
	if err != nil {
		return err
	}

	check := func(fs FeatureSet) error {
		for name, f := range fs {
			prod, _ := resolveProduct(ctx, sc, prodMap, f)
			if prod == nil {
				return ErrProductMissingFeature
			}

			for i, p := range f.Billing.Prices {
				md := map[string]string{ManagedByKey: ManagedByValue}

				maps.Copy(md, p.Metadata)

				price, err := sc.FindPriceForProduct(ctx, prod.ID, p.PriceID, p.UnitAmount, "", p.Interval, p.Nickname, p.LookupKey, md)
				if err != nil {
					return err
				}

				if price == nil {
					return ErrMatchingPriceNotFound
				}

				f.Billing.Prices[i].PriceID = price.ID
			}

			f.ProductID = prod.ID

			fs[name] = f
		}

		return nil
	}

	if err := check(c.Modules); err != nil {
		return err
	}

	return check(c.Addons)
}

func (c *Catalog) getProductMap(ctx context.Context, sc *entitlements.StripeClient) (map[string]*stripe.Product, error) {
	if c == nil || sc == nil {
		return nil, ErrContextandClientRequired
	}

	products, err := sc.ListProducts(ctx)
	if err != nil {
		log.Error().Err(err).Msg("failed to list products from Stripe")
		return nil, err
	}

	prodMap := map[string]*stripe.Product{}

	for _, p := range products {
		if p.ID != "" {
			prodMap[p.ID] = p
		}

		if p.Name != "" {
			prodMap[p.Name] = p
		}
	}

	return prodMap, nil
}

// EnsurePrices verifies prices exist in Stripe and creates them when missing.
// New products are created using the feature display name and description.
// Matching is performed by unit amount, interval, nickname, lookup key and
// metadata instead of a fixed price ID. The discovered Stripe price ID is stored back in the catalog
// struct but not persisted to disk.
func (c *Catalog) EnsurePrices(ctx context.Context, sc *entitlements.StripeClient, currency string) error {
	prodMap, err := c.getProductMap(ctx, sc)
	if err != nil {
		return err
	}

	create := func(name string, f Feature) (Feature, error) {
		log.Info().Str("feature", name).Msg("ensuring feature prices exist in Stripe")

		prod, _ := resolveProduct(ctx, sc, prodMap, f)

		lookup := name
		if f.LookupKey != "" {
			lookup = f.LookupKey
		}

		var feature *stripe.EntitlementsFeature

		feature, err = sc.GetFeatureByLookupKey(ctx, lookup)
		if err != nil {
			log.Err(err).Msg("failed to get feature by lookup key")
			return f, err
		}

		if prod == nil {
			metadata := map[string]string{
				ManagedByKey: ManagedByValue,
				"module":     name,
			}

			prod, err = sc.CreateProduct(ctx, f.DisplayName, f.Description, metadata)
			if err != nil {
				log.Error().Err(err).Msg("failed to create product")
				return f, ErrFailedToCreateProduct
			}

			prodMap[f.DisplayName] = prod
		} else if prod.Metadata == nil || prod.Metadata["module"] == "" {
			// Product exists, check if it has the module metadata
			// Need to update the product metadata
			existingMetadata := make(map[string]string)

			if prod.Metadata != nil {
				for k, v := range prod.Metadata {
					existingMetadata[k] = v
				}
			}

			// Add missing metadata
			existingMetadata[ManagedByKey] = ManagedByValue
			existingMetadata["module"] = name

			// Update the product with the new metadata
			updateParams := &stripe.ProductUpdateParams{}
			updateParams = sc.UpdateProductWithOptions(updateParams,
				entitlements.WithUpdateProductMetadata(existingMetadata))

			updatedProd, err := sc.UpdateProductWithParams(ctx, prod.ID, updateParams)
			if err != nil {
				log.Error().Err(err).Msg("failed to update product metadata")
				return f, fmt.Errorf("failed to update product metadata: %w", err)
			}

			prod = updatedProd
			prodMap[f.DisplayName] = prod
		}

		if feature == nil {
			log.Info().Str("feature", name).Msg("creating entitlement feature in Stripe")

			feature, err = sc.CreateProductFeatureWithOptions(ctx,
				&stripe.EntitlementsFeatureCreateParams{},
				entitlements.WithFeatureName(f.DisplayName),
				entitlements.WithFeatureLookupKey(lookup),
			)
			if err != nil {
				log.Error().Err(err).Msg("failed to create entitlement feature")
				return f, err
			}

			log.Info().Str("feature", name).Str("id", feature.ID).Msg("created entitlement feature in Stripe")
			_, _ = sc.AttachFeatureToProductWithOptions(ctx,
				&stripe.ProductFeatureCreateParams{},
				entitlements.WithProductFeatureProductID(prod.ID),
				entitlements.WithProductFeatureEntitlementFeatureID(feature.ID),
			)
		}

		f.ProductID = prod.ID

		var monthPriceID, yearPriceID string

		for i, p := range f.Billing.Prices {
			md := map[string]string{ManagedByKey: ManagedByValue}

			maps.Copy(md, p.Metadata)

			price, err := sc.FindPriceForProduct(ctx, prod.ID, p.PriceID, p.UnitAmount, currency, p.Interval, p.Nickname, p.LookupKey, md)
			if err != nil && !notFound(err) {
				return f, err
			}

			if price == nil {
				log.Info().Str("feature", name).Msg("creating missing price in Stripe")

				price, err = sc.CreatePrice(ctx, prod.ID, p.UnitAmount, currency, p.Interval, p.Nickname, p.LookupKey, md)
				if err != nil {
					log.Error().Err(err).Msg("failed to create price")
					return f, ErrFailedToCreatePrice
				}
			}

			f.Billing.Prices[i].PriceID = price.ID

			switch p.Interval {
			case "month":
				monthPriceID = price.ID
			case "year":
				yearPriceID = price.ID
			}
		}

		if monthPriceID != "" && yearPriceID != "" {
			if err := sc.TagPriceUpsell(ctx, monthPriceID, yearPriceID); err != nil {
				log.Error().Err(err).Msg("failed to tag price upsell")
				return f, err
			}
		}

		return f, nil
	}

	ensure := func(fs FeatureSet) error {
		for name, feat := range fs {
			var err error

			feat, err = create(name, feat)
			if err != nil {
				log.Error().Err(err).Msg("failed to create feature")
				return err
			}

			fs[name] = feat
		}

		return nil
	}

	if err := ensure(c.Modules); err != nil {
		log.Error().Err(err).Msg("failed to ensure module features")
		return err
	}

	return ensure(c.Addons)
}

// SaveCatalog writes the catalog to disk in YAML format, as well as computing and updating the SHA
func (c *Catalog) SaveCatalog(path string) (string, error) {
	if c == nil {
		return "", nil
	}

	origData, err := os.ReadFile(path)
	if err != nil && !os.IsNotExist(err) {
		return "", err
	}

	var orig Catalog
	if len(origData) > 0 {
		if err := yaml.Unmarshal(origData, &orig); err != nil {
			return "", err
		}
	}

	// carry over version if not set and ensure SHA present
	if c.Version == "" && orig.Version != "" {
		c.Version = orig.Version
	}

	if c.SHA == "" {
		c.SHA = computeSHA(c.Version)
	}

	newData, err := yaml.Marshal(c)
	if err != nil {
		return "", err
	}

	diff := ""

	if !bytes.Equal(origData, newData) {
		ver := c.Version
		if ver == "" {
			ver = models.DefaultRevision
		}

		if bumped, berr := models.BumpPatch(ver); berr == nil {
			c.Version = bumped

			c.SHA = computeSHA(c.Version)

			if newData, err = yaml.Marshal(c); err != nil {
				return "", err
			}
		}

		u := difflib.UnifiedDiff{
			A:        difflib.SplitLines(string(origData)),
			B:        difflib.SplitLines(string(newData)),
			FromFile: "catalog(old)",
			ToFile:   "catalog(new)",
			Context:  3, //nolint:mnd
		}

		diff, _ = difflib.GetUnifiedDiffString(u)
	}

	if err := os.WriteFile(path, newData, 0o644); err != nil { //nolint:gosec,mnd
		return "", err
	}

	return diff, nil
}

// resolveProduct attempts to locate the Stripe product for a feature using the
// most specific information available. It tries any referenced price IDs first,
// then lookup keys, and finally falls back to the feature display name.
func resolveProduct(ctx context.Context, sc *entitlements.StripeClient, prodMap map[string]*stripe.Product, feat Feature) (*stripe.Product, error) {
	for _, p := range feat.Billing.Prices {
		if p.PriceID != "" {
			pr, err := sc.GetPrice(ctx, p.PriceID)
			if err == nil && pr != nil && pr.Product != nil {
				return sc.GetProductByID(ctx, pr.Product.ID)
			}
		}

		if p.LookupKey != "" {
			pr, err := sc.GetPriceByLookupKey(ctx, p.LookupKey)
			if err == nil && pr != nil && pr.Product != nil {
				return sc.GetProductByID(ctx, pr.Product.ID)
			}
		}
	}

	if prod, ok := prodMap[feat.DisplayName]; ok {
		return prod, nil
	}

	return nil, nil
}
