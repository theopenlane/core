package catalog

import (
	"bytes"
	"context"
	"crypto/sha256"
	_ "embed"
	"encoding/hex"
	"io/fs"
	"maps"
	"os"

	"github.com/goccy/go-yaml"
	"github.com/pmezard/go-difflib/difflib"
	"github.com/rs/zerolog/log"

	"github.com/stripe/stripe-go/v82"
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
	Prices []Price `json:"prices" yaml:"prices" jsonschema:"description=List of price options for this feature"`
}

// Feature defines a purchasable module or addon feature
type Feature struct {
	DisplayName string  `json:"display_name" yaml:"display_name" jsonschema:"description=Human-readable name for the feature,example=Advanced Reporting"`
	Description string  `json:"description" yaml:"description,omitempty" jsonschema:"description=Optional description of the feature,example=Provides advanced analytics and reporting capabilities"`
	Billing     Billing `json:"billing" yaml:"billing" jsonschema:"description=Billing information for the feature"`
	Audience    string  `json:"audience" yaml:"audience" jsonschema:"enum=public,enum=private,enum=beta,description=Intended audience for the feature,example=public"`
	Usage       *Usage  `json:"usage,omitempty" yaml:"usage,omitempty" jsonschema:"description=Usage limits granted by the feature"`
}

// Usage defines usage limits granted by a feature.
type Usage struct {
	EvidenceStorageGB int64 `json:"evidence_storage_gb,omitempty" yaml:"evidence_storage_gb,omitempty" jsonschema:"description=Storage limit in GB for evidence,example=10"`
	RecordCount       int64 `json:"record_count,omitempty" yaml:"record_count,omitempty" jsonschema:"description=Maximum number of records allowed,example=1000"`
}

// FeatureSet is a mapping of feature identifiers to metadata
type FeatureSet map[string]Feature

// Catalog contains all modules and addons offered by Openlane
type Catalog struct {
	Version string     `json:"version" yaml:"version" jsonschema:"description=Catalog version,example=1.0.0"`
	SHA     string     `json:"sha" yaml:"sha" jsonschema:"description=SHA of the catalog version"`
	Modules FeatureSet `json:"modules" yaml:"modules" jsonschema:"description=Set of modules available in the catalog"`
	Addons  FeatureSet `json:"addons" yaml:"addons" jsonschema:"description=Set of addons available in the catalog"`
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
		prod, _ := resolveProduct(ctx, sc, prodMap, f)

		var err error

		if prod == nil {
			prod, err = sc.CreateProduct(ctx, name, f.DisplayName, f.Description, map[string]string{ManagedByKey: ManagedByValue})
			if err != nil {
				return f, ErrFailedToCreateProduct
			}

			prodMap[f.DisplayName] = prod

			lookup := name
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

			price, err := sc.FindPriceForProduct(ctx, prod.ID, p.PriceID, p.UnitAmount, currency, p.Interval, p.Nickname, p.LookupKey, md)
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
