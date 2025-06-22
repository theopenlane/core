package catalog

import (
	"context"
	_ "embed"
	"io/fs"
	"maps"
	"os"

	"github.com/goccy/go-yaml"

	"github.com/stripe/stripe-go/v82"
	"github.com/xeipuuv/gojsonschema"

	"github.com/theopenlane/core/pkg/entitlements"
)

//go:embed catalog.schema.json
var schemaBytes []byte

// ManagedByKey is the metadata key applied to Stripe resources created via the catalog.
const ManagedByKey = "managed_by"

// ManagedByValue identifies objects managed by the catalog automation.
const ManagedByValue = "module-manager"

// Billing describes pricing details for a module or addon.
// Price describes a single price option for a module or addon.
type Price struct {
	Interval   string            `json:"interval" yaml:"interval"`
	UnitAmount int64             `json:"unit_amount" yaml:"unit_amount"`
	Nickname   string            `json:"nickname,omitempty" yaml:"nickname,omitempty"`
	LookupKey  string            `json:"lookup_key,omitempty" yaml:"lookup_key,omitempty"`
	Metadata   map[string]string `json:"metadata,omitempty" yaml:"metadata,omitempty"`
	PriceID    string            `json:"-" yaml:"-"`
}

// Billing contains one or more price options for a module or addon.
type Billing struct {
	Prices []Price `json:"prices" yaml:"prices"`
}

// Feature defines a purchasable module or addon feature.
type Feature struct {
	DisplayName string  `json:"display_name" yaml:"display_name"`
	Description string  `json:"description,omitempty" yaml:"description,omitempty"`
	Billing     Billing `json:"billing" yaml:"billing"`
	Audience    string  `json:"audience" yaml:"audience"`
	Usage       Usage   `json:"usage,omitempty" yaml:"usage,omitempty"`
}

// Usage defines usage limits granted by a feature.
type Usage struct {
	EvidenceStorageGB int64 `json:"evidence_storage_gb,omitempty" yaml:"evidence_storage_gb,omitempty"`
}

// FeatureSet is a mapping of feature identifiers to metadata.
type FeatureSet map[string]Feature

// Catalog contains all modules and addons offered by Openlane.
type Catalog struct {
	Modules FeatureSet `json:"modules" yaml:"modules"`
	Addons  FeatureSet `json:"addons" yaml:"addons"`
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
			prod, err = sc.CreateProduct(ctx, f.DisplayName, f.Description, map[string]string{ManagedByKey: ManagedByValue})
			if err != nil {
				return f, ErrFailedToCreateProduct
			}
			prodMap[f.DisplayName] = prod
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
