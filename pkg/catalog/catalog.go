package catalog

import (
	"context"
	_ "embed"
	"io/fs"
	"maps"
	"os"
	"regexp"
	"strings"

	"github.com/goccy/go-yaml"

	"github.com/stripe/stripe-go/v82"
	"github.com/xeipuuv/gojsonschema"

	"github.com/theopenlane/core/pkg/entitlements"
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
	LookupKey  string            `json:"lookup_key,omitempty" yaml:"lookup_key,omitempty" jsonschema:"description=Optional lookup key for referencing the price,example=price_compliance_monthly"`
	Metadata   map[string]string `json:"metadata,omitempty" yaml:"metadata,omitempty" jsonschema:"description=Additional metadata for the price,example={\"tier\":\"premium\"}"`
	PriceID    string            `json:"price_id,omitempty" yaml:"price_id,omitempty" jsonschema:"description=Stripe price ID,example=price_1N2Yw2A1b2c3d4e5"`
}

// Billing contains one or more price options for a module or addon
type Billing struct {
	Prices []Price `json:"prices" yaml:"prices" jsonschema:"description=List of price options for this feature"`
}

// Feature defines a purchasable module or addon feature
type Feature struct {
	DisplayName string  `json:"display_name" yaml:"display_name" jsonschema:"description=Human-readable name for the feature,example=Advanced Reporting"`
	Description string  `json:"description" yaml:"description,omitempty" jsonschema:"description=Optional description of the feature,example=Provides advanced analytics and reporting capabilities"`
	Billing     Billing `json:"billing" yaml:"billing" jsonschema:"description=Billing information for the feature"`
	Audience    string  `json:"audience" yaml:"audience" jsonschema:"enum=public,enum=private,enum=beta,description=Intended audience for the feature,example=public"`
	Usage       *Usage  `json:"usage,omitempty" yaml:"usage" jsonschema:"description=Usage limits granted by the feature"`
}

// Usage defines usage limits granted by a feature.
type Usage struct {
	EvidenceStorageGB int64 `json:"evidence_storage_gb,omitempty" yaml:"evidence_storage_gb,omitempty" jsonschema:"description=Storage limit in GB for evidence,example=10"`
	RecordCount       int64 `json:"record_count,omitempty" yaml:"record_count,omitempty" jsonschema:"description=Maximum number of records allowed,example=1000"`
}

// FeatureSet is a mapping of feature identifiers to metadata.
type FeatureSet map[string]Feature

// Catalog contains all modules and addons offered by Openlane.
type Catalog struct {
	Modules FeatureSet `json:"modules" yaml:"modules" jsonschema:"description=Set of modules available in the catalog"`
	Addons  FeatureSet `json:"addons" yaml:"addons" jsonschema:"description=Set of addons available in the catalog"`
}

// Visible returns modules and addons filtered by audience.
// Providing "" returns everything.
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

	res, err := gojsonschema.Validate(schema, doc)
	if err != nil {
		return nil, err
	}

	if !res.Valid() {
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
	if c == nil || sc == nil {
		return nil
	}

	products, err := sc.ListProducts(ctx)
	if err != nil {
		return err
	}

	prodMap := map[string]*stripe.Product{}

	for _, p := range products {
		prodMap[p.Name] = p
	}

	check := func(fs FeatureSet) error {
		for name, f := range fs {
			prod, ok := prodMap[f.DisplayName]
			if !ok {
				return ErrProductMissingFeature
			}

			for i, p := range f.Billing.Prices {
				md := map[string]string{ManagedByKey: ManagedByValue}

				maps.Copy(md, p.Metadata)

				price, err := sc.FindPriceForProduct(ctx, prod.ID, p.UnitAmount, "", p.Interval, p.Nickname, p.LookupKey, md)
				if err != nil {
					return err
				}

				if price == nil {
					return ErrMatchingPriceNotFound
				}

				f.Billing.Prices[i].PriceID = price.ID
			}

			fs[name] = f
		}

		return nil
	}

	if err := check(c.Modules); err != nil {
		return err
	}

	return check(c.Addons)
}

// makeLookupKey converts a feature or product name into a lookup key.
// It lowercases the string, replaces spaces with underscores, and
// removes characters that are not letters, digits or underscores.
func makeLookupKey(name string) string {
	key := strings.ToLower(name)
	key = strings.ReplaceAll(key, " ", "_")
	re := regexp.MustCompile(`[^a-z0-9_]+`)
	return re.ReplaceAllString(key, "")
}

// EnsurePrices verifies prices exist in Stripe and creates them when missing.
// New products are created using the feature display name and description.
// Matching is performed by unit amount, interval, nickname, lookup key and
// metadata instead of a fixed price ID. The discovered Stripe price ID is stored back in the catalog
// struct but not persisted to disk.
func (c *Catalog) EnsurePrices(ctx context.Context, sc *entitlements.StripeClient, currency string) error {
	if c == nil || sc == nil {
		return nil
	}

	products, err := sc.ListProducts(ctx)
	if err != nil {
		return err
	}

	prodMap := map[string]*stripe.Product{}
	for _, p := range products {
		prodMap[p.Name] = p
	}

	create := func(name string, f Feature) (Feature, error) {
		prod, ok := prodMap[f.DisplayName]
		var err error
		if !ok {
			prod, err = sc.CreateProduct(ctx, name, f.DisplayName, f.Description, map[string]string{ManagedByKey: ManagedByValue})
			if err != nil {
				return f, ErrFailedToCreateProduct
			}

			prodMap[f.DisplayName] = prod

			lookup := makeLookupKey(f.DisplayName)
			feature, ferr := sc.CreateProductFeatureWithOptions(ctx,
				&stripe.EntitlementsFeatureCreateParams{},
				entitlements.WithFeatureName(f.DisplayName),
				entitlements.WithFeatureLookupKey(lookup),
			)

			if ferr == nil {
				_, ferr = sc.AttachFeatureToProductWithOptions(ctx,
					&stripe.ProductFeatureCreateParams{},
					entitlements.WithProductFeatureProductID(prod.ID),
					entitlements.WithProductFeatureEntitlementFeatureID(feature.ID),
				)
			}

			if ferr != nil {
				return f, ferr
			}
		}

		for i, p := range f.Billing.Prices {
			md := map[string]string{ManagedByKey: ManagedByValue}

			maps.Copy(md, p.Metadata)

			price, err := sc.FindPriceForProduct(ctx, prod.ID, p.UnitAmount, currency, p.Interval, p.Nickname, p.LookupKey, md)
			if err != nil {
				return f, err
			}

			if price == nil {
				price, err = sc.CreatePrice(ctx, prod.ID, p.UnitAmount, currency, p.Interval, p.Nickname, p.LookupKey, md)
				if err != nil {
					return f, ErrFailedToCreatePrice
				}
			}

			f.Billing.Prices[i].PriceID = price.ID
		}

		return f, nil
	}

	ensure := func(fs FeatureSet) error {
		for name, feat := range fs {

			var err error

			feat, err = create(name, feat)
			if err != nil {
				return err
			}

			fs[name] = feat
		}

		return nil
	}

	if err := ensure(c.Modules); err != nil {
		return err
	}

	return ensure(c.Addons)
}

// SaveCatalog writes the catalog to disk in YAML format,
// preserving fields that were omitted in the original file.
func (c *Catalog) SaveCatalog(path string) error {
	if c == nil {
		return nil
	}

	// Read the original YAML to preserve omitted fields
	origData, err := os.ReadFile(path)
	if err != nil && !os.IsNotExist(err) {
		return err
	}

	// Unmarshal original YAML into a map for field presence tracking
	var origMap map[string]any
	if len(origData) > 0 {
		_ = yaml.Unmarshal(origData, &origMap)
	}

	// Marshal the current catalog to a map
	newData, err := yaml.Marshal(c)
	if err != nil {
		return err
	}
	var newMap map[string]any
	_ = yaml.Unmarshal(newData, &newMap)

	// Helper to recursively remove fields that were not present in the original
	var prune func(newNode, origNode any)
	prune = func(newNode, origNode any) {
		switch n := newNode.(type) {
		case map[string]any:
			orig, _ := origNode.(map[string]any)
			for k := range n {
				if orig == nil || orig[k] == nil {
					// Remove keys not present in original
					delete(n, k)
				} else {
					prune(n[k], orig[k])
				}
			}
		case []any:
			origArr, _ := origNode.([]any)
			for i := range n {
				var origElem any
				if origArr != nil && i < len(origArr) {
					origElem = origArr[i]
				}
				prune(n[i], origElem)
			}
		}
	}

	// Only prune if original exists
	if len(origMap) > 0 {
		prune(newMap, origMap)
	}

	// Marshal pruned map back to YAML
	finalData, err := yaml.Marshal(newMap)
	if err != nil {
		return err
	}

	return os.WriteFile(path, finalData, 0o644) // nolint:mnd
}
